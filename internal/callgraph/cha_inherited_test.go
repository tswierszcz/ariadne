package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// TestCHAInheritedBody: a call on an interface whose implementing class
// inherits the body from a superclass outside the interface's hierarchy
// (testdata/runtime: LoudGreeter extends BaseGreeter implements Greeter)
// gets a CHA edge to that body — the one the JVM ran in the recorded trace.
// A class declaring the method keeps its own edge; an abstract declaration
// is never a target.
func TestCHAInheritedBody(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/runtime/rtm.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	g := BuildWith(h, CHA{})
	from := mustID(t, g, "rtm.Gaps", "inherited", "(J)J")
	to := mustID(t, g, "rtm.BaseGreeter", "greet", "(J)J")
	e, ok := g.EdgeBetween(from, to)
	if !ok || e.Basis != BasisCHA {
		t.Fatalf("no CHA edge Gaps#inherited -> BaseGreeter#greet (got %+v, %v)", e, ok)
	}
	targets := CHA{}.DispatchTargets(h, classfile.Call{Owner: "rtm.Greeter", Name: "greet", Desc: "(J)J", Kind: classfile.CallInterface})
	if len(targets) != 1 || targets[0] != "rtm.BaseGreeter" {
		t.Errorf("Greeter#greet targets %v, want [rtm.BaseGreeter]", targets)
	}
	// Shape#area: both implementations declare it — unchanged.
	shapes := CHA{}.DispatchTargets(h, classfile.Call{Owner: "rtm.Shape", Name: "area", Desc: "(J)J", Kind: classfile.CallInterface})
	if len(shapes) != 2 {
		t.Errorf("Shape#area targets %v, want Circle and Square", shapes)
	}
}
