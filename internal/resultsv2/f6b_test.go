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
)

// TestCandidateApplicability (CVE:F6b): a version-candidates instance is
// "affected" only when every candidate version is; otherwise which
// version is present decides, and that is not known.
func TestCandidateApplicability(t *testing.T) {
	c := cves.CVE{Affected: []cves.Affected{{Package: "g:a", Version: "1.0"}, {Package: "g:a", Version: "1.1"}, {Package: "other:x", Version: "2.0"}}}
	for _, tc := range []struct {
		versions []string
		want     string
	}{
		{[]string{"1.0", "1.1"}, "affected"},
		{[]string{"1.0", "2.0"}, "undetermined"},
		{[]string{"2.0"}, "undetermined"},
	} {
		if got := candidateApplicability(c, "g:a", tc.versions); got != tc.want {
			t.Errorf("%v = %s, want %s", tc.versions, got, tc.want)
		}
	}
}

// TestV1ExportWithCandidates (CVE:F9.4): an advisory with a confirmed
// instance and a version-candidates instance exports to the same v1 row
// as the engine's aggregate; the candidate finding keeps its reason and
// is undetermined.
func TestV1ExportWithCandidates(t *testing.T) {
	in := testInputs()
	g := input.GAV{Group: "g", Artifact: "lib"}
	in.Inventory.Instances = append(in.Inventory.Instances, component.Instance{ID: "c-emb", ArtifactID: "a-x", GAV: &g,
		Detection: component.DetectedEmbedded, HostInstance: "c-x", VersionCandidates: []string{"1", "3"}})
	in.Results[1].Findings = append(in.Results[1].Findings, reach.InstanceFinding{Instance: "c-emb", Status: reach.StatusReachable,
		Reasons: []findings.ReasonCode{findings.VersionCandidates}, Path: []string{"app#m()V", "g.A#gone()V"}})
	in.Results[1].Status, in.Results[1].Reasons = reach.Aggregate(in.Results[1].Findings)
	doc := Build(in)
	if m := CheckV1(doc, in.Results); m != nil {
		t.Fatalf("v1 export disagrees: %v", m)
	}
	for _, f := range doc.Findings {
		if f.ComponentInstanceID == "c-emb" && (f.Applicability != "undetermined" || f.Identity != "version_candidates" || f.ApplicabilityCheck != nil) {
			t.Errorf("candidate finding %+v", f)
		}
	}
	for _, c := range doc.Components {
		if c.ID == "c-emb" && (c.GAV != "g:lib:{1,3}" || !reflect.DeepEqual(c.VersionCandidates, []string{"1", "3"})) {
			t.Errorf("component %+v", c)
		}
	}
}

// TestAnalysisID (CVE:F9.9): the ID is a function of the content — the
// same document gives the same ID (an ID already set does not count), a
// different verdict a different one.
func TestAnalysisID(t *testing.T) {
	a, err := WithAnalysisID(Build(testInputs()))
	if err != nil || !strings.HasPrefix(a.AnalysisID, "sha256:") {
		t.Fatalf("%q %v", a.AnalysisID, err)
	}
	again, _ := WithAnalysisID(a)
	if again.AnalysisID != a.AnalysisID {
		t.Errorf("not stable: %s vs %s", again.AnalysisID, a.AnalysisID)
	}
	in := testInputs()
	in.Results[0].Findings[0].Status = reach.StatusFramework
	in.Results[0].Status = reach.StatusFramework
	b, _ := WithAnalysisID(Build(in))
	if b.AnalysisID == a.AnalysisID {
		t.Errorf("different results, same ID")
	}
}
