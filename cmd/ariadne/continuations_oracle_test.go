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
)

const contFixture = "../../testdata/continuations/"

// TestContinuationsOracle (ASYNC_CALLBACKS A3 slice 1): testdata/
// continuations (scripts/continuations-oracle.sh) runs CompletionStage
// pipelines on the real JVM and records every callback body that ran.
// Each must be reached from main. A lambda or method reference already was
// (the edge from the invokedynamic that creates it); a Runnable class
// handed to runAsync was not — the gap this slice closes. Every
// registration with a lambda, a method reference or an allocated class
// now also has a callback edge to it; the one receiving its callback
// through a parameter is a reported gap, as in A1. The callbacks the JVM did not run here — a
// normal-branch function on a failed stage, an exceptionally on a
// succeeding one, a continuation of a future nothing completes — keep
// their "may run" edge: the model does not decide completion.
func TestContinuationsOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", contFixture + "app.jar",
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
	reached := reachableFrom(ex, "fx.c.Main#main([Ljava/lang/String;)V")
	obs := (benchFixture{observed: contFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	if len(obs) != 12 {
		t.Fatalf("observed.txt has %d bodies; regenerate with scripts/continuations-oracle.sh", len(obs))
	}
	var missed []string
	for b := range obs {
		if !reached[b] {
			missed = append(missed, b)
		}
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
	}

	var got []string
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisCallback) && strings.HasPrefix(id.From, "fx.c.Main#") {
			got = append(got, id.From[len("fx.c.Main#"):strings.Index(id.From, "(")]+" -> "+id.To)
		}
	}
	sort.Strings(got)
	want := []string{
		"async -> fx.c.Main#after()V",
		"async -> fx.c.Task#run()V",
		"chain -> fx.c.Main#lambda$chain$0()Ljava/lang/String;",
		"chain -> fx.c.Main#lambda$chain$1(Ljava/lang/String;)V",
		"chain -> fx.c.Main#upper(Ljava/lang/String;)Ljava/lang/String;",
		"combine -> fx.c.Main#lambda$combine$0(Ljava/lang/Integer;Ljava/lang/Integer;)Ljava/lang/Integer;",
		"compose -> fx.c.Main#next(Ljava/lang/Integer;)Ljava/util/concurrent/CompletableFuture;",
		"failure -> fx.c.Main#lambda$failure$0()Ljava/lang/String;",
		"failure -> fx.c.Main#lambda$failure$1(Ljava/lang/String;)Ljava/lang/String;",
		"failure -> fx.c.Main#lambda$failure$2(Ljava/lang/Throwable;)Ljava/lang/String;",
		"failure -> fx.c.Main#lambda$failure$3(Ljava/lang/String;Ljava/lang/Throwable;)Ljava/lang/String;",
		"failure -> fx.c.Main#lambda$failure$4(Ljava/lang/String;Ljava/lang/Throwable;)V",
		"failure -> fx.c.Main#lambda$failure$5(Ljava/lang/Throwable;)Ljava/lang/String;",
		"never -> fx.c.Main#lambda$never$0(Ljava/lang/String;)V",
		"success -> fx.c.Main#lambda$success$0(Ljava/lang/Throwable;)Ljava/lang/String;",
		"throughInterface -> fx.c.Main#lambda$throughInterface$0(Ljava/lang/String;)V",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("callback edges:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	// The registration in attach gets its callback through a parameter: a
	// reported gap (the lambda itself is reached from where it is created).
	h, _, err := hierarchy.Build(contFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	cbs, _ := callback.ResolveCallbacks(h, callback.FindRegistrationSites(h))
	gaps := map[string]string{}
	conds := map[string]string{}
	for _, c := range cbs {
		if c.UnresolvedReason != "" {
			gaps[c.Site.CallerName] = c.UnresolvedReason
		}
		if c.Site.Model.Condition != "" {
			conds[c.Site.CallerName+" "+c.Site.Model.Name] = c.Site.Model.Condition
		}
	}
	if len(gaps) != 1 || gaps["attach"] == "" {
		t.Errorf("callback gaps = %v, want only attach (the callback is a parameter)", gaps)
	}
	wantConds := map[string]string{"failure exceptionally": callback.ConditionExceptional, "success exceptionally": callback.ConditionExceptional}
	if len(conds) != len(wantConds) || conds["failure exceptionally"] != wantConds["failure exceptionally"] || conds["success exceptionally"] != wantConds["success exceptionally"] {
		t.Errorf("conditional registrations = %v, want %v", conds, wantConds)
	}
}
