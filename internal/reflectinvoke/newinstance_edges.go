package reflectinvoke

import (
	"fmt"

	"ariadne/internal/hierarchy"
)

// NewInstanceEdgeResult is one ResolvedNewInstance connected to the
// exact "<init>" the resolved parameter types prove — InvokeEdgeResult's
// own counterpart for allocation, but exact by construction, not just
// by parameter shape on a dispatch target: a constructor is declared
// directly on OwnerFQCN (never inherited, never overridden — JVMS
// 5.4.3.3's own "a class or interface's instance initialization method
// is never inherited"), so there is no hierarchy walk, and the
// resolved Method's own full descriptor (return type always "V") is
// used verbatim as the real call-graph target. A Gap means the
// receiver never resolved, a parameter type this phase can't turn back
// into a descriptor, OwnerFQCN declares no constructor matching the
// resolved parameter types, or a proven argument-count mismatch — the
// same four Gap reasons InvokeEdgeResult already reports, adapted to
// allocation.
type NewInstanceEdgeResult struct {
	ResolvedNewInstance ResolvedNewInstance
	TargetDesc          string // "" when Gap != ""
	Gap                 string
}

// FindNewInstanceEdges resolves, for every ResolvedNewInstance with a
// known OwnerFQCN, the exact "<init>" descriptor its resolved parameter
// types declare on that class — FindInvokeEdges's own counterpart, but
// without ResolveByParams's hierarchy walk, since a constructor is
// never inherited.
func FindNewInstanceEdges(h *hierarchy.Hierarchy, resolved []ResolvedNewInstance) []NewInstanceEdgeResult {
	var out []NewInstanceEdgeResult
	for _, ri := range resolved {
		if ri.OwnerFQCN == "" {
			continue
		}
		r := NewInstanceEdgeResult{ResolvedNewInstance: ri}

		if ri.ArgCountKnown && ri.ArgCount != len(ri.ParamTypes) {
			r.Gap = fmt.Sprintf("proven argument-count mismatch: the literal args array has %d element(s), the resolved Constructor takes %d",
				ri.ArgCount, len(ri.ParamTypes))
			out = append(out, r)
			continue
		}

		paramDescs := make([]string, len(ri.ParamTypes))
		ok := true
		for i, t := range ri.ParamTypes {
			d, descOK := paramDescriptor(t)
			if !descOK {
				ok = false
				break
			}
			paramDescs[i] = d
		}
		if !ok {
			r.Gap = "a parameter type is an array or otherwise not a plain class literal — dispatch resolution by parameter shape is not supported for it"
			out = append(out, r)
			continue
		}

		ci, found := h.Lookup(ri.OwnerFQCN)
		if !found {
			r.Gap = fmt.Sprintf("%s is not present in the analyzed hierarchy", ri.OwnerFQCN)
			out = append(out, r)
			continue
		}
		ctor, found := ci.MethodByNameAndParams("<init>", paramDescs)
		if !found {
			r.Gap = fmt.Sprintf("%s does not declare a constructor with %d parameter(s) matching these types",
				ri.OwnerFQCN, len(paramDescs))
		} else {
			r.TargetDesc = ctor.Descriptor
		}
		out = append(out, r)
	}
	return out
}

// NewInstanceEdge is one (newInstance call site, constructor) pair
// REFLECTION R2 wants as a call-graph edge — InvokeEdge's own
// counterpart, but exact (OwnerFQCN#"<init>"TargetDesc), never
// overload-insensitive: FindNewInstanceEdges already picked the one
// constructor the resolved parameter types prove, so there is nothing
// left to be insensitive about.
type NewInstanceEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetDesc             string
}

// NewInstanceEdgeTargets returns every NewInstanceEdgeResult with a
// resolved TargetDesc as a concrete edge to add — a Gap contributes
// nothing, EdgeTargets's own counterpart.
func NewInstanceEdgeTargets(results []NewInstanceEdgeResult) []NewInstanceEdge {
	var out []NewInstanceEdge
	for _, r := range results {
		if r.Gap != "" {
			continue
		}
		s := r.ResolvedNewInstance.Site
		out = append(out, NewInstanceEdge{
			CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, PC: s.PC,
			TargetFQCN: r.ResolvedNewInstance.OwnerFQCN, TargetDesc: r.TargetDesc,
		})
	}
	return out
}
