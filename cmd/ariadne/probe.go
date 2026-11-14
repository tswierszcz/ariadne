package main

import (
	"fmt"
	"sort"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/reach"
	"ariadne/internal/samcall"
)

// lambdaProbeSites turns the functional-interface calls whose lambda or
// method reference stayed unresolved (samcall gaps) into probe sites for
// ROADMAP B0 b0-4: the candidates are the classes that create a lambda of
// the called interface or of one of its subinterfaces (a resolved
// LambdaMetafactory invokedynamic naming it). Ordinary classes
// implementing the interface are not candidates here — CHA, or the
// skipped-receiver part of the probe, covers them.
func lambdaProbeSites(h *hierarchy.Hierarchy, results []samcall.Result) []reach.GapSite {
	creators := map[string]map[string]bool{} // SAM interface -> creating classes
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind != classfile.CallDynamic || c.SamInterface == "" {
					continue
				}
				if creators[c.SamInterface] == nil {
					creators[c.SamInterface] = map[string]bool{}
				}
				creators[c.SamInterface][fqcn] = true
			}
		}
	}
	ownersOf := map[string][]string{}
	owners := func(iface string) []string {
		if o, ok := ownersOf[iface]; ok {
			return o
		}
		set := map[string]bool{}
		for c := range creators[iface] {
			set[c] = true
		}
		for _, sub := range h.Subtypes(iface) {
			for c := range creators[sub] {
				set[c] = true
			}
		}
		o := make([]string, 0, len(set))
		for c := range set {
			o = append(o, c)
		}
		sort.Strings(o)
		ownersOf[iface] = o
		return o
	}
	var out []reach.GapSite
	for _, r := range results {
		if r.Site.Kind != samcall.KindCall || len(r.Gaps) == 0 {
			continue
		}
		// Owners is never nil: a lambda gap is bounded by the lambdas the
		// input creates (an empty set is relevant to no instance).
		out = append(out, reach.GapSite{Kind: "lambda", CallerFQCN: r.Site.CallerFQCN, CallerName: r.Site.CallerName,
			CallerDesc: r.Site.CallerDesc, Owners: append([]string{}, owners(r.Site.Owner)...)})
	}
	return out
}

// probeLine renders the probe summary for the log.
func probeLine(s *reach.ProbeSummary) string {
	if s == nil {
		return "no NOT_REACHABLE finding"
	}
	out := ""
	for _, v := range reach.ProbeVariants {
		if out != "" {
			out += ", "
		}
		out += fmt.Sprintf("%s %d/%d", v, s.Marked[v], s.NotReachable)
	}
	return out
}
