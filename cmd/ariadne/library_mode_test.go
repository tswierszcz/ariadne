package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const libraryFixture = "../../testdata/library/"

// libraryExposed runs -roots api over testdata/library and returns, per
// advisory, the exposed API method names (owner#name, "lib." dropped)
// and their passes, plus coverage.api.
func libraryExposed(t *testing.T, extra ...string) (map[string][]string, map[string]map[string]string, map[string]any) {
	t.Helper()
	dir := t.TempDir()
	out := filepath.Join(dir, "r.json")
	// The OSV and fix-commit responses this input needs, copied from a
	// real run's cache (testdata/library/cache), so the test runs offline
	// on a fresh checkout.
	cache := filepath.Join(dir, "cache")
	copyDir(t, libraryFixture+"cache", cache)
	args := append([]string{"-offline", "-cache-dir", cache, "-app", libraryFixture + "lib.jar",
		"-deps", libraryFixture + "deps", "-profile", "generic", "-roots", "api", "-json-format", "v2",
		"-out", filepath.Join(dir, "r.md"), "-json", out}, extra...)
	if code, stderr := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Findings []struct {
			AdvisoryID   string `json:"advisory_id"`
			Reachability string `json:"reachability"`
			ExposedAPI   []struct {
				Method string   `json:"method"`
				Pass   string   `json:"pass"`
				Path   []string `json:"path"`
			} `json:"exposed_api"`
		} `json:"findings"`
		Coverage struct {
			API map[string]any `json:"api"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	names, keys := map[string][]string{}, map[string]map[string]string{}
	for _, f := range doc.Findings {
		keys[f.AdvisoryID] = map[string]string{}
		for _, x := range f.ExposedAPI {
			n := strings.TrimPrefix(x.Method[:strings.IndexByte(x.Method, '(')], "lib.")
			names[f.AdvisoryID] = append(names[f.AdvisoryID], n)
			keys[f.AdvisoryID][x.Method] = x.Pass
			if len(x.Path) == 0 || x.Path[0] != x.Method {
				t.Errorf("%s %s: example path %v does not start at the method", f.AdvisoryID, x.Method, x.Path)
			}
		}
		sort.Strings(names[f.AdvisoryID])
		if f.Reachability != "REACHABLE" {
			t.Errorf("%s: %s, want REACHABLE", f.AdvisoryID, f.Reachability)
		}
	}
	return names, keys, doc.Coverage.API
}

// TestLibraryMode (LIBRARY_MODE LB1): -roots api on the control library
// gives exactly the exposed API testdata/library/EXPECTED.MD wrote down
// before the implementation — CHA, RTA and -api-exclude — and agrees
// with the real JVM (observed.txt): under RTA every exposed method is one
// the JVM ran, and the code no client can call is never exposed.
func TestLibraryMode(t *testing.T) {
	never := []string{"api.Engine#expand", "api.Formatter#unsafe", "api.Hidden#run", "api.ExpandingListener#onEvent", "api.Formatter#hook"}
	for _, c := range []struct {
		name  string
		extra []string
		want  []string
	}{
		{"cha", nil, []string{"api.Codecs#decodeAll", "api.Pipeline#run", "api.Templates#render", "internal.ExpandCodec#decode", "internal.Util#expand"}},
		{"rta", []string{"-callgraph-mode", "rta"}, []string{"api.Codecs#decodeAll", "api.Templates#render", "internal.ExpandCodec#decode", "internal.Util#expand"}},
		{"exclude", []string{"-api-exclude", "lib.internal"}, []string{"api.Codecs#decodeAll", "api.Pipeline#run", "api.Templates#render"}},
	} {
		names, keys, cov := libraryExposed(t, c.extra...)
		for _, cve := range []string{"CVE-2022-42889", "CVE-2025-48924"} {
			if got := names[cve]; !reflect.DeepEqual(got, c.want) {
				t.Errorf("%s %s: exposed %v, want %v", c.name, cve, got, c.want)
			}
			for _, n := range never {
				for _, g := range names[cve] {
					if g == n {
						t.Errorf("%s %s: %s exposed", c.name, cve, n)
					}
				}
			}
			for k, pass := range keys[cve] {
				if pass != "app_only" {
					t.Errorf("%s %s: %s pass %s", c.name, cve, k, pass)
				}
			}
		}
		if cov == nil || cov["jpms"] != "classpath" {
			t.Errorf("%s: coverage.api %v", c.name, cov)
		}
		if c.name == "rta" {
			obs := (benchFixture{observed: libraryFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
			for k := range keys["CVE-2022-42889"] {
				if !obs[k] {
					t.Errorf("rta: %s exposed, the JVM never ran it", k)
				}
			}
			if !obs["org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;"] {
				t.Error("observed.txt: replace never ran; regenerate with scripts/library-oracle.sh")
			}
		}
	}
}

func copyDir(t *testing.T, from, to string) {
	t.Helper()
	if err := os.MkdirAll(to, 0o755); err != nil {
		t.Fatal(err)
	}
	ents, err := os.ReadDir(from)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ents {
		b, err := os.ReadFile(filepath.Join(from, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(to, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
