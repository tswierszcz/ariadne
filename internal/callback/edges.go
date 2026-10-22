package callback

import (
	"fmt"

	"ariadne/internal/hierarchy"
)

// CallbackEdgeResult is ASYNC_CALLBACKS A2's own result: one resolved
// callback identity (A1) connected to the method that will actually
// run, or a Gap explaining why that couldn't be done. For a
// TargetFQCN-based identity (a direct `new`), that's the class that
// actually *declares* CallbackMethod — JVM method resolution walks the
// superclass chain, since TargetFQCN itself may only inherit it; a Gap
// there should be rare on real bytecode (a class the compiler accepted
// as a Runnable/Callable argument must, by Java's own type rules, have
// that method somewhere in its hierarchy — a Gap means ariadne's own
// hierarchy is missing a class on that chain, e.g. an unresolved
// dependency, not that the callback is somehow unreachable). For an
// ImplOwner-based identity (a lambda/method reference,
// DYNAMIC_LINKAGE L0/L1), DeclaringMethod is already exactly that impl
// method — no hierarchy walk needed or possible, since it was resolved
// once via LambdaMetafactory, not CHA.
type CallbackEdgeResult struct {
	ResolvedCallback ResolvedCallback
	DeclaringFQCN    string // "" when Gap != ""
	DeclaringMethod  string // name
	DeclaringDesc    string
	Gap              string
}

// FindCallbackEdges resolves, for every ResolvedCallback with a known
// TargetFQCN or ImplOwner, which method will actually run.
func FindCallbackEdges(h *hierarchy.Hierarchy, resolved []ResolvedCallback) []CallbackEdgeResult {
	var out []CallbackEdgeResult
	for _, rc := range resolved {
		switch {
		case rc.ImplOwner != "":
			// A lambda/method reference: already the exact declaring
			// method, resolved via LambdaMetafactory — never a Gap here,
			// since nothing further needs resolving.
			out = append(out, CallbackEdgeResult{
				ResolvedCallback: rc, DeclaringFQCN: rc.ImplOwner, DeclaringMethod: rc.ImplName, DeclaringDesc: rc.ImplDesc,
			})
		case rc.TargetFQCN != "":
			r := CallbackEdgeResult{ResolvedCallback: rc}
			declaring, ok := h.ResolveExact(rc.TargetFQCN, rc.Site.Model.CallbackMethod, rc.Site.Model.CallbackMethodDesc)
			if !ok {
				r.Gap = fmt.Sprintf("%s does not resolve %s#%s%s anywhere in its hierarchy",
					rc.TargetFQCN, rc.Site.Model.CallbackInterface, rc.Site.Model.CallbackMethod, rc.Site.Model.CallbackMethodDesc)
			} else {
				r.DeclaringFQCN, r.DeclaringMethod, r.DeclaringDesc = declaring, rc.Site.Model.CallbackMethod, rc.Site.Model.CallbackMethodDesc
			}
			out = append(out, r)
		}
	}
	return out
}

// CallbackEdge is one (registration call site, target method) pair
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
			TargetFQCN: r.DeclaringFQCN, TargetMethod: r.DeclaringMethod, TargetMethodDesc: r.DeclaringDesc,
		})
	}
	return out
}
