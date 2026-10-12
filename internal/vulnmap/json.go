package vulnmap

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// WriteJSON writes m to path as indented, deterministic JSON (keys sorted),
// matching the plan's `vulnmap.json: {cve_id: {"methods": [...], "source":
// "manual|diff|llm"}}` shape.
func WriteJSON(path string, m map[string]Entry) error {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// encoding/json already sorts map[string]... keys when marshaling, so
	// a plain Marshal would already be deterministic; building an
	// ordered intermediate anyway keeps this independent of that detail.
	ordered := make(map[string]Entry, len(m))
	for _, id := range ids {
		ordered[id] = m[id]
	}

	b, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling vulnmap: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// ReadJSON reads a vulnmap.json file previously written by WriteJSON.
func ReadJSON(path string) (map[string]Entry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var m map[string]Entry
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	return m, nil
}
