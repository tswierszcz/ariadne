package classfile

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
)

// Fingerprint is a hash of a class's content that does not depend on how
// the class file was serialized (CVE:F6a). Tools that rewrite classes —
// maven-shade, any ASM ClassWriter pass — keep the code but reorder the
// constant pool, which changes every index in the bytecode and can widen
// an ldc into ldc_w (shifting branch offsets). The fingerprint covers the
// class header, fields and methods (sorted), and per method its
// instructions with every constant-pool operand replaced by the constant
// it names, ldc_w/goto_w/jsr_w folded into ldc/goto/jsr, and branch
// targets as instruction indexes. Not covered: attributes other than
// Code (annotations, signatures, debug tables), exception tables and
// switch tables (only their opcode is hashed).
func Fingerprint(cf *ClassFile) (string, error) {
	h := sha256.New()
	w := func(parts ...string) {
		for _, p := range parts {
			h.Write([]byte(p))
			h.Write([]byte{0})
		}
	}
	w("class", cf.ThisClass, cf.SuperClass, fmt.Sprint(cf.AccessFlags))
	ifs := append([]string(nil), cf.Interfaces...)
	sort.Strings(ifs)
	w(ifs...)

	fields := make([]string, len(cf.Fields))
	for i, f := range cf.Fields {
		fields[i] = fmt.Sprintf("%d %s %s", f.AccessFlags, f.Name, f.Descriptor)
	}
	sort.Strings(fields)
	w(fields...)

	type mfp struct{ key, body string }
	ms := make([]mfp, 0, len(cf.Methods))
	for _, m := range cf.Methods {
		body, err := cf.codeFingerprint(m.Code)
		if err != nil {
			return "", fmt.Errorf("%s#%s%s: %w", cf.ThisClass, m.Name, m.Descriptor, err)
		}
		ms = append(ms, mfp{fmt.Sprintf("%d %s %s", m.AccessFlags, m.Name, m.Descriptor), body})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].key < ms[j].key })
	for _, m := range ms {
		w("method", m.key, m.body)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (cf *ClassFile) codeFingerprint(code []byte) (string, error) {
	if code == nil {
		return "", nil
	}
	ins, err := decodeInstructions(code)
	if err != nil {
		return "", err
	}
	index := make(map[int]int, len(ins))
	for i, in := range ins {
		index[in.pc] = i
	}
	target := func(pc int, off int32) string {
		if i, ok := index[pc+int(off)]; ok {
			return fmt.Sprintf("@%d", i)
		}
		return "@?"
	}
	var b strings.Builder
	for _, in := range ins {
		op := in.opcode
		var arg string
		switch {
		case op == opLdc:
			op, arg = opLdc, cf.pool.describe(uint16(in.operand[0]))
		case op == 0x13 || op == 0x14: // ldc_w, ldc2_w
			if op == 0x13 {
				op = opLdc
			}
			arg = cf.pool.describe(u16(in.operand))
		case op >= opGetStatic && op <= opInvokeInterface, op == opInvokeDynamic, op == opNew,
			op == 0xbd, op == 0xc0, op == 0xc1: // anewarray, checkcast, instanceof
			arg = cf.pool.describe(u16(in.operand))
		case op == 0xc5: // multianewarray
			arg = fmt.Sprintf("%s %d", cf.pool.describe(u16(in.operand)), in.operand[2])
		case (op >= opIfeqFirst && op <= opJsr) || op == opIfnull || op == opIfnonnull:
			arg = target(in.pc, int32(int16(u16(in.operand))))
		case op == opGotoW || op == opJsrW:
			arg = target(in.pc, be32(in.operand, 0))
			if op == opGotoW {
				op = opGoto
			} else {
				op = opJsr
			}
		case op == opTableSwitch || op == opLookupSwitch:
			// switch tables carry padding that depends on the pc
		default:
			arg = hex.EncodeToString(in.operand)
		}
		fmt.Fprintf(&b, "%02x %s;", op, arg)
	}
	return b.String(), nil
}

// describe renders a constant-pool entry by what it names, never by its
// index.
func (p pool) describe(idx uint16) string {
	e, err := p.get(idx)
	if err != nil {
		return "?"
	}
	switch e.tag {
	case tagUtf8:
		return "u:" + e.str
	case tagClass:
		n, _ := p.utf8(e.nameIndex)
		return "c:" + n
	case tagString:
		s, _ := p.utf8(e.nameIndex)
		return "s:" + s
	case tagMethodType:
		s, _ := p.utf8(e.nameIndex)
		return "mt:" + s
	case tagInteger:
		return fmt.Sprintf("i:%d", int32(e.raw32))
	case tagFloat:
		return fmt.Sprintf("f:%x", e.raw32)
	case tagLong:
		return fmt.Sprintf("l:%d", int64(e.raw64))
	case tagDouble:
		return fmt.Sprintf("d:%x", math.Float64bits(math.Float64frombits(e.raw64)))
	case tagFieldref, tagMethodref, tagInterfaceMethodref:
		owner := p.describe(e.classIndex)
		nt := p.describe(e.nameAndTypeIndex)
		return fmt.Sprintf("r%d:%s.%s", e.tag, owner, nt)
	case tagNameAndType:
		n, _ := p.utf8(e.nameIndex)
		d, _ := p.utf8(e.descriptorIndex)
		return n + ":" + d
	case tagMethodHandle:
		return fmt.Sprintf("mh%d:%s", e.referenceKind, p.describe(e.referenceIndex))
	case tagDynamic, tagInvokeDynamic:
		// the bootstrap table index may be renumbered too: name and type only
		return fmt.Sprintf("dyn%d:%s", e.tag, p.describe(e.nameAndTypeIndex))
	default:
		return fmt.Sprintf("t%d", e.tag)
	}
}
