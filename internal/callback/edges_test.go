package callback

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func runMethod(access uint16) hierarchy.Method {
	return hierarchy.Method{Method: classfile.Method{Name: "run", Descriptor: "()V", AccessFlags: access}}
}

func TestFindCallbackEdges_resolvesDeclaringClass(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.BaseRunnable", Methods: []hierarchy.Method{runMethod(classfile.AccPublic)}},
		{FQCN: "test.MyRunnable", SuperClass: "test.BaseRunnable"}, // inherits run(), doesn't override it
	})
	model, ok := Lookup(fqcnThread, "<init>", "(Ljava/lang/Runnable;)V")
	if !ok {
		t.Fatal("DefaultCatalog is missing Thread(Runnable)")
	}
	resolved := []ResolvedCallback{
		{Site: RegistrationSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 11, Model: model, Outcome: OutcomeCataloged},
			TargetFQCN: "test.MyRunnable"},
	}

	results := FindCallbackEdges(h, resolved)
	if len(results) != 1 {
		t.Fatalf("results = %+v, want exactly 1", results)
	}
	if results[0].Gap != "" {
		t.Errorf("Gap = %q, want empty", results[0].Gap)
	}
	if results[0].DeclaringFQCN != "test.BaseRunnable" {
		t.Errorf("DeclaringFQCN = %q, want test.BaseRunnable (run() is inherited, not overridden)", results[0].DeclaringFQCN)
	}

	edges := EdgeTargets(results)
	if len(edges) != 1 {
		t.Fatalf("edges = %+v, want exactly 1", edges)
	}
	want := CallbackEdge{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 11,
		TargetFQCN: "test.BaseRunnable", TargetMethod: "run", TargetMethodDesc: "()V"}
	if edges[0] != want {
		t.Errorf("edges[0] = %+v, want %+v", edges[0], want)
	}
}

func TestFindCallbackEdges_unresolvableTargetIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.NoRunMethod"}, // declares nothing at all
	})
	model, _ := Lookup(fqcnThread, "<init>", "(Ljava/lang/Runnable;)V")
	resolved := []ResolvedCallback{
		{Site: RegistrationSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 1, Model: model, Outcome: OutcomeCataloged},
			TargetFQCN: "test.NoRunMethod"},
	}

	results := FindCallbackEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
	if edges := EdgeTargets(results); len(edges) != 0 {
		t.Errorf("edges = %+v, want none (a Gap contributes no edge)", edges)
	}
}

// TestFindCallbackEdges_unresolvedIdentityContributesNothing: a
// ResolvedCallback with no TargetFQCN (A1 couldn't resolve it) must not
// appear in FindCallbackEdges' own results at all.
func TestFindCallbackEdges_unresolvedIdentityContributesNothing(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	model, _ := Lookup(fqcnThread, "<init>", "(Ljava/lang/Runnable;)V")
	resolved := []ResolvedCallback{
		{Site: RegistrationSite{Model: model, Outcome: OutcomeCataloged}, UnresolvedReason: "the argument is typed"},
	}
	if results := FindCallbackEdges(h, resolved); len(results) != 0 {
		t.Errorf("results = %+v, want none", results)
	}
}

// TestFindCallbackEdges_lambdaSkipsHierarchyResolution: an
// ImplOwner-based identity (a lambda/method reference) must connect
// straight to that implementation method, never through
// hierarchy.ResolveExact (it is already the exact declaring method),
// and never produce a Gap.
func TestFindCallbackEdges_lambdaSkipsHierarchyResolution(t *testing.T) {
	h := hierarchy.NewForTesting(nil) // empty hierarchy: if ResolveExact were called, this would fail to resolve anything
	model, _ := Lookup(fqcnThread, "<init>", "(Ljava/lang/Runnable;)V")
	resolved := []ResolvedCallback{
		{Site: RegistrationSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9, Model: model, Outcome: OutcomeCataloged},
			ImplOwner: "test.Caller", ImplName: "lambda$m$0", ImplDesc: "()V"},
	}

	results := FindCallbackEdges(h, resolved)
	if len(results) != 1 {
		t.Fatalf("results = %+v, want exactly 1", results)
	}
	if results[0].Gap != "" {
		t.Errorf("Gap = %q, want empty", results[0].Gap)
	}
	if results[0].DeclaringFQCN != "test.Caller" || results[0].DeclaringMethod != "lambda$m$0" || results[0].DeclaringDesc != "()V" {
		t.Errorf("Declaring = %s#%s%s, want test.Caller#lambda$m$0()V", results[0].DeclaringFQCN, results[0].DeclaringMethod, results[0].DeclaringDesc)
	}

	edges := EdgeTargets(results)
	want := CallbackEdge{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9,
		TargetFQCN: "test.Caller", TargetMethod: "lambda$m$0", TargetMethodDesc: "()V"}
	if len(edges) != 1 || edges[0] != want {
		t.Errorf("edges = %+v, want [%+v]", edges, want)
	}
}
