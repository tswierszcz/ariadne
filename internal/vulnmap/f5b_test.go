package vulnmap

import (
	"context"
	"reflect"
	"testing"
)

// fakeBytecode is a tiny "analyzed classes" table: class → name → descs.
type fakeBytecode map[string]map[string][]string

func (b fakeBytecode) fn() func(string, string) ([]string, bool) {
	return func(cls, name string) ([]string, bool) {
		ms, ok := b[cls]
		if !ok {
			return nil, false
		}
		return ms[name], true
	}
}

func methodsBySelector(e Entry) map[string]FixMethod {
	out := map[string]FixMethod{}
	for _, c := range e.FixCommits {
		for _, f := range c.Files {
			for _, m := range f.Methods {
				out[m.Selector] = m
			}
		}
	}
	return out
}

// TestBytecodeVerification: one matching overload → exact selector; a
// class or method absent from the analyzed classes is said so; several
// matching overloads stay a name selector (never guessed).
func TestBytecodeVerification(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	r := s.resolver(t, nil)
	r.Bytecode = fakeBytecode{"p.Parser": {
		"load":       {"(Ljava/lang/String;)Ljava/lang/Object;", "(I)Ljava/lang/Object;"},
		"unsafeLoad": {"(Ljava/lang/String;)Ljava/lang/Object;", "(Lother/String;)Ljava/lang/Object;"},
	}}.fn()

	e, ok, err := r.Resolve(context.Background(), f5CVE("CVE-F5B-1", f5Fix))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	want := []string{"p.Parser#load(Ljava/lang/String;)Ljava/lang/Object;", "p.Parser#unsafeLoad"}
	if !reflect.DeepEqual(e.Methods, want) {
		t.Fatalf("Methods = %v, want %v", e.Methods, want)
	}
	ms := methodsBySelector(e)
	if m := ms[want[0]]; m.Bytecode != BytecodeExact || !reflect.DeepEqual(m.SourceParams, []string{"String"}) {
		t.Errorf("load = %+v", m)
	}
	if m := ms["p.Parser#unsafeLoad"]; m.Bytecode != BytecodeAmbiguous {
		t.Errorf("unsafeLoad = %+v, want ambiguous (two String overloads)", m)
	}
	if m := ms["p.Parser#checkSafe"]; m.Kind != FixMethodAdded || m.Bytecode != "" {
		t.Errorf("added helper verified: %+v", m)
	}

	r.Bytecode = fakeBytecode{"p.Parser": {"load": {"(I)V"}}}.fn()
	e, _, _ = r.Resolve(context.Background(), f5CVE("CVE-F5B-1", f5Fix))
	ms = methodsBySelector(e)
	if ms["p.Parser#load"].Bytecode != BytecodeParamMismatch || ms["p.Parser#unsafeLoad"].Bytecode != BytecodeMethodMissing {
		t.Errorf("mismatch/missing: %+v", ms)
	}
	r.Bytecode = fakeBytecode{}.fn()
	e, _, _ = r.Resolve(context.Background(), f5CVE("CVE-F5B-1", f5Fix))
	if ms := methodsBySelector(e); ms["p.Parser#load"].Bytecode != BytecodeClassMissing {
		t.Errorf("class missing: %+v", ms)
	}
}

// TestReleaseFit: of two fix commits (one per release line), the one with
// no candidate in the analyzed classes is left out — but only when the
// other one has a present candidate.
func TestReleaseFit(t *testing.T) {
	s := newF5Server(t)
	s.standard()
	s.api("o/r", f5Fix, f5Parent)
	const other, otherParent = "5555555555555555555555555555555555555555", "6666666666666666666666666666666666666666"
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
	s.serve("/o/r/commit/"+other+".diff", otherDiff)
	s.serve("/o/r/"+otherParent+"/src/main/java/q/Legacy.java", "package q;\npublic class Legacy {\n    void parse() { bad(); }\n}\n")
	s.serve("/o/r/"+other+"/src/main/java/q/Legacy.java", "package q;\npublic class Legacy {\n    void parse() { good(); }\n}\n")
	s.api("o/r", other, otherParent)

	cve := f5CVE("CVE-F5B-2", f5Fix, other)
	cve.References[1].Type = "WEB"
	r := s.resolver(t, nil)
	r.Bytecode = fakeBytecode{"p.Parser": {"load": {"(Ljava/lang/String;)Ljava/lang/Object;"}}}.fn()
	e, _, err := r.Resolve(context.Background(), cve)
	if err != nil {
		t.Fatal(err)
	}
	if e.FixCommits[0].ReleaseFit != FitPresent || e.FixCommits[1].ReleaseFit != FitNonePresent {
		t.Fatalf("fit = %q, %q", e.FixCommits[0].ReleaseFit, e.FixCommits[1].ReleaseFit)
	}
	if e.FixCommits[0].RefType != "FIX" || e.FixCommits[1].RefType != "WEB" {
		t.Errorf("ref types = %q, %q", e.FixCommits[0].RefType, e.FixCommits[1].RefType)
	}
	for _, m := range e.Methods {
		if m == "q.Legacy#parse" {
			t.Errorf("a commit with no present candidate contributed %s", m)
		}
	}

	// Neither commit present: no conclusion, every candidate kept.
	r.Bytecode = fakeBytecode{}.fn()
	e, _, _ = r.Resolve(context.Background(), cve)
	if e.FixCommits[1].ReleaseFit != "" || len(e.Methods) != 3 {
		t.Errorf("no present commit: fit=%q methods=%v", e.FixCommits[1].ReleaseFit, e.Methods)
	}
}

// TestNestedClassAndRename: a member class's method maps under Outer$Inner;
// a removed method and an added one with the same body are paired as a
// possible rename, without changing the mapping.
func TestNestedClassAndRename(t *testing.T) {
	s := newF5Server(t)
	const oldSrc = `package p;
public class Outer {
    static class Inner {
        void legacyName(String s) {
            int a = compute(s);
            store(a, s);
            notifyAll(a);
        }
    }
}
`
	const newSrc = `package p;
public class Outer {
    static class Inner {
        void newName(String s) {
            int a = compute(s);
            store(a, s);
            notifyAll(a);
        }
    }
}
`
	const diff = `diff --git a/src/main/java/p/Outer.java b/src/main/java/p/Outer.java
index 1111111..2222222 100644
--- a/src/main/java/p/Outer.java
+++ b/src/main/java/p/Outer.java
@@ -3,6 +3,6 @@ public class Outer {
     static class Inner {
-        void legacyName(String s) {
+        void newName(String s) {
             int a = compute(s);
             store(a, s);
             notifyAll(a);
`
	s.serve("/o/r/commit/"+f5Fix+".diff", diff)
	s.serve("/o/r/"+f5Parent+"/src/main/java/p/Outer.java", oldSrc)
	s.serve("/o/r/"+f5Fix+"/src/main/java/p/Outer.java", newSrc)
	s.api("o/r", f5Fix, f5Parent)

	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-F5B-3", f5Fix))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if want := []string{"p.Outer$Inner#legacyName"}; !reflect.DeepEqual(e.Methods, want) {
		t.Errorf("Methods = %v, want %v", e.Methods, want)
	}
	ms := methodsBySelector(e)
	if ms["p.Outer$Inner#legacyName"].PossibleRename != "p.Outer$Inner#newName" || ms["p.Outer$Inner#newName"].PossibleRename != "p.Outer$Inner#legacyName" {
		t.Errorf("rename pairing: %+v", ms)
	}
}

func TestParamsCompatible(t *testing.T) {
	for _, tc := range []struct {
		src        []string
		desc, name string
		want       bool
	}{
		{[]string{"String", "int[]"}, "(Ljava/lang/String;[I)V", "m", true},
		{[]string{"Entry"}, "(Ljava/util/Map$Entry;)V", "m", true},
		{[]string{"?"}, "(Ljava/lang/Object;)V", "m", true},
		{[]string{"?[]"}, "([Ljava/lang/Object;)V", "m", true},
		{[]string{"?[]"}, "(Ljava/lang/Object;)V", "m", false},
		{[]string{"long"}, "(I)V", "m", false},
		{[]string{"int"}, "(Lp/Outer;I)V", "<init>", true}, // inner-class constructor
		{[]string{"int"}, "(Lp/Outer;I)V", "m", false},
		{nil, "(Ljava/lang/String;I)V", "<init>", true}, // enum constructor
	} {
		if got := paramsCompatible(tc.src, tc.desc, tc.name, ""); got != tc.want {
			t.Errorf("paramsCompatible(%v, %s, %s) = %v", tc.src, tc.desc, tc.name, got)
		}
	}
}

func TestCollapseSelectors(t *testing.T) {
	got := collapseSelectors([]string{"a.B#m(I)V", "a.B#m", "a.B#n(I)V", "a.C#m(J)V"})
	if want := []string{"a.B#m", "a.B#n(I)V", "a.C#m(J)V"}; !reflect.DeepEqual(got, want) {
		t.Errorf("collapseSelectors = %v, want %v", got, want)
	}
}
