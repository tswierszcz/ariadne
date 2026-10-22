package dataflow

import (
	"testing"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

// TestAnalyze_lambdaProducesFunctionalValue is DYNAMIC_LINKAGE L0/L1's
// own first example: an invokedynamic resolved through a
// LambdaMetafactory bootstrap must produce a real valueflow.Functional
// value (not a generic Unknown) at the call it's passed to.
//
// Bytecode (static void m()):
//
//	pc0: invokedynamic #4 (run()Ljava/lang/Runnable;) -> pushes the lambda
//	pc5: invokestatic #10 (test.Sink#accept(Runnable)V)
//	pc8: return
func TestAnalyze_lambdaProducesFunctionalValue(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "run"},                              // 1
		{Tag: classfile.TagUtf8, Str: "()Ljava/lang/Runnable;"},           // 2
		{Tag: classfile.TagNameAndType, NameIndex: 1, DescriptorIndex: 2}, // 3
		{Tag: classfile.TagInvokeDynamic, NameAndTypeIndex: 3},            // 4 (bootstrap attr index 0)
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                        // 5
		{Tag: classfile.TagClass, NameIndex: 5},                           // 6
		{Tag: classfile.TagUtf8, Str: "accept"},                           // 7
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Runnable;)V"},          // 8
		{Tag: classfile.TagNameAndType, NameIndex: 7, DescriptorIndex: 8}, // 9
		{Tag: classfile.TagMethodref, ClassIndex: 6, NameAndTypeIndex: 9}, // 10
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	cf.Bootstrap = []classfile.BootstrapMethod{
		{
			Kind: classfile.RefInvokeStatic, Owner: "java.lang.invoke.LambdaMetafactory", Name: "metafactory",
			Args: []classfile.BootstrapArg{
				{},
				{IsMethodHandle: true, Kind: classfile.RefInvokeStatic, Owner: "test.Caller", Name: "lambda$m$0", Desc: "()V"},
				{},
			},
		},
	}
	code := []byte{
		0xba, 0x00, 0x04, 0x00, 0x00, // pc0 invokedynamic #4
		0xb8, 0x00, 0x0a, // pc5 invokestatic #10 (test.Sink#accept(Runnable)V)
		0xb1, // pc8 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	val := facts.ArgumentAt(bytecodeir.Site{Method: "test.Caller#m", Descriptor: "()V", PC: 5, Op: bytecodeir.OpCall}, 0)
	if val.Kind != valueflow.KindFunctional {
		t.Fatalf("ArgumentAt = %+v, want KindFunctional", val)
	}
	fv := val.Func
	if fv.SamInterface != "java.lang.Runnable" || fv.SamMethod != "run" {
		t.Errorf("SamInterface/SamMethod = %q/%q, want java.lang.Runnable/run", fv.SamInterface, fv.SamMethod)
	}
	if fv.ImplOwner != "test.Caller" || fv.ImplName != "lambda$m$0" || fv.ImplDesc != "()V" {
		t.Errorf("Impl = %s#%s%s, want test.Caller#lambda$m$0()V", fv.ImplOwner, fv.ImplName, fv.ImplDesc)
	}
	if fv.RefKind != classfile.RefInvokeStatic {
		t.Errorf("RefKind = %d, want RefInvokeStatic", fv.RefKind)
	}
	if len(fv.Captured) != 0 {
		t.Errorf("Captured = %+v, want none (no free variables)", fv.Captured)
	}
}

// TestAnalyze_lambdaCapturesFreeVariable confirms a captured free
// variable (here, a literal passed through a local before the
// invokedynamic) ends up in FunctionalValue.Captured with its own
// already-resolved Value, not re-widened to Unknown.
//
// Bytecode (static void m()):
//
//	pc0: ldc #11 ("x")
//	pc2: invokedynamic #4 ((Ljava/lang/String;)Ljava/lang/Runnable;) -> captures "x"
//	pc7: invokestatic #10 (test.Sink#accept(Runnable)V)
//	pc10: return
func TestAnalyze_lambdaCapturesFreeVariable(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "run"},                                      // 1
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;)Ljava/lang/Runnable;"}, // 2
		{Tag: classfile.TagNameAndType, NameIndex: 1, DescriptorIndex: 2},         // 3
		{Tag: classfile.TagInvokeDynamic, NameAndTypeIndex: 3},                    // 4
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                                // 5
		{Tag: classfile.TagClass, NameIndex: 5},                                   // 6
		{Tag: classfile.TagUtf8, Str: "accept"},                                   // 7
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Runnable;)V"},                  // 8
		{Tag: classfile.TagNameAndType, NameIndex: 7, DescriptorIndex: 8},         // 9
		{Tag: classfile.TagMethodref, ClassIndex: 6, NameAndTypeIndex: 9},         // 10
		{Tag: classfile.TagUtf8, Str: "x"},                                        // 11
		{Tag: classfile.TagString, NameIndex: 11},                                 // 12
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	cf.Bootstrap = []classfile.BootstrapMethod{
		{
			Kind: classfile.RefInvokeStatic, Owner: "java.lang.invoke.LambdaMetafactory", Name: "metafactory",
			Args: []classfile.BootstrapArg{
				{},
				{IsMethodHandle: true, Kind: classfile.RefInvokeStatic, Owner: "test.Caller", Name: "lambda$m$0", Desc: "(Ljava/lang/String;)V"},
				{},
			},
		},
	}
	code := []byte{
		0x12, 0x0c, // pc0 ldc #12 ("x")
		0xba, 0x00, 0x04, 0x00, 0x00, // pc2 invokedynamic #4
		0xb8, 0x00, 0x0a, // pc7 invokestatic #10
		0xb1, // pc10 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	val := facts.ArgumentAt(bytecodeir.Site{Method: "test.Caller#m", Descriptor: "()V", PC: 7, Op: bytecodeir.OpCall}, 0)
	if val.Kind != valueflow.KindFunctional {
		t.Fatalf("ArgumentAt = %+v, want KindFunctional", val)
	}
	if len(val.Func.Captured) != 1 {
		t.Fatalf("Captured = %+v, want exactly 1", val.Func.Captured)
	}
	want := valueflow.Const("java.lang.String", "x")
	if !val.Func.Captured[0].Equal(want) {
		t.Errorf("Captured[0] = %+v, want %+v", val.Func.Captured[0], want)
	}
}
