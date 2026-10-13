package vulnmap

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

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

// resolveFromDiffs tries every fix-commit reference on cve in turn and
// returns the union of methods found across all of them (a CVE can list
// more than one fix commit, e.g. one per affected branch).
func (r *Resolver) resolveFromDiffs(ctx context.Context, cve cves.CVE) ([]string, error) {
	var methods []string
	seen := make(map[string]bool)

	for _, commit := range fixCommits(cve) {
		m, err := r.resolveFromCommit(ctx, commit)
		if err != nil {
			return nil, err
		}
		for _, name := range m {
			if !seen[name] {
				seen[name] = true
				methods = append(methods, name)
			}
		}
	}
	return methods, nil
}

func (r *Resolver) resolveFromCommit(ctx context.Context, commit commitRef) ([]string, error) {
	diffURL := fmt.Sprintf("%s/%s/%s/commit/%s.diff", r.GithubBaseURL, commit.owner, commit.repo, commit.sha)
	diffBody, err := r.fetch(ctx, diffURL)
	if err != nil {
		return nil, fmt.Errorf("fetching fix commit diff %s: %w", diffURL, err)
	}

	files, err := javadiff.ParseUnifiedDiff(string(diffBody))
	if err != nil {
		return nil, fmt.Errorf("parsing diff %s: %w", diffURL, err)
	}

	var methods []string
	for _, f := range files {
		if !isNonTestJavaFile(f.Path) {
			continue
		}
		rawURL := fmt.Sprintf("%s/%s/%s/%s/%s", r.RawBaseURL, commit.owner, commit.repo, commit.sha, f.Path)
		src, err := r.fetch(ctx, rawURL)
		if err != nil {
			// The file may have been renamed/deleted by a later commit on
			// the default branch that raw.githubusercontent.com's cache
			// doesn't reflect, or the repo could be private/gone; either
			// way, skip this one file rather than failing the whole CVE.
			continue
		}
		fqcn, ok := javadiff.FQCN(f.Path, string(src))
		if !ok {
			continue
		}
		for _, name := range javadiff.FindMethodsForLines(string(src), f.ChangedLines) {
			methods = append(methods, fqcn+"#"+name)
		}
	}
	return methods, nil
}

func (r *Resolver) fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return httpcache.Fetch(r.HTTPClient, r.Cache, httpcache.Key("GET", url), req)
}
