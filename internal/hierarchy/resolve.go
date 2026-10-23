package hierarchy

// resolveBy walks fqcn's method-resolution order (JVMS 5.4.3.3/5.4.3.4):
// fqcn itself, then up the superclass chain, then — only if that finds
// nothing — every superinterface reachable from any class in that chain,
// breadth-first so a closer interface wins over a farther one. match
// decides whether a given class/interface is the one being looked for.
func (h *Hierarchy) resolveBy(fqcn string, match func(*ClassInfo) bool) (declaringClass string, found bool) {
	chain := h.superclassChain(fqcn)
	for _, c := range chain {
		if ci, ok := h.classes[c]; ok && match(ci) {
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
		if match(ci) {
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

// ResolveMethod finds which class in fqcn's superclass/superinterface
// chain actually declares a method named name, following the JVM's own
// lookup order (see resolveBy).
//
// Matching is by name only, ignoring descriptor/overloads — the same
// conservative, overload-insensitive approach the plan uses throughout,
// since bytecode call sites, CVE advisories and vuln_methods.yaml entries
// rarely agree on an exact descriptor for the same logical method. Kept
// as an explicit legacy/approximate mode (JVM_GRAPH G2) for exactly that
// use — today, only LLM-suggested-method grounding in cmd/ariadne uses
// it, since advisories and vulnmap entries never carry a descriptor
// either.
func (h *Hierarchy) ResolveMethod(fqcn, name string) (declaringClass string, found bool) {
	return h.resolveBy(fqcn, func(ci *ClassInfo) bool {
		_, ok := ci.MethodByName(name)
		return ok
	})
}

// ResolveExact finds which class in fqcn's superclass/superinterface
// chain actually declares a method with this exact name and descriptor,
// following the same JVM lookup order ResolveMethod uses, but without
// ResolveMethod's overload blindness (JVM_GRAPH G2). internal/callgraph
// uses this to lead a call whose declared receiver type doesn't itself
// implement the method — because it only inherits it — to the class
// that actually does, instead of a dead-end node with no outgoing edges
// of its own; a bridge method generated for a covariant-return/generic
// override already has the exact erased descriptor this looks for, so
// it needs no special handling here — it resolves like any other
// method, and its own Calls (the delegating invocation to the real,
// covariant-signature implementation) carry the analysis the rest of
// the way.
func (h *Hierarchy) ResolveExact(fqcn, name, desc string) (declaringClass string, found bool) {
	return h.resolveBy(fqcn, func(ci *ClassInfo) bool {
		_, ok := ci.MethodByNameAndDesc(name, desc)
		return ok
	})
}

// ResolveByParams is ResolveExact's REFLECTION R2 counterpart: it finds
// which class in fqcn's superclass/superinterface chain declares a
// method matching name and paramDescs, the same JVM lookup order, but
// without requiring a return type — a resolved `Class#getMethod` call
// proves the parameter types, never the return type (ClassInfo's own
// MethodByNameAndParams doc comment explains why matching without it is
// still exact, not an overload-insensitive guess).
func (h *Hierarchy) ResolveByParams(fqcn, name string, paramDescs []string) (declaringClass string, found bool) {
	return h.resolveBy(fqcn, func(ci *ClassInfo) bool {
		_, ok := ci.MethodByNameAndParams(name, paramDescs)
		return ok
	})
}

// ResolveField finds which class/interface actually declares a field
// with this exact name and descriptor, following the JVM's own field
// resolution order (JVMS 5.4.3.2) — deliberately *not* the same order as
// ResolveMethod/ResolveExact: a field checks fqcn itself, then every
// superinterface (recursively) *before* the superclass, the reverse of
// a method's superclass-first order. Used by JVM_GRAPH G3 to find the
// class whose <clinit> a getstatic/putstatic should trigger an edge to.
func (h *Hierarchy) ResolveField(fqcn, name, desc string) (declaringClass string, found bool) {
	return h.resolveField(fqcn, name, desc, make(map[string]bool))
}

func (h *Hierarchy) resolveField(fqcn, name, desc string, seen map[string]bool) (declaringClass string, found bool) {
	if seen[fqcn] {
		return "", false // malformed/cyclic hierarchy; don't recurse forever
	}
	seen[fqcn] = true

	ci, ok := h.classes[fqcn]
	if !ok {
		return "", false
	}
	if _, ok := ci.FieldByNameAndDesc(name, desc); ok {
		return fqcn, true
	}
	for _, iface := range ci.Interfaces {
		if d, ok := h.resolveField(iface, name, desc, seen); ok {
			return d, true
		}
	}
	if ci.SuperClass != "" {
		return h.resolveField(ci.SuperClass, name, desc, seen)
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
