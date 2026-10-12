package classfile

import (
	"bytes"
	"fmt"
)

// parseBootstrapMethods parses a class-level BootstrapMethods attribute,
// resolving each bootstrap method's own target and every MethodHandle
// among its arguments (see BootstrapArg).
func parseBootstrapMethods(data []byte, p pool) ([]BootstrapMethod, error) {
	rd := newReader(bytes.NewReader(data))
	count := rd.u2()
	methods := make([]BootstrapMethod, 0, count)
	for i := uint16(0); i < count && rd.err == nil; i++ {
		refIdx := rd.u2()
		argCount := rd.u2()
		argIdxs := make([]uint16, argCount)
		for j := range argIdxs {
			argIdxs[j] = rd.u2()
		}
		if rd.err != nil {
			break
		}

		kind, owner, name, desc, isIface, err := p.methodHandle(refIdx)
		if err != nil {
			return nil, fmt.Errorf("bootstrap method %d: %w", i, err)
		}
		args := make([]BootstrapArg, len(argIdxs))
		for j, idx := range argIdxs {
			args[j] = resolveBootstrapArg(p, idx)
		}
		methods = append(methods, BootstrapMethod{
			Kind: kind, Owner: owner, Name: name, Desc: desc, IsInterface: isIface,
			Args: args,
		})
	}
	if rd.err != nil {
		return nil, rd.err
	}
	return methods, nil
}

// resolveBootstrapArg resolves one bootstrap argument if it's a
// MethodHandle; any other constant kind (MethodType, Class, a constant
// value, ...) is left as a zero-value BootstrapArg, since ariadne only
// ever needs the implementation-method handle.
func resolveBootstrapArg(p pool, idx uint16) BootstrapArg {
	e, err := p.get(idx)
	if err != nil || e.tag != tagMethodHandle {
		return BootstrapArg{}
	}
	kind, owner, name, desc, isIface, err := p.methodHandle(idx)
	if err != nil {
		return BootstrapArg{}
	}
	return BootstrapArg{
		IsMethodHandle: true,
		Kind:           kind, Owner: owner, Name: name, Desc: desc, IsInterface: isIface,
	}
}
