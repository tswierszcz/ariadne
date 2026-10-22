package callback

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
)

// TestResolveOne_directAllocationResolves is A1's own positive example:
// `new Thread(new MyRunnable())` must resolve the Thread constructor's
// callback argument to test.MyRunnable, not just "some Runnable".
//
// Bytecode (static void m()):
//
//	pc0:  new #2 (java.lang.Thread)
//	pc3:  dup
//	pc4:  new #4 (test.MyRunnable)
//	pc7:  dup
//	pc8:  invokespecial #8 (test.MyRunnable#<init>()V)
//	pc11: invokespecial #11 (java.lang.Thread#<init>(Runnable)V)
//	pc14: pop
//	pc15: return
func TestResolveOne_directAllocationResolves(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "java/lang/Thread"},                  // 1
		{Tag: classfile.TagClass, NameIndex: 1},                            // 2
		{Tag: classfile.TagUtf8, Str: "test/MyRunnable"},                   // 3
		{Tag: classfile.TagClass, NameIndex: 3},                            // 4
		{Tag: classfile.TagUtf8, Str: "<init>"},                            // 5
		{Tag: classfile.TagUtf8, Str: "()V"},                               // 6
		{Tag: classfile.TagNameAndType, NameIndex: 5, DescriptorIndex: 6},  // 7
		{Tag: classfile.TagMethodref, ClassIndex: 4, NameAndTypeIndex: 7},  // 8: MyRunnable#<init>()V
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Runnable;)V"},           // 9
		{Tag: classfile.TagNameAndType, NameIndex: 5, DescriptorIndex: 9},  // 10
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 10}, // 11: Thread#<init>(Runnable)V
	}
	cf := classfile.NewClassFileForTesting("test.Caller", pool)
	code := []byte{
		0xbb, 0x00, 0x02, // pc0  new #2 (Thread)
		0x59,             // pc3  dup
		0xbb, 0x00, 0x04, // pc4  new #4 (MyRunnable)
		0x59,             // pc7  dup
		0xb7, 0x00, 0x08, // pc8  invokespecial #8 (MyRunnable#<init>()V)
		0xb7, 0x00, 0x0b, // pc11 invokespecial #11 (Thread#<init>(Runnable)V)
		0x57, // pc14 pop
		0xb1, // pc15 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	facts, err := dataflow.Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("dataflow.Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	model, ok := Lookup(fqcnThread, "<init>", "(Ljava/lang/Runnable;)V")
	if !ok {
		t.Fatal("DefaultCatalog is missing Thread(Runnable)")
	}
	site := RegistrationSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 11, Model: model, Outcome: OutcomeCataloged}

	rc := resolveOne("test.Caller", site, facts)
	if rc.TargetFQCN != "test.MyRunnable" {
		t.Errorf("TargetFQCN = %q, want test.MyRunnable", rc.TargetFQCN)
	}
	if rc.UnresolvedReason != "" {
		t.Errorf("UnresolvedReason = %q, want empty", rc.UnresolvedReason)
	}
}

// TestResolveOne_parameterStaysUnresolved is A1's own negative example:
// a Runnable passed straight through as a parameter must never be
// fabricated into a concrete class.
//
// Bytecode (static void m(Executor, Runnable)): slot0=Executor,
// slot1=Runnable.
//
//	pc0: aload_0
//	pc1: aload_1
//	pc2: invokevirtual #6 (Executor#execute(Runnable)V)
//	pc5: return
func TestResolveOne_parameterStaysUnresolved(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "java/util/concurrent/Executor"},    // 1
		{Tag: classfile.TagClass, NameIndex: 1},                           // 2
		{Tag: classfile.TagUtf8, Str: "execute"},                          // 3
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Runnable;)V"},          // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4}, // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5}, // 6
	}
	cf := classfile.NewClassFileForTesting("test.Caller2", pool)
	code := []byte{
		0x2a,             // pc0 aload_0
		0x2b,             // pc1 aload_1
		0xb6, 0x00, 0x06, // pc2 invokevirtual #6
		0xb1, // pc5 return
	}
	desc := "(Ljava/util/concurrent/Executor;Ljava/lang/Runnable;)V"
	m := classfile.Method{Name: "m", Descriptor: desc, AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}

	facts, err := dataflow.Analyze(cf, "test.Caller2", m)
	if err != nil {
		t.Fatalf("dataflow.Analyze: %v", err)
	}

	model, ok := Lookup(fqcnExecutor, "execute", "(Ljava/lang/Runnable;)V")
	if !ok {
		t.Fatal("DefaultCatalog is missing Executor#execute")
	}
	site := RegistrationSite{CallerFQCN: "test.Caller2", CallerName: "m", CallerDesc: desc, PC: 2, Model: model, Outcome: OutcomeCataloged}

	rc := resolveOne("test.Caller2", site, facts)
	if rc.TargetFQCN != "" {
		t.Errorf("TargetFQCN = %q, want empty (the argument was an unresolved parameter)", rc.TargetFQCN)
	}
	if rc.UnresolvedReason == "" {
		t.Errorf("UnresolvedReason is empty, want a non-empty explanation")
	}
}

// TestResolveOne_lambdaResolvesToImplementationMethod is A1's own
// DYNAMIC_LINKAGE L0/L1 extension: a lambda passed directly as the
// callback argument resolves straight to its own implementation
// method, not to a concrete class.
//
// Bytecode (static void m()): `new Thread(() -> {})`.
//
//	pc0:  new #2 (java.lang.Thread)
//	pc3:  dup
//	pc4:  invokedynamic #10 (run()Ljava/lang/Runnable;)
//	pc9:  invokespecial #6 (Thread#<init>(Runnable)V)
//	pc12: pop
//	pc13: return
func TestResolveOne_lambdaResolvesToImplementationMethod(t *testing.T) {
	pool := []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "java/lang/Thread"},                 // 1
		{Tag: classfile.TagClass, NameIndex: 1},                           // 2
		{Tag: classfile.TagUtf8, Str: "<init>"},                           // 3
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Runnable;)V"},          // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4}, // 5
		{Tag: classfile.TagMethodref, ClassIndex: 2, NameAndTypeIndex: 5}, // 6: Thread#<init>(Runnable)V
		{Tag: classfile.TagUtf8, Str: "run"},                              // 7
		{Tag: classfile.TagUtf8, Str: "()Ljava/lang/Runnable;"},           // 8
		{Tag: classfile.TagNameAndType, NameIndex: 7, DescriptorIndex: 8}, // 9
		{Tag: classfile.TagInvokeDynamic, NameAndTypeIndex: 9},            // 10 (bootstrap attr index 0)
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
		0xbb, 0x00, 0x02, // pc0 new #2 (Thread)
		0x59,                         // pc3 dup
		0xba, 0x00, 0x0a, 0x00, 0x00, // pc4 invokedynamic #10
		0xb7, 0x00, 0x06, // pc9 invokespecial #6 (Thread#<init>(Runnable)V)
		0x57, // pc12 pop
		0xb1, // pc13 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 0}

	facts, err := dataflow.Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("dataflow.Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	model, ok := Lookup(fqcnThread, "<init>", "(Ljava/lang/Runnable;)V")
	if !ok {
		t.Fatal("DefaultCatalog is missing Thread(Runnable)")
	}
	site := RegistrationSite{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 9, Model: model, Outcome: OutcomeCataloged}

	rc := resolveOne("test.Caller", site, facts)
	if rc.TargetFQCN != "" {
		t.Errorf("TargetFQCN = %q, want empty (this is a lambda, not a `new`)", rc.TargetFQCN)
	}
	if rc.ImplOwner != "test.Caller" || rc.ImplName != "lambda$m$0" || rc.ImplDesc != "()V" {
		t.Errorf("Impl = %s#%s%s, want test.Caller#lambda$m$0()V", rc.ImplOwner, rc.ImplName, rc.ImplDesc)
	}
	if rc.UnresolvedReason != "" {
		t.Errorf("UnresolvedReason = %q, want empty", rc.UnresolvedReason)
	}
}
