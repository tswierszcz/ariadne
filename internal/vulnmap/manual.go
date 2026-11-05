// Package vulnmap maps a CVE to the methods that are actually vulnerable,
// so internal/reach has something concrete to search the call graph for.
// Sources, in priority order (ties always go to the earlier source):
//  1. a manual table (vuln_methods.yaml), checked by hand against each
//     advisory/fix commit — always wins when it has an entry;
//  2. a GitHub fix-commit diff, parsed into the methods its changed lines
//     fall inside (internal/javadiff);
//  3. an LLM asked to read the advisory and, when there is one, the fix
//     diff (llm*.go; opt-in with -llm; see docs/phase4-llm.md). It only
//     fills gaps left by sources 1 and 2, and its answer is checked against
//     the analyzed bytecode before use.
package vulnmap

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"ariadne/internal/mavenversion"

	"gopkg.in/yaml.v3"
)

//go:embed vuln_methods.yaml
var defaultManualYAML []byte

// Mapping roles, coverage and review states (docs/RESULTS_CONTRACT.MD §1,
// CVE:F1b). Role is entry-level: every method of one entry has the same
// role.
const (
	RoleVulnerable = "vulnerable"
	RoleEntrypoint = "entrypoint"
	RoleSink       = "sink"

	CoverageComplete = "complete_under_declared_assumptions"
	CoveragePartial  = "partial"
	CoverageUnknown  = "unknown"

	ReviewReviewed   = "reviewed"
	ReviewUnreviewed = "unreviewed"
)

// MappingMeta is a manual entry's role, coverage and review state. The
// legacy list form carries none; LoadManualMeta then reports the
// contract's default for legacy manual mappings (§6): partial, reviewed,
// role unknown ("").
type MappingMeta struct {
	Role     string `yaml:"role"`
	Coverage string `yaml:"coverage"`
	Review   string `yaml:"review"`
	// Variants are the versioned method lists of a schema-v2 entry
	// (CVE:F4); nil for an unversioned entry.
	Variants []Variant `yaml:"-"`
}

// Variant is one versioned method list of a mapping (CVE:F4): it applies
// to the component instances of Package (any package when empty) whose
// version is in Versions (Maven range syntax).
type Variant struct {
	Package  string             `yaml:"package"`
	Versions string             `yaml:"versions"`
	Methods  []string           `yaml:"methods"`
	Range    mavenversion.Range `yaml:"-"`
}

// manualEntry accepts every form of one vuln_methods.yaml value: the
// legacy list of "Class#method" strings, an object with methods plus
// role/coverage/review, or (schema v2, CVE:F4) an object with variants.
type manualEntry struct {
	Methods []string
	MappingMeta
}

func keysOf(n *yaml.Node) []string {
	if n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	var keys []string
	for i := 0; i+1 < len(n.Content); i += 2 {
		keys = append(keys, n.Content[i].Value)
	}
	return keys
}

func checkKeys(n *yaml.Node, what string, allowed ...string) error {
	ok := map[string]bool{}
	for _, a := range allowed {
		ok[a] = true
	}
	for _, k := range keysOf(n) {
		if !ok[k] {
			return fmt.Errorf("%s: unknown field %q (allowed: %s)", what, k, strings.Join(allowed, ", "))
		}
	}
	return nil
}

func (e *manualEntry) UnmarshalYAML(n *yaml.Node) error {
	node := n
	if node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	if node.Kind == yaml.SequenceNode {
		return n.Decode(&e.Methods)
	}
	if err := checkKeys(node, "mapping", "methods", "role", "coverage", "review", "variants"); err != nil {
		return err
	}
	var obj struct {
		Methods     []string    `yaml:"methods"`
		Variants    []yaml.Node `yaml:"variants"`
		MappingMeta `yaml:",inline"`
	}
	if err := n.Decode(&obj); err != nil {
		return err
	}
	e.Methods, e.MappingMeta = obj.Methods, obj.MappingMeta
	for i := range obj.Variants {
		if err := checkKeys(&obj.Variants[i], fmt.Sprintf("variants[%d]", i), "package", "versions", "methods"); err != nil {
			return err
		}
		var v Variant
		if err := obj.Variants[i].Decode(&v); err != nil {
			return err
		}
		e.Variants = append(e.Variants, v)
	}
	return nil
}

var advisoryID = regexp.MustCompile(`^[A-Z][A-Z0-9]*-[0-9A-Za-z][0-9A-Za-z._:-]*$`)
var packageName = regexp.MustCompile(`^[^:\s]+:[^:\s]+$`)

// validate checks one entry: identifiers, selectors, ranges, and that a
// mapping is either unversioned (methods) or versioned (variants).
func (e *manualEntry) validate(id string) error {
	if !advisoryID.MatchString(id) {
		return fmt.Errorf("%q is not an advisory identifier", id)
	}
	switch {
	case len(e.Variants) > 0 && len(e.Methods) > 0:
		return fmt.Errorf("%s: give either methods or variants, not both", id)
	case len(e.Variants) == 0 && len(e.Methods) == 0:
		return fmt.Errorf("%s: no methods", id)
	}
	for _, m := range e.Methods {
		if _, err := ParseSelector(m); err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
	}
	for i := range e.Variants {
		v := &e.Variants[i]
		if v.Package != "" && !packageName.MatchString(v.Package) {
			return fmt.Errorf("%s: variants[%d]: package %q is not group:artifact", id, i, v.Package)
		}
		r, err := mavenversion.ParseRange(v.Versions)
		if err != nil {
			return fmt.Errorf("%s: variants[%d]: %w", id, i, err)
		}
		v.Range = r
		if len(v.Methods) == 0 {
			return fmt.Errorf("%s: variants[%d]: no methods", id, i)
		}
		for _, m := range v.Methods {
			if _, err := ParseSelector(m); err != nil {
				return fmt.Errorf("%s: variants[%d]: %w", id, i, err)
			}
		}
	}
	return nil
}

// AllMethods is every method of the entry (the union of its variants).
func (e manualEntry) AllMethods() []string {
	if len(e.Variants) == 0 {
		return e.Methods
	}
	seen := map[string]bool{}
	var out []string
	for _, v := range e.Variants {
		for _, m := range v.Methods {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	sort.Strings(out)
	return out
}

func loadManualEntries(path string) (map[string]manualEntry, error) {
	data := defaultManualYAML
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		data = b
	}
	var m map[string]manualEntry
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing vuln_methods.yaml: %w", err)
	}
	for id, e := range m {
		if err := e.validate(id); err != nil {
			return nil, fmt.Errorf("vuln_methods.yaml: %w", err)
		}
		m[id] = e
		switch e.Role {
		case "", RoleVulnerable, RoleEntrypoint, RoleSink:
		default:
			return nil, fmt.Errorf("vuln_methods.yaml: %s: unknown role %q", id, e.Role)
		}
		switch e.Coverage {
		case "", CoverageComplete, CoveragePartial, CoverageUnknown:
		default:
			return nil, fmt.Errorf("vuln_methods.yaml: %s: unknown coverage %q", id, e.Coverage)
		}
		switch e.Review {
		case "", ReviewReviewed, ReviewUnreviewed:
		default:
			return nil, fmt.Errorf("vuln_methods.yaml: %s: unknown review %q", id, e.Review)
		}
	}
	return m, nil
}

// LoadManual parses a vuln_methods.yaml-shaped file: a mapping from CVE
// id to a list of "Class#method" strings, or to an object with `methods`
// and optional `role`/`coverage`/`review`. path == "" loads ariadne's
// own embedded default table instead of reading a file.
func LoadManual(path string) (map[string][]string, error) {
	entries, err := loadManualEntries(path)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string, len(entries))
	for id, e := range entries {
		m[id] = e.AllMethods()
	}
	return m, nil
}

// LoadManualMeta returns every manual entry's role/coverage/review, with
// the legacy defaults filled in (coverage partial, review reviewed).
func LoadManualMeta(path string) (map[string]MappingMeta, error) {
	entries, err := loadManualEntries(path)
	if err != nil {
		return nil, err
	}
	m := make(map[string]MappingMeta, len(entries))
	for id, e := range entries {
		meta := e.MappingMeta
		if meta.Coverage == "" {
			meta.Coverage = CoveragePartial
		}
		if meta.Review == "" {
			meta.Review = ReviewReviewed
		}
		m[id] = meta
	}
	return m, nil
}
