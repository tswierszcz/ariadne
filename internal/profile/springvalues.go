// FRAMEWORKS F1's own motivating example, made real: before this file,
// a profile could only tell that a method carried *some* HTTP-mapping
// annotation, or a class carried *some* @Profile/@Component — the
// actual path, HTTP verb, profile name, and bean qualifier (all real
// element values the bytecode always carries) were invisible. This
// sources exactly the three facts FRAMEWORKS F1's own "Gotowe"
// requires values for: route, profile, qualifier.
//
// Scoped to Spring MVC's own well-known annotation shapes (the F0/F1
// pilot) — not a general annotation-value-to-business-fact engine.
package profile

import (
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// httpMappingVerbs maps each of Spring MVC's own HTTP mapping
// annotation types to the HTTP verb(s) it implies when the
// annotation's own "method" element is absent. A composed annotation
// (@GetMapping et al.) never declares a "method" element at all —
// Spring bakes the verb into the annotation type itself, mirroring
// @RequestMapping's own element shape rather than proxying through a
// meta-annotation at the bytecode level — so only @RequestMapping's own
// entry can ever be overridden by an explicit "method" element; nil
// here (for @RequestMapping with no "method" given either) means
// "every verb", never guessed as one specific verb.
var httpMappingVerbs = map[string][]string{
	"org.springframework.web.bind.annotation.RequestMapping": nil,
	"org.springframework.web.bind.annotation.GetMapping":     {"GET"},
	"org.springframework.web.bind.annotation.PostMapping":    {"POST"},
	"org.springframework.web.bind.annotation.PutMapping":     {"PUT"},
	"org.springframework.web.bind.annotation.DeleteMapping":  {"DELETE"},
	"org.springframework.web.bind.annotation.PatchMapping":   {"PATCH"},
}

// Route is one HTTP endpoint mapping found on an application method,
// with its real declared path(s) and verb(s) — not just the bare fact
// that *a* mapping annotation is present.
type Route struct {
	FQCN, Method, Descriptor string
	// Paths is every path this mapping resolves to, combining the
	// controller class's own @RequestMapping prefix (if any) with the
	// method's own — Spring concatenates them. Always at least one
	// entry ("/" when neither level gives a path, Spring's own
	// default).
	Paths []string
	// HTTPMethods is every verb this mapping declares; nil means
	// unspecified (matches every verb), never defaulted to "GET".
	HTTPMethods []string
}

// Routes finds every Spring MVC HTTP mapping on h's application
// classes.
func Routes(h *hierarchy.Hierarchy) []Route {
	var out []Route
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp {
			continue
		}
		classPaths, classVerbs := mappingShape(ci.AnnotationDetails)
		for _, m := range ci.Methods {
			for _, ann := range m.AnnotationDetails {
				verbsIfUnspecified, known := httpMappingVerbs[ann.Type]
				if !known {
					continue
				}
				methodPaths := stringArrayValue(firstOf(ann, "value", "path"))
				methodVerbs := enumArrayValue(annValue(ann, "method"))
				if methodVerbs == nil {
					methodVerbs = verbsIfUnspecified
				}
				out = append(out, Route{
					FQCN:        fqcn,
					Method:      m.Name,
					Descriptor:  m.Descriptor,
					Paths:       combinePaths(classPaths, methodPaths),
					HTTPMethods: combineVerbs(classVerbs, methodVerbs),
				})
			}
		}
	}
	return out
}

// mappingShape reads a class-level @RequestMapping's own path/method
// (a controller's shared path prefix) — classes are never annotated
// with a composed mapping (@GetMapping on a type makes no sense in
// Spring MVC and is not modeled).
func mappingShape(anns []classfile.Annotation) (paths, verbs []string) {
	for _, ann := range anns {
		if ann.Type != "org.springframework.web.bind.annotation.RequestMapping" {
			continue
		}
		return stringArrayValue(firstOf(ann, "value", "path")), enumArrayValue(annValue(ann, "method"))
	}
	return nil, nil
}

func combinePaths(classPaths, methodPaths []string) []string {
	if len(classPaths) == 0 {
		classPaths = []string{""}
	}
	if len(methodPaths) == 0 {
		methodPaths = []string{""}
	}
	seen := map[string]bool{}
	var out []string
	for _, c := range classPaths {
		for _, m := range methodPaths {
			p := joinPath(c, m)
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	return out
}

func joinPath(a, b string) string {
	a = strings.TrimSuffix(a, "/")
	if b != "" && !strings.HasPrefix(b, "/") {
		b = "/" + b
	}
	p := a + b
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func combineVerbs(classVerbs, methodVerbs []string) []string {
	if len(methodVerbs) > 0 {
		return methodVerbs
	}
	return classVerbs
}

// stereotypeAnnotations are Spring's own component-scan stereotypes,
// each taking an optional "value" element naming the bean (JVM
// spec-wise: a plain String, not an array, unlike the mapping
// annotations above).
var stereotypeAnnotations = map[string]bool{
	"org.springframework.stereotype.Component":               true,
	"org.springframework.stereotype.Service":                 true,
	"org.springframework.stereotype.Repository":              true,
	"org.springframework.stereotype.Controller":              true,
	"org.springframework.web.bind.annotation.RestController": true,
}

// BeanName is one @Component-family stereotype's own declared name —
// FRAMEWORKS F1's own qualifier/bean-identity value, previously
// invisible (only the bare stereotype presence was known).
type BeanName struct {
	FQCN string
	// Name is the explicit bean name given to the stereotype's own
	// "value" element, or "" when none was given (Spring itself then
	// derives a default from the class's simple name — not modeled
	// here, since that derivation needs no annotation value at all).
	Name    string
	Primary bool
}

// BeanNames finds every @Component-family stereotype on h's
// application classes, with its own declared name (if any) and
// @Primary status.
func BeanNames(h *hierarchy.Hierarchy) []BeanName {
	var out []BeanName
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp {
			continue
		}
		var name string
		found := false
		for _, ann := range ci.AnnotationDetails {
			if !stereotypeAnnotations[ann.Type] {
				continue
			}
			found = true
			if v, ok := ann.Value("value"); ok {
				if s, ok := v.Const.(string); ok {
					name = s
				}
			}
		}
		if !found {
			continue
		}
		out = append(out, BeanName{
			FQCN:    fqcn,
			Name:    name,
			Primary: hasAny(ci.Annotations, []string{"org.springframework.context.annotation.Primary"}),
		})
	}
	return out
}

// configurationFQCN and beanMethodFQCN are Spring's own factory-method
// shape: a `@Bean` method is only ever meaningful inside a
// `@Configuration` class (a plain `@Component`'s own methods are never
// treated as bean factories by Spring's container).
const (
	configurationFQCN = "org.springframework.context.annotation.Configuration"
	beanMethodFQCN    = "org.springframework.context.annotation.Bean"
)

// BeanMethods finds every `@Bean`-annotated method on h's
// `@Configuration`-annotated application classes — FRAMEWORKS F2's own
// "known factories" case. The bean's FQCN is the method's own return
// type (the object the factory method actually constructs), never the
// declaring @Configuration class itself; Name is the method's own
// `@Bean("...")`/`value()`/`name()` element when given, or (Spring's
// own real default) the method's own name otherwise.
//
// A method whose return type isn't an ordinary object type (void, or a
// primitive/array — never a valid bean type) contributes nothing,
// honestly, rather than guessing. Ambiguous or programmatic factories
// (a return type that is an interface with more than one app-class
// implementor, or any construction logic beyond the declared return
// type) are deliberately out of this slice's own scope — see
// FRAMEWORKS_IMPLEMENTATION_PLAN.MD's own F2 "niejednoznaczne... fabryki
// zachowują alternatywy/luki" requirement; BeanMethods never resolves
// those, it only reports the method's declared return type as-is,
// identically to how BeanNames already treats one stereotype as one
// bean without resolving interface/implementation ambiguity itself.
func BeanMethods(h *hierarchy.Hierarchy) []BeanName {
	var out []BeanName
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp {
			continue
		}
		if !hasAny(ci.Annotations, []string{configurationFQCN}) {
			continue
		}
		for _, m := range ci.Methods {
			ann, found := findAnnotation(m.AnnotationDetails, beanMethodFQCN)
			if !found {
				continue
			}
			retType, ok := methodReturnTypeFQCN(m.Descriptor)
			if !ok {
				continue
			}
			names := stringArrayValue(firstOf(ann, "value", "name"))
			name := m.Name
			if len(names) > 0 {
				name = names[0]
			}
			out = append(out, BeanName{
				FQCN:    retType,
				Name:    name,
				Primary: hasAnnotationType(m.AnnotationDetails, "org.springframework.context.annotation.Primary"),
			})
		}
	}
	return out
}

// findAnnotation returns the first of anns with the given type.
func findAnnotation(anns []classfile.Annotation, fqcn string) (classfile.Annotation, bool) {
	for _, a := range anns {
		if a.Type == fqcn {
			return a, true
		}
	}
	return classfile.Annotation{}, false
}

// methodReturnTypeFQCN reads a method descriptor's own return type as
// an FQCN — "" (ok=false) for void, a primitive, or an array, none of
// which Spring ever treats as a `@Bean` method's constructed object.
func methodReturnTypeFQCN(descriptor string) (string, bool) {
	i := strings.IndexByte(descriptor, ')')
	if i < 0 || i+1 >= len(descriptor) {
		return "", false
	}
	return classfile.FieldTypeFQCN(descriptor[i+1:])
}

// ProfileCondition is one `@Profile(...)` usage, with its declared
// profile name(s)/expression(s) and — only when the caller gave an
// explicit active-profile list — whether it is actually active.
// FRAMEWORKS F1's own explicit requirement: "brak pliku lub
// nieobsługiwany warunek daje unknown, nie false" — Active stays
// ConditionUnknown, never a guessed ConditionFalse, whenever no active
// profile list was given at all.
type ProfileCondition struct {
	FQCN, Method string // Method is "" for a class-level @Profile.
	Names        []string
	Active       Condition
}

// ProfileConditions finds every `@Profile(...)` usage on h's
// application classes/methods. activeProfiles nil means "not given by
// the caller" (every condition reports ConditionUnknown); a non-nil
// (possibly empty) slice means the caller knows exactly which profiles
// are active, so every condition resolves to ConditionTrue/False.
func ProfileConditions(h *hierarchy.Hierarchy, activeProfiles []string) []ProfileCondition {
	var active map[string]bool
	if activeProfiles != nil {
		active = make(map[string]bool, len(activeProfiles))
		for _, p := range activeProfiles {
			active[p] = true
		}
	}

	var out []ProfileCondition
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp {
			continue
		}
		if pc, ok := profileCondition(ci.AnnotationDetails, active, fqcn, ""); ok {
			out = append(out, pc)
		}
		for _, m := range ci.Methods {
			if pc, ok := profileCondition(m.AnnotationDetails, active, fqcn, m.Name); ok {
				out = append(out, pc)
			}
		}
	}
	return out
}

func profileCondition(anns []classfile.Annotation, active map[string]bool, fqcn, method string) (ProfileCondition, bool) {
	for _, ann := range anns {
		if ann.Type != "org.springframework.context.annotation.Profile" {
			continue
		}
		names := stringArrayValue(firstOf(ann, "value"))
		cond := ConditionUnknown
		if active != nil {
			cond = ConditionFalse
			for _, n := range names {
				neg := strings.HasPrefix(n, "!")
				bare := strings.TrimPrefix(n, "!")
				if neg && !active[bare] {
					cond = ConditionTrue
					break
				}
				if !neg && active[bare] {
					cond = ConditionTrue
					break
				}
			}
		}
		return ProfileCondition{FQCN: fqcn, Method: method, Names: names, Active: cond}, true
	}
	return ProfileCondition{}, false
}

// annValue is Annotation.Value without the ok bool, for a call site
// that only wants to chain straight into stringArrayValue/
// enumArrayValue, which already treat "absent" (the zero ElementValue)
// as "no value" on their own.
func annValue(a classfile.Annotation, name string) classfile.ElementValue {
	v, _ := a.Value(name)
	return v
}

// firstOf returns the first of names that a has an explicit value for
// — Spring's own "value"/"path" alias pair on the mapping annotations,
// where a usage site gives at most one of the two.
func firstOf(a classfile.Annotation, names ...string) classfile.ElementValue {
	for _, name := range names {
		if v, ok := a.Value(name); ok {
			return v
		}
	}
	return classfile.ElementValue{}
}

// stringArrayValue reads a String/String[] element value generically:
// javac always wraps a single value in a one-element array when the
// annotation element's own declared type is an array (JLS), so both
// shapes are handled, and the zero ElementValue (an absent element) is
// honestly nil, not guessed at.
func stringArrayValue(ev classfile.ElementValue) []string {
	switch ev.Kind {
	case classfile.ElementArray:
		out := make([]string, 0, len(ev.Array))
		for _, e := range ev.Array {
			if e.Kind == classfile.ElementConst {
				if s, ok := e.Const.(string); ok {
					out = append(out, s)
				}
			}
		}
		return out
	case classfile.ElementConst:
		if s, ok := ev.Const.(string); ok {
			return []string{s}
		}
	}
	return nil
}

// enumArrayValue reads an enum/enum[] element value's constant
// name(s) — Spring's own RequestMethod[] "method" element.
func enumArrayValue(ev classfile.ElementValue) []string {
	switch ev.Kind {
	case classfile.ElementArray:
		out := make([]string, 0, len(ev.Array))
		for _, e := range ev.Array {
			if e.Kind == classfile.ElementEnum {
				out = append(out, e.EnumConst)
			}
		}
		return out
	case classfile.ElementEnum:
		return []string{ev.EnumConst}
	}
	return nil
}
