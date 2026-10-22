package classfile

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
		}
	}
	return &ClassFile{ThisClass: thisClass, pool: p}
}
