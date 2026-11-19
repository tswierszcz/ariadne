package callback

import (
	"sort"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// forwarder resolves a registration site's parameter argument at the
// registering method's callers (ASYNC_CALLBACKS a3-2): a task handed
// through methods that only pass it on (SelectorManager#execute(task) →
// Executor#execute(task)) to the API that runs it.
type forwarder struct {
	h      *hierarchy.Hierarchy
	reader *classfile.JarClassReader
	// callers: method key (owner#name+desc, the declaring class as
	// resolved) → the call sites naming it.
	callers map[string][]callerSite
	facts   map[string]*dataflow.MethodFacts
	site    RegistrationSite
	budget  int
}

type callerSite struct {
	fqcn string
	m    hierarchy.Method
	pc   int
}

func newForwarder(h *hierarchy.Hierarchy, r *classfile.JarClassReader) *forwarder {
	return &forwarder{h: h, reader: r, facts: map[string]*dataflow.MethodFacts{}}
}

func (f *forwarder) index() {
	if f.callers != nil {
		return
	}
	f.callers = map[string][]callerSite{}
	for _, fqcn := range f.h.SortedFQCNs() {
		ci, _ := f.h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind == classfile.CallDynamic {
					continue
				}
				owner := c.Owner
				if d, ok := f.h.ResolveExact(c.Owner, c.Name, c.Desc); ok {
					owner = d
				}
				k := owner + "#" + c.Name + c.Desc
				f.callers[k] = append(f.callers[k], callerSite{fqcn: fqcn, m: m, pc: c.PC})
			}
		}
	}
}

// resolve returns the targets callers of method pass as its parameter
// param; chain holds the caller methods already followed.
func (f *forwarder) resolve(method string, param int, chain []string, depth int) []ResolvedCallback {
	if depth == 0 {
		f.budget = forwardCallers
	}
	if depth >= forwardDepth {
		return nil
	}
	f.index()
	var out []ResolvedCallback
	seen := map[string]bool{}
	for _, cs := range f.callers[method] {
		if f.budget <= 0 {
			break
		}
		f.budget--
		key := cs.fqcn + "#" + cs.m.Name + cs.m.Descriptor
		facts := f.analyze(cs.fqcn, cs.m)
		if facts == nil {
			continue
		}
		v, ok := argumentAt(facts, bytecodeir.Site{Method: cs.fqcn + "#" + cs.m.Name, Descriptor: cs.m.Descriptor, PC: cs.pc, Op: bytecodeir.OpCall}, param)
		if !ok {
			continue
		}
		via := append(append([]string(nil), chain...), key)
		var found []ResolvedCallback
		switch {
		case v.Kind == valueflow.KindAlloc:
			found = []ResolvedCallback{{Site: f.site, TargetFQCN: v.Type, Forwarded: via}}
		case v.Kind == valueflow.KindFunctional && v.Func != nil && v.Func.ImplOwner != "":
			found = []ResolvedCallback{{Site: f.site, ImplOwner: v.Func.ImplOwner, ImplName: v.Func.ImplName, ImplDesc: v.Func.ImplDesc, Forwarded: via}}
		case v.Kind == valueflow.KindTyped && v.Parameter != nil && v.Parameter.Index >= 0:
			found = f.resolve(key, v.Parameter.Index, via, depth+1)
		}
		for _, r := range found {
			id := r.TargetFQCN + "|" + r.ImplOwner + "#" + r.ImplName + r.ImplDesc
			if !seen[id] {
				seen[id] = true
				out = append(out, r)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].TargetFQCN+out[i].ImplOwner+out[i].ImplName+out[i].ImplDesc, out[j].TargetFQCN+out[j].ImplOwner+out[j].ImplName+out[j].ImplDesc
		return a < b
	})
	return out
}

func (f *forwarder) analyze(fqcn string, m hierarchy.Method) *dataflow.MethodFacts {
	key := fqcn + "#" + m.Name + m.Descriptor
	if facts, ok := f.facts[key]; ok {
		return facts
	}
	f.facts[key] = nil
	ci, ok := f.h.Lookup(fqcn)
	if !ok || ci.SourceJar == "" || len(m.Code) == 0 {
		return nil
	}
	cf, err := f.reader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		return nil
	}
	facts, err := dataflow.AnalyzeWithParameterRefs(cf, fqcn, m.Method)
	if err != nil {
		return nil
	}
	if aborted, _ := facts.Aborted(); aborted {
		return nil
	}
	f.facts[key] = facts
	return facts
}

func argumentAt(f *dataflow.MethodFacts, at bytecodeir.Site, i int) (v valueflow.Value, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return f.ArgumentAt(at, i), true
}
