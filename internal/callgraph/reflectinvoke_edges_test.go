package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestApplyReflectInvokeEdges(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "m", Descriptor: "()V"}},
		}},
		{FQCN: "test.Target", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "bar", Descriptor: "()V"}},
			{Method: classfile.Method{Name: "bar", Descriptor: "(I)V"}}, // overload: both should get an edge
		}},
	})
	g := Build(h, nil)

	added := ApplyReflectInvokeEdges(g, []ReflectInvokeEdge{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9, TargetFQCN: "test.Target", TargetMethod: "bar"},
	})
	if added != 2 {
		t.Fatalf("ApplyReflectInvokeEdges added %d edges, want 2 (both overloads)", added)
	}

	callerID, ok := g.ID("test.Caller", "m", "()V")
	if !ok {
		t.Fatal("test.Caller#m not found")
	}
	for _, desc := range []string{"()V", "(I)V"} {
		targetID, ok := g.ID("test.Target", "bar", desc)
		if !ok {
			t.Fatalf("test.Target#bar%s not found", desc)
		}
		edge, ok := g.EdgeBetween(callerID, targetID)
		if !ok {
			t.Fatalf("missing edge to test.Target#bar%s", desc)
		}
		if edge.Kind != classfile.CallSynthetic || edge.Basis != BasisReflection {
			t.Errorf("edge to bar%s = %+v, want Kind=CallSynthetic Basis=BasisReflection", desc, edge)
		}
	}

	if addedAgain := ApplyReflectInvokeEdges(g, []ReflectInvokeEdge{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9, TargetFQCN: "test.Target", TargetMethod: "bar"},
	}); addedAgain != 0 {
		t.Errorf("re-applying the same edge added %d, want 0", addedAgain)
	}
}
