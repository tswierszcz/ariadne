package methodhandle

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
)

// TestResolveOne_handleReceiverResolves is L2's own positive example,
// mirroring internal/dataflow's own
// TestAnalyze_receiverAtReadsBackHandleValue but exercised through this
// package's own resolveOne.
//
// Bytecode (static void m(Lookup, MethodType)): findStatic then
// invokeExact on the result.
func TestResolveOne_handleReceiverResolves(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Foo"},                              // 1
		{Tag: classfile.TagClass, NameIndex: 1},                                // 2
		{Tag: classfile.TagUtf8, Str: "bar"},                                   // 3
		{Tag: classfile.TagString, NameIndex: 3},                               // 4
		{Tag: classfile.TagUtf8, Str: "java/lang/invoke/MethodHandles$Lookup"}, // 5
		{Tag: classfile.TagClass, NameIndex: 5},                                // 6
		{Tag: classfile.TagUtf8, Str: "findStatic"},                            // 7
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Class;Ljava/lang/String;Ljava/lang/invoke/MethodType;)Ljava/lang/invoke/MethodHandle;"}, // 8
		{Tag: classfile.TagNameAndType, NameIndex: 7, DescriptorIndex: 8},                                                                  // 9
		{Tag: classfile.TagMethodref, ClassIndex: 6, NameAndTypeIndex: 9},                                                                  // 10
		{Tag: classfile.TagUtf8, Str: "java/lang/invoke/MethodHandle"},                                                                     // 11
		{Tag: classfile.TagClass, NameIndex: 11},                                                                                           // 12
		{Tag: classfile.TagUtf8, Str: "invokeExact"},                                                                                       // 13
		{Tag: classfile.TagUtf8, Str: "()V"},                                                                                               // 14
		{Tag: classfile.TagNameAndType, NameIndex: 13, DescriptorIndex: 14},                                                                // 15
		{Tag: classfile.TagMethodref, ClassIndex: 12, NameAndTypeIndex: 15},                                                                // 16
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x2a,       // pc0 aload_0 (lookup)
		0x12, 0x02, // pc1 ldc #2 (test.Foo.class)
		0x12, 0x04, // pc3 ldc #4 ("bar")
		0x2b,             // pc5 aload_1 (mt)
		0xb6, 0x00, 0x0a, // pc6 invokevirtual #10 (findStatic)
		0xb6, 0x00, 0x10, // pc9 invokevirtual #16 (invokeExact)
		0xb1, // pc12 return
	}
	desc := "(Ljava/lang/invoke/MethodHandles$Lookup;Ljava/lang/invoke/MethodType;)V"
	m := classfile.Method{Name: "m", Descriptor: desc, AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}

	facts, err := dataflow.Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("dataflow.Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	site := InvokeSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: desc, PC: 9, Exact: true}
	ri, _ := resolveOne(hierarchy.NewForTesting(nil), "test.Caller", site, facts)
	if ri.RequestedOwner != "test.Foo" || ri.Method != "bar" {
		t.Errorf("Requested = %s#%s, want test.Foo#bar", ri.RequestedOwner, ri.Method)
	}
	if ri.UnresolvedReason != "" {
		t.Errorf("UnresolvedReason = %q, want empty", ri.UnresolvedReason)
	}
}

// TestResolveOne_fieldReceiverFallsThroughToResolveStaticField confirms
// resolveOne dispatches a KindField receiver to
// dataflow.ResolveStaticField rather than treating it as a plain
// unresolved value — exercised here against a field whose owner isn't
// in the hierarchy at all, so it stays honestly unresolved without
// needing real jar bytes (ResolveStaticField's own early-exit path,
// before any file I/O). The positive, same-class-writer case is
// covered at the dataflow layer (TestMergeFieldWriters_*) and verified
// end-to-end against real petclinic/WebGoat bytecode (see
// RECOMMENDATION's own status record).
func TestResolveOne_fieldReceiverFallsThroughToResolveStaticField(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Caller"},                       // 1
		{Tag: classfile.TagClass, NameIndex: 1},                            // 2
		{Tag: classfile.TagUtf8, Str: "HANDLE"},                            // 3
		{Tag: classfile.TagUtf8, Str: "Ljava/lang/invoke/MethodHandle;"},   // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},  // 5
		{Tag: classfile.TagFieldref, ClassIndex: 2, NameAndTypeIndex: 5},   // 6: test.Caller#HANDLE
		{Tag: classfile.TagUtf8, Str: "java/lang/invoke/MethodHandle"},     // 7
		{Tag: classfile.TagClass, NameIndex: 7},                            // 8
		{Tag: classfile.TagUtf8, Str: "invokeExact"},                       // 9
		{Tag: classfile.TagUtf8, Str: "()V"},                               // 10
		{Tag: classfile.TagNameAndType, NameIndex: 9, DescriptorIndex: 10}, // 11
		{Tag: classfile.TagMethodref, ClassIndex: 8, NameAndTypeIndex: 11}, // 12
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0xb2, 0x00, 0x06, // pc0 getstatic #6 (HANDLE)
		0xb6, 0x00, 0x0c, // pc3 invokevirtual #12 (invokeExact)
		0xb1, // pc6 return
	}
	m := classfile.Method{Name: "invokeHandle", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	facts, err := dataflow.Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("dataflow.Analyze: %v", err)
	}

	// An empty hierarchy: test.Caller#HANDLE cannot resolve, so
	// ResolveStaticField must return KindUnknown without touching any
	// jar, and resolveOne must surface that as UnresolvedReason, not
	// panic or silently default to something else.
	h := hierarchy.NewForTesting(nil)
	site := InvokeSite{CallerFQCN: "test.Caller", CallerName: "invokeHandle", CallerDesc: "()V", PC: 3, Exact: true}
	ri, _ := resolveOne(h, "test.Caller", site, facts)
	if ri.RequestedOwner != "" {
		t.Errorf("RequestedOwner = %q, want empty", ri.RequestedOwner)
	}
	if ri.UnresolvedReason == "" {
		t.Errorf("UnresolvedReason is empty, want a non-empty explanation")
	}
}
