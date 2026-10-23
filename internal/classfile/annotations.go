package classfile

import (
	"bytes"
	"fmt"
	"strings"
)

// ElementKind is an annotation element_value's own shape (JVM spec
// 4.7.16.1) — FRAMEWORKS F1's own "wartości, typy, tablice" requirement:
// before F1, ariadne only kept an annotation's type FQCN and discarded
// every element value entirely (see Annotation's own doc comment for
// why that was no longer enough).
type ElementKind string

const (
	// ElementConst is a literal primitive/String — exactly one of
	// ElementValue's Const is set, holding a Go value of the matching
	// type (int8/uint16/int32/int64/float32/float64/bool/rune/string).
	ElementConst ElementKind = "const"
	// ElementEnum is an enum constant — ElementValue.EnumType/EnumConst.
	ElementEnum ElementKind = "enum"
	// ElementClass is a `Foo.class` literal — ElementValue.ClassFQCN.
	ElementClass ElementKind = "class"
	// ElementAnnotation is a nested annotation — ElementValue.Nested.
	ElementAnnotation ElementKind = "annotation"
	// ElementArray is an array of element values, each independently
	// one of the above kinds (never itself ElementArray — JVM spec
	// 4.7.16.1 does not nest arrays inside arrays for an annotation
	// element) — ElementValue.Array.
	ElementArray ElementKind = "array"
)

// ElementValue is one annotation element's value, fully decoded —
// distinct from a mere "this element exists" fact. Exactly one field
// group is meaningful, selected by Kind, following the same
// one-Kind-many-fields discipline DATA_FLOW's own valueflow.Value uses.
type ElementValue struct {
	Kind ElementKind

	// Const is set only when Kind == ElementConst.
	Const any

	// EnumType/EnumConst are set only when Kind == ElementEnum: the
	// enum type's own FQCN and the constant's name (e.g.
	// "org.springframework.web.bind.annotation.RequestMethod", "GET").
	EnumType, EnumConst string

	// ClassFQCN is set only when Kind == ElementClass: the referenced
	// type's FQCN for an ordinary reference type ("Lfoo/Bar;" ->
	// "foo.Bar"), or the raw descriptor itself for a primitive or array
	// type (e.g. "I", "[Ljava/lang/String;") — rare in practice for an
	// annotation element, and not worth a full descriptor-to-source
	// renderer just for this.
	ClassFQCN string

	// Nested is set only when Kind == ElementAnnotation.
	Nested *Annotation

	// Array is set only when Kind == ElementArray — always non-nil
	// (possibly empty) for that Kind.
	Array []ElementValue
}

// Annotation is one `@Foo(...)` usage, fully decoded: its type and every
// explicitly-given element's value. FRAMEWORKS F1's own motivating gap —
// before F1, ariadne's profile system could tell a class was annotated
// `@Profile` or `@RequestMapping`, but never which profile name or which
// path, exactly the information F1's own "Gotowe" requires ("wartości
// qualifiera, profilu i trasy mają źródło"). An element not explicitly
// given at the usage site (relying on the annotation type's own
// declared default) is simply absent from Values — resolving that
// default needs the annotation type's own class file, read via
// hierarchy access this package deliberately doesn't have (see
// internal/dataflow's own re-read-the-class-file precedent for
// everything else interprocedural).
type Annotation struct {
	Type   string
	Values map[string]ElementValue
}

// Value is Values[name] with the ok idiom, so a caller doesn't need to
// know Values can be nil for an annotation with no explicit elements.
func (a Annotation) Value(name string) (ElementValue, bool) {
	v, ok := a.Values[name]
	return v, ok
}

// annotationNames adapts a []Annotation to the plain FQCN list every
// existing consumer (internal/profile's own hasAny-style name matching)
// already expects — FRAMEWORKS F1's own explicit requirement to keep
// the old name lists working through an adapter rather than forcing
// every such consumer to learn the richer shape at once.
func annotationNames(anns []Annotation) []string {
	if anns == nil {
		return nil
	}
	names := make([]string, len(anns))
	for i, a := range anns {
		names[i] = a.Type
	}
	return names
}

// parseAnnotations parses a RuntimeVisibleAnnotations/
// RuntimeInvisibleAnnotations attribute's bytes into every annotation
// present, values included. The two attributes share exactly the same
// body shape (JVM spec 4.7.16/4.7.17); FRAMEWORKS F1's own requirement
// to not ignore invisible annotations means a caller merges both
// attributes' results into one list, same as ariadne already merges
// multiple sources elsewhere (e.g. JVM_GRAPH G1's duplicate-class
// warnings) rather than this function knowing about retention policy at
// all.
func parseAnnotations(data []byte, p pool) ([]Annotation, error) {
	rd := newReader(bytes.NewReader(data))
	count := rd.u2()
	out := make([]Annotation, 0, count)
	for i := uint16(0); i < count && rd.err == nil; i++ {
		a, err := parseAnnotation(rd, p)
		if err != nil {
			return nil, fmt.Errorf("annotation %d: %w", i, err)
		}
		out = append(out, a)
	}
	if rd.err != nil {
		return nil, rd.err
	}
	return out, nil
}

// parseParameterAnnotations parses a RuntimeVisibleParameterAnnotations/
// RuntimeInvisibleParameterAnnotations attribute's bytes (JVM spec
// 4.7.18/4.7.19: a u1 num_parameters, then that many annotation lists,
// each shaped exactly like RuntimeVisibleAnnotations' own body) into one
// []Annotation slice per parameter.
//
// Known javac quirk, not handled here: for a few parameter kinds (e.g. a
// non-static inner class constructor's implicit outer-class parameter,
// or an enum constructor's implicit name/ordinal parameters), some javac
// versions emit a num_parameters lower than the method descriptor's own
// parameter count, omitting the synthetic ones from the front. Nothing
// in ariadne looks at parameter annotations on a constructor today (CDI
// observer methods, @Observes' own use, are never constructors), so this
// is not realigned against the descriptor; a caller matching
// ParameterAnnotations[i] against descriptor parameter i could be off by
// one for such a method.
func parseParameterAnnotations(data []byte, p pool) ([][]Annotation, error) {
	rd := newReader(bytes.NewReader(data))
	numParams := rd.u1()
	out := make([][]Annotation, 0, numParams)
	for i := 0; i < int(numParams) && rd.err == nil; i++ {
		count := rd.u2()
		anns := make([]Annotation, 0, count)
		for j := uint16(0); j < count && rd.err == nil; j++ {
			a, err := parseAnnotation(rd, p)
			if err != nil {
				return nil, fmt.Errorf("parameter %d, annotation %d: %w", i, j, err)
			}
			anns = append(anns, a)
		}
		out = append(out, anns)
	}
	if rd.err != nil {
		return nil, rd.err
	}
	return out, nil
}

// mergeParameterAnnotations combines a Visible and an Invisible
// RuntimeParameterAnnotations attribute's own per-parameter lists —
// either may arrive first, and either may be absent (nil), in which
// case the other's shape is used as-is; both present merges them
// pairwise. A method never declares both attributes with a different
// num_parameters in practice (they describe the same parameter list),
// so no attempt is made to reconcile mismatched lengths beyond simply
// extending the shorter one with empty slices.
func mergeParameterAnnotations(existing, incoming [][]Annotation) [][]Annotation {
	if existing == nil {
		return incoming
	}
	if incoming == nil {
		return existing
	}
	n := len(existing)
	if len(incoming) > n {
		n = len(incoming)
	}
	out := make([][]Annotation, n)
	for i := 0; i < n; i++ {
		var a, b []Annotation
		if i < len(existing) {
			a = existing[i]
		}
		if i < len(incoming) {
			b = incoming[i]
		}
		out[i] = append(append([]Annotation{}, a...), b...)
	}
	return out
}

// parseAnnotationDefault parses an AnnotationDefault attribute's bytes
// (JVM spec 4.7.22): unlike RuntimeVisibleAnnotations, this attribute's
// body is a single element_value directly, with no count prefix — it
// appears only on a method declared directly on an annotation
// interface, giving that element's default value.
func parseAnnotationDefault(data []byte, p pool) (ElementValue, error) {
	rd := newReader(bytes.NewReader(data))
	ev, err := parseElementValue(rd, p)
	if err != nil {
		return ElementValue{}, err
	}
	if rd.err != nil {
		return ElementValue{}, rd.err
	}
	return ev, nil
}

// parseAnnotation reads one `annotation` structure (type_index,
// num_element_value_pairs, then that many (name_index, element_value)
// pairs) and returns it fully decoded.
func parseAnnotation(rd *reader, p pool) (Annotation, error) {
	typeIdx := rd.u2()
	numPairs := rd.u2()
	var values map[string]ElementValue
	if numPairs > 0 {
		values = make(map[string]ElementValue, numPairs)
	}
	for i := uint16(0); i < numPairs && rd.err == nil; i++ {
		nameIdx := rd.u2()
		ev, err := parseElementValue(rd, p)
		if err != nil {
			return Annotation{}, err
		}
		if rd.err != nil {
			break
		}
		name, err := p.utf8(nameIdx)
		if err != nil {
			return Annotation{}, fmt.Errorf("element_name_index: %w", err)
		}
		values[name] = ev
	}
	if rd.err != nil {
		return Annotation{}, rd.err
	}
	desc, err := p.utf8(typeIdx)
	if err != nil {
		return Annotation{}, fmt.Errorf("type_index: %w", err)
	}
	return Annotation{Type: descriptorToFQCN(desc), Values: values}, nil
}

// descriptorToFQCN converts a field descriptor for a class type, "Lfoo/Bar;",
// to "foo.Bar". Annotation type_index always points to such a descriptor.
func descriptorToFQCN(desc string) string {
	d := strings.TrimPrefix(desc, "L")
	d = strings.TrimSuffix(d, ";")
	return binaryNameToFQCN(d)
}

// FieldTypeFQCN converts a reference-type field descriptor
// ("Lfoo/Bar;") to its dotted FQCN ("foo.Bar") — for a caller matching
// a field's or parameter's own declared type against a known class
// (FRAMEWORKS F2's own DI candidate lookup). ok is false for a
// primitive or array descriptor ("I", "[Ljava/lang/String;"): a
// dependency-injection target is always a single object reference,
// never a primitive or a collection-of-beans injection (the latter is
// a real Spring feature, deliberately out of this narrow scope).
func FieldTypeFQCN(desc string) (string, bool) {
	if strings.HasPrefix(desc, "L") && strings.HasSuffix(desc, ";") {
		return descriptorToFQCN(desc), true
	}
	return "", false
}

// classLiteralFQCN is ElementClass's own descriptor reader: unlike
// annotation type_index (always a reference type), a `Foo.class`
// element_value's class_info_index can name a primitive or array type
// too (JVM spec 4.7.16.1's own "return descriptor" wording) — "I" for
// `int.class`, "[Ljava/lang/String;" for `String[].class`. Only the
// ordinary reference-type shape is converted to a dotted FQCN; the rest
// is kept as its raw descriptor, a documented simplification (see
// ElementValue.ClassFQCN's own doc comment).
func classLiteralFQCN(desc string) string {
	if strings.HasPrefix(desc, "L") && strings.HasSuffix(desc, ";") {
		return descriptorToFQCN(desc)
	}
	return desc
}

// parseElementValue reads one element_value structure (JVM spec
// 4.7.16.1) and returns it fully decoded. The tag byte selects the
// shape, including 'e' (enum constant), 'c' (class literal), '@' (nested
// annotation) and '[' (array, recursive).
func parseElementValue(rd *reader, p pool) (ElementValue, error) {
	tag := rd.u1()
	switch tag {
	case 'B':
		idx := rd.u2()
		v, err := p.integer(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("byte const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: int8(v)}, nil
	case 'C':
		idx := rd.u2()
		v, err := p.integer(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("char const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: rune(v)}, nil
	case 'D':
		idx := rd.u2()
		v, err := p.doubleVal(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("double const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: v}, nil
	case 'F':
		idx := rd.u2()
		v, err := p.float32Val(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("float const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: v}, nil
	case 'I':
		idx := rd.u2()
		v, err := p.integer(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("int const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: v}, nil
	case 'J':
		idx := rd.u2()
		v, err := p.long(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("long const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: v}, nil
	case 'S':
		idx := rd.u2()
		v, err := p.integer(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("short const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: int16(v)}, nil
	case 'Z':
		idx := rd.u2()
		v, err := p.integer(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("boolean const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: v != 0}, nil
	case 's':
		idx := rd.u2()
		v, err := p.utf8(idx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("string const_value_index: %w", err)
		}
		return ElementValue{Kind: ElementConst, Const: v}, nil
	case 'e':
		typeIdx := rd.u2()
		constIdx := rd.u2()
		typeDesc, err := p.utf8(typeIdx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("enum type_name_index: %w", err)
		}
		constName, err := p.utf8(constIdx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("enum const_name_index: %w", err)
		}
		return ElementValue{Kind: ElementEnum, EnumType: descriptorToFQCN(typeDesc), EnumConst: constName}, nil
	case 'c':
		classIdx := rd.u2()
		// class_info_index points at a Utf8 return-descriptor, *not* a
		// Class constant pool entry (JVM spec 4.7.16.1) — p.utf8, not
		// p.className.
		desc, err := p.utf8(classIdx)
		if err != nil {
			return ElementValue{}, fmt.Errorf("class_info_index: %w", err)
		}
		return ElementValue{Kind: ElementClass, ClassFQCN: classLiteralFQCN(desc)}, nil
	case '@':
		a, err := parseAnnotation(rd, p)
		if err != nil {
			return ElementValue{}, err
		}
		return ElementValue{Kind: ElementAnnotation, Nested: &a}, nil
	case '[':
		n := rd.u2()
		arr := make([]ElementValue, 0, n)
		for i := uint16(0); i < n && rd.err == nil; i++ {
			ev, err := parseElementValue(rd, p)
			if err != nil {
				return ElementValue{}, err
			}
			arr = append(arr, ev)
		}
		if rd.err != nil {
			return ElementValue{}, rd.err
		}
		return ElementValue{Kind: ElementArray, Array: arr}, nil
	default:
		return ElementValue{}, fmt.Errorf("unknown element_value tag %q", tag)
	}
}
