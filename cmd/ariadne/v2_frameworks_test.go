package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ariadne/internal/resultsv2"
)

// TestV2Frameworks (FRAMEWORKS F5 follow-up): a v2 run of testdata/di
// carries the profile's framework facts — the same injection resolutions
// the JVM made (TestDIInjectionOracle's "# inject" lines), the @Profile
// states under -active-profiles default, and the bean lifecycle roots — and
// the section is deterministic.
func TestV2Frameworks(t *testing.T) {
	read := func() resultsv2.Document {
		dir := t.TempDir()
		out := filepath.Join(dir, "r.json")
		args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", diFixture + "app.jar",
			"-deps", diFixture + "deps", "-profile", "spring", "-roots", "profile", "-active-profiles", "default",
			"-out", filepath.Join(dir, "r.md"), "-json", out, "-json-format", "v2"}
		if code, msg := runWithEnv(args, nil); code != 0 {
			t.Fatalf("exit %d: %s", code, msg)
		}
		raw, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		var doc resultsv2.Document
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		return doc
	}
	doc := read()
	f := doc.Frameworks
	if f == nil {
		t.Fatal("no frameworks section")
	}
	if f.Profile != "spring" || f.RulesVersion == "" || len(f.ActiveProfiles) != 1 || f.ActiveProfiles[0] != "default" {
		t.Errorf("profile %q, rules %q, active %v", f.Profile, f.RulesVersion, f.ActiveProfiles)
	}
	inj := map[string]string{}
	for _, i := range f.Injections {
		if i.Class == "fx.beans.Shop" && i.ParamIndex == -1 {
			inj[i.Member] = i.Resolved
		}
	}
	for m, want := range map[string]string{"store": "fx.beans.FastStore", "slow": "fx.beans.SlowStore", "archive": "fx.beans.ArchiveStore", "cold": "fx.beans.ColdStore"} {
		if inj[m] != want {
			t.Errorf("injection Shop#%s: %q, want %q (what the JVM injected)", m, inj[m], want)
		}
	}
	cond := map[string]string{}
	for _, c := range f.Conditions {
		if c.Method == "" {
			cond[c.Class] = c.Active
		}
	}
	for cls, want := range map[string]string{"fx.beans.UsTax": "true", "fx.beans.EuTax": "false", "fx.beans.CloudEuTax": "false", "fx.beans.Dormant": "false"} {
		if cond[cls] != want {
			t.Errorf("@Profile on %s: %q, want %q", cls, cond[cls], want)
		}
	}
	var cold bool
	for _, b := range f.Beans {
		if b.Class == "fx.beans.ColdStore" && b.Name == "coldStore" && !b.Explicit && b.Source == "stereotype" {
			cold = true
		}
	}
	if !cold {
		t.Error("bean fx.beans.ColdStore with default name coldStore missing")
	}
	if f.Summary.RootsByKind["constructor"] == 0 || f.Summary.Injections != len(f.Injections) || f.Summary.Beans != len(f.Beans) {
		t.Errorf("summary %+v", f.Summary)
	}
	a, _ := json.Marshal(f)
	b, _ := json.Marshal(read().Frameworks)
	if string(a) != string(b) {
		t.Error("frameworks section differs between two runs")
	}
}
