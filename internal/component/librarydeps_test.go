package component

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/input"
)

func pomXML(gav, parent, props, deps, managed string) string {
	p := strings.Split(gav, ":")
	s := "<project><modelVersion>4.0.0</modelVersion>"
	if parent != "" {
		q := strings.Split(parent, ":")
		s += fmt.Sprintf("<parent><groupId>%s</groupId><artifactId>%s</artifactId><version>%s</version></parent>", q[0], q[1], q[2])
	}
	s += fmt.Sprintf("<groupId>%s</groupId><artifactId>%s</artifactId><version>%s</version>", p[0], p[1], p[2])
	s += "<properties>" + props + "</properties><dependencies>" + deps + "</dependencies>"
	if managed != "" {
		s += "<dependencyManagement><dependencies>" + managed + "</dependencies></dependencyManagement>"
	}
	return s + "</project>"
}

func dep(gav string, extra string) string {
	p := strings.Split(gav, ":")
	v := ""
	if len(p) > 2 && p[2] != "" {
		v = "<version>" + p[2] + "</version>"
	}
	return fmt.Sprintf("<dependency><groupId>%s</groupId><artifactId>%s</artifactId>%s%s</dependency>", p[0], p[1], v, extra)
}

func writeLibJar(t *testing.T, entries map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "lib-1.0.jar")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}
	zw.Close()
	f.Close()
	return path
}

func depInstance(id, gav string) Instance {
	p := strings.Split(gav, ":")
	return Instance{ID: id, GAV: &input.GAV{Group: p[0], Artifact: p[1], Version: p[2]}, Detection: DetectedPOMProperties}
}

// TestLibraryDeclarations (LIBRARY_MODE LB3, lb3-1): the consumer tree as
// Maven builds it — nearest wins, exclusions, provided and optional not
// transitive, managed versions and scopes from the parent, a range
// followed through the input's version — and the user-supplied rest.
func TestLibraryDeclarations(t *testing.T) {
	central := map[string]string{
		"p:parent:1": pomXML("p:parent:1", "", "<x.version>2.0</x.version>",
			"", dep("m:managed:3.1", "")+dep("m:scoped:1.0", "<scope>provided</scope>")),
		// a: compile, brings b (compile), c (runtime), an optional d, a test e.
		"g:a:1.0": pomXML("g:a:1.0", "", "", dep("g:b:1.0", "")+dep("g:c:1.0", "<scope>runtime</scope>")+
			dep("g:d:1.0", "<optional>true</optional>")+dep("g:e:1.0", "<scope>test</scope>"), ""),
		// b brings c 9.9 (loses to a's nearer c) and f.
		"g:b:1.0": pomXML("g:b:1.0", "", "", dep("g:c:9.9", "")+dep("g:f:1.0", ""), ""),
		"g:c:1.0": pomXML("g:c:1.0", "", "", "", ""),
		"g:f:1.0": pomXML("g:f:1.0", "", "", "", ""),
		// h is provided by the user; its compile dep i is user-supplied too.
		"g:h:2.0": pomXML("g:h:2.0", "", "", dep("g:i:1.0", "")+dep("g:j:1.0", ""), ""),
		"g:i:1.0": pomXML("g:i:1.0", "", "", "", ""),
		"g:j:1.0": pomXML("g:j:1.0", "", "", "", ""),
		// r: in range [1.0,2.0) — followed through the input's 1.5.
		"g:r:1.5":       pomXML("g:r:1.5", "", "", dep("g:s:1.0", ""), ""),
		"g:s:1.0":       pomXML("g:s:1.0", "", "", "", ""),
		"m:managed:3.1": pomXML("m:managed:3.1", "", "", "", ""),
		"g:x:2.0":       pomXML("g:x:2.0", "", "", "", ""),
	}
	fetch := func(_ context.Context, url string) ([]byte, error) {
		for gav, body := range central {
			if url == POMURL("https://repo.test", parseGAVMust(gav)) {
				return []byte(body), nil
			}
		}
		return nil, fmt.Errorf("404 %s", url)
	}
	lib := pomXML("lib:lib:1.0", "p:parent:1", "<r.range>[1.0,2.0)</r.range>",
		dep("g:a:1.0", "")+
			dep("g:h:2.0", "<scope>provided</scope><exclusions><exclusion><groupId>g</groupId><artifactId>j</artifactId></exclusion></exclusions>")+
			dep("g:r:${r.range}", "")+
			dep("m:managed:", "")+ // version from the parent's dependencyManagement
			dep("m:scoped:1.0", "")+ // scope provided from the parent's dependencyManagement
			dep("g:x:${x.version}", "<optional>true</optional>")+
			dep("g:t:1.0", "<scope>test</scope>")+
			dep("g:u:${missing}", ""),
		"")
	jar := writeLibJar(t, map[string]string{"META-INF/maven/lib/lib/pom.xml": lib})
	inv := &Inventory{Instances: []Instance{
		{ID: "app", Detection: DetectedApp},
		depInstance("ia", "g:a:1.0"),
		depInstance("ic", "g:c:1.1"),                                     // differs from the declared 1.0
		depInstance("ih", "g:h:2.0"),                                     // provided, present
		depInstance("ir", "g:r:1.5"),                                     // in range
		depInstance("iz", "z:extra:1.0"),                                 // undeclared
		{ID: "ib", ArtifactID: "jar-b", Detection: DetectedUnidentified}, // b-1.0.jar without pom.properties
	}, Artifacts: []Artifact{{ID: "jar-b", Location: "deps:b-1.0.jar"}}}
	got := inv.LibraryDeclarations(context.Background(), fetch, "https://repo.test", jar, "lib-1.0.jar", "")
	if got.POM != "lib-1.0.jar!/META-INF/maven/lib/lib/pom.xml" || got.Library != "lib:lib:1.0" {
		t.Fatalf("pom %q library %q", got.POM, got.Library)
	}
	type row struct{ pkg, by, scope, via, version, inst, match string }
	var rows []row
	for _, d := range got.Declared {
		rows = append(rows, row{d.Package, d.SuppliedBy, d.Scope + map[bool]string{true: "+optional"}[d.Optional], strings.Join(d.Via, ">"), d.Version, d.Instance, d.Match})
	}
	want := []row{
		{"g:a", "library", "compile", "", "1.0", "ia", "declared"},
		{"g:b", "library", "compile", "g:a", "1.0", "ib", "unidentified"},
		{"g:c", "library", "runtime", "g:a", "1.0", "ic", "differs"}, // nearer than b's 9.9
		{"g:f", "library", "compile", "g:a>g:b", "1.0", "", "absent"},
		{"g:h", "user", "provided", "", "2.0", "ih", "declared"},
		{"g:i", "user", "provided", "g:h", "1.0", "", "absent"}, // j excluded
		{"g:r", "library", "compile", "", "[1.0,2.0)", "ir", "in_range"},
		{"g:s", "library", "compile", "g:r", "1.0", "", "absent"},
		{"g:u", "library", "compile", "", "", "", "absent"},
		{"g:x", "user", "compile+optional", "", "2.0", "", "absent"},
		{"m:managed", "library", "compile", "", "3.1", "", "absent"},
		{"m:scoped", "user", "provided", "", "1.0", "", "absent"},
		{"z:extra", "undeclared", "", "", "", "iz", ""},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("declared:\n got %v\nwant %v", rows, want)
	}
	if d := got.ForInstance("ih"); d == nil || d.SuppliedBy != SuppliedByUser {
		t.Errorf("ForInstance(ih) = %+v", d)
	}
	if n, absent := got.Count(SuppliedByLibrary); n != 8 || absent != 4 {
		t.Errorf("Count(library) = %d, %d; want 8, 4", n, absent)
	}
	notes := strings.Join(got.Notes, "\n")
	if !strings.Contains(notes, "g:u: version unresolved") {
		t.Errorf("notes lack the unresolved g:u:\n%s", notes)
	}
}

func parseGAVMust(s string) input.GAV {
	p := strings.Split(s, ":")
	return input.GAV{Group: p[0], Artifact: p[1], Version: p[2]}
}

// TestLibraryPOMChoice: a jar with several poms needs one named by the
// jar; none at all is a note, not an error.
func TestLibraryPOMChoice(t *testing.T) {
	two := writeLibJar(t, map[string]string{
		"META-INF/maven/g/lib/pom.xml":    pomXML("g:lib:1.0", "", "", "", ""),
		"META-INF/maven/o/shaded/pom.xml": pomXML("o:shaded:2.0", "", "", "", ""),
	})
	inv := &Inventory{}
	if got := inv.LibraryDeclarations(context.Background(), nil, "", two, "lib-1.0.jar", ""); got.Library != "g:lib:1.0" {
		t.Errorf("named pom: %+v", got)
	}
	if got := inv.LibraryDeclarations(context.Background(), nil, "", two, "other.jar", ""); got.POM != "" || len(got.Notes) != 1 || !strings.Contains(got.Notes[0], "-library-pom") {
		t.Errorf("ambiguous: %+v", got)
	}
	none := writeLibJar(t, map[string]string{"x/Y.class": ""})
	if got := inv.LibraryDeclarations(context.Background(), nil, "", none, "lib.jar", ""); got.POM != "" || len(got.Notes) != 1 {
		t.Errorf("none: %+v", got)
	}
	pf := filepath.Join(t.TempDir(), "pom.xml")
	os.WriteFile(pf, []byte(pomXML("g:given:3.0", "", "", "", "")), 0o644)
	if got := inv.LibraryDeclarations(context.Background(), nil, "", none, "lib.jar", pf); got.Library != "g:given:3.0" || got.POM != pf {
		t.Errorf("-library-pom: %+v", got)
	}
}

// TestLibraryDeclarationsBOM: versions managed by an imported BOM
// (scope import), resolved against the BOM's properties; the pom's own
// management wins over the import.
func TestLibraryDeclarationsBOM(t *testing.T) {
	bom := pomXML("b:bom:1", "", "<k.version>4.2</k.version>", "",
		dep("g:k:${k.version}", "")+dep("g:own:9.9", ""))
	fetch := func(_ context.Context, url string) ([]byte, error) {
		if url == POMURL("https://repo.test", parseGAVMust("b:bom:1")) {
			return []byte(bom), nil
		}
		return nil, fmt.Errorf("404 %s", url)
	}
	lib := pomXML("lib:lib:1.0", "", "", dep("g:k:", "")+dep("g:own:", ""),
		dep("g:own:1.0", "")+dep("b:bom:1", "<type>pom</type><scope>import</scope>"))
	jar := writeLibJar(t, map[string]string{"META-INF/maven/lib/lib/pom.xml": lib})
	got := (&Inventory{}).LibraryDeclarations(context.Background(), fetch, "https://repo.test", jar, "lib-1.0.jar", "")
	versions := map[string]string{}
	for _, d := range got.Declared {
		versions[d.Package] = d.Version
	}
	if want := map[string]string{"g:k": "4.2", "g:own": "1.0"}; !reflect.DeepEqual(versions, want) {
		t.Errorf("versions %v, want %v (notes %v)", versions, want, got.Notes)
	}
}
