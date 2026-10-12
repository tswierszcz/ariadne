package hierarchy

// ResolveMethod finds which class in fqcn's superclass/superinterface
// chain actually declares a method named name, following the JVM's own
// lookup order: fqcn itself, then up the superclass chain, and only if
// that finds nothing, the interfaces reachable from any class in that
// chain (breadth-first, so a closer interface wins over a farther one).
//
// Matching is by name only, ignoring descriptor/overloads — the same
// conservative, overload-insensitive approach the plan uses throughout,
// since bytecode call sites, CVE advisories and vuln_methods.yaml entries
// rarely agree on an exact descriptor for the same logical method.
func (h *Hierarchy) ResolveMethod(fqcn, name string) (declaringClass string, found bool) {
	chain := h.superclassChain(fqcn)
	for _, c := range chain {
		ci, ok := h.classes[c]
		if !ok {
			continue
		}
		if _, ok := ci.MethodByName(name); ok {
			return c, true
		}
	}

	visited := make(map[string]bool)
	var queue []string
	for _, c := range chain {
		ci, ok := h.classes[c]
		if !ok {
			continue
		}
		for _, iface := range ci.Interfaces {
			if !visited[iface] {
				visited[iface] = true
				queue = append(queue, iface)
			}
		}
	}
	for len(queue) > 0 {
		iface := queue[0]
		queue = queue[1:]
		ci, ok := h.classes[iface]
		if !ok {
			continue
		}
		if _, ok := ci.MethodByName(name); ok {
			return iface, true
		}
		for _, parent := range ci.Interfaces {
			if !visited[parent] {
				visited[parent] = true
				queue = append(queue, parent)
			}
		}
	}
	return "", false
}

// superclassChain returns fqcn and every superclass above it, in order,
// stopping at java.lang.Object or at the first class ariadne never read
// (a JDK type outside the app+dependency universe).
func (h *Hierarchy) superclassChain(fqcn string) []string {
	var chain []string
	seen := make(map[string]bool)
	cur := fqcn
	for cur != "" && !seen[cur] {
		seen[cur] = true
		chain = append(chain, cur)
		ci, ok := h.classes[cur]
		if !ok {
			break
		}
		cur = ci.SuperClass
	}
	return chain
}
