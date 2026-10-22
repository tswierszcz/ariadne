package valueflow

import "ariadne/internal/bytecodeir"

// Facts is what DATA_FLOW D2's local-value analysis will expose to any
// consumer model — SERVICE_LOADING S1 (the Class argument to
// ServiceLoader.load), REFLECTION (Class.forName's string argument),
// ASYNC_CALLBACKS (a Runnable/Callable instance), DYNAMIC_LINKAGE (a
// MethodHandle's target) — deliberately independent of CVE/vulnmap and
// of any specific framework, per ROADMAP.MD B3's own model-registry
// design ("Zaprojektować API faktów dla modeli B3, tak aby nie zależało
// od CVE ani konkretnego frameworka").
//
// D0 defines only this shape; D1 (the CFG) and D2 (the local-value
// analysis) provide the real implementation. Until then, no production
// code should assume a Facts implementation exists — see
// facts_test.go's fakeFacts for how a future consumer's own tests can
// use this contract today without waiting for D1/D2.
type Facts interface {
	// ArgumentAt returns the Value of a call's argument at index i
	// (0-based, not counting the receiver of an instance call), as
	// known at the call site identified by at. A Site whose Op is not
	// bytecodeir.OpCall, or an out-of-range i, is a caller error — an
	// implementation may panic rather than return a disguised Unknown,
	// so a bug in the caller doesn't read as "the analysis genuinely
	// doesn't know".
	ArgumentAt(at bytecodeir.Site, i int) Value

	// ReceiverAt returns the Value of an instance call's own receiver
	// (the object the call is made on — `this` inside the invoked
	// method), as known at the call site identified by at. DYNAMIC_LINKAGE
	// L2's own first consumer: `handle.invoke(...)`/`invokeExact(...)`
	// needs the receiver — the MethodHandle itself — not an argument, to
	// connect back to the HandleValue a `Lookup.find*` call produced
	// earlier. A Site whose Op is not bytecodeir.OpCall, or one with no
	// receiver at all (a static call), is a caller error — an
	// implementation may panic rather than return a disguised Unknown,
	// same discipline as ArgumentAt.
	ReceiverAt(at bytecodeir.Site) Value
}
