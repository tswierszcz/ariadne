package profile

import (
	"reflect"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestWebXMLRoots_servletLifecycleMethodsFound(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN: "demo.MainServlet",
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "doGet", Descriptor: "(Ljavax/servlet/http/HttpServletRequest;Ljavax/servlet/http/HttpServletResponse;)V"}},
				{Method: classfile.Method{Name: "helperNotALifecycleMethod", Descriptor: "()V"}},
			},
		},
	})

	roots := WebXMLRoots(h, []string{"demo.MainServlet"}, nil, nil)
	want := []WebXMLRoot{{FQCN: "demo.MainServlet", Method: "doGet", Kind: WebRootServlet}}
	if !reflect.DeepEqual(roots, want) {
		t.Errorf("WebXMLRoots = %+v, want %+v (helper method must not become a root)", roots, want)
	}
}

func TestWebXMLRoots_filterAndListener(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN: "demo.AuthFilter",
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "doFilter", Descriptor: "()V"}},
			},
		},
		{
			FQCN: "demo.StartupListener",
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "contextInitialized", Descriptor: "()V"}},
			},
		},
	})

	roots := WebXMLRoots(h, nil, []string{"demo.AuthFilter"}, []string{"demo.StartupListener"})
	want := []WebXMLRoot{
		{FQCN: "demo.AuthFilter", Method: "doFilter", Kind: WebRootFilter},
		{FQCN: "demo.StartupListener", Method: "contextInitialized", Kind: WebRootListener},
	}
	if !reflect.DeepEqual(roots, want) {
		t.Errorf("WebXMLRoots = %+v, want %+v", roots, want)
	}
}

func TestWebXMLRoots_classNotInHierarchyIsSkipped(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	roots := WebXMLRoots(h, []string{"demo.NotFound"}, nil, nil)
	if len(roots) != 0 {
		t.Errorf("WebXMLRoots(unresolvable class) = %+v, want none", roots)
	}
}

func TestWebXMLRoots_abstractMethodSkipped(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN: "demo.AbstractServlet",
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "doGet", Descriptor: "()V", AccessFlags: classfile.AccAbstract}},
			},
		},
	})
	roots := WebXMLRoots(h, []string{"demo.AbstractServlet"}, nil, nil)
	if len(roots) != 0 {
		t.Errorf("WebXMLRoots(abstract method) = %+v, want none", roots)
	}
}

func TestWebXMLAppRoots_resolvesRealGraphNodes(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "demo.MainServlet",
			IsApp: true,
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "doGet", Descriptor: "()V"}},
				{Method: classfile.Method{Name: "notALifecycleMethod", Descriptor: "()V"}},
			},
		},
	})
	g := callgraph.Build(h, nil)

	roots := WebXMLAppRoots(g, h, []string{"demo.MainServlet"}, nil, nil)
	if len(roots) != 1 {
		t.Fatalf("WebXMLAppRoots = %v, want exactly 1 root", roots)
	}
	if g.Key(roots[0]) != "demo.MainServlet#doGet()V" {
		t.Errorf("WebXMLAppRoots root = %q, want demo.MainServlet#doGet()V", g.Key(roots[0]))
	}
}
