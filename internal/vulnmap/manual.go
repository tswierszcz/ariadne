// Package vulnmap maps a CVE to the methods that are actually vulnerable,
// so internal/reach has something concrete to search the call graph for.
// Sources, in priority order (ties always go to the earlier source):
//  1. a manual table (vuln_methods.yaml), checked by hand against each
//     advisory/fix commit — always wins when it has an entry;
//  2. a GitHub fix-commit diff, parsed into the methods its changed lines
//     fall inside (internal/javadiff).
//
// A third source, an LLM asked to read the advisory and diff, is in the
// plan but not implemented here.
package vulnmap

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed vuln_methods.yaml
var defaultManualYAML []byte

// LoadManual parses a vuln_methods.yaml-shaped file: a mapping from CVE
// id to a list of "Class#method" strings. path == "" loads ariadne's own
// embedded default table instead of reading a file.
func LoadManual(path string) (map[string][]string, error) {
	data := defaultManualYAML
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		data = b
	}
	var m map[string][]string
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing vuln_methods.yaml: %w", err)
	}
	return m, nil
}
