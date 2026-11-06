// Package relocate applies explicit package relocations (CVE:F7a): a
// component's original package prefix "from" appears in an artifact as
// "to" (maven-shade's pattern/shadedPattern, Gradle Shadow's relocate).
//
// A prefix matches only at a package boundary: "org.objectweb" relocates
// "org.objectweb.asm.Type" and "org.objectweb" itself, never
// "org.objectwebx.Y". When several rules match, the longest "from" wins.
// Names are rewritten wherever a class name can occur: binary names
// ("a.b.C$D"), internal names ("a/b/C$D"), field and method descriptors
// (including arrays) and method selectors ("a.b.C#m(La/b/D;)V").
package relocate

import (
	"fmt"
	"sort"
	"strings"
)

// Rule relocates package prefix From (dotted) to To.
type Rule struct {
	From, To string
	// Source says where the rule came from (e.g. "declared",
	// "build-config: META-INF/maven/g/a/pom.xml"); evidence only.
	Source string
}

// Set is an ordered set of rules, longest From first.
type Set struct {
	rules []Rule
}

// New validates rules and returns a Set. Two rules with the same From
// and different To are a conflict; a rule whose From or To is not a
// package name is rejected.
func New(rules []Rule) (Set, error) {
	seen := map[string]string{}
	var out []Rule
	for _, r := range rules {
		if !validPackage(r.From) || !validPackage(r.To) {
			return Set{}, fmt.Errorf("relocation %q -> %q: not package names", r.From, r.To)
		}
		if r.From == r.To {
			continue
		}
		if to, ok := seen[r.From]; ok {
			if to != r.To {
				return Set{}, fmt.Errorf("relocation of %q to both %q and %q", r.From, to, r.To)
			}
			continue
		}
		seen[r.From] = r.To
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool { return len(out[i].From) > len(out[j].From) })
	return Set{rules: out}, nil
}

func validPackage(p string) bool {
	if p == "" || strings.HasPrefix(p, ".") || strings.HasSuffix(p, ".") || strings.Contains(p, "..") {
		return false
	}
	for _, c := range p {
		if !(c == '.' || c == '_' || c == '$' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// Empty reports whether the set has no rules.
func (s Set) Empty() bool { return len(s.rules) == 0 }

// Rules returns the rules, longest From first.
func (s Set) Rules() []Rule { return append([]Rule(nil), s.rules...) }

// Inverse maps relocated names back to the original ones.
func (s Set) Inverse() Set {
	inv := make([]Rule, len(s.rules))
	for i, r := range s.rules {
		inv[i] = Rule{From: r.To, To: r.From, Source: r.Source}
	}
	sort.SliceStable(inv, func(i, j int) bool { return len(inv[i].From) > len(inv[j].From) })
	return Set{rules: inv}
}

// Merge returns the union of two sets; a conflict is an error.
func Merge(a, b Set) (Set, error) {
	return New(append(a.Rules(), b.Rules()...))
}

// Name rewrites one class or package name, dotted or slashed (sep is
// detected from the name: a name with '/' is internal form).
func (s Set) Name(name string) string {
	sep := "."
	if strings.Contains(name, "/") {
		sep = "/"
	}
	for _, r := range s.rules {
		from, to := r.From, r.To
		if sep == "/" {
			from, to = strings.ReplaceAll(from, ".", "/"), strings.ReplaceAll(to, ".", "/")
		}
		if name == from {
			return to
		}
		if strings.HasPrefix(name, from+sep) {
			return to + name[len(from):]
		}
	}
	return name
}

// Changes reports whether Name would rewrite name.
func (s Set) Changes(name string) bool { return s.Name(name) != name }

// Descriptor rewrites every class name of a field or method descriptor.
func (s Set) Descriptor(desc string) string {
	if s.Empty() || !strings.Contains(desc, "L") {
		return desc
	}
	var b strings.Builder
	for i := 0; i < len(desc); {
		c := desc[i]
		if c != 'L' {
			b.WriteByte(c)
			i++
			continue
		}
		end := strings.IndexByte(desc[i:], ';')
		if end < 0 {
			b.WriteString(desc[i:])
			break
		}
		b.WriteByte('L')
		b.WriteString(s.Name(desc[i+1 : i+end]))
		b.WriteByte(';')
		i += end + 1
	}
	return b.String()
}

// Selector rewrites a method selector "Class#name[(params)[ret]]".
func (s Set) Selector(sel string) string {
	cls, rest, ok := strings.Cut(sel, "#")
	if !ok {
		return sel
	}
	if i := strings.IndexByte(rest, '('); i >= 0 {
		rest = rest[:i] + s.Descriptor(rest[i:])
	}
	return s.Name(cls) + "#" + rest
}

// String rewrites a string constant that is a class or package name in
// dotted or internal form starting with a relocated prefix (shading
// tools rewrite such literals too); any other string is unchanged.
func (s Set) String(str string) string {
	if strings.ContainsAny(str, " ;()<>") {
		return str
	}
	return s.Name(str)
}

// Under returns the rule whose target package contains name (dotted or
// internal form), if any.
func (s Set) Under(name string) (Rule, bool) {
	sep := "."
	if strings.Contains(name, "/") {
		sep = "/"
	}
	for _, r := range s.rules {
		to := r.To
		if sep == "/" {
			to = strings.ReplaceAll(to, ".", "/")
		}
		if name == to || strings.HasPrefix(name, to+sep) {
			return r, true
		}
	}
	return Rule{}, false
}
