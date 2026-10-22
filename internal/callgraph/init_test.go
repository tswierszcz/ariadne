package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func hasInitEdge(g *Graph, from uint32, toClass string) bool {
	toID, ok := g.ID(toClass, "<clinit>", "()V")
	if !ok {
		return false
	}
	for _, e := range g.Edges(from) {
		if e.To == toID && e.Basis == BasisInit {
			return true
		}
	}
	return false
}

func clinitMethod() hierarchy.Method {
	return hierarchy.Method{Method: classfile.Method{Name: "<clinit>", Descriptor: "()V", AccessFlags: classfile.AccStatic}}
}

// TestInit_inheritedStaticField is JVM_GRAPH G3's "odziedziczone pole
// statyczne" fixture: a getstatic on a subclass that declares no field
// of its own must still trigger the superclass's <clinit> — the one
// that actually declares the field — via hierarchy.ResolveField's
// JVMS-5.4.3.2 order.
func TestInit_inheritedStaticField(t *testing.T) {
	base := &hierarchy.ClassInfo{
		FQCN:       "test.Base",
		SuperClass: "java.lang.Object",
		Fields:     []classfile.Field{{Name: "counter", Descriptor: "I"}},
		Methods:    []hierarchy.Method{clinitMethod()},
	}
	sub := &hierarchy.ClassInfo{FQCN: "test.Sub", SuperClass: "test.Base"} // no fields, no <clinit> of its own
	caller := &hierarchy.ClassInfo{
		FQCN:       "test.Caller",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{{
			Method:   classfile.Method{Name: "run", Descriptor: "()V", AccessFlags: classfile.AccPublic},
			InitRefs: []classfile.InitRef{{Kind: classfile.InitGetStatic, Owner: "test.Sub", Name: "counter", Desc: "I"}},
		}},
	}

	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{base, sub, caller})
	g := Build(h, nil)

	callerID, _ := g.ID("test.Caller", "run", "()V")
	if !hasInitEdge(g, callerID, "test.Base") {
		t.Errorf("no init edge to test.Base#<clinit> for a getstatic on test.Sub (which inherits the field); edges: %v", targetNames(g, g.Edges(callerID)))
	}
}

// TestInit_noClinitProducesNoEdge is G3's "brak <clinit>"/"przypadki bez
// inicjalizacji" fixture: `new` on a class with no static initializer at
// all must not fabricate an edge to a <clinit> node that doesn't exist.
func TestInit_noClinitProducesNoEdge(t *testing.T) {
	plain := &hierarchy.ClassInfo{FQCN: "test.NoInit", SuperClass: "java.lang.Object"} // no <clinit>
	caller := &hierarchy.ClassInfo{
		FQCN:       "test.Caller2",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{{
			Method:   classfile.Method{Name: "run", Descriptor: "()V", AccessFlags: classfile.AccPublic},
			InitRefs: []classfile.InitRef{{Kind: classfile.InitNew, Owner: "test.NoInit"}},
		}},
	}

	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{plain, caller})
	g := Build(h, nil)

	callerID, _ := g.ID("test.Caller2", "run", "()V")
	for _, e := range g.Edges(callerID) {
		if e.Basis == BasisInit {
			t.Errorf("unexpected init edge %+v for a class with no <clinit>", e)
		}
	}
	if _, ok := g.ID("test.NoInit", "<clinit>", "()V"); ok {
		t.Errorf("test.NoInit#<clinit> node exists even though the class declares none")
	}
}

// TestInit_cycleDoesNotHang is G3's "cykl" fixture: two classes whose
// <clinit>s instantiate each other. The JVM handles this at runtime via
// a per-thread "being initialized" flag; ariadne's static graph has no
// such concept, so this is simply a cycle in the graph — Build must
// still terminate (it's a single finite pass, no recursion that could
// loop) and both edges must exist.
func TestInit_cycleDoesNotHang(t *testing.T) {
	a := &hierarchy.ClassInfo{
		FQCN:       "test.CycleA",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{{
			Method:   classfile.Method{Name: "<clinit>", Descriptor: "()V", AccessFlags: classfile.AccStatic},
			InitRefs: []classfile.InitRef{{Kind: classfile.InitNew, Owner: "test.CycleB"}},
		}},
	}
	b := &hierarchy.ClassInfo{
		FQCN:       "test.CycleB",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{{
			Method:   classfile.Method{Name: "<clinit>", Descriptor: "()V", AccessFlags: classfile.AccStatic},
			InitRefs: []classfile.InitRef{{Kind: classfile.InitNew, Owner: "test.CycleA"}},
		}},
	}

	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{a, b})
	g := Build(h, nil) // must return, not hang

	aClinit, _ := g.ID("test.CycleA", "<clinit>", "()V")
	bClinit, _ := g.ID("test.CycleB", "<clinit>", "()V")
	if !hasInitEdge(g, aClinit, "test.CycleB") {
		t.Errorf("missing init edge test.CycleA#<clinit> -> test.CycleB#<clinit>")
	}
	if !hasInitEdge(g, bClinit, "test.CycleA") {
		t.Errorf("missing init edge test.CycleB#<clinit> -> test.CycleA#<clinit>")
	}
}

// TestInit_staticCallAndSuperclassChain covers both remaining G3
// triggers in one shape: invokestatic requiring its resolved declaring
// class's <clinit>, and a class with its own <clinit> requiring its
// superclass's <clinit> to run first (JVMS 5.5), independent of
// anything that <clinit>'s own body does.
func TestInit_staticCallAndSuperclassChain(t *testing.T) {
	base := &hierarchy.ClassInfo{
		FQCN:       "test.Base2",
		SuperClass: "java.lang.Object",
		Methods:    []hierarchy.Method{clinitMethod()},
	}
	sub := &hierarchy.ClassInfo{
		FQCN:       "test.Sub2",
		SuperClass: "test.Base2",
		Methods: []hierarchy.Method{
			clinitMethod(),
			{Method: classfile.Method{Name: "staticMethod", Descriptor: "()V", AccessFlags: classfile.AccStatic | classfile.AccPublic}},
		},
	}
	caller := &hierarchy.ClassInfo{
		FQCN:       "test.Caller3",
		SuperClass: "java.lang.Object",
		Methods: []hierarchy.Method{{
			Method: classfile.Method{Name: "run", Descriptor: "()V", AccessFlags: classfile.AccPublic},
			Calls:  []classfile.Call{{Kind: classfile.CallStatic, Owner: "test.Sub2", Name: "staticMethod", Desc: "()V"}},
		}},
	}

	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{base, sub, caller})
	g := Build(h, nil)

	callerID, _ := g.ID("test.Caller3", "run", "()V")
	if !hasInitEdge(g, callerID, "test.Sub2") {
		t.Errorf("invokestatic on test.Sub2#staticMethod produced no init edge to test.Sub2#<clinit>; edges: %v", targetNames(g, g.Edges(callerID)))
	}

	sub2Clinit, _ := g.ID("test.Sub2", "<clinit>", "()V")
	if !hasInitEdge(g, sub2Clinit, "test.Base2") {
		t.Errorf("test.Sub2#<clinit> has no superclass-chain init edge to test.Base2#<clinit>")
	}
}
