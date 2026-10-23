package hierarchy

import (
	"ariadne/internal/classfile"
	"testing"
)

// TestResolveByParams_matchesExactParamsIgnoringReturnType confirms
// REFLECTION R2's own exact-parameter resolution follows the same
// inheritance order as ResolveExact, and ignores return type (a
// getMethod call never proves one).
func TestResolveByParams_matchesExactParamsIgnoringReturnType(t *testing.T) {
	h := NewForTesting([]*ClassInfo{
		{FQCN: "test.Base", Methods: []Method{
			{Method: classfile.Method{Name: "bar", Descriptor: "(Ljava/lang/String;)I", AccessFlags: classfile.AccPublic}},
		}},
		{FQCN: "test.Foo", SuperClass: "test.Base"}, // inherits bar(), doesn't override it
	})
	declaring, found := h.ResolveByParams("test.Foo", "bar", []string{"Ljava/lang/String;"})
	if !found || declaring != "test.Base" {
		t.Fatalf("ResolveByParams = (%q, %v), want (test.Base, true)", declaring, found)
	}
}

// TestResolveByParams_paramCountMismatchIsNotFound confirms a
// same-named method with a different parameter count is never matched
// — distinct overloads must stay distinct.
func TestResolveByParams_paramCountMismatchIsNotFound(t *testing.T) {
	h := NewForTesting([]*ClassInfo{
		{FQCN: "test.Foo", Methods: []Method{
			{Method: classfile.Method{Name: "bar", Descriptor: "(Ljava/lang/String;I)V", AccessFlags: classfile.AccPublic}},
		}},
	})
	if _, found := h.ResolveByParams("test.Foo", "bar", []string{"Ljava/lang/String;"}); found {
		t.Errorf("ResolveByParams matched a different-arity overload, want not found")
	}
}

// TestResolveByParams_noSuchMethodIsNotFound confirms a name with no
// declaration anywhere in the hierarchy is honestly unresolved.
func TestResolveByParams_noSuchMethodIsNotFound(t *testing.T) {
	h := NewForTesting([]*ClassInfo{{FQCN: "test.Foo"}})
	if _, found := h.ResolveByParams("test.Foo", "bar", nil); found {
		t.Errorf("ResolveByParams found a method that doesn't exist")
	}
}
