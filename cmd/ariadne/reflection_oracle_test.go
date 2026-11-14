package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/reflectinvoke"
)

const reflectionFixture = "../../testdata/reflection/"

// TestReflectionOracle (REFLECTION R0–R2, ROADMAP B4): testdata/reflection
// (scripts/reflection-oracle.sh) runs one case per method on the real JVM
// and records the bodies that ran. Positive cases — forName and getMethod
// literals, a lookup on a supertype with the target allocated as the
// subtype, a static method with a null target, getDeclaredMethod of a
// private method, a Method cached in a static field, Constructor#newInstance
// — must be reachable from main through reflection edges to exactly the
// bodies the JVM ran. Negative cases — a proven argument-count mismatch
// (the JVM throws), and the supertype body the JVM did not select — must
// have no reflection edge. The unresolved case (a method name from the
// input) must be reported as a gap and have no invented target; it is the
// only body the JVM ran that the graph does not reach.
func TestReflectionOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", reflectionFixture + "app.jar",
		"-profile", "generic", "-roots", "profile",
		"-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-graph-out", graph}
	if code, out := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, out)
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

	const unresolved = "fx.r.Target#chosenAtRunTime()V"
	reach := reachableFrom(ex, "fx.r.Main#main([Ljava/lang/String;)V")
	obs := (benchFixture{observed: reflectionFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	if len(obs) != 8 || !obs[unresolved] {
		t.Fatalf("observed.txt has %d bodies (%v); regenerate with scripts/reflection-oracle.sh", len(obs), obs)
	}
	for b := range obs {
		switch {
		case b == unresolved && reach[b]:
			t.Errorf("%s is reached: the unresolved case resolved — update this oracle", b)
		case b != unresolved && !reach[b]:
			t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
		}
	}

	var got []string
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisReflection) {
			got = append(got, id.From+" -> "+id.To)
		}
	}
	sort.Strings(got)
	want := []string{
		"fx.r.Main#cachedInField()V -> fx.r.Target#viaField()V",
		"fx.r.Main#constructor()V -> fx.r.Made#<init>()V",
		"fx.r.Main#declaredPrivate()V -> fx.r.Target#secret()V",
		"fx.r.Main#literal()V -> fx.r.Target#hello()V",
		"fx.r.Main#staticWithNull()V -> fx.r.Target#stat(Ljava/lang/String;)Ljava/lang/String;",
		"fx.r.Main#supertypeLookup()V -> fx.r.Sub#greet()V",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("reflection edges:\n%s\nwant (every one a body the JVM ran, and no other):\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	for _, e := range got {
		if to := e[strings.Index(e, " -> ")+4:]; !obs[to] {
			t.Errorf("reflection edge to a body the JVM never ran: %s", e)
		}
	}

	// The two cases without an edge are reported, each with its reason.
	h, _, err := hierarchy.Build(reflectionFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	resolved, _ := reflectinvoke.ResolveInvokes(h, reflectinvoke.FindInvokeSites(h))
	gaps := map[string]string{}
	for _, ri := range resolved {
		if ri.OwnerFQCN == "" {
			gaps[ri.Site.CallerName] = "unresolved: " + ri.UnresolvedReason
		}
	}
	for _, r := range reflectinvoke.FindInvokeEdges(h, resolved) {
		if r.Gap != "" {
			gaps[r.ResolvedInvoke.Site.CallerName] = r.Gap
		}
	}
	if len(gaps) != 2 || !strings.HasPrefix(gaps["nameFromInput"], "unresolved: ") || !strings.Contains(gaps["wrongArity"], "argument-count mismatch") {
		t.Errorf("gaps = %v, want nameFromInput unresolved and wrongArity's argument-count mismatch", gaps)
	}
}
