package main

import (
	"sort"
	"strconv"

	"ariadne/internal/alloc"
	"ariadne/internal/callgraph"
	"ariadne/internal/deserialize"
	"ariadne/internal/hierarchy"
	"ariadne/internal/profile"
	"ariadne/internal/reflectinvoke"
	"ariadne/internal/serviceload"
)

// allocInputs gathers the models' facts for GRAPH_PRECISION P1's catalog.
func allocInputs(h *hierarchy.Hierarchy, g *callgraph.Graph, cfg config, roots []uint32,
	beans, beanMethods, repositories []profile.BeanName,
	deserJava, deserXStream map[string][]deserialize.Target, deserSites []deserialize.Site,
	loadSites []serviceload.ServiceLoadSite, newInstanceSites []reflectinvoke.NewInstanceSite,
	newInstanceTargets []reflectinvoke.NewInstanceEdge) alloc.Inputs {
	in := alloc.Inputs{H: h, G: g, EntryRoots: cfg.roots == "profile", ExternalInputs: cfg.roots == "all",
		DeserCandidates: map[string][]string{}, DeserSites: map[string][]alloc.Site{}}
	for _, b := range beans {
		in.Container = append(in.Container, alloc.ContainerBean{Class: b.FQCN, Detail: "stereotype"})
	}
	for _, b := range beanMethods {
		in.Container = append(in.Container, alloc.ContainerBean{Class: b.FQCN, Detail: "@Bean " + b.Name})
	}
	for _, b := range repositories {
		in.ContainerProxies = append(in.ContainerProxies, alloc.ContainerBean{Class: b.FQCN, Detail: "Spring Data repository"})
	}
	for format, byClass := range map[string]map[string][]deserialize.Target{deserialize.FormatJava: deserJava, deserialize.FormatXStream: deserXStream} {
		for cl := range byClass {
			in.DeserCandidates[format] = append(in.DeserCandidates[format], cl)
		}
		sort.Strings(in.DeserCandidates[format])
	}
	for _, s := range deserSites {
		if s.Format == "" {
			continue
		}
		in.DeserSites[s.Format] = append(in.DeserSites[s.Format], alloc.Site{Method: s.CallerFQCN + "#" + s.CallerName + s.CallerDesc, PC: s.PC})
	}
	for _, r := range roots {
		in.Roots = append(in.Roots, g.Key(r))
	}
	sort.Strings(in.Roots)
	for _, s := range loadSites {
		if s.Interface == "" {
			in.UnresolvedService = append(in.UnresolvedService, alloc.Site{Method: s.CallerFQCN + "#" + s.CallerName + s.CallerDesc, PC: s.PC})
		}
	}
	resolved := map[string]bool{}
	for _, t := range newInstanceTargets {
		resolved[t.CallerFQCN+"#"+t.CallerName+t.CallerDesc+"@"+strconv.Itoa(t.PC)] = true
	}
	for _, s := range newInstanceSites {
		site := alloc.Site{Method: s.CallerFQCN + "#" + s.CallerName + s.CallerDesc, PC: s.PC}
		if !resolved[site.String()] {
			in.UnresolvedReflection = append(in.UnresolvedReflection, site)
		}
	}
	return in
}

// allocExport is the catalog's part of -graph-out: per class, its origin
// kinds and site counts (sites themselves stay in memory: the file would
// otherwise repeat every deserialization entry per candidate class).
func allocExport(c alloc.Catalog) []callgraph.TypeRecord {
	var out []callgraph.TypeRecord
	for _, t := range c.Types {
		r := callgraph.TypeRecord{Class: t.Class}
		for _, o := range t.Origins {
			k := o.Kind
			if o.Detail != "" {
				k += ":" + o.Detail
			}
			r.Origins = append(r.Origins, k+"="+strconv.Itoa(len(o.Sites)))
		}
		out = append(out, r)
	}
	return out
}
