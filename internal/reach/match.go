package reach

import (
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// matchVulnerableMethod resolves one "FQCN#method" string (as
// vulnmap.Entry.Methods gives it, no descriptor) to every graph node that
// represents a call to it. This is overload-insensitive, like everywhere
// else in ariadne, and additionally follows inheritance: a call written
// against a subclass that does not itself override the method (so the
// call dispatches, at run time, to the vulnerable class's own
// implementation) matches too, even though the call graph necessarily
// interned it under the subclass's name — classfile.ExtractCalls reports
// a Methodref's literal declared owner, which is whatever static type the
// original Java source used at the call site, not the type that actually
// defines the method.
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
