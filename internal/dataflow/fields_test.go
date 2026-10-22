package dataflow

import (
	"testing"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

// TestAnalyze_staticFieldReadProducesKindField is DATA_FLOW D3's own
// first example: a getstatic must identify which field it read (not
// just its declared type), read back out by passing it straight to a
// consuming call — same pattern as the lambda/handle tests.
//
// Bytecode (static void m()):
//
//	pc0: getstatic #6 (test.Caller#HANDLE:Ljava/lang/invoke/MethodHandle;)
//	pc3: invokestatic #12 (test.Sink#accept(MethodHandle)V)
//	pc6: return
func TestAnalyze_staticFieldReadProducesKindField(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Caller"},                        // 1
		{Tag: classfile.TagClass, NameIndex: 1},                             // 2
		{Tag: classfile.TagUtf8, Str: "HANDLE"},                             // 3
		{Tag: classfile.TagUtf8, Str: "Ljava/lang/invoke/MethodHandle;"},    // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},   // 5
		{Tag: classfile.TagFieldref, ClassIndex: 2, NameAndTypeIndex: 5},    // 6: test.Caller#HANDLE
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                          // 7
		{Tag: classfile.TagClass, NameIndex: 7},                             // 8
		{Tag: classfile.TagUtf8, Str: "accept"},                             // 9
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/invoke/MethodHandle;)V"}, // 10
		{Tag: classfile.TagNameAndType, NameIndex: 9, DescriptorIndex: 10},  // 11
		{Tag: classfile.TagMethodref, ClassIndex: 8, NameAndTypeIndex: 11},  // 12: test.Sink#accept
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0xb2, 0x00, 0x06, // pc0 getstatic #6
		0xb8, 0x00, 0x0c, // pc3 invokestatic #12
		0xb1, // pc6 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	val := facts.ArgumentAt(bytecodeir.Site{Method: "test.Caller#m", Descriptor: "()V", PC: 3, Op: bytecodeir.OpCall}, 0)
	if val.Kind != valueflow.KindField {
		t.Fatalf("ArgumentAt = %+v, want KindField", val)
	}
	want := valueflow.FieldRef{Owner: "test.Caller", Name: "HANDLE", Desc: "Ljava/lang/invoke/MethodHandle;"}
	if *val.Field != want {
		t.Errorf("Field = %+v, want %+v", *val.Field, want)
	}
}

// TestAnalyze_staticFieldWriteRecordsFieldValueAt is D3's own second
// example: a putstatic's own stored value must be recoverable via
// FieldValueAt — here, a string literal assigned to a static field.
//
// Bytecode (static void m()):
//
//	pc0: ldc #7 ("x")
//	pc2: putstatic #6 (test.Caller#NAME:Ljava/lang/String;)
//	pc5: return
func TestAnalyze_staticFieldWriteRecordsFieldValueAt(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Caller"},                      // 1
		{Tag: classfile.TagClass, NameIndex: 1},                           // 2
		{Tag: classfile.TagUtf8, Str: "NAME"},                             // 3
		{Tag: classfile.TagUtf8, Str: "Ljava/lang/String;"},               // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4}, // 5
		{Tag: classfile.TagFieldref, ClassIndex: 2, NameAndTypeIndex: 5},  // 6: test.Caller#NAME
		{Tag: classfile.TagUtf8, Str: "x"},                                // 7
		{Tag: classfile.TagString, NameIndex: 7},                          // 8
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x08, // pc0 ldc #8 ("x")
		0xb3, 0x00, 0x06, // pc2 putstatic #6
		0xb1, // pc5 return
	}
	m := classfile.Method{Name: "<clinit>", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	val := facts.FieldValueAt(bytecodeir.Site{Method: "test.Caller#<clinit>", Descriptor: "()V", PC: 2, Op: bytecodeir.OpFieldPut})
	want := valueflow.Const("java.lang.String", "x")
	if !val.Equal(want) {
		t.Errorf("FieldValueAt = %+v, want %+v", val, want)
	}
}
