package reach

import "ariadne/internal/callgraph"

// bfsResult is one multi-source BFS's outcome: which nodes were reached
// from any root, and the parent pointer each one was first reached
// through, for shortest-path reconstruction.
type bfsResult struct {
	visited []bool
	parent  []int32 // -1 for a root or an unvisited node
	// clones lists a method's receiver-specialized copies
	// (callgraph.Graph.AddClone, LIBRARY_MODE lb2-4): path to a method
	// the search never entered itself goes through its nearest copy.
	clones func(uint32) []uint32
}

// bfs runs a multi-source breadth-first search over g starting from
// every node in roots simultaneously, so the parent pointers it records
// give the shortest path from *some* root, not a specific one.
//
// isSink(id), when true, stops the search from expanding id's own
// outgoing edges while still marking id itself as visited (with a valid
// path to it) — id becomes a dead end rather than invisible. This is
// exactly the REACHABLE/FRAMEWORK split from section 3.6/5: a framework
// node is reached and can itself be a REACHABLE match (the app called it
// directly), but anything only reachable by continuing *through* several
// hops of framework-internal calls is not, since those calls are
// typically reflection-mediated and the static call graph can't see
// where they actually go.
func bfs(g *callgraph.Graph, roots []uint32, isSink func(id uint32) bool) bfsResult {
	return bfsKeep(g, roots, isSink, nil)
}

// noCHA keeps every edge except a CHA expansion. A target that another
// rule also reaches from the same caller stays reachable through that
// edge, so only CHA-only steps are cut (ROADMAP B0 b0-2).
func noCHA(_ uint32, e callgraph.Edge) bool { return e.Basis != callgraph.BasisCHA }

// bfsKeep is bfs over the edges keep admits (every edge when keep is nil).
func bfsKeep(g *callgraph.Graph, roots []uint32, isSink func(id uint32) bool, keep func(u uint32, e callgraph.Edge) bool) bfsResult {
	n := g.NodeCount()
	res := bfsResult{
		visited: make([]bool, n),
		parent:  make([]int32, n),
		clones:  g.Clones,
	}
	for i := range res.parent {
		res.parent[i] = -1
	}

	queue := make([]uint32, 0, len(roots))
	for _, r := range roots {
		if int(r) >= n || res.visited[r] {
			continue
		}
		res.visited[r] = true
		queue = append(queue, r)
	}

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		if isSink(u) {
			continue
		}
		for _, e := range g.Edges(u) {
			if res.visited[e.To] || (keep != nil && !keep(u, e)) {
				continue
			}
			res.visited[e.To] = true
			res.parent[e.To] = int32(u)
			queue = append(queue, e.To)
		}
	}
	return res
}

// path reconstructs the shortest root-to-target path bfs found, as a
// sequence of node ids from a root to target inclusive. ok is false if
// target was never visited.
func (r bfsResult) path(target uint32) ([]uint32, bool) {
	if int(target) >= len(r.visited) {
		return nil, false
	}
	if !r.visited[target] {
		// A method entered only as receiver-specialized copies is
		// reached through the copy with the shortest path (the first
		// copy on a tie: copies are created in a deterministic order).
		var best []uint32
		if r.clones != nil {
			for _, c := range r.clones(target) {
				if p, ok := r.path(c); ok && (best == nil || len(p) < len(best)) {
					best = p
				}
			}
		}
		return best, best != nil
	}
	var rev []uint32
	cur := int32(target)
	for cur != -1 {
		rev = append(rev, uint32(cur))
		cur = r.parent[cur]
	}
	out := make([]uint32, len(rev))
	for i, id := range rev {
		out[len(rev)-1-i] = id
	}
	return out, true
}
