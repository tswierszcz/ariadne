// Package libapi chooses the roots of library mode (-roots api,
// LIBRARY_MODE LB1, docs/RESULTS_CONTRACT.MD §8): the methods a user of
// the library can call or override from outside its packages.
package libapi

import (
	"archive/zip"
	"ariadne/internal/component"
	"io"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Version versions the API definition and the client-code sites;
// snapshots record it. lb2-1: ClientSites (LIBRARY_MODE LB2); lb2-2: a
// class a client can construct is not a client-dispatch stand-in
// (Constructible, measured on commons-configuration2); lb2-3: methods an
// API class inherits from a dependency supertype are roots (Inherited);
// lb2-4: each inherited root is specialized to the receiver class
// (Specialize); lb2-5: classes the library announces to Spring are
// entries (AnnouncedRoots); lb3-1: the library's dependency declarations
// and who supplies each dependency (Coverage.Dependencies, read by
// component.LibraryDeclarations); lb2-6: the beans of a CDI bean archive
// (META-INF/beans.xml, Jandex index) are entries (CDIAnnouncements);
// lb2-7: calls on objects the library's caller or container
// supplies are marked (SuppliedSites, exposed_api[].supplied_dispatch);
// lb2-8: a library announcing classes to Spring gets the spring profile's
// framework packages (Coverage.FrameworkAdded).
const Version = "lb2-8"

const (
	accPrivate   = 0x0002
	accProtected = 0x0004
	accStatic    = 0x0008
	accBridge    = 0x0040
	accSynthetic = 0x1000
)

// Options narrow the API.
type Options struct {
	// Exclude are package prefixes the user calls internal.
	Exclude []string
	// ModulePath: the library runs on the module path (-jpms-scenario
	// module-path), so a module-info's unqualified exports decide which
	// packages are API, and its provides which providers are roots.
	ModulePath bool
}

// Provider is a service provider the library registers.
type Provider struct {
	Service, Class string
}

// Coverage is what the API selection counted (v2 coverage.api).
type Coverage struct {
	Classes int `json:"classes"`
	Methods int `json:"methods"`
	// Inherited counts the roots that are dependency methods an API class
	// inherits (lb2-3); they are part of Methods.
	Inherited int `json:"inherited_methods"`
	// Specialized is what receiver specialization of the inherited
	// roots did (lb2-4); set by the caller after Specialize.
	Specialized *SpecializeStats `json:"receiver_specialization,omitempty"`
	// Supplied (lb2-7): calls on caller-supplied objects
	// (SuppliedSites) and the analysis behind them.
	Supplied *SuppliedCoverage `json:"supplied,omitempty"`
	// FrameworkAdded (lb2-8): framework package patterns of the spring
	// profile added because the library announces classes to Spring.
	FrameworkAdded []string `json:"framework_packages_added,omitempty"`
	// Announced counts the classes the library announces to a framework
	// (lb2-5: spring.factories, META-INF/spring/*.imports, and what they
	// pull in), whose entries are roots; Announcements lists the
	// announcements read.
	Announced     int            `json:"announced_classes"`
	Announcements []Announcement `json:"announcements,omitempty"`
	// Dependencies (lb3-1) are the library pom's declarations, matched to
	// the input's instances.
	Dependencies *component.LibraryDependencies `json:"dependencies,omitempty"`
	Providers    int                            `json:"providers"`
	Excluded     []string                       `json:"excluded_packages,omitempty"`
	JPMS         string                         `json:"jpms"` // "exports" or "classpath"
	Exports      []string                       `json:"exports,omitempty"`
}

// Module reads the library's module-info (root or a multi-release
// version entry); nil when it has none.
func Module(appJar string) (*classfile.ModuleInfo, error) {
	r, err := zip.OpenReader(appJar)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var best *zip.File
	for _, f := range r.File {
		if f.Name == "module-info.class" || (strings.HasPrefix(f.Name, "META-INF/versions/") && strings.HasSuffix(f.Name, "/module-info.class")) {
			if best == nil || f.Name == "module-info.class" {
				best = f
			}
		}
	}
	if best == nil {
		return nil, nil
	}
	rc, err := best.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	cf, err := classfile.Parse(io.LimitReader(rc, 4<<20))
	if err != nil {
		return nil, err
	}
	return cf.Module, nil
}

// selector holds the API's access rules (contract §8), shared by Roots
// and ClientSites.
type selector struct {
	h         *hierarchy.Hierarchy
	exclude   []string
	exported  map[string]bool // nil: every package (classpath)
	providers []Provider
	cov       Coverage
	memo      map[string]bool
}

func newSelector(h *hierarchy.Hierarchy, mod *classfile.ModuleInfo, providers []Provider, opt Options) *selector {
	s := &selector{h: h, exclude: opt.Exclude, providers: providers, memo: map[string]bool{},
		cov: Coverage{JPMS: "classpath", Excluded: append([]string(nil), opt.Exclude...)}}
	if opt.ModulePath && mod != nil {
		s.cov.JPMS = "exports"
		s.exported = map[string]bool{}
		for _, e := range mod.Exports {
			if len(e.To) == 0 {
				s.exported[e.Package] = true
				s.cov.Exports = append(s.cov.Exports, e.Package)
			}
		}
		sort.Strings(s.cov.Exports)
		s.providers = nil
		for _, p := range mod.Provides {
			for _, w := range p.With {
				s.providers = append(s.providers, Provider{Service: p.Service, Class: w})
			}
		}
	}
	return s
}

// inScope: the class's package is neither excluded nor unexported.
func (s *selector) inScope(fqcn string) bool {
	pkg := ""
	if i := strings.LastIndexByte(fqcn, '.'); i >= 0 {
		pkg = fqcn[:i]
	}
	for _, ex := range s.exclude {
		if pkg == ex || strings.HasPrefix(pkg, ex+".") {
			return false
		}
	}
	return s.exported == nil || s.exported[pkg]
}

// accessible: a public library class whose enclosing classes are all
// accessible. A member class (Outer$Inner) is accessible only when every
// enclosing class is; its own public flag also covers a protected member
// (compiled public).
func (s *selector) accessible(fqcn string) bool {
	if v, ok := s.memo[fqcn]; ok {
		return v
	}
	ci, ok := s.h.Lookup(fqcn)
	v := ok && ci.IsApp && ci.AccessFlags&classfile.AccPublic != 0
	if v {
		if i := strings.LastIndexByte(fqcn, '$'); i > 0 {
			if _, outer := s.h.Lookup(fqcn[:i]); outer {
				v = s.accessible(fqcn[:i])
			}
		}
	}
	s.memo[fqcn] = v
	return v
}

// Roots returns the API methods of the application input (the library)
// as graph nodes, sorted by key, and what was counted. providers are the
// library's own META-INF/services registrations; under ModulePath with a
// module-info its provides are used instead.
//
// inherited maps each root that is a dependency method an accessible API
// class inherits (public, or protected in a non-final class; an instance
// method with a body; resolved from that class to the dependency's
// declaration, so not overridden on the way) to those API classes, sorted.
func Roots(g *callgraph.Graph, h *hierarchy.Hierarchy, mod *classfile.ModuleInfo, providers []Provider, opt Options) (roots []uint32, cov Coverage, inherited map[uint32][]string) {
	sel := newSelector(h, mod, providers, opt)
	cov = sel.cov
	providers = sel.providers
	inScope, accessible := sel.inScope, sel.accessible

	seen := map[uint32]bool{}
	add := func(fqcn, name, desc string) bool {
		id, ok := g.ID(fqcn, name, desc)
		if !ok || seen[id] {
			return false
		}
		seen[id] = true
		roots = append(roots, id)
		return true
	}
	for _, fqcn := range h.SortedFQCNs() {
		if !accessible(fqcn) || !inScope(fqcn) {
			continue
		}
		ci, _ := h.Lookup(fqcn)
		cov.Classes++
		final := ci.AccessFlags&classfile.AccFinal != 0
		for _, m := range ci.Methods {
			f := m.AccessFlags
			if f&(accSynthetic|accBridge) != 0 {
				continue
			}
			if m.Name == "<clinit>" || f&classfile.AccPublic != 0 || (f&accProtected != 0 && !final) {
				add(fqcn, m.Name, m.Descriptor)
			}
		}
	}
	for _, p := range providers {
		ci, ok := h.Lookup(p.Class)
		if !ok || !ci.IsApp || !inScope(p.Class) {
			continue
		}
		cov.Providers++
		// ServiceLoader needs a public class and a public no-argument
		// constructor (or a provider() method on the module path); its
		// public methods are API once it is accessible.
		add(p.Class, "<init>", "()V")
		for _, m := range ci.Methods {
			if m.AccessFlags&classfile.AccPublic != 0 && m.AccessFlags&(accSynthetic|accBridge) == 0 {
				add(p.Class, m.Name, m.Descriptor)
			}
			if m.Name == "<clinit>" {
				add(p.Class, m.Name, m.Descriptor)
			}
		}
	}
	inherited = map[uint32][]string{}
	for _, fqcn := range h.SortedFQCNs() {
		if !accessible(fqcn) || !inScope(fqcn) {
			continue
		}
		ci, _ := h.Lookup(fqcn)
		final := ci.AccessFlags&classfile.AccFinal != 0
		for _, sup := range dependencySupertypes(h, fqcn) {
			sci, _ := h.Lookup(sup)
			for _, m := range sci.Methods {
				f := m.AccessFlags
				if m.Name == "<init>" || m.Name == "<clinit>" || f&(accStatic|accSynthetic|accBridge|accPrivate|classfile.AccAbstract) != 0 {
					continue
				}
				if f&classfile.AccPublic == 0 && (f&accProtected == 0 || final) {
					continue
				}
				if decl, ok := h.ResolveExact(fqcn, m.Name, m.Descriptor); !ok || decl != sup {
					continue // overridden closer to the API class
				}
				id, ok := g.ID(sup, m.Name, m.Descriptor)
				if !ok {
					continue
				}
				if !seen[id] {
					seen[id] = true
					roots = append(roots, id)
					cov.Inherited++
				}
				inherited[id] = append(inherited[id], fqcn) // fqcn ascending: sorted
			}
		}
	}
	sort.Slice(roots, func(i, j int) bool { return g.Key(roots[i]) < g.Key(roots[j]) })
	cov.Methods = len(roots)
	return roots, cov, inherited
}

// dependencySupertypes lists fqcn's transitive supertypes (superclasses
// and interfaces) that are dependency classes, in breadth-first order.
func dependencySupertypes(h *hierarchy.Hierarchy, fqcn string) []string {
	seen := map[string]bool{fqcn: true}
	queue := []string{fqcn}
	var out []string
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		ci, ok := h.Lookup(cur)
		if !ok {
			continue
		}
		if cur != fqcn && !ci.IsApp {
			out = append(out, cur)
		}
		next := append([]string{ci.SuperClass}, ci.Interfaces...)
		for _, s := range next {
			if s != "" && !seen[s] {
				seen[s] = true
				queue = append(queue, s)
			}
		}
	}
	return out
}

// Constructible reports whether a client can get an object of a library
// class without the library creating it: an accessible, in-scope
// interface (a client implements it, inheriting its default bodies), a
// class with a public constructor, or a non-final class with a protected
// one (a client subclass inherits its bodies).
func Constructible(h *hierarchy.Hierarchy, mod *classfile.ModuleInfo, opt Options) func(fqcn string) bool {
	sel := newSelector(h, mod, nil, opt)
	return func(fqcn string) bool {
		ci, ok := h.Lookup(fqcn)
		if !ok || !sel.accessible(fqcn) || !sel.inScope(fqcn) {
			return false
		}
		if ci.IsInterface() {
			return true
		}
		for _, m := range ci.Methods {
			if m.Name != "<init>" {
				continue
			}
			if m.AccessFlags&classfile.AccPublic != 0 || (m.AccessFlags&accProtected != 0 && ci.AccessFlags&classfile.AccFinal == 0) {
				return true
			}
		}
		return false
	}
}

// ClientSite is a call the library makes on a method a client of the
// library can implement or override (LIBRARY_MODE LB2,
// docs/RESULTS_CONTRACT.MD §8 "Client code"): the code that runs there
// may be the client's, which the input does not have.
type ClientSite struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Declared is the resolved method the call names (owner#name+desc):
	// its own body, when it has one, is the library's default; any other
	// dispatch target stands in for a client implementation.
	Declared string
	// Abstract: the declared method is abstract (an interface or abstract
	// method), so the client's implementation is the only kind of code a
	// client-created object runs there.
	Abstract bool
}

// ClientSites lists, in deterministic order, the virtual and interface
// calls in the library's own methods whose resolved method a client can
// implement or override: declared in an accessible, in-scope library
// type the client can subtype (an interface, or a class that is not final
// and has a public or protected constructor), with public or protected
// access, not static and not final. Calls into dependency or JDK types
// are out of scope (other gap mechanisms cover callbacks there).
func ClientSites(h *hierarchy.Hierarchy, mod *classfile.ModuleInfo, opt Options) []ClientSite {
	sel := newSelector(h, mod, nil, opt)
	subtypable := map[string]bool{}
	canSubtype := func(fqcn string) bool {
		if v, ok := subtypable[fqcn]; ok {
			return v
		}
		ci, ok := h.Lookup(fqcn)
		v := ok && sel.accessible(fqcn) && sel.inScope(fqcn)
		if v && !ci.IsInterface() {
			v = ci.AccessFlags&classfile.AccFinal == 0
			ctor := false
			for _, m := range ci.Methods {
				if m.Name == "<init>" && m.AccessFlags&(classfile.AccPublic|accProtected) != 0 {
					ctor = true
					break
				}
			}
			v = v && ctor
		}
		subtypable[fqcn] = v
		return v
	}
	var out []ClientSite
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind != classfile.CallVirtual && c.Kind != classfile.CallInterface {
					continue
				}
				decl, ok := h.ResolveExact(c.Owner, c.Name, c.Desc)
				if !ok || !canSubtype(decl) {
					continue
				}
				dci, _ := h.Lookup(decl)
				dm, _ := dci.MethodByNameAndDesc(c.Name, c.Desc)
				f := dm.AccessFlags
				if f&(classfile.AccPublic|accProtected) == 0 || f&(accPrivate|accStatic|classfile.AccFinal) != 0 || dm.Name == "<init>" {
					continue
				}
				out = append(out, ClientSite{CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC,
					Declared: decl + "#" + c.Name + c.Desc, Abstract: f&classfile.AccAbstract != 0})
			}
		}
	}
	return out
}
