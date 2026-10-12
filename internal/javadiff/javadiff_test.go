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
		byPath[f.Path] = f
	}

	yaml, ok := byPath["src/main/java/org/yaml/snakeyaml/Yaml.java"]
	if !ok {
		t.Fatalf("Yaml.java not found in %+v", files)
	}
	sort.Ints(yaml.ChangedLines)
	if want := []int{12, 13, 33}; !reflect.DeepEqual(yaml.ChangedLines, want) {
		t.Errorf("Yaml.java ChangedLines = %v, want %v", yaml.ChangedLines, want)
	}

	if _, ok := byPath["src/test/java/org/yaml/snakeyaml/YamlTest.java"]; !ok {
		t.Errorf("expected the test file to still be parsed (filtering src/main is the caller's job, not ParseUnifiedDiff's)")
	}
}

func TestParseUnifiedDiff_deletedFileSkipped(t *testing.T) {
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
	if len(files) != 0 {
		t.Fatalf("got %+v, want no files (pure deletion has nothing on the new side)", files)
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
		{"inside anonymous class's run method, not withAnon", 25, []string{"run"}},
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
