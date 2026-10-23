package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestApplyReflectNewInstanceEdges(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "m", Descriptor: "()V"}},
		}},
		{FQCN: "test.GoodTarget", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "<init>", Descriptor: "(Ljava/lang/String;)V"}},
			{Method: classfile.Method{Name: "<clinit>", Descriptor: "()V"}},
		}},
		{FQCN: "test.NoClinitTarget", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}},
		}},
	})
	g := Build(h, nil)

	added := ApplyReflectNewInstanceEdges(g, []ReflectNewInstanceEdge{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 11, TargetFQCN: "test.GoodTarget", TargetDesc: "(Ljava/lang/String;)V"},
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 11, TargetFQCN: "test.NoClinitTarget", TargetDesc: "()V"},
	})
	// GoodTarget: <init> + <clinit> = 2; NoClinitTarget: <init> only = 1.
	if added != 3 {
		t.Fatalf("ApplyReflectNewInstanceEdges added %d edges, want 3", added)
	}

	callerID, ok := g.ID("test.Caller", "m", "()V")
	if !ok {
		t.Fatal("test.Caller#m not found")
	}
	ctorID, ok := g.ID("test.GoodTarget", "<init>", "(Ljava/lang/String;)V")
	if !ok {
		t.Fatal("test.GoodTarget#<init> not found")
	}
	clinitID, ok := g.ID("test.GoodTarget", "<clinit>", "()V")
	if !ok {
		t.Fatal("test.GoodTarget#<clinit> not found")
	}

	ctorEdge, ok := g.EdgeBetween(callerID, ctorID)
	if !ok {
		t.Fatal("missing edge to test.GoodTarget#<init>")
	}
	if ctorEdge.Kind != classfile.CallSynthetic || ctorEdge.Basis != BasisReflection {
		t.Errorf("ctor edge = %+v, want Kind=CallSynthetic Basis=BasisReflection", ctorEdge)
	}
	if len(ctorEdge.Sites) != 1 || ctorEdge.Sites[0].PC != 11 {
		t.Errorf("ctor edge Sites = %+v, want [{PC:11}] (the real newInstance call site)", ctorEdge.Sites)
	}

	clinitEdge, ok := g.EdgeBetween(callerID, clinitID)
	if !ok {
		t.Fatal("missing edge to test.GoodTarget#<clinit>")
	}
	if clinitEdge.Kind != classfile.CallInit || clinitEdge.Basis != BasisReflection {
		t.Errorf("clinit edge = %+v, want Kind=CallInit Basis=BasisReflection", clinitEdge)
	}

	noClinitCtorID, ok := g.ID("test.NoClinitTarget", "<init>", "()V")
	if !ok {
		t.Fatal("test.NoClinitTarget#<init> not found")
	}
	if _, ok := g.EdgeBetween(callerID, noClinitCtorID); !ok {
		t.Error("missing edge to test.NoClinitTarget#<init>")
	}

	// Applying the same edges again must not double-count.
	addedAgain := ApplyReflectNewInstanceEdges(g, []ReflectNewInstanceEdge{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 11, TargetFQCN: "test.GoodTarget", TargetDesc: "(Ljava/lang/String;)V"},
	})
	if addedAgain != 0 {
		t.Errorf("re-applying the same edge added %d, want 0", addedAgain)
	}
}
