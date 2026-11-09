package servletdispatch

import (
	"sort"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// objectTypes are the Servlet API types whose instances the container and
// the registered filters pass along a request: the request, the response,
// their wrappers and the filter chain. A call on one of them is a
// javax./jakarta. receiver CHA never expands (callgraph.DefaultSkip), so
// a filter's chain.doFilter(...) or a controller's request.getHeader(...)
// keeps only its edge to the interface method and never reaches the
// container's or a wrapper's body. jakarta.* is not under the skip rule —
// CHA already expands those calls — so only javax.servlet is listed.
var objectTypes = func() []string {
	var out []string
	for _, p := range []string{"javax.servlet."} {
		for _, t := range []string{"FilterChain", "ServletRequest", "ServletResponse", "ServletRequestWrapper", "ServletResponseWrapper",
			"http.HttpServletRequest", "http.HttpServletResponse", "http.HttpServletRequestWrapper", "http.HttpServletResponseWrapper"} {
			out = append(out, p+t)
		}
	}
	return out
}()

// ObjectEdges returns, for every virtual/interface call on a request,
// response, wrapper or filter-chain type (objectTypes), a candidate edge to
// the body each concrete implementation in the input runs for it
// (hierarchy.ResolveExact), with the methods that create that
// implementation (`new`) as its creators. The candidates go through a
// callgraph creator gate: an implementation counts only once a method
// creating it is reachable — the container's own request/response and
// chain classes, and the wrappers of the filters that run, not the
// wrappers of filters nobody registers (dozens on a Spring/Tomcat class
// path). The result is sorted.
func ObjectEdges(h *hierarchy.Hierarchy) []callgraph.SAMEdge {
	creators := map[string][]callgraph.MethodRef{}
	type siteCall struct {
		caller hierarchy.Method
		owner  string
		call   classfile.Call
	}
	var sites []siteCall
	isObjectType := map[string]bool{}
	for _, t := range objectTypes {
		isObjectType[t] = true
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, r := range m.InitRefs {
				if r.Kind == classfile.InitNew {
					creators[r.Owner] = append(creators[r.Owner], callgraph.MethodRef{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor})
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
	var out []callgraph.SAMEdge
	for _, s := range sites {
		k := key{s.call.Owner, s.call.Name, s.call.Desc}
		byDecl := map[string][]callgraph.MethodRef{}
		var decls []string
		for _, b := range bodies[k] {
			if _, ok := byDecl[b.decl]; !ok {
				decls = append(decls, b.decl)
			}
			byDecl[b.decl] = append(byDecl[b.decl], creators[b.class]...)
		}
		for _, d := range decls {
			out = append(out, callgraph.SAMEdge{CallerFQCN: s.owner, CallerName: s.caller.Name, CallerDesc: s.caller.Descriptor, PC: s.call.PC,
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
