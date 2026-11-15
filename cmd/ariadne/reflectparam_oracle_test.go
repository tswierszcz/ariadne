package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

const reflectParamFixture = "../../testdata/reflectparam/"

// TestReflectParamOracle (REFLECTION R3 r3-3): testdata/reflectparam
// (scripts/reflectparam-oracle.sh) resolves Method objects by literal
// lookups — kept in static fields (the Thymeleaf shape) or looked up at the
// call — and passes them to helpers that invoke them. Through their
// callers (reflectinvoke.ResolveThroughParameters) the graph reaches every
// target the JVM ran except Secret#check, whose Method one caller looks up
// by a name known only at run time: that helper keeps its model gap, with
// the resolved target (Known#check) reached anyway. A Method that is looked
// up but never passed (Unused#check) is not reached.
func TestReflectParamOracle(t *testing.T) {
	dir := t.TempDir()
	graph, out := filepath.Join(dir, "g.txt"), filepath.Join(dir, "r.json")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", reflectParamFixture + "app.jar",
		"-profile", "generic", "-roots", "profile", "-out", filepath.Join(dir, "r.md"), "-json", out, "-json-format", "v2", "-graph-out", graph}
	if code, o := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, o)
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
	all := (benchFixture{observed: reflectParamFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	const secret = "fx.app.Secret#check(Ljava/lang/Object;)Z"
	if len(all) < 10 || !all[secret] {
		t.Fatalf("observed.txt has %d bodies; regenerate with scripts/reflectparam-oracle.sh", len(all))
	}
	var missed []string
	for b := range all {
		if strings.HasPrefix(b, "fx.app.") && !reach[b] && b != secret {
			missed = append(missed, b)
		}
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
	}
	if reach[secret] {
		t.Errorf("the graph reaches %s, whose Method is looked up by a run-time name", secret)
	}
	if reach["fx.app.Unused#check(Ljava/lang/Object;)Z"] {
		t.Error("the graph reaches Unused#check, whose Method is never passed")
	}
	for id := range ex.Edges {
		if strings.HasSuffix(id.To, "#check(Ljava/lang/Object;)Z") && id.Basis != string(callgraph.BasisReflection) {
			t.Errorf("edge to a check method with basis %s, want reflection: %s -> %s", id.Basis, id.From, id.To)
		}
	}

	// The partial helper stays a model gap; the fully resolved ones do not.
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Coverage struct {
			ReflectionCensus *struct {
				Sites       int            `json:"sites"`
				ByCategory  map[string]int `json:"by_category"`
				Convertible int            `json:"convertible"`
			} `json:"reflection_census"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	c := doc.Coverage.ReflectionCensus
	if c == nil || c.Sites != 1 || c.ByCategory["parameter"] != 1 || c.Convertible != 0 {
		t.Errorf("reflection census = %+v, want the one partial parameter site (Ops#mixed), none convertible", c)
	}
}
