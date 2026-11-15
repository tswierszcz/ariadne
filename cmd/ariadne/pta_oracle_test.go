package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

const ptaFixture = "../../testdata/pta/"

// areaTargets runs the fixture in mode and returns, per caller, the
// Shape#area implementations its edges reach.
func areaTargets(t *testing.T, mode string) map[string][]string {
	t.Helper()
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", ptaFixture + "app.jar",
		"-profile", "generic", "-roots", "profile", "-callgraph-mode", mode,
		"-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-graph-out", graph}
	if code, out := runWithEnv(args, nil); code != 0 {
		t.Fatalf("%s: exit %d: %s", mode, code, out)
	}
	f, err := os.Open(graph)
	if err != nil {
		t.Fatal(err)
	}
	ex, err := callgraph.Read(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
	for id := range ex.Edges {
		if strings.HasSuffix(id.To, "#area()D") && id.To != "fx.app.Shape#area()D" {
			out[id.From] = append(out[id.From], strings.TrimSuffix(strings.TrimPrefix(id.To, "fx.app."), "#area()D"))
		}
	}
	for k := range out {
		sort.Strings(out[k])
	}
	return out
}

// TestPTAOracle (GRAPH_PRECISION P3 p3-1): testdata/pta
// (scripts/pta-oracle.sh) calls Shape#area from one method per case.
// Circle, Square and Hexagon are all created by reachable code, so RTA
// gives every site all three; receiver narrowing keeps exactly what the
// receiver's allocations select — two receivers, a final field written in
// the constructor, a static factory, a private method's parameter, a
// factory returning one of two allocations — and keeps all three where
// there is no proof (an array element, a non-final field, a final field
// written from a constructor parameter). Every target
// the JVM ran is kept.
func TestPTAOracle(t *testing.T) {
	rta, pta := areaTargets(t, "rta"), areaTargets(t, "pta")
	all := []string{"Circle", "Hexagon", "Square"}
	m := func(name string) string { return "fx.app.Main#" + name }
	want := map[string][]string{
		m("twoReceiversA()D"):                    {"Circle"},
		m("twoReceiversB()D"):                    {"Square"},
		m("factory()D"):                          {"Square"},
		"fx.app.Holder#area()D":                  {"Circle"},
		m("param(Lfx/app/Shape;)D"):              {"Circle", "Square"},
		m("twoAllocations(Ljava/lang/String;)D"): {"Circle", "Hexagon"},
		m("array()D"):                            all,
		m("unknown()D"):                          all,
		"fx.app.Wrapper#area()D":                 all,
	}
	for caller, w := range want {
		if !reflect.DeepEqual(rta[caller], all) {
			t.Errorf("RTA at %s: %v, want all three (the baseline P3 must improve on)", caller, rta[caller])
		}
		if !reflect.DeepEqual(pta[caller], w) {
			t.Errorf("PTA at %s: %v, want %v", caller, pta[caller], w)
		}
	}

	// Soundness: every call the JVM made is an edge in the PTA graph.
	raw, err := os.ReadFile(ptaFixture + "observed.txt")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(string(raw), "\n") {
		from, to, ok := strings.Cut(line, " -> ")
		if !ok || strings.HasPrefix(line, "#") {
			continue
		}
		n++
		target := strings.TrimSuffix(strings.TrimPrefix(to, "fx.app."), "#area()D")
		found := false
		for _, got := range pta[from] {
			found = found || got == target
		}
		if !found {
			t.Errorf("the JVM ran %s -> %s, the PTA graph has no such edge", from, to)
		}
	}
	if n < 10 {
		t.Fatalf("observed.txt has %d calls; regenerate with scripts/pta-oracle.sh", n)
	}
}
