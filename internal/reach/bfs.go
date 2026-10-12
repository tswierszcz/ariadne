package reach

import "ariadne/internal/callgraph"

// bfsResult is one multi-source BFS's outcome: which nodes were reached
// from any root, and the parent pointer each one was first reached
// through, for shortest-path reconstruction.
type bfsResult struct {
	visited []bool
	parent  []int32 // -1 for a root or an unvisited node
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
	n := g.NodeCount()
	res := bfsResult{
		visited: make([]bool, n),
		parent:  make([]int32, n),
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
			if res.visited[e.To] {
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
	if int(target) >= len(r.visited) || !r.visited[target] {
		return nil, false
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
