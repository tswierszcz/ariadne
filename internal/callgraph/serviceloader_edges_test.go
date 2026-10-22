package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestApplyServiceLoaderEdges(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "m", Descriptor: "()V"}},
		}},
		{FQCN: "test.GoodProvider", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}},
			{Method: classfile.Method{Name: "<clinit>", Descriptor: "()V"}},
		}},
		{FQCN: "test.NoClinitProvider", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}},
		}},
	})
	g := Build(h, nil)

	added := ApplyServiceLoaderEdges(g, []ServiceLoaderEdge{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 7, ProviderFQCN: "test.GoodProvider"},
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 7, ProviderFQCN: "test.NoClinitProvider"},
	})
	// GoodProvider: <init> + <clinit> = 2; NoClinitProvider: <init> only = 1.
	if added != 3 {
		t.Fatalf("ApplyServiceLoaderEdges added %d edges, want 3", added)
	}

	callerID, ok := g.ID("test.Caller", "m", "()V")
	if !ok {
		t.Fatal("test.Caller#m not found")
	}
	ctorID, ok := g.ID("test.GoodProvider", "<init>", "()V")
	if !ok {
		t.Fatal("test.GoodProvider#<init> not found")
	}
	clinitID, ok := g.ID("test.GoodProvider", "<clinit>", "()V")
	if !ok {
		t.Fatal("test.GoodProvider#<clinit> not found")
	}

	ctorEdge, ok := g.EdgeBetween(callerID, ctorID)
	if !ok {
		t.Fatal("missing edge to test.GoodProvider#<init>")
	}
	if ctorEdge.Kind != classfile.CallSynthetic || ctorEdge.Basis != BasisServiceLoader {
		t.Errorf("ctor edge = %+v, want Kind=CallSynthetic Basis=BasisServiceLoader", ctorEdge)
	}
	if len(ctorEdge.Sites) != 1 || ctorEdge.Sites[0].PC != 7 {
		t.Errorf("ctor edge Sites = %+v, want [{PC:7}] (the real ServiceLoader.load call site)", ctorEdge.Sites)
	}

	clinitEdge, ok := g.EdgeBetween(callerID, clinitID)
	if !ok {
		t.Fatal("missing edge to test.GoodProvider#<clinit>")
	}
	if clinitEdge.Kind != classfile.CallInit || clinitEdge.Basis != BasisServiceLoader {
		t.Errorf("clinit edge = %+v, want Kind=CallInit Basis=BasisServiceLoader", clinitEdge)
	}

	noClinitCtorID, ok := g.ID("test.NoClinitProvider", "<init>", "()V")
	if !ok {
		t.Fatal("test.NoClinitProvider#<init> not found")
	}
	if _, ok := g.EdgeBetween(callerID, noClinitCtorID); !ok {
		t.Error("missing edge to test.NoClinitProvider#<init>")
	}

	// Applying the same edges again must not double-count.
	addedAgain := ApplyServiceLoaderEdges(g, []ServiceLoaderEdge{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 7, ProviderFQCN: "test.GoodProvider"},
	})
	if addedAgain != 0 {
		t.Errorf("re-applying the same edge added %d, want 0", addedAgain)
	}
}
