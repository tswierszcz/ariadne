package runtimevuln

import (
	"strings"
	"testing"
	"time"

	"ariadne/internal/runtimeinfo"
)

// A synthetic catalog (not real advisories: ids and versions are made up
// to exercise every rule).
const testCatalog = `
schema: 1
version: "t1"
source: test
note: synthetic
advisories:
  - date: 2026-04-21
    url: https://openjdk.org/groups/vulnerability/advisories/2026-04-21
    last_affected: {"8": "8u461", "17": "17.0.16", "25": "25"}
    earlier_affected: false
    cves:
      - {id: CVE-2026-00001, component: xml/jaxp, cvss: 7.5, lines: [8, 17, 25]}
      - {id: CVE-2026-00002, component: core-libs, cvss: 3.7, lines: [25]}
  - date: 2026-07-21
    url: https://openjdk.org/groups/vulnerability/advisories/2026-07-21
    last_affected: {"17": "17.0.17", "25": "25.0.1"}
    earlier_affected: true
    cves:
      - {id: CVE-2026-00003, component: security-libs/java.security, cvss: 5.9, lines: [17, 25]}
`

func provider(t *testing.T) *OpenJDK {
	c, err := ParseCatalog([]byte(testCatalog))
	if err != nil {
		t.Fatal(err)
	}
	return &OpenJDK{C: c, Distributions: map[string]bool{"Temurin": true}}
}

func temurin(v string) runtimeinfo.Runtime {
	rel, _ := runtimeinfo.ParseVersion(v)
	return runtimeinfo.Runtime{Distribution: "Temurin", VersionString: v, Version: &rel}
}

func apps(a Assessment) map[string]string {
	out := map[string]string{}
	for _, f := range a.Findings {
		out[f.CVE] = f.Applicability
	}
	return out
}

var now = time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

func TestAssess(t *testing.T) {
	p := provider(t)
	cases := []struct {
		version string
		want    map[string]string
		notes   int
	}{
		// exactly the last affected release; the second advisory's
		// "and earlier" covers 17.0.16 too
		{"17.0.16+8", map[string]string{"CVE-2026-00001": Affected, "CVE-2026-00002": NotAffected, "CVE-2026-00003": Affected}, 0},
		// later than the first advisory's 17.0.16: fixed there
		{"17.0.17+10", map[string]string{"CVE-2026-00001": NotAffected, "CVE-2026-00002": NotAffected, "CVE-2026-00003": Affected}, 0},
		// earlier, without an "and earlier" statement: undetermined
		{"17.0.9+9", map[string]string{"CVE-2026-00001": Undetermined, "CVE-2026-00002": NotAffected, "CVE-2026-00003": Affected}, 0},
		{"25.0.2+10", map[string]string{"CVE-2026-00001": NotAffected, "CVE-2026-00002": NotAffected, "CVE-2026-00003": NotAffected}, 0},
		// JDK 8 (8uNNN): only the first advisory covers line 8
		{"1.8.0_461-b11", map[string]string{"CVE-2026-00001": Affected, "CVE-2026-00002": NotAffected}, 1},
		// line 21: no advisory covers it — no findings, a note, never "not affected"
		{"21.0.8+9", map[string]string{}, 1},
	}
	for _, c := range cases {
		a := p.Assess(temurin(c.version), now)
		if a.State != Performed {
			t.Fatalf("%s: state %s (%s)", c.version, a.State, a.Reason)
		}
		got := apps(a)
		if len(got) != len(c.want) {
			t.Errorf("%s: %v, want %v", c.version, got, c.want)
		}
		for cve, w := range c.want {
			if got[cve] != w {
				t.Errorf("%s %s = %s, want %s", c.version, cve, got[cve], w)
			}
		}
		if len(a.Notes) != c.notes {
			t.Errorf("%s notes = %v", c.version, a.Notes)
		}
		for _, f := range a.Findings {
			if f.Reachability != "not_assessed" || !strings.HasSuffix(f.URL, f.Advisory) {
				t.Errorf("%s: finding %+v", c.version, f)
			}
		}
	}
}

func TestAssess_notPerformed(t *testing.T) {
	p := provider(t)
	corretto := temurin("17.0.16+8")
	corretto.Distribution = "Corretto"
	if a := p.Assess(corretto, now); a.State != NotPerformed || !strings.Contains(a.Reason, "Corretto is not covered") || len(a.Findings) != 0 {
		t.Errorf("corretto: %+v", a)
	}
	unknown := temurin("17.0.16+8")
	unknown.Distribution = ""
	if a := p.Assess(unknown, now); a.State != NotPerformed || !strings.Contains(a.Reason, "unidentified distribution") {
		t.Errorf("unidentified: %+v", a)
	}
	if a := p.Assess(runtimeinfo.Runtime{Distribution: "Temurin"}, now); a.State != NotPerformed || !strings.Contains(a.Reason, "version is unknown") {
		t.Errorf("no version: %+v", a)
	}
	empty := &OpenJDK{Distributions: map[string]bool{"Temurin": true}}
	if a := empty.Assess(temurin("17.0.16+8"), now); a.State != NotPerformed || !strings.Contains(a.Reason, "no advisories yet") {
		t.Errorf("empty catalog: %+v", a)
	}
	late := p.Assess(temurin("17.0.16+8"), now.AddDate(0, 6, 0))
	if len(late.Notes) != 1 || !strings.Contains(late.Notes[0], "2026-07-21") {
		t.Errorf("stale catalog notes = %v", late.Notes)
	}
}

func TestParseCatalog_rejects(t *testing.T) {
	bad := map[string]string{
		"schema":  "schema: 2\nversion: x\nsource: s\nnote: n\nadvisories: []\n",
		"field":   "schema: 1\nversion: x\nsource: s\nnote: n\nextra: 1\nadvisories: []\n",
		"url":     "schema: 1\nversion: x\nsource: s\nnote: n\nadvisories: [{date: 2026-04-21, url: https://example.org/2026-04-21, last_affected: {}, earlier_affected: false, cves: []}]\n",
		"cve id":  "schema: 1\nversion: x\nsource: s\nnote: n\nadvisories: [{date: 2026-04-21, url: https://openjdk.org/groups/vulnerability/advisories/2026-04-21, last_affected: {\"17\": \"17.0.1\"}, earlier_affected: false, cves: [{id: CVE-26-1, component: c, cvss: 1, lines: [17]}]}]\n",
		"line":    "schema: 1\nversion: x\nsource: s\nnote: n\nadvisories: [{date: 2026-04-21, url: https://openjdk.org/groups/vulnerability/advisories/2026-04-21, last_affected: {\"17\": \"17.0.1\"}, earlier_affected: false, cves: [{id: CVE-2026-0001, component: c, cvss: 1, lines: [21]}]}]\n",
		"version": "schema: 1\nversion: x\nsource: s\nnote: n\nadvisories: [{date: 2026-04-21, url: https://openjdk.org/groups/vulnerability/advisories/2026-04-21, last_affected: {\"17\": \"latest\"}, earlier_affected: false, cves: []}]\n",
	}
	for name, y := range bad {
		if _, err := ParseCatalog([]byte(y)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// TestEmbeddedCatalog: the embedded catalog parses (it is empty until the
// first advisories are curated).
func TestEmbeddedCatalog(t *testing.T) {
	p, err := NewOpenJDK()
	if err != nil {
		t.Fatal(err)
	}
	if !p.Distributions["Temurin"] || len(p.Distributions) != 1 {
		t.Errorf("distributions = %v", p.Distributions)
	}
	if EmbeddedCatalogVersion() == "" {
		t.Error("no catalog version")
	}
}
