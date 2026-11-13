package reach

import (
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// TestPreferPathWithoutCHAOnlySteps (ROADMAP B0 b0-2), over the call-graph
// and RTA oracle fixtures (virtual dispatch, inheritance, interfaces): for
// every method, the verdict is the one the plain BFS passes give, and the
// reported path has no CHA-only step whenever such a path exists — so a
// `heuristic` character means reachability depends on a CHA expansion.
func TestPreferPathWithoutCHAOnlySteps(t *testing.T) {
	for _, jar := range []string{"../../testdata/callgraph/cg.jar", "../../testdata/rta/rta.jar"} {
		h, _, err := hierarchy.Build(jar, nil)
		if err != nil {
			t.Fatal(err)
		}
		g := callgraph.Build(h, nil)
		// Roots: the fixtures' main methods only (every app method would
		// make each one a root of its own).
		var roots []uint32
		for id := uint32(0); int(id) < g.NodeCount(); id++ {
			if strings.HasSuffix(g.Key(id), "#main([Ljava/lang/String;)V") {
				roots = append(roots, id)
			}
		}
		if len(roots) == 0 {
			t.Fatalf("%s: no main method", jar)
		}
		e := NewWithRoots(g, h, NewFrameworkMatcher(nil), roots)
		strictOnly, heuristic := 0, 0
		for id := uint32(0); int(id) < g.NodeCount(); id++ {
			var r Result
			e.classify(&r, []uint32{id})
			want := StatusNotReachable
			if e.appOnly.visited[id] {
				want = StatusReachable
			} else if e.full.visited[id] {
				want = StatusFramework
			}
			if r.Status != want {
				t.Fatalf("%s %s: status %s, want %s", jar, g.Key(id), r.Status, want)
			}
			if r.Status == StatusNotReachable {
				continue
			}
			steps, chars := pathEvidence(g, r.Path)
			hasHeur := false
			for _, c := range chars {
				if c == CharHeuristic {
					hasHeur = true
				}
			}
			strictAvail := (r.Status == StatusReachable && e.appOnlyNoCHA.visited[id]) || (r.Status == StatusFramework && e.fullNoCHA.visited[id])
			if strictAvail && hasHeur {
				t.Errorf("%s %s: a path without CHA-only steps exists, reported %v", jar, g.Key(id), steps)
			}
			if !strictAvail {
				if !hasHeur && len(r.Path) > 1 {
					t.Errorf("%s %s: no CHA-free path, yet the reported one has no heuristic step: %v", jar, g.Key(id), chars)
				}
				heuristic++
			} else {
				strictOnly++
			}
		}
		if heuristic == 0 || strictOnly == 0 {
			t.Errorf("%s: fixture exercises only one case (CHA-free %d, CHA-dependent %d)", jar, strictOnly, heuristic)
		}
	}
}
