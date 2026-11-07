package deserialize

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

type controlCase struct {
	ID         string            `yaml:"id"`
	Component  string            `yaml:"component"`
	Role       string            `yaml:"role"`
	Chain      string            `yaml:"chain"`
	Vulnerable string            `yaml:"vulnerable"`
	Fixed      string            `yaml:"fixed"`
	Source     string            `yaml:"source"`
	Expected   map[string]string `yaml:"expected"`
	CheckedOn  string            `yaml:"checked_on"`
}

// TestDeserializationControlSet (DESERIALIZATION Z4): the source-documented
// CVE control set is well formed, names only catalog chains, and covers
// every advisory a catalog chain names — a chain added without a
// documented case fails here.
func TestDeserializationControlSet(t *testing.T) {
	b, err := os.ReadFile("../../testdata/deserialize/cves.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Cases []controlCase `yaml:"cases"`
	}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	chains := map[string]ChainSpec{}
	for _, c := range ChainCatalog {
		chains[c.ID] = c
	}
	covered := map[string]bool{}
	roles := map[string]int{}
	for _, c := range doc.Cases {
		roles[c.Role]++
		if c.ID == "" || c.Component == "" || c.Source == "" || c.Vulnerable == "" || c.CheckedOn == "" || len(c.Expected) == 0 {
			t.Errorf("%s/%s: incomplete case", c.ID, c.Component)
		}
		switch c.Role {
		case "chain", "entry":
			spec, ok := chains[c.Chain]
			if !ok {
				t.Errorf("%s: chain %q not in the catalog", c.ID, c.Chain)
				continue
			}
			names := false
			for _, a := range spec.Advisories {
				if a == c.ID {
					names = true
				}
			}
			if !names {
				t.Errorf("%s: catalog chain %s does not name it", c.ID, c.Chain)
			}
			covered[c.Chain+"\x00"+c.ID] = true
		case "callback":
			if c.Chain != "" {
				t.Errorf("%s: a callback case names a chain", c.ID)
			}
		default:
			t.Errorf("%s: role %q", c.ID, c.Role)
		}
	}
	for _, spec := range ChainCatalog {
		for _, a := range spec.Advisories {
			if !covered[spec.ID+"\x00"+a] {
				t.Errorf("chain %s names %s, which has no control case", spec.ID, a)
			}
		}
	}
	for _, r := range []string{"chain", "entry", "callback"} {
		if roles[r] == 0 {
			t.Errorf("no %s case", r)
		}
	}
}
