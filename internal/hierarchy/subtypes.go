package hierarchy

// Subtypes returns every transitive subtype of fqcn (every class that
// directly or indirectly extends it, or — when fqcn is an interface —
// implements it, directly or through another interface). The result is
// memoized, since CHA expansion (internal/callgraph) calls this once per
// distinct virtual/interface call target, and the same target is called
// from many call sites.
//
// A class ariadne never read (typically a JDK type, since the hierarchy
// only contains the app's own classes and its declared dependency jars)
// has no known subtypes and returns an empty slice, not an error: CHA
// expansion over such a type is exactly what section 3.6's skip list
// exists to avoid in the first place.
func (h *Hierarchy) Subtypes(fqcn string) []string {
	if cached, ok := h.subtypesCache[fqcn]; ok {
		return cached
	}

	visited := map[string]bool{fqcn: true}
	var out []string
	queue := append([]string(nil), h.direct[fqcn]...)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true
		out = append(out, cur)
		queue = append(queue, h.direct[cur]...)
	}

	h.subtypesCache[fqcn] = out
	return out
}
