package callgraph

import (
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// SyntheticEdge mirrors a profile's synthetic_edges entry (PLAN.md
// section 4): every transitive subtype of FromSubtypesOf gets, for each
// method it itself declares, an edge to the same-named method on To.
// This stands in for a proxy the bytecode has no implementation of at
// all — the motivating case is Spring Data: an app's own
// `OwnerRepository extends JpaRepository<...>` interface has zero
// bytecode subtypes (Spring generates the implementation at runtime), so
// without this edge CHA finds nothing to call and a vulnerability in
// `SimpleJpaRepository` itself would never show up as reachable through
// a repository call, even though that's exactly what runs.
type SyntheticEdge struct {
	FromSubtypesOf string
	To             string
}

// ApplySyntheticEdges adds every edge edges describes to g (which must
// have been built from h) and returns how many distinct edges were
// actually added, for the CLI to log. An edge whose To class, or whose
// matching method on it, isn't part of the analyzed universe (most
// commonly because that dependency isn't actually on the classpath) is
// silently skipped rather than treated as an error — the same spirit as
// Build itself never failing on a callee outside the hierarchy.
func ApplySyntheticEdges(g *Graph, h *hierarchy.Hierarchy, edges []SyntheticEdge) int {
	added := 0
	for _, se := range edges {
		toInfo, ok := h.Lookup(se.To)
		if !ok {
			continue
		}
		for _, sub := range h.Subtypes(se.FromSubtypesOf) {
			subInfo, ok := h.Lookup(sub)
			if !ok {
				continue
			}
			for _, m := range subInfo.Methods {
				fromID, ok := g.ID(sub, m.Name, m.Descriptor)
				if !ok {
					continue // defensive: Build's pass 1 interns every declared method, so this shouldn't happen
				}
				for _, toMethod := range toInfo.MethodsByName(m.Name) {
					toID, ok := g.ID(se.To, toMethod.Name, toMethod.Descriptor)
					if !ok {
						continue
					}
					e := Edge{To: toID, Kind: classfile.CallSynthetic}
					if hasEdge(g, fromID, e) {
						continue
					}
					g.addEdge(fromID, toID, classfile.CallSynthetic)
					added++
				}
			}
		}
	}
	return added
}

func hasEdge(g *Graph, from uint32, e Edge) bool {
	for _, existing := range g.adjacency[from] {
		if existing == e {
			return true
		}
	}
	return false
}
