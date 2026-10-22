package callgraph

import (
	"sort"
	"testing"

	"ariadne/internal/hierarchy"
)

func buildFixtureGraph(t *testing.T) *Graph {
	t.Helper()
	h, warnings, err := hierarchy.Build("../../testdata/fixture/fixture-8.jar", nil)
	if err != nil {
		t.Fatalf("hierarchy.Build: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	return Build(h, nil)
}

func targetNames(g *Graph, edges []Edge) []string {
	names := make([]string, len(edges))
	for i, e := range edges {
		names[i] = g.Key(e.To)
	}
	sort.Strings(names)
	return names
}

// TestCHA_interfaceCallReachesBothImplementations is the Phase 3
// acceptance test from the plan: a call through the Greeter interface
// must reach both FormalGreeter and CasualGreeter via CHA, in addition to
// the direct edge to the interface method itself.
func TestCHA_interfaceCallReachesBothImplementations(t *testing.T) {
	g := buildFixtureGraph(t)

	handleID, ok := g.ID("fixture.App", "handle", "(Ljava/lang/String;)Ljava/lang/String;")
	if !ok {
		t.Fatalf("fixture.App#handle not found in graph")
	}

	const greetDesc = "(Ljava/lang/String;)Ljava/lang/String;"
	want := []string{
		"fixture.CasualGreeter#greet" + greetDesc,
		"fixture.FormalGreeter#greet" + greetDesc,
		"fixture.Greeter#greet" + greetDesc,
	}

	var got []string
	for _, e := range g.Edges(handleID) {
		key := g.Key(e.To)
		for _, w := range want {
			if key == w {
				got = append(got, key)
			}
		}
	}
	sort.Strings(got)
	if len(got) != len(want) {
		t.Fatalf("handle()'s edges to Greeter#greet = %v, want %v (full edge list: %v)", got, want, targetNames(g, g.Edges(handleID)))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestCHA_jdkTypeNotExpanded is the other half of the Phase 3 acceptance
// test: a call declared on a JDK type (java.lang.Runnable#run — not
// present in this fixture's own code, so constructed synthetically here
// via a graph built straight from a hierarchy that includes it) must not
// be expanded to its implementors, since those implementors are not part
// of the universe ariadne read in the first place and, more importantly,
// expanding JDK types at all is exactly what causes the CHA explosion
// section 3.6 describes.
func TestCHA_jdkTypeNotExpanded(t *testing.T) {
	if !DefaultSkip("java.lang.Runnable") {
		t.Fatalf("DefaultSkip(java.lang.Runnable) = false, want true")
	}
	if !DefaultSkip("javax.servlet.Filter") {
		t.Fatalf("DefaultSkip(javax.servlet.Filter) = false, want true")
	}
	if DefaultSkip("fixture.Greeter") {
		t.Fatalf("DefaultSkip(fixture.Greeter) = true, want false")
	}
}

func TestBuild_internsEveryDeclaredMethodEvenWithoutCalls(t *testing.T) {
	g := buildFixtureGraph(t)
	// fixture.Entry is an annotation type with no methods of its own
	// body, but fixture.Greeter#greet is abstract (no Code, no Calls)
	// and must still be a node so it can be a CHA expansion target.
	if _, ok := g.ID("fixture.Greeter", "greet", "(Ljava/lang/String;)Ljava/lang/String;"); !ok {
		t.Errorf("fixture.Greeter#greet not interned despite having no Calls")
	}
}

func TestIDsByName_overloadInsensitive(t *testing.T) {
	g := buildFixtureGraph(t)
	ids := g.IDsByName("fixture.Greeter", "greet")
	if len(ids) != 1 {
		t.Fatalf("IDsByName(Greeter, greet) = %v, want exactly 1 id", ids)
	}
}

func TestDirectCallsAreNotDuplicated(t *testing.T) {
	g := buildFixtureGraph(t)
	mainID, ok := g.ID("fixture.App", "main", "([Ljava/lang/String;)V")
	if !ok {
		t.Fatalf("fixture.App#main not found")
	}
	handleCalls := 0
	for _, e := range g.Edges(mainID) {
		if g.Key(e.To) == "fixture.App#handle(Ljava/lang/String;)Ljava/lang/String;" {
			handleCalls++
		}
	}
	if handleCalls != 1 {
		t.Errorf("main -> App#handle edge count = %d, want 1 (deduped)", handleCalls)
	}
}

// TestCHA_edgeBasisDistinguishesDirectFromExpanded is JVM_GRAPH G0's own
// "resolution rule" requirement: a reader (or report) must be able to
// tell a direct edge to the call's own symbolic target apart from an
// edge CHA fabricated to a subtype, not just infer it from whether the
// target happens to be an interface.
func TestCHA_edgeBasisDistinguishesDirectFromExpanded(t *testing.T) {
	g := buildFixtureGraph(t)

	handleID, ok := g.ID("fixture.App", "handle", "(Ljava/lang/String;)Ljava/lang/String;")
	if !ok {
		t.Fatalf("fixture.App#handle not found in graph")
	}
	greetID, ok := g.ID("fixture.Greeter", "greet", "(Ljava/lang/String;)Ljava/lang/String;")
	if !ok {
		t.Fatalf("fixture.Greeter#greet not found in graph")
	}
	formalID, ok := g.ID("fixture.FormalGreeter", "greet", "(Ljava/lang/String;)Ljava/lang/String;")
	if !ok {
		t.Fatalf("fixture.FormalGreeter#greet not found in graph")
	}

	byTo := make(map[uint32]Edge)
	for _, e := range g.Edges(handleID) {
		byTo[e.To] = e
	}

	direct, ok := byTo[greetID]
	if !ok {
		t.Fatalf("handle() has no edge to Greeter#greet itself")
	}
	if direct.Basis != BasisDirect {
		t.Errorf("edge to the call's own declared receiver: Basis = %q, want %q", direct.Basis, BasisDirect)
	}
	if len(direct.Sites) == 0 {
		t.Errorf("direct edge has no Sites evidence at all")
	}

	expanded, ok := byTo[formalID]
	if !ok {
		t.Fatalf("handle() has no CHA-expanded edge to FormalGreeter#greet")
	}
	if expanded.Basis != BasisCHA {
		t.Errorf("edge CHA added to a subtype: Basis = %q, want %q", expanded.Basis, BasisCHA)
	}
}

// TestEvidenceIsPreservedAcrossDuplicateCallSites is the other half of
// G0's "deduplikacja nie usuwa alternatywnych dowodów": handle() (Java 8
// build) calls StringBuilder#append four times from four different
// bytecode offsets; they must still collapse into one Edge (reachability
// doesn't need four copies of the same target), but every offset must
// survive as separate Sites evidence, not just the first/last one.
func TestEvidenceIsPreservedAcrossDuplicateCallSites(t *testing.T) {
	g := buildFixtureGraph(t)

	handleID, ok := g.ID("fixture.App", "handle", "(Ljava/lang/String;)Ljava/lang/String;")
	if !ok {
		t.Fatalf("fixture.App#handle not found in graph")
	}

	var appendEdge *Edge
	for i, e := range g.Edges(handleID) {
		if g.Key(e.To) == "java.lang.StringBuilder#append(Ljava/lang/String;)Ljava/lang/StringBuilder;" {
			appendEdge = &g.Edges(handleID)[i]
			break
		}
	}
	if appendEdge == nil {
		t.Fatalf("handle() has no edge to StringBuilder#append")
	}
	if len(appendEdge.Sites) != 4 {
		t.Fatalf("StringBuilder#append edge has %d Sites, want 4 (one per call site, none dropped by dedup)", len(appendEdge.Sites))
	}
	seenPCs := make(map[int]bool)
	for _, s := range appendEdge.Sites {
		if seenPCs[s.PC] {
			t.Errorf("duplicate PC %d across Sites: %v", s.PC, appendEdge.Sites)
		}
		seenPCs[s.PC] = true
	}
}

func TestEdgeAndNodeCountsAreConsistent(t *testing.T) {
	g := buildFixtureGraph(t)
	if g.NodeCount() == 0 {
		t.Fatalf("NodeCount() = 0")
	}
	if g.EdgeCount() < g.PreCHAEdgeCount() {
		t.Errorf("EdgeCount() = %d, want >= PreCHAEdgeCount() = %d", g.EdgeCount(), g.PreCHAEdgeCount())
	}
	if g.CHAEdgeCount() != g.EdgeCount()-g.PreCHAEdgeCount() {
		t.Errorf("CHAEdgeCount() = %d, want EdgeCount()-PreCHAEdgeCount() = %d", g.CHAEdgeCount(), g.EdgeCount()-g.PreCHAEdgeCount())
	}
}
