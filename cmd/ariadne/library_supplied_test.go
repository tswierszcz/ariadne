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

const suppliedFixture = "../../testdata/library-supplied/"

// TestLibrarySupplied (LIBRARY_MODE LB2, lb2-7) follows
// testdata/library-supplied/EXPECTED.MD: verdicts and exposed methods are
// those of lb2-6; the API methods that reach ClassUtils#getClass only
// through a StringLookup the caller supplies — as a parameter (y) or
// through a field set only from one (z) — are supplied_dispatch, the
// library's own lookup (aa), a field also set to the library's default
// (bb), a value stored in a dependency object (cc) and the library calling
// (y) with its own lookup (dd) are not. With only
// (y) and (z) in the API the finding carries SUPPLIED_DISPATCH_ONLY.
func TestLibrarySupplied(t *testing.T) {
	type exposed struct {
		Method   string `json:"method"`
		Supplied bool   `json:"supplied_dispatch"`
	}
	run := func(t *testing.T, mode string, extra ...string) map[string]struct {
		status, reasons string
		exposed         []string
	} {
		dir := t.TempDir()
		out := filepath.Join(dir, "r.json")
		cache := filepath.Join(dir, "cache")
		copyDir(t, libraryFixture+"cache", cache)
		args := append([]string{"-offline", "-cache-dir", cache, "-app", suppliedFixture + "lib.jar",
			"-deps", libraryFixture + "deps", "-profile", "generic", "-roots", "api", "-callgraph-mode", mode,
			"-json-format", "v2", "-out", filepath.Join(dir, "r.md"), "-json", out}, extra...)
		if code, stderr := runWithEnv(args, nil); code != 0 {
			t.Fatalf("%s: exit %d: %s", mode, code, stderr)
		}
		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		var doc struct {
			Findings []struct {
				AdvisoryID   string    `json:"advisory_id"`
				Reachability string    `json:"reachability"`
				Reasons      []string  `json:"reason_codes"`
				ExposedAPI   []exposed `json:"exposed_api"`
			} `json:"findings"`
		}
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		got := map[string]struct {
			status, reasons string
			exposed         []string
		}{}
		for _, f := range doc.Findings {
			var ex []string
			for _, x := range f.ExposedAPI {
				name, _, _ := strings.Cut(x.Method, "(")
				if x.Supplied {
					name += " supplied"
				}
				ex = append(ex, name)
			}
			sort.Strings(ex)
			r := ""
			for _, c := range f.Reasons {
				r += c + " "
			}
			got[f.AdvisoryID] = struct {
				status, reasons string
				exposed         []string
			}{f.Reachability, r, ex}
		}
		return got
	}
	for _, mode := range []string{"cha", "rta"} {
		got := run(t, mode)
		text, lang := got["CVE-2022-42889"], got["CVE-2025-48924"]
		if text.status != "REACHABLE" || !reflect.DeepEqual(text.exposed, []string{"sp.api.Registrar#expand"}) {
			t.Errorf("%s: CVE-2022-42889 %+v", mode, text)
		}
		want := []string{"sp.api.Constants#constant", "sp.api.Mixed#get", "sp.api.Registrar#expand", "sp.api.Wrapper#wrap",
			"sp.supplied.Holder#get supplied", "sp.supplied.Resolver#resolve supplied"}
		if lang.status != "REACHABLE" || lang.reasons != "" || !reflect.DeepEqual(lang.exposed, want) {
			t.Errorf("%s: CVE-2025-48924 %+v\nwant exposed %v, no reasons", mode, lang, want)
		}
	}
	got := run(t, "cha", "-api-exclude", "sp.api")
	if s := got["CVE-2022-42889"].status; s != "NOT_REACHABLE" {
		t.Errorf("-api-exclude sp.api: CVE-2022-42889 %s, want NOT_REACHABLE", s)
	}
	lang := got["CVE-2025-48924"]
	if lang.status != "REACHABLE" || lang.reasons != "SUPPLIED_DISPATCH_ONLY " ||
		!reflect.DeepEqual(lang.exposed, []string{"sp.supplied.Holder#get supplied", "sp.supplied.Resolver#resolve supplied"}) {
		t.Errorf("-api-exclude sp.api: CVE-2025-48924 %+v, want REACHABLE, SUPPLIED_DISPATCH_ONLY, (y) and (z) supplied", lang)
	}
}
