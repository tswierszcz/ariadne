package callgraph

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
)

const cgJar = "../../testdata/callgraph/cg.jar"

// modelTargets: the fixture methods with code the graph reaches from a
// site method within two hops (the call itself, then a bridge or a
// helper such as callSuper), excluding the trace helper.
func modelTargets(t *testing.T, g *Graph, h *hierarchy.Hierarchy, site string) map[string]bool {
	t.Helper()
	ids := g.IDsByName("cg.Sites", site)
	if len(ids) != 1 {
		t.Fatalf("site %s: %d nodes", site, len(ids))
	}
	out := map[string]bool{}
	frontier := []uint32{ids[0]}
	for hop := 0; hop < 2; hop++ {
		var next []uint32
		for _, from := range frontier {
			for _, e := range g.Edges(from) {
				k := g.Key(e.To)
				if !strings.HasPrefix(k, "cg.") || strings.HasPrefix(k, "cg.Trace#") {
					continue
				}
				owner, rest, _ := strings.Cut(k, "#")
				name := rest[:strings.IndexByte(rest, '(')]
				ci, ok := h.Lookup(owner)
				if !ok {
					continue
				}
				m, ok := ci.MethodByNameAndDesc(name, rest[len(name):])
				if !ok || len(m.Code) == 0 {
					continue // an abstract or interface declaration: no body runs there
				}
				out[k] = true
				next = append(next, e.To)
			}
		}
		frontier = next
	}
	return out
}

// TestKnownTargets (GRAPH_PRECISION P0): the graph against what the real
// JVM ran at each call site of testdata/callgraph (scripts/callgraph-oracle.sh,
// every receiver the fixture can create). No observed body may be missing
// (overloads, a default method, a generic and a covariant bridge, an
// inherited body, a super call); every extra target is listed with why.
func TestKnownTargets(t *testing.T) {
	h, _, err := hierarchy.Build(cgJar, nil)
	if err != nil {
		t.Fatal(err)
	}
	g := Build(h, nil)
	observed := map[string]map[string]bool{}
	f, err := os.Open("../../testdata/callgraph/observed.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := sc.Text()
		if strings.HasPrefix(l, "#") || l == "" {
			continue
		}
		site, m, _ := strings.Cut(l, "\t")
		if observed[site] == nil {
			observed[site] = map[string]bool{}
		}
		observed[site][m] = true
	}
	// Extras: bodies the graph connects that the JVM did not run for the
	// site, each with its reason. Intermediate bodies (a bridge, the
	// helper a super call goes through) are listed too.
	wantExtras := map[string]string{
		"bridge\tcg.StrFn#apply(Ljava/lang/Object;)Ljava/lang/Object;": "the compiler's bridge: runs, but records nothing (synthetic body)",
		"covariant\tcg.Dog#self()Lcg/Animal;":                          "the compiler's covariant bridge (synthetic body)",
		"superCall\tcg.Child#callSuper()V":                             "the helper the site calls; the super call is its own invokespecial",
	}
	var gotExtras []string
	for site, obs := range observed {
		got := modelTargets(t, g, h, site)
		for m := range obs {
			if !got[m] {
				t.Errorf("%s: the JVM ran %s, the graph has no edge to it", site, m)
			}
		}
		for m := range got {
			if !obs[m] {
				gotExtras = append(gotExtras, site+"\t"+m)
			}
		}
	}
	sort.Strings(gotExtras)
	var want []string
	for k := range wantExtras {
		want = append(want, k)
	}
	sort.Strings(want)
	if strings.Join(gotExtras, "\n") != strings.Join(want, "\n") {
		t.Errorf("extra targets:\n%s\nwant (with reasons):\n%s", strings.Join(gotExtras, "\n"), strings.Join(want, "\n"))
	}
	// Every CHA edge leads to a method its class really declares (a
	// name-only match would wire up a node for a method the subtype does
	// not have — the JVM_GRAPH G2 bug).
	for site := range observed {
		for _, e := range g.Edges(g.IDsByName("cg.Sites", site)[0]) {
			if e.Basis != BasisCHA {
				continue
			}
			owner, rest, _ := strings.Cut(g.Key(e.To), "#")
			name := rest[:strings.IndexByte(rest, '(')]
			ci, ok := h.Lookup(owner)
			if _, declared := ci.MethodByNameAndDesc(name, rest[len(name):]); !ok || !declared {
				t.Errorf("%s: CHA edge to %s, which its class does not declare", site, g.Key(e.To))
			}
		}
	}
	// The overload the String override does not override is not wired up.
	for _, e := range g.Edges(g.IDsByName("cg.Sites", "overloadInt")[0]) {
		if g.Key(e.To) == "cg.OverSub#f(Ljava/lang/String;)V" {
			t.Error("overloadInt reaches OverSub#f(String)")
		}
	}
}
