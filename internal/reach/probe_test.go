package reach

import (
	"reflect"
	"testing"

	"ariadne/internal/findings"
	"ariadne/internal/vulnmap"
)

// TestGapProbe (ROADMAP B0 b0-4): the probe counts, per NOT_REACHABLE
// finding, unbounded and instance-bounded site gaps, lambda gaps whose
// creators are in the instance and skipped-receiver calls whose skipped
// expansion lands in it — and changes nothing else.
func TestGapProbe(t *testing.T) {
	diff := vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#unreachable"}, Source: vulnmap.SourceDiff}
	caller := GapSite{CallerFQCN: "reachfixture.AppEntry", CallerName: "callsVulnerableDirectly", CallerDesc: "()V"}
	site := func(kind string, owners ...string) GapSite {
		s := caller
		s.Kind, s.Owners = kind, owners
		return s
	}
	unreachableCaller := GapSite{Kind: "lambda", CallerFQCN: "reachfixture.VulnerableLib", CallerName: "unreachable", CallerDesc: "()V", Owners: []string{"reachfixture.VulnerableLib"}}

	e := buildInstanceEngine(t, false)
	e.SetGaps([]GapSite{site("reflection"), site("reflection", "reachfixture.VulnerableLib"), site("callback", "reachfixture.FrameworkThing")})
	before := analyzeAffected(t, e, diff, "1.0")

	e.SetProbe([]GapSite{site("lambda", "reachfixture.VulnerableLib"), site("lambda", "reachfixture.FrameworkThing"), site("lambda"), unreachableCaller}, nil)
	inst := before.Findings[0].Instance
	// A skipped JDK target whose expansion lands in the instance, one in
	// another component only, and one on java.lang.Object.
	e.probe.skipped = append(e.probe.skipped,
		skippedTarget{calls: 3, instances: map[string]bool{inst: true}},
		skippedTarget{calls: 5, instances: map[string]bool{"elsewhere": true}},
		skippedTarget{calls: 2, object: true, instances: map[string]bool{inst: true}})

	results := []Result{before}
	results[0].Findings = append([]InstanceFinding(nil), before.Findings...)
	e.ApplyProbe(results)
	got := results[0].Findings[0].GapProbe
	want := &GapProbe{Unbounded: 1, Bounded: 1, BoundedByKind: map[string]int{"reflection": 1}, Lambda: 1, SkippedReceiver: 5, SkippedReceiverObject: 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("probe = %+v, want %+v", got, want)
	}
	// Measurement only: the verdict, reasons and model gaps are untouched.
	f := results[0].Findings[0]
	if f.Status != StatusNotReachable || !reflect.DeepEqual(f.Reasons, before.Findings[0].Reasons) || !reflect.DeepEqual(f.ModelGaps, before.Findings[0].ModelGaps) ||
		!reflect.DeepEqual(results[0].Reasons, reasons(findings.MappingPartial, findings.ModelIncomplete)) {
		t.Errorf("finding changed: %+v / row %v", f, results[0].Reasons)
	}

	s := SummarizeProbe(results)
	wantMarked := map[string]int{ProbeCurrent: 1, ProbeScoped: 1, ProbeScopedLambda: 1, ProbeScopedNoObject: 1, ProbeScopedSkipped: 1}
	if s == nil || s.NotReachable != 1 || !reflect.DeepEqual(s.Marked, wantMarked) || s.Version != GapProbeVersion {
		t.Errorf("summary = %+v", s)
	}
	if rows := ProbeRows(results); len(rows) != 1 || rows[0].GAV != "org.vulnerable:lib:1.0" {
		t.Errorf("rows = %+v", rows)
	}
}

// TestGapProbeVariants: each variant marks exactly what it counts.
func TestGapProbeVariants(t *testing.T) {
	cases := []struct {
		p    GapProbe
		want []string
	}{
		{GapProbe{}, nil},
		{GapProbe{Unbounded: 4}, []string{ProbeCurrent}},
		{GapProbe{Bounded: 1}, []string{ProbeCurrent, ProbeScoped, ProbeScopedLambda, ProbeScopedNoObject, ProbeScopedSkipped}},
		{GapProbe{Lambda: 1}, []string{ProbeScopedLambda, ProbeScopedNoObject, ProbeScopedSkipped}},
		{GapProbe{SkippedReceiver: 2, SkippedReceiverObject: 2}, []string{ProbeScopedSkipped}},
		{GapProbe{SkippedReceiver: 3, SkippedReceiverObject: 2}, []string{ProbeScopedNoObject, ProbeScopedSkipped}},
	}
	for _, c := range cases {
		if got := c.p.Marks(); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%+v marks %v, want %v", c.p, got, c.want)
		}
	}
}

// TestGapProbeSkippedFromGraph: skipped-receiver calls are read from the
// graph — a direct virtual edge from a reachable method to a type the skip
// rule names counts once, with no instance when CHA's expansion of it
// selects only the skipped type's own body (VulnerableLibSub inherits
// riskyInstance).
func TestGapProbeSkippedFromGraph(t *testing.T) {
	e := buildInstanceEngine(t, false)
	e.SetProbe(nil, func(fqcn string) bool { return fqcn == "reachfixture.VulnerableLib" })
	_, targets, calls := e.ProbeInputs()
	if targets != 1 || calls != 1 {
		t.Fatalf("skipped targets %d calls %d, want 1 1", targets, calls)
	}
	if st := e.probe.skipped[0]; st.object || len(st.instances) != 0 {
		t.Errorf("target = %+v, want no instance", st)
	}
	// Without the skip rule naming it, nothing is skipped.
	e.SetProbe(nil, func(string) bool { return false })
	if _, targets, _ := e.ProbeInputs(); targets != 0 {
		t.Errorf("targets = %d, want 0", targets)
	}
}
