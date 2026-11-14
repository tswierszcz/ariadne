// Package spidispatch models calls on a service-provider interface the
// call graph never CHA-expands (a java./javax./jdk./sun. receiver,
// internal/callgraph.DefaultSkip) to the providers the input registers for
// that interface in META-INF/services (SERVICE_LOADING inventory),
// registered as a ROADMAP B3 model.
//
// The motivating case came from a runtime trace (RUNTIME_NATIVE N1):
// Hibernate Validator's AbstractConfigurationImpl#buildValidatorFactory
// calls provider.buildValidatorFactory(this) on a
// javax.validation.spi.ValidationProvider — a skipped receiver, so the
// graph stopped there while the JVM ran HibernateValidator#
// buildValidatorFactory, the provider hibernate-validator registers in
// META-INF/services/javax.validation.spi.ValidationProvider. The same
// shape covers JDBC drivers (java.sql.Driver#connect), and any other
// JDK/javax SPI with registered providers.
//
// A provider-based API hands out more than the provider: Bean Validation's
// provider creates the Configuration, which builds the ValidatorFactory,
// which hands out Validators; a JDBC Driver returns Connections, which
// return Statements and ResultSets. All are interfaces of the same skipped
// API, implemented in the provider's library. So the model also expands a
// call on a *product* interface — one reached from the service through
// return types of its methods, transitively, within the same API family
// (the service's package without a trailing ".spi": javax.validation,
// java.sql) — to the implementations of that interface found in the jars
// of the registered providers. An interface the service does not hand out
// (a ServletContainerInitializer returns nothing) has no products.
//
// Lifting the skip for these interfaces would fan out to every
// implementation on the class path; this model expands exactly the calls
// on a registered service interface (to its registered providers) and on
// another interface of its API family (to implementations in the
// providers' own jars). An implementation from elsewhere (constructed
// directly by other code, a module "provides") stays unexpanded, as
// before. The registration shows the provider may be loaded by
// ServiceLoader, not that this call site receives it: the edges are "may
// run", like CHA edges.
package spidispatch

import (
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/jdkapi"
	"ariadne/internal/models"
)

// ModelID is the registry identifier of this model.
const ModelID = "service-provider:dispatch"

// Version versions the model's rules (snapshots record it). p2: facades
// (DriverManager#getConnection → Driver#connect). p3: a product's
// implementations include those of its JDK subinterfaces (jdkapi).
const Version = "p3"

// Facade is a static JDK method that runs a registered service's method
// itself: the input calls the facade, never the service, so without it the
// providers are unreachable (ROADMAP B4 b4-3 measured this for JDBC: a
// driver reached only through DriverManager was invisible). Every
// overload of Owner#Name is a facade; the call reaches Service#Method
// (Desc) of every registered provider of Service.
type Facade struct {
	Owner, Name           string
	Service, Method, Desc string
}

// Facades are the known facades. DriverManager#getConnection asks each
// registered driver to connect (JDBC 4: drivers are loaded through
// ServiceLoader and register themselves).
var Facades = []Facade{
	{Owner: "java.sql.DriverManager", Name: "getConnection", Service: "java.sql.Driver", Method: "connect",
		Desc: "(Ljava/lang/String;Ljava/util/Properties;)Ljava/sql/Connection;"},
}

// facadeOf returns the facade a static call names, if any.
func facadeOf(owner, name string) (Facade, bool) {
	for _, f := range Facades {
		if f.Owner == owner && f.Name == name {
			return f, true
		}
	}
	return Facade{}, false
}

// Registration is one META-INF/services entry the model may dispatch to.
type Registration struct {
	Service, Provider, Source string
}

// Catalog is what the model knows: per skipped service interface, its
// registered providers present in the input, and the call signatures in
// the input whose receiver is such an interface.
type Catalog struct {
	Providers map[string][]string // service -> providers (sorted)
	// Families: API family root (service package without ".spi") -> the
	// jars of its registered providers.
	Families   map[string]map[string]bool
	Signatures []models.Signature
	products   map[string][]string // product interface -> implementations in provider jars
	// Products: product interface -> the service it is reached from.
	Products map[string]string
	// Skipped counts registrations left out: service not skipped by CHA
	// (already expanded), or provider not in the input / not a subtype.
	NotSkipped, ProviderMissing int
}

// Find builds the catalog from the registrations, given the CHA skip rule.
func Find(h *hierarchy.Hierarchy, regs []Registration, skip func(string) bool) Catalog {
	c := Catalog{Providers: map[string][]string{}, Families: map[string]map[string]bool{}, products: map[string][]string{}, Products: map[string]string{}}
	seen := map[[2]string]bool{}
	for _, r := range regs {
		if !skip(r.Service) {
			c.NotSkipped++
			continue
		}
		ci, ok := h.Lookup(r.Provider)
		if !ok || ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) != 0 || !isSubtype(h, r.Provider, r.Service) {
			c.ProviderMissing++
			continue
		}
		k := [2]string{r.Service, r.Provider}
		if seen[k] {
			continue
		}
		seen[k] = true
		c.Providers[r.Service] = append(c.Providers[r.Service], r.Provider)
		root := familyRoot(r.Service)
		if c.Families[root] == nil {
			c.Families[root] = map[string]bool{}
		}
		c.Families[root][ci.SourceJar] = true
	}
	for s := range c.Providers {
		sort.Strings(c.Providers[s])
	}
	// Return types per skipped interface: from its declared methods when it
	// is in the input, and from the calls on it the input makes.
	returns := map[string]map[string]bool{}
	addRet := func(owner, desc string) {
		i := strings.LastIndexByte(desc, ')')
		if i < 0 || !strings.HasPrefix(desc[i+1:], "L") || !strings.HasSuffix(desc, ";") {
			return
		}
		ret := strings.ReplaceAll(desc[i+2:len(desc)-1], "/", ".")
		if returns[owner] == nil {
			returns[owner] = map[string]bool{}
		}
		returns[owner][ret] = true
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if skip(fqcn) {
			for _, m := range ci.Methods {
				addRet(fqcn, m.Descriptor)
			}
		}
		for _, m := range ci.Methods {
			for _, call := range m.Calls {
				if skip(call.Owner) {
					addRet(call.Owner, call.Desc)
				}
				// A facade call is a call on its service's method.
				if f, ok := facadeOf(call.Owner, call.Name); ok && call.Kind == classfile.CallStatic {
					addRet(f.Service, f.Desc)
				}
			}
		}
	}
	services := make([]string, 0, len(c.Providers))
	for s := range c.Providers {
		services = append(services, s)
	}
	sort.Strings(services)
	for _, s := range services {
		root := familyRoot(s) + "."
		queue := []string{s}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			rets := make([]string, 0, len(returns[cur]))
			for r := range returns[cur] {
				rets = append(rets, r)
			}
			sort.Strings(rets)
			for _, r := range rets {
				if !strings.HasPrefix(r, root) || !skip(r) {
					continue
				}
				if _, isService := c.Providers[r]; isService {
					continue
				}
				if _, done := c.Products[r]; done {
					continue
				}
				c.Products[r] = s
				queue = append(queue, r)
			}
		}
	}

	sigs := map[models.Signature]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, call := range m.Calls {
				if f, ok := facadeOf(call.Owner, call.Name); ok && call.Kind == classfile.CallStatic {
					if _, ok := c.Providers[f.Service]; ok {
						sigs[models.Signature{Owner: call.Owner, Name: call.Name, Desc: call.Desc}] = true
					}
					continue
				}
				if call.Kind != classfile.CallInterface && call.Kind != classfile.CallVirtual {
					continue
				}
				if _, ok := c.Providers[call.Owner]; ok || len(c.productsOf(h, call.Owner, skip)) > 0 {
					sigs[models.Signature{Owner: call.Owner, Name: call.Name, Desc: call.Desc}] = true
				}
			}
		}
	}
	for s := range sigs {
		c.Signatures = append(c.Signatures, s)
	}
	sort.Slice(c.Signatures, func(i, j int) bool {
		a, b := c.Signatures[i], c.Signatures[j]
		if a.Owner != b.Owner {
			return a.Owner < b.Owner
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Desc < b.Desc
	})
	return c
}

// familyRoot: the API family of a service — its package, without a
// trailing ".spi" segment.
func familyRoot(service string) string {
	pkg := service
	if i := strings.LastIndexByte(pkg, '.'); i >= 0 {
		pkg = pkg[:i]
	}
	return strings.TrimSuffix(pkg, ".spi")
}

// productsOf: for a skipped interface of a registered API family that is
// not itself a registered service, the concrete implementations in the
// family's provider jars (memoized; nil when none or not in a family).
func (c *Catalog) productsOf(h *hierarchy.Hierarchy, iface string, skip func(string) bool) []string {
	if impls, ok := c.products[iface]; ok {
		return impls
	}
	var impls []string
	if service, isProduct := c.Products[iface]; isProduct {
		jars := c.Families[familyRoot(service)]
		{
			for _, sub := range jdkapi.Subtypes(h, iface) {
				ci, ok := h.Lookup(sub)
				if ok && jars[ci.SourceJar] && ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) == 0 {
					impls = append(impls, sub)
				}
			}
		}
		sort.Strings(impls)
	}
	c.products[iface] = impls
	return impls
}

// isSubtype: the provider is the service or one of its transitive subtypes
// (a class implementing an interface of the input, or of the JDK when the
// service itself is not in the input: then any declared superinterface
// chain naming it counts).
func isSubtype(h *hierarchy.Hierarchy, cls, target string) bool {
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
	return walk(cls)
}

// Register adds the model for the catalog's signatures.
func Register(registry *models.Registry, c Catalog) {
	if len(c.Signatures) == 0 {
		return
	}
	registry.Register(&models.Model{
		ID:         ModelID,
		Source:     "calls on a JDK/javax service-provider interface (never CHA-expanded) dispatched to the providers registered for it in META-INF/services",
		Signatures: c.Signatures,
	})
}

// Evaluator dispatches a matched call to the registered providers.
type Evaluator struct {
	h *hierarchy.Hierarchy
	c Catalog
}

// NewEvaluator returns an evaluator over the catalog.
func NewEvaluator(h *hierarchy.Hierarchy, c Catalog) *Evaluator { return &Evaluator{h: h, c: c} }

// Evaluate implements models.Evaluator.
func (e *Evaluator) Evaluate(m *models.Model, site models.CallSite) models.Fact {
	fact := models.Fact{ModelID: m.ID}
	if f, ok := facadeOf(site.Call.Owner, site.Call.Name); ok {
		if kind, ok := e.callKind(site); ok && kind == classfile.CallStatic {
			fact.Targets = e.targets(e.c.Providers[f.Service], f.Service, f.Method, f.Desc, classfile.CallInterface)
		}
		return fact
	}
	providers, ok := e.c.Providers[site.Call.Owner]
	if !ok {
		providers = e.c.products[site.Call.Owner] // computed by Find for every matched signature
	}
	if len(providers) == 0 {
		return fact
	}
	kind, ok := e.callKind(site)
	if !ok || (kind != classfile.CallVirtual && kind != classfile.CallInterface) {
		return fact
	}
	fact.Targets = e.targets(providers, site.Call.Owner, site.Call.Name, site.Call.Desc, kind)
	return fact
}

// targets: the body each provider runs for owner#name+desc, once per
// declaring class, sorted.
func (e *Evaluator) targets(providers []string, owner, name, desc string, kind classfile.CallKind) []models.Target {
	var out []models.Target
	seen := map[string]bool{}
	for _, p := range providers {
		decl, ok := e.h.ResolveExact(p, name, desc)
		if !ok || decl == owner || seen[decl] {
			continue
		}
		ci, _ := e.h.Lookup(decl)
		if meth, ok := ci.MethodByNameAndDesc(name, desc); !ok || meth.IsAbstract() {
			continue
		}
		seen[decl] = true
		out = append(out, models.Target{Method: models.Signature{Owner: decl, Name: name, Desc: desc}, Kind: kind})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Method.Owner < out[j].Method.Owner })
	return out
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
