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
	// Forwarded (ASYNC_CALLBACKS a3-2): the argument was a parameter of
	// the registering method, resolved at its callers — the caller
	// methods (owner#name+desc, outermost last) through which this
	// target reached the site. A site can then have several
	// ResolvedCallbacks, one per target its callers pass.
	Forwarded []string
	// Origin (a3-2/a3-3): for a forwarded target, the outermost call
	// site passing it — where the object enters the chain. A3-3's queue
	// hand-offs draw their edge from there, so each caller reaches only
	// what it hands in.
	Origin *Origin
	// SamInterface/SamMethod: for a forwarded lambda/method reference,
	// the functional interface method it implements (a3-3 matches it
	// against the method a drain calls).
	SamInterface, SamMethod string
}

// Origin is a call site: owner#name+desc at PC.
type Origin struct {
	FQCN, Name, Desc string
	PC               int
}

// Forwarding bounds (a3-2): how many caller levels a parameter is
// followed through, and how many caller sites in all, per registration
// site.
const (
	forwardDepth   = 3
	forwardCallers = 32
)

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
//
// a3-2: when the argument is the registering method's own parameter, it
// is resolved at that method's callers: each caller site that passes a
// direct `new` or a lambda/method reference contributes a target, and a
// caller passing its own parameter on is followed further (forwardDepth
// levels, forwardCallers caller sites per registration site). The site
// then has one ResolvedCallback per distinct target, each with Forwarded.
func ResolveCallbacks(h *hierarchy.Hierarchy, sites []RegistrationSite) ([]ResolvedCallback, []string) {
	var out []ResolvedCallback
	var warnings []string
	var jarReader classfile.JarClassReader
	defer jarReader.Close()
	fw := newForwarder(h, &jarReader)

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
				f, err := dataflow.AnalyzeWithParameterRefs(cf, fqcn, m.Method)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis: %v", fqcn, s.CallerName, s.CallerDesc, err))
					continue
				}
				factsByMethod[key] = f
				facts = f
			}
			rc := resolveOne(fqcn, s, facts)
			if p, ok := parameterArgument(fqcn, s, facts); ok && rc.TargetFQCN == "" && rc.ImplOwner == "" {
				fw.site = s
				if found := fw.resolve(fqcn+"#"+s.CallerName+s.CallerDesc, p, nil, 0); len(found) > 0 {
					out = append(out, found...)
					continue
				}
				rc.UnresolvedReason = fmt.Sprintf("the argument is parameter %d of %s#%s%s, and no caller (of at most %d, %d levels) passes a direct `new` or a lambda/method reference",
					p, fqcn, s.CallerName, s.CallerDesc, forwardCallers, forwardDepth)
			}
			out = append(out, rc)
		}
	}
	return out, warnings
}

// parameterArgument: the site's callback argument is the registering
// method's own parameter; its index (receiver excluded).
func parameterArgument(fqcn string, s RegistrationSite, facts *dataflow.MethodFacts) (int, bool) {
	if aborted, _ := facts.Aborted(); aborted {
		return 0, false
	}
	val, ok := argumentAt(facts, bytecodeir.Site{Method: fqcn + "#" + s.CallerName, Descriptor: s.CallerDesc, PC: s.PC, Op: bytecodeir.OpCall}, s.Model.CallbackArgIndex)
	if ok && val.Kind == valueflow.KindTyped && val.Parameter != nil && val.Parameter.Index >= 0 {
		return val.Parameter.Index, true
	}
	return 0, false
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
