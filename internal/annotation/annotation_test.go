package annotation

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestResolveDefault_unresolvableTypeIsFalse(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	_, ok, _ := ResolveDefault(h, "test.NoSuchAnnotation", "value")
	if ok {
		t.Error("ResolveDefault(unresolvable type) = true, want false")
	}
}

func TestResolveDefault_noDefaultDeclaredIsFalse(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN: "test.Config",
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "value"}}, // mandatory element, no default
			},
		},
	})
	_, ok, _ := ResolveDefault(h, "test.Config", "value")
	if ok {
		t.Error("ResolveDefault(no default) = true, want false")
	}
}

func TestResolveDefault_declaredDefaultResolves(t *testing.T) {
	want := classfile.ElementValue{Kind: classfile.ElementConst, Const: "default"}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN: "test.Config",
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "value", DefaultValue: &want}},
			},
		},
	})
	got, ok, _ := ResolveDefault(h, "test.Config", "value")
	if !ok {
		t.Fatal("ResolveDefault(declared default) = false, want true")
	}
	if got.Kind != want.Kind || got.Const != want.Const {
		t.Errorf("ResolveDefault = %+v, want %+v", got, want)
	}
}

func TestEffectiveValue_explicitValueWinsOverDefault(t *testing.T) {
	// No hierarchy entry at all for test.Config — if EffectiveValue
	// fell through to ResolveDefault here, it would come back false;
	// the point of this test is that it must not even try, since the
	// usage site already gave an explicit value.
	h := hierarchy.NewForTesting(nil)
	ann := classfile.Annotation{Type: "test.Config", Values: map[string]classfile.ElementValue{
		"value": {Kind: classfile.ElementConst, Const: "explicit"},
	}}

	got, ok, _ := EffectiveValue(h, ann, "value")
	if !ok || got.Const != "explicit" {
		t.Errorf("EffectiveValue = %+v, ok=%v, want const \"explicit\"", got, ok)
	}
}

func TestEffectiveValue_fallsBackToDeclaredDefault(t *testing.T) {
	want := classfile.ElementValue{Kind: classfile.ElementConst, Const: "default"}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN: "test.Config",
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "value", DefaultValue: &want}},
			},
		},
	})
	ann := classfile.Annotation{Type: "test.Config"} // no explicit value at the usage site

	got, ok, _ := EffectiveValue(h, ann, "value")
	if !ok || got.Const != "default" {
		t.Errorf("EffectiveValue = %+v, ok=%v, want const \"default\"", got, ok)
	}
}

func TestRepeatableContainer_unresolvableTypeIsFalse(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	_, ok, _ := RepeatableContainer(h, "test.NoSuchAnnotation")
	if ok {
		t.Error("RepeatableContainer(unresolvable type) = true, want false")
	}
}

// TestExpandRepeatable_nonArrayLeftAlone confirms an ordinary annotation
// with no "value" array element (the common case) is passed through
// unchanged, with no hierarchy/jar access attempted at all.
func TestExpandRepeatable_nonArrayLeftAlone(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	anns := []classfile.Annotation{{Type: "test.Component"}}

	got, warnings := ExpandRepeatable(h, anns)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if len(got) != 1 || got[0].Type != "test.Component" {
		t.Errorf("ExpandRepeatable = %+v, want unchanged", got)
	}
}

// TestExpandRepeatable_mismatchedNestedTypesLeftAlone confirms shape
// detection requires every array element to be the *same* nested
// annotation type — an array mixing two different annotation types is
// never mistaken for a repeatable container.
func TestExpandRepeatable_mismatchedNestedTypesLeftAlone(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	anns := []classfile.Annotation{{
		Type: "test.Mixed",
		Values: map[string]classfile.ElementValue{
			"value": {Kind: classfile.ElementArray, Array: []classfile.ElementValue{
				{Kind: classfile.ElementAnnotation, Nested: &classfile.Annotation{Type: "test.A"}},
				{Kind: classfile.ElementAnnotation, Nested: &classfile.Annotation{Type: "test.B"}},
			}},
		},
	}}

	got, _ := ExpandRepeatable(h, anns)
	if len(got) != 1 || got[0].Type != "test.Mixed" {
		t.Errorf("ExpandRepeatable(mismatched nested types) = %+v, want unchanged", got)
	}
}

// TestExpandRepeatable_containerMismatchLeftAlone confirms that shape
// (a) alone — same-typed nested annotations in a "value" array — is
// not enough: the nested type's own @Repeatable must actually name
// *this* annotation as its container. Here test.Item is unresolvable in
// the hierarchy at all, so RepeatableContainer comes back false and the
// outer annotation is left untouched, not misidentified as a container
// by shape alone (see ExpandRepeatable's own doc comment on why this
// matters: an unrelated annotation-array element, e.g. Spring's
// @ComponentScan#excludeFilters, has exactly this same shape).
func TestExpandRepeatable_containerMismatchLeftAlone(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	anns := []classfile.Annotation{{
		Type: "test.Items",
		Values: map[string]classfile.ElementValue{
			"value": {Kind: classfile.ElementArray, Array: []classfile.ElementValue{
				{Kind: classfile.ElementAnnotation, Nested: &classfile.Annotation{Type: "test.Item"}},
			}},
		},
	}}

	got, _ := ExpandRepeatable(h, anns)
	if len(got) != 1 || got[0].Type != "test.Items" {
		t.Errorf("ExpandRepeatable(unresolvable nested type) = %+v, want unchanged", got)
	}
}
