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
