package reflectinvoke

import (
	"fmt"
	"strings"

	"ariadne/internal/hierarchy"
)

// InvokeEdgeResult is one ResolvedInvoke connected to the class that
// actually declares the reflected method's real implementation — exact
// by parameter types (hierarchy.ResolveByParams), the precision
// REFLECTION R1's own literal `Class[]` tracking makes possible, unlike
// DYNAMIC_LINKAGE L2's own name-only HandleEdgeResult (a find* call
// never proves parameter types). A Gap means either the receiver never
// resolved, a parameter type this phase can't turn back into a
// descriptor (an array or primitive Class literal — see
// paramDescriptor's own doc comment), no declaring class exists
// anywhere in the dispatch root's hierarchy, or a proven argument-count
// mismatch (R2's own "nie zakładać skutecznego invocation przy
// dowiedzionej niezgodności" — never fabricate a successful invocation
// once incompatibility is proven, even though possession of the Method
// object alone was already real).
type InvokeEdgeResult struct {
	ResolvedInvoke ResolvedInvoke
	DeclaringFQCN  string // "" when Gap != ""
	Gap            string
}

// paramDescriptor turns one getMethod parameter-type Class literal's
// own FQCN back into a field descriptor, so the real target method can
// be found by parameter *shape*, not just name — getMethod itself never
// gives a return type to match against (hierarchy.ResolveByParams's own
// doc comment). Only a plain object/class type is supported: an array
// type's own Class literal is read out of the constant pool already in
// its internal array-descriptor form ("[Ljava/lang/String;"), which
// REFLECTION R1's own binaryNameToFQCN blindly dot-converts into an
// invalid FQCN shape ("[Ljava.lang.String;") rather than a real one —
// recognized here and rejected rather than silently mis-converted back.
func paramDescriptor(fqcn string) (string, bool) {
	if fqcn == "" || strings.ContainsAny(fqcn, "[;") {
		return "", false
	}
	return "L" + strings.ReplaceAll(fqcn, ".", "/") + ";", true
}

// FindInvokeEdges resolves, for every ResolvedInvoke with a known
// OwnerFQCN/MethodName, which class actually declares the real
// dispatch target — the invocation target's own concrete allocation
// type when REFLECTION R2 proved one (real dispatch, not an
// unconditional pin to OwnerFQCN, which might only be a
// supertype/interface the lookup was performed on), falling back to
// OwnerFQCN itself otherwise, the only basis left.
func FindInvokeEdges(h *hierarchy.Hierarchy, resolved []ResolvedInvoke) []InvokeEdgeResult {
	var out []InvokeEdgeResult
	for _, ri := range resolved {
		if ri.OwnerFQCN == "" {
			continue
		}
		r := InvokeEdgeResult{ResolvedInvoke: ri}

		if ri.ArgCountKnown && ri.ArgCount != len(ri.ParamTypes) {
			r.Gap = fmt.Sprintf("proven argument-count mismatch: the literal args array has %d element(s), the resolved Method takes %d",
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

		dispatchRoot := ri.OwnerFQCN
		if ri.ObjType != "" {
			dispatchRoot = ri.ObjType
		}
		declaring, found := h.ResolveByParams(dispatchRoot, ri.MethodName, paramDescs)
		if !found {
			r.Gap = fmt.Sprintf("%s does not resolve a method named %q with %d parameter(s) anywhere in its hierarchy",
				dispatchRoot, ri.MethodName, len(paramDescs))
		} else {
			r.DeclaringFQCN = declaring
		}
		out = append(out, r)
	}
	return out
}

// InvokeEdge is one (invoke call site, target method name) pair
// REFLECTION R2 wants as a call-graph edge — plain identity, same
// decoupling internal/methodhandle.HandleEdge already uses for
// internal/callgraph. No descriptor on the target: the edge connects to
// every overload of TargetMethod on TargetFQCN (internal/callgraph's
// own IDsByName), the same overload-insensitive final edge every other
// synthetic-edge producer in this codebase uses (JVM_GRAPH G2's own
// convention for vulnerable-method matching) — ResolveByParams already
// did the precise part (choosing *which* declaring class), this is just
// the edge shape.
type InvokeEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetMethod           string
}

// EdgeTargets returns every InvokeEdgeResult with a resolved
// DeclaringFQCN as a concrete edge to add — a Gap contributes nothing.
func EdgeTargets(results []InvokeEdgeResult) []InvokeEdge {
	var out []InvokeEdge
	for _, r := range results {
		if r.Gap != "" {
			continue
		}
		s := r.ResolvedInvoke.Site
		out = append(out, InvokeEdge{
			CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, PC: s.PC,
			TargetFQCN: r.DeclaringFQCN, TargetMethod: r.ResolvedInvoke.MethodName,
		})
	}
	return out
}
