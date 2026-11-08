package alloc

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func cls(fqcn string, flags uint16, ms ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: "java.lang.Object", AccessFlags: flags, IsApp: true, Methods: ms}
}

func method(name, desc string, flags uint16, refs ...classfile.InitRef) hierarchy.Method {
	m := hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc, AccessFlags: flags}, InitRefs: refs}
	return m
}

func origin(c Catalog, class, kind string) *Origin {
	for _, t := range c.Types {
		if t.Class != class {
			continue
		}
		for i := range t.Origins {
			if t.Origins[i].Kind == kind {
				return &t.Origins[i]
			}
		}
	}
	return nil
}

// TestOriginsAndGaps: container beans (concrete only), container proxies
// as bounded gaps, entry classes of instance roots (profile roots only),
// the roots=all external-input rule, unresolved model sites as gaps,
// `new` of an abstract class ignored, sites deduplicated and sorted.
func TestOriginsAndGaps(t *testing.T) {
	pub := uint16(classfile.AccPublic)
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		cls("app.Ctl", pub, method("list", "(Ljava/lang/String;)V", pub), method("main", "([Ljava/lang/String;)V", pub|classfile.AccStatic)),
		cls("app.Svc", pub, method("run", "()V", pub,
			classfile.InitRef{Kind: classfile.InitNew, Owner: "app.Impl", PC: 9},
			classfile.InitRef{Kind: classfile.InitNew, Owner: "app.Impl", PC: 4},
			classfile.InitRef{Kind: classfile.InitNew, Owner: "app.Abs", PC: 12})),
		cls("app.Impl", pub),
		cls("app.Abs", pub|classfile.AccAbstract),
		cls("app.Repo", pub|classfile.AccInterface|classfile.AccAbstract),
		cls("app.Bean", pub),
	})
	in := Inputs{H: h, EntryRoots: true,
		Roots:             []string{"app.Ctl#list(Ljava/lang/String;)V", "app.Ctl#main([Ljava/lang/String;)V"},
		Container:         []ContainerBean{{Class: "app.Bean", Detail: "stereotype"}, {Class: "app.Repo", Detail: "@Bean repo"}},
		ContainerProxies:  []ContainerBean{{Class: "app.Repo", Detail: "Spring Data repository"}},
		UnresolvedService: []Site{{Method: "app.Svc#run()V", PC: 20}},
	}
	c := Build(in)
	if o := origin(c, "app.Impl", KindNew); o == nil || len(o.Sites) != 2 || o.Sites[0].PC != 4 {
		t.Errorf("new origin of app.Impl: %+v", o)
	}
	if origin(c, "app.Abs", KindNew) != nil {
		t.Error("an abstract class got a `new` origin")
	}
	if o := origin(c, "app.Bean", KindContainer); o == nil || o.Sites[0].Method != ContainerOrigin {
		t.Errorf("container origin: %+v", o)
	}
	if origin(c, "app.Repo", KindContainer) != nil {
		t.Error("an interface bean got a container origin (its class is generated: a gap)")
	}
	if o := origin(c, "app.Ctl", KindEntry); o == nil || len(o.Sites) != 1 || o.Sites[0].Method != "app.Ctl#list(Ljava/lang/String;)V" {
		t.Errorf("entry origin (instance root only): %+v", o)
	}
	kinds := map[string]int{}
	for _, g := range c.Gaps {
		kinds[g.Kind]++
		if g.Kind == GapContainerProxy && g.Bound != "app.Repo" {
			t.Errorf("container-proxy gap bound %q", g.Bound)
		}
	}
	if kinds[GapContainerProxy] != 1 || kinds[GapService] != 1 || kinds[GapExternalInput] != 0 {
		t.Errorf("gaps %v", kinds)
	}

	// roots=all: no entry origin, one external-input rule counting the
	// reference parameters of the roots.
	in.EntryRoots, in.ExternalInputs = false, true
	c = Build(in)
	if origin(c, "app.Ctl", KindEntry) != nil {
		t.Error("entry origin with roots=all")
	}
	if !c.ExternalInputs || c.RootParams != 2 {
		t.Errorf("external inputs %v, root params %d", c.ExternalInputs, c.RootParams)
	}
	s := Summarize(c, nil)
	if s.Gaps[GapExternalInput] != 1 || s.ByOrigin[KindNew] != 1 || s.ByOrigin[KindContainer] != 1 {
		t.Errorf("summary %+v", s)
	}
}
