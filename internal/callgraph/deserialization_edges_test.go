package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestApplyDeserializationEdges(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Reader", Methods: []hierarchy.Method{{Method: classfile.Method{Name: "read", Descriptor: "()V"}}}},
		{FQCN: "test.Gadget", Methods: []hierarchy.Method{{Method: classfile.Method{Name: "readObject", Descriptor: "(Ljava/io/ObjectInputStream;)V"}}}},
	})
	g := Build(h, nil)
	e := DeserializationEdge{CallerFQCN: "test.Reader", CallerName: "read", CallerDesc: "()V", PC: 7,
		TargetFQCN: "test.Gadget", TargetName: "readObject", TargetDesc: "(Ljava/io/ObjectInputStream;)V"}
	missing := e
	missing.TargetFQCN = "test.Absent"
	if added := ApplyDeserializationEdges(g, []DeserializationEdge{e, e, missing}); added != 1 {
		t.Fatalf("added %d, want 1 (duplicate and missing target skipped)", added)
	}
	from, _ := g.ID("test.Reader", "read", "()V")
	to, _ := g.ID("test.Gadget", "readObject", "(Ljava/io/ObjectInputStream;)V")
	edge, ok := g.EdgeBetween(from, to)
	if !ok || edge.Kind != classfile.CallSynthetic || edge.Basis != BasisDeserialization || len(edge.Sites) != 1 || edge.Sites[0].PC != 7 {
		t.Errorf("edge = %+v %v", edge, ok)
	}
}
