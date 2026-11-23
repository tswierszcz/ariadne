package reach

import (
	"testing"

	"ariadne/internal/cves"
	"ariadne/internal/findings"
	"ariadne/internal/vulnmap"
)

// TestWithComponentsAltVersion (LIBRARY_MODE LB3 lb3-2, mode "approx"): the
// engine copied over an inventory in which lib:overload has another version
// picks the mapping variant of that version — on the same code — and
// reports MAPPING_VERSION_UNCOVERED when no variant covers it; the original
// engine keeps its own version.
func TestWithComponentsAltVersion(t *testing.T) {
	e := overloadEngine(t, "1.0")
	var id string
	for _, in := range e.components.Instances {
		if in.GAV != nil && in.GAV.Artifact == "overload" {
			id = in.ID
		}
	}
	if id == "" {
		t.Fatal("no lib:overload instance")
	}
	entry := variants(t,
		[2]string{"[1.0,2.0)", "lib.overload.C#parse(Ljava/lang/String;)V"},
		[2]string{"[2.0,3.0)", "lib.overload.C#parse(Ljava/io/InputStream;)V"})
	analyze := func(eng *Engine, version string) InstanceFinding {
		t.Helper()
		cve := cves.CVE{ID: "ARIADNE-TEST", Affected: []cves.Affected{{Package: "lib:overload", Version: version}}}
		r := eng.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: entry})[0]
		for _, f := range r.Findings {
			if f.Instance == id {
				return f
			}
		}
		t.Fatalf("%s: no finding for %s", version, id)
		return InstanceFinding{}
	}
	if f := analyze(e, "1.0"); f.Status != StatusNotReachable {
		t.Errorf("1.0 (variant: the uncalled String overload): %s %v", f.Status, f.Reasons)
	}
	alt := e.WithComponents(e.components.WithInstanceVersion(id, "2.5"))
	if f := analyze(alt, "2.5"); f.Status != StatusReachable {
		t.Errorf("2.5 (variant: the called InputStream overload): %s %v", f.Status, f.Reasons)
	}
	unc := e.WithComponents(e.components.WithInstanceVersion(id, "4.0"))
	if f := analyze(unc, "4.0"); f.Status != StatusUnknown || len(f.Reasons) == 0 || f.Reasons[0] != findings.MappingVersionUncovered {
		t.Errorf("4.0 (no variant): %s %v, want UNKNOWN MAPPING_VERSION_UNCOVERED", f.Status, f.Reasons)
	}
	if f := analyze(e, "1.0"); f.Status != StatusNotReachable {
		t.Error("the original engine changed")
	}
}
