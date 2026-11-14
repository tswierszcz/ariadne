package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callback"
	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/methodhandle"
)

const asyncFixture = "../../testdata/async/"

// TestAsyncOracle (ASYNC_CALLBACKS A0–A2, DYNAMIC_LINKAGE L2 MethodHandle;
// ROADMAP B4): testdata/async (scripts/async-oracle.sh) runs one case per
// method on the real JVM and records the bodies that ran, from any thread.
// Positive cases — a Thread with a Runnable allocated in place, lambdas
// handed to an ExecutorService, a Callable to submit, a synchronous
// AccessController#doPrivileged action, MethodHandles.Lookup
// findVirtual/findStatic with literals — must be reachable from main
// through callback / methodhandle edges, each edge to a body the JVM ran
// (a Callable/PrivilegedAction edge lands on the bridge, which calls it).
// Negative cases — a Runnable allocated but never registered, a find* of a
// method that does not exist (the JVM throws) — have no edge. The
// unresolved cases — a callback reaching the registration through a
// parameter, a handle whose method name comes from the input — are
// reported as gaps and are the only bodies the JVM ran that the graph
// misses.
func TestAsyncOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", asyncFixture + "app.jar",
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

	unresolved := map[string]bool{"fx.a.FromParam#run()V": true, "fx.a.Handled#chosenAtRunTime()V": true}
	reach := reachableFrom(ex, "fx.a.Main#main([Ljava/lang/String;)V")
	obs := (benchFixture{observed: asyncFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	if len(obs) != 10 {
		t.Fatalf("observed.txt has %d bodies (%v); regenerate with scripts/async-oracle.sh", len(obs), obs)
	}
	for b := range obs {
		switch {
		case unresolved[b] && reach[b]:
			t.Errorf("%s is reached: an unresolved case resolved — update this oracle", b)
		case !unresolved[b] && !reach[b]:
			t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
		}
	}
	if reach["fx.a.Unused#run()V"] {
		t.Error("fx.a.Unused#run is reached, but it is never registered and the JVM never ran it")
	}

	var got []string
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisCallback) || id.Basis == string(callgraph.BasisMethodHandle) {
			got = append(got, id.Basis+" "+id.From+" -> "+id.To)
		}
	}
	sort.Strings(got)
	want := []string{
		"callback fx.a.Main#callable()V -> fx.a.Job#call()Ljava/lang/Object;",
		"callback fx.a.Main#lambdaToExecutor()V -> fx.a.Main#lambda$lambdaToExecutor$0()V",
		"callback fx.a.Main#lambdaToExecutor()V -> fx.a.Main#lambda$lambdaToExecutor$1()V",
		"callback fx.a.Main#privileged()V -> fx.a.Action#run()Ljava/lang/Object;",
		"callback fx.a.Main#thread()V -> fx.a.Task#run()V",
		"methodhandle fx.a.Main#handleStatic()V -> fx.a.Handled#stat(I)I",
		"methodhandle fx.a.Main#handleVirtual()V -> fx.a.Handled#virt()V",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("callback/methodhandle edges:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	// Every target ran on the JVM — directly, or as the bridge whose
	// direct callee in the same class ran.
	for _, e := range got {
		to := e[strings.Index(e, " -> ")+4:]
		if obs[to] {
			continue
		}
		owner := to[:strings.Index(to, "#")+1]
		ran := false
		for id := range ex.Edges {
			if id.From == to && id.Basis == string(callgraph.BasisDirect) && strings.HasPrefix(id.To, owner) && obs[id.To] {
				ran = true
			}
		}
		if !ran {
			t.Errorf("edge to a body the JVM never ran: %s", e)
		}
	}

	// The cases without an edge are reported, each with its reason.
	h, _, err := hierarchy.Build(asyncFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	cbs, _ := callback.ResolveCallbacks(h, callback.FindRegistrationSites(h))
	cbGaps := map[string]string{}
	for _, c := range cbs {
		if c.UnresolvedReason != "" {
			cbGaps[c.Site.CallerName] = c.UnresolvedReason
		}
	}
	if len(cbGaps) != 1 || cbGaps["runOn"] == "" {
		t.Errorf("callback gaps = %v, want only runOn (the callback is a parameter)", cbGaps)
	}
	handles, _ := methodhandle.ResolveInvokes(h, methodhandle.FindInvokeSites(h))
	hGaps := map[string]string{}
	for _, r := range handles {
		if r.RequestedOwner == "" {
			hGaps[r.Site.CallerName] = "unresolved: " + r.UnresolvedReason
		}
	}
	for _, r := range methodhandle.FindHandleEdges(h, handles) {
		if r.Gap != "" {
			hGaps[r.ResolvedInvoke.Site.CallerName] = r.Gap
		}
	}
	if len(hGaps) != 2 || !strings.HasPrefix(hGaps["handleFromInput"], "unresolved: ") || !strings.Contains(hGaps["handleMissing"], `"missing"`) {
		t.Errorf("method-handle gaps = %v, want handleFromInput unresolved and handleMissing's missing method", hGaps)
	}
}
