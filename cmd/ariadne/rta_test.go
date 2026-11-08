package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// rtaRun runs the CLI on testdata/rta with mode and roots and returns the
// exported graph and the v2 call-graph section.
func rtaRun(t *testing.T, mode, roots string) (callgraph.Export, map[string]json.RawMessage) {
	t.Helper()
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	os.MkdirAll(filepath.Join(dir, "deps"), 0o755)
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", "../../testdata/rta/rta.jar",
		"-deps", filepath.Join(dir, "deps"), "-profile", "generic", "-roots", roots, "-callgraph-mode", mode,
		"-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-json-format", "v2", "-graph-out", graph}
	if code, out := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	f, err := os.Open(graph)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ex, err := callgraph.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		CallGraph map[string]json.RawMessage `json:"callgraph"`
	}
	b, _ := os.ReadFile(filepath.Join(dir, "r.json"))
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	return ex, doc.CallGraph
}

// reachableFrom: the methods the exported graph reaches from root.
func reachableFrom(ex callgraph.Export, root string) map[string]bool {
	adj := map[string][]string{}
	for id := range ex.Edges {
		adj[id.From] = append(adj[id.From], id.To)
	}
	seen := map[string]bool{root: true}
	queue := []string{root}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, n := range adj[cur] {
			if !seen[n] {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	return seen
}

func chaEdges(ex callgraph.Export) map[string]bool {
	out := map[string]bool{}
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisCHA) {
			out[id.From+" -> "+id.To] = true
		}
	}
	return out
}

func observedBodies(t *testing.T) []string {
	t.Helper()
	f, err := os.Open("../../testdata/rta/observed.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if l := sc.Text(); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out
}

// TestRTAOracle (GRAPH_PRECISION P2): the RTA graph against the method
// bodies the real JVM ran in testdata/rta (scripts/rta-oracle.sh): three
// implementations called through an interface (one created only in a
// factory), a dispatch cycle whose allocations are discovered late, a
// ServiceLoader provider, and two classes from a reflective factory the
// analysis cannot resolve. Every body the JVM ran stays reachable from
// main, except one expected miss that an active unknown-allocation gap
// covers; every CHA edge RTA removes is listed with why. Runs the whole
// CLI offline, without a JDK.
func TestRTAOracle(t *testing.T) {
	const main = "rt.Oracle#main([Ljava/lang/String;)V"
	exCHA, _ := rtaRun(t, callgraph.ModeCHA, "profile")
	ex, v2 := rtaRun(t, callgraph.ModeRTA, "profile")
	if ex.Config.Mode != callgraph.ModeRTA || ex.Config.UnknownAllocationPolicy != callgraph.PolicyCHAFallbackWithGap {
		t.Fatalf("config %+v", ex.Config)
	}
	var res callgraph.RTAResult
	if err := json.Unmarshal(v2["rta"], &res); err != nil {
		t.Fatalf("v2 callgraph.rta: %v", err)
	}

	// Expected misses: a body the JVM ran that RTA does not reach, each
	// covered by an active gap of the given kind.
	wantMiss := map[string]string{
		"rt.Unnamed#area()D": "reflection-unresolved", // created by a reflective factory and named nowhere the analysis sees
	}
	reach := reachableFrom(ex, main)
	reachCHA := reachableFrom(exCHA, main)
	for _, b := range observedBodies(t) {
		if !reachCHA[b] {
			t.Errorf("CHA does not reach %s (the baseline is broken)", b)
		}
		if reach[b] {
			continue
		}
		kind, ok := wantMiss[b]
		if !ok {
			t.Errorf("the JVM ran %s, RTA does not reach it", b)
			continue
		}
		if res.ActiveGaps[kind] == 0 {
			t.Errorf("%s is missed but no %s gap is active (gaps %v)", b, kind, res.ActiveGaps)
		}
	}

	// Removed edges, each with why.
	wantRemoved := map[string]string{
		"rt.Oracle#total(Lrt/Shape;)D -> rt.Ghost#area()D":   "Ghost is never created",
		"rt.Oracle#total(Lrt/Shape;)D -> rt.Hidden#area()D":  "Hidden is created only in a method nothing calls",
		"rt.Oracle#total(Lrt/Shape;)D -> rt.Unnamed#area()D": "the expected miss above",
	}
	before, after := chaEdges(exCHA), chaEdges(ex)
	var removed []string
	for e := range before {
		if !after[e] {
			removed = append(removed, e)
		}
	}
	for e := range after {
		if !before[e] {
			t.Errorf("RTA added a CHA edge %s", e)
		}
	}
	sort.Strings(removed)
	for _, e := range removed {
		if _, ok := wantRemoved[e]; !ok {
			t.Errorf("RTA removed %s, not an expected removal", e)
		}
	}
	for e, why := range wantRemoved {
		if after[e] {
			t.Errorf("RTA kept %s (%s)", e, why)
		}
	}
	if res.CHAEdgesRemoved != len(removed) || res.CHAEdgesBefore != len(before) {
		t.Errorf("result counts %d of %d removed, graphs say %d of %d", res.CHAEdgesRemoved, res.CHAEdgesBefore, len(removed), len(before))
	}
	// Dynamic is created by the same unresolved factory: kept only by the
	// fallback (an annotated class).
	if !after["rt.Oracle#total(Lrt/Shape;)D -> rt.Dynamic#area()D"] || res.CHAEdgesFallback == 0 {
		t.Errorf("Dynamic#area should be kept by the fallback (fallback edges %d)", res.CHAEdgesFallback)
	}

	// roots=all is a separate scenario: a root's Shape parameter may be
	// any Shape (external input), so nothing is removed.
	exAll, v2All := rtaRun(t, callgraph.ModeRTA, "all")
	var resAll callgraph.RTAResult
	json.Unmarshal(v2All["rta"], &resAll)
	if resAll.CHAEdgesRemoved != 0 || resAll.ActiveGaps["external-input"] == 0 {
		t.Errorf("roots=all: %d removed, gaps %v; want 0 removed and an active external-input gap", resAll.CHAEdgesRemoved, resAll.ActiveGaps)
	}
	if len(exAll.Edges) != len(exCHA.Edges) {
		t.Errorf("roots=all: %d edges, CHA %d", len(exAll.Edges), len(exCHA.Edges))
	}
}

// TestNameableClasses: the unknown-allocation fallback of testdata/rta
// holds the annotated class and the class-literal one, not a class named
// only by a string built at run time nor one only `new` names.
func TestNameableClasses(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/rta/rta.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, c := range nameableClasses(h) {
		got[c] = true
	}
	for c, want := range map[string]bool{
		"rt.Dynamic": true,  // @Scanned
		"rt.Unnamed": false, // its name is built at run time
		"rt.Ghost":   false, // never named
		"rt.Square":  false, // only `new Square()`: not nameable, but created
	} {
		if got[c] != want {
			t.Errorf("%s nameable %v, want %v", c, got[c], want)
		}
	}
	if got["rt.Shape"] || got["rt.Scanned"] {
		t.Error("an interface/annotation type is in the fallback (only concrete classes can be created)")
	}
}
