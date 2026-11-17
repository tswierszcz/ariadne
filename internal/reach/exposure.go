package reach

import (
	"sort"
	"strings"

	"ariadne/internal/callgraph"
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
	// ClientDispatch (LIBRARY_MODE LB2): every path from Method in its
	// pass dispatches, at a client-code site, to a library body other
	// than the called method's own — a stand-in for code a client
	// supplies there. Path is then such a path.
	ClientDispatch bool
	// InheritedBy (lb2-3): when Method is a dependency method that API
	// classes inherit, those classes — the receivers a client calls it on.
	InheritedBy []string
}

// ClientCall is a call site where a client of the library may supply the
// code that runs (libapi.ClientSites): Declared (owner#name+desc) is the
// called method; a dispatch from this site to any other body stands in
// for client code.
type ClientCall struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	Declared                           string
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
	// strict is reverse without the client-dispatch edges; nil when
	// there are none.
	strict    [][]uint32
	cache     map[string][]ExposedMethod
	inherited map[uint32][]string
}

// SetExposure makes Analyze list, on every positive finding, the roots
// that reach its mapped code (ExposedAPI). Meant for -roots api, where
// the roots are the library's API; with other roots it lists those.
// client are the client-code sites (LB2) and made reports whether the
// input itself creates a class (any creation origin but an assumed
// entry). An edge whose every site is a client-code site, whose target is
// not the site's called method, and whose target's class (or a subtype
// inheriting its body) the input never creates is a client-dispatch
// edge: the body runs there only on an object the client creates and
// passes. A root that reaches the code only through one is marked
// ClientDispatch. inherited names, for roots that are dependency methods
// API classes inherit, those classes (ExposedMethod.InheritedBy). It
// returns how many client-dispatch edges there are.
func (e *Engine) SetExposure(client []ClientCall, made func(fqcn string) bool, inherited map[uint32][]string) int {
	n := e.g.NodeCount()
	x := &exposure{e: e, isRoot: make([]bool, n), reverse: make([][]uint32, n), cache: map[string][]ExposedMethod{}, inherited: inherited}
	for _, r := range e.roots {
		if int(r) < n {
			x.isRoot[r] = true
		}
	}
	sites := map[uint32]map[int]string{}
	for _, c := range client {
		if id, ok := e.g.ID(c.CallerFQCN, c.CallerName, c.CallerDesc); ok {
			if sites[id] == nil {
				sites[id] = map[int]string{}
			}
			sites[id][c.PC] = c.Declared
		}
	}
	dispatch := func(u uint32, ed callgraph.Edge) bool {
		at := sites[u]
		if at == nil || len(ed.Sites) == 0 {
			return false
		}
		to := e.g.Key(ed.To)
		for _, s := range ed.Sites {
			if d, ok := at[s.PC]; !ok || d == to {
				return false
			}
		}
		owner, _, _ := strings.Cut(to, "#")
		if made != nil {
			if made(owner) {
				return false
			}
			for _, sub := range e.h.Subtypes(owner) {
				if made(sub) {
					return false
				}
			}
		}
		return true
	}
	clientEdges := 0
	for u := 0; u < n; u++ {
		for _, ed := range e.g.Edges(uint32(u)) {
			x.reverse[ed.To] = append(x.reverse[ed.To], uint32(u))
			if dispatch(uint32(u), ed) {
				clientEdges++
			}
		}
	}
	if clientEdges > 0 {
		x.strict = make([][]uint32, n)
		for u := 0; u < n; u++ {
			for _, ed := range e.g.Edges(uint32(u)) {
				if !dispatch(uint32(u), ed) {
					x.strict[ed.To] = append(x.strict[ed.To], uint32(u))
				}
			}
		}
	}
	e.exposure = x
	return clientEdges
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
	// lb2-4: a target's receiver-specialized copies are the target too.
	for _, t := range ts {
		ts = append(ts, x.e.g.Clones(t)...)
	}
	appNext := x.backward(x.reverse, ts, true)
	fullNext := x.backward(x.reverse, ts, false)
	appStrict, fullStrict := appNext, fullNext
	if x.strict != nil {
		appStrict = x.backward(x.strict, ts, true)
		fullStrict = x.backward(x.strict, ts, false)
	}
	var out []ExposedMethod
	byMethod := map[string]int{}
	for id := range x.isRoot {
		if !x.isRoot[id] {
			continue
		}
		pass, next, strict := "", fullNext, fullStrict
		switch {
		case appNext[id] != notVisited:
			pass, next, strict = PassAppOnly, appNext, appStrict
		case fullNext[id] != notVisited:
			pass = PassFull
		default:
			continue
		}
		client := strict[id] == notVisited
		if !client {
			next = strict
		}
		path := []uint32{uint32(id)}
		for cur := uint32(id); next[cur] != isTarget; {
			cur = uint32(next[cur])
			path = append(path, cur)
		}
		rendered := renderPath(x.e.g, path)
		edges, char := pathEvidence(x.e.g, rendered)
		m := ExposedMethod{Method: x.e.g.Key(uint32(id)), Pass: pass, Path: rendered, PathEdges: edges, Character: char, ClientDispatch: client, InheritedBy: x.inherited[uint32(id)]}
		if i, ok := byMethod[m.Method]; ok {
			// lb2-4: one root per inheriting API class (receiver-
			// specialized copies of the same method) — one entry naming
			// every class whose copy reaches the code, with the best
			// path (app-only over full, no client dispatch over client
			// dispatch, then shorter).
			o := &out[i]
			by := append(append([]string(nil), o.InheritedBy...), m.InheritedBy...)
			sort.Strings(by)
			if better(m, *o) {
				*o = m
			}
			o.InheritedBy = uniqueStrings(by)
			continue
		}
		byMethod[m.Method] = len(out)
		out = append(out, m)
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
func (x *exposure) backward(reverse [][]uint32, targets []uint32, appOnly bool) []int32 {
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
		for _, u := range reverse[v] {
			if next[u] != notVisited || (appOnly && x.e.isFrameworkNode(u)) {
				continue
			}
			next[u] = int32(v)
			queue = append(queue, u)
		}
	}
	return next
}

// better orders two exposures of one method (receiver-specialized
// copies): an app-only pass, then no client dispatch, then a shorter path.
func better(a, b ExposedMethod) bool {
	if (a.Pass == PassAppOnly) != (b.Pass == PassAppOnly) {
		return a.Pass == PassAppOnly
	}
	if a.ClientDispatch != b.ClientDispatch {
		return !a.ClientDispatch
	}
	return len(a.Path) < len(b.Path)
}

func uniqueStrings(sorted []string) []string {
	var out []string
	for i, s := range sorted {
		if i == 0 || s != sorted[i-1] {
			out = append(out, s)
		}
	}
	return out
}
