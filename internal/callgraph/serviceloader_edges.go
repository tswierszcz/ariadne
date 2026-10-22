package callgraph

import "ariadne/internal/classfile"

// ServiceLoaderEdge is one SERVICE_LOADING S2 edge to add: a resolved
// ServiceLoader.load(...) call site's caller method, to a registered
// provider's constructor. Plain identity rather than a serviceload
// package type, so this package (which owns Graph's unexported
// addEdge) doesn't need to import internal/serviceload — the caller
// (cmd/ariadne) does that join and hands over the result.
type ServiceLoaderEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	// PC is the originating call site's bytecode offset — real evidence
	// carried onto the constructor edge (not the related <clinit> edge,
	// which has no single instruction to point at).
	PC           int
	ProviderFQCN string
}

// ApplyServiceLoaderEdges adds, for every edge, a synthetic edge from its
// caller to the provider's "<init>()V" node, plus a related
// BasisInit/BasisServiceLoader edge to the provider's own "<clinit>()V"
// node if it has one (SERVICE_LOADING S2's own "cele konstruktora ...
// oraz związane inicjalizacje"). A caller or provider node missing from
// g (shouldn't happen — both came from the same Hierarchy g was built
// from) is skipped rather than treated as an error, same spirit as
// ApplySyntheticEdges. Returns how many distinct edges were actually
// added, for the CLI to log.
func ApplyServiceLoaderEdges(g *Graph, edges []ServiceLoaderEdge) int {
	added := 0
	for _, e := range edges {
		fromID, ok := g.ID(e.CallerFQCN, e.CallerName, e.CallerDesc)
		if !ok {
			continue
		}
		if ctorID, ok := g.ID(e.ProviderFQCN, "<init>", "()V"); ok {
			if !hasEdge(g, fromID, ctorID, classfile.CallSynthetic, BasisServiceLoader) {
				g.addEdge(fromID, ctorID, classfile.CallSynthetic, BasisServiceLoader, &CallSite{PC: e.PC})
				added++
			}
		}
		if clinitID, ok := g.ID(e.ProviderFQCN, "<clinit>", "()V"); ok {
			if !hasEdge(g, fromID, clinitID, classfile.CallInit, BasisServiceLoader) {
				g.addEdge(fromID, clinitID, classfile.CallInit, BasisServiceLoader, nil)
				added++
			}
		}
	}
	return added
}
