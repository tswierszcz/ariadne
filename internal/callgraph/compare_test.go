package callgraph

import (
	"bytes"
	"strings"
	"testing"
)

func exportFixture() Export {
	return Export{
		Config: Config{Mode: ModeCHA, Builder: BuilderVersion, SkipRule: "jdk", UnknownAllocationPolicy: PolicyAllSubtypes, Roots: "all"},
		Edges: map[EdgeID][]int{
			{From: "a.A#m()V", To: "a.I#run()V", Kind: "interface", Basis: "direct"}:       {3},
			{From: "a.A#m()V", To: "a.Impl1#run()V", Kind: "interface", Basis: "cha"}:      {3},
			{From: "a.A#m()V", To: "a.Impl2#run()V", Kind: "interface", Basis: "cha"}:      {3},
			{From: "a.Impl2#run()V", To: "lib.V#bad()V", Kind: "virtual", Basis: "direct"}: {7, 9},
		},
		Findings: map[string]FindingPath{
			"CVE-1": {ID: "CVE-1", Status: "REACHABLE", Path: []string{"a.A#m()V", "a.Impl2#run()V", "lib.V#bad()V"}},
			"CVE-2": {ID: "CVE-2", Status: "NOT_REACHABLE"},
		},
	}
}

// TestExportRoundTrip: Write is deterministic and Read gives back the same
// graph; a file of another format is refused.
func TestExportRoundTrip(t *testing.T) {
	ex := exportFixture()
	var a, b bytes.Buffer
	if err := Write(&a, ex); err != nil {
		t.Fatal(err)
	}
	if err := Write(&b, ex); err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Fatal("two writes differ")
	}
	back, err := Read(strings.NewReader(a.String()))
	if err != nil {
		t.Fatal(err)
	}
	if c := Compare(ex, back); !c.Empty() {
		t.Errorf("round trip differs: %+v", c)
	}
	if _, err := Read(strings.NewReader("# other-format\n")); err == nil {
		t.Error("another format was accepted")
	}
}

// TestCompareExplains: removing a CHA edge is reported by basis, and the
// finding whose path used it is listed with that edge as the reason;
// different settings are always disclosed.
func TestCompareExplains(t *testing.T) {
	before := exportFixture()
	after := exportFixture()
	after.Config.Mode = "rta"
	delete(after.Edges, EdgeID{From: "a.A#m()V", To: "a.Impl2#run()V", Kind: "interface", Basis: "cha"})
	after.Edges[EdgeID{From: "a.Impl2#run()V", To: "lib.V#bad()V", Kind: "virtual", Basis: "direct"}] = []int{7}
	after.Findings["CVE-1"] = FindingPath{ID: "CVE-1", Status: "NOT_REACHABLE"}
	c := Compare(before, after)
	if len(c.ConfigDifferences) != 1 || !strings.Contains(c.ConfigDifferences[0], "mode") {
		t.Errorf("config differences %v", c.ConfigDifferences)
	}
	if len(c.Removed) != 1 || c.RemovedByBasis["cha"] != 1 || len(c.Added) != 0 || c.EvidenceChanged != 1 {
		t.Errorf("edges: removed %v by basis %v added %v evidence %d", c.Removed, c.RemovedByBasis, c.Added, c.EvidenceChanged)
	}
	if len(c.Findings) != 1 || c.Findings[0].ID != "CVE-1" || c.Findings[0].Old != "REACHABLE" || c.Findings[0].New != "NOT_REACHABLE" ||
		len(c.Findings[0].Because) != 1 || c.Findings[0].Because[0] != "removed a.A#m()V -> a.Impl2#run()V" {
		t.Errorf("findings %+v", c.Findings)
	}
	if c.Empty() {
		t.Error("comparison reported empty")
	}
}
