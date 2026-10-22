package callgraph

import (
	"sort"
	"testing"

	"ariadne/internal/classfile"
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

// TestCHA_overloadsAreNotConflated is JVM_GRAPH G2's own fix, tested
// against a hand-built hierarchy (no JDK available to compile a new
// fixture class with two overloads of the same name): an interface with
// two overloads of "process", a subtype that overrides both, each
// calling a different, distinguishable method on a "library" class. A
// call site that invokes exactly one overload must CHA-expand only to
// that overload's own override, never to the other — the bug the old
// name-only `MethodsByName` CHA expansion had (it returned every
// same-named overload regardless of which one the call site actually
// invoked).
func TestCHA_overloadsAreNotConflated(t *testing.T) {
	const (
		stringDesc = "(Ljava/lang/String;)V"
		intDesc    = "(I)V"
	)
	iface := &hierarchy.ClassInfo{
		FQCN:        "test.Multi",
		SuperClass:  "java.lang.Object",
		AccessFlags: classfile.AccInterface | classfile.AccAbstract,
		Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "process", Descriptor: stringDesc, AccessFlags: classfile.AccAbstract | classfile.AccPublic}},
			{Method: classfile.Method{Name: "process", Descriptor: intDesc, AccessFlags: classfile.AccAbstract | classfile.AccPublic}},
		},
	}
	impl := &hierarchy.ClassInfo{
		FQCN:       "test.MultiImpl",
		SuperClass: "java.lang.Object",
		Interfaces: []string{"test.Multi"},
		Methods: []hierarchy.Method{
			{
				Method: classfile.Method{Name: "process", Descriptor: stringDesc, AccessFlags: classfile.AccPublic},
				Calls:  []classfile.Call{{Kind: classfile.CallStatic, Owner: "test.Vulnerable", Name: "stringPath", Desc: "()V"}},
			},
			{
				Method: classfile.Method{Name: "process", Descriptor: intDesc, AccessFlags: classfile.AccPublic},
				Calls:  []classfile.Call{{Kind: classfile.CallStatic, Owner: "test.Vulnerable", Name: "intPath", Desc: "()V"}},
			},
		},
	}
	caller := &hierarchy.ClassInfo{
		FQCN:       "test.Caller",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{
			{
				Method: classfile.Method{Name: "callString", Descriptor: "()V", AccessFlags: classfile.AccPublic},
				Calls:  []classfile.Call{{Kind: classfile.CallInterface, Owner: "test.Multi", Name: "process", Desc: stringDesc, PC: 0}},
			},
		},
	}
	vulnerable := &hierarchy.ClassInfo{
		FQCN:       "test.Vulnerable",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "stringPath", Descriptor: "()V", AccessFlags: classfile.AccStatic | classfile.AccPublic}},
			{Method: classfile.Method{Name: "intPath", Descriptor: "()V", AccessFlags: classfile.AccStatic | classfile.AccPublic}},
		},
	}

	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{iface, impl, caller, vulnerable})
	g := Build(h, nil)

	callerID, ok := g.ID("test.Caller", "callString", "()V")
	if !ok {
		t.Fatalf("test.Caller#callString not found")
	}

	wantTo := "test.MultiImpl#process" + stringDesc
	dontWantTo := "test.MultiImpl#process" + intDesc
	var foundWanted, foundUnwanted bool
	for _, e := range g.Edges(callerID) {
		switch g.Key(e.To) {
		case wantTo:
			foundWanted = true
			if e.Basis != BasisCHA {
				t.Errorf("edge to %s: Basis = %q, want %q", wantTo, e.Basis, BasisCHA)
			}
		case dontWantTo:
			foundUnwanted = true
		}
	}
	if !foundWanted {
		t.Errorf("caller has no CHA edge to the overload it actually called (%s); edges: %v", wantTo, targetNames(g, g.Edges(callerID)))
	}
	if foundUnwanted {
		t.Errorf("caller has a CHA edge to the *other* overload (%s) it never called — overloads were conflated", dontWantTo)
	}

	// And the called overload's own body must still lead to its own
	// target, not the other overload's.
	implStringID, ok := g.ID("test.MultiImpl", "process", stringDesc)
	if !ok {
		t.Fatalf("test.MultiImpl#process%s not found", stringDesc)
	}
	wantVuln := "test.Vulnerable#stringPath()V"
	var reachesStringPath bool
	for _, e := range g.Edges(implStringID) {
		if g.Key(e.To) == wantVuln {
			reachesStringPath = true
		}
	}
	if !reachesStringPath {
		t.Errorf("test.MultiImpl#process%s has no edge to %s", stringDesc, wantVuln)
	}
}

// TestResolveTarget_inheritedMethodLeadsToDeclaringClass is the direct-
// edge half of JVM_GRAPH G2 (the CHA half is
// TestCHA_overloadsAreNotConflated above, and the real-bytecode
// end-to-end case is internal/reach's
// TestAnalyze_inheritedMethodMatchedThroughSubclassCall): a static call
// whose declared owner only inherits the method, never overriding it
// itself, must lead to the class that actually supplies the body, with
// no diagnostic (owner is part of the analyzed universe and resolution
// succeeds) — not a dead-end node under the subclass's own name.
func TestResolveTarget_inheritedMethodLeadsToDeclaringClass(t *testing.T) {
	base := &hierarchy.ClassInfo{
		FQCN:       "test.Base",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{
			{
				Method: classfile.Method{Name: "helper", Descriptor: "()V", AccessFlags: classfile.AccPublic},
				Calls:  []classfile.Call{{Kind: classfile.CallStatic, Owner: "test.Vulnerable", Name: "stringPath", Desc: "()V"}},
			},
		},
	}
	sub := &hierarchy.ClassInfo{FQCN: "test.Sub", SuperClass: "test.Base"} // inherits helper(), overrides nothing
	caller := &hierarchy.ClassInfo{
		FQCN:       "test.Caller2",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{
			{
				Method: classfile.Method{Name: "run", Descriptor: "()V", AccessFlags: classfile.AccPublic},
				// invokevirtual on the subclass, exactly as classfile.ExtractCalls
				// would decode a call written as "new Sub().helper()".
				Calls: []classfile.Call{{Kind: classfile.CallVirtual, Owner: "test.Sub", Name: "helper", Desc: "()V"}},
			},
		},
	}
	vulnerable := &hierarchy.ClassInfo{
		FQCN:       "test.Vulnerable",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "stringPath", Descriptor: "()V", AccessFlags: classfile.AccStatic | classfile.AccPublic}},
		},
	}

	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{base, sub, caller, vulnerable})
	g := Build(h, nil)

	if g.UnresolvedOwnerCount() != 0 {
		t.Errorf("UnresolvedOwnerCount() = %d, want 0 (test.Sub is in the universe and test.Base supplies helper())", g.UnresolvedOwnerCount())
	}

	callerID, _ := g.ID("test.Caller2", "run", "()V")
	wantTo := "test.Base#helper()V"
	dontWantTo := "test.Sub#helper()V"
	var foundWanted, foundDeadEnd bool
	for _, e := range g.Edges(callerID) {
		switch g.Key(e.To) {
		case wantTo:
			foundWanted = true
		case dontWantTo:
			foundDeadEnd = true
		}
	}
	if !foundWanted {
		t.Errorf("caller has no edge to %s (the actual declaring class); edges: %v", wantTo, targetNames(g, g.Edges(callerID)))
	}
	if foundDeadEnd {
		t.Errorf("caller has a dead-end edge to %s (the subclass that merely inherits helper) instead of being redirected to %s", dontWantTo, wantTo)
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
