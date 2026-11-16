package reach

import (
	"sort"
	"strings"
)

// ExposedMethod is one root (an API method under -roots api) from which a
// finding's mapped code is reached (LIBRARY_MODE LB1,
// docs/RESULTS_CONTRACT.MD §8).
type ExposedMethod struct {
	Method string
	// Pass is "app_only" when the method reaches the code without passing
	// through framework code (the REACHABLE pass), "full" otherwise.
	Pass string
	// Path is one shortest path from Method to a mapped method (in the
	// method's pass); PathEdges/Character its evidence, as a finding's.
	Path      []string
	PathEdges []PathEdge
	Character []string
}

// Exposed passes.
const (
	PassAppOnly = "app_only"
	PassFull    = "full"
)

// exposure answers, per set of matched nodes, which roots reach them:
// one backward BFS per pass over the reversed graph, with the forward
// passes' rule that a framework node is not passed through in the
// app-only pass.
type exposure struct {
	e       *Engine
	isRoot  []bool
	reverse [][]uint32
	cache   map[string][]ExposedMethod
}

// SetExposure makes Analyze list, on every positive finding, the roots
// that reach its mapped code (ExposedAPI). Meant for -roots api, where
// the roots are the library's API; with other roots it lists those.
func (e *Engine) SetExposure() {
	n := e.g.NodeCount()
	x := &exposure{e: e, isRoot: make([]bool, n), reverse: make([][]uint32, n), cache: map[string][]ExposedMethod{}}
	for _, r := range e.roots {
		if int(r) < n {
			x.isRoot[r] = true
		}
	}
	for u := 0; u < n; u++ {
		for _, ed := range e.g.Edges(uint32(u)) {
			x.reverse[ed.To] = append(x.reverse[ed.To], uint32(u))
		}
	}
	e.exposure = x
}

func (x *exposure) of(targets []uint32) []ExposedMethod {
	if len(targets) == 0 {
		return nil
	}
	ts := append([]uint32(nil), targets...)
	sort.Slice(ts, func(i, j int) bool { return ts[i] < ts[j] })
	var kb strings.Builder
	for _, t := range ts {
		kb.WriteString(x.e.g.Key(t))
		kb.WriteByte('\n')
	}
	key := kb.String()
	if out, ok := x.cache[key]; ok {
		return out
	}
	appNext := x.backward(ts, true)
	fullNext := x.backward(ts, false)
	var out []ExposedMethod
	for id := range x.isRoot {
		if !x.isRoot[id] {
			continue
		}
		pass, next := "", fullNext
		switch {
		case appNext[id] != notVisited:
			pass, next = PassAppOnly, appNext
		case fullNext[id] != notVisited:
			pass = PassFull
		default:
			continue
		}
		path := []uint32{uint32(id)}
		for cur := uint32(id); next[cur] != isTarget; {
			cur = uint32(next[cur])
			path = append(path, cur)
		}
		rendered := renderPath(x.e.g, path)
		edges, char := pathEvidence(x.e.g, rendered)
		out = append(out, ExposedMethod{Method: x.e.g.Key(uint32(id)), Pass: pass, Path: rendered, PathEdges: edges, Character: char})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Method < out[j].Method })
	x.cache[key] = out
	return out
}

const (
	notVisited = int32(-2)
	isTarget   = int32(-1)
)

// backward returns, for every node, the next node on a shortest path to
// a target (isTarget for a target itself, notVisited when no path). In
// the app-only pass an edge u → v is followed only when u is not a
// framework node, as the forward pass does not expand one.
func (x *exposure) backward(targets []uint32, appOnly bool) []int32 {
	next := make([]int32, x.e.g.NodeCount())
	for i := range next {
		next[i] = notVisited
	}
	queue := make([]uint32, 0, len(targets))
	for _, t := range targets {
		if next[t] == notVisited {
			next[t] = isTarget
			queue = append(queue, t)
		}
	}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, u := range x.reverse[v] {
			if next[u] != notVisited || (appOnly && x.e.isFrameworkNode(u)) {
				continue
			}
			next[u] = int32(v)
			queue = append(queue, u)
		}
	}
	return next
}

