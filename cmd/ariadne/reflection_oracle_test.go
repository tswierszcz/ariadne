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

// TestReflectionOracle (REFLECTION R0–R2, R4, ROADMAP B4): testdata/reflection
// (scripts/reflection-oracle.sh) runs one case per method on the real JVM
// and records the bodies that ran. Positive cases — forName and getMethod
// literals, a lookup on a supertype with the target allocated as the
// subtype, a static method with a null target, getDeclaredMethod of a
// private method, a Method cached in a static field, Constructor#newInstance,
// Class.forName initializing a class (directly, through its superclass, with
// a non-constant initialize) — must be reachable from main through reflection edges to exactly the
// bodies the JVM ran. Negative cases — a proven argument-count mismatch
// (the JVM throws), the supertype body the JVM did not select, and
// Class.forName with initialize=false — must have no reflection edge; nor
// may R4's: a lookup alone, getMethod of a non-public method,
// getDeclaredMethod of an inherited one, getConstructor of a non-public
// constructor. privateWithoutAccess keeps an edge the JVM rejects
// (setAccessible is not tracked) — the one stated over-approximation. The unresolved case (a method name from the
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
	if len(obs) != 11 || !obs[unresolved] {
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
		"fx.r.Main#forNameInit()V -> fx.r.Inited#<clinit>()V",
		"fx.r.Main#forNameMaybe(Z)V -> fx.r.MaybeInited#<clinit>()V",
		"fx.r.Main#forNameSuper()V -> fx.r.InitedBase#<clinit>()V",
		"fx.r.Main#literal()V -> fx.r.Target#hello()V",
		"fx.r.Main#privateWithoutAccess()V -> fx.r.Target#locked()V",
		"fx.r.Main#staticWithNull()V -> fx.r.Target#stat(Ljava/lang/String;)Ljava/lang/String;",
		"fx.r.Main#supertypeLookup()V -> fx.r.Sub#greet()V",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("reflection edges:\n%s\nwant (every one a body the JVM ran, and no other):\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	// R4: setAccessible is not tracked, so a private method invoked without
	// it keeps its edge although the JVM throws IllegalAccessException — the
	// one stated over-approximation (an extra edge, never a missed one).
	const overApprox = "fx.r.Target#locked()V"
	for _, e := range got {
		if to := e[strings.Index(e, " -> ")+4:]; !obs[to] && to != overApprox {
			t.Errorf("reflection edge to a body the JVM never ran: %s", e)
		}
	}

	// R4 r4-1: initialize=false loads the class without initializing it.
	if notInit := "fx.r.NotInited#<clinit>()V"; reach[notInit] || obs[notInit] {
		t.Errorf("%s: reached %v, ran %v — want neither (Class.forName with initialize=false)", notInit, reach[notInit], obs[notInit])
	}

	// R4: a lookup alone, and lookups the JVM rejects with
	// NoSuchMethodException, never reach the body.
	for _, b := range []string{"fx.r.Target#lookedUp()V", "fx.r.Target#packagePrivate()V", "fx.r.Base#baseOnly()V", "fx.r.Hidden#<init>()V"} {
		if reach[b] || obs[b] {
			t.Errorf("%s: reached %v, ran %v — want neither", b, reach[b], obs[b])
		}
	}

	// The cases without an edge are reported, each with its reason.
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
	if len(gaps) != 4 || !strings.HasPrefix(gaps["nameFromInput"], "unresolved: ") || !strings.Contains(gaps["wrongArity"], "argument-count mismatch") ||
		!strings.Contains(gaps["getMethodNotPublic"], "public methods only") || !strings.Contains(gaps["declaredInherited"], "is inherited from fx.r.Base") {
		t.Errorf("gaps = %v, want nameFromInput unresolved, wrongArity's argument-count mismatch and the two rejected lookups", gaps)
	}
	nr, _ := reflectinvoke.ResolveNewInstances(h, reflectinvoke.FindNewInstanceSites(h))
	ctorGaps := map[string]string{}
	for _, r := range reflectinvoke.FindNewInstanceEdges(h, nr) {
		if r.Gap != "" {
			ctorGaps[r.ResolvedNewInstance.Site.CallerName] = r.Gap
		}
	}
	if len(ctorGaps) != 1 || !strings.Contains(ctorGaps["constructorNotPublic"], "public constructors only") {
		t.Errorf("constructor gaps = %v, want constructorNotPublic's non-public constructor", ctorGaps)
	}
}
