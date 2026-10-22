package classfile

import "testing"

// TestExtractCalls_initRefs is JVM_GRAPH G3's bytecode-decoding half,
// tested against a hand-assembled constant pool and Code array rather
// than a compiled fixture class: this environment has no JDK (javac is
// unavailable), but new/getstatic/putstatic/ldc are simple enough to
// assemble by hand and this is, if anything, a more precise test of the
// decoder itself than a compiled .class would be — it pins down exactly
// which bytes produce which InitRef, independent of what any particular
// javac version happens to emit.
func TestExtractCalls_initRefs(t *testing.T) {
	p := pool{
		{},                             // index 0 is unused
		{tag: tagUtf8, str: "test/A"},  // 1
		{tag: tagClass, nameIndex: 1},  // 2: test.A
		{tag: tagUtf8, str: "test/B"},  // 3
		{tag: tagClass, nameIndex: 3},  // 4: test.B
		{tag: tagUtf8, str: "counter"}, // 5
		{tag: tagUtf8, str: "I"},       // 6
		{tag: tagNameAndType, nameIndex: 5, descriptorIndex: 6}, // 7
		{tag: tagFieldref, classIndex: 4, nameAndTypeIndex: 7},  // 8: test.B#counter:I
	}
	cf := &ClassFile{pool: p}

	code := []byte{
		opNew, 0x00, 0x02, // new #2 (test.A)                 pc=0
		opGetStatic, 0x00, 0x08, // getstatic #8 (test.B#counter) pc=3
		opPutStatic, 0x00, 0x08, // putstatic #8 (test.B#counter) pc=6
		opLdc, 0x02, // ldc #2 (the Class literal test.A.class) — must NOT trigger init  pc=9
	}

	calls, initRefs, unresolved, err := cf.ExtractCalls(code)
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}
	if len(calls) != 0 {
		t.Errorf("calls = %+v, want none (no invoke* instructions in this bytecode)", calls)
	}
	if unresolved != 0 {
		t.Errorf("unresolved = %d, want 0", unresolved)
	}

	want := []InitRef{
		{Kind: InitNew, Owner: "test.A", PC: 0},
		{Kind: InitGetStatic, Owner: "test.B", Name: "counter", Desc: "I", PC: 3},
		{Kind: InitPutStatic, Owner: "test.B", Name: "counter", Desc: "I", PC: 6},
	}
	if len(initRefs) != len(want) {
		t.Fatalf("initRefs = %+v, want %+v", initRefs, want)
	}
	for i, w := range want {
		if initRefs[i] != w {
			t.Errorf("initRefs[%d] = %+v, want %+v", i, initRefs[i], w)
		}
	}
	// The ldc of a Class constant (test.A.class) must not have produced a
	// fourth InitRef — "samo posiadanie literału klasy nie staje się
	// regułą inicjalizacji" (JVM_GRAPH G3's own functional spec).
}
