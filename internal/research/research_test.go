package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const contractDir = "../../docs/research-interface/v1"

func loadFixture(t *testing.T) *Bundle {
	t.Helper()
	b, err := Load(filepath.Join(contractDir, "fixtures/bundle"))
	if err != nil {
		t.Fatalf("loading the RI0 fixture bundle: %v", err)
	}
	return b
}

func readFile(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// The RI0 fixture requests replayed against the fixture bundle give the
// documented results: resolution, search state, entries and paths, model
// state and relevant gaps (work units follow this implementation's
// research-work-1 accounting and are not compared).
func TestFixtureConformance(t *testing.T) {
	b := loadFixture(t)
	var cases struct {
		Cases []struct {
			ID, Path, Definition, Request string
			SemanticValid                 bool `json:"semantic_valid"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readFile(t, filepath.Join(contractDir, "cases.json")), &cases); err != nil {
		t.Fatal(err)
	}
	compared := 0
	for _, c := range cases.Cases {
		if c.Definition != "Response" || !c.SemanticValid || c.Request == "" {
			continue
		}
		var want map[string]any
		json.Unmarshal(readFile(t, filepath.Join(contractDir, c.Path)), &want)
		if want["status"] != "ok" || strings.Contains(string(readFile(t, filepath.Join(contractDir, c.Path))), `"work_complete": false`) {
			continue // a budget cutoff the fixture shows for illustration; TestBudget covers it
		}
		req, qerr := ParseRequest(b, readFile(t, filepath.Join(contractDir, c.Request)))
		if qerr != nil {
			t.Errorf("%s: request rejected: %v", c.ID, qerr)
			continue
		}
		if req.Limits != nil && req.Limits.MaxWorkUnits != nil {
			continue // budget cases depend on the accounting, checked below
		}
		got, qerr := Query(b, req)
		if qerr != nil {
			t.Errorf("%s: %v", c.ID, qerr)
			continue
		}
		gb, _ := json.Marshal(got)
		var g map[string]any
		json.Unmarshal(gb, &g)
		compared++
		if req.Operation != "reachability" {
			if !jsonEqual(g["records"], want["records"]) {
				t.Errorf("%s: records\n got %v\nwant %v", c.ID, g["records"], want["records"])
			}
			gp, wp := g["page"].(map[string]any), want["page"].(map[string]any)
			for _, k := range []string{"returned_count", "total_count", "truncated", "work_complete"} {
				if !jsonEqual(gp[k], wp[k]) {
					t.Errorf("%s: page.%s = %v, want %v", c.ID, k, gp[k], wp[k])
				}
			}
			continue
		}
		gr, wr := g["results"].([]any), want["results"].([]any)
		if len(gr) != len(wr) {
			t.Fatalf("%s: %d results, want %d", c.ID, len(gr), len(wr))
		}
		for i := range gr {
			x, y := gr[i].(map[string]any), wr[i].(map[string]any)
			for _, k := range []string{"target_id", "resolution", "matched_node_ids", "total_entry_count", "entries_truncated", "root_basis"} {
				if !jsonEqual(x[k], y[k]) {
					t.Errorf("%s: %s = %v, want %v", c.ID, k, x[k], y[k])
				}
			}
			if xs, ys := x["search"].(map[string]any), y["search"].(map[string]any); xs["state"] != ys["state"] || xs["work_complete"] != ys["work_complete"] {
				t.Errorf("%s: search = %v, want %v", c.ID, xs, ys)
			}
			if !jsonEqual(x["model"], y["model"]) {
				t.Errorf("%s: model = %v, want %v", c.ID, x["model"], y["model"])
			}
			if !jsonEqual(x["entries"], y["entries"]) {
				t.Errorf("%s: entries\n got %v\nwant %v", c.ID, x["entries"], y["entries"])
			}
		}
	}
	if compared < 5 {
		t.Fatalf("compared %d fixture responses, expected at least 5", compared)
	}
}

func jsonEqual(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

func request(t *testing.T, b *Bundle, body string) (any, *Error) {
	t.Helper()
	raw := strings.ReplaceAll(body, "@ID@", b.AnalysisID)
	req, err := ParseRequest(b, []byte(raw))
	if err != nil {
		return nil, err
	}
	return Query(b, req)
}

func TestBudget(t *testing.T) {
	b := loadFixture(t)
	// One unit: the shared forward search stops at once; no witness.
	got, err := request(t, b, `{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"reachability",
		"limits":{"max_work_units":1},"targets":[{"id":"a","method_id":"m-write"},{"id":"b","method_id":"m-write"}],"roots":{"selection":"indexed"}}`)
	if err != nil {
		t.Fatal(err)
	}
	rs := got.(ReachabilityResponse).Results
	for i, r := range rs {
		if r.Search.State != "incomplete" || r.Search.WorkComplete || r.TotalEntryCount != nil || len(r.Entries) != 0 {
			t.Errorf("target %d: %+v", i, r)
		}
	}
	if rs[1].Search.WorkUnits != 0 {
		t.Errorf("later target charged %d units after the budget ended", rs[1].Search.WorkUnits)
	}
	// Enough for the shared search and part of the path search: the
	// witnesses found stay, the search is not complete.
	got, err = request(t, b, `{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"reachability",
		"limits":{"max_work_units":7},"targets":[{"id":"a","method_id":"m-write"}],"roots":{"selection":"indexed"}}`)
	if err != nil {
		t.Fatal(err)
	}
	r := got.(ReachabilityResponse).Results[0]
	if r.Search.State != "found" || r.Search.WorkComplete || r.TotalEntryCount != nil || !r.EntriesTruncated || len(r.Entries) != 1 {
		t.Errorf("found under a partial budget: %+v", r)
	}
	if r.Search.WorkUnits > 7 {
		t.Errorf("work %d over the limit", r.Search.WorkUnits)
	}
}

func TestRequestErrors(t *testing.T) {
	b := loadFixture(t)
	for _, c := range []struct{ body, code, loc string }{
		{`{"schema_version":"ariadne-research-query-2","analysis_id":"@ID@","request_id":"r","operation":"entries"}`, "UNSUPPORTED_SCHEMA", "/schema_version"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"entries","operation":"methods"}`, "INVALID_REQUEST", "/"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"sha256:` + strings.Repeat("0", 64) + `","request_id":"r","operation":"entries"}`, "ANALYSIS_MISMATCH", "/analysis_id"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"trace-config"}`, "UNSUPPORTED_OPERATION", "/operation"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"entries","filters":{"reachable_only":true}}`, "UNSUPPORTED_FILTER", "/filters/reachable_only"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"entries","bogus":1}`, "INVALID_REQUEST", "/"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"entries","targets":[]}`, "INVALID_REQUEST", "/targets"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"reachability","targets":[{"id":"a","method_id":"m-write"},{"id":"a","method_id":"m-write"}],"roots":{"selection":"indexed"}}`, "INVALID_REQUEST", "/targets/1/id"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"reachability","targets":[{"id":"a","method_id":"m-write","selector":{"instance_id":"c-library","class":"A","name":"b","descriptor":"()V"}}],"roots":{"selection":"indexed"}}`, "INVALID_REQUEST", "/targets/0"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"reachability","targets":[{"id":"a","selector":{"instance_id":"c-library","class":"A","name":"b","descriptor":"(Lx)V"}}],"roots":{"selection":"indexed"}}`, "INVALID_REQUEST", "/targets/0/selector/descriptor"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"reachability","targets":[{"id":"a","method_id":"m-write"}],"roots":{"selection":"subset","node_ids":["n-write"]}}`, "ROOTS_OUTSIDE_INDEX", "/roots/node_ids/0"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"callers","node_id":"n-nope"}`, "INVALID_REQUEST", "/node_id"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"methods","cursor":"c1.bm9wZQ"}`, "INVALID_CURSOR", "/cursor"},
		{`{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"methods","limits":{"max_records":0}}`, "INVALID_REQUEST", "/limits/max_records"},
	} {
		_, err := request(t, b, c.body)
		if err == nil || err.Code != c.code || err.Location != c.loc {
			t.Errorf("%s\n  error %+v, want %s at %s", c.body, err, c.code, c.loc)
		}
	}
}

// Paging: every record once, in id order, and a cursor bound to its
// filters and limit.
func TestPaging(t *testing.T) {
	b := loadFixture(t)
	var ids []string
	cursor := "null"
	for page := 0; page < 10; page++ {
		got, err := request(t, b, `{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"methods","limits":{"max_records":2},"cursor":`+cursor+`}`)
		if err != nil {
			t.Fatal(err)
		}
		d := got.(DiscoveryResponse)
		for _, r := range d.Records {
			ids = append(ids, r.(Node).NodeID)
		}
		if *d.Page.TotalCount != 3 {
			t.Fatalf("total_count %d", *d.Page.TotalCount)
		}
		if d.Page.NextCursor == nil {
			break
		}
		c, _ := json.Marshal(*d.Page.NextCursor)
		cursor = string(c)
		// The same cursor with another limit is rejected.
		if _, err := request(t, b, `{"schema_version":"ariadne-research-query-1","analysis_id":"@ID@","request_id":"r","operation":"methods","limits":{"max_records":3},"cursor":`+cursor+`}`); err == nil || err.Code != "INVALID_CURSOR" {
			t.Errorf("cursor accepted under another limit: %v", err)
		}
	}
	if strings.Join(ids, ",") != "n-extract,n-unused,n-write" {
		t.Errorf("paged ids %v", ids)
	}
}

func copyDir(t *testing.T, from string) string {
	t.Helper()
	to := filepath.Join(t.TempDir(), "bundle")
	os.MkdirAll(to, 0o755)
	entries, _ := os.ReadDir(from)
	for _, e := range entries {
		os.WriteFile(filepath.Join(to, e.Name()), readFile(t, filepath.Join(from, e.Name())), 0o644)
	}
	return to
}

func TestLoadRejectsCorruption(t *testing.T) {
	src := filepath.Join(contractDir, "fixtures/bundle")
	for name, mutate := range map[string]func(dir string){
		"record changed": func(dir string) {
			p := filepath.Join(dir, "records.jsonl")
			os.WriteFile(p, []byte(strings.Replace(string(readFile(t, p)), "writeEntry", "writeEntrz", 1)), 0o644)
		},
		"identity changed": func(dir string) {
			p := filepath.Join(dir, "identity.json")
			os.WriteFile(p, append(readFile(t, p), ' '), 0o644)
		},
		"dataset missing": func(dir string) { os.Remove(filepath.Join(dir, "records.jsonl")) },
		"dataset symlink": func(dir string) {
			p := filepath.Join(dir, "records.jsonl")
			real := filepath.Join(t.TempDir(), "records.jsonl")
			os.WriteFile(real, readFile(t, p), 0o644)
			os.Remove(p)
			os.Symlink(real, p)
		},
		"manifest revision": func(dir string) {
			p := filepath.Join(dir, "manifest.json")
			os.WriteFile(p, []byte(strings.Replace(string(readFile(t, p)), "1.0.0-draft.1", "1.0.0-draft.2", 1)), 0o644)
		},
	} {
		dir := copyDir(t, src)
		mutate(dir)
		if _, err := Load(dir); err == nil {
			t.Errorf("%s: loaded", name)
		}
	}
	if _, err := safeMember(src, "../bundle/records.jsonl"); err == nil {
		t.Error("traversal accepted")
	}
	if _, err := safeMember(src, "/etc/passwd"); err == nil {
		t.Error("absolute path accepted")
	}
}

// Publish writes a bundle Load verifies; it refuses an existing
// destination, and of two concurrent publishers to one name at most one
// succeeds while the published bundle stays valid.
func TestPublishAndReload(t *testing.T) {
	b := loadFixture(t)
	dir := filepath.Join(t.TempDir(), "out")
	ident := b.Identity
	ident.Provenance.Kind = "analysis"
	id, err := Publish(dir, ident, b.Records, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	again, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if again.AnalysisID != id || len(again.Nodes) != len(b.Nodes) || len(again.Edges) != len(b.Edges) {
		t.Fatalf("reloaded %s", again)
	}
	// Identical content gives the identical analysis id.
	id2, err := Publish(filepath.Join(t.TempDir(), "out2"), ident, b.Records, time.Now())
	if err != nil || id2 != id {
		t.Fatalf("second publication %s %v, want %s", id2, err, id)
	}
	if _, err := Publish(dir, ident, b.Records, time.Now()); err == nil {
		t.Fatal("published over an existing bundle")
	}
	race := filepath.Join(t.TempDir(), "race")
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Publish(race, ident, b.Records, time.Now()); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if ok != 1 {
		t.Fatalf("%d concurrent publications succeeded, want exactly 1", ok)
	}
	if _, err := Load(race); err != nil {
		t.Fatal(err)
	}
	left, _ := filepath.Glob(filepath.Join(filepath.Dir(race), ".race.partial-*"))
	if len(left) != 0 {
		t.Fatalf("partial directories left: %v", left)
	}
}

func TestDescriptor(t *testing.T) {
	for d, ok := range map[string]bool{
		"()V": true, "(I[JLjava/lang/String;)[[Ljava/lang/Object;": true, "(Ljava/nio/file/Path;[B)V": true,
		"()": false, "(V)V": false, "(L;)V": false, "(Ljava.lang.String;)V": false, "(I)VV": false, "I": false, "(Lx)V": false,
	} {
		if ValidMethodDescriptor(d) != ok {
			t.Errorf("ValidMethodDescriptor(%q) != %v", d, ok)
		}
	}
}

// Records are rejected for a duplicate, unknown or missing member at any
// level, invalid UTF-8 or a mismatched record_type.
func TestRecordStrictness(t *testing.T) {
	good := `{"basis":"direct","character":"jvm_rule","client_dispatch":false,"edge_id":"e-1","from":"n-a","kind":"static","record_type":"edge","rule":"r","sites":[{"caller_node_id":"n-a","condition":null,"kind":"static","pc":4}],"supplied_dispatch":false,"to":"n-b"}`
	if _, err := decodeRecord([]byte(good)); err != nil {
		t.Fatalf("good record rejected: %v", err)
	}
	for name, bad := range map[string]string{
		"duplicate":      strings.Replace(good, `"kind":"static","record_type"`, `"kind":"static","kind":"static","record_type"`, 1),
		"unknown":        strings.Replace(good, `"to":"n-b"}`, `"to":"n-b","extra":1}`, 1),
		"missing":        strings.Replace(good, `"rule":"r",`, ``, 1),
		"nested unknown": strings.Replace(good, `"pc":4}`, `"pc":4,"x":1}`, 1),
		"nested missing": strings.Replace(good, `"condition":null,`, ``, 1),
		"case":           strings.Replace(good, `"edge_id"`, `"Edge_id"`, 1),
		"utf8":           strings.Replace(good, `"rule":"r"`, "\"rule\":\"\xff\"", 1),
		"type":           strings.Replace(good, `"record_type":"edge"`, `"record_type":"node"`, 1),
	} {
		if _, err := decodeRecord([]byte(bad)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
