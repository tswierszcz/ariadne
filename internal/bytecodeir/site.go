// Package bytecodeir is DATA_FLOW D0's own contract: a common identity
// for one bytecode instruction, shared by D1's CFG builder and D2's
// local-value analysis (and, through them, every consumer model —
// SERVICE_LOADING S1, REFLECTION, ASYNC_CALLBACKS, DYNAMIC_LINKAGE) so
// each one doesn't invent its own notion of "which instruction, in
// which method, produced this fact".
//
// This package is the contract only (DATA_FLOW D0) — it does not decode
// real bytecode into a Site itself. That is D1's own scope
// ("internal/classfile/code.go" gains a CFG builder that produces real
// Sites); D0 exists so D1 and D2 agree on the shape before either is
// built, the same reasoning JVM_GRAPH G0 applied one level up (call
// graph edges) rather than every consumer re-deriving its own evidence
// format.
package bytecodeir

// Site identifies one bytecode instruction within one method — the
// same four pieces of information JVM_GRAPH's own Call/CallSite already
// anchor a call-graph edge to (method, descriptor, offset), plus the
// instruction's own operation kind, since a data-flow fact is about an
// instruction, not necessarily a call.
type Site struct {
	Method     string // "fqcn#name" (no descriptor — see Descriptor)
	Descriptor string
	PC         int // bytecode offset within this method's Code array
	Op         OpKind
}

// OpKind is the operation category an instruction belongs to, coarse
// enough for D0's contract purposes (D1's real CFG may record the exact
// opcode too, once it exists) but specific enough to tell a data-flow
// consumer what kind of fact to expect at a Site.
type OpKind string

const (
	OpConst    OpKind = "const"    // ldc, iconst/lconst/fconst/dconst, aconst_null
	OpLoad     OpKind = "load"     // *load (local variable)
	OpStore    OpKind = "store"    // *store (local variable)
	OpFieldGet OpKind = "fieldget" // getfield, getstatic
	OpFieldPut OpKind = "fieldput" // putfield, putstatic
	OpArrayGet OpKind = "arrayget" // *aload (array element)
	OpArrayPut OpKind = "arrayput" // *astore (array element)
	OpNew      OpKind = "new"      // new, newarray, anewarray, multianewarray
	OpCall     OpKind = "call"     // invoke*
	OpBranch   OpKind = "branch"   // if*, goto, *switch
	OpReturn   OpKind = "return"   // *return
	OpThrow    OpKind = "throw"    // athrow
	OpCast     OpKind = "cast"     // checkcast, instanceof
	// OpUnsupported marks an instruction D1's CFG builder recognizes but
	// does not model the semantics of (the plan's own example: old
	// jsr/ret) — a Site with this Op always pairs with a Value of
	// KindUnknown (internal/valueflow) carrying a Reason, never a
	// fabricated fact.
	OpUnsupported OpKind = "unsupported"
)
