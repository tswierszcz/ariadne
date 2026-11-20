package libapi

import (
	"archive/zip"
	"bufio"
	"io"
	"path"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// Micronaut bean definitions (LIBRARY_MODE LB2, lb2-9): Micronaut's
// annotation processor writes, next to each bean, a generated
// $X$Definition class (package-private, synthetic) whose build,
// injectBean and initialize methods — and $X$Definition$Exec#dispatch for
// executable methods such as @EventListener — call the bean's constructor,
// factory, @Inject, @PostConstruct and listener methods directly, whatever
// their access (measured 2026-11-20, testdata/library-micronaut). The
// framework finds the definitions through the jar's own list of
// BeanDefinitionReference classes and calls them from outside the library,
// so they are announcements: the library methods the generated code calls
// are entries. Private bean methods are called through reflection instead
// (no edge): they are not found.

const (
	// Micronaut 3.x/4.x: one empty entry per reference class.
	micronautRefDir = "META-INF/micronaut/io.micronaut.inject.BeanDefinitionReference/"
	// Micronaut 1.x/2.x: a ServiceLoader file listing the reference classes.
	micronautServices = "META-INF/services/io.micronaut.inject.BeanDefinitionReference"
	// MicronautKey is the Key of a Micronaut announcement.
	MicronautKey = "io.micronaut.inject.BeanDefinitionReference"
)

// MicronautAnnouncements lists the bean definitions a library jar
// announces to Micronaut (File: the listing entry — the reference's own
// entry under META-INF/micronaut/, or the services file; Key: MicronautKey;
// Class: the definition class — the reference class without its
// "$Reference" suffix, or the listed class itself when it has none).
// Sorted by class.
func MicronautAnnouncements(jar string) ([]Announcement, error) {
	zr, err := zip.OpenReader(jar)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	seen := map[string]bool{}
	var out []Announcement
	addRef := func(file, ref string) {
		def := strings.TrimSuffix(strings.TrimSpace(ref), "$Reference")
		if def == "" || seen[def] {
			return
		}
		seen[def] = true
		out = append(out, Announcement{File: file, Key: MicronautKey, Class: def})
	}
	for _, f := range zr.File {
		switch {
		case strings.HasPrefix(f.Name, micronautRefDir) && !strings.HasSuffix(f.Name, "/"):
			addRef(f.Name, path.Base(f.Name))
		case f.Name == micronautServices:
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
			sc := bufio.NewScanner(strings.NewReader(string(b)))
			for sc.Scan() {
				line := sc.Text()
				if i := strings.IndexByte(line, '#'); i >= 0 {
					line = line[:i]
				}
				if line = strings.TrimSpace(line); line != "" {
					addRef(micronautServices, line)
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Class < out[j].Class })
	return out, nil
}

// isMicronaut reports whether an announcement is a Micronaut definition.
func (a Announcement) isMicronaut() bool { return a.Key == MicronautKey }

// micronautTargets returns, for a definition class, the library methods
// its generated code calls directly: every direct edge out of a method of
// the definition or of a class nested in it ($Exec), into a class of the
// input that is not itself generated (synthetic) and not a superclass of
// the generated class (Micronaut's own base definitions, when the library
// is Micronaut itself). Sorted by key.
func micronautTargets(g *callgraph.Graph, h *hierarchy.Hierarchy, def string) []uint32 {
	var out []uint32
	seen := map[uint32]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		if fqcn != def && !strings.HasPrefix(fqcn, def+"$") {
			continue
		}
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp {
			continue
		}
		ancestors := map[string]bool{}
		for sup := ci.SuperClass; sup != "" && !ancestors[sup]; {
			ancestors[sup] = true
			si, ok := h.Lookup(sup)
			if !ok {
				break
			}
			sup = si.SuperClass
		}
		for _, m := range ci.Methods {
			id, ok := g.ID(fqcn, m.Name, m.Descriptor)
			if !ok {
				continue
			}
			for _, e := range g.Edges(id) {
				if e.Basis != callgraph.BasisDirect || seen[e.To] {
					continue
				}
				owner := g.Key(e.To)
				if i := strings.IndexByte(owner, '#'); i >= 0 {
					owner = owner[:i]
				}
				oi, ok := h.Lookup(owner)
				if !ok || !oi.IsApp || oi.AccessFlags&accSynthetic != 0 || owner == def || strings.HasPrefix(owner, def+"$") || ancestors[owner] {
					continue
				}
				seen[e.To] = true
				out = append(out, e.To)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return g.Key(out[i]) < g.Key(out[j]) })
	return out
}
