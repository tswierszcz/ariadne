package dataflow

import (
	"fmt"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ResolveCallResult is DATA_FLOW D3's own third interprocedural
// summary: given an invokestatic/invokespecial call's target exactly as
// the bytecode named it (valueflow.CallResultRef — the owner is the
// literal Methodref owner, not necessarily the declaring class, same
// convention ResolveStaticField follows for fields), resolve the real
// declaring class (JVM_GRAPH G2's own exact method resolution, shared
// via hierarchy.ResolveExact) and merge every value its own
// return-with-value instructions produce — the same budgeted
// widening/merge rule the field summaries already use at a control-flow
// join, one level up: a value crossing from one method into another
// through a direct call's return, not a field.
//
// Unlike a field write (which can come from several writer methods),
// one call site names exactly one target method, so there is only ever
// one method to analyze here — but that method's own return value can
// itself be a KindField (a getter returning a cached field) or another
// KindCallResult (one helper calling another), so this chases through
// chaseFieldOrCall the same way ResolveField's own public entry point
// does, sharing one visited set across both kinds so a chain crossing
// between fields and calls (a field initialized from a call whose own
// return reads a field, and so on) is still cycle-safe — DATA_FLOW D0's
// own "Rekurencja zbiega lub ujawnia limit": a revisited field or call
// is reported Unknown, never looped forever.
func ResolveCallResult(h *hierarchy.Hierarchy, ref valueflow.CallResultRef) (valueflow.Value, []string) {
	visited := map[string]bool{callKey(ref): true}
	raw, warnings := resolveCallResultOnce(h, ref)
	v, chaseWarnings := chaseFieldOrCall(h, raw, visited)
	return v, append(warnings, chaseWarnings...)
}

// resolveCallResultOnce is ResolveCallResult's own one-hop resolution,
// with no further chasing — the piece chaseFieldOrCall calls back into
// when a field chase lands on a call result.
func resolveCallResultOnce(h *hierarchy.Hierarchy, ref valueflow.CallResultRef) (valueflow.Value, []string) {
	var warnings []string
	declaring, ok := h.ResolveExact(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf(
			"%s does not resolve method %s%s anywhere in its hierarchy", ref.Owner, ref.Name, ref.Desc)), warnings
	}
	ci, ok := h.Lookup(declaring)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf("declaring class %s not found in the analyzed hierarchy", declaring)), warnings
	}
	m, ok := ci.MethodByNameAndDesc(ref.Name, ref.Desc)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf(
			"%s does not declare a method named %s with descriptor %s directly", declaring, ref.Name, ref.Desc)), warnings
	}

	var jarReader classfile.JarClassReader
	defer jarReader.Close()
	cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf(
			"re-reading %s (%s) for a call-result summary of %s#%s%s: %v", ci.SourceJar, ci.SourceEntry, declaring, ref.Name, ref.Desc, err))
		return valueflow.Unknown("could not re-read the declaring class to find its own return value"), warnings
	}

	v, mergeWarnings := mergeReturns(cf, declaring, m)
	return v, append(warnings, mergeWarnings...)
}

// callKey is chaseFieldOrCall's own visited-set key for a
// CallResultRef — distinct from a FieldRef's key (fieldKey) by its
// "call:" prefix, since the two share one visited set when a chase
// crosses between them.
func callKey(ref valueflow.CallResultRef) string {
	return "call:" + ref.Owner + "#" + ref.Name + ref.Desc
}

// chaseFieldOrCall follows a value that is itself an unresolved
// KindField or KindCallResult one more hop (resolveFieldOnce/
// resolveCallResultOnce), repeating as long as each hop's own result is
// again one of those two kinds, until it lands on a terminal Value
// (Const/Typed/Alloc/Handle/Functional/Null/Set/Unknown) or visited
// catches a cycle. Shared by ResolveField and ResolveCallResult's own
// public entry points so a chain crossing between the two kinds stays
// cycle-safe under one visited set, not two independent ones that
// could ping-pong between them forever.
func chaseFieldOrCall(h *hierarchy.Hierarchy, v valueflow.Value, visited map[string]bool) (valueflow.Value, []string) {
	switch v.Kind {
	case valueflow.KindField:
		key := fieldKey(*v.Field)
		if visited[key] {
			return valueflow.Unknown(fmt.Sprintf("field/call-result chain revisits %s — a cycle, not resolved", key)), nil
		}
		visited[key] = true
		raw, warnings := resolveFieldOnce(h, *v.Field)
		next, chaseWarnings := chaseFieldOrCall(h, raw, visited)
		return next, append(warnings, chaseWarnings...)
	case valueflow.KindCallResult:
		key := callKey(*v.Call)
		if visited[key] {
			return valueflow.Unknown(fmt.Sprintf("field/call-result chain revisits %s — a cycle, not resolved", key)), nil
		}
		visited[key] = true
		raw, warnings := resolveCallResultOnce(h, *v.Call)
		next, chaseWarnings := chaseFieldOrCall(h, raw, visited)
		return next, append(warnings, chaseWarnings...)
	default:
		return v, nil
	}
}

// mergeReturns is ResolveCallResult's own I/O-free core, split out the
// same way mergeFieldWriters is: given an already-obtained
// *classfile.ClassFile and m's own metadata, analyze m once and merge
// every value-returning return site's value it recorded, in the
// deterministic order Analyze's own preregistration pass found them
// (MethodFacts.returnOrder) — never Go's randomized map iteration,
// which this codebase's own byte-identical-results.json determinism
// contract cannot tolerate.
func mergeReturns(cf *classfile.ClassFile, declaring string, m hierarchy.Method) (valueflow.Value, []string) {
	var warnings []string
	facts, err := Analyze(cf, declaring, m.Method)
	if err != nil {
		return valueflow.Unknown(fmt.Sprintf("local-value analysis: %v", err)), warnings
	}
	if aborted, reason := facts.Aborted(); aborted {
		warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis aborted: %s", declaring, m.Name, m.Descriptor, reason))
	}
	if len(facts.returnOrder) == 0 {
		return valueflow.Unknown(fmt.Sprintf(
			"%s#%s%s has no value-returning return in its own body (void, abstract, or native)", declaring, m.Name, m.Descriptor)), warnings
	}

	var result valueflow.Value
	haveAny := false
	for _, site := range facts.returnOrder {
		v := facts.returns[site]
		if !haveAny {
			result, haveAny = v, true
		} else {
			result = valueflow.Merge(result, v)
		}
	}
	return result, warnings
}
