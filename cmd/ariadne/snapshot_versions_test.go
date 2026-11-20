package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
	"testing"

	"ariadne/internal/analysis"
	"ariadne/internal/callgraph"
)

// basisVersions says, for every call-graph edge basis, which snapshot
// derivation entry versions the rules that add such edges (ROADMAP B3
// b3-4): an algorithm key, or "config:models" for the fixed-point
// participants recorded with their versions in the snapshot configuration.
// A change to those rules without a version bump would leave a snapshot
// comparison unable to say why a verdict moved.
var basisVersions = map[callgraph.EdgeBasis][]string{
	callgraph.BasisDirect:          {"callgraph.builder"},
	callgraph.BasisCHA:             {"callgraph.builder"},
	callgraph.BasisInit:            {"callgraph.builder"},
	callgraph.BasisSynthetic:       {"callgraph.builder"},
	callgraph.BasisServiceLoader:   {"config:models"},
	callgraph.BasisModel:           {"config:models"},
	callgraph.BasisCallback:        {"callback.resolver", "config:models"},
	callgraph.BasisMethodHandle:    {"methodhandle.edges"},
	callgraph.BasisReflection:      {"reflection.invoke", "reflection.parameters"},
	callgraph.BasisDeserialization: {"deserialize.models"},
	callgraph.BasisSAM:             {"samcall.rules", "config:models"},
}

// TestAlgorithmVersions_everyEdgeBasis reads every EdgeBasis constant
// declared in internal/callgraph, so a new kind of edge fails here until
// basisVersions names the version of its rules.
func TestAlgorithmVersions_everyEdgeBasis(t *testing.T) {
	declared := map[string]bool{}
	pkgs, err := parser.ParseDir(token.NewFileSet(), "../../internal/callgraph", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pkgs {
		for _, f := range p.Files {
			for _, d := range f.Decls {
				g, ok := d.(*ast.GenDecl)
				if !ok || g.Tok != token.CONST {
					continue
				}
				for _, s := range g.Specs {
					vs := s.(*ast.ValueSpec)
					if id, ok := vs.Type.(*ast.Ident); ok && id.Name == "EdgeBasis" {
						for _, n := range vs.Names {
							declared[n.Name] = true
						}
					}
				}
			}
		}
	}
	if len(declared) != len(basisVersions) {
		t.Errorf("internal/callgraph declares %d edge bases, basisVersions covers %d: %v", len(declared), len(basisVersions), declared)
	}
	algos := algorithmVersions(callgraph.ModeCHA)
	for basis, keys := range basisVersions {
		for _, k := range keys {
			if k == "config:models" {
				continue
			}
			if algos[k] == "" {
				t.Errorf("basis %s: algorithm %q has no version", basis, k)
			}
		}
	}
}

// TestParticipantCatalog_versioned: every fixed-point participant has a
// version, and the snapshot's "models" configuration carries it.
func TestParticipantCatalog_versioned(t *testing.T) {
	for _, p := range participantCatalog() {
		if p.Version == "" {
			t.Errorf("participant %s has no version", p.ID)
		}
	}
	for _, l := range analysis.Labels(participantCatalog()) {
		if !strings.Contains(l, "@") {
			t.Errorf("label %q carries no version", l)
		}
	}
}
