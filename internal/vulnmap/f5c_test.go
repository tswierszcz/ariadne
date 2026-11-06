package vulnmap

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

// CVE:F5c fixtures: a second fix commit (another release line) beside the
// standard one, tags and the compare API.

const (
	f5cOther       = "5555555555555555555555555555555555555555"
	f5cOtherParent = "6666666666666666666666666666666666666666"
)

func (s *f5Server) otherCommit() {
	const otherDiff = `diff --git a/src/main/java/q/Legacy.java b/src/main/java/q/Legacy.java
index 1111111..2222222 100644
--- a/src/main/java/q/Legacy.java
+++ b/src/main/java/q/Legacy.java
@@ -2,3 +2,3 @@
 public class Legacy {
-    void parse() { bad(); }
+    void parse() { good(); }
 }
`
	s.serve("/o/r/commit/"+f5cOther+".diff", otherDiff)
	s.serve("/o/r/"+f5cOtherParent+"/src/main/java/q/Legacy.java", "package q;\npublic class Legacy {\n    void parse() { bad(); }\n}\n")
	s.serve("/o/r/"+f5cOther+"/src/main/java/q/Legacy.java", "package q;\npublic class Legacy {\n    void parse() { good(); }\n}\n")
	s.api("o/r", f5cOther, f5cOtherParent)
}

// tags serves matching-refs answers: name → refs returned (a prefix query
// may return longer names, which must not count as the tag).
func (s *f5Server) tags(answers map[string][]string) {
	s.mux.HandleFunc("/repos/o/r/git/matching-refs/tags/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/repos/o/r/git/matching-refs/tags/")
		var refs []string
		for _, ref := range answers[name] {
			refs = append(refs, `{"ref":"refs/tags/`+ref+`"}`)
		}
		w.Write([]byte("[" + strings.Join(refs, ",") + "]"))
	})
}

func (s *f5Server) compare(tag, sha, status string) {
	s.compareBody(tag, sha, `{"status":"`+status+`","ahead_by":1,"behind_by":0}`)
}

// compareBody serves a compare answer; only page 2 (no file list) is
// served, so a query for page 1 fails the test.
func (s *f5Server) compareBody(tag, sha, body string) {
	s.mux.HandleFunc("/repos/o/r/compare/"+tag+"..."+sha, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("per_page") != "1" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write([]byte(body))
	})
}

func f5cCVE(id string) cves.CVE {
	c := f5CVE(id, f5Fix, f5cOther)
	c.Affected = []cves.Affected{{Package: "org.o:lib", Version: "1.2.3"}}
	return c
}

func lines(e Entry) []string {
	var out []string
	for _, c := range e.FixCommits {
		out = append(out, c.ReleaseLine+"|"+c.Excluded)
	}
	return out
}

// TestReleaseLine_otherLineExcluded: the tag is found by exact name only
// (v1.2.30 for the "v1.2.3" prefix query does not count; "1.2.3" does);
// a commit descending from it keeps its methods, a diverged one is left
// out with its reason.
func TestReleaseLine_otherLineExcluded(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	s.otherCommit()
	s.tags(map[string][]string{"v1.2.3": {"v1.2.30"}, "1.2.3": {"1.2.3", "1.2.3.1"}})
	s.compare("1.2.3", f5Fix, "ahead")
	s.compare("1.2.3", f5cOther, "diverged")

	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5cCVE("CVE-F5C-1"))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if got, want := lines(e), []string{LineDescends + "|", LineOther + "|" + ExcludedReleaseLine}; !reflect.DeepEqual(got, want) {
		t.Fatalf("lines = %v, want %v", got, want)
	}
	if e.FixCommits[0].ReleaseTag != "1.2.3" {
		t.Errorf("tag = %q, want the exact 1.2.3", e.FixCommits[0].ReleaseTag)
	}
	if want := []string{"p.Parser#load", "p.Parser#unsafeLoad"}; !reflect.DeepEqual(e.Methods, want) {
		t.Errorf("Methods = %v, want %v (the other line's q.Legacy#parse left out)", e.Methods, want)
	}
}

// TestReleaseLine_noDescendingCommitKeepsAll: without a commit on the
// analyzed release's line there is nothing to prefer: a diverged commit
// stays (visible), and a commit the release already contains is evidence
// only.
func TestReleaseLine_noDescendingCommitKeepsAll(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	s.otherCommit()
	s.tags(map[string][]string{"v1.2.3": {"v1.2.3"}})
	s.compare("v1.2.3", f5Fix, "behind")
	s.compare("v1.2.3", f5cOther, "diverged")

	e, _, err := s.resolver(t, nil).Resolve(context.Background(), f5cCVE("CVE-F5C-2"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := lines(e), []string{LineInRelease + "|", LineOther + "|"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("lines = %v, want %v", got, want)
	}
	if len(e.Methods) != 3 {
		t.Errorf("Methods = %v, want all three candidates", e.Methods)
	}
}

// TestReleaseLine_unknown: no tag, several analyzed versions with a tag,
// and an offline cache without the lookups all leave the line unknown
// with a reason — never an error, never an exclusion.
func TestReleaseLine_unknown(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	s.otherCommit()
	s.tags(map[string][]string{"v1.2.3": {"v1.2.3"}, "v2.0.0": {"v2.0.0"}})
	s.compare("v1.2.3", f5Fix, "ahead")
	s.compare("v1.2.3", f5cOther, "diverged")

	cve := f5cCVE("CVE-F5C-3")
	cve.Affected = []cves.Affected{{Package: "org.o:lib", Version: "9.9"}}
	e, _, err := s.resolver(t, nil).Resolve(context.Background(), cve)
	if err != nil {
		t.Fatal(err)
	}
	c := e.FixCommits[0]
	if c.ReleaseLine != "" || !strings.Contains(c.ReleaseLineReason, "no tag for analyzed version 9.9") || len(e.Methods) != 3 {
		t.Errorf("no tag: %+v methods=%v", c, e.Methods)
	}

	cve.Affected = []cves.Affected{{Package: "org.o:lib", Version: "1.2.3"}, {Package: "org.o:lib-extra", Version: "2.0.0"}}
	e, _, _ = s.resolver(t, nil).Resolve(context.Background(), cve)
	if c := e.FixCommits[0]; c.ReleaseLine != "" || !strings.Contains(c.ReleaseLineReason, "several analyzed versions (1.2.3, 2.0.0)") {
		t.Errorf("two versions: %+v", c)
	}

	// Offline: the diff and sources are cached, the tag lookups are not.
	cache := mustCache(t)
	r := s.resolver(t, cache)
	r.GithubAPIBaseURL = s.srv.URL + "/nothing" // commit metadata and tags 404: not cached
	if _, _, err := r.Resolve(context.Background(), f5cCVE("CVE-F5C-3")); err != nil {
		t.Fatal(err)
	}
	off, err := httpcache.New(cache.Dir(), true)
	if err != nil {
		t.Fatal(err)
	}
	e, ok, err := s.resolver(t, off).Resolve(context.Background(), f5cCVE("CVE-F5C-3"))
	if err != nil || !ok {
		t.Fatalf("offline: ok=%v err=%v", ok, err)
	}
	if c := e.FixCommits[0]; c.ReleaseLine != "" || c.Excluded != "" || !strings.HasPrefix(c.ReleaseLineReason, "tag lookup unavailable") || len(e.Methods) != 3 {
		// new-side-only (no metadata): Parser's two touched methods and Legacy#parse
		t.Errorf("offline: %+v methods=%v", c, e.Methods)
	}
}

// TestReleaseLine_andReleaseFit: the diverged commit is left out by its
// line even though its candidates are present, and is then no reference
// for the fit rule: the descending commit stays although none of its
// candidates is in the analyzed classes. Against a present commit that
// is not excluded, the fit rule applies to a descending commit too
// (descending history may be a later major line).
func TestReleaseLine_andReleaseFit(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	s.otherCommit()
	s.tags(map[string][]string{"v1.2.3": {"v1.2.3"}})
	s.compare("v1.2.3", f5Fix, "ahead")
	s.compare("v1.2.3", f5cOther, "diverged")

	r := s.resolver(t, nil)
	r.Bytecode = fakeBytecode{"q.Legacy": {"parse": {"()V"}}}.fn()
	e, _, err := r.Resolve(context.Background(), f5cCVE("CVE-F5C-4"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := lines(e), []string{LineDescends + "|", LineOther + "|" + ExcludedReleaseLine}; !reflect.DeepEqual(got, want) {
		t.Fatalf("lines = %v, want %v", got, want)
	}
	if want := []string{"p.Parser#load", "p.Parser#unsafeLoad"}; !reflect.DeepEqual(e.Methods, want) {
		t.Errorf("Methods = %v, want %v", e.Methods, want)
	}

	// The other commit is in the release (not excluded) and present: the
	// descending commit with no present candidate is left out by fit.
	s3 := newF5Server(t)
	s3.standard()
	s3.api("o/r", f5Fix, f5Parent)
	s3.otherCommit()
	s3.tags(map[string][]string{"v1.2.3": {"v1.2.3"}})
	s3.compare("v1.2.3", f5Fix, "ahead")
	s3.compare("v1.2.3", f5cOther, "behind")
	r3 := s3.resolver(t, nil)
	r3.Bytecode = r.Bytecode
	e, _, _ = r3.Resolve(context.Background(), f5cCVE("CVE-F5C-4"))
	if c := e.FixCommits[0]; c.ReleaseFit != FitNonePresent || c.Excluded != ExcludedReleaseFit || !reflect.DeepEqual(e.Methods, []string{"q.Legacy#parse()V"}) {
		t.Errorf("descending commit with no present candidate: %+v methods=%v", c, e.Methods)
	}

	// Without history (no tag), the CVE:F5b fit rule decides as before.
	s2 := newF5Server(t)
	s2.standard()
	s2.api("o/r", f5Fix, f5Parent)
	s2.otherCommit()
	s2.tags(nil)
	r2 := s2.resolver(t, nil)
	r2.Bytecode = r.Bytecode
	e, _, _ = r2.Resolve(context.Background(), f5cCVE("CVE-F5C-4"))
	if e.FixCommits[0].Excluded != ExcludedReleaseFit || !reflect.DeepEqual(e.Methods, []string{"q.Legacy#parse()V"}) {
		t.Errorf("fit only: %+v methods=%v", e.FixCommits[0], e.Methods)
	}
}

// TestFixMethodsPrompt: the LLM sees both sides of the fix — changed,
// removed and added methods, presence in the analyzed classes — and a
// commit on another release line is marked in <fix_methods> and on its
// "# commit" line in the diff.
func TestFixMethodsPrompt(t *testing.T) {
	fixes := []FixCommit{
		{Repo: "example/lib", SHA: "abc1234", Base: BaseParent, ReleaseTag: "v1.2.3", ReleaseLine: LineDescends, Files: []FixFile{{Methods: []FixMethod{
			{Selector: "org.example.Parser#parse(Ljava/lang/String;)V", Kind: FixMethodChanged, SourceParams: []string{"String"}, Bytecode: BytecodeExact},
			{Selector: "org.example.Parser#unsafe", Kind: FixMethodRemoved, Bytecode: BytecodeMethodMissing},
			{Selector: "org.example.Parser#check", Kind: FixMethodAdded, SourceParams: []string{"String"}},
		}}}},
		{Repo: "example/lib", SHA: "def5678", Base: BaseNewSideOnly, ReleaseTag: "v1.2.3", ReleaseLine: LineOther, Excluded: ExcludedReleaseLine, Files: []FixFile{{Methods: []FixMethod{
			{Selector: "org.example.Legacy#parse", Kind: FixMethodChanged},
		}}}},
		{Repo: "example/lib", SHA: "0000000", Error: "fetching diff: 404"},
	}
	cve := testCVE()
	commitURL := "https://github.com/example/lib/commit/abc1234"
	var kept []fileDiff
	for _, f := range splitDiffFiles(fakeDiff) {
		if keepForLLM(f.path) {
			kept = append(kept, f)
		}
	}
	other := "https://github.com/example/lib/commit/def5678"
	d, ev, err := buildUserPrompt(cve, []string{commitURL, other}, []commitDiff{{url: commitURL, files: kept}, {url: other, note: lineNote(fixes[1]), files: kept}}, fixes)
	if err != nil || ev != EvidenceDiff {
		t.Fatalf("ev=%q err=%v", ev, err)
	}
	readGolden(t, "user_d.golden", d)

	if renderFixMethods(fixes[2:]) != "" {
		t.Error("a commit whose diff failed rendered methods")
	}
	if p, _, _ := buildUserPrompt(cve, nil, nil, nil); strings.Contains(p, "fix_methods") {
		t.Error("an empty list rendered a <fix_methods> section")
	}
}

// TestLLMCommitDiffsNotesOtherLine: the diff shown to the model marks a
// commit excluded by its release line.
func TestLLMCommitDiffsNotesOtherLine(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.otherCommit()
	r := s.resolver(t, nil)
	fixes := []FixCommit{{Repo: "o/r", SHA: f5cOther, ReleaseTag: "v1", ReleaseLine: LineOther, Excluded: ExcludedReleaseLine}}
	_, cds, err := r.llmCommitDiffs(context.Background(), f5CVE("CVE-F5C-5", f5Fix, f5cOther[:12]), fixes)
	if err != nil || len(cds) != 1 {
		// the abbreviated reference to the other commit is served under
		// its full SHA only, so just the standard commit is read here
		t.Fatalf("cds=%d err=%v", len(cds), err)
	}
	if cds[0].note != "" {
		t.Errorf("the descending commit got a note: %q", cds[0].note)
	}
	s.serve("/o/r/commit/"+f5cOther[:12]+".diff", "")
	_, cds, _ = r.llmCommitDiffs(context.Background(), f5CVE("CVE-F5C-5", f5cOther[:12]), fixes)
	if len(cds) != 1 || !strings.Contains(cds[0].note, "another release line") {
		t.Errorf("abbreviated reference not matched to its full-SHA evidence: %+v", cds)
	}
}

// TestReleaseLine_offBranchTag: a tag on a release commit made off the
// branch (Apache Tomcat) makes every later fix on that branch "diverged"
// by exactly the tagged commit; with the merge base being the tag's
// parent the commit descends from the release. A real divergence (more
// commits behind, or another merge base) stays another line.
func TestReleaseLine_offBranchTag(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	s.otherCommit()
	s.tags(map[string][]string{"1.2.3": {"1.2.3"}})
	s.compareBody("1.2.3", f5Fix, `{"status":"diverged","ahead_by":1080,"behind_by":1,"merge_base_commit":{"sha":"aaa"},"base_commit":{"sha":"tag","parents":[{"sha":"aaa"}]}}`)
	s.compareBody("1.2.3", f5cOther, `{"status":"diverged","ahead_by":20000,"behind_by":1107,"merge_base_commit":{"sha":"bbb"},"base_commit":{"sha":"tag","parents":[{"sha":"aaa"}]}}`)

	e, _, err := s.resolver(t, nil).Resolve(context.Background(), f5cCVE("CVE-F5C-6"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := lines(e), []string{LineDescends + "|", LineOther + "|" + ExcludedReleaseLine}; !reflect.DeepEqual(got, want) {
		t.Fatalf("lines = %v, want %v", got, want)
	}
	if e.FixCommits[0].ReleaseLineReason != NoteOffBranchTag || e.FixCommits[1].ReleaseLineReason != "" {
		t.Errorf("notes = %q, %q", e.FixCommits[0].ReleaseLineReason, e.FixCommits[1].ReleaseLineReason)
	}

	// behind by one, but the merge base is not the tag's parent.
	s2 := newF5Server(t)
	s2.standard()
	s2.api("o/r", f5Fix, f5Parent)
	s2.tags(map[string][]string{"1.2.3": {"1.2.3"}})
	s2.compareBody("1.2.3", f5Fix, `{"status":"diverged","behind_by":1,"merge_base_commit":{"sha":"ccc"},"base_commit":{"sha":"tag","parents":[{"sha":"aaa"}]}}`)
	one := f5CVE("CVE-F5C-6", f5Fix)
	one.Affected = []cves.Affected{{Package: "org.o:lib", Version: "1.2.3"}}
	e, _, _ = s2.resolver(t, nil).Resolve(context.Background(), one)
	if e.FixCommits[0].ReleaseLine != LineOther {
		t.Errorf("unrelated merge base: %+v", e.FixCommits[0])
	}

	// the tag's parent is the merge base, but the release has more
	// commits the fix lacks: a branch forked before the release.
	s3 := newF5Server(t)
	s3.standard()
	s3.api("o/r", f5Fix, f5Parent)
	s3.tags(map[string][]string{"1.2.3": {"1.2.3"}})
	s3.compareBody("1.2.3", f5Fix, `{"status":"diverged","behind_by":5,"merge_base_commit":{"sha":"aaa"},"base_commit":{"sha":"tag","parents":[{"sha":"aaa"}]}}`)
	e, _, _ = s3.resolver(t, nil).Resolve(context.Background(), one)
	if e.FixCommits[0].ReleaseLine != LineOther {
		t.Errorf("behind by several: %+v", e.FixCommits[0])
	}
}
