// Package valueflow is DATA_FLOW D0's value-domain contract: the
// abstract values D1's CFG and D2's local-value analysis produce, and
// every later consumer (D3's interprocedural summaries, D4's taint,
// SERVICE_LOADING S1, REFLECTION, ASYNC_CALLBACKS) reads. D0 only
// defines the domain, the widening/merge order, and the Facts API
// (facts.go) — it does not itself compute a single real Value from
// bytecode; that is D1/D2's own scope.
package valueflow

import "ariadne/internal/bytecodeir"

// Kind is the value domain DATA_FLOW D0 must settle before any real
// analysis (D1/D2) can produce or consume a Value: an exact statically
// known constant, an upper-bound type with no known constant, a
// specific allocation site, null, a small explicit set of alternatives,
// or unknown-with-a-reason. Kept as a flat enum, not a type hierarchy,
// so a consumer can switch on it exhaustively without reflection.
type Kind string

const (
	// KindConst is an exact, statically known value — a string, a
	// *Class literal's FQCN, a boxed primitive. Distinct from KindTyped:
	// "this is definitely the string \"foo\"" vs. "this is definitely a
	// java.lang.String, but which one is unknown" are different claims,
	// and conflating them would silently overstate precision.
	KindConst Kind = "const"
	// KindTyped is an upper-bound static type only — no known constant.
	// The common case for a parameter, a field read without a narrower
	// summary, or anything D1/D2 can type but not pin down further.
	KindTyped Kind = "typed"
	// KindAlloc identifies a specific `new`/factory-call Site this value
	// came from — distinct from KindConst: knowing *which* allocation
	// produced a value is not the same as knowing what's inside it.
	KindAlloc Kind = "alloc"
	// KindNull is the compile-time-visible null literal or a value
	// narrowed to definitely-null by control flow.
	KindNull Kind = "null"
	// KindSet is a small, explicit set of alternative Values — e.g. two
	// branches joining with two different constants. Never silently
	// truncated: once len(Candidates) would exceed the budget passed to
	// Merge, the result widens to KindUnknown with a reason instead of
	// keeping a partial list that looks complete (DATA_FLOW D0's own
	// requirement: "Nadmiar kandydatów zastępować odpowiednim stanem
	// unknown z powodem, nie przyciętą listą udającą pełną").
	KindSet Kind = "set"
	// KindUnknown means the value could not be determined — budget
	// exceeded, an unsupported operation (bytecodeir.OpUnsupported), an
	// external call with no summary, or a genuinely dynamic value (user
	// input, a config-read result). Reason is required whenever Kind is
	// KindUnknown; "unknown" without a reason is not an acceptable
	// result from any producer in this package's contract.
	KindUnknown Kind = "unknown"
	// KindFunctional identifies a lambda/method-reference object created
	// by invokedynamic through LambdaMetafactory (DYNAMIC_LINKAGE L0/L1)
	// — see FunctionalValue. Distinct from KindAlloc: the created object
	// isn't a `new` of any class ariadne reads; it's a synthetic
	// implementation of a functional interface that forwards straight to
	// a resolved real method.
	KindFunctional Kind = "functional"
)

// FunctionalValue is DYNAMIC_LINKAGE L0/L1's own model for a lambda or
// method reference: which functional interface/method it implements,
// which real method it dispatches straight to once that method runs,
// and what it closed over at creation. A value in DATA_FLOW's own
// domain — "a program can pass this around, store it, or never invoke
// it" — not a call-graph edge (L0's own "odrębny od gotowej krawędzi
// wywołania"); a consumer that wants an edge builds one from this.
type FunctionalValue struct {
	// Site is the invokedynamic instruction that created this value —
	// its identity, the same role bytecodeir.Site plays for KindAlloc.
	Site bytecodeir.Site
	// SamInterface/SamMethod are the functional interface and its
	// single abstract method this value implements (e.g.
	// "java.lang.Runnable"/"run") — what a consumer eventually calls.
	SamInterface, SamMethod string
	// ImplOwner/ImplName/ImplDesc are the real method SamMethod
	// dispatches straight to: a lambda body's own synthetic method, a
	// method reference's target, or a constructor reference's
	// constructor. Resolved once via the LambdaMetafactory bootstrap,
	// never through CHA — this is not an ordinary virtual call site.
	// "" when the bootstrap couldn't be resolved (e.g.
	// StringConcatFactory): the value still exists as evidence a lambda
	// was created, just with no known implementation.
	ImplOwner, ImplName, ImplDesc string
	// RefKind is the implementation MethodHandle's own JVMS 5.4.3.5
	// reference kind (classfile.RefInvokeStatic and friends) — 0 when
	// ImplOwner is "".
	RefKind byte
	// Captured holds every free-variable value this lambda/method
	// reference closed over at creation, in argument order — DATA_FLOW's
	// own domain: two creation sites of the same lambda shape may still
	// carry different captured data, so this lives per-Value, not
	// per-shape.
	Captured []Value
}

// Value is one fact's abstract value in DATA_FLOW's domain. Exactly one
// of Const/Type/AllocSite/Candidates/Reason is meaningful, selected by
// Kind — see each field's own comment for which.
type Value struct {
	Kind Kind

	// Const is set only when Kind == KindConst.
	Const any

	// Type is set when Kind == KindConst (the constant's own runtime
	// type, e.g. "java.lang.String"), KindTyped (the upper bound with
	// no known constant), or KindAlloc (the allocated class's own FQCN,
	// e.g. "com.example.MyRunnable" for `new MyRunnable()` — the static
	// type `new`/`anewarray`/`multianewarray` itself names, not a claim
	// about the object's contents; ASYNC_CALLBACKS A1's own first
	// consumer, to tell which concrete Runnable/Callable a registration
	// call's argument actually is).
	Type string

	// AllocSite is set only when Kind == KindAlloc.
	AllocSite *bytecodeir.Site

	// Candidates holds every alternative Value when Kind == KindSet.
	// Always non-empty and within Budget when this Kind is used — a
	// Merge call that would exceed the budget returns KindUnknown
	// instead of a Candidates slice long enough to misrepresent as
	// exhaustive.
	Candidates []Value

	// Reason explains a KindUnknown value. Required (non-empty) whenever
	// Kind == KindUnknown.
	Reason string

	// Func is set only when Kind == KindFunctional.
	Func *FunctionalValue
}

// Const returns an exact known constant of the given runtime type.
func Const(typ string, v any) Value { return Value{Kind: KindConst, Type: typ, Const: v} }

// Typed returns an upper-bound-only value: known type, unknown constant.
func Typed(typ string) Value { return Value{Kind: KindTyped, Type: typ} }

// Alloc returns a value identified by the allocation site that produced
// it and that site's own static type (the class `new` named, or the
// element type for `anewarray`/`multianewarray`) — not a claim about the
// object's contents, just which allocation and what static type it was.
func Alloc(site bytecodeir.Site, typ string) Value {
	return Value{Kind: KindAlloc, AllocSite: &site, Type: typ}
}

// Functional returns a value identified as a lambda/method reference.
func Functional(fv FunctionalValue) Value { return Value{Kind: KindFunctional, Func: &fv} }

// Null returns the null value.
func Null() Value { return Value{Kind: KindNull} }

// Unknown returns an unknown value with a mandatory, human-readable
// reason — never construct a bare KindUnknown Value literal outside
// this package; use this constructor so Reason can never be forgotten.
func Unknown(reason string) Value {
	if reason == "" {
		reason = "no reason given"
	}
	return Value{Kind: KindUnknown, Reason: reason}
}

// Equal reports whether two Values represent the same fact — used by
// Merge to decide whether two inputs collapse to one, rather than
// widening. Two KindUnknown values are never Equal (even with the same
// Reason): each occurrence is its own fact, not a deduplicated instance.
func (v Value) Equal(other Value) bool {
	if v.Kind != other.Kind {
		return false
	}
	switch v.Kind {
	case KindConst:
		return v.Type == other.Type && v.Const == other.Const
	case KindTyped:
		return v.Type == other.Type
	case KindAlloc:
		return v.AllocSite != nil && other.AllocSite != nil && *v.AllocSite == *other.AllocSite
	case KindNull:
		return true
	case KindFunctional:
		return v.Func != nil && other.Func != nil && v.Func.Site == other.Func.Site
	default: // KindSet, KindUnknown
		return false
	}
}
