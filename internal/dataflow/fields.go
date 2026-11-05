package dataflow

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ResolveStaticField is DATA_FLOW D3's own first interprocedural
// summary: given a static field exactly as the bytecode named it
// (valueflow.FieldRef — the owner is the literal Fieldref owner, not
// necessarily the declaring class), find every method *on the field's
// own declaring class* that writes it (classfile.InitPutStatic,
// already extracted once by hierarchy.Build — no new bytecode scan
// needed to find candidates) and merge every value they store into one
// summary Value, using the same widening/merge rules D2's own local
// analysis already applies at a control-flow join.
//
// Scoped to the declaring class's own methods only — the motivating
// real pattern (confirmed on petclinic/WebGoat while building
// DYNAMIC_LINKAGE L2: a private static final MethodHandle/Logger/etc.
// set once in <clinit> and read from a different method) is exactly
// this shape. A write reached only through a subclass reference, or
// from a class other than the declaring one, is a rarer pattern left
// for a later increment — "no writer found in its own declaring class"
// is reported honestly as Unknown, never silently assumed impossible.
func ResolveStaticField(h *hierarchy.Hierarchy, ref valueflow.FieldRef) (valueflow.Value, []string) {
	var warnings []string
	declaring, ok := h.ResolveField(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf(
			"%s does not resolve field %s:%s anywhere in its hierarchy", ref.Owner, ref.Name, ref.Desc)), warnings
	}
	ci, ok := h.Lookup(declaring)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf("declaring class %s not found in the analyzed hierarchy", declaring)), warnings
	}

	var jarReader classfile.JarClassReader
	defer jarReader.Close()
	cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf(
			"re-reading %s (%s) for a field summary of %s#%s: %v", ci.SourceJar, ci.SourceEntry, declaring, ref.Name, err))
		return valueflow.Unknown("could not re-read the declaring class to find its writers"), warnings
	}

	v, mergeWarnings := mergeFieldWriters(cf, declaring, ci.Methods, ref)
	return v, append(warnings, mergeWarnings...)
}

// mergeFieldWriters is ResolveStaticField's own I/O-free core: given an
// already-obtained *classfile.ClassFile for declaring and its own
// methods (hierarchy.ClassInfo.Methods, already carrying each one's
// InitRefs), find and merge every static write of ref. Split out from
// ResolveStaticField so it's testable with
// classfile.NewClassFileForTesting directly — same reasoning
// internal/callback's own resolveOne/ResolveCallbacks split follows.
func mergeFieldWriters(cf *classfile.ClassFile, declaring string, methods []hierarchy.Method, ref valueflow.FieldRef) (valueflow.Value, []string) {
	var warnings []string
	var result valueflow.Value
	haveAny := false
	for _, m := range methods {
		var writes []classfile.InitRef
		for _, ir := range m.InitRefs {
			if ir.Kind == classfile.InitPutStatic && ir.Owner == declaring && ir.Name == ref.Name && ir.Desc == ref.Desc {
				writes = append(writes, ir)
			}
		}
		if len(writes) == 0 {
			continue
		}

		facts, err := Analyze(cf, declaring, m.Method)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis: %v", declaring, m.Name, m.Descriptor, err))
			continue
		}
		if aborted, reason := facts.Aborted(); aborted {
			warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis aborted: %s", declaring, m.Name, m.Descriptor, reason))
		}
		for _, ir := range writes {
			site := bytecodeir.Site{Method: declaring + "#" + m.Name, Descriptor: m.Descriptor, PC: ir.PC, Op: bytecodeir.OpFieldPut}
			v := facts.FieldValueAt(site)
			if !haveAny {
				result, haveAny = v, true
			} else {
				result = valueflow.Merge(result, v)
			}
		}
	}
	if !haveAny {
		return valueflow.Unknown(fmt.Sprintf("no writer of %s#%s found in its own declaring class", declaring, ref.Name)), warnings
	}
	return result, warnings
}

// ResolveInstanceField is DATA_FLOW D3's second interprocedural summary.
// An instance field in general needs real points-to (which allocation
// "this" refers to at the read site) to resolve safely across methods —
// out of scope here, same as everywhere else in this codebase that
// stops short of points-to. This increment covers only the one shape
// that is safe without it: a *final* field written only by a
// `this.field = ...` store inside a constructor (<init>) of its own
// declaring class. Finality means no method outside a constructor can
// ever change it again, and a self-receiver write means the value does
// not depend on which instance is asking — so merging every constructor
// write (the same widening rule ResolveStaticField already uses) is as
// safe as the static case. A non-final field, or one with no such
// constructor write, is reported Unknown, never guessed.
func ResolveInstanceField(h *hierarchy.Hierarchy, ref valueflow.FieldRef) (valueflow.Value, []string) {
	var warnings []string
	declaring, ok := h.ResolveField(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf(
			"%s does not resolve field %s:%s anywhere in its hierarchy", ref.Owner, ref.Name, ref.Desc)), warnings
	}
	ci, ok := h.Lookup(declaring)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf("declaring class %s not found in the analyzed hierarchy", declaring)), warnings
	}
	field, ok := ci.FieldByNameAndDesc(ref.Name, ref.Desc)
	if !ok || field.AccessFlags&classfile.AccFinal == 0 {
		return valueflow.Unknown(fmt.Sprintf(
			"%s#%s is not a final field — instance-field summaries in this increment are scoped to a final field set in a constructor", declaring, ref.Name)), warnings
	}

	var jarReader classfile.JarClassReader
	defer jarReader.Close()
	cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf(
			"re-reading %s (%s) for a field summary of %s#%s: %v", ci.SourceJar, ci.SourceEntry, declaring, ref.Name, err))
		return valueflow.Unknown("could not re-read the declaring class to find its constructor writes"), warnings
	}

	v, mergeWarnings := mergeConstructorFieldWriters(cf, declaring, ci.Methods, ref.Name, ref.Desc)
	return v, append(warnings, mergeWarnings...)
}

// mergeConstructorFieldWriters is ResolveInstanceField's own I/O-free
// core, split out the same way mergeFieldWriters is. Unlike
// mergeFieldWriters, it cannot use classfile.InitRef as its structural
// list of candidate writer sites — getfield/putfield never trigger JVM
// class initialization (JVMS 5.5), so JVM_GRAPH G3 never extracted them
// — so it decodes each constructor's own local ops once (fieldPutPCs)
// to find candidate sites, then consults MethodFacts.fieldWrites
// directly (not the exported, panic-on-absence FieldValueAt: a site
// found here that step.go chose not to record — a write through a
// non-self receiver — is a legitimate "not this increment's shape",
// not a caller error).
func mergeConstructorFieldWriters(cf *classfile.ClassFile, declaring string, methods []hierarchy.Method, name, desc string) (valueflow.Value, []string) {
	var warnings []string
	var result valueflow.Value
	haveAny := false
	for _, m := range methods {
		if m.Name != "<init>" {
			continue
		}
		pcs, err := fieldPutPCs(cf, m.Method, declaring, name, desc)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s#%s%s: decoding field-put sites: %v", declaring, m.Name, m.Descriptor, err))
			continue
		}
		if len(pcs) == 0 {
			continue
		}

		facts, err := Analyze(cf, declaring, m.Method)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis: %v", declaring, m.Name, m.Descriptor, err))
			continue
		}
		if aborted, reason := facts.Aborted(); aborted {
			warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis aborted: %s", declaring, m.Name, m.Descriptor, reason))
		}
		for _, pc := range pcs {
			site := bytecodeir.Site{Method: declaring + "#" + m.Name, Descriptor: m.Descriptor, PC: pc, Op: bytecodeir.OpFieldPut}
			v, ok := facts.fieldWrites[site]
			if !ok {
				continue // not a `this.field = ...` write, or its block was never reached — never fabricate
			}
			if !haveAny {
				result, haveAny = v, true
			} else {
				result = valueflow.Merge(result, v)
			}
		}
	}
	if !haveAny {
		return valueflow.Unknown(fmt.Sprintf("no self (`this.field = ...`) write of %s#%s found in any constructor of its own declaring class", declaring, name)), warnings
	}
	return result, warnings
}

// ResolveField is DATA_FLOW D3's single entry point for a KindField
// receiver: it resolves ref to its declaring class and real access
// flags exactly once (resolveFieldOnce), then — since a field's own
// stored value can itself be an unresolved KindField (a field whose
// writer just reads another field) or KindCallResult (`private static
// final MethodHandle H = findIt();` — a field initialized from a
// direct call) — chases that one more hop via chaseFieldOrCall, the
// same cycle-guarded chase ResolveCallResult's own public entry point
// uses. A caller that instead tried ResolveStaticField/
// ResolveInstanceField directly (static first, falling through to
// instance on any Unknown) would, for a *static* field whose <clinit>-
// computed value just isn't provable, get that call's honest reason
// silently overwritten by ResolveInstanceField's "not written in a
// constructor" — true of every static field, and misleading about why
// resolution actually failed.
func ResolveField(h *hierarchy.Hierarchy, ref valueflow.FieldRef) (valueflow.Value, []string) {
	visited := map[string]bool{fieldKey(ref): true}
	raw, warnings := resolveFieldOnce(h, ref)
	v, chaseWarnings := chaseFieldOrCall(h, raw, visited)
	return v, append(warnings, chaseWarnings...)
}

// ResolveFieldOneHop resolves the exact field summary without chasing a
// stored KindField or KindCallResult. Callers that need provenance can then
// traverse those values themselves and preserve their own cycle/depth rules.
func ResolveFieldOneHop(h *hierarchy.Hierarchy, ref valueflow.FieldRef) (valueflow.Value, []string) {
	return resolveFieldOnce(h, ref)
}

// resolveFieldOnce is ResolveField's own one-hop dispatch, with no
// further chasing — the piece chaseFieldOrCall calls back into when a
// call-result chase lands on a field.
func resolveFieldOnce(h *hierarchy.Hierarchy, ref valueflow.FieldRef) (valueflow.Value, []string) {
	declaring, ok := h.ResolveField(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf(
			"%s does not resolve field %s:%s anywhere in its hierarchy", ref.Owner, ref.Name, ref.Desc)), nil
	}
	ci, ok := h.Lookup(declaring)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf("declaring class %s not found in the analyzed hierarchy", declaring)), nil
	}
	field, ok := ci.FieldByNameAndDesc(ref.Name, ref.Desc)
	if !ok {
		return valueflow.Unknown(fmt.Sprintf(
			"%s does not declare a field named %s with descriptor %s directly", declaring, ref.Name, ref.Desc)), nil
	}
	if field.AccessFlags&classfile.AccStatic != 0 {
		return ResolveStaticField(h, ref)
	}
	return ResolveInstanceField(h, ref)
}

// fieldKey is chaseFieldOrCall's own visited-set key for a FieldRef —
// distinct from a CallResultRef's key (callKey) by its "field:" prefix,
// since the two share one visited set when a chase crosses between
// them.
func fieldKey(ref valueflow.FieldRef) string {
	return "field:" + ref.Owner + "#" + ref.Name + ":" + ref.Desc
}

// fieldPutPCs decodes m's own local ops and returns the byte offset of
// every instance putfield instruction targeting declaring#name:desc —
// the structural candidate list mergeConstructorFieldWriters needs,
// decoded straight from the method's Code the same way Analyze itself
// does (DecodeLocalOps), since there is no pre-extracted InitRef-style
// list for instance field writes to reuse.
func fieldPutPCs(cf *classfile.ClassFile, m classfile.Method, declaring, name, desc string) ([]int, error) {
	if len(m.Code) == 0 {
		return nil, nil
	}
	ops, err := cf.DecodeLocalOps(m.Code)
	if err != nil {
		return nil, err
	}
	var pcs []int
	for _, op := range ops {
		if op.Kind == classfile.LocalOpFieldPut && !op.IsStatic &&
			op.FieldOwner == declaring && op.FieldName == name && op.FieldDesc == desc {
			pcs = append(pcs, op.PC)
		}
	}
	return pcs, nil
}
