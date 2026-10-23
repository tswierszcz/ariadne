package dataflow

import (
	"testing"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

// classSinkPool builds a constant pool shared by every test below: a
// `test.Sink#accept(Class)V` sink to read a resolved argument back
// out, and `java.lang.Class#forName(String)Class` at a fixed index (9)
// — extra entries for a second overload or an unresolved case are
// appended by each test.
func classSinkPool(extra ...classfile.PoolEntry) []classfile.PoolEntry {
	base := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                             // 1
		{Tag: classfile.TagClass, NameIndex: 1},                                // 2
		{Tag: classfile.TagUtf8, Str: "accept"},                                // 3
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Class;)V"},                  // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},      // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5},      // 6: test.Sink#accept(Class)V
		{Tag: classfile.TagUtf8, Str: "java/lang/Class"},                       // 7
		{Tag: classfile.TagClass, NameIndex: 7},                                // 8
		{Tag: classfile.TagUtf8, Str: "forName"},                               // 9
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;)Ljava/lang/Class;"}, // 10
		{Tag: classfile.TagNameAndType, NameIndex: 9, DescriptorIndex: 10},     // 11
		{Tag: classfile.TagMethodref, ClassIndex: 8, NameAndTypeIndex: 11},     // 12: Class#forName(String)Class
	}
	return append(base, extra...)
}

// argAt reads back the Sink#accept call's own single argument — the
// Value classForNameValue (or the generic fallback) pushed.
func argAt(t *testing.T, cf *classfile.ClassFile, fqcn string, m classfile.Method, acceptPC int) valueflow.Value {
	t.Helper()
	facts, err := Analyze(cf, fqcn, m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}
	return facts.ArgumentAt(bytecodeir.Site{Method: fqcn + "#" + m.Name, Descriptor: m.Descriptor, PC: acceptPC, Op: bytecodeir.OpCall}, 0)
}

// TestClassForNameValue_literalResolves is REFLECTION R1's own
// "jednoznaczna nazwa" case:
//
//	pc0: ldc #14 ("test.Foo")
//	pc2: invokestatic #12 (Class#forName)
//	pc5: invokestatic #6  (Sink#accept)
//	pc8: return
func TestClassForNameValue_literalResolves(t *testing.T) {
	pool := classSinkPool(
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "test.Foo"}, // 13
		classfile.PoolEntry{Tag: classfile.TagString, NameIndex: 13}, // 14
	)
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x0e, // pc0 ldc #14 ("test.Foo")
		0xb8, 0x00, 0x0c, // pc2 invokestatic #12 (Class#forName)
		0xb8, 0x00, 0x06, // pc5 invokestatic #6 (Sink#accept)
		0xb1, // pc8 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	val := argAt(t, cf, "test.Caller", m, 5)
	if val.Kind != valueflow.KindConst || val.Type != "java.lang.Class" || val.Const != "test.Foo" {
		t.Fatalf("ArgumentAt = %+v, want KindConst/java.lang.Class/test.Foo", val)
	}
}

// TestClassForNameValue_throughLocalVariableResolvesSame confirms R1's
// own "Gotowe": a literal name stored in a local and reloaded resolves
// identically to passing it straight through.
//
//	pc0: ldc #14 ("test.Bar")
//	pc2: astore_0
//	pc3: aload_0
//	pc4: invokestatic #12 (Class#forName)
//	pc7: invokestatic #6  (Sink#accept)
//	pc10: return
func TestClassForNameValue_throughLocalVariableResolvesSame(t *testing.T) {
	pool := classSinkPool(
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "test.Bar"},
		classfile.PoolEntry{Tag: classfile.TagString, NameIndex: 13},
	)
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x0e, // pc0 ldc #14 ("test.Bar")
		0x4b,             // pc2 astore_0
		0x2a,             // pc3 aload_0
		0xb8, 0x00, 0x0c, // pc4 invokestatic #12 (Class#forName)
		0xb8, 0x00, 0x06, // pc7 invokestatic #6 (Sink#accept)
		0xb1, // pc10 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 1}

	val := argAt(t, cf, "test.Caller", m, 7)
	if val.Kind != valueflow.KindConst || val.Type != "java.lang.Class" || val.Const != "test.Bar" {
		t.Fatalf("ArgumentAt = %+v, want KindConst/java.lang.Class/test.Bar", val)
	}
}

// TestClassForNameValue_runtimeComputedNameIsNotFabricated confirms
// R1's own explicit non-goal ("poza pierwszym zakresem: dowolne
// runtime-computed nazwy"): a parameter passed straight to forName
// falls through to the generic KindCallResult, never a guessed Class.
//
//	pc0: aload_0 (the String parameter)
//	pc1: invokestatic #12 (Class#forName)
//	pc4: invokestatic #6  (Sink#accept)
//	pc7: return
func TestClassForNameValue_runtimeComputedNameIsNotFabricated(t *testing.T) {
	pool := classSinkPool()
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x2a,             // pc0 aload_0 (name parameter)
		0xb8, 0x00, 0x0c, // pc1 invokestatic #12 (Class#forName)
		0xb8, 0x00, 0x06, // pc4 invokestatic #6 (Sink#accept)
		0xb1, // pc7 return
	}
	m := classfile.Method{Name: "m", Descriptor: "(Ljava/lang/String;)V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 1}

	val := argAt(t, cf, "test.Caller", m, 4)
	if val.Kind != valueflow.KindCallResult {
		t.Fatalf("ArgumentAt = %+v, want KindCallResult (not fabricated)", val)
	}
	if val.Call.Owner != classOwner || val.Call.Name != "forName" {
		t.Errorf("CallResultRef = %+v, want Owner=%s Name=forName", val.Call, classOwner)
	}
}

// TestClassForNameValue_threeArgOverloadResolves confirms the
// `forName(String, boolean, ClassLoader)` overload is recognized too,
// by its own distinct descriptor, same name-argument index (0).
//
//	pc0: ldc #16 ("test.Baz")
//	pc2: iconst_1
//	pc3: aload_0 (ClassLoader parameter)
//	pc4: invokestatic #15 (Class#forName, 3-arg)
//	pc7: invokestatic #6  (Sink#accept)
//	pc10: return
func TestClassForNameValue_threeArgOverloadResolves(t *testing.T) {
	pool := classSinkPool(
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;ZLjava/lang/ClassLoader;)Ljava/lang/Class;"}, // 13
		classfile.PoolEntry{Tag: classfile.TagNameAndType, NameIndex: 9, DescriptorIndex: 13},                             // 14
		classfile.PoolEntry{Tag: classfile.TagMethodref, ClassIndex: 8, NameAndTypeIndex: 14},                             // 15: Class#forName(String,boolean,ClassLoader)
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "test.Baz"},                                                      // 16
		classfile.PoolEntry{Tag: classfile.TagString, NameIndex: 16},                                                      // 17
	)
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x11, // pc0 ldc #17 ("test.Baz")
		0x04,             // pc2 iconst_1
		0x2a,             // pc3 aload_0 (ClassLoader parameter)
		0xb8, 0x00, 0x0f, // pc4 invokestatic #15 (Class#forName, 3-arg)
		0xb8, 0x00, 0x06, // pc7 invokestatic #6 (Sink#accept)
		0xb1, // pc10 return
	}
	m := classfile.Method{Name: "m", Descriptor: "(Ljava/lang/ClassLoader;)V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 1}

	val := argAt(t, cf, "test.Caller", m, 7)
	if val.Kind != valueflow.KindConst || val.Type != "java.lang.Class" || val.Const != "test.Baz" {
		t.Fatalf("ArgumentAt = %+v, want KindConst/java.lang.Class/test.Baz", val)
	}
}
