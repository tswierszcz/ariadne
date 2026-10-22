package hierarchy

import (
	"archive/zip"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"ariadne/internal/input"
)

func buildFixture(t *testing.T) *Hierarchy {
	t.Helper()
	h, warnings, err := Build("../../testdata/fixture/fixture-8.jar", nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	return h
}

func TestBuild_readsAllFixtureClasses(t *testing.T) {
	h := buildFixture(t)
	for _, fqcn := range []string{"fixture.App", "fixture.Greeter", "fixture.FormalGreeter", "fixture.CasualGreeter", "fixture.Util", "fixture.Entry"} {
		if _, ok := h.Lookup(fqcn); !ok {
			t.Errorf("Lookup(%q) not found", fqcn)
		}
	}
}

func TestSubtypes_interfaceToImplementations(t *testing.T) {
	h := buildFixture(t)
	got := h.Subtypes("fixture.Greeter")
	sort.Strings(got)
	want := []string{"fixture.CasualGreeter", "fixture.FormalGreeter"}
	if len(got) != len(want) {
		t.Fatalf("Subtypes(fixture.Greeter) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Subtypes(fixture.Greeter)[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSubtypes_jdkTypeIsEmpty(t *testing.T) {
	h := buildFixture(t)
	// java.lang.Runnable was never read (it's a JDK type, not part of the
	// app or its dependencies), so it must resolve to no known subtypes
	// rather than error — this is exactly the case section 3.6's CHA
	// skip-list exists to avoid expanding in the first place.
	if got := h.Subtypes("java.lang.Runnable"); len(got) != 0 {
		t.Errorf("Subtypes(java.lang.Runnable) = %v, want empty", got)
	}
}

func TestResolveMethod_declaredDirectly(t *testing.T) {
	h := buildFixture(t)
	declaring, found := h.ResolveMethod("fixture.FormalGreeter", "greet")
	if !found || declaring != "fixture.FormalGreeter" {
		t.Errorf("ResolveMethod(FormalGreeter, greet) = (%q, %v), want (fixture.FormalGreeter, true)", declaring, found)
	}
}

func TestResolveMethod_inheritedFromInterface(t *testing.T) {
	h := buildFixture(t)
	// fixture.Greeter itself only declares an abstract "greet"; resolving
	// a call ON the interface type should find that declaration.
	declaring, found := h.ResolveMethod("fixture.Greeter", "greet")
	if !found || declaring != "fixture.Greeter" {
		t.Errorf("ResolveMethod(Greeter, greet) = (%q, %v), want (fixture.Greeter, true)", declaring, found)
	}
}

func TestResolveMethod_notFound(t *testing.T) {
	h := buildFixture(t)
	_, found := h.ResolveMethod("fixture.FormalGreeter", "noSuchMethod")
	if found {
		t.Errorf("ResolveMethod(FormalGreeter, noSuchMethod) unexpectedly found")
	}
}

func TestBuild_duplicateClassWarnsAndDedupes(t *testing.T) {
	// Reading the same jar as both the app and (artificially) one of its
	// own dependencies should warn about every duplicate FQCN and still
	// produce exactly one entry per class, not an error and not two.
	solo, _, err := Build("../../testdata/fixture/fixture-8.jar", nil)
	if err != nil {
		t.Fatalf("Build (solo): %v", err)
	}

	dup, warnings, err := Build("../../testdata/fixture/fixture-8.jar", []input.DepJar{
		{Path: "../../testdata/fixture/fixture-8.jar"},
	})
	if err != nil {
		t.Fatalf("Build (dup): %v", err)
	}
	if len(warnings) != solo.Len() {
		t.Errorf("got %d duplicate warnings, want %d (one per class)", len(warnings), solo.Len())
	}
	if dup.Len() != solo.Len() {
		t.Errorf("Build with a duplicate dep jar has %d classes, want %d (deduplicated)", dup.Len(), solo.Len())
	}

	// JVM_GRAPH G1: the plain-string warnings are a rendering of
	// structured evidence (Duplicates()), not the only record of what
	// happened.
	dups := dup.Duplicates()
	if len(dups) != solo.Len() {
		t.Fatalf("Duplicates() has %d entries, want %d (one per class)", len(dups), solo.Len())
	}
	for _, d := range dups {
		if d.FQCN == "" {
			t.Errorf("Duplicate has empty FQCN: %+v", d)
		}
		if d.Kept.SHA256 == "" || d.Shadowed.SHA256 == "" {
			t.Errorf("Duplicate %s: Kept/Shadowed SHA256 not hashed: %+v", d.FQCN, d)
		}
		// Both occurrences are literally the same file in this test, so
		// their content hashes must agree even though they came from two
		// different SourceRefs.
		if d.Kept.SHA256 != d.Shadowed.SHA256 {
			t.Errorf("Duplicate %s: Kept/Shadowed SHA256 differ (%s vs %s) for the same underlying file",
				d.FQCN, d.Kept.SHA256, d.Shadowed.SHA256)
		}
	}
}

// TestBuild_duplicateWinnerFollowsExplicitOrderNotReadOrder is JVM_GRAPH
// G1's own "Gotowe" criterion: which duplicate occurrence wins must
// depend only on the explicit order dependency jars were given in, not
// on which worker happened to read its jar first. tmpA and tmpB are two
// on-disk copies of the same fixture jar at two different paths (two
// "artifacts" that happen to collide on every class), so the winner can
// only be told apart by path — swapping their order in depJars must
// swap which path's copy of fixture.App is kept, every time.
func TestBuild_duplicateWinnerFollowsExplicitOrderNotReadOrder(t *testing.T) {
	tmpA := copyFixtureJar(t, "a")
	tmpB := copyFixtureJar(t, "b")
	const fqcn = "fixture.App"

	forward, _, err := Build(tmpA, []input.DepJar{{Path: tmpB}})
	if err != nil {
		t.Fatalf("Build (A app, B dep): %v", err)
	}
	ci, ok := forward.Lookup(fqcn)
	if !ok {
		t.Fatalf("%s not found", fqcn)
	}
	if ci.SourceJar != tmpA {
		t.Errorf("app job isn't index 0 in job order: kept SourceJar = %s, want the app jar %s", ci.SourceJar, tmpA)
	}

	// Two dependency jars, compared both ways: the app job is always
	// index 0 and would otherwise win every collision regardless of
	// dependency order, so use an empty app jar (no classes at all) to
	// isolate dependency-vs-dependency ordering specifically.
	emptyApp := emptyJar(t)

	depsAB, _, err := Build(emptyApp, []input.DepJar{{Path: tmpA}, {Path: tmpB}})
	if err != nil {
		t.Fatalf("Build (deps A,B): %v", err)
	}
	ciAB, _ := depsAB.Lookup(fqcn)
	if ciAB.SourceJar != tmpA {
		t.Errorf("deps=[A,B]: kept SourceJar = %s, want first dep jar %s", ciAB.SourceJar, tmpA)
	}

	depsBA, _, err := Build(emptyApp, []input.DepJar{{Path: tmpB}, {Path: tmpA}})
	if err != nil {
		t.Fatalf("Build (deps B,A): %v", err)
	}
	ciBA, _ := depsBA.Lookup(fqcn)
	if ciBA.SourceJar != tmpB {
		t.Errorf("deps=[B,A]: kept SourceJar = %s, want first dep jar %s (reversing the list must flip the winner)", ciBA.SourceJar, tmpB)
	}
}

// copyFixtureJar copies the fixture jar to a fresh path under t.TempDir()
// (named with suffix so two calls in one test don't collide), so a test
// can give Build two or three distinct jar *paths* that are nonetheless
// byte-identical in content — isolating "which path wins" from "which
// content wins", which a single shared path (as
// TestBuild_duplicateClassWarnsAndDedupes uses) can't do.
func copyFixtureJar(t *testing.T, suffix string) string {
	t.Helper()
	src, err := os.ReadFile("../../testdata/fixture/fixture-8.jar")
	if err != nil {
		t.Fatalf("reading fixture jar: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "fixture-"+suffix+".jar")
	if err := os.WriteFile(dst, src, 0o644); err != nil {
		t.Fatalf("writing %s: %v", dst, err)
	}
	return dst
}

// emptyJar writes a valid, empty zip archive (no entries at all) and
// returns its path — a stand-in "app" with no classes of its own, so a
// test can isolate dependency-vs-dependency collision ordering without
// the always-first app job winning every collision by construction.
func emptyJar(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "empty.jar")
	f, err := os.Create(dst)
	if err != nil {
		t.Fatalf("creating %s: %v", dst, err)
	}
	if err := zip.NewWriter(f).Close(); err != nil {
		t.Fatalf("writing empty zip %s: %v", dst, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing %s: %v", dst, err)
	}
	return dst
}

func TestLen(t *testing.T) {
	h := buildFixture(t)
	if h.Len() != len(h.All()) {
		t.Errorf("Len() = %d, All() has %d entries", h.Len(), len(h.All()))
	}
	if h.Len() == 0 {
		t.Errorf("Len() = 0, want > 0")
	}
}
