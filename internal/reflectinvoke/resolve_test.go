package reflectinvoke

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
)

// TestResolveOne_positivePath exercises REFLECTION R2's own full
// pipeline on real bytecode:
//
//	pc0:  new #2 (test.Target)
//	pc3:  dup
//	pc4:  invokespecial #26 (test.Target#<init>)
//	pc7:  astore_0                                  -- local0 = t
//	pc8:  ldc #2  (test.Target.class)                -- getMethod receiver
//	pc10: ldc #6  ("bar")
//	pc12: iconst_1
//	pc13: anewarray #4 (java/lang/Class)
//	pc16: dup
//	pc17: iconst_0
//	pc18: ldc #8  (java.lang.String.class)
//	pc20: aastore
//	pc21: invokevirtual #12 (Class#getMethod)
//	pc24: astore_1                                  -- local1 = md
//	pc25: iconst_1
//	pc26: anewarray #14 (java/lang/Object)
//	pc29: dup
//	pc30: iconst_0
//	pc31: ldc #16 ("x")
//	pc33: aastore
//	pc34: astore_2                                  -- local2 = args
//	pc35: aload_1 (md)
//	pc36: aload_0 (t)
//	pc37: aload_2 (args)
//	pc38: invokevirtual #22 (Method#invoke)
//	pc41: pop
//	pc42: return
func TestResolveOne_positivePath(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Target"},                                                      // 1
		{Tag: classfile.TagClass, NameIndex: 1},                                                           // 2: test.Target
		{Tag: classfile.TagUtf8, Str: "java/lang/Class"},                                                  // 3
		{Tag: classfile.TagClass, NameIndex: 3},                                                           // 4: java.lang.Class
		{Tag: classfile.TagUtf8, Str: "bar"},                                                              // 5
		{Tag: classfile.TagString, NameIndex: 5},                                                          // 6: "bar"
		{Tag: classfile.TagUtf8, Str: "java/lang/String"},                                                 // 7
		{Tag: classfile.TagClass, NameIndex: 7},                                                           // 8: java.lang.String.class
		{Tag: classfile.TagUtf8, Str: "getMethod"},                                                        // 9
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/reflect/Method;"}, // 10
		{Tag: classfile.TagNameAndType, NameIndex: 9, DescriptorIndex: 10},                                // 11
		{Tag: classfile.TagMethodref, ClassIndex: 4, NameAndTypeIndex: 11},                                // 12: Class#getMethod
		{Tag: classfile.TagUtf8, Str: "java/lang/Object"},                                                 // 13
		{Tag: classfile.TagClass, NameIndex: 13},                                                          // 14: java.lang.Object
		{Tag: classfile.TagUtf8, Str: "x"},                                                                // 15
		{Tag: classfile.TagString, NameIndex: 15},                                                         // 16: "x"
		{Tag: classfile.TagUtf8, Str: "java/lang/reflect/Method"},                                         // 17
		{Tag: classfile.TagClass, NameIndex: 17},                                                          // 18: java.lang.reflect.Method
		{Tag: classfile.TagUtf8, Str: "invoke"},                                                           // 19
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;"},        // 20
		{Tag: classfile.TagNameAndType, NameIndex: 19, DescriptorIndex: 20},                               // 21
		{Tag: classfile.TagMethodref, ClassIndex: 18, NameAndTypeIndex: 21},                               // 22: Method#invoke
		{Tag: classfile.TagUtf8, Str: "<init>"},                                                           // 23
		{Tag: classfile.TagUtf8, Str: "()V"},                                                              // 24
		{Tag: classfile.TagNameAndType, NameIndex: 23, DescriptorIndex: 24},                               // 25
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 25},                                // 26: test.Target#<init>()V
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0xbb, 0x00, 0x02, // pc0  new #2
		0x59,             // pc3  dup
		0xb7, 0x00, 0x1a, // pc4  invokespecial #26
		0x4b,       // pc7  astore_0
		0x12, 0x02, // pc8  ldc #2 (test.Target.class)
		0x12, 0x06, // pc10 ldc #6 ("bar")
		0x04,             // pc12 iconst_1
		0xbd, 0x00, 0x04, // pc13 anewarray #4 (Class[])
		0x59,       // pc16 dup
		0x03,       // pc17 iconst_0
		0x12, 0x08, // pc18 ldc #8 (java.lang.String.class)
		0x53,             // pc20 aastore
		0xb6, 0x00, 0x0c, // pc21 invokevirtual #12 (Class#getMethod)
		0x4c,             // pc24 astore_1
		0x04,             // pc25 iconst_1
		0xbd, 0x00, 0x0e, // pc26 anewarray #14 (Object[])
		0x59,       // pc29 dup
		0x03,       // pc30 iconst_0
		0x12, 0x10, // pc31 ldc #16 ("x")
		0x53,             // pc33 aastore
		0x4d,             // pc34 astore_2
		0x2b,             // pc35 aload_1 (md)
		0x2a,             // pc36 aload_0 (t)
		0x2c,             // pc37 aload_2 (args)
		0xb6, 0x00, 0x16, // pc38 invokevirtual #22 (Method#invoke)
		0x57, // pc41 pop
		0xb1, // pc42 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 3}

	facts, err := dataflow.Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	site := InvokeSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 38}
	ri, warnings := resolveOne(hierarchy.NewForTesting(nil), "test.Caller", site, facts)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if ri.UnresolvedReason != "" {
		t.Fatalf("UnresolvedReason = %q, want empty", ri.UnresolvedReason)
	}
	if ri.OwnerFQCN != "test.Target" || ri.MethodName != "bar" {
		t.Errorf("OwnerFQCN/MethodName = %q/%q, want test.Target/bar", ri.OwnerFQCN, ri.MethodName)
	}
	if len(ri.ParamTypes) != 1 || ri.ParamTypes[0] != "java.lang.String" {
		t.Errorf("ParamTypes = %v, want [java.lang.String]", ri.ParamTypes)
	}
	if ri.ObjType != "test.Target" {
		t.Errorf("ObjType = %q, want test.Target (real dispatch root)", ri.ObjType)
	}
	if !ri.ArgCountKnown || ri.ArgCount != 1 {
		t.Errorf("ArgCount/ArgCountKnown = %d/%v, want 1/true", ri.ArgCount, ri.ArgCountKnown)
	}
}

// TestResolveOne_nonReflectiveReceiverIsUnresolved confirms a Method
// receiver that isn't a provable KindReflectiveMethod (here: a plain
// method parameter) is honestly left unresolved, never fabricated.
//
//	pc0: aload_0 (Method parameter)
//	pc1: aconst_null (obj)
//	pc2: aconst_null (args)
//	pc3: invokevirtual #22 (Method#invoke)
//	pc6: pop
//	pc7: return
func TestResolveOne_nonReflectiveReceiverIsUnresolved(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "java/lang/reflect/Method"},                                  // 1
		{Tag: classfile.TagClass, NameIndex: 1},                                                    // 2
		{Tag: classfile.TagUtf8, Str: "invoke"},                                                    // 3
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;"}, // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},                          // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5},                          // 6
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x2a,             // pc0 aload_0 (Method parameter)
		0x01,             // pc1 aconst_null
		0x01,             // pc2 aconst_null
		0xb6, 0x00, 0x06, // pc3 invokevirtual #6 (Method#invoke)
		0x57, // pc6 pop
		0xb1, // pc7 return
	}
	m := classfile.Method{Name: "m", Descriptor: "(Ljava/lang/reflect/Method;)V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 1}

	facts, err := dataflow.Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	site := InvokeSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "(Ljava/lang/reflect/Method;)V", PC: 3}
	ri, _ := resolveOne(hierarchy.NewForTesting(nil), "test.Caller", site, facts)
	if ri.OwnerFQCN != "" {
		t.Errorf("OwnerFQCN = %q, want empty", ri.OwnerFQCN)
	}
	if ri.UnresolvedReason == "" {
		t.Errorf("UnresolvedReason is empty, want a non-empty explanation")
	}
}
