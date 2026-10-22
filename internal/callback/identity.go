package callback

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ResolvedCallback is ASYNC_CALLBACKS A1's own result: one cataloged
// RegistrationSite (A0) with its callback argument resolved, where
// possible, to either a concrete class (a direct `new`) or a
// lambda/method reference's own implementation (DYNAMIC_LINKAGE L0/L1).
// Exactly one of TargetFQCN and ImplOwner is set on a resolved
// ResolvedCallback; both are "" when UnresolvedReason explains why
// neither applies.
type ResolvedCallback struct {
	Site RegistrationSite
	// TargetFQCN is the concrete class whose CallbackMethod will
	// actually run — set only when the argument was a provably exact
	// allocation (valueflow.KindAlloc) of a class directly passed to the
	// registration call.
	TargetFQCN string
	// ImplOwner/ImplName/ImplDesc are set instead of TargetFQCN when the
	// argument was a lambda or method reference (valueflow.KindFunctional,
	// DYNAMIC_LINKAGE L0/L1): the real method it dispatches straight to
	// when CallbackMethod is invoked — already the exact declaring
	// method (resolved once via LambdaMetafactory, not CHA), so a
	// consumer building a call-graph edge from this needs no further
	// hierarchy lookup, unlike TargetFQCN's own case.
	ImplOwner, ImplName, ImplDesc string
	// UnresolvedReason explains why neither TargetFQCN nor ImplOwner is
	// set — e.g. a parameter, field read, or an unresolved dynamic
	// bootstrap (StringConcatFactory and the like, never guessed at).
	UnresolvedReason string
}

// ResolveCallbacks is ASYNC_CALLBACKS A1: for every cataloged
// RegistrationSite (A0), resolve its callback argument via DATA_FLOW
// D0-D2's local-value analysis — a concrete class for a direct `new`
// allocation, or a lambda/method reference's own implementation method
// (DYNAMIC_LINKAGE L0/L1, valueflow.KindFunctional) — using the same
// re-read-the-class-file-for-its-pool approach SERVICE_LOADING S1 uses
// (classfile.ClassFile's constant pool is not retained past
// hierarchy.Build). A site whose Outcome isn't OutcomeCataloged (A0's
// own "unsupported-api" diagnostic) carries no model to resolve against
// and is skipped, not guessed at.
//
// The "Thread subclass overrides run() directly, with no Runnable
// object anywhere" pattern A1's own plan text calls out
// ("rozdzielić przekazane zadanie i override run") is deliberately not
// handled here: A0's catalog only matches Thread's Runnable-taking
// constructors, so that pattern produces no RegistrationSite at all to
// resolve — a separate detection, not a resolution step, and left as
// this increment's own open scope rather than silently claimed.
func ResolveCallbacks(h *hierarchy.Hierarchy, sites []RegistrationSite) ([]ResolvedCallback, []string) {
	var out []ResolvedCallback
	var warnings []string
	var jarReader classfile.JarClassReader
	defer jarReader.Close()

	byClass := map[string][]RegistrationSite{}
	for _, s := range sites {
		if s.Outcome != OutcomeCataloged {
			continue
		}
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
				"re-reading %s (%s) for callback identity analysis: %v", ci.SourceJar, ci.SourceEntry, err))
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
			out = append(out, resolveOne(fqcn, s, facts))
		}
	}
	return out, warnings
}

func resolveOne(fqcn string, s RegistrationSite, facts *dataflow.MethodFacts) ResolvedCallback {
	rc := ResolvedCallback{Site: s}
	argSite := bytecodeir.Site{Method: fqcn + "#" + s.CallerName, Descriptor: s.CallerDesc, PC: s.PC, Op: bytecodeir.OpCall}
	val := facts.ArgumentAt(argSite, s.Model.CallbackArgIndex)

	if aborted, reason := facts.Aborted(); aborted {
		rc.UnresolvedReason = "local-value analysis for this method aborted: " + reason
		return rc
	}
	switch val.Kind {
	case valueflow.KindAlloc:
		rc.TargetFQCN = val.Type
	case valueflow.KindFunctional:
		if val.Func.ImplOwner == "" {
			rc.UnresolvedReason = fmt.Sprintf(
				"the argument is a lambda/method reference whose bootstrap could not be resolved (%s#%s)", val.Func.SamInterface, val.Func.SamMethod)
			break
		}
		rc.ImplOwner, rc.ImplName, rc.ImplDesc = val.Func.ImplOwner, val.Func.ImplName, val.Func.ImplDesc
	case valueflow.KindUnknown:
		rc.UnresolvedReason = "the argument is unknown: " + val.Reason
	default:
		rc.UnresolvedReason = fmt.Sprintf(
			"the argument is %s, not a direct `new` allocation or a resolved lambda/method reference", val.Kind)
	}
	return rc
}
