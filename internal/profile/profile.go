// Package profile supplies the framework-specific knowledge PLAN.md
// section 4 describes: which classes mark an app as using a given
// framework (auto-detection), which dependency packages count as
// "framework" code for the REACHABLE/FRAMEWORK split (internal/reach),
// which methods are entry points when --roots profile replaces the
// default "every application method", and synthetic call-graph edges
// (internal/callgraph) for proxies the bytecode has no implementation of
// at all — notably Spring Data repositories.
//
// A profile is plain YAML (the shape PLAN.md section 4's own example
// uses), with two shipped by default (spring, generic), embedded with
// //go:embed and overridable/extendable with -profile-file. Nothing here
// talks to the network or touches global state; Select is the one
// entry point cmd/ariadne calls once the hierarchy is built.
package profile

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

//go:embed profiles/spring.yaml
var springYAML []byte

//go:embed profiles/generic.yaml
var genericYAML []byte

// SyntheticEdge mirrors PLAN.md section 4's synthetic_edges entry; see
// callgraph.SyntheticEdge, which this is converted to 1:1 (profile does
// not import callgraph's own type to keep the YAML shape and the
// callgraph-internal representation free to evolve independently).
type SyntheticEdge struct {
	FromSubtypesOf string `yaml:"from_subtypes_of"`
	To             string `yaml:"to"`
}

// FrameworkPackage is one framework_packages entry: a "groupId:artifact"
// glob pattern (artifact "*" matches any), as a plain YAML scalar, or —
// PLAN.md's own jackson-databind example — a mapping naming the pattern
// under "pkg" plus "when_annotation", a list of annotation FQCNs that
// gates the pattern: it only applies when the analyzed app actually
// carries at least one of them somewhere.
//
// That condition is evaluated once for the whole app (see
// ResolvePatterns), not per call site: a precise per-call-site version
// (jackson-databind only counts as framework-driven on *this specific*
// path if *this specific* caller has the annotation) would need
// internal/reach's BFS sink test to see the calling node as well as the
// callee, which is a much larger change to bfs.go than this gains over
// the simpler whole-app check — an app that uses @RequestBody anywhere
// is, in practice, exactly the kind of app where Jackson deserializes
// attacker-controlled JSON somewhere, which is what the condition is
// really trying to capture.
type FrameworkPackage struct {
	Pattern        string
	WhenAnnotation []string
}

// UnmarshalYAML accepts either a bare string ("org.hibernate:*") or a
// mapping ({pkg: ..., when_annotation: [...]}), matching PLAN.md section
// 4's own heterogeneous framework_packages example.
func (p *FrameworkPackage) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		return value.Decode(&p.Pattern)
	}
	var m struct {
		Pkg            string   `yaml:"pkg"`
		WhenAnnotation []string `yaml:"when_annotation"`
	}
	if err := value.Decode(&m); err != nil {
		return err
	}
	p.Pattern = m.Pkg
	p.WhenAnnotation = m.WhenAnnotation
	return nil
}

// Roots configures --roots profile: the entry-point set used instead of
// reach's own default of every application method.
type Roots struct {
	// Annotations: a method carrying one of these is a root by itself;
	// a class carrying one makes every one of its public methods a root
	// (PLAN.md section 4: "method-level, or class-level = all public
	// methods").
	Annotations []string `yaml:"annotations"`
	// Overrides is accepted (so a profile YAML matching PLAN.md's own
	// shape parses) but not yet matched against anything: no profile
	// shipped today needs it (spring.yaml's own list is empty, exactly
	// PLAN.md's own example), and "overrides method X of class Y" needs
	// hierarchy.ResolveMethod wired into root selection in a way nothing
	// has needed before now. Left for the servlet/jakarta profiles
	// PLAN.md's Phase 7 would add (HttpServlet#doGet, Filter#doFilter).
	Overrides []string `yaml:"overrides"`
	// Main: every public static "main(String[])" method is also a root.
	Main bool `yaml:"main"`
}

// Profile is one framework's knowledge, loaded from YAML.
type Profile struct {
	Name string `yaml:"name"`
	// Detect: auto-detection selects this profile when at least one of
	// these classes is present anywhere in the analyzed app or its
	// dependencies.
	Detect            []string           `yaml:"detect"`
	Roots             Roots              `yaml:"roots"`
	FrameworkPackages []FrameworkPackage `yaml:"framework_packages"`
	SyntheticEdges    []SyntheticEdge    `yaml:"synthetic_edges"`
}

// Catalog is the set of profiles -profile can select by name or
// auto-detect among.
type Catalog struct {
	profiles map[string]Profile
	// order is insertion order (default catalog order, then any names
	// AddFile appended), so auto-detection's merge order — and therefore
	// its merged Name string — is deterministic rather than at the mercy
	// of Go's randomized map iteration.
	order []string
}

func parseProfile(data []byte) (Profile, error) {
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Profile{}, err
	}
	if p.Name == "" {
		return Profile{}, fmt.Errorf(`profile has no "name"`)
	}
	return p, nil
}

func (c *Catalog) put(p Profile) {
	if c.profiles == nil {
		c.profiles = map[string]Profile{}
	}
	if _, exists := c.profiles[p.Name]; !exists {
		c.order = append(c.order, p.Name)
	}
	c.profiles[p.Name] = p
}

// LoadDefaultCatalog parses ariadne's two embedded profiles (spring,
// generic). generic must always be present: it is auto-detection's
// fallback when nothing else matches.
func LoadDefaultCatalog() (Catalog, error) {
	var c Catalog
	for _, raw := range [][]byte{springYAML, genericYAML} {
		p, err := parseProfile(raw)
		if err != nil {
			return Catalog{}, fmt.Errorf("parsing embedded profile: %w", err)
		}
		c.put(p)
	}
	return c, nil
}

// AddFile parses one profile from a YAML file (the same per-profile
// shape as the embedded defaults — not a list) and adds it to the
// catalog, replacing any existing profile of the same name. This is
// -profile-file: PLAN.md section 4's "overridable with --profile-file".
func (c *Catalog) AddFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	p, err := parseProfile(data)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}
	c.put(p)
	return nil
}

// Get returns the named profile, if the catalog has one.
func (c Catalog) Get(name string) (Profile, bool) {
	p, ok := c.profiles[name]
	return p, ok
}

// Names lists every profile in the catalog, in a fixed order.
func (c Catalog) Names() []string {
	return append([]string(nil), c.order...)
}

// Select picks the active profile for a run.
//
//   - name == "" or "auto" (the default): every catalog profile other
//     than "generic" whose Detect list names a class present in h is
//     merged together (PLAN.md section 4: "if several profiles match,
//     they are merged"); none match -> the catalog's own "generic"
//     fallback, unmodified.
//   - any other name: exactly that profile from the catalog (an error if
//     it isn't there).
func Select(catalog Catalog, name string, h *hierarchy.Hierarchy) (Profile, error) {
	if name != "" && name != "auto" {
		p, ok := catalog.Get(name)
		if !ok {
			return Profile{}, fmt.Errorf("unknown profile %q (known: %s)", name, strings.Join(catalog.Names(), ", "))
		}
		return p, nil
	}

	var matched []Profile
	for _, pname := range catalog.order {
		if pname == "generic" {
			continue // generic is the fallback, never auto-detected into
		}
		p := catalog.profiles[pname]
		if classPresent(h, p.Detect) {
			matched = append(matched, p)
		}
	}
	if len(matched) == 0 {
		generic, ok := catalog.Get("generic")
		if !ok {
			return Profile{}, fmt.Errorf(`catalog has no "generic" fallback profile`)
		}
		return generic, nil
	}
	return merge(matched), nil
}

func classPresent(h *hierarchy.Hierarchy, classes []string) bool {
	for _, fqcn := range classes {
		if _, ok := h.Lookup(fqcn); ok {
			return true
		}
	}
	return false
}

func merge(profiles []Profile) Profile {
	var m Profile
	var names []string
	for _, p := range profiles {
		names = append(names, p.Name)
		m.Detect = append(m.Detect, p.Detect...)
		m.Roots.Annotations = append(m.Roots.Annotations, p.Roots.Annotations...)
		m.Roots.Overrides = append(m.Roots.Overrides, p.Roots.Overrides...)
		m.Roots.Main = m.Roots.Main || p.Roots.Main
		m.FrameworkPackages = append(m.FrameworkPackages, p.FrameworkPackages...)
		m.SyntheticEdges = append(m.SyntheticEdges, p.SyntheticEdges...)
	}
	m.Name = strings.Join(names, "+")
	return m
}

// ResolvePatterns turns p's FrameworkPackages into the plain
// "groupId:artifact" pattern list reach.NewFrameworkMatcher takes,
// evaluating each pattern's WhenAnnotation condition (if any) once
// against the whole analyzed app (see FrameworkPackage's own comment for
// why per-call-site evaluation isn't done).
func (p Profile) ResolvePatterns(h *hierarchy.Hierarchy) []string {
	var out []string
	for _, fp := range p.FrameworkPackages {
		if len(fp.WhenAnnotation) == 0 || appHasAnyAnnotation(h, fp.WhenAnnotation) {
			out = append(out, fp.Pattern)
		}
	}
	return out
}

func appHasAnyAnnotation(h *hierarchy.Hierarchy, needles []string) bool {
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		if hasAny(ci.Annotations, needles) {
			return true
		}
		for _, m := range ci.Methods {
			if hasAny(m.Annotations, needles) {
				return true
			}
		}
	}
	return false
}

func hasAny(haystack, needles []string) bool {
	for _, n := range needles {
		for _, hv := range haystack {
			if hv == n {
				return true
			}
		}
	}
	return false
}

// mainDescriptor is the JVM descriptor of "void main(String[])".
const mainDescriptor = "([Ljava/lang/String;)V"

func isMainMethod(m hierarchy.Method) bool {
	return m.Name == "main" && m.IsStatic() && m.IsPublic() && m.Descriptor == mainDescriptor
}

// AppRoots returns the root method nodes for --roots profile (PLAN.md
// section 4 point 4), instead of reach's own default of every
// application method: every method carrying one of Roots.Annotations,
// every public method of a class itself carrying one of them, and, when
// Roots.Main is true, every public static main(String[]).
func (p Profile) AppRoots(g *callgraph.Graph, h *hierarchy.Hierarchy) []uint32 {
	var roots []uint32
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		classAnnotated := hasAny(ci.Annotations, p.Roots.Annotations)
		for _, m := range ci.Methods {
			include := (classAnnotated && m.IsPublic()) || hasAny(m.Annotations, p.Roots.Annotations)
			if !include && p.Roots.Main && isMainMethod(m) {
				include = true
			}
			if !include {
				continue
			}
			if id, ok := g.ID(fqcn, m.Name, m.Descriptor); ok {
				roots = append(roots, id)
			}
		}
	}
	return roots
}
