package profile

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// rulesSourceDigest is ruleSourceDigest's value for RulesVersion. A change
// to the rule code (comments aside) changes it: bump RulesVersion, then
// update this value to the one the failure prints.
const rulesSourceDigest = "398baee861dd5605"

// ruleSourceDigest hashes the non-test Go files of the packages whose rules
// RulesVersion names, parsed without comments and printed back, so a
// comment or formatting edit needs no bump.
func ruleSourceDigest(t *testing.T) string {
	t.Helper()
	var files []string
	for _, dir := range []string{".", "../springxml"} {
		m, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range m {
			if !strings.HasSuffix(f, "_test.go") && filepath.Base(f) != "rules.go" {
				files = append(files, f)
			}
		}
	}
	sort.Strings(files)
	h := sha256.New()
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		af, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := printer.Fprint(&buf, fset, af); err != nil {
			t.Fatal(err)
		}
		h.Write([]byte(filepath.ToSlash(f) + "\n"))
		h.Write(buf.Bytes())
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// TestRulesVersion_tracksRuleSource (FRAMEWORKS F5 f5-1): the root and
// framework-fact rules cannot change without a RulesVersion bump, or a
// snapshot comparison could not say why a verdict moved — the profile-rule
// counterpart of TestAlgorithmVersions_everyEdgeBasis (ROADMAP B3 b3-4).
func TestRulesVersion_tracksRuleSource(t *testing.T) {
	if got := ruleSourceDigest(t); got != rulesSourceDigest {
		t.Errorf("rule source digest %s, recorded %s for RulesVersion %q: the root/framework-fact rules changed — bump RulesVersion and set rulesSourceDigest to %s",
			got, rulesSourceDigest, RulesVersion, got)
	}
}

// TestBuiltinDigest_followsProfileContent: the digest covers every built-in
// profile and changes when one of them does.
func TestBuiltinDigest_followsProfileContent(t *testing.T) {
	d := BuiltinDigest()
	if len(d) != 16 {
		t.Fatalf("digest %q", d)
	}
	if n := len(builtinYAMLs()); n != 7 {
		t.Errorf("%d built-in profiles hashed, want 7", n)
	}
	saved := stripesYAML
	defer func() { stripesYAML = saved }()
	stripesYAML = append(append([]byte{}, saved...), '\n')
	if BuiltinDigest() == d {
		t.Error("digest unchanged after a built-in profile changed")
	}
}
