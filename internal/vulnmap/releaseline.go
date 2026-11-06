package vulnmap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"ariadne/internal/cves"
)

// Release line of a fix commit relative to the analyzed release (CVE:F5c),
// read from git history: the analyzed version's tag in the commit's
// repository compared with the commit (GitHub compare API).
const (
	// LineDescends: the analyzed release's tag is an ancestor of the fix
	// commit — the fix was made in history that contains this release:
	// its own line, or (in a repository merged forward) a later line.
	LineDescends = "descends_from_release"
	// LineInRelease: the fix commit is an ancestor of (or is) the tag —
	// the analyzed release already contains this commit.
	LineInRelease = "in_release"
	// LineOther: the commit and the tag diverged — the fix was made on a
	// release line that does not contain the analyzed release.
	LineOther = "other_line"
)

// Why a commit was left out of the mapping (FixCommit.Excluded).
const (
	ExcludedReleaseLine = "release_line" // LineOther while another commit is LineDescends
	ExcludedReleaseFit  = "release_fit"  // FitNonePresent (CVE:F5b)
)

// tagCandidates are the tag names tried, in order, for version v of
// artifact a in repository repo. Exact names only: a prefix match never
// picks a tag (5.3.2 must not find v5.3.20).
func tagCandidates(v, artifact, repo string) []string {
	out := []string{"v" + v, v}
	if artifact != "" {
		out = append(out, artifact+"-"+v)
	}
	if repo != "" && repo != artifact {
		out = append(out, repo+"-"+v)
	}
	out = append(out, "release-"+v, "v_"+v, "version-"+v)
	return out
}

type matchingRef struct {
	Ref string `json:"ref"`
}

type compareResult struct {
	Status     string               `json:"status"`
	BehindBy   int                  `json:"behind_by"`
	MergeBase  struct{ SHA string } `json:"merge_base_commit"`
	BaseCommit struct {
		Parents []struct{ SHA string } `json:"parents"`
	} `json:"base_commit"`
}

// releaseTag finds the tag of one analyzed version in owner/repo. It
// returns the tag, the version it belongs to, and a reason when none
// could be settled. Several versions with a tag are ambiguous: the
// mapping is per advisory, not per component instance.
func (r *Resolver) releaseTag(ctx context.Context, owner, repo string, aff []cves.Affected) (tag, version, reason string) {
	type pv struct{ version, artifact string }
	var vs []pv
	seen := map[string]bool{}
	for _, a := range aff {
		if a.Version == "" {
			continue
		}
		_, artifact, _ := strings.Cut(a.Package, ":")
		k := a.Version + "\x00" + artifact
		if !seen[k] {
			seen[k] = true
			vs = append(vs, pv{a.Version, artifact})
		}
	}
	if len(vs) == 0 {
		return "", "", "no analyzed version for the affected packages"
	}
	sort.Slice(vs, func(i, j int) bool {
		if vs[i].version != vs[j].version {
			return vs[i].version < vs[j].version
		}
		return vs[i].artifact < vs[j].artifact
	})

	found := map[string]string{} // version → tag
	var versions []string
	tried := map[string]bool{}
	for _, v := range vs {
		if _, ok := found[v.version]; ok {
			continue
		}
		for _, name := range tagCandidates(v.version, v.artifact, repo) {
			if tried[name] {
				continue
			}
			tried[name] = true
			ok, err := r.tagExists(ctx, owner, repo, name)
			if err != nil {
				return "", "", "tag lookup unavailable: " + err.Error()
			}
			if ok {
				found[v.version] = name
				versions = append(versions, v.version)
				break
			}
		}
	}
	switch len(versions) {
	case 0:
		var ver []string
		for _, v := range vs {
			ver = append(ver, v.version)
		}
		return "", "", fmt.Sprintf("no tag for analyzed version %s in %s/%s", strings.Join(dedupe(ver), ", "), owner, repo)
	case 1:
		return found[versions[0]], versions[0], ""
	default:
		return "", "", fmt.Sprintf("tags for several analyzed versions (%s): no single release to compare with", strings.Join(versions, ", "))
	}
}

func dedupe(ss []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func (r *Resolver) apiHeader() http.Header {
	hdr := http.Header{"Accept": {"application/vnd.github+json"}, "X-Github-Api-Version": {"2022-11-28"}}
	if r.GithubToken != "" {
		hdr.Set("Authorization", "Bearer "+r.GithubToken)
	}
	return hdr
}

// tagExists asks GitHub's matching-refs API, which answers 200 with a
// (possibly empty) list, so a missing tag is a cacheable answer, not an
// error.
func (r *Resolver) tagExists(ctx context.Context, owner, repo, name string) (bool, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/git/matching-refs/tags/%s", r.GithubAPIBaseURL, owner, repo, escapeRef(name))
	b, err := r.fetch(ctx, u, r.apiHeader())
	if err != nil {
		return false, err
	}
	var refs []matchingRef
	if err := json.Unmarshal(b, &refs); err != nil {
		return false, fmt.Errorf("decoding %s: %w", u, err)
	}
	for _, ref := range refs {
		if ref.Ref == "refs/tags/"+name {
			return true, nil
		}
	}
	return false, nil
}

// escapeRef path-escapes each "/"-separated segment of a ref name.
func escapeRef(name string) string {
	parts := strings.Split(name, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

// NoteOffBranchTag explains a LineDescends commit whose release tag
// points at a release commit made off the branch (e.g. Apache Tomcat's
// "Tag x.y.z" commits): the tag's only commit missing from the fix
// commit's history is the tagged commit itself, and the merge base is
// its parent.
const NoteOffBranchTag = "release tag is a commit off the branch; its parent is an ancestor of the fix commit"

// compareLine relates the fix commit sha to tag. It reads page 2 of the
// compare with one commit per page: GitHub sends the changed files only
// on the first page, and the status, counts and merge base on every page
// (≈10 KB instead of up to ≈1 MB per compare).
func (r *Resolver) compareLine(ctx context.Context, owner, repo, tag, sha string) (line, note string, err error) {
	u := fmt.Sprintf("%s/repos/%s/%s/compare/%s...%s?per_page=1&page=2", r.GithubAPIBaseURL, owner, repo, escapeRef(tag), sha)
	b, err := r.fetch(ctx, u, r.apiHeader())
	if err != nil {
		return "", "", err
	}
	var c compareResult
	if err := json.Unmarshal(b, &c); err != nil {
		return "", "", fmt.Errorf("decoding %s: %w", u, err)
	}
	switch c.Status {
	case "ahead":
		return LineDescends, "", nil
	case "behind", "identical":
		return LineInRelease, "", nil
	case "diverged":
		if c.BehindBy == 1 && c.MergeBase.SHA != "" {
			for _, p := range c.BaseCommit.Parents {
				if p.SHA == c.MergeBase.SHA {
					return LineDescends, NoteOffBranchTag, nil
				}
			}
		}
		return LineOther, "", nil
	default:
		return "", "", fmt.Errorf("%s: unexpected compare status %q", u, c.Status)
	}
}

// noteReleaseLine sets each readable commit's release line. Git history
// is an enrichment like commit metadata: a lookup that fails (rate limit,
// an offline cache without it, no tag) leaves the line unknown with a
// reason — never a global error, never an exclusion. A commit on another
// line is excluded only when another commit of the same advisory
// descends from the analyzed release.
func (r *Resolver) noteReleaseLine(ctx context.Context, cve cves.CVE, commits []FixCommit) {
	tags := map[string][3]string{} // repo → tag, version, reason
	for i := range commits {
		c := &commits[i]
		if c.Error != "" {
			continue
		}
		owner, repo, _ := strings.Cut(c.Repo, "/")
		t, ok := tags[c.Repo]
		if !ok {
			tag, ver, reason := r.releaseTag(ctx, owner, repo, cve.Affected)
			t = [3]string{tag, ver, reason}
			tags[c.Repo] = t
		}
		if t[0] == "" {
			c.ReleaseLineReason = t[2]
			continue
		}
		c.ReleaseTag = t[0]
		line, note, err := r.compareLine(ctx, owner, repo, t[0], c.SHA)
		if err != nil {
			c.ReleaseLineReason = "compare unavailable: " + err.Error()
			continue
		}
		c.ReleaseLine, c.ReleaseLineReason = line, note
	}
	descends := false
	for _, c := range commits {
		if c.ReleaseLine == LineDescends {
			descends = true
		}
	}
	if !descends {
		return
	}
	for i := range commits {
		if commits[i].ReleaseLine == LineOther {
			commits[i].Excluded = ExcludedReleaseLine
		}
	}
}
