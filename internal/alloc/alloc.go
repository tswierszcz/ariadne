// Package alloc is GRAPH_PRECISION P1's catalog of the classes the
// analyzed program may create, explicitly (`new`) or through a modeled
// mechanism (a ServiceLoader provider, a resolved reflective
// constructor, a container bean, a deserialization candidate, a class
// whose entry points a framework calls), each with its origin — the
// method (and PC) whose reachability makes the class created — plus the
// gaps: places a class may be created by a mechanism the analysis does
// not model. RTA (P2) narrows virtual dispatch to created classes and
// must treat a gap per the unknown-allocation policy fixed in P0
// (callgraph.PolicyCHAFallbackWithGap), never as "not created".
//
// P1 is a catalog only: it changes no edge and no verdict.
package alloc

import (
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// CatalogVersion versions the origin kinds, gap rules and APIs below.
// p1-2: the servlets/filters the app registers (internal/servletdispatch)
// are container-created types.
const CatalogVersion = "p1-2"

// Origin kinds.
const (
	KindNew             = "new"             // a `new` instruction
	KindService         = "service"         // a ServiceLoader provider edge (serviceloader basis)
	KindReflection      = "reflection"      // Constructor#newInstance with a resolved target
	KindModel           = "model"           // a registered library model's constructor edge (B3; today the ServiceLoader models materializing providers)
	KindContainer       = "container"       // a container bean (Spring stereotype, @Bean return type) or a registered servlet/filter
	KindDeserialization = "deserialization" // a class a deserialization entry may create
	KindEntry           = "entry"           // a class a framework instantiates to call its entry points
)

// Gap kinds.
const (
	GapReflection     = "reflection-unresolved" // newInstance with an unresolved class/constructor
	GapProxy          = "dynamic-proxy"         // Proxy.newProxyInstance: a generated class
	GapUnsafe         = "unsafe-allocation"     // Unsafe.allocateInstance
	GapHandle         = "constructor-handle"    // MethodHandles.Lookup#findConstructor
	GapService        = "service-unresolved"    // a ServiceLoader load site with an unknown service
	GapContainerProxy = "container-proxy"       // a container-generated implementation (repository, AOP proxy)
	GapExternalInput  = "external-input"        // roots=all: parameters of root methods may be any subtype
)

// ContainerOrigin is the container's own pseudo-site: beans exist when
// the container runs, not when a particular method is reached.
const ContainerOrigin = "<container>"

// Site is a place a class is created (or may be created).
type Site struct {
	Method string `json:"method"` // owner#name+desc, or ContainerOrigin
	PC     int    `json:"pc"`     // -1 when not tied to an instruction
}

func (s Site) String() string {
	if s.PC < 0 {
		return s.Method
	}
	return s.Method + "@" + strconv.Itoa(s.PC)
}

// Origin is one way a class is created, with every site.
type Origin struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
	Sites  []Site `json:"sites"`
}

// Created is one class the program may create.
type Created struct {
	Class   string   `json:"class"`
	Origins []Origin `json:"origins"`
}

// Gap is a place a class may be created by an unmodeled mechanism.
type Gap struct {
	Kind string `json:"kind"`
	Site Site   `json:"site"`
	// Bound names the type the created object is known to have (an
	// interface for a proxy or a container-generated implementation), ""
	// when nothing bounds it (any class).
	Bound  string `json:"bound,omitempty"`
	Reason string `json:"reason"`
}

// Lambda is an invokedynamic site whose bootstrap was resolved: it
// creates an instance of a functional interface implemented by Impl.
// Calls on that instance are dispatched by the lambda's own edges (to
// Impl), not by CHA over the interface.
type Lambda struct {
	Site Site   `json:"site"`
	Impl string `json:"impl"`
}

// Catalog is the result.
type Catalog struct {
	Version string
	Types   []Created
	Gaps    []Gap
	Lambdas []Lambda
	// ExternalInputs: roots=all, so every root method's reference
	// parameters may be of any subtype (recorded as one rule, not per
	// method); RootParams counts those parameters.
	ExternalInputs bool
	RootParams     int
}

// ContainerBean is a bean class the container creates.
type ContainerBean struct {
	Class  string
	Detail string // "stereotype", "@Bean <factory method>"
}

// Inputs are the models' facts, gathered by the caller so this package
// depends on none of them.
type Inputs struct {
	H *hierarchy.Hierarchy
	G *callgraph.Graph // the final graph (every model edge added)
	// Container beans; ContainerProxies are interfaces the container
	// implements with a generated class (repositories, AOP proxies).
	Container        []ContainerBean
	ContainerProxies []ContainerBean
	// Deserialization: per format, the candidate classes and the entry
	// sites of that format.
	DeserCandidates map[string][]string
	DeserSites      map[string][]Site
	// Roots are the BFS roots' method keys; Entry marks that they are the
	// profile's entry points (instance methods imply a created class).
	Roots          []string
	EntryRoots     bool
	ExternalInputs bool
	// Unresolved model sites, reported as gaps.
	UnresolvedService    []Site
	UnresolvedReflection []Site
}

// gapCalls: calls creating an object of a class the analysis cannot
// name, by owner#name (any descriptor).
var gapCalls = map[string]struct{ kind, reason string }{
	"java.lang.reflect.Proxy#newProxyInstance":              {GapProxy, "a generated proxy class implementing interfaces given at run time"},
	"sun.misc.Unsafe#allocateInstance":                      {GapUnsafe, "allocates any class without a constructor"},
	"jdk.internal.misc.Unsafe#allocateInstance":             {GapUnsafe, "allocates any class without a constructor"},
	"java.lang.invoke.MethodHandles$Lookup#findConstructor": {GapHandle, "a constructor handle: its class is not tracked by the catalog"},
}

// Build computes the catalog.
func Build(in Inputs) Catalog {
	c := Catalog{Version: CatalogVersion, ExternalInputs: in.ExternalInputs}
	byClass := map[string]map[string]*Origin{} // class -> kind+detail -> origin
	add := func(class, kind, detail string, site Site) {
		if byClass[class] == nil {
			byClass[class] = map[string]*Origin{}
		}
		k := kind + "\x00" + detail
		o := byClass[class][k]
		if o == nil {
			o = &Origin{Kind: kind, Detail: detail}
			byClass[class][k] = o
		}
		o.Sites = append(o.Sites, site)
	}
	concrete := func(class string) bool {
		ci, ok := in.H.Lookup(class)
		return !ok || ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) == 0
	}

	// `new`, lambdas and gap APIs, from every method's bytecode facts.
	for _, fqcn := range in.H.SortedFQCNs() {
		ci, _ := in.H.Lookup(fqcn)
		for _, m := range ci.Methods {
			key := fqcn + "#" + m.Name + m.Descriptor
			for _, ref := range m.InitRefs {
				if ref.Kind == classfile.InitNew && concrete(ref.Owner) {
					add(ref.Owner, KindNew, "", Site{Method: key, PC: ref.PC})
				}
			}
			for _, call := range m.Calls {
				if call.Kind == classfile.CallDynamic {
					if call.RefKind != 0 {
						c.Lambdas = append(c.Lambdas, Lambda{Site: Site{Method: key, PC: call.PC}, Impl: call.Owner + "#" + call.Name + call.Desc})
					}
					continue
				}
				if g, ok := gapCalls[call.Owner+"#"+call.Name]; ok {
					c.Gaps = append(c.Gaps, Gap{Kind: g.kind, Site: Site{Method: key, PC: call.PC}, Reason: g.reason})
				}
			}
		}
	}

	// Model edges into a constructor: the class is created at the edge's
	// source, by that model.
	if in.G != nil {
		kinds := map[callgraph.EdgeBasis]string{callgraph.BasisServiceLoader: KindService, callgraph.BasisReflection: KindReflection, callgraph.BasisModel: KindModel}
		for id := 0; id < in.G.NodeCount(); id++ {
			from := in.G.Key(uint32(id))
			for _, e := range in.G.Edges(uint32(id)) {
				kind, ok := kinds[e.Basis]
				if !ok {
					continue
				}
				owner, rest, _ := strings.Cut(in.G.Key(e.To), "#")
				if !strings.HasPrefix(rest, "<init>(") {
					continue
				}
				if len(e.Sites) == 0 {
					add(owner, kind, "", Site{Method: from, PC: -1})
				}
				for _, s := range e.Sites {
					add(owner, kind, "", Site{Method: from, PC: s.PC})
				}
			}
		}
	}

	for _, b := range in.Container {
		if concrete(b.Class) {
			add(b.Class, KindContainer, b.Detail, Site{Method: ContainerOrigin, PC: -1})
		}
	}
	for _, b := range in.ContainerProxies {
		c.Gaps = append(c.Gaps, Gap{Kind: GapContainerProxy, Site: Site{Method: ContainerOrigin, PC: -1}, Bound: b.Class,
			Reason: "the container implements it with a generated class (" + b.Detail + ")"})
	}
	formats := make([]string, 0, len(in.DeserCandidates))
	for f := range in.DeserCandidates {
		formats = append(formats, f)
	}
	sort.Strings(formats)
	for _, f := range formats {
		sites := in.DeserSites[f]
		if len(sites) == 0 {
			continue // no entry of this format: nothing creates them
		}
		for _, cl := range in.DeserCandidates[f] {
			for _, s := range sites {
				add(cl, KindDeserialization, f, s)
			}
		}
	}
	for _, r := range in.Roots {
		owner, rest, _ := strings.Cut(r, "#")
		name := rest
		if i := strings.IndexByte(rest, '('); i >= 0 {
			name = rest[:i]
		}
		ci, ok := in.H.Lookup(owner)
		if !ok {
			continue
		}
		m, ok := ci.MethodByNameAndDesc(name, rest[len(name):])
		if !ok {
			continue
		}
		if in.ExternalInputs {
			if args, err := classfile.ArgumentSlots(m.Descriptor, true); err == nil {
				for _, a := range args {
					if strings.HasPrefix(a.Type, "L") || strings.HasPrefix(a.Type, "[") {
						c.RootParams++
					}
				}
			}
		}
		if in.EntryRoots && !m.IsStatic() && name != "<init>" && concrete(owner) {
			add(owner, KindEntry, "", Site{Method: r, PC: -1})
		}
	}
	if in.ExternalInputs {
		c.Gaps = append(c.Gaps, Gap{Kind: GapExternalInput, Site: Site{Method: "<roots=all>", PC: -1},
			Reason: "every application method is a root: its reference parameters may be objects of any subtype created outside the analyzed code (" + strconv.Itoa(c.RootParams) + " parameters)"})
	}
	for _, s := range in.UnresolvedService {
		c.Gaps = append(c.Gaps, Gap{Kind: GapService, Site: s, Reason: "ServiceLoader load site whose service class is not resolved: its providers are unknown"})
	}
	for _, s := range in.UnresolvedReflection {
		c.Gaps = append(c.Gaps, Gap{Kind: GapReflection, Site: s, Reason: "Constructor#newInstance whose class or constructor is not resolved"})
	}

	// Deterministic output.
	classes := make([]string, 0, len(byClass))
	for cl := range byClass {
		classes = append(classes, cl)
	}
	sort.Strings(classes)
	for _, cl := range classes {
		ct := Created{Class: cl}
		keys := make([]string, 0, len(byClass[cl]))
		for k := range byClass[cl] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			o := byClass[cl][k]
			sortSites(o.Sites)
			o.Sites = dedupSites(o.Sites)
			ct.Origins = append(ct.Origins, *o)
		}
		c.Types = append(c.Types, ct)
	}
	sort.SliceStable(c.Gaps, func(i, j int) bool {
		a, b := c.Gaps[i], c.Gaps[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Site.String() != b.Site.String() {
			return a.Site.String() < b.Site.String()
		}
		return a.Bound < b.Bound
	})
	sort.SliceStable(c.Lambdas, func(i, j int) bool { return c.Lambdas[i].Site.String() < c.Lambdas[j].Site.String() })
	return c
}

func sortSites(s []Site) {
	sort.Slice(s, func(i, j int) bool {
		if s[i].Method != s[j].Method {
			return s[i].Method < s[j].Method
		}
		return s[i].PC < s[j].PC
	})
}

func dedupSites(s []Site) []Site {
	out := s[:0]
	for i, x := range s {
		if i == 0 || x != s[i-1] {
			out = append(out, x)
		}
	}
	return out
}

// Summary counts the catalog for the report and v2.
type Summary struct {
	Version  string         `json:"catalog"`
	Types    int            `json:"types"`
	ByOrigin map[string]int `json:"types_by_origin"`
	// OnlyModeled counts classes created only by a model (no `new`).
	OnlyModeled int            `json:"types_only_modeled"`
	Lambdas     int            `json:"lambda_sites"`
	Gaps        map[string]int `json:"gaps_by_kind,omitempty"`
	// ExternalInputs/RootParams: see Catalog.
	ExternalInputs bool `json:"external_inputs"`
	RootParams     int  `json:"root_reference_parameters,omitempty"`
	// CHATargets and CHATargetsNotCreated: CHA-basis edges, and of them
	// those whose target's class is in no origin — informational, the
	// most RTA could remove before gaps are applied; not an improvement
	// by itself.
	CHATargets           int `json:"cha_edges"`
	CHATargetsNotCreated int `json:"cha_edges_to_classes_not_created"`
}

// Summarize counts c against g.
func Summarize(c Catalog, g *callgraph.Graph) Summary {
	s := Summary{Version: c.Version, Types: len(c.Types), ByOrigin: map[string]int{}, Lambdas: len(c.Lambdas),
		ExternalInputs: c.ExternalInputs, RootParams: c.RootParams}
	created := map[string]bool{}
	for _, t := range c.Types {
		created[t.Class] = true
		onlyModel := true
		for _, o := range t.Origins {
			s.ByOrigin[o.Kind]++
			if o.Kind == KindNew {
				onlyModel = false
			}
		}
		if onlyModel {
			s.OnlyModeled++
		}
	}
	if len(c.Gaps) > 0 {
		s.Gaps = map[string]int{}
		for _, gp := range c.Gaps {
			s.Gaps[gp.Kind]++
		}
	}
	if g != nil {
		for id := 0; id < g.NodeCount(); id++ {
			for _, e := range g.Edges(uint32(id)) {
				if e.Basis != callgraph.BasisCHA {
					continue
				}
				s.CHATargets++
				owner, _, _ := strings.Cut(g.Key(e.To), "#")
				if !created[owner] {
					s.CHATargetsNotCreated++
				}
			}
		}
	}
	return s
}

func countText(m map[string]int) string {
	if len(m) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + " " + strconv.Itoa(m[k])
	}
	return strings.Join(parts, ", ")
}

// OriginText and GapText render the counts for the report.
func (s Summary) OriginText() string { return countText(s.ByOrigin) }
func (s Summary) GapText() string    { return countText(s.Gaps) }
