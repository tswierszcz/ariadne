package reflectinvoke

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func baseNewInstanceSite() NewInstanceSite {
	return NewInstanceSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9}
}

// TestFindNewInstanceEdges_exactConstructorResolves confirms the
// positive path: a resolved OwnerFQCN/ParamTypes finds the exact
// "<init>" declared directly on that class, full descriptor (return
// type always "V") carried through verbatim.
func TestFindNewInstanceEdges_exactConstructorResolves(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Target", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "<init>", Descriptor: "(Ljava/lang/String;)V", AccessFlags: classfile.AccPublic}},
		}},
	})
	resolved := []ResolvedNewInstance{{
		Site: baseNewInstanceSite(), OwnerFQCN: "test.Target", ParamTypes: []string{"java.lang.String"},
	}}
	results := FindNewInstanceEdges(h, resolved)
	if len(results) != 1 || results[0].Gap != "" {
		t.Fatalf("results = %+v, want exactly 1 with no Gap", results)
	}
	if results[0].TargetDesc != "(Ljava/lang/String;)V" {
		t.Errorf("TargetDesc = %q, want (Ljava/lang/String;)V", results[0].TargetDesc)
	}
}

// TestFindNewInstanceEdges_argCountMismatchIsAGap confirms a proven
// argument-count mismatch never fabricates a successful-allocation
// edge — FindInvokeEdges's own counterpart, applied to allocation.
func TestFindNewInstanceEdges_argCountMismatchIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Target", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "<init>", Descriptor: "(Ljava/lang/String;)V", AccessFlags: classfile.AccPublic}},
		}},
	})
	resolved := []ResolvedNewInstance{{
		Site: baseNewInstanceSite(), OwnerFQCN: "test.Target",
		ParamTypes: []string{"java.lang.String"}, ArgCount: 2, ArgCountKnown: true,
	}}
	results := FindNewInstanceEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
	if edges := NewInstanceEdgeTargets(results); len(edges) != 0 {
		t.Errorf("edges = %+v, want none (a Gap contributes no edge)", edges)
	}
}

// TestFindNewInstanceEdges_arrayParamTypeIsAGap confirms a parameter
// type that isn't a plain class literal is rejected, not mis-converted.
func TestFindNewInstanceEdges_arrayParamTypeIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{{FQCN: "test.Target"}})
	resolved := []ResolvedNewInstance{{
		Site: baseNewInstanceSite(), OwnerFQCN: "test.Target", ParamTypes: []string{"[Ljava.lang.String;"},
	}}
	results := FindNewInstanceEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
}

// TestFindNewInstanceEdges_noSuchConstructorIsAGap confirms a class
// that declares no constructor matching the resolved parameter types
// is a Gap, not a dropped result — never, say, silently matching an
// unrelated overload.
func TestFindNewInstanceEdges_noSuchConstructorIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Target", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "<init>", Descriptor: "()V", AccessFlags: classfile.AccPublic}},
		}},
	})
	resolved := []ResolvedNewInstance{{
		Site: baseNewInstanceSite(), OwnerFQCN: "test.Target", ParamTypes: []string{"java.lang.String"},
	}}
	results := FindNewInstanceEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
	if edges := NewInstanceEdgeTargets(results); len(edges) != 0 {
		t.Errorf("edges = %+v, want none (a Gap contributes no edge)", edges)
	}
}

// TestFindNewInstanceEdges_classNotInHierarchyIsAGap confirms a
// resolved OwnerFQCN absent from the analyzed hierarchy (shouldn't
// normally happen — the Class literal that produced it did resolve to
// something — but defended against regardless) is a Gap, not a panic.
func TestFindNewInstanceEdges_classNotInHierarchyIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	resolved := []ResolvedNewInstance{{Site: baseNewInstanceSite(), OwnerFQCN: "test.Missing"}}
	results := FindNewInstanceEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
}

// TestFindNewInstanceEdges_unresolvedContributesNothing confirms an
// unresolved receiver (OwnerFQCN == "") is skipped entirely, not
// reported as a Gap of its own — FindInvokeEdges's own counterpart.
func TestFindNewInstanceEdges_unresolvedContributesNothing(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	resolved := []ResolvedNewInstance{{Site: baseNewInstanceSite(), UnresolvedReason: "the receiver is typed"}}
	if results := FindNewInstanceEdges(h, resolved); len(results) != 0 {
		t.Errorf("results = %+v, want none", results)
	}
}

// TestNewInstanceEdgeTargets_resolvedGivesOneEdgePerResult confirms
// the plain identity conversion.
func TestNewInstanceEdgeTargets_resolvedGivesOneEdgePerResult(t *testing.T) {
	results := []NewInstanceEdgeResult{{
		ResolvedNewInstance: ResolvedNewInstance{Site: baseNewInstanceSite(), OwnerFQCN: "test.Target"},
		TargetDesc:          "(Ljava/lang/String;)V",
	}}
	want := NewInstanceEdge{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9, TargetFQCN: "test.Target", TargetDesc: "(Ljava/lang/String;)V"}
	edges := NewInstanceEdgeTargets(results)
	if len(edges) != 1 || edges[0] != want {
		t.Errorf("edges = %+v, want [%+v]", edges, want)
	}
}
