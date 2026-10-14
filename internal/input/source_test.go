package input

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareAppInput_plainJarUsedAsIs(t *testing.T) {
	jar := "../../testdata/fixture/fixture-8.jar"
	p, err := PrepareAppInput(jar)
	if err != nil {
		t.Fatalf("PrepareAppInput: %v", err)
	}
	defer p.Cleanup()
	if p.AppClassesJar != jar {
		t.Errorf("AppClassesJar = %q, want the original path unchanged (no temp copy needed)", p.AppClassesJar)
	}
	if p.ExtractedLibs != "" {
		t.Errorf("ExtractedLibs = %q, want empty for a plain jar", p.ExtractedLibs)
	}
}

func TestPrepareAppInput_directoryIsPackedIntoAJar(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "com", "example"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "com", "example", "Foo.class"), []byte("fake class bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := PrepareAppInput(dir)
	if err != nil {
		t.Fatalf("PrepareAppInput: %v", err)
	}
	defer p.Cleanup()

	zr, err := zip.OpenReader(p.AppClassesJar)
	if err != nil {
		t.Fatalf("opening produced jar: %v", err)
	}
	defer zr.Close()
	found := false
	for _, f := range zr.File {
		if f.Name == "com/example/Foo.class" {
			found = true
		}
	}
	if !found {
		t.Errorf("produced jar has no com/example/Foo.class entry")
	}

	// Cleanup must actually remove the temp file.
	path := p.AppClassesJar
	if err := p.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("temp jar %s still exists after Cleanup", path)
	}
}

func TestPrepareAppInput_springBootFatJar(t *testing.T) {
	fat := filepath.Join(t.TempDir(), "app.jar")
	writeFatJar(t, fat)

	p, err := PrepareAppInput(fat)
	if err != nil {
		t.Fatalf("PrepareAppInput: %v", err)
	}
	defer p.Cleanup()

	if p.ExtractedLibs == "" {
		t.Fatalf("ExtractedLibs is empty, want a directory of extracted dependency jars")
	}

	zr, err := zip.OpenReader(p.AppClassesJar)
	if err != nil {
		t.Fatalf("opening repacked app-classes jar: %v", err)
	}
	defer zr.Close()
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	wantEntry := "com/example/App.class"
	if !contains(names, wantEntry) {
		t.Errorf("repacked jar entries = %v, want it to contain %q (BOOT-INF/classes/ prefix stripped)", names, wantEntry)
	}
	for _, n := range names {
		if n == "BOOT-INF/classes/com/example/App.class" {
			t.Errorf("repacked jar still has the BOOT-INF/classes/ prefix: %q", n)
		}
	}

	libPath := filepath.Join(p.ExtractedLibs, "dep-1.0.jar")
	if _, err := os.Stat(libPath); err != nil {
		t.Errorf("extracted lib jar %s not found: %v", libPath, err)
	}
}

func TestPrepareAppInput_war(t *testing.T) {
	war := filepath.Join(t.TempDir(), "app.war")
	writeZip(t, war, map[string][]byte{
		"WEB-INF/classes/com/example/App.class": []byte("fake app class bytes"),
		"WEB-INF/lib/dep-1.0.jar":               []byte("fake nested jar bytes"),
	})

	p, err := PrepareAppInput(war)
	if err != nil {
		t.Fatalf("PrepareAppInput: %v", err)
	}
	defer p.Cleanup()

	if p.ExtractedLibs == "" {
		t.Fatalf("ExtractedLibs is empty, want a directory of extracted dependency jars")
	}

	names := jarEntryNames(t, p.AppClassesJar)
	if !contains(names, "com/example/App.class") {
		t.Errorf("repacked jar entries = %v, want it to contain com/example/App.class (WEB-INF/classes/ prefix stripped)", names)
	}
	for _, n := range names {
		if n == "WEB-INF/classes/com/example/App.class" {
			t.Errorf("repacked jar still has the WEB-INF/classes/ prefix: %q", n)
		}
	}

	libPath := filepath.Join(p.ExtractedLibs, "dep-1.0.jar")
	if _, err := os.Stat(libPath); err != nil {
		t.Errorf("extracted lib jar %s not found: %v", libPath, err)
	}
}

func TestPrepareAppInput_ear(t *testing.T) {
	warBytes := zipBytes(t, map[string][]byte{
		"WEB-INF/classes/com/example/web/WebApp.class": []byte("fake web app class bytes"),
		"WEB-INF/lib/web-only-dep-1.0.jar":             []byte("fake web-only dep bytes"),
	})

	ear := filepath.Join(t.TempDir(), "app.ear")
	writeZip(t, ear, map[string][]byte{
		"META-INF/application.xml": []byte("<application/>"),
		"app.war":                  warBytes,
		"lib/shared-dep-2.0.jar":   []byte("fake shared dep bytes"),
	})

	p, err := PrepareAppInput(ear)
	if err != nil {
		t.Fatalf("PrepareAppInput: %v", err)
	}
	defer p.Cleanup()

	if p.ExtractedLibs == "" {
		t.Fatalf("ExtractedLibs is empty, want a directory of extracted dependency jars")
	}

	names := jarEntryNames(t, p.AppClassesJar)
	if !contains(names, "com/example/web/WebApp.class") {
		t.Errorf("repacked jar entries = %v, want the nested WAR's class (WEB-INF/classes/ prefix stripped)", names)
	}

	for _, want := range []string{"web-only-dep-1.0.jar", "shared-dep-2.0.jar"} {
		if _, err := os.Stat(filepath.Join(p.ExtractedLibs, want)); err != nil {
			t.Errorf("extracted lib jar %s not found: %v", want, err)
		}
	}
}

// TestPrepareAppInput_ear_detectedWithoutApplicationXml confirms EAR
// detection also works via a nested *.war entry alone, since
// META-INF/application.xml has been optional since Java EE 5.
func TestPrepareAppInput_ear_detectedWithoutApplicationXml(t *testing.T) {
	warBytes := zipBytes(t, map[string][]byte{
		"WEB-INF/classes/com/example/App.class": []byte("fake app class bytes"),
	})

	ear := filepath.Join(t.TempDir(), "app.ear")
	writeZip(t, ear, map[string][]byte{
		"app.war": warBytes,
	})

	p, err := PrepareAppInput(ear)
	if err != nil {
		t.Fatalf("PrepareAppInput: %v", err)
	}
	defer p.Cleanup()

	names := jarEntryNames(t, p.AppClassesJar)
	if !contains(names, "com/example/App.class") {
		t.Errorf("repacked jar entries = %v, want the nested WAR's class even with no application.xml", names)
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// writeFatJar builds a minimal Spring Boot style fat jar: one app class
// under BOOT-INF/classes/ and one nested dependency jar under
// BOOT-INF/lib/.
func writeFatJar(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)

	w, err := zw.Create("BOOT-INF/classes/com/example/App.class")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("fake app class bytes"))

	// A nested "jar" — content doesn't need to be a real jar for this
	// test, since PrepareAppInput only extracts it byte-for-byte.
	w, err = zw.Create("BOOT-INF/lib/dep-1.0.jar")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("fake nested jar bytes"))

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

// zipBytes builds an in-memory zip archive from entries (name -> raw
// content, written byte-for-byte, same as writeFatJar's "fake jar"
// entries — the test never needs these to be real jars) and returns its
// serialized bytes, for embedding as a nested archive (a WAR inside an
// EAR) or for writing straight to disk via writeZip.
func zipBytes(t *testing.T, entries map[string][]byte) []byte {
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
	return buf.Bytes()
}

// writeZip writes entries (see zipBytes) to a new zip file at path.
func writeZip(t *testing.T, path string, entries map[string][]byte) {
	t.Helper()
	if err := os.WriteFile(path, zipBytes(t, entries), 0o644); err != nil {
		t.Fatal(err)
	}
}

// jarEntryNames lists every entry name in the jar at path.
func jarEntryNames(t *testing.T, path string) []string {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("opening %s: %v", path, err)
	}
	defer zr.Close()
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return names
}
