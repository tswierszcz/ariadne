package servletdispatch

import (
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/models"
)

const (
	jxReq   = "Ljakarta/servlet/ServletRequest;"
	jxRes   = "Ljakarta/servlet/ServletResponse;"
	jxHReq  = "Ljakarta/servlet/http/HttpServletRequest;"
	jxHRes  = "Ljakarta/servlet/http/HttpServletResponse;"
	jxChain = "Ljakarta/servlet/FilterChain;"
)

var (
	serviceDesc  = "(" + jxReq + jxRes + ")V"
	httpDesc     = "(" + jxHReq + jxHRes + ")V"
	doFilterDesc = "(" + jxReq + jxRes + jxChain + ")V"
)

func method(name, desc string, abstract bool, calls ...classfile.Call) hierarchy.Method {
	m := hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc}, Calls: calls}
	if abstract {
		m.AccessFlags = classfile.AccAbstract
	}
	return m
}

func iface(fqcn string, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: "java.lang.Object", AccessFlags: classfile.AccInterface | classfile.AccAbstract, Methods: methods}
}

func class(fqcn, super string, ifaces []string, app bool, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: super, Interfaces: ifaces, Methods: methods, IsApp: app}
}

// fixture: the jakarta servlet API (Servlet, Filter, GenericServlet,
// HttpServlet with its template self-calls), a container calling
// Servlet#service and Filter#doFilter through the interfaces, a registered
// HttpServlet subclass overriding doGet and calling super.doPost, a
// registered filter, an unregistered servlet and filter.
func fixture(extra ...*hierarchy.ClassInfo) *hierarchy.Hierarchy {
	classes := []*hierarchy.ClassInfo{
		iface("jakarta.servlet.Servlet", method("service", serviceDesc, true)),
		iface("jakarta.servlet.Filter", method("doFilter", doFilterDesc, true)),
		class("jakarta.servlet.GenericServlet", "java.lang.Object", []string{"jakarta.servlet.Servlet"}, false,
			method("service", serviceDesc, true)),
		class("jakarta.servlet.http.HttpServlet", "jakarta.servlet.GenericServlet", nil, false,
			method("service", serviceDesc, false, classfile.Call{Kind: classfile.CallVirtual, Owner: "jakarta.servlet.http.HttpServlet", Name: "service", Desc: httpDesc, PC: 10}),
			method("service", httpDesc, false,
				classfile.Call{Kind: classfile.CallVirtual, Owner: "jakarta.servlet.http.HttpServlet", Name: "doGet", Desc: httpDesc, PC: 20},
				classfile.Call{Kind: classfile.CallVirtual, Owner: "jakarta.servlet.http.HttpServlet", Name: "doPost", Desc: httpDesc, PC: 30}),
			method("doGet", httpDesc, false),
			method("doPost", httpDesc, false)),
		class("org.container.Chain", "java.lang.Object", nil, false,
			method("internalDoFilter", "()V", false,
				classfile.Call{Kind: classfile.CallInterface, Owner: "jakarta.servlet.Filter", Name: "doFilter", Desc: doFilterDesc, PC: 5},
				classfile.Call{Kind: classfile.CallInterface, Owner: "jakarta.servlet.Servlet", Name: "service", Desc: serviceDesc, PC: 9})),
		class("app.Hello", "jakarta.servlet.http.HttpServlet", nil, true,
			method("doGet", httpDesc, false),
			method("doPost", httpDesc, false, classfile.Call{Kind: classfile.CallSpecial, Owner: "jakarta.servlet.http.HttpServlet", Name: "doPost", Desc: httpDesc, PC: 3})),
		class("app.Audit", "java.lang.Object", []string{"jakarta.servlet.Filter"}, true, method("doFilter", doFilterDesc, false)),
		class("app.Unused", "jakarta.servlet.http.HttpServlet", nil, true, method("doGet", httpDesc, false)),
		class("app.UnusedFilter", "java.lang.Object", []string{"jakarta.servlet.Filter"}, true, method("doFilter", doFilterDesc, false)),
	}
	return hierarchy.NewForTesting(append(classes, extra...))
}

func site(caller, callerName, callerDesc, owner, name, desc string, pc int) models.CallSite {
	return models.CallSite{
		Caller: models.Signature{Owner: caller, Name: callerName, Desc: callerDesc},
		Call:   models.Signature{Owner: owner, Name: name, Desc: desc},
		PC:     pc,
	}
}

func targets(f models.Fact) []string {
	var out []string
	for _, t := range f.Targets {
		out = append(out, t.Method.Owner+"#"+t.Method.Name+"/"+string(t.Kind))
	}
	return out
}

func evaluate(t *testing.T, h *hierarchy.Hierarchy, in Inputs, s models.CallSite) models.Fact {
	t.Helper()
	regs, unresolved := Find(h, in)
	reg := models.NewRegistry()
	Register(reg)
	ms := reg.Lookup(s.Call, nil)
	if len(ms) != 1 || ms[0].ID != ModelID {
		t.Fatalf("registry lookup %v: %v", s.Call, ms)
	}
	return NewEvaluator(h, regs, unresolved).Evaluate(ms[0], s)
}

func TestWebXMLRegistrationsBoundTheDispatch(t *testing.T) {
	h := fixture()
	in := Inputs{WebXMLServlets: []string{"app.Hello", "app.Missing"}, WebXMLFilters: []string{"app.Audit"}}
	regs, unresolved := Find(h, in)
	want := []Registration{{"app.Audit", KindFilter, "web.xml"}, {"app.Hello", KindServlet, "web.xml"}}
	if !reflect.DeepEqual(regs, want) || len(unresolved) != 0 {
		t.Fatalf("Find = %v %v, want %v", regs, unresolved, want)
	}

	// Servlet#service from the container: the method the JVM selects on
	// app.Hello is HttpServlet's own body — never app.Unused.
	f := evaluate(t, h, in, site("org.container.Chain", "internalDoFilter", "()V", "jakarta.servlet.Servlet", "service", serviceDesc, 9))
	if got := targets(f); !reflect.DeepEqual(got, []string{"jakarta.servlet.http.HttpServlet#service/interface"}) || f.Incomplete {
		t.Errorf("service targets = %v (incomplete %t)", got, f.Incomplete)
	}
	f = evaluate(t, h, in, site("org.container.Chain", "internalDoFilter", "()V", "jakarta.servlet.Filter", "doFilter", doFilterDesc, 5))
	if got := targets(f); !reflect.DeepEqual(got, []string{"app.Audit#doFilter/interface"}) {
		t.Errorf("doFilter targets = %v", got)
	}
	// HttpServlet's template self-call: to the registered override only.
	f = evaluate(t, h, in, site("jakarta.servlet.http.HttpServlet", "service", httpDesc, "jakarta.servlet.http.HttpServlet", "doGet", httpDesc, 20))
	if got := targets(f); !reflect.DeepEqual(got, []string{"app.Hello#doGet/virtual"}) {
		t.Errorf("doGet targets = %v", got)
	}
}

func TestSuperCallAndForeignCallerAreNotExpanded(t *testing.T) {
	h := fixture()
	in := Inputs{WebXMLServlets: []string{"app.Hello"}}
	// super.doPost (invokespecial) already has its direct edge.
	if f := evaluate(t, h, in, site("app.Hello", "doPost", httpDesc, "jakarta.servlet.http.HttpServlet", "doPost", httpDesc, 3)); len(f.Targets) != 0 {
		t.Errorf("invokespecial expanded to %v", targets(f))
	}
	// A template signature called from outside HttpServlet: the receiver is
	// not HttpServlet's own `this`, so the model does not claim it.
	other := class("app.Caller", "java.lang.Object", nil, true,
		method("run", "()V", false, classfile.Call{Kind: classfile.CallVirtual, Owner: "jakarta.servlet.http.HttpServlet", Name: "doGet", Desc: httpDesc, PC: 1}))
	h = fixture(other)
	if f := evaluate(t, h, in, site("app.Caller", "run", "()V", "jakarta.servlet.http.HttpServlet", "doGet", httpDesc, 1)); len(f.Targets) != 0 {
		t.Errorf("foreign caller expanded to %v", targets(f))
	}
}

func TestNoRegistrationNoEdge(t *testing.T) {
	h := fixture()
	f := evaluate(t, h, Inputs{}, site("org.container.Chain", "internalDoFilter", "()V", "jakarta.servlet.Servlet", "service", serviceDesc, 9))
	if len(f.Targets) != 0 || f.Incomplete {
		t.Errorf("no registration: targets %v incomplete %t", targets(f), f.Incomplete)
	}
}

func bootApp(ann classfile.Annotation) *hierarchy.ClassInfo {
	c := class("app.Main", "java.lang.Object", nil, true)
	c.Annotations = []string{ann.Type}
	c.AnnotationDetails = []classfile.Annotation{ann}
	return c
}

func TestSpringBootCatalog(t *testing.T) {
	boot := []*hierarchy.ClassInfo{
		class("org.springframework.boot.autoconfigure.web.servlet.DispatcherServletAutoConfiguration", "java.lang.Object", nil, false),
		class("org.springframework.web.servlet.DispatcherServlet", "jakarta.servlet.http.HttpServlet", nil, false, method("doGet", httpDesc, false)),
		class("org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration", "java.lang.Object", nil, false),
		class("org.springframework.boot.web.servlet.filter.OrderedFormContentFilter", "java.lang.Object", []string{"jakarta.servlet.Filter"}, false, method("doFilter", doFilterDesc, false)),
		// SecurityFilterAutoConfiguration and DelegatingFilterProxy without
		// Spring Security on the class path: its @ConditionalOnClass fails.
		class("org.springframework.boot.autoconfigure.security.servlet.SecurityFilterAutoConfiguration", "java.lang.Object", nil, false),
		class("org.springframework.web.filter.DelegatingFilterProxy", "java.lang.Object", []string{"jakarta.servlet.Filter"}, false, method("doFilter", doFilterDesc, false)),
	}
	ann := classfile.Annotation{Type: "org.springframework.boot.autoconfigure.SpringBootApplication"}

	// Not a Boot application: nothing from the catalog.
	if regs, _ := Find(fixture(boot...), Inputs{}); len(regs) != 0 {
		t.Errorf("without @SpringBootApplication: %v", regs)
	}
	// Boot application: DispatcherServlet; WebMvcAutoConfiguration's filter
	// needs WebMvcConfigurer (absent here); no security filter.
	regs, _ := Find(fixture(append(boot, bootApp(ann))...), Inputs{})
	want := []Registration{{"org.springframework.web.servlet.DispatcherServlet", KindServlet, "Spring Boot DispatcherServletAutoConfiguration"}}
	if !reflect.DeepEqual(regs, want) {
		t.Errorf("Boot app: %v, want %v", regs, want)
	}
	// Excluded auto-configuration.
	excl := ann
	excl.Values = map[string]classfile.ElementValue{"exclude": {Kind: classfile.ElementArray, Array: []classfile.ElementValue{
		{Kind: classfile.ElementClass, ClassFQCN: "org.springframework.boot.autoconfigure.web.servlet.DispatcherServletAutoConfiguration"}}}}
	if regs, _ := Find(fixture(append(boot, bootApp(excl))...), Inputs{}); len(regs) != 0 {
		t.Errorf("excluded auto-configuration still registered: %v", regs)
	}
}

func TestSpringBeansAndUnknownRegistrations(t *testing.T) {
	in := Inputs{Beans: []Bean{
		{Class: "app.Audit", Detail: "stereotype"},
		{Class: "app.Plain", Detail: "stereotype"}, // not in the hierarchy
		{Class: "org.springframework.boot.web.servlet.FilterRegistrationBean", Detail: "@Bean myFilter", Method: "@Bean myFilter"},
	}}
	// Not a Boot application: Spring does not register filter beans.
	if regs, unresolved := Find(fixture(), in); len(regs) != 0 || len(unresolved) != 0 {
		t.Errorf("plain Spring: %v %v", regs, unresolved)
	}
	h := fixture(
		class("org.springframework.boot.web.servlet.FilterRegistrationBean", "java.lang.Object", nil, false),
		bootApp(classfile.Annotation{Type: "org.springframework.boot.autoconfigure.EnableAutoConfiguration"}))
	regs, unresolved := Find(h, in)
	if want := []Registration{{"app.Audit", KindFilter, "Spring bean (stereotype)"}}; !reflect.DeepEqual(regs, want) {
		t.Errorf("regs = %v, want %v", regs, want)
	}
	if len(unresolved) != 1 || unresolved[0].Where != "@Bean myFilter" {
		t.Fatalf("unresolved = %v", unresolved)
	}
	// A container dispatch is then incomplete, with the reason.
	f := evaluate(t, h, in, site("org.container.Chain", "internalDoFilter", "()V", "jakarta.servlet.Filter", "doFilter", doFilterDesc, 5))
	if !f.Incomplete || !strings.Contains(f.Reason, "@Bean myFilter") || !reflect.DeepEqual(targets(f), []string{"app.Audit#doFilter/interface"}) {
		t.Errorf("fact = %+v", f)
	}
}

func TestAnnotatedRegistrations(t *testing.T) {
	ws := class("app.Annotated", "jakarta.servlet.http.HttpServlet", nil, true, method("doGet", httpDesc, false))
	ws.Annotations = []string{"jakarta.servlet.annotation.WebServlet"}
	regs, _ := Find(fixture(ws), Inputs{})
	if want := []Registration{{"app.Annotated", KindServlet, "@WebServlet"}}; !reflect.DeepEqual(regs, want) {
		t.Errorf("regs = %v, want %v", regs, want)
	}
}
