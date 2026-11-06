package vulnmap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

// CVE:F5 fixtures: one fake GitHub serving .diff, raw files at two
// revisions and the commit-metadata API.

const (
	f5Parent = "1111111111111111111111111111111111111111"
	f5Fix    = "2222222222222222222222222222222222222222"
)

// The fix removes the vulnerable `unsafeLoad`, hardens `load` by calling
// a new helper `checkSafe`, and leaves `other` alone.
const f5Old = `package p;

public class Parser {
    public Object load(String s) {
        return unsafeLoad(s);
    }

    Object unsafeLoad(String s) {
        return eval(s);
    }

    void other() {
    }
}
`

const f5New = `package p;

public class Parser {
    public Object load(String s) {
        checkSafe(s);
        return safeLoad(s);
    }

    void checkSafe(String s) {
        if (s == null) throw new IllegalArgumentException();
    }

    void other() {
    }
}
`

const f5Diff = `diff --git a/src/main/java/p/Parser.java b/src/main/java/p/Parser.java
index aaaaaaa..bbbbbbb 100644
--- a/src/main/java/p/Parser.java
+++ b/src/main/java/p/Parser.java
@@ -4,10 +4,11 @@ public class Parser {
     public Object load(String s) {
-        return unsafeLoad(s);
+        checkSafe(s);
+        return safeLoad(s);
     }
 
-    Object unsafeLoad(String s) {
-        return eval(s);
+    void checkSafe(String s) {
+        if (s == null) throw new IllegalArgumentException();
     }
 
     void other() {
`

type f5Server struct {
	mux     *http.ServeMux
	srv     *httptest.Server
	apiHits int
}

func newF5Server(t *testing.T) *f5Server {
	t.Helper()
	s := &f5Server{mux: http.NewServeMux()}
	s.srv = httptest.NewServer(s.mux)
	t.Cleanup(s.srv.Close)
	return s
}

func (s *f5Server) serve(path, body string) {
	s.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) })
}

func (s *f5Server) api(repo, sha string, parents ...string) {
	var ps []string
	for _, p := range parents {
		ps = append(ps, `{"sha":"`+p+`"}`)
	}
	s.mux.HandleFunc("/repos/"+repo+"/commits/"+sha, func(w http.ResponseWriter, r *http.Request) {
		s.apiHits++
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"sha":"` + sha + `","parents":[` + strings.Join(ps, ",") + `]}`))
	})
}

func (s *f5Server) resolver(t *testing.T, cache *httpcache.Cache) *Resolver {
	if cache == nil {
		cache = mustCache(t)
	}
	r := NewResolver(nil, cache)
	r.HTTPClient = s.srv.Client()
	r.GithubBaseURL, r.RawBaseURL, r.GithubAPIBaseURL = s.srv.URL, s.srv.URL, s.srv.URL
	return r
}

func f5CVE(id string, shas ...string) cves.CVE {
	c := cves.CVE{ID: id}
	for _, sha := range shas {
		c.References = append(c.References, cves.Reference{Type: "FIX", URL: "https://github.com/o/r/commit/" + sha})
	}
	return c
}

func (s *f5Server) standard() {
	s.serve("/o/r/commit/"+f5Fix+".diff", f5Diff)
	s.serve("/o/r/"+f5Parent+"/src/main/java/p/Parser.java", f5Old)
	s.serve("/o/r/"+f5Fix+"/src/main/java/p/Parser.java", f5New)
}

// TestFixDiff_bothSides: the removed method is a candidate, the changed
// one too, the new helper is not (CVE:F5 criteria 1 and 2).
func TestFixDiff_bothSides(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)

	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-F5-1", f5Fix))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if want := []string{"p.Parser#load", "p.Parser#unsafeLoad"}; !reflect.DeepEqual(e.Methods, want) {
		t.Fatalf("Methods = %v, want %v (old side: removed + changed; never the added helper)", e.Methods, want)
	}
	if e.NewSideOnly() {
		t.Error("NewSideOnly with a single-parent commit")
	}
	fc := e.FixCommits[0]
	if fc.Base != BaseParent || fc.Parent != f5Parent || fc.SHA != f5Fix {
		t.Errorf("commit evidence = %+v", fc)
	}
	f := fc.Files[0]
	want := map[string]string{"p.Parser#load": FixMethodChanged, "p.Parser#unsafeLoad": FixMethodRemoved, "p.Parser#checkSafe": FixMethodAdded}
	if !reflect.DeepEqual(kinds(f.Methods), want) {
		t.Errorf("file methods = %v, want %v", f.Methods, want)
	}
	if f.OldSHA256 == "" || f.NewSHA256 == "" || f.OldSHA256 == f.NewSHA256 {
		t.Errorf("source hashes old=%q new=%q", f.OldSHA256, f.NewSHA256)
	}
}

// TestFixDiff_metadataUnavailable: no API answer → the post-fix side only
// (the pre-F5 behavior), visibly marked.
func TestFixDiff_metadataUnavailable(t *testing.T) {
	s := newF5Server(t)
	s.standard() // no API handler: 404

	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-F5-2", f5Fix))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if want := []string{"p.Parser#checkSafe", "p.Parser#load"}; !reflect.DeepEqual(e.Methods, want) {
		t.Errorf("Methods = %v, want %v (new side only)", e.Methods, want)
	}
	if !e.NewSideOnly() || e.FixCommits[0].Base != BaseNewSideOnly || !strings.Contains(e.FixCommits[0].BaseReason, "metadata unavailable") {
		t.Errorf("evidence = %+v, want new_side_only with reason", e.FixCommits[0])
	}
}

// TestFixDiff_mergeCommit: two parents → no guessed base.
func TestFixDiff_mergeCommit(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent, "3333333333333333333333333333333333333333")

	e, _, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-F5-3", f5Fix))
	if err != nil {
		t.Fatal(err)
	}
	fc := e.FixCommits[0]
	if fc.Base != BaseNewSideOnly || fc.Parent != "" || !strings.Contains(fc.BaseReason, "merge commit") {
		t.Errorf("merge commit evidence = %+v", fc)
	}
}

// TestFixDiff_renameAndDelete: a renamed class is mapped under its
// pre-fix name; a deleted file's methods are all removed candidates; an
// added file contributes nothing.
func TestFixDiff_renameAndDelete(t *testing.T) {
	s := newF5Server(t)
	const diff = `diff --git a/src/main/java/a/Old.java b/src/main/java/b/Moved.java
similarity index 80%
rename from src/main/java/a/Old.java
rename to src/main/java/b/Moved.java
index 1111111..2222222 100644
--- a/src/main/java/a/Old.java
+++ b/src/main/java/b/Moved.java
@@ -1,5 +1,5 @@
-package a;
+package b;
 public class Old {
     void m() {
-        bad();
+        good();
     }
diff --git a/src/main/java/a/Gone.java b/src/main/java/a/Gone.java
deleted file mode 100644
index 3333333..0000000
--- a/src/main/java/a/Gone.java
+++ /dev/null
@@ -1,4 +0,0 @@
-package a;
-public class Gone {
-    void g() { bad(); }
-}
diff --git a/src/main/java/a/Guard.java b/src/main/java/a/Guard.java
new file mode 100644
index 0000000..4444444
--- /dev/null
+++ b/src/main/java/a/Guard.java
@@ -0,0 +1,4 @@
+package a;
+public class Guard {
+    static void check() { }
+}
`
	s.serve("/o/r/commit/"+f5Fix+".diff", diff)
	s.serve("/o/r/"+f5Parent+"/src/main/java/a/Old.java", "package a;\npublic class Old {\n    void m() {\n        bad();\n    }\n}\n")
	s.serve("/o/r/"+f5Fix+"/src/main/java/b/Moved.java", "package b;\npublic class Old {\n    void m() {\n        good();\n    }\n}\n")
	s.serve("/o/r/"+f5Parent+"/src/main/java/a/Gone.java", "package a;\npublic class Gone {\n    void g() { bad(); }\n}\n")
	s.serve("/o/r/"+f5Fix+"/src/main/java/a/Guard.java", "package a;\npublic class Guard {\n    static void check() { }\n}\n")
	s.api("o/r", f5Fix, f5Parent)

	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-F5-4", f5Fix))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if want := []string{"a.Gone#g", "a.Old#m"}; !reflect.DeepEqual(e.Methods, want) {
		t.Errorf("Methods = %v, want %v", e.Methods, want)
	}
	if got := kinds(e.FixCommits[0].Files[2].Methods); !reflect.DeepEqual(got, map[string]string{"a.Guard#check": FixMethodAdded}) {
		t.Errorf("added file methods = %v", got)
	}
}

// TestFixDiff_noSources: the diff is readable but neither side's source
// is: no method, evidence says why.
func TestFixDiff_noSources(t *testing.T) {
	s := newF5Server(t)
	s.serve("/o/r/commit/"+f5Fix+".diff", f5Diff)
	s.api("o/r", f5Fix, f5Parent)

	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-F5-5", f5Fix))
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v entry=%+v, want no mapping", ok, err, e)
	}
}

// TestFixDiff_unreadableCommitIsPerCVE: a failing diff for one CVE gives
// it SourceError and does not stop the others (CVE:F5 task 10).
func TestFixDiff_unreadableCommitIsPerCVE(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	const broken = "4444444444444444444444444444444444444444"
	s.mux.HandleFunc("/o/r/commit/"+broken+".diff", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	got, err := s.resolver(t, nil).ResolveAll(context.Background(), []cves.CVE{f5CVE("CVE-OK", f5Fix), f5CVE("CVE-BROKEN", broken)})
	if err != nil {
		t.Fatalf("ResolveAll: %v (a per-CVE source failure must not fail the run)", err)
	}
	if !got["CVE-OK"].Mapped() {
		t.Errorf("CVE-OK not mapped: %+v", got["CVE-OK"])
	}
	b, ok := got["CVE-BROKEN"]
	if !ok || b.Mapped() || b.SourceError == "" {
		t.Errorf("CVE-BROKEN = %+v ok=%v, want an unmapped entry with SourceError", b, ok)
	}

	// One unreadable commit next to a readable one: the readable one
	// still maps, the failure is kept as evidence.
	e, _, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-MIXED", broken, f5Fix))
	if err != nil || !e.Mapped() || e.FixCommits[0].Error == "" {
		t.Errorf("mixed: err=%v entry=%+v", err, e)
	}
}

// TestFixDiff_offlineMissStopsRun: an offline cache without the diff is a
// contract failure, not a per-CVE gap.
func TestFixDiff_offlineMissStopsRun(t *testing.T) {
	s := newF5Server(t)
	c, err := httpcache.New(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.resolver(t, c).ResolveAll(context.Background(), []cves.CVE{f5CVE("CVE-OFF", f5Fix)})
	if !errors.Is(err, httpcache.ErrOffline) {
		t.Fatalf("err = %v, want ErrOffline", err)
	}
}

// TestFixDiff_offlineWithoutMetadata: an older cache with the diff and the
// post-fix file but no commit metadata degrades to new-side-only instead
// of failing (metadata is an enrichment).
func TestFixDiff_offlineWithoutMetadata(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	dir := t.TempDir()
	online, _ := httpcache.New(dir, false)
	r := s.resolver(t, online)
	r.GithubAPIBaseURL = s.srv.URL + "/nowhere" // first run: no metadata cached
	if _, _, err := r.Resolve(context.Background(), f5CVE("CVE-OLD", f5Fix)); err != nil {
		t.Fatal(err)
	}
	offline, _ := httpcache.New(dir, true)
	r = s.resolver(t, offline)
	e, ok, err := r.Resolve(context.Background(), f5CVE("CVE-OLD", f5Fix))
	if err != nil || !ok || !e.NewSideOnly() {
		t.Fatalf("ok=%v err=%v newSideOnly=%v", ok, err, e.NewSideOnly())
	}
}

// TestFixDiff_tokenNotCached: the token is sent and never stored.
func TestFixDiff_tokenSent(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	var auth string
	s.mux.HandleFunc("/repos/o/r/commits/"+f5Fix, func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.Write([]byte(`{"sha":"` + f5Fix + `","parents":[{"sha":"` + f5Parent + `"}]}`))
	})
	r := s.resolver(t, nil)
	r.GithubToken = "t0k"
	if _, _, err := r.Resolve(context.Background(), f5CVE("CVE-TOK", f5Fix)); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer t0k" {
		t.Errorf("Authorization = %q", auth)
	}
}

func kinds(ms []FixMethod) map[string]string {
	out := map[string]string{}
	for _, m := range ms {
		out[m.Selector] = m.Kind
	}
	return out
}
