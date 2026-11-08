package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

// TestSPIDispatchOracle (model service-provider:dispatch): testdata/spidispatch
// (scripts/spidispatch-oracle.sh) — a javax.* service-provider interface
// the graph never CHA-expands, its registered provider and its product.
// Every body the JVM ran is reachable from main through model edges; the
// unregistered provider and the product implementation outside the
// provider's jar are never targets; the provider jar's never-created codec
// is a target by design (jar-bounded, not allocation-aware).
func TestSPIDispatchOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", "../../testdata/spidispatch/app.jar",
		"-deps", "../../testdata/spidispatch/deps", "-profile", "generic", "-roots", "profile",
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
	reach := reachableFrom(ex, "fx.app.Main#main([Ljava/lang/String;)V")
	obs := (benchFixture{observed: "../../testdata/spidispatch/observed.txt", bodyColumn: 0}).observedBodies(t)
	for b := range obs {
		if !reach[b] {
			t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
		}
	}
	model := map[string]bool{}
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisModel) {
			model[id.From+" -> "+id.To] = true
		}
	}
	want := []string{
		"javax.ariadnefx.Codecs#open()Ljavax/ariadnefx/Codec; -> fx.impl.Base64Provider#create()Ljavax/ariadnefx/Codec;",
		"fx.app.Main#run(Ljava/lang/String;)Ljava/lang/String; -> fx.impl.Base64Codec#encode(Ljava/lang/String;)Ljava/lang/String;",
		"fx.app.Main#run(Ljava/lang/String;)Ljava/lang/String; -> fx.impl.UnusedCodec#encode(Ljava/lang/String;)Ljava/lang/String;",
	}
	for _, e := range want {
		if !model[e] {
			t.Errorf("missing model edge %s", e)
		}
	}
	for e := range model {
		for _, bad := range []string{"fx.other.RogueCodec#", "fx.other.RogueProvider#"} {
			if strings.Contains(e, bad) {
				t.Errorf("model edge to an unregistered/out-of-jar class: %s", e)
			}
		}
	}
}
