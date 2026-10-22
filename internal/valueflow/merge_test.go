package valueflow

import (
	"fmt"
	"testing"

	"ariadne/internal/bytecodeir"
)

// DATA_FLOW D0's own "Gotowe": the contract must have worked examples
// for a branch, a loop, a call, and an unsupported operation. These
// four tests are exactly that — D1/D2 don't exist yet, so each one
// hand-builds the Values/Sites D1/D2 would eventually produce for the
// described bytecode shape, and checks the D0 contract (Merge, Unknown,
// the OpKind enum) represents it correctly. They are not a claim that
// ariadne can analyze real bytecode yet.

// Example 1: branch. Two branches of an `if` assign different string
// constants to the same local; at the join point after the branch,
// Merge must combine them into one KindSet, not silently pick one or
// widen prematurely (two candidates is well within DefaultBudget).
func TestExample_branch(t *testing.T) {
	thenBranch := Const("java.lang.String", "yes")
	elseBranch := Const("java.lang.String", "no")

	joined := Merge(thenBranch, elseBranch)

	if joined.Kind != KindSet {
		t.Fatalf("Merge(branch) = %+v, want KindSet", joined)
	}
	if len(joined.Candidates) != 2 {
		t.Fatalf("Merge(branch).Candidates = %v, want 2 entries", joined.Candidates)
	}
}

// Example 2: loop. A loop-carried local accumulates a different string
// constant on each of five iterations — more than DefaultBudget (4)
// distinct candidates. The contract's own requirement ("Nadmiar
// kandydatów zastępować odpowiednim stanem unknown z powodem") means
// this must widen to KindUnknown with a reason once the budget is
// exceeded, not silently stop growing the Candidates list.
func TestExample_loop(t *testing.T) {
	acc := Const("java.lang.String", "seed")
	for i, s := range []string{"a", "b", "c", "d", "e"} {
		acc = Merge(acc, Const("java.lang.String", s))
		if acc.Kind == KindUnknown {
			break
		}
		_ = i
	}

	if acc.Kind != KindUnknown {
		t.Fatalf("after merging 6 distinct constants, Kind = %v, want KindUnknown (budget exceeded)", acc.Kind)
	}
	if acc.Reason == "" {
		t.Errorf("KindUnknown value has no Reason")
	}
}

// Example 3: call. A value coming from an external method call's return
// has no summary available yet (D3's own scope, not D0-D2) — the
// contract represents this as KindUnknown with an explicit reason
// naming the call site, not as a fabricated constant or an empty Value.
func TestExample_call(t *testing.T) {
	callSite := bytecodeir.Site{
		Method:     "com.example.App#handle",
		Descriptor: "(Ljava/lang/String;)Ljava/lang/String;",
		PC:         42,
		Op:         bytecodeir.OpCall,
	}
	result := Unknown(fmt.Sprintf("return value of %s at pc=%d: no interprocedural summary available (DATA_FLOW D3)", callSite.Method, callSite.PC))

	if result.Kind != KindUnknown {
		t.Fatalf("Kind = %v, want KindUnknown", result.Kind)
	}
	if result.Reason == "" {
		t.Errorf("call-result Value has no Reason")
	}
}

// Example 4: unsupported operation. An old jsr/ret pair — the plan's
// own named example of an instruction D1's CFG builder recognizes but
// does not model — pairs bytecodeir.OpUnsupported with a KindUnknown
// Value carrying a reason, never a fabricated fact pretending the
// operation was understood.
func TestExample_unsupportedOperation(t *testing.T) {
	site := bytecodeir.Site{
		Method:     "com.example.Legacy#oldStyleFinally",
		Descriptor: "()V",
		PC:         17,
		Op:         bytecodeir.OpUnsupported,
	}
	value := Unknown(fmt.Sprintf("jsr/ret at pc=%d is not modeled", site.PC))

	if site.Op != bytecodeir.OpUnsupported {
		t.Fatalf("Op = %v, want OpUnsupported", site.Op)
	}
	if value.Kind != KindUnknown || value.Reason == "" {
		t.Fatalf("value = %+v, want KindUnknown with a reason", value)
	}
}
