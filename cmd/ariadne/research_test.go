package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/research"
)

// RESEARCH_INTERFACE RI1–RI3 on the control library testdata/research
// (scripts/research-fixture.sh; EXPECTED.MD there): an index without
// advisories or network, then queries answered from the bundle alone,
// checked against what the JVM ran (observed.txt).

func researchIndex(t *testing.T, dir string, extra ...string) (string, *research.Bundle) {
	t.Helper()
	out := filepath.Join(dir, "bundle")
	args := append([]string{"research", "index", "-app", "../../testdata/research/lib.jar", "-deps", "../../testdata/research/deps",
		"-deps-list", "../../testdata/research/deps.txt", "-roots", "api", "-profile", "generic",
		"-cache-dir", filepath.Join(dir, "empty-cache"), "-out", out}, extra...)
	if code, msg := runWithEnv(args, nil); code != 0 {
		t.Fatalf("research index: exit %d: %s", code, msg)
	}
	b, err := research.Load(out)
	if err != nil {
		t.Fatal(err)
	}
	return out, b
}

func researchQuery(t *testing.T, bundle string, body map[string]any) (int, map[string]any) {
	t.Helper()
	dir := t.TempDir()
	req := filepath.Join(dir, "request.json")
	raw, _ := json.Marshal(body)
	os.WriteFile(req, raw, 0o644)
	resp := filepath.Join(dir, "response.json")
	code, _ := runWithEnv([]string{"research", "query", "-bundle", bundle, "-request", req, "-out", resp}, nil)
	var doc map[string]any
	b, err := os.ReadFile(resp)
	if err != nil {
		t.Fatalf("no response document (exit %d): %v", code, err)
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	return code, doc
}

func instanceOf(b *research.Bundle, location string) string {
	for _, c := range b.Components {
		if c.Location == location {
			return c.InstanceID
		}
	}
	return ""
}

func TestResearchIndexAndQuery(t *testing.T) {
	dir := t.TempDir()
	bundle, b := researchIndex(t, dir)
	if b.Identity.Provenance.Kind != "analysis" || b.Identity.Configuration.RootsMode != "api" {
		t.Fatalf("identity %+v", b.Identity)
	}
	// Deterministic: a second index of the same input has the same id.
	if _, b2 := researchIndex(t, t.TempDir()); b2.AnalysisID != b.AnalysisID {
		t.Fatalf("analysis id %s, then %s", b.AnalysisID, b2.AnalysisID)
	}
	lib, v1, v2 := instanceOf(b, "app:lib.jar"), instanceOf(b, "deps:fx-dep-1.0.jar"), instanceOf(b, "deps:fx-dep-2.0.jar")
	if lib == "" || v1 == "" || v2 == "" {
		t.Fatalf("components %+v", b.Components)
	}
	observed := map[string]bool{}
	f, err := os.Open("../../testdata/research/observed.txt")
	if err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if l := sc.Text(); l != "" && !strings.HasPrefix(l, "#") {
			observed[l] = true
		}
	}
	f.Close()
	const P = "Ljava/nio/file/Path;"
	type tc struct {
		id, inst, class, name, desc string
		resolution, search          string
	}
	cases := []tc{
		{"write", lib, "fx.research.Extractor", "writeEntry", "(" + P + "[B)V", "matched", "found"},
		{"unused", lib, "fx.research.Extractor", "unused", "(" + P + ")V", "matched", "not_found"},
		{"wipe", lib, "fx.research.internal.Cleaner", "wipe", "(" + P + ")V", "matched", "found"},
		{"resolve", v1, "fx.dep.Paths2", "resolve", "(" + P + "Ljava/lang/String;)" + P, "matched", "found"},
		{"legacy", v1, "fx.dep.Paths2", "legacyResolve", "(" + P + "Ljava/lang/String;)" + P, "matched", "not_found"},
		{"safe", v2, "fx.dep.Paths2", "safeResolve", "(" + P + "Ljava/lang/String;)" + P, "shadowed", "not_evaluated"},
		{"missing", lib, "fx.research.Extractor", "absent", "()V", "missing", "not_evaluated"},
		{"save-bytes", lib, "fx.research.Extractor", "save", "([B)V", "matched", "found"},
	}
	var targets []any
	for _, c := range cases {
		targets = append(targets, map[string]any{"id": c.id, "selector": map[string]any{"instance_id": c.inst, "class": c.class, "name": c.name, "descriptor": c.desc}})
	}
	code, doc := researchQuery(t, bundle, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "t1",
		"operation": "reachability", "targets": targets, "roots": map[string]any{"selection": "indexed"}})
	if code != 0 || doc["status"] != "ok" {
		t.Fatalf("exit %d: %v", code, doc)
	}
	results := doc["results"].([]any)
	nodes := map[string]map[string]any{}
	for _, e := range doc["evidence"].([]any) {
		r := e.(map[string]any)
		if r["record_type"] == "node" {
			nodes[r["node_id"].(string)] = r
		}
	}
	keyOf := func(id string) string {
		m := nodes[id]["method"].(map[string]any)
		return fmt.Sprintf("%s#%s%s", m["class"], m["name"], m["descriptor"])
	}
	for i, c := range cases {
		r := results[i].(map[string]any)
		search := r["search"].(map[string]any)
		if r["target_id"] != c.id || r["resolution"] != c.resolution || search["state"] != c.search {
			t.Errorf("%s: resolution %v search %v, want %s %s", c.id, r["resolution"], search["state"], c.resolution, c.search)
			continue
		}
		key := c.class + "#" + c.name + c.desc
		// What the JVM ran is found; found targets name a root that ran.
		if observed[key] && c.search != "found" {
			t.Errorf("%s: the JVM ran %s, the query says %s", c.id, key, c.search)
		}
		for _, e := range r["entries"].([]any) {
			p := e.(map[string]any)["path"].(map[string]any)
			ns := p["node_ids"].([]any)
			if last := keyOf(ns[len(ns)-1].(string)); last != key {
				t.Errorf("%s: path ends at %s", c.id, last)
			}
			if len(p["edge_ids"].([]any)) != len(ns)-1 {
				t.Errorf("%s: path shape %v", c.id, p)
			}
		}
		// The unbounded reflection site in plugin() is reachable: every
		// matched target's model is incomplete, unresolved ones undetermined.
		model := r["model"].(map[string]any)["state"]
		if (c.resolution == "matched") != (model == "incomplete") {
			t.Errorf("%s: model %v", c.id, model)
		}
	}
	// wipe: three API entries; Handler#process reaches it through a CHA step
	// to a library class, which is not client dispatch.
	wipe := results[2].(map[string]any)
	if wipe["total_entry_count"].(float64) != 3 {
		t.Errorf("wipe entries %v", wipe["entries"])
	}
	for _, e := range wipe["entries"].([]any) {
		if e.(map[string]any)["client_dispatch"] != false {
			t.Errorf("wipe entry marked client dispatch: %v", e)
		}
	}

	// Discovery: the Files operations the library calls, JDK bodies absent.
	code, doc = researchQuery(t, bundle, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "t2",
		"operation": "calls", "filters": map[string]any{"declared_class_prefix": "java.nio.file.Files"}})
	if code != 0 {
		t.Fatalf("calls: exit %d %v", code, doc)
	}
	var called []string
	for _, r := range doc["records"].([]any) {
		c := r.(map[string]any)
		if c["body_in_input"] != false {
			t.Errorf("JDK call with a body in the input: %v", c)
		}
		called = append(called, c["declared_name"].(string))
	}
	if strings.Join(called, ",") == "" || !strings.Contains(strings.Join(called, ","), "write") {
		t.Errorf("Files calls %v", called)
	}

	// Errors: an invalid request exits 2 with an error envelope.
	code, doc = researchQuery(t, bundle, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "t3",
		"operation": "entries", "filters": map[string]any{"reachable_only": true}})
	if code != 2 || doc["status"] != "error" || doc["error"].(map[string]any)["code"] != "UNSUPPORTED_FILTER" || doc["request_id"] != "t3" {
		t.Errorf("unsupported filter: exit %d %v", code, doc)
	}
	// A corrupted bundle exits 1 with BUNDLE_CORRUPT.
	nodesFile := filepath.Join(bundle, "nodes.jsonl")
	raw, _ := os.ReadFile(nodesFile)
	os.WriteFile(nodesFile, []byte(strings.Replace(string(raw), `"body":"bytecode"`, `"body":"native"`, 1)), 0o644)
	code, doc = researchQuery(t, bundle, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "t4", "operation": "entries"})
	if code != 1 || doc["error"].(map[string]any)["code"] != "BUNDLE_CORRUPT" {
		t.Errorf("corrupt bundle: exit %d %v", code, doc)
	}
}

func TestResearchIndexFlags(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"research", "index", "-app", "x.jar", "-out", filepath.Join(dir, "b")}, "-roots are required"},
		{[]string{"research", "index", "-app", "x.jar", "-roots", "api", "-out", dir}, "already exists"},
		{[]string{"research", "index", "-app", "x.jar", "-roots", "api", "-llm", "-out", filepath.Join(dir, "c")}, "flag provided but not defined: -llm"},
		{[]string{"research", "frobnicate"}, "unknown command"},
		{[]string{"research", "index", "-app", "x.jar", "-roots", "api", "-out", filepath.Join(dir, "no", "such", "b")}, "parent directory does not exist"},
	} {
		code, msg := runWithEnv(c.args, nil)
		if code != 2 || !strings.Contains(msg, c.want) {
			t.Errorf("%v: exit %d %q, want 2 and %q", c.args, code, msg, c.want)
		}
	}
}

// -scope reachable keeps every reachability answer of -scope all for the
// methods it keeps, drops what nothing reaches, and says so on a miss.
func TestResearchScopeReachable(t *testing.T) {
	_, all := researchIndex(t, t.TempDir())
	dir := t.TempDir()
	bundle, b := researchIndex(t, dir, "-scope", "reachable")
	if b.Identity.Configuration.Settings["scope"] != "reachable" || len(b.Nodes) >= len(all.Nodes) {
		t.Fatalf("scope %v: %d nodes, all has %d", b.Identity.Configuration.Settings["scope"], len(b.Nodes), len(all.Nodes))
	}
	// Every node and edge kept is identical to its -scope all record.
	allNodes := map[string]research.Node{}
	for _, n := range all.Nodes {
		allNodes[n.NodeID] = n
	}
	for _, n := range b.Nodes {
		if !jsonSame(n, allNodes[n.NodeID]) {
			t.Fatalf("node %s differs from -scope all", n.NodeID)
		}
	}
	lib, v1 := instanceOf(b, "app:lib.jar"), instanceOf(b, "deps:fx-dep-1.0.jar")
	const P = "Ljava/nio/file/Path;"
	sel := func(inst, class, name, desc string) map[string]any {
		return map[string]any{"instance_id": inst, "class": class, "name": name, "descriptor": desc}
	}
	_, doc := researchQuery(t, bundle, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "s1",
		"operation": "reachability", "roots": map[string]any{"selection": "indexed"}, "targets": []any{
			map[string]any{"id": "write", "selector": sel(lib, "fx.research.Extractor", "writeEntry", "("+P+"[B)V")},
			map[string]any{"id": "unused", "selector": sel(lib, "fx.research.Extractor", "unused", "("+P+")V")},
			map[string]any{"id": "legacy", "selector": sel(v1, "fx.dep.Paths2", "legacyResolve", "("+P+"Ljava/lang/String;)"+P)},
		}})
	rs := doc["results"].([]any)
	state := func(i int) (string, string) {
		r := rs[i].(map[string]any)
		return r["resolution"].(string), r["search"].(map[string]any)["state"].(string)
	}
	if res, st := state(0); res != "matched" || st != "found" {
		t.Errorf("writeEntry %s %s", res, st)
	}
	// The library's own dead method stays (the input's methods are kept).
	if res, st := state(1); res != "matched" || st != "not_found" {
		t.Errorf("unused %s %s", res, st)
	}
	// A dependency method nothing reaches is outside the scope.
	if res, _ := state(2); res != "missing" || !strings.Contains(rs[2].(map[string]any)["search"].(map[string]any)["reason"].(string), "-scope all") {
		t.Errorf("legacyResolve %v", rs[2])
	}
	if code, _ := runWithEnv([]string{"research", "index", "-app", "x.jar", "-roots", "api", "-scope", "some", "-out", filepath.Join(dir, "z")}, nil); code != 2 {
		t.Errorf("-scope some: exit %d", code)
	}
}

func jsonSame(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}
