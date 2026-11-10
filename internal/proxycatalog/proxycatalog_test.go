package proxycatalog

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"ariadne/internal/appref"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func writeJar(t *testing.T, name string, files map[string]string) appref.Jar {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for n, body := range files {
		e, _ := w.Create(n)
		e.Write([]byte(body))
	}
	w.Close()
	f.Close()
	return appref.Jar{Path: p, Name: name}
}

func iface(fqcn string, app bool, methods ...string) *hierarchy.ClassInfo {
	ci := &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: "java.lang.Object", AccessFlags: classfile.AccInterface | classfile.AccAbstract, IsApp: app}
	for _, m := range methods {
		ci.Methods = append(ci.Methods, hierarchy.Method{Method: classfile.Method{Name: m, Descriptor: "()Ljava/lang/Object;", AccessFlags: classfile.AccAbstract}})
	}
	return ci
}

func cls(fqcn string, app bool, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: "java.lang.Object", Methods: methods, IsApp: app}
}

func meth(name, desc string, calls []classfile.Call, news ...string) hierarchy.Method {
	m := hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc}, Calls: calls}
	for _, n := range news {
		m.InitRefs = append(m.InitRefs, classfile.InitRef{Kind: classfile.InitNew, Owner: n})
	}
	return m
}

const (
	mapperProxy   = "org.apache.ibatis.binding.MapperProxy"
	proxyFactory  = "org.apache.ibatis.binding.MapperProxyFactory"
	categoryMap   = "demo.mapper.CategoryMapper"
	scannedMap    = "demo.scanned.ItemMapper"
	annotatedMap  = "demo.other.OrderMapper"
	notAMapper    = "demo.service.Helper"
	catalogSvc    = "demo.service.CatalogService"
	getCategory   = "getCategory"
	objDesc       = "()Ljava/lang/Object;"
	interfaceCall = classfile.CallInterface
)

func call(owner, name string, pc int) classfile.Call {
	return classfile.Call{Kind: interfaceCall, Owner: owner, Name: name, Desc: objDesc, PC: pc}
}

func fixture(withFactory bool) *hierarchy.Hierarchy {
	annotated := iface(annotatedMap, true, "getOrder")
	annotated.Annotations = []string{mapperAnnotation}
	classes := []*hierarchy.ClassInfo{
		iface(categoryMap, true, getCategory),
		iface(scannedMap, true, "getItem"),
		annotated,
		iface(notAMapper, true, "help"),
		cls(mapperProxy, false, meth("invoke", invokeDesc, nil)),
		cls(catalogSvc, true, meth("view", "()V", []classfile.Call{
			call(categoryMap, getCategory, 1), call(scannedMap, "getItem", 5), call(annotatedMap, "getOrder", 9), call(notAMapper, "help", 13),
		})),
	}
	if withFactory {
		classes = append(classes, cls(proxyFactory, false, meth("newInstance", "()Ljava/lang/Object;", nil, mapperProxy)))
	}
	return hierarchy.NewForTesting(classes)
}

func TestFindMappers(t *testing.T) {
	app := writeJar(t, "app.jar", map[string]string{
		"demo/mapper/CategoryMapper.xml": `<?xml version="1.0"?><!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="demo.mapper.CategoryMapper"><select id="getCategory">SELECT 1</select></mapper>`,
		"applicationContext.xml": `<beans xmlns:mybatis="http://mybatis.org/schema/mybatis-spring"><mybatis:scan base-package="demo.scanned"/></beans>`,
		"other.xml":              `<mapper namespace="demo.service.CatalogService"/>`, // a class, not an interface
	})
	r, err := Find(fixture(true), []appref.Jar{app})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, m := range r.Mappers {
		got[m.Interface] = m.Why
	}
	want := map[string]string{
		categoryMap:  "namespace of app.jar!/demo/mapper/CategoryMapper.xml",
		scannedMap:   "scanned by app.jar!/applicationContext.xml",
		annotatedMap: "@Mapper",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mappers %v, want %v", got, want)
	}
	if r.ByProxy["mybatis:mapper"] != 3 || len(r.Edges) != 3 {
		t.Errorf("edges %d (%v), want 3 mapper calls", len(r.Edges), r.ByProxy)
	}

	// The gate: the edges join the graph once the factory creating
	// MapperProxy is reachable.
	h := fixture(true)
	g := callgraph.BuildWith(h, callgraph.CHA{})
	gate := callgraph.NewCreatorGate(g, h, r.Edges, callgraph.BasisModel)
	svc, _ := g.ID(catalogSvc, "view", "()V")
	if st := gate.Apply(g, []uint32{svc}); st.Added != 0 || st.Gated != 3 {
		t.Errorf("factory unreachable: %+v", st)
	}
	factory, _ := g.ID(proxyFactory, "newInstance", "()Ljava/lang/Object;")
	// The three calls are in one method: one (caller, handler) edge.
	if st := gate.Apply(g, []uint32{svc, factory}); st.Added != 1 || st.Gated != 0 {
		t.Errorf("factory reachable: %+v", st)
	}
	invoke, _ := g.ID(mapperProxy, "invoke", invokeDesc)
	if _, ok := g.EdgeBetween(svc, invoke); !ok {
		t.Error("no CatalogService#view → MapperProxy#invoke edge")
	}

	// No class creating MapperProxy: no candidates, reported.
	r, _ = Find(fixture(false), []appref.Jar{app})
	if len(r.Edges) != 0 || !reflect.DeepEqual(r.NoCreator, []string{"mybatis:mapper"}) {
		t.Errorf("without a creator: edges %d, no-creator %v", len(r.Edges), r.NoCreator)
	}
}

func TestFixedInterfacesAndCallerClass(t *testing.T) {
	const (
		em       = "javax.persistence.EntityManager"
		shared   = "org.springframework.orm.jpa.SharedEntityManagerCreator$SharedEntityManagerInvocationHandler"
		creator  = "org.springframework.orm.jpa.SharedEntityManagerCreator"
		template = "org.mybatis.spring.SqlSessionTemplate"
		session  = "org.apache.ibatis.session.SqlSession"
		intercep = "org.mybatis.spring.SqlSessionTemplate$SqlSessionInterceptor"
		repoBase = "org.springframework.data.repository.Repository"
		ownerRep = "demo.OwnerRepository"
		aop      = "org.springframework.aop.framework.JdkDynamicAopProxy"
		aopMaker = "org.springframework.aop.framework.DefaultAopProxyFactory"
	)
	repo := iface(ownerRep, true, "findAll")
	repo.Interfaces = []string{repoBase}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		iface(em, false, "find"),
		iface(session, false, "selectOne"),
		iface(repoBase, false),
		repo,
		cls(shared, false, meth("invoke", invokeDesc, nil)),
		cls(creator, false, meth("createSharedEntityManager", "()V", nil, shared)),
		cls(intercep, false, meth("invoke", invokeDesc, nil)),
		cls(aop, false, meth("invoke", invokeDesc, nil)),
		cls(aopMaker, false, meth("createAopProxy", "()V", nil, aop)),
		cls(template, false, meth("<init>", "()V", nil, intercep), meth("selectOne", objDesc, []classfile.Call{call(session, "selectOne", 3)})),
		cls("demo.Dao", true, meth("load", "()V", []classfile.Call{call(em, "find", 1), call(session, "selectOne", 4), call(ownerRep, "findAll", 7)})),
	})
	r, err := Find(h, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"spring-orm:shared-entity-manager": 1, "mybatis-spring:sqlsession": 1, "spring-data:repository": 1}
	if !reflect.DeepEqual(r.ByProxy, want) {
		t.Errorf("by proxy %v, want %v (SqlSession only inside SqlSessionTemplate)", r.ByProxy, want)
	}
	if !reflect.DeepEqual(r.Repositories, []string{ownerRep}) {
		t.Errorf("repositories %v", r.Repositories)
	}
	for _, e := range r.Edges {
		if e.TargetFQCN == intercep && e.CallerFQCN != template {
			t.Errorf("SqlSession edge from %s", e.CallerFQCN)
		}
	}
}
