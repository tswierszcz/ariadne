package reach

import (
	"reflect"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/findings"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/mavenversion"
	"ariadne/internal/vulnmap"
)

const overloadDir = "../../testdata/overloadfixture/"

// overloadEngine analyzes testdata/overloadfixture (scripts/build-overload-fixture.sh)
// with lib.jar identified as lib:overload:<version>.
func overloadEngine(t *testing.T, version string) *Engine {
	t.Helper()
	deps := []input.DepJar{{Path: overloadDir + "lib.jar", GAV: &input.GAV{Group: "lib", Artifact: "overload", Version: version}, Source: input.SourceDepsList}}
	h, _, err := hierarchy.Build(overloadDir+"app.jar", deps)
	if err != nil {
		t.Fatal(err)
	}
	e := New(callgraph.Build(h, nil), h, NewFrameworkMatcher(nil))
	inv, err := component.Build(component.Inputs{AppInput: overloadDir + "app.jar", AppJar: overloadDir + "app.jar", DepJars: deps}, h)
	if err != nil {
		t.Fatal(err)
	}
	e.SetComponents(inv)
	return e
}

func analyzeOverload(t *testing.T, e *Engine, version string, entry vulnmap.Entry) Result {
	t.Helper()
	cve := cves.CVE{ID: "ARIADNE-TEST", Affected: []cves.Affected{{Package: "lib:overload", Version: version}}}
	return e.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: entry})[0]
}

func diffEntry(m ...string) vulnmap.Entry {
	return vulnmap.Entry{Methods: m, Source: vulnmap.SourceDiff}
}

// ARIADNE-TEST-0104 and the inheritance case: a descriptor selects one
// overload; a shared name alone no longer links them; an inherited call
// keeps reaching the declaring class's body.
func TestExactSelectors(t *testing.T) {
	e := overloadEngine(t, "1.0")
	cases := []struct {
		spec string
		want Status
	}{
		{"lib.overload.C#parse", StatusReachable},                          // name: every overload (legacy behavior)
		{"lib.overload.C#parse(Ljava/lang/String;)V", StatusNotReachable},  // exact: the vulnerable overload is not called
		{"lib.overload.C#parse(Ljava/lang/String;)", StatusNotReachable},   // params: same
		{"lib.overload.C#parse(Ljava/io/InputStream;)V", StatusReachable},  // the called overload
		{"lib.overload.Base#handle(Ljava/lang/String;)V", StatusReachable}, // inherited through Sub
		{"lib.overload.Base#handle(Ljava/lang/Integer;)V", StatusNotReachable},
		{"lib.overload.C#parse(Ljava/lang/String;)I", StatusUnknown}, // no such method: METHOD_NOT_FOUND
	}
	for _, c := range cases {
		if r := analyzeOverload(t, e, "1.0", diffEntry(c.spec)); r.Status != c.want {
			t.Errorf("%s: %s %v, want %s (path %v)", c.spec, r.Status, r.Reasons, c.want, r.Path)
		}
	}
}

func variants(t *testing.T, vs ...[2]string) vulnmap.Entry {
	t.Helper()
	e := vulnmap.Entry{Source: vulnmap.SourceManual}
	for _, v := range vs {
		r, err := mavenversion.ParseRange(v[0])
		if err != nil {
			t.Fatal(err)
		}
		e.Variants = append(e.Variants, vulnmap.Variant{Versions: v[0], Range: r, Methods: []string{v[1]}})
		e.Methods = append(e.Methods, v[1])
	}
	return e.WithDefaults()
}

// ARIADNE-TEST-0101 (and the version-mapping matrix): two version ranges
// select disjoint overloads; an uncovered, conflicting or irregular
// version stays undetermined with a reason code.
func TestVersionedMappings(t *testing.T) {
	str, opts := "lib.overload.B#parse(Ljava/lang/String;)V", "lib.overload.B#parse(Llib/overload/Options;Ljava/lang/String;)V"
	two := [][2]string{{"[1.0,2.0)", str}, {"[2.0,3.0)", opts}}
	cases := []struct {
		version string
		entry   vulnmap.Entry
		status  Status
		reasons []findings.ReasonCode
	}{
		{"1.5", variants(t, two...), StatusReachable, nil},
		{"2.1", variants(t, two...), StatusNotReachable, reasons(findings.MappingPartial)},
		{"3.1", variants(t, two...), StatusUnknown, reasons(findings.MappingVersionUncovered)},
		{"2.0", variants(t, [2]string{"[1.0,2.0]", str}, [2]string{"[2.0,3.0)", opts}), StatusUnknown, reasons(findings.MappingConflict)},
		{"2.0-SNAPSHOT", variants(t, two...), StatusUnknown, reasons(findings.VersionUnknown)},
	}
	for _, c := range cases {
		r := analyzeOverload(t, overloadEngine(t, c.version), c.version, c.entry)
		if r.Status != c.status || !reflect.DeepEqual(r.Reasons, c.reasons) {
			t.Errorf("%s: %s %v, want %s %v (findings %+v)", c.version, r.Status, r.Reasons, c.status, c.reasons, r.Findings)
		}
		if s, _ := Aggregate(r.Findings); s != r.Status {
			t.Errorf("%s: aggregate %s != row %s", c.version, s, r.Status)
		}
	}
}
