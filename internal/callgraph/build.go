package callgraph

import (
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Build interns every method in h and builds the call graph between
// them: a direct edge for every call site classfile.ExtractCalls found,
// plus — for each virtual/interface call whose declared receiver type
// skip doesn't exempt — an edge to the same-named method on every
// transitive subtype of that receiver (CHA expansion, section 3.3/3.6).
//
// Passing nil for skip uses DefaultSkip. Build never fails: a callee
// outside the hierarchy (almost always a JDK type ariadne never read) is
// still interned as a node with no outgoing edges of its own, so paths
// can still end there.
func Build(h *hierarchy.Hierarchy, skip SkipFunc) *Graph {
	if skip == nil {
		skip = DefaultSkip
	}
	g := &Graph{
		idOf:          make(map[string]uint32),
		byClassMethod: make(map[string][]uint32),
	}

	// Both passes iterate SortedFQCNs rather than ranging over h.All()
	// directly: node/edge insertion order is otherwise at the mercy of
	// Go's randomized map iteration, and since a BFS records the first
	// parent it reaches a node through, that randomness would make which
	// of several equally-short paths internal/reach reports nondeterministic
	// between two runs of the exact same analysis.
	classes := h.All()
	fqcns := h.SortedFQCNs()

	// Pass 1: intern every declared method first, so a method with no
	// outgoing calls at all (a simple getter, an empty override) still
	// gets a node — internal/reach needs it there to use as a root.
	for _, fqcn := range fqcns {
		for _, m := range classes[fqcn].Methods {
			g.intern(fqcn, m.Name, m.Descriptor)
		}
	}

	// Pass 2: add edges. A per-method "seen" index dedupes the common case
	// of the same target called the same way from several call sites in
	// one method (or a direct call and a CHA-expanded edge landing on the
	// same node), without the cost of deduping across the whole graph —
	// but unlike a plain seen-set, a repeat still gets its call site
	// appended as evidence onto the one Edge that survives (JVM_GRAPH G0:
	// "deduplikacja nie usuwa alternatywnych dowodów").
	seen := make(map[edgeKey]int)
	for _, fqcn := range fqcns {
		ci := classes[fqcn]
		for _, m := range ci.Methods {
			if len(m.Calls) == 0 {
				g.unresolvedDynamic += m.UnresolvedDynamic
				continue
			}
			callerID, _ := g.ID(fqcn, m.Name, m.Descriptor)
			clear(seen)

			for _, call := range m.Calls {
				// JVM_GRAPH G2: call.Owner is the bytecode's literal
				// Methodref owner — the call's declared/static receiver
				// type, which does not have to be the class that actually
				// declares the method (an inherited, non-overridden call).
				// Resolve to the real declaring class so the direct edge
				// leads to an available implementation body instead of a
				// dead-end node with no Calls of its own; an owner outside
				// the analyzed universe (almost always a JDK type) is left
				// exactly as the bytecode named it, same as before this
				// phase.
				toOwner, resolveOK := resolveTarget(h, call.Owner, call.Name, call.Desc)
				if !resolveOK {
					g.unresolvedOwner++
				}
				toID := g.intern(toOwner, call.Name, call.Desc)
				if addOrMerge(g, seen, callerID, toID, call.Kind, BasisDirect, call.PC) {
					g.preCHAEdges++
				}

				if call.Kind != classfile.CallVirtual && call.Kind != classfile.CallInterface {
					continue
				}
				if skip(call.Owner) {
					continue
				}
				// CHA expands from the call's own declared receiver type
				// (call.Owner, deliberately not toOwner above): any
				// subtype of *that* static type could be the runtime
				// receiver, regardless of which class actually supplies
				// the inherited body for Owner itself.
				for _, sub := range h.Subtypes(call.Owner) {
					if skip(sub) {
						continue
					}
					subInfo, ok := h.Lookup(sub)
					if !ok {
						continue
					}
					// Exact descriptor match (JVM_GRAPH G2), not
					// MethodsByName's name-only match: a subtype that
					// overrides a *different* overload of the same name
					// must not be wired up as if it overrode this one. A
					// bridge method generated for a covariant-return or
					// generic override already carries the exact erased
					// descriptor being looked for here, so it needs no
					// special handling — it resolves like any other
					// declared method, and its own Calls (the delegating
					// invocation to the real, covariant-signature
					// implementation) carry the analysis the rest of the
					// way.
					if _, ok := subInfo.MethodByNameAndDesc(call.Name, call.Desc); !ok {
						continue
					}
					subID := g.intern(sub, call.Name, call.Desc)
					addOrMerge(g, seen, callerID, subID, call.Kind, BasisCHA, call.PC)
				}
			}
			g.unresolvedDynamic += m.UnresolvedDynamic
		}
	}
	// The report wants "how many edges did CHA add", which is just the
	// difference between the final edge count and the direct-call-only
	// count accumulated above.
	g.chaEdges = g.EdgeCount() - g.preCHAEdges

	return g
}

// resolveTarget finds the class that actually declares name+desc,
// following the JVM's own method-resolution order (JVM_GRAPH G2,
// hierarchy.ResolveExact), so a direct edge can lead to an available
// implementation body instead of a dead-end node on a class that merely
// inherits the method without overriding it.
//
// ok is true both for the ordinary case (owner already declares the
// method, or owner is outside the analyzed universe — almost always a
// JDK type ariadne never read, where there's nothing more precise to
// resolve to) and for a successful resolution to a different declaring
// class. ok is false when owner *is* part of the analyzed universe but
// resolution still found no declaring class anywhere in its
// superclass/superinterface chain — expected to be common (see
// Graph.UnresolvedOwnerCount's own doc comment: this is dominated by
// methods actually inherited from the JDK or another class outside the
// input, one level removed, not by genuine bugs), so this is
// informational, not a reliable signal to investigate on its own.
// resolvedOwner falls back to owner unchanged either way, so the caller
// still gets a valid (if imprecise) dead-end node instead of nothing.
func resolveTarget(h *hierarchy.Hierarchy, owner, name, desc string) (resolvedOwner string, ok bool) {
	ownerInfo, inUniverse := h.Lookup(owner)
	if !inUniverse {
		return owner, true
	}
	if _, declared := ownerInfo.MethodByNameAndDesc(name, desc); declared {
		return owner, true
	}
	if declaring, found := h.ResolveExact(owner, name, desc); found {
		return declaring, true
	}
	return owner, false
}

// edgeKey identifies an edge for seen-set dedup purposes, deliberately
// excluding Sites (which accumulates evidence rather than identifying
// the edge).
type edgeKey struct {
	to    uint32
	kind  classfile.CallKind
	basis EdgeBasis
}

// addOrMerge adds a new (to, kind, basis) edge from `from` the first time
// it's seen, or appends pc as another piece of evidence onto the existing
// edge on every repeat. Reports whether it actually created a new edge
// (used for the preCHAEdges/chaEdges bookkeeping).
func addOrMerge(g *Graph, seen map[edgeKey]int, from, to uint32, kind classfile.CallKind, basis EdgeBasis, pc int) bool {
	k := edgeKey{to: to, kind: kind, basis: basis}
	if idx, ok := seen[k]; ok {
		adj := g.adjacency[from]
		adj[idx].Sites = append(adj[idx].Sites, CallSite{PC: pc})
		return false
	}
	idx := len(g.adjacency[from])
	g.addEdge(from, to, kind, basis, &CallSite{PC: pc})
	seen[k] = idx
	return true
}
