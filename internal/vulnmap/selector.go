package vulnmap

import (
	"fmt"
	"regexp"
	"strings"
)

// Match precisions of a method selector (CVE:F4).
const (
	PrecisionName   = "name"   // Class#name: every overload
	PrecisionParams = "params" // Class#name(params): one parameter list, any return type
	PrecisionExact  = "exact"  // Class#name(params)ret: one JVM method
	PrecisionMixed  = "mixed"  // a mapping whose selectors differ
)

// Selector is one parsed "Class#name[(params)[ret]]" method selector.
type Selector struct {
	Class, Name string
	// Params is the parameter descriptor including parentheses, e.g.
	// "(Ljava/lang/String;)"; Return the return descriptor ("V", ...).
	Params, Return string
}

// Precision is the selector's match precision.
func (s Selector) Precision() string {
	switch {
	case s.Params == "":
		return PrecisionName
	case s.Return == "":
		return PrecisionParams
	}
	return PrecisionExact
}

func (s Selector) String() string { return s.Class + "#" + s.Name + s.Params + s.Return }

var (
	fieldType = `\[*(?:[BCDFIJSZ]|L[^;\[\(\)\.]+;)`
	paramsRE  = regexp.MustCompile(`^\((?:` + fieldType + `)*\)$`)
	returnRE  = regexp.MustCompile(`^(?:V|` + fieldType + `)$`)
	classRE   = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*(\.[A-Za-z_$][A-Za-z0-9_$]*)*$`)
	methodRE  = regexp.MustCompile(`^(<init>|<clinit>|[A-Za-z_$][A-Za-z0-9_$]*)$`)
)

// ParseSelector parses and validates a selector.
func ParseSelector(spec string) (Selector, error) {
	cls, rest, ok := strings.Cut(spec, "#")
	if !ok {
		return Selector{}, fmt.Errorf("selector %q: want Class#method", spec)
	}
	s := Selector{Class: cls, Name: rest}
	if i := strings.IndexByte(rest, '('); i >= 0 {
		j := strings.IndexByte(rest, ')')
		if j < i {
			return Selector{}, fmt.Errorf("selector %q: unbalanced descriptor", spec)
		}
		s.Name, s.Params, s.Return = rest[:i], rest[i:j+1], rest[j+1:]
		if !paramsRE.MatchString(s.Params) {
			return Selector{}, fmt.Errorf("selector %q: %q is not a JVM parameter descriptor", spec, s.Params)
		}
		if s.Return != "" && !returnRE.MatchString(s.Return) {
			return Selector{}, fmt.Errorf("selector %q: %q is not a JVM return descriptor", spec, s.Return)
		}
	}
	if !classRE.MatchString(s.Class) {
		return Selector{}, fmt.Errorf("selector %q: %q is not a binary class name", spec, s.Class)
	}
	if !methodRE.MatchString(s.Name) {
		return Selector{}, fmt.Errorf("selector %q: %q is not a method name", spec, s.Name)
	}
	if s.Name == "<init>" && s.Return != "" && s.Return != "V" {
		return Selector{}, fmt.Errorf("selector %q: a constructor returns V", spec)
	}
	return s, nil
}

// MappingPrecision is the precision shared by every selector, or
// PrecisionMixed; "" for none. Unparsable selectors count as name.
func MappingPrecision(methods []string) string {
	p := ""
	for _, m := range methods {
		q := PrecisionName
		if s, err := ParseSelector(m); err == nil {
			q = s.Precision()
		}
		if p == "" {
			p = q
		} else if p != q {
			return PrecisionMixed
		}
	}
	return p
}
