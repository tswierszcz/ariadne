package callgraph

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

const rtaJar = "../../testdata/rta/rta.jar"

// rtaFixture builds the RTA candidate graph of testdata/rta and inputs:
// main as the root, every `new` as a creation at its method, and the
// unresolved reflective factory as an unknown creation whose fallback is
// rt.Dynamic.
func rtaFixture(t *testing.T) (*hierarchy.Hierarchy, *Graph, RTAInputs) {
	t.Helper()
	h, _, err := hierarchy.Build(rtaJar, nil)
	if err != nil {
		t.Fatal(err)
	}
	g := BuildWith(h, RTA{})
	main, ok := g.ID("rt.Oracle", "main", "([Ljava/lang/String;)V")
	if !ok {
		t.Fatal("no main")
	}
	in := RTAInputs{Roots: []uint32{main}}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			id, _ := g.ID(fqcn, m.Name, m.Descriptor)
			for _, r := range m.InitRefs {
				if r.Kind == classfile.InitNew {
					in.Creations = append(in.Creations, Creation{Site: id, Classes: []string{r.Owner}})
				}
			}
		}
	}
	factory, ok := g.ID("rt.UnknownFactory", "create", "(Ljava/lang/String;)Lrt/Shape;")
	if !ok {
		t.Fatal("no factory")
	}
	in.Unknown = []UnknownCreation{{Kind: "reflection-unresolved", Site: factory, Fallback: []string{"rt.Dynamic"}}}
	return h, g, in
}

func edgeSet(g *Graph) []string {
	var out []string
	for from := range g.adjacency {
		for _, e := range g.adjacency[from] {
			out = append(out, fmt.Sprintf("%s -> %s %s/%s", g.Key(uint32(from)), g.Key(e.To), e.Kind, e.Basis))
		}
	}
	sort.Strings(out)
	return out
}

// TestRTAOrderIndependent: the result does not depend on the order of the
// roots, the creations, the unknown sites or each method's edges.
func TestRTAOrderIndependent(t *testing.T) {
	h, g, in := rtaFixture(t)
	want := ApplyRTA(g, h, in)
	wantEdges := edgeSet(g)
	if want.CHAEdgesRemoved == 0 {
		t.Fatal("fixture removes nothing: the test would prove nothing")
	}
	for seed := int64(1); seed <= 5; seed++ {
		_, g2, in2 := rtaFixture(t)
		r := rand.New(rand.NewSource(seed))
		r.Shuffle(len(in2.Creations), func(i, j int) { in2.Creations[i], in2.Creations[j] = in2.Creations[j], in2.Creations[i] })
		in2.Roots = append(in2.Roots, in2.Roots...)
		for id := range g2.adjacency {
			adj := g2.adjacency[id]
			r.Shuffle(len(adj), func(i, j int) { adj[i], adj[j] = adj[j], adj[i] })
		}
		got := ApplyRTA(g2, h, in2)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("seed %d: result %+v, want %+v", seed, got, want)
		}
		if !reflect.DeepEqual(edgeSet(g2), wantEdges) {
			t.Errorf("seed %d: different edges", seed)
		}
	}
}

// TestRTAFallbackAndGap: without the reachable unknown factory, its
// fallback class is not live; with it, the class is live only by the
// fallback and its edge is counted as kept by the fallback.
func TestRTAFallbackAndGap(t *testing.T) {
	h, g, in := rtaFixture(t)
	res := ApplyRTA(g, h, in)
	if !reflect.DeepEqual(res.FallbackOnly, []string{"rt.Dynamic"}) || res.CHAEdgesFallback == 0 || res.ActiveGaps["reflection-unresolved"] != 1 {
		t.Errorf("with the factory reachable: fallback %v, fallback edges %d, gaps %v", res.FallbackOnly, res.CHAEdgesFallback, res.ActiveGaps)
	}
	_, g2, in2 := rtaFixture(t)
	in2.Unknown[0].Site = mustID(t, g2, "rt.Hidden", "unused", "()Lrt/Shape;") // nothing calls it
	res2 := ApplyRTA(g2, h, in2)
	if len(res2.FallbackOnly) != 0 || res2.ActiveGaps != nil {
		t.Errorf("factory unreachable: fallback %v, gaps %v", res2.FallbackOnly, res2.ActiveGaps)
	}
	if _, ok := g2.EdgeBetween(mustID(t, g2, "rt.Oracle", "total", "(Lrt/Shape;)D"), mustID(t, g2, "rt.Dynamic", "area", "()D")); ok {
		t.Error("Dynamic#area kept without a reachable creation or gap")
	}
}

func mustID(t *testing.T, g *Graph, owner, name, desc string) uint32 {
	t.Helper()
	id, ok := g.ID(owner, name, desc)
	if !ok {
		t.Fatalf("no node %s#%s%s", owner, name, desc)
	}
	return id
}

// TestStrategyForRTA: the mode is recognized and records its policy.
func TestStrategyForRTA(t *testing.T) {
	s, err := StrategyFor(ModeRTA)
	if err != nil || s.Mode() != ModeRTA || s.UnknownAllocationPolicy() != PolicyCHAFallbackWithGap {
		t.Fatalf("%v %v", s, err)
	}
}
