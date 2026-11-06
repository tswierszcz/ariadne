package callgraph

import "ariadne/internal/classfile"

// DeserializationEdge is one DESERIALIZATION Z1 edge to add: an entry
// call site and a method serialization may run there. Plain identity,
// like CallbackEdge, so this package does not import internal/deserialize.
type DeserializationEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetName, TargetDesc string
}

// ApplyDeserializationEdges adds a synthetic BasisDeserialization edge
// per distinct (caller, target); a node missing from g is skipped.
// Returns how many edges were added.
func ApplyDeserializationEdges(g *Graph, edges []DeserializationEdge) int {
	added := 0
	for _, e := range edges {
		fromID, ok := g.ID(e.CallerFQCN, e.CallerName, e.CallerDesc)
		if !ok {
			continue
		}
		toID, ok := g.ID(e.TargetFQCN, e.TargetName, e.TargetDesc)
		if !ok {
			continue
		}
		if hasEdge(g, fromID, toID, classfile.CallSynthetic, BasisDeserialization) {
			continue
		}
		g.addEdge(fromID, toID, classfile.CallSynthetic, BasisDeserialization, &CallSite{PC: e.PC})
		added++
	}
	return added
}
