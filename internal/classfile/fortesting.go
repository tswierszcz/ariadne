package classfile

import (
	"fmt"
	"strings"
)

// PoolEntry is one constant-pool entry, exposed only so a test in another
// package can build a *ClassFile whose DecodeLocalOps/ExtractCalls
// resolve real references (ldc of a string/class, methodref, fieldref)
// without hand-assembling a full binary .class file — this repo's own
// internal tests already hand-build cpEntry directly (see
// internal/classfile/initrefs_test.go); NewClassFileForTesting is the
// same pattern, usable from outside this package. Mirrors
// hierarchy.NewForTesting's own reasoning for ClassInfo.
//
// Field meanings are exactly cpEntry's own (see pool.go's own doc
// comment on each tag): set only the fields the chosen Tag uses.
type PoolEntry struct {
	Tag byte

	Str string // TagUtf8

	NameIndex uint16 // TagClass/TagString/TagMethodType/TagModule/TagPackage

	ClassIndex       uint16 // TagFieldref/TagMethodref/TagInterfaceMethodref
	NameAndTypeIndex uint16 // TagFieldref/TagMethodref/TagInterfaceMethodref/TagDynamic/TagInvokeDynamic

	DescriptorIndex uint16 // TagNameAndType

	ReferenceKind  byte   // TagMethodHandle
	ReferenceIndex uint16 // TagMethodHandle

	BootstrapMethodAttrIndex uint16 // TagDynamic/TagInvokeDynamic

	Raw32 uint32 // TagInteger (int32 bits)/TagFloat (IEEE 754 bits)
	Raw64 uint64 // TagLong (int64 bits)/TagDouble (IEEE 754 bits)
}

// Constant pool tags, exported for NewClassFileForTesting callers — the
// same values as this package's own unexported tagXxx constants (pool.go).
const (
	TagUtf8               = tagUtf8
	TagInteger            = tagInteger
	TagFloat              = tagFloat
	TagLong               = tagLong
	TagDouble             = tagDouble
	TagClass              = tagClass
	TagString             = tagString
	TagFieldref           = tagFieldref
	TagMethodref          = tagMethodref
	TagInterfaceMethodref = tagInterfaceMethodref
	TagNameAndType        = tagNameAndType
	TagMethodHandle       = tagMethodHandle
	TagMethodType         = tagMethodType
	TagDynamic            = tagDynamic
	TagInvokeDynamic      = tagInvokeDynamic
)

// NewClassFileForTesting builds a *ClassFile with ThisClass and a
// constant pool built verbatim from entries, index 0 left unused same as
// the real constant pool (entries[0] becomes pool index 1, and so on —
// callers index from 1, matching every real bytecode operand). Bootstrap
// methods and methods/fields are left empty; set them directly on the
// returned *ClassFile's exported fields if a test needs them.
func NewClassFileForTesting(thisClass string, entries []PoolEntry) *ClassFile {
	p := make(pool, len(entries)+1)
	for i, e := range entries {
		p[i+1] = cpEntry{
			tag:                      e.Tag,
			str:                      e.Str,
			nameIndex:                e.NameIndex,
			classIndex:               e.ClassIndex,
			nameAndTypeIndex:         e.NameAndTypeIndex,
			descriptorIndex:          e.DescriptorIndex,
			referenceKind:            e.ReferenceKind,
			referenceIndex:           e.ReferenceIndex,
			bootstrapMethodAttrIndex: e.BootstrapMethodAttrIndex,
			raw32:                    e.Raw32,
			raw64:                    e.Raw64,
		}
	}
	return &ClassFile{ThisClass: thisClass, pool: p}
}

// RelocateClassForTesting returns class file b with every occurrence of
// the internal package prefix from+"/" replaced by to+"/" in its Utf8
// constants — class names, descriptors and signatures — which is what a
// shading tool's relocation does (CVE:F7a fixtures; no ASM, no JDK).
// The constant pool is rewritten in place, so indices do not move.
func RelocateClassForTesting(b []byte, from, to string) ([]byte, error) {
	if len(b) < 10 {
		return nil, fmt.Errorf("class file too short")
	}
	count := int(uint16(b[8])<<8 | uint16(b[9]))
	out := append([]byte(nil), b[:10]...)
	pos := 10
	for i := 1; i < count; i++ {
		if pos >= len(b) {
			return nil, fmt.Errorf("truncated constant pool")
		}
		tag := b[pos]
		var n int
		switch tag {
		case tagUtf8:
			l := int(uint16(b[pos+1])<<8 | uint16(b[pos+2]))
			s := string(b[pos+3 : pos+3+l])
			s = strings.ReplaceAll(s, from+"/", to+"/")
			out = append(out, tagUtf8, byte(len(s)>>8), byte(len(s)))
			out = append(out, s...)
			pos += 3 + l
			continue
		case tagInteger, tagFloat, tagFieldref, tagMethodref, tagInterfaceMethodref, tagNameAndType, tagDynamic, tagInvokeDynamic:
			n = 5
		case tagLong, tagDouble:
			n = 9
			i++
		case tagClass, tagString, tagMethodType, tagModule, tagPackage:
			n = 3
		case tagMethodHandle:
			n = 4
		default:
			return nil, fmt.Errorf("unknown constant pool tag %d", tag)
		}
		out = append(out, b[pos:pos+n]...)
		pos += n
	}
	return append(out, b[pos:]...), nil
}
