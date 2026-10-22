package callback

import (
	"fmt"

	"ariadne/internal/hierarchy"
)

// CallbackEdgeResult is ASYNC_CALLBACKS A2's own result: one resolved
// callback identity (A1) with the class that actually *declares* its
// CallbackMethod (JVM method resolution walks the superclass chain —
// TargetFQCN itself may only inherit the implementation), or a Gap
// explaining why that couldn't be found. A Gap here should be rare on
// real bytecode: a class the compiler accepted as a Runnable/Callable
// argument must, by Java's own type rules, have that method somewhere
// in its hierarchy — a Gap means ariadne's own hierarchy is missing a
// class on that chain (e.g. an unresolved dependency), not that the
// callback is somehow unreachable.
type CallbackEdgeResult struct {
	ResolvedCallback ResolvedCallback
	DeclaringFQCN    string // "" when Gap != ""
	Gap              string
}

// FindCallbackEdges resolves, for every ResolvedCallback with a known
// TargetFQCN, which class actually declares the callback method that
// will run.
func FindCallbackEdges(h *hierarchy.Hierarchy, resolved []ResolvedCallback) []CallbackEdgeResult {
	var out []CallbackEdgeResult
	for _, rc := range resolved {
		if rc.TargetFQCN == "" {
			continue
		}
		r := CallbackEdgeResult{ResolvedCallback: rc}
		declaring, ok := h.ResolveExact(rc.TargetFQCN, rc.Site.Model.CallbackMethod, rc.Site.Model.CallbackMethodDesc)
		if !ok {
			r.Gap = fmt.Sprintf("%s does not resolve %s#%s%s anywhere in its hierarchy",
				rc.TargetFQCN, rc.Site.Model.CallbackInterface, rc.Site.Model.CallbackMethod, rc.Site.Model.CallbackMethodDesc)
		} else {
			r.DeclaringFQCN = declaring
		}
		out = append(out, r)
	}
	return out
}

// CallbackEdge is one (registration call site, callback method) pair
// ASYNC_CALLBACKS A2 wants as a call-graph edge — plain identity, same
// decoupling SERVICE_LOADING S2 uses for internal/callgraph.
type CallbackEdge struct {
	CallerFQCN, CallerName, CallerDesc         string
	PC                                         int
	TargetFQCN, TargetMethod, TargetMethodDesc string
}

// EdgeTargets returns every CallbackEdgeResult with a resolved
// DeclaringFQCN as a concrete edge to add — a Gap contributes nothing,
// never a fabricated target.
func EdgeTargets(results []CallbackEdgeResult) []CallbackEdge {
	var out []CallbackEdge
	for _, r := range results {
		if r.Gap != "" {
			continue
		}
		s := r.ResolvedCallback.Site
		out = append(out, CallbackEdge{
			CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, PC: s.PC,
			TargetFQCN: r.DeclaringFQCN, TargetMethod: s.Model.CallbackMethod, TargetMethodDesc: s.Model.CallbackMethodDesc,
		})
	}
	return out
}
