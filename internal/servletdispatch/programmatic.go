package servletdispatch

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// Programmatic registration (s3): a servlet container with no web.xml
// entry for a servlet or filter learns about it from code run at start-up
// — a ServletContainerInitializer (declared in
// META-INF/services/javax.servlet.ServletContainerInitializer, found by the
// container's ServiceLoader) receives the classes matching its
// @HandlesTypes types, and either it or those classes call
// ServletContext#addServlet/addFilter. Spring MVC without Boot works this
// way: SpringServletContainerInitializer (@HandlesTypes
// WebApplicationInitializer) creates every concrete WebApplicationInitializer
// reflectively, and AbstractDispatcherServletInitializer#onStartup
// registers the DispatcherServlet that createDispatcherServlet returns.
//
// The model takes as start-up classes the initializers themselves and the
// concrete classes of their @HandlesTypes types (the container hands them
// over; Spring creates exactly the concrete ones). A registration counts
// only in a method declared by a start-up class or one of its superclasses
// — code a start-up object runs — so the same call in a class nothing
// creates at start-up (an abstract initializer no class extends, a helper
// of a framework the application does not use) registers nothing. The
// registered class is the argument's value (DATA_FLOW): a `new`, a class
// literal, a class-name string, followed through the callers of a
// parameter; a factory call's or field's declared type stands for its
// concrete subtypes, unless it is the bare Servlet/Filter/Class/Object
// type — that registration is unresolved (counted, never "every servlet").

// StartupClass is a class the container creates or hands to an
// initializer at start-up.
type StartupClass struct {
	Class, Detail string
}

// maxArgDepth bounds how far a registered value is followed (parameters
// through callers, each level counted once).
const maxArgDepth = 6

var sciServices = map[string]string{
	"javax.servlet.ServletContainerInitializer":   "javax.servlet.annotation.HandlesTypes",
	"jakarta.servlet.ServletContainerInitializer": "jakarta.servlet.annotation.HandlesTypes",
}

// IsInitializerService reports whether service is the
// ServletContainerInitializer service (javax or jakarta).
func IsInitializerService(service string) bool {
	_, ok := sciServices[service]
	return ok
}

// StartupClasses returns the initializers (provider classes of the
// ServletContainerInitializer service, as given) and the concrete classes of
// their @HandlesTypes types present in h, sorted by class.
func StartupClasses(h *hierarchy.Hierarchy, initializers []string) []StartupClass {
	seen := map[string]bool{}
	var out []StartupClass
	add := func(class, detail string) {
		if seen[class] || !concrete(h, class) {
			return
		}
		seen[class] = true
		out = append(out, StartupClass{Class: class, Detail: detail})
	}
	for _, sci := range initializers {
		ci, ok := h.Lookup(sci)
		if !ok {
			continue
		}
		add(sci, "ServletContainerInitializer")
		for _, ann := range ci.AnnotationDetails {
			if ann.Type != sciServices["javax.servlet.ServletContainerInitializer"] && ann.Type != sciServices["jakarta.servlet.ServletContainerInitializer"] {
				continue
			}
			v, ok := ann.Value("value")
			if !ok {
				continue
			}
			vals := []classfile.ElementValue{v}
			if v.Kind == classfile.ElementArray {
				vals = v.Array
			}
			for _, e := range vals {
				if e.Kind != classfile.ElementClass {
					continue
				}
				for _, s := range append([]string{e.ClassFQCN}, h.Subtypes(e.ClassFQCN)...) {
					add(s, "@HandlesTypes("+e.ClassFQCN+") of "+sci)
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Class < out[j].Class })
	return out
}

func concrete(h *hierarchy.Hierarchy, class string) bool {
	ci, ok := h.Lookup(class)
	return ok && ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) == 0
}

type registrationAPI struct {
	kind Kind
}

// registrationCalls are ServletContext's addServlet/addFilter overloads
// (name, then a class name, a Class or an instance).
var registrationCalls = func() map[string]registrationAPI {
	out := map[string]registrationAPI{}
	for _, p := range []string{"javax", "jakarta"} {
		ctx := p + ".servlet.ServletContext"
		for _, arg := range []string{"Ljava/lang/String;", "Ljava/lang/Class;", "L" + p + "/servlet/Servlet;"} {
			out[ctx+"#addServlet(Ljava/lang/String;"+arg+")L"+p+"/servlet/ServletRegistration$Dynamic;"] = registrationAPI{KindServlet}
		}
		for _, arg := range []string{"Ljava/lang/String;", "Ljava/lang/Class;", "L" + p + "/servlet/Filter;"} {
			out[ctx+"#addFilter(Ljava/lang/String;"+arg+")L"+p+"/servlet/FilterRegistration$Dynamic;"] = registrationAPI{KindFilter}
		}
	}
	return out
}()

// FindProgrammatic returns the servlets and filters the start-up classes
// register with ServletContext#addServlet/addFilter, and the
// registrations whose class it cannot name. Not de-duplicated against
// other sources (Find does that).
func FindProgrammatic(h *hierarchy.Hierarchy, idx *dataflow.CallerIndex, startup []StartupClass) ([]Registration, []Unresolved) {
	runners := map[string]bool{}
	for _, s := range startup {
		for c := s.Class; c != ""; {
			if runners[c] {
				break
			}
			runners[c] = true
			ci, ok := h.Lookup(c)
			if !ok {
				break
			}
			c = ci.SuperClass
		}
	}
	classes := make([]string, 0, len(runners))
	for c := range runners {
		classes = append(classes, c)
	}
	sort.Strings(classes)
	var regs []Registration
	var unresolved []Unresolved
	for _, c := range classes {
		ci, ok := h.Lookup(c)
		if !ok {
			continue
		}
		for mi := range ci.Methods {
			m := &ci.Methods[mi]
			for _, call := range m.Calls {
				api, ok := registrationCalls[call.Owner+"#"+call.Name+call.Desc]
				if !ok {
					continue
				}
				method := dataflow.MethodID{Owner: c, Name: m.Name, Desc: m.Descriptor}
				where := fmt.Sprintf("%s#%s%s pc=%d", c, m.Name, m.Descriptor, call.PC)
				facts, reason := idx.Facts(method)
				if reason != "" {
					unresolved = append(unresolved, Unresolved{Where: where, Reason: reason})
					continue
				}
				site := bytecodeir.Site{Method: c + "#" + m.Name, Descriptor: m.Descriptor, PC: call.PC, Op: bytecodeir.OpCall}
				r := &argResolver{h: h, idx: idx, seen: map[string]bool{}}
				r.value(facts.ArgumentAt(site, 1), method, 0)
				for _, class := range r.classes {
					regs = append(regs, Registration{Class: class, Kind: api.kind, Basis: "ServletContext#" + call.Name + " in " + c + "#" + m.Name + " (" + r.how[class] + ")"})
				}
				for _, u := range r.unknown {
					unresolved = append(unresolved, Unresolved{Where: where, Reason: u})
				}
			}
		}
	}
	return regs, unresolved
}

type argResolver struct {
	h       *hierarchy.Hierarchy
	idx     *dataflow.CallerIndex
	seen    map[string]bool
	classes []string
	how     map[string]string
	unknown []string
}

func (r *argResolver) add(class, how string) {
	if r.how == nil {
		r.how = map[string]string{}
	}
	if _, ok := r.how[class]; ok {
		return
	}
	r.how[class] = how
	r.classes = append(r.classes, class)
}

// declared adds the concrete subtypes of a declared type (itself
// included), or an unresolved reason for a type that bounds nothing.
func (r *argResolver) declared(t, how string) {
	switch t {
	case "", "java.lang.Object", "java.lang.Class", "java.lang.String":
		r.unknown = append(r.unknown, how+" of type "+t+": the registered class is chosen at run time")
		return
	}
	if isAPIType(t) {
		r.unknown = append(r.unknown, how+" of the bare "+t+" type")
		return
	}
	for _, s := range append([]string{t}, r.h.Subtypes(t)...) {
		if concrete(r.h, s) {
			r.add(s, how+" declared "+t)
		}
	}
}

func (r *argResolver) value(v valueflow.Value, ctx dataflow.MethodID, depth int) {
	switch v.Kind {
	case valueflow.KindConst:
		if s, ok := v.Const.(string); ok && (v.Type == "java.lang.String" || v.Type == "java.lang.Class") {
			name := strings.ReplaceAll(s, "/", ".")
			if _, ok := r.h.Lookup(name); ok {
				what := "class literal"
				if v.Type == "java.lang.String" {
					what = "class name"
				}
				r.add(name, what)
			} else {
				r.unknown = append(r.unknown, "names "+name+", which is not in the analyzed input")
			}
			return
		}
		r.unknown = append(r.unknown, fmt.Sprintf("a %s constant", v.Type))
	case valueflow.KindAlloc:
		r.add(v.Type, "new")
	case valueflow.KindSet:
		for _, c := range v.Candidates {
			r.value(c, ctx, depth)
		}
	case valueflow.KindNull:
	case valueflow.KindTyped:
		if v.Parameter == nil {
			r.declared(v.Type, "a value")
			return
		}
		key := fmt.Sprintf("%s/%d", ctx, v.Parameter.Index)
		if r.seen[key] {
			return
		}
		r.seen[key] = true
		if depth >= maxArgDepth {
			r.unknown = append(r.unknown, fmt.Sprintf("parameter chain longer than %d methods", maxArgDepth))
			return
		}
		pv := r.idx.ResolveParameter(ctx, v.Parameter.Index)
		r.unknown = append(r.unknown, pv.Unknown...)
		if len(pv.Values) == 0 && len(pv.Unknown) == 0 {
			r.unknown = append(r.unknown, fmt.Sprintf("no caller passes parameter %d of %s", v.Parameter.Index, ctx))
		}
		for i, c := range pv.Values {
			r.value(c, pv.In[i], depth+1)
		}
	case valueflow.KindCallTrace:
		r.declared(descType(returnDesc(v.CallTrace.Desc)), "the result of "+v.CallTrace.Owner+"#"+v.CallTrace.Name)
	case valueflow.KindCallResult:
		r.declared(descType(returnDesc(v.Call.Desc)), "the result of "+v.Call.Owner+"#"+v.Call.Name)
	case valueflow.KindField:
		r.declared(descType(v.Field.Desc), "field "+v.Field.Owner+"."+v.Field.Name)
	case valueflow.KindUnknown:
		if v.Type != "" {
			// a virtual call's result: its declared return type
			r.declared(v.Type, "a call result")
			return
		}
		r.unknown = append(r.unknown, v.Reason)
	default:
		r.unknown = append(r.unknown, fmt.Sprintf("a %s value", v.Kind))
	}
}

func returnDesc(desc string) string {
	if i := strings.LastIndexByte(desc, ')'); i >= 0 {
		return desc[i+1:]
	}
	return ""
}

func descType(desc string) string {
	if len(desc) > 2 && desc[0] == 'L' && desc[len(desc)-1] == ';' {
		return strings.ReplaceAll(desc[1:len(desc)-1], "/", ".")
	}
	return ""
}
