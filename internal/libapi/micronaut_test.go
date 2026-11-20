package libapi

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const micronautFixture = "../../testdata/library-micronaut/"

var micronautDefs = []string{
	"mn.internal.$AuditListener$ApplicationEventListener$onStartup1$Intercepted$Definition",
	"mn.internal.$AuditListener$Definition",
	"mn.internal.$CachedStore$Definition",
	"mn.internal.$Producers$Banner0$Definition",
	"mn.internal.$Producers$Definition",
	"mn.internal.$Wiring$Definition",
}

// remicronaut copies the fixture's lib.jar without its
// META-INF/micronaut/ listing; services adds the 2.x ServiceLoader file
// listing the same references instead.
func remicronaut(t *testing.T, services bool) string {
	t.Helper()
	zr, err := zip.OpenReader(micronautFixture + "lib.jar")
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
	var refs []string
	for _, e := range zr.File {
		if strings.HasPrefix(e.Name, micronautRefDir) {
			if !strings.HasSuffix(e.Name, "/") {
				refs = append(refs, path.Base(e.Name))
			}
			continue
		}
		w, _ := zw.Create(e.Name)
		rc, _ := e.Open()
		io.Copy(w, rc)
		rc.Close()
	}
	if services {
		w, _ := zw.Create(micronautServices)
		io.WriteString(w, "# references\n"+strings.Join(refs, "\n")+"\n")
	}
	zw.Close()
	f.Close()
	return out
}

func micronautClasses(t *testing.T, jar string) []string {
	t.Helper()
	anns, err := MicronautAnnouncements(jar)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, a := range anns {
		if a.Key != MicronautKey || !a.isMicronaut() || a.isCDI() || SpringAnnouncement(a) {
			t.Errorf("announcement %+v", a)
		}
		got = append(got, a.Class)
	}
	return got
}

// TestMicronautAnnouncements on testdata/library-micronaut: the six
// definitions Micronaut 3.9.7 lists under META-INF/micronaut/ (the
// reference class without "$Reference"), the same from a 2.x services
// file, and nothing from a jar listing none.
func TestMicronautAnnouncements(t *testing.T) {
	if got := micronautClasses(t, micronautFixture+"lib.jar"); !reflect.DeepEqual(got, micronautDefs) {
		t.Errorf("3.x listing:\n got %v\nwant %v", got, micronautDefs)
	}
	if got := micronautClasses(t, remicronaut(t, true)); !reflect.DeepEqual(got, micronautDefs) {
		t.Errorf("2.x services file:\n got %v\nwant %v", got, micronautDefs)
	}
	if got := micronautClasses(t, remicronaut(t, false)); len(got) != 0 {
		t.Errorf("no listing: %v", got)
	}
	anns, _ := MicronautAnnouncements(micronautFixture + "lib.jar")
	if f := anns[2].File; f != micronautRefDir+"mn.internal.$CachedStore$Definition$Reference" {
		t.Errorf("file %q", f)
	}
}

// TestMicronautRoots: the entries are what the definitions' generated
// code calls — constructors, the factory method, @PostConstruct, @Inject
// and the @EventListener method through $Exec#dispatch — plus the beans'
// public instance methods; never Expand#expand, Unmanaged or a generated
// class. -api-exclude takes them all out.
func TestMicronautRoots(t *testing.T) {
	jar := micronautFixture + "lib.jar"
	h, _, err := hierarchy.Build(jar, []input.DepJar{{Path: fixture + "deps/commons-text-1.9.jar"}})
	if err != nil {
		t.Fatal(err)
	}
	anns, err := MicronautAnnouncements(jar)
	if err != nil {
		t.Fatal(err)
	}
	g := callgraph.Build(h, callgraph.DefaultSkip)
	roots, prov, classes := AnnouncedRoots(g, h, anns, Options{})
	want := []string{
		"mn.internal.AuditListener#<init>",
		"mn.internal.AuditListener#onStartup",
		"mn.internal.CachedStore#<init>",
		"mn.internal.CachedStore#get",
		"mn.internal.CachedStore#warm",
		"mn.internal.Producers#<init>",
		"mn.internal.Producers#banner",
		"mn.internal.Wiring#<init>",
		"mn.internal.Wiring#describe",
		"mn.internal.Wiring#init",
	}
	if got := names(g, roots); !reflect.DeepEqual(got, want) || classes != 4 {
		t.Errorf("roots (%d classes)\n got %v\nwant %v", classes, got, want)
	}
	id, _ := g.ID("mn.internal.Producers", "banner", "()Lmn/api/Banner;")
	if p := prov[id]; len(p) != 1 || p[0] != micronautRefDir+"mn.internal.$Producers$Banner0$Definition$Reference "+MicronautKey+": mn.internal.$Producers$Banner0$Definition" {
		t.Errorf("provenance %v", p)
	}
	if r, _, n := AnnouncedRoots(g, h, anns, Options{Exclude: []string{"mn.internal"}}); len(r) != 0 || n != 0 {
		t.Errorf("-api-exclude mn.internal: %d roots, %d classes", len(r), n)
	}
}
