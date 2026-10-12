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

// Edge is one call-site edge, naming both its target and how the call
// was made (kept mainly for the report; reachability itself only needs
// the target).
type Edge struct {
	To   uint32
	Kind classfile.CallKind
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

func (g *Graph) addEdge(from uint32, to uint32, kind classfile.CallKind) {
	g.adjacency[from] = append(g.adjacency[from], Edge{To: to, Kind: kind})
}
