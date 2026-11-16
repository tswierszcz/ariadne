package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
)

// TestCutSitesAfter: a cut drops the method's sites past the offset from
// edges of every basis (an edge left with none goes, and the totals
// follow), keeps the earlier ones, and refuses a site past the offset
// that a model adds afterwards, through every way edges are added.
func TestCutSitesAfter(t *testing.T) {
	_, g, _ := rtaFixture(t)
	// The method with the most call sites, and an offset in the middle.
	var from uint32
	var pcs []int
	for id := range g.adjacency {
		var ps []int
		for _, e := range g.adjacency[id] {
			for _, s := range e.Sites {
				ps = append(ps, s.PC)
			}
		}
		if len(ps) > len(pcs) {
			from, pcs = uint32(id), ps
		}
	}
	if len(pcs) < 2 {
		t.Fatal("fixture has no method with two call sites")
	}
	lo, hi := pcs[0], pcs[0]
	for _, pc := range pcs {
		lo, hi = min(lo, pc), max(hi, pc)
	}
	if lo == hi {
		t.Fatal("fixture method's sites share one offset")
	}
	cut := lo
	total := g.EdgeCount()
	sites, edges := g.CutSitesAfter(from, cut)
	if sites == 0 || g.EdgeCount() != total-edges {
		t.Errorf("removed %d sites / %d edges; edge count %d -> %d", sites, edges, total, g.EdgeCount())
	}
	for _, e := range g.adjacency[from] {
		for _, s := range e.Sites {
			if s.PC > cut {
				t.Errorf("site at pc %d left after a cut at %d", s.PC, cut)
			}
		}
	}
	if after, ok := g.SiteCut(from); !ok || after != cut {
		t.Errorf("SiteCut = %d, %v", after, ok)
	}

	// Later additions past the cut are refused, earlier ones are not.
	to := g.adjacency[from][0].To
	if g.addEdge(from, to, classfile.CallSynthetic, BasisCallback, &CallSite{PC: hi + 1}) {
		t.Error("addEdge past the cut was accepted")
	}
	if !g.addEdge(from, to, classfile.CallSynthetic, BasisCallback, &CallSite{PC: cut}) {
		t.Error("addEdge at the cut was refused")
	}
	key := g.Key(from)
	cls, rest, _ := cutKey(key)
	name, desc := splitNameDesc(rest)
	tcls, trest, _ := cutKey(g.Key(to))
	tname, tdesc := splitNameDesc(trest)
	if n := ApplyModelEdges(g, []ModelEdge{{CallerFQCN: cls, CallerName: name, CallerDesc: desc, TargetFQCN: tcls, TargetName: tname, TargetDesc: tdesc, PC: hi + 1}}); n != 0 {
		t.Errorf("ApplyModelEdges past the cut added %d", n)
	}

	// A wider cut on the same method changes nothing.
	if s, e := g.CutSitesAfter(from, hi); s != 0 || e != 0 {
		t.Errorf("a later offset removed %d/%d", s, e)
	}
}

func cutKey(k string) (string, string, bool) {
	for i := 0; i < len(k); i++ {
		if k[i] == '#' {
			return k[:i], k[i+1:], true
		}
	}
	return k, "", false
}

func splitNameDesc(s string) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			return s[:i], s[i:]
		}
	}
	return s, ""
}
