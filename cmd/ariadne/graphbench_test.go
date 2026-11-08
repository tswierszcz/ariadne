package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

// benchFixture is a fixture with a real-JVM oracle of the method bodies
// that ran when its main ran.
type benchFixture struct {
	name, jar, observed, main string
	pkg                       string // the fixture's package prefix
	trace                     string // the method a traced body calls to record that it ran
	bodyColumn                int    // observed file: column holding the body
	// neverReached: method-key prefixes the graph must not reach in any
	// mode (servlets/filters on the class path the app never registers).
	neverReached []string
}

var benchFixtures = []benchFixture{
	{name: "callgraph", jar: "../../testdata/callgraph/cg.jar", observed: "../../testdata/callgraph/observed.tsv",
		main: "cg.Oracle#main([Ljava/lang/String;)V", pkg: "cg.", trace: "cg.Trace#hit", bodyColumn: 1},
	{name: "rta", jar: "../../testdata/rta/rta.jar", observed: "../../testdata/rta/observed.txt",
		main: "rt.Oracle#main([Ljava/lang/String;)V", pkg: "rt.", trace: "rt.Ran#here", bodyColumn: 0},
	{name: "servlet", jar: "../../testdata/servletdispatch/sd.war", observed: "../../testdata/servletdispatch/observed.txt",
		main: "sd.Oracle#main([Ljava/lang/String;)V", pkg: "sd.", trace: "sd.Ran#here", bodyColumn: 0,
		neverReached: []string{"sd.UnusedServlet#", "sd.UnusedFilter#"}},
}

func (f benchFixture) observedBodies(t *testing.T) map[string]bool {
	t.Helper()
	fh, err := os.Open(f.observed)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	out := map[string]bool{}
	sc := bufio.NewScanner(fh)
	for sc.Scan() {
		l := sc.Text()
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		out[strings.Split(l, "\t")[f.bodyColumn]] = true
	}
	return out
}

// bodies: the fixture's traced methods — those that record when they run
// (only for them does "not observed" mean "did not run").
func (f benchFixture) bodies(t *testing.T) map[string]bool {
	t.Helper()
	prepared, err := input.PrepareAppInput(f.jar) // a WAR's WEB-INF/classes
	if err != nil {
		t.Fatal(err)
	}
	defer prepared.Cleanup()
	h, _, err := hierarchy.Build(prepared.AppClassesJar, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		if !strings.HasPrefix(fqcn, f.pkg) {
			continue
		}
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Owner+"#"+c.Name == f.trace {
					out[fqcn+"#"+m.Name+m.Descriptor] = true
					break
				}
			}
		}
	}
	return out
}

func benchRun(t *testing.T, jar, mode, roots string) callgraph.Export {
	t.Helper()
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	os.MkdirAll(filepath.Join(dir, "deps"), 0o755)
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", jar, "-deps", filepath.Join(dir, "deps"),
		"-profile", "generic", "-roots", roots, "-callgraph-mode", mode, "-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-graph-out", graph}
	if code, out := runWithEnv(args, nil); code != 0 {
		t.Fatalf("%s %s %s: exit %d: %s", jar, mode, roots, code, out)
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
	return ex
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestGraphBenchmarkFixtures (GRAPH_PRECISION P4): per fixture with a
// real-JVM oracle, mode and roots scenario — the method bodies the JVM
// ran that the graph does not reach from main (missed expected targets)
// and the traced fixture bodies it reaches that the JVM did not run (extra: an
// over-approximation, or code this run did not exercise — not proven
// wrong). Logs the table (go test -v); asserts CHA misses nothing, RTA
// misses only the documented case, and RTA's extras are a subset of
// CHA's.
func TestGraphBenchmarkFixtures(t *testing.T) {
	expectedRTAMiss := map[string]bool{"rt.Unnamed#area()D": true} // a class named nowhere, created by a reflective factory (active gap)
	var table []string
	for _, fx := range benchFixtures {
		obs, all := fx.observedBodies(t), fx.bodies(t)
		for _, roots := range []string{"profile", "all"} {
			extras := map[string]map[string]bool{}
			for _, mode := range callgraph.Modes {
				reach := reachableFrom(benchRun(t, fx.jar, mode, roots), fx.main)
				missed, extra := map[string]bool{}, map[string]bool{}
				for b := range obs {
					if !reach[b] {
						missed[b] = true
					}
				}
				for b := range all {
					if reach[b] && !obs[b] {
						extra[b] = true
					}
				}
				extras[mode] = extra
				for b := range reach {
					for _, p := range fx.neverReached {
						if strings.HasPrefix(b, p) {
							t.Errorf("%s %s %s: reaches %s, which the app never registers", fx.name, roots, mode, b)
						}
					}
				}
				table = append(table, fmt.Sprintf("%-9s %-7s %-3s observed %2d  reached %2d/%2d  missed %d %v  extra %d %v",
					fx.name, roots, mode, len(obs), len(all)-countNotIn(all, reach), len(all), len(missed), sortedKeys(missed), len(extra), sortedKeys(extra)))
				for b := range missed {
					if mode == callgraph.ModeCHA || !expectedRTAMiss[b] {
						t.Errorf("%s %s %s: missed %s", fx.name, roots, mode, b)
					}
				}
			}
			for b := range extras[callgraph.ModeRTA] {
				if !extras[callgraph.ModeCHA][b] {
					t.Errorf("%s %s: RTA reaches %s, CHA does not", fx.name, roots, b)
				}
			}
		}
	}
	for _, l := range table {
		t.Log(l)
	}
}

func countNotIn(all, reach map[string]bool) int {
	n := 0
	for b := range all {
		if !reach[b] {
			n++
		}
	}
	return n
}
