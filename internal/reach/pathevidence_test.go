package reach

import (
	"encoding/json"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

// TestPathEvidence (ROADMAP B0/B1): a positive result's path is rebuilt
// step by step from the graph — every step has the edges between its two
// methods with kind and basis — and the finding's characters are the set
// of the steps'. v1 JSON does not carry them.
func TestPathEvidence(t *testing.T) {
	e := overloadEngine(t, "1.0")
	for _, spec := range []string{"lib.overload.C#parse(Ljava/io/InputStream;)V", "lib.overload.Base#handle(Ljava/lang/String;)V"} {
		r := analyzeOverload(t, e, "1.0", diffEntry(spec))
		if r.Status != StatusReachable || len(r.Path) < 2 {
			t.Fatalf("%s: %s %v", spec, r.Status, r.Path)
		}
		if len(r.PathEdges) != len(r.Path)-1 || len(r.Character) == 0 {
			t.Fatalf("%s: %d steps for %d nodes, character %v", spec, len(r.PathEdges), len(r.Path), r.Character)
		}
		for i, s := range r.PathEdges {
			if s.From != r.Path[i] || s.To != r.Path[i+1] || len(s.Evidence) == 0 || s.Character == "" {
				t.Errorf("%s step %d: %+v", spec, i, s)
			}
			for _, ev := range s.Evidence {
				if ev.Basis == "" || ev.Character != basisCharacter(callgraph.EdgeBasis(ev.Basis)) {
					t.Errorf("%s step %d evidence %+v", spec, i, ev)
				}
			}
		}
		for _, f := range r.Findings {
			if len(f.PathEdges) != len(f.Path)-1 || strings.Join(f.Character, ",") == "" {
				t.Errorf("%s finding %s: %d steps for %d nodes", spec, f.Instance, len(f.PathEdges), len(f.Path))
			}
		}
		b, _ := json.Marshal(r)
		if strings.Contains(string(b), "path_edges") || strings.Contains(string(b), "character") {
			t.Errorf("%s: v1 JSON carries path evidence: %s", spec, b)
		}
	}
}

// TestBasisCharacterAndLines: every basis maps to its character, and the
// report annotates only steps no bytecode call backs.
func TestBasisCharacterAndLines(t *testing.T) {
	want := map[callgraph.EdgeBasis]string{
		callgraph.BasisDirect: CharJVMRule, callgraph.BasisInit: CharJVMRule, callgraph.BasisCHA: CharHeuristic,
		callgraph.BasisSynthetic: CharDeclaration, callgraph.BasisSAM: CharLibraryModel, callgraph.BasisModel: CharLibraryModel,
		callgraph.BasisReflection: CharLibraryModel, callgraph.BasisServiceLoader: CharLibraryModel, callgraph.BasisCallback: CharLibraryModel,
		callgraph.BasisMethodHandle: CharLibraryModel, callgraph.BasisDeserialization: CharLibraryModel, "future": CharUnclassified,
	}
	for b, c := range want {
		if got := basisCharacter(b); got != c {
			t.Errorf("%s -> %s, want %s", b, got, c)
		}
	}
	r := Result{Path: []string{"a.A#m()V", "b.B#n()V", "c.C#o()V"}, PathEdges: []PathEdge{
		{Character: CharJVMRule, Evidence: []EdgeEvidence{{Basis: "direct"}}},
		{Character: CharHeuristic, Evidence: []EdgeEvidence{{Basis: "cha"}, {Basis: "cha"}}},
	}}
	lines := r.PathLines()
	if lines[1] != " → b.B#n()V" || lines[2] != " →[heuristic: cha] c.C#o()V" {
		t.Errorf("lines = %q", lines)
	}
	if s := r.PathSummary(); s != "2 steps: jvm_rule 1, heuristic 1" {
		t.Errorf("summary = %q", s)
	}
}
