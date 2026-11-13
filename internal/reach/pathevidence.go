package reach

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
)

// Characters of a path step and of a finding (ROADMAP B0: "wynik reguły
// JVM, model biblioteki, heurystyka, deklaracja wejściowa"). They are kinds
// of evidence, not a confidence score.
const (
	// CharJVMRule: a call instruction in the bytecode, or a class
	// initialization the JVM performs.
	CharJVMRule = "jvm_rule"
	// CharLibraryModel: a modeled API (ServiceLoader, callbacks, method
	// handles, reflection, deserialization, SAM calls, coordinator models).
	CharLibraryModel = "library_model"
	// CharHeuristic: a CHA expansion — every subtype's override is assumed
	// possible, whether or not such an object exists at that call.
	CharHeuristic = "heuristic"
	// CharDeclaration: an edge a profile declares (synthetic_edges).
	CharDeclaration = "input_declaration"
	// CharUnclassified: an edge whose basis this version does not know.
	CharUnclassified = "unclassified"
)

var charOrder = []string{CharJVMRule, CharLibraryModel, CharDeclaration, CharHeuristic, CharUnclassified}

// basisCharacter maps an edge basis to its character.
func basisCharacter(b callgraph.EdgeBasis) string {
	switch b {
	case callgraph.BasisDirect, callgraph.BasisInit:
		return CharJVMRule
	case callgraph.BasisCHA:
		return CharHeuristic
	case callgraph.BasisSynthetic:
		return CharDeclaration
	case callgraph.BasisServiceLoader, callgraph.BasisCallback, callgraph.BasisMethodHandle, callgraph.BasisReflection,
		callgraph.BasisDeserialization, callgraph.BasisSAM, callgraph.BasisModel:
		return CharLibraryModel
	}
	return CharUnclassified
}

// EdgeEvidence is one rule that produced a path step's edge.
type EdgeEvidence struct {
	Kind      string `json:"kind"`
	Basis     string `json:"basis"`
	Character string `json:"character"`
	// PCs are the bytecode offsets of the call sites (none for an edge
	// without a call site, e.g. a declared one).
	PCs []int `json:"pcs,omitempty"`
}

// PathEdge is one step of a reported path with every edge between its
// two methods (ROADMAP B1: two reasons for the same edge are both kept).
// Character is the step's strongest evidence: a step a bytecode call
// also backs is not a heuristic one.
type PathEdge struct {
	From      string         `json:"from"`
	To        string         `json:"to"`
	Character string         `json:"character"`
	Evidence  []EdgeEvidence `json:"evidence"`
}

// pathEvidence rebuilds a rendered path's steps from the graph and
// returns them with the set of the steps' characters, in charOrder.
func pathEvidence(g *callgraph.Graph, path []string) ([]PathEdge, []string) {
	if len(path) < 2 {
		return nil, nil
	}
	rank := map[string]int{}
	for i, c := range charOrder {
		rank[c] = i
	}
	seen := map[string]bool{}
	var steps []PathEdge
	for i := 0; i+1 < len(path); i++ {
		step := PathEdge{From: path[i], To: path[i+1], Character: CharUnclassified}
		from, ok1 := g.IDOfKey(path[i])
		to, ok2 := g.IDOfKey(path[i+1])
		if ok1 && ok2 {
			for _, e := range g.EdgesBetween(from, to) {
				ev := EdgeEvidence{Kind: string(e.Kind), Basis: string(e.Basis), Character: basisCharacter(e.Basis)}
				for _, s := range e.Sites {
					ev.PCs = append(ev.PCs, s.PC)
				}
				sort.Ints(ev.PCs)
				step.Evidence = append(step.Evidence, ev)
				if rank[ev.Character] < rank[step.Character] {
					step.Character = ev.Character
				}
			}
		}
		seen[step.Character] = true
		steps = append(steps, step)
	}
	var chars []string
	for _, c := range charOrder {
		if seen[c] {
			chars = append(chars, c)
		}
	}
	return steps, chars
}

// PathLines renders Path for the report: one node per element, each
// arrow after the first node annotated with its step's character and
// bases when the step is not backed by a bytecode call (ROADMAP B0).
func (r Result) PathLines() []string {
	out := make([]string, len(r.Path))
	for i, node := range r.Path {
		if i == 0 {
			out[i] = node
			continue
		}
		arrow := " → "
		if i-1 < len(r.PathEdges) {
			if s := r.PathEdges[i-1]; s.Character != CharJVMRule {
				var bases []string
				for _, ev := range s.Evidence {
					bases = append(bases, ev.Basis)
				}
				arrow = " →[" + s.Character
				if len(bases) > 0 {
					arrow += ": " + joinUnique(bases)
				}
				arrow += "] "
			}
		}
		out[i] = arrow + node
	}
	return out
}

// PathSummary is the report's one-line account of the path's evidence.
func (r Result) PathSummary() string {
	if len(r.PathEdges) == 0 {
		return ""
	}
	n := map[string]int{}
	for _, s := range r.PathEdges {
		n[s.Character]++
	}
	var parts []string
	for _, c := range charOrder {
		if n[c] > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", c, n[c]))
		}
	}
	return fmt.Sprintf("%d steps: %s", len(r.PathEdges), strings.Join(parts, ", "))
}

func joinUnique(xs []string) string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return strings.Join(out, ", ")
}
