package callgraph

import "ariadne/internal/classfile"

// ReflectInvokeEdge is one REFLECTION R2 edge to add: a resolved
// java.lang.reflect.Method#invoke call site, connected to its target's
// name on the class real dispatch resolved it to. Plain identity, not
// an internal/reflectinvoke type — same decoupling
// internal/methodhandle's own HandleEdge already uses.
type ReflectInvokeEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetMethod           string
}

// ApplyReflectInvokeEdges adds, for every edge, a synthetic edge from
// its caller to *every* node matching TargetFQCN#TargetMethod
// (IDsByName — overload-insensitive, same reasoning
// ApplyMethodHandleEdges already applies). A caller node missing from
// g, or a target name with no match at all, is skipped rather than
// treated as an error. Returns how many distinct edges were actually
// added, for the CLI to log.
func ApplyReflectInvokeEdges(g *Graph, edges []ReflectInvokeEdge) int {
	added := 0
	for _, e := range edges {
		fromID, ok := g.ID(e.CallerFQCN, e.CallerName, e.CallerDesc)
		if !ok {
			continue
		}
		for _, toID := range g.IDsByName(e.TargetFQCN, e.TargetMethod) {
			if hasEdge(g, fromID, toID, classfile.CallSynthetic, BasisReflection) {
				continue
			}
			g.addEdge(fromID, toID, classfile.CallSynthetic, BasisReflection, &CallSite{PC: e.PC})
			added++
		}
	}
	return added
}
