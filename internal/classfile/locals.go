package classfile

import "fmt"

// Argument is one parameter's local-variable slot assignment (JVMS
// 2.6.1) — DATA_FLOW D1's own "argumenty wynikające z deskryptora i
// kategorie jedno-/dwusłowowe [...] zmiennych lokalnych".
type Argument struct {
	Slot  int    // the local-variable slot this parameter starts at
	Width int    // 1, or 2 for a long/double (JVMS 2.6.1: a category-2 type occupies two slots)
	Type  string // the raw field descriptor, e.g. "J", "Ljava/lang/String;", "[I"
}

// ArgumentSlots returns, in order, every parameter descriptor's slot
// assignment: an instance method's implicit `this` occupies slot 0
// (so the first real parameter starts at slot 1); a long/double
// parameter occupies two consecutive slots, same as a local variable
// of that type would (the reason a later parameter's slot number isn't
// just its index).
func ArgumentSlots(descriptor string, isStatic bool) ([]Argument, error) {
	params, err := parseParamDescriptors(descriptor)
	if err != nil {
		return nil, err
	}
	slot := 0
	if !isStatic {
		slot = 1 // `this`
	}
	args := make([]Argument, len(params))
	for i, p := range params {
		w := slotWidth(p)
		args[i] = Argument{Slot: slot, Width: w, Type: p}
		slot += w
	}
	return args, nil
}

// slotWidth is 2 for exactly "J" (long) or "D" (double) — an array of
// either, e.g. "[J", is still a single reference and so width 1, like
// every other reference type.
func slotWidth(fieldDescriptor string) int {
	if fieldDescriptor == "J" || fieldDescriptor == "D" {
		return 2
	}
	return 1
}

// ParamDescriptors is parseParamDescriptors, exported for a caller
// (REFLECTION R2) that needs a method's parameter descriptors alone,
// without a return type to match against — a reflective `getMethod`
// lookup never provides one, unlike ArgumentSlots's own callers, which
// always have a full descriptor already.
func ParamDescriptors(descriptor string) ([]string, error) {
	return parseParamDescriptors(descriptor)
}

// parseParamDescriptors splits a method descriptor's parameter section
// "(...)" into individual field descriptors (JVMS 4.3.2/4.3.3).
func parseParamDescriptors(descriptor string) ([]string, error) {
	if len(descriptor) == 0 || descriptor[0] != '(' {
		return nil, fmt.Errorf("invalid method descriptor %q: missing leading '('", descriptor)
	}
	i := 1
	var params []string
	for i < len(descriptor) && descriptor[i] != ')' {
		start := i
		for descriptor[i] == '[' {
			i++
			if i >= len(descriptor) {
				return nil, fmt.Errorf("invalid method descriptor %q: truncated array type", descriptor)
			}
		}
		switch descriptor[i] {
		case 'B', 'C', 'D', 'F', 'I', 'J', 'S', 'Z':
			i++
		case 'L':
			for i < len(descriptor) && descriptor[i] != ';' {
				i++
			}
			if i >= len(descriptor) {
				return nil, fmt.Errorf("invalid method descriptor %q: unterminated object type", descriptor)
			}
			i++ // consume ';'
		default:
			return nil, fmt.Errorf("invalid method descriptor %q: unknown field type %q at index %d", descriptor, descriptor[i], i)
		}
		params = append(params, descriptor[start:i])
	}
	if i >= len(descriptor) {
		return nil, fmt.Errorf("invalid method descriptor %q: missing closing ')'", descriptor)
	}
	return params, nil
}
