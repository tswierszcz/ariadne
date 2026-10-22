package serviceload

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

func writeZip(t *testing.T, path string, entries map[string][]byte) {
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

func TestInventory_foundAndMissingProviders(t *testing.T) {
	appJar := filepath.Join(t.TempDir(), "app.jar")
	writeZip(t, appJar, map[string][]byte{}) // no services of its own

	depJar := filepath.Join(t.TempDir(), "dep.jar")
	writeZip(t, depJar, map[string][]byte{
		"META-INF/services/test.Foo": []byte("test.FooImpl\ntest.Missing\n"),
	})

	// A hierarchy that knows about test.FooImpl but not test.Missing —
	// built by hand (hierarchy.NewForTesting), independent of these
	// synthetic jars' own (empty/irrelevant) .class content, exactly
	// like S0's "cross-reference against the analyzed hierarchy" needs.
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.FooImpl", SuperClass: "java.lang.Object"},
	})

	gav := &input.GAV{Group: "g", Artifact: "a", Version: "1"}
	regs, warnings, err := Inventory(h, appJar, []input.DepJar{{Path: depJar, GAV: gav}})
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(regs) != 2 {
		t.Fatalf("got %d registrations, want 2: %+v", len(regs), regs)
	}

	byProvider := make(map[string]ServiceRegistration)
	for _, r := range regs {
		byProvider[r.ProviderClass] = r
	}

	found, ok := byProvider["test.FooImpl"]
	if !ok {
		t.Fatalf("no registration for test.FooImpl")
	}
	if !found.ProviderFound {
		t.Errorf("test.FooImpl: ProviderFound = false, want true")
	}
	if found.Interface != "test.Foo" {
		t.Errorf("test.FooImpl: Interface = %q, want test.Foo", found.Interface)
	}
	if found.GAV != gav {
		t.Errorf("test.FooImpl: GAV = %v, want %v", found.GAV, gav)
	}
	if found.SourceFile != "META-INF/services/test.Foo" {
		t.Errorf("test.FooImpl: SourceFile = %q", found.SourceFile)
	}
	if found.SourceSHA256 == "" {
		t.Errorf("test.FooImpl: SourceSHA256 is empty")
	}

	missing, ok := byProvider["test.Missing"]
	if !ok {
		t.Fatalf("no registration for test.Missing")
	}
	if missing.ProviderFound {
		t.Errorf("test.Missing: ProviderFound = true, want false (not in the analyzed hierarchy)")
	}

	// A missing-provider warning must exist and name the class — this is
	// S0's own "diagnostyka", not a silent gap.
	var sawMissingWarning bool
	for _, w := range warnings {
		if strings.Contains(w, "test.Missing") && strings.Contains(w, "not found") {
			sawMissingWarning = true
		}
	}
	if !sawMissingWarning {
		t.Errorf("warnings = %v, want one mentioning test.Missing not found", warnings)
	}
}

func TestInventory_noServiceFilesAtAll(t *testing.T) {
	appJar := filepath.Join(t.TempDir(), "app.jar")
	writeZip(t, appJar, map[string][]byte{"com/example/App.class": []byte("irrelevant")})

	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "com.example.App", SuperClass: "java.lang.Object"},
	})

	regs, warnings, err := Inventory(h, appJar, nil)
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(regs) != 0 {
		t.Errorf("got %d registrations, want 0: %+v", len(regs), regs)
	}
	if len(warnings) != 0 {
		t.Errorf("got %d warnings, want 0: %v", len(warnings), warnings)
	}
}

func TestInventory_malformedEntryDoesNotCreateARegistration(t *testing.T) {
	appJar := filepath.Join(t.TempDir(), "app.jar")
	writeZip(t, appJar, map[string][]byte{})

	depJar := filepath.Join(t.TempDir(), "dep.jar")
	writeZip(t, depJar, map[string][]byte{
		"META-INF/services/test.Foo": []byte("not a valid name!!\n"),
	})

	h := hierarchy.NewForTesting(nil)
	regs, warnings, err := Inventory(h, appJar, []input.DepJar{{Path: depJar}})
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(regs) != 0 {
		t.Errorf("got %d registrations for a malformed entry, want 0: %+v", len(regs), regs)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v, want exactly 1 (the malformed name)", warnings)
	}
}
