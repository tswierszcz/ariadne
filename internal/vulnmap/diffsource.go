package vulnmap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
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
	refType          string // the OSV reference type: "FIX" when any reference to the commit is one
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
	Repo       string `json:"repo"`             // owner/name
	SHA        string `json:"sha"`              // full SHA when commit metadata was read
	Parent     string `json:"parent,omitempty"` // the base revision, for BaseParent
	Base       string `json:"base"`             // BaseParent | BaseNewSideOnly
	BaseReason string `json:"base_reason,omitempty"`
	// RefType is the advisory's reference type for this commit (CVE:F5b):
	// "FIX" for a fix reference, otherwise the type of another link
	// ("WEB", "ADVISORY", …) — a weaker kind of evidence.
	RefType string `json:"ref_type,omitempty"`
	// ReleaseFit says whether this commit's candidates exist in the
	// analyzed classes (CVE:F5b): FitPresent, FitNonePresent (none does
	// although another commit's do — its methods are left out of the
	// mapping), FitNoCandidates, or "" when not checked.
	ReleaseFit string    `json:"release_fit,omitempty"`
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
	// Methods are the methods the diff touched, in file order; in a
	// new-side-only file every one is "changed" because the old side is
	// unknown.
	Methods []FixMethod `json:"methods,omitempty"`
	// NewSideOnly marks a file whose old side could not be read although
	// the commit's base was known.
	NewSideOnly bool   `json:"new_side_only,omitempty"`
	Unavailable string `json:"unavailable,omitempty"` // why neither side could be read
}

// Bytecode verification of a candidate in the analyzed classes (CVE:F5b).
const (
	BytecodeExact         = "exact"           // one overload matches the source parameters: the selector carries its descriptor
	BytecodeAmbiguous     = "ambiguous"       // several overloads match the source parameters: name selector
	BytecodeParamMismatch = "params_mismatch" // overloads exist, none matches the source parameters: name selector
	BytecodeMethodMissing = "method_missing"  // the class is present, no method of that name
	BytecodeClassMissing  = "class_missing"   // the class is not among the analyzed classes
)

// Release fit of a fix commit (CVE:F5b).
const (
	FitPresent      = "candidates_present"
	FitNonePresent  = "no_candidate_present"
	FitNoCandidates = "no_candidates"
)

// FixMethod is one method a fix commit touched.
type FixMethod struct {
	// Selector is "fqcn#name", or "fqcn#name(params)ret" once bytecode
	// verification found exactly one matching overload.
	Selector string `json:"selector"`
	Kind     string `json:"kind"` // FixMethodChanged | FixMethodRemoved | FixMethodAdded
	// SourceParams are the parameter types read from the source (simple
	// names; "?" for a type variable).
	SourceParams []string `json:"source_params,omitempty"`
	// Bytecode is the verification outcome (Bytecode* constants); empty
	// when not verified (added methods, no analyzed classes).
	Bytecode string `json:"bytecode,omitempty"`
	// PossibleRename names, for a removed method, an added method of the
	// same class with a similar body, and vice versa (CVE:F5b heuristic,
	// evidence only).
	PossibleRename string `json:"possible_rename,omitempty"`

	// key is the source identity (Class#name(params)), for matching
	// sides and rename pairs.
	key   string
	body  map[string]bool
	class string
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
	seen := make(map[string]int)
	for _, r := range cve.References {
		m := githubCommitRe.FindStringSubmatch(r.URL)
		if m == nil {
			continue
		}
		key := m[1] + "/" + m[2] + "@" + m[3]
		if i, ok := seen[key]; ok {
			if r.Type == "FIX" {
				refs[i].refType = "FIX"
			}
			continue
		}
		seen[key] = len(refs)
		refs = append(refs, commitRef{owner: m[1], repo: m[2], sha: m[3], refType: r.Type})
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
	var commits []FixCommit
	for _, commit := range fixCommits(cve) {
		fc, err := r.resolveFromCommit(ctx, commit)
		if err != nil {
			return nil, nil, err
		}
		commits = append(commits, fc)
	}
	r.noteReleaseFit(commits)

	var methods []string
	seen := make(map[string]bool)
	for _, fc := range commits {
		if fc.ReleaseFit == FitNonePresent {
			continue
		}
		for _, f := range fc.Files {
			for _, m := range f.Methods {
				if m.Kind != FixMethodAdded && !seen[m.Selector] {
					seen[m.Selector] = true
					methods = append(methods, m.Selector)
				}
			}
		}
	}
	return collapseSelectors(methods), commits, nil
}

// collapseSelectors drops an exact selector ("C#m(desc)") when the same
// list also has the name selector "C#m", which already matches every
// overload (another commit or side could not narrow it): the broader one
// wins, never the guess.
func collapseSelectors(ms []string) []string {
	byName := map[string]bool{}
	for _, m := range ms {
		if !strings.Contains(m, "(") {
			byName[m] = true
		}
	}
	out := ms[:0]
	for _, m := range ms {
		if i := strings.IndexByte(m, '('); i >= 0 && byName[m[:i]] {
			continue
		}
		out = append(out, m)
	}
	return out
}

// noteReleaseFit sets each commit's ReleaseFit from its candidates'
// bytecode verification. A commit none of whose candidates exists in the
// analyzed classes is FitNonePresent — and left out of the mapping — only
// when another commit of the same advisory has a present candidate: the
// fix then evidently targets another release line. When no commit has
// one, every candidate stays (the engine reports METHOD_NOT_FOUND).
func (r *Resolver) noteReleaseFit(commits []FixCommit) {
	if r.Bytecode == nil {
		return
	}
	anyPresent := false
	for i := range commits {
		c := &commits[i]
		if c.Error != "" {
			continue
		}
		cands, present := 0, 0
		for _, f := range c.Files {
			for _, m := range f.Methods {
				if m.Kind == FixMethodAdded {
					continue
				}
				cands++
				if m.Bytecode != BytecodeClassMissing && m.Bytecode != BytecodeMethodMissing {
					present++
				}
			}
		}
		switch {
		case cands == 0:
			c.ReleaseFit = FitNoCandidates
		case present > 0:
			c.ReleaseFit, anyPresent = FitPresent, true
		default:
			c.ReleaseFit = FitNonePresent
		}
	}
	if !anyPresent {
		for i := range commits {
			if commits[i].ReleaseFit == FitNonePresent {
				commits[i].ReleaseFit = "" // nothing to compare against: no conclusion
			}
		}
	}
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
	fc := FixCommit{Repo: commit.owner + "/" + commit.repo, SHA: commit.sha, RefType: commit.refType}
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
	ff := FixFile{OldPath: f.OldPath, NewPath: f.NewPath, Change: string(f.Change)}
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
		ff.OldSHA256 = ""
		if !newOK {
			ff.Unavailable = "source not readable at the fix revision"
			return ff
		}
		for _, m := range javadiff.MethodsForLines(newSrc, f.NewLines) {
			ff.Methods = append(ff.Methods, fixMethod(newSrc, m, FixMethodChanged))
		}
		r.verify(ff.Methods)
		return ff
	}
	if !newOK && f.NewPath != "" {
		ff.Unavailable = "source not readable at the fix revision"
		ff.NewSideOnly = true
		return ff
	}

	oldDecl, newDecl := map[string]bool{}, map[string]bool{}
	for _, m := range javadiff.Declared(oldSrc) {
		oldDecl[m.Key()] = true
	}
	for _, m := range javadiff.Declared(newSrc) {
		newDecl[m.Key()] = true
	}
	seen := map[string]bool{}
	if oldOK {
		for _, m := range javadiff.MethodsForLines(oldSrc, f.OldLines) {
			kind := FixMethodRemoved
			if newDecl[m.Key()] {
				kind = FixMethodChanged
			}
			seen[m.Key()] = true
			ff.Methods = append(ff.Methods, fixMethod(oldSrc, m, kind))
		}
	}
	if newOK {
		for _, m := range javadiff.MethodsForLines(newSrc, f.NewLines) {
			if seen[m.Key()] {
				continue
			}
			seen[m.Key()] = true
			kind := FixMethodAdded
			src := newSrc
			if oldDecl[m.Key()] {
				// Present before the fix: the pre-fix declaration is the
				// one in the vulnerable release (its class may differ after
				// a rename).
				kind, src = FixMethodChanged, oldSrc
				for _, om := range javadiff.Declared(oldSrc) {
					if om.Key() == m.Key() {
						m = om
						break
					}
				}
			}
			ff.Methods = append(ff.Methods, fixMethod(src, m, kind))
		}
	}
	notePossibleRenames(ff.Methods)
	r.verify(ff.Methods)
	return ff
}

func fixMethod(src string, m javadiff.MethodRef, kind string) FixMethod {
	cls := m.Class
	if pkg := javadiff.Package(src); pkg != "" {
		cls = pkg + "." + cls
	}
	return FixMethod{Selector: cls + "#" + m.Name, Kind: kind, SourceParams: m.Params,
		key: m.Key(), class: cls, body: bodyTokens(src, m)}
}

// verify checks every non-added candidate against the analyzed classes
// and narrows its selector to a descriptor when exactly one overload
// matches the source parameters. It never guesses: no match or several
// keep the name selector.
func (r *Resolver) verify(ms []FixMethod) {
	if r.Bytecode == nil {
		return
	}
	for i := range ms {
		m := &ms[i]
		if m.Kind == FixMethodAdded {
			continue
		}
		cls, name, _ := strings.Cut(m.Selector, "#")
		descs, classFound := r.Bytecode(cls, name)
		switch {
		case !classFound:
			m.Bytecode = BytecodeClassMissing
		case len(descs) == 0:
			m.Bytecode = BytecodeMethodMissing
		default:
			var match []string
			for _, d := range descs {
				if paramsCompatible(m.SourceParams, d, name, cls) {
					match = append(match, d)
				}
			}
			switch len(match) {
			case 1:
				m.Bytecode, m.Selector = BytecodeExact, m.Selector+match[0]
			case 0:
				m.Bytecode = BytecodeParamMismatch
			default:
				m.Bytecode = BytecodeAmbiguous
			}
		}
	}
}

// paramsCompatible reports whether source parameter types (simple names,
// "?" wildcards) fit a JVM method descriptor. A constructor of a member
// or enum class may carry leading synthetic parameters (outer instance;
// enum name and ordinal); only trailing parameters are compared then.
func paramsCompatible(src []string, desc, name, cls string) bool {
	dps := descParamSimpleNames(desc)
	if len(dps) < len(src) {
		return false
	}
	if extra := len(dps) - len(src); extra > 0 {
		if name != "<init>" || extra > 2 {
			return false
		}
		dps = dps[extra:]
	}
	for i, p := range src {
		if p != "?" && strings.TrimSuffix(p, strings.Repeat("[]", strings.Count(p, "[]"))) == "?" {
			// "?[]": only the dimensions are known
			if strings.Count(p, "[]") != strings.Count(dps[i], "[]") {
				return false
			}
			continue
		}
		if p == "?" {
			continue
		}
		if p != dps[i] {
			return false
		}
	}
	return true
}

// descParamSimpleNames turns "(Ljava/util/Map$Entry;[I)V" into
// ["Entry", "int[]"].
func descParamSimpleNames(desc string) []string {
	i, j := strings.IndexByte(desc, '('), strings.IndexByte(desc, ')')
	if i < 0 || j < i {
		return nil
	}
	ps := desc[i+1 : j]
	var out []string
	for k := 0; k < len(ps); {
		dims := 0
		for k < len(ps) && ps[k] == '[' {
			dims++
			k++
		}
		if k >= len(ps) {
			break
		}
		var t string
		if ps[k] == 'L' {
			end := strings.IndexByte(ps[k:], ';')
			if end < 0 {
				return out
			}
			t = ps[k+1 : k+end]
			k += end + 1
			if x := strings.LastIndexAny(t, "/$"); x >= 0 {
				t = t[x+1:]
			}
		} else {
			t = primitiveNames[ps[k]]
			k++
		}
		out = append(out, t+strings.Repeat("[]", dims))
	}
	return out
}

var primitiveNames = map[byte]string{'Z': "boolean", 'B': "byte", 'C': "char", 'S': "short", 'I': "int", 'J': "long", 'F': "float", 'D': "double"}

var identRe = regexp.MustCompile(`[A-Za-z_]\w*`)

func bodyTokens(src string, m javadiff.MethodRef) map[string]bool {
	if m.StartLine <= 0 || m.EndLine < m.StartLine {
		return nil
	}
	lines := strings.Split(src, "\n")
	if m.EndLine > len(lines) {
		return nil
	}
	out := map[string]bool{}
	for _, t := range identRe.FindAllString(strings.Join(lines[m.StartLine:m.EndLine], "\n"), -1) {
		out[t] = true
	}
	return out
}

// notePossibleRenames pairs a removed method with an added method of the
// same class whose body shares most identifiers (Jaccard >= 0.6, both
// bodies non-trivial). A heuristic, recorded as evidence only: it never
// changes the mapping, and an ambiguous pairing (several candidates) is
// not recorded.
func notePossibleRenames(ms []FixMethod) {
	const threshold = 0.6
	for i := range ms {
		if ms[i].Kind != FixMethodRemoved || len(ms[i].body) < 3 {
			continue
		}
		best, n := -1, 0
		for j := range ms {
			if ms[j].Kind != FixMethodAdded || ms[j].class != ms[i].class || len(ms[j].body) < 3 {
				continue
			}
			if jaccard(ms[i].body, ms[j].body) >= threshold {
				best, n = j, n+1
			}
		}
		if n == 1 {
			ms[i].PossibleRename, ms[best].PossibleRename = ms[best].Selector, ms[i].Selector
		}
	}
}

func jaccard(a, b map[string]bool) float64 {
	inter := 0
	for k := range a {
		if b[k] {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
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
