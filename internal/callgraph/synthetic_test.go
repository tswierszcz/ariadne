package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestApplySyntheticEdges(t *testing.T) {
	h, warnings, err := hierarchy.Build("../../testdata/fixture/fixture-8.jar", nil)
	if err != nil {
		t.Fatalf("hierarchy.Build: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	g := Build(h, nil)

	// fixture.Greeter's subtypes are CasualGreeter and FormalGreeter
	// (see build_test.go); modeling FormalGreeter as the "proxy target"
	// stands in for the Spring Data case (every Repository subtype's
	// interface method really dispatches, at runtime, to
	// SimpleJpaRepository's implementation).
	added := ApplySyntheticEdges(g, h, []SyntheticEdge{
		{FromSubtypesOf: "fixture.Greeter", To: "fixture.FormalGreeter"},
	})
	// One edge per (subtype method, same-named To method) pair: greet and
	// <init> both exist on both CasualGreeter and FormalGreeter (which is
	// a subtype of Greeter too, so it also gets a self-edge), for 2x2 = 4.
	if added != 4 {
		t.Errorf("ApplySyntheticEdges added %d edges, want 4", added)
	}

	greetDesc := "(Ljava/lang/String;)Ljava/lang/String;"
	casualID, ok := g.ID("fixture.CasualGreeter", "greet", greetDesc)
	if !ok {
		t.Fatal("fixture.CasualGreeter#greet not found")
	}
	formalID, ok := g.ID("fixture.FormalGreeter", "greet", greetDesc)
	if !ok {
		t.Fatal("fixture.FormalGreeter#greet not found")
	}

	var found bool
	for _, e := range g.Edges(casualID) {
		if e.To == formalID && e.Kind == classfile.CallSynthetic {
			found = true
		}
	}
	if !found {
		t.Error("expected a synthetic edge CasualGreeter#greet -> FormalGreeter#greet")
	}
}

func TestApplySyntheticEdges_unknownToIsSkippedNotAnError(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/fixture/fixture-8.jar", nil)
	if err != nil {
		t.Fatalf("hierarchy.Build: %v", err)
	}
	g := Build(h, nil)

	added := ApplySyntheticEdges(g, h, []SyntheticEdge{
		{FromSubtypesOf: "fixture.Greeter", To: "no.such.Class"},
	})
	if added != 0 {
		t.Errorf("added = %d, want 0 (the To class isn't part of the analyzed universe)", added)
	}
}

func TestApplySyntheticEdges_unknownFromSubtypesOfIsSkipped(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/fixture/fixture-8.jar", nil)
	if err != nil {
		t.Fatalf("hierarchy.Build: %v", err)
	}
	g := Build(h, nil)

	added := ApplySyntheticEdges(g, h, []SyntheticEdge{
		{FromSubtypesOf: "no.such.Interface", To: "fixture.FormalGreeter"},
	})
	if added != 0 {
		t.Errorf("added = %d, want 0 (no such interface has no known subtypes)", added)
	}
}
