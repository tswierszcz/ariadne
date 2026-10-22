package dataflow

import (
	"testing"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

// calleePool is the shared tail every test below needs: a Methodref for
// test.Callee#take(Ljava/lang/String;)V, plus the String constants "a"
// and "b" ldc can reference. Index numbers are documented at each use
// site since classfile.PoolEntry is 1-indexed like the real pool.
func calleePool() []classfile.PoolEntry {
	return []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "a"},                                // 1
		{Tag: classfile.TagString, NameIndex: 1},                          // 2: "a"
		{Tag: classfile.TagUtf8, Str: "b"},                                // 3
		{Tag: classfile.TagString, NameIndex: 3},                          // 4: "b"
		{Tag: classfile.TagUtf8, Str: "test/Callee"},                      // 5
		{Tag: classfile.TagClass, NameIndex: 5},                           // 6
		{Tag: classfile.TagUtf8, Str: "take"},                             // 7
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;)V"},            // 8
		{Tag: classfile.TagNameAndType, NameIndex: 7, DescriptorIndex: 8}, // 9
		{Tag: classfile.TagMethodref, ClassIndex: 6, NameAndTypeIndex: 9}, // 10: test.Callee#take
	}
}

func takeArgSite(pc int) bytecodeir.Site {
	return bytecodeir.Site{Method: "test.Owner#m", Descriptor: "()V", PC: pc, Op: bytecodeir.OpCall}
}

// TestAnalyze_literalThroughVariable is DATA_FLOW D2's own first named
// example: `String s = "a"; take(s);` must resolve to the exact literal
// at the call, not widen it away just because it passed through a local.
//
// Bytecode:
//
//	pc=0: ldc #2 ("a")
//	pc=2: astore_1
//	pc=3: aload_1
//	pc=4: invokestatic #10 (test.Callee#take)
//	pc=7: return
func TestAnalyze_literalThroughVariable(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Owner", calleePool())
	code := []byte{
		0x12, 0x02, // pc0 ldc #2 ("a")
		0x4c,             // pc2 astore_1
		0x2b,             // pc3 aload_1
		0xb8, 0x00, 0x0a, // pc4 invokestatic #10
		0xb1, // pc7 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}

	facts, err := Analyze(cf, cf.ThisClass, m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	got := facts.ArgumentAt(takeArgSite(4), 0)
	want := valueflow.Const("java.lang.String", "a")
	if !got.Equal(want) {
		t.Errorf("ArgumentAt = %+v, want %+v", got, want)
	}
}

// TestAnalyze_twoBranchVariantsJoin is D2's own second named example: two
// different string literals assigned on an if/else's two arms must join
// into a KindSet of exactly those two candidates at the point they're
// both still live, not silently collapse to one or widen for no reason.
//
// Bytecode:
//
//	pc=0:  iconst_0                 fake condition
//	pc=1:  ifeq +9  -> pc=10        else arm
//	pc=4:  ldc #2 ("a")             then arm
//	pc=6:  astore_1
//	pc=7:  goto +6  -> pc=13
//	pc=10: ldc #4 ("b")             else arm
//	pc=12: astore_1
//	pc=13: aload_1                  join point
//	pc=14: invokestatic #10 (take)
//	pc=17: return
func TestAnalyze_twoBranchVariantsJoin(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Owner", calleePool())
	code := []byte{
		0x03,             // pc0 iconst_0
		0x99, 0x00, 0x09, // pc1 ifeq +9 -> pc10
		0x12, 0x02, // pc4 ldc #2 ("a")
		0x4c,             // pc6 astore_1
		0xa7, 0x00, 0x06, // pc7 goto +6 -> pc13
		0x12, 0x04, // pc10 ldc #4 ("b")
		0x4c,             // pc12 astore_1
		0x2b,             // pc13 aload_1
		0xb8, 0x00, 0x0a, // pc14 invokestatic #10
		0xb1, // pc17 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}

	facts, err := Analyze(cf, cf.ThisClass, m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	got := facts.ArgumentAt(takeArgSite(14), 0)
	if got.Kind != valueflow.KindSet {
		t.Fatalf("ArgumentAt = %+v, want KindSet", got)
	}
	a := valueflow.Const("java.lang.String", "a")
	b := valueflow.Const("java.lang.String", "b")
	if len(got.Candidates) != 2 ||
		!((got.Candidates[0].Equal(a) && got.Candidates[1].Equal(b)) ||
			(got.Candidates[0].Equal(b) && got.Candidates[1].Equal(a))) {
		t.Errorf("ArgumentAt.Candidates = %+v, want exactly {%+v, %+v}", got.Candidates, a, b)
	}
}

// TestAnalyze_unknownParameterStaysUnknown is D2's own third named
// example: `take(p)` where p is a parameter the caller could have passed
// anything as must stay KindTyped (known type, no known constant) or
// explicitly Unknown — never fabricate a specific string.
//
// Bytecode (instance method void m(String p)): p is slot 1 (slot 0 is
// `this`).
//
//	pc=0: aload_1
//	pc=1: invokestatic #10 (take)
//	pc=4: return
func TestAnalyze_unknownParameterStaysUnknown(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Owner", calleePool())
	code := []byte{
		0x2b,             // pc0 aload_1 (the parameter)
		0xb8, 0x00, 0x0a, // pc1 invokestatic #10
		0xb1, // pc4 return
	}
	m := classfile.Method{Name: "m", Descriptor: "(Ljava/lang/String;)V", Code: code, MaxLocals: 2}

	facts, err := Analyze(cf, cf.ThisClass, m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	got := facts.ArgumentAt(bytecodeir.Site{Method: "test.Owner#m", Descriptor: "(Ljava/lang/String;)V", PC: 1, Op: bytecodeir.OpCall}, 0)
	if got.Kind != valueflow.KindTyped || got.Type != "Ljava/lang/String;" {
		t.Errorf("ArgumentAt = %+v, want KindTyped(\"Ljava/lang/String;\")", got)
	}
}

// TestAnalyze_jsrRetAborts confirms the analysis reports Aborted rather
// than fabricating facts or panicking on the one instruction family this
// increment deliberately does not model (same limitation DATA_FLOW D1's
// CFG.Unsupported already declared).
func TestAnalyze_jsrRetAborts(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Owner", calleePool())
	code := []byte{
		0xa8, 0x00, 0x03, // pc0 jsr +3 -> pc3
		0x4c,       // pc3 astore_1
		0xa9, 0x01, // pc4 ret 1
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 2}

	facts, err := Analyze(cf, cf.ThisClass, m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); !aborted || reason == "" {
		t.Errorf("Aborted() = (%v, %q), want (true, non-empty)", aborted, reason)
	}
}

// TestAnalyze_noCodeIsNotAnError covers an abstract/native method (nil
// Code) — must produce empty, non-aborted Facts, not an error.
func TestAnalyze_noCodeIsNotAnError(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Owner", nil)
	m := classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic | classfile.AccAbstract}

	facts, err := Analyze(cf, cf.ThisClass, m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Errorf("Aborted() = (true, %q), want (false, \"\") for a Code-less method", reason)
	}
}
