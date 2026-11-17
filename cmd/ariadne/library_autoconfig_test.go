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

const autoconfigFixture = "../../testdata/library-autoconfig/"

// TestLibraryAutoconfig (LIBRARY_MODE LB2, lb2-5): -roots api over the
// control starter (testdata/library-autoconfig, EXPECTED.MD). None of its
// vulnerable code is public API; the classes it announces in
// spring.factories and AutoConfiguration.imports — with their member
// classes and @Imports — are entries, each exposed method carrying its
// provenance. Every exposed method ran on Spring Boot 2.4.5
// (observed.txt) except ModernAutoConfiguration's, whose imports file
// Boot 2.4 does not read; the never-announced UnusedConfiguration is
// never exposed.
func TestLibraryAutoconfig(t *testing.T) {
	want := []string{
		"ac.autoconfigure.GreeterAutoConfiguration#banner",
		"ac.autoconfigure.GreeterAutoConfiguration$NestedConfiguration#nestedBanner",
		"ac.autoconfigure.ImportedConfiguration#init",
		"ac.autoconfigure.ModernAutoConfiguration#modernBanner",
		"ac.autoconfigure.StartupListener#onApplicationEvent",
	}
	via := map[string]string{
		"ac.autoconfigure.GreeterAutoConfiguration#banner":                           "META-INF/spring.factories org.springframework.boot.autoconfigure.EnableAutoConfiguration: ac.autoconfigure.GreeterAutoConfiguration",
		"ac.autoconfigure.GreeterAutoConfiguration$NestedConfiguration#nestedBanner": " > member ac.autoconfigure.GreeterAutoConfiguration$NestedConfiguration",
		"ac.autoconfigure.ImportedConfiguration#init":                                " > @Import ac.autoconfigure.ImportedConfiguration",
		"ac.autoconfigure.ModernAutoConfiguration#modernBanner":                      "AutoConfiguration.imports org.springframework.boot.autoconfigure.AutoConfiguration: ",
		"ac.autoconfigure.StartupListener#onApplicationEvent":                        "META-INF/spring.factories org.springframework.context.ApplicationListener: ",
	}
	obs := (benchFixture{observed: autoconfigFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	if !obs["org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;"] {
		t.Fatal("observed.txt: replace never ran; regenerate with scripts/library-autoconfig-oracle.sh")
	}
	for _, mode := range []string{"cha", "rta"} {
		dir := t.TempDir()
		out := filepath.Join(dir, "r.json")
		cache := filepath.Join(dir, "cache")
		copyDir(t, libraryFixture+"cache", cache)
		args := []string{"-offline", "-cache-dir", cache, "-app", autoconfigFixture + "lib.jar",
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
					Announced     int `json:"announced_classes"`
					Announcements []struct {
						File, Key, Class string
					} `json:"announcements"`
				} `json:"api"`
			} `json:"coverage"`
		}
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		if doc.Coverage.API.Announced != 5 || len(doc.Coverage.API.Announcements) != 3 {
			t.Errorf("%s: coverage.api %d announced classes, %d announcements; want 5, 3", mode, doc.Coverage.API.Announced, len(doc.Coverage.API.Announcements))
		}
		if len(doc.Findings) != 2 {
			t.Fatalf("%s: %d findings, want CVE-2022-42889 and CVE-2025-48924", mode, len(doc.Findings))
		}
		for _, f := range doc.Findings {
			if f.Reachability != "REACHABLE" {
				t.Errorf("%s %s: %s, want REACHABLE", mode, f.AdvisoryID, f.Reachability)
			}
			var got []string
			for _, x := range f.ExposedAPI {
				name := x.Method[:strings.IndexByte(x.Method, '(')]
				got = append(got, name)
				if len(x.AnnouncedBy) != 1 || !strings.Contains(x.AnnouncedBy[0], via[name]) {
					t.Errorf("%s %s: %s announced by %v, want one containing %q", mode, f.AdvisoryID, name, x.AnnouncedBy, via[name])
				}
				if !obs[x.Method] && name != "ac.autoconfigure.ModernAutoConfiguration#modernBanner" {
					t.Errorf("%s %s: %s exposed, the JVM never ran it", mode, f.AdvisoryID, x.Method)
				}
			}
			sort.Strings(got)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s %s: exposed\n got %v\nwant %v", mode, f.AdvisoryID, got, want)
			}
		}
	}
	for k := range obs {
		if strings.Contains(k, "UnusedConfiguration") || strings.Contains(k, "ModernAutoConfiguration") {
			t.Errorf("observed.txt: %s ran; the fixture expects it never to", k)
		}
	}
}
