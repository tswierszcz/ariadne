package callgraph

import (
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// MethodRef names one method of the graph.
type MethodRef struct{ FQCN, Name, Desc string }

// SAMEdge is one DYNAMIC_LINKAGE L1 edge candidate (internal/samcall): a
// call site and the lambda/method-reference implementation it may run.
// Plain identity, the same decoupling as CallbackEdge/HandleEdge.
type SAMEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Kind is the site's own call kind (interface/virtual for a SAM call,
	// any for a call handing the value outside the input).
	Kind                               classfile.CallKind
	TargetFQCN, TargetName, TargetDesc string
	// Virtual: an instance-method reference — the JVM selects the body by
	// the receiver, so the target is expanded like a virtual call on it
	// (the declaring body plus CHA's dispatch targets).
	Virtual bool
	// Creators: the methods whose invokedynamic creates the values the
	// site may receive with this implementation.
	Creators []MethodRef
}

// SAMStats count what a SAMGate did.
type SAMStats struct {
	// Added: edges added. Covered: (caller, target) pairs already in the
	// graph — typically the creating method's own invokedynamic edge,
	// when the value is used where it is created. Gated: candidates left
	// out because no method creating the value is reachable from the
	// roots (the value never exists in a run of the reachable code).
	// Missing: candidates with no node (an implementation outside the
	// input, e.g. a JDK method reference) or no reachable-able creator node.
	Added, Covered, Gated, Missing int
}

type samCandidate struct {
	from, to uint32
	kind     classfile.CallKind
	pc       int
	creators []uint32
	done     bool
}

// SAMGate holds the SAM-call edge candidates until a creator of the value
// is reachable: resolution is context-insensitive (a helper's parameter
// gets what every caller passes), so a lambda whose creating method never
// runs from the roots would otherwise reach a SAM call through a caller
// that does run — a path no execution takes ("an unused lambda does not
// prove its body runs", DYNAMIC_LINKAGE).
type SAMGate struct {
	cands []samCandidate
	stats SAMStats
	basis EdgeBasis
}

// NewSAMGate resolves the candidates to graph nodes. Dispatch targets of
// virtual method references are computed once per owner/name/descriptor,
// like BuildWith's.
func NewSAMGate(g *Graph, h *hierarchy.Hierarchy, edges []SAMEdge) *SAMGate {
	return NewCreatorGate(g, h, edges, BasisSAM)
}

// NewCreatorGate is NewSAMGate for candidate edges of another basis: the
// same "add once a creator of the receiver is reachable" rule, used by a
// library model whose targets depend on which classes reachable code
// creates (internal/servletdispatch's request/response objects, basis
// BasisModel, counted with the other model edges).
func NewCreatorGate(g *Graph, h *hierarchy.Hierarchy, edges []SAMEdge, basis EdgeBasis) *SAMGate {
	s := &SAMGate{basis: basis}
	type key struct{ owner, name, desc string }
	dispatch := map[key][]string{}
	cha := CHA{Skip: DefaultSkip}
	targetsOf := func(e SAMEdge) []string {
		owner, _ := resolveTarget(h, e.TargetFQCN, e.TargetName, e.TargetDesc)
		if !e.Virtual {
			return []string{owner}
		}
		k := key{e.TargetFQCN, e.TargetName, e.TargetDesc}
		if t, ok := dispatch[k]; ok {
			return t
		}
		kind := classfile.CallVirtual
		if ci, ok := h.Lookup(e.TargetFQCN); ok && ci.IsInterface() {
			kind = classfile.CallInterface
		}
		t := append([]string{owner}, cha.DispatchTargets(h, classfile.Call{Kind: kind, Owner: e.TargetFQCN, Name: e.TargetName, Desc: e.TargetDesc})...)
		dispatch[k] = t
		return t
	}
	for _, e := range edges {
		from, ok := g.ID(e.CallerFQCN, e.CallerName, e.CallerDesc)
		if !ok {
			s.stats.Missing++
			continue
		}
		var creators []uint32
		for _, c := range e.Creators {
			if id, ok := g.ID(c.FQCN, c.Name, c.Desc); ok {
				creators = append(creators, id)
			}
		}
		if len(creators) == 0 {
			s.stats.Missing++
			continue
		}
		kind := e.Kind
		if kind == "" {
			kind = classfile.CallSynthetic
		}
		for _, owner := range targetsOf(e) {
			to, ok := g.ID(owner, e.TargetName, e.TargetDesc)
			if !ok {
				s.stats.Missing++
				continue
			}
			s.cands = append(s.cands, samCandidate{from: from, to: to, kind: kind, pc: e.PC, creators: creators})
		}
	}
	return s
}

// Apply adds every pending candidate one of whose creators is reachable
// from roots, to a fixed point (an added edge can make another creator
// reachable). It can be called again after other edges were added (the
// library models); candidates already added are not repeated. It returns
// the running totals.
func (s *SAMGate) Apply(g *Graph, roots []uint32) SAMStats {
	reach := make([]bool, g.NodeCount())
	var queue []uint32
	visit := func(id uint32) {
		if !reach[id] {
			reach[id] = true
			queue = append(queue, id)
		}
	}
	bfs := func() {
		for len(queue) > 0 {
			n := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			for _, e := range g.adjacency[n] {
				visit(e.To)
			}
		}
	}
	for _, r := range roots {
		visit(r)
	}
	bfs()
	for {
		changed := false
		for i := range s.cands {
			c := &s.cands[i]
			if c.done {
				continue
			}
			live := false
			for _, cr := range c.creators {
				live = live || reach[cr]
			}
			if !live {
				continue
			}
			c.done, changed = true, true
			s.add(g, c)
			if reach[c.from] {
				visit(c.to)
			}
		}
		if !changed {
			break
		}
		bfs()
	}
	s.stats.Gated = 0
	for _, c := range s.cands {
		if !c.done {
			s.stats.Gated++
		}
	}
	return s.stats
}

func (s *SAMGate) add(g *Graph, c *samCandidate) {
	idx, other := -1, false
	for i, ex := range g.adjacency[c.from] {
		if ex.To != c.to {
			continue
		}
		if ex.Basis == s.basis && ex.Kind == c.kind {
			idx = i
			break
		}
		other = true
	}
	switch {
	case idx >= 0:
		ex := &g.adjacency[c.from][idx]
		for _, site := range ex.Sites {
			if site.PC == c.pc {
				return
			}
		}
		ex.Sites = append(ex.Sites, CallSite{PC: c.pc})
	case other:
		s.stats.Covered++
	default:
		g.addEdge(c.from, c.to, c.kind, s.basis, &CallSite{PC: c.pc})
		if s.basis == BasisSAM {
			g.samEdges++
		} else {
			g.modelEdges++
		}
		s.stats.Added++
	}
}

// SAMEdgeCount is how many edges SAMGate added (BasisSAM).
func (g *Graph) SAMEdgeCount() int { return g.samEdges }
