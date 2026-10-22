package classfile

import "fmt"

// Opcodes DATA_FLOW D2's local-value analysis needs that JVM_GRAPH/D1
// never had to name: every instruction with its own stack/locals effect,
// JVMS chapter 6. Grouped by the real spec mnemonic ranges; nothing here
// duplicates a constant opcodes.go already declares (ldc, invoke*,
// getstatic/putstatic, new, athrow, if*/goto/switch/return, ifnull/
// ifnonnull, goto_w/jsr_w, wide — see opcodes.go's own comment on why
// those were named first, for JVM_GRAPH/D1's narrower needs).
const (
	opNop        = 0x00
	opAconstNull = 0x01
	opIconstM1   = 0x02
	opIconst0    = 0x03
	opIconst5    = 0x08
	opLconst0    = 0x09
	opLconst1    = 0x0a
	opFconst0    = 0x0b
	opFconst2    = 0x0d
	opDconst0    = 0x0e
	opDconst1    = 0x0f
	opBipush     = 0x10
	opSipush     = 0x11
	opLdcW       = 0x13
	opLdc2W      = 0x14

	opIload  = 0x15
	opLload  = 0x16
	opFload  = 0x17
	opDload  = 0x18
	opAload  = 0x19
	opIload0 = 0x1a
	opIload3 = 0x1d
	opLload0 = 0x1e
	opLload3 = 0x21
	opFload0 = 0x22
	opFload3 = 0x25
	opDload0 = 0x26
	opDload3 = 0x29
	opAload0 = 0x2a
	opAload3 = 0x2d

	opIaload = 0x2e
	opLaload = 0x2f
	opFaload = 0x30
	opDaload = 0x31
	opAaload = 0x32
	opBaload = 0x33
	opCaload = 0x34
	opSaload = 0x35

	opIstore  = 0x36
	opLstore  = 0x37
	opFstore  = 0x38
	opDstore  = 0x39
	opAstore  = 0x3a
	opIstore0 = 0x3b
	opIstore3 = 0x3e
	opLstore0 = 0x3f
	opLstore3 = 0x42
	opFstore0 = 0x43
	opFstore3 = 0x46
	opDstore0 = 0x47
	opDstore3 = 0x4a
	opAstore0 = 0x4b
	opAstore3 = 0x4e

	opIastore = 0x4f
	opLastore = 0x50
	opFastore = 0x51
	opDastore = 0x52
	opAastore = 0x53
	opBastore = 0x54
	opCastore = 0x55
	opSastore = 0x56

	opPop    = 0x57
	opPop2   = 0x58
	opDup    = 0x59
	opDupX1  = 0x5a
	opDupX2  = 0x5b
	opDup2   = 0x5c
	opDup2X1 = 0x5d
	opDup2X2 = 0x5e
	opSwap   = 0x5f

	opIadd = 0x60
	opDrem = 0x73 // iadd..drem (0x60-0x73): binary arithmetic, pop2 push1
	opIneg = 0x74
	opDneg = 0x77 // ineg..dneg (0x74-0x77): unary arithmetic, pop1 push1
	opIshl = 0x78
	opLxor = 0x83 // ishl..lxor (0x78-0x83): binary bitwise/shift, pop2 push1

	opI2l = 0x85
	opD2f = 0x90 // i2l..d2f (0x85-0x90): widening/narrowing conversions, pop1 push1
	opI2b = 0x91
	opI2s = 0x93 // i2b..i2s (0x91-0x93): narrowing int conversions, pop1 push1

	opLcmp  = 0x94
	opDcmpg = 0x98 // lcmp..dcmpg (0x94-0x98): comparisons, pop2 push1(int)

	opGetField      = 0xb4
	opPutField      = 0xb5
	opNewArray      = 0xbc
	opAnewArray     = 0xbd
	opArrayLength   = 0xbe
	opCheckCast     = 0xc0
	opInstanceOf    = 0xc1
	opMonitorEnter  = 0xc2
	opMonitorExit   = 0xc3
	opMultiNewArray = 0xc5
)

// LocalOpKind is the stack/locals effect DATA_FLOW D2's abstract
// interpreter must apply for one decoded instruction — coarser than the
// real opcode (many opcodes share one effect shape) but precise enough
// that the interpreter never needs to re-decode raw bytes itself.
type LocalOpKind string

const (
	LocalOpNop           LocalOpKind = "nop"           // no stack/locals effect at all
	LocalOpConst         LocalOpKind = "const"         // push 1 (Width tells 1 or 2 words)
	LocalOpLoad          LocalOpKind = "load"          // push 1: a copy of local Slot
	LocalOpStore         LocalOpKind = "store"         // pop 1, write into local Slot
	LocalOpIinc          LocalOpKind = "iinc"          // no stack effect; widens local Slot
	LocalOpArrayLoad     LocalOpKind = "arrayload"     // pop 2 (arrayref, index), push 1
	LocalOpArrayStore    LocalOpKind = "arraystore"    // pop 3 (arrayref, index, value)
	LocalOpFieldGet      LocalOpKind = "fieldget"      // pop (IsStatic?0:1), push 1
	LocalOpFieldPut      LocalOpKind = "fieldput"      // pop (IsStatic?1:2)
	LocalOpNew           LocalOpKind = "new"           // push 1 (Alloc)
	LocalOpNewArray      LocalOpKind = "newarray"      // pop 1 (count), push 1 (Alloc)
	LocalOpMultiNewArray LocalOpKind = "multinewarray" // pop Dims (counts), push 1 (Alloc)
	LocalOpArrayLength   LocalOpKind = "arraylength"   // pop 1, push 1
	LocalOpCheckCast     LocalOpKind = "checkcast"     // pop 1, push 1 (same value, narrowed type)
	LocalOpInstanceOf    LocalOpKind = "instanceof"    // pop 1, push 1
	LocalOpCall          LocalOpKind = "call"          // pop args(+receiver), push 0/1
	LocalOpDup           LocalOpKind = "dup"
	LocalOpDupX1         LocalOpKind = "dup_x1"
	LocalOpDupX2         LocalOpKind = "dup_x2"
	LocalOpDup2          LocalOpKind = "dup2"
	LocalOpDup2X1        LocalOpKind = "dup2_x1"
	LocalOpDup2X2        LocalOpKind = "dup2_x2"
	LocalOpSwap          LocalOpKind = "swap"
	LocalOpPop           LocalOpKind = "pop"
	LocalOpPop2          LocalOpKind = "pop2"
	LocalOpUnaryNumeric  LocalOpKind = "unary"   // pop 1, push 1, result type unknown-but-typed
	LocalOpBinaryNumeric LocalOpKind = "binary"  // pop 2, push 1, result type unknown-but-typed
	LocalOpCompare       LocalOpKind = "compare" // pop 2, push 1 (int)
	LocalOpConvert       LocalOpKind = "convert" // pop 1, push 1
	LocalOpMonitor       LocalOpKind = "monitor" // pop 1
	LocalOpThrow         LocalOpKind = "throw"   // pop 1 (terminal — see CFG)
	LocalOpReturn        LocalOpKind = "return"  // pop 0/1 (terminal — see CFG)
	LocalOpBranch        LocalOpKind = "branch"  // pop 0/1/2 (control flow handled by CFG)
	// LocalOpUnsupported mirrors CFG.Unsupported's own jsr/ret handling:
	// the interpreter must not fabricate a stack effect for it.
	LocalOpUnsupported LocalOpKind = "unsupported"
)

// ConstKind narrows LocalOpConst to exactly the two literal kinds DATA_FLOW
// D2 is in scope for ("stałe napisowe/klasowe" — string/class constants);
// every other constant-producing instruction (int/float/long/double
// literals, MethodType/MethodHandle/Dynamic constants) still gets a
// LocalOpConst so the stack stays balanced, just with ConstKind == "".
type ConstKind string

const (
	ConstKindString ConstKind = "string"
	ConstKindClass  ConstKind = "class"
	ConstKindNull   ConstKind = "null"
)

// LocalOp is one instruction, decoded enough for DATA_FLOW D2's abstract
// interpreter to apply its stack/locals effect without itself touching
// the constant pool (classfile.pool stays unexported, same reasoning as
// calls.go's own ExtractCalls).
type LocalOp struct {
	PC   int
	Kind LocalOpKind

	// Width is 1 or 2 "words" for every push-producing Kind (see each
	// LocalOpKind's own comment) — dup/dup2/pop/pop2/swap instead consult
	// each existing stack entry's own recorded width, not this field (see
	// internal/dataflow). For LocalOpReturn it is the popped return
	// value's width (0 for void). For LocalOpBranch it is the number of
	// values popped (0 for goto/goto_w, 2 for if_icmp*/if_acmp*, 1
	// otherwise) — branches push nothing, so Width can't mean push-width
	// there; it is repurposed rather than adding a rarely-used field.
	Width int

	Slot int // LocalOpLoad/Store/Iinc

	ConstKind ConstKind // LocalOpConst only; "" if untracked (see ConstKind's own doc)
	ConstStr  string    // the string value, or a Class literal's FQCN

	FieldOwner, FieldName, FieldDesc string // LocalOpFieldGet/FieldPut
	IsStatic                         bool   // LocalOpFieldGet/FieldPut: getstatic/putstatic

	// TypeName is the resolved class (LocalOpNew/CheckCast/InstanceOf/
	// MultiNewArray) or raw element descriptor (LocalOpNewArray, e.g. "I"
	// for `newarray int`, or a resolved FQCN for `anewarray`).
	TypeName string
	Dims     int // LocalOpMultiNewArray only

	CallOwner, CallName, CallDesc string
	CallKind                      CallKind // LocalOpCall only
	CallHasReceiver               bool     // true for virtual/special/interface

	// CallRefKind is the resolved implementation's MethodHandle
	// reference kind (JVMS 5.4.3.5, e.g. RefInvokeStatic,
	// RefNewInvokeSpecial) — DYNAMIC_LINKAGE L0's own "rodzaj
	// REF_invokeVirtual/REF_invokeStatic/REF_newInvokeSpecial jest
	// dostępny dla modelu". Set only for LocalOpCall with
	// CallKind == CallDynamic and a resolved LambdaMetafactory impl; 0
	// otherwise (0 is never a valid JVMS reference kind, so it
	// unambiguously means "not applicable/not resolved").
	CallRefKind byte
	// SamInterface/SamMethod are the functional interface this
	// invokedynamic's result implements and the single abstract method
	// name being created for (e.g. "java.lang.Runnable"/"run") — read
	// directly off the invokedynamic's own NameAndType, independent of
	// whether the bootstrap itself could be resolved (DYNAMIC_LINKAGE
	// L0: "zachować dynamiczne stałe... bez wykonywania bootstrapu").
	// "" when the opcode isn't invokedynamic.
	SamInterface, SamMethod string
	// CallImplDesc is the resolved implementation method's own
	// descriptor — distinct from CallDesc, which for LocalOpCall with
	// CallKind == CallDynamic holds the invokedynamic *call site's own*
	// descriptor (captured-argument types and the functional interface
	// as return type), not the implementation's. "" when CallOwner is
	// "" (unresolved) or CallKind != CallDynamic.
	CallImplDesc string
}

// DecodeLocalOps decodes code into one LocalOp per instruction, resolving
// every constant-pool reference a push/pop effect needs. It never panics;
// a malformed operand or an out-of-range constant-pool index is returned
// as an error, same contract as ExtractCalls/BuildCFG.
func (c *ClassFile) DecodeLocalOps(code []byte) ([]LocalOp, error) {
	instrs, err := decodeInstructions(code)
	if err != nil {
		return nil, err
	}
	ops := make([]LocalOp, 0, len(instrs))
	for _, in := range instrs {
		op, err := c.decodeOneLocalOp(in)
		if err != nil {
			return nil, fmt.Errorf("pc=%d: %w", in.pc, err)
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func (c *ClassFile) decodeOneLocalOp(in instruction) (LocalOp, error) {
	pc := in.pc
	switch in.opcode {
	case opNop, opMonitorEnter, opMonitorExit:
		kind := LocalOpNop
		if in.opcode != opNop {
			kind = LocalOpMonitor
		}
		return LocalOp{PC: pc, Kind: kind}, nil

	case opAconstNull:
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: 1, ConstKind: ConstKindNull}, nil

	case opLdc:
		return c.decodeLdc(pc, uint16(in.operand[0]), 1)
	case opLdcW:
		return c.decodeLdc(pc, u16(in.operand), 1)
	case opLdc2W:
		idx := u16(in.operand)
		if _, err := c.pool.get(idx); err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: 2}, nil

	case opBipush, opSipush:
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: 1}, nil

	default:
		if op := localOpByRange(in); op != nil {
			op.PC = pc
			return *op, nil
		}
	}

	switch {
	case in.opcode >= opIconstM1 && in.opcode <= opIconst5:
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: 1}, nil
	case in.opcode == opLconst0 || in.opcode == opLconst1:
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: 2}, nil
	case in.opcode >= opFconst0 && in.opcode <= opFconst2:
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: 1}, nil
	case in.opcode == opDconst0 || in.opcode == opDconst1:
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: 2}, nil

	case in.opcode == opIload || in.opcode == opFload || in.opcode == opAload:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 1, Slot: int(in.operand[0])}, nil
	case in.opcode == opLload || in.opcode == opDload:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 2, Slot: int(in.operand[0])}, nil
	case in.opcode >= opIload0 && in.opcode <= opIload3:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 1, Slot: int(in.opcode - opIload0)}, nil
	case in.opcode >= opLload0 && in.opcode <= opLload3:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 2, Slot: int(in.opcode - opLload0)}, nil
	case in.opcode >= opFload0 && in.opcode <= opFload3:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 1, Slot: int(in.opcode - opFload0)}, nil
	case in.opcode >= opDload0 && in.opcode <= opDload3:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 2, Slot: int(in.opcode - opDload0)}, nil
	case in.opcode >= opAload0 && in.opcode <= opAload3:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 1, Slot: int(in.opcode - opAload0)}, nil

	case in.opcode == opIstore || in.opcode == opFstore || in.opcode == opAstore:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: int(in.operand[0])}, nil
	case in.opcode == opLstore || in.opcode == opDstore:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: int(in.operand[0])}, nil
	case in.opcode >= opIstore0 && in.opcode <= opIstore3:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: int(in.opcode - opIstore0)}, nil
	case in.opcode >= opLstore0 && in.opcode <= opLstore3:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: int(in.opcode - opLstore0)}, nil
	case in.opcode >= opFstore0 && in.opcode <= opFstore3:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: int(in.opcode - opFstore0)}, nil
	case in.opcode >= opDstore0 && in.opcode <= opDstore3:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: int(in.opcode - opDstore0)}, nil
	case in.opcode >= opAstore0 && in.opcode <= opAstore3:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: int(in.opcode - opAstore0)}, nil
	case in.opcode == opWideIinc: // 0x84 is both plain iinc and the widened-form marker
		return LocalOp{PC: pc, Kind: LocalOpIinc, Slot: int(in.operand[0])}, nil

	case in.opcode == opLaload || in.opcode == opDaload:
		return LocalOp{PC: pc, Kind: LocalOpArrayLoad, Width: 2}, nil
	case in.opcode >= opIaload && in.opcode <= opSaload:
		return LocalOp{PC: pc, Kind: LocalOpArrayLoad, Width: 1}, nil
	case in.opcode >= opIastore && in.opcode <= opSastore:
		return LocalOp{PC: pc, Kind: LocalOpArrayStore}, nil

	case in.opcode == opGetStatic || in.opcode == opGetField:
		idx := u16(in.operand)
		owner, name, desc, err := c.pool.fieldref(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpFieldGet, Width: fieldWidth(desc),
			FieldOwner: owner, FieldName: name, FieldDesc: desc, IsStatic: in.opcode == opGetStatic}, nil
	case in.opcode == opPutStatic || in.opcode == opPutField:
		idx := u16(in.operand)
		owner, name, desc, err := c.pool.fieldref(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpFieldPut,
			FieldOwner: owner, FieldName: name, FieldDesc: desc, IsStatic: in.opcode == opPutStatic}, nil

	case in.opcode == opNew:
		idx := u16(in.operand)
		owner, err := c.pool.className(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpNew, Width: 1, TypeName: owner}, nil
	case in.opcode == opNewArray:
		return LocalOp{PC: pc, Kind: LocalOpNewArray, Width: 1, TypeName: primitiveArrayType(in.operand[0])}, nil
	case in.opcode == opAnewArray:
		idx := u16(in.operand)
		owner, err := c.pool.className(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpNewArray, Width: 1, TypeName: owner}, nil
	case in.opcode == opMultiNewArray:
		idx := u16(in.operand)
		owner, err := c.pool.className(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpMultiNewArray, Width: 1, TypeName: owner, Dims: int(in.operand[2])}, nil
	case in.opcode == opArrayLength:
		return LocalOp{PC: pc, Kind: LocalOpArrayLength, Width: 1}, nil

	case in.opcode == opCheckCast:
		idx := u16(in.operand)
		owner, err := c.pool.className(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpCheckCast, Width: 1, TypeName: owner}, nil
	case in.opcode == opInstanceOf:
		idx := u16(in.operand)
		owner, err := c.pool.className(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpInstanceOf, Width: 1, TypeName: owner}, nil

	case in.opcode == opInvokeVirtual || in.opcode == opInvokeSpecial || in.opcode == opInvokeStatic:
		idx := u16(in.operand)
		owner, name, desc, _, err := c.pool.methodref(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpCall, Width: returnWidth(desc),
			CallOwner: owner, CallName: name, CallDesc: desc, CallKind: kindFor(in.opcode),
			CallHasReceiver: in.opcode != opInvokeStatic}, nil
	case in.opcode == opInvokeInterface:
		idx := u16(in.operand)
		owner, name, desc, _, err := c.pool.methodref(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpCall, Width: returnWidth(desc),
			CallOwner: owner, CallName: name, CallDesc: desc, CallKind: CallInterface, CallHasReceiver: true}, nil
	case in.opcode == opInvokeDynamic:
		idx := u16(in.operand)
		e, err := c.pool.get(idx)
		if err != nil {
			return LocalOp{}, err
		}
		if e.tag != tagInvokeDynamic {
			return LocalOp{}, fmt.Errorf("constant pool index %d is tag %d, want InvokeDynamic", idx, e.tag)
		}
		samMethod, desc, err := c.pool.nameAndType(e.nameAndTypeIndex)
		if err != nil {
			return LocalOp{}, err
		}
		call, ok, err := c.resolveInvokeDynamic(idx, pc)
		if err != nil {
			return LocalOp{}, err
		}
		op := LocalOp{PC: pc, Kind: LocalOpCall, Width: returnWidth(desc), CallDesc: desc, CallKind: CallDynamic,
			SamInterface: samInterfaceOf(desc), SamMethod: samMethod}
		if ok {
			op.CallOwner, op.CallName, op.CallImplDesc, op.CallRefKind = call.Owner, call.Name, call.Desc, call.RefKind
		}
		return op, nil

	case in.opcode == opAthrow:
		return LocalOp{PC: pc, Kind: LocalOpThrow}, nil
	case in.opcode >= opIreturn && in.opcode <= opReturnLast:
		w := 0
		if in.opcode != opReturnLast {
			w = 1
			if in.opcode == 0xad || in.opcode == 0xaf { // lreturn, dreturn
				w = 2
			}
		}
		return LocalOp{PC: pc, Kind: LocalOpReturn, Width: w}, nil

	case isBranchOpcode(in.opcode):
		return LocalOp{PC: pc, Kind: LocalOpBranch, Width: branchPopCount(in.opcode)}, nil

	case in.opcode == opJsr || in.opcode == opJsrW || in.opcode == opRet:
		return LocalOp{PC: pc, Kind: LocalOpUnsupported}, nil

	case in.opcode == opWide:
		return c.decodeWide(pc, in.operand)
	}

	return LocalOp{}, fmt.Errorf("unhandled opcode %#x", in.opcode)
}

// localOpByRange handles every fixed-shape numeric opcode range (pop/
// push counts only — no constant-pool resolution, no per-instruction
// data) in one table instead of repeating the same four lines per range.
func localOpByRange(in instruction) *LocalOp {
	switch {
	case in.opcode >= opIadd && in.opcode <= opDrem:
		return &LocalOp{Kind: LocalOpBinaryNumeric, Width: wordWidthOf(in.opcode)}
	case in.opcode >= opIneg && in.opcode <= opDneg:
		return &LocalOp{Kind: LocalOpUnaryNumeric, Width: wordWidthOf(in.opcode)}
	case in.opcode >= opIshl && in.opcode <= opLxor:
		return &LocalOp{Kind: LocalOpBinaryNumeric, Width: shiftWidthOf(in.opcode)}
	case in.opcode >= opI2l && in.opcode <= opD2f:
		return &LocalOp{Kind: LocalOpConvert, Width: convertResultWidth(in.opcode)}
	case in.opcode >= opI2b && in.opcode <= opI2s:
		return &LocalOp{Kind: LocalOpConvert, Width: 1}
	case in.opcode >= opLcmp && in.opcode <= opDcmpg:
		return &LocalOp{Kind: LocalOpCompare, Width: 1}
	case in.opcode == opPop:
		return &LocalOp{Kind: LocalOpPop}
	case in.opcode == opPop2:
		return &LocalOp{Kind: LocalOpPop2}
	case in.opcode == opDup:
		return &LocalOp{Kind: LocalOpDup}
	case in.opcode == opDupX1:
		return &LocalOp{Kind: LocalOpDupX1}
	case in.opcode == opDupX2:
		return &LocalOp{Kind: LocalOpDupX2}
	case in.opcode == opDup2:
		return &LocalOp{Kind: LocalOpDup2}
	case in.opcode == opDup2X1:
		return &LocalOp{Kind: LocalOpDup2X1}
	case in.opcode == opDup2X2:
		return &LocalOp{Kind: LocalOpDup2X2}
	case in.opcode == opSwap:
		return &LocalOp{Kind: LocalOpSwap}
	}
	return nil
}

// wordWidthOf reports a binary/unary arithmetic opcode's own operand/
// result width: each group of 4 (iX, lX, fX, dX) follows int/long/float/
// double in that fixed order (JVMS 6.5).
func wordWidthOf(op byte) int {
	switch (op - opIadd) % 4 {
	case 1, 3: // long, double
		return 2
	default:
		return 1
	}
}

func shiftWidthOf(op byte) int {
	// ishl,lshl,ishr,lshr,iushr,lushr,iand,land,ior,lor,ixor,lxor:
	// alternating int/long, 12 opcodes from opIshl.
	if (op-opIshl)%2 == 1 {
		return 2
	}
	return 1
}

// convertResultWidth is the pushed value's width for a widening/narrowing
// numeric conversion (JVMS 6.5 i2l..d2f, 0x85-0x90): i2l=0x85(->long,2)
// i2f=0x86(->float,1) i2d=0x87(->double,2) l2i=0x88(->int,1) l2f=0x89
// (->float,1) l2d=0x8a(->double,2) f2i=0x8b(->int,1) f2l=0x8c(->long,2)
// f2d=0x8d(->double,2) d2i=0x8e(->int,1) d2l=0x8f(->long,2) d2f=0x90
// (->float,1).
func convertResultWidth(op byte) int {
	switch op {
	case 0x85, 0x87, 0x8a, 0x8c, 0x8d, 0x8f:
		return 2
	default:
		return 1
	}
}

func primitiveArrayType(atype byte) string {
	switch atype {
	case 4:
		return "[Z"
	case 5:
		return "[C"
	case 6:
		return "[F"
	case 7:
		return "[D"
	case 8:
		return "[B"
	case 9:
		return "[S"
	case 10:
		return "[I"
	case 11:
		return "[J"
	default:
		return "[?"
	}
}

func fieldWidth(desc string) int {
	if desc == "J" || desc == "D" {
		return 2
	}
	return 1
}

// returnWidth is 0 for void, 2 for long/double, 1 otherwise.
func returnWidth(methodDesc string) int {
	ret, err := returnDescriptor(methodDesc)
	if err != nil || ret == "V" {
		return 0
	}
	return fieldWidth(ret)
}

func returnDescriptor(methodDesc string) (string, error) {
	i := indexByte(methodDesc, ')')
	if i < 0 || i+1 >= len(methodDesc) {
		return "", fmt.Errorf("invalid method descriptor %q: missing ')'", methodDesc)
	}
	return methodDesc[i+1:], nil
}

// samInterfaceOf returns an invokedynamic call descriptor's return
// type as an FQCN — the functional interface the created object
// implements (always an object type; a SAM factory's result is never a
// primitive or array). "" if the descriptor is malformed.
func samInterfaceOf(desc string) string {
	ret, err := returnDescriptor(desc)
	if err != nil || len(ret) < 2 || ret[0] != 'L' {
		return ""
	}
	return descriptorToFQCN(ret)
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func isBranchOpcode(op byte) bool {
	switch {
	case op >= opIfeqFirst && op <= opIfAcmpneLast, op == opIfnull, op == opIfnonnull:
		return true
	case op == opGoto || op == opGotoW:
		return true
	case op == opTableSwitch || op == opLookupSwitch:
		return true
	}
	return false
}

// branchPopCount is how many stack values a conditional branch consumes:
// if_icmp*/if_acmp* compare two; every other if*, plus tableswitch/
// lookupswitch, consume one; goto/goto_w consume none.
func branchPopCount(op byte) int {
	switch {
	case op == opGoto || op == opGotoW:
		return 0
	case op >= 0x9f && op <= 0xa6: // if_icmpeq..if_acmpne
		return 2
	default:
		return 1
	}
}

func (c *ClassFile) decodeLdc(pc int, idx uint16, width int) (LocalOp, error) {
	e, err := c.pool.get(idx)
	if err != nil {
		return LocalOp{}, err
	}
	switch e.tag {
	case tagString:
		s, err := c.pool.utf8(e.nameIndex)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: width, ConstKind: ConstKindString, ConstStr: s}, nil
	case tagClass:
		name, err := c.pool.className(idx)
		if err != nil {
			return LocalOp{}, err
		}
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: width, ConstKind: ConstKindClass, ConstStr: name}, nil
	default:
		// Integer/Float/MethodType/MethodHandle/Dynamic: out of D2's
		// declared scope — keep the stack balanced, track nothing.
		return LocalOp{PC: pc, Kind: LocalOpConst, Width: width}, nil
	}
}

// decodeWide handles a `wide` prefix (JVMS 6.5 wide): decodeInstructions
// already packed the widened opcode and its 16-bit index (plus, for
// iinc, a 16-bit constant) into in.operand, so this only needs to
// re-dispatch with the correct slot number.
func (c *ClassFile) decodeWide(pc int, operand []byte) (LocalOp, error) {
	if len(operand) < 3 {
		return LocalOp{}, fmt.Errorf("truncated wide instruction")
	}
	widened := operand[0]
	slot := int(u16(operand[1:3]))
	switch widened {
	case opIload, opFload, opAload:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 1, Slot: slot}, nil
	case opLload, opDload:
		return LocalOp{PC: pc, Kind: LocalOpLoad, Width: 2, Slot: slot}, nil
	case opIstore, opFstore, opAstore:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: slot}, nil
	case opLstore, opDstore:
		return LocalOp{PC: pc, Kind: LocalOpStore, Slot: slot}, nil
	case opWideIinc:
		return LocalOp{PC: pc, Kind: LocalOpIinc, Slot: slot}, nil
	case opRet:
		return LocalOp{PC: pc, Kind: LocalOpUnsupported, Slot: slot}, nil
	default:
		return LocalOp{}, fmt.Errorf("unknown widened opcode %#x", widened)
	}
}
