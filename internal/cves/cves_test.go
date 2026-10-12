package cves

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ariadne/internal/httpcache"
	"ariadne/internal/input"
)

// fakeOSV serves a tiny, fixed OSV dataset:
//   - org.yaml:snakeyaml:1.27 is affected by GHSA-3mc7-4q67-w48m, whose
//     alias is CVE-2022-1471 (real SnakeYAML CVE, used as a realistic
//     fixture).
//   - com.example:clean:1.0 has no vulnerabilities.
func fakeOSV(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/querybatch", func(w http.ResponseWriter, r *http.Request) {
		var req osvQueryBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		resp := osvQueryBatchResponse{Results: make([]osvBatchResult, len(req.Queries))}
		for i, q := range req.Queries {
			if q.Package.Name == "org.yaml:snakeyaml" && q.Version == "1.27" {
				resp.Results[i] = osvBatchResult{Vulns: []osvVulnID{{ID: "GHSA-3mc7-4q67-w48m"}}}
			}
		}
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/vulns/GHSA-3mc7-4q67-w48m", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(osvVuln{
			ID:       "GHSA-3mc7-4q67-w48m",
			Aliases:  []string{"CVE-2022-1471"},
			Summary:  "SnakeYAML Constructor Deserialization Remote Code Execution",
			Severity: []osvSeverity{{Type: "CVSS_V3", Score: "9.8"}},
			References: []osvReference{
				{Type: "FIX", URL: "https://github.com/snakeyaml/snakeyaml/commit/abc123"},
			},
		})
	})
	return httptest.NewServer(mux)
}

func TestResolve_dedupesAliasesAndPicksCVEAsPrimary(t *testing.T) {
	srv := fakeOSV(t)
	defer srv.Close()

	cache, err := httpcache.New(t.TempDir(), false)
	if err != nil {
		t.Fatalf("httpcache.New: %v", err)
	}
	client := &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Cache: cache}

	gavs := []input.GAV{
		{Group: "org.yaml", Artifact: "snakeyaml", Version: "1.27"},
		{Group: "com.example", Artifact: "clean", Version: "1.0"},
	}
	got, err := client.Resolve(context.Background(), gavs)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d CVEs, want 1: %+v", len(got), got)
	}
	cve := got[0]
	if cve.ID != "CVE-2022-1471" {
		t.Errorf("ID = %q, want CVE-2022-1471 (should prefer the CVE alias over the GHSA id)", cve.ID)
	}
	if len(cve.Aliases) != 1 || cve.Aliases[0] != "GHSA-3mc7-4q67-w48m" {
		t.Errorf("Aliases = %v, want [GHSA-3mc7-4q67-w48m]", cve.Aliases)
	}
	if len(cve.Affected) != 1 || cve.Affected[0] != (Affected{Package: "org.yaml:snakeyaml", Version: "1.27"}) {
		t.Errorf("Affected = %+v", cve.Affected)
	}
	if len(cve.References) != 1 || cve.References[0].Type != "FIX" {
		t.Errorf("References = %+v", cve.References)
	}
}

func TestResolve_noVulnsReturnsEmpty(t *testing.T) {
	srv := fakeOSV(t)
	defer srv.Close()

	cache, err := httpcache.New(t.TempDir(), false)
	if err != nil {
		t.Fatalf("httpcache.New: %v", err)
	}
	client := &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Cache: cache}

	got, err := client.Resolve(context.Background(), []input.GAV{
		{Group: "com.example", Artifact: "clean", Version: "1.0"},
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d CVEs, want 0: %+v", len(got), got)
	}
}

func TestResolve_offlineCacheHitNoNetwork(t *testing.T) {
	srv := fakeOSV(t)
	gavs := []input.GAV{{Group: "org.yaml", Artifact: "snakeyaml", Version: "1.27"}}

	dir := t.TempDir()
	onlineCache, err := httpcache.New(dir, false)
	if err != nil {
		t.Fatalf("httpcache.New: %v", err)
	}
	client := &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Cache: onlineCache}
	if _, err := client.Resolve(context.Background(), gavs); err != nil {
		t.Fatalf("priming Resolve: %v", err)
	}
	srv.Close() // prove the second run touches no network

	offlineCache, err := httpcache.New(dir, true)
	if err != nil {
		t.Fatalf("httpcache.New (offline): %v", err)
	}
	client.Cache = offlineCache
	got, err := client.Resolve(context.Background(), gavs)
	if err != nil {
		t.Fatalf("offline Resolve: %v", err)
	}
	if len(got) != 1 || got[0].ID != "CVE-2022-1471" {
		t.Fatalf("offline Resolve = %+v", got)
	}
}

func TestDedupe_mergesMultipleRecordsSharingAnAlias(t *testing.T) {
	// Simulates two OSV records (e.g. a GHSA and a separately-ingested
	// CVE record) that both claim the same alias, which should collapse
	// into a single CVE entry.
	vulns := []osvVuln{
		{ID: "GHSA-aaaa", Aliases: []string{"CVE-2099-0001"}, Summary: "first"},
		{ID: "CVE-2099-0001", Aliases: []string{"GHSA-aaaa"}, Summary: "second"},
	}
	affectedBy := map[string][]pkgVersion{
		"GHSA-aaaa":     {{Package: "a:b", Version: "1.0"}},
		"CVE-2099-0001": {{Package: "a:b", Version: "1.0"}},
	}
	got := dedupe(vulns, affectedBy)
	if len(got) != 1 {
		t.Fatalf("got %d CVEs, want 1: %+v", len(got), got)
	}
	if got[0].ID != "CVE-2099-0001" {
		t.Errorf("ID = %q, want CVE-2099-0001", got[0].ID)
	}
	if len(got[0].Aliases) != 1 || got[0].Aliases[0] != "GHSA-aaaa" {
		t.Errorf("Aliases = %v, want [GHSA-aaaa]", got[0].Aliases)
	}
}

func TestDedupe_carriesDetailsFromFirstNonEmptyRecord(t *testing.T) {
	vulns := []osvVuln{
		{ID: "GHSA-aaaa", Aliases: []string{"CVE-2099-0001"}, Summary: "s"},
		{ID: "CVE-2099-0001", Aliases: []string{"GHSA-aaaa"}, Details: "long advisory text"},
	}
	affectedBy := map[string][]pkgVersion{
		"GHSA-aaaa":     {{Package: "a:b", Version: "1.0"}},
		"CVE-2099-0001": {{Package: "a:b", Version: "1.0"}},
	}
	got := dedupe(vulns, affectedBy)
	if len(got) != 1 || got[0].Details != "long advisory text" {
		t.Fatalf("Details = %+v, want the non-empty one from the second record", got)
	}
}
