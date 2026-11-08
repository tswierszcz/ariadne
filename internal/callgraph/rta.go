package callgraph

import (
	"ariadne/internal/classfile"
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/hierarchy"
)

// GRAPH_PRECISION P2: rapid type analysis. The graph is built exactly as
// for CHA (the CHA set is RTA's candidate set), every model edge is
// added, and then ApplyRTA narrows it: a fixed point over the methods
// reachable from the roots and the classes those methods create (the P1
// catalog, translated by the caller into RTAInputs) decides which classes
// are live, and a CHA-basis edge survives only if some live class
// dispatches the call to that edge's target (exact descriptor, JVM
// method selection — the same resolution the CHA edge was built with).
// Every other edge (direct, init, model, synthetic, …) is followed and
// kept as it is.
//
// Unknown allocation (policy PolicyCHAFallbackWithGap): a creation site
// the analysis does not model (unresolved reflection, Unsafe, a
// constructor handle, a ServiceLoader with an unknown service, roots=all
// parameters) makes its fallback classes live as soon as the site is
// reachable — the classes such a mechanism could create, bounded by what
// the caller can justify — and is reported as an active gap. A CHA edge
// kept only because of fallback classes is counted apart, so the report
// says how much of the graph the policy kept rather than proved.

// ModeRTA is the -callgraph-mode value of rapid type analysis.
const ModeRTA = "rta"

// RTA is the strategy recorded for an RTA graph: at build time it
// produces the CHA candidate set; ApplyRTA narrows it.
type RTA struct {
	Skip SkipFunc
}

// Mode implements Strategy.
func (RTA) Mode() string { return ModeRTA }

// UnknownAllocationPolicy implements Strategy.
func (RTA) UnknownAllocationPolicy() string { return PolicyCHAFallbackWithGap }

// SkipRule implements Strategy.
func (r RTA) SkipRule() string { return CHA{Skip: r.Skip}.SkipRule() }

// DispatchTargets implements Strategy: the CHA candidates.
func (r RTA) DispatchTargets(h *hierarchy.Hierarchy, call classfile.Call) []string {
	return CHA{Skip: r.Skip}.DispatchTargets(h, call)
}

// RTAVersion versions the fixed point and the fallback rules.
const RTAVersion = "p2-1"

// Creation makes Classes live once Site (a method node) is reachable; a
// creation with Always set is live from the start (container beans,
// classes whose entry points a framework calls).
type Creation struct {
	Site    uint32
	Always  bool
	Classes []string
}

// UnknownCreation is a creation site of an unmodeled mechanism: when it
// is reachable (or Always), the gap is active and Fallback becomes live.
// Fallback may be empty when the mechanism creates no class of the
// analyzed input (a dynamic proxy's generated class).
type UnknownCreation struct {
	Kind     string
	Site     uint32
	Always   bool
	Fallback []string
}

// RTAInputs are the roots and the creation facts.
type RTAInputs struct {
	Roots     []uint32
	Creations []Creation
	Unknown   []UnknownCreation
}

// RTAResult describes what the fixed point found and what ApplyRTA
// removed.
type RTAResult struct {
	Version          string         `json:"version"`
	ReachableMethods int            `json:"reachable_methods"`
	LiveClasses      int            `json:"live_classes"`
	FallbackClasses  int            `json:"live_only_by_fallback"`
	CHAEdgesBefore   int            `json:"cha_edges_before"`
	CHAEdgesRemoved  int            `json:"cha_edges_removed"`
	CHAEdgesFallback int            `json:"cha_edges_kept_by_fallback"`
	ActiveGaps       map[string]int `json:"active_gaps,omitempty"`
	// Live lists the live classes (sorted), for tests and -graph-out.
	Live []string `json:"-"`
	// FallbackOnly lists the classes live only by fallback (sorted).
	FallbackOnly []string `json:"-"`
}

// GapText renders ActiveGaps for the report ("none" when empty).
func (r RTAResult) GapText() string {
	if len(r.ActiveGaps) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(r.ActiveGaps))
	for k := range r.ActiveGaps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + " " + strconv.Itoa(r.ActiveGaps[k])
	}
	return strings.Join(parts, ", ")
}

const (
	liveNone = iota
	liveFallback
	liveCreated
)

type rtaState struct {
	g *Graph
	h *hierarchy.Hierarchy

	reach      []bool
	work       []uint32
	live       map[string]int // class -> liveCreated/liveFallback
	createdAt  map[uint32][]string
	unknownAt  map[uint32][]int
	unknown    []UnknownCreation
	active     []bool
	enabled    map[uint32]bool
	checked    map[uint32]bool
	pendingBy  map[string][]uint32 // declaring class -> CHA targets not yet enabled
	ancestors  map[string][]string
	targetName map[uint32][3]string
}

// ApplyRTA runs the fixed point on g (built with the RTA strategy and
// every model edge already added) and removes the CHA-basis edges no
// live class dispatches to. The result does not depend on the order of
// roots, creations or edges.
func ApplyRTA(g *Graph, h *hierarchy.Hierarchy, in RTAInputs) RTAResult {
	s := &rtaState{
		g: g, h: h,
		reach:      make([]bool, g.NodeCount()),
		live:       map[string]int{},
		createdAt:  map[uint32][]string{},
		unknownAt:  map[uint32][]int{},
		unknown:    in.Unknown,
		active:     make([]bool, len(in.Unknown)),
		enabled:    map[uint32]bool{},
		checked:    map[uint32]bool{},
		pendingBy:  map[string][]uint32{},
		ancestors:  map[string][]string{},
		targetName: map[uint32][3]string{},
	}
	for _, c := range in.Creations {
		if c.Always {
			for _, cl := range c.Classes {
				s.markLive(cl, liveCreated)
			}
			continue
		}
		s.createdAt[c.Site] = append(s.createdAt[c.Site], c.Classes...)
	}
	for i, u := range in.Unknown {
		if u.Always {
			s.activate(i)
			continue
		}
		s.unknownAt[u.Site] = append(s.unknownAt[u.Site], i)
	}
	for _, r := range in.Roots {
		s.markReach(r)
	}
	for len(s.work) > 0 {
		m := s.work[len(s.work)-1]
		s.work = s.work[:len(s.work)-1]
		s.visit(m)
	}

	res := RTAResult{Version: RTAVersion, ActiveGaps: map[string]int{}}
	for _, r := range s.reach {
		if r {
			res.ReachableMethods++
		}
	}
	for cl, how := range s.live {
		res.Live = append(res.Live, cl)
		if how == liveFallback {
			res.FallbackOnly = append(res.FallbackOnly, cl)
		}
	}
	sort.Strings(res.Live)
	sort.Strings(res.FallbackOnly)
	res.LiveClasses, res.FallbackClasses = len(res.Live), len(res.FallbackOnly)
	for i, a := range s.active {
		if a {
			res.ActiveGaps[s.unknown[i].Kind]++
		}
	}
	if len(res.ActiveGaps) == 0 {
		res.ActiveGaps = nil
	}

	// Narrow: decide every distinct CHA target once against the final
	// live set (also targets only unreachable methods call — the decision
	// is the target's, not the caller's).
	decision := map[uint32]int{}
	removedCHA := 0
	for id := range g.adjacency {
		adj := g.adjacency[id]
		kept := adj[:0]
		for _, e := range adj {
			if e.Basis != BasisCHA {
				kept = append(kept, e)
				continue
			}
			res.CHAEdgesBefore++
			d, ok := decision[e.To]
			if !ok {
				d = s.dispatchedBy(e.To)
				decision[e.To] = d
			}
			switch d {
			case liveNone:
				res.CHAEdgesRemoved++
				removedCHA++
				continue
			case liveFallback:
				res.CHAEdgesFallback++
			}
			kept = append(kept, e)
		}
		// Clear the tail so removed edges' Sites can be collected.
		for i := len(kept); i < len(adj); i++ {
			adj[i] = Edge{}
		}
		g.adjacency[id] = kept
	}
	g.chaEdges -= removedCHA
	return res
}

func (s *rtaState) markReach(id uint32) {
	if int(id) >= len(s.reach) || s.reach[id] {
		return
	}
	s.reach[id] = true
	s.work = append(s.work, id)
}

func (s *rtaState) visit(m uint32) {
	for _, cl := range s.createdAt[m] {
		s.markLive(cl, liveCreated)
	}
	for _, i := range s.unknownAt[m] {
		s.activate(i)
	}
	for _, e := range s.g.adjacency[m] {
		if e.Basis != BasisCHA {
			s.markReach(e.To)
			continue
		}
		s.consider(e.To)
	}
}

func (s *rtaState) activate(i int) {
	if s.active[i] {
		return
	}
	s.active[i] = true
	for _, cl := range s.unknown[i].Fallback {
		s.markLive(cl, liveFallback)
	}
}

// markLive makes cl live and enables every pending CHA target cl
// dispatches to.
func (s *rtaState) markLive(cl string, how int) {
	if s.live[cl] >= how {
		return
	}
	was := s.live[cl]
	s.live[cl] = how
	if was != liveNone {
		return // already enabled what it dispatches to
	}
	for _, a := range s.ancestorsOf(cl) {
		pend := s.pendingBy[a]
		if len(pend) == 0 {
			continue
		}
		rest := pend[:0]
		for _, t := range pend {
			if s.dispatches(cl, t) {
				s.enable(t)
				continue
			}
			rest = append(rest, t)
		}
		s.pendingBy[a] = rest
	}
}

func (s *rtaState) enable(t uint32) {
	s.enabled[t] = true
	s.markReach(t)
}

// consider follows a CHA edge to t if a live class dispatches to t, and
// otherwise parks t until one does.
func (s *rtaState) consider(t uint32) {
	if s.enabled[t] {
		s.markReach(t)
		return
	}
	if s.checked[t] {
		return
	}
	s.checked[t] = true
	if s.dispatchedBy(t) != liveNone {
		s.enable(t)
		return
	}
	owner := s.target(t)[0]
	s.pendingBy[owner] = append(s.pendingBy[owner], t)
}

// dispatchedBy reports whether some live class (created, else fallback)
// selects t's declaration for t's name and descriptor.
func (s *rtaState) dispatchedBy(t uint32) int {
	owner := s.target(t)[0]
	best := liveNone
	try := func(cl string) {
		if how := s.live[cl]; how > best && s.dispatches(cl, t) {
			best = how
		}
	}
	try(owner)
	for _, sub := range s.h.Subtypes(owner) {
		if best == liveCreated {
			break
		}
		try(sub)
	}
	return best
}

// dispatches reports whether a receiver of class cl runs t's declaration.
func (s *rtaState) dispatches(cl string, t uint32) bool {
	n := s.target(t)
	got, ok := s.h.ResolveExact(cl, n[1], n[2])
	return ok && got == n[0]
}

func (s *rtaState) target(t uint32) [3]string {
	if n, ok := s.targetName[t]; ok {
		return n
	}
	key := s.g.Key(t)
	owner, rest, _ := strings.Cut(key, "#")
	name, desc := rest, ""
	if i := strings.IndexByte(rest, '('); i >= 0 {
		name, desc = rest[:i], rest[i:]
	}
	n := [3]string{owner, name, desc}
	s.targetName[t] = n
	return n
}

// ancestorsOf: cl, its superclasses and every interface it implements,
// as far as the analyzed universe knows them.
func (s *rtaState) ancestorsOf(cl string) []string {
	if a, ok := s.ancestors[cl]; ok {
		return a
	}
	seen := map[string]bool{}
	var out []string
	var walk func(string)
	walk = func(c string) {
		if c == "" || seen[c] {
			return
		}
		seen[c] = true
		out = append(out, c)
		ci, ok := s.h.Lookup(c)
		if !ok {
			return
		}
		walk(ci.SuperClass)
		for _, i := range ci.Interfaces {
			walk(i)
		}
	}
	walk(cl)
	s.ancestors[cl] = out
	return out
}
