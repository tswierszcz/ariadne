package vulnmap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

func TestLoadManual_embeddedDefault(t *testing.T) {
	m, err := LoadManual("")
	if err != nil {
		t.Fatalf("LoadManual: %v", err)
	}
	// The plan's own seed entries must survive.
	if got := m["CVE-2022-1471"]; len(got) == 0 {
		t.Errorf("CVE-2022-1471 (SnakeYAML) has no methods")
	}
	if got := m["CVE-2021-44228"]; len(got) == 0 {
		t.Errorf("CVE-2021-44228 (Log4Shell) has no methods")
	}
	// The XStream family (WebGoat control case) uses a YAML anchor/alias
	// to share one method list; confirm the alias actually resolved
	// rather than yaml.v3 silently giving every aliased CVE an empty list.
	a, b := m["CVE-2013-7285"], m["CVE-2021-39144"]
	if len(a) == 0 || len(b) == 0 {
		t.Fatalf("XStream anchor/alias entries are empty: CVE-2013-7285=%v CVE-2021-39144=%v", a, b)
	}
	if a[0] != b[0] {
		t.Errorf("aliased entries differ: %v vs %v", a, b)
	}
}

func TestResolve_manualTableWinsOverDiff(t *testing.T) {
	r := NewResolver(map[string][]string{"CVE-2022-1471": {"org.yaml.snakeyaml.Yaml#load"}}, mustCache(t))
	entry, ok, err := r.Resolve(context.Background(), cves.CVE{ID: "CVE-2022-1471"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !ok || entry.Source != SourceManual {
		t.Fatalf("Resolve = (%+v, %v), want source=manual", entry, ok)
	}
}

func TestResolve_manualTableMatchesByAlias(t *testing.T) {
	r := NewResolver(map[string][]string{"GHSA-xxxx": {"a.B#m"}}, mustCache(t))
	entry, ok, err := r.Resolve(context.Background(), cves.CVE{ID: "CVE-2099-0001", Aliases: []string{"GHSA-xxxx"}})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !ok || entry.Source != SourceManual || entry.Methods[0] != "a.B#m" {
		t.Fatalf("Resolve = (%+v, %v), want a manual hit via the alias", entry, ok)
	}
}

// fakeGitHub serves a fixed fix-commit diff and the post-fix file
// content, standing in for github.com and raw.githubusercontent.com.
func fakeGitHub(t *testing.T) *httptest.Server {
	t.Helper()
	const diff = `diff --git a/src/main/java/org/yaml/snakeyaml/Yaml.java b/src/main/java/org/yaml/snakeyaml/Yaml.java
index 1111111..2222222 100644
--- a/src/main/java/org/yaml/snakeyaml/Yaml.java
+++ b/src/main/java/org/yaml/snakeyaml/Yaml.java
@@ -5,3 +5,4 @@ public class Yaml {
     public Object load(String s) {
+        checkSafe(s);
         return doLoad(s);
     }
`
	const file = `package org.yaml.snakeyaml;

public class Yaml {
    public Object load(String s) {
        checkSafe(s);
        return doLoad(s);
    }
}
`
	mux := http.NewServeMux()
	mux.HandleFunc("/yaml/snakeyaml/commit/abc1234.diff", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(diff))
	})
	mux.HandleFunc("/yaml/snakeyaml/abc1234/src/main/java/org/yaml/snakeyaml/Yaml.java", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(file))
	})
	return httptest.NewServer(mux)
}

func TestResolve_fallsBackToFixCommitDiff(t *testing.T) {
	srv := fakeGitHub(t)
	defer srv.Close()

	r := NewResolver(nil, mustCache(t))
	r.HTTPClient = srv.Client()
	r.GithubBaseURL = srv.URL
	r.RawBaseURL = srv.URL

	cve := cves.CVE{
		ID: "CVE-2022-1471",
		References: []cves.Reference{
			{Type: "FIX", URL: "https://github.com/yaml/snakeyaml/commit/abc1234"},
		},
	}
	entry, ok, err := r.Resolve(context.Background(), cve)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !ok {
		t.Fatalf("Resolve did not find a diff-based mapping")
	}
	if entry.Source != SourceDiff {
		t.Errorf("Source = %q, want %q", entry.Source, SourceDiff)
	}
	want := []string{"org.yaml.snakeyaml.Yaml#load"}
	if len(entry.Methods) != 1 || entry.Methods[0] != want[0] {
		t.Errorf("Methods = %v, want %v", entry.Methods, want)
	}
}

func TestResolve_noFixReferenceIsUnknown(t *testing.T) {
	r := NewResolver(nil, mustCache(t))
	_, ok, err := r.Resolve(context.Background(), cves.CVE{ID: "CVE-2099-9999"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if ok {
		t.Errorf("Resolve unexpectedly found something for a CVE with no references at all")
	}
}

func TestResolveAll_onlyIncludesResolvedCVEs(t *testing.T) {
	srv := fakeGitHub(t)
	defer srv.Close()

	r := NewResolver(map[string][]string{"CVE-1": {"a.B#m"}}, mustCache(t))
	r.HTTPClient = srv.Client()
	r.GithubBaseURL = srv.URL
	r.RawBaseURL = srv.URL

	list := []cves.CVE{
		{ID: "CVE-1"}, // manual hit
		{ID: "CVE-2", References: []cves.Reference{{Type: "FIX", URL: "https://github.com/yaml/snakeyaml/commit/abc1234"}}}, // diff hit
		{ID: "CVE-3"}, // no source at all -> absent from the result
	}
	got, err := r.ResolveAll(context.Background(), list)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2: %+v", len(got), got)
	}
	if _, ok := got["CVE-3"]; ok {
		t.Errorf("CVE-3 should not have an entry")
	}
}

func mustCache(t *testing.T) *httpcache.Cache {
	t.Helper()
	c, err := httpcache.New(t.TempDir(), false)
	if err != nil {
		t.Fatalf("httpcache.New: %v", err)
	}
	return c
}
