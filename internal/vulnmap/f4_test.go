package vulnmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSelector(t *testing.T) {
	ok := map[string]string{
		"a.B#m":                        PrecisionName,
		"a.B$Inner#<init>":             PrecisionName,
		"a.B#m(Ljava/lang/String;)":    PrecisionParams,
		"a.B#m(Ljava/lang/String;[I)V": PrecisionExact,
		"a.B#<init>(La/Opts;)V":        PrecisionExact,
		"a.B#m()[[Ljava/lang/Object;":  PrecisionExact,
	}
	for spec, want := range ok {
		s, err := ParseSelector(spec)
		if err != nil || s.Precision() != want || s.String() != spec {
			t.Errorf("%s: %+v %v, want precision %s", spec, s, err, want)
		}
	}
	for _, bad := range []string{"a.B", "a.B#", "#m", "a.B#m(", "a.B#m(String)V", "a.B#m(Ljava/lang/String)V", "a.B#m()X", "a b#m", "a.B#<init>()I", "a.B#m(Ljava.lang.String;)V"} {
		if _, err := ParseSelector(bad); err == nil {
			t.Errorf("ParseSelector(%q) accepted", bad)
		}
	}
	if MappingPrecision([]string{"a.B#m", "a.B#n(I)V"}) != PrecisionMixed || MappingPrecision([]string{"a.B#n(I)V"}) != PrecisionExact {
		t.Error("MappingPrecision")
	}
}

// Schema v2: versioned variants load with parsed ranges; unknown fields,
// bad ranges, bad descriptors, bad ids and ambiguous entries are errors.
func TestManualSchemaV2(t *testing.T) {
	dir := t.TempDir()
	write := func(body string) string {
		p := filepath.Join(dir, "m.yaml")
		os.WriteFile(p, []byte(body), 0o644)
		return p
	}
	good := write(`
CVE-2099-1:
  role: vulnerable
  variants:
    - package: lib:b
      versions: "[1.0,2.0)"
      methods: ["lib.B#parse(Ljava/lang/String;)V"]
    - versions: "[2.0,)"
      methods: ["lib.B#parse(Llib/Options;Ljava/lang/String;)V"]
CVE-2099-2: [a.B#m]
`)
	meta, err := LoadManualMeta(good)
	if err != nil {
		t.Fatal(err)
	}
	v := meta["CVE-2099-1"].Variants
	if len(v) != 2 || v[0].Package != "lib:b" || v[0].Range.String() != "[1.0,2.0)" || meta["CVE-2099-2"].Variants != nil {
		t.Errorf("variants = %+v", v)
	}
	m, _ := LoadManual(good)
	if len(m["CVE-2099-1"]) != 2 {
		t.Errorf("union of variant methods = %v", m["CVE-2099-1"])
	}
	for name, body := range map[string]string{
		"unknown field":         "CVE-1: {methods: [a.B#m], severity: high}",
		"unknown variant field": "CVE-1: {variants: [{versions: \"[1,2)\", methods: [a.B#m], note: x}]}",
		"bad range":             "CVE-1: {variants: [{versions: \"1.0\", methods: [a.B#m]}]}",
		"bad descriptor":        "CVE-1: {methods: [\"a.B#m(String)V\"]}",
		"both forms":            "CVE-1: {methods: [a.B#m], variants: [{versions: \"[1,2)\", methods: [a.B#m]}]}",
		"bad id":                "not an id: [a.B#m]",
		"bad package":           "CVE-1: {variants: [{package: lib, versions: \"[1,2)\", methods: [a.B#m]}]}",
		"empty variant":         "CVE-1: {variants: [{versions: \"[1,2)\"}]}",
	} {
		if _, err := LoadManual(write(body)); err == nil {
			t.Errorf("%s: accepted", name)
		} else if testing.Verbose() {
			t.Logf("%s: %v", name, strings.SplitN(err.Error(), "\n", 2)[0])
		}
	}
}

// CVE:F4 keeps the LLM's per-method roles.
func TestGroundMethodsWithRoles(t *testing.T) {
	ans := llmAnswer{Methods: []llmMethod{
		{Class: "a.B", Method: "x", Role: "entry_point"},
		{Class: "a.B", Method: "y", Role: "vulnerable"},
		{Class: "a.B", Method: "y", Role: "entry_point"},
	}}
	ms, roles, _ := groundMethodsWithRoles(ans, nil)
	if len(ms) != 2 || roles["a.B#x"] != "entry_point" || roles["a.B#y"] != "vulnerable" {
		t.Errorf("methods %v roles %v", ms, roles)
	}
}
