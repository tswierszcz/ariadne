package vulnmap

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"sync"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

const (
	SourceManual = "manual"
	SourceDiff   = "diff"
	SourceLLM    = "llm"
)

// Entry is one CVE's resolved method mapping, exactly as Phase 4's
// cves.json/vulnmap.json shape is described in the plan.
type Entry struct {
	Methods []string `json:"methods"`
	Source  string   `json:"source"`
	// Model and Evidence are set only for Source == SourceLLM: the model
	// that answered, and whether it saw a fix diff (EvidenceDiff) or the
	// advisory text alone (EvidenceAdvisory).
	Model    string `json:"model,omitempty"`
	Evidence string `json:"evidence,omitempty"`
	// Role/Coverage/Review are the mapping's structure (CVE:F1b,
	// docs/RESULTS_CONTRACT.MD): manual entries carry what
	// vuln_methods.yaml declares; a diff mapping is vulnerable/partial/
	// unreviewed (methods changed by a fix are never assumed complete);
	// an LLM mapping is vulnerable/unknown/unreviewed. Empty in an entry
	// read from an older cache file; WithDefaults fills it in.
	Role     string `json:"role,omitempty"`
	Coverage string `json:"coverage,omitempty"`
	Review   string `json:"review,omitempty"`
	// SourceHashes are the SHA-256 of every fetched body (fix diff, raw
	// file, LLM record) this mapping was derived from, sorted (CVE:F3
	// derivation key); Reused marks an LLM answer given for an earlier
	// version of the advisory (CVE:F2 reuse rule).
	SourceHashes []string `json:"-"`
	Reused       bool     `json:"-"`
	// Variants are a versioned manual mapping's method lists (CVE:F4):
	// the engine chooses one per component instance; Methods is their
	// union. MethodRoles are an LLM mapping's per-method roles
	// ("vulnerable" / "entry_point").
	Variants    []Variant         `json:"-"`
	MethodRoles map[string]string `json:"-"`
	// FixCommits is a diff mapping's per-commit evidence (CVE:F5): base
	// revision, files, and which methods were changed, removed or added
	// by the fix. Only changed and removed methods are in Methods.
	FixCommits []FixCommit `json:"fix_commits,omitempty"`
	// SourceError is set, with no Methods, when a CVE had fix-commit
	// references but none could be read and no other source mapped it
	// (reason SOURCE_UNAVAILABLE).
	SourceError string `json:"source_error,omitempty"`
}

// Mapped reports whether the entry names at least one method.
func (e Entry) Mapped() bool { return len(e.Methods) > 0 }

// ResolverVersion versions the mapping resolution (manual → diff → LLM
// order, diff attribution, grounding). Bump it on any change that can
// change a mapping for the same inputs; it is part of every derivation
// key (CVE:F3).
const ResolverVersion = "5" // 2: CVE:F4 versioned manual variants, LLM per-method roles; 3: CVE:F5a both sides of a fix diff; 4: CVE:F5b bytecode verification, nested classes, release fit; 5: CVE:F5c release line from git tags

// PromptVersion is the LLM prompt version (part of derivation keys).
const PromptVersion = promptVersion

type evidenceKey struct{}

// evidence collects, per Resolve call, the hashes of the bodies used.
type evidence struct {
	mu     sync.Mutex
	hashes map[string]bool
	reused bool
}

func noteEvidence(ctx context.Context, body []byte) {
	if ev, ok := ctx.Value(evidenceKey{}).(*evidence); ok {
		ev.mu.Lock()
		ev.hashes[sha256Hex(body)] = true
		ev.mu.Unlock()
	}
}

func sha256Hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func noteReused(ctx context.Context) {
	if ev, ok := ctx.Value(evidenceKey{}).(*evidence); ok {
		ev.mu.Lock()
		ev.reused = true
		ev.mu.Unlock()
	}
}

// WithDefaults fills Role/Coverage/Review from Source when they are empty.
func (e Entry) WithDefaults() Entry {
	switch e.Source {
	case SourceDiff:
		e.Role, e.Coverage, e.Review = orDefault(e.Role, RoleVulnerable), orDefault(e.Coverage, CoveragePartial), orDefault(e.Review, ReviewUnreviewed)
	case SourceLLM:
		e.Role, e.Coverage, e.Review = orDefault(e.Role, RoleVulnerable), orDefault(e.Coverage, CoverageUnknown), orDefault(e.Review, ReviewUnreviewed)
	case SourceManual:
		e.Coverage, e.Review = orDefault(e.Coverage, CoveragePartial), orDefault(e.Review, ReviewReviewed)
	}
	return e
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

const resolveWorkers = 8

// Resolver maps CVEs to vulnerable methods using the manual table first, a
// GitHub fix-commit diff second, and (when LLM is set) an LLM third.
type Resolver struct {
	Manual map[string][]string
	// ManualMeta is each manual entry's role/coverage/review
	// (LoadManualMeta); nil leaves the legacy defaults.
	ManualMeta map[string]MappingMeta
	HTTPClient *http.Client
	Cache      *httpcache.Cache

	// GithubBaseURL/RawBaseURL/GithubAPIBaseURL default to the real
	// github.com/raw.githubusercontent.com/api.github.com and only need
	// overriding in tests, so a fake HTTP server can stand in for all.
	GithubBaseURL    string
	RawBaseURL       string
	GithubAPIBaseURL string
	// GithubToken, when set, authenticates commit-metadata requests
	// (higher rate limit). It is sent, never cached.
	GithubToken string

	// LLM enables source 3; nil leaves it disabled. MethodExists grounds
	// the model's answer in the analyzed bytecode (nil skips the check).
	LLM          *LLMSource
	MethodExists func(fqcn, method string) bool
	// Bytecode, when set, verifies fix-diff candidates in the analyzed
	// classes (CVE:F5b): the descriptors of the methods named name that
	// class fqcn declares, and whether the class is present at all.
	Bytecode func(fqcn, name string) (descs []string, classFound bool)
}

// NewResolver builds a Resolver from an already-loaded manual table (see
// LoadManual); httpClient may be nil to use http.DefaultClient.
func NewResolver(manual map[string][]string, cache *httpcache.Cache) *Resolver {
	return &Resolver{
		Manual:           manual,
		HTTPClient:       http.DefaultClient,
		Cache:            cache,
		GithubBaseURL:    "https://github.com",
		RawBaseURL:       "https://raw.githubusercontent.com",
		GithubAPIBaseURL: "https://api.github.com",
	}
}

// Resolve maps one CVE to its vulnerable methods, trying the CVE's own id
// and every alias against the manual table before falling back to its
// fix-commit diffs, then (if enabled) asking the LLM. ok is false when no
// source found anything — internal/reach treats that as status UNKNOWN.
func (r *Resolver) Resolve(ctx context.Context, cve cves.CVE) (Entry, bool, error) {
	ev := &evidence{hashes: map[string]bool{}}
	e, ok, err := r.resolve(context.WithValue(ctx, evidenceKey{}, ev), cve)
	if ok {
		for h := range ev.hashes {
			e.SourceHashes = append(e.SourceHashes, h)
		}
		sort.Strings(e.SourceHashes)
		e.Reused = ev.reused
	}
	return e, ok, err
}

func (r *Resolver) resolve(ctx context.Context, cve cves.CVE) (Entry, bool, error) {
	for _, id := range append([]string{cve.ID}, cve.Aliases...) {
		if methods, ok := r.Manual[id]; ok && len(methods) > 0 {
			meta := r.ManualMeta[id]
			return Entry{Methods: methods, Source: SourceManual, Role: meta.Role, Coverage: meta.Coverage, Review: meta.Review, Variants: meta.Variants}.WithDefaults(), true, nil
		}
	}

	methods, commits, err := r.resolveFromDiffs(ctx, cve)
	if err != nil {
		return Entry{}, false, err
	}
	if len(methods) > 0 {
		sort.Strings(methods)
		return Entry{Methods: methods, Source: SourceDiff, FixCommits: commits}.WithDefaults(), true, nil
	}

	if r.LLM != nil {
		e, ok, err := r.resolveFromLLM(ctx, cve, commits)
		if ok || err != nil {
			return e, ok, err
		}
	}
	// No mapping. When every fix commit failed to be read, say so
	// (SOURCE_UNAVAILABLE) rather than leave a bare "no mapping".
	if len(commits) > 0 && allFailed(commits) {
		msgs := make([]string, len(commits))
		for i, c := range commits {
			msgs[i] = c.Repo + "@" + c.SHA + ": " + c.Error
		}
		return Entry{SourceError: strings.Join(msgs, "; "), FixCommits: commits}, true, nil
	}
	return Entry{}, false, nil
}

func allFailed(commits []FixCommit) bool {
	for _, c := range commits {
		if c.Error == "" {
			return false
		}
	}
	return true
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
