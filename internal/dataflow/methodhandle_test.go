package dataflow

import (
	"testing"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

func sinkPool(extra ...classfile.PoolEntry) []classfile.PoolEntry {
	base := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                          // 1
		{Tag: classfile.TagClass, NameIndex: 1},                             // 2
		{Tag: classfile.TagUtf8, Str: "accept"},                             // 3
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/invoke/MethodHandle;)V"}, // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},   // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5},   // 6: test.Sink#accept(MethodHandle)V
	}
	return append(base, extra...)
}

// TestAnalyze_lookupFindStaticProducesHandleValue is DYNAMIC_LINKAGE
// L2's own "stały cel" example: Lookup.findStatic(Foo.class, "bar", mt)
// with a literal Class and method name must resolve to a real
// HandleValue, read back out by passing it straight to a second call.
//
// Bytecode (static void m(Lookup lookup, MethodType mt)):
//
//	pc0:  aload_0           (lookup)
//	pc1:  ldc #7            (test.Foo.class)
//	pc3:  ldc #9            ("bar")
//	pc5:  aload_1           (mt)
//	pc6:  invokevirtual #13 (Lookup#findStatic(Class,String,MethodType)MethodHandle)
//	pc9:  invokestatic #6   (test.Sink#accept(MethodHandle)V)
//	pc12: return
func TestAnalyze_lookupFindStaticProducesHandleValue(t *testing.T) {
	pool := sinkPool(
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "test/Foo"},                                                                                          // 7
		classfile.PoolEntry{Tag: classfile.TagClass, NameIndex: 7},                                                                                            // 8
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "bar"},                                                                                               // 9
		classfile.PoolEntry{Tag: classfile.TagString, NameIndex: 9},                                                                                           // 10
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "java/lang/invoke/MethodHandles$Lookup"},                                                             // 11
		classfile.PoolEntry{Tag: classfile.TagClass, NameIndex: 11},                                                                                           // 12
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "findStatic"},                                                                                        // 13
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Class;Ljava/lang/String;Ljava/lang/invoke/MethodType;)Ljava/lang/invoke/MethodHandle;"}, // 14
		classfile.PoolEntry{Tag: classfile.TagNameAndType, NameIndex: 13, DescriptorIndex: 14},                                                                // 15
		classfile.PoolEntry{Tag: classfile.TagMethodref, ClassIndex: 12, NameAndTypeIndex: 15},                                                                // 16: Lookup#findStatic
	)
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x2a,       // pc0 aload_0 (lookup)
		0x12, 0x08, // pc1 ldc #8 (test.Foo.class)
		0x12, 0x0a, // pc3 ldc #10 ("bar")
		0x2b,             // pc5 aload_1 (mt)
		0xb6, 0x00, 0x10, // pc6 invokevirtual #16 (findStatic)
		0xb8, 0x00, 0x06, // pc9 invokestatic #6 (test.Sink#accept)
		0xb1, // pc12 return
	}
	desc := "(Ljava/lang/invoke/MethodHandles$Lookup;Ljava/lang/invoke/MethodType;)V"
	m := classfile.Method{Name: "m", Descriptor: desc, AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	val := facts.ArgumentAt(bytecodeir.Site{Method: "test.Caller#m", Descriptor: desc, PC: 9, Op: bytecodeir.OpCall}, 0)
	if val.Kind != valueflow.KindHandle {
		t.Fatalf("ArgumentAt = %+v, want KindHandle", val)
	}
	hv := val.Handle
	if hv.RequestedOwner != "test.Foo" || hv.Name != "bar" {
		t.Errorf("RequestedOwner/Name = %q/%q, want test.Foo/bar", hv.RequestedOwner, hv.Name)
	}
	if hv.RefKind != classfile.RefInvokeStatic {
		t.Errorf("RefKind = %d, want RefInvokeStatic", hv.RefKind)
	}
}

// TestAnalyze_lookupFindVirtualWithUnknownNameStaysUnknown is L2's own
// "nieznany lookup" example: a non-literal method-name argument must
// never be guessed at — the find* call's own return value stays
// KindUnknown, read back out the same way as the positive test.
//
// Bytecode (static void m(Lookup lookup, String name, MethodType mt)):
//
//	pc0:  aload_0           (lookup)
//	pc1:  ldc #8            (test.Foo.class)
//	pc3:  aload_1           (name — a parameter, not a literal)
//	pc4:  aload_2           (mt)
//	pc5:  invokevirtual #16 (Lookup#findVirtual(Class,String,MethodType)MethodHandle)
//	pc8:  invokestatic #6   (test.Sink#accept(MethodHandle)V)
//	pc11: return
func TestAnalyze_lookupFindVirtualWithUnknownNameStaysUnknown(t *testing.T) {
	pool := sinkPool(
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "test/Foo"},                                                                                          // 7
		classfile.PoolEntry{Tag: classfile.TagClass, NameIndex: 7},                                                                                            // 8
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "java/lang/invoke/MethodHandles$Lookup"},                                                             // 9
		classfile.PoolEntry{Tag: classfile.TagClass, NameIndex: 9},                                                                                            // 10
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "findVirtual"},                                                                                       // 11
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Class;Ljava/lang/String;Ljava/lang/invoke/MethodType;)Ljava/lang/invoke/MethodHandle;"}, // 12
		classfile.PoolEntry{Tag: classfile.TagNameAndType, NameIndex: 11, DescriptorIndex: 12},                                                                // 13
		classfile.PoolEntry{Tag: classfile.TagMethodref, ClassIndex: 10, NameAndTypeIndex: 13},                                                                // 14: Lookup#findVirtual
	)
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x2a,       // pc0 aload_0 (lookup)
		0x12, 0x08, // pc1 ldc #8 (test.Foo.class)
		0x2b,             // pc3 aload_1 (name, a parameter)
		0x2c,             // pc4 aload_2 (mt)
		0xb6, 0x00, 0x0e, // pc5 invokevirtual #14 (findVirtual)
		0xb8, 0x00, 0x06, // pc8 invokestatic #6 (test.Sink#accept)
		0xb1, // pc11 return
	}
	desc := "(Ljava/lang/invoke/MethodHandles$Lookup;Ljava/lang/String;Ljava/lang/invoke/MethodType;)V"
	m := classfile.Method{Name: "m", Descriptor: desc, AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 3}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	val := facts.ArgumentAt(bytecodeir.Site{Method: "test.Caller#m", Descriptor: desc, PC: 8, Op: bytecodeir.OpCall}, 0)
	if val.Kind != valueflow.KindUnknown {
		t.Fatalf("ArgumentAt = %+v, want KindUnknown", val)
	}
}

// TestAnalyze_receiverAtReadsBackHandleValue confirms Facts.ReceiverAt
// surfaces the real receiver — here, a HandleValue flowing from a
// `findStatic` call straight into `invokeExact`'s own receiver slot —
// DYNAMIC_LINKAGE L2's own prerequisite for connecting invoke/
// invokeExact to the handle it was called on.
//
// Bytecode (static void m(Lookup lookup, MethodType mt)):
//
//	pc0:  aload_0           (lookup)
//	pc1:  ldc #8            (test.Foo.class)
//	pc3:  ldc #10           ("bar")
//	pc5:  aload_1           (mt)
//	pc6:  invokevirtual #16 (Lookup#findStatic(Class,String,MethodType)MethodHandle)
//	pc9:  invokevirtual #19 (MethodHandle#invokeExact([Ljava/lang/Object;)V, varargs-free for this test)
//	pc12: return
func TestAnalyze_receiverAtReadsBackHandleValue(t *testing.T) {
	pool := sinkPool(
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "test/Foo"},                                                                                          // 7
		classfile.PoolEntry{Tag: classfile.TagClass, NameIndex: 7},                                                                                            // 8
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "bar"},                                                                                               // 9
		classfile.PoolEntry{Tag: classfile.TagString, NameIndex: 9},                                                                                           // 10
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "java/lang/invoke/MethodHandles$Lookup"},                                                             // 11
		classfile.PoolEntry{Tag: classfile.TagClass, NameIndex: 11},                                                                                           // 12
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "findStatic"},                                                                                        // 13
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Class;Ljava/lang/String;Ljava/lang/invoke/MethodType;)Ljava/lang/invoke/MethodHandle;"}, // 14
		classfile.PoolEntry{Tag: classfile.TagNameAndType, NameIndex: 13, DescriptorIndex: 14},                                                                // 15
		classfile.PoolEntry{Tag: classfile.TagMethodref, ClassIndex: 12, NameAndTypeIndex: 15},                                                                // 16: Lookup#findStatic
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "java/lang/invoke/MethodHandle"},                                                                     // 17
		classfile.PoolEntry{Tag: classfile.TagClass, NameIndex: 17},                                                                                           // 18
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "invokeExact"},                                                                                       // 19
		classfile.PoolEntry{Tag: classfile.TagUtf8, Str: "()V"},                                                                                               // 20
		classfile.PoolEntry{Tag: classfile.TagNameAndType, NameIndex: 19, DescriptorIndex: 20},                                                                // 21
		classfile.PoolEntry{Tag: classfile.TagMethodref, ClassIndex: 18, NameAndTypeIndex: 21},                                                                // 22: MethodHandle#invokeExact()V
	)
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x2a,       // pc0 aload_0 (lookup)
		0x12, 0x08, // pc1 ldc #8 (test.Foo.class)
		0x12, 0x0a, // pc3 ldc #10 ("bar")
		0x2b,             // pc5 aload_1 (mt)
		0xb6, 0x00, 0x10, // pc6 invokevirtual #16 (findStatic)
		0xb6, 0x00, 0x16, // pc9 invokevirtual #22 (invokeExact)
		0xb1, // pc12 return
	}
	desc := "(Ljava/lang/invoke/MethodHandles$Lookup;Ljava/lang/invoke/MethodType;)V"
	m := classfile.Method{Name: "m", Descriptor: desc, AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	receiver := facts.ReceiverAt(bytecodeir.Site{Method: "test.Caller#m", Descriptor: desc, PC: 9, Op: bytecodeir.OpCall})
	if receiver.Kind != valueflow.KindHandle {
		t.Fatalf("ReceiverAt = %+v, want KindHandle", receiver)
	}
	if receiver.Handle.RequestedOwner != "test.Foo" || receiver.Handle.Name != "bar" {
		t.Errorf("Handle = %s#%s, want test.Foo#bar", receiver.Handle.RequestedOwner, receiver.Handle.Name)
	}
}
