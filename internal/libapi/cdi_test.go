package libapi

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const cdiFixture = "../../testdata/library-cdi/"

// TestDiscoveryMode: beans.xml's explicit mode wins; without one an empty
// or pre-4.0 file is "all", a 4.0+ one "annotated".
func TestDiscoveryMode(t *testing.T) {
	for in, want := range map[string]string{
		"":                       "all",
		"<beans/>":               "all",
		`<beans version="2.0"/>`: "all",
		`<beans version="4.0"/>`: "annotated",
		`<beans bean-discovery-mode="ALL" version="4.0"/>`:         "all",
		`<beans bean-discovery-mode="annotated"/>`:                 "annotated",
		`<?xml version="1.0"?><beans bean-discovery-mode="none"/>`: "none",
		"not xml <": "all",
	} {
		if got := discoveryMode([]byte(in)); got != want {
			t.Errorf("discoveryMode(%q) = %q, want %q", in, got, want)
		}
	}
}

// rejar copies the fixture's lib.jar with its META-INF/beans.xml replaced
// (beans == nil drops it) and, if jandex, an empty Jandex index added.
func rejar(t *testing.T, beans []byte, jandex bool) string {
	t.Helper()
	zr, err := zip.OpenReader(cdiFixture + "lib.jar")
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	out := filepath.Join(t.TempDir(), "lib.jar")
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for _, e := range zr.File {
		if e.Name == beansXML {
			continue
		}
		w, _ := zw.Create(e.Name)
		rc, _ := e.Open()
		io.Copy(w, rc)
		rc.Close()
	}
	if beans != nil {
		w, _ := zw.Create(beansXML)
		w.Write(beans)
	}
	if jandex {
		zw.Create(jandexIdx)
	}
	zw.Close()
	f.Close()
	return out
}

func cdiClasses(t *testing.T, jar string) (map[string]string, *hierarchy.Hierarchy) {
	t.Helper()
	h, _, err := hierarchy.Build(jar, []input.DepJar{{Path: fixture + "deps/commons-text-1.9.jar"}})
	if err != nil {
		t.Fatal(err)
	}
	anns, err := CDIAnnouncements(jar, h)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, a := range anns {
		got[a.Class] = a.File + " " + a.Key
	}
	return got, h
}

// TestCDIAnnouncements on testdata/library-cdi: annotated mode gives the
// four beans (never Unmanaged); "all" every concrete class with a
// no-argument constructor; a Jandex index alone is annotated; "none" and
// a plain jar nothing.
func TestCDIAnnouncements(t *testing.T) {
	const ann = "META-INF/beans.xml bean-discovery-mode=annotated @jakarta.enterprise.context."
	got, _ := cdiClasses(t, cdiFixture+"lib.jar")
	want := map[string]string{
		"cd.internal.AuditObserver": ann + "Dependent",
		"cd.internal.CachedStore":   ann + "ApplicationScoped",
		"cd.internal.Producers":     ann + "ApplicationScoped",
		"cd.internal.Wiring":        ann + "Dependent",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("annotated:\n got %v\nwant %v", got, want)
	}
	all := "META-INF/beans.xml bean-discovery-mode=all"
	got, _ = cdiClasses(t, rejar(t, []byte(`<beans bean-discovery-mode="all"/>`), false))
	want = map[string]string{}
	for _, c := range []string{"cd.api.Startup", "cd.internal.AuditObserver", "cd.internal.CachedStore", "cd.internal.Expand",
		"cd.internal.Producers", "cd.internal.Unmanaged", "cd.internal.Wiring"} {
		want[c] = all
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("all:\n got %v\nwant %v", got, want)
	}
	got, _ = cdiClasses(t, rejar(t, nil, true))
	if len(got) != 4 || got["cd.internal.Producers"] != "META-INF/jandex.idx bean-discovery-mode=annotated @jakarta.enterprise.context.ApplicationScoped" {
		t.Errorf("jandex only: %v", got)
	}
	if got, _ = cdiClasses(t, rejar(t, []byte(`<beans bean-discovery-mode="none"/>`), false)); len(got) != 0 {
		t.Errorf("none: %v", got)
	}
	if got, _ = cdiClasses(t, rejar(t, nil, false)); len(got) != 0 {
		t.Errorf("no bean archive: %v", got)
	}
}

// TestCDIRoots: a bean's entries are its constructors, the methods the
// container calls whatever their access, and its public instance
// methods — never Expand#expand or Unmanaged's.
func TestCDIRoots(t *testing.T) {
	jar := cdiFixture + "lib.jar"
	h, _, err := hierarchy.Build(jar, []input.DepJar{{Path: fixture + "deps/commons-text-1.9.jar"}})
	if err != nil {
		t.Fatal(err)
	}
	anns, err := CDIAnnouncements(jar, h)
	if err != nil {
		t.Fatal(err)
	}
	g := callgraph.Build(h, callgraph.DefaultSkip)
	roots, prov, classes := AnnouncedRoots(g, h, anns, Options{})
	want := []string{
		"cd.internal.AuditObserver#<init>",
		"cd.internal.AuditObserver#onStartup",
		"cd.internal.CachedStore#<init>",
		"cd.internal.CachedStore#get",
		"cd.internal.CachedStore#warm",
		"cd.internal.Producers#<init>",
		"cd.internal.Producers#banner",
		"cd.internal.Wiring#<init>",
		"cd.internal.Wiring#init",
		"cd.internal.Wiring#name",
	}
	if got := names(g, roots); !reflect.DeepEqual(got, want) || classes != 4 {
		t.Errorf("roots (%d classes)\n got %v\nwant %v", classes, got, want)
	}
	id, _ := g.ID("cd.internal.Producers", "banner", "()Lcd/api/Banner;")
	if p := prov[id]; len(p) != 1 || p[0] != "META-INF/beans.xml bean-discovery-mode=annotated @jakarta.enterprise.context.ApplicationScoped: cd.internal.Producers" {
		t.Errorf("provenance %v", p)
	}
	if r, _, n := AnnouncedRoots(g, h, anns, Options{Exclude: []string{"cd.internal"}}); len(r) != 0 || n != 0 {
		t.Errorf("-api-exclude cd.internal: %d roots, %d classes", len(r), n)
	}
}
