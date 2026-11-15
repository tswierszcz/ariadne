package callgraph

import "testing"

// TestPruneCHASites: only CHA edges lose sites; an edge with no site left
// is removed and counted, an edge with a site left keeps it, and the CHA
// total follows; other bases are untouched whatever the predicate says.
func TestPruneCHASites(t *testing.T) {
	_, g, _ := rtaFixture(t)
	before, cha := 0, g.CHAEdgeCount()
	nonCHA := 0
	for id := range g.adjacency {
		for _, e := range g.adjacency[id] {
			before++
			if e.Basis != BasisCHA {
				nonCHA++
			}
		}
	}
	if cha == 0 {
		t.Fatal("fixture has no CHA edge")
	}
	// Drop every CHA site: every CHA edge goes, nothing else.
	edges, sites := g.PruneCHASites(func(uint32, Edge, int) bool { return true })
	if edges != cha || sites < edges || g.CHAEdgeCount() != 0 {
		t.Errorf("removed %d edges / %d sites, CHA left %d; want %d edges, ≥ that many sites, 0 left", edges, sites, g.CHAEdgeCount(), cha)
	}
	after := 0
	for id := range g.adjacency {
		for _, e := range g.adjacency[id] {
			after++
			if e.Basis == BasisCHA {
				t.Errorf("CHA edge left: %s -> %s", g.Key(uint32(id)), g.Key(e.To))
			}
		}
	}
	if after != nonCHA || before-after != cha {
		t.Errorf("edges %d -> %d, non-CHA %d", before, after, nonCHA)
	}

	// Keeping every site removes nothing.
	_, g2, _ := rtaFixture(t)
	if e, s := g2.PruneCHASites(func(uint32, Edge, int) bool { return false }); e != 0 || s != 0 || g2.CHAEdgeCount() != cha {
		t.Errorf("keep-all removed %d/%d", e, s)
	}
}

// TestStrategyForPTA: "pta" is a mode of its own, built as RTA.
func TestStrategyForPTA(t *testing.T) {
	s, err := StrategyFor(ModePTA)
	if err != nil || s.Mode() != ModePTA || s.UnknownAllocationPolicy() != PolicyCHAFallbackWithGap {
		t.Fatalf("StrategyFor(pta) = %v, %v", s, err)
	}
}
