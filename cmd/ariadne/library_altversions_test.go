package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const altVersionsFixture = "../../testdata/library-altversions/"

type altVersionsDoc struct {
	Findings []json.RawMessage `json:"findings"`
	Coverage struct {
		API struct {
			Alt *struct {
				Strategy string `json:"strategy"`
				Mode     string `json:"mode"`
				Rules    string `json:"rules_version"`
				Deps     []struct {
					Package  string `json:"package"`
					Note     string `json:"note"`
					Versions []struct {
						Version     string   `json:"version"`
						ChosenBy    []string `json:"chosen_by"`
						Basis       string   `json:"basis"`
						NotAssessed string   `json:"not_assessed"`
						Findings    []struct {
							Advisory string `json:"advisory"`
							Status   string `json:"status"`
						} `json:"findings"`
					} `json:"versions"`
				} `json:"dependencies"`
			} `json:"alternate_versions"`
		} `json:"api"`
	} `json:"coverage"`
}

// runAltVersions runs the lb3-2 control case; withOSV false leaves out the
// OSV answer for the chosen version.
func runAltVersions(t *testing.T, format string, withOSV bool, extra ...string) (string, int, string) {
	t.Helper()
	dir := t.TempDir()
	cache := filepath.Join(dir, "cache")
	copyDir(t, libraryFixture+"cache", cache)
	copyDir(t, libraryDepsFixture+"cache", cache)
	copyDir(t, altVersionsFixture+"cache", cache)
	if !withOSV {
		for _, f := range []string{"0a5922ea92cabd2aafefd133a434c7b63c17fd3fe36b97ab9f185b9ce84a6c2d.json", "0a5922ea92cabd2aafefd133a434c7b63c17fd3fe36b97ab9f185b9ce84a6c2d.meta.json"} {
			os.Remove(filepath.Join(cache, f))
		}
	}
	out := filepath.Join(dir, "r.json")
	args := append([]string{"-offline", "-cache-dir", cache, "-app", libraryDepsFixture + "lib.jar", "-deps", libraryFixture + "deps",
		"-library-pom", altVersionsFixture + "pom.xml", "-profile", "generic", "-json-format", format,
		"-out", filepath.Join(dir, "r.md"), "-json", out}, extra...)
	code, msg := runWithEnv(args, nil)
	return out, code, msg
}

// TestLibraryAltVersions (LIBRARY_MODE LB3 lb3-2): testdata/library-altversions/EXPECTED.MD.
func TestLibraryAltVersions(t *testing.T) {
	read := func(path string) altVersionsDoc {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var d altVersionsDoc
		if err := json.Unmarshal(b, &d); err != nil {
			t.Fatal(err)
		}
		return d
	}
	base, code, msg := runAltVersions(t, "v2", true, "-roots", "api")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, msg)
	}
	alt, code, msg := runAltVersions(t, "v2", true, "-roots", "api", "-alt-versions", "newest")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, msg)
	}
	db, da := read(base), read(alt)
	if db.Coverage.API.Alt != nil {
		t.Error("alternate_versions without -alt-versions")
	}
	fb, _ := json.Marshal(db.Findings)
	fa, _ := json.Marshal(da.Findings)
	if string(fb) != string(fa) {
		t.Error("the run's own findings changed with -alt-versions")
	}
	a := da.Coverage.API.Alt
	if a == nil || a.Strategy != "newest" || a.Mode != "approx" || a.Rules != "lb3-2" || len(a.Deps) != 2 {
		t.Fatalf("alternate_versions = %+v", a)
	}
	lang, text := a.Deps[0], a.Deps[1]
	if lang.Package != "org.apache.commons:commons-lang3" || len(lang.Versions) != 1 {
		t.Fatalf("commons-lang3 set: %+v", lang)
	}
	v := lang.Versions[0]
	if v.Version != "3.17.0" || strings.Join(v.ChosenBy, ",") != "range_max" || !strings.Contains(v.Basis, "code of the analyzed 3.11") {
		t.Errorf("commons-lang3 version: %+v", v)
	}
	if len(v.Findings) != 1 || v.Findings[0].Advisory != "CVE-2025-48924" || v.Findings[0].Status != "REACHABLE" {
		t.Errorf("commons-lang3 3.17.0 findings: %+v", v.Findings)
	}
	if text.Package != "org.apache.commons:commons-text" || len(text.Versions) != 0 || !strings.Contains(text.Note, "published versions unavailable") {
		t.Errorf("commons-text set: %+v", text)
	}

	// v1 is byte-identical with and without the flag.
	v1a, c1, m1 := runAltVersions(t, "v1", true, "-roots", "api")
	v1b, c2, m2 := runAltVersions(t, "v1", true, "-roots", "api", "-alt-versions", "newest")
	if c1 != 0 || c2 != 0 {
		t.Fatalf("v1 runs: %d %s / %d %s", c1, m1, c2, m2)
	}
	b1, _ := os.ReadFile(v1a)
	b2, _ := os.ReadFile(v1b)
	if string(b1) != string(b2) {
		t.Error("v1 differs with -alt-versions")
	}

	// Without the OSV answer: chosen, not assessed, the run succeeds.
	miss, code, msg := runAltVersions(t, "v2", false, "-roots", "api", "-alt-versions", "newest")
	if code != 0 {
		t.Fatalf("without the OSV answer: exit %d: %s", code, msg)
	}
	m := read(miss).Coverage.API.Alt
	if m == nil || len(m.Deps[0].Versions) != 1 || !strings.Contains(m.Deps[0].Versions[0].NotAssessed, "advisories unavailable") || len(m.Deps[0].Versions[0].Findings) != 0 {
		t.Errorf("without the OSV answer: %+v", m)
	}
}

func TestLibraryAltVersionsFlags(t *testing.T) {
	for name, args := range map[string][]string{
		"no -roots api":   {"-roots", "all", "-alt-versions", "newest"},
		"unknown":         {"-roots", "api", "-alt-versions", "every"},
		"rerun not built": {"-roots", "api", "-alt-versions", "newest", "-alt-versions-mode", "rerun"},
	} {
		_, code, msg := runAltVersions(t, "v2", true, args...)
		if code == 0 {
			t.Errorf("%s: exit 0, want a rejection", name)
		}
		if name == "rerun not built" && !strings.Contains(msg, "not built yet") {
			t.Errorf("%s: %s", name, msg)
		}
	}
}
