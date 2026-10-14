package classfile

import (
	"bytes"
	"fmt"
	"strings"
)

// parseAnnotations parses a RuntimeVisibleAnnotations attribute's bytes
// into the FQCNs of the annotation types present (ariadne's profile system
// only needs to know which annotations are on a class/method, never their
// element values).
func parseAnnotations(data []byte, p pool) ([]string, error) {
	rd := newReader(bytes.NewReader(data))
	count := rd.u2()
	names := make([]string, 0, count)
	for i := uint16(0); i < count && rd.err == nil; i++ {
		name, err := skipAnnotation(rd, p)
		if err != nil {
			return nil, fmt.Errorf("annotation %d: %w", i, err)
		}
		names = append(names, name)
	}
	if rd.err != nil {
		return nil, rd.err
	}
	return names, nil
}

// parseParameterAnnotations parses a RuntimeVisibleParameterAnnotations
// attribute's bytes (JVM spec 4.7.18: a u1 num_parameters, then that many
// annotation lists, each shaped exactly like RuntimeVisibleAnnotations'
// own body) into one FQCN slice per parameter.
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
func parseParameterAnnotations(data []byte, p pool) ([][]string, error) {
	rd := newReader(bytes.NewReader(data))
	numParams := rd.u1()
	out := make([][]string, 0, numParams)
	for i := 0; i < int(numParams) && rd.err == nil; i++ {
		count := rd.u2()
		names := make([]string, 0, count)
		for j := uint16(0); j < count && rd.err == nil; j++ {
			name, err := skipAnnotation(rd, p)
			if err != nil {
				return nil, fmt.Errorf("parameter %d, annotation %d: %w", i, j, err)
			}
			names = append(names, name)
		}
		out = append(out, names)
	}
	if rd.err != nil {
		return nil, rd.err
	}
	return out, nil
}

// skipAnnotation reads one `annotation` structure (type_index,
// num_element_value_pairs, then that many (name_index, element_value)
// pairs) and returns its FQCN. It must correctly skip every
// element_value, however deeply nested, even though ariadne never looks
// at the values themselves.
func skipAnnotation(rd *reader, p pool) (string, error) {
	typeIdx := rd.u2()
	numPairs := rd.u2()
	for i := uint16(0); i < numPairs && rd.err == nil; i++ {
		rd.u2() // element_name_index
		if err := skipElementValue(rd, p); err != nil {
			return "", err
		}
	}
	if rd.err != nil {
		return "", rd.err
	}
	desc, err := p.utf8(typeIdx)
	if err != nil {
		return "", fmt.Errorf("type_index: %w", err)
	}
	return descriptorToFQCN(desc), nil
}

// descriptorToFQCN converts a field descriptor for a class type, "Lfoo/Bar;",
// to "foo.Bar". Annotation type_index always points to such a descriptor.
func descriptorToFQCN(desc string) string {
	d := strings.TrimPrefix(desc, "L")
	d = strings.TrimSuffix(d, ";")
	return binaryNameToFQCN(d)
}

// skipElementValue reads one element_value structure, discarding it; see
// JVM spec 4.7.16.1. The tag byte selects the shape, including 'e' (enum
// constant), 'c' (class literal), '@' (nested annotation) and '['
// (array, recursive).
func skipElementValue(rd *reader, p pool) error {
	tag := rd.u1()
	switch tag {
	case 'B', 'C', 'D', 'F', 'I', 'J', 'S', 'Z', 's':
		rd.u2() // const_value_index
	case 'e':
		rd.u2() // type_name_index
		rd.u2() // const_name_index
	case 'c':
		rd.u2() // class_info_index
	case '@':
		if _, err := skipAnnotation(rd, p); err != nil {
			return err
		}
	case '[':
		n := rd.u2()
		for i := uint16(0); i < n && rd.err == nil; i++ {
			if err := skipElementValue(rd, p); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown element_value tag %q", tag)
	}
	return rd.err
}
