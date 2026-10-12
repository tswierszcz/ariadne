package vulnmap

import (
	"context"
	"net/http"
	"sort"
	"sync"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

const (
	SourceManual = "manual"
	SourceDiff   = "diff"
)

// Entry is one CVE's resolved method mapping, exactly as Phase 4's
// cves.json/vulnmap.json shape is described in the plan.
type Entry struct {
	Methods []string `json:"methods"`
	Source  string   `json:"source"`
}

const resolveWorkers = 8

// Resolver maps CVEs to vulnerable methods using the manual table first
// and a GitHub fix-commit diff second.
type Resolver struct {
	Manual     map[string][]string
	HTTPClient *http.Client
	Cache      *httpcache.Cache

	// GithubBaseURL/RawBaseURL default to the real github.com/
	// raw.githubusercontent.com and only need overriding in tests, so a
	// fake HTTP server can stand in for both.
	GithubBaseURL string
	RawBaseURL    string
}

// NewResolver builds a Resolver from an already-loaded manual table (see
// LoadManual); httpClient may be nil to use http.DefaultClient.
func NewResolver(manual map[string][]string, cache *httpcache.Cache) *Resolver {
	return &Resolver{
		Manual:        manual,
		HTTPClient:    http.DefaultClient,
		Cache:         cache,
		GithubBaseURL: "https://github.com",
		RawBaseURL:    "https://raw.githubusercontent.com",
	}
}

// Resolve maps one CVE to its vulnerable methods, trying the CVE's own id
// and every alias against the manual table before falling back to its
// fix-commit diffs. ok is false when neither source found anything —
// internal/reach treats that as status UNKNOWN.
func (r *Resolver) Resolve(ctx context.Context, cve cves.CVE) (Entry, bool, error) {
	for _, id := range append([]string{cve.ID}, cve.Aliases...) {
		if methods, ok := r.Manual[id]; ok && len(methods) > 0 {
			return Entry{Methods: methods, Source: SourceManual}, true, nil
		}
	}

	methods, err := r.resolveFromDiffs(ctx, cve)
	if err != nil {
		return Entry{}, false, err
	}
	if len(methods) == 0 {
		return Entry{}, false, nil
	}
	sort.Strings(methods)
	return Entry{Methods: methods, Source: SourceDiff}, true, nil
}

// ResolveAll resolves every CVE in list, in parallel (fix-commit/raw-file
// fetches are the slow part, and are independent per CVE). The result
// only has an entry for a CVE that resolved to at least one method; a CVE
// with no entry is UNKNOWN as far as vulnmap is concerned.
func (r *Resolver) ResolveAll(ctx context.Context, list []cves.CVE) (map[string]Entry, error) {
	type result struct {
		id    string
		entry Entry
		ok    bool
		err   error
	}
	jobs := make(chan cves.CVE)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < resolveWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for cve := range jobs {
				entry, ok, err := r.Resolve(ctx, cve)
				results <- result{id: cve.ID, entry: entry, ok: ok, err: err}
			}
		}()
	}
	go func() {
		for _, c := range list {
			jobs <- c
		}
		close(jobs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	out := make(map[string]Entry)
	var firstErr error
	for res := range results {
		if res.err != nil && firstErr == nil {
			firstErr = res.err
			continue
		}
		if res.ok {
			out[res.id] = res.entry
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}
