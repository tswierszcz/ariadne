package dataflow

import (
	"fmt"
	"strings"

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

// ResolveCallResultOneHop resolves only the exact invokestatic/invokespecial
// target named by ref and merges its immediate return values. If a return is
// itself a KindCallResult or KindField, it is left intact for consumers that
// need to preserve provenance boundaries or apply their own depth limit.
func ResolveCallResultOneHop(h *hierarchy.Hierarchy, ref valueflow.CallResultRef) (valueflow.Value, []string) {
	return resolveCallResultOnce(h, ref)
}

// ResolveCallResultOneHopWithParameterRefs preserves a returned formal
// parameter as a typed value carrying its descriptor index. A caller can
// then substitute that marker using the arguments at ref.Site.
func ResolveCallResultOneHopWithParameterRefs(h *hierarchy.Hierarchy, ref valueflow.CallResultRef) (valueflow.Value, []string) {
	return resolveCallResultOnceMode(h, ref, true)
}

// ResolveCallArgumentAtCall analyzes the method containing ref.Site and
// returns the exact argument at index. It does not resolve virtual dispatch.
func ResolveCallArgumentAtCall(h *hierarchy.Hierarchy, ref valueflow.CallResultRef, index int) (valueflow.Value, []string) {
	var warnings []string
	separator := strings.LastIndexByte(ref.Site.Method, '#')
	if separator <= 0 || ref.Site.Descriptor == "" {
		return valueflow.Unknown("helper call site has no usable caller method identity"), warnings
	}
	callerOwner, callerName := ref.Site.Method[:separator], ref.Site.Method[separator+1:]
	ci, ok := h.Lookup(callerOwner)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf("caller class %s is not in the analyzed hierarchy", callerOwner)), warnings
	}
	callerMethod, ok := ci.MethodByNameAndDesc(callerName, ref.Site.Descriptor)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf("caller method %s#%s%s is not in the analyzed hierarchy", callerOwner, callerName, ref.Site.Descriptor)), warnings
	}
	var jarReader classfile.JarClassReader
	defer jarReader.Close()
	cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("re-reading %s (%s) for helper call arguments: %v", ci.SourceJar, ci.SourceEntry, err))
		return valueflow.Unknown("could not re-read the caller class to substitute a helper parameter"), warnings
	}
	facts, err := Analyze(cf, callerOwner, callerMethod.Method)
	if err != nil {
		return valueflow.Unknown(fmt.Sprintf("caller local-value analysis failed: %v", err)), warnings
	}
	if aborted, reason := facts.Aborted(); aborted {
		warnings = append(warnings, fmt.Sprintf("%s#%s%s: caller local-value analysis aborted: %s", callerOwner, callerName, ref.Site.Descriptor, reason))
	}
	args, err := classfile.ArgumentSlots(ref.Desc, true)
	if err != nil {
		return valueflow.Unknown(fmt.Sprintf("invalid helper descriptor %q: %v", ref.Desc, err)), warnings
	}
	if index < 0 || index >= len(args) {
		return valueflow.Unknown(fmt.Sprintf("helper parameter index %d is outside descriptor %s", index, ref.Desc)), warnings
	}
	return facts.ArgumentAt(ref.Site, index), warnings
}

// resolveCallResultOnce is ResolveCallResult's own one-hop resolution,
// with no further chasing — the piece chaseFieldOrCall calls back into
// when a field chase lands on a call result.
func resolveCallResultOnce(h *hierarchy.Hierarchy, ref valueflow.CallResultRef) (valueflow.Value, []string) {
	return resolveCallResultOnceMode(h, ref, false)
}

func resolveCallResultOnceMode(h *hierarchy.Hierarchy, ref valueflow.CallResultRef, parameterRefs bool) (valueflow.Value, []string) {
	if isDoPrivileged(ref) {
		return resolvePrivileged(h, ref, parameterRefs)
	}
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

	v, mergeWarnings := mergeReturnsMode(cf, declaring, m, parameterRefs)
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
	return mergeReturnsMode(cf, declaring, m, false)
}

func mergeReturnsMode(cf *classfile.ClassFile, declaring string, m hierarchy.Method, parameterRefs bool) (valueflow.Value, []string) {
	var warnings []string
	var facts *MethodFacts
	var err error
	if parameterRefs {
		facts, err = AnalyzeWithParameterRefs(cf, declaring, m.Method)
	} else {
		facts, err = Analyze(cf, declaring, m.Method)
	}
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

// isDoPrivileged names every java.security.AccessController.doPrivileged
// overload: each runs its action's run() on the calling thread and
// returns that result (the ASYNC_CALLBACKS catalog models the edge).
func isDoPrivileged(ref valueflow.CallResultRef) bool {
	return ref.Owner == "java.security.AccessController" && ref.Name == "doPrivileged" &&
		(strings.HasPrefix(ref.Desc, "(Ljava/security/PrivilegedAction;") || strings.HasPrefix(ref.Desc, "(Ljava/security/PrivilegedExceptionAction;")) &&
		strings.HasSuffix(ref.Desc, ")Ljava/lang/Object;")
}

const accBridge = 0x0040

// resolvePrivileged is DATA_FLOW D3's library summary for doPrivileged:
// its result is whatever the action's run() returns. The action must be
// provable at the call site — an allocation of a known class (its own
// non-bridge, no-argument run()), or a lambda/method reference whose
// implementation takes no arguments (nothing captured). Anything else is
// an explicit unknown.
func resolvePrivileged(h *hierarchy.Hierarchy, ref valueflow.CallResultRef, parameterRefs bool) (valueflow.Value, []string) {
	action, warnings := ResolveCallArgumentAtCall(h, ref, 0)
	if action.Kind == valueflow.KindField || action.Kind == valueflow.KindCallResult {
		visited := map[string]bool{}
		var more []string
		action, more = chaseFieldOrCall(h, action, visited)
		warnings = append(warnings, more...)
	}
	switch action.Kind {
	case valueflow.KindAlloc:
		ci, ok := h.Lookup(action.Type)
		if !ok {
			return valueflow.Unknown(fmt.Sprintf("doPrivileged action class %s is outside the analyzed hierarchy", action.Type)), warnings
		}
		var run *hierarchy.Method
		for i := range ci.Methods {
			m := &ci.Methods[i]
			if m.Name == "run" && strings.HasPrefix(m.Descriptor, "()") && !m.IsStatic() && m.AccessFlags&accBridge == 0 {
				if run != nil {
					return valueflow.Unknown(fmt.Sprintf("doPrivileged action %s declares several run() methods", action.Type)), warnings
				}
				run = m
			}
		}
		if run == nil {
			return valueflow.Unknown(fmt.Sprintf("doPrivileged action %s does not declare its own run()", action.Type)), warnings
		}
		return mergeReturnsOf(h, ci, *run, parameterRefs, warnings)
	case valueflow.KindFunctional:
		f := action.Func
		if f == nil || f.ImplOwner == "" {
			return valueflow.Unknown("doPrivileged action is a lambda whose implementation was not resolved"), warnings
		}
		if !strings.HasPrefix(f.ImplDesc, "()") {
			return valueflow.Unknown(fmt.Sprintf("doPrivileged action lambda %s#%s%s takes captured or bound arguments, which are not tracked", f.ImplOwner, f.ImplName, f.ImplDesc)), warnings
		}
		ci, ok := h.Lookup(f.ImplOwner)
		if !ok {
			return valueflow.Unknown(fmt.Sprintf("lambda implementation class %s is outside the analyzed hierarchy", f.ImplOwner)), warnings
		}
		m, ok := ci.MethodByNameAndDesc(f.ImplName, f.ImplDesc)
		if !ok || !m.IsStatic() {
			return valueflow.Unknown(fmt.Sprintf("lambda implementation %s#%s%s is not a static method in the analyzed input", f.ImplOwner, f.ImplName, f.ImplDesc)), warnings
		}
		return mergeReturnsOf(h, ci, m, parameterRefs, warnings)
	case valueflow.KindUnknown:
		return valueflow.Unknown("doPrivileged action is unknown: " + action.Reason), warnings
	default:
		return valueflow.Unknown(fmt.Sprintf("doPrivileged action is %s, not a provable allocation or lambda", action.Kind)), warnings
	}
}

func mergeReturnsOf(h *hierarchy.Hierarchy, ci *hierarchy.ClassInfo, m hierarchy.Method, parameterRefs bool, warnings []string) (valueflow.Value, []string) {
	var jarReader classfile.JarClassReader
	defer jarReader.Close()
	cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		return valueflow.Unknown(fmt.Sprintf("could not re-read %s: %v", ci.FQCN, err)), warnings
	}
	v, more := mergeReturnsMode(cf, ci.FQCN, m, parameterRefs)
	return v, append(warnings, more...)
}
