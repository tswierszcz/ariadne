package resultsv2

import (
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/findings"
	"ariadne/internal/input"
	"ariadne/internal/reach"
	"ariadne/internal/vulnmap"
)

func testInputs() Inputs {
	gav := func(v string) *input.GAV { return &input.GAV{Group: "g", Artifact: "lib", Version: v} }
	inv := &component.Inventory{
		Artifacts: []component.Artifact{
			{ID: "a-app", Location: "app:app.jar", Role: component.RoleApp, SHA256: "11"},
			{ID: "a-1", Location: "deps:lib-1.jar", Role: component.RoleDependency, SHA256: "22"},
			{ID: "a-2", Location: "deps:lib-2.jar", Role: component.RoleDependency, SHA256: "33"},
			{ID: "a-x", Location: "deps:x.jar", Role: component.RoleDependency, SHA256: "44"},
		},
		Instances: []component.Instance{
			{ID: "c-app", ArtifactID: "a-app", Detection: component.DetectedApp},
			{ID: "c-1", ArtifactID: "a-1", GAV: gav("1"), Detection: component.DetectedPOMProperties},
			{ID: "c-2", ArtifactID: "a-2", GAV: gav("2"), Detection: component.DetectedPOMProperties, Shadowed: 1},
			{ID: "c-x", ArtifactID: "a-x", Detection: component.DetectedUnidentified},
		},
		Conflicts: []component.Conflict{{FQCN: "g.A", KeptInstance: "c-1", ShadowedInstance: "c-2"}},
	}
	results := []reach.Result{
		{CVE: "CVE-B", Status: reach.StatusReachable, Reasons: []findings.ReasonCode{findings.MappingEntrypointOnly}, Findings: []reach.InstanceFinding{
			{Instance: "c-1", Status: reach.StatusReachable, Path: []string{"app#m()V", "g.A#vuln()V"}},
			{Instance: "c-2", Status: reach.StatusUnknown, Reasons: []findings.ReasonCode{findings.ClasspathConflict}},
		}},
		{CVE: "CVE-A", Status: reach.StatusUnknown, Reasons: []findings.ReasonCode{findings.MethodNotFound}, Findings: []reach.InstanceFinding{
			{Instance: "c-1", Status: reach.StatusUnknown, Reasons: []findings.ReasonCode{findings.MethodNotFound}},
		}},
		{CVE: "CVE-C", Status: reach.StatusNotReachable, Reasons: []findings.ReasonCode{findings.MappingPartial, findings.ModelIncomplete}},
	}
	return Inputs{
		ToolVersion: "test", Offline: true, Inventory: inv, Results: results,
		CVEs: []cves.CVE{
			{ID: "CVE-B", Affected: []cves.Affected{{Package: "g:lib", Version: "1"}, {Package: "g:lib", Version: "2"}}},
			{ID: "CVE-A", Affected: []cves.Affected{{Package: "g:lib", Version: "1"}}},
			{ID: "CVE-C", Affected: []cves.Affected{{Package: "g:other", Version: "9"}}},
		},
		Mappings: map[string]vulnmap.Entry{
			"CVE-B": {Methods: []string{"g.A#vuln"}, Source: vulnmap.SourceManual, Role: vulnmap.RoleEntrypoint},
			"CVE-A": {Methods: []string{"g.A#gone"}, Source: vulnmap.SourceDiff},
		},
		Gaps: GlobalGaps{UnresolvedSites: 3, Reachable: 2, Unbounded: 1},
	}
}

func TestBuild(t *testing.T) {
	doc := Build(testInputs())
	if doc.SchemaVersion != "2" || doc.Snapshot.Kind != "inputs" || !strings.HasPrefix(doc.Snapshot.ID, "sha256:") || !doc.Snapshot.Offline {
		t.Errorf("header = %+v %+v", doc.SchemaVersion, doc.Snapshot)
	}
	idents := map[string]string{}
	for _, c := range doc.Components {
		idents[c.ID] = c.Identity
	}
	// c-2's shadowed copy has different content (SameContent false).
	if want := map[string]string{"c-app": "identified", "c-1": "identified", "c-2": "conflict", "c-x": "unidentified"}; !reflect.DeepEqual(idents, want) {
		t.Errorf("identities = %v", idents)
	}
	if got := []string{doc.Advisories[0].ID, doc.Advisories[1].ID, doc.Advisories[2].ID}; !reflect.DeepEqual(got, []string{"CVE-A", "CVE-B", "CVE-C"}) {
		t.Errorf("advisories not sorted: %v", got)
	}
	if m := doc.Advisories[0].Mapping; m == nil || m.Source != "diff" || m.Role != "vulnerable" || m.Coverage != "partial" || m.Review != "unreviewed" || m.MatchPrecision != "name" {
		t.Errorf("diff mapping defaults = %+v", m)
	}
	if doc.Advisories[2].Mapping != nil {
		t.Errorf("unmapped advisory has a mapping")
	}
	if len(doc.Findings) != 3 || doc.Findings[0].AdvisoryID != "CVE-A" || doc.Findings[2].Identity != "conflict" || doc.Findings[1].Applicability != "affected" {
		t.Errorf("findings = %+v", doc.Findings)
	}
	for _, f := range doc.Findings {
		for _, c := range f.ReasonCodes {
			if c == findings.MappingPartial || c == findings.MappingEntrypointOnly {
				t.Errorf("mapping-axis code %s on a finding", c)
			}
		}
	}
	// CVE-C has no instance: coverage, without the mapping-axis code.
	if w := doc.Coverage.AdvisoriesWithoutInstance; len(w) != 1 || w[0].CVE != "CVE-C" || !reflect.DeepEqual(w[0].Reasons, []findings.ReasonCode{findings.ModelIncomplete}) {
		t.Errorf("without instance = %+v", w)
	}
	c := doc.Coverage.Counts
	if c.Advisories != 3 || c.Findings != 3 || c.FindingsByReachability["UNKNOWN"] != 2 || c.AdvisoriesByV1Status["NOT_REACHABLE"] != 1 {
		t.Errorf("counts = %+v", c)
	}
	if !reflect.DeepEqual(doc.Coverage.UnidentifiedComponents, []string{"c-x"}) {
		t.Errorf("unidentified = %v", doc.Coverage.UnidentifiedComponents)
	}
}

// The snapshot reference depends on artifact content and stable location
// only, not on where the files were read from.
func TestSnapshotIDIgnoresPaths(t *testing.T) {
	a, b := testInputs(), testInputs()
	for i := range b.Inventory.Artifacts {
		b.Inventory.Artifacts[i].Path = "/tmp/elsewhere/" + b.Inventory.Artifacts[i].Location
	}
	if Build(a).Snapshot.ID != Build(b).Snapshot.ID {
		t.Error("snapshot ID depends on read paths")
	}
	b.Inventory.Artifacts[1].SHA256 = "99"
	if Build(a).Snapshot.ID == Build(b).Snapshot.ID {
		t.Error("snapshot ID ignores content")
	}
}

func TestExportV1(t *testing.T) {
	in := testInputs()
	doc := Build(in)
	got := ExportV1(doc)
	want := []V1Row{
		{CVE: "CVE-A", Status: "UNKNOWN", Reasons: []findings.ReasonCode{findings.MethodNotFound}},
		{CVE: "CVE-B", Status: "REACHABLE"},
		{CVE: "CVE-C", Status: "NOT_REACHABLE", Reasons: []findings.ReasonCode{findings.ModelIncomplete}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExportV1 = %+v", got)
	}
	if m := CheckV1(doc, in.Results); m != nil {
		t.Errorf("consistent inputs reported: %v", m)
	}
	in.Results[0].Status = reach.StatusFramework
	if m := CheckV1(doc, in.Results); len(m) != 1 || !strings.Contains(m[0], "CVE-B") {
		t.Errorf("mismatch not reported: %v", m)
	}
}
