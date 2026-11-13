package runtimevuln

import (
	"os"
	"strings"
	"testing"
)

// TestParseAdvisoryHTML on the synthetic page (testdata/README.MD): the
// affected-versions sentence with "and earlier", the OpenJDK matrix with a
// two-row header (line columns aligned from the right), and the OpenJFX
// matrix after it left out.
func TestParseAdvisoryHTML(t *testing.T) {
	b, err := os.ReadFile("testdata/advisory-synthetic.html")
	if err != nil {
		t.Fatal(err)
	}
	a, err := ParseAdvisoryHTML("2026-04-21", string(b))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"8": "8u461", "11": "11.0.28", "17": "17.0.16", "21": "21.0.8", "25": "25"}
	if len(a.LastAffected) != len(want) {
		t.Errorf("last affected = %v", a.LastAffected)
	}
	for k, v := range want {
		if a.LastAffected[k] != v {
			t.Errorf("line %s = %q, want %q", k, a.LastAffected[k], v)
		}
	}
	if !a.EarlierAffected || !strings.HasSuffix(a.URL, "/2026-04-21") {
		t.Errorf("earlier %v url %s", a.EarlierAffected, a.URL)
	}
	if len(a.CVEs) != 2 || a.CVEs[0].ID != "CVE-2026-90001" || a.CVEs[0].Component != "xml/jaxp" || a.CVEs[0].CVSS != 7.5 ||
		len(a.CVEs[0].Lines) != 5 || a.CVEs[1].ID != "CVE-2026-90002" || len(a.CVEs[1].Lines) != 2 || a.CVEs[1].Lines[0] != 21 {
		t.Errorf("cves = %+v", a.CVEs)
	}
	noSentence := strings.Replace(string(b), "affected versions are", "versions are", 1)
	if _, err := ParseAdvisoryHTML("2026-04-21", noSentence); err == nil {
		t.Error("page without the affected-versions sentence accepted")
	}
	without := strings.Replace(string(b), ", and earlier", "", 1)
	if a, _ := ParseAdvisoryHTML("2026-04-21", without); a.EarlierAffected {
		t.Error("earlier_affected without the statement")
	}
}
