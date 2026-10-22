package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestApplyCallbackEdges(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "m", Descriptor: "()V"}},
		}},
		{FQCN: "test.MyRunnable", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "run", Descriptor: "()V"}},
		}},
	})
	g := Build(h, nil)

	added := ApplyCallbackEdges(g, []CallbackEdge{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 11,
			TargetFQCN: "test.MyRunnable", TargetMethod: "run", TargetMethodDesc: "()V"},
	})
	if added != 1 {
		t.Fatalf("ApplyCallbackEdges added %d edges, want 1", added)
	}

	callerID, ok := g.ID("test.Caller", "m", "()V")
	if !ok {
		t.Fatal("test.Caller#m not found")
	}
	runID, ok := g.ID("test.MyRunnable", "run", "()V")
	if !ok {
		t.Fatal("test.MyRunnable#run not found")
	}
	edge, ok := g.EdgeBetween(callerID, runID)
	if !ok {
		t.Fatal("missing edge to test.MyRunnable#run")
	}
	if edge.Kind != classfile.CallSynthetic || edge.Basis != BasisCallback {
		t.Errorf("edge = %+v, want Kind=CallSynthetic Basis=BasisCallback", edge)
	}
	if len(edge.Sites) != 1 || edge.Sites[0].PC != 11 {
		t.Errorf("edge Sites = %+v, want [{PC:11}]", edge.Sites)
	}

	if addedAgain := ApplyCallbackEdges(g, []CallbackEdge{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 11,
			TargetFQCN: "test.MyRunnable", TargetMethod: "run", TargetMethodDesc: "()V"},
	}); addedAgain != 0 {
		t.Errorf("re-applying the same edge added %d, want 0", addedAgain)
	}
}
