package main

import (
	"sort"
	"strings"

	"ariadne/internal/alloc"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/reach"
)

// rtaInputs translates GRAPH_PRECISION P1's catalog into the facts of
// P2's fixed point: every origin site becomes a creation at that method
// (the container's pseudo-site and a site not in the graph: live from
// the start), every gap an unknown creation with its fallback classes.
func rtaInputs(h *hierarchy.Hierarchy, g *callgraph.Graph, c alloc.Catalog, roots []uint32, providers []string) callgraph.RTAInputs {
	in := callgraph.RTAInputs{Roots: roots}
	siteID := func(s alloc.Site) (uint32, bool) {
		owner, rest, ok := strings.Cut(s.Method, "#")
		if !ok {
			return 0, false
		}
		i := strings.IndexByte(rest, '(')
		if i < 0 {
			return 0, false
		}
		return g.ID(owner, rest[:i], rest[i:])
	}
	always := []string{}
	bySite := map[uint32][]string{}
	for _, t := range c.Types {
		for _, o := range t.Origins {
			for _, s := range o.Sites {
				if id, ok := siteID(s); ok {
					bySite[id] = append(bySite[id], t.Class)
				} else {
					always = append(always, t.Class)
				}
			}
		}
	}
	if len(always) > 0 {
		in.Creations = append(in.Creations, callgraph.Creation{Always: true, Classes: always})
	}
	ids := make([]uint32, 0, len(bySite))
	for id := range bySite {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		in.Creations = append(in.Creations, callgraph.Creation{Site: id, Classes: bySite[id]})
	}

	var nameable []string
	for _, gp := range c.Gaps {
		u := callgraph.UnknownCreation{Kind: gp.Kind}
		switch gp.Kind {
		case alloc.GapReflection, alloc.GapUnsafe, alloc.GapHandle:
			if nameable == nil {
				nameable = nameableClasses(h)
			}
			u.Fallback = nameable
		case alloc.GapService:
			u.Fallback = providers
		case alloc.GapExternalInput:
			u.Fallback = rootParameterClasses(h, g, roots)
		case alloc.GapProxy, alloc.GapContainerProxy:
			// A generated class: no class of the input is created. A
			// container proxy delegates to the bean (live as a container
			// origin) or is backed by synthetic edges.
		}
		if id, ok := siteID(gp.Site); ok {
			u.Site = id
		} else {
			u.Always = true
		}
		in.Unknown = append(in.Unknown, u)
	}
	return in
}

// nameableClasses: the concrete classes reflection could create without
// the analysis resolving the call — those the code can name at run time
// (a class literal or a class-name string constant anywhere in the
// input), classes carrying a class-level annotation (what frameworks
// scan for: entities, components, DTOs), and classes used as a field,
// parameter or return type (what binders, ORMs and serializers
// instantiate from member types). Not covered: a class named only in a
// resource or configuration file, or only in a generic signature.
func nameableClasses(h *hierarchy.Hierarchy) []string {
	named := map[string]bool{}
	addDesc := func(desc string) {
		for i := 0; i < len(desc); i++ {
			if desc[i] != 'L' {
				continue
			}
			j := strings.IndexByte(desc[i:], ';')
			if j < 0 {
				return
			}
			named[strings.ReplaceAll(desc[i+1:i+j], "/", ".")] = true
			i += j
		}
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, n := range ci.NamedClasses {
			named[n] = true
		}
		for _, a := range ci.Annotations {
			if a != "kotlin.Metadata" {
				named[fqcn] = true
				break
			}
		}
		for _, f := range ci.Fields {
			addDesc(f.Descriptor)
		}
		for _, m := range ci.Methods {
			addDesc(m.Descriptor)
		}
	}
	var out []string
	for n := range named {
		if concreteClass(h, n) {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// rootParameterClasses: with -roots all, a root's reference parameter
// may be any concrete subtype of its declared type (external input).
func rootParameterClasses(h *hierarchy.Hierarchy, g *callgraph.Graph, roots []uint32) []string {
	types := map[string]bool{}
	for _, r := range roots {
		_, rest, _ := strings.Cut(g.Key(r), "#")
		i := strings.IndexByte(rest, '(')
		if i < 0 {
			continue
		}
		args, err := classfile.ArgumentSlots(rest[i:], true)
		if err != nil {
			continue
		}
		for _, a := range args {
			t := strings.TrimLeft(a.Type, "[")
			if strings.HasPrefix(t, "L") && strings.HasSuffix(t, ";") {
				types[strings.ReplaceAll(t[1:len(t)-1], "/", ".")] = true
			}
		}
	}
	set := map[string]bool{}
	for t := range types {
		if concreteClass(h, t) {
			set[t] = true
		}
		for _, s := range h.Subtypes(t) {
			if concreteClass(h, s) {
				set[s] = true
			}
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func concreteClass(h *hierarchy.Hierarchy, fqcn string) bool {
	ci, ok := h.Lookup(fqcn)
	return ok && ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) == 0
}

// rtaGapSites: in RTA mode, reachable creation sites of a mechanism no
// other gap rule covers (Unsafe, constructor handles) are unbounded model
// gaps too (docs/RESULTS_CONTRACT.MD §4), since the fallback classes are
// a bound the analysis chose, not one the code proves.
func rtaGapSites(c alloc.Catalog) []reach.GapSite {
	var out []reach.GapSite
	for _, gp := range c.Gaps {
		if gp.Kind != alloc.GapUnsafe && gp.Kind != alloc.GapHandle {
			continue
		}
		owner, rest, ok := strings.Cut(gp.Site.Method, "#")
		i := strings.IndexByte(rest, '(')
		if !ok || i < 0 {
			continue
		}
		out = append(out, reach.GapSite{Kind: "allocation", CallerFQCN: owner, CallerName: rest[:i], CallerDesc: rest[i:]})
	}
	return out
}
