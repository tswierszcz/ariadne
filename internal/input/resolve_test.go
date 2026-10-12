package input

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// writeJar builds a minimal jar at path, with the given pom.properties
// entries (mavenPath -> properties content) and optionally some filler
// class entries so the jar isn't trivially empty.
func writeJar(t *testing.T, path string, pomProps map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, content := range pomProps {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
}

func TestResolveDepJars_pomProperties(t *testing.T) {
	dir := t.TempDir()
	writeJar(t, filepath.Join(dir, "snakeyaml-1.27.jar"), map[string]string{
		"META-INF/maven/org.yaml/snakeyaml/pom.properties": "groupId=org.yaml\nartifactId=snakeyaml\nversion=1.27\n",
	})

	got, err := ResolveDepJars(dir, nil)
	if err != nil {
		t.Fatalf("ResolveDepJars: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d DepJars, want 1: %+v", len(got), got)
	}
	dj := got[0]
	if dj.GAV == nil {
		t.Fatalf("GAV is nil, want org.yaml:snakeyaml:1.27")
	}
	want := GAV{Group: "org.yaml", Artifact: "snakeyaml", Version: "1.27"}
	if *dj.GAV != want {
		t.Errorf("GAV = %+v, want %+v", *dj.GAV, want)
	}
	if dj.Source != SourcePOMProperties {
		t.Errorf("Source = %q, want %q", dj.Source, SourcePOMProperties)
	}
	if len(dj.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", dj.Warnings)
	}
}

func TestResolveDepJars_shadedJarWarns(t *testing.T) {
	dir := t.TempDir()
	writeJar(t, filepath.Join(dir, "shaded-1.0.jar"), map[string]string{
		"META-INF/maven/com.example/a/pom.properties": "groupId=com.example\nartifactId=a\nversion=1.0\n",
		"META-INF/maven/com.example/b/pom.properties": "groupId=com.example\nartifactId=b\nversion=2.0\n",
	})

	got, err := ResolveDepJars(dir, nil)
	if err != nil {
		t.Fatalf("ResolveDepJars: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d DepJars, want 1", len(got))
	}
	if len(got[0].Warnings) == 0 {
		t.Errorf("expected a warning about multiple pom.properties, got none")
	}
}

func TestResolveDepJars_shadedJarPicksMatchingFileName(t *testing.T) {
	dir := t.TempDir()
	// Mirrors org.ehcache:ehcache:3.9.2, a fat jar that embeds
	// pom.properties for several internal modules that were never
	// published under their own coordinates. The jar's own name should
	// win over whichever entry happens to be read first.
	writeJar(t, filepath.Join(dir, "ehcache-3.9.2.jar"), map[string]string{
		"META-INF/maven/org.ehcache.modules/ehcache-107/pom.properties": "groupId=org.ehcache.modules\nartifactId=ehcache-107\nversion=3.9.2\n",
		"META-INF/maven/org.ehcache/ehcache/pom.properties":             "groupId=org.ehcache\nartifactId=ehcache\nversion=3.9.2\n",
	})

	got, err := ResolveDepJars(dir, nil)
	if err != nil {
		t.Fatalf("ResolveDepJars: %v", err)
	}
	if len(got) != 1 || got[0].GAV == nil {
		t.Fatalf("got %+v, want one identified DepJar", got)
	}
	want := GAV{Group: "org.ehcache", Artifact: "ehcache", Version: "3.9.2"}
	if *got[0].GAV != want {
		t.Errorf("GAV = %+v, want %+v", *got[0].GAV, want)
	}
}

func TestResolveDepJars_fallsBackToDepsList(t *testing.T) {
	dir := t.TempDir()
	// No pom.properties at all inside this jar.
	writeJar(t, filepath.Join(dir, "weird-2.3.jar"), nil)

	depsList := []GAV{
		{Group: "com.example", Artifact: "weird", Version: "2.3"},
	}
	got, err := ResolveDepJars(dir, depsList)
	if err != nil {
		t.Fatalf("ResolveDepJars: %v", err)
	}
	if len(got) != 1 || got[0].GAV == nil {
		t.Fatalf("got %+v, want one identified DepJar", got)
	}
	if *got[0].GAV != depsList[0] {
		t.Errorf("GAV = %+v, want %+v", *got[0].GAV, depsList[0])
	}
	if got[0].Source != SourceDepsList {
		t.Errorf("Source = %q, want %q", got[0].Source, SourceDepsList)
	}
}

func TestResolveDepJars_unidentified(t *testing.T) {
	dir := t.TempDir()
	writeJar(t, filepath.Join(dir, "mystery-9.9.jar"), nil)

	got, err := ResolveDepJars(dir, nil)
	if err != nil {
		t.Fatalf("ResolveDepJars: %v", err)
	}
	if len(got) != 1 || got[0].GAV != nil {
		t.Fatalf("got %+v, want one unidentified DepJar", got)
	}
}
