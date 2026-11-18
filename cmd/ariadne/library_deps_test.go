package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const libraryDepsFixture = "../../testdata/library-deps/"

// TestLibraryDeps (LIBRARY_MODE LB3, lb3-1): -roots api over the control
// library testdata/library-deps (EXPECTED.MD). Its pom's declarations are
// read, the consumer tree is built from the poms in the fixture's cache
// (offline), and the dependencies it says the library supplies are
// exactly what real Maven gives a consumer (maven-consumer.txt, from
// scripts/library-deps-oracle.sh). Verdicts are unchanged; findings and
// components carry who supplies the instance.
func TestLibraryDeps(t *testing.T) {
	oracle := mavenConsumer(t)
	for _, mode := range []string{"cha", "rta"} {
		doc := runLibraryDeps(t, "api", mode)
		type row struct{ pkg, by, scope, via, version, version_match string }
		var rows []row
		var supplied []string
		for _, d := range doc.Coverage.API.Dependencies.Declared {
			scope := d.Scope
			if d.Optional {
				scope += "+optional"
			}
			rows = append(rows, row{d.Package, d.SuppliedBy, scope, strings.Join(d.Via, ">"), d.Version, d.Match})
			if d.SuppliedBy == "library" {
				supplied = append(supplied, d.Package+":"+d.Version+":"+d.Scope)
			}
		}
		want := []row{
			{"commons-io:commons-io", "library", "runtime", "", "2.6", "absent"},
			{"org.apache.commons:commons-lang3", "library", "compile", "", "3.11", "declared"},
			{"org.apache.commons:commons-text", "user", "provided", "", "[1.6,1.10.0)", "in_range"},
			{"org.springframework:spring-core", "user", "compile+optional", "", "5.3.0", "absent"},
			{"org.springframework:spring-jcl", "user", "compile+optional", "org.springframework:spring-core", "5.3.0", "absent"},
		}
		if !reflect.DeepEqual(rows, want) {
			t.Errorf("%s: declared\n got %v\nwant %v", mode, rows, want)
		}
		sort.Strings(supplied)
		if !reflect.DeepEqual(supplied, oracle) {
			t.Errorf("%s: supplied by the library %v, Maven gives a consumer %v", mode, supplied, oracle)
		}
		if d := doc.Coverage.API.Dependencies; d.POM != "lib.jar!/META-INF/maven/fx/library-deps/pom.xml" || d.Library != "fx:library-deps:1.0" {
			t.Errorf("%s: pom %q library %q", mode, d.POM, d.Library)
		}
		by := map[string]string{}
		for _, f := range doc.Findings {
			if f.Reachability != "REACHABLE" || len(f.ExposedAPI) != 1 {
				t.Errorf("%s: %s %s with %d exposed methods; want REACHABLE through Render#render", mode, f.AdvisoryID, f.Reachability, len(f.ExposedAPI))
			}
			by[f.AdvisoryID] = f.SuppliedBy
		}
		if want := map[string]string{"CVE-2022-42889": "user", "CVE-2025-48924": "library"}; !reflect.DeepEqual(by, want) {
			t.Errorf("%s: supplied_by %v, want %v", mode, by, want)
		}
		declared := map[string]string{}
		for _, c := range doc.Components {
			if c.Declaration != nil {
				declared[c.GAV] = c.Declaration.SuppliedBy
			}
		}
		if want := map[string]string{"org.apache.commons:commons-text:1.9": "user", "org.apache.commons:commons-lang3:3.11": "library"}; !reflect.DeepEqual(declared, want) {
			t.Errorf("%s: components' declarations %v, want %v", mode, declared, want)
		}
		md, err := os.ReadFile(filepath.Join(doc.dir, "r.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(md), "commons-io:commons-io 2.6 reaches users through the library (direct, runtime) but is not in the input") {
			t.Errorf("%s: no warning for commons-io", mode)
		}
		if !strings.Contains(string(md), "| `org.apache.commons:commons-text` | user | provided | direct | `[1.6,1.10.0)` (`${text.range}`) | `1.9` | in_range |") {
			t.Errorf("%s: report lacks the commons-text row", mode)
		}
	}
	// Other root sets read nothing.
	doc := runLibraryDeps(t, "all", "cha")
	if doc.Coverage.API.Dependencies != nil {
		t.Errorf("-roots all: dependencies %+v", doc.Coverage.API.Dependencies)
	}
	for _, f := range doc.Findings {
		if f.SuppliedBy != "" {
			t.Errorf("-roots all: %s supplied_by %q", f.AdvisoryID, f.SuppliedBy)
		}
	}
	if code, stderr := runWithEnv([]string{"-app", libraryDepsFixture + "lib.jar", "-out", filepath.Join(t.TempDir(), "r.md"), "-json", filepath.Join(t.TempDir(), "r.json"), "-library-pom", "x.xml"}, nil); code == 0 || !strings.Contains(stderr, "-library-pom needs -roots api") {
		t.Errorf("-library-pom without -roots api: exit %d, %s", code, stderr)
	}
}

type libraryDepsDoc struct {
	dir      string
	Findings []struct {
		AdvisoryID   string            `json:"advisory_id"`
		Reachability string            `json:"reachability"`
		SuppliedBy   string            `json:"supplied_by"`
		ExposedAPI   []json.RawMessage `json:"exposed_api"`
	} `json:"findings"`
	Components []struct {
		GAV         string `json:"gav"`
		Declaration *struct {
			SuppliedBy string `json:"supplied_by"`
		} `json:"declaration"`
	} `json:"components"`
	Coverage struct {
		API struct {
			Dependencies *struct {
				POM      string `json:"pom"`
				Library  string `json:"library"`
				Declared []struct {
					Package    string   `json:"package"`
					SuppliedBy string   `json:"supplied_by"`
					Scope      string   `json:"scope"`
					Optional   bool     `json:"optional"`
					Via        []string `json:"via"`
					Version    string   `json:"version"`
					Match      string   `json:"version_match"`
				} `json:"declared"`
			} `json:"dependencies"`
		} `json:"api"`
	} `json:"coverage"`
}

func runLibraryDeps(t *testing.T, roots, mode string) libraryDepsDoc {
	t.Helper()
	dir := t.TempDir()
	out := filepath.Join(dir, "r.json")
	cache := filepath.Join(dir, "cache")
	copyDir(t, libraryFixture+"cache", cache)
	copyDir(t, libraryDepsFixture+"cache", cache)
	args := []string{"-offline", "-cache-dir", cache, "-app", libraryDepsFixture + "lib.jar",
		"-deps", libraryFixture + "deps", "-profile", "generic", "-roots", roots, "-callgraph-mode", mode,
		"-json-format", "v2", "-out", filepath.Join(dir, "r.md"), "-json", out}
	if code, stderr := runWithEnv(args, nil); code != 0 {
		t.Fatalf("%s %s: exit %d: %s", roots, mode, code, stderr)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	doc := libraryDepsDoc{dir: dir}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

// mavenConsumer reads maven-consumer.txt: group:artifact:version:scope.
func mavenConsumer(t *testing.T) []string {
	t.Helper()
	f, err := os.Open(libraryDepsFixture + "maven-consumer.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if l := strings.TrimSpace(sc.Text()); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	sort.Strings(out)
	if len(out) == 0 {
		t.Fatal("maven-consumer.txt is empty; run scripts/library-deps-oracle.sh")
	}
	return out
}
