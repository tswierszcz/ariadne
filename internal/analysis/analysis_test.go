package analysis

import (
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/models"
)

type evaluatorFunc func(*models.Model, models.CallSite) models.Fact

func (f evaluatorFunc) Evaluate(m *models.Model, s models.CallSite) models.Fact { return f(m, s) }

func TestRun_repeatsUntilModeledTargetExposesAnotherModelSite(t *testing.T) {
	app := &hierarchy.ClassInfo{FQCN: "app.Entry", IsApp: true, Methods: []hierarchy.Method{{Method: classfile.Method{Name: "run", Descriptor: "()V"}, Calls: []classfile.Call{{Kind: classfile.CallStatic, Owner: "lib.API", Name: "first", Desc: "()V", PC: 3}}}}}
	api := &hierarchy.ClassInfo{FQCN: "lib.API", GAV: &input.GAV{Group: "lib", Artifact: "api", Version: "1.0"}, Methods: []hierarchy.Method{
		{Method: classfile.Method{Name: "first", Descriptor: "()V"}},
		{Method: classfile.Method{Name: "second", Descriptor: "()V"}},
	}}
	one := &hierarchy.ClassInfo{FQCN: "app.One", Methods: []hierarchy.Method{{Method: classfile.Method{Name: "execute", Descriptor: "()V"}, Calls: []classfile.Call{{Kind: classfile.CallStatic, Owner: "lib.API", Name: "second", Desc: "()V", PC: 7}}}}}
	two := &hierarchy.ClassInfo{FQCN: "app.Two", Methods: []hierarchy.Method{{Method: classfile.Method{Name: "execute", Descriptor: "()V"}}}}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{app, api, one, two})
	g := callgraph.Build(h, nil)
	registry := models.NewRegistry()
	registry.Register(&models.Model{ID: "m:first", Artifact: "lib:api", Signatures: []models.Signature{{Owner: "lib.API", Name: "first", Desc: "()V"}}})
	registry.Register(&models.Model{ID: "m:second", Artifact: "lib:api", Signatures: []models.Signature{{Owner: "lib.API", Name: "second", Desc: "()V"}}})
	calls := 0
	eval := evaluatorFunc(func(m *models.Model, s models.CallSite) models.Fact {
		calls++
		if m.ID == "m:first" {
			return models.Fact{Targets: []models.Target{{Method: models.Signature{Owner: "app.One", Name: "execute", Desc: "()V"}}}}
		}
		return models.Fact{Targets: []models.Target{{Method: models.Signature{Owner: "app.Two", Name: "execute", Desc: "()V"}}}}
	})
	root, _ := g.ID("app.Entry", "run", "()V")
	result, err := Run(h, g, registry, eval, []uint32{root}, Options{MaxIterations: 4})
	if err != nil {
		t.Fatal(err)
	}
	if result.Incomplete || result.Iterations != 3 || result.AddedEdges != 2 || calls != 2 {
		t.Fatalf("result=%+v calls=%d, want complete fixed point after two additions", result, calls)
	}
	if g.ModelEdgeCount() != 2 {
		t.Fatalf("Graph.ModelEdgeCount() = %d, want 2 distinct model edges", g.ModelEdgeCount())
	}
	if _, ok := g.EdgeBetween(root, mustID(t, g, "app.One", "execute", "()V")); !ok {
		t.Fatal("first model edge missing")
	}
	firstID := mustID(t, g, "app.One", "execute", "()V")
	if _, ok := g.EdgeBetween(firstID, mustID(t, g, "app.Two", "execute", "()V")); !ok {
		t.Fatal("second iteration model edge missing")
	}
}

func TestRun_reportsIterationBudgetAsIncomplete(t *testing.T) {
	h, g, registry, eval, root := limitFixture(t)
	result, err := Run(h, g, registry, eval, []uint32{root}, Options{MaxIterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Incomplete || len(result.Reasons) == 0 || result.AddedEdges != 1 {
		t.Fatalf("result=%+v, want visible incomplete limit after edge discovery", result)
	}
}

func TestRun_reportsUnmodeledTargetAsIncomplete(t *testing.T) {
	h, g, registry, _, root := limitFixture(t)
	eval := evaluatorFunc(func(*models.Model, models.CallSite) models.Fact {
		return models.Fact{Targets: []models.Target{{Method: models.Signature{Owner: "missing.Type", Name: "go", Desc: "()V"}}}}
	})
	result, err := Run(h, g, registry, eval, []uint32{root}, Options{MaxIterations: 2})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Incomplete || len(result.Reasons) == 0 || result.AddedEdges != 0 {
		t.Fatalf("result=%+v, want explicit unresolved-target gap", result)
	}
}

func TestRun_ignoresUnreachableAndWrongVersionSites(t *testing.T) {
	call := classfile.Call{Kind: classfile.CallStatic, Owner: "lib.API", Name: "first", Desc: "()V"}
	app := &hierarchy.ClassInfo{FQCN: "app.Entry", IsApp: true, Methods: []hierarchy.Method{{Method: classfile.Method{Name: "run", Descriptor: "()V"}}}}
	api := &hierarchy.ClassInfo{FQCN: "lib.API", GAV: &input.GAV{Group: "lib", Artifact: "api", Version: "1.0"}, Methods: []hierarchy.Method{{Method: classfile.Method{Name: "first", Descriptor: "()V"}, Calls: []classfile.Call{call}}}}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{app, api})
	g := callgraph.Build(h, nil)
	r := models.NewRegistry()
	r.Register(&models.Model{ID: "m:wrong-version", Artifact: "lib:api", Versions: models.VersionRange{Min: "2.0"}, Signatures: []models.Signature{{Owner: "lib.API", Name: "first", Desc: "()V"}}})
	calls := 0
	root, _ := g.ID("app.Entry", "run", "()V")
	result, err := Run(h, g, r, evaluatorFunc(func(*models.Model, models.CallSite) models.Fact { calls++; return models.Fact{} }), []uint32{root}, Options{MaxIterations: 2})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 || len(result.Facts) != 0 {
		t.Fatalf("calls=%d facts=%v, want no evaluation", calls, result.Facts)
	}
}

func limitFixture(t *testing.T) (*hierarchy.Hierarchy, *callgraph.Graph, *models.Registry, models.Evaluator, uint32) {
	t.Helper()
	app := &hierarchy.ClassInfo{FQCN: "app.Entry", IsApp: true, Methods: []hierarchy.Method{{Method: classfile.Method{Name: "run", Descriptor: "()V"}, Calls: []classfile.Call{{Kind: classfile.CallStatic, Owner: "lib.API", Name: "first", Desc: "()V"}}}}}
	api := &hierarchy.ClassInfo{FQCN: "lib.API", GAV: &input.GAV{Group: "lib", Artifact: "api", Version: "1.0"}, Methods: []hierarchy.Method{{Method: classfile.Method{Name: "first", Descriptor: "()V"}}}}
	target := &hierarchy.ClassInfo{FQCN: "app.Target", Methods: []hierarchy.Method{{Method: classfile.Method{Name: "go", Descriptor: "()V"}}}}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{app, api, target})
	g := callgraph.Build(h, nil)
	r := models.NewRegistry()
	r.Register(&models.Model{ID: "m:first", Signatures: []models.Signature{{Owner: "lib.API", Name: "first", Desc: "()V"}}})
	root, _ := g.ID("app.Entry", "run", "()V")
	e := evaluatorFunc(func(*models.Model, models.CallSite) models.Fact {
		return models.Fact{Targets: []models.Target{{Method: models.Signature{Owner: "app.Target", Name: "go", Desc: "()V"}}}}
	})
	return h, g, r, e, root
}

func mustID(t *testing.T, g *callgraph.Graph, owner, name, desc string) uint32 {
	t.Helper()
	id, ok := g.ID(owner, name, desc)
	if !ok {
		t.Fatalf("missing graph node %s#%s%s", owner, name, desc)
	}
	return id
}
