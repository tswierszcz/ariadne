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

// TestTypesInExport (GRAPH_PRECISION P1): created types round-trip,
// compare by class, and a v1 file (no types) still reads and compares
// without inventing type differences.
func TestTypesInExport(t *testing.T) {
	before := exportFixture()
	before.Types = []TypeRecord{{Class: "a.Impl1", Origins: []string{"new=2"}}, {Class: "a.Impl2", Origins: []string{"new=1"}}}
	after := exportFixture()
	after.Types = []TypeRecord{{Class: "a.Impl1", Origins: []string{"new=2"}}, {Class: "a.Impl3", Origins: []string{"service=1"}}}
	var buf bytes.Buffer
	if err := Write(&buf, after); err != nil {
		t.Fatal(err)
	}
	back, err := Read(strings.NewReader(buf.String()))
	if err != nil || len(back.Types) != 2 || back.Types[1].Origins[0] != "service=1" {
		t.Fatalf("round trip: %+v %v", back.Types, err)
	}
	c := Compare(before, back)
	if len(c.TypesAdded) != 1 || c.TypesAdded[0] != "a.Impl3" || len(c.TypesRemoved) != 1 || c.TypesRemoved[0] != "a.Impl2" {
		t.Errorf("types added %v removed %v", c.TypesAdded, c.TypesRemoved)
	}
	v1 := strings.Replace(buf.String(), "# "+ExportFormat, "# "+ExportFormatV1, 1)
	v1 = strings.Join(filterLines(strings.Split(v1, "\n"), "type\t"), "\n")
	old, err := Read(strings.NewReader(v1))
	if err != nil {
		t.Fatalf("v1: %v", err)
	}
	if c := Compare(old, back); c.TypesAdded != nil || c.TypesRemoved != nil {
		t.Errorf("v1 vs v2 invented type differences: %v %v", c.TypesAdded, c.TypesRemoved)
	}
}

func filterLines(ls []string, prefix string) []string {
	var out []string
	for _, l := range ls {
		if !strings.HasPrefix(l, prefix) {
			out = append(out, l)
		}
	}
	return out
}
