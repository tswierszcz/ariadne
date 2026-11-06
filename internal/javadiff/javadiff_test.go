package javadiff

import (
	"reflect"
	"sort"
	"testing"
)

func TestParseUnifiedDiff(t *testing.T) {
	// A realistic two-hunk diff, shaped like what
	// `https://github.com/<o>/<r>/commit/<sha>.diff` serves: a changed
	// line, an added line, and a hunk with only a context+addition (no
	// deletion) to make sure line-number bookkeeping doesn't depend on
	// deletions being present.
	diff := `diff --git a/src/main/java/org/yaml/snakeyaml/Yaml.java b/src/main/java/org/yaml/snakeyaml/Yaml.java
index 1111111..2222222 100644
--- a/src/main/java/org/yaml/snakeyaml/Yaml.java
+++ b/src/main/java/org/yaml/snakeyaml/Yaml.java
@@ -10,5 +10,6 @@ public class Yaml {
     private int x;

-    public Object load(String s) {
+    public Object load(String s) {
+        checkSafe(s);
         return doLoad(s);
     }
@@ -30,3 +31,4 @@ public class Yaml {
     private void helper() {
         int y = 1;
+        y++;
     }
diff --git a/src/test/java/org/yaml/snakeyaml/YamlTest.java b/src/test/java/org/yaml/snakeyaml/YamlTest.java
index 3333333..4444444 100644
--- a/src/test/java/org/yaml/snakeyaml/YamlTest.java
+++ b/src/test/java/org/yaml/snakeyaml/YamlTest.java
@@ -1,2 +1,3 @@
 public class YamlTest {
+    // a new test
 }
`
	files, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2: %+v", len(files), files)
	}

	byPath := map[string]FileDiff{}
	for _, f := range files {
		byPath[f.Path()] = f
	}

	yaml, ok := byPath["src/main/java/org/yaml/snakeyaml/Yaml.java"]
	if !ok {
		t.Fatalf("Yaml.java not found in %+v", files)
	}
	sort.Ints(yaml.NewLines)
	if want := []int{12, 13, 33}; !reflect.DeepEqual(yaml.NewLines, want) {
		t.Errorf("Yaml.java NewLines = %v, want %v", yaml.NewLines, want)
	}
	if want := []int{12}; !reflect.DeepEqual(yaml.OldLines, want) {
		t.Errorf("Yaml.java OldLines = %v, want %v", yaml.OldLines, want)
	}
	if yaml.Change != ChangeModified || yaml.OldPath != yaml.NewPath {
		t.Errorf("Yaml.java = %+v, want modified in place", yaml)
	}

	if _, ok := byPath["src/test/java/org/yaml/snakeyaml/YamlTest.java"]; !ok {
		t.Errorf("expected the test file to still be parsed (filtering src/main is the caller's job, not ParseUnifiedDiff's)")
	}
}

func TestParseUnifiedDiff_deletedFileKept(t *testing.T) {
	diff := `diff --git a/Old.java b/Old.java
deleted file mode 100644
index 1111111..0000000
--- a/Old.java
+++ /dev/null
@@ -1,3 +0,0 @@
-public class Old {
-    void m() {}
-}
`
	files, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff: %v", err)
	}
	want := []FileDiff{{OldPath: "Old.java", Change: ChangeDeleted, OldLines: []int{1, 2, 3}}}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("got %+v, want %+v (CVE:F5: a deleted file's old lines are the evidence)", files, want)
	}
}

func TestParseUnifiedDiff_bothSides(t *testing.T) {
	// An added file, a rename with a content change, a pure rename, a
	// pure-deletion hunk, and a removed line whose content starts with
	// "-- " (shown as "--- ...") that must not be read as a file header.
	diff := `diff --git a/src/New.java b/src/New.java
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/src/New.java
@@ -0,0 +1,2 @@
+class New {
+}
diff --git a/src/a/Old.java b/src/b/Moved.java
similarity index 90%
rename from src/a/Old.java
rename to src/b/Moved.java
index 2222222..3333333 100644
--- a/src/a/Old.java
+++ b/src/b/Moved.java
@@ -3,4 +3,3 @@ class Old {
     void m() {
--- x;
-        y();
+        z();
     }
diff --git a/src/P.java b/src/Q.java
similarity index 100%
rename from src/P.java
rename to src/Q.java
diff --git a/src/D.java b/src/D.java
index 4444444..5555555 100644
--- a/src/D.java
+++ b/src/D.java
@@ -10,3 +10,2 @@
 a
-b
 c
@@ -20 +19 @@
-x
+y
`
	files, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff: %v", err)
	}
	want := []FileDiff{
		{NewPath: "src/New.java", Change: ChangeAdded, NewLines: []int{1, 2}},
		{OldPath: "src/a/Old.java", NewPath: "src/b/Moved.java", Change: ChangeRenamed, OldLines: []int{4, 5}, NewLines: []int{4}},
		{OldPath: "src/P.java", NewPath: "src/Q.java", Change: ChangeRenamed},
		{OldPath: "src/D.java", NewPath: "src/D.java", Change: ChangeModified, OldLines: []int{11, 20}, NewLines: []int{19}},
	}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("got\n%+v\nwant\n%+v", files, want)
	}
}

func TestDeclaredMethods(t *testing.T) {
	src := `package p;
class A {
    A() {}
    void one() { int x = 1; }
    abstract void noBody();
    class Inner { void two() {
    } }
}
`
	got := DeclaredMethods(src)
	want := map[string]bool{"<init>": true, "one": true, "two": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DeclaredMethods = %v, want %v", got, want)
	}
}

func TestFindMethodsForLines(t *testing.T) {
	src := `package org.yaml.snakeyaml;

public class Yaml {
    private int x;

    public Object load(String s) {
        checkSafe(s);
        return doLoad(s);
    }

    public Yaml() {
        this.x = 1;
    }

    private void helper() {
        int y = 1;
        if (y > 0) {
            y++;
        }
    }

    private void withAnon() {
        Runnable r = new Runnable() {
            public void run() {
                System.out.println("hi");
            }
        };
    }
}
`
	tests := []struct {
		name string
		line int
		want []string
	}{
		{"inside load body", 7, []string{"load"}},
		{"inside constructor", 11, []string{"<init>"}},
		{"inside if-block nested in helper (transparent)", 18, []string{"helper"}},
		// CVE:F5b: the anonymous class's run is Outer$1#run in bytecode, a
		// compiler-numbered name; the line belongs to the method that
		// creates the anonymous class (it exists under its own name).
		{"inside anonymous class's run method: the enclosing withAnon", 25, []string{"withAnon"}},
		{"field declaration, no enclosing method", 4, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FindMethodsForLines(src, []int{tc.line})
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("FindMethodsForLines(line %d) = %v, want %v", tc.line, got, tc.want)
			}
		})
	}
}

func TestFindMethodsForLines_dedupesAcrossMultipleLines(t *testing.T) {
	src := `public class C {
    void m() {
        int a = 1;
        int b = 2;
    }
}
`
	got := FindMethodsForLines(src, []int{3, 4})
	if !reflect.DeepEqual(got, []string{"m"}) {
		t.Errorf("got %v, want [m] (deduped)", got)
	}
}

func TestStripCommentsAndStrings_bracesInStringsAndCommentsIgnored(t *testing.T) {
	src := `public class C {
    // a comment with a brace {
    String s = "also a brace: {";
    /* block {
       comment } */
    void m() {
        char c = '{';
    }
}
`
	got := FindMethodsForLines(src, []int{7})
	if !reflect.DeepEqual(got, []string{"m"}) {
		t.Errorf("got %v, want [m] (stray braces in comments/strings/chars must not desync brace matching)", got)
	}
}

func TestFQCN(t *testing.T) {
	tests := []struct {
		path string
		src  string
		want string
	}{
		{"src/main/java/org/yaml/snakeyaml/Yaml.java", "package org.yaml.snakeyaml;\n\npublic class Yaml {}\n", "org.yaml.snakeyaml.Yaml"},
		{"Standalone.java", "public class Standalone {}\n", "Standalone"},
		{"NotJava.txt", "whatever", ""},
	}
	for _, tc := range tests {
		got, ok := FQCN(tc.path, tc.src)
		if tc.want == "" {
			if ok {
				t.Errorf("FQCN(%q) = (%q, true), want ok=false", tc.path, got)
			}
			continue
		}
		if !ok || got != tc.want {
			t.Errorf("FQCN(%q) = (%q, %v), want (%q, true)", tc.path, got, ok, tc.want)
		}
	}
}

func TestFindMethodsForLines_oneLineMethod(t *testing.T) {
	src := "class A {\n    void g() { bad(); }\n    void h() {\n    }\n}\n"
	if got := FindMethodsForLines(src, []int{2}); !reflect.DeepEqual(got, []string{"g"}) {
		t.Errorf("line 2 owner = %v, want [g]", got)
	}
	if got := FindMethodsForLines(src, []int{4}); len(got) != 0 {
		t.Errorf("closing-brace line owner = %v, want none", got)
	}
}

func TestMethodsForLines_nestedClassesAndParams(t *testing.T) {
	src := `package p;
public class Outer {
    public <T> void gen(T t, java.util.Map.Entry<String, Integer> e, final String... rest) {
        x();
    }
    static class Inner {
        Inner(int[] a, @Deprecated long b) {
            y();
        }
        class Deep { void d(byte[][] b) { z(); } }
    }
    void local() {
        class Local { void lm() {
            w();
        } }
        Runnable r = () -> {
            v();
        };
    }
}
class Second {
    void s(Object o) {
        u();
    }
}
`
	got := MethodsForLines(src, []int{4, 8, 10, 14, 17, 23})
	var keys []string
	for _, m := range got {
		keys = append(keys, m.Key())
	}
	want := []string{
		"Outer#gen(?,Entry,String[])",
		"Outer$Inner#<init>(int[],long)",
		"Outer$Inner$Deep#d(byte[][])",
		"Outer#local()", // local class and lambda lines belong to the enclosing method
		"Second#s(Object)",
	}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("got %v\nwant %v", keys, want)
	}
	if got[0].StartLine != 3 || got[0].EndLine != 5 {
		t.Errorf("gen lines = %d-%d, want 3-5", got[0].StartLine, got[0].EndLine)
	}
}
