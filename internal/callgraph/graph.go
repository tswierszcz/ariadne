// Package callgraph interns every method in a hierarchy.Hierarchy into a
// compact integer id and builds the directed call graph between them,
// expanding each virtual/interface call via class-hierarchy analysis
// (CHA): a call to T#m also reaches every subtype of T that declares m,
// since the JVM could dispatch to any of them at run time.
package callgraph

import (
	"fmt"

	"ariadne/internal/classfile"
)

// EdgeBasis says why an edge exists, independent of Kind (which names the
// bytecode call/dispatch mechanism — virtual, interface, static, ...).
// Kept as its own field (RECOMMENDATION JVM_GRAPH G0) so a future
// "initializes" edge (G3, a <clinit> edge from new/getstatic/putstatic/
// invokestatic) doesn't have to overload Kind's meaning — it adds a new
// EdgeBasis value instead, next to BasisCalls here.
type EdgeBasis string

const (
	// BasisDirect is a direct edge to exactly the call site's own symbolic
	// target, found straight in the caller's bytecode.
	BasisDirect EdgeBasis = "direct"
	// BasisCHA is an edge class-hierarchy-analysis expansion added to a
	// subtype of the call's declared receiver (section 3.5 of PLAN.md) —
	// the resolution rule a reader needs to know this edge is an
	// over-approximation, not a direct call.
	BasisCHA EdgeBasis = "cha"
	// BasisSynthetic is an edge from a profile's synthetic_edges (PLAN.md
	// section 4), standing in for a runtime-generated proxy with no
	// bytecode implementation at all (e.g. Spring Data).
	BasisSynthetic EdgeBasis = "synthetic"
)

// CallSite is one piece of evidence for why an edge exists: the bytecode
// offset of the invoke instruction that produced it (or, for a
// BasisCHA edge, of the original call whose declared receiver triggered
// the CHA expansion to this edge's actual target).
type CallSite struct {
	PC int
}

// Edge is one call-site edge, naming its target, how the call was made
// (Kind), and why the edge exists (Basis). Sites records every bytecode
// offset that produced this exact (To, Kind, Basis) edge — several call
// sites in one method that happen to call the same target the same way
// collapse into one Edge (reachability itself only needs the target), but
// their evidence is kept, not discarded, so a report can still point at a
// specific line of bytecode. Sites is empty for a BasisSynthetic edge,
// which has no real call site at all.
type Edge struct {
	To    uint32
	Kind  classfile.CallKind
	Basis EdgeBasis
	Sites []CallSite
}

// Graph is the whole interned call graph: every method ariadne read is a
// node (even one with no edges, and even one no edge points at), and
// Adjacency holds the outgoing edges found by ExtractCalls plus CHA.
type Graph struct {
	// keys[id] is the node's "fqcn#name(desc)" key; ids are dense,
	// starting at 0, in the order nodes were first interned.
	keys []string
	// idOf is the reverse of keys.
	idOf map[string]uint32
	// byClassMethod indexes "fqcn#name" (no descriptor) to every id with
	// that name, for the overload-insensitive matching internal/reach
	// needs to line up a vulnerable method from vuln_methods.yaml (which
	// has no descriptor) with the graph's own nodes.
	byClassMethod map[string][]uint32

	adjacency [][]Edge // adjacency[id] = outgoing edges from id

	// chaEdges/unresolvedDynamic are running totals kept for the report
	// (graph size before/after CHA, how many dynamic call sites never
	// resolved to an edge at all).
	preCHAEdges       int
	chaEdges          int
	unresolvedDynamic int
	// unresolvedOwner counts direct calls whose declared receiver type is
	// itself part of the analyzed universe, but whose method could not
	// be resolved to any declaring class within it (JVM_GRAPH G2). This
	// is expected to be a large number on any real app: it is dominated
	// by methods genuinely inherited from a class outside the analyzed
	// input — almost always the JDK (Thread#start, Object#toString,
	// AbstractList#get, ...), which ariadne never reads by design, one
	// level removed through an app/dependency class that itself extends
	// it. It is not, on its own, evidence of a bug or a gap worth
	// chasing; see UnresolvedOwnerCount's own doc comment.
	unresolvedOwner int
}

// NodeCount is the number of distinct methods in the graph.
func (g *Graph) NodeCount() int { return len(g.keys) }

// EdgeCount is the total number of outgoing edges across every node,
// after CHA expansion.
func (g *Graph) EdgeCount() int {
	n := 0
	for _, es := range g.adjacency {
		n += len(es)
	}
	return n
}

// PreCHAEdgeCount is EdgeCount before CHA expansion added any edges,
// useful for the report to show how much CHA actually added.
func (g *Graph) PreCHAEdgeCount() int { return g.preCHAEdges }

// CHAEdgeCount is how many of EdgeCount's edges were added by CHA
// expansion rather than found directly by classfile.ExtractCalls.
func (g *Graph) CHAEdgeCount() int { return g.chaEdges }

// UnresolvedDynamicCount is how many invokedynamic call sites across the
// whole graph had a bootstrap method ariadne couldn't resolve to a real
// target (StringConcatFactory and the like; see classfile.ExtractCalls).
func (g *Graph) UnresolvedDynamicCount() int { return g.unresolvedDynamic }

// UnresolvedOwnerCount is how many direct calls (JVM_GRAPH G2) named a
// receiver type that is part of the analyzed universe, but whose method
// resolution still failed to find any declaring class in its
// superclass/superinterface chain. Measured on real petclinic/WebGoat
// runs: a few thousand, the large majority a class in the analyzed
// input inheriting a method from the JDK (or any other class outside
// the input) one or more levels up a chain this package never reads
// past — e.g. an app class extending java.lang.Thread and calling the
// inherited start(). That is the expected, common case, not a sign of
// a problem: ariadne cannot tell "the JDK provides this, as expected"
// apart from "this is a genuine hole in the dependency set" once
// resolution has to cross outside the universe it actually read, so
// this count is informational only, not an actionable diagnostic on
// its own.
func (g *Graph) UnresolvedOwnerCount() int { return g.unresolvedOwner }

// Key renders a node id back to its "fqcn#name(desc)" string.
func (g *Graph) Key(id uint32) string {
	if int(id) >= len(g.keys) {
		return fmt.Sprintf("<invalid node %d>", id)
	}
	return g.keys[id]
}

// ID looks up an existing node by its exact fqcn/name/desc; it never
// creates one (only Build does, while reading the hierarchy), so a
// target outside the universe ariadne read (almost always a JDK type)
// correctly reports "not found" rather than fabricating a dangling node.
func (g *Graph) ID(fqcn, name, desc string) (uint32, bool) {
	id, ok := g.idOf[methodKey(fqcn, name, desc)]
	return id, ok
}

// IDsByName returns every node for fqcn#name, regardless of descriptor —
// the overload-insensitive lookup internal/reach uses to match a
// vulnerable method named only "Class#method" against the graph.
func (g *Graph) IDsByName(fqcn, name string) []uint32 {
	return g.byClassMethod[classMethodKey(fqcn, name)]
}

// Edges returns id's outgoing edges.
func (g *Graph) Edges(id uint32) []Edge {
	if int(id) >= len(g.adjacency) {
		return nil
	}
	return g.adjacency[id]
}

// EdgeBetween returns the edge from -> to, if any — used to attach
// resolution-rule/call-site evidence to one step of an already-
// reconstructed path (internal/reach) without the BFS itself having to
// carry that evidence through every hop. When CHA added more than one
// edge between the same pair (different Kind, e.g. both a virtual and
// an interface call site happened to exist), the first one found is
// returned; this is report evidence, not a correctness-critical lookup.
func (g *Graph) EdgeBetween(from, to uint32) (Edge, bool) {
	if int(from) >= len(g.adjacency) {
		return Edge{}, false
	}
	for _, e := range g.adjacency[from] {
		if e.To == to {
			return e, true
		}
	}
	return Edge{}, false
}

func methodKey(fqcn, name, desc string) string {
	return fqcn + "#" + name + desc
}

func classMethodKey(fqcn, name string) string {
	return fqcn + "#" + name
}

// intern returns fqcn#name(desc)'s id, creating and indexing a new node
// if this is the first time it's been seen (as either a caller or, more
// often, a callee that's never itself a caller — an abstract or
// unresolved method still needs a node so edges can point at it).
func (g *Graph) intern(fqcn, name, desc string) uint32 {
	key := methodKey(fqcn, name, desc)
	if id, ok := g.idOf[key]; ok {
		return id
	}
	id := uint32(len(g.keys))
	g.keys = append(g.keys, key)
	g.idOf[key] = id
	g.adjacency = append(g.adjacency, nil)
	cmKey := classMethodKey(fqcn, name)
	g.byClassMethod[cmKey] = append(g.byClassMethod[cmKey], id)
	return id
}

// addEdge always appends a new Edge; callers that want to merge evidence
// into an existing (to, kind, basis) edge instead (build.go's addOrMerge)
// must check for that themselves first.
func (g *Graph) addEdge(from uint32, to uint32, kind classfile.CallKind, basis EdgeBasis, site *CallSite) {
	var sites []CallSite
	if site != nil {
		sites = []CallSite{*site}
	}
	g.adjacency[from] = append(g.adjacency[from], Edge{To: to, Kind: kind, Basis: basis, Sites: sites})
}
