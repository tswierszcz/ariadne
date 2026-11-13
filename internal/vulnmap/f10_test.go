package vulnmap

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

// CVE:F10 part A: a pull-request reference resolves to the commit that
// merged it; a two-parent merge commit is read against its first parent
// (the change the pull request brought), never as "new side only".
func TestFixDiff_pullRequestMergeCommit(t *testing.T) {
	s := newF5Server(t)
	const other = "3333333333333333333333333333333333333333"
	// The merge commit f5Fix: first parent f5Parent (the branch), second
	// the pull request's head.
	s.api("o/r", f5Fix, f5Parent, other)
	s.serve("/repos/o/r/pulls/7", `{"merged":true,"merge_commit_sha":"`+f5Fix+`","commits":2}`)
	s.serve("/repos/o/r/pulls/8", `{"merged":false,"merge_commit_sha":"4444444444444444444444444444444444444444","commits":1}`)
	s.serve("/o/r/compare/"+f5Parent+"..."+f5Fix+".diff", f5Diff)
	s.serve("/o/r/"+f5Parent+"/src/main/java/p/Parser.java", f5Old)
	s.serve("/o/r/"+f5Fix+"/src/main/java/p/Parser.java", f5New)

	cve := cves.CVE{ID: "CVE-F10-1", References: []cves.Reference{
		{Type: "WEB", URL: "https://github.com/o/r/pull/7"},
		{Type: "WEB", URL: "https://github.com/o/r/pull/7/files"}, // same pull request once
		{Type: "WEB", URL: "https://github.com/o/r/pull/8"},
		{Type: "WEB", URL: "https://github.com/github/advisory-database/pull/3281"}, // never fetched
	}}
	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), cve)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if want := []string{"p.Parser#load", "p.Parser#unsafeLoad"}; !reflect.DeepEqual(e.Methods, want) {
		t.Fatalf("Methods = %v, want %v (both sides of the merged change)", e.Methods, want)
	}
	var merged, notMerged *FixCommit
	for i := range e.FixCommits {
		switch e.FixCommits[i].PullRequest {
		case "#7":
			merged = &e.FixCommits[i]
		case "#8":
			notMerged = &e.FixCommits[i]
		}
	}
	if merged == nil || merged.Base != BaseParent || merged.Parent != f5Parent || !strings.Contains(merged.BaseReason, "pull request #7") {
		t.Errorf("merged pull request evidence = %+v", merged)
	}
	if notMerged == nil || notMerged.Error != "pull request not merged" {
		t.Errorf("unmerged pull request evidence = %+v", notMerged)
	}
}

// A pull request whose merge commit the advisory also lists as a commit
// is read once; a squash merge (one parent) is an ordinary commit.
func TestFixDiff_pullRequestSquashAndDuplicate(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	s.serve("/repos/o/r/pulls/9", `{"merged":true,"merge_commit_sha":"`+f5Fix+`","commits":1}`)

	// Only the pull request: squash merge, read like a commit.
	cve := cves.CVE{ID: "CVE-F10-2", References: []cves.Reference{{Type: "FIX", URL: "https://github.com/o/r/pull/9"}}}
	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), cve)
	if err != nil || !ok || len(e.FixCommits) != 1 || e.FixCommits[0].PullRequest != "#9" || e.FixCommits[0].Base != BaseParent {
		t.Fatalf("squash merge: ok=%v err=%v commits=%+v", ok, err, e.FixCommits)
	}

	// The same commit also listed directly: one FixCommit.
	cve.References = append(cve.References, cves.Reference{Type: "FIX", URL: "https://github.com/o/r/commit/" + f5Fix})
	e, _, _ = s.resolver(t, nil).Resolve(context.Background(), cve)
	if len(e.FixCommits) != 1 || e.FixCommits[0].PullRequest != "" {
		t.Errorf("duplicate: commits=%+v, want the direct commit only", e.FixCommits)
	}
}

// CVE:F10 part B: a fix that only changes a field initializer maps to
// <clinit>; next to a changed method the initializer is left out (its
// <clinit> runs whenever the class is used).
func TestFixDiff_fieldInitializers(t *testing.T) {
	const old = `package p;

public class Etag {
    private static final String PATTERN = "a*";

    boolean matches(String s) {
        return s.matches(PATTERN);
    }
}
`
	const onlyField = `package p;

public class Etag {
    private static final String PATTERN = "a";

    boolean matches(String s) {
        return s.matches(PATTERN);
    }
}
`
	const both = `package p;

public class Etag {
    private static final String PATTERN = "a";

    boolean matches(String s) {
        return s != null && s.matches(PATTERN);
    }
}
`
	diffOnlyField := `diff --git a/src/main/java/p/Etag.java b/src/main/java/p/Etag.java
--- a/src/main/java/p/Etag.java
+++ b/src/main/java/p/Etag.java
@@ -3,3 +3,3 @@ package p;
 public class Etag {
-    private static final String PATTERN = "a*";
+    private static final String PATTERN = "a";
 
`
	diffBoth := diffOnlyField + `@@ -6,3 +6,3 @@ public class Etag {
     boolean matches(String s) {
-        return s.matches(PATTERN);
+        return s != null && s.matches(PATTERN);
     }
`
	for name, tc := range map[string]struct {
		diff, newSrc string
		want         []string
	}{
		"only the field":   {diffOnlyField, onlyField, []string{"p.Etag#<clinit>"}},
		"field and method": {diffBoth, both, []string{"p.Etag#matches"}},
	} {
		s := newF5Server(t)
		s.api("o/r", f5Fix, f5Parent)
		s.serve("/o/r/commit/"+f5Fix+".diff", tc.diff)
		s.serve("/o/r/"+f5Parent+"/src/main/java/p/Etag.java", old)
		s.serve("/o/r/"+f5Fix+"/src/main/java/p/Etag.java", tc.newSrc)
		e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-F10-B", f5Fix))
		if err != nil || !ok || !reflect.DeepEqual(e.Methods, tc.want) {
			t.Errorf("%s: ok=%v err=%v methods=%v, want %v", name, ok, err, e.Methods, tc.want)
		}
	}
}

// An explicit constructor changed next to an instance field initializer
// keeps its selector (both are "C#<init>").
func TestFixDiff_constructorNextToFieldInitializer(t *testing.T) {
	const old = `package p;

public class Enc {
    private int rounds = 10;

    public Enc(int r) {
        rounds = r;
    }
}
`
	const fixed = `package p;

public class Enc {
    private int rounds = 12;

    public Enc(int r) {
        if (r > 30) throw new IllegalArgumentException();
        rounds = r;
    }
}
`
	diff := `diff --git a/src/main/java/p/Enc.java b/src/main/java/p/Enc.java
--- a/src/main/java/p/Enc.java
+++ b/src/main/java/p/Enc.java
@@ -3,6 +3,7 @@ package p;
 public class Enc {
-    private int rounds = 10;
+    private int rounds = 12;
 
     public Enc(int r) {
+        if (r > 30) throw new IllegalArgumentException();
         rounds = r;
     }
`
	s := newF5Server(t)
	s.api("o/r", f5Fix, f5Parent)
	s.serve("/o/r/commit/"+f5Fix+".diff", diff)
	s.serve("/o/r/"+f5Parent+"/src/main/java/p/Enc.java", old)
	s.serve("/o/r/"+f5Fix+"/src/main/java/p/Enc.java", fixed)
	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-F10-B2", f5Fix))
	if err != nil || !ok || !reflect.DeepEqual(e.Methods, []string{"p.Enc#<init>"}) {
		t.Errorf("ok=%v err=%v methods=%v, want the constructor", ok, err, e.Methods)
	}
}

// Offline, a pull request the cache lacks (its lookup failed online, so
// nothing was stored) is recorded per reference, not a stopped run; a
// commit the advisory names directly still stops it
// (TestFixDiff_offlineMissStopsRun).
func TestFixDiff_pullRequestOfflineMiss(t *testing.T) {
	c, err := httpcache.New(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	r := NewResolver(nil, c)
	r.GithubBaseURL, r.RawBaseURL, r.GithubAPIBaseURL = "http://127.0.0.1:1", "http://127.0.0.1:1", "http://127.0.0.1:1"
	cve := cves.CVE{ID: "CVE-F10-OFF", References: []cves.Reference{{Type: "WEB", URL: "https://github.com/o/r/pull/5"}}}
	e, ok, err := r.Resolve(context.Background(), cve)
	if err != nil {
		t.Fatalf("offline pull request miss stopped the run: %v", err)
	}
	if !ok || len(e.FixCommits) != 1 || e.FixCommits[0].PullRequest != "#5" || e.FixCommits[0].Error == "" {
		t.Errorf("ok=%v commits=%+v, want the miss recorded", ok, e.FixCommits)
	}
}
