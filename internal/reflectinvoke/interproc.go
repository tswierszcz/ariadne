package reflectinvoke

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ThroughParametersVersion versions the interprocedural step (REFLECTION
// R3, slice 3). Snapshots record it.
const ThroughParametersVersion = "r3-3"

// ResolveThroughParameters is REFLECTION R3's measured shape: a
// Method#invoke whose Method receiver is a parameter of the calling
// method, and whose callers in the input pass a Method that R1/R2 resolve
// — directly, or read from a field or returned by a static helper one step
// away (DATA_FLOW D3). The census (Provenance) found exactly this shape on
// petclinic and WebGoat: Thymeleaf's BinaryOperationExpression#
// doComposeBinaryOperationExpression receives isLeftAllowed/isRightAllowed
// Methods its six subclasses keep in static fields.
//
// For such a site every resolved caller value becomes one ResolvedInvoke
// (ThroughParameter set), with the invocation target and argument array
// read at the site as R2 does. When every caller value resolves — the
// callers are all known (dataflow.CallerIndex: statically exact call
// sites) and each passes a resolved Method — the unresolved entry is
// replaced, so the site is no model gap any more; otherwise the resolved
// entries are added and the unresolved one stays, with a reason saying how
// many values resolved. Constructor#newInstance has no such site in the
// measured applications and is left as it is.
func ResolveThroughParameters(h *hierarchy.Hierarchy, idx *dataflow.CallerIndex, resolved []ResolvedInvoke) ([]ResolvedInvoke, ParameterStats, []string) {
	var stats ParameterStats
	var warnings []string
	var reader classfile.JarClassReader
	defer reader.Close()
	facts := map[string]*dataflow.MethodFacts{}
	factsFor := func(s InvokeSite) *dataflow.MethodFacts {
		key := s.CallerFQCN + "#" + s.CallerName + s.CallerDesc
		if f, ok := facts[key]; ok {
			return f
		}
		facts[key] = nil
		ci, ok := h.Lookup(s.CallerFQCN)
		if !ok {
			return nil
		}
		m, found := ci.MethodByNameAndDesc(s.CallerName, s.CallerDesc)
		if !found {
			return nil
		}
		cf, err := reader.ReadClass(ci.SourceJar, ci.SourceEntry)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("re-reading %s for R3 parameter resolution: %v", s.CallerFQCN, err))
			return nil
		}
		f, err := dataflow.AnalyzeWithParameterRefs(cf, s.CallerFQCN, m.Method)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: local-value analysis: %v", key, err))
			return nil
		}
		if aborted, _ := f.Aborted(); aborted {
			return nil
		}
		facts[key] = f
		return f
	}

	var out []ResolvedInvoke
	for _, ri := range resolved {
		if ri.UnresolvedReason == "" {
			out = append(out, ri)
			continue
		}
		f := factsFor(ri.Site)
		if f == nil {
			out = append(out, ri)
			continue
		}
		s := ri.Site
		site := bytecodeir.Site{Method: s.CallerFQCN + "#" + s.CallerName, Descriptor: s.CallerDesc, PC: s.PC, Op: bytecodeir.OpCall}
		recv := f.ReceiverAt(site)
		if recv.Kind != valueflow.KindTyped || recv.Parameter == nil {
			out = append(out, ri)
			continue
		}
		stats.ParameterSites++
		pv := idx.ResolveParameter(dataflow.MethodID{Owner: s.CallerFQCN, Name: s.CallerName, Desc: s.CallerDesc}, recv.Parameter.Index)
		var targets []valueflow.ReflectiveMethodValue
		seen := map[string]bool{}
		unresolvedValues := 0
		for _, v := range pv.Values {
			r := oneHop(h, v)
			if r.Kind != valueflow.KindReflectiveMethod {
				unresolvedValues++
				continue
			}
			key := r.ReflectiveMethod.OwnerFQCN + "#" + r.ReflectiveMethod.Name + "(" + strings.Join(r.ReflectiveMethod.ParamTypes, ",") + ")"
			if !seen[key] {
				seen[key] = true
				targets = append(targets, *r.ReflectiveMethod)
			}
		}
		if len(targets) == 0 {
			out = append(out, ri)
			continue
		}
		sort.Slice(targets, func(i, j int) bool {
			if targets[i].OwnerFQCN != targets[j].OwnerFQCN {
				return targets[i].OwnerFQCN < targets[j].OwnerFQCN
			}
			return targets[i].Name < targets[j].Name
		})
		stats.ResolvedSites++
		base := ResolvedInvoke{Site: s, ThroughParameter: true}
		switch obj := f.ArgumentAt(site, 0); obj.Kind {
		case valueflow.KindAlloc:
			base.ObjType = obj.Type
		case valueflow.KindNull:
			base.ObjIsNull = true
		}
		switch args := f.ArgumentAt(site, 1); args.Kind {
		case valueflow.KindArray:
			base.ArgCount, base.ArgCountKnown = len(args.Array.Elements), true
		case valueflow.KindNull:
			base.ArgCount, base.ArgCountKnown = 0, true
		}
		for _, t := range targets {
			r := base
			r.OwnerFQCN, r.MethodName, r.ParamTypes, r.Declared = t.OwnerFQCN, t.Name, t.ParamTypes, t.Declared
			out = append(out, r)
			stats.Targets++
		}
		if unresolvedValues > 0 || len(pv.Unknown) > 0 {
			stats.PartialSites++
			ri.PartlyThroughParameter = true
			ri.UnresolvedReason = fmt.Sprintf("the receiver is parameter %d: %d caller value(s) resolved, %d not, %d caller contribution(s) undetermined",
				recv.Parameter.Index, len(targets), unresolvedValues, len(pv.Unknown))
			out = append(out, ri)
		}
	}
	return out, stats, warnings
}

// ParameterStats counts what ResolveThroughParameters did.
type ParameterStats struct {
	// ParameterSites: unresolved sites whose receiver is a parameter.
	ParameterSites int
	// ResolvedSites: of those, sites with at least one resolved caller
	// value; PartialSites: of those, sites that stay a gap too.
	ResolvedSites, PartialSites int
	// Targets: resolved Method identities added.
	Targets int
}
