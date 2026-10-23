package classfile

import (
	"reflect"
	"testing"
)

// TestParseAnnotations_fullElementValueShapes is FRAMEWORKS F1's own
// motivating example: a single hand-assembled annotation exercising
// every element_value shape at once (JVM spec 4.7.16.1) — a string
// const, an int const, an enum constant, a class literal, an array, and
// a nested annotation — built from a raw pool and raw attribute bytes
// the same way JVM_GRAPH G3's TestExtractCalls_initRefs pins down
// new/getstatic/putstatic, rather than relying on a compiled fixture
// (no javac in this environment).
//
// Represents:
//
//	@Config(name="x", level=3, method=RequestMethod.GET, type=Foo.class,
//	         tags={"a","b"}, nested=@Inner(flag=true))
func TestParseAnnotations_fullElementValueShapes(t *testing.T) {
	p := pool{
		{},                                          // 0 unused
		{tag: tagUtf8, str: "Ltest/Config;"},        // 1: annotation type
		{tag: tagUtf8, str: "name"},                 // 2
		{tag: tagUtf8, str: "x"},                    // 3
		{tag: tagUtf8, str: "level"},                // 4
		{tag: tagInteger, raw32: 3},                 // 5
		{tag: tagUtf8, str: "method"},               // 6
		{tag: tagUtf8, str: "Ltest/RequestMethod;"}, // 7
		{tag: tagUtf8, str: "GET"},                  // 8
		{tag: tagUtf8, str: "type"},                 // 9
		{tag: tagUtf8, str: "Ltest/Foo;"},           // 10
		{tag: tagUtf8, str: "tags"},                 // 11
		{tag: tagUtf8, str: "a"},                    // 12
		{tag: tagUtf8, str: "b"},                    // 13
		{tag: tagUtf8, str: "nested"},               // 14
		{tag: tagUtf8, str: "Ltest/Inner;"},         // 15
		{tag: tagUtf8, str: "flag"},                 // 16
		{tag: tagInteger, raw32: 1},                 // 17
	}

	data := []byte{
		0x00, 0x01, // num_annotations = 1
		0x00, 0x01, // type_index = 1 (Config)
		0x00, 0x06, // num_element_value_pairs = 6

		0x00, 0x02, 's', 0x00, 0x03, // name = "x"
		0x00, 0x04, 'I', 0x00, 0x05, // level = 3
		0x00, 0x06, 'e', 0x00, 0x07, 0x00, 0x08, // method = RequestMethod.GET
		0x00, 0x09, 'c', 0x00, 0x0A, // type = Foo.class
		0x00, 0x0B, '[', 0x00, 0x02, // tags = { ... } (2 elements)
		's', 0x00, 0x0C, // "a"
		's', 0x00, 0x0D, // "b"
		0x00, 0x0E, '@', // nested = @Inner(...)
		0x00, 0x0F, // nested annotation type_index = 15 (Inner)
		0x00, 0x01, // nested num_element_value_pairs = 1
		0x00, 0x10, 'Z', 0x00, 0x11, // flag = true
	}

	anns, err := parseAnnotations(data, p)
	if err != nil {
		t.Fatalf("parseAnnotations: %v", err)
	}
	if len(anns) != 1 {
		t.Fatalf("len(anns) = %d, want 1", len(anns))
	}
	a := anns[0]
	if a.Type != "test.Config" {
		t.Errorf("Type = %q, want test.Config", a.Type)
	}

	nameVal, ok := a.Value("name")
	if !ok || nameVal.Kind != ElementConst || nameVal.Const != "x" {
		t.Errorf("name = %+v, ok=%v, want const \"x\"", nameVal, ok)
	}

	levelVal, ok := a.Value("level")
	if !ok || levelVal.Kind != ElementConst || levelVal.Const != int32(3) {
		t.Errorf("level = %+v, ok=%v, want const int32(3)", levelVal, ok)
	}

	methodVal, ok := a.Value("method")
	if !ok || methodVal.Kind != ElementEnum || methodVal.EnumType != "test.RequestMethod" || methodVal.EnumConst != "GET" {
		t.Errorf("method = %+v, ok=%v, want enum test.RequestMethod#GET", methodVal, ok)
	}

	typeVal, ok := a.Value("type")
	if !ok || typeVal.Kind != ElementClass || typeVal.ClassFQCN != "test.Foo" {
		t.Errorf("type = %+v, ok=%v, want class test.Foo", typeVal, ok)
	}

	tagsVal, ok := a.Value("tags")
	if !ok || tagsVal.Kind != ElementArray {
		t.Fatalf("tags = %+v, ok=%v, want an array", tagsVal, ok)
	}
	wantTags := []ElementValue{
		{Kind: ElementConst, Const: "a"},
		{Kind: ElementConst, Const: "b"},
	}
	if !reflect.DeepEqual(tagsVal.Array, wantTags) {
		t.Errorf("tags array = %+v, want %+v", tagsVal.Array, wantTags)
	}

	nestedVal, ok := a.Value("nested")
	if !ok || nestedVal.Kind != ElementAnnotation || nestedVal.Nested == nil {
		t.Fatalf("nested = %+v, ok=%v, want a nested annotation", nestedVal, ok)
	}
	if nestedVal.Nested.Type != "test.Inner" {
		t.Errorf("nested.Type = %q, want test.Inner", nestedVal.Nested.Type)
	}
	flagVal, ok := nestedVal.Nested.Value("flag")
	if !ok || flagVal.Kind != ElementConst || flagVal.Const != true {
		t.Errorf("nested.flag = %+v, ok=%v, want const true", flagVal, ok)
	}
}

// TestAnnotationNames confirms the adapter FRAMEWORKS F1 requires: a
// caller that only ever looked at the old plain-FQCN list still gets
// exactly that, in the same order, from the richer []Annotation form.
func TestAnnotationNames(t *testing.T) {
	anns := []Annotation{{Type: "a.B"}, {Type: "c.D"}}
	got := annotationNames(anns)
	want := []string{"a.B", "c.D"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("annotationNames = %v, want %v", got, want)
	}
	if annotationNames(nil) != nil {
		t.Errorf("annotationNames(nil) = %v, want nil", annotationNames(nil))
	}
}

// TestParseAnnotationDefault exercises AnnotationDefault's own shape
// (JVM spec 4.7.22): a single element_value with no count prefix, the
// one attribute body shape that differs from every other annotation
// attribute.
func TestParseAnnotationDefault(t *testing.T) {
	p := pool{
		{},                       // 0 unused
		{tag: tagUtf8, str: "x"}, // 1
	}
	data := []byte{'s', 0x00, 0x01} // default value: the string "x"

	ev, err := parseAnnotationDefault(data, p)
	if err != nil {
		t.Fatalf("parseAnnotationDefault: %v", err)
	}
	if ev.Kind != ElementConst || ev.Const != "x" {
		t.Errorf("parseAnnotationDefault = %+v, want const \"x\"", ev)
	}
}

func TestMergeParameterAnnotations(t *testing.T) {
	visible := [][]Annotation{{{Type: "a.V"}}, nil}
	invisible := [][]Annotation{nil, {{Type: "a.I"}}}

	got := mergeParameterAnnotations(visible, invisible)
	want := [][]Annotation{{{Type: "a.V"}}, {{Type: "a.I"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeParameterAnnotations = %+v, want %+v", got, want)
	}

	if got := mergeParameterAnnotations(nil, visible); !reflect.DeepEqual(got, visible) {
		t.Errorf("mergeParameterAnnotations(nil, visible) = %+v, want %+v", got, visible)
	}
	if got := mergeParameterAnnotations(visible, nil); !reflect.DeepEqual(got, visible) {
		t.Errorf("mergeParameterAnnotations(visible, nil) = %+v, want %+v", got, visible)
	}
}

// TestClassLiteralFQCN_primitiveAndArrayStayRaw confirms the documented
// simplification: only an ordinary reference-type class literal becomes
// a dotted FQCN; a primitive or array type's own descriptor is kept
// as-is rather than rendered as pseudo-source.
func TestClassLiteralFQCN_primitiveAndArrayStayRaw(t *testing.T) {
	cases := map[string]string{
		"Ltest/Foo;":          "test.Foo",
		"I":                   "I",
		"[Ljava/lang/String;": "[Ljava/lang/String;",
	}
	for desc, want := range cases {
		if got := classLiteralFQCN(desc); got != want {
			t.Errorf("classLiteralFQCN(%q) = %q, want %q", desc, got, want)
		}
	}
}
