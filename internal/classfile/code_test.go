package classfile

import "testing"

// TestParseCodeAttribute_exceptionTableAndLineNumbers is DATA_FLOW D1's
// own prerequisite, tested against a hand-assembled Code attribute
// rather than a compiled fixture class: no existing fixture source has
// a try/catch block, and this environment has no javac to add one. The
// binary format (JVM spec 4.7.3/4.7.12) is simple enough to assemble by
// hand, and doing so pins down exactly which bytes produce which
// ExceptionHandler/LineNumberEntry, independent of any particular
// javac's own compiled output.
func TestParseCodeAttribute_exceptionTableAndLineNumbers(t *testing.T) {
	p := pool{
		{},                                      // index 0 unused
		{tag: tagUtf8, str: "test/MyException"}, // 1
		{tag: tagClass, nameIndex: 1},           // 2: test.MyException
	}

	// Code attribute body: max_stack=2, max_locals=1, 3 bytes of code
	// (irrelevant content for this test), one exception handler
	// (catch test.MyException, pc range [0,3) -> handler at pc 3), one
	// catch-all handler (catch_type=0, pc range [0,3) -> handler at
	// pc 6), and a LineNumberTable with two entries.
	data := []byte{
		0x00, 0x02, // max_stack = 2
		0x00, 0x01, // max_locals = 1
		0x00, 0x00, 0x00, 0x03, // code_length = 3
		0xb1, 0xb1, 0xb1, // 3 bytes of (irrelevant) code
		0x00, 0x02, // exception_table_length = 2
		// handler 1: [0,3) -> 3, catch test.MyException (#2)
		0x00, 0x00, 0x00, 0x03, 0x00, 0x03, 0x00, 0x02,
		// handler 2: [0,3) -> 6, catch-all (0)
		0x00, 0x00, 0x00, 0x03, 0x00, 0x06, 0x00, 0x00,
		0x00, 0x01, // attributes_count = 1
		// attribute 0: name_index referring to a Utf8 we haven't added
		// yet for "LineNumberTable" — add it as pool index 3 below and
		// reference it here.
	}
	// Append the LineNumberTable attribute, referencing pool index 3
	// ("LineNumberTable") which we add now that data's fixed prefix is
	// settled.
	p = append(p, cpEntry{tag: tagUtf8, str: "LineNumberTable"}) // index 3
	lineNumberTableInfo := []byte{
		0x00, 0x02, // line_number_table_length = 2
		0x00, 0x00, 0x00, 0x05, // pc=0, line=5
		0x00, 0x02, 0x00, 0x06, // pc=2, line=6
	}
	data = append(data,
		0x00, 0x03, // attribute_name_index = 3 ("LineNumberTable")
		byte(len(lineNumberTableInfo)>>24), byte(len(lineNumberTableInfo)>>16),
		byte(len(lineNumberTableInfo)>>8), byte(len(lineNumberTableInfo)),
	)
	data = append(data, lineNumberTableInfo...)

	got, err := parseCodeAttribute(data, p)
	if err != nil {
		t.Fatalf("parseCodeAttribute: %v", err)
	}

	if got.MaxStack != 2 || got.MaxLocals != 1 {
		t.Errorf("MaxStack/MaxLocals = %d/%d, want 2/1", got.MaxStack, got.MaxLocals)
	}
	if len(got.Code) != 3 {
		t.Errorf("Code = %v, want 3 bytes", got.Code)
	}

	wantHandlers := []ExceptionHandler{
		{StartPC: 0, EndPC: 3, HandlerPC: 3, CatchType: "test.MyException"},
		{StartPC: 0, EndPC: 3, HandlerPC: 6, CatchType: ""},
	}
	if len(got.ExceptionTable) != len(wantHandlers) {
		t.Fatalf("ExceptionTable = %+v, want %+v", got.ExceptionTable, wantHandlers)
	}
	for i, w := range wantHandlers {
		if got.ExceptionTable[i] != w {
			t.Errorf("ExceptionTable[%d] = %+v, want %+v", i, got.ExceptionTable[i], w)
		}
	}

	wantLines := []LineNumberEntry{{PC: 0, Line: 5}, {PC: 2, Line: 6}}
	if len(got.LineNumbers) != len(wantLines) {
		t.Fatalf("LineNumbers = %+v, want %+v", got.LineNumbers, wantLines)
	}
	for i, w := range wantLines {
		if got.LineNumbers[i] != w {
			t.Errorf("LineNumbers[%d] = %+v, want %+v", i, got.LineNumbers[i], w)
		}
	}
}

// TestParseCodeAttribute_noLineNumberTableIsNilNotError confirms D1's
// own requirement: absent debug info must not stop analysis or be
// treated as an error.
func TestParseCodeAttribute_noLineNumberTableIsNilNotError(t *testing.T) {
	p := pool{{}}
	data := []byte{
		0x00, 0x01, // max_stack
		0x00, 0x01, // max_locals
		0x00, 0x00, 0x00, 0x01, // code_length = 1
		0xb1,       // 1 byte of code (return)
		0x00, 0x00, // exception_table_length = 0
		0x00, 0x00, // attributes_count = 0
	}
	got, err := parseCodeAttribute(data, p)
	if err != nil {
		t.Fatalf("parseCodeAttribute: %v", err)
	}
	if got.LineNumbers != nil {
		t.Errorf("LineNumbers = %v, want nil (no debug info present)", got.LineNumbers)
	}
	if got.ExceptionTable != nil {
		t.Errorf("ExceptionTable = %v, want nil (no handlers present)", got.ExceptionTable)
	}
}
