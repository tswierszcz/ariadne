package callgraph

import "ariadne/internal/classfile"

// CallbackEdge is one ASYNC_CALLBACKS A2 edge to add: a resolved
// callback-registration call site, connected to the class that actually
// declares its callback method. Plain identity, not an
// internal/callback type — same decoupling SERVICE_LOADING S2 already
// uses for ServiceLoaderEdge, so this package doesn't need to import
// internal/callback.
type CallbackEdge struct {
	CallerFQCN, CallerName, CallerDesc         string
	PC                                         int
	TargetFQCN, TargetMethod, TargetMethodDesc string
}

// ApplyCallbackEdges adds, for every edge, a synthetic edge from its
// caller to the callback method on TargetFQCN (which internal/callback's
// own FindCallbackEdges already resolved to the actual declaring
// class). A caller or target node missing from g is skipped rather than
// treated as an error, same spirit as ApplySyntheticEdges/
// ApplyServiceLoaderEdges. Returns how many distinct edges were
// actually added, for the CLI to log.
func ApplyCallbackEdges(g *Graph, edges []CallbackEdge) int {
	added := 0
	for _, e := range edges {
		fromID, ok := g.ID(e.CallerFQCN, e.CallerName, e.CallerDesc)
		if !ok {
			continue
		}
		toID, ok := g.ID(e.TargetFQCN, e.TargetMethod, e.TargetMethodDesc)
		if !ok {
			continue
		}
		if hasEdge(g, fromID, toID, classfile.CallSynthetic, BasisCallback) {
			continue
		}
		g.addEdge(fromID, toID, classfile.CallSynthetic, BasisCallback, &CallSite{PC: e.PC})
		added++
	}
	return added
}
