package vulnmap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
	"ariadne/internal/javadiff"
)

// githubCommitRe matches a GitHub commit reference URL, e.g.
// "https://github.com/x-stream/xstream/commit/abc123...", tolerating a
// trailing "#diff-..." fragment some advisories append.
var githubCommitRe = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/commit/([0-9a-fA-F]{7,40})`)

type commitRef struct {
	owner, repo, sha string
}

// Fix-commit bases (CVE:F5): which pre-fix revision the old side of a
// commit was read from.
const (
	// BaseParent: the commit's single parent — both sides were read.
	BaseParent = "parent"
	// BaseNewSideOnly: only the post-fix revision could be read (commit
	// metadata unavailable, merge or root commit). Methods the fix removed
	// are missing and helpers it added may be included.
	BaseNewSideOnly = "new_side_only"
)

// Method kinds of a fix-commit file (CVE:F5).
const (
	FixMethodChanged = "changed" // declared before and after the fix, and touched by it
	FixMethodRemoved = "removed" // declared before the fix only
	FixMethodAdded   = "added"   // declared after the fix only: not code of the vulnerable release
)

// FixCommit is the evidence one fix-commit reference contributed to a
// diff mapping (CVE:F5): the commit, the revision its old side was read
// from, and per file which methods it touched on which side.
type FixCommit struct {
	Repo       string    `json:"repo"`             // owner/name
	SHA        string    `json:"sha"`              // full SHA when commit metadata was read
	Parent     string    `json:"parent,omitempty"` // the base revision, for BaseParent
	Base       string    `json:"base"`             // BaseParent | BaseNewSideOnly
	BaseReason string    `json:"base_reason,omitempty"`
	Files      []FixFile `json:"files,omitempty"`
	// Error is why the commit's diff itself could not be obtained; such
	// a commit contributes no methods.
	Error string `json:"error,omitempty"`
}

// FixFile is one non-test .java file of a fix commit.
type FixFile struct {
	OldPath string `json:"old_path,omitempty"`
	NewPath string `json:"new_path,omitempty"`
	Change  string `json:"change"`
	// OldSHA256/NewSHA256 hash the source read for each side.
	OldSHA256 string `json:"old_sha256,omitempty"`
	NewSHA256 string `json:"new_sha256,omitempty"`
	// Methods maps "fqcn#name" to its kind (FixMethodChanged/Removed/
	// Added); in a new-side-only file every touched method is "changed"
	// because the old side is unknown.
	Methods map[string]string `json:"methods,omitempty"`
	// NewSideOnly marks a file whose old side could not be read although
	// the commit's base was known.
	NewSideOnly bool   `json:"new_side_only,omitempty"`
	Unavailable string `json:"unavailable,omitempty"` // why neither side could be read
}

// NewSideOnly reports whether any commit or file of a diff mapping was
// read on the post-fix side only.
func (e Entry) NewSideOnly() bool {
	for _, c := range e.FixCommits {
		if c.Error != "" {
			continue
		}
		if c.Base == BaseNewSideOnly {
			return true
		}
		for _, f := range c.Files {
			if f.NewSideOnly {
				return true
			}
		}
	}
	return false
}

// fixCommits extracts every GitHub commit reference among a CVE's "FIX"
// references (and, since OSV doesn't always tag a GitHub commit URL as
// type FIX specifically, among every reference — a non-fix reference
// that happens to be a commit link is still worth trying).
func fixCommits(cve cves.CVE) []commitRef {
	var refs []commitRef
	seen := make(map[commitRef]bool)
	for _, r := range cve.References {
		m := githubCommitRe.FindStringSubmatch(r.URL)
		if m == nil {
			continue
		}
		c := commitRef{owner: m[1], repo: m[2], sha: m[3]}
		if !seen[c] {
			seen[c] = true
			refs = append(refs, c)
		}
	}
	return refs
}

// isContractError reports a failure of the run's global data contract
// (offline cache miss, data older than -max-data-age, a lookup a replayed
// snapshot did not record): it must stop the run, never become a
// per-CVE "source unavailable" (CVE:F5 task 10).
func isContractError(err error) bool {
	return errors.Is(err, httpcache.ErrOffline) || errors.Is(err, httpcache.ErrTooOld) || errors.Is(err, httpcache.ErrNotInSnapshot)
}

// resolveFromDiffs tries every fix-commit reference on cve in turn and
// returns the union of the pre-fix methods found across all of them (a
// CVE can list more than one fix commit, e.g. one per affected branch),
// with the per-commit evidence. A commit whose diff cannot be fetched is
// recorded and skipped; only a contract error is returned.
func (r *Resolver) resolveFromDiffs(ctx context.Context, cve cves.CVE) ([]string, []FixCommit, error) {
	var methods []string
	var commits []FixCommit
	seen := make(map[string]bool)

	for _, commit := range fixCommits(cve) {
		fc, err := r.resolveFromCommit(ctx, commit)
		if err != nil {
			return nil, nil, err
		}
		commits = append(commits, fc)
		for _, f := range fc.Files {
			for _, name := range sortedKeys(f.Methods) {
				if f.Methods[name] != FixMethodAdded && !seen[name] {
					seen[name] = true
					methods = append(methods, name)
				}
			}
		}
	}
	return methods, commits, nil
}

// commitMeta is the part of GitHub's "get a commit" response used here.
// The patch text itself still comes from the .diff, since the API may
// truncate or omit per-file patches.
type commitMeta struct {
	SHA     string `json:"sha"`
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
}

func (r *Resolver) resolveFromCommit(ctx context.Context, commit commitRef) (FixCommit, error) {
	fc := FixCommit{Repo: commit.owner + "/" + commit.repo, SHA: commit.sha}
	diffURL := fmt.Sprintf("%s/%s/%s/commit/%s.diff", r.GithubBaseURL, commit.owner, commit.repo, commit.sha)
	diffBody, err := r.fetch(ctx, diffURL, nil)
	if err != nil {
		if isContractError(err) {
			return fc, fmt.Errorf("fetching fix commit diff %s: %w", diffURL, err)
		}
		fc.Error = fmt.Sprintf("fetching diff: %v", err)
		return fc, nil
	}
	files, err := javadiff.ParseUnifiedDiff(string(diffBody))
	if err != nil {
		fc.Error = fmt.Sprintf("parsing diff: %v", err)
		return fc, nil
	}

	// The base revision: the single parent, from the commit metadata.
	// Metadata is an enrichment — when it cannot be read (rate limit,
	// offline cache without it, an older snapshot) the commit falls back
	// to the post-fix side, visibly.
	newRev := commit.sha
	fc.Base = BaseNewSideOnly
	meta, err := r.commitMeta(ctx, commit)
	switch {
	case err != nil:
		fc.BaseReason = "commit metadata unavailable: " + err.Error()
	case len(meta.Parents) == 0:
		fc.BaseReason = "root commit: no pre-fix revision"
	case len(meta.Parents) > 1:
		fc.BaseReason = fmt.Sprintf("merge commit with %d parents: no unambiguous pre-fix revision", len(meta.Parents))
	default:
		fc.Base, fc.Parent = BaseParent, meta.Parents[0].SHA
	}
	if err == nil && meta.SHA != "" {
		fc.SHA, newRev = meta.SHA, meta.SHA
	}

	for _, f := range files {
		if !isNonTestJavaFile(f.Path()) {
			continue
		}
		fc.Files = append(fc.Files, r.fixFile(ctx, commit, newRev, fc.Parent, f))
	}
	return fc, nil
}

// fixFile reads both sides of one file and classifies the methods the
// diff touched. parent is "" when only the new side can be read.
func (r *Resolver) fixFile(ctx context.Context, commit commitRef, newRev, parent string, f javadiff.FileDiff) FixFile {
	ff := FixFile{OldPath: f.OldPath, NewPath: f.NewPath, Change: string(f.Change), Methods: map[string]string{}}
	read := func(rev, path string) (string, string, bool) {
		if rev == "" || path == "" {
			return "", "", false
		}
		url := fmt.Sprintf("%s/%s/%s/%s/%s", r.RawBaseURL, commit.owner, commit.repo, rev, path)
		b, err := r.fetch(ctx, url, nil)
		if err != nil {
			// A file missing at that revision, or a private/gone repo:
			// skip this side of this file rather than the whole CVE.
			return "", "", false
		}
		return string(b), sha256Hex(b), true
	}
	newSrc, newHash, newOK := read(newRev, f.NewPath)
	oldSrc, oldHash, oldOK := read(parent, f.OldPath)
	ff.NewSHA256, ff.OldSHA256 = newHash, oldHash

	if parent == "" || (!oldOK && f.OldPath != "") {
		// Post-fix side only (the pre-CVE:F5 behavior): every method the
		// diff touched in the new file, of unknown kind.
		ff.NewSideOnly = parent != ""
		if !newOK {
			ff.Unavailable = "source not readable at the fix revision"
			ff.OldSHA256 = ""
			return ff
		}
		if fqcn, ok := javadiff.FQCN(f.NewPath, newSrc); ok {
			for _, m := range javadiff.FindMethodsForLines(newSrc, f.NewLines) {
				ff.Methods[fqcn+"#"+m] = FixMethodChanged
			}
		}
		return ff
	}
	if !newOK && f.NewPath != "" {
		ff.Unavailable = "source not readable at the fix revision"
		ff.NewSideOnly = true
		return ff
	}

	oldFQCN, _ := javadiff.FQCN(f.OldPath, oldSrc)
	newFQCN, _ := javadiff.FQCN(f.NewPath, newSrc)
	oldDecl, newDecl := map[string]bool{}, map[string]bool{}
	if oldOK {
		oldDecl = javadiff.DeclaredMethods(oldSrc)
	}
	if newOK {
		newDecl = javadiff.DeclaredMethods(newSrc)
	}
	if oldOK && oldFQCN != "" {
		for _, m := range javadiff.FindMethodsForLines(oldSrc, f.OldLines) {
			kind := FixMethodRemoved
			if newDecl[m] {
				kind = FixMethodChanged
			}
			ff.Methods[oldFQCN+"#"+m] = kind
		}
	}
	if newOK && newFQCN != "" {
		for _, m := range javadiff.FindMethodsForLines(newSrc, f.NewLines) {
			if oldDecl[m] && oldFQCN != "" {
				// Present before the fix: the pre-fix class is the one in
				// the vulnerable release (it differs after a rename).
				ff.Methods[oldFQCN+"#"+m] = FixMethodChanged
			} else if _, dup := ff.Methods[newFQCN+"#"+m]; !dup {
				ff.Methods[newFQCN+"#"+m] = FixMethodAdded
			}
		}
	}
	return ff
}

func (r *Resolver) commitMeta(ctx context.Context, commit commitRef) (commitMeta, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s", r.GithubAPIBaseURL, commit.owner, commit.repo, commit.sha)
	hdr := http.Header{"Accept": {"application/vnd.github+json"}, "X-Github-Api-Version": {"2022-11-28"}}
	if r.GithubToken != "" {
		hdr.Set("Authorization", "Bearer "+r.GithubToken)
	}
	b, err := r.fetch(ctx, url, hdr)
	if err != nil {
		return commitMeta{}, err
	}
	var m commitMeta
	if err := json.Unmarshal(b, &m); err != nil {
		return commitMeta{}, fmt.Errorf("decoding %s: %w", url, err)
	}
	if m.SHA != "" && !strings.HasPrefix(strings.ToLower(m.SHA), strings.ToLower(commit.sha)) {
		return commitMeta{}, fmt.Errorf("%s answered for commit %s", url, m.SHA)
	}
	return m, nil
}

// fetch GETs url through the cache. hdr is sent but never stored (the
// token stays out of the cache).
func (r *Resolver) fetch(ctx context.Context, url string, hdr http.Header) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range hdr {
		req.Header[k] = v
	}
	b, err := httpcache.Fetch(r.HTTPClient, r.Cache, httpcache.Key("GET", url), httpcache.KindForURL(url), req)
	if err == nil {
		noteEvidence(ctx, b)
	}
	return b, err
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
