package findings

import (
	"gopkg.in/yaml.v3"

	"os"
	"regexp"
	"strings"
	"testing"
)

// TestReasonCodesDocumented keeps the contract document and the code
// identical: every constant is documented in docs/RESULTS_CONTRACT.MD's
// reason-code table, and every code in that table exists here.
func TestReasonCodesDocumented(t *testing.T) {
	doc, err := os.ReadFile("../../docs/RESULTS_CONTRACT.MD")
	if err != nil {
		t.Fatal(err)
	}
	text := string(doc)
	start, end := strings.Index(text, "## 3. Reason codes"), strings.Index(text, "## 4.")
	if start < 0 || end < start {
		t.Fatal("docs/RESULTS_CONTRACT.MD has no \"## 3. Reason codes\" section before \"## 4.\"")
	}
	documented := map[string]bool{}
	for _, m := range regexp.MustCompile("(?m)^\\| `([A-Z_]+)` \\|").FindAllStringSubmatch(text[start:end], -1) {
		documented[m[1]] = true
	}
	defined := map[string]bool{}
	for _, c := range All {
		if defined[string(c)] {
			t.Errorf("duplicate code %s", c)
		}
		defined[string(c)] = true
		if !documented[string(c)] {
			t.Errorf("%s is not documented in docs/RESULTS_CONTRACT.MD", c)
		}
	}
	for c := range documented {
		if !defined[c] {
			t.Errorf("docs/RESULTS_CONTRACT.MD documents %s, which internal/findings does not define", c)
		}
	}
}

type controlCase struct {
	ID        string `yaml:"id"`
	Kind      string `yaml:"kind"`
	Extension string `yaml:"extension"`
	Matrix    string `yaml:"matrix"`
	Source    string `yaml:"source"`
	Expected  struct {
		Applicability string   `yaml:"applicability"`
		Reachability  string   `yaml:"reachability"`
		Reasons       []string `yaml:"reasons"`
	} `yaml:"expected"`
}

// TestControlSetCoversExtensions checks CVE:F0's completion criterion
// "the control set covers all five extensions": every extension has at
// least one mechanics and one quality case, every quality case cites a
// source, and every expectation uses the contract's vocabulary.
func TestControlSetCoversExtensions(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/cve-control/cases.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Cases []controlCase `yaml:"cases"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for _, c := range All {
		codes[string(c)] = true
	}
	applicability := map[string]bool{"affected": true, "not_affected": true, "undetermined": true}
	reachability := map[string]bool{"REACHABLE": true, "FRAMEWORK": true, "NOT_REACHABLE": true, "UNKNOWN": true, "NOT_APPLICABLE": true, "not_assessed": true}
	type cov struct{ mechanics, quality int }
	byExt := map[string]*cov{}
	ids := map[string]bool{}
	for _, c := range doc.Cases {
		if ids[c.ID] {
			t.Errorf("duplicate case id %s", c.ID)
		}
		ids[c.ID] = true
		if byExt[c.Extension] == nil {
			byExt[c.Extension] = &cov{}
		}
		switch c.Kind {
		case "mechanics":
			byExt[c.Extension].mechanics++
		case "quality":
			byExt[c.Extension].quality++
			if c.Source == "" {
				t.Errorf("%s: a quality case needs a source", c.ID)
			}
		default:
			t.Errorf("%s: kind %q", c.ID, c.Kind)
		}
		if !applicability[c.Expected.Applicability] || !reachability[c.Expected.Reachability] {
			t.Errorf("%s: expected %+v uses states outside the contract", c.ID, c.Expected)
		}
		for _, r := range c.Expected.Reasons {
			if !codes[r] {
				t.Errorf("%s: unknown reason code %s", c.ID, r)
			}
		}
	}
	for _, ext := range []string{"version-mapping", "pre-fix", "shaded", "freshness", "jdk"} {
		c := byExt[ext]
		if c == nil || c.mechanics == 0 || c.quality == 0 {
			t.Errorf("extension %s: want at least one mechanics and one quality case, got %+v", ext, c)
		}
	}
}
