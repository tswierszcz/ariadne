package cves

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"ariadne/internal/httpcache"
	"ariadne/internal/input"
)

const (
	// DefaultBaseURL is OSV's public API.
	DefaultBaseURL = "https://api.osv.dev"
	// osvBatchLimit is the documented maximum number of queries per
	// querybatch request.
	osvBatchLimit = 1000
	// maxExtraPages bounds how many extra pages we'll fetch for a package
	// whose vulnerability count overflows one querybatch page. In
	// practice no real Maven package comes close to this.
	maxExtraPages = 20
	// detailWorkers is how many /v1/vulns/{id} lookups run in parallel.
	detailWorkers = 8
)

// Reference is one reference URL from an advisory, e.g. a GitHub fix commit.
type Reference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Affected records that this CVE was reported against one of the app's
// resolved dependency GAVs.
type Affected struct {
	Package string `json:"package"` // group:artifact
	Version string `json:"version"`
}

// CVE is one deduplicated vulnerability: all GHSA/CVE aliases for the same
// underlying advisory are merged into a single entry.
type CVE struct {
	ID         string      `json:"id"` // a CVE id when one exists among the aliases, else the OSV/GHSA id
	Aliases    []string    `json:"aliases"`
	Summary    string      `json:"summary"`
	Severity   string      `json:"severity,omitempty"` // CVSS vector/score, verbatim from OSV, if present
	Affected   []Affected  `json:"affected"`
	References []Reference `json:"references"`
}

// Client queries OSV, with every request going through a disk cache.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Cache      *httpcache.Cache
}

// NewClient returns a Client with sensible defaults; httpClient may be nil
// to use http.DefaultClient.
func NewClient(cache *httpcache.Cache) *Client {
	return &Client{BaseURL: DefaultBaseURL, HTTPClient: http.DefaultClient, Cache: cache}
}

// Resolve looks up every unique package:version in gavs against OSV and
// returns the deduplicated, sorted CVE list.
func (c *Client) Resolve(ctx context.Context, gavs []input.GAV) ([]CVE, error) {
	unique := uniquePackageVersions(gavs)
	if len(unique) == 0 {
		return nil, nil
	}

	vulnIDsByPV, err := c.queryBatch(ctx, unique)
	if err != nil {
		return nil, err
	}

	// Collect the set of distinct vuln ids, and which package:versions
	// reported each one.
	affectedBy := make(map[string][]pkgVersion) // vuln id -> affected GAVs
	for pv, ids := range vulnIDsByPV {
		for _, id := range ids {
			affectedBy[id] = append(affectedBy[id], pv)
		}
	}
	if len(affectedBy) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(affectedBy))
	for id := range affectedBy {
		ids = append(ids, id)
	}
	vulns, err := c.fetchDetails(ctx, ids)
	if err != nil {
		return nil, err
	}

	return dedupe(vulns, affectedBy), nil
}

type pkgVersion struct {
	Package string // group:artifact
	Version string
}

func uniquePackageVersions(gavs []input.GAV) []pkgVersion {
	seen := make(map[pkgVersion]bool)
	var out []pkgVersion
	for _, g := range gavs {
		pv := pkgVersion{Package: g.PackageName(), Version: g.Version}
		if !seen[pv] {
			seen[pv] = true
			out = append(out, pv)
		}
	}
	return out
}

// queryBatch returns, for each package:version, the list of OSV/GHSA
// vulnerability ids that affect it.
func (c *Client) queryBatch(ctx context.Context, pvs []pkgVersion) (map[pkgVersion][]string, error) {
	result := make(map[pkgVersion][]string, len(pvs))

	for start := 0; start < len(pvs); start += osvBatchLimit {
		end := start + osvBatchLimit
		if end > len(pvs) {
			end = len(pvs)
		}
		chunk := pvs[start:end]

		reqBody := osvQueryBatchRequest{Queries: make([]osvQuery, len(chunk))}
		for i, pv := range chunk {
			reqBody.Queries[i] = osvQuery{
				Package: osvPackage{Ecosystem: "Maven", Name: pv.Package},
				Version: pv.Version,
			}
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshaling OSV querybatch request: %w", err)
		}

		key := httpcache.Key("POST", c.BaseURL+"/v1/querybatch", string(body))
		respBody, err := httpcache.Fetch(c.HTTPClient, c.Cache, key, c.jsonRequest(ctx, "POST", "/v1/querybatch", body))
		if err != nil {
			return nil, fmt.Errorf("OSV querybatch: %w", err)
		}

		var resp osvQueryBatchResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return nil, fmt.Errorf("decoding OSV querybatch response: %w", err)
		}
		if len(resp.Results) != len(chunk) {
			return nil, fmt.Errorf("OSV querybatch returned %d results for %d queries", len(resp.Results), len(chunk))
		}

		for i, r := range resp.Results {
			pv := chunk[i]
			ids := idsOf(r.Vulns)

			token := r.NextPageToken
			for page := 0; token != "" && page < maxExtraPages; page++ {
				more, nextToken, err := c.queryOnePage(ctx, pv, token)
				if err != nil {
					return nil, err
				}
				ids = append(ids, more...)
				token = nextToken
			}
			result[pv] = ids
		}
	}
	return result, nil
}

// queryOnePage fetches one extra page of results for a single
// package:version via the non-batch /v1/query endpoint, which is the only
// one OSV supports passing a page_token to.
func (c *Client) queryOnePage(ctx context.Context, pv pkgVersion, pageToken string) (ids []string, nextToken string, err error) {
	reqBody := map[string]any{
		"package":    osvPackage{Ecosystem: "Maven", Name: pv.Package},
		"version":    pv.Version,
		"page_token": pageToken,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("marshaling OSV query request: %w", err)
	}
	key := httpcache.Key("POST", c.BaseURL+"/v1/query", string(body))
	respBody, err := httpcache.Fetch(c.HTTPClient, c.Cache, key, c.jsonRequest(ctx, "POST", "/v1/query", body))
	if err != nil {
		return nil, "", fmt.Errorf("OSV query (page): %w", err)
	}
	var page osvBatchResult
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, "", fmt.Errorf("decoding OSV query (page) response: %w", err)
	}
	return idsOf(page.Vulns), page.NextPageToken, nil
}

func idsOf(vs []osvVulnID) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.ID
	}
	return out
}

// fetchDetails fetches GET /v1/vulns/{id} for each id, in parallel.
func (c *Client) fetchDetails(ctx context.Context, ids []string) ([]osvVuln, error) {
	sort.Strings(ids) // deterministic order regardless of map iteration

	type result struct {
		vuln osvVuln
		err  error
	}
	jobs := make(chan string)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < detailWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				v, err := c.fetchOne(ctx, id)
				results <- result{vuln: v, err: err}
			}
		}()
	}
	go func() {
		for _, id := range ids {
			jobs <- id
		}
		close(jobs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	vulns := make([]osvVuln, 0, len(ids))
	var firstErr error
	for r := range results {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		if r.err == nil {
			vulns = append(vulns, r.vuln)
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return vulns, nil
}

func (c *Client) fetchOne(ctx context.Context, id string) (osvVuln, error) {
	key := httpcache.Key("GET", c.BaseURL+"/v1/vulns/"+id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/v1/vulns/"+id, nil)
	if err != nil {
		return osvVuln{}, fmt.Errorf("building request for %s: %w", id, err)
	}
	body, err := httpcache.Fetch(c.HTTPClient, c.Cache, key, req)
	if err != nil {
		return osvVuln{}, fmt.Errorf("fetching OSV vuln %s: %w", id, err)
	}
	var v osvVuln
	if err := json.Unmarshal(body, &v); err != nil {
		return osvVuln{}, fmt.Errorf("decoding OSV vuln %s: %w", id, err)
	}
	return v, nil
}

func (c *Client) jsonRequest(ctx context.Context, method, path string, body []byte) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		// Only non-recoverable programmer errors (bad method/URL) reach
		// here; the caller always passes fixed, known-good values.
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// dedupe merges vulnerabilities that alias each other (a GHSA id and its
// CVE alias, say) into one CVE entry, preferring a CVE id as the primary
// id when the group has one.
func dedupe(vulns []osvVuln, affectedBy map[string][]pkgVersion) []CVE {
	// Union-find over ids: every vuln's own id and all its aliases belong
	// to the same group.
	parent := make(map[string]string)
	find := func(x string) string {
		for parent[x] != x {
			x = parent[x]
		}
		return x
	}
	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}
	ensure := func(id string) {
		if _, ok := parent[id]; !ok {
			parent[id] = id
		}
	}

	byID := make(map[string]osvVuln, len(vulns))
	for _, v := range vulns {
		byID[v.ID] = v
		ensure(v.ID)
		for _, a := range v.Aliases {
			ensure(a)
			union(v.ID, a)
		}
	}

	// recordIDs: root -> ids for which we actually fetched an osvVuln
	// record (used to merge summary/severity/references/affected).
	// allIDs: root -> every id in the group, including aliases that OSV
	// never returned a /v1/vulns/{id} record for on their own (a CVE
	// alias of a GHSA-keyed record, typically) — pickPrimaryID needs the
	// full set so it can still prefer a CVE id that only exists as an
	// alias.
	recordIDs := make(map[string][]string)
	for _, v := range vulns {
		root := find(v.ID)
		recordIDs[root] = append(recordIDs[root], v.ID)
	}
	allIDs := make(map[string][]string)
	for id := range parent {
		root := find(id)
		allIDs[root] = append(allIDs[root], id)
	}

	var out []CVE
	for root, memberIDs := range recordIDs {
		full := allIDs[root]
		sort.Strings(full)
		primary := pickPrimaryID(full)

		// Merge fields from every member (there's usually exactly one,
		// since OSV already cross-links GHSA/CVE pairs via aliases on a
		// single record, but handle the general case).
		var allAliases []string
		seenAlias := map[string]bool{}
		var summary, severity string
		var affected []Affected
		var refs []Reference
		seenAffected := map[Affected]bool{}
		seenRef := map[Reference]bool{}

		for _, id := range memberIDs {
			v := byID[id]
			if summary == "" {
				summary = v.Summary
			}
			// A GHSA-sourced qualitative label ("HIGH", "CRITICAL", ...) is
			// far more report-friendly than the raw CVSS vector string, so
			// prefer it; only fall back to the vector when OSV has no such
			// label for this advisory (common for CVE-only, non-GHSA records).
			if severity == "" && v.DatabaseSpecific.Severity != "" {
				severity = v.DatabaseSpecific.Severity
			}
			if severity == "" && len(v.Severity) > 0 {
				severity = v.Severity[0].Type + ":" + v.Severity[0].Score
			}
			for _, a := range append([]string{v.ID}, v.Aliases...) {
				if a != primary && !seenAlias[a] {
					seenAlias[a] = true
					allAliases = append(allAliases, a)
				}
			}
			for _, pv := range affectedBy[v.ID] {
				a := Affected{Package: pv.Package, Version: pv.Version}
				if !seenAffected[a] {
					seenAffected[a] = true
					affected = append(affected, a)
				}
			}
			for _, r := range v.References {
				ref := Reference{Type: r.Type, URL: r.URL}
				if !seenRef[ref] {
					seenRef[ref] = true
					refs = append(refs, ref)
				}
			}
		}

		sort.Strings(allAliases)
		sort.Slice(affected, func(i, j int) bool {
			if affected[i].Package != affected[j].Package {
				return affected[i].Package < affected[j].Package
			}
			return affected[i].Version < affected[j].Version
		})
		sort.Slice(refs, func(i, j int) bool { return refs[i].URL < refs[j].URL })

		out = append(out, CVE{
			ID:         primary,
			Aliases:    allAliases,
			Summary:    summary,
			Severity:   severity,
			Affected:   affected,
			References: refs,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// pickPrimaryID prefers a CVE id over a bare GHSA/other id, since CVE ids
// are what the rest of the report and vuln_methods.yaml key on.
func pickPrimaryID(ids []string) string {
	for _, id := range ids {
		if strings.HasPrefix(id, "CVE-") {
			return id
		}
	}
	return ids[0] // already sorted, so this is deterministic
}
