package snapshot

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ariadne/internal/httpcache"
)

func record(t *testing.T, dir string) Manifest {
	t.Helper()
	r, err := NewRecorder(dir)
	if err != nil {
		t.Fatal(err)
	}
	r.Observe(httpcache.Observation{Namespace: "llm", Key: "k1", Outcome: httpcache.OutcomeMiss}, nil)
	r.Observe(httpcache.Observation{Namespace: "llm", Key: "k1", Outcome: httpcache.OutcomeHit, SHA256: HashBytes([]byte("answer"))}, []byte("answer"))
	r.Observe(httpcache.Observation{Namespace: "llm", Key: "k1", Outcome: httpcache.OutcomeMiss}, nil) // never downgrades a hit
	r.Observe(httpcache.Observation{Namespace: "main", Key: "k2", Outcome: httpcache.OutcomeError, ErrClass: "other", Error: "HTTP 404"}, nil)
	obs, err := r.Observations()
	if err != nil {
		t.Fatal(err)
	}
	m := Manifest{CreatedAt: time.Now(), Tool: Tool{Version: "v1", Reproducible: true}, Algorithms: map[string]string{"a": "1"},
		Inputs: []Input{{Role: "dependency", Location: "deps:x.jar", SHA256: "aa"}}, Observations: obs}
	if err := Write(dir, m); err != nil {
		t.Fatal(err)
	}
	got, err := Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestRecordAndReplay(t *testing.T) {
	dir := t.TempDir()
	m := record(t, dir)
	if len(m.Observations) != 2 || m.Observations[0].Outcome != httpcache.OutcomeHit {
		t.Fatalf("observations = %+v", m.Observations)
	}
	r := NewReplayer(dir, m)
	if o, b, ok, err := r.Lookup("llm", "k1"); !ok || err != nil || string(b) != "answer" || o.Outcome != httpcache.OutcomeHit {
		t.Errorf("hit: %+v %q %v %v", o, b, ok, err)
	}
	if o, _, ok, _ := r.Lookup("main", "k2"); !ok || o.Error != "HTTP 404" {
		t.Errorf("negative observation: %+v", o)
	}
	if _, _, ok, _ := r.Lookup("main", "other"); ok {
		t.Error("unrecorded key found")
	}
	if _, err := NewRecorder(dir); err == nil {
		t.Error("recorder overwrote an existing snapshot")
	}
}

// ARIADNE-TEST-0404: an object that no longer matches its recorded hash
// is an unambiguous error, never served.
func TestCorruptObjectAndEditedManifest(t *testing.T) {
	dir := t.TempDir()
	m := record(t, dir)
	sum := HashBytes([]byte("answer"))
	os.WriteFile(filepath.Join(dir, "objects", sum[:2], sum), []byte("tampered"), 0o644)
	if _, _, _, err := NewReplayer(dir, m).Lookup("llm", "k1"); !errors.Is(err, ErrCorrupt) {
		t.Errorf("corrupt object: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
	os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(strings.Replace(string(b), "HTTP 404", "HTTP 500", 1)), 0o644)
	if _, err := Read(dir); err == nil || !strings.Contains(err.Error(), "edited manifest") {
		t.Errorf("edited manifest: %v", err)
	}
}

func TestDataSetIDIgnoresTimes(t *testing.T) {
	a := []httpcache.Observation{{Namespace: "main", Key: "k", Outcome: "hit", SHA256: "x", Meta: &httpcache.Meta{FetchedAt: time.Now()}}}
	b := []httpcache.Observation{{Namespace: "main", Key: "k", Outcome: "hit", SHA256: "x", Meta: &httpcache.Meta{FetchedAt: time.Now().Add(time.Hour)}}}
	if DataSetID(a) != DataSetID(b) {
		t.Error("data set ID depends on fetch time")
	}
	b[0].SHA256 = "y"
	if DataSetID(a) == DataSetID(b) {
		t.Error("data set ID ignores content")
	}
}

func TestCheckToolAndInputs(t *testing.T) {
	m := Manifest{Tool: Tool{Version: "v1"}, Algorithms: map[string]string{"reach.engine": "1"}, Inputs: []Input{{Role: "dependency", Location: "deps:x.jar", SHA256: "aa"}}}
	if err := CheckTool(m, Tool{Version: "v1"}, map[string]string{"reach.engine": "1"}); err != nil {
		t.Errorf("same tool: %v", err)
	}
	if err := CheckTool(m, Tool{Version: "v2"}, map[string]string{"reach.engine": "1"}); err == nil {
		t.Error("other version accepted")
	}
	if err := CheckTool(m, Tool{Version: "v1"}, map[string]string{"reach.engine": "2"}); err == nil || !strings.Contains(err.Error(), "reach.engine 1 → 2") {
		t.Errorf("algorithm change: %v", err)
	}
	dirty := Manifest{Tool: Tool{Version: "v1-dirty", Dirty: true, BinarySHA256: "b1"}}
	if err := CheckTool(dirty, Tool{Version: "v1-dirty", BinarySHA256: "b2"}, nil); err == nil {
		t.Error("other dirty binary accepted")
	}
	if err := CheckInputs(m, []Input{{Role: "dependency", Location: "deps:x.jar", SHA256: "aa"}}); err != nil {
		t.Errorf("same inputs: %v", err)
	}
	for _, in := range [][]Input{
		{{Role: "dependency", Location: "deps:x.jar", SHA256: "bb"}},
		{},
		{{Role: "dependency", Location: "deps:x.jar", SHA256: "aa"}, {Role: "dependency", Location: "deps:y.jar", SHA256: "cc"}},
	} {
		if err := CheckInputs(m, in); err == nil {
			t.Errorf("mismatching inputs %v accepted", in)
		}
	}
}

// Changes of advisory, mapping or algorithm invalidate the right
// derivations, and only those.
func TestCompare(t *testing.T) {
	d := func(id, adv, mapping string) Derivation {
		x := Derivation{Advisory: id, AdvisoryHash: adv, Affected: []string{"g:a@1"}, Source: "diff", MappingHash: mapping, Resolver: "1"}
		x.ComputeKey()
		return x
	}
	old := Manifest{Algorithms: map[string]string{"r": "1"}, Derivations: []Derivation{d("A", "a1", "m"), d("B", "b1", "m"), d("C", "c1", "m"), d("GONE", "g", "m")},
		Verdicts: map[string]string{"A": "REACHABLE", "B": "NOT_REACHABLE", "C": "FRAMEWORK", "GONE": "UNKNOWN"}}
	cur := Manifest{Algorithms: map[string]string{"r": "1"}, Derivations: []Derivation{d("A", "a1", "m"), d("B", "b2", "m"), d("C", "c1", "m2"), d("NEW", "n", "m")},
		Verdicts: map[string]string{"A": "REACHABLE", "B": "UNKNOWN", "C": "FRAMEWORK", "NEW": "REACHABLE"}}
	got := map[string]Change{}
	for _, c := range Compare(old, cur) {
		got[c.Advisory] = c
	}
	if _, ok := got["A"]; ok {
		t.Error("unchanged advisory reported")
	}
	if c := got["B"]; c.Verdict != "NOT_REACHABLE → UNKNOWN" || !strings.Contains(strings.Join(c.Reasons, ";"), "advisory content changed") {
		t.Errorf("B = %+v", c)
	}
	if c := got["C"]; c.Verdict != "" || !strings.Contains(strings.Join(c.Reasons, ";"), "mapping changed") {
		t.Errorf("C = %+v", c)
	}
	if got["NEW"].Reasons[0] != "new advisory" || got["NEW"].Verdict != "(none) → REACHABLE" || got["GONE"].Reasons[0] != "advisory no longer reported" {
		t.Errorf("new/gone = %+v %+v", got["NEW"], got["GONE"])
	}
	cur.Algorithms["r"] = "2"
	if n := len(Compare(old, cur)); n != 5 {
		t.Errorf("algorithm change invalidates %d advisories, want all 5", n)
	}
}

func TestReplayerKeepsFailures(t *testing.T) {
	dir := t.TempDir()
	m := record(t, dir)
	r := NewReplayer(dir, m)
	r.Lookup("main", "unrecorded")
	sum := HashBytes([]byte("answer"))
	os.WriteFile(filepath.Join(dir, "objects", sum[:2], sum), []byte("tampered"), 0o644)
	r.Lookup("llm", "k1")
	if f := r.Failures(); len(f) != 2 {
		t.Errorf("failures = %v", f)
	}
}
