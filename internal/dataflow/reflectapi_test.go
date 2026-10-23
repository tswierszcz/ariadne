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

// TestReflectiveMethodValue_literalGetMethodResolves is REFLECTION
// R1's own full array-literal case:
//
//	pc0:  ldc #8  (test.Foo.class)        -- receiver
//	pc2:  ldc #10 ("bar")                 -- method name
//	pc4:  iconst_1                        -- array length 1
//	pc5:  anewarray #12 (java/lang/Class)
//	pc8:  dup
//	pc9:  iconst_0
//	pc10: ldc #14 (java.lang.String.class)
//	pc12: aastore
//	pc13: invokevirtual #18 (Class#getMethod)
//	pc16: invokestatic #6 (Sink#accept(Method)V)
//	pc19: return
func TestReflectiveMethodValue_literalGetMethodResolves(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                                                        // 1
		{Tag: classfile.TagClass, NameIndex: 1},                                                           // 2
		{Tag: classfile.TagUtf8, Str: "accept"},                                                           // 3
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/reflect/Method;)V"},                                    // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},                                 // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5},                                 // 6: Sink#accept(Method)V
		{Tag: classfile.TagUtf8, Str: "test/Foo"},                                                         // 7
		{Tag: classfile.TagClass, NameIndex: 7},                                                           // 8: test.Foo.class
		{Tag: classfile.TagUtf8, Str: "bar"},                                                              // 9
		{Tag: classfile.TagString, NameIndex: 9},                                                          // 10: "bar"
		{Tag: classfile.TagUtf8, Str: "java/lang/Class"},                                                  // 11
		{Tag: classfile.TagClass, NameIndex: 11},                                                          // 12
		{Tag: classfile.TagUtf8, Str: "java/lang/String"},                                                 // 13
		{Tag: classfile.TagClass, NameIndex: 13},                                                          // 14: java.lang.String.class
		{Tag: classfile.TagUtf8, Str: "getMethod"},                                                        // 15
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;"}, // 16
		{Tag: classfile.TagNameAndType, NameIndex: 15, DescriptorIndex: 16},                               // 17
		{Tag: classfile.TagMethodref, ClassIndex: 12, NameAndTypeIndex: 17},                               // 18: Class#getMethod
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x08, // pc0  ldc #8  (test.Foo.class)
		0x12, 0x0a, // pc2  ldc #10 ("bar")
		0x04,             // pc4  iconst_1
		0xbd, 0x00, 0x0c, // pc5  anewarray #12 (java/lang/Class)
		0x59,       // pc8  dup
		0x03,       // pc9  iconst_0
		0x12, 0x0e, // pc10 ldc #14 (java.lang.String.class)
		0x53,             // pc12 aastore
		0xb6, 0x00, 0x12, // pc13 invokevirtual #18 (Class#getMethod)
		0xb8, 0x00, 0x06, // pc16 invokestatic #6 (Sink#accept)
		0xb1, // pc19 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	val := argAt(t, cf, "test.Caller", m, 16)
	if val.Kind != valueflow.KindReflectiveMethod {
		t.Fatalf("ArgumentAt = %+v, want KindReflectiveMethod", val)
	}
	rm := val.ReflectiveMethod
	if rm.OwnerFQCN != "test.Foo" || rm.Name != "bar" || rm.Declared {
		t.Errorf("OwnerFQCN/Name/Declared = %q/%q/%v, want test.Foo/bar/false", rm.OwnerFQCN, rm.Name, rm.Declared)
	}
	if len(rm.ParamTypes) != 1 || rm.ParamTypes[0] != "java.lang.String" {
		t.Errorf("ParamTypes = %v, want [java.lang.String]", rm.ParamTypes)
	}
}

// TestReflectiveMethodValue_twoElementArrayResolves confirms the
// mutation-through-a-shared-pointer mechanism (ArrayValue's own doc
// comment) holds across more than one `dup`+`aastore` write, not just
// one — `getMethod("qux", String.class, Integer.class)`.
//
//	pc0:  ldc #8  (test.Foo.class)         -- receiver
//	pc2:  ldc #10 ("qux")
//	pc4:  iconst_2
//	pc5:  anewarray #12 (java/lang/Class)
//	pc8:  dup
//	pc9:  iconst_0
//	pc10: ldc #14 (java.lang.String.class)
//	pc12: aastore
//	pc13: dup
//	pc14: iconst_1
//	pc15: ldc #16 (java.lang.Integer.class)
//	pc17: aastore
//	pc18: invokevirtual #20 (Class#getMethod)
//	pc21: invokestatic #6 (Sink#accept)
//	pc24: return
func TestReflectiveMethodValue_twoElementArrayResolves(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Sink"},
		{Tag: classfile.TagClass, NameIndex: 1},
		{Tag: classfile.TagUtf8, Str: "accept"},
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/reflect/Method;)V"},
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5}, // 6
		{Tag: classfile.TagUtf8, Str: "test/Foo"},
		{Tag: classfile.TagClass, NameIndex: 7}, // 8
		{Tag: classfile.TagUtf8, Str: "qux"},
		{Tag: classfile.TagString, NameIndex: 9}, // 10
		{Tag: classfile.TagUtf8, Str: "java/lang/Class"},
		{Tag: classfile.TagClass, NameIndex: 11}, // 12
		{Tag: classfile.TagUtf8, Str: "java/lang/String"},
		{Tag: classfile.TagClass, NameIndex: 13}, // 14
		{Tag: classfile.TagUtf8, Str: "java/lang/Integer"},
		{Tag: classfile.TagClass, NameIndex: 15}, // 16
		{Tag: classfile.TagUtf8, Str: "getMethod"},
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;"},
		{Tag: classfile.TagNameAndType, NameIndex: 17, DescriptorIndex: 18},
		{Tag: classfile.TagMethodref, ClassIndex: 12, NameAndTypeIndex: 19}, // 20
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x08, // pc0  ldc #8 (test.Foo.class)
		0x12, 0x0a, // pc2  ldc #10 ("qux")
		0x05,             // pc4  iconst_2
		0xbd, 0x00, 0x0c, // pc5  anewarray #12 (java/lang/Class)
		0x59,       // pc8  dup
		0x03,       // pc9  iconst_0
		0x12, 0x0e, // pc10 ldc #14 (java.lang.String.class)
		0x53,       // pc12 aastore
		0x59,       // pc13 dup
		0x04,       // pc14 iconst_1
		0x12, 0x10, // pc15 ldc #16 (java.lang.Integer.class)
		0x53,             // pc17 aastore
		0xb6, 0x00, 0x14, // pc18 invokevirtual #20 (Class#getMethod)
		0xb8, 0x00, 0x06, // pc21 invokestatic #6 (Sink#accept)
		0xb1, // pc24 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	val := argAt(t, cf, "test.Caller", m, 21)
	if val.Kind != valueflow.KindReflectiveMethod {
		t.Fatalf("ArgumentAt = %+v, want KindReflectiveMethod", val)
	}
	rm := val.ReflectiveMethod
	want := []string{"java.lang.String", "java.lang.Integer"}
	if rm.Name != "qux" || len(rm.ParamTypes) != 2 || rm.ParamTypes[0] != want[0] || rm.ParamTypes[1] != want[1] {
		t.Errorf("Name/ParamTypes = %q/%v, want qux/%v", rm.Name, rm.ParamTypes, want)
	}
}

// TestReflectiveMethodValue_zeroArgArrayResolves confirms a no-
// parameter lookup (`getDeclaredMethod("baz")`, compiled as a literal
// zero-length `Class[]`) resolves with an empty ParamTypes, not
// skipped as if the array were never written to.
//
//	pc0: ldc #8  (test.Foo.class)
//	pc2: ldc #10 ("baz")
//	pc4: iconst_0
//	pc5: anewarray #12 (java/lang/Class)
//	pc8: invokevirtual #18 (Class#getDeclaredMethod)
//	pc11: invokestatic #6 (Sink#accept)
//	pc14: return
func TestReflectiveMethodValue_zeroArgArrayResolves(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Sink"},
		{Tag: classfile.TagClass, NameIndex: 1},
		{Tag: classfile.TagUtf8, Str: "accept"},
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/reflect/Method;)V"},
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5}, // 6
		{Tag: classfile.TagUtf8, Str: "test/Foo"},
		{Tag: classfile.TagClass, NameIndex: 7}, // 8
		{Tag: classfile.TagUtf8, Str: "baz"},
		{Tag: classfile.TagString, NameIndex: 9}, // 10
		{Tag: classfile.TagUtf8, Str: "java/lang/Class"},
		{Tag: classfile.TagClass, NameIndex: 11}, // 12
		{Tag: classfile.TagUtf8, Str: "getDeclaredMethod"},
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;"},
		{Tag: classfile.TagNameAndType, NameIndex: 13, DescriptorIndex: 14},
		{Tag: classfile.TagMethodref, ClassIndex: 12, NameAndTypeIndex: 15}, // 16
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x08, // pc0 ldc #8 (test.Foo.class)
		0x12, 0x0a, // pc2 ldc #10 ("baz")
		0x03,             // pc4 iconst_0
		0xbd, 0x00, 0x0c, // pc5 anewarray #12 (java/lang/Class)
		0xb6, 0x00, 0x10, // pc8 invokevirtual #16 (Class#getDeclaredMethod)
		0xb8, 0x00, 0x06, // pc11 invokestatic #6 (Sink#accept)
		0xb1, // pc14 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	val := argAt(t, cf, "test.Caller", m, 11)
	if val.Kind != valueflow.KindReflectiveMethod {
		t.Fatalf("ArgumentAt = %+v, want KindReflectiveMethod", val)
	}
	rm := val.ReflectiveMethod
	if rm.OwnerFQCN != "test.Foo" || rm.Name != "baz" || !rm.Declared {
		t.Errorf("OwnerFQCN/Name/Declared = %q/%q/%v, want test.Foo/baz/true", rm.OwnerFQCN, rm.Name, rm.Declared)
	}
	if len(rm.ParamTypes) != 0 {
		t.Errorf("ParamTypes = %v, want empty", rm.ParamTypes)
	}
}

// TestReflectiveMethodValue_nonLiteralNameIsNotFabricated confirms a
// runtime-computed method name falls through to the generic
// KindUnknown (a virtual call's return value is never resolved
// generically — CHA ambiguity, same as any other virtual call), never
// fabricated.
//
//	pc0:  ldc #8  (test.Foo.class)
//	pc2:  aload_0 (name parameter)
//	pc3:  iconst_0
//	pc4:  anewarray #10 (java/lang/Class)
//	pc7:  invokevirtual #14 (Class#getMethod)
//	pc10: invokestatic #6 (Sink#accept)
//	pc13: return
func TestReflectiveMethodValue_nonLiteralNameIsNotFabricated(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Sink"},
		{Tag: classfile.TagClass, NameIndex: 1},
		{Tag: classfile.TagUtf8, Str: "accept"},
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/reflect/Method;)V"},
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5}, // 6
		{Tag: classfile.TagUtf8, Str: "test/Foo"},
		{Tag: classfile.TagClass, NameIndex: 7}, // 8
		{Tag: classfile.TagUtf8, Str: "java/lang/Class"},
		{Tag: classfile.TagClass, NameIndex: 9}, // 10
		{Tag: classfile.TagUtf8, Str: "getMethod"},
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;"},
		{Tag: classfile.TagNameAndType, NameIndex: 11, DescriptorIndex: 12},
		{Tag: classfile.TagMethodref, ClassIndex: 10, NameAndTypeIndex: 13}, // 14
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x08, // pc0 ldc #8 (test.Foo.class)
		0x2a,             // pc1 aload_0 (name parameter)
		0x03,             // pc2 iconst_0
		0xbd, 0x00, 0x0a, // pc3 anewarray #10 (java/lang/Class)
		0xb6, 0x00, 0x0e, // pc6 invokevirtual #14 (Class#getMethod)
		0xb8, 0x00, 0x06, // pc9 invokestatic #6 (Sink#accept)
		0xb1, // pc12 return
	}
	m := classfile.Method{Name: "m", Descriptor: "(Ljava/lang/String;)V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 1}

	val := argAt(t, cf, "test.Caller", m, 10)
	if val.Kind != valueflow.KindUnknown {
		t.Fatalf("ArgumentAt = %+v, want KindUnknown (not fabricated)", val)
	}
}
