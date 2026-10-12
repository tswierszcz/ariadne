package reach

import (
	"reflect"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/cves"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/vulnmap"
)

const fixtureDir = "../../testdata/reachfixture/"

// buildEngine wires up testdata/reachfixture as three separate inputs —
// an app jar and two dependency jars, one tagged as a framework package
// and one as a plain dependency — exactly so REACHABLE, FRAMEWORK and
// NOT_REACHABLE each have a real, distinct call-graph shape to be tested
// against instead of being asserted only against a hand-built graph.
func buildEngine(t *testing.T) (*Engine, *callgraph.Graph, *hierarchy.Hierarchy) {
	t.Helper()
	depJars := []input.DepJar{
		{Path: fixtureDir + "frameworkthing.jar", GAV: &input.GAV{Group: "org.framework", Artifact: "thing", Version: "1.0"}},
		{Path: fixtureDir + "vulnerablelib.jar", GAV: &input.GAV{Group: "org.vulnerable", Artifact: "lib", Version: "1.0"}},
	}
	h, warnings, err := hierarchy.Build(fixtureDir+"appentry.jar", depJars)
	if err != nil {
		t.Fatalf("hierarchy.Build: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	g := callgraph.Build(h, nil)
	e := New(g, h, NewFrameworkMatcher([]string{"org.framework:*"}))
	return e, g, h
}

func analyzeOne(t *testing.T, e *Engine, methods []string) Result {
	t.Helper()
	cve := cves.CVE{ID: "CVE-TEST-0001"}
	vm := map[string]vulnmap.Entry{cve.ID: {Methods: methods, Source: vulnmap.SourceManual}}
	results := e.Analyze([]cves.CVE{cve}, vm)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	return results[0]
}

func TestAnalyze_reachableDirectDependencyCall(t *testing.T) {
	e, _, _ := buildEngine(t)
	r := analyzeOne(t, e, []string{"reachfixture.VulnerableLib#vulnerable"})
	if r.Status != StatusReachable {
		t.Fatalf("Status = %q, want REACHABLE (path: %v)", r.Status, r.Path)
	}
	want := []string{
		"reachfixture.AppEntry#callsVulnerableDirectly()V",
		"reachfixture.VulnerableLib#vulnerable()V",
	}
	if !reflect.DeepEqual(r.Path, want) {
		t.Errorf("Path = %v, want %v", r.Path, want)
	}
	if r.Entry == nil || r.Entry.From != want[0] || r.Entry.To != want[1] {
		t.Errorf("Entry = %+v, want From=%q To=%q", r.Entry, want[0], want[1])
	}
}

func TestAnalyze_reachableDirectCallIntoFrameworkPackage(t *testing.T) {
	e, _, _ := buildEngine(t)
	// riskyMethod is itself the "vulnerable" method, and it lives in a
	// framework-tagged package — but the app calls it in one hop, so
	// this is the section 3.6/5 exception: still REACHABLE, not
	// FRAMEWORK, because there's no need to trace *through* any
	// framework-internal call to get there.
	r := analyzeOne(t, e, []string{"reachfixture.FrameworkThing#riskyMethod"})
	if r.Status != StatusReachable {
		t.Fatalf("Status = %q, want REACHABLE (path: %v)", r.Status, r.Path)
	}
}

func TestAnalyze_frameworkMediatedCall(t *testing.T) {
	e, _, _ := buildEngine(t)
	// frameworkOnlyVulnerable is called only from FrameworkThing#doWork
	// (itself only reachable from the app), never directly by app code:
	// the only path to it runs *through* a framework-tagged method, so
	// it must be FRAMEWORK, not REACHABLE, even though a path does
	// exist.
	r := analyzeOne(t, e, []string{"reachfixture.VulnerableLib#frameworkOnlyVulnerable"})
	if r.Status != StatusFramework {
		t.Fatalf("Status = %q, want FRAMEWORK (path: %v)", r.Status, r.Path)
	}
	want := []string{
		"reachfixture.AppEntry#callsFrameworkWhichCallsVulnerable()V",
		"reachfixture.FrameworkThing#doWork()V",
		"reachfixture.VulnerableLib#frameworkOnlyVulnerable()V",
	}
	if !reflect.DeepEqual(r.Path, want) {
		t.Errorf("Path = %v, want %v", r.Path, want)
	}
	if r.Entry == nil || r.Entry.From != want[0] || r.Entry.To != want[1] {
		t.Errorf("Entry = %+v, want From=%q To=%q", r.Entry, want[0], want[1])
	}
}

func TestAnalyze_directCallIntoFrameworkPackageIsReachableNotFramework(t *testing.T) {
	// doWork is called directly by the app (one hop, app -> framework),
	// so even though it lives in a framework-tagged package, it's
	// REACHABLE, not FRAMEWORK: there's no framework-internal call to
	// trace through to get to it.
	e, _, _ := buildEngine(t)
	r := analyzeOne(t, e, []string{"reachfixture.FrameworkThing#doWork"})
	if r.Status != StatusReachable {
		t.Fatalf("Status = %q, want REACHABLE for a direct one-hop call into a framework package", r.Status)
	}
}

func TestAnalyze_inheritedMethodMatchedThroughSubclassCall(t *testing.T) {
	// callsInheritedMethodThroughSubclass calls riskyInstance() on a
	// VulnerableLibSub instance; VulnerableLibSub doesn't override it,
	// so the bytecode's Methodref owner is the subclass (confirmed by
	// hand with javap) even though VulnerableLib is the "vulnerable
	// class" named in vulnmap. matchVulnerableMethod must still find
	// this call.
	e, _, _ := buildEngine(t)
	r := analyzeOne(t, e, []string{"reachfixture.VulnerableLib#riskyInstance"})
	if r.Status != StatusReachable {
		t.Fatalf("Status = %q, want REACHABLE (path: %v)", r.Status, r.Path)
	}
	want := []string{
		"reachfixture.AppEntry#callsInheritedMethodThroughSubclass()V",
		"reachfixture.VulnerableLibSub#riskyInstance()V",
	}
	if !reflect.DeepEqual(r.Path, want) {
		t.Errorf("Path = %v, want %v", r.Path, want)
	}
}

func TestAnalyze_notReachable(t *testing.T) {
	e, _, _ := buildEngine(t)
	r := analyzeOne(t, e, []string{"reachfixture.VulnerableLib#unreachable"})
	if r.Status != StatusNotReachable {
		t.Fatalf("Status = %q, want NOT_REACHABLE", r.Status)
	}
	if r.Path != nil {
		t.Errorf("Path = %v, want nil", r.Path)
	}
}

func TestAnalyze_unknownWithNoVulnmapEntry(t *testing.T) {
	e, _, _ := buildEngine(t)
	results := e.Analyze([]cves.CVE{{ID: "CVE-NO-MAPPING"}}, map[string]vulnmap.Entry{})
	if len(results) != 1 || results[0].Status != StatusUnknown {
		t.Fatalf("got %+v, want a single UNKNOWN result", results)
	}
}

func TestAnalyze_unmatchedMethodIsNotReachable(t *testing.T) {
	e, _, _ := buildEngine(t)
	// A vulnmap entry that names a method nowhere in this universe
	// (neither declared nor called) must not panic and must be
	// NOT_REACHABLE, not UNKNOWN (ariadne *does* have a mapping for the
	// CVE; it just can't find the method in this app's graph).
	r := analyzeOne(t, e, []string{"com.example.NoSuchClass#noSuchMethod"})
	if r.Status != StatusNotReachable {
		t.Fatalf("Status = %q, want NOT_REACHABLE", r.Status)
	}
}

func TestFrameworkMatcher(t *testing.T) {
	m := NewFrameworkMatcher([]string{"org.springframework.boot:*", "org.hibernate:*", "com.example:exact-artifact"})
	cases := []struct {
		gav  *input.GAV
		want bool
	}{
		{&input.GAV{Group: "org.springframework.boot", Artifact: "spring-boot-starter-web"}, true},
		{&input.GAV{Group: "org.hibernate", Artifact: "hibernate-core"}, true},
		{&input.GAV{Group: "com.example", Artifact: "exact-artifact"}, true},
		{&input.GAV{Group: "com.example", Artifact: "other-artifact"}, false},
		{&input.GAV{Group: "org.yaml", Artifact: "snakeyaml"}, false},
		{nil, false},
	}
	for _, tc := range cases {
		if got := m.Match(tc.gav); got != tc.want {
			t.Errorf("Match(%+v) = %v, want %v", tc.gav, got, tc.want)
		}
	}
}

func analyzeEntry(t *testing.T, e *Engine, entry vulnmap.Entry) Result {
	t.Helper()
	cve := cves.CVE{ID: "CVE-TEST-0002"}
	return e.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: entry})[0]
}

// An LLM that saw only the advisory must never produce NOT_REACHABLE, but
// a positive finding, or any verdict backed by a fix diff, is unchanged.
func TestAnalyze_llmAdvisoryOnlyNotReachableIsDowngradedToUnknown(t *testing.T) {
	e, _, _ := buildEngine(t)
	notReached := []string{"reachfixture.VulnerableLib#unreachable"}

	if r := analyzeEntry(t, e, vulnmap.Entry{Methods: notReached, Source: vulnmap.SourceManual}); r.Status != StatusNotReachable {
		t.Fatalf("control: manual entry Status = %q, want NOT_REACHABLE (is the fixture method still unreachable?)", r.Status)
	}

	r := analyzeEntry(t, e, vulnmap.Entry{Methods: notReached, Source: vulnmap.SourceLLM, Model: "m/x", Evidence: vulnmap.EvidenceAdvisory})
	if r.Status != StatusUnknown {
		t.Errorf("advisory-only LLM NOT_REACHABLE: Status = %q, want UNKNOWN", r.Status)
	}
	if r.Note == "" || r.Model != "m/x" || r.Evidence != vulnmap.EvidenceAdvisory || r.Source != vulnmap.SourceLLM {
		t.Errorf("result lost its provenance: %+v", r)
	}

	r = analyzeEntry(t, e, vulnmap.Entry{Methods: notReached, Source: vulnmap.SourceLLM, Model: "m/x", Evidence: vulnmap.EvidenceDiff})
	if r.Status != StatusNotReachable || r.Note != "" {
		t.Errorf("diff-backed LLM NOT_REACHABLE: Status = %q Note = %q, want NOT_REACHABLE with no note", r.Status, r.Note)
	}

	r = analyzeEntry(t, e, vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#vulnerable"}, Source: vulnmap.SourceLLM, Model: "m/x", Evidence: vulnmap.EvidenceAdvisory})
	if r.Status != StatusReachable || r.Note != "" {
		t.Errorf("advisory-only LLM REACHABLE: Status = %q Note = %q, want REACHABLE unchanged", r.Status, r.Note)
	}
}
