package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/component"
	"ariadne/internal/input"
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

// RI4 source locations: the package path of the file the class names,
// the method's line range, and with -sources (a directory or a jar) the
// hash of that file; the binding stays declared.
func TestResearchSourceLocations(t *testing.T) {
	src := "../../testdata/research/lib"
	jar := filepath.Join(t.TempDir(), "lib-sources.jar")
	writeZipOf(t, src, jar)
	for _, sources := range []string{src, jar} {
		_, b := researchIndex(t, t.TempDir(), "-sources", sources)
		want, _ := os.ReadFile(filepath.Join(src, "fx/research/Extractor.java"))
		wantSHA := fmt.Sprintf("sha256:%x", sha256.Sum256(want))
		found := false
		for _, n := range b.Nodes {
			m := n.Method
			if m == nil {
				continue
			}
			switch {
			case m.Class == "fx.research.Extractor" && m.Name == "writeEntry":
				found = true
				s := n.Source
				if s.Path == nil || *s.Path != "fx/research/Extractor.java" || s.Binding != "declared" || s.SourceSHA256 == nil || *s.SourceSHA256 != wantSHA {
					t.Errorf("%s: writeEntry source %+v", sources, s)
				}
				lines := strings.Split(string(want), "\n")
				if s.LineStart == nil || *s.LineStart < 1 || *s.LineEnd > len(lines) || !strings.Contains(strings.Join(lines[*s.LineStart-1:*s.LineEnd], "\n"), "Files.write") {
					t.Errorf("%s: writeEntry lines %v-%v do not cover its body", sources, s.LineStart, s.LineEnd)
				}
			case m.Class == "fx.dep.Paths2":
				// A dependency class: its path is declared, its file is not in these sources.
				if n.Source.Path == nil || *n.Source.Path != "fx/dep/Paths2.java" || n.Source.SourceSHA256 != nil {
					t.Errorf("%s: Paths2 source %+v", sources, n.Source)
				}
			case m.InstanceID == nil && n.Source.Path != nil:
				t.Errorf("external node %s#%s has a source path", m.Class, m.Name)
			}
		}
		if !found {
			t.Fatal("writeEntry not in the bundle")
		}
		ok := false
		for _, in := range b.Identity.Inputs {
			ok = ok || (in.ID == "in-sources" && in.Role == "sources")
		}
		if !ok {
			t.Errorf("%s: no sources input in the identity", sources)
		}
	}
	// Without -sources: paths, no hashes.
	_, b := researchIndex(t, t.TempDir())
	for _, n := range b.Nodes {
		if n.Source.SourceSHA256 != nil {
			t.Fatalf("hash without -sources: %+v", n.Source)
		}
	}
}

func writeZipOf(t *testing.T, dir, out string) {
	t.Helper()
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		w, _ := zw.Create(filepath.ToSlash(rel))
		b, _ := os.ReadFile(p)
		w.Write(b)
		return nil
	})
	zw.Close()
	f.Close()
}

// RI1: embedded library candidates that were not identified are named in
// coverage.limits; confirmed ones and modules of the host are not.
func TestResearchEmbeddedLimit(t *testing.T) {
	gav := func(a, v string) input.GAV { return input.GAV{Group: "g", Artifact: a, Version: v} }
	inv := &component.Inventory{Embedded: []component.Embedded{
		{Kind: component.EmbeddedLibrary, GAV: gav("ok", "1"), Status: component.EmbeddedConfirmed},
		{Kind: component.EmbeddedLibrary, GAV: gav("cand", "1"), Status: component.EmbeddedVersionCandidates},
		{Kind: component.EmbeddedModule, GAV: gav("mod", "1")},
		{Kind: component.EmbeddedLibrary, GAV: gav("b", "2"), Status: component.EmbeddedReferenceUnavailable},
		{Kind: component.EmbeddedLibrary, GAV: gav("a", "1"), Status: component.EmbeddedReferenceUnavailable},
		{Kind: component.EmbeddedLibrary, GAV: gav("c", "3"), Status: component.EmbeddedPartial},
	}}
	got := embeddedLimit(inv)
	want := "Embedded components: 3 library candidates not identified (partial: g:c:3; reference_unavailable: g:a:1, g:b:2); whatever classes of theirs the host carries are attributed to the host component. reference_unavailable: the reference jar is in neither -cache-dir nor ARIADNE_MAVEN_LOCAL (research index does not use the network)."
	if got != want {
		t.Fatalf("limit:\n got %s\nwant %s", got, want)
	}
	if l := embeddedLimit(&component.Inventory{Embedded: inv.Embedded[:3]}); l != "" {
		t.Fatalf("all identified, got limit %q", l)
	}
}

// RH-A0: `research check` accepts this producer's response (with replay)
// and rejects a broken one; `research query -verify full` verifies the
// datasets an operation does not read, which the default only checks for
// presence and size.
func TestResearchCheckAndVerify(t *testing.T) {
	dir := t.TempDir()
	bundle, b := researchIndex(t, dir)
	req := filepath.Join(dir, "req.json")
	raw, _ := json.Marshal(map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "c1", "operation": "entries"})
	os.WriteFile(req, raw, 0o644)
	resp := filepath.Join(dir, "resp.json")
	if code, msg := runWithEnv([]string{"research", "query", "-bundle", bundle, "-request", req, "-out", resp}, nil); code != 0 {
		t.Fatalf("query: %d %s", code, msg)
	}
	var stdout bytes.Buffer
	check := func(response string) (int, map[string]any) {
		stdout.Reset()
		var stderr bytes.Buffer
		code := run([]string{"research", "check", "-bundle", bundle, "-request", req, "-response", response, "-replay"}, &stdout, &stderr, func(string) string { return "" })
		var v map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &v); err != nil {
			t.Fatalf("check output %q (%s): %v", stdout.String(), stderr.String(), err)
		}
		return code, v
	}
	if code, v := check(resp); code != 0 || v["valid"] != true || v["replay"] != "identical" {
		t.Fatalf("own response: exit %d verdict %v", code, v)
	}
	good, _ := os.ReadFile(resp)
	bad := filepath.Join(dir, "bad.json")
	os.WriteFile(bad, bytes.Replace(good, []byte(`"request_id": "c1"`), []byte(`"request_id": "c2"`), 1), 0o644)
	code, v := check(bad)
	if code != 1 || v["valid"] != false || v["replay"] != "differs" || !strings.Contains(stdout.String(), `"rule": "binding"`) {
		t.Fatalf("broken response: exit %d verdict %s", code, stdout.String())
	}
	// calls.jsonl is not read by entries: a damaged copy passes the default
	// verification and fails -verify full.
	calls := filepath.Join(bundle, "calls.jsonl")
	cb, _ := os.ReadFile(calls)
	cb[10] ^= 1
	os.WriteFile(calls, cb, 0o644)
	if code, msg := runWithEnv([]string{"research", "query", "-bundle", bundle, "-request", req, "-out", resp}, nil); code != 0 {
		t.Fatalf("default verification: %d %s", code, msg)
	}
	if code, _ := runWithEnv([]string{"research", "query", "-bundle", bundle, "-request", req, "-out", resp, "-verify", "full"}, nil); code != 1 {
		t.Fatalf("-verify full on a damaged dataset: exit %d, want 1", code)
	}
	if b, _ := os.ReadFile(resp); !strings.Contains(string(b), `"BUNDLE_CORRUPT"`) && !strings.Contains(string(b), `"status": "error"`) {
		t.Fatalf("-verify full response: %s", b)
	}
}

// RH-A2 cases (j) and (k) of testdata/research/EXPECTED.MD: entry dispatch
// flags recomputed over every path (client code standing in at a library
// interface; a caller-supplied field; a parameter of the entry itself,
// lb2-7, stated as a warning), and a callback site's condition kept on its
// own edge. The JVM ran none of the targets (observed.txt).
func TestResearchDispatchAndConditions(t *testing.T) {
	dir := t.TempDir()
	bundle, b := researchIndex(t, dir)
	lib, v1 := instanceOf(b, "app:lib.jar"), instanceOf(b, "deps:fx-dep-1.0.jar")
	const P = "Ljava/nio/file/Path;"
	sel := func(inst, class, name, desc string) map[string]any {
		return map[string]any{"instance_id": inst, "class": class, "name": name, "descriptor": desc}
	}
	_, doc := researchQuery(t, bundle, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "d1",
		"operation": "reachability", "roots": map[string]any{"selection": "indexed"}, "targets": []any{
			map[string]any{"id": "purge", "selector": sel(lib, "fx.research.internal.Cleaner", "purge", "("+P+")V")},
			map[string]any{"id": "purger", "selector": sel(v1, "fx.dep.Purger", "call", "("+P+")V")},
			map[string]any{"id": "recover", "selector": sel(lib, "fx.research.Workers", "recover", "(Ljava/lang/Throwable;)"+P)},
		}})
	if w, _ := doc["warnings"].([]any); len(w) != 1 || w[0] != research.SuppliedParameterWarning {
		t.Errorf("warnings %v, want the lb2-7 warning", doc["warnings"])
	}
	names := map[string]string{}
	for _, n := range b.Nodes {
		if n.Method != nil {
			names[n.NodeID] = n.Method.Class + "#" + n.Method.Name
		}
	}
	type flags struct{ client, supplied bool }
	got := map[string]map[string]flags{}
	for _, r := range doc["results"].([]any) {
		rm := r.(map[string]any)
		if st := rm["search"].(map[string]any)["state"]; st != "found" {
			t.Errorf("%s: %v", rm["target_id"], st)
		}
		m := map[string]flags{}
		for _, e := range rm["entries"].([]any) {
			em := e.(map[string]any)
			m[names[em["root_node_id"].(string)]] = flags{em["client_dispatch"].(bool), em["supplied_dispatch"].(bool)}
		}
		got[rm["target_id"].(string)] = m
	}
	want := map[string]map[string]flags{
		"purge": {"fx.research.Drains#drain": {true, false}, "fx.research.Pipeline#run": {true, false}, "fx.research.internal.Cleaner#purge": {false, false}},
		// Drains#hand: lb2-7, caller-supplied on paths that start there, not reported.
		"purger":  {"fx.research.Drains#hand": {false, false}, "fx.research.Relay#send": {false, true}},
		"recover": {"fx.research.Workers#later": {false, false}},
	}
	for target, entries := range want {
		for root, f := range entries {
			if g, ok := got[target][root]; !ok || g != f {
				t.Errorf("%s from %s: %+v (present %v), want %+v", target, root, g, ok, f)
			}
		}
	}
	// The JVM ran none of the three targets.
	obs, _ := os.ReadFile("../../testdata/research/observed.txt")
	for _, m := range []string{"Cleaner#purge", "Purger#call", "Workers#recover"} {
		if strings.Contains(string(obs), m) {
			t.Errorf("observed.txt has %s", m)
		}
	}
	// (k): later's two callbacks are separate edges; only recover's site
	// carries the exceptional condition.
	var later string
	for id, n := range names {
		if n == "fx.research.Workers#later" {
			later = id
		}
	}
	_, doc = researchQuery(t, bundle, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "d2",
		"operation": "callees", "node_id": later})
	conds := map[string][]any{}
	for _, r := range doc["records"].([]any) {
		e := r.(map[string]any)
		if e["basis"] != "callback" {
			continue
		}
		for _, s := range e["sites"].([]any) {
			conds[names[e["to"].(string)]] = append(conds[names[e["to"].(string)]], s.(map[string]any)["condition"])
		}
	}
	if fmt.Sprint(conds["fx.research.Workers#recover"]) != "[exceptional]" || fmt.Sprint(conds["fx.research.Workers#archive"]) != "[<nil>]" {
		t.Errorf("callback site conditions %v", conds)
	}
}

// RH-A2: a root-subset query over an RTA graph says the graph was pruned
// for every indexed root; over CHA it says nothing.
func TestResearchSubsetPrunedWarning(t *testing.T) {
	for _, mode := range []string{"cha", "rta"} {
		dir := t.TempDir()
		bundle, b := researchIndex(t, dir, "-callgraph-mode", mode)
		var root string
		for _, n := range b.Nodes {
			if n.Method != nil && n.Method.Class == "fx.research.Extractor" && n.Method.Name == "extract" {
				root = n.NodeID
			}
		}
		lib := instanceOf(b, "app:lib.jar")
		_, doc := researchQuery(t, bundle, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": "p1",
			"operation": "reachability", "roots": map[string]any{"selection": "subset", "node_ids": []any{root}}, "targets": []any{
				map[string]any{"id": "w", "selector": map[string]any{"instance_id": lib, "class": "fx.research.Extractor", "name": "writeEntry", "descriptor": "(Ljava/nio/file/Path;[B)V"}}}})
		pruned := false
		for _, w := range doc["warnings"].([]any) {
			pruned = pruned || strings.Contains(w.(string), "was pruned (rta)")
		}
		if pruned != (mode == "rta") {
			t.Errorf("%s: warnings %v", mode, doc["warnings"])
		}
	}
}

// RI6: JFR filters by method name, so a target with overloads names the
// other overloads it traces too; an unresolved target gets no filter; a
// request for another analysis is refused.
func TestResearchTraceConfigOverloads(t *testing.T) {
	dir := t.TempDir()
	bundle, b := researchIndex(t, dir)
	lib := instanceOf(b, "app:lib.jar")
	write := func(analysis string) string {
		raw, _ := json.Marshal(map[string]any{"schema_version": research.TraceRequestSchema, "analysis_id": analysis, "request_id": "o1",
			"targets": []any{
				map[string]any{"id": "save", "selector": map[string]any{"instance_id": lib, "class": "fx.research.Extractor", "name": "save", "descriptor": "([B)V"}},
				map[string]any{"id": "absent", "selector": map[string]any{"instance_id": lib, "class": "fx.research.Extractor", "name": "absent", "descriptor": "()V"}}}})
		p := filepath.Join(t.TempDir(), "req.json")
		os.WriteFile(p, raw, 0o644)
		return p
	}
	cfg, jfc := filepath.Join(dir, "cfg.json"), filepath.Join(dir, "t.jfc")
	if code, msg := runWithEnv([]string{"research", "trace-config", "-bundle", bundle, "-request", write(b.AnalysisID), "-out", cfg, "-jfc", jfc}, nil); code != 0 {
		t.Fatalf("trace-config: %d %s", code, msg)
	}
	var c research.TraceConfig
	raw, _ := os.ReadFile(cfg)
	json.Unmarshal(raw, &c)
	if len(c.Filters) != 1 || c.Filters[0] != "fx.research.Extractor::save" {
		t.Errorf("filters %v", c.Filters)
	}
	other := ""
	for _, n := range b.Nodes {
		if n.Method != nil && n.Method.Class == "fx.research.Extractor" && n.Method.Name == "save" && n.Method.Descriptor != "([B)V" && n.CanonicalNodeID == n.NodeID {
			other = n.NodeID
		}
	}
	if s := c.Targets[0]; len(s.AlsoTraced) != 1 || s.AlsoTraced[0] != other {
		t.Errorf("save overloads traced too: %v, want [%s]", s.AlsoTraced, other)
	}
	if a := c.Targets[1]; a.Resolution != "missing" || a.Filter != nil {
		t.Errorf("absent target %+v", a)
	}
	if code, _ := runWithEnv([]string{"research", "trace-config", "-bundle", bundle, "-request", write("sha256:" + strings.Repeat("0", 64)), "-out", cfg, "-jfc", jfc}, nil); code != 2 {
		t.Errorf("another analysis: exit %d, want 2", code)
	}
}
