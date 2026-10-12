package callgraph

import (
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Build interns every method in h and builds the call graph between
// them: a direct edge for every call site classfile.ExtractCalls found,
// plus — for each virtual/interface call whose declared receiver type
// skip doesn't exempt — an edge to the same-named method on every
// transitive subtype of that receiver (CHA expansion, section 3.3/3.6).
//
// Passing nil for skip uses DefaultSkip. Build never fails: a callee
// outside the hierarchy (almost always a JDK type ariadne never read) is
// still interned as a node with no outgoing edges of its own, so paths
// can still end there.
func Build(h *hierarchy.Hierarchy, skip SkipFunc) *Graph {
	if skip == nil {
		skip = DefaultSkip
	}
	g := &Graph{
		idOf:          make(map[string]uint32),
		byClassMethod: make(map[string][]uint32),
	}

	// Both passes iterate SortedFQCNs rather than ranging over h.All()
	// directly: node/edge insertion order is otherwise at the mercy of
	// Go's randomized map iteration, and since a BFS records the first
	// parent it reaches a node through, that randomness would make which
	// of several equally-short paths internal/reach reports nondeterministic
	// between two runs of the exact same analysis.
	classes := h.All()
	fqcns := h.SortedFQCNs()

	// Pass 1: intern every declared method first, so a method with no
	// outgoing calls at all (a simple getter, an empty override) still
	// gets a node — internal/reach needs it there to use as a root.
	for _, fqcn := range fqcns {
		for _, m := range classes[fqcn].Methods {
			g.intern(fqcn, m.Name, m.Descriptor)
		}
	}

	// Pass 2: add edges. A per-method "seen" set dedupes the common case
	// of the same target called from several call sites in one method
	// (or a direct call and a CHA-expanded edge landing on the same
	// node), without the cost of deduping across the whole graph.
	seen := make(map[Edge]bool)
	for _, fqcn := range fqcns {
		ci := classes[fqcn]
		for _, m := range ci.Methods {
			if len(m.Calls) == 0 {
				g.unresolvedDynamic += m.UnresolvedDynamic
				continue
			}
			callerID, _ := g.ID(fqcn, m.Name, m.Descriptor)
			clear(seen)

			for _, call := range m.Calls {
				toID := g.intern(call.Owner, call.Name, call.Desc)
				if addOnce(g, seen, callerID, toID, call.Kind) {
					g.preCHAEdges++
				}

				if call.Kind != classfile.CallVirtual && call.Kind != classfile.CallInterface {
					continue
				}
				if skip(call.Owner) {
					continue
				}
				for _, sub := range h.Subtypes(call.Owner) {
					if skip(sub) {
						continue
					}
					subInfo, ok := h.Lookup(sub)
					if !ok {
						continue
					}
					for _, subMethod := range subInfo.MethodsByName(call.Name) {
						subID := g.intern(sub, subMethod.Name, subMethod.Descriptor)
						addOnce(g, seen, callerID, subID, call.Kind)
					}
				}
			}
			g.unresolvedDynamic += m.UnresolvedDynamic
		}
	}
	// The report wants "how many edges did CHA add", which is just the
	// difference between the final edge count and the direct-call-only
	// count accumulated above.
	g.chaEdges = g.EdgeCount() - g.preCHAEdges

	return g
}

// addOnce adds the edge if seen doesn't already have it, and reports
// whether it actually added a new edge.
func addOnce(g *Graph, seen map[Edge]bool, from, to uint32, kind classfile.CallKind) bool {
	e := Edge{To: to, Kind: kind}
	if seen[e] {
		return false
	}
	seen[e] = true
	g.addEdge(from, to, kind)
	return true
}
