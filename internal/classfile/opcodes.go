package classfile

import "fmt"

// Opcodes this package decodes operands for. Only the few actually needed
// to build a call graph get names; the rest of the 0..0xc9 range is still
// walked correctly (see operandLen) so the walker never desyncs on an
// instruction it doesn't care about.
const (
	opLdc             = 0x12
	opInvokeVirtual   = 0xb6
	opInvokeSpecial   = 0xb7
	opInvokeStatic    = 0xb8
	opInvokeInterface = 0xb9
	opInvokeDynamic   = 0xba
	opWide            = 0xc4
	opTableSwitch     = 0xaa
	opLookupSwitch    = 0xab
	opWideIinc        = 0x84 // the opcode byte that follows a `wide` prefix for iinc
)

// operandLen is the number of operand bytes that follow each one-byte
// fixed-length opcode (JVM spec chapter 6). tableswitch, lookupswitch and
// wide have variable length and are handled separately by decode.
// Index -1 marks exactly those three plus the reserved opcodes.
var operandLen = [256]int{
	0x00: 0, 0x01: 0, 0x02: 0, 0x03: 0, 0x04: 0, 0x05: 0, 0x06: 0, 0x07: 0,
	0x08: 0, 0x09: 0, 0x0a: 0, 0x0b: 0, 0x0c: 0, 0x0d: 0, 0x0e: 0, 0x0f: 0,
	0x10: 1, 0x11: 2, 0x12: 1, 0x13: 2, 0x14: 2, 0x15: 1, 0x16: 1, 0x17: 1,
	0x18: 1, 0x19: 1, 0x1a: 0, 0x1b: 0, 0x1c: 0, 0x1d: 0, 0x1e: 0, 0x1f: 0,
	0x20: 0, 0x21: 0, 0x22: 0, 0x23: 0, 0x24: 0, 0x25: 0, 0x26: 0, 0x27: 0,
	0x28: 0, 0x29: 0, 0x2a: 0, 0x2b: 0, 0x2c: 0, 0x2d: 0, 0x2e: 0, 0x2f: 0,
	0x30: 0, 0x31: 0, 0x32: 0, 0x33: 0, 0x34: 0, 0x35: 0, 0x36: 1, 0x37: 1,
	0x38: 1, 0x39: 1, 0x3a: 1, 0x3b: 0, 0x3c: 0, 0x3d: 0, 0x3e: 0, 0x3f: 0,
	0x40: 0, 0x41: 0, 0x42: 0, 0x43: 0, 0x44: 0, 0x45: 0, 0x46: 0, 0x47: 0,
	0x48: 0, 0x49: 0, 0x4a: 0, 0x4b: 0, 0x4c: 0, 0x4d: 0, 0x4e: 0, 0x4f: 0,
	0x50: 0, 0x51: 0, 0x52: 0, 0x53: 0, 0x54: 0, 0x55: 0, 0x56: 0, 0x57: 0,
	0x58: 0, 0x59: 0, 0x5a: 0, 0x5b: 0, 0x5c: 0, 0x5d: 0, 0x5e: 0, 0x5f: 0,
	0x60: 0, 0x61: 0, 0x62: 0, 0x63: 0, 0x64: 0, 0x65: 0, 0x66: 0, 0x67: 0,
	0x68: 0, 0x69: 0, 0x6a: 0, 0x6b: 0, 0x6c: 0, 0x6d: 0, 0x6e: 0, 0x6f: 0,
	0x70: 0, 0x71: 0, 0x72: 0, 0x73: 0, 0x74: 0, 0x75: 0, 0x76: 0, 0x77: 0,
	0x78: 0, 0x79: 0, 0x7a: 0, 0x7b: 0, 0x7c: 0, 0x7d: 0, 0x7e: 0, 0x7f: 0,
	0x80: 0, 0x81: 0, 0x82: 0, 0x83: 0, 0x84: 2, 0x85: 0, 0x86: 0, 0x87: 0,
	0x88: 0, 0x89: 0, 0x8a: 0, 0x8b: 0, 0x8c: 0, 0x8d: 0, 0x8e: 0, 0x8f: 0,
	0x90: 0, 0x91: 0, 0x92: 0, 0x93: 0, 0x94: 0, 0x95: 0, 0x96: 0, 0x97: 0,
	0x98: 0, 0x99: 2, 0x9a: 2, 0x9b: 2, 0x9c: 2, 0x9d: 2, 0x9e: 2, 0x9f: 2,
	0xa0: 2, 0xa1: 2, 0xa2: 2, 0xa3: 2, 0xa4: 2, 0xa5: 2, 0xa6: 2, 0xa7: 2,
	0xa8: 2, 0xa9: 1, 0xaa: -1, 0xab: -1, 0xac: 0, 0xad: 0, 0xae: 0, 0xaf: 0,
	0xb0: 0, 0xb1: 0, 0xb2: 2, 0xb3: 2, 0xb4: 2, 0xb5: 2, 0xb6: 2, 0xb7: 2,
	0xb8: 2, 0xb9: 4, 0xba: 4, 0xbb: 2, 0xbc: 1, 0xbd: 2, 0xbe: 0, 0xbf: 0,
	0xc0: 2, 0xc1: 2, 0xc2: 0, 0xc3: 0, 0xc4: -1, 0xc5: 3, 0xc6: 2, 0xc7: 2,
	0xc8: 4, 0xc9: 4,
	// 0xca and 0xcb..0xfd are unassigned/reserved for internal JVM use and
	// never appear in a real class file's Code attribute; 0xfe/0xff
	// (impdep1/impdep2) are likewise reserved. Leave them at the zero
	// value (0); decode treats any opcode outside 0x00-0xc9 as an error
	// instead of trusting that zero.
}

// instruction is one decoded bytecode instruction: its position, opcode,
// and operand bytes exactly as they appear in the code array (big-endian,
// unsigned; decode doesn't interpret branch offsets or switch tables,
// since call extraction never needs to follow control flow).
type instruction struct {
	pc      int
	opcode  byte
	operand []byte
}

// decodeInstructions walks a method's raw bytecode into a sequence of
// instructions. It must stay in sync on every opcode, including the
// variable-length tableswitch/lookupswitch/wide, since getting any one
// instruction's length wrong desyncs every instruction after it.
func decodeInstructions(code []byte) ([]instruction, error) {
	var out []instruction
	pc := 0
	for pc < len(code) {
		op := code[pc]
		start := pc
		pc++

		switch op {
		case opTableSwitch, opLookupSwitch:
			pc = skipSwitch(code, start, op)
		case opWide:
			if pc >= len(code) {
				return nil, fmt.Errorf("truncated wide instruction at pc=%d", start)
			}
			widened := code[pc]
			if widened == opWideIinc {
				pc += 1 + 2 + 2 // opcode + indexbyte1,2 + constbyte1,2
			} else {
				pc += 1 + 2 // opcode + indexbyte1,2
			}
		default:
			n, ok := fixedOperandLen(op)
			if !ok {
				return nil, fmt.Errorf("unknown opcode %#x at pc=%d", op, start)
			}
			pc += n
		}
		if pc > len(code) {
			return nil, fmt.Errorf("instruction at pc=%d (opcode %#x) runs past the end of the code array", start, op)
		}
		out = append(out, instruction{pc: start, opcode: op, operand: code[start+1 : pc]})
	}
	return out, nil
}

// fixedOperandLen looks up a fixed-length opcode's operand length. Every
// opcode from 0x00 to 0xc9 is assigned by the JVM spec and covered by
// operandLen; everything above that (0xca "breakpoint" and the
// unassigned/reserved range through 0xff) never appears in a real class
// file, so it is rejected rather than trusted as an implicit zero-length
// instruction.
func fixedOperandLen(op byte) (int, bool) {
	if op > 0xc9 {
		return 0, false
	}
	n := operandLen[op]
	if n < 0 {
		return 0, false // tableswitch/lookupswitch/wide: decode handles these before reaching here
	}
	return n, true
}

// skipSwitch returns the pc immediately after a tableswitch/lookupswitch
// instruction that starts at opcodeStart. Both pad with zero bytes after
// the opcode so the first 4-byte field begins at an address that's a
// multiple of 4 (measured from the start of the method's bytecode, i.e.
// opcodeStart+1 here since pc=0 is the method's first byte).
func skipSwitch(code []byte, opcodeStart int, op byte) int {
	pc := opcodeStart + 1
	pad := (4 - (pc % 4)) % 4
	pc += pad // default_byte now at a multiple of 4

	pc += 4 // defaultbyte1-4
	if op == opTableSwitch {
		low := be32(code, pc)
		pc += 4
		high := be32(code, pc)
		pc += 4
		n := int(high) - int(low) + 1
		if n < 0 {
			n = 0
		}
		pc += n * 4
	} else { // lookupswitch
		npairs := be32(code, pc)
		pc += 4
		pc += int(npairs) * 8
	}
	return pc
}

func be32(b []byte, at int) int32 {
	if at+4 > len(b) {
		return 0
	}
	return int32(uint32(b[at])<<24 | uint32(b[at+1])<<16 | uint32(b[at+2])<<8 | uint32(b[at+3]))
}

// u16 reads a big-endian unsigned 16-bit operand (indexbyte1/indexbyte2
// pairs, as used by every CP-referencing instruction this package cares
// about).
func u16(operand []byte) uint16 {
	return uint16(operand[0])<<8 | uint16(operand[1])
}
