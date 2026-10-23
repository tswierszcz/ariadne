package reflectinvoke

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func baseSite() InvokeSite {
	return InvokeSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9}
}

// TestFindInvokeEdges_realDispatchPrefersObjType confirms R2's own
// "real dispatch" requirement: when the invocation target's own
// concrete allocation type is known, dispatch resolves against it, not
// blindly against OwnerFQCN (the Class the lookup happened to be
// performed on, here an unrelated supertype that doesn't itself
// declare the method at all).
func TestFindInvokeEdges_realDispatchPrefersObjType(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Base"}, // OwnerFQCN: declares nothing of its own
		{FQCN: "test.Impl", SuperClass: "test.Base", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "bar", Descriptor: "(Ljava/lang/String;)V", AccessFlags: classfile.AccPublic}},
		}},
	})
	resolved := []ResolvedInvoke{{
		Site: baseSite(), OwnerFQCN: "test.Base", MethodName: "bar",
		ParamTypes: []string{"java.lang.String"}, ObjType: "test.Impl",
	}}
	results := FindInvokeEdges(h, resolved)
	if len(results) != 1 || results[0].Gap != "" {
		t.Fatalf("results = %+v, want exactly 1 with no Gap", results)
	}
	if results[0].DeclaringFQCN != "test.Impl" {
		t.Errorf("DeclaringFQCN = %q, want test.Impl (real dispatch, not OwnerFQCN)", results[0].DeclaringFQCN)
	}
}

// TestFindInvokeEdges_fallsBackToOwnerFQCNWhenObjTypeUnknown confirms
// the conservative fallback: without a provable invocation-target type,
// dispatch resolves against the Class the lookup was performed on, the
// only basis left — never a guessed subtype.
func TestFindInvokeEdges_fallsBackToOwnerFQCNWhenObjTypeUnknown(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Base", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "bar", Descriptor: "()V", AccessFlags: classfile.AccPublic}},
		}},
	})
	resolved := []ResolvedInvoke{{Site: baseSite(), OwnerFQCN: "test.Base", MethodName: "bar"}}
	results := FindInvokeEdges(h, resolved)
	if len(results) != 1 || results[0].Gap != "" || results[0].DeclaringFQCN != "test.Base" {
		t.Fatalf("results = %+v, want exactly 1 resolving to test.Base", results)
	}
}

// TestFindInvokeEdges_argCountMismatchIsAGap confirms a proven
// argument-count mismatch never fabricates a successful-invocation
// edge, even though the receiver and parameter types resolved fine.
func TestFindInvokeEdges_argCountMismatchIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Base", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "bar", Descriptor: "(Ljava/lang/String;)V", AccessFlags: classfile.AccPublic}},
		}},
	})
	resolved := []ResolvedInvoke{{
		Site: baseSite(), OwnerFQCN: "test.Base", MethodName: "bar",
		ParamTypes: []string{"java.lang.String"}, ArgCount: 2, ArgCountKnown: true,
	}}
	results := FindInvokeEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
	if edges := EdgeTargets(results); len(edges) != 0 {
		t.Errorf("edges = %+v, want none (a Gap contributes no edge)", edges)
	}
}

// TestFindInvokeEdges_arrayParamTypeIsAGap confirms a parameter type
// that isn't a plain class literal (an array-typed Class literal, read
// out of the constant pool in its own internal descriptor shape) is
// rejected, not mis-converted into an invalid descriptor.
func TestFindInvokeEdges_arrayParamTypeIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{{FQCN: "test.Base"}})
	resolved := []ResolvedInvoke{{
		Site: baseSite(), OwnerFQCN: "test.Base", MethodName: "bar",
		ParamTypes: []string{"[Ljava.lang.String;"},
	}}
	results := FindInvokeEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
}

// TestFindInvokeEdges_noSuchMethodIsAGap confirms an unresolvable
// dispatch target is a Gap, not a dropped result.
func TestFindInvokeEdges_noSuchMethodIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{{FQCN: "test.NoSuchMethod"}})
	resolved := []ResolvedInvoke{{Site: baseSite(), OwnerFQCN: "test.NoSuchMethod", MethodName: "bar"}}
	results := FindInvokeEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
	if edges := EdgeTargets(results); len(edges) != 0 {
		t.Errorf("edges = %+v, want none (a Gap contributes no edge)", edges)
	}
}

// TestFindInvokeEdges_unresolvedInvokeContributesNothing confirms an
// unresolved receiver (OwnerFQCN == "") is skipped entirely, not
// reported as a Gap of its own — the Gap concept is for a resolved
// Method whose *dispatch* still fails, not for a Method that was never
// identified in the first place (ResolvedInvoke.UnresolvedReason
// already covers that).
func TestFindInvokeEdges_unresolvedInvokeContributesNothing(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	resolved := []ResolvedInvoke{{Site: baseSite(), UnresolvedReason: "the receiver is typed"}}
	if results := FindInvokeEdges(h, resolved); len(results) != 0 {
		t.Errorf("results = %+v, want none", results)
	}
}

// TestEdgeTargets_resolvedGivesOneEdgePerResult confirms the plain
// identity conversion.
func TestEdgeTargets_resolvedGivesOneEdgePerResult(t *testing.T) {
	results := []InvokeEdgeResult{{
		ResolvedInvoke: ResolvedInvoke{Site: baseSite(), MethodName: "bar"},
		DeclaringFQCN:  "test.Impl",
	}}
	want := InvokeEdge{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9, TargetFQCN: "test.Impl", TargetMethod: "bar"}
	edges := EdgeTargets(results)
	if len(edges) != 1 || edges[0] != want {
		t.Errorf("edges = %+v, want [%+v]", edges, want)
	}
}
