package callgraph

import (
	"sort"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// ObjectEdges returns, for every virtual/interface call whose owner is one
// of types (receivers CHA's skip rule never expands), a candidate edge to
// the body each concrete implementation in the input runs for it
// (hierarchy.ResolveExact), with the methods that create that
// implementation (`new`) as its creators. The candidates are meant for a
// creator gate (NewCreatorGate): an implementation counts only once a
// method creating it is reachable from the roots, so the many
// implementations a class path carries but the reachable code never
// creates are left out. An implementation with no `new` in the input
// (created reflectively or outside it) is not a candidate. Shared by the
// library models whose receivers are created objects:
// internal/servletdispatch (request/response/chain) and
// internal/jdbcdispatch (connections, statements, result sets). The result
// is sorted.
func ObjectEdges(h *hierarchy.Hierarchy, types []string) []SAMEdge {
	creators := map[string][]MethodRef{}
	type siteCall struct {
		caller hierarchy.Method
		owner  string
		call   classfile.Call
	}
	var sites []siteCall
	isObjectType := map[string]bool{}
	for _, t := range types {
		isObjectType[t] = true
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, r := range m.InitRefs {
				if r.Kind == classfile.InitNew {
					creators[r.Owner] = append(creators[r.Owner], MethodRef{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor})
				}
			}
			for _, c := range m.Calls {
				if (c.Kind == classfile.CallInterface || c.Kind == classfile.CallVirtual) && isObjectType[c.Owner] {
					sites = append(sites, siteCall{m, fqcn, c})
				}
			}
		}
	}
	type key struct{ owner, name, desc string }
	done := map[key]bool{}
	type body struct{ decl, class string }
	bodies := map[key][]body{}
	for _, s := range sites {
		k := key{s.call.Owner, s.call.Name, s.call.Desc}
		if done[k] {
			continue
		}
		done[k] = true
		for _, sub := range h.Subtypes(s.call.Owner) {
			ci, ok := h.Lookup(sub)
			if !ok || ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) != 0 || len(creators[sub]) == 0 {
				continue
			}
			decl, ok := h.ResolveExact(sub, s.call.Name, s.call.Desc)
			if !ok || decl == s.call.Owner {
				continue
			}
			dci, _ := h.Lookup(decl)
			if m, ok := dci.MethodByNameAndDesc(s.call.Name, s.call.Desc); !ok || m.IsAbstract() {
				continue
			}
			bodies[k] = append(bodies[k], body{decl, sub})
		}
	}
	var out []SAMEdge
	for _, s := range sites {
		k := key{s.call.Owner, s.call.Name, s.call.Desc}
		byDecl := map[string][]MethodRef{}
		var decls []string
		for _, b := range bodies[k] {
			if _, ok := byDecl[b.decl]; !ok {
				decls = append(decls, b.decl)
			}
			byDecl[b.decl] = append(byDecl[b.decl], creators[b.class]...)
		}
		for _, d := range decls {
			out = append(out, SAMEdge{CallerFQCN: s.owner, CallerName: s.caller.Name, CallerDesc: s.caller.Descriptor, PC: s.call.PC,
				Kind: s.call.Kind, TargetFQCN: d, TargetName: s.call.Name, TargetDesc: s.call.Desc, Creators: byDecl[d]})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.CallerFQCN != b.CallerFQCN {
			return a.CallerFQCN < b.CallerFQCN
		}
		if a.CallerName+a.CallerDesc != b.CallerName+b.CallerDesc {
			return a.CallerName+a.CallerDesc < b.CallerName+b.CallerDesc
		}
		if a.PC != b.PC {
			return a.PC < b.PC
		}
		return a.TargetFQCN < b.TargetFQCN
	})
	return out
}
