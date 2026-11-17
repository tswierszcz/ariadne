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
func libraryExposed(t *testing.T, extra ...string) (map[string][]string, map[string]map[string]string, map[string]any, libraryClient) {
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
				Method         string   `json:"method"`
				Pass           string   `json:"pass"`
				Path           []string `json:"path"`
				ClientDispatch bool     `json:"client_dispatch"`
				InheritedBy    []string `json:"inherited_by"`
			} `json:"exposed_api"`
		} `json:"findings"`
		Coverage struct {
			API  map[string]any `json:"api"`
			Gaps []struct {
				Mechanism string `json:"mechanism"`
				Count     int    `json:"count"`
				Reachable *int   `json:"reachable"`
			} `json:"gaps"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	names, keys := map[string][]string{}, map[string]map[string]string{}
	client := libraryClient{dispatch: map[string][]string{}, inherited: map[string]map[string][]string{}}
	for _, g := range doc.Coverage.Gaps {
		if g.Mechanism == "client_code" {
			client.sites = g.Count
			if g.Reachable != nil {
				client.reachable = *g.Reachable
			}
		}
	}
	for _, f := range doc.Findings {
		keys[f.AdvisoryID] = map[string]string{}
		client.inherited[f.AdvisoryID] = map[string][]string{}
		for _, x := range f.ExposedAPI {
			if len(x.InheritedBy) > 0 {
				// lb2-3: a dependency method API classes inherit; checked
				// on its own (TestLibraryMode, cases g and h).
				client.inherited[f.AdvisoryID][x.Method] = x.InheritedBy
				continue
			}
			n := strings.TrimPrefix(x.Method[:strings.IndexByte(x.Method, '(')], "lib.")
			names[f.AdvisoryID] = append(names[f.AdvisoryID], n)
			if x.ClientDispatch {
				client.dispatch[f.AdvisoryID] = append(client.dispatch[f.AdvisoryID], n)
			}
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
	return names, keys, doc.Coverage.API, client
}

// libraryClient is what a library-mode run says about client code
// (LIBRARY_MODE LB2): the client_code gap's sites and reachable sites,
// and per advisory the exposed methods marked client_dispatch.
type libraryClient struct {
	sites, reachable int
	dispatch         map[string][]string
	// inherited: per advisory, exposed dependency methods → the API
	// classes inheriting them (lb2-3).
	inherited map[string]map[string][]string
}

// TestLibraryMode (LIBRARY_MODE LB1, LB2): -roots api on the control
// library gives exactly the exposed API testdata/library/EXPECTED.MD wrote
// down before the implementation — CHA, RTA and -api-exclude — and agrees
// with the real JVM (observed.txt): under RTA every exposed method is one
// the JVM ran, and the code no client can call is never exposed. LB2: the
// three client-code sites are a client_code gap, and under CHA only
// Pipeline#run (case d, the one exposed method the JVM does not confirm)
// is marked client_dispatch. lb2-3: the dependency methods Expander (g)
// inherits are exposed with inherited_by, and SafeExpander (h), which
// overrides replace(String), does not inherit that one.
func TestLibraryMode(t *testing.T) {
	never := []string{"api.Engine#expand", "api.Formatter#unsafe", "api.Hidden#run", "api.ExpandingListener#onEvent", "api.Formatter#hook", "api.SafeExpander#replace"}
	for _, c := range []struct {
		name     string
		extra    []string
		want     []string
		dispatch []string
	}{
		{"cha", nil, []string{"api.Codecs#decodeAll", "api.Pipeline#run", "api.Templates#render", "internal.ExpandCodec#decode", "internal.Util#expand"}, []string{"api.Pipeline#run"}},
		{"rta", []string{"-callgraph-mode", "rta"}, []string{"api.Codecs#decodeAll", "api.Templates#render", "internal.ExpandCodec#decode", "internal.Util#expand"}, nil},
		{"exclude", []string{"-api-exclude", "lib.internal"}, []string{"api.Codecs#decodeAll", "api.Pipeline#run", "api.Templates#render"}, []string{"api.Pipeline#run"}},
	} {
		names, keys, cov, client := libraryExposed(t, c.extra...)
		if cov["inherited_methods"] == nil || cov["inherited_methods"].(float64) == 0 {
			t.Errorf("%s: coverage.api.inherited_methods %v", c.name, cov["inherited_methods"])
		}
		if client.sites != 3 || client.reachable != 3 {
			t.Errorf("%s: client_code gap %d sites, %d reachable; want 3, 3", c.name, client.sites, client.reachable)
		}
		for _, cve := range []string{"CVE-2022-42889", "CVE-2025-48924"} {
			// lb2-3: (g) Expander inherits every replace overload; (h)
			// SafeExpander overrides replace(String), so it does not
			// inherit that one. lb2-4: (i) StrictExpander inherits them
			// all, and they are mapped code of CVE-2022-42889 themselves;
			// but its own substitute, which they call on this, never gets
			// to the lookups behind CVE-2025-48924 (observed-strict.txt) —
			// only the inherited resolveVariable does.
			inh := client.inherited[cve]
			strict := []string{"lib.api.StrictExpander"}
			if cve == "CVE-2025-48924" {
				strict = nil
			}
			replace := "org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;"
			if got, want := inh[replace], append([]string{"lib.api.Expander"}, strict...); !reflect.DeepEqual(got, want) {
				t.Errorf("%s %s: %s inherited by %v, want %v", c.name, cve, replace, got, want)
			}
			if got, want := inh["org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/Object;)Ljava/lang/String;"], append([]string{"lib.api.Expander", "lib.api.SafeExpander"}, strict...); !reflect.DeepEqual(got, want) {
				t.Errorf("%s %s: replace(Object) inherited by %v, want %v", c.name, cve, got, want)
			}
			if cve == "CVE-2025-48924" {
				resolve := "org.apache.commons.text.StringSubstitutor#resolveVariable(Ljava/lang/String;Lorg/apache/commons/text/TextStringBuilder;II)Ljava/lang/String;"
				if got := inh[resolve]; !reflect.DeepEqual(got, []string{"lib.api.Expander", "lib.api.SafeExpander", "lib.api.StrictExpander"}) {
					t.Errorf("%s %s: resolveVariable inherited by %v, want all three expanders", c.name, cve, got)
				}
			}
			if got := names[cve]; !reflect.DeepEqual(got, c.want) {
				t.Errorf("%s %s: exposed %v, want %v", c.name, cve, got, c.want)
			}
			if got := client.dispatch[cve]; !reflect.DeepEqual(got, c.dispatch) {
				t.Errorf("%s %s: client_dispatch %v, want %v", c.name, cve, got, c.dispatch)
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
			if !obs["org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;"] || !obs["lib.api.Expander#<init>()V"] || !obs["lib.api.SafeExpander#replace(Ljava/lang/String;)Ljava/lang/String;"] {
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
