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
)

// TestCreatedTypesOracle (GRAPH_PRECISION P1): the created-type catalog
// against what the real JVM created in testdata/alloc
// (scripts/alloc-oracle.sh). Every class the JVM created is in the
// catalog with the origin of the mechanism that created it, or — a
// generated proxy — covered by a gap of its kind; every catalog class the
// JVM did not create is listed with why. Runs the whole CLI (models
// included) on the fixture, offline, without a JDK.
func TestCreatedTypesOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	os.MkdirAll(filepath.Join(dir, "deps"), 0o755)
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", "../../testdata/alloc/alloc.jar",
		"-deps", filepath.Join(dir, "deps"), "-profile", "generic", "-out", filepath.Join(dir, "r.md"),
		"-json", filepath.Join(dir, "r.json"), "-json-format", "v2", "-graph-out", graph}
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
	origins := map[string]string{}
	for _, tr := range ex.Types {
		if strings.HasPrefix(tr.Class, "al.") {
			origins[tr.Class] = strings.Join(tr.Origins, ",")
		}
	}
	var doc struct {
		CallGraph struct {
			CreatedTypes struct {
				Gaps map[string]int `json:"gaps_by_kind"`
			} `json:"created_types"`
		} `json:"callgraph"`
	}
	b, _ := os.ReadFile(filepath.Join(dir, "r.json"))
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	obs, err := os.Open("../../testdata/alloc/observed.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer obs.Close()
	observed := map[string]bool{}
	sc := bufio.NewScanner(obs)
	for sc.Scan() {
		if l := sc.Text(); l != "" && !strings.HasPrefix(l, "#") {
			observed[l] = true
		}
	}
	// The origin each observed class must have (the mechanism that created it).
	wantOrigin := map[string]string{
		"al.Square":     "new=",
		"al.BigSquare":  "new=",
		"al.Circle":     "new=",
		"al.Hexagon":    "reflection=",
		"al.PluginImpl": "model=", // ServiceLoader providers are materialized by the B3 service model
		"al.Saved":      "deserialization:java=",
	}
	for cl := range observed {
		if cl == "<proxy>" {
			if doc.CallGraph.CreatedTypes.Gaps["dynamic-proxy"] == 0 {
				t.Error("the JVM created a dynamic proxy, but the catalog has no dynamic-proxy gap")
			}
			continue
		}
		o, ok := origins[cl]
		if !ok {
			t.Errorf("the JVM created %s, the catalog does not have it", cl)
			continue
		}
		if !strings.Contains(o, wantOrigin[cl]) {
			t.Errorf("%s: origins %s, want %s…", cl, o, wantOrigin[cl])
		}
	}
	// Catalog classes the JVM did not create, with why.
	wantExtras := map[string]string{
		"al.Never": "allocated only in a method nothing calls: the catalog is not reachability-aware (RTA, P2, is)",
	}
	var extras []string
	for cl := range origins {
		if !observed[cl] {
			extras = append(extras, cl)
		}
	}
	sort.Strings(extras)
	for _, cl := range extras {
		if _, ok := wantExtras[cl]; !ok {
			t.Errorf("catalog class %s (%s) was not created on the JVM and is not an expected extra", cl, origins[cl])
		}
	}
	if !contains(extras, "al.Never") {
		t.Errorf("al.Never (an unexecuted `new`) should be in the catalog: extras %v", extras)
	}
	if _, ok := origins["al.AbstractShape"]; ok {
		t.Error("an abstract class is in the catalog")
	}
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
