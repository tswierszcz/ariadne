package methodhandle

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestFindInvokeSites(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "m", Descriptor: "()V"}, Calls: []classfile.Call{
				{Owner: methodHandleOwner, Name: "invokeExact", Desc: "()V", PC: 9},
				{Owner: methodHandleOwner, Name: "invoke", Desc: "(I)I", PC: 20},
				{Owner: "java.lang.String", Name: "valueOf", Desc: "(I)Ljava/lang/String;", PC: 1},
			}},
		}},
	})

	sites := FindInvokeSites(h)
	if len(sites) != 2 {
		t.Fatalf("sites = %+v, want exactly 2", sites)
	}
	var sawExact, sawInvoke bool
	for _, s := range sites {
		if s.PC == 9 && s.Exact {
			sawExact = true
		}
		if s.PC == 20 && !s.Exact {
			sawInvoke = true
		}
	}
	if !sawExact || !sawInvoke {
		t.Errorf("sites = %+v, want one invokeExact at pc=9 and one invoke at pc=20", sites)
	}
}

func TestFindHandleEdges_resolvesDeclaringClass(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Base", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "bar", Descriptor: "()V", AccessFlags: classfile.AccPublic}},
		}},
		{FQCN: "test.Foo", SuperClass: "test.Base"}, // inherits bar(), doesn't override it
	})
	resolved := []ResolvedInvoke{
		{Site: InvokeSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9},
			RequestedOwner: "test.Foo", Method: "bar"},
	}

	results := FindHandleEdges(h, resolved)
	if len(results) != 1 {
		t.Fatalf("results = %+v, want exactly 1", results)
	}
	if results[0].Gap != "" {
		t.Errorf("Gap = %q, want empty", results[0].Gap)
	}
	if results[0].DeclaringFQCN != "test.Base" {
		t.Errorf("DeclaringFQCN = %q, want test.Base (bar() is inherited, not overridden)", results[0].DeclaringFQCN)
	}

	edges := EdgeTargets(results)
	want := HandleEdge{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9, TargetFQCN: "test.Base", TargetMethod: "bar"}
	if len(edges) != 1 || edges[0] != want {
		t.Errorf("edges = %+v, want [%+v]", edges, want)
	}
}

func TestFindHandleEdges_unresolvableTargetIsAGap(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.NoSuchMethod"},
	})
	resolved := []ResolvedInvoke{
		{Site: InvokeSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 1},
			RequestedOwner: "test.NoSuchMethod", Method: "bar"},
	}

	results := FindHandleEdges(h, resolved)
	if len(results) != 1 || results[0].Gap == "" {
		t.Fatalf("results = %+v, want exactly 1 with a non-empty Gap", results)
	}
	if edges := EdgeTargets(results); len(edges) != 0 {
		t.Errorf("edges = %+v, want none (a Gap contributes no edge)", edges)
	}
}

func TestFindHandleEdges_unresolvedInvokeContributesNothing(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	resolved := []ResolvedInvoke{
		{Site: InvokeSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 1}, UnresolvedReason: "the receiver is typed"},
	}
	if results := FindHandleEdges(h, resolved); len(results) != 0 {
		t.Errorf("results = %+v, want none", results)
	}
}
