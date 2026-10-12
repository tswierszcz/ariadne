// Package cves resolves a list of Maven coordinates (GAVs) to known
// vulnerabilities via OSV (https://osv.dev), and deduplicates results that
// are the same vulnerability under different ids (a GHSA id and its CVE
// alias, say) into one entry.
package cves

// osvQuery is one element of a querybatch request.
type osvQuery struct {
	Package osvPackage `json:"package"`
	Version string     `json:"version"`
}

type osvPackage struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
}

type osvQueryBatchRequest struct {
	Queries []osvQuery `json:"queries"`
}

// osvQueryBatchResponse is one page of querybatch results, one entry per
// query, in the same order as the request.
type osvQueryBatchResponse struct {
	Results []osvBatchResult `json:"results"`
}

type osvBatchResult struct {
	Vulns         []osvVulnID `json:"vulns"`
	NextPageToken string      `json:"next_page_token"`
}

type osvVulnID struct {
	ID string `json:"id"`
}

// osvVuln is the detail returned by GET /v1/vulns/{id}. Only the fields
// ariadne uses are modeled; OSV's schema has more.
type osvVuln struct {
	ID         string         `json:"id"`
	Aliases    []string       `json:"aliases"`
	Summary    string         `json:"summary"`
	Details    string         `json:"details"`
	Severity   []osvSeverity  `json:"severity"`
	Affected   []osvAffected  `json:"affected"`
	References []osvReference `json:"references"`
}

type osvSeverity struct {
	Type  string `json:"type"` // e.g. "CVSS_V3"
	Score string `json:"score"`
}

type osvAffected struct {
	Package  osvPackage `json:"package"`
	Ranges   []osvRange `json:"ranges"`
	Versions []string   `json:"versions"`
}

type osvRange struct {
	Type   string              `json:"type"`
	Events []map[string]string `json:"events"`
}

type osvReference struct {
	Type string `json:"type"` // e.g. "FIX", "ADVISORY", "WEB"
	URL  string `json:"url"`
}
