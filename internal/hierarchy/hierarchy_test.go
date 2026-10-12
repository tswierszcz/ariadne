package hierarchy

import (
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
