package reach

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/findings"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/vulnmap"
)

// buildInstanceEngine is buildEngine plus the CVE:F1a component inventory.
// When withCopy is set, a second copy of vulnerablelib.jar (identified as
// org.vulnerable:lib:0.9) is listed after the original, so every one of
// its classes loses to the 1.0 copy under the classpath order.
func buildInstanceEngine(t *testing.T, withCopy bool) *Engine {
	t.Helper()
	deps := []input.DepJar{
		{Path: fixtureDir + "frameworkthing.jar", GAV: &input.GAV{Group: "org.framework", Artifact: "thing", Version: "1.0"}, Source: input.SourceDepsList},
		{Path: fixtureDir + "vulnerablelib.jar", GAV: &input.GAV{Group: "org.vulnerable", Artifact: "lib", Version: "1.0"}, Source: input.SourceDepsList},
	}
	if withCopy {
		copyPath := filepath.Join(t.TempDir(), "vulnerablelib-0.9.jar")
		b, err := os.ReadFile(fixtureDir + "vulnerablelib.jar")
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile(copyPath, b, 0o644)
		deps = append(deps, input.DepJar{Path: copyPath, GAV: &input.GAV{Group: "org.vulnerable", Artifact: "lib", Version: "0.9"}, Source: input.SourceDepsList})
	}
	h, _, err := hierarchy.Build(fixtureDir+"appentry.jar", deps)
	if err != nil {
		t.Fatal(err)
	}
	g := callgraph.Build(h, nil)
	e := New(g, h, NewFrameworkMatcher([]string{"org.framework:*"}))
	inv, err := component.Build(component.Inputs{AppInput: fixtureDir + "appentry.jar", AppJar: fixtureDir + "appentry.jar", DepJars: deps}, h)
	if err != nil {
		t.Fatal(err)
	}
	e.SetComponents(inv)
	return e
}

func analyzeAffected(t *testing.T, e *Engine, entry vulnmap.Entry, affected ...string) Result {
	t.Helper()
	cve := cves.CVE{ID: "ARIADNE-TEST"}
	for _, v := range affected {
		cve.Affected = append(cve.Affected, cves.Affected{Package: "org.vulnerable:lib", Version: v})
	}
	return e.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: entry})[0]
}

func reasons(rs ...findings.ReasonCode) []findings.ReasonCode { return rs }

// TestContractCases executes the contract cases of the CVE:F0 control set
// (testdata/cve-control/cases.yaml, ARIADNE-TEST-0001..0009) that CVE:F1b
// implements, plus option B's MAPPING_PARTIAL rule.
func TestContractCases(t *testing.T) {
	diff := func(m ...string) vulnmap.Entry { return vulnmap.Entry{Methods: m, Source: vulnmap.SourceDiff} }
	e := buildInstanceEngine(t, false)

	// 0001: reachable direct call.
	if r := analyzeAffected(t, e, diff("reachfixture.VulnerableLib#vulnerable"), "1.0"); r.Status != StatusReachable || len(r.Reasons) != 0 || len(r.Findings) != 1 || r.Findings[0].Status != StatusReachable {
		t.Errorf("0001: %s %v %+v", r.Status, r.Reasons, r.Findings)
	}
	// 0002 + option B: exists, not reached, partial mapping.
	if r := analyzeAffected(t, e, diff("reachfixture.VulnerableLib#unreachable"), "1.0"); r.Status != StatusNotReachable || !reflect.DeepEqual(r.Reasons, reasons(findings.MappingPartial)) {
		t.Errorf("0002: %s %v", r.Status, r.Reasons)
	}
	complete := vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#unreachable"}, Source: vulnmap.SourceManual, Role: vulnmap.RoleVulnerable, Coverage: vulnmap.CoverageComplete}
	if r := analyzeAffected(t, e, complete, "1.0"); r.Status != StatusNotReachable || len(r.Reasons) != 0 {
		t.Errorf("complete mapping: %s %v, want NOT_REACHABLE without reasons", r.Status, r.Reasons)
	}
	// 0003: method not found.
	if r := analyzeAffected(t, e, diff("reachfixture.VulnerableLib#gone"), "1.0"); r.Status != StatusUnknown || !reflect.DeepEqual(r.Reasons, reasons(findings.MethodNotFound)) || r.Path != nil {
		t.Errorf("0003: %s %v", r.Status, r.Reasons)
	}
	// 0007: entry-point-only mapping, reachable.
	ep := vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#vulnerable"}, Source: vulnmap.SourceManual, Role: vulnmap.RoleEntrypoint}
	if r := analyzeAffected(t, e, ep, "1.0"); r.Status != StatusReachable || !reflect.DeepEqual(r.Reasons, reasons(findings.MappingEntrypointOnly)) {
		t.Errorf("0007: %s %v", r.Status, r.Reasons)
	}

	// 0008: only the shadowed 0.9 copy is affected — its path must not be
	// borrowed from the selected 1.0 copy.
	ec := buildInstanceEngine(t, true)
	r := analyzeAffected(t, ec, diff("reachfixture.VulnerableLib#vulnerable"), "0.9")
	if r.Status != StatusUnknown || !reflect.DeepEqual(r.Reasons, reasons(findings.ClasspathConflict)) || r.Path != nil {
		t.Errorf("0008: %s %v path=%v", r.Status, r.Reasons, r.Path)
	}
	// Two affected instances: separate findings; the positive one wins.
	r = analyzeAffected(t, ec, diff("reachfixture.VulnerableLib#vulnerable"), "1.0", "0.9")
	if r.Status != StatusReachable || len(r.Findings) != 2 {
		t.Fatalf("two instances: %s %+v", r.Status, r.Findings)
	}
	byGAV := map[string]InstanceFinding{}
	for _, f := range r.Findings {
		byGAV[f.GAV] = f
	}
	if byGAV["org.vulnerable:lib:1.0"].Status != StatusReachable || byGAV["org.vulnerable:lib:0.9"].Status != StatusUnknown ||
		!reflect.DeepEqual(byGAV["org.vulnerable:lib:0.9"].Reasons, reasons(findings.ClasspathConflict)) ||
		byGAV["org.vulnerable:lib:1.0"].Instance == byGAV["org.vulnerable:lib:0.9"].Instance {
		t.Errorf("two instances findings = %+v", r.Findings)
	}
}

// TestModelGapRule covers contract §4 in its report-only form (CVE:F1b,
// ARIADNE-TEST-0004): a reachable unresolved site that could dispatch into
// the affected instance keeps NOT_REACHABLE but adds MODEL_INCOMPLETE and
// the gap count, on the advisory and on the instance finding; an
// unreachable site, or one bounded to other components, adds nothing; a
// positive path is reported regardless.
func TestModelGapRule(t *testing.T) {
	diff := func(m ...string) vulnmap.Entry { return vulnmap.Entry{Methods: m, Source: vulnmap.SourceDiff} }
	reachable := GapSite{Kind: "reflection", CallerFQCN: "reachfixture.AppEntry", CallerName: "callsVulnerableDirectly", CallerDesc: "()V"}
	partial := reasons(findings.MappingPartial)
	gapped := reasons(findings.MappingPartial, findings.ModelIncomplete)
	cases := []struct {
		name    string
		gap     GapSite
		method  string
		status  Status
		reasons []findings.ReasonCode
		gaps    map[string]int
	}{
		{"unbounded reachable", reachable, "reachfixture.VulnerableLib#unreachable", StatusNotReachable, gapped, map[string]int{"reflection": 1}},
		{"bounded into affected", withOwners(reachable, "reachfixture.VulnerableLib"), "reachfixture.VulnerableLib#unreachable", StatusNotReachable, gapped, map[string]int{"reflection": 1}},
		{"bounded elsewhere", withOwners(reachable, "reachfixture.FrameworkThing"), "reachfixture.VulnerableLib#unreachable", StatusNotReachable, partial, nil},
		{"unreachable site", GapSite{Kind: "reflection", CallerFQCN: "reachfixture.VulnerableLib", CallerName: "unreachable", CallerDesc: "()V"}, "reachfixture.VulnerableLib#unreachable", StatusNotReachable, partial, nil},
		{"positive path wins", reachable, "reachfixture.VulnerableLib#vulnerable", StatusReachable, nil, nil},
	}
	for _, c := range cases {
		e := buildInstanceEngine(t, false)
		e.SetGaps([]GapSite{c.gap})
		r := analyzeAffected(t, e, diff(c.method), "1.0")
		if r.Status != c.status || !reflect.DeepEqual(r.Reasons, c.reasons) || !reflect.DeepEqual(r.ModelGaps, c.gaps) {
			t.Errorf("%s: %s %v %v, want %s %v %v", c.name, r.Status, r.Reasons, r.ModelGaps, c.status, c.reasons, c.gaps)
		}
		if len(r.Findings) != 1 || r.Findings[0].Status != c.status || !reflect.DeepEqual(r.Findings[0].ModelGaps, c.gaps) {
			t.Errorf("%s: instance finding %+v", c.name, r.Findings)
		}
	}
}

func withOwners(s GapSite, owners ...string) GapSite { s.Owners = owners; return s }

// TestAggregateNeverHidesUnresolvedInstance: contract §5 — without a
// positive finding, one unresolved instance makes the advisory UNKNOWN,
// even when another instance's finding is NOT_REACHABLE.
func TestAggregateNeverHidesUnresolvedInstance(t *testing.T) {
	e := buildInstanceEngine(t, true)
	r := analyzeAffected(t, e, vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#unreachable"}, Source: vulnmap.SourceDiff}, "1.0", "0.9")
	if r.Status != StatusUnknown || !reflect.DeepEqual(r.Reasons, reasons(findings.ClasspathConflict)) {
		t.Fatalf("aggregate = %s %v (findings %+v), want UNKNOWN [CLASSPATH_CONFLICT]", r.Status, r.Reasons, r.Findings)
	}
}
