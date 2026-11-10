// Package proxycatalog links calls on interfaces that frameworks
// implement with JDK dynamic proxies to the InvocationHandler the JVM
// runs (FRAMEWORKS:F3, slice 3: a catalog model).
//
// A call on a MyBatis mapper interface (CatalogService calling
// CategoryMapper#getCategory) has no implementation in the input: MyBatis
// hands out Proxy.newProxyInstance(…, new Class[]{mapperInterface},
// new MapperProxy(…)), so the graph used to stop at the interface method
// while the JVM runs MapperProxy#invoke → MapperMethod#execute → the
// SqlSession → the JDBC driver. Slice 2 measured why a general rule cannot
// close this: at none of the measured Proxy.newProxyInstance sites is the
// interface set bounded — it comes from configuration (the mapper
// scanner, Spring AOP's advised interfaces) or from objects created at run
// time. The interface set is, however, what the framework documents, so
// this model takes it from a catalog (Proxies, Version):
//
//   - MyBatis mapper proxies: every mapper interface → MapperProxy#invoke.
//     A mapper is an interface named as the namespace of a MyBatis mapper
//     XML in the input, annotated @Mapper, or in a package a @MapperScan or
//     a Spring XML <mybatis:scan>/MapperScannerConfigurer in the input
//     scans;
//   - mybatis-spring's SqlSessionTemplate: its calls on its own SqlSession
//     proxy → SqlSessionTemplate$SqlSessionInterceptor#invoke;
//   - Spring Data repositories: every repository interface (an
//     application interface extending Spring Data's Repository) →
//     JdkDynamicAopProxy#invoke;
//   - Spring ORM's JPA proxies: EntityManager → the shared and the
//     extended EntityManager handlers, Query/TypedQuery/
//     StoredProcedureQuery → the deferred-query handler,
//     EntityManagerFactory → the managed factory handler (javax and
//     jakarta).
//
// Each edge is a candidate for the creator gate (callgraph.NewCreatorGate,
// shared with the SAM, servlet-object and JDBC models): it is added only
// once a method that creates the handler (`new`) is reachable from the
// roots, so a framework on the class path that never builds its proxies
// adds nothing. The edge says the handler may run; what the handler then
// calls through Method#invoke is reflection (REFLECTION R2's scope), but
// MyBatis's MapperProxy calls MapperMethod#execute directly.
package proxycatalog

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"sort"
	"strings"

	"ariadne/internal/appref"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// ModelID identifies the model in logs; its edges carry
// callgraph.BasisModel.
const ModelID = "framework:proxies"

// Version versions the catalog and the mapper rules. Snapshots record it.
const Version = "x1"

const invokeDesc = "(Ljava/lang/Object;Ljava/lang/reflect/Method;[Ljava/lang/Object;)Ljava/lang/Object;"

// Proxy is one catalog entry: calls on Interfaces (or on the interfaces
// the entry's own rule finds) reach Handler#invoke.
type Proxy struct {
	Name    string
	Handler string
	// Interfaces are the fixed proxied interfaces; nil for an entry whose
	// interfaces are found in the input (mappers, repositories).
	Interfaces []string
	// CallerClass, when set, restricts the entry to calls made inside
	// that class (SqlSessionTemplate calls its own SqlSession proxy).
	CallerClass string
}

// Proxies is the catalog, checked in the bytecode of mybatis 3.4.1, 3.4.6,
// 3.5.6, 3.5.16; mybatis-spring 1.3.0, 2.0.7, 3.0.3; spring-aop and
// spring-orm 4.3.4.RELEASE, 5.3.6, 6.1.14: each handler declares
// invoke(Object, Method, Object[]) and the listed creator `new`s it.
var Proxies = []Proxy{
	{Name: "mybatis:mapper", Handler: "org.apache.ibatis.binding.MapperProxy"},
	{Name: "mybatis-spring:sqlsession", Handler: "org.mybatis.spring.SqlSessionTemplate$SqlSessionInterceptor",
		Interfaces: []string{"org.apache.ibatis.session.SqlSession"}, CallerClass: "org.mybatis.spring.SqlSessionTemplate"},
	{Name: "spring-data:repository", Handler: "org.springframework.aop.framework.JdkDynamicAopProxy"},
	{Name: "spring-orm:shared-entity-manager", Handler: "org.springframework.orm.jpa.SharedEntityManagerCreator$SharedEntityManagerInvocationHandler",
		Interfaces: []string{"javax.persistence.EntityManager", "jakarta.persistence.EntityManager"}},
	{Name: "spring-orm:extended-entity-manager", Handler: "org.springframework.orm.jpa.ExtendedEntityManagerCreator$ExtendedEntityManagerInvocationHandler",
		Interfaces: []string{"javax.persistence.EntityManager", "jakarta.persistence.EntityManager"}},
	{Name: "spring-orm:deferred-query", Handler: "org.springframework.orm.jpa.SharedEntityManagerCreator$DeferredQueryInvocationHandler",
		Interfaces: []string{"javax.persistence.Query", "javax.persistence.TypedQuery", "javax.persistence.StoredProcedureQuery",
			"jakarta.persistence.Query", "jakarta.persistence.TypedQuery", "jakarta.persistence.StoredProcedureQuery"}},
	{Name: "spring-orm:entity-manager-factory", Handler: "org.springframework.orm.jpa.AbstractEntityManagerFactoryBean$ManagedEntityManagerFactoryInvocationHandler",
		Interfaces: []string{"javax.persistence.EntityManagerFactory", "jakarta.persistence.EntityManagerFactory"}},
}

const (
	mapperAnnotation     = "org.apache.ibatis.annotations.Mapper"
	mapperScanAnnotation = "org.mybatis.spring.annotation.MapperScan"
	scannerConfigurer    = "org.mybatis.spring.mapper.MapperScannerConfigurer"
	springDataRepository = "org.springframework.data.repository.Repository"
)

// Mapper is one MyBatis mapper interface and why it is one.
type Mapper struct {
	Interface string
	Why       string
}

// Result is the catalog applied to one input.
type Result struct {
	Mappers      []Mapper
	Repositories []string
	// Edges are the candidates for callgraph.NewCreatorGate.
	Edges []callgraph.SAMEdge
	// ByProxy counts candidates per catalog entry.
	ByProxy map[string]int
	// NoCreator lists the entries whose handler is in the input but has
	// no method creating it (no candidates then).
	NoCreator []string
}

// Find detects mappers and repositories and builds the candidate edges.
// jars are the analyzed class path's jars (mapper XML and Spring XML are
// read from them).
func Find(h *hierarchy.Hierarchy, jars []appref.Jar) (Result, error) {
	r := Result{ByProxy: map[string]int{}}
	mappers, err := findMappers(h, jars)
	if err != nil {
		return r, err
	}
	r.Mappers = mappers
	for _, fqcn := range h.Subtypes(springDataRepository) {
		if ci, ok := h.Lookup(fqcn); ok && ci.IsApp && ci.IsInterface() {
			r.Repositories = append(r.Repositories, fqcn)
		}
	}
	sort.Strings(r.Repositories)

	creators := map[string][]callgraph.MethodRef{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, ref := range m.InitRefs {
				if ref.Kind == classfile.InitNew {
					creators[ref.Owner] = append(creators[ref.Owner], callgraph.MethodRef{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor})
				}
			}
		}
	}
	// interface → the entries (handlers) its calls reach
	type target struct {
		proxy    Proxy
		creators []callgraph.MethodRef
	}
	byIface := map[string][]target{}
	for _, p := range Proxies {
		if _, ok := h.Lookup(p.Handler); !ok {
			continue
		}
		cr := dedupe(creators[p.Handler])
		if len(cr) == 0 {
			r.NoCreator = append(r.NoCreator, p.Name)
			continue
		}
		ifaces := p.Interfaces
		switch p.Name {
		case "mybatis:mapper":
			for _, m := range mappers {
				ifaces = append(ifaces, m.Interface)
			}
		case "spring-data:repository":
			ifaces = append(ifaces, r.Repositories...)
		}
		for _, i := range ifaces {
			byIface[i] = append(byIface[i], target{p, cr})
		}
	}
	if len(byIface) == 0 {
		return r, nil
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind != classfile.CallInterface && c.Kind != classfile.CallVirtual {
					continue
				}
				for _, t := range byIface[c.Owner] {
					if t.proxy.CallerClass != "" && fqcn != t.proxy.CallerClass {
						continue
					}
					r.Edges = append(r.Edges, callgraph.SAMEdge{
						CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC, Kind: c.Kind,
						TargetFQCN: t.proxy.Handler, TargetName: "invoke", TargetDesc: invokeDesc,
						Creators: t.creators,
					})
					r.ByProxy[t.proxy.Name]++
				}
			}
		}
	}
	return r, nil
}

func dedupe(in []callgraph.MethodRef) []callgraph.MethodRef {
	seen := map[callgraph.MethodRef]bool{}
	var out []callgraph.MethodRef
	for _, m := range in {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

// findMappers applies the mapper rules of the package comment.
func findMappers(h *hierarchy.Hierarchy, jars []appref.Jar) ([]Mapper, error) {
	found := map[string]string{}
	add := func(iface, why string) {
		ci, ok := h.Lookup(iface)
		if !ok || !ci.IsInterface() {
			return
		}
		if _, dup := found[iface]; !dup {
			found[iface] = why
		}
	}
	var packages []struct{ pkg, why string }
	for _, j := range jars {
		r, err := zip.OpenReader(j.Path)
		if err != nil {
			return nil, err
		}
		for _, f := range r.File {
			if !strings.HasSuffix(f.Name, ".xml") || f.UncompressedSize64 > 4<<20 {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				r.Close()
				return nil, err
			}
			data, _ := io.ReadAll(rc)
			rc.Close()
			where := j.Name + "!/" + f.Name
			if bytes.Contains(data, []byte("<mapper")) && bytes.Contains(data, []byte("namespace")) {
				if ns := mapperNamespace(data); ns != "" {
					add(ns, "namespace of "+where)
				}
			}
			if bytes.Contains(data, []byte("base-package")) || bytes.Contains(data, []byte(scannerConfigurer)) {
				for _, p := range scannedPackages(data) {
					packages = append(packages, struct{ pkg, why string }{p, "scanned by " + where})
				}
			}
		}
		r.Close()
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, a := range ci.Annotations {
			if a == mapperAnnotation {
				add(fqcn, "@Mapper")
			}
		}
		for _, a := range ci.AnnotationDetails {
			if a.Type != mapperScanAnnotation {
				continue
			}
			for _, name := range []string{"value", "basePackages", "basePackageClasses"} {
				v, ok := a.Value(name)
				if !ok {
					continue
				}
				vals := []classfile.ElementValue{v}
				if v.Kind == classfile.ElementArray {
					vals = v.Array
				}
				for _, e := range vals {
					switch e.Kind {
					case classfile.ElementConst:
						if s, ok := e.Const.(string); ok {
							for _, p := range splitPackages(s) {
								packages = append(packages, struct{ pkg, why string }{p, "@MapperScan on " + fqcn})
							}
						}
					case classfile.ElementClass:
						if i := strings.LastIndex(e.ClassFQCN, "."); i > 0 {
							packages = append(packages, struct{ pkg, why string }{e.ClassFQCN[:i], "@MapperScan on " + fqcn})
						}
					}
				}
			}
		}
	}
	if len(packages) > 0 {
		for _, fqcn := range h.SortedFQCNs() {
			ci, _ := h.Lookup(fqcn)
			if !ci.IsApp || !ci.IsInterface() {
				continue
			}
			for _, p := range packages {
				if strings.HasPrefix(fqcn, p.pkg+".") {
					add(fqcn, p.why)
					break
				}
			}
		}
	}
	out := make([]Mapper, 0, len(found))
	for i, why := range found {
		out = append(out, Mapper{Interface: i, Why: why})
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Interface < out[b].Interface })
	return out, nil
}

// mapperNamespace returns the namespace of a MyBatis mapper XML (root
// element <mapper namespace="...">), or "".
func mapperNamespace(data []byte) string {
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	for {
		tok, err := d.Token()
		if err != nil {
			return ""
		}
		if se, ok := tok.(xml.StartElement); ok {
			if se.Name.Local != "mapper" {
				return ""
			}
			for _, a := range se.Attr {
				if a.Name.Local == "namespace" {
					return strings.TrimSpace(a.Value)
				}
			}
			return ""
		}
	}
}

// scannedPackages returns the packages a Spring XML scans for mappers:
// <mybatis:scan base-package="..."> and a MapperScannerConfigurer bean's
// basePackage property. A placeholder package is skipped.
func scannedPackages(data []byte) []string {
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	var out []string
	inConfigurer := false
	depth := 0
	for {
		tok, err := d.Token()
		if err != nil {
			return out
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if inConfigurer {
				depth++
			}
			attr := func(name string) string {
				for _, a := range t.Attr {
					if a.Name.Local == name {
						return a.Value
					}
				}
				return ""
			}
			switch {
			case t.Name.Local == "scan" && strings.Contains(t.Name.Space, "mybatis"):
				out = append(out, splitPackages(attr("base-package"))...)
			case t.Name.Local == "bean" && attr("class") == scannerConfigurer:
				inConfigurer, depth = true, 0
			case inConfigurer && t.Name.Local == "property" && attr("name") == "basePackage":
				out = append(out, splitPackages(attr("value"))...)
			}
		case xml.EndElement:
			if inConfigurer {
				if depth == 0 {
					inConfigurer = false
				} else {
					depth--
				}
			}
		}
	}
}

func splitPackages(s string) []string {
	var out []string
	for _, p := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n' }) {
		if p != "" && !strings.Contains(p, "${") && !strings.ContainsAny(p, "*?") {
			out = append(out, p)
		}
	}
	return out
}
