package research

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// RESEARCH_HARNESS RH-A0: the committed producer acceptance package
// (docs/research-interface/acceptance, scripts/research-acceptance.sh) is
// intact, every real response in it passes CheckResponse and is what this
// producer still answers (replay), and every mutation breaks the rule it
// names. A failing replay means a query changed what it answers: regenerate
// the package and say so in the PR (the consumer compares verdicts on it).

const acceptanceDir = "../../docs/research-interface/acceptance"

type acceptanceCase struct {
	ID            string   `json:"id"`
	Kind          string   `json:"kind"`
	Bundle        string   `json:"bundle"`
	Request       string   `json:"request"`
	Response      string   `json:"response"`
	ExpectedValid bool     `json:"expected_valid"`
	ExpectedRules []string `json:"expected_rules"`
}

func TestAcceptancePackage(t *testing.T) {
	f, err := os.Open(filepath.Join(acceptanceDir, "SHA256SUMS"))
	if err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(f)
	files := 0
	for sc.Scan() {
		sum, name, ok := strings.Cut(sc.Text(), "  ")
		if !ok {
			t.Fatalf("SHA256SUMS line %q", sc.Text())
		}
		raw, err := os.ReadFile(filepath.Join(acceptanceDir, name))
		if err != nil {
			t.Fatal(err)
		}
		s := sha256.Sum256(raw)
		if hex.EncodeToString(s[:]) != sum {
			t.Errorf("%s: hash differs from SHA256SUMS", name)
		}
		files++
	}
	f.Close()
	var pkg struct {
		ContractRevision string           `json:"contract_revision"`
		Rules            []string         `json:"rules"`
		Cases            []acceptanceCase `json:"cases"`
	}
	raw, err := os.ReadFile(filepath.Join(acceptanceDir, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.ContractRevision != ContractRevision {
		t.Fatalf("package for contract %s, producer implements %s", pkg.ContractRevision, ContractRevision)
	}
	bundles := map[string]*Bundle{}
	exercised := map[string]bool{}
	responses, mutations := 0, 0
	for _, c := range pkg.Cases {
		b := bundles[c.Bundle]
		if b == nil {
			if b, err = Load(filepath.Join(acceptanceDir, c.Bundle)); err != nil {
				t.Fatalf("%s: %v", c.Bundle, err)
			}
			bundles[c.Bundle] = b
		}
		req, err1 := os.ReadFile(filepath.Join(acceptanceDir, c.Request))
		resp, err2 := os.ReadFile(filepath.Join(acceptanceDir, c.Response))
		if err1 != nil || err2 != nil {
			t.Fatalf("%s: %v %v", c.ID, err1, err2)
		}
		got := CheckResponse(b, req, resp)
		if c.ExpectedValid {
			responses++
			if len(got) > 0 {
				t.Errorf("%s: a real response breaks %v", c.ID, got)
			}
			if r := Replay(b, req, resp); r != nil {
				t.Errorf("%s: this producer now answers differently (regenerate with scripts/research-acceptance.sh)", c.ID)
			}
			continue
		}
		mutations++
		for _, rule := range c.ExpectedRules {
			exercised[rule] = true
			found := false
			for _, v := range got {
				found = found || v.Rule == rule
			}
			if !found {
				t.Errorf("%s: expected rule %s, got %v", c.ID, rule, got)
			}
		}
	}
	// Rules no case of the control library can break (ACCEPTANCE.md §3).
	notExercised := map[string]bool{RuleErrorEnvelope: true, RuleRecordType: true, RuleReplay: true}
	for _, r := range Rules {
		if !exercised[r] && !notExercised[r] {
			t.Errorf("rule %s: no mutation breaks it", r)
		}
	}
	t.Logf("%d files, %d responses, %d mutations", files, responses, mutations)
}

// The v1 contract cases (RI0 fixtures): CheckResponse gives the semantic
// verdict validate.mjs gives, on the fixture bundle.
func TestCheckContractCases(t *testing.T) {
	v1 := "../../docs/research-interface/v1"
	b, err := Load(filepath.Join(v1, "fixtures/bundle"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(v1, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Cases []struct {
			ID            string `json:"id"`
			Path          string `json:"path"`
			Definition    string `json:"definition"`
			Request       string `json:"request"`
			SemanticValid *bool  `json:"semantic_valid"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &suite); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, c := range suite.Cases {
		if c.Definition != "Response" || c.Request == "" || c.SemanticValid == nil {
			continue
		}
		req, _ := os.ReadFile(filepath.Join(v1, c.Request))
		resp, _ := os.ReadFile(filepath.Join(v1, c.Path))
		got := CheckResponse(b, req, resp)
		if valid := len(got) == 0; valid != *c.SemanticValid {
			t.Errorf("%s: valid=%v, the contract case says %v: %v", c.ID, valid, *c.SemanticValid, got)
		}
		n++
	}
	if n == 0 {
		t.Fatal("no contract response case checked")
	}
}
