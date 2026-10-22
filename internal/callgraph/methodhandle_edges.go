package callgraph

import "ariadne/internal/classfile"

// HandleEdge is one DYNAMIC_LINKAGE L2 edge to add: a resolved
// MethodHandle#invoke/invokeExact call site, connected to its target's
// name on the class that declares it. Plain identity, not an
// internal/methodhandle type — same decoupling SERVICE_LOADING S2/
// ASYNC_CALLBACKS A2 already use.
type HandleEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetMethod           string
}

// ApplyMethodHandleEdges adds, for every edge, a synthetic edge from
// its caller to *every* node matching TargetFQCN#TargetMethod
// (IDsByName — overload-insensitive, since nothing about a
// MethodHandles.Lookup find* call pins down one descriptor). A caller
// node missing from g, or a target name with no match at all, is
// skipped rather than treated as an error, same spirit as
// ApplyServiceLoaderEdges/ApplyCallbackEdges. Returns how many distinct
// edges were actually added, for the CLI to log.
func ApplyMethodHandleEdges(g *Graph, edges []HandleEdge) int {
	added := 0
	for _, e := range edges {
		fromID, ok := g.ID(e.CallerFQCN, e.CallerName, e.CallerDesc)
		if !ok {
			continue
		}
		for _, toID := range g.IDsByName(e.TargetFQCN, e.TargetMethod) {
			if hasEdge(g, fromID, toID, classfile.CallSynthetic, BasisMethodHandle) {
				continue
			}
			g.addEdge(fromID, toID, classfile.CallSynthetic, BasisMethodHandle, &CallSite{PC: e.PC})
			added++
		}
	}
	return added
}
