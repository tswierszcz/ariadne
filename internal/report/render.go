package report

import (
	"ariadne/internal/atomicfile"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"ariadne/internal/reach"
)

//go:embed templates/report.md.tmpl
var markdownTemplate string

// RenderMarkdown fills the embedded report.md template with data.
func RenderMarkdown(data Data) (string, error) {
	tmpl, err := template.New("report.md").Funcs(template.FuncMap{"gapCounts": GapCounts, "join": strings.Join, "deref": func(p *int) int { return *p }}).Parse(markdownTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing report template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering report: %w", err)
	}
	return buf.String(), nil
}

// WriteMarkdown renders data and writes it to path.
func WriteMarkdown(path string, data Data) error {
	rendered, err := RenderMarkdown(data)
	if err != nil {
		return err
	}
	return atomicfile.WriteFile(path, []byte(rendered))
}

// WriteResultsJSON writes results to path as deterministic JSON
// (ResultsJSON), atomically.
func WriteResultsJSON(path string, results []reach.Result) error {
	b, err := ResultsJSON(results)
	if err != nil {
		return err
	}
	return atomicfile.WriteFile(path, b)
}

// ResultsJSON renders the v1 results as deterministic JSON: sorted by CVE
// id, with each result's own Methods slice sorted too, matching the
// plan's "results.json must be deterministic" requirement so two runs
// against the same cache produce byte-identical output.
func ResultsJSON(results []reach.Result) ([]byte, error) {
	sorted := append([]reach.Result(nil), results...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CVE < sorted[j].CVE })
	for i := range sorted {
		sort.Strings(sorted[i].Methods)
	}

	b, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling results: %w", err)
	}
	return append(b, '\n'), nil
}
