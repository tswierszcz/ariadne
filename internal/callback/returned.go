package callback

import (
	"fmt"
	"sort"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// Returned tasks (ASYNC_CALLBACKS a3-4): a task a method *returns* to the
// code that runs it — Jetty's EatWhatYouKill runs (runTask: task.run())
// what ExecutionStrategy.Producer#produce returns; ManagedSelector's
// SelectorProducer returns what Selectable#onSelected returns, and
// ChannelEndPoint#onSelected returns one of its final Runnable fields
// (_runFillable…), each an anonymous class created in its constructor.
// Neither the run() call (a JDK interface, never CHA-expanded) nor A2's
// executor sites see those objects: the value is the result of an
// instance call.
//
// Sites: a Runnable#run/Callable#call call, and A2's cataloged
// registration sites (Executor#execute…), whose task resolves —
//   - a parameter: at the method's callers (a3-2's caller index);
//   - the result of a call: the return values of the method it names —
//     the exact target of invokestatic/invokespecial, or for an
//     invokevirtual/invokeinterface declared to return Runnable/Callable
//     (dataflow traces those) the bodies its receiver's classes run, when
//     those are known (receiverTypes: a `new`, a parameter, a private or
//     final field), else every implementation in the input, up to
//     MaxReturnImpls;
//   - a field: its writers (a final field set in a constructor, or a
//     static field — dataflow.ResolveFieldOneHop);
//   - `new` or a lambda: the task.
//
// Only tasks that come out of at least one method's return are this
// slice's (a direct `new` at the site is A2's, a parameter alone a3-2's
// or L1's). A return of the returning method's own parameter is not
// followed (its callers are every caller of a helper — not one task).
//
// The edge goes from the site to the method the task runs (Runnable#run
// of its class, a lambda's implementation), with condition "produced",
// and is gated (callgraph.NewCreatorGate) on the innermost method whose
// return handed the task out (ChannelEndPoint#onSelected): resolution is
// context-insensitive — a run site gets what every implementation it may
// call returns — so a task counts only once code returning it is
// reachable. Every chain is bounded by maxReturnDepth links and each site
// by maxReturnWork analyses (a site past them is counted, its edges so
// far kept).
const (
	// ReturnedVersion names the rules; a snapshot records it.
	ReturnedVersion = "a3-4"
	// ConditionProduced: the target runs when a method returning the
	// task (the edge's creator) hands it to the site.
	ConditionProduced = "produced"
	// MaxReturnImpls bounds the implementations of one virtual/interface
	// call whose returns are followed; past it the call is a gap.
	MaxReturnImpls = 16
	// maxReturnDepth bounds the nesting of returns, fields and callers
	// followed for one value.
	maxReturnDepth = 8
	// maxReturnWork bounds the method analyses one site may cause.
	maxReturnWork = 256
)

// GatedEdge is one candidate edge for a creator gate
// (callgraph.NewCreatorGate): the site, the method it may run, and the
// methods one of which must be reachable for the edge to count — for a
// returned task (a3-4) the methods returning it, for a listener (a3-5)
// the methods registering it.
type GatedEdge struct {
	CallerFQCN, CallerName, CallerDesc         string
	PC                                         int
	Kind                                       classfile.CallKind
	TargetFQCN, TargetMethod, TargetMethodDesc string
	Creators                                   []MethodID
	Condition                                  string
}

// MethodID names one method.
type MethodID struct{ FQCN, Name, Desc string }

// ReturnedStats summarizes FindReturnedTasks for the run log.
type ReturnedStats struct {
	// Sites: Runnable#run/Callable#call calls and cataloged
	// registration sites analyzed; Resolved: those with at least one
	// returned task.
	Sites, Resolved int
	// Returners: distinct methods whose return values handed out a task.
	Returners int
	// TooManyImpls: virtual/interface calls with more than
	// MaxReturnImpls implementations (not followed); Unresolved: field
	// or call values that resolved to nothing usable; Budget: sites
	// whose resolution hit maxReturnWork or maxReturnDepth.
	TooManyImpls, Unresolved, Budget int
}

// taskObj is one task a value may be: its identity and the innermost
// method whose return handed it out ("" fqcn: none — not this slice's).
type taskObj struct {
	rc      ResolvedCallback
	creator handoffMethod
}

type returnedResolver struct {
	h     *hierarchy.Hierarchy
	fw    *forwarder
	stats *ReturnedStats
	// returns: method key → the tasks its return values may be
	// (context-free: a returned parameter is not followed); a key mapped
	// to nil while its own returns are being resolved breaks cycles.
	returns   map[string][]taskObj
	returners map[string]bool
	work      int
	exhausted bool
}

const (
	runDesc  = "()V"
	callDesc = "()Ljava/lang/Object;"
)

func isTaskRun(c classfile.Call) bool {
	if c.Kind != classfile.CallInterface && c.Kind != classfile.CallVirtual {
		return false
	}
	return (c.Owner == fqcnRunnable && c.Name == "run" && c.Desc == runDesc) ||
		(c.Owner == fqcnCallable && c.Name == "call" && c.Desc == callDesc)
}

// FindReturnedTasks finds the sites whose task comes out of a method's
// return and their candidate edges (see above). sites are A2's
// registration sites (FindRegistrationSites); only cataloged ones are
// used.
func FindReturnedTasks(h *hierarchy.Hierarchy, sites []RegistrationSite) ([]GatedEdge, ReturnedStats, []string) {
	var (
		stats    ReturnedStats
		warnings []string
		jar      classfile.JarClassReader
	)
	defer jar.Close()
	r := &returnedResolver{h: h, fw: newForwarder(h, &jar), stats: &stats, returns: map[string][]taskObj{}, returners: map[string]bool{}}

	type site struct {
		at   handoffMethod
		m    hierarchy.Method
		pc   int
		kind classfile.CallKind
		// arg: the task's argument index, -1 for the receiver.
		arg                       int
		iface, method, methodDesc string
	}
	var all []site
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if ci.SourceJar == "" {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if !isTaskRun(c) {
					continue
				}
				all = append(all, site{at: handoffMethod{fqcn, m.Name, m.Descriptor}, m: m, pc: c.PC, kind: c.Kind, arg: -1,
					iface: c.Owner, method: c.Name, methodDesc: c.Desc})
			}
		}
	}
	for _, s := range sites {
		if s.Outcome != OutcomeCataloged {
			continue
		}
		ci, ok := h.Lookup(s.CallerFQCN)
		if !ok || ci.SourceJar == "" {
			continue
		}
		m, ok := ci.MethodByNameAndDesc(s.CallerName, s.CallerDesc)
		if !ok {
			continue
		}
		all = append(all, site{at: handoffMethod{s.CallerFQCN, s.CallerName, s.CallerDesc}, m: m, pc: s.PC, kind: classfile.CallSynthetic,
			arg: s.Model.CallbackArgIndex, iface: s.Model.CallbackInterface, method: s.Model.CallbackMethod, methodDesc: s.Model.CallbackMethodDesc})
	}

	var edges []GatedEdge
	seen := map[string]int{}
	for _, s := range all {
		stats.Sites++
		facts := r.fw.analyze(s.at.fqcn, s.m)
		if facts == nil {
			continue
		}
		at := bytecodeir.Site{Method: s.at.fqcn + "#" + s.at.name, Descriptor: s.at.desc, PC: s.pc, Op: bytecodeir.OpCall}
		var v valueflow.Value
		if s.arg < 0 {
			v = receiverAt(facts, at)
		} else {
			var ok bool
			if v, ok = argumentAt(facts, at, s.arg); !ok {
				continue
			}
		}
		r.work, r.exhausted = 0, false
		objs := r.value(v, s.at, 0, handoffMethod{})
		if r.exhausted {
			stats.Budget++
		}
		resolved := false
		for _, o := range objs {
			if o.creator.fqcn == "" {
				continue
			}
			t, ok := r.runs(o.rc, s.iface, s.method, s.methodDesc)
			if !ok {
				continue
			}
			resolved = true
			k := fmt.Sprintf("%s@%d>%s", s.at.key(), s.pc, t.key())
			creator := MethodID{o.creator.fqcn, o.creator.name, o.creator.desc}
			if i, ok := seen[k]; ok {
				if !hasCreator(edges[i].Creators, creator) {
					edges[i].Creators = append(edges[i].Creators, creator)
				}
				continue
			}
			seen[k] = len(edges)
			edges = append(edges, GatedEdge{
				CallerFQCN: s.at.fqcn, CallerName: s.at.name, CallerDesc: s.at.desc, PC: s.pc, Kind: s.kind,
				TargetFQCN: t.fqcn, TargetMethod: t.name, TargetMethodDesc: t.desc,
				Creators: []MethodID{creator}, Condition: ConditionProduced,
			})
		}
		if resolved {
			stats.Resolved++
		}
	}
	stats.Returners = len(r.returners)
	for i := range edges {
		sort.Slice(edges[i].Creators, func(a, b int) bool {
			x, y := edges[i].Creators[a], edges[i].Creators[b]
			return x.FQCN+"#"+x.Name+x.Desc < y.FQCN+"#"+y.Name+y.Desc
		})
	}
	sort.SliceStable(edges, func(i, j int) bool {
		a, b := edges[i], edges[j]
		ka := a.CallerFQCN + "#" + a.CallerName + a.CallerDesc + fmt.Sprintf("@%06d", a.PC) + a.TargetFQCN + "#" + a.TargetMethod + a.TargetMethodDesc
		kb := b.CallerFQCN + "#" + b.CallerName + b.CallerDesc + fmt.Sprintf("@%06d", b.PC) + b.TargetFQCN + "#" + b.TargetMethod + b.TargetMethodDesc
		return ka < kb
	})
	return edges, stats, warnings
}

func hasCreator(cs []MethodID, c MethodID) bool {
	for _, x := range cs {
		if x == c {
			return true
		}
	}
	return false
}

// value: the tasks v (in method ctx) may be; creator is the innermost
// returning method so far.
func (r *returnedResolver) value(v valueflow.Value, ctx handoffMethod, depth int, creator handoffMethod) []taskObj {
	if depth > maxReturnDepth {
		r.exhausted = true
		return nil
	}
	switch v.Kind {
	case valueflow.KindAlloc:
		return []taskObj{{rc: ResolvedCallback{TargetFQCN: v.Type}, creator: creator}}
	case valueflow.KindFunctional:
		if v.Func == nil || v.Func.ImplOwner == "" {
			return nil
		}
		return []taskObj{{rc: ResolvedCallback{ImplOwner: v.Func.ImplOwner, ImplName: v.Func.ImplName, ImplDesc: v.Func.ImplDesc,
			SamInterface: v.Func.SamInterface, SamMethod: v.Func.SamMethod}, creator: creator}}
	case valueflow.KindSet:
		var out []taskObj
		for _, c := range v.Candidates {
			out = append(out, r.value(c, ctx, depth, creator)...)
		}
		return out
	case valueflow.KindTyped:
		if v.Parameter == nil || v.Parameter.Index < 0 || ctx.fqcn == "" {
			return nil
		}
		return r.parameter(ctx, v.Parameter.Index, depth, creator)
	case valueflow.KindCallResult:
		if v.Call == nil {
			return nil
		}
		decl, ok := r.h.ResolveExact(v.Call.Owner, v.Call.Name, v.Call.Desc)
		if !ok {
			return nil
		}
		return r.returnsOf(handoffMethod{decl, v.Call.Name, v.Call.Desc}, depth)
	case valueflow.KindCallTrace:
		t := v.CallTrace
		if t == nil || !dataflow.IsTaskReturn(t.Desc) {
			return nil
		}
		var impls []handoffMethod
		if types, ok := r.receiverTypes(t.Receiver, ctx, depth+1); ok {
			impls = r.dispatch(types, t.Name, t.Desc)
		} else if impls, ok = r.implementations(t.Owner, t.Name, t.Desc); !ok {
			r.stats.TooManyImpls++
			return nil
		}
		var out []taskObj
		for _, m := range impls {
			out = append(out, r.returnsOf(m, depth)...)
		}
		return out
	case valueflow.KindField:
		if v.Field == nil {
			return nil
		}
		fv, _ := dataflow.ResolveFieldOneHop(r.h, *v.Field)
		if fv.Kind == valueflow.KindUnknown {
			r.stats.Unresolved++
			return nil
		}
		// A field's writers are a constructor or <clinit>: their
		// parameters are not this method's.
		return r.value(fv, handoffMethod{}, depth+1, creator)
	}
	return nil
}

// parameter: what the callers of ctx pass as its parameter i.
func (r *returnedResolver) parameter(ctx handoffMethod, i, depth int, creator handoffMethod) []taskObj {
	r.fw.index()
	var out []taskObj
	for _, cs := range r.fw.callers[ctx.key()] {
		if !r.spend() {
			break
		}
		facts := r.fw.analyze(cs.fqcn, cs.m)
		if facts == nil {
			continue
		}
		at := bytecodeir.Site{Method: cs.fqcn + "#" + cs.m.Name, Descriptor: cs.m.Descriptor, PC: cs.pc, Op: bytecodeir.OpCall}
		arg, ok := argumentAt(facts, at, i)
		if !ok {
			continue
		}
		out = append(out, r.value(arg, handoffMethod{cs.fqcn, cs.m.Name, cs.m.Descriptor}, depth+1, creator)...)
	}
	return out
}

// returnsOf: the tasks m's return values may be, each with its innermost
// returning method (m itself unless a nested return is more specific).
func (r *returnedResolver) returnsOf(m handoffMethod, depth int) []taskObj {
	k := m.key()
	if objs, ok := r.returns[k]; ok {
		return objs
	}
	r.returns[k] = nil
	if !r.spend() {
		delete(r.returns, k)
		return nil
	}
	ci, ok := r.h.Lookup(m.fqcn)
	if !ok {
		return nil
	}
	hm, ok := ci.MethodByNameAndDesc(m.name, m.desc)
	if !ok {
		return nil
	}
	facts := r.fw.analyze(m.fqcn, hm)
	if facts == nil {
		return nil
	}
	var out []taskObj
	for _, rv := range facts.ReturnValues() {
		// A returned parameter is not followed (see the type comment).
		out = append(out, r.value(rv, handoffMethod{}, depth+1, m)...)
	}
	if len(out) > 0 {
		r.returners[k] = true
	}
	if !r.exhausted {
		r.returns[k] = out
	} else {
		delete(r.returns, k)
	}
	return out
}

func (r *returnedResolver) spend() bool {
	if r.work >= maxReturnWork {
		r.exhausted = true
		return false
	}
	r.work++
	return true
}

// implementations: the bodies a virtual/interface call of owner#name+desc
// may run in the input — the declaring class's own, if concrete, and each
// concrete subtype's resolution; false past MaxReturnImpls.
func (r *returnedResolver) implementations(owner, name, desc string) ([]handoffMethod, bool) {
	seen := map[string]bool{}
	var out []handoffMethod
	for _, c := range append([]string{owner}, r.h.Subtypes(owner)...) {
		ci, ok := r.h.Lookup(c)
		if !ok || ci.IsInterface() || ci.AccessFlags&classfile.AccAbstract != 0 {
			continue
		}
		decl, ok := r.h.ResolveExact(c, name, desc)
		if !ok || seen[decl] {
			continue
		}
		dci, ok := r.h.Lookup(decl)
		if !ok {
			continue
		}
		if dm, ok := dci.MethodByNameAndDesc(name, desc); !ok || len(dm.Code) == 0 {
			continue
		}
		seen[decl] = true
		out = append(out, handoffMethod{decl, name, desc})
		if len(out) > MaxReturnImpls {
			return nil, false
		}
	}
	return out, true
}

// objID is one object a receiver may be: its class, and the `new` that
// created it ("" for `this`, whose allocation is not one site).
type objID struct{ typ, site string }

// receiverTypes: the classes a call's receiver v (in method ctx) may be,
// when every way it gets its value is known (receiverObjs). An empty, ok
// result: nothing in the input supplies a receiver (a strategy nothing
// constructs — Jetty's ProduceConsume — runs no producer's task). false:
// unknown, the caller falls back to CHA.
func (r *returnedResolver) receiverTypes(v *valueflow.Value, ctx handoffMethod, depth int) ([]string, bool) {
	objs, ok := r.receiverObjs(v, ctx, depth)
	if !ok {
		return nil, false
	}
	var out []string
	for _, o := range objs {
		out = append(out, o.typ)
	}
	return out, true
}

// receiverObjs: the objects a receiver v (in method ctx) may be — a
// `new`, `this` (its class and concrete subclasses), a parameter (what
// the callers pass), or a private or final instance field (what its
// class's methods store; a constructor parameter from the callers of
// that constructor). false: some way it gets its value is unknown.
func (r *returnedResolver) receiverObjs(v *valueflow.Value, ctx handoffMethod, depth int) ([]objID, bool) {
	if v == nil || depth > maxReturnDepth || !r.spend() {
		return nil, false
	}
	switch v.Kind {
	case valueflow.KindAlloc:
		site := ""
		if v.AllocSite != nil {
			site = fmt.Sprintf("%s%s@%d", v.AllocSite.Method, v.AllocSite.Descriptor, v.AllocSite.PC)
		}
		return []objID{{v.Type, site}}, true
	case valueflow.KindSet:
		var out []objID
		for i := range v.Candidates {
			os, ok := r.receiverObjs(&v.Candidates[i], ctx, depth)
			if !ok {
				return nil, false
			}
			out = append(out, os...)
		}
		return out, true
	case valueflow.KindTyped:
		if v.Parameter == nil && ctx.fqcn != "" && v.Type == ctx.fqcn && ctx.name != "<clinit>" {
			// `this` (AnalyzeWithParameterRefs types local 0 as the
			// method's class, without a parameter mark).
			if m, ok := r.method(ctx); ok && !m.IsStatic() {
				return r.thisObjs(ctx.fqcn), true
			}
		}
		if v.Parameter == nil || v.Parameter.Index < 0 || ctx.fqcn == "" {
			return nil, false
		}
		r.fw.index()
		var out []objID
		for _, cs := range r.fw.callers[ctx.key()] {
			facts := r.fw.analyze(cs.fqcn, cs.m)
			if facts == nil {
				return nil, false
			}
			at := bytecodeir.Site{Method: cs.fqcn + "#" + cs.m.Name, Descriptor: cs.m.Descriptor, PC: cs.pc, Op: bytecodeir.OpCall}
			arg, ok := argumentAt(facts, at, v.Parameter.Index)
			if !ok {
				return nil, false
			}
			os, ok := r.receiverObjs(&arg, handoffMethod{cs.fqcn, cs.m.Name, cs.m.Descriptor}, depth+1)
			if !ok {
				return nil, false
			}
			out = append(out, os...)
		}
		return out, true
	case valueflow.KindField:
		return r.fieldObjs(*v.Field, depth)
	}
	return nil, false
}

func (r *returnedResolver) method(m handoffMethod) (hierarchy.Method, bool) {
	ci, ok := r.h.Lookup(m.fqcn)
	if !ok {
		return hierarchy.Method{}, false
	}
	return ci.MethodByNameAndDesc(m.name, m.desc)
}

// thisObjs: `this` in a method of fqcn — fqcn, if concrete, and its
// concrete subclasses.
func (r *returnedResolver) thisObjs(fqcn string) []objID {
	var out []objID
	for _, c := range append([]string{fqcn}, r.h.Subtypes(fqcn)...) {
		if ci, ok := r.h.Lookup(c); ok && !ci.IsInterface() && ci.AccessFlags&classfile.AccAbstract == 0 {
			out = append(out, objID{typ: c})
		}
	}
	return out
}

// fieldObjs: receiverObjs of an instance field written only in its own
// class (private: its nest; final: its constructors).
func (r *returnedResolver) fieldObjs(ref valueflow.FieldRef, depth int) ([]objID, bool) {
	decl, ok := r.h.ResolveField(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		return nil, false
	}
	ci, ok := r.h.Lookup(decl)
	if !ok {
		return nil, false
	}
	f, ok := ci.FieldByNameAndDesc(ref.Name, ref.Desc)
	if !ok || f.AccessFlags&classfile.AccStatic != 0 || f.AccessFlags&(classfile.AccPrivate|classfile.AccFinal) == 0 {
		return nil, false
	}
	writers := []string{decl}
	if f.AccessFlags&classfile.AccFinal == 0 {
		// A private field: nestmates (javac 11+) write it directly.
		for _, c := range r.h.SortedFQCNs() {
			if c != decl && nest(c) == nest(decl) {
				writers = append(writers, c)
			}
		}
	}
	var out []objID
	for _, w := range writers {
		wci, _ := r.h.Lookup(w)
		for _, m := range wci.Methods {
			// A final field is assigned only in its constructors.
			if len(m.Code) == 0 || (f.AccessFlags&classfile.AccFinal != 0 && m.Name != "<init>") {
				continue
			}
			if !r.spend() {
				return nil, false
			}
			facts := r.fw.analyze(w, m)
			if facts == nil {
				return nil, false
			}
			for _, fw := range facts.FieldWrites() {
				if fw.Static || fw.Field.Name != ref.Name || fw.Field.Desc != ref.Desc {
					continue
				}
				if d, ok := r.h.ResolveField(fw.Field.Owner, fw.Field.Name, fw.Field.Desc); !ok || d != decl {
					continue
				}
				if fw.Value.Kind == valueflow.KindNull {
					continue
				}
				val := fw.Value
				os, ok := r.receiverObjs(&val, handoffMethod{w, m.Name, m.Descriptor}, depth+1)
				if !ok {
					return nil, false
				}
				out = append(out, os...)
			}
		}
	}
	return out, true
}

// dispatch: the bodies a call of name+desc runs on receivers of types.
func (r *returnedResolver) dispatch(types []string, name, desc string) []handoffMethod {
	seen := map[string]bool{}
	var out []handoffMethod
	for _, t := range types {
		decl, ok := r.h.ResolveExact(t, name, desc)
		if !ok || seen[decl] {
			continue
		}
		seen[decl] = true
		out = append(out, handoffMethod{decl, name, desc})
	}
	return out
}

// runs: the method the task runs for iface#method — its class's
// resolution, or a lambda of that interface method's implementation.
func (r *returnedResolver) runs(rc ResolvedCallback, iface, method, desc string) (handoffMethod, bool) {
	switch {
	case rc.TargetFQCN != "":
		decl, ok := r.h.ResolveExact(rc.TargetFQCN, method, desc)
		if !ok {
			return handoffMethod{}, false
		}
		return handoffMethod{decl, method, desc}, true
	case rc.ImplOwner != "":
		if rc.SamMethod != method {
			return handoffMethod{}, false
		}
		return handoffMethod{rc.ImplOwner, rc.ImplName, rc.ImplDesc}, true
	}
	return handoffMethod{}, false
}
