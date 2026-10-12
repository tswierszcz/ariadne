// Package report turns a finished analysis run into report.md (via an
// embedded text/template) and results.json (deterministic: sorted by CVE
// id, with each CVE's own method list sorted too).
package report

import (
	"sort"
	"time"

	"ariadne/internal/cves"
	"ariadne/internal/reach"
)

// Header carries everything section 6's report header asks for.
type Header struct {
	App             string
	GeneratedAt     time.Time
	AriadneVersion  string
	Profile         string
	MaxBytecodeVer  int
	GraphNodes      int
	GraphEdges      int
	PreCHAEdges     int
	CHAEdges        int
	UnresolvedDyn   int
	DepsTotal       int
	DepsIdentified  int
	UnidentifiedJar []string
}

// Row is one line of the CVE table.
type Row struct {
	CVE      string
	Package  string
	Version  string
	Severity string
	Status   reach.Status
	Methods  string
	Source   string
}

// Data is everything the Markdown/JSON templates need, already computed
// and sorted — the template itself does no logic beyond formatting.
type Data struct {
	Header
	StatusCounts    map[string]int
	SeverityCounts  map[string]int
	Rows            []Row
	Reachable       []reach.Result
	Framework       []reach.Result
	TotalCVEs       int
	WarningMessages []string
}

// Build assembles Data from a finished run's pieces. severityOf/packageOf
// come from the matching cves.CVE (Analyze's Result doesn't carry them,
// since internal/reach only needs the vulnerable methods, not the
// advisory metadata).
func Build(h Header, list []cves.CVE, results []reach.Result, warnings []string) Data {
	byID := make(map[string]cves.CVE, len(list))
	for _, c := range list {
		byID[c.ID] = c
	}

	sorted := append([]reach.Result(nil), results...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CVE < sorted[j].CVE })
	for i := range sorted {
		sort.Strings(sorted[i].Methods)
	}

	data := Data{
		Header:          h,
		StatusCounts:    map[string]int{},
		SeverityCounts:  map[string]int{},
		TotalCVEs:       len(sorted),
		WarningMessages: warnings,
	}

	for _, r := range sorted {
		data.StatusCounts[string(r.Status)]++
		cve := byID[r.CVE]
		sev := severityLabel(cve.Severity)
		data.SeverityCounts[sev]++

		for _, pkg := range cve.Affected {
			data.Rows = append(data.Rows, Row{
				CVE:      r.CVE,
				Package:  pkg.Package,
				Version:  pkg.Version,
				Severity: sev,
				Status:   r.Status,
				Methods:  joinMethods(r.Methods),
				Source:   r.Source,
			})
		}
		if len(cve.Affected) == 0 {
			data.Rows = append(data.Rows, Row{
				CVE: r.CVE, Severity: sev, Status: r.Status,
				Methods: joinMethods(r.Methods), Source: r.Source,
			})
		}

		switch r.Status {
		case reach.StatusReachable:
			data.Reachable = append(data.Reachable, r)
		case reach.StatusFramework:
			data.Framework = append(data.Framework, r)
		}
	}
	sort.Slice(data.Rows, func(i, j int) bool { return data.Rows[i].CVE < data.Rows[j].CVE })

	return data
}

func joinMethods(methods []string) string {
	s := ""
	for i, m := range methods {
		if i > 0 {
			s += ", "
		}
		s += "`" + m + "`"
	}
	return s
}

// severityLabel extracts a short label ("CRITICAL", "9.8", ...) from
// OSV's raw "TYPE:score" severity string, or "unknown" when absent.
func severityLabel(raw string) string {
	if raw == "" {
		return "unknown"
	}
	return raw
}
