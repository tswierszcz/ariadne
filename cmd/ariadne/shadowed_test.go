package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestShadowedClassesReport: report.md lists components that define the
// same classes, with how many differ in content. The input is commons-lang3
// 3.11 three times: the jar, an identical copy, and a jar with Validate
// whose string constant was changed (same length — a valid class with
// other bytes).
func TestShadowedClassesReport(t *testing.T) {
	dir := t.TempDir()
	deps := filepath.Join(dir, "deps")
	if err := os.Mkdir(deps, 0o755); err != nil {
		t.Fatal(err)
	}
	orig, err := os.ReadFile(libraryFixture + "deps/commons-lang3-3.11.jar")
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(orig), int64(len(orig)))
	if err != nil {
		t.Fatal(err)
	}
	// Without META-INF/maven the jars stay unidentified: no OSV query.
	var stripped bytes.Buffer
	sw := zip.NewWriter(&stripped)
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "META-INF/maven/") || f.FileInfo().IsDir() {
			continue
		}
		rc, _ := f.Open()
		b, _ := io.ReadAll(rc)
		rc.Close()
		w, _ := sw.Create(f.Name)
		w.Write(b)
	}
	sw.Close()
	for _, n := range []string{"a-commons-lang3.jar", "b-lang3-copy.jar"} {
		if err := os.WriteFile(filepath.Join(deps, n), stripped.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var cls []byte
	for _, f := range zr.File {
		if f.Name == "org/apache/commons/lang3/Validate.class" {
			rc, _ := f.Open()
			cls, _ = io.ReadAll(rc)
			rc.Close()
		}
	}
	changed := bytes.Replace(cls, []byte("The validated object is null"), []byte("The validated object is NULL"), 1)
	if bytes.Equal(changed, cls) {
		t.Fatal("constant not found")
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("org/apache/commons/lang3/Validate.class")
	w.Write(changed)
	zw.Close()
	if err := os.WriteFile(filepath.Join(deps, "c-patched.jar"), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	md := filepath.Join(dir, "r.md")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", "../../testdata/native/app.jar", "-deps", deps,
		"-profile", "generic", "-out", md, "-json", filepath.Join(dir, "r.json")}
	if code, msg := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, msg)
	}
	rep, _ := os.ReadFile(md)
	out := string(rep)
	if !strings.Contains(out, "## Classes in several copies") {
		t.Fatalf("no section:\n%s", out)
	}
	sec := out[strings.Index(out, "## Classes in several copies"):]
	sec = sec[:strings.Index(sec, "## CVEs")]
	lines := strings.Split(sec, "\n")
	var rows []string
	for _, l := range lines {
		if strings.HasPrefix(l, "| ") && !strings.HasPrefix(l, "| Kept") {
			rows = append(rows, l)
		}
	}
	if len(rows) != 2 {
		t.Fatalf("rows %q", rows)
	}
	// The patched Validate first (1 class, 1 different), then the identical copy.
	if !strings.Contains(rows[0], "| 1 | 1 | `org.apache.commons.lang3.Validate` |") || !strings.Contains(rows[0], "c-patched.jar") {
		t.Errorf("first row %q", rows[0])
	}
	if !strings.Contains(rows[1], "b-lang3-copy.jar") || !strings.Contains(rows[1], "| 0 |") {
		t.Errorf("second row %q", rows[1])
	}
	if !strings.Contains(sec, "of them with different content") || !strings.Contains(sec, "1 of them") {
		t.Errorf("summary line: %s", sec)
	}
}
