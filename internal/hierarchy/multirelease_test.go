package hierarchy

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"ariadne/internal/input"
)

// readFixtureEntryBytes extracts one raw .class entry's bytes from the
// committed fixture jar — real, javac-compiled bytecode, used below as
// stand-in "variant content" distinguishable by its own ThisClass, so a
// full readJarClasses/BuildWithTarget round trip can confirm which
// variant was actually parsed, not just which zip.File selectClassEntries
// picked.
func readFixtureEntryBytes(t *testing.T, entry string) []byte {
	t.Helper()
	zr, err := zip.OpenReader("../../testdata/fixture/fixture-8.jar")
	if err != nil {
		t.Fatalf("opening fixture jar: %v", err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name != entry {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening %s: %v", entry, err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("reading %s: %v", entry, err)
		}
		return data
	}
	t.Fatalf("entry %s not found in fixture jar", entry)
	return nil
}

// TestBuildWithTarget_selectsRealVariantBytecode is the full-pipeline
// counterpart to the selectClassEntries-only tests above: a synthetic
// multi-release dependency jar whose base entry is really
// fixture.Greeter's compiled bytecode and whose version-17 variant is
// really fixture.FormalGreeter's — different real classes standing in
// for "two variants of the same logical class" distinguishable by their
// own ThisClass, so the resulting Hierarchy proves readJarClasses parsed
// the *selected* bytes, not just that selectClassEntries named the right
// zip.File.
func TestBuildWithTarget_selectsRealVariantBytecode(t *testing.T) {
	greeterBytes := readFixtureEntryBytes(t, "fixture/Greeter.class")
	formalBytes := readFixtureEntryBytes(t, "fixture/FormalGreeter.class")

	dep := filepath.Join(t.TempDir(), "mr-dep.jar")
	writeZipFile(t, dep, map[string][]byte{
		"META-INF/MANIFEST.MF":                  []byte(mrManifest),
		"test/Thing.class":                      greeterBytes,
		"META-INF/versions/17/test/Thing.class": formalBytes,
	})

	// An empty app jar, not fixture-8.jar itself: fixture.Greeter and
	// fixture.FormalGreeter already exist natively in fixture-8.jar, and
	// Build keeps whichever job (app, index 0) it sees first on a name
	// collision — which would hide what this test is actually checking.
	appJar := emptyJar(t)

	withoutTarget, _, err := Build(appJar, []input.DepJar{{Path: dep}})
	if err != nil {
		t.Fatalf("Build (no target): %v", err)
	}
	if _, ok := withoutTarget.Lookup("fixture.Greeter"); !ok {
		t.Errorf("without -target-java, expected the base entry's class (fixture.Greeter, from Thing.class) to be read")
	}

	withTarget, _, err := BuildWithTarget(appJar, []input.DepJar{{Path: dep}}, 17)
	if err != nil {
		t.Fatalf("BuildWithTarget(17): %v", err)
	}
	ci, ok := withTarget.Lookup("fixture.FormalGreeter")
	if !ok {
		t.Fatalf("with -target-java 17, expected the version-17 variant's class (fixture.FormalGreeter, from Thing.class) to be read")
	}
	if ci.VariantVersion != 17 {
		t.Errorf("VariantVersion = %d, want 17", ci.VariantVersion)
	}
	if ci.SourceEntry != "META-INF/versions/17/test/Thing.class" {
		t.Errorf("SourceEntry = %q, want the versioned entry path", ci.SourceEntry)
	}
}

func writeZipFile(t *testing.T, path string, entries map[string][]byte) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func zipReader(t *testing.T, entries map[string][]byte) *zip.Reader {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func selectedFor(t *testing.T, sels []selectedEntry, baseName string) (selectedEntry, bool) {
	t.Helper()
	for _, s := range sels {
		if s.file.Name == baseName || s.file.Name == "META-INF/versions/"+itoa(s.variant)+"/"+baseName {
			return s, true
		}
	}
	return selectedEntry{}, false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

const mrManifest = "Manifest-Version: 1.0\nMulti-Release: true\n"

// TestSelectClassEntries_notMultiReleaseIgnoresVersionsDir is JVM_GRAPH
// G4's own rule: without "Multi-Release: true" in the manifest,
// META-INF/versions/* is just an inert resource, even with a target set
// — never an active variant.
func TestSelectClassEntries_notMultiReleaseIgnoresVersionsDir(t *testing.T) {
	zr := zipReader(t, map[string][]byte{
		"com/example/Foo.class":                      []byte("base"),
		"META-INF/versions/17/com/example/Foo.class": []byte("variant17"),
	})
	sels, warnings := selectClassEntries(zr, 17)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	sel, ok := selectedFor(t, sels, "com/example/Foo.class")
	if !ok {
		t.Fatalf("com/example/Foo.class not selected at all; sels=%v", sels)
	}
	if sel.variant != 0 {
		t.Errorf("variant = %d, want 0 (base) — not a declared multi-release jar", sel.variant)
	}
}

// TestSelectClassEntries_unsetTargetFallsBackToBase is G4's documented
// fallback: even a genuinely multi-release jar never activates a
// variant without an explicit target — it must not guess from the
// highest version present.
func TestSelectClassEntries_unsetTargetFallsBackToBase(t *testing.T) {
	zr := zipReader(t, map[string][]byte{
		"META-INF/MANIFEST.MF":                       []byte(mrManifest),
		"com/example/Foo.class":                      []byte("base"),
		"META-INF/versions/17/com/example/Foo.class": []byte("variant17"),
	})
	sels, _ := selectClassEntries(zr, 0)
	sel, ok := selectedFor(t, sels, "com/example/Foo.class")
	if !ok {
		t.Fatalf("com/example/Foo.class not selected; sels=%v", sels)
	}
	if sel.variant != 0 {
		t.Errorf("variant = %d, want 0 (base) when no target is given", sel.variant)
	}
}

// TestSelectClassEntries_picksHighestQualifyingVariant covers the core
// selection rule across three targets against the same jar (variants at
// 9, 11 and 17): each target must pick the highest variant <= itself,
// never the highest variant overall.
func TestSelectClassEntries_picksHighestQualifyingVariant(t *testing.T) {
	const path = "com/example/Foo.class"
	entries := map[string][]byte{
		"META-INF/MANIFEST.MF":         []byte(mrManifest),
		path:                           []byte("base"),
		"META-INF/versions/9/" + path:  []byte("v9"),
		"META-INF/versions/11/" + path: []byte("v11"),
		"META-INF/versions/17/" + path: []byte("v17"),
	}

	cases := []struct {
		target      int
		wantVariant int
	}{
		{target: 8, wantVariant: 0}, // below every variant: base
		{target: 9, wantVariant: 9},
		{target: 10, wantVariant: 9}, // highest <= 10 is 9
		{target: 11, wantVariant: 11},
		{target: 16, wantVariant: 11}, // highest <= 16 is 11
		{target: 17, wantVariant: 17},
		{target: 21, wantVariant: 17}, // highest <= 21 is 17, not a guess at 21
	}
	for _, tc := range cases {
		zr := zipReader(t, entries)
		sels, _ := selectClassEntries(zr, tc.target)
		sel, ok := selectedFor(t, sels, path)
		if !ok {
			t.Fatalf("target=%d: %s not selected; sels=%v", tc.target, path, sels)
		}
		if sel.variant != tc.wantVariant {
			t.Errorf("target=%d: variant = %d, want %d", tc.target, sel.variant, tc.wantVariant)
		}
	}
}

// TestSelectClassEntries_malformedVersionDirWarnsAndFallsBackToBase is
// G4's "raportować niezgodne wejścia": a non-numeric or below-minimum
// version directory must not be silently trusted — it's skipped with a
// warning, and the base entry still wins.
func TestSelectClassEntries_malformedVersionDirWarnsAndFallsBackToBase(t *testing.T) {
	const path = "com/example/Foo.class"
	zr := zipReader(t, map[string][]byte{
		"META-INF/MANIFEST.MF":                   []byte(mrManifest),
		path:                                     []byte("base"),
		"META-INF/versions/not-a-number/" + path: []byte("bogus"),
		"META-INF/versions/5/" + path:            []byte("too-old"), // MR jars start at 9
	})
	sels, warnings := selectClassEntries(zr, 17)
	if len(warnings) != 2 {
		t.Fatalf("warnings = %v, want 2 (one per malformed version directory)", warnings)
	}
	sel, ok := selectedFor(t, sels, path)
	if !ok {
		t.Fatalf("%s not selected at all; sels=%v", path, sels)
	}
	if sel.variant != 0 {
		t.Errorf("variant = %d, want 0 (base) — both version directories are invalid", sel.variant)
	}
}

// TestSelectClassEntries_moduleInfoAlwaysSkipped confirms module-info is
// never selected, in either the base or a versioned location.
func TestSelectClassEntries_moduleInfoAlwaysSkipped(t *testing.T) {
	zr := zipReader(t, map[string][]byte{
		"META-INF/MANIFEST.MF":                   []byte(mrManifest),
		"module-info.class":                      []byte("base module-info"),
		"META-INF/versions/17/module-info.class": []byte("variant module-info"),
		"com/example/Foo.class":                  []byte("real class"),
	})
	sels, _ := selectClassEntries(zr, 17)
	for _, s := range sels {
		if s.file.Name == "module-info.class" || s.file.Name == "META-INF/versions/17/module-info.class" {
			t.Errorf("module-info.class was selected: %+v", s)
		}
	}
	if _, ok := selectedFor(t, sels, "com/example/Foo.class"); !ok {
		t.Errorf("com/example/Foo.class was not selected")
	}
}
