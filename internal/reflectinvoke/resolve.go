package reflectinvoke

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ResolvedInvoke is one InvokeSite with its Method receiver resolved,
// where possible, to the getMethod/getDeclaredMethod call that produced
// it (REFLECTION R1's own KindReflectiveMethod), plus what this
// particular call site's own arguments prove about the invocation.
type ResolvedInvoke struct {
	Site InvokeSite
	// OwnerFQCN/MethodName/ParamTypes are ReflectiveMethodValue's own
	// fields — the Class the lookup was performed on, not necessarily
	// the class that actually declares the method (FindInvokeEdges
	// resolves that, same split DYNAMIC_LINKAGE L2 already uses for
	// RequestedOwner).
	OwnerFQCN, MethodName string
	ParamTypes            []string
	// ObjType is the invocation target's own concrete allocation type
	// (valueflow.KindAlloc's Type) when provable — real dispatch should
	// resolve against this, not blindly against OwnerFQCN, which may
	// only be a supertype/interface the lookup happened to be performed
	// on. "" when not provable (a parameter, a field read, or any other
	// value this analysis cannot pin to one allocation site) — real
	// dispatch then falls back to OwnerFQCN itself, the only basis left,
	// not a guess at a subtype.
	ObjType string
	// ObjIsNull is true when the invocation target is the null literal
	// — a static method invoke, which needs no dispatch at all.
	ObjIsNull bool
	// ArgCount/ArgCountKnown report the literal `Object[]` argument
	// array's own length when REFLECTION R1's same KindArray literal-
	// construction tracking proves it (or 0/true for a null args
	// array). false when the array isn't a provable literal at this
	// call site (an existing variable, a field, ...) — arity then
	// cannot be checked at all, so a mismatch is never claimed, only
	// possibly disproved.
	ArgCount      int
	ArgCountKnown bool
	// UnresolvedReason explains why OwnerFQCN/MethodName are "" — the
	// Method receiver wasn't a provable valueflow.KindReflectiveMethod.
	UnresolvedReason string
}

// ResolveInvokes is REFLECTION R2's own receiver-resolution step: for
// every InvokeSite, read back its Method receiver and its own two
// arguments via DATA_FLOW D2's Facts.ReceiverAt/ArgumentAt. Same
// re-read-the-class-file-for-its-pool approach
// internal/methodhandle.ResolveInvokes already uses.
func ResolveInvokes(h *hierarchy.Hierarchy, sites []InvokeSite) ([]ResolvedInvoke, []string) {
	var out []ResolvedInvoke
	var warnings []string
	var jarReader classfile.JarClassReader
	defer jarReader.Close()

	byClass := map[string][]InvokeSite{}
	for _, s := range sites {
		byClass[s.CallerFQCN] = append(byClass[s.CallerFQCN], s)
	}

	for _, fqcn := range h.SortedFQCNs() {
		classSites, ok := byClass[fqcn]
		if !ok {
			continue
		}
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf(
				"re-reading %s (%s) for Method#invoke call-site analysis: %v", ci.SourceJar, ci.SourceEntry, err))
			continue
		}

		factsByMethod := map[[2]string]*dataflow.MethodFacts{}
		for _, s := range classSites {
			key := [2]string{s.CallerName, s.CallerDesc}
			facts, ok := factsByMethod[key]
			if !ok {
				m, found := ci.MethodByNameAndDesc(s.CallerName, s.CallerDesc)
				if !found {
					warnings = append(warnings, fmt.Sprintf(
						"%s#%s%s: method not found in the re-read class file", fqcn, s.CallerName, s.CallerDesc))
					continue
				}
				f, err := dataflow.Analyze(cf, fqcn, m.Method)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis: %v", fqcn, s.CallerName, s.CallerDesc, err))
					continue
				}
				factsByMethod[key] = f
				facts = f
			}
			ri, resolveWarnings := resolveOne(h, fqcn, s, facts)
			out = append(out, ri)
			warnings = append(warnings, resolveWarnings...)
		}
	}
	return out, warnings
}

// resolveOne reads back one InvokeSite's Method receiver, chasing a
// KindField/KindCallResult indirection one more step — the "find a
// Method once, cache it in a field, invoke from elsewhere" and "look it
// up in a helper method, invoke from elsewhere" patterns
// internal/methodhandle's own resolveOne already chases for a
// MethodHandle receiver — plus the invocation target (arg 0) and
// argument array (arg 1), read directly: neither needs a cross-method
// chase, since real dispatch only cares about a provable allocation
// site or null, and a proven-mismatched arity is useful even when only
// provable within this one method.
func resolveOne(h *hierarchy.Hierarchy, fqcn string, s InvokeSite, facts *dataflow.MethodFacts) (ResolvedInvoke, []string) {
	ri := ResolvedInvoke{Site: s}
	callSite := bytecodeir.Site{Method: fqcn + "#" + s.CallerName, Descriptor: s.CallerDesc, PC: s.PC, Op: bytecodeir.OpCall}

	if aborted, reason := facts.Aborted(); aborted {
		ri.UnresolvedReason = "local-value analysis for this method aborted: " + reason
		return ri, nil
	}

	val := facts.ReceiverAt(callSite)
	var warnings []string
	switch val.Kind {
	case valueflow.KindField:
		resolved, fieldWarnings := dataflow.ResolveField(h, *val.Field)
		warnings = fieldWarnings
		val = resolved
	case valueflow.KindCallResult:
		resolved, callWarnings := dataflow.ResolveCallResult(h, *val.Call)
		warnings = callWarnings
		val = resolved
	}

	switch val.Kind {
	case valueflow.KindReflectiveMethod:
		ri.OwnerFQCN = val.ReflectiveMethod.OwnerFQCN
		ri.MethodName = val.ReflectiveMethod.Name
		ri.ParamTypes = val.ReflectiveMethod.ParamTypes
	case valueflow.KindUnknown:
		ri.UnresolvedReason = "the receiver is unknown: " + val.Reason
		return ri, warnings
	default:
		ri.UnresolvedReason = fmt.Sprintf("the receiver is %s, not a resolved getMethod/getDeclaredMethod result", val.Kind)
		return ri, warnings
	}

	obj := facts.ArgumentAt(callSite, 0)
	switch obj.Kind {
	case valueflow.KindAlloc:
		ri.ObjType = obj.Type
	case valueflow.KindNull:
		ri.ObjIsNull = true
	}

	args := facts.ArgumentAt(callSite, 1)
	switch args.Kind {
	case valueflow.KindArray:
		ri.ArgCount = len(args.Array.Elements)
		ri.ArgCountKnown = true
	case valueflow.KindNull:
		ri.ArgCount = 0
		ri.ArgCountKnown = true
	}

	return ri, warnings
}
