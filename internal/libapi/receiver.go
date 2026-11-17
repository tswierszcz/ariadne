package libapi

import (
	"fmt"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
)

// SpecializeStats is what Specialize did, for the log and v2 coverage.
type SpecializeStats struct {
	// Clones counts the receiver-specialized copies of methods.
	Clones int `json:"clones"`
	// ThisSites counts the copied call sites whose receiver is the
	// method's own unchanged this.
	ThisSites int `json:"this_sites"`
	// Dropped counts the edges from those sites to a body no class of
	// the receiver's type selects.
	Dropped int `json:"edges_dropped"`
	// Unanalyzed counts the copied methods whose receivers could not be
	// told apart (class not re-read, local-value analysis aborted): all
	// their edges are kept.
	Unanalyzed int `json:"unanalyzed_methods"`
}

// Specialize gives each inherited root (lb2-3) one copy per API class
// that inherits it, run on a receiver of that class (LIBRARY_MODE lb2-4):
// in the copy, a virtual or interface call on the method's own unchanged
// this goes only to the bodies the API class and its subtypes select —
// YAMLFactory's createParser, not JsonFactory's — and a call on this to a
// method the API class also inherits from a proper supertype goes to that
// method's copy for the same class, transitively. Every other edge is
// copied unchanged. dispatch is the graph's own dispatch rule (the
// strategy that built it), asked as if the call named the API class.
//
// It returns the roots with each inherited root replaced by its copies
// and inherited re-keyed to them (each copy names its one class).
func Specialize(g *callgraph.Graph, h *hierarchy.Hierarchy, roots []uint32, inherited map[uint32][]string, dispatch func(classfile.Call) []string) ([]uint32, map[uint32][]string, SpecializeStats) {
	s := &specializer{g: g, h: h, dispatch: dispatch, memo: map[ctx]uint32{}, facts: map[uint32]*dataflow.MethodFacts{}, targets: map[string]map[string]bool{}}
	defer s.reader.Close()
	var out []uint32
	byClone := map[uint32][]string{}
	for _, r := range roots {
		classes := inherited[r]
		if len(classes) == 0 {
			out = append(out, r)
			continue
		}
		for _, c := range classes {
			id := s.clone(r, c)
			out = append(out, id)
			byClone[id] = []string{c}
		}
	}
	for len(s.queue) > 0 {
		cx := s.queue[0]
		s.queue = s.queue[1:]
		g.SetCloneEdges(s.memo[cx], s.edges(cx))
	}
	s.stats.Clones = g.CloneCount()
	return out, byClone, s.stats
}

type ctx struct {
	orig uint32
	recv string // the API class the receiver is (or a subtype of)
}

type specializer struct {
	g        *callgraph.Graph
	h        *hierarchy.Hierarchy
	dispatch func(classfile.Call) []string
	reader   classfile.JarClassReader
	memo     map[ctx]uint32
	queue    []ctx
	facts    map[uint32]*dataflow.MethodFacts // nil: not analyzable
	targets  map[string]map[string]bool       // recv#name+desc+kind → declaring classes
	stats    SpecializeStats
}

func (s *specializer) clone(orig uint32, recv string) uint32 {
	cx := ctx{orig, recv}
	if id, ok := s.memo[cx]; ok {
		return id
	}
	id := s.g.AddClone(orig)
	s.memo[cx] = id
	s.queue = append(s.queue, cx)
	return id
}

func splitKey(key string) (owner, name, desc string) {
	owner, rest, _ := strings.Cut(key, "#")
	name, desc = rest, ""
	if i := strings.IndexByte(rest, '('); i >= 0 {
		name, desc = rest[:i], rest[i:]
	}
	return owner, name, desc
}

// thisFacts returns orig's local-value facts with its receiver marked,
// or nil when they cannot be had.
func (s *specializer) thisFacts(orig uint32) *dataflow.MethodFacts {
	if f, ok := s.facts[orig]; ok {
		return f
	}
	s.facts[orig] = nil
	owner, name, desc := splitKey(s.g.Key(orig))
	ci, ok := s.h.Lookup(owner)
	if !ok || ci.SourceJar == "" {
		return nil
	}
	m, ok := ci.MethodByNameAndDesc(name, desc)
	if !ok || m.IsStatic() {
		return nil
	}
	cf, err := s.reader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		return nil
	}
	f, err := dataflow.AnalyzeWithThis(cf, owner, m.Method)
	if err != nil {
		return nil
	}
	if aborted, _ := f.Aborted(); aborted {
		return nil
	}
	s.facts[orig] = f
	return f
}

// onThis reports whether the call at pc in orig is on its own this.
func onThis(f *dataflow.MethodFacts, key string, pc int) (on bool) {
	owner, name, desc := splitKey(key)
	defer func() {
		if recover() != nil { // no receiver recorded there: not a this-call
			on = false
		}
	}()
	return f.ReceiverAt(bytecodeir.Site{Method: owner + "#" + name, Descriptor: desc, PC: pc, Op: bytecodeir.OpCall}).IsThis()
}

// selects reports the declaring classes a receiver of class recv (or a
// subtype) dispatches name+desc to; empty when dispatch finds none.
func (s *specializer) selects(recv, name, desc string, kind classfile.CallKind) map[string]bool {
	k := fmt.Sprintf("%s#%s%s/%s", recv, name, desc, kind)
	if t, ok := s.targets[k]; ok {
		return t
	}
	t := map[string]bool{}
	// The body the API class itself selects (dispatch lists what its
	// subtypes add), then theirs.
	if d, ok := s.h.ResolveExact(recv, name, desc); ok {
		t[d] = true
	}
	for _, d := range s.dispatch(classfile.Call{Kind: kind, Owner: recv, Name: name, Desc: desc}) {
		t[d] = true
	}
	s.targets[k] = t
	return t
}

func (s *specializer) edges(cx ctx) []callgraph.Edge {
	orig := s.g.Origin(cx.orig)
	src := s.g.Edges(orig)
	f := s.thisFacts(orig)
	if f == nil {
		s.stats.Unanalyzed++
		return append([]callgraph.Edge(nil), src...)
	}
	key := s.g.Key(orig)
	// inherits: the body runs on this receiver in a proper supertype of
	// the API class — selected from it (virtual, interface) or named
	// exactly (invokespecial: a private method, a super call).
	inherits := func(owner, name, desc string, kind classfile.CallKind) bool {
		if owner == cx.recv {
			return false
		}
		if kind == classfile.CallSpecial {
			return s.isSupertype(owner, cx.recv)
		}
		decl, ok := s.h.ResolveExact(cx.recv, name, desc)
		return ok && decl == owner
	}
	var out []callgraph.Edge
	for _, e := range src {
		if len(e.Sites) == 0 || (e.Basis != callgraph.BasisDirect && e.Basis != callgraph.BasisCHA) ||
			(e.Kind != classfile.CallVirtual && e.Kind != classfile.CallInterface && e.Kind != classfile.CallSpecial) {
			out = append(out, e)
			continue
		}
		owner, name, desc := splitKey(s.g.Key(e.To))
		var this, other []callgraph.CallSite
		for _, site := range e.Sites {
			if name != "<init>" && onThis(f, key, site.PC) {
				this = append(this, site)
			} else {
				other = append(other, site)
			}
		}
		if len(other) > 0 {
			o := e
			o.Sites = other
			out = append(out, o)
		}
		if len(this) == 0 {
			continue
		}
		s.stats.ThisSites += len(this)
		to := e.To
		if e.Kind != classfile.CallSpecial {
			// A receiver of the API class's type selects these bodies;
			// none found (the method comes from outside the input) keeps
			// the edge, as the graph itself does.
			if t := s.selects(cx.recv, name, desc, e.Kind); len(t) > 0 && !t[owner] {
				s.stats.Dropped++
				continue
			}
		}
		// The body the API class itself runs, declared in a proper
		// supertype, runs on the same receiver: its copy.
		if inherits(owner, name, desc, e.Kind) {
			if m, ok := s.h.Lookup(owner); ok {
				if mm, ok := m.MethodByNameAndDesc(name, desc); ok && !mm.IsStatic() && mm.AccessFlags&classfile.AccAbstract == 0 {
					to = s.clone(e.To, cx.recv)
				}
			}
		}
		t := e
		t.To, t.Sites = to, this
		out = append(out, t)
	}
	return out
}

// isSupertype reports whether sup is a proper supertype of fqcn.
func (s *specializer) isSupertype(sup, fqcn string) bool {
	seen := map[string]bool{fqcn: true}
	queue := []string{fqcn}
	for len(queue) > 0 {
		ci, ok := s.h.Lookup(queue[0])
		queue = queue[1:]
		if !ok {
			continue
		}
		for _, n := range append([]string{ci.SuperClass}, ci.Interfaces...) {
			if n == sup {
				return true
			}
			if n != "" && !seen[n] {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	return false
}
