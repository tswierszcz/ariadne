// Package classfile is a pure-Go parser for JVM .class files, plus a
// bytecode walker that extracts method-call edges (including lambdas and
// method references resolved through LambdaMetafactory) straight from the
// Code attribute. It needs no JDK: ariadne builds its own call graph
// instead of shelling out to a bytecode-based tool.
package classfile

import (
	"fmt"
	"math"
)

// Constant pool tags (JVM spec table 4.4-A), including every tag added up
// through Java 21 (Dynamic/InvokeDynamic/Module/Package).
const (
	tagUtf8               = 1
	tagInteger            = 3
	tagFloat              = 4
	tagLong               = 5
	tagDouble             = 6
	tagClass              = 7
	tagString             = 8
	tagFieldref           = 9
	tagMethodref          = 10
	tagInterfaceMethodref = 11
	tagNameAndType        = 12
	tagMethodHandle       = 15
	tagMethodType         = 16
	tagDynamic            = 17
	tagInvokeDynamic      = 18
	tagModule             = 19
	tagPackage            = 20
)

// MethodHandle reference kinds (JVM spec table 5.4.3.5). The complete
// table, kept here even though field-access kinds (1-4) never produce a
// LambdaMetafactory implementation method and so never matter to
// ImplMethod/resolveInvokeDynamic — DYNAMIC_LINKAGE L0's own "rodzaj...
// jest dostępny dla modelu" asks for the full set to be nameable, not
// just the ones this package's own call-graph resolution currently uses.
const (
	RefGetField         = 1
	RefGetStatic        = 2
	RefPutField         = 3
	RefPutStatic        = 4
	RefInvokeVirtual    = 5 // an unbound instance method reference, e.g. String::length
	RefInvokeStatic     = 6
	RefInvokeSpecial    = 7
	RefNewInvokeSpecial = 8 // a constructor reference, e.g. Foo::new
	RefInvokeInterface  = 9
)

// cpEntry is a constant pool entry. Not every field is meaningful for
// every tag; see the comment on each tag's usage below. A Long/Double
// entry occupies its own slot plus an unused placeholder slot right after
// it, per the JVM spec's "irregular" pool indexing.
type cpEntry struct {
	tag byte

	// tagUtf8
	str string

	// tagClass: name index (-> Utf8, slash-separated binary name)
	// tagString: string index (-> Utf8)
	// tagMethodType: descriptor index (-> Utf8)
	// tagModule, tagPackage: name index (-> Utf8)
	nameIndex uint16

	// tagFieldref/tagMethodref/tagInterfaceMethodref: owning class index (-> Class)
	classIndex uint16

	// tagFieldref/tagMethodref/tagInterfaceMethodref: name-and-type index
	// tagDynamic/tagInvokeDynamic: name-and-type index
	nameAndTypeIndex uint16

	// tagNameAndType: descriptor index (-> Utf8)
	descriptorIndex uint16

	// tagMethodHandle
	referenceKind  byte
	referenceIndex uint16

	// tagDynamic/tagInvokeDynamic: index into the class's bootstrap_methods table
	bootstrapMethodAttrIndex uint16

	// tagInteger: the raw 4 bytes, read as a two's-complement int32.
	// tagFloat: the raw 4 bytes, IEEE 754 bit pattern (math.Float32frombits).
	// FRAMEWORKS F1's own need: an annotation element_value's 'B'/'C'/'I'/
	// 'S'/'Z' const_value_index point at an Integer entry (narrowed by the
	// element_value's own tag, JVM spec 4.7.16.1), and 'F' at a Float one.
	raw32 uint32

	// tagLong/tagDouble: the raw 8 bytes (hi u4 in the upper 32 bits, lo
	// u4 in the lower 32, matching the two u4 reads readPool already did),
	// read as int64 or IEEE 754 double bits (math.Float64frombits) by an
	// element_value's 'J'/'D' tag.
	raw64 uint64
}

// pool is a class file's constant pool, 1-indexed as the JVM spec requires
// (pool[0] is unused).
type pool []cpEntry

func (p pool) get(idx uint16) (cpEntry, error) {
	if idx == 0 || int(idx) >= len(p) {
		return cpEntry{}, fmt.Errorf("constant pool index %d out of range (pool has %d entries)", idx, len(p))
	}
	return p[idx], nil
}

func (p pool) utf8(idx uint16) (string, error) {
	e, err := p.get(idx)
	if err != nil {
		return "", err
	}
	if e.tag != tagUtf8 {
		return "", fmt.Errorf("constant pool index %d is tag %d, want Utf8", idx, e.tag)
	}
	return e.str, nil
}

// className resolves a Class constant to its binary name with '/' already
// converted to '.' (so "fixture/App" becomes "fixture.App"; nested classes
// keep their '$', e.g. "fixture.App$1").
func (p pool) className(idx uint16) (string, error) {
	e, err := p.get(idx)
	if err != nil {
		return "", err
	}
	if e.tag != tagClass {
		return "", fmt.Errorf("constant pool index %d is tag %d, want Class", idx, e.tag)
	}
	name, err := p.utf8(e.nameIndex)
	if err != nil {
		return "", err
	}
	return binaryNameToFQCN(name), nil
}

func binaryNameToFQCN(binaryName string) string {
	out := make([]byte, len(binaryName))
	for i := 0; i < len(binaryName); i++ {
		if binaryName[i] == '/' {
			out[i] = '.'
		} else {
			out[i] = binaryName[i]
		}
	}
	return string(out)
}

// nameAndType resolves a NameAndType constant to (name, descriptor).
func (p pool) nameAndType(idx uint16) (name, desc string, err error) {
	e, err := p.get(idx)
	if err != nil {
		return "", "", err
	}
	if e.tag != tagNameAndType {
		return "", "", fmt.Errorf("constant pool index %d is tag %d, want NameAndType", idx, e.tag)
	}
	name, err = p.utf8(e.nameIndex)
	if err != nil {
		return "", "", err
	}
	desc, err = p.utf8(e.descriptorIndex)
	if err != nil {
		return "", "", err
	}
	return name, desc, nil
}

// methodref resolves a Methodref/InterfaceMethodref constant to its owning
// class (FQCN), name and descriptor.
func (p pool) methodref(idx uint16) (owner, name, desc string, isInterface bool, err error) {
	e, err := p.get(idx)
	if err != nil {
		return "", "", "", false, err
	}
	if e.tag != tagMethodref && e.tag != tagInterfaceMethodref {
		return "", "", "", false, fmt.Errorf("constant pool index %d is tag %d, want Methodref/InterfaceMethodref", idx, e.tag)
	}
	owner, err = p.className(e.classIndex)
	if err != nil {
		return "", "", "", false, err
	}
	name, desc, err = p.nameAndType(e.nameAndTypeIndex)
	if err != nil {
		return "", "", "", false, err
	}
	return owner, name, desc, e.tag == tagInterfaceMethodref, nil
}

// fieldref resolves a Fieldref constant to its owning class (FQCN), name
// and descriptor — JVM_GRAPH G3's new/getstatic/putstatic init-trigger
// extraction needs this the same way methodref already serves ExtractCalls.
func (p pool) fieldref(idx uint16) (owner, name, desc string, err error) {
	e, err := p.get(idx)
	if err != nil {
		return "", "", "", err
	}
	if e.tag != tagFieldref {
		return "", "", "", fmt.Errorf("constant pool index %d is tag %d, want Fieldref", idx, e.tag)
	}
	owner, err = p.className(e.classIndex)
	if err != nil {
		return "", "", "", err
	}
	name, desc, err = p.nameAndType(e.nameAndTypeIndex)
	if err != nil {
		return "", "", "", err
	}
	return owner, name, desc, nil
}

// integer resolves an Integer constant to its int32 value — used by an
// annotation element_value's 'B'/'C'/'I'/'S'/'Z' tags (FRAMEWORKS F1),
// each narrowing the same int32 differently per JVM spec 4.7.16.1.
func (p pool) integer(idx uint16) (int32, error) {
	e, err := p.get(idx)
	if err != nil {
		return 0, err
	}
	if e.tag != tagInteger {
		return 0, fmt.Errorf("constant pool index %d is tag %d, want Integer", idx, e.tag)
	}
	return int32(e.raw32), nil
}

// float32Val resolves a Float constant to its float32 value.
func (p pool) float32Val(idx uint16) (float32, error) {
	e, err := p.get(idx)
	if err != nil {
		return 0, err
	}
	if e.tag != tagFloat {
		return 0, fmt.Errorf("constant pool index %d is tag %d, want Float", idx, e.tag)
	}
	return math.Float32frombits(e.raw32), nil
}

// long resolves a Long constant to its int64 value.
func (p pool) long(idx uint16) (int64, error) {
	e, err := p.get(idx)
	if err != nil {
		return 0, err
	}
	if e.tag != tagLong {
		return 0, fmt.Errorf("constant pool index %d is tag %d, want Long", idx, e.tag)
	}
	return int64(e.raw64), nil
}

// doubleVal resolves a Double constant to its float64 value.
func (p pool) doubleVal(idx uint16) (float64, error) {
	e, err := p.get(idx)
	if err != nil {
		return 0, err
	}
	if e.tag != tagDouble {
		return 0, fmt.Errorf("constant pool index %d is tag %d, want Double", idx, e.tag)
	}
	return math.Float64frombits(e.raw64), nil
}

// methodHandle resolves a MethodHandle constant: its reference kind, and
// (for the kinds we care about, all of which point at a
// Methodref/InterfaceMethodref) the target method.
func (p pool) methodHandle(idx uint16) (kind byte, owner, name, desc string, isInterface bool, err error) {
	e, err := p.get(idx)
	if err != nil {
		return 0, "", "", "", false, err
	}
	if e.tag != tagMethodHandle {
		return 0, "", "", "", false, fmt.Errorf("constant pool index %d is tag %d, want MethodHandle", idx, e.tag)
	}
	owner, name, desc, isInterface, err = p.methodref(e.referenceIndex)
	if err != nil {
		return 0, "", "", "", false, err
	}
	return e.referenceKind, owner, name, desc, isInterface, nil
}
