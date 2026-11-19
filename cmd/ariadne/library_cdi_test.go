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

const cdiFixture = "../../testdata/library-cdi/"

// TestLibraryCDI (LIBRARY_MODE LB2, lb2-6): -roots api over the control
// CDI bean archive (testdata/library-cdi, EXPECTED.MD). Its vulnerable
// code is reached only from package-private producer, observer,
// @PostConstruct and @Inject methods a CDI container runs; those four are
// exposed, with their bean archive as provenance, and every one ran on
// Weld SE (observed.txt). Unmanaged, which is not a bean, never is.
func TestLibraryCDI(t *testing.T) {
	want := []string{
		"cd.internal.AuditObserver#onStartup(Lcd/api/Startup;)V",
		"cd.internal.CachedStore#warm()V",
		"cd.internal.Producers#banner()Lcd/api/Banner;",
		"cd.internal.Wiring#init(Lcd/api/Banner;)V",
	}
	obs := (benchFixture{observed: cdiFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	if !obs["org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;"] {
		t.Fatal("observed.txt: replace never ran; regenerate with scripts/library-cdi-oracle.sh")
	}
	for _, m := range want {
		if !obs[m] {
			t.Errorf("observed.txt: %s did not run on Weld", m)
		}
	}
	for m := range obs {
		if strings.HasPrefix(m, "cd.internal.Unmanaged#") {
			t.Errorf("observed.txt: %s ran, but Unmanaged is not a bean", m)
		}
	}
	for _, mode := range []string{"cha", "rta"} {
		dir := t.TempDir()
		out := filepath.Join(dir, "r.json")
		cache := filepath.Join(dir, "cache")
		copyDir(t, libraryFixture+"cache", cache)
		args := []string{"-offline", "-cache-dir", cache, "-app", cdiFixture + "lib.jar",
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
				if len(x.AnnouncedBy) != 1 || !strings.HasPrefix(x.AnnouncedBy[0], "META-INF/beans.xml bean-discovery-mode=annotated @jakarta.enterprise.context.") {
					t.Errorf("%s: %s announced_by %v", mode, x.Method, x.AnnouncedBy)
				}
				if !obs[x.Method] {
					t.Errorf("%s: %s exposed but never ran on Weld", mode, x.Method)
				}
			}
			sort.Strings(got)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s: %s exposed_api\n got %v\nwant %v", mode, f.AdvisoryID, got, want)
			}
		}
	}
}
