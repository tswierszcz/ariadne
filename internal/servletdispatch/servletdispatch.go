// Package servletdispatch models the servlet container's dispatch into the
// servlets and filters an application registers (FRAMEWORKS F2/JVM_GRAPH,
// registered as a ROADMAP B3 model).
//
// The call graph never CHA-expands a receiver under javax./jakarta.
// (internal/callgraph.DefaultSkip): a container's
// `servlet.service(req, res)` (javax.servlet.Servlet#service) or
// `filter.doFilter(req, res, chain)` (javax.servlet.Filter#doFilter) keeps
// only its direct edge to the interface method, so an application's
// DispatcherServlet is never reached from the container. Lifting the skip
// for these interfaces would fan out to every Servlet/Filter class on the
// class path (dozens of framework servlets nobody registers); this model
// instead expands exactly those call signatures to the classes the
// application registers with the container:
//
//   - servlets/filters named in web.xml;
//   - application classes annotated @WebServlet/@WebFilter;
//   - in a Spring Boot application, Spring beans whose class is a
//     Servlet/Filter (a stereotype class or a @Bean method's declared
//     return type — Boot registers every Servlet/Filter bean);
//   - a fixed catalog of Spring Boot auto-configurations that register a
//     servlet or filter (DispatcherServlet, the character-encoding,
//     form-content and request-context filters, Spring Security's
//     DelegatingFilterProxy and its FilterChainProxy delegate), applied only
//     when the application is a Spring Boot application, the
//     auto-configuration class is on the class path and the application
//     does not exclude it.
//
// HttpServlet's own template dispatch (service(ServletRequest, ...) calling
// this.service(HttpServletRequest, ...) calling this.doGet/doPost/...) is a
// javax receiver too, so the same model expands those self-calls — only
// when the caller is HttpServlet itself, whose `this` is the dispatched
// servlet — to the registered HttpServlet subclasses' overrides.
//
// Every target is the method the JVM would select on the registered class
// (hierarchy.ResolveExact), so an inherited body (HttpServlet#service for
// DispatcherServlet) is the target, never a fabricated node. Edges carry
// BasisModel; the registrations carry their own basis, reported by the CLI.
package servletdispatch

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/models"
)

// ModelID is the registry identifier of this model.
const ModelID = "servlet-container:dispatch"

// Version versions the model's semantics (signatures, registration sources,
// Spring Boot catalog). Snapshots record it.
const Version = "s1"

// MaxChain is how many model iterations one dispatch can need: the
// container's Servlet#service edge, then HttpServlet#service(Http...), then
// doGet/doPost/... — each found only once the previous edge is in.
const MaxChain = 3

// Kind is what the container dispatches to.
type Kind string

const (
	KindServlet Kind = "servlet"
	KindFilter  Kind = "filter"
)

// Registration is one class the container dispatches requests to.
type Registration struct {
	Class string
	Kind  Kind
	// Basis says why the class is considered registered, e.g. "web.xml",
	// "@WebFilter", "Spring bean (stereotype)",
	// "Spring Boot DispatcherServletAutoConfiguration".
	Basis string
}

// Unresolved is a registration the model knows exists but cannot name the
// class of (a @Bean returning a ServletRegistrationBean/FilterRegistrationBean
// or a bare Servlet/Filter interface).
type Unresolved struct {
	Where, Reason string
}

// Bean is a Spring bean class with how it was declared.
type Bean struct {
	Class, Detail string
	// Method is the declaring "owner#name" of a @Bean method, "" for a
	// stereotype class.
	Method string
}

// Inputs are the registration facts gathered elsewhere.
type Inputs struct {
	WebXMLServlets, WebXMLFilters []string
	Beans                         []Bean
}

type api struct {
	servlet, filter, httpServlet, genericServlet string
	req, res, httpReq, httpRes, chain            string
}

var apis = []api{
	{"javax.servlet.Servlet", "javax.servlet.Filter", "javax.servlet.http.HttpServlet", "javax.servlet.GenericServlet",
		"Ljavax/servlet/ServletRequest;", "Ljavax/servlet/ServletResponse;", "Ljavax/servlet/http/HttpServletRequest;", "Ljavax/servlet/http/HttpServletResponse;", "Ljavax/servlet/FilterChain;"},
	{"jakarta.servlet.Servlet", "jakarta.servlet.Filter", "jakarta.servlet.http.HttpServlet", "jakarta.servlet.GenericServlet",
		"Ljakarta/servlet/ServletRequest;", "Ljakarta/servlet/ServletResponse;", "Ljakarta/servlet/http/HttpServletRequest;", "Ljakarta/servlet/http/HttpServletResponse;", "Ljakarta/servlet/FilterChain;"},
}

// httpTemplateMethods are HttpServlet's own self-dispatched methods taking
// (HttpServletRequest, HttpServletResponse) and returning void.
var httpTemplateMethods = []string{"service", "doGet", "doPost", "doPut", "doDelete", "doHead", "doOptions", "doTrace", "doPatch"}

type sigRole struct {
	kind Kind
	// selfOnly: expand only when the caller is the call's owner class
	// (HttpServlet's `this` dispatch).
	selfOnly bool
}

func signatures() map[models.Signature]sigRole {
	out := map[models.Signature]sigRole{}
	for _, a := range apis {
		service := "(" + a.req + a.res + ")V"
		out[models.Signature{Owner: a.servlet, Name: "service", Desc: service}] = sigRole{kind: KindServlet}
		out[models.Signature{Owner: a.genericServlet, Name: "service", Desc: service}] = sigRole{kind: KindServlet}
		out[models.Signature{Owner: a.httpServlet, Name: "service", Desc: service}] = sigRole{kind: KindServlet}
		out[models.Signature{Owner: a.filter, Name: "doFilter", Desc: "(" + a.req + a.res + a.chain + ")V"}] = sigRole{kind: KindFilter}
		for _, m := range httpTemplateMethods {
			out[models.Signature{Owner: a.httpServlet, Name: m, Desc: "(" + a.httpReq + a.httpRes + ")V"}] = sigRole{kind: KindServlet, selfOnly: true}
		}
		out[models.Signature{Owner: a.httpServlet, Name: "getLastModified", Desc: "(" + a.httpReq + ")J"}] = sigRole{kind: KindServlet, selfOnly: true}
	}
	return out
}

// Register adds the model to registry.
func Register(registry *models.Registry) {
	sigs := signatures()
	list := make([]models.Signature, 0, len(sigs))
	for s := range sigs {
		list = append(list, s)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Owner != list[j].Owner {
			return list[i].Owner < list[j].Owner
		}
		if list[i].Name != list[j].Name {
			return list[i].Name < list[j].Name
		}
		return list[i].Desc < list[j].Desc
	})
	registry.Register(&models.Model{
		ID:         ModelID,
		Source:     "Servlet specification (Servlet#service, Filter#doFilter, HttpServlet template methods) dispatched to the servlets/filters registered by web.xml, @WebServlet/@WebFilter, Spring beans and Spring Boot auto-configuration",
		Signatures: list,
	})
}

// bootCatalog: Spring Boot auto-configurations (2.x and 3.x share the class
// names) that register a servlet or filter with the embedded container
// without any application code naming it. Each entry was read from the
// auto-configuration's source: the bean is unconditional for a servlet web
// application, or conditional on a property whose default enables it
// (HttpEncoding: server.servlet.encoding.enabled, FormContentFilter:
// spring.mvc.formcontent.filter.enabled, both matchIfMissing=true), or on
// the absence of an application bean of the same type
// (RequestContextFilter). HiddenHttpMethodFilter is off by default
// (matchIfMissing=false) and not listed.
var bootCatalog = []struct {
	autoConfig, class string
	kind              Kind
	note              string
	// requires are the auto-configuration's @ConditionalOnClass classes
	// (beyond the registered class itself).
	requires []string
}{
	{"org.springframework.boot.autoconfigure.web.servlet.DispatcherServletAutoConfiguration", "org.springframework.web.servlet.DispatcherServlet", KindServlet, "", nil},
	{"org.springframework.boot.autoconfigure.web.servlet.HttpEncodingAutoConfiguration", "org.springframework.boot.web.servlet.filter.OrderedCharacterEncodingFilter", KindFilter, "", nil},
	{"org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration", "org.springframework.boot.web.servlet.filter.OrderedFormContentFilter", KindFilter, "",
		[]string{"org.springframework.web.servlet.DispatcherServlet", "org.springframework.web.servlet.config.annotation.WebMvcConfigurer"}},
	{"org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration", "org.springframework.boot.web.servlet.filter.OrderedRequestContextFilter", KindFilter, "",
		[]string{"org.springframework.web.servlet.DispatcherServlet", "org.springframework.web.servlet.config.annotation.WebMvcConfigurer"}},
	{"org.springframework.boot.autoconfigure.security.servlet.SecurityFilterAutoConfiguration", "org.springframework.web.filter.DelegatingFilterProxy", KindFilter, "",
		[]string{"org.springframework.security.web.context.AbstractSecurityWebApplicationInitializer", "org.springframework.security.config.http.SessionCreationPolicy", "org.springframework.security.web.FilterChainProxy"}},
	{"org.springframework.boot.autoconfigure.security.servlet.SecurityFilterAutoConfiguration", "org.springframework.security.web.FilterChainProxy", KindFilter, " (springSecurityFilterChain, DelegatingFilterProxy's delegate)",
		[]string{"org.springframework.security.web.context.AbstractSecurityWebApplicationInitializer", "org.springframework.security.config.http.SessionCreationPolicy"}},
}

var bootAnnotations = []string{
	"org.springframework.boot.autoconfigure.SpringBootApplication",
	"org.springframework.boot.autoconfigure.EnableAutoConfiguration",
}

var registrationBeanTypes = map[string]bool{
	"org.springframework.boot.web.servlet.ServletRegistrationBean":                         true,
	"org.springframework.boot.web.servlet.FilterRegistrationBean":                          true,
	"org.springframework.boot.web.servlet.DelegatingFilterProxyRegistrationBean":           true,
	"org.springframework.boot.web.servlet.RegistrationBean":                                true,
	"org.springframework.boot.web.servlet.AbstractFilterRegistrationBean":                  true,
	"org.springframework.boot.web.servlet.DynamicRegistrationBean":                         true,
	"org.springframework.boot.web.servlet.ServletContextInitializer":                       true,
	"org.springframework.boot.autoconfigure.web.servlet.DispatcherServletRegistrationBean": true,
}

// Find collects the registered servlets and filters present in h, sorted
// by class then kind, and the registrations it cannot name.
func Find(h *hierarchy.Hierarchy, in Inputs) ([]Registration, []Unresolved) {
	isServlet := subtypeSet(h, func(a api) string { return a.servlet })
	isFilter := subtypeSet(h, func(a api) string { return a.filter })
	seen := map[string]bool{}
	var out []Registration
	add := func(class string, kind Kind, basis string) {
		ci, ok := h.Lookup(class)
		if !ok || ci.IsInterface() {
			return
		}
		if (kind == KindServlet && !isServlet[class]) || (kind == KindFilter && !isFilter[class]) {
			return
		}
		key := class + "|" + string(kind)
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, Registration{Class: class, Kind: kind, Basis: basis})
	}
	for _, c := range in.WebXMLServlets {
		add(c, KindServlet, "web.xml")
	}
	for _, c := range in.WebXMLFilters {
		add(c, KindFilter, "web.xml")
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		for _, a := range ci.Annotations {
			switch a {
			case "javax.servlet.annotation.WebServlet", "jakarta.servlet.annotation.WebServlet":
				add(fqcn, KindServlet, "@WebServlet")
			case "javax.servlet.annotation.WebFilter", "jakarta.servlet.annotation.WebFilter":
				add(fqcn, KindFilter, "@WebFilter")
			}
		}
	}
	// Spring Boot registers every Servlet/Filter bean with the embedded
	// container (ServletContextInitializerBeans); plain Spring MVC does not
	// (a filter bean needs a DelegatingFilterProxy), so beans count only
	// in a Boot application.
	excluded, boot := bootApplication(h)
	var unresolved []Unresolved
	for _, b := range in.Beans {
		if !boot {
			break
		}
		basis := "Spring bean (" + b.Detail + ")"
		if isServlet[b.Class] {
			add(b.Class, KindServlet, basis)
		}
		if isFilter[b.Class] {
			add(b.Class, KindFilter, basis)
		}
		where := b.Class
		if b.Method != "" {
			where = b.Method
		}
		switch {
		case registrationBeanTypes[b.Class]:
			unresolved = append(unresolved, Unresolved{Where: where, Reason: "returns " + b.Class + ": the registered servlet/filter class is chosen at run time"})
		case isAPIType(b.Class):
			unresolved = append(unresolved, Unresolved{Where: where, Reason: "returns the bare " + b.Class + " interface"})
		}
	}
	if boot {
		for _, e := range bootCatalog {
			if excluded[e.autoConfig] {
				continue
			}
			if !allPresent(h, append([]string{e.autoConfig}, e.requires...)) {
				continue
			}
			add(e.class, e.kind, "Spring Boot "+e.autoConfig[strings.LastIndexByte(e.autoConfig, '.')+1:]+e.note)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Class != out[j].Class {
			return out[i].Class < out[j].Class
		}
		return out[i].Kind < out[j].Kind
	})
	sort.Slice(unresolved, func(i, j int) bool { return unresolved[i].Where < unresolved[j].Where })
	return out, unresolved
}

func allPresent(h *hierarchy.Hierarchy, classes []string) bool {
	for _, c := range classes {
		if _, ok := h.Lookup(c); !ok {
			return false
		}
	}
	return true
}

func isAPIType(class string) bool {
	for _, a := range apis {
		if class == a.servlet || class == a.filter {
			return true
		}
	}
	return false
}

func subtypeSet(h *hierarchy.Hierarchy, pick func(api) string) map[string]bool {
	out := map[string]bool{}
	for _, a := range apis {
		for _, s := range h.Subtypes(pick(a)) {
			out[s] = true
		}
	}
	return out
}

// bootApplication reports whether an application class enables Spring
// Boot auto-configuration, and which auto-configurations it excludes
// (exclude/excludeName).
func bootApplication(h *hierarchy.Hierarchy) (map[string]bool, bool) {
	found := false
	excluded := map[string]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		for _, ann := range ci.AnnotationDetails {
			if ann.Type != bootAnnotations[0] && ann.Type != bootAnnotations[1] {
				continue
			}
			found = true
			for _, name := range []string{"exclude", "excludeName"} {
				v, ok := ann.Value(name)
				if !ok {
					continue
				}
				vals := []classfile.ElementValue{v}
				if v.Kind == classfile.ElementArray {
					vals = v.Array
				}
				for _, e := range vals {
					switch e.Kind {
					case classfile.ElementClass:
						excluded[e.ClassFQCN] = true
					case classfile.ElementConst:
						if s, ok := e.Const.(string); ok {
							excluded[s] = true
						}
					}
				}
			}
		}
	}
	return excluded, found
}

// Evaluator answers the model at a call site.
type Evaluator struct {
	h          *hierarchy.Hierarchy
	regs       []Registration
	unresolved []Unresolved
	sigs       map[models.Signature]sigRole
}

// NewEvaluator returns an evaluator over the given registrations.
func NewEvaluator(h *hierarchy.Hierarchy, regs []Registration, unresolved []Unresolved) *Evaluator {
	return &Evaluator{h: h, regs: regs, unresolved: unresolved, sigs: signatures()}
}

// Evaluate implements models.Evaluator.
func (e *Evaluator) Evaluate(m *models.Model, site models.CallSite) models.Fact {
	fact := models.Fact{ModelID: m.ID}
	role, ok := e.sigs[site.Call]
	if !ok {
		return fact
	}
	kind, ok := e.callKind(site)
	if !ok || (kind != classfile.CallVirtual && kind != classfile.CallInterface) {
		return fact // invokespecial (super.service) is already a direct edge
	}
	if role.selfOnly && site.Caller.Owner != site.Call.Owner {
		return fact
	}
	subtypes := map[string]bool{}
	for _, s := range e.h.Subtypes(site.Call.Owner) {
		subtypes[s] = true
	}
	seen := map[string]bool{}
	for _, r := range e.regs {
		if r.Kind != role.kind || !subtypes[r.Class] {
			continue
		}
		decl, ok := e.h.ResolveExact(r.Class, site.Call.Name, site.Call.Desc)
		if !ok || decl == site.Call.Owner || seen[decl] {
			continue
		}
		ci, _ := e.h.Lookup(decl)
		if meth, ok := ci.MethodByNameAndDesc(site.Call.Name, site.Call.Desc); !ok || meth.IsAbstract() {
			continue
		}
		seen[decl] = true
		fact.Targets = append(fact.Targets, models.Target{
			Method: models.Signature{Owner: decl, Name: site.Call.Name, Desc: site.Call.Desc},
			Kind:   kind,
		})
	}
	sort.Slice(fact.Targets, func(i, j int) bool { return fact.Targets[i].Method.Owner < fact.Targets[j].Method.Owner })
	if !role.selfOnly && len(e.unresolved) > 0 {
		fact.Incomplete = true
		fact.Reason = fmt.Sprintf("%d servlet/filter registration(s) whose class is not known statically (first: %s %s)",
			len(e.unresolved), e.unresolved[0].Where, e.unresolved[0].Reason)
	}
	return fact
}

func (e *Evaluator) callKind(site models.CallSite) (classfile.CallKind, bool) {
	ci, ok := e.h.Lookup(site.Caller.Owner)
	if !ok {
		return "", false
	}
	m, ok := ci.MethodByNameAndDesc(site.Caller.Name, site.Caller.Desc)
	if !ok {
		return "", false
	}
	for _, c := range m.Calls {
		if c.PC == site.PC && c.Owner == site.Call.Owner && c.Name == site.Call.Name && c.Desc == site.Call.Desc {
			return c.Kind, true
		}
	}
	return "", false
}
