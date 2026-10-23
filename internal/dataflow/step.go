package dataflow

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

// step applies one instruction's stack/locals effect to st, returning a
// non-empty reason only when it hits something it refuses to guess past
// (a stack underflow — which should never happen on valid bytecode but
// is checked anyway, per this codebase's own "niepoprawne wejścia nie
// powodują paniki" rule — or jsr/ret).
func (it *interp) step(st *absState, op classfile.LocalOp) string {
	switch op.Kind {
	case classfile.LocalOpNop:
		return ""

	case classfile.LocalOpConst:
		st.push(constValue(op), op.Width)
		return ""

	case classfile.LocalOpLoad:
		v, ok := st.locals[op.Slot]
		if !ok {
			v = valueflow.Unknown("read of a local variable slot outside the method's declared MaxLocals")
		}
		st.push(v, op.Width)
		return ""

	case classfile.LocalOpStore:
		e, ok := st.pop()
		if !ok {
			return "stack underflow at a *store instruction"
		}
		st.locals[op.Slot] = e.v
		return ""

	case classfile.LocalOpIinc:
		// The incremented value is deliberately not tracked — D2's own
		// declared scope is string/class constants, not arithmetic.
		st.locals[op.Slot] = valueflow.Typed("I")
		return ""

	case classfile.LocalOpArrayLoad:
		if _, ok := st.popN(2); !ok {
			return "stack underflow at an array-load instruction"
		}
		st.push(valueflow.Unknown("array element value is not tracked by this analysis increment"), op.Width)
		return ""

	case classfile.LocalOpArrayStore:
		if _, ok := st.popN(3); !ok {
			return "stack underflow at an array-store instruction"
		}
		return ""

	case classfile.LocalOpFieldGet:
		n := 0
		if !op.IsStatic {
			n = 1
		}
		if n > 0 {
			if _, ok := st.popN(n); !ok {
				return "stack underflow at a field-get instruction"
			}
		}
		// DATA_FLOW D3: identify *which* field this is (static or
		// instance), rather than discarding to a bare Typed(desc) — a
		// consumer with hierarchy access (ResolveStaticField/
		// ResolveInstanceField) can go find out what every writer
		// actually stored there.
		st.push(valueflow.Field(valueflow.FieldRef{Owner: op.FieldOwner, Name: op.FieldName, Desc: op.FieldDesc}), op.Width)
		return ""

	case classfile.LocalOpFieldPut:
		n := 1
		if !op.IsStatic {
			n = 2
		}
		entries, ok := st.popN(n)
		if !ok {
			return "stack underflow at a field-put instruction"
		}
		if op.IsStatic {
			// DATA_FLOW D3: record what was actually stored, for
			// Facts.FieldValueAt — the other half of ResolveStaticField's
			// own cross-method summary.
			it.facts.fieldWrites[it.site(op.PC, bytecodeir.OpFieldPut)] = entries[0].v
		} else if it.m.Name == "<init>" && entries[0].v.Kind == valueflow.KindTyped && entries[0].v.Type == it.fqcn {
			// DATA_FLOW D3's instance-field increment: only a provable
			// `this.field = ...` write inside a constructor is recorded
			// — entries[0] is the objectref, and slot 0 in a constructor
			// is always `this` (Typed(it.fqcn), set once at method entry
			// in analyze.go's run and never reassigned by valid
			// bytecode). A write through any other receiver (a different
			// instance, a supertype constructor's own copy of a field)
			// is left unrecorded — ResolveInstanceField's own merge
			// treats an unrecorded site as "no usable writer here",
			// never fabricating a value. See ResolveInstanceField's doc
			// comment for why this narrow shape is safe without real
			// points-to.
			it.facts.fieldWrites[it.site(op.PC, bytecodeir.OpFieldPut)] = entries[1].v
		}
		return ""

	case classfile.LocalOpNew:
		st.push(valueflow.Alloc(it.site(op.PC, bytecodeir.OpNew), op.TypeName), op.Width)
		return ""

	case classfile.LocalOpNewArray:
		if _, ok := st.pop(); !ok {
			return "stack underflow at an array-allocation instruction"
		}
		st.push(valueflow.Alloc(it.site(op.PC, bytecodeir.OpNew), op.TypeName), op.Width)
		return ""

	case classfile.LocalOpMultiNewArray:
		if _, ok := st.popN(op.Dims); !ok {
			return "stack underflow at a multianewarray instruction"
		}
		st.push(valueflow.Alloc(it.site(op.PC, bytecodeir.OpNew), op.TypeName), op.Width)
		return ""

	case classfile.LocalOpArrayLength:
		if _, ok := st.pop(); !ok {
			return "stack underflow at arraylength"
		}
		st.push(valueflow.Typed("I"), op.Width)
		return ""

	case classfile.LocalOpCheckCast:
		e, ok := st.pop()
		if !ok {
			return "stack underflow at checkcast"
		}
		v := e.v
		if v.Kind == valueflow.KindTyped {
			v = valueflow.Typed(op.TypeName)
		}
		st.push(v, op.Width)
		return ""

	case classfile.LocalOpInstanceOf:
		if _, ok := st.pop(); !ok {
			return "stack underflow at instanceof"
		}
		st.push(valueflow.Typed("Z"), op.Width)
		return ""

	case classfile.LocalOpCall:
		return it.stepCall(st, op)

	case classfile.LocalOpDup:
		e, ok := st.pop()
		if !ok || e.width != 1 {
			return "dup on a missing or category-2 top-of-stack value"
		}
		st.push(e.v, 1)
		st.push(e.v, 1)
		return ""

	case classfile.LocalOpDupX1:
		return stepDupX1(st)
	case classfile.LocalOpDupX2:
		return stepDupX2(st)
	case classfile.LocalOpDup2:
		return stepDup2(st)
	case classfile.LocalOpDup2X1:
		return stepDup2X1(st)
	case classfile.LocalOpDup2X2:
		return stepDup2X2(st)

	case classfile.LocalOpSwap:
		vs, ok := st.popN(2)
		if !ok || vs[0].width != 1 || vs[1].width != 1 {
			return "swap on fewer than two category-1 values"
		}
		st.pushEntry(vs[1])
		st.pushEntry(vs[0])
		return ""

	case classfile.LocalOpPop:
		e, ok := st.pop()
		if !ok || e.width != 1 {
			return "pop on a missing or category-2 top-of-stack value"
		}
		return ""

	case classfile.LocalOpPop2:
		e, ok := st.pop()
		if !ok {
			return "stack underflow at pop2"
		}
		if e.width == 2 {
			return ""
		}
		if _, ok := st.pop(); !ok {
			return "pop2 on a single category-1 value with nothing beneath it"
		}
		return ""

	case classfile.LocalOpUnaryNumeric:
		if _, ok := st.pop(); !ok {
			return "stack underflow at a unary numeric op"
		}
		st.push(valueflow.Unknown("arithmetic result is not tracked by this analysis increment"), op.Width)
		return ""

	case classfile.LocalOpBinaryNumeric:
		if _, ok := st.popN(2); !ok {
			return "stack underflow at a binary numeric op"
		}
		st.push(valueflow.Unknown("arithmetic result is not tracked by this analysis increment"), op.Width)
		return ""

	case classfile.LocalOpCompare:
		if _, ok := st.popN(2); !ok {
			return "stack underflow at a comparison op"
		}
		st.push(valueflow.Typed("I"), op.Width)
		return ""

	case classfile.LocalOpConvert:
		if _, ok := st.pop(); !ok {
			return "stack underflow at a conversion op"
		}
		st.push(valueflow.Unknown("converted numeric value is not tracked by this analysis increment"), op.Width)
		return ""

	case classfile.LocalOpMonitor:
		if _, ok := st.pop(); !ok {
			return "stack underflow at monitorenter/monitorexit"
		}
		return ""

	case classfile.LocalOpThrow:
		if _, ok := st.pop(); !ok {
			return "stack underflow at athrow"
		}
		return "" // terminal: D1's CFG gives athrow no normal successor

	case classfile.LocalOpReturn:
		if op.Width > 0 {
			e, ok := st.pop()
			if !ok {
				return "stack underflow at a return instruction"
			}
			// DATA_FLOW D3: record what this return site actually
			// returns, for ResolveCallResult's own cross-method
			// summary — the return-value half of the same
			// identify-here/resolve-there split KindField already
			// uses for fields.
			it.facts.returns[it.site(op.PC, bytecodeir.OpReturn)] = e.v
		}
		return "" // terminal

	case classfile.LocalOpBranch:
		if _, ok := st.popN(op.Width); !ok {
			return "stack underflow at a branch instruction"
		}
		return ""

	case classfile.LocalOpUnsupported:
		return "jsr/ret subroutine: not modeled by this analysis (same limitation as DATA_FLOW D1's CFG)"

	default:
		return fmt.Sprintf("unhandled LocalOpKind %q", op.Kind)
	}
}

func constValue(op classfile.LocalOp) valueflow.Value {
	switch op.ConstKind {
	case classfile.ConstKindString:
		return valueflow.Const("java.lang.String", op.ConstStr)
	case classfile.ConstKindClass:
		return valueflow.Const("java.lang.Class", op.ConstStr)
	case classfile.ConstKindNull:
		return valueflow.Null()
	default:
		return valueflow.Unknown("constant kind outside DATA_FLOW D2's declared scope (not a string/class literal)")
	}
}

func (it *interp) stepCall(st *absState, op classfile.LocalOp) string {
	n, err := callArgCount(op.CallDesc)
	if err != nil {
		return fmt.Sprintf("invalid call descriptor %q: %v", op.CallDesc, err)
	}
	popN := n
	if op.CallHasReceiver {
		popN++
	}
	entries, ok := st.popN(popN)
	if !ok {
		return "stack underflow at a call instruction"
	}
	argEntries := entries
	if op.CallHasReceiver {
		argEntries = entries[1:]
	}
	args := make([]valueflow.Value, len(argEntries))
	for i, e := range argEntries {
		args[i] = e.v
	}
	it.facts.calls[it.site(op.PC, bytecodeir.OpCall)] = args
	if op.CallHasReceiver {
		it.facts.receivers[it.site(op.PC, bytecodeir.OpCall)] = entries[0].v
	}

	if op.CallKind == classfile.CallDynamic && op.CallOwner != "" {
		// DYNAMIC_LINKAGE L0/L1: a resolved LambdaMetafactory creation
		// site produces a real FunctionalValue, not a generic Unknown
		// — args here are exactly the captured free variables (indy
		// never has a receiver, so popN == len(args) above already).
		st.push(valueflow.Functional(valueflow.FunctionalValue{
			Site:         it.site(op.PC, bytecodeir.OpCall),
			SamInterface: op.SamInterface,
			SamMethod:    op.SamMethod,
			ImplOwner:    op.CallOwner,
			ImplName:     op.CallName,
			ImplDesc:     op.CallImplDesc,
			RefKind:      op.CallRefKind,
			Captured:     args,
		}), op.Width)
		return ""
	}
	if hv, ok := lookupHandleValue(it.site(op.PC, bytecodeir.OpCall), op, args); ok {
		// DYNAMIC_LINKAGE L2: a MethodHandles.Lookup find* call whose
		// Class/name arguments are provable literals produces a real
		// HandleValue, not a generic Unknown.
		st.push(valueflow.Handle(hv), op.Width)
		return ""
	}
	if cv, ok := classForNameValue(op, args); ok {
		// REFLECTION R1: a Class.forName call whose class-name argument
		// is a provable literal resolves straight to the same
		// KindConst/"java.lang.Class" shape a `Foo.class` literal
		// already produces, not a generic Unknown.
		st.push(cv, op.Width)
		return ""
	}
	if op.Width > 0 {
		if op.CallKind == classfile.CallStatic || op.CallKind == classfile.CallSpecial {
			// DATA_FLOW D3's second interprocedural summary: an
			// invokestatic/invokespecial target is already exactly
			// resolved (no CHA ambiguity, unlike a virtual/interface
			// call, which could dispatch to any override at runtime) —
			// a consumer with hierarchy access
			// (internal/dataflow.ResolveCallResult) can go analyze that
			// one method's own return value(s), the same
			// identify-here/resolve-there split KindField already uses.
			st.push(valueflow.CallResult(valueflow.CallResultRef{Owner: op.CallOwner, Name: op.CallName, Desc: op.CallDesc}), op.Width)
			return ""
		}
		reason := "call return value is not tracked by this analysis increment"
		switch {
		case op.CallKind == classfile.CallDynamic:
			reason = fmt.Sprintf("invokedynamic bootstrap for %s#%s could not be resolved to a LambdaMetafactory implementation", op.SamInterface, op.SamMethod)
		case op.CallOwner == lookupOwner:
			reason = fmt.Sprintf("MethodHandles.Lookup#%s target is not statically known (Class/name argument not a literal)", op.CallName)
		case op.CallOwner == classOwner && op.CallName == "forName":
			reason = "Class.forName target is not statically known (class-name argument not a literal)"
		}
		st.push(valueflow.Unknown(reason), op.Width)
	}
	return ""
}
