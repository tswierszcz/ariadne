package profile

import (
	"os"
	"path/filepath"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

func buildFixture(t *testing.T) (*hierarchy.Hierarchy, *callgraph.Graph) {
	t.Helper()
	h, warnings, err := hierarchy.Build("../../testdata/fixture/fixture-8.jar", nil)
	if err != nil {
		t.Fatalf("hierarchy.Build: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	return h, callgraph.Build(h, nil)
}

func TestLoadDefaultCatalog(t *testing.T) {
	c, err := LoadDefaultCatalog()
	if err != nil {
		t.Fatalf("LoadDefaultCatalog: %v", err)
	}
	spring, ok := c.Get("spring")
	if !ok {
		t.Fatal(`catalog has no "spring" profile`)
	}
	if len(spring.Detect) == 0 {
		t.Error("spring profile has no detect markers")
	}
	if len(spring.FrameworkPackages) == 0 {
		t.Error("spring profile has no framework packages")
	}
	if len(spring.SyntheticEdges) == 0 {
		t.Error("spring profile has no synthetic edges")
	}
	generic, ok := c.Get("generic")
	if !ok {
		t.Fatal(`catalog has no "generic" fallback profile`)
	}
	if len(generic.Detect) != 0 {
		t.Error("generic profile should never auto-detect (it's the fallback)")
	}
	if len(generic.FrameworkPackages) != 0 {
		t.Error("generic profile should have no framework packages")
	}
}

func TestFrameworkPackage_UnmarshalsScalarAndConditional(t *testing.T) {
	c, err := LoadDefaultCatalog()
	if err != nil {
		t.Fatalf("LoadDefaultCatalog: %v", err)
	}
	spring, _ := c.Get("spring")

	var plain, conditional *FrameworkPackage
	for i, fp := range spring.FrameworkPackages {
		fp := fp
		if fp.Pattern == "org.hibernate:*" {
			plain = &spring.FrameworkPackages[i]
		}
		if fp.Pattern == "com.fasterxml.jackson.core:jackson-databind" {
			conditional = &spring.FrameworkPackages[i]
		}
	}
	if plain == nil {
		t.Fatal("org.hibernate:* (a bare scalar entry) did not parse")
	}
	if len(plain.WhenAnnotation) != 0 {
		t.Errorf("a bare scalar entry should have no WhenAnnotation, got %v", plain.WhenAnnotation)
	}
	if conditional == nil {
		t.Fatal("jackson-databind (a {pkg, when_annotation} entry) did not parse")
	}
	if fp := conditional; len(fp.WhenAnnotation) == 0 {
		t.Error("jackson-databind's when_annotation did not parse")
	}
}

func TestSelect_explicitName(t *testing.T) {
	c, err := LoadDefaultCatalog()
	if err != nil {
		t.Fatalf("LoadDefaultCatalog: %v", err)
	}
	h, _ := buildFixture(t)

	p, err := Select(c, "generic", h)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if p.Name != "generic" {
		t.Errorf("Name = %q, want generic", p.Name)
	}

	if _, err := Select(c, "no-such-profile", h); err == nil {
		t.Error("Select with an unknown profile name should fail")
	}
}

func TestSelect_autoFallsBackToGeneric(t *testing.T) {
	c, err := LoadDefaultCatalog()
	if err != nil {
		t.Fatalf("LoadDefaultCatalog: %v", err)
	}
	// The fixture has none of spring's detect markers at all.
	h, _ := buildFixture(t)

	p, err := Select(c, "auto", h)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if p.Name != "generic" {
		t.Errorf("Name = %q, want generic (nothing in the fixture matches spring's detect markers)", p.Name)
	}
}

func TestSelect_autoDetectsAndMerges(t *testing.T) {
	// A synthetic catalog whose detect markers are classes the fixture
	// jar actually has (fixture.App, fixture.Util), so auto-detection has
	// something real to match without needing an actual Spring Boot jar
	// on hand.
	var c Catalog
	c.put(Profile{Name: "a", Detect: []string{"fixture.App"}, FrameworkPackages: []FrameworkPackage{{Pattern: "g1:a1"}}})
	c.put(Profile{Name: "b", Detect: []string{"fixture.Util"}, FrameworkPackages: []FrameworkPackage{{Pattern: "g2:a2"}}})
	c.put(Profile{Name: "c", Detect: []string{"no.such.Class"}, FrameworkPackages: []FrameworkPackage{{Pattern: "g3:a3"}}})
	c.put(Profile{Name: "generic"})

	h, _ := buildFixture(t)
	p, err := Select(c, "auto", h)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if p.Name != "a+b" {
		t.Errorf("Name = %q, want \"a+b\" (a and b's markers are present, c's is not)", p.Name)
	}
	got := p.ResolvePatterns(h)
	want := []string{"g1:a1", "g2:a2"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("ResolvePatterns = %v, want %v (merged from both matched profiles, c's not included)", got, want)
	}
}

func TestResolvePatterns_conditionalPattern(t *testing.T) {
	h, _ := buildFixture(t)

	// fixture.Entry is the fixture's own stand-in "route annotation"
	// (see testdata/fixture/src/fixture/Entry.java); fixture.App#handle
	// carries it, so a pattern gated on it should be included...
	present := Profile{FrameworkPackages: []FrameworkPackage{
		{Pattern: "g:a", WhenAnnotation: []string{"fixture.Entry"}},
	}}
	if got := present.ResolvePatterns(h); len(got) != 1 {
		t.Errorf("ResolvePatterns = %v, want [g:a] (fixture.App#handle carries @Entry)", got)
	}

	// ...but a pattern gated on an annotation nothing in the fixture
	// carries should be left out.
	absent := Profile{FrameworkPackages: []FrameworkPackage{
		{Pattern: "g:a", WhenAnnotation: []string{"no.such.Annotation"}},
	}}
	if got := absent.ResolvePatterns(h); len(got) != 0 {
		t.Errorf("ResolvePatterns = %v, want [] (nothing in the fixture carries this annotation)", got)
	}
}

func TestAppRoots(t *testing.T) {
	h, g := buildFixture(t)

	prof := Profile{Roots: Roots{Annotations: []string{"fixture.Entry"}, Main: true}}
	roots := prof.AppRoots(g, h)

	want := map[string]bool{
		"fixture.App#handle(Ljava/lang/String;)Ljava/lang/String;": true, // carries @Entry
		"fixture.App#main([Ljava/lang/String;)V":                   true, // Roots.Main
	}
	got := map[string]bool{}
	for _, id := range roots {
		got[g.Key(id)] = true
	}
	for k := range want {
		if !got[k] {
			t.Errorf("AppRoots is missing expected root %q; got %v", k, got)
		}
	}
	if len(got) != len(want) {
		t.Errorf("AppRoots = %v, want exactly %v (greetWith/greet etc. are neither annotated nor main)", got, want)
	}
}

func TestAppRoots_overrides(t *testing.T) {
	h, g := buildFixture(t)

	// fixture.Greeter#greet has two app implementations (FormalGreeter,
	// CasualGreeter), neither named in Roots.Overrides directly -- both
	// should become roots via hierarchy.Subtypes, the same way an app's
	// own HttpServlet subclass would for a real servlet profile's
	// "javax.servlet.http.HttpServlet#doGet".
	prof := Profile{Roots: Roots{Overrides: []string{"fixture.Greeter#greet"}}}
	roots := prof.AppRoots(g, h)

	want := map[string]bool{
		"fixture.FormalGreeter#greet(Ljava/lang/String;)Ljava/lang/String;": true,
		"fixture.CasualGreeter#greet(Ljava/lang/String;)Ljava/lang/String;": true,
	}
	got := map[string]bool{}
	for _, id := range roots {
		got[g.Key(id)] = true
	}
	if len(got) != len(want) {
		t.Fatalf("AppRoots = %v, want exactly %v", got, want)
	}
	for k := range want {
		if !got[k] {
			t.Errorf("AppRoots is missing expected override root %q; got %v", k, got)
		}
	}
}

func TestAppRoots_overrides_malformedSpecIgnored(t *testing.T) {
	h, g := buildFixture(t)
	prof := Profile{Roots: Roots{Overrides: []string{"no-hash-sign-here"}}}
	if roots := prof.AppRoots(g, h); len(roots) != 0 {
		t.Errorf("AppRoots = %v, want none (a spec with no '#' should be ignored, not panic)", roots)
	}
}

func TestAppRoots_mainFalseExcludesMain(t *testing.T) {
	h, g := buildFixture(t)
	prof := Profile{Roots: Roots{Annotations: []string{"fixture.Entry"}, Main: false}}
	roots := prof.AppRoots(g, h)
	for _, id := range roots {
		if g.Key(id) == "fixture.App#main([Ljava/lang/String;)V" {
			t.Error("AppRoots included main even though Roots.Main is false")
		}
	}
}

func TestCatalogAddFile_overridesByName(t *testing.T) {
	var c Catalog
	c.put(Profile{Name: "spring", Detect: []string{"old.Marker"}})

	dir := t.TempDir()
	path := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(path, []byte("name: spring\ndetect: [\"new.Marker\"]\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := c.AddFile(path); err != nil {
		t.Fatalf("AddFile: %v", err)
	}
	p, ok := c.Get("spring")
	if !ok {
		t.Fatal(`"spring" missing after AddFile`)
	}
	if len(p.Detect) != 1 || p.Detect[0] != "new.Marker" {
		t.Errorf("Detect = %v, want [new.Marker] (AddFile should replace, not merge)", p.Detect)
	}
	if got := c.Names(); len(got) != 1 {
		t.Errorf("Names() = %v, want exactly one entry (replacing an existing name must not add a duplicate)", got)
	}
}
