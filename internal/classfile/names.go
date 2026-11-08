package classfile

import (
	"sort"
	"strings"
)

// NamedClasses returns the classes this class names in a form code can
// turn into a Class object at run time (GRAPH_PRECISION P2): class
// literals (an ldc/ldc_w of a Class constant, e.g. `Foo.class`) and string
// constants spelled like a binary class name ("com.example.Foo" or
// "com/example/Foo", the argument Class.forName and friends take). Array
// classes are left out. Sorted, deduplicated, dotted FQCNs; a malformed
// method body contributes nothing rather than failing the class.
func (cf *ClassFile) NamedClasses() []string {
	seen := map[string]bool{}
	for i := 1; i < len(cf.pool); i++ {
		e := cf.pool[i]
		if e.tag != tagString {
			continue
		}
		s, err := cf.pool.utf8(e.nameIndex)
		if err != nil || !looksLikeClassName(s) {
			continue
		}
		seen[binaryNameToFQCN(s)] = true
	}
	for _, m := range cf.Methods {
		if m.Code == nil {
			continue
		}
		ins, err := decodeInstructions(m.Code)
		if err != nil {
			continue
		}
		for _, in := range ins {
			var idx uint16
			switch in.opcode {
			case opLdc:
				idx = uint16(in.operand[0])
			case 0x13: // ldc_w
				idx = u16(in.operand)
			default:
				continue
			}
			e, err := cf.pool.get(idx)
			if err != nil || e.tag != tagClass {
				continue
			}
			name, err := cf.pool.utf8(e.nameIndex)
			if err != nil || strings.HasPrefix(name, "[") {
				continue
			}
			seen[binaryNameToFQCN(name)] = true
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// looksLikeClassName: at least one package separator, every segment a
// Java identifier (letters, digits, '_', '$', not starting with a digit).
func looksLikeClassName(s string) bool {
	if len(s) < 3 || len(s) > 512 || !strings.ContainsAny(s, "./") {
		return false
	}
	if strings.Contains(s, ".") && strings.Contains(s, "/") {
		return false
	}
	seg := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '.' || c == '/':
			if seg == 0 {
				return false
			}
			seg = 0
			continue
		case c >= '0' && c <= '9':
			if seg == 0 {
				return false
			}
		case c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c >= 0x80:
		default:
			return false
		}
		seg++
	}
	return seg > 0
}
