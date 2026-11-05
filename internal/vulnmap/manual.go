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
}

// manualEntry accepts both forms of one vuln_methods.yaml value: the
// legacy list of "Class#method" strings, or an object with methods plus
// role/coverage/review.
type manualEntry struct {
	Methods []string
	MappingMeta
}

func (e *manualEntry) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.SequenceNode {
		return n.Decode(&e.Methods)
	}
	var obj struct {
		Methods     []string `yaml:"methods"`
		MappingMeta `yaml:",inline"`
	}
	if err := n.Decode(&obj); err != nil {
		return err
	}
	e.Methods, e.MappingMeta = obj.Methods, obj.MappingMeta
	return nil
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
		m[id] = e.Methods
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
