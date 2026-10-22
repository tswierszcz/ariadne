package methodhandle

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
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
	ri := resolveOne("test.Caller", site, facts)
	if ri.RequestedOwner != "test.Foo" || ri.Method != "bar" {
		t.Errorf("Requested = %s#%s, want test.Foo#bar", ri.RequestedOwner, ri.Method)
	}
	if ri.UnresolvedReason != "" {
		t.Errorf("UnresolvedReason = %q, want empty", ri.UnresolvedReason)
	}
}
