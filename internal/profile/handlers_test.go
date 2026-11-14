package profile

import (
	"sort"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func hmethod(name, desc string, flags uint16) hierarchy.Method {
	return hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc, AccessFlags: flags}}
}

// TestHandlerRoots (Roots.Handlers, the stripes profile's convention): in
// an app class implementing the type, every public, non-static,
// non-abstract, no-argument method returning the type's Returns or a
// subtype of it is a root; a method with parameters, a private or static
// one, another return type, or a library class's method is not.
func TestHandlerRoots(t *testing.T) {
	const (
		bean       = "net.sourceforge.stripes.action.ActionBean"
		resolution = "net.sourceforge.stripes.action.Resolution"
		forward    = "net.sourceforge.stripes.action.ForwardResolution"
		pub        = classfile.AccPublic
	)
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: bean, AccessFlags: classfile.AccInterface | classfile.AccAbstract},
		{FQCN: resolution, AccessFlags: classfile.AccInterface | classfile.AccAbstract},
		{FQCN: forward, SuperClass: "java.lang.Object", Interfaces: []string{resolution}},
		{FQCN: "app.AccountBean", SuperClass: "app.AbstractBean", IsApp: true, Methods: []hierarchy.Method{
			hmethod("signon", "()Lnet/sourceforge/stripes/action/Resolution;", pub),
			hmethod("view", "()Lnet/sourceforge/stripes/action/ForwardResolution;", pub),
			hmethod("withArg", "(Ljava/lang/String;)Lnet/sourceforge/stripes/action/Resolution;", pub),
			hmethod("hidden", "()Lnet/sourceforge/stripes/action/Resolution;", 0x0002),
			hmethod("factory", "()Lnet/sourceforge/stripes/action/Resolution;", pub|classfile.AccStatic),
			hmethod("getUsername", "()Ljava/lang/String;", pub),
		}},
		{FQCN: "app.AbstractBean", SuperClass: "java.lang.Object", Interfaces: []string{bean}, IsApp: true,
			AccessFlags: classfile.AccAbstract, Methods: []hierarchy.Method{
				hmethod("cancel", "()Lnet/sourceforge/stripes/action/Resolution;", pub),
				hmethod("abstractOne", "()Lnet/sourceforge/stripes/action/Resolution;", pub|classfile.AccAbstract),
			}},
		{FQCN: "lib.LibraryBean", SuperClass: "java.lang.Object", Interfaces: []string{bean}, Methods: []hierarchy.Method{
			hmethod("handle", "()Lnet/sourceforge/stripes/action/Resolution;", pub),
		}},
	})
	g := callgraph.Build(h, nil)
	p := Profile{Roots: Roots{Handlers: []Handler{{Implements: bean, Returns: resolution}}}}
	var got []string
	for _, id := range p.AppRoots(g, h) {
		got = append(got, g.Key(id))
	}
	sort.Strings(got)
	want := []string{
		"app.AbstractBean#cancel()Lnet/sourceforge/stripes/action/Resolution;",
		"app.AccountBean#signon()Lnet/sourceforge/stripes/action/Resolution;",
		"app.AccountBean#view()Lnet/sourceforge/stripes/action/ForwardResolution;",
	}
	if len(got) != len(want) {
		t.Fatalf("roots = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("roots = %v, want %v", got, want)
		}
	}
}

// TestStripesProfile: the embedded stripes profile is detected by
// ActionBean and carries the handler convention.
func TestStripesProfile(t *testing.T) {
	c, err := LoadDefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	p, ok := c.Get("stripes")
	if !ok {
		t.Fatal("no stripes profile")
	}
	if len(p.Detect) != 1 || p.Detect[0] != "net.sourceforge.stripes.action.ActionBean" ||
		len(p.Roots.Handlers) != 1 || p.Roots.Handlers[0].Returns != "net.sourceforge.stripes.action.Resolution" {
		t.Errorf("stripes profile = %+v", p)
	}
}
