package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const micronautFixture = "../../testdata/library-micronaut/"

// TestLibraryMicronaut_measuredGap (LIBRARY_MODE LB2, Micronaut
// measurement 2026-11-20): Micronaut's generated bean definitions are
// package-private or synthetic, so -roots api does not root them, and the
// package-private bean methods they call — which ran on the JVM
// (observed.txt) — leave the advisories NOT_REACHABLE. This pins the
// measured false negative (EXPECTED.MD); a Micronaut announcement model
// is expected to turn it into REACHABLE with four exposed methods.
func TestLibraryMicronaut_measuredGap(t *testing.T) {
	obs := (benchFixture{observed: micronautFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	for _, m := range []string{
		"org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;",
		"mn.internal.AuditListener#onStartup(Lmn/api/Startup;)V",
		"mn.internal.CachedStore#warm()V",
		"mn.internal.Producers#banner()Lmn/api/Banner;",
		"mn.internal.Wiring#init(Lmn/api/Banner;)V",
	} {
		if !obs[m] {
			t.Errorf("observed.txt: %s did not run; regenerate with scripts/library-micronaut-oracle.sh", m)
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
				AdvisoryID   string            `json:"advisory_id"`
				Reachability string            `json:"reachability"`
				ExposedAPI   []json.RawMessage `json:"exposed_api"`
			} `json:"findings"`
		}
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		if len(doc.Findings) != 2 {
			t.Fatalf("%s: %d findings, want 2", mode, len(doc.Findings))
		}
		for _, f := range doc.Findings {
			if f.Reachability != "NOT_REACHABLE" || len(f.ExposedAPI) != 0 {
				t.Errorf("%s: %s %s with %d exposed methods — if a Micronaut model landed, update this test and EXPECTED.MD", mode, f.AdvisoryID, f.Reachability, len(f.ExposedAPI))
			}
		}
	}
}
