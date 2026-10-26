package dataflow

import (
	"testing"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// TestAnalyze_staticCallResultProducesKindCallResult is DATA_FLOW D3's
// call-result increment, first example: an invokestatic call's return
// value must identify *which* method was called, same as a field read
// already identifies which field, read back out by passing it straight
// to a sink call (same pattern as the field/handle tests).
//
// Bytecode (static void m()):
//
//	pc0: invokestatic #6  (test.Caller#target()Ljava/lang/String;)
//	pc3: invokestatic #12 (test.Sink#accept(String)V)
//	pc6: return
func TestAnalyze_staticCallResultProducesKindCallResult(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Caller"},                       // 1
		{Tag: classfile.TagClass, NameIndex: 1},                            // 2
		{Tag: classfile.TagUtf8, Str: "target"},                            // 3
		{Tag: classfile.TagUtf8, Str: "()Ljava/lang/String;"},              // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},  // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5},  // 6: test.Caller#target()Ljava/lang/String;
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                         // 7
		{Tag: classfile.TagClass, NameIndex: 7},                            // 8
		{Tag: classfile.TagUtf8, Str: "accept"},                            // 9
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;)V"},             // 10
		{Tag: classfile.TagNameAndType, NameIndex: 9, DescriptorIndex: 10}, // 11
		{Tag: classfile.TagMethodref, ClassIndex: 8, NameAndTypeIndex: 11}, // 12: test.Sink#accept(String)V
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0xb8, 0x00, 0x06, // pc0 invokestatic #6 (target)
		0xb8, 0x00, 0x0c, // pc3 invokestatic #12 (Sink#accept)
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
	if val.Kind != valueflow.KindCallResult {
		t.Fatalf("ArgumentAt = %+v, want KindCallResult", val)
	}
	want := valueflow.CallResultRef{
		Owner: "test.Caller", Name: "target", Desc: "()Ljava/lang/String;",
		Site: bytecodeir.Site{Method: "test.Caller#m", Descriptor: "()V", PC: 0, Op: bytecodeir.OpCall},
	}
	if *val.Call != want {
		t.Errorf("Call = %+v, want %+v", *val.Call, want)
	}
}

func TestAnalyze_preservesBoundedVirtualReceiverTrace(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "java/util/stream/Stream"},            // 1
		{Tag: classfile.TagClass, NameIndex: 1},                             // 2
		{Tag: classfile.TagUtf8, Str: "first"},                              // 3
		{Tag: classfile.TagUtf8, Str: "()Ljava/util/stream/Stream;"},        // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},   // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5},   // 6
		{Tag: classfile.TagUtf8, Str: "second"},                             // 7
		{Tag: classfile.TagNameAndType, NameIndex: 7, DescriptorIndex: 4},   // 8
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 8},   // 9
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                          // 10
		{Tag: classfile.TagClass, NameIndex: 10},                            // 11
		{Tag: classfile.TagUtf8, Str: "accept"},                             // 12
		{Tag: classfile.TagUtf8, Str: "(Ljava/util/stream/Stream;)V"},       // 13
		{Tag: classfile.TagNameAndType, NameIndex: 12, DescriptorIndex: 13}, // 14
		{Tag: classfile.TagMethodref, ClassIndex: 11, NameAndTypeIndex: 14}, // 15
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0x01,             // pc0 aconst_null
		0xb6, 0x00, 0x06, // pc1 invokevirtual Fluent#first
		0x4b,             // pc4 astore_0
		0x2a,             // pc5 aload_0
		0xb6, 0x00, 0x09, // pc6 invokevirtual Fluent#second
		0x4c,             // pc9 astore_1
		0x2b,             // pc10 aload_1
		0xb8, 0x00, 0x0f, // pc11 invokestatic Sink#accept
		0xb1, // pc14 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}
	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}
	arg := facts.ArgumentAt(bytecodeir.Site{Method: "test.Caller#m", Descriptor: "()V", PC: 11, Op: bytecodeir.OpCall}, 0)
	if arg.Kind != valueflow.KindCallTrace || arg.CallTrace == nil {
		t.Fatalf("sink argument=%+v, want virtual-call trace", arg)
	}
	if arg.CallTrace.Site.PC != 6 || arg.CallTrace.Receiver == nil || arg.CallTrace.Receiver.Kind != valueflow.KindCallTrace || arg.CallTrace.Receiver.CallTrace.Site.PC != 1 {
		t.Fatalf("trace=%+v, want second@6 -> first@1", arg.CallTrace)
	}
	if got := valueflow.CallTraceDepth(arg); got != 2 {
		t.Fatalf("trace depth=%d, want 2", got)
	}
}

// targetReturningConst builds a hierarchy.Method for a static method
// named name that always returns the string literal "x" — reused by
// several tests below.
func targetReturningConst(name string) hierarchy.Method {
	code := []byte{
		0x12, 0x04, // pc0 ldc #4 ("x")
		0xb0, // pc2 areturn
	}
	return hierarchy.Method{
		Method: classfile.Method{Name: name, Descriptor: "()Ljava/lang/String;", AccessFlags: classfile.AccStatic, Code: code},
	}
}

func constStringPool() []classfile.PoolEntry {
	return []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Caller"}, // 1
		{Tag: classfile.TagClass, NameIndex: 1},      // 2
		{Tag: classfile.TagUtf8, Str: "x"},           // 3
		{Tag: classfile.TagString, NameIndex: 3},     // 4
	}
}

func TestMergeReturns_singleReturnResolves(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", constStringPool())
	m := targetReturningConst("target")

	v, warnings := mergeReturns(cf, "test.Caller", m)
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	want := valueflow.Const("java.lang.String", "x")
	if !v.Equal(want) {
		t.Errorf("mergeReturns = %+v, want %+v", v, want)
	}
}

func TestMergeReturns_voidMethodIsHonestUnknown(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", constStringPool())
	m := hierarchy.Method{Method: classfile.Method{Name: "m", Descriptor: "()V", Code: []byte{0xb1}}}

	v, _ := mergeReturns(cf, "test.Caller", m)
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("mergeReturns(void method) = %+v, want KindUnknown", v)
	}
}

func TestResolveCallResult_unresolvableOwnerIsUnknown(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	v, _ := ResolveCallResult(h, valueflow.CallResultRef{Owner: "test.NoSuchClass", Name: "target", Desc: "()Ljava/lang/String;"})
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("ResolveCallResult(unresolvable owner) = %+v, want KindUnknown", v)
	}
}

// TestChaseFieldOrCall_cycleDoesNotHang confirms the visited guard:
// revisiting a call-result ref already being resolved (mutual
// recursion through calls, or a field/call chain crossing back on
// itself) reports Unknown instead of recursing forever — DATA_FLOW D0's
// own "recursion converges or reveals a limit". Exercised directly
// against the unexported chaseFieldOrCall with a pre-seeded visited
// set, the same way other packages' cycle tests (e.g. JVM_GRAPH G3's
// TestInit_cycleDoesNotHang) avoid needing a real, compiled
// mutual-recursion fixture.
func TestChaseFieldOrCall_cycleDoesNotHang(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	ref := valueflow.CallResultRef{Owner: "test.Caller", Name: "target", Desc: "()Ljava/lang/String;"}
	visited := map[string]bool{callKey(ref): true}

	v, _ := chaseFieldOrCall(h, valueflow.CallResult(ref), visited)
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("chaseFieldOrCall(cycle) = %+v, want KindUnknown", v)
	}
}

// TestChaseFieldOrCall_fieldToCallResultChases confirms the asymmetry
// this increment closes: a field whose own resolved value is itself a
// KindCallResult (`private static final MethodHandle H = findIt();`)
// gets chased one more hop into ResolveCallResult, not left as an
// unresolved KindCallResult the caller has to know to chase itself.
// Both sides of the chase fail cheaply here (no SourceJar to read), but
// the point under test is that chaseFieldOrCall's switch actually
// dispatches KindCallResult into resolveCallResultOnce rather than
// returning it unchanged.
func TestChaseFieldOrCall_fieldToCallResultChases(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	v := valueflow.CallResult(valueflow.CallResultRef{Owner: "test.NoSuchClass", Name: "target", Desc: "()Ljava/lang/String;"})

	got, _ := chaseFieldOrCall(h, v, map[string]bool{})
	if got.Kind != valueflow.KindUnknown {
		t.Fatalf("chaseFieldOrCall(callresult) = %+v, want KindUnknown (chased, not passed through)", got)
	}
	if got.Reason == "" {
		t.Errorf("chaseFieldOrCall(callresult) Reason is empty, want resolveCallResultOnce's own unresolvable-owner message")
	}
}
