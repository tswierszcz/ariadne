package research

import "strings"

// ValidMethodDescriptor reports whether d is a JVM method descriptor
// (JVMS 4.3.3): "(" parameter field types ")" and a return type, which is
// a field type or V.
func ValidMethodDescriptor(d string) bool {
	if !strings.HasPrefix(d, "(") {
		return false
	}
	i := 1
	for i < len(d) && d[i] != ')' {
		n := fieldType(d[i:])
		if n == 0 {
			return false
		}
		i += n
	}
	if i >= len(d) || d[i] != ')' {
		return false
	}
	ret := d[i+1:]
	if ret == "V" {
		return true
	}
	return ret != "" && fieldType(ret) == len(ret)
}

// fieldType returns the length of the field type descriptor at the start
// of s, or 0 when there is none.
func fieldType(s string) int {
	dims := 0
	for dims < len(s) && s[dims] == '[' {
		dims++
	}
	if dims > 255 || dims == len(s) {
		return 0
	}
	switch s[dims] {
	case 'B', 'C', 'D', 'F', 'I', 'J', 'S', 'Z':
		return dims + 1
	case 'L':
		end := strings.IndexByte(s[dims:], ';')
		if end <= 1 {
			return 0
		}
		name := s[dims+1 : dims+end]
		if strings.ContainsAny(name, ".[;") {
			return 0
		}
		for _, part := range strings.Split(name, "/") {
			if part == "" {
				return 0
			}
		}
		return dims + end + 1
	}
	return 0
}
