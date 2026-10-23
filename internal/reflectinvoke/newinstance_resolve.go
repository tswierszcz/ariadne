package reflectinvoke

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ResolvedNewInstance is one NewInstanceSite with its Constructor
// receiver resolved, where possible, to the getConstructor/
// getDeclaredConstructor call that produced it (REFLECTION R2's own
// valueflow.KindReflectiveConstructor) — ResolvedInvoke's own
// counterpart for allocation rather than invocation. Unlike
// ResolvedInvoke, there is no separate "invocation target" argument to
// read: Constructor#newInstance allocates and constructs the new
// instance itself, with no receiver object of its own to dispatch
// against — the resolved OwnerFQCN is always the exact class to
// instantiate, never just an upper bound a real dispatch might refine.
type ResolvedNewInstance struct {
	Site NewInstanceSite
	// OwnerFQCN/ParamTypes are ReflectiveConstructorValue's own fields.
	OwnerFQCN  string
	ParamTypes []string
	// ArgCount/ArgCountKnown report the literal `Object[]` argument
	// array's own length when REFLECTION R1's same KindArray literal-
	// construction tracking proves it (or 0/true for a null args
	// array) — ResolvedInvoke's own ArgCount/ArgCountKnown, same
	// reasoning.
	ArgCount      int
	ArgCountKnown bool
	// UnresolvedReason explains why OwnerFQCN is "" — the Constructor
	// receiver wasn't a provable valueflow.KindReflectiveConstructor.
	UnresolvedReason string
}

// ResolveNewInstances is REFLECTION R2's own second-half receiver-
// resolution step — ResolveInvokes's own counterpart, reading back a
// Constructor receiver instead of a Method one.
func ResolveNewInstances(h *hierarchy.Hierarchy, sites []NewInstanceSite) ([]ResolvedNewInstance, []string) {
	var out []ResolvedNewInstance
	var warnings []string
	var jarReader classfile.JarClassReader
	defer jarReader.Close()

	byClass := map[string][]NewInstanceSite{}
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
				"re-reading %s (%s) for Constructor#newInstance call-site analysis: %v", ci.SourceJar, ci.SourceEntry, err))
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
			ri, resolveWarnings := resolveOneNewInstance(h, fqcn, s, facts)
			out = append(out, ri)
			warnings = append(warnings, resolveWarnings...)
		}
	}
	return out, warnings
}

// resolveOneNewInstance reads back one NewInstanceSite's Constructor
// receiver, chasing a KindField/KindCallResult indirection one more
// step — resolveOne's own counterpart — plus the literal args array
// (arg 0; there is no separate target-object argument the way
// Method#invoke has one at arg 0 and its args array at arg 1).
func resolveOneNewInstance(h *hierarchy.Hierarchy, fqcn string, s NewInstanceSite, facts *dataflow.MethodFacts) (ResolvedNewInstance, []string) {
	ri := ResolvedNewInstance{Site: s}
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
	case valueflow.KindReflectiveConstructor:
		ri.OwnerFQCN = val.ReflectiveConstructor.OwnerFQCN
		ri.ParamTypes = val.ReflectiveConstructor.ParamTypes
	case valueflow.KindUnknown:
		ri.UnresolvedReason = "the receiver is unknown: " + val.Reason
		return ri, warnings
	default:
		ri.UnresolvedReason = fmt.Sprintf("the receiver is %s, not a resolved getConstructor/getDeclaredConstructor result", val.Kind)
		return ri, warnings
	}

	args := facts.ArgumentAt(callSite, 0)
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
