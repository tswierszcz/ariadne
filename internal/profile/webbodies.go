package profile

import (
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// WebBodyKind says why WebBodyRoots made a method a root.
type WebBodyKind string

const (
	WebBodyGetter      WebBodyKind = "response body getter"
	WebBodyConstructor WebBodyKind = "request body constructor"
	WebBodySetter      WebBodyKind = "request body setter"
)

// WebBodyRoot is one method Jackson calls reflectively for a Spring MVC
// handler's JSON body.
type WebBodyRoot struct {
	Handler, FQCN, Name, Desc string
	Kind                      WebBodyKind
}

const springWeb = "org.springframework.web.bind.annotation."

// webHandlerAnnotations mark the methods Spring MVC invokes for a request
// whose return value its message converters may write.
var webHandlerAnnotations = []string{
	springWeb + "RequestMapping", springWeb + "GetMapping", springWeb + "PostMapping",
	springWeb + "PutMapping", springWeb + "DeleteMapping", springWeb + "PatchMapping",
	springWeb + "ExceptionHandler",
}

// responseBodyClassAnnotations make every handler of the class write its
// return value as the body (@RestController and @RestControllerAdvice carry
// @ResponseBody as a meta-annotation).
var responseBodyClassAnnotations = []string{
	springWeb + "ResponseBody", springWeb + "RestController", springWeb + "RestControllerAdvice",
}

// WebBodyRoots is FRAMEWORKS F5 f5-2's rule for the JSON bodies Spring MVC
// hands to Jackson, which calls the application's types reflectively — a
// real-JVM oracle (testdata/mvc) measured them unreachable:
//
//   - a handler method of an application class (a mapping or
//     @ExceptionHandler) that writes its return value as the body (the
//     method or its class carries @ResponseBody, @RestController or
//     @RestControllerAdvice): Jackson serializes the declared return type
//     through its getters — every public non-static no-argument get*/is*
//     method, in the class and its application superclasses, then the same
//     for each application type a getter returns;
//   - a handler parameter carrying @RequestBody: Jackson creates the
//     declared parameter type through its no-argument constructor (every
//     constructor when it has none: a @JsonCreator or parameter names) and
//     fills it through its setters — every public non-static one-argument
//     set* method — then the same for each application type a setter
//     takes.
//
// A declared type that is abstract or an interface includes its application
// subtypes (Jackson writes the runtime type; polymorphic reading names
// one). Only application types count; arrays count as their element type.
// Not modeled: generic element types (List<Item> erases to List), field
// access, @JsonProperty on private members, builders and mix-ins.
func WebBodyRoots(h *hierarchy.Hierarchy) []WebBodyRoot {
	var out []WebBodyRoot
	seen := map[string]bool{}
	add := func(r WebBodyRoot) {
		k := string(r.Kind) + "\t" + r.FQCN + "#" + r.Name + r.Desc
		if !seen[k] {
			seen[k] = true
			out = append(out, r)
		}
	}
	written, read := map[string]bool{}, map[string]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		classBody := hasAny(ci.Annotations, responseBodyClassAnnotations)
		for _, m := range ci.Methods {
			if m.IsStatic() || !hasAny(m.Annotations, webHandlerAnnotations) {
				continue
			}
			handler := fqcn + "#" + m.Name + m.Descriptor
			if classBody || hasAny(m.Annotations, []string{springWeb + "ResponseBody"}) {
				if t := descClass(returnDesc(m.Descriptor)); t != "" {
					writeType(h, t, handler, written, add)
				}
			}
			params, err := classfile.ParamDescriptors(m.Descriptor)
			if err != nil {
				continue
			}
			for i, p := range params {
				if i < len(m.ParameterAnnotations) && hasAny(m.ParameterAnnotations[i], []string{springWeb + "RequestBody"}) {
					if t := descClass(p); t != "" {
						readType(h, t, handler, read, add)
					}
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.FQCN != b.FQCN {
			return a.FQCN < b.FQCN
		}
		if a.Name+a.Desc != b.Name+b.Desc {
			return a.Name+a.Desc < b.Name+b.Desc
		}
		return a.Kind < b.Kind
	})
	return out
}

// bodyTypes is fqcn and, when it is abstract or an interface, its
// application subtypes — the application classes among them.
func bodyTypes(h *hierarchy.Hierarchy, fqcn string) []*hierarchy.ClassInfo {
	var out []*hierarchy.ClassInfo
	ci, ok := h.Lookup(fqcn)
	if ok && ci.IsApp {
		out = append(out, ci)
	}
	if !ok || ci.IsInterface() || ci.AccessFlags&classfile.AccAbstract != 0 {
		for _, s := range h.Subtypes(fqcn) {
			if sc, ok := h.Lookup(s); ok && sc.IsApp {
				out = append(out, sc)
			}
		}
	}
	return out
}

func writeType(h *hierarchy.Hierarchy, fqcn, handler string, seen map[string]bool, add func(WebBodyRoot)) {
	if seen[fqcn] {
		return
	}
	seen[fqcn] = true
	for _, ci := range bodyTypes(h, fqcn) {
		for cur := ci; cur != nil && cur.IsApp; cur = superApp(h, cur) {
			for _, m := range cur.Methods {
				if !isGetter(m) {
					continue
				}
				add(WebBodyRoot{Handler: handler, FQCN: cur.FQCN, Name: m.Name, Desc: m.Descriptor, Kind: WebBodyGetter})
				if t := descClass(returnDesc(m.Descriptor)); t != "" {
					writeType(h, t, handler, seen, add)
				}
			}
		}
	}
}

func readType(h *hierarchy.Hierarchy, fqcn, handler string, seen map[string]bool, add func(WebBodyRoot)) {
	if seen[fqcn] {
		return
	}
	seen[fqcn] = true
	for _, ci := range bodyTypes(h, fqcn) {
		if ci.IsInterface() || ci.AccessFlags&classfile.AccAbstract != 0 {
			continue
		}
		ctors := ci.MethodsByName("<init>")
		var noArg []hierarchy.Method
		for _, c := range ctors {
			if c.Descriptor == "()V" {
				noArg = append(noArg, c)
			}
		}
		if len(noArg) > 0 {
			ctors = noArg
		}
		for _, c := range ctors {
			add(WebBodyRoot{Handler: handler, FQCN: ci.FQCN, Name: c.Name, Desc: c.Descriptor, Kind: WebBodyConstructor})
		}
		for cur := ci; cur != nil && cur.IsApp; cur = superApp(h, cur) {
			for _, m := range cur.Methods {
				if !isSetter(m) {
					continue
				}
				add(WebBodyRoot{Handler: handler, FQCN: cur.FQCN, Name: m.Name, Desc: m.Descriptor, Kind: WebBodySetter})
				if ps, err := classfile.ParamDescriptors(m.Descriptor); err == nil && len(ps) == 1 {
					if t := descClass(ps[0]); t != "" {
						readType(h, t, handler, seen, add)
					}
				}
			}
		}
	}
}

func superApp(h *hierarchy.Hierarchy, ci *hierarchy.ClassInfo) *hierarchy.ClassInfo {
	sc, ok := h.Lookup(ci.SuperClass)
	if !ok {
		return nil
	}
	return sc
}

func isGetter(m hierarchy.Method) bool {
	if !m.IsPublic() || m.IsStatic() || !strings.HasPrefix(m.Descriptor, "()") || strings.HasSuffix(m.Descriptor, ")V") {
		return false
	}
	if m.Name == "getClass" {
		return false
	}
	return beanName(m.Name, "get") || beanName(m.Name, "is") && (m.Descriptor == "()Z" || m.Descriptor == "()Ljava/lang/Boolean;")
}

func isSetter(m hierarchy.Method) bool {
	if !m.IsPublic() || m.IsStatic() || !beanName(m.Name, "set") {
		return false
	}
	ps, err := classfile.ParamDescriptors(m.Descriptor)
	return err == nil && len(ps) == 1
}

func beanName(name, prefix string) bool {
	return len(name) > len(prefix) && strings.HasPrefix(name, prefix) && name[len(prefix)] >= 'A' && name[len(prefix)] <= 'Z'
}

func returnDesc(desc string) string {
	if i := strings.LastIndexByte(desc, ')'); i >= 0 {
		return desc[i+1:]
	}
	return ""
}

// descClass is the class a field descriptor names (an array's element
// class), or "" for a primitive.
func descClass(d string) string {
	d = strings.TrimLeft(d, "[")
	if !strings.HasPrefix(d, "L") || !strings.HasSuffix(d, ";") {
		return ""
	}
	return strings.ReplaceAll(d[1:len(d)-1], "/", ".")
}

// WebBodyAppRoots is WebBodyRoots' call-graph-node form.
func WebBodyAppRoots(g *callgraph.Graph, roots []WebBodyRoot) []uint32 {
	var out []uint32
	for _, r := range roots {
		if id, ok := g.ID(r.FQCN, r.Name, r.Desc); ok {
			out = append(out, id)
		}
	}
	return out
}
