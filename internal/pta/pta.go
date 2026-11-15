// Package pta is GRAPH_PRECISION P3's first slice: receiver allocation
// narrowing, a bounded points-to analysis over the shared value domain
// (DATA_FLOW D2/D3, internal/valueflow) — not a second stack interpreter.
//
// For a virtual/interface call site the analysis asks which classes the
// receiver can be an instance of, from proof only:
//
//   - a `new T` in the calling method (valueflow.KindAlloc) — the
//     allocation key is the `new` site; its class is exact;
//   - a small set of such values joined by control flow (KindSet);
//   - a static helper's result (KindCallResult, a factory) and a final
//     field's value (KindField: a static final field is written only in
//     its class's <clinit>, an instance final field only in its
//     constructors — dataflow.ResolveField), each resolved to the same
//     kinds, a bounded number of steps;
//   - a parameter of a static or private method or a constructor, when
//     every caller in the input passes such values (dataflow.CallerIndex)
//     and nothing but a direct call enters the method (no root, no
//     reflection/model/synthetic edge): only then is the caller set the
//     complete one.
//
// Context: none — one set per call site, whichever caller. Anything
// else — an array element, a virtual call's result, a non-final field, an
// unknown caller, a budget overflow — is no proof, and the site keeps
// every target RTA kept. With a proof, a CHA edge's site survives only if
// one of the proven classes selects that edge's target (JVM method
// selection, hierarchy.ResolveExact); the direct edge to the call's own
// resolved method is never touched. The proof does not depend on which
// edges the graph has, so one pass is the fixed point.
package pta

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// Version versions the evidence rules. Snapshots record it.
const Version = "p3-1"

// maxDepth bounds the steps through helpers, fields and parameters.
const maxDepth = 3

// maxTypes bounds a proven set; a larger one is no proof.
const maxTypes = 8

// Result says what the narrowing proved and removed.
type Result struct {
	Version string `json:"version"`
	// CallerMethods: reachable methods with at least one CHA edge.
	CallerMethods int `json:"caller_methods"`
	// Sites: their call sites with CHA edges; Proven: of those, sites
	// with a proven receiver type set; Narrowed: proven sites that lost
	// at least one CHA target.
	Sites    int `json:"sites"`
	Proven   int `json:"proven"`
	Narrowed int `json:"narrowed"`
	// ProofBy counts proven sites by the evidence's outermost kind.
	ProofBy map[string]int `json:"proof_by,omitempty"`
	// CHAEdgesBefore/CHAEdgesRemoved/CHASitesRemoved: the CHA edges the
	// graph had after RTA, and what this step removed.
	CHAEdgesBefore  int    `json:"cha_edges_before"`
	CHAEdgesRemoved int    `json:"cha_edges_removed"`
	CHASitesRemoved int    `json:"cha_sites_removed"`
	Duration        string `json:"-"`
}

// Narrow proves receiver types at the CHA call sites of methods reachable
// from roots and removes the CHA sites the proof rules out.
func Narrow(g *callgraph.Graph, h *hierarchy.Hierarchy, roots []uint32) (Result, []string) {
	start := time.Now()
	res := Result{Version: Version, ProofBy: map[string]int{}}
	reach := reachable(g, roots)
	entered := enteredOtherwise(g, roots)

	type siteKey struct {
		from uint32
		pc   int
	}
	type callInfo struct{ name, desc string }
	sites := map[siteKey]callInfo{}
	callers := map[uint32]bool{}
	for id := 0; id < g.NodeCount(); id++ {
		if !reach[id] {
			continue
		}
		for _, e := range g.Edges(uint32(id)) {
			if e.Basis != callgraph.BasisCHA {
				continue
			}
			res.CHAEdgesBefore++
			_, name, desc := splitKey(g.Key(e.To))
			for _, s := range e.Sites {
				sites[siteKey{uint32(id), s.PC}] = callInfo{name, desc}
			}
			callers[uint32(id)] = true
		}
	}
	res.CallerMethods = len(callers)
	res.Sites = len(sites)

	p := &prover{h: h, g: g, idx: dataflow.NewCallerIndex(h), entered: entered, facts: map[uint32]*dataflow.MethodFacts{}}
	defer p.close()
	// allowed[site] = declaring classes the proven types select.
	allowed := map[siteKey]map[string]bool{}
	keys := make([]siteKey, 0, len(sites))
	for k := range sites {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].from != keys[j].from {
			return keys[i].from < keys[j].from
		}
		return keys[i].pc < keys[j].pc
	})
	for _, k := range keys {
		types, by, ok := p.receiverTypes(k.from, k.pc)
		if !ok {
			continue
		}
		c := sites[k]
		decl := map[string]bool{}
		resolved := true
		for _, t := range types {
			d, found := h.ResolveExact(t, c.name, c.desc)
			if !found {
				// An array, or a class outside the input: the JVM's
				// selection is not known here — no proof.
				resolved = false
				break
			}
			decl[d] = true
		}
		if !resolved {
			continue
		}
		res.Proven++
		res.ProofBy[by]++
		allowed[k] = decl
	}

	narrowed := map[siteKey]bool{}
	res.CHAEdgesRemoved, res.CHASitesRemoved = g.PruneCHASites(func(from uint32, e callgraph.Edge, pc int) bool {
		decl, ok := allowed[siteKey{from, pc}]
		if !ok {
			return false
		}
		owner, _, _ := splitKey(g.Key(e.To))
		if decl[owner] {
			return false
		}
		narrowed[siteKey{from, pc}] = true
		return true
	})
	res.Narrowed = len(narrowed)
	if len(res.ProofBy) == 0 {
		res.ProofBy = nil
	}
	res.Duration = time.Since(start).Round(time.Millisecond).String()
	return res, p.warnings
}

type prover struct {
	h        *hierarchy.Hierarchy
	g        *callgraph.Graph
	idx      *dataflow.CallerIndex
	reader   classfile.JarClassReader
	entered  map[uint32]bool
	facts    map[uint32]*dataflow.MethodFacts
	warnings []string
}

func (p *prover) close() {
	p.idx.Close()
	p.reader.Close()
}

func (p *prover) methodFacts(id uint32) *dataflow.MethodFacts {
	if f, ok := p.facts[id]; ok {
		return f
	}
	p.facts[id] = nil
	fqcn, name, desc := splitKey(p.g.Key(id))
	ci, ok := p.h.Lookup(fqcn)
	if !ok {
		return nil
	}
	m, ok := ci.MethodByNameAndDesc(name, desc)
	if !ok {
		return nil
	}
	cf, err := p.reader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		p.warnings = append(p.warnings, fmt.Sprintf("re-reading %s for receiver narrowing: %v", fqcn, err))
		return nil
	}
	f, err := dataflow.AnalyzeWithParameterRefs(cf, fqcn, m.Method)
	if err != nil {
		return nil
	}
	if aborted, _ := f.Aborted(); aborted {
		return nil
	}
	p.facts[id] = f
	return f
}

// receiverTypes proves the classes the receiver at from's call site pc
// can be an instance of; by names the outermost evidence.
func (p *prover) receiverTypes(from uint32, pc int) ([]string, string, bool) {
	f := p.methodFacts(from)
	if f == nil {
		return nil, "", false
	}
	fqcn, name, desc := splitKey(p.g.Key(from))
	v := f.ReceiverAt(bytecodeir.Site{Method: fqcn + "#" + name, Descriptor: desc, PC: pc, Op: bytecodeir.OpCall})
	set := map[string]bool{}
	if !p.prove(v, dataflow.MethodID{Owner: fqcn, Name: name, Desc: desc}, 0, set) || len(set) == 0 || len(set) > maxTypes {
		return nil, "", false
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out, evidenceKind(v), true
}

// prove adds the classes v is proven to be an instance of to set; false
// when v is not fully proven. in is the method v was read in.
func (p *prover) prove(v valueflow.Value, in dataflow.MethodID, depth int, set map[string]bool) bool {
	if depth > maxDepth || len(set) > maxTypes {
		return false
	}
	switch v.Kind {
	case valueflow.KindAlloc:
		if v.Type == "" {
			return false
		}
		set[v.Type] = true
		return true
	case valueflow.KindSet:
		for _, c := range v.Candidates {
			if !p.prove(c, in, depth, set) {
				return false
			}
		}
		return len(v.Candidates) > 0
	case valueflow.KindCallResult:
		r, _ := dataflow.ResolveCallResult(p.h, *v.Call)
		return p.prove(r, dataflow.MethodID{Owner: v.Call.Owner, Name: v.Call.Name, Desc: v.Call.Desc}, depth+1, set)
	case valueflow.KindField:
		if !p.finalField(*v.Field) {
			return false
		}
		r, _ := dataflow.ResolveField(p.h, *v.Field)
		// The writer (a constructor, <clinit>) is not in: a parameter
		// in r is the writer's, not in's — no proof through it.
		return p.prove(r, dataflow.MethodID{}, depth+1, set)
	case valueflow.KindTyped:
		if v.Parameter == nil || in.Owner == "" || !p.completeCallers(in) {
			return false
		}
		pv := p.idx.ResolveParameter(in, v.Parameter.Index)
		if !pv.Complete() {
			return false
		}
		for i, c := range pv.Values {
			at := in
			if i < len(pv.In) {
				at = pv.In[i]
			} else if c.Kind == valueflow.KindTyped {
				// A caller's own parameter, with no record of which
				// caller: no proof.
				return false
			}
			if !p.prove(c, at, depth+1, set) {
				return false
			}
		}
		return true
	}
	return false
}

// finalField: the field is declared final (only <clinit> or the
// constructors write it, so the writers dataflow.ResolveField merges are
// all of them).
func (p *prover) finalField(ref valueflow.FieldRef) bool {
	decl, ok := p.h.ResolveField(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		return false
	}
	ci, ok := p.h.Lookup(decl)
	if !ok {
		return false
	}
	f, ok := ci.FieldByNameAndDesc(ref.Name, ref.Desc)
	return ok && f.AccessFlags&accFinal != 0
}

const (
	accFinal   = 0x0010
	accPrivate = 0x0002
)

// completeCallers: the input's call sites are all of m's callers — m is
// static, private or a constructor, not a root, and no edge but a direct
// or class-init call enters it.
func (p *prover) completeCallers(m dataflow.MethodID) bool {
	id, ok := p.g.ID(m.Owner, m.Name, m.Desc)
	if !ok || p.entered[id] {
		return false
	}
	ci, ok := p.h.Lookup(m.Owner)
	if !ok {
		return false
	}
	hm, ok := ci.MethodByNameAndDesc(m.Name, m.Desc)
	if !ok {
		return false
	}
	return m.Name == "<init>" || hm.IsStatic() || hm.AccessFlags&accPrivate != 0
}

// enteredOtherwise marks the roots and every target of an edge that is
// not a plain direct or class-init call: methods frameworks, reflection
// or models enter, whose callers the input does not show.
func enteredOtherwise(g *callgraph.Graph, roots []uint32) map[uint32]bool {
	out := map[uint32]bool{}
	for _, r := range roots {
		out[r] = true
	}
	for id := 0; id < g.NodeCount(); id++ {
		for _, e := range g.Edges(uint32(id)) {
			switch e.Basis {
			case callgraph.BasisDirect, callgraph.BasisInit, callgraph.BasisCHA:
			default:
				out[e.To] = true
			}
		}
	}
	return out
}

func reachable(g *callgraph.Graph, roots []uint32) []bool {
	seen := make([]bool, g.NodeCount())
	work := append([]uint32(nil), roots...)
	for _, r := range roots {
		if int(r) < len(seen) {
			seen[r] = true
		}
	}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		for _, e := range g.Edges(n) {
			if !seen[e.To] {
				seen[e.To] = true
				work = append(work, e.To)
			}
		}
	}
	return seen
}

func evidenceKind(v valueflow.Value) string {
	switch v.Kind {
	case valueflow.KindAlloc:
		return "allocation"
	case valueflow.KindSet:
		return "allocations"
	case valueflow.KindCallResult:
		return "factory"
	case valueflow.KindField:
		return "final field"
	case valueflow.KindTyped:
		return "parameter"
	}
	return string(v.Kind)
}

// splitKey splits a node key "fqcn#name(desc)".
func splitKey(key string) (fqcn, name, desc string) {
	i := strings.IndexByte(key, '#')
	if i < 0 {
		return key, "", ""
	}
	fqcn, rest := key[:i], key[i+1:]
	j := strings.IndexByte(rest, '(')
	if j < 0 {
		return fqcn, rest, ""
	}
	return fqcn, rest[:j], rest[j:]
}
