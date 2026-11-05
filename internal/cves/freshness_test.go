package cves

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"ariadne/internal/httpcache"
	"ariadne/internal/input"
)

var lib = []input.GAV{{Group: "org.x", Artifact: "lib", Version: "1.0"}}

// osvServer answers querybatch with ids() for org.x:lib:1.0 and serves
// records from recs (404 for any other id).
func osvServer(t *testing.T, ids func() []string, recs map[string]osvVuln) (*httptest.Server, *atomic.Int32) {
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/querybatch", func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var req osvQueryBatchRequest
		json.NewDecoder(r.Body).Decode(&req)
		resp := osvQueryBatchResponse{Results: make([]osvBatchResult, len(req.Queries))}
		for i := range req.Queries {
			for _, id := range ids() {
				resp.Results[i].Vulns = append(resp.Results[i].Vulns, osvVulnID{ID: id})
			}
		}
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/vulns/", func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		v, ok := recs[r.URL.Path[len("/v1/vulns/"):]]
		if !ok {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(v)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &calls
}

// ARIADNE-TEST-0402: the cached GHSA record is withdrawn, its CVE alias
// has an active record that the query did not return — the alias is
// fetched and the advisory stays active, with both sources kept.
func TestWithdrawnRecordWithActiveAlias(t *testing.T) {
	recs := map[string]osvVuln{
		"GHSA-old": {ID: "GHSA-old", Aliases: []string{"CVE-2099-1"}, Summary: "retracted text", Withdrawn: "2026-10-01T00:00:00Z", Modified: "2026-10-01T00:00:00Z"},
		"CVE-2099-1": {ID: "CVE-2099-1", Summary: "active", Published: "2026-01-01T00:00:00Z", Modified: "2026-09-01T00:00:00Z",
			Affected: []osvAffected{{Package: osvPackage{Ecosystem: "Maven", Name: "org.x:lib"}, Ranges: []osvRange{{Type: "ECOSYSTEM", Events: []map[string]string{{"introduced": "0"}, {"fixed": "1.1"}}}}}}},
	}
	srv, _ := osvServer(t, func() []string { return []string{"GHSA-old"} }, recs)
	cache, _ := httpcache.New(t.TempDir(), false)
	r, err := (&Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Cache: cache}).ResolveAll(context.Background(), lib)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.CVEs) != 1 || len(r.Withdrawn) != 0 {
		t.Fatalf("resolution = %+v", r)
	}
	c := r.CVEs[0]
	if c.ID != "CVE-2099-1" || c.Summary != "active" || len(c.Affected) != 1 || len(c.Sources) != 2 ||
		c.Sources[1].Withdrawn == "" || c.Modified != "2026-09-01T00:00:00Z" || c.Published != "2026-01-01T00:00:00Z" ||
		len(c.Sources[0].Ranges) != 1 || c.Sources[0].Ranges[0].Events[1]["fixed"] != "1.1" {
		t.Errorf("cve = %+v", c)
	}
}

func TestAllSourcesWithdrawn(t *testing.T) {
	recs := map[string]osvVuln{"GHSA-gone": {ID: "GHSA-gone", Aliases: []string{"CVE-2099-2"}, Withdrawn: "2026-10-01T00:00:00Z"}}
	srv, _ := osvServer(t, func() []string { return []string{"GHSA-gone"} }, recs)
	cache, _ := httpcache.New(t.TempDir(), false)
	r, err := (&Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Cache: cache}).ResolveAll(context.Background(), lib)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.CVEs) != 0 || len(r.Withdrawn) != 1 || r.Withdrawn[0].ID != "CVE-2099-2" || len(r.Withdrawn[0].Affected) != 1 {
		t.Errorf("resolution = %+v", r)
	}
}

// ARIADNE-TEST-0401 through the real client: an empty cached querybatch
// answer is refreshed after the TTL and the new advisory appears.
func TestNewAdvisoryAfterEmptyAnswerExpires(t *testing.T) {
	var ids atomic.Value
	ids.Store([]string{})
	recs := map[string]osvVuln{"GHSA-new": {ID: "GHSA-new", Aliases: []string{"CVE-2099-3"}}}
	srv, _ := osvServer(t, func() []string { return ids.Load().([]string) }, recs)
	dir := t.TempDir()
	now := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	resolve := func() []CVE {
		cache, _ := httpcache.New(dir, false)
		cache.SetClock(func() time.Time { return now })
		got, err := (&Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Cache: cache}).Resolve(context.Background(), lib)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	if got := resolve(); len(got) != 0 {
		t.Fatalf("initially %v", got)
	}
	ids.Store([]string{"GHSA-new"})
	now = now.Add(time.Hour)
	if got := resolve(); len(got) != 0 {
		t.Fatalf("within TTL the cached empty answer must be used, got %v", got)
	}
	now = now.Add(httpcache.DefaultTTL)
	if got := resolve(); len(got) != 1 || got[0].ID != "CVE-2099-3" {
		t.Fatalf("after TTL: %v", got)
	}
}
