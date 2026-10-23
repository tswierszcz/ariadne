package reflectinvoke

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
)

// TestResolveOneNewInstance_positivePath exercises REFLECTION R2's own
// second-half pipeline on real bytecode — resolveOne's own counterpart
// for allocation rather than invocation:
//
//	pc0:  ldc #2  (test.Target.class)                -- getConstructor receiver
//	pc2:  iconst_1
//	pc3:  anewarray #4 (java/lang/Class)
//	pc6:  dup
//	pc7:  iconst_0
//	pc8:  ldc #6  (java.lang.String.class)
//	pc10: aastore
//	pc11: invokevirtual #10 (Class#getConstructor)
//	pc14: astore_0                                   -- local0 = ctor
//	pc15: iconst_1
//	pc16: anewarray #12 (java/lang/Object)
//	pc19: dup
//	pc20: iconst_0
//	pc21: ldc #14 ("x")
//	pc23: aastore
//	pc24: astore_1                                   -- local1 = args
//	pc25: aload_0 (ctor)
//	pc26: aload_1 (args)
//	pc27: invokevirtual #20 (Constructor#newInstance)
//	pc30: pop
//	pc31: return
func TestResolveOneNewInstance_positivePath(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Target"},                                         // 1
		{Tag: classfile.TagClass, NameIndex: 1},                                              // 2: test.Target
		{Tag: classfile.TagUtf8, Str: "java/lang/Class"},                                     // 3
		{Tag: classfile.TagClass, NameIndex: 3},                                              // 4: java.lang.Class
		{Tag: classfile.TagUtf8, Str: "java/lang/String"},                                    // 5
		{Tag: classfile.TagClass, NameIndex: 5},                                              // 6: java.lang.String.class
		{Tag: classfile.TagUtf8, Str: "getConstructor"},                                      // 7
		{Tag: classfile.TagUtf8, Str: "([Ljava/lang/Class;)Ljava/lang/reflect/Constructor;"}, // 8
		{Tag: classfile.TagNameAndType, NameIndex: 7, DescriptorIndex: 8},                    // 9
		{Tag: classfile.TagMethodref, ClassIndex: 4, NameAndTypeIndex: 9},                    // 10: Class#getConstructor
		{Tag: classfile.TagUtf8, Str: "java/lang/Object"},                                    // 11
		{Tag: classfile.TagClass, NameIndex: 11},                                             // 12: java.lang.Object
		{Tag: classfile.TagUtf8, Str: "x"},                                                   // 13
		{Tag: classfile.TagString, NameIndex: 13},                                            // 14: "x"
		{Tag: classfile.TagUtf8, Str: "java/lang/reflect/Constructor"},                       // 15
		{Tag: classfile.TagClass, NameIndex: 15},                                             // 16: java.lang.reflect.Constructor
		{Tag: classfile.TagUtf8, Str: "newInstance"},                                         // 17
		{Tag: classfile.TagUtf8, Str: "([Ljava/lang/Object;)Ljava/lang/Object;"},             // 18
		{Tag: classfile.TagNameAndType, NameIndex: 17, DescriptorIndex: 18},                  // 19
		{Tag: classfile.TagMethodref, ClassIndex: 16, NameAndTypeIndex: 19},                  // 20: Constructor#newInstance
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x12, 0x02, // pc0  ldc #2 (test.Target.class)
		0x04,             // pc2  iconst_1
		0xbd, 0x00, 0x04, // pc3  anewarray #4 (Class[])
		0x59,       // pc6  dup
		0x03,       // pc7  iconst_0
		0x12, 0x06, // pc8  ldc #6 (java.lang.String.class)
		0x53,             // pc10 aastore
		0xb6, 0x00, 0x0a, // pc11 invokevirtual #10 (Class#getConstructor)
		0x4b,             // pc14 astore_0
		0x04,             // pc15 iconst_1
		0xbd, 0x00, 0x0c, // pc16 anewarray #12 (Object[])
		0x59,       // pc19 dup
		0x03,       // pc20 iconst_0
		0x12, 0x0e, // pc21 ldc #14 ("x")
		0x53,             // pc23 aastore
		0x4c,             // pc24 astore_1
		0x2a,             // pc25 aload_0 (ctor)
		0x2b,             // pc26 aload_1 (args)
		0xb6, 0x00, 0x14, // pc27 invokevirtual #20 (Constructor#newInstance)
		0x57, // pc30 pop
		0xb1, // pc31 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}

	facts, err := dataflow.Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	site := NewInstanceSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 27}
	ri, warnings := resolveOneNewInstance(hierarchy.NewForTesting(nil), "test.Caller", site, facts)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if ri.UnresolvedReason != "" {
		t.Fatalf("UnresolvedReason = %q, want empty", ri.UnresolvedReason)
	}
	if ri.OwnerFQCN != "test.Target" {
		t.Errorf("OwnerFQCN = %q, want test.Target", ri.OwnerFQCN)
	}
	if len(ri.ParamTypes) != 1 || ri.ParamTypes[0] != "java.lang.String" {
		t.Errorf("ParamTypes = %v, want [java.lang.String]", ri.ParamTypes)
	}
	if !ri.ArgCountKnown || ri.ArgCount != 1 {
		t.Errorf("ArgCount/ArgCountKnown = %d/%v, want 1/true", ri.ArgCount, ri.ArgCountKnown)
	}
}

// TestResolveOneNewInstance_nonReflectiveReceiverIsUnresolved confirms
// a Constructor receiver that isn't a provable
// KindReflectiveConstructor (here: a plain method parameter) is
// honestly left unresolved, never fabricated —
// TestResolveOne_nonReflectiveReceiverIsUnresolved's own counterpart.
//
//	pc0: aload_0 (Constructor parameter)
//	pc1: aconst_null (args)
//	pc2: invokevirtual #6 (Constructor#newInstance)
//	pc5: pop
//	pc6: return
func TestResolveOneNewInstance_nonReflectiveReceiverIsUnresolved(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "java/lang/reflect/Constructor"},           // 1
		{Tag: classfile.TagClass, NameIndex: 1},                                  // 2
		{Tag: classfile.TagUtf8, Str: "newInstance"},                             // 3
		{Tag: classfile.TagUtf8, Str: "([Ljava/lang/Object;)Ljava/lang/Object;"}, // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},        // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5},        // 6
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x2a,             // pc0 aload_0 (Constructor parameter)
		0x01,             // pc1 aconst_null
		0xb6, 0x00, 0x06, // pc2 invokevirtual #6 (Constructor#newInstance)
		0x57, // pc5 pop
		0xb1, // pc6 return
	}
	m := classfile.Method{Name: "m", Descriptor: "(Ljava/lang/reflect/Constructor;)V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 1}

	facts, err := dataflow.Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	site := NewInstanceSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "(Ljava/lang/reflect/Constructor;)V", PC: 2}
	ri, _ := resolveOneNewInstance(hierarchy.NewForTesting(nil), "test.Caller", site, facts)
	if ri.OwnerFQCN != "" {
		t.Errorf("OwnerFQCN = %q, want empty", ri.OwnerFQCN)
	}
	if ri.UnresolvedReason == "" {
		t.Errorf("UnresolvedReason is empty, want a non-empty explanation")
	}
}
