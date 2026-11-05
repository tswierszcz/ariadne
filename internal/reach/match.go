package reach

import (
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/vulnmap"
)

// matchVulnerableMethod resolves one selector (vulnmap.Entry.Methods) to
// every graph node that represents a call to it. A name-only selector
// ("FQCN#method" — what fix-commit diffs and LLM answers give, and the
// legacy manual table) is overload-insensitive; it additionally follows
// inheritance: a call written against a subclass that does not itself
// override the method (so the call dispatches, at run time, to the
// vulnerable class's own implementation) matches too. Before G2,
// classfile.ExtractCalls's literal Methodref owner (whatever static type
// the original Java source used at the call site, not necessarily the
// type that actually defines the method) meant the call graph itself
// could intern such a call under the subclass's name with no outgoing
// edges of its own, and this inheritance check was the only thing that
// still found it; G2 fixed that connectivity gap at the source for calls
// whose declaring class is part of the analyzed universe, but this check
// stays as a second line of defense — e.g. for a declaring class that
// Build itself couldn't resolve to (outside the analyzed universe
// entirely), where Owner is still whatever the bytecode happened to name.
//
// CVE:F4: a selector may carry a descriptor. "Class#name(params)ret"
// matches exactly that JVM method, "Class#name(params)" that parameter
// list with any return type; a subtype overrides only when it declares
// the same descriptor (or parameter list). Only "Class#name" keeps the
// overload-insensitive behavior. An unparsable selector matches nothing.
func matchVulnerableMethod(g *callgraph.Graph, h *hierarchy.Hierarchy, spec string) []uint32 {
	sel, err := vulnmap.ParseSelector(spec)
	if err != nil {
		return nil
	}
	keep := func(id uint32) bool {
		if sel.Params == "" {
			return true
		}
		_, rest, _ := strings.Cut(g.Key(id), "#")
		desc := rest[strings.IndexByte(rest, '('):]
		if sel.Return != "" {
			return desc == sel.Params+sel.Return
		}
		return strings.HasPrefix(desc, sel.Params)
	}
	overrides := func(ci *hierarchy.ClassInfo) bool {
		for _, m := range ci.MethodsByName(sel.Name) {
			switch {
			case sel.Params == "":
				return true
			case sel.Return != "" && m.Descriptor == sel.Params+sel.Return:
				return true
			case sel.Return == "" && strings.HasPrefix(m.Descriptor, sel.Params):
				return true
			}
		}
		return false
	}

	var ids []uint32
	add := func(fqcn string) {
		for _, id := range g.IDsByName(fqcn, sel.Name) {
			if keep(id) {
				ids = append(ids, id)
			}
		}
	}
	add(sel.Class)
	for _, sub := range h.Subtypes(sel.Class) {
		if subInfo, ok := h.Lookup(sub); ok && overrides(subInfo) {
			continue // the subtype overrides it: a different method, not the vulnerable one
		}
		add(sub)
	}
	return dedupeIDs(ids)
}

func dedupeIDs(ids []uint32) []uint32 {
	if len(ids) < 2 {
		return ids
	}
	seen := make(map[uint32]bool, len(ids))
	out := ids[:0]
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}
