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
// RegistrationSite (A0) with its callback argument's concrete class
// resolved, where possible.
type ResolvedCallback struct {
	Site RegistrationSite
	// TargetFQCN is the concrete class whose CallbackMethod will
	// actually run — set only when the argument was a provably exact
	// allocation (valueflow.KindAlloc) of a class directly passed to the
	// registration call. "" otherwise; see UnresolvedReason.
	TargetFQCN string
	// UnresolvedReason explains why TargetFQCN is "". A1's own declared
	// scope deliberately stops at "a `new X(...)` passed directly or
	// through a local variable" (DATA_FLOW D0-D2's own domain) — a
	// lambda or method reference is reported as unresolved here with a
	// reason naming that, not silently treated as "unknown receiver":
	// DYNAMIC_LINKAGE L1 (not yet implemented) is what would resolve
	// those, and this phase does not pretend to.
	UnresolvedReason string
}

// ResolveCallbacks is ASYNC_CALLBACKS A1: for every cataloged
// RegistrationSite (A0), resolve the concrete class passed as its
// callback argument via DATA_FLOW D0-D2's local-value analysis — same
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
	case valueflow.KindUnknown:
		rc.UnresolvedReason = "the argument is unknown: " + val.Reason
	default:
		rc.UnresolvedReason = fmt.Sprintf(
			"the argument is %s, not a direct `new` allocation (lambdas/method references need DYNAMIC_LINKAGE L1, not yet implemented)", val.Kind)
	}
	return rc
}
