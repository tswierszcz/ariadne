// Package dataflow is DATA_FLOW D2's local-value analysis: a monotone
// abstract interpreter over D1's CFG that produces real
// internal/valueflow.Value facts from real bytecode — string/class
// constants propagated through variables and branches, call arguments
// exposed to any consumer via valueflow.Facts.
//
// Scope, per D2's own plan: local values only (no interprocedural
// summaries — a call's return value is always KindUnknown; that is D3).
// Arithmetic results are deliberately not tracked (D2's own "stałe
// napisowe/klasowe" — string/class literals, not numeric constants);
// every arithmetic/conversion/comparison op still gets a real,
// width-correct stack effect so later instructions' own facts stay
// sound, just with a Typed/Unknown result instead of a guessed number.
package dataflow

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/valueflow"
)

// MethodFacts is one method's analysis result: valueflow.Facts for every
// call site found, plus whether the analysis gave up partway (Aborted)
// and why.
type MethodFacts struct {
	calls     map[bytecodeir.Site][]valueflow.Value
	receivers map[bytecodeir.Site]valueflow.Value
	aborted   bool
	reason    string
}

var _ valueflow.Facts = (*MethodFacts)(nil)

// ArgumentAt implements valueflow.Facts. Per that interface's own
// contract, a Site that isn't a real call site found in this method (the
// wrong Op, or one this method's bytecode never contains) is a caller
// error and panics; a call site that exists but whose analysis Aborted
// still returns a (Kind: KindUnknown) Value rather than panicking — the
// site is real, only its facts are incomplete.
func (f *MethodFacts) ArgumentAt(at bytecodeir.Site, i int) valueflow.Value {
	if at.Op != bytecodeir.OpCall {
		panic(fmt.Sprintf("dataflow: ArgumentAt: Site.Op = %q, want %q", at.Op, bytecodeir.OpCall))
	}
	args, ok := f.calls[at]
	if !ok {
		panic(fmt.Sprintf("dataflow: ArgumentAt: no call recorded at %+v", at))
	}
	if i < 0 || i >= len(args) {
		panic(fmt.Sprintf("dataflow: ArgumentAt: argument index %d out of range (call has %d arguments)", i, len(args)))
	}
	return args[i]
}

// ReceiverAt implements valueflow.Facts, same panic/Aborted discipline
// as ArgumentAt: a Site with no receiver recorded (wrong Op, a static
// call, or one this method's bytecode never contains) is a caller
// error and panics.
func (f *MethodFacts) ReceiverAt(at bytecodeir.Site) valueflow.Value {
	if at.Op != bytecodeir.OpCall {
		panic(fmt.Sprintf("dataflow: ReceiverAt: Site.Op = %q, want %q", at.Op, bytecodeir.OpCall))
	}
	v, ok := f.receivers[at]
	if !ok {
		panic(fmt.Sprintf("dataflow: ReceiverAt: no receiver recorded at %+v (static call, or not a real call site in this method)", at))
	}
	return v
}

// Aborted reports whether the local-value analysis gave up before
// reaching every instruction (an unsupported jsr/ret subroutine, or an
// internal stack-shape inconsistency it does not trust itself to recover
// from) and, if so, why. Every call site still has an entry in Facts —
// unreached ones simply hold KindUnknown with this same reason.
func (f *MethodFacts) Aborted() (bool, string) { return f.aborted, f.reason }
