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
// x2: Spring Data base repositories (BaseRepositoryEntry).
// x3: Spring's transactional proxies (TransactionalEntry).
const Version = "x3"

// BaseRepositoryEntry names the x2 rule: a call on a Spring Data
// repository interface also reaches the body a base repository class runs
// for it. Spring Data's proxy (JdkDynamicAopProxy → its interceptors →
// RepositoryMethodInvoker) runs the base class's method through
// Method#invoke, which no literal resolves (ROADMAP B4 b4-5 measured CRUD
// methods unreached on a real JVM). A base class is a concrete subtype of
// Repository in the input (SimpleJpaRepository, a store's own); it counts
// once a reachable method creates it — `new`s it, or belongs to a class
// naming it as a literal (a factory's getRepositoryBaseClass returns
// SimpleJpaRepository.class, and Spring instantiates it reflectively).
// Query methods declared on the repository itself have no base-class body
// and get no edge.
const BaseRepositoryEntry = "spring-data:base-repository"

// TransactionalEntry names the x3 rule: a call on a method Spring runs in
// a transaction reaches TransactionInterceptor#invoke. Spring hands out
// a proxy (CGLIB for a class, JDK for an interface) for every bean with a
// transactional method; the proxy runs the interceptor chain, and
// TransactionInterceptor → TransactionAspectSupport#invokeWithinTransaction
// begins, commits or rolls back through the PlatformTransactionManager
// around the call (measured on a real JVM: testdata/tx). The graph used to
// go from the call straight to the method's body.
//
// A method is transactional when @Transactional (Spring's, or JTA's javax/
// jakarta one) is on it, on its class, or on a supertype declaring it or
// on that supertype itself — the places Spring's annotation source looks;
// only public, non-static, non-final methods (Spring's default
// publicMethodsOnly; a final method cannot be proxied). A call site gets
// the edge when the method it names, on the call's owner or on any
// concrete subtype that may receive it, is transactional — except a call
// from inside that same class (self-invocation does not go through the
// proxy). Gate: a reachable method creating the interceptor — `new
// TransactionInterceptor` (@EnableTransactionManagement's configuration)
// or naming the class (the <tx:annotation-driven> parser registers it by
// class) — so an application that never enables transaction management
// gets nothing.
const TransactionalEntry = "spring-tx:transactional"

const (
	transactionInterceptor = "org.springframework.transaction.interceptor.TransactionInterceptor"
	interceptorInvokeDesc  = "(Lorg/aopalliance/intercept/MethodInvocation;)Ljava/lang/Object;"
)

var transactionalAnnotations = []string{
	"org.springframework.transaction.annotation.Transactional",
	"javax.transaction.Transactional",
	"jakarta.transaction.Transactional",
}

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
// read from them); extraXML are Spring XML files outside the class path,
// keyed by where they are — a WAR's WEB-INF/applicationContext.xml
// (input.PreparedApp.WebInfBeansXMLs), where a <mybatis:scan> or a
// MapperScannerConfigurer is usually declared.
func Find(h *hierarchy.Hierarchy, jars []appref.Jar, extraXML map[string][]byte) (Result, error) {
	r := Result{ByProxy: map[string]int{}}
	mappers, err := findMappers(h, jars, extraXML)
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
	r.Edges = append(r.Edges, baseRepositoryEdges(h, r.Repositories, creators, &r)...)
	r.Edges = append(r.Edges, transactionalEdges(h, creators, &r)...)
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

// transactionalEdges builds the x3 candidates (TransactionalEntry).
func transactionalEdges(h *hierarchy.Hierarchy, creators map[string][]callgraph.MethodRef, r *Result) []callgraph.SAMEdge {
	if _, ok := h.Lookup(transactionInterceptor); !ok {
		return nil
	}
	gate := append([]callgraph.MethodRef(nil), creators[transactionInterceptor]...)
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if fqcn == transactionInterceptor {
			continue
		}
		for _, named := range ci.NamedClasses {
			if named != transactionInterceptor {
				continue
			}
			for _, m := range ci.Methods {
				gate = append(gate, callgraph.MethodRef{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor})
			}
		}
	}
	gate = dedupe(gate)
	if len(gate) == 0 {
		r.NoCreator = append(r.NoCreator, TransactionalEntry)
		return nil
	}

	annotated := func(anns []string) bool { return hasAnyAnnotation(anns, transactionalAnnotations) }
	// typeTx: types carrying the annotation themselves; methodTx: methods
	// carrying it, by "type#name+desc".
	typeTx := map[string]bool{}
	methodTx := map[string]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if annotated(ci.Annotations) {
			typeTx[fqcn] = true
		}
		for _, m := range ci.Methods {
			if annotated(m.Annotations) {
				methodTx[fqcn+"#"+m.Name+m.Descriptor] = true
			}
		}
	}
	if len(typeTx) == 0 && len(methodTx) == 0 {
		return nil
	}
	// declares: the type or one of its supertypes that declares the
	// method and is annotated (on the method or on itself).
	var declares func(fqcn, name, desc string, seen map[string]bool) bool
	declares = func(fqcn, name, desc string, seen map[string]bool) bool {
		if fqcn == "" || seen[fqcn] {
			return false
		}
		seen[fqcn] = true
		ci, ok := h.Lookup(fqcn)
		if !ok {
			return false
		}
		if m, ok := ci.MethodByNameAndDesc(name, desc); ok {
			if methodTx[fqcn+"#"+name+desc] || typeTx[fqcn] {
				return m.AccessFlags&classfile.AccStatic == 0
			}
		}
		if declares(ci.SuperClass, name, desc, seen) {
			return true
		}
		for _, i := range ci.Interfaces {
			if declares(i, name, desc, seen) {
				return true
			}
		}
		return false
	}
	// transactional: Spring would run fqcn's method name+desc (fqcn the
	// bean's class, or an interface Spring proxies) in a transaction.
	memo := map[string]bool{}
	transactional := func(fqcn, name, desc string) bool {
		key := fqcn + "#" + name + desc
		if v, ok := memo[key]; ok {
			return v
		}
		v := false
		ci, ok := h.Lookup(fqcn)
		if ok {
			impl := fqcn
			if !ci.IsInterface() {
				if d, found := h.ResolveExact(fqcn, name, desc); found {
					impl = d
				}
			}
			if ic, ok := h.Lookup(impl); ok {
				if m, ok := ic.MethodByNameAndDesc(name, desc); ok {
					final := m.AccessFlags&0x0010 != 0
					public := m.AccessFlags&classfile.AccPublic != 0
					v = public && !final && declares(fqcn, name, desc, map[string]bool{})
				}
			}
		}
		memo[key] = v
		return v
	}
	// Owners worth looking at: the annotated types, their subtypes and
	// supertypes.
	relevant := map[string]bool{}
	var up func(string)
	up = func(fqcn string) {
		if fqcn == "" || relevant[fqcn] {
			return
		}
		relevant[fqcn] = true
		if ci, ok := h.Lookup(fqcn); ok {
			up(ci.SuperClass)
			for _, i := range ci.Interfaces {
				up(i)
			}
		}
	}
	seeds := map[string]bool{}
	for t := range typeTx {
		seeds[t] = true
	}
	for k := range methodTx {
		seeds[k[:strings.IndexByte(k, '#')]] = true
	}
	for t := range seeds {
		up(t)
		for _, sub := range h.Subtypes(t) {
			up(sub)
		}
	}

	// callTx: per called owner#name+desc, whether the owner's method is
	// transactional, and up to two concrete subtypes whose method is — a
	// call is transactional when the owner's is, or a concrete subtype's
	// other than the calling class (a self-call does not pass the proxy).
	// Two witnesses decide that for every caller, so the subtypes are
	// walked once per called method, not once per call site.
	type txCall struct {
		owner bool
		subs  []string
	}
	callMemo := map[string]txCall{}
	callTx := func(owner, name, desc, caller string) bool {
		key := owner + "#" + name + desc
		t, ok := callMemo[key]
		if !ok {
			t.owner = transactional(owner, name, desc)
			if !t.owner {
				for _, sub := range h.Subtypes(owner) {
					if sc, ok := h.Lookup(sub); ok && sc.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) == 0 && transactional(sub, name, desc) {
						t.subs = append(t.subs, sub)
						if len(t.subs) == 2 {
							break
						}
					}
				}
			}
			callMemo[key] = t
		}
		if t.owner {
			return true
		}
		for _, sub := range t.subs {
			if sub != caller {
				return true
			}
		}
		return false
	}

	var out []callgraph.SAMEdge
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if (c.Kind != classfile.CallInterface && c.Kind != classfile.CallVirtual) || !relevant[c.Owner] || c.Owner == fqcn {
					continue
				}
				if !callTx(c.Owner, c.Name, c.Desc, fqcn) {
					continue
				}
				out = append(out, callgraph.SAMEdge{
					CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC, Kind: c.Kind,
					TargetFQCN: transactionInterceptor, TargetName: "invoke", TargetDesc: interceptorInvokeDesc,
					Creators: gate,
				})
				r.ByProxy[TransactionalEntry]++
			}
		}
	}
	return out
}

func hasAnyAnnotation(anns, want []string) bool {
	for _, a := range anns {
		for _, w := range want {
			if a == w {
				return true
			}
		}
	}
	return false
}

// baseRepositoryEdges builds the x2 candidates (BaseRepositoryEntry).
func baseRepositoryEdges(h *hierarchy.Hierarchy, repos []string, creators map[string][]callgraph.MethodRef, r *Result) []callgraph.SAMEdge {
	if len(repos) == 0 {
		return nil
	}
	isRepo := map[string]bool{}
	for _, i := range repos {
		isRepo[i] = true
	}
	var bases []string
	for _, fqcn := range h.Subtypes(springDataRepository) {
		ci, ok := h.Lookup(fqcn)
		if ok && ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) == 0 {
			bases = append(bases, fqcn)
		}
	}
	sort.Strings(bases)
	if len(bases) == 0 {
		return nil
	}
	isBase := map[string]bool{}
	for _, b := range bases {
		isBase[b] = true
	}
	baseCreators := map[string][]callgraph.MethodRef{}
	for _, b := range bases {
		baseCreators[b] = append(baseCreators[b], creators[b]...)
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, named := range ci.NamedClasses {
			if !isBase[named] || named == fqcn {
				continue
			}
			for _, m := range ci.Methods {
				baseCreators[named] = append(baseCreators[named], callgraph.MethodRef{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor})
			}
		}
	}
	var out []callgraph.SAMEdge
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind != classfile.CallInterface || !isRepo[c.Owner] || queryOrDefault(h, c.Owner, c.Name, c.Desc) {
					continue
				}
				byDecl := map[string][]callgraph.MethodRef{}
				var decls []string
				for _, b := range bases {
					cr := baseCreators[b]
					if len(cr) == 0 {
						continue
					}
					decl, declDesc, ok := baseMethodFor(h, b, c.Name, c.Desc)
					if !ok {
						continue
					}
					k := decl + "#" + declDesc
					if _, seen := byDecl[k]; !seen {
						decls = append(decls, k)
					}
					byDecl[k] = append(byDecl[k], cr...)
				}
				for _, k := range decls {
					d, dd, _ := strings.Cut(k, "#")
					out = append(out, callgraph.SAMEdge{
						CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC, Kind: c.Kind,
						TargetFQCN: d, TargetName: c.Name, TargetDesc: dd, Creators: dedupe(byDecl[k]),
					})
					r.ByProxy[BaseRepositoryEntry]++
				}
			}
		}
	}
	return out
}

// queryOrDefault: the repository's own declaration of the called method is
// a query method (an annotation named *.Query — Spring Data runs it as a
// query even when a base class has a method of that name) or a default
// method (it runs its own body).
func queryOrDefault(h *hierarchy.Hierarchy, repo, name, desc string) bool {
	decl, ok := h.ResolveExact(repo, name, desc)
	if !ok {
		return false
	}
	ci, _ := h.Lookup(decl)
	m, ok := ci.MethodByNameAndDesc(name, desc)
	if !ok {
		return false
	}
	if !m.IsAbstract() {
		return true
	}
	for _, a := range m.Annotations {
		if strings.HasSuffix(a, ".Query") {
			return true
		}
	}
	return false
}

// baseMethodFor: the body a base repository class runs for a repository
// method, the way Spring Data picks it — the exact descriptor, else a method
// of that name in the class chain with as many parameters, each accepting
// the call's (an erased type variable is Object and accepts any reference);
// the return type is not compared. Bridges and synthetic methods are not
// candidates.
func baseMethodFor(h *hierarchy.Hierarchy, base, name, desc string) (decl, declDesc string, ok bool) {
	if d, ok := h.ResolveExact(base, name, desc); ok {
		dci, _ := h.Lookup(d)
		if dm, ok := dci.MethodByNameAndDesc(name, desc); ok && !dm.IsAbstract() {
			return d, desc, true
		}
	}
	want, err := classfile.ParamDescriptors(desc)
	if err != nil {
		return "", "", false
	}
	for cur := base; cur != ""; {
		ci, found := h.Lookup(cur)
		if !found {
			break
		}
		for _, m := range ci.Methods {
			if m.Name != name || m.IsAbstract() || m.AccessFlags&(accBridge|accSynthetic) != 0 {
				continue
			}
			got, err := classfile.ParamDescriptors(m.Descriptor)
			if err != nil || len(got) != len(want) {
				continue
			}
			match := true
			for i := range got {
				if !accepts(h, got[i], want[i]) {
					match = false
					break
				}
			}
			if match {
				return cur, m.Descriptor, true
			}
		}
		cur = ci.SuperClass
	}
	return "", "", false
}

const (
	accBridge    = 0x0040
	accSynthetic = 0x1000
)

// accepts: a parameter of type param accepts an argument of type arg.
func accepts(h *hierarchy.Hierarchy, param, arg string) bool {
	if param == arg {
		return true
	}
	if !strings.HasPrefix(param, "L") || !(strings.HasPrefix(arg, "L") || strings.HasPrefix(arg, "[")) {
		return false
	}
	if param == "Ljava/lang/Object;" {
		return true
	}
	if !strings.HasPrefix(arg, "L") {
		return false
	}
	target := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(param, "L"), ";"), "/", ".")
	from := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(arg, "L"), ";"), "/", ".")
	seen := map[string]bool{}
	var walk func(string) bool
	walk = func(c string) bool {
		if c == "" || seen[c] {
			return false
		}
		if c == target {
			return true
		}
		seen[c] = true
		ci, ok := h.Lookup(c)
		if !ok {
			return false
		}
		if walk(ci.SuperClass) {
			return true
		}
		for _, i := range ci.Interfaces {
			if walk(i) {
				return true
			}
		}
		return false
	}
	return walk(from)
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
func findMappers(h *hierarchy.Hierarchy, jars []appref.Jar, extraXML map[string][]byte) ([]Mapper, error) {
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
	readXML := func(where string, data []byte) {
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
			readXML(j.Name+"!/"+f.Name, data)
		}
		r.Close()
	}
	extra := make([]string, 0, len(extraXML))
	for where := range extraXML {
		extra = append(extra, where)
	}
	sort.Strings(extra)
	for _, where := range extra {
		readXML(where, extraXML[where])
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
