package servletdispatch

import (
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// TestObjectEdgesGatedByCreator: a call on HttpServletRequest reaches the
// body of the container's request class, created by reachable code, and
// not the body of a wrapper only an unreachable method creates.
func TestObjectEdgesGatedByCreator(t *testing.T) {
	const getHeader = "(Ljava/lang/String;)Ljava/lang/String;"
	req := "javax.servlet.http.HttpServletRequest" // jakarta.* is not skipped: CHA already expands it
	creator := func(name, created string) hierarchy.Method {
		m := method(name, "()V", false)
		m.InitRefs = []classfile.InitRef{{Kind: classfile.InitNew, Owner: created, PC: 1}}
		return m
	}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		iface(req, method("getHeader", getHeader, true)),
		class("org.container.Req", "java.lang.Object", []string{req}, false, method("getHeader", getHeader, false)),
		class("org.container.Conn", "java.lang.Object", nil, false, creator("accept", "org.container.Req")),
		class("app.Wrap", "java.lang.Object", []string{req}, true, method("getHeader", getHeader, false)),
		class("app.UnusedFilter", "java.lang.Object", nil, true, creator("wrap", "app.Wrap")),
		class("app.NeverCreated", "java.lang.Object", []string{req}, true, method("getHeader", getHeader, false)),
		class("app.Ctl", "java.lang.Object", nil, true,
			method("handle", "()V", false, classfile.Call{Kind: classfile.CallInterface, Owner: req, Name: "getHeader", Desc: getHeader, PC: 7})),
	})
	edges := ObjectEdges(h)
	if len(edges) != 2 {
		t.Fatalf("candidates %+v, want Req and Wrap (NeverCreated has no creator)", edges)
	}
	g := callgraph.Build(h, nil)
	id := func(owner, name, desc string) uint32 {
		n, ok := g.ID(owner, name, desc)
		if !ok {
			t.Fatalf("no node %s#%s", owner, name)
		}
		return n
	}
	gate := callgraph.NewCreatorGate(g, h, edges, callgraph.BasisModel)
	st := gate.Apply(g, []uint32{id("org.container.Conn", "accept", "()V"), id("app.Ctl", "handle", "()V")})
	if st.Added != 1 || st.Gated != 1 {
		t.Errorf("stats %+v, want 1 added, 1 gated", st)
	}
	from := id("app.Ctl", "handle", "()V")
	if e, ok := g.EdgeBetween(from, id("org.container.Req", "getHeader", getHeader)); !ok || e.Basis != callgraph.BasisModel {
		t.Errorf("no model edge to the container's request body (%+v, %v)", e, ok)
	}
	if _, ok := g.EdgeBetween(from, id("app.Wrap", "getHeader", getHeader)); ok {
		t.Error("edge to a wrapper only an unreachable method creates")
	}
}

// TestBootCatalogThymeleafResourceFilter: ResourceUrlEncodingFilter is
// registered only with Thymeleaf's MVC configuration and webjars-locator.
func TestBootCatalogThymeleafResourceFilter(t *testing.T) {
	boot := class("app.Main", "java.lang.Object", nil, true)
	boot.AnnotationDetails = []classfile.Annotation{{Type: "org.springframework.boot.autoconfigure.SpringBootApplication"}}
	filter := "org.springframework.web.servlet.resource.ResourceUrlEncodingFilter"
	base := []*hierarchy.ClassInfo{
		boot,
		class("org.springframework.boot.autoconfigure.thymeleaf.ThymeleafAutoConfiguration", "java.lang.Object", nil, false),
		class("org.springframework.boot.autoconfigure.thymeleaf.ThymeleafAutoConfiguration$ThymeleafWebMvcConfiguration", "java.lang.Object", nil, false),
		class("org.thymeleaf.spring5.SpringTemplateEngine", "java.lang.Object", nil, false),
		class(filter, "java.lang.Object", []string{"jakarta.servlet.Filter"}, false, method("doFilter", doFilterDesc, false)),
	}
	has := func(regs []Registration) bool {
		for _, r := range regs {
			if r.Class == filter {
				return true
			}
		}
		return false
	}
	if regs, _ := Find(fixture(base...), Inputs{}); has(regs) {
		t.Error("registered without webjars-locator")
	}
	withLocator := append(base, class("org.webjars.WebJarAssetLocator", "java.lang.Object", nil, false))
	if regs, _ := Find(fixture(withLocator...), Inputs{}); !has(regs) {
		t.Error("not registered with webjars-locator")
	}
}
