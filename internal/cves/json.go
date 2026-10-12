package cves

import (
	"encoding/json"
	"fmt"
	"os"
)

// WriteJSON writes cves to path as indented, deterministic JSON (Resolve
// already returns the list sorted by CVE id).
func WriteJSON(path string, list []CVE) error {
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling CVE list: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// ReadJSON reads a cves.json file previously written by WriteJSON.
func ReadJSON(path string) ([]CVE, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var list []CVE
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	return list, nil
}
