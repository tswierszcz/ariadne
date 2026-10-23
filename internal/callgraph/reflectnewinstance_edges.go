package callgraph

import "ariadne/internal/classfile"

// ReflectNewInstanceEdge is one REFLECTION R2 edge to add: a resolved
// java.lang.reflect.Constructor#newInstance call site, connected to
// the exact "<init>" its resolved parameter types declared on
// TargetFQCN — ReflectInvokeEdge's own counterpart for allocation, but
// exact (TargetDesc), never a bare name: FindNewInstanceEdges already
// picked the one constructor the resolved types prove, not every
// overload.
type ReflectNewInstanceEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetDesc             string
}

// ApplyReflectNewInstanceEdges adds, for every edge, a synthetic edge
// from its caller to TargetFQCN's own "<init>"TargetDesc node, plus a
// related BasisReflection edge (classfile.CallInit, not CallSynthetic —
// the same Kind/Basis split ApplyServiceLoaderEdges already uses) to
// TargetFQCN's own "<clinit>()V" node if it has one — the same "cele
// konstruktora ... oraz związane inicjalizacje" reasoning
// (SERVICE_LOADING S2), since newInstance triggers class
// initialization exactly as `new` does (JVM_GRAPH G3). A caller or
// target node missing from g is skipped rather than treated as an
// error. Returns how many distinct edges were actually added, for the
// CLI to log.
func ApplyReflectNewInstanceEdges(g *Graph, edges []ReflectNewInstanceEdge) int {
	added := 0
	for _, e := range edges {
		fromID, ok := g.ID(e.CallerFQCN, e.CallerName, e.CallerDesc)
		if !ok {
			continue
		}
		if ctorID, ok := g.ID(e.TargetFQCN, "<init>", e.TargetDesc); ok {
			if !hasEdge(g, fromID, ctorID, classfile.CallSynthetic, BasisReflection) {
				g.addEdge(fromID, ctorID, classfile.CallSynthetic, BasisReflection, &CallSite{PC: e.PC})
				added++
			}
		}
		if clinitID, ok := g.ID(e.TargetFQCN, "<clinit>", "()V"); ok {
			if !hasEdge(g, fromID, clinitID, classfile.CallInit, BasisReflection) {
				g.addEdge(fromID, clinitID, classfile.CallInit, BasisReflection, nil)
				added++
			}
		}
	}
	return added
}
