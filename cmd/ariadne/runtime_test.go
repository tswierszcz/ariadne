package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/runtimeevents"
)

type runtimeDoc struct {
	CallGraph struct {
		Edges int `json:"edges"`
	} `json:"callgraph"`
	Findings []json.RawMessage     `json:"findings"`
	Runtime  *runtimeevents.Result `json:"runtime"`
}

func runtimeRun(t *testing.T, jar string, extra ...string) (runtimeDoc, string) {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "deps"), 0o755)
	args := append([]string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", jar, "-deps", filepath.Join(dir, "deps"),
		"-profile", "generic", "-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-json-format", "v2"}, extra...)
	if code, out := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	var doc runtimeDoc
	b, _ := os.ReadFile(filepath.Join(dir, "r.json"))
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	md, _ := os.ReadFile(filepath.Join(dir, "r.md"))
	return doc, string(md)
}

// TestRuntimeTrace (RUNTIME_NATIVE N1): the JFR-recorded fixture trace of
// testdata/runtime (scripts/runtime-trace-fixture.sh) imported next to the
// analysis: every class matched by hash, the two definitions of rtm.Plugin
// kept apart, the reflective call the static graph cannot see reported as
// not in the graph, and the graph unchanged. Against another build of the
// jar the changed class is a different version, its evidence unused.
func TestRuntimeTrace(t *testing.T) {
	const trace = "../../testdata/runtime/run.trace"
	plain, _ := runtimeRun(t, "../../testdata/runtime/rtm.jar")
	doc, md := runtimeRun(t, "../../testdata/runtime/rtm.jar", "-runtime-trace", trace)
	rt := doc.Runtime
	if rt == nil {
		t.Fatal("no runtime section in v2")
	}
	if doc.CallGraph.Edges != plain.CallGraph.Edges || len(doc.Findings) != len(plain.Findings) {
		t.Errorf("the trace changed the analysis: edges %d vs %d, findings %d vs %d", doc.CallGraph.Edges, plain.CallGraph.Edges, len(doc.Findings), len(plain.Findings))
	}
	if rt.ClassCounts[runtimeevents.StatusMatched] != 7 || len(rt.ClassCounts) != 1 {
		t.Errorf("classes %v, want 7 matched", rt.ClassCounts)
	}
	if strings.Join(rt.MultiDefined, ",") != "rtm.Plugin" {
		t.Errorf("several definitions %v, want rtm.Plugin", rt.MultiDefined)
	}
	if rt.ExecutedCounts["traced"] != 1 || rt.ExecutedCounts["sampled"] == 0 || rt.ExecutedOther != 0 {
		t.Errorf("executed %v (+%d not usable)", rt.ExecutedCounts, rt.ExecutedOther)
	}
	if rt.CallCounts[runtimeevents.CallNotInGraph] != 2 || rt.CallCounts[runtimeevents.CallInGraph] == 0 {
		t.Errorf("calls %v: want the 2 reflective calls (one per loader) not in the graph", rt.CallCounts)
	}
	for _, want := range []string{"## Runtime evidence", "**incomplete**", "`rtm.Main#reflective(Ljava/lang/ClassLoader;J)J` (line", "`rtm.Plugin#work(J)J`", "The absence of an event\nproves nothing"} {
		if !strings.Contains(md, want) {
			t.Errorf("report lacks %q", want)
		}
	}

	v2, _ := runtimeRun(t, "../../testdata/runtime/rtm-v2.jar", "-runtime-trace", trace)
	if v2.Runtime.ClassCounts[runtimeevents.StatusDifferentVersion] != 1 || v2.Runtime.ExecutedOther == 0 {
		t.Errorf("another build: classes %v, %d executed not usable; want Square a different version and its samples unused",
			v2.Runtime.ClassCounts, v2.Runtime.ExecutedOther)
	}
}

// TestRuntimeTraceSnapshot: the trace is an input of the snapshot and the
// importer's rules an algorithm version; a bad trace fails before analysis.
func TestRuntimeTraceSnapshot(t *testing.T) {
	dir := t.TempDir()
	snap := filepath.Join(dir, "snap")
	runtimeRun(t, "../../testdata/runtime/rtm.jar", "-runtime-trace", "../../testdata/runtime/run.trace", "-snapshot-out", snap)
	b, err := os.ReadFile(filepath.Join(snap, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := string(b)
	for _, want := range []string{`"role": "runtime-trace"`, `"runtime.import": "` + runtimeevents.ImportVersion + `"`, `"runtime-trace": "run.trace"`} {
		if !strings.Contains(m, want) {
			t.Errorf("manifest lacks %s", want)
		}
	}
	bad := filepath.Join(dir, "bad.trace")
	os.WriteFile(bad, []byte(`{"record":"header","format":"other"}`+"\n"), 0o644)
	code, out := runWithEnv([]string{"-offline", "-app", "../../testdata/runtime/rtm.jar", "-profile", "generic", "-out", filepath.Join(dir, "x.md"),
		"-json", filepath.Join(dir, "x.json"), "-runtime-trace", bad}, nil)
	if code == 0 || !strings.Contains(out, "-runtime-trace") {
		t.Errorf("a bad trace: exit %d, %q", code, out)
	}
}
