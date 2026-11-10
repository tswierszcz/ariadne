package analysis

import (
	"fmt"
	"sort"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/models"
)

// gateStep adapts a creator gate (running totals) to a Step (per call).
func gateStep(name string, gate *callgraph.SAMGate) Step {
	prev := 0
	return Step{Name: name, Apply: func(g *callgraph.Graph, roots []uint32) int {
		total := gate.Apply(g, roots).Added
		n := total - prev
		prev = total
		return n
	}}
}

// stepFixture: app.Entry#run calls lib.API#first (model m:first → the
// factory app.Factory#make, which creates app.Gated) and app.Iface#go
// (a gate candidate to app.Gated#go, created by the factory). app.Gated#go
// calls lib.API#second (model m:second → app.Last#done). The last edge
// needs the order model → gate → model: the gates-models-gates sequence
// the CLI used before misses it.
func stepFixture(t *testing.T) (*hierarchy.Hierarchy, *callgraph.Graph, *models.Registry, models.Evaluator, uint32, []callgraph.SAMEdge) {
	t.Helper()
	sv := func(name string, calls ...classfile.Call) hierarchy.Method {
		return hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: "()V"}, Calls: calls}
	}
	call := func(kind classfile.CallKind, owner, name string, pc int) classfile.Call {
		return classfile.Call{Kind: kind, Owner: owner, Name: name, Desc: "()V", PC: pc}
	}
	gav := &input.GAV{Group: "lib", Artifact: "api", Version: "1.0"}
	classes := []*hierarchy.ClassInfo{
		{FQCN: "app.Entry", IsApp: true, Methods: []hierarchy.Method{sv("run", call(classfile.CallStatic, "lib.API", "first", 3), call(classfile.CallInterface, "app.Iface", "go", 9))}},
		{FQCN: "lib.API", GAV: gav, Methods: []hierarchy.Method{sv("first"), sv("second")}},
		{FQCN: "app.Iface", AccessFlags: classfile.AccInterface | classfile.AccAbstract, Methods: []hierarchy.Method{{Method: classfile.Method{Name: "go", Descriptor: "()V", AccessFlags: classfile.AccAbstract}}}},
		{FQCN: "app.Factory", Methods: []hierarchy.Method{sv("make")}},
		{FQCN: "app.Gated", Methods: []hierarchy.Method{sv("go", call(classfile.CallStatic, "lib.API", "second", 5))}},
		{FQCN: "app.Last", Methods: []hierarchy.Method{sv("done")}},
		{FQCN: "app.Other", Methods: []hierarchy.Method{sv("go")}},
	}
	h := hierarchy.NewForTesting(classes)
	g := callgraph.Build(h, nil)
	registry := models.NewRegistry()
	registry.Register(&models.Model{ID: "m:first", Artifact: "lib:api", Signatures: []models.Signature{{Owner: "lib.API", Name: "first", Desc: "()V"}}})
	registry.Register(&models.Model{ID: "m:second", Artifact: "lib:api", Signatures: []models.Signature{{Owner: "lib.API", Name: "second", Desc: "()V"}}})
	eval := evaluatorFunc(func(m *models.Model, s models.CallSite) models.Fact {
		if m.ID == "m:first" {
			return models.Fact{Targets: []models.Target{{Method: models.Signature{Owner: "app.Factory", Name: "make", Desc: "()V"}}}}
		}
		return models.Fact{Targets: []models.Target{{Method: models.Signature{Owner: "app.Last", Name: "done", Desc: "()V"}}}}
	})
	root, _ := g.ID("app.Entry", "run", "()V")
	cands := []callgraph.SAMEdge{
		{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 9, Kind: classfile.CallInterface,
			TargetFQCN: "app.Gated", TargetName: "go", TargetDesc: "()V", Creators: []callgraph.MethodRef{{FQCN: "app.Factory", Name: "make", Desc: "()V"}}},
		// Created by app.Last#done: only after the second model edge.
		{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 9, Kind: classfile.CallInterface,
			TargetFQCN: "app.Other", TargetName: "go", TargetDesc: "()V", Creators: []callgraph.MethodRef{{FQCN: "app.Last", Name: "done", Desc: "()V"}}},
	}
	return h, g, registry, eval, root, cands
}

func edgeSet(g *callgraph.Graph) []string {
	var out []string
	for id := 0; id < g.NodeCount(); id++ {
		for _, e := range g.Edges(uint32(id)) {
			out = append(out, fmt.Sprintf("%s>%s/%v", g.Key(uint32(id)), g.Key(e.To), e.Basis))
		}
	}
	sort.Strings(out)
	return out
}

func TestRun_stepsAndModelsShareOneFixedPoint(t *testing.T) {
	h, g, registry, eval, root, cands := stepFixture(t)
	gate := callgraph.NewCreatorGate(g, h, cands, callgraph.BasisModel)
	result, err := Run(h, g, registry, eval, []uint32{root}, Options{MaxIterations: 8, Steps: []Step{gateStep("gate", gate)}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Incomplete || result.AddedEdges != 2 || result.StepEdges["gate"] != 2 || len(result.LastAdders) != 0 {
		t.Fatalf("result=%+v, want both model edges and both gate edges at a fixed point", result)
	}
	gated := mustID(t, g, "app.Gated", "go", "()V")
	if _, ok := g.EdgeBetween(gated, mustID(t, g, "app.Last", "done", "()V")); !ok {
		t.Fatal("model edge behind the gate missing")
	}
	if _, ok := g.EdgeBetween(root, mustID(t, g, "app.Other", "go", "()V")); !ok {
		t.Fatal("gate edge behind the second model edge missing")
	}

	// The sequence the CLI ran before (gates, models to their own fixed
	// point, gates once more) misses the second model edge.
	h2, g2, registry2, eval2, root2, cands2 := stepFixture(t)
	gate2 := callgraph.NewCreatorGate(g2, h2, cands2, callgraph.BasisModel)
	gate2.Apply(g2, []uint32{root2})
	if _, err := Run(h2, g2, registry2, eval2, []uint32{root2}, Options{MaxIterations: 8}); err != nil {
		t.Fatal(err)
	}
	gate2.Apply(g2, []uint32{root2})
	if _, ok := g2.EdgeBetween(mustID(t, g2, "app.Gated", "go", "()V"), mustID(t, g2, "app.Last", "done", "()V")); ok {
		t.Fatal("fixture does not exercise the joint fixed point")
	}
}

func TestRun_stepBudgetIsVisible(t *testing.T) {
	h, g, registry, eval, root, cands := stepFixture(t)
	gate := callgraph.NewCreatorGate(g, h, cands, callgraph.BasisModel)
	result, err := Run(h, g, registry, eval, []uint32{root}, Options{MaxIterations: 2, Steps: []Step{gateStep("gate", gate)}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Incomplete || len(result.LastAdders) == 0 || len(result.Reasons) == 0 {
		t.Fatalf("result=%+v, want the limit reported with the participants still adding", result)
	}
}

// TestRun_orderDoesNotChangeTheGraph: the same participants in every
// order give the same edges (ROADMAP B3: "kolejność uruchomienia modeli
// nie może losowo zmieniać wyniku").
func TestRun_orderDoesNotChangeTheGraph(t *testing.T) {
	var want []string
	for _, order := range [][2]int{{0, 1}, {1, 0}} {
		h, g, registry, eval, root, cands := stepFixture(t)
		steps := []Step{
			gateStep("gate-a", callgraph.NewCreatorGate(g, h, cands[:1], callgraph.BasisModel)),
			gateStep("gate-b", callgraph.NewCreatorGate(g, h, cands[1:], callgraph.BasisModel)),
		}
		if _, err := Run(h, g, registry, eval, []uint32{root}, Options{MaxIterations: 8, Steps: []Step{steps[order[0]], steps[order[1]]}}); err != nil {
			t.Fatal(err)
		}
		got := edgeSet(g)
		if want == nil {
			want = got
			continue
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("order %v: edges\n%v\nwant\n%v", order, got, want)
		}
	}
}
