package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/resultsv2"
	"ariadne/internal/snapshot"
)

// CVE:F9 for the F8 increment: the target runtime assessment end to end,
// through the CLI, on a small offline application.

const rtFixtures = "../../internal/runtimeinfo/testdata/"

type rtRun struct {
	dir  string
	code int
	msg  string
}

func runTarget(t *testing.T, format string, extra ...string) rtRun {
	t.Helper()
	dir := t.TempDir()
	args := append([]string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", reflectionFixture + "app.jar",
		"-profile", "generic", "-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-json-format", format}, extra...)
	code, msg := runWithEnv(args, nil)
	return rtRun{dir: dir, code: code, msg: msg}
}

func (r rtRun) file(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(r.dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func (r rtRun) v2(t *testing.T) resultsv2.Document {
	t.Helper()
	var d resultsv2.Document
	if err := json.Unmarshal(r.file(t, "r.json"), &d); err != nil {
		t.Fatal(err)
	}
	return d
}

// TestTargetRuntimeE2E: the v1 rows never depend on the target runtime
// (F9 task 4); v2 reports the assessment separately (task 2) — performed
// for Temurin with every advisory's reachability not_assessed, not
// performed with a reason for another distribution or without a runtime;
// the report has its section; two runs give the same analysis id (task 5).
func TestTargetRuntimeE2E(t *testing.T) {
	none := runTarget(t, "v1")
	tem := runTarget(t, "v1", "-jdk-home", rtFixtures+"temurin-17")
	deb := runTarget(t, "v1", "-jdk-home", rtFixtures+"debian-25")
	for _, r := range []rtRun{none, tem, deb} {
		if r.code != 0 {
			t.Fatalf("exit %d: %s", r.code, r.msg)
		}
	}
	if string(none.file(t, "r.json")) != string(tem.file(t, "r.json")) || string(none.file(t, "r.json")) != string(deb.file(t, "r.json")) {
		t.Error("v1 results differ with a target runtime: runtime advisories must never enter the v1 rows")
	}
	if !strings.Contains(string(tem.file(t, "r.md")), "Target runtime") {
		t.Error("report without the Target runtime section")
	}

	for name, c := range map[string]struct {
		args      []string
		performed bool
	}{
		"none":    {nil, false},
		"temurin": {[]string{"-jdk-home", rtFixtures + "temurin-17"}, true},
		"debian":  {[]string{"-jdk-home", rtFixtures + "debian-25"}, false},
	} {
		a, b := runTarget(t, "v2", c.args...), runTarget(t, "v2", c.args...)
		if a.code != 0 || b.code != 0 {
			t.Fatalf("%s: exit %d/%d: %s %s", name, a.code, b.code, a.msg, b.msg)
		}
		da, db := a.v2(t), b.v2(t)
		if da.AnalysisID == "" || da.AnalysisID != db.AnalysisID {
			t.Errorf("%s: analysis ids %q and %q of two identical runs differ", name, da.AnalysisID, db.AnalysisID)
		}
		tr := da.TargetRuntime
		if tr == nil {
			t.Fatalf("%s: no target_runtime section", name)
		}
		raw, _ := json.Marshal(tr)
		if c.performed {
			if tr.Advisories == nil || !strings.Contains(string(raw), "affected") {
				t.Errorf("%s: assessment %s, want advisories with an affected release", name, raw)
			}
			if strings.Contains(string(raw), `"reachability":"`) && !strings.Contains(string(raw), `"reachability":"not_assessed"`) {
				t.Errorf("%s: a runtime advisory has a reachability other than not_assessed: %s", name, raw)
			}
		} else if tr.Assessment != resultsv2.AssessmentNotPerformed || tr.Reason == "" {
			t.Errorf("%s: assessment %q reason %q, want not_performed with a reason", name, tr.Assessment, tr.Reason)
		}
	}
}

// TestTargetRuntimeSnapshot: a snapshot records the runtime's identity in
// its configuration and a custom catalog's content as an input (task 5);
// replaying it gives the same document.
func TestTargetRuntimeSnapshot(t *testing.T) {
	cat, err := os.ReadFile("../../internal/runtimevuln/openjdk_advisories.yaml")
	if err != nil {
		t.Fatal(err)
	}
	catFile := filepath.Join(t.TempDir(), "catalog.yaml")
	if err := os.WriteFile(catFile, cat, 0o644); err != nil {
		t.Fatal(err)
	}
	snap := filepath.Join(t.TempDir(), "snap")
	out := runTarget(t, "v2", "-jdk-home", rtFixtures+"temurin-17", "-openjdk-catalog", catFile, "-snapshot-out", snap)
	if out.code != 0 {
		t.Fatalf("exit %d: %s", out.code, out.msg)
	}
	m, err := snapshot.Read(snap)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(m.Config["target-runtime"], "17.0.9") {
		t.Errorf("config target-runtime %q, want the runtime's identity", m.Config["target-runtime"])
	}
	found := false
	for _, in := range m.Inputs {
		if in.Role == "openjdk-catalog" && in.SHA256 != "" {
			found = true
		}
	}
	if !found {
		t.Errorf("inputs %+v: the custom catalog's content is not recorded", m.Inputs)
	}
	replay := runTarget(t, "v2", "-jdk-home", rtFixtures+"temurin-17", "-openjdk-catalog", catFile, "-snapshot-in", snap)
	if replay.code != 0 {
		t.Fatalf("replay exit %d: %s", replay.code, replay.msg)
	}
	// A replay is marked as one (snapshot.data_set.replayed, freshness
	// "replayed"), so its analysis id differs by design; the runtime
	// assessment and the findings must not.
	same := func(a, b resultsv2.Document) bool {
		ja, _ := json.Marshal([]any{a.TargetRuntime, a.Findings})
		jb, _ := json.Marshal([]any{b.TargetRuntime, b.Findings})
		return string(ja) == string(jb)
	}
	if !same(out.v2(t), replay.v2(t)) {
		t.Error("the replay's target runtime assessment or findings differ from the recorded run's")
	}

	// The catalog-age note is dated by the snapshot, not the wall clock: a
	// snapshot dated a year after the catalog's newest advisory replays
	// with the note today, although a live run today has none.
	raw, err := os.ReadFile(filepath.Join(snap, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var mm map[string]any
	if err := json.Unmarshal(raw, &mm); err != nil {
		t.Fatal(err)
	}
	mm["created_at"] = "2027-11-17T00:00:00Z"
	raw, _ = json.Marshal(mm)
	if err := os.WriteFile(filepath.Join(snap, "manifest.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	late := runTarget(t, "v2", "-jdk-home", rtFixtures+"temurin-17", "-openjdk-catalog", catFile, "-snapshot-in", snap)
	if late.code != 0 {
		t.Fatalf("replay of the re-dated snapshot: exit %d: %s", late.code, late.msg)
	}
	note := func(d resultsv2.Document) bool {
		j, _ := json.Marshal(d.TargetRuntime)
		return strings.Contains(string(j), "advisories published since are not included")
	}
	if note(out.v2(t)) {
		t.Skip("the embedded catalog is already older than StaleAfter today; the dating check needs a fresh catalog")
	}
	if !note(late.v2(t)) {
		t.Error("replaying a snapshot dated 2027-11-17: no catalog-age note — the note is dated by the wall clock, not the snapshot")
	}
}

// TestTargetRuntimeErrors: bad runtime inputs fail the run before any
// result is written (task 8: a global error must not leave a report that
// looks complete).
func TestTargetRuntimeErrors(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(bad, []byte("vendor: [unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][]string{
		"both flags":       {"-jdk-home", rtFixtures + "temurin-17", "-runtime-file", bad},
		"no release file":  {"-jdk-home", t.TempDir()},
		"bad runtime file": {"-runtime-file", bad},
		"bad catalog":      {"-jdk-home", rtFixtures + "temurin-17", "-openjdk-catalog", bad},
		"missing catalog":  {"-jdk-home", rtFixtures + "temurin-17", "-openjdk-catalog", filepath.Join(t.TempDir(), "none.yaml")},
	} {
		r := runTarget(t, "v2", args...)
		if r.code == 0 {
			t.Errorf("%s: exit 0, want a failure", name)
		}
		for _, f := range []string{"r.md", "r.json"} {
			if _, err := os.Stat(filepath.Join(r.dir, f)); err == nil {
				t.Errorf("%s: %s written despite the failure (%s)", name, f, strings.TrimSpace(r.msg))
			}
		}
	}
}
