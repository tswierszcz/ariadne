package profile

import (
	"reflect"
	"sort"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

const (
	componentFQCN = "org.springframework.stereotype.Component"
	serviceFQCN   = "org.springframework.stereotype.Service"
	profileFQCN   = "org.springframework.context.annotation.Profile"
)

func ann(fqcn string, values map[string]classfile.ElementValue) classfile.Annotation {
	return classfile.Annotation{Type: fqcn, Values: values}
}

func str(s string) classfile.ElementValue {
	return classfile.ElementValue{Kind: classfile.ElementConst, Const: s}
}

func bmethod(name, desc string, flags uint16, anns ...classfile.Annotation) hierarchy.Method {
	m := hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc, AccessFlags: flags | classfile.AccPublic, AnnotationDetails: anns}}
	for _, a := range anns {
		m.Annotations = append(m.Annotations, a.Type)
	}
	return m
}

func bclass(fqcn, super string, app bool, flags uint16, anns []classfile.Annotation, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	c := &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: super, IsApp: app, AccessFlags: flags, AnnotationDetails: anns, Methods: methods}
	for _, a := range anns {
		c.Annotations = append(c.Annotations, a.Type)
	}
	return c
}

func newing(m hierarchy.Method, classes ...string) hierarchy.Method {
	for i, c := range classes {
		m.InitRefs = append(m.InitRefs, classfile.InitRef{Kind: classfile.InitNew, Owner: c, PC: 10 + i})
	}
	return m
}

func beanFixture() *hierarchy.Hierarchy {
	comp := []classfile.Annotation{ann(componentFQCN, nil)}
	return hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		bclass("java.lang.Object", "", false, 0, nil),
		// One constructor: picked whatever its parameters.
		bclass("app.Single", "java.lang.Object", true, 0, []classfile.Annotation{ann(serviceFQCN, nil)},
			bmethod("<init>", "(Lapp/Repo;)V", 0)),
		// Several constructors: the @Autowired one.
		bclass("app.Chosen", "java.lang.Object", true, 0, comp,
			bmethod("<init>", "()V", 0),
			bmethod("<init>", "(Lapp/Repo;)V", 0, ann(autowiredFQCN, nil))),
		// Several, none annotated: the no-argument one; lifecycle methods
		// here and in an application superclass; a static one is not.
		bclass("app.Base", "java.lang.Object", true, classfile.AccAbstract, nil,
			bmethod("<init>", "()V", 0),
			bmethod("ready", "()V", 0, ann("javax.annotation.PostConstruct", nil))),
		bclass("app.Fallback", "app.Base", true, 0, comp,
			bmethod("<init>", "()V", 0),
			bmethod("<init>", "(I)V", 0),
			bmethod("setRepo", "(Lapp/Repo;)V", 0, ann(autowiredFQCN, nil)),
			bmethod("bye", "()V", 0, ann("jakarta.annotation.PreDestroy", nil)),
			bmethod("helper", "()V", 0),
			bmethod("make", "()V", classfile.AccStatic, ann(autowiredFQCN, nil))),
		// Several, none annotated, no no-argument one: Spring fails; none.
		bclass("app.Broken", "java.lang.Object", true, 0, comp,
			bmethod("<init>", "(I)V", 0),
			bmethod("<init>", "(J)V", 0)),
		// Not beans: no stereotype; abstract; a library class.
		bclass("app.Plain", "java.lang.Object", true, 0, nil, bmethod("<init>", "()V", 0)),
		bclass("app.AbstractBean", "java.lang.Object", true, classfile.AccAbstract, comp, bmethod("<init>", "()V", 0)),
		bclass("lib.LibBean", "java.lang.Object", false, 0, comp, bmethod("<init>", "()V", 0)),
		// A @Profile bean: unknown without -active-profiles.
		bclass("app.Dormant", "java.lang.Object", true, 0,
			[]classfile.Annotation{ann(componentFQCN, nil), ann(profileFQCN, map[string]classfile.ElementValue{"value": strArrayElement("prod")})},
			bmethod("<init>", "()V", 0)),
		// @Bean products: an explicit initMethod/destroyMethod on a concrete
		// return type; an inferred close on the class a method declared to
		// return an interface creates; destroyMethod="" turns inference off.
		bclass("app.Config", "java.lang.Object", true, 0, []classfile.Annotation{ann(configurationFQCN, nil)},
			bmethod("<init>", "()V", 0),
			bmethod("engine", "()Lapp/Engine;", 0, ann(beanMethodFQCN, map[string]classfile.ElementValue{"initMethod": str("start"), "destroyMethod": str("stop")})),
			newing(bmethod("pool", "()Llib/Pool;", 0, ann(beanMethodFQCN, nil)), "lib.HikariPool"),
			bmethod("quiet", "()Lapp/Quiet;", 0, ann(beanMethodFQCN, map[string]classfile.ElementValue{"destroyMethod": str("")}))),
		bclass("app.Engine", "java.lang.Object", true, 0, nil,
			bmethod("start", "()V", 0), bmethod("stop", "()V", 0), bmethod("close", "()V", 0)),
		bclass("lib.Pool", "", false, classfile.AccInterface|classfile.AccAbstract, nil, bmethod("close", "()V", classfile.AccAbstract)),
		bclass("lib.PoolBase", "java.lang.Object", false, 0, nil, bmethod("shutdown", "()V", 0), bmethod("close", "()V", 0)),
		{FQCN: "lib.HikariPool", SuperClass: "lib.PoolBase", Interfaces: []string{"lib.Pool"}},
		bclass("app.Quiet", "java.lang.Object", true, 0, nil, bmethod("close", "()V", 0)),
	})
}

func rootSet(roots []BeanRoot) []string {
	var out []string
	for _, r := range roots {
		out = append(out, r.FQCN+"#"+r.Name+r.Desc+" "+string(r.Kind))
	}
	sort.Strings(out)
	return out
}

// TestBeanLifecycleRoots: the constructor Spring picks (the only one, the
// @Autowired one, the no-argument one, none when it cannot pick), injection
// and lifecycle methods of the bean and its application superclasses (not a
// static one), a @Bean's explicit and inferred init/destroy methods — and
// nothing for a class with no stereotype, an abstract one or a library one.
// A @Profile bean counts while its profile is unknown and drops out once
// -active-profiles says it is inactive.
func TestBeanLifecycleRoots(t *testing.T) {
	h := beanFixture()
	common := []string{
		"app.Base#ready()V init annotation",
		"app.Chosen#<init>(Lapp/Repo;)V constructor",
		"app.Config#<init>()V constructor",
		"app.Engine#start()V @Bean initMethod",
		"app.Engine#stop()V @Bean destroyMethod",
		"app.Fallback#<init>()V constructor",
		"app.Fallback#bye()V destroy annotation",
		"app.Fallback#setRepo(Lapp/Repo;)V injection method",
		"app.Single#<init>(Lapp/Repo;)V constructor",
		"lib.PoolBase#close()V @Bean destroyMethod",
	}
	unknown := append([]string{"app.Dormant#<init>()V constructor"}, common...)
	sort.Strings(unknown)
	if got := rootSet(BeanLifecycleRoots(h, nil)); !reflect.DeepEqual(got, unknown) {
		t.Errorf("no -active-profiles:\n got %q\nwant %q", got, unknown)
	}
	if got := rootSet(BeanLifecycleRoots(h, []string{"default"})); !reflect.DeepEqual(got, common) {
		t.Errorf("-active-profiles default:\n got %q\nwant %q", got, common)
	}
	if got := rootSet(BeanLifecycleRoots(h, []string{"prod"})); !reflect.DeepEqual(got, unknown) {
		t.Errorf("-active-profiles prod:\n got %q\nwant %q", got, unknown)
	}
}

// TestSpringProfileHasBeanRoots: the embedded spring profile turns the
// rule on, and merging keeps it on.
func TestSpringProfileHasBeanRoots(t *testing.T) {
	c, err := LoadDefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	p, _ := c.Get("spring")
	g, _ := c.Get("generic")
	if !p.Roots.Beans {
		t.Fatal("spring profile: roots.beans is off")
	}
	if g.Roots.Beans || !merge([]Profile{g, p}).Roots.Beans {
		t.Error("generic must not set roots.beans; a merge with spring must keep it")
	}
}
