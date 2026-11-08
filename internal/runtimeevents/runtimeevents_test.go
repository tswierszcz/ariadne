package runtimeevents

import (
	"os"
	"strings"
	"testing"
)

const fixtureTrace = "../../testdata/runtime/run.trace"

func readFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(fixtureTrace)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func parse(t *testing.T, s string) (*Trace, error) {
	t.Helper()
	return Parse(strings.NewReader(s), "test")
}

// TestParseFixture: the JFR-converted fixture trace is valid and carries
// every evidence kind it declares.
func TestParseFixture(t *testing.T) {
	tr, err := parse(t, readFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if tr.Header.Format != Format || tr.Header.Completeness.Complete || tr.Header.Sampling.Kind != "periodic" {
		t.Errorf("header %+v", tr.Header)
	}
	if len(tr.Loaded) == 0 || len(tr.Executed) == 0 || len(tr.Calls) == 0 {
		t.Errorf("loaded %d, executed %d, calls %d", len(tr.Loaded), len(tr.Executed), len(tr.Calls))
	}
	plugins := 0
	for _, c := range tr.Classes {
		if c.Name == "rtm.Plugin" {
			plugins++
		}
	}
	if plugins != 2 {
		t.Errorf("rtm.Plugin defined %d times, want 2 (two loaders)", plugins)
	}
}

// TestParseRejects: every contract violation rejects the whole file.
func TestParseRejects(t *testing.T) {
	good := readFixture(t)
	lines := strings.Split(strings.TrimRight(good, "\n"), "\n")
	header := lines[0]
	replace := func(old, new string) string { return strings.Replace(good, old, new, 1) }
	cases := map[string]string{
		"empty":                "",
		"wrong format":         replace(`"format":"ariadne-runtime-trace-1"`, `"format":"ariadne-runtime-trace-9"`),
		"no header first":      strings.Join(append(lines[1:2], lines...), "\n"),
		"unknown header field": replace(`"data_policy":"identifiers-only"`, `"data_policy":"identifiers-only","payload":"x"`),
		"payload policy":       replace(`"data_policy":"identifiers-only"`, `"data_policy":"full"`),
		"sampled yet complete": replace(`"complete":false`, `"complete":true`),
		"unknown record":       good + `{"record":"guess","id":"z1"}` + "\n",
		"unknown field":        good + `{"record":"loaded","id":"z1","class":"c1","time":5}` + "\n",
		"dangling class":       good + `{"record":"loaded","id":"z1","class":"c999"}` + "\n",
		"dangling method":      good + `{"record":"call","id":"z1","caller":"m1","callee":"m999","basis":"stack-adjacent","source":"sampled","count":1}` + "\n",
		"duplicate evidence":   good + `{"record":"loaded","id":"e1","class":"c1"}` + "\n",
		"zero count":           good + `{"record":"executed","id":"z1","method":"m1","basis":"sampled","count":0}` + "\n",
		"bad basis":            good + `{"record":"executed","id":"z1","method":"m1","basis":"guessed","count":1}` + "\n",
		"undeclared kind":      good + `{"record":"handoff","id":"z1","from":"m1","to":"m2","mechanism":"executor","basis":"token"}` + "\n",
		"unknown loader":       header + "\n" + `{"record":"class","id":"c1","name":"a.B","loader":"l9","source_basis":"observed"}` + "\n",
		"slashed class name":   header + "\n" + `{"record":"loader","id":"l1","name":"app","type":""}` + "\n" + `{"record":"class","id":"c1","name":"a/B","loader":"l1","source_basis":"observed"}` + "\n",
		"same class same loader": header + "\n" + `{"record":"loader","id":"l1","name":"app","type":""}` + "\n" +
			`{"record":"class","id":"c1","name":"a.B","loader":"l1","source_basis":"observed"}` + "\n" + `{"record":"class","id":"c2","name":"a.B","loader":"l1","source_basis":"observed"}` + "\n",
		"bad sha": header + "\n" + `{"record":"artifact","id":"a1","name":"x.jar","sha256":"XYZ"}` + "\n",
		"no desc": header + "\n" + `{"record":"loader","id":"l1","name":"app","type":""}` + "\n" + `{"record":"class","id":"c1","name":"a.B","loader":"l1","source_basis":"observed"}` + "\n" +
			`{"record":"method","id":"m1","class":"c1","name":"run","desc":""}` + "\n",
		"long line": header + "\n" + `{"record":"loader","id":"l1","name":"` + strings.Repeat("x", MaxLine) + `","type":""}` + "\n",
	}
	for name, s := range cases {
		if _, err := parse(t, s); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := parse(t, good); err != nil {
		t.Fatalf("the unmodified fixture is rejected: %v", err)
	}
}

// TestCorrelateStatuses: a class is usable evidence only when its hash
// matches the analyzed class; the two definitions stay apart.
func TestCorrelateStatuses(t *testing.T) {
	tr, err := parse(t, readFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	var squareSHA string
	for _, c := range tr.Classes {
		if c.Name == "rtm.Square" {
			squareSHA = c.SHA256
		}
	}
	lookup := func(other bool) func(string) (string, string, bool) {
		return func(fqcn string) (string, string, bool) {
			for _, c := range tr.Classes {
				if c.Name != fqcn || fqcn == "rtm.Lib" {
					continue
				}
				if other && fqcn == "rtm.Square" {
					return strings.Repeat("0", 64), "", true
				}
				return c.SHA256, "", true
			}
			return "", "", false // rtm.Lib is not in the analyzed input
		}
	}
	r := Correlate(tr, Inputs{Lookup: lookup(false)})
	st := map[string][]string{}
	for _, c := range r.Classes {
		st[c.Name] = append(st[c.Name], c.Status)
	}
	if strings.Join(st["rtm.Square"], ",") != StatusMatched || strings.Join(st["rtm.Lib"], ",") != StatusNotInInput {
		t.Errorf("statuses %v", st)
	}
	if strings.Join(st["rtm.Plugin"], ",") != StatusMatched+","+StatusMatched || strings.Join(r.MultiDefined, ",") != "rtm.Plugin" {
		t.Errorf("two definitions: %v, multi %v", st["rtm.Plugin"], r.MultiDefined)
	}
	r2 := Correlate(tr, Inputs{Lookup: lookup(true)})
	for _, c := range r2.Classes {
		if c.Name == "rtm.Square" && c.Status != StatusDifferentVersion {
			t.Errorf("another build of Square: %s", c.Status)
		}
	}
	// No hash at all: name only, never usable.
	var out []string
	for _, l := range strings.Split(readFixture(t), "\n") {
		if strings.Contains(l, `"name":"rtm.Square"`) {
			l = strings.Replace(l, `,"sha256":"`+squareSHA+`"`, "", 1)
			l = strings.Replace(l, `,"artifact":"a1"`, "", 1)
		}
		out = append(out, l)
	}
	stripped := strings.Join(out, "\n")
	tr3, err := parse(t, stripped)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range Correlate(tr3, Inputs{Lookup: lookup(false)}).Classes {
		if c.Name == "rtm.Square" && c.Status != StatusNameOnly {
			t.Errorf("Square without hashes: %s, want %s", c.Status, StatusNameOnly)
		}
	}
}

// TestObservedVulnerable: a mapping without a descriptor matches any
// overload, with one only that descriptor; unusable classes never match.
func TestObservedVulnerable(t *testing.T) {
	r := Result{Executed: []MethodObservation{
		{Evidence: "e1", Method: "a.B#run(I)V", ClassStatus: StatusMatched, InGraph: true},
		{Evidence: "e2", Method: "a.B#run(J)V", ClassStatus: StatusMatched, InGraph: true},
		{Evidence: "e3", Method: "a.B#runner()V", ClassStatus: StatusMatched, InGraph: true},
		{Evidence: "e4", Method: "a.C#run()V", ClassStatus: StatusDifferentVersion, InGraph: true},
	}}
	ids := func(os []MethodObservation) string {
		var s []string
		for _, o := range os {
			s = append(s, o.Evidence)
		}
		return strings.Join(s, ",")
	}
	if got := ids(r.ObservedVulnerable([]string{"a.B#run"})); got != "e1,e2" {
		t.Errorf("name mapping: %s", got)
	}
	if got := ids(r.ObservedVulnerable([]string{"a.B#run(J)V"})); got != "e2" {
		t.Errorf("descriptor mapping: %s", got)
	}
	if got := ids(r.ObservedVulnerable([]string{"a.C#run"})); got != "" {
		t.Errorf("different version used: %s", got)
	}
}
