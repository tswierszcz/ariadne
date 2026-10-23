// Package annotation is FRAMEWORKS F1's own cross-class annotation
// resolution: internal/classfile decodes one class file's own
// annotations in isolation (no hierarchy access, by design — see its
// own doc comments), but a default value, a meta-annotation, or a
// repeatable annotation's container are all facts about a *different*
// class than the one being analyzed (the annotation type itself, not
// its usage site) — the same "re-read the class file through
// hierarchy access" shape internal/dataflow already uses for static
// fields, instance fields and call results, one layer up.
package annotation

import (
	"fmt"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// repeatableFQCN is java.lang.annotation.Repeatable's own FQCN — never
// read from a dependency jar (it's a JDK annotation type, outside the
// analyzed universe like every other java.lang.* type), so its own
// shape (a single "value" element naming the container class) is
// hard-coded here rather than resolved through ClassAnnotations.
const repeatableFQCN = "java.lang.annotation.Repeatable"

// ClassAnnotations re-reads fqcn's own class file and returns its
// class-level annotations — the shared building block every function
// in this package uses to look at a *different* class than the one
// whose usage site is being analyzed (an annotation type's own
// @Target/@Retention/@Repeatable, or a composed annotation's own
// meta-annotations). A class outside the analyzed hierarchy (most
// commonly a JDK annotation type ariadne never reads, e.g.
// java.lang.annotation.Target itself) is reported as a plain "not
// found" rather than an error — callers that only care about a
// specific, well-known meta-annotation (RepeatableContainer's own
// Repeatable lookup) special-case what they need instead.
func ClassAnnotations(h *hierarchy.Hierarchy, fqcn string) ([]classfile.Annotation, []string) {
	ci, ok := h.Lookup(fqcn)
	if !ok {
		return nil, nil
	}
	var jarReader classfile.JarClassReader
	defer jarReader.Close()
	cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		return nil, []string{fmt.Sprintf("re-reading %s (%s) for %s's own annotations: %v", ci.SourceJar, ci.SourceEntry, fqcn, err)}
	}
	return cf.AnnotationDetails, nil
}

// ResolveDefault finds annotationType's own declared default for
// elementName (its AnnotationDefault attribute, JVM spec 4.7.22) —
// what a usage site that omits this element actually gets. false means
// no default is declared (the element is mandatory at every usage
// site) or annotationType/elementName could not be resolved at all;
// either way, a caller must not guess a value, exactly D0's own
// "widen-or-report-unknown" discipline one layer up.
func ResolveDefault(h *hierarchy.Hierarchy, annotationType, elementName string) (classfile.ElementValue, bool, []string) {
	ci, ok := h.Lookup(annotationType)
	if !ok {
		return classfile.ElementValue{}, false, []string{fmt.Sprintf("%s not found in the analyzed hierarchy", annotationType)}
	}
	m, ok := ci.MethodByName(elementName)
	if !ok || m.DefaultValue == nil {
		return classfile.ElementValue{}, false, nil
	}
	return *m.DefaultValue, true, nil
}

// EffectiveValue is ResolveDefault's own usual entry point: ann's
// explicit value for elementName if the usage site gave one, otherwise
// ann.Type's own declared default. false means neither exists — the
// element is simply absent, not a value ariadne failed to read.
func EffectiveValue(h *hierarchy.Hierarchy, ann classfile.Annotation, elementName string) (classfile.ElementValue, bool, []string) {
	if v, ok := ann.Value(elementName); ok {
		return v, true, nil
	}
	return ResolveDefault(h, ann.Type, elementName)
}

// RepeatableContainer reports the container annotation type
// annotationType declares itself repeatable through (its own
// @java.lang.annotation.Repeatable(Container.class) meta-annotation),
// if any. false means annotationType is not repeatable, or its
// @Repeatable's own "value" element isn't the expected class literal —
// never guessed from shape alone (see ExpandRepeatable's own doc
// comment for why that matters).
func RepeatableContainer(h *hierarchy.Hierarchy, annotationType string) (string, bool, []string) {
	metas, warnings := ClassAnnotations(h, annotationType)
	for _, meta := range metas {
		if meta.Type != repeatableFQCN {
			continue
		}
		v, ok := meta.Value("value")
		if !ok || v.Kind != classfile.ElementClass {
			return "", false, warnings
		}
		return v.ClassFQCN, true, warnings
	}
	return "", false, warnings
}

// ExpandRepeatable replaces every container-shaped entry in anns with
// the repeated annotations it actually holds — Java's own compiler
// trick for "more than one of the same annotation": `@PropertySource`
// applied twice compiles to one `@PropertySources({@PropertySource,
// @PropertySource})`, invisible as *two* `@PropertySource` facts unless
// this unwrapping happens. An entry only expands when its repeated
// element's type both (a) looks like a container (a single "value"
// array of same-typed nested annotations) *and* (b) that nested type's
// own @Repeatable actually names this entry's type as its container —
// condition (b) is what makes this safe: an ordinary annotation with an
// unrelated nested-annotation-array element (e.g. Spring's
// @ComponentScan#excludeFilters) matches (a) but never (b), so it is
// left alone, never misidentified as a repeatable container by shape
// alone.
func ExpandRepeatable(h *hierarchy.Hierarchy, anns []classfile.Annotation) ([]classfile.Annotation, []string) {
	var out []classfile.Annotation
	var warnings []string
	for _, a := range anns {
		expanded, ok, w := expandOne(h, a)
		warnings = append(warnings, w...)
		if !ok {
			out = append(out, a)
			continue
		}
		out = append(out, expanded...)
	}
	return out, warnings
}

func expandOne(h *hierarchy.Hierarchy, a classfile.Annotation) ([]classfile.Annotation, bool, []string) {
	valueVal, ok := a.Value("value")
	if !ok || valueVal.Kind != classfile.ElementArray || len(valueVal.Array) == 0 {
		return nil, false, nil
	}

	elemType := ""
	for _, ev := range valueVal.Array {
		if ev.Kind != classfile.ElementAnnotation || ev.Nested == nil {
			return nil, false, nil
		}
		if elemType == "" {
			elemType = ev.Nested.Type
		} else if elemType != ev.Nested.Type {
			return nil, false, nil
		}
	}
	if elemType == "" {
		return nil, false, nil
	}

	container, ok, warnings := RepeatableContainer(h, elemType)
	if !ok || container != a.Type {
		return nil, false, warnings
	}

	out := make([]classfile.Annotation, len(valueVal.Array))
	for i, ev := range valueVal.Array {
		out[i] = *ev.Nested
	}
	return out, true, warnings
}
