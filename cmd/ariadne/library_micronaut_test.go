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

const micronautFixture = "../../testdata/library-micronaut/"

// TestLibraryMicronaut (LIBRARY_MODE LB2, lb2-9): -roots api over the
// control Micronaut library (testdata/library-micronaut, EXPECTED.MD).
// Its vulnerable code is reached only from package-private factory,
// @EventListener, @PostConstruct and @Inject methods that Micronaut's
// generated bean definitions call; those four are exposed, with the
// definition's listing entry as provenance, and every one ran on the JVM
// (observed.txt). Unmanaged, which has no definition, never is. Before
// lb2-9 both advisories were NOT_REACHABLE (measured 2026-11-20).
func TestLibraryMicronaut(t *testing.T) {
	want := []string{
		"mn.internal.AuditListener#onStartup(Lmn/api/Startup;)V",
		"mn.internal.CachedStore#warm()V",
		"mn.internal.Producers#banner()Lmn/api/Banner;",
		"mn.internal.Wiring#init(Lmn/api/Banner;)V",
	}
	obs := (benchFixture{observed: micronautFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	if !obs["org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;"] {
		t.Fatal("observed.txt: replace never ran; regenerate with scripts/library-micronaut-oracle.sh")
	}
	for _, m := range want {
		if !obs[m] {
			t.Errorf("observed.txt: %s did not run on Micronaut", m)
		}
	}
	for m := range obs {
		if strings.HasPrefix(m, "mn.internal.Unmanaged#") {
			t.Errorf("observed.txt: %s ran, but Unmanaged is not a bean", m)
		}
	}
	for _, mode := range []string{"cha", "rta"} {
		dir := t.TempDir()
		out := filepath.Join(dir, "r.json")
		cache := filepath.Join(dir, "cache")
		copyDir(t, libraryFixture+"cache", cache)
		args := []string{"-offline", "-cache-dir", cache, "-app", micronautFixture + "lib.jar",
			"-deps", libraryFixture + "deps", "-profile", "generic", "-roots", "api", "-callgraph-mode", mode,
			"-json-format", "v2", "-out", filepath.Join(dir, "r.md"), "-json", out}
		if code, stderr := runWithEnv(args, nil); code != 0 {
			t.Fatalf("%s: exit %d: %s", mode, code, stderr)
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
					Method      string   `json:"method"`
					AnnouncedBy []string `json:"announced_by"`
				} `json:"exposed_api"`
			} `json:"findings"`
			Coverage struct {
				API struct {
					Announced int `json:"announced_classes"`
				} `json:"api"`
			} `json:"coverage"`
		}
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		if doc.Coverage.API.Announced != 4 {
			t.Errorf("%s: %d announced classes, want 4", mode, doc.Coverage.API.Announced)
		}
		if len(doc.Findings) != 2 {
			t.Fatalf("%s: %d findings, want 2", mode, len(doc.Findings))
		}
		for _, f := range doc.Findings {
			if f.Reachability != "REACHABLE" {
				t.Errorf("%s: %s %s, want REACHABLE", mode, f.AdvisoryID, f.Reachability)
			}
			var got []string
			for _, x := range f.ExposedAPI {
				got = append(got, x.Method)
				if len(x.AnnouncedBy) != 1 || !strings.HasPrefix(x.AnnouncedBy[0], "META-INF/micronaut/io.micronaut.inject.BeanDefinitionReference/mn.internal.$") {
					t.Errorf("%s: %s announced_by %v", mode, x.Method, x.AnnouncedBy)
				}
				if !obs[x.Method] {
					t.Errorf("%s: %s exposed but never ran on Micronaut", mode, x.Method)
				}
			}
			sort.Strings(got)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s: %s exposed_api\n got %v\nwant %v", mode, f.AdvisoryID, got, want)
			}
		}
	}
}
