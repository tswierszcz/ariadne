package valueflow

import "fmt"

// DefaultBudget bounds how many distinct alternatives Merge keeps in a
// KindSet before widening to KindUnknown (DATA_FLOW D0's own "budżety").
// D1/D2 may use a tighter budget inside a loop body, where the same join
// point is revisited every iteration (see MergeWithBudget and the loop
// example in merge_test.go) — the exact number is a policy knob, not
// part of the contract itself.
const DefaultBudget = 4

// Merge joins two Values found at a control-flow join point — DATA_FLOW
// D0's "kolejność poszerzania i łączenia stanów": identical values
// collapse to one, different values combine into a KindSet bounded by
// DefaultBudget, and either input already being KindUnknown — or the
// combined candidate count exceeding the budget — widens the result to
// KindUnknown with an explicit reason, never a Candidates list truncated
// to fit while still claiming to be a KindSet.
func Merge(a, b Value) Value {
	return MergeWithBudget(a, b, DefaultBudget)
}

// MergeWithBudget is Merge with an explicit budget.
func MergeWithBudget(a, b Value, budget int) Value {
	if a.Kind == KindUnknown {
		return a
	}
	if b.Kind == KindUnknown {
		return b
	}
	if a.Equal(b) {
		return a
	}

	candidates := candidatesOf(a)
	for _, c := range candidatesOf(b) {
		if !containsEqual(candidates, c) {
			candidates = append(candidates, c)
		}
	}

	if len(candidates) == 1 {
		return candidates[0]
	}
	if len(candidates) > budget {
		return Unknown(fmt.Sprintf(
			"merge produced %d distinct candidates, exceeding the budget of %d",
			len(candidates), budget))
	}
	return Value{Kind: KindSet, Candidates: candidates}
}

func candidatesOf(v Value) []Value {
	if v.Kind == KindSet {
		return append([]Value(nil), v.Candidates...)
	}
	return []Value{v}
}

func containsEqual(vs []Value, v Value) bool {
	for _, existing := range vs {
		if existing.Equal(v) {
			return true
		}
	}
	return false
}
