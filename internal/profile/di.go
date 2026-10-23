// FRAMEWORKS F2's own "pierwsze DI" slice: finding Spring's own
// dependency-injection points and resolving each one, by declared
// type, against the bean catalog F1 already built (BeanNames) — using
// @Qualifier and @Primary to disambiguate when more than one bean
// matches by type, the same way Spring's own container does.
//
// Scoped deliberately narrow: field and constructor injection only
// (no setter injection — rare in modern Spring code, and this
// increment's own marginal value for it is low), single-bean-by-type
// only (no `List<Bean>`/`Map<String,Bean>` collection injection), and
// resolution against annotation-declared beans only (BeanNames) — not
// yet combined with F1's own XML `<bean>` inventory, which lives in a
// different package (internal/springxml) this one does not import,
// to keep this increment's own blast radius small.
package profile

import (
	"fmt"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

const (
	autowiredFQCN = "org.springframework.beans.factory.annotation.Autowired"
	qualifierFQCN = "org.springframework.beans.factory.annotation.Qualifier"
)

// InjectionPoint is one DI injection point found on an application
// class — a field, or a constructor parameter of the one constructor
// Spring actually uses to build the bean.
type InjectionPoint struct {
	FQCN string
	// Member is the field's own name, or "<init>" for a constructor
	// parameter.
	Member string
	// ParamIndex is -1 for a field, or the constructor parameter's own
	// index (0-based) otherwise.
	ParamIndex int
	// Type is the injected type's own FQCN (the field's/parameter's
	// declared type, not any resolved bean's — that is
	// ResolvedInjection's own job).
	Type string
	// Qualifier is "" when no `@Qualifier` is given.
	Qualifier string
}

// InjectionPoints finds every DI injection point on h's application
// classes: a field carrying `@Autowired`, and each parameter of the
// constructor Spring would actually use to build the class — the one
// `@Autowired`-annotated constructor if there is one, or (Spring
// 4.3's own implicit-autowiring rule) the sole constructor when a
// class declares exactly one and none is annotated. A class with
// several constructors and none `@Autowired` has no basis to pick one,
// so it contributes no constructor injection points at all — never a
// guessed choice.
func InjectionPoints(h *hierarchy.Hierarchy) []InjectionPoint {
	var out []InjectionPoint
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp {
			continue
		}
		for _, f := range ci.Fields {
			if !hasAnnotationType(f.AnnotationDetails, autowiredFQCN) {
				continue
			}
			typ, ok := classfile.FieldTypeFQCN(f.Descriptor)
			if !ok {
				continue
			}
			out = append(out, InjectionPoint{
				FQCN: fqcn, Member: f.Name, ParamIndex: -1,
				Type: typ, Qualifier: qualifierValue(f.AnnotationDetails),
			})
		}
		out = append(out, constructorInjectionPoints(fqcn, ci)...)
	}
	return out
}

func constructorInjectionPoints(fqcn string, ci *hierarchy.ClassInfo) []InjectionPoint {
	var ctors []hierarchy.Method
	for _, m := range ci.Methods {
		if m.Name == "<init>" {
			ctors = append(ctors, m)
		}
	}
	var chosen *hierarchy.Method
	for i := range ctors {
		if hasAnnotationType(ctors[i].AnnotationDetails, autowiredFQCN) {
			chosen = &ctors[i]
			break
		}
	}
	if chosen == nil && len(ctors) == 1 {
		chosen = &ctors[0]
	}
	if chosen == nil {
		return nil
	}

	args, err := classfile.ArgumentSlots(chosen.Descriptor, false)
	if err != nil {
		return nil
	}
	var out []InjectionPoint
	for i, a := range args {
		typ, ok := classfile.FieldTypeFQCN(a.Type)
		if !ok {
			continue
		}
		qualifier := ""
		if i < len(chosen.ParameterAnnotationDetails) {
			qualifier = qualifierValue(chosen.ParameterAnnotationDetails[i])
		}
		out = append(out, InjectionPoint{FQCN: fqcn, Member: "<init>", ParamIndex: i, Type: typ, Qualifier: qualifier})
	}
	return out
}

func hasAnnotationType(anns []classfile.Annotation, fqcn string) bool {
	for _, a := range anns {
		if a.Type == fqcn {
			return true
		}
	}
	return false
}

func qualifierValue(anns []classfile.Annotation) string {
	for _, a := range anns {
		if a.Type != qualifierFQCN {
			continue
		}
		if v, ok := a.Value("value"); ok {
			if s, ok := v.Const.(string); ok {
				return s
			}
		}
	}
	return ""
}

// ResolvedInjection is one InjectionPoint resolved against a bean
// catalog.
type ResolvedInjection struct {
	InjectionPoint
	// Candidates is every known bean (exact type match or a subtype)
	// assignable to Type, in hierarchy order.
	Candidates []string
	// Resolved is the one candidate Spring would actually wire here,
	// or "" when that can't be determined without guessing.
	Resolved string
	// Reason explains why Resolved is "" (no candidate, or more than
	// one with no @Qualifier/@Primary to disambiguate) — "" when
	// Resolved is non-empty.
	Reason string
}

// ResolveInjections resolves every InjectionPoint against beans — the
// same disambiguation order Spring's own container uses: an exact
// type match first (trivial when there is only one candidate), then
// @Qualifier (the injection point's own declared qualifier matched
// against a candidate bean's own explicit name), then @Primary.
// More than one candidate surviving all three is reported Unknown
// (Resolved == "", Reason explains why), never picked arbitrarily.
func ResolveInjections(h *hierarchy.Hierarchy, points []InjectionPoint, beans []BeanName) []ResolvedInjection {
	byFQCN := make(map[string]BeanName, len(beans))
	beanFQCNs := make([]string, 0, len(beans))
	for _, b := range beans {
		byFQCN[b.FQCN] = b
		beanFQCNs = append(beanFQCNs, b.FQCN)
	}

	out := make([]ResolvedInjection, len(points))
	for i, p := range points {
		ri := ResolvedInjection{InjectionPoint: p}
		subtypeSet := make(map[string]bool)
		for _, s := range h.Subtypes(p.Type) {
			subtypeSet[s] = true
		}
		for _, beanFQCN := range beanFQCNs {
			if beanFQCN == p.Type || subtypeSet[beanFQCN] {
				ri.Candidates = append(ri.Candidates, beanFQCN)
			}
		}

		switch len(ri.Candidates) {
		case 0:
			ri.Reason = fmt.Sprintf("no known bean assignable to %s", p.Type)
		case 1:
			ri.Resolved = ri.Candidates[0]
		default:
			ri.Resolved, ri.Reason = disambiguate(ri.Candidates, byFQCN, p.Qualifier)
		}
		out[i] = ri
	}
	return out
}

// disambiguate picks the one candidate a @Qualifier match or a unique
// @Primary settles on, or reports why neither did.
func disambiguate(candidates []string, byFQCN map[string]BeanName, qualifier string) (resolved, reason string) {
	if qualifier != "" {
		var matches []string
		for _, c := range candidates {
			if byFQCN[c].Name == qualifier {
				matches = append(matches, c)
			}
		}
		if len(matches) == 1 {
			return matches[0], ""
		}
	}
	var primaries []string
	for _, c := range candidates {
		if byFQCN[c].Primary {
			primaries = append(primaries, c)
		}
	}
	if len(primaries) == 1 {
		return primaries[0], ""
	}
	return "", fmt.Sprintf("%d candidates, no @Qualifier/@Primary disambiguates them", len(candidates))
}
