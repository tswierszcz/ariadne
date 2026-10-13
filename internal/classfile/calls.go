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
	// CallSynthetic is never produced by ExtractCalls; internal/callgraph
	// uses it to mark an edge added from a profile's synthetic_edges
	// (PLAN.md section 4) rather than found in any method's own bytecode —
	// standing in for a runtime-generated proxy (e.g. a Spring Data
	// repository interface) that has no bytecode implementation at all.
	CallSynthetic CallKind = "synthetic"
)

// Call is one call-site edge out of a method: owner/name/desc identify
// the target method exactly as the bytecode (or, for a dynamic call, the
// resolved LambdaMetafactory implementation) names it.
type Call struct {
	Kind  CallKind
	Owner string // FQCN
	Name  string
	Desc  string
}

// ExtractCalls decodes a method's bytecode and returns every call-site
// edge out of it: invokevirtual/invokespecial/invokestatic/invokeinterface
// resolved straight from the constant pool, and invokedynamic resolved
// through the class's BootstrapMethods when the bootstrap is a
// LambdaMetafactory (a plain lambda or method reference); any other
// invokedynamic (e.g. string concatenation via StringConcatFactory) is
// skipped, not errored, and counted by the caller if it wants to report
// how many dynamic call sites went unresolved.
func (c *ClassFile) ExtractCalls(code []byte) ([]Call, int, error) {
	instrs, err := decodeInstructions(code)
	if err != nil {
		return nil, 0, err
	}

	var calls []Call
	unresolvedDynamic := 0
	for _, in := range instrs {
		switch in.opcode {
		case opInvokeVirtual, opInvokeSpecial, opInvokeStatic:
			idx := u16(in.operand)
			owner, name, desc, _, err := c.pool.methodref(idx)
			if err != nil {
				return nil, 0, fmt.Errorf("pc=%d: %w", in.pc, err)
			}
			calls = append(calls, Call{Kind: kindFor(in.opcode), Owner: owner, Name: name, Desc: desc})

		case opInvokeInterface:
			idx := u16(in.operand) // operand: indexbyte1,2, count, 0
			owner, name, desc, _, err := c.pool.methodref(idx)
			if err != nil {
				return nil, 0, fmt.Errorf("pc=%d: %w", in.pc, err)
			}
			calls = append(calls, Call{Kind: CallInterface, Owner: owner, Name: name, Desc: desc})

		case opInvokeDynamic:
			idx := u16(in.operand) // operand: indexbyte1,2, 0, 0
			call, ok, err := c.resolveInvokeDynamic(idx)
			if err != nil {
				return nil, 0, fmt.Errorf("pc=%d: %w", in.pc, err)
			}
			if ok {
				calls = append(calls, call)
			} else {
				unresolvedDynamic++
			}
		}
	}
	return calls, unresolvedDynamic, nil
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
func (c *ClassFile) resolveInvokeDynamic(idx uint16) (Call, bool, error) {
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
	return Call{Kind: CallDynamic, Owner: impl.Owner, Name: impl.Name, Desc: impl.Desc}, true, nil
}
