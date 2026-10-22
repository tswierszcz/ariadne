package dataflow

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// writerMethod builds a hierarchy.Method for test.Caller's <clinit>
// that stores a string literal into test.Caller#NAME:Ljava/lang/String;
// at the given PC — reused by both tests below, mirroring D2's own
// TestAnalyze_staticFieldWriteRecordsFieldValueAt bytecode shape.
func writerMethod(name string, pc int) hierarchy.Method {
	code := []byte{
		0x12, 0x08, // pc0 ldc #8 ("x")
		0xb3, 0x00, 0x06, // pc2 putstatic #6 (test.Caller#NAME)
		0xb1, // pc5 return
	}
	return hierarchy.Method{
		Method: classfile.Method{Name: name, Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code},
		InitRefs: []classfile.InitRef{
			{Kind: classfile.InitPutStatic, Owner: "test.Caller", Name: "NAME", Desc: "Ljava/lang/String;", PC: pc},
		},
	}
}

func fieldWriterPool() []classfile.PoolEntry {
	return []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Caller"},                      // 1
		{Tag: classfile.TagClass, NameIndex: 1},                           // 2
		{Tag: classfile.TagUtf8, Str: "NAME"},                             // 3
		{Tag: classfile.TagUtf8, Str: "Ljava/lang/String;"},               // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4}, // 5
		{Tag: classfile.TagFieldref, ClassIndex: 2, NameAndTypeIndex: 5},  // 6: test.Caller#NAME
		{Tag: classfile.TagUtf8, Str: "x"},                                // 7
		{Tag: classfile.TagString, NameIndex: 7},                          // 8
	}
}

func TestMergeFieldWriters_singleWriterResolves(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", fieldWriterPool())
	methods := []hierarchy.Method{writerMethod("<clinit>", 2)}
	ref := valueflow.FieldRef{Owner: "test.Caller", Name: "NAME", Desc: "Ljava/lang/String;"}

	v, warnings := mergeFieldWriters(cf, "test.Caller", methods, ref)
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	want := valueflow.Const("java.lang.String", "x")
	if !v.Equal(want) {
		t.Errorf("mergeFieldWriters = %+v, want %+v", v, want)
	}
}

func TestMergeFieldWriters_noWriterIsHonestUnknown(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", fieldWriterPool())
	// A method that does not write the field at all.
	methods := []hierarchy.Method{{Method: classfile.Method{Name: "m", Descriptor: "()V", Code: []byte{0xb1}}}}
	ref := valueflow.FieldRef{Owner: "test.Caller", Name: "NAME", Desc: "Ljava/lang/String;"}

	v, _ := mergeFieldWriters(cf, "test.Caller", methods, ref)
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("mergeFieldWriters(no writer) = %+v, want KindUnknown", v)
	}
}

func TestResolveStaticField_unresolvableOwnerIsUnknown(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	v, _ := ResolveStaticField(h, valueflow.FieldRef{Owner: "test.NoSuchClass", Name: "X", Desc: "I"})
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("ResolveStaticField(unresolvable owner) = %+v, want KindUnknown", v)
	}
}
