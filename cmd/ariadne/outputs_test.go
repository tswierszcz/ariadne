package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// CVE:F9.8/F9.9: end-to-end checks of the run's output artifacts on the
// small reach fixture, without network (no dependency jars, so no OSV
// query is needed; with dependencies and an empty offline cache the OSV
// lookup fails).

const fixtureApp = "../../testdata/reachfixture/appentry.jar"

type outRun struct {
	dir, md, json, cache string
}

func newOutRun(t *testing.T) outRun {
	dir := t.TempDir()
	r := outRun{dir: dir, md: filepath.Join(dir, "report.md"), json: filepath.Join(dir, "results.json"), cache: filepath.Join(dir, "cache")}
	os.MkdirAll(filepath.Join(dir, "deps"), 0o755)
	return r
}

func (r outRun) run(t *testing.T, extra ...string) (int, string) {
	t.Helper()
	args := append([]string{"-offline", "-cache-dir", r.cache, "-app", fixtureApp, "-deps", filepath.Join(r.dir, "deps"), "-out", r.md, "-json", r.json}, extra...)
	return runWithEnv(args, nil)
}

func fileSHA(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

var reportID = regexp.MustCompile("\\*\\*Analysis ID:\\*\\* `([^`]+)` — results file `([^`]+)` \\(([^,]+), sha256 `([0-9a-f]+)`\\)")

// TestOutputsShareAnalysisID: the report names the v2 document's
// analysis_id and the results file's hash; the snapshot manifest carries
// the same ID; a second run on the same input and data gives the same ID
// and byte-identical v2.
func TestOutputsShareAnalysisID(t *testing.T) {
	r := newOutRun(t)
	snap := filepath.Join(r.dir, "snap")
	if code, stderr := r.run(t, "-json-format", "v2", "-snapshot-out", snap); code != 0 {
		t.Fatalf("code %d: %s", code, stderr)
	}
	var doc struct {
		AnalysisID string `json:"analysis_id"`
	}
	b, _ := os.ReadFile(r.json)
	if err := json.Unmarshal(b, &doc); err != nil || !strings.HasPrefix(doc.AnalysisID, "sha256:") {
		t.Fatalf("v2 analysis_id %q %v", doc.AnalysisID, err)
	}
	md, _ := os.ReadFile(r.md)
	m := reportID.FindStringSubmatch(string(md))
	if m == nil || m[1] != doc.AnalysisID || m[2] != "results.json" || m[3] != "v2" || m[4] != fileSHA(t, r.json) {
		t.Fatalf("report header %v, want %s and the results hash", m, doc.AnalysisID)
	}
	var man struct {
		AnalysisID string `json:"analysis_id"`
	}
	mb, _ := os.ReadFile(filepath.Join(snap, "manifest.json"))
	if err := json.Unmarshal(mb, &man); err != nil || man.AnalysisID != doc.AnalysisID {
		t.Errorf("manifest analysis_id %q, want %q (%v)", man.AnalysisID, doc.AnalysisID, err)
	}

	first := fileSHA(t, r.json)
	if code, stderr := r.run(t, "-json-format", "v2", "-snapshot-out", filepath.Join(r.dir, "snap2")); code != 0 {
		t.Fatalf("second run %d: %s", code, stderr)
	}
	if fileSHA(t, r.json) != first {
		t.Errorf("v2 differs between two runs on the same input")
	}
	es, _ := os.ReadDir(r.dir)
	for _, e := range es {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temporary file left: %s", e.Name())
		}
	}
}

// TestV1ReportNamesResultsHash: with the default v1 format the results
// file has no analysis_id field; the report still binds to it by hash.
func TestV1ReportNamesResultsHash(t *testing.T) {
	r := newOutRun(t)
	if code, stderr := r.run(t); code != 0 {
		t.Fatalf("code %d: %s", code, stderr)
	}
	md, _ := os.ReadFile(r.md)
	m := reportID.FindStringSubmatch(string(md))
	if m == nil || m[3] != "v1" || m[4] != fileSHA(t, r.json) {
		t.Fatalf("report header %v", m)
	}
}

// TestFailedWriteReplacesNothing: the results file cannot be written (its
// directory does not exist). Before CVE:F9.9 the report was already
// written by then; now no output is replaced — the earlier report stays.
func TestFailedWriteReplacesNothing(t *testing.T) {
	r := newOutRun(t)
	os.WriteFile(r.md, []byte("earlier report\n"), 0o644)
	r.json = filepath.Join(r.dir, "missing", "results.json")
	code, stderr := r.run(t)
	if code == 0 || !strings.Contains(stderr, "no output file was replaced") {
		t.Fatalf("code %d stderr %q", code, stderr)
	}
	if b, _ := os.ReadFile(r.md); string(b) != "earlier report\n" {
		t.Errorf("report replaced: %q", b)
	}
}

// TestGlobalDataErrorWritesNothing: a dependency with an identity needs
// an OSV answer; offline with an empty cache that is a global error —
// no report, no results file.
func TestGlobalDataErrorWritesNothing(t *testing.T) {
	r := newOutRun(t)
	src, _ := os.ReadFile("../../testdata/reachfixture/vulnerablelib.jar")
	os.WriteFile(filepath.Join(r.dir, "deps", "lib-1.0.jar"), src, 0o644)
	list := filepath.Join(r.dir, "deps.txt")
	os.WriteFile(list, []byte("org.vulnerable:lib:jar:1.0:compile\n"), 0o644)
	code, stderr := r.run(t, "-deps-list", list)
	if code == 0 || !strings.Contains(stderr, "offline and not in cache") {
		t.Fatalf("code %d stderr %q", code, stderr)
	}
	for _, p := range []string{r.md, r.json} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("%s written after a global error", p)
		}
	}
}

// TestBrokenSnapshotWritesNothing: replaying a corrupt manifest, or one
// of an unsupported format, fails before any output.
func TestBrokenSnapshotWritesNothing(t *testing.T) {
	for name, body := range map[string]string{
		"corrupt":   "{not json",
		"format 99": `{"format_version": 99}`,
	} {
		r := newOutRun(t)
		snap := filepath.Join(r.dir, "snap")
		os.MkdirAll(snap, 0o755)
		os.WriteFile(filepath.Join(snap, "manifest.json"), []byte(body), 0o644)
		args := []string{"-cache-dir", r.cache, "-app", fixtureApp, "-deps", filepath.Join(r.dir, "deps"), "-out", r.md, "-json", r.json, "-snapshot-in", snap}
		code, stderr := runWithEnv(args, nil)
		if code == 0 {
			t.Errorf("%s: accepted", name)
		}
		if _, err := os.Stat(r.md); err == nil {
			t.Errorf("%s: report written (%s)", name, stderr)
		}
	}
}
