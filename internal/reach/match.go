package reach

import (
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// matchVulnerableMethod resolves one "FQCN#method" string (as
// vulnmap.Entry.Methods gives it, no descriptor) to every graph node that
// represents a call to it. This is overload-insensitive by necessity —
// vulnmap sources (the manual table, fix-commit diffs, LLM answers) never
// carry a descriptor, so there is no exact alternative to fall back to
// here the way internal/callgraph.Build now has via
// hierarchy.ResolveExact (JVM_GRAPH G2) — and additionally follows
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
func matchVulnerableMethod(g *callgraph.Graph, h *hierarchy.Hierarchy, spec string) []uint32 {
	fqcn, name, ok := strings.Cut(spec, "#")
	if !ok {
		return nil
	}

	ids := g.IDsByName(fqcn, name)
	for _, sub := range h.Subtypes(fqcn) {
		if subInfo, ok := h.Lookup(sub); ok {
			if _, declared := subInfo.MethodByName(name); declared {
				continue // the subtype overrides it: a different method, not the vulnerable one
			}
		}
		ids = append(ids, g.IDsByName(sub, name)...)
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
