package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/reach"
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
	if rt.ClassCounts[runtimeevents.StatusMatched] != 12 || len(rt.ClassCounts) != 1 {
		t.Errorf("classes %v, want 12 matched", rt.ClassCounts)
	}
	if strings.Join(rt.MultiDefined, ",") != "rtm.Plugin" {
		t.Errorf("several definitions %v, want rtm.Plugin", rt.MultiDefined)
	}
	if rt.ExecutedCounts["traced"] != 1 || rt.ExecutedCounts["sampled"] == 0 || rt.ExecutedOther != 0 {
		t.Errorf("executed %v (+%d not usable)", rt.ExecutedCounts, rt.ExecutedOther)
	}
	if rt.CallCounts[runtimeevents.CallNotInGraph] != 2 || rt.CallCounts[runtimeevents.CallInGraph] == 0 {
		t.Errorf("calls %v: want 2 not in the graph", rt.CallCounts)
	}
	// The body Greeter inherits from BaseGreeter is a CHA edge now
	// (TestCHAInheritedBody), and the lambda created in Gaps#sam and run
	// through Work#run in Gaps#drive a SAM-call edge (DYNAMIC_LINKAGE L1,
	// internal/samcall). What is left: the reflective call under each loader.
	causes := map[string]int{}
	for _, g := range rt.GapCauses {
		causes[g.Cause] = g.Calls
	}
	if causes[runtimeevents.GapReflection] != 2 || len(causes) != 1 {
		t.Errorf("gap causes %v, want reflection 2", causes)
	}
	if strings.Contains(md, "| sam-call |") {
		t.Error("the report still lists a sam-call gap")
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

// TestRuntimeFilter: -runtime-filter-out selects the mapped methods of the
// chosen verdicts whose class is in the input, as JFR "Class::method"
// filters (constructors included, XML-escaped), and coverage marks the
// advisories traced but not entered.
func TestRuntimeFilter(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/runtime/rtm.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	results := []reach.Result{
		{CVE: "CVE-A", Status: reach.StatusFramework, Methods: []string{"rtm.Lib#parse", "rtm.Square#<init>()V"}},
		{CVE: "CVE-B", Status: reach.StatusNotReachable, Methods: []string{"rtm.Circle#area(J)J", "absent.Lib#x"}},
		{CVE: "CVE-C", Status: reach.StatusNotApplicable, Methods: []string{"rtm.Plugin#work"}},
	}
	f := runtimeFilter(results, h, defaultRuntimeFilterStatus)
	if strings.Join(f.Filters, ";") != "rtm.Circle::area;rtm.Lib::parse;rtm.Square::<init>" || f.Advisories != 2 || f.Skipped != 1 {
		t.Errorf("filter %+v", f)
	}
	jfc := string(f.jfcFile(defaultRuntimeFilterStatus))
	for _, want := range []string{`<event name="jdk.MethodTrace">`, `<setting name="enabled">true</setting>`, `rtm.Square::&lt;init&gt;`, `of 3 mapped vulnerable methods of 2 advisories`} {
		if !strings.Contains(jfc, want) {
			t.Errorf("jfc lacks %q:\n%s", want, jfc)
		}
	}
	if f := runtimeFilter(results, h, "REACHABLE"); len(f.Filters) != 0 || !strings.Contains(string(f.jfcFile("REACHABLE")), `<setting name="enabled">false</setting>`) {
		t.Errorf("no REACHABLE finding: %+v", f)
	}

	rt := runtimeevents.Result{MethodFilters: []string{"rtm.Lib::parse", "rtm.Circle::area"}}
	traced := runtimeTraced(rt, results, h, nil)
	obs := map[string]map[string][]runtimeevents.MethodObservation{"CVE-A": {"": {{Evidence: "e1", Method: "rtm.Lib#parse(Ljava/lang/String;)J", Basis: "traced", Count: 3}}}}
	var noEntry []string
	for _, r := range tracedNoEntry(traced, obs, results) {
		noEntry = append(noEntry, r.Advisory)
	}
	if !traced["CVE-A"][""] || !traced["CVE-B"][""] || traced["CVE-C"] != nil || strings.Join(noEntry, ",") != "CVE-B" {
		t.Errorf("traced %v, not entered %v; want A and B traced, only B not entered", traced, noEntry)
	}
}

// TestRuntimeFilterFlag: the CLI writes the settings file (here empty: the
// fixture has no advisories) and warns; an unknown verdict is rejected.
func TestRuntimeFilterFlag(t *testing.T) {
	dir := t.TempDir()
	jfc := filepath.Join(dir, "v.jfc")
	_, md := runtimeRun(t, "../../testdata/runtime/rtm.jar", "-runtime-filter-out", jfc)
	b, err := os.ReadFile(jfc)
	if err != nil || !strings.Contains(string(b), `<setting name="enabled">false</setting>`) || !strings.Contains(md, "-runtime-filter-out: no mapped vulnerable method") {
		t.Errorf("settings file %q (err %v); report warning present: %v", b, err, strings.Contains(md, "-runtime-filter-out"))
	}
	code, out := runWithEnv([]string{"-app", "x", "-out", "r.md", "-json", "r.json", "-runtime-filter-status", "MAYBE"}, nil)
	if code == 0 || !strings.Contains(out, "-runtime-filter-status") {
		t.Errorf("unknown verdict: exit %d, %q", code, out)
	}
}

// TestGapClassifier: causes decided from the caller's bytecode at the
// measured index, on the real classes of testdata/runtime.
func TestGapClassifier(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/runtime/rtm.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	pc := func(owner, name, desc, callOwner, callName string) int {
		ci, _ := h.Lookup(owner)
		m, _ := ci.MethodByNameAndDesc(name, desc)
		for _, c := range m.Calls {
			if c.Owner == callOwner && c.Name == callName {
				return c.PC
			}
		}
		t.Fatalf("no call %s#%s in %s#%s", callOwner, callName, owner, name)
		return -1
	}
	shapes := pc("rtm.Main", "shapes", "(J)J", "rtm.Shape", "area")
	cases := []struct {
		mode, caller, callee string
		bci                  int
		basis, cause, detail string
	}{
		{"cha", "rtm.Main#shapes(J)J", "rtm.Square#area(J)J", shapes, "stack-adjacent", runtimeevents.GapDispatchMissing, "rtm.Shape"},
		{"rta", "rtm.Main#shapes(J)J", "rtm.Square#area(J)J", shapes, "stack-adjacent", runtimeevents.GapRTAPruned, "rtm.Shape"},
		{"cha", "rtm.Gaps#inherited(J)J", "rtm.BaseGreeter#greet(J)J", pc("rtm.Gaps", "inherited", "(J)J", "rtm.Greeter", "greet"), "stack-adjacent", runtimeevents.GapInherited, "rtm.Greeter"},
		{"cha", "rtm.Gaps#drive(Lrtm/Work;J)J", "rtm.Gaps#lambda$sam$0(J)J", pc("rtm.Gaps", "drive", "(Lrtm/Work;J)J", "rtm.Work", "run"), "stack-through-runtime", runtimeevents.GapSAMCall, "rtm.Work"},
		{"cha", "rtm.Main#reflective(Ljava/lang/ClassLoader;J)J", "rtm.Plugin#work(J)J", pc("rtm.Main", "reflective", "(Ljava/lang/ClassLoader;J)J", "java.lang.reflect.Method", "invoke"), "stack-through-runtime", runtimeevents.GapReflection, "java.lang.reflect.Method#invoke"},
		{"cha", "rtm.Main#main([Ljava/lang/String;)V", "rtm.LoadedOnly#<clinit>()V", 0, "stack-adjacent", runtimeevents.GapClassInit, "rtm.LoadedOnly"},
		{"cha", "rtm.Main#shapes(J)J", "rtm.Square#area(J)J", -1, "stack-adjacent", runtimeevents.GapNoCallAtSite, "no bytecode index"},
	}
	for _, c := range cases {
		cause, detail := gapClassifier(h, c.mode)(c.caller, c.callee, c.bci, c.basis)
		if cause != c.cause || detail != c.detail {
			t.Errorf("%s %s -> %s: %s (%s), want %s (%s)", c.mode, c.caller, c.callee, cause, detail, c.cause, c.detail)
		}
	}
}
