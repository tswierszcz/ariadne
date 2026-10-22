package methodhandle

import (
	"fmt"

	"ariadne/internal/hierarchy"
)

// HandleEdgeResult is one ResolvedInvoke connected to the class that
// actually declares the handle's target method — overload-insensitive
// by name alone (hierarchy.ResolveMethod), the same precision level the
// rest of ariadne already uses for CVE-vulnerable-method matching
// (PLAN.md's own convention): a find* call never carries an exact
// descriptor to match against in the first place. A Gap means no class
// on RequestedOwner's own superclass chain declares a method with that
// name at all — on real bytecode this should be rare (the JDK's own
// ReflectiveOperationException aside, the lookup itself would have
// failed at runtime), not evidence the call is unreachable.
type HandleEdgeResult struct {
	ResolvedInvoke ResolvedInvoke
	DeclaringFQCN  string // "" when Gap != ""
	Gap            string
}

// FindHandleEdges resolves, for every ResolvedInvoke with a known
// RequestedOwner, which class actually declares the handle's target.
func FindHandleEdges(h *hierarchy.Hierarchy, resolved []ResolvedInvoke) []HandleEdgeResult {
	var out []HandleEdgeResult
	for _, ri := range resolved {
		if ri.RequestedOwner == "" {
			continue
		}
		r := HandleEdgeResult{ResolvedInvoke: ri}
		declaring, ok := h.ResolveMethod(ri.RequestedOwner, ri.Method)
		if !ok {
			r.Gap = fmt.Sprintf("%s does not resolve a method named %q anywhere in its hierarchy", ri.RequestedOwner, ri.Method)
		} else {
			r.DeclaringFQCN = declaring
		}
		out = append(out, r)
	}
	return out
}

// HandleEdge is one (invoke call site, target method name) pair
// DYNAMIC_LINKAGE L2 wants as a call-graph edge — plain identity, same
// decoupling SERVICE_LOADING S2/ASYNC_CALLBACKS A2 use for
// internal/callgraph. No descriptor: the edge connects to every
// overload of TargetMethod on TargetFQCN (internal/callgraph's own
// IDsByName), since a MethodHandle's own lookup never pinned one down
// exactly either.
type HandleEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetMethod           string
}

// EdgeTargets returns every HandleEdgeResult with a resolved
// DeclaringFQCN as a concrete edge to add — a Gap contributes nothing.
func EdgeTargets(results []HandleEdgeResult) []HandleEdge {
	var out []HandleEdge
	for _, r := range results {
		if r.Gap != "" {
			continue
		}
		s := r.ResolvedInvoke.Site
		out = append(out, HandleEdge{
			CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, PC: s.PC,
			TargetFQCN: r.DeclaringFQCN, TargetMethod: r.ResolvedInvoke.Method,
		})
	}
	return out
}
