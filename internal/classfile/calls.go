package classfile

import "fmt"

// CallKind says how a call was made, which matters for CHA: virtual and
// interface calls get expanded to every overriding subtype, special and
// static calls are already fully resolved, and dynamic calls (lambdas,
// method references) resolve to exactly one target method via
// LambdaMetafactory.
type CallKind string

const (
	CallVirtual   CallKind = "virtual"
	CallInterface CallKind = "interface"
	CallSpecial   CallKind = "special"
	CallStatic    CallKind = "static"
	CallDynamic   CallKind = "dynamic"
	// CallInit is never produced by ExtractCalls; internal/callgraph uses
	// it for a class-initialization edge (JVM_GRAPH G3, callgraph.BasisInit)
	// — there is no single invoke* instruction for the JVM implicitly
	// running a <clinit>, so no real CallKind describes it either.
	CallInit CallKind = "init"
	// CallSynthetic is never produced by ExtractCalls; internal/callgraph
	// uses it to mark an edge added from a profile's synthetic_edges
	// (PLAN.md section 4) rather than found in any method's own bytecode —
	// standing in for a runtime-generated proxy (e.g. a Spring Data
	// repository interface) that has no bytecode implementation at all.
	CallSynthetic CallKind = "synthetic"
)

// Call is one call-site edge out of a method: owner/name/desc identify
// the target method exactly as the bytecode (or, for a dynamic call, the
// resolved LambdaMetafactory implementation) names it. PC is the byte
// offset of the invoke* instruction within the method's Code array —
// evidence for *which* call site produced this edge, kept so
// internal/callgraph/internal/reach can point a report at a specific
// line of bytecode instead of just naming the target method (RECOMMENDATION
// ROADMAP.MD B1/JVM_GRAPH G0: "miejsce wywołania: stabilna metoda + offset
// bytecode + rodzaj operacji"). It is 0 for a call synthesized outside any
// real bytecode (a profile's synthetic_edges, CallSynthetic).
type Call struct {
	Kind  CallKind
	Owner string // FQCN
	Name  string
	Desc  string
	PC    int
	// RefKind is the resolved implementation's MethodHandle reference
	// kind (JVMS 5.4.3.5) for a CallDynamic edge whose LambdaMetafactory
	// bootstrap was resolved — DYNAMIC_LINKAGE L0. 0 (never a valid
	// JVMS reference kind) for every other Kind, or an unresolved
	// dynamic call.
	RefKind byte
	// SamInterface/SamMethod are, for a resolved CallDynamic
	// (LambdaMetafactory), the functional interface the created object
	// implements and its abstract method's name — read from the
	// invokedynamic NameAndType (DYNAMIC_LINKAGE L1, SAM-call
	// materialization finds the interface calls that may run Owner#Name
	// from them). "" for every other call.
	SamInterface, SamMethod string
}

// InitKind says which instruction produced an InitRef — the three JVMS
// 5.5 triggers ariadne can see in a method body without a full
// data-flow/value analysis (invokestatic is also a Call and handled via
// that, not duplicated here — see InitRef's own doc comment).
type InitKind string

const (
	InitNew       InitKind = "new"
	InitGetStatic InitKind = "getstatic"
	InitPutStatic InitKind = "putstatic"
)

// InitRef is one reference that, per the JVM's class-initialization
// rules (JVMS 5.5), requires some class to be initialized (its <clinit>
// run) before this instruction executes — the fourth kind of
// JVM_GRAPH-G0-style edge basis, "initializes", next to "calls":
//   - InitNew: `new` always initializes exactly the class it names —
//     Owner is that class, Name/Desc are empty.
//   - InitGetStatic/InitPutStatic: a `getstatic`/`putstatic` initializes
//     whichever class actually *declares* the named field (Owner is the
//     bytecode's literal Fieldref owner, which — exactly like a
//     Methodref's owner before JVM_GRAPH G2 — is not necessarily the
//     declaring class if the field is inherited; the caller resolves
//     that, the same way internal/callgraph already resolves a call's
//     declaring class).
//
// A compile-time-constant static field (a `static final` primitive/
// String) never produces a getstatic/putstatic in the first place — the
// JLS requires the compiler to inline its value (typically as `ldc`)
// instead — so every InitRef this type actually represents is already,
// by construction of how javac/kotlinc emit bytecode, a genuine
// initialization trigger. ariadne does not need to separately detect
// field constant-ness (G3's own "samo posiadanie literału klasy nie
// staje się regułą inicjalizacji" is about `Foo.class`/ldc, which this
// type never produces at all, not about constant static fields).
//
// invokestatic is deliberately not a separate InitKind: it is already a
// Call (Kind: CallStatic), and internal/callgraph already resolves its
// declaring class via hierarchy.ResolveExact for the ordinary call edge
// (JVM_GRAPH G2) — reusing that result for the init edge avoids
// resolving the same reference twice.
type InitRef struct {
	Kind  InitKind
	Owner string // FQCN
	Name  string // "" for InitNew
	Desc  string // "" for InitNew
	PC    int
}

// ExtractCalls decodes a method's bytecode and returns every call-site
// edge out of it: invokevirtual/invokespecial/invokestatic/invokeinterface
// resolved straight from the constant pool, and invokedynamic resolved
// through the class's BootstrapMethods when the bootstrap is a
// LambdaMetafactory (a plain lambda or method reference); any other
// invokedynamic (e.g. string concatenation via StringConcatFactory) is
// skipped, not errored, and counted by the caller if it wants to report
// how many dynamic call sites went unresolved. It also returns every
// InitRef found in the same pass (JVM_GRAPH G3) — one walk over the
// decoded instructions serves both, since new/getstatic/putstatic sit
// right alongside the invoke* opcodes in the same switch.
func (c *ClassFile) ExtractCalls(code []byte) ([]Call, []InitRef, int, error) {
	instrs, err := decodeInstructions(code)
	if err != nil {
		return nil, nil, 0, err
	}

	var calls []Call
	var initRefs []InitRef
	unresolvedDynamic := 0
	for _, in := range instrs {
		switch in.opcode {
		case opInvokeVirtual, opInvokeSpecial, opInvokeStatic:
			idx := u16(in.operand)
			owner, name, desc, _, err := c.pool.methodref(idx)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("pc=%d: %w", in.pc, err)
			}
			calls = append(calls, Call{Kind: kindFor(in.opcode), Owner: owner, Name: name, Desc: desc, PC: in.pc})

		case opInvokeInterface:
			idx := u16(in.operand) // operand: indexbyte1,2, count, 0
			owner, name, desc, _, err := c.pool.methodref(idx)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("pc=%d: %w", in.pc, err)
			}
			calls = append(calls, Call{Kind: CallInterface, Owner: owner, Name: name, Desc: desc, PC: in.pc})

		case opInvokeDynamic:
			idx := u16(in.operand) // operand: indexbyte1,2, 0, 0
			call, ok, err := c.resolveInvokeDynamic(idx, in.pc)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("pc=%d: %w", in.pc, err)
			}
			if ok {
				calls = append(calls, call)
			} else {
				unresolvedDynamic++
			}

		case opNew:
			idx := u16(in.operand)
			owner, err := c.pool.className(idx)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("pc=%d: %w", in.pc, err)
			}
			initRefs = append(initRefs, InitRef{Kind: InitNew, Owner: owner, PC: in.pc})

		case opGetStatic, opPutStatic:
			idx := u16(in.operand)
			owner, name, desc, err := c.pool.fieldref(idx)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("pc=%d: %w", in.pc, err)
			}
			kind := InitGetStatic
			if in.opcode == opPutStatic {
				kind = InitPutStatic
			}
			initRefs = append(initRefs, InitRef{Kind: kind, Owner: owner, Name: name, Desc: desc, PC: in.pc})
		}
	}
	return calls, initRefs, unresolvedDynamic, nil
}

func kindFor(opcode byte) CallKind {
	switch opcode {
	case opInvokeVirtual:
		return CallVirtual
	case opInvokeSpecial:
		return CallSpecial
	case opInvokeStatic:
		return CallStatic
	default:
		return CallVirtual
	}
}

// resolveInvokeDynamic looks up the CONSTANT_InvokeDynamic entry at idx,
// finds its bootstrap method, and — only when that bootstrap is
// LambdaMetafactory — resolves to the real implementation method a call
// through the resulting lambda/method reference will dispatch to.
func (c *ClassFile) resolveInvokeDynamic(idx uint16, pc int) (Call, bool, error) {
	e, err := c.pool.get(idx)
	if err != nil {
		return Call{}, false, err
	}
	if e.tag != tagInvokeDynamic {
		return Call{}, false, fmt.Errorf("constant pool index %d is tag %d, want InvokeDynamic", idx, e.tag)
	}
	if int(e.bootstrapMethodAttrIndex) >= len(c.Bootstrap) {
		return Call{}, false, fmt.Errorf("invokedynamic bootstrap_method_attr_index %d out of range (%d bootstrap methods)", e.bootstrapMethodAttrIndex, len(c.Bootstrap))
	}
	bsm := c.Bootstrap[e.bootstrapMethodAttrIndex]
	impl, ok := bsm.ImplMethod()
	if !ok {
		return Call{}, false, nil
	}
	samMethod, desc, err := c.pool.nameAndType(e.nameAndTypeIndex)
	if err != nil {
		return Call{}, false, err
	}
	return Call{Kind: CallDynamic, Owner: impl.Owner, Name: impl.Name, Desc: impl.Desc, PC: pc, RefKind: impl.Kind,
		SamInterface: samInterfaceOf(desc), SamMethod: samMethod}, true, nil
}
