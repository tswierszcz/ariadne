// Package libapi chooses the roots of library mode (-roots api,
// LIBRARY_MODE LB1, docs/RESULTS_CONTRACT.MD §8): the methods a user of
// the library can call or override from outside its packages.
package libapi

import (
	"archive/zip"
	"io"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Version versions the API definition; snapshots record it.
const Version = "lb1-1"

const (
	accProtected = 0x0004
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
	Classes   int      `json:"classes"`
	Methods   int      `json:"methods"`
	Providers int      `json:"providers"`
	Excluded  []string `json:"excluded_packages,omitempty"`
	JPMS      string   `json:"jpms"` // "exports" or "classpath"
	Exports   []string `json:"exports,omitempty"`
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

// Roots returns the API methods of the application input (the library)
// as graph nodes, sorted by key, and what was counted. providers are the
// library's own META-INF/services registrations; under ModulePath with a
// module-info its provides are used instead.
func Roots(g *callgraph.Graph, h *hierarchy.Hierarchy, mod *classfile.ModuleInfo, providers []Provider, opt Options) ([]uint32, Coverage) {
	cov := Coverage{JPMS: "classpath", Excluded: append([]string(nil), opt.Exclude...)}
	var exported map[string]bool
	if opt.ModulePath && mod != nil {
		cov.JPMS = "exports"
		exported = map[string]bool{}
		for _, e := range mod.Exports {
			if len(e.To) == 0 {
				exported[e.Package] = true
				cov.Exports = append(cov.Exports, e.Package)
			}
		}
		sort.Strings(cov.Exports)
		providers = nil
		for _, p := range mod.Provides {
			for _, w := range p.With {
				providers = append(providers, Provider{Service: p.Service, Class: w})
			}
		}
	}
	inScope := func(fqcn string) bool {
		pkg := ""
		if i := strings.LastIndexByte(fqcn, '.'); i >= 0 {
			pkg = fqcn[:i]
		}
		for _, ex := range opt.Exclude {
			if pkg == ex || strings.HasPrefix(pkg, ex+".") {
				return false
			}
		}
		return exported == nil || exported[pkg]
	}
	memo := map[string]bool{}
	var accessible func(fqcn string) bool
	accessible = func(fqcn string) bool {
		if v, ok := memo[fqcn]; ok {
			return v
		}
		ci, ok := h.Lookup(fqcn)
		v := ok && ci.IsApp && ci.AccessFlags&classfile.AccPublic != 0
		if v {
			// A member class (Outer$Inner) is accessible only when every
			// enclosing class is; its own public flag also covers a
			// protected member (compiled public).
			if i := strings.LastIndexByte(fqcn, '$'); i > 0 {
				if _, outer := h.Lookup(fqcn[:i]); outer {
					v = accessible(fqcn[:i])
				}
			}
		}
		memo[fqcn] = v
		return v
	}

	seen := map[uint32]bool{}
	var roots []uint32
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
	sort.Slice(roots, func(i, j int) bool { return g.Key(roots[i]) < g.Key(roots[j]) })
	cov.Methods = len(roots)
	return roots, cov
}
