package classfile

import "testing"

func decodeOne(t *testing.T, code []byte) LocalOp {
	t.Helper()
	cf := &ClassFile{}
	ops, err := cf.DecodeLocalOps(code)
	if err != nil {
		t.Fatalf("DecodeLocalOps: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("DecodeLocalOps = %+v, want exactly 1 op", ops)
	}
	return ops[0]
}

func TestDecodeLocalOps_wideLoadStoreIinc(t *testing.T) {
	cf := &ClassFile{}

	// wide iload 300 (300 needs a 16-bit index, i.e. an actual `wide` form)
	ops, err := cf.DecodeLocalOps([]byte{opWide, opIload, 0x01, 0x2c})
	if err != nil {
		t.Fatalf("DecodeLocalOps(wide iload): %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != LocalOpLoad || ops[0].Slot != 300 || ops[0].Width != 1 {
		t.Errorf("wide iload 300 = %+v, want LocalOpLoad slot=300 width=1", ops[0])
	}

	ops, err = cf.DecodeLocalOps([]byte{opWide, opWideIinc, 0x01, 0x2c, 0x00, 0x05})
	if err != nil {
		t.Fatalf("DecodeLocalOps(wide iinc): %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != LocalOpIinc || ops[0].Slot != 300 {
		t.Errorf("wide iinc 300 = %+v, want LocalOpIinc slot=300", ops[0])
	}

	ops, err = cf.DecodeLocalOps([]byte{opWide, opLstore, 0x00, 0x05})
	if err != nil {
		t.Fatalf("DecodeLocalOps(wide lstore): %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != LocalOpStore || ops[0].Slot != 5 {
		t.Errorf("wide lstore 5 = %+v, want LocalOpStore slot=5", ops[0])
	}
}

func TestDecodeLocalOps_arithmeticWidths(t *testing.T) {
	cases := []struct {
		name string
		op   byte
		kind LocalOpKind
		want int
	}{
		{"iadd", opIadd, LocalOpBinaryNumeric, 1},
		{"ladd", opIadd + 1, LocalOpBinaryNumeric, 2},
		{"fadd", opIadd + 2, LocalOpBinaryNumeric, 1},
		{"dadd", opIadd + 3, LocalOpBinaryNumeric, 2},
		{"ineg", opIneg, LocalOpUnaryNumeric, 1},
		{"lneg", opIneg + 1, LocalOpUnaryNumeric, 2},
		{"dneg", opIneg + 3, LocalOpUnaryNumeric, 2},
		{"ishl", opIshl, LocalOpBinaryNumeric, 1},
		{"lshl", opIshl + 1, LocalOpBinaryNumeric, 2},
		{"lxor", opLxor, LocalOpBinaryNumeric, 2},
		{"i2l", 0x85, LocalOpConvert, 2},
		{"i2f", 0x86, LocalOpConvert, 1},
		{"l2d", 0x8a, LocalOpConvert, 2},
		{"d2f", 0x90, LocalOpConvert, 1},
		{"i2b", opI2b, LocalOpConvert, 1},
		{"lcmp", opLcmp, LocalOpCompare, 1},
		{"dcmpg", opDcmpg, LocalOpCompare, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			op := decodeOne(t, []byte{c.op})
			if op.Kind != c.kind || op.Width != c.want {
				t.Errorf("%s = %+v, want Kind=%q Width=%d", c.name, op, c.kind, c.want)
			}
		})
	}
}

// TestDecodeLocalOps_dupFamily pins each dup/dup2/pop/swap variant's own
// decoded Kind — the actual stack-shuffle semantics are exercised at the
// internal/dataflow layer (that package owns the width-dependent
// behavior); this only confirms the opcode-to-Kind mapping itself.
func TestDecodeLocalOps_dupFamily(t *testing.T) {
	cases := []struct {
		op   byte
		kind LocalOpKind
	}{
		{opPop, LocalOpPop}, {opPop2, LocalOpPop2},
		{opDup, LocalOpDup}, {opDupX1, LocalOpDupX1}, {opDupX2, LocalOpDupX2},
		{opDup2, LocalOpDup2}, {opDup2X1, LocalOpDup2X1}, {opDup2X2, LocalOpDup2X2},
		{opSwap, LocalOpSwap},
	}
	for _, c := range cases {
		op := decodeOne(t, []byte{c.op})
		if op.Kind != c.kind {
			t.Errorf("opcode %#x = %+v, want Kind=%q", c.op, op, c.kind)
		}
	}
}

func TestDecodeLocalOps_ldcStringAndClass(t *testing.T) {
	p := pool{
		{},
		{tag: tagUtf8, str: "hello"},       // 1
		{tag: tagString, nameIndex: 1},     // 2
		{tag: tagUtf8, str: "test/Target"}, // 3
		{tag: tagClass, nameIndex: 3},      // 4
		{tag: tagInteger},                  // 5: out of D2's declared scope
	}
	cf := &ClassFile{pool: p}

	ops, err := cf.DecodeLocalOps([]byte{opLdc, 0x02, opLdc, 0x04, opLdc, 0x05})
	if err != nil {
		t.Fatalf("DecodeLocalOps: %v", err)
	}
	if len(ops) != 3 {
		t.Fatalf("got %d ops, want 3", len(ops))
	}
	if ops[0].ConstKind != ConstKindString || ops[0].ConstStr != "hello" {
		t.Errorf("ldc string = %+v, want ConstKindString \"hello\"", ops[0])
	}
	if ops[1].ConstKind != ConstKindClass || ops[1].ConstStr != "test.Target" {
		t.Errorf("ldc class = %+v, want ConstKindClass \"test.Target\"", ops[1])
	}
	if ops[2].ConstKind != "" {
		t.Errorf("ldc integer = %+v, want ConstKind \"\" (out of D2's scope)", ops[2])
	}
}

func TestDecodeLocalOps_fieldAndArrayWidths(t *testing.T) {
	p := pool{
		{},
		{tag: tagUtf8, str: "test/Holder"}, // 1
		{tag: tagClass, nameIndex: 1},      // 2
		{tag: tagUtf8, str: "count"},       // 3
		{tag: tagUtf8, str: "J"},           // 4: long field
		{tag: tagNameAndType, nameIndex: 3, descriptorIndex: 4}, // 5
		{tag: tagFieldref, classIndex: 2, nameAndTypeIndex: 5},  // 6
	}
	cf := &ClassFile{pool: p}

	ops, err := cf.DecodeLocalOps([]byte{opGetStatic, 0x00, 0x06})
	if err != nil {
		t.Fatalf("DecodeLocalOps: %v", err)
	}
	if ops[0].Kind != LocalOpFieldGet || ops[0].Width != 2 || ops[0].FieldDesc != "J" || !ops[0].IsStatic {
		t.Errorf("getstatic long field = %+v, want FieldGet width=2 static", ops[0])
	}

	ops, err = cf.DecodeLocalOps([]byte{opLaload})
	if err != nil {
		t.Fatalf("DecodeLocalOps(laload): %v", err)
	}
	if ops[0].Kind != LocalOpArrayLoad || ops[0].Width != 2 {
		t.Errorf("laload = %+v, want ArrayLoad width=2", ops[0])
	}
}

func TestDecodeLocalOps_invalidOpcodeErrorsNotPanics(t *testing.T) {
	cf := &ClassFile{}
	if _, err := cf.DecodeLocalOps([]byte{0xca}); err == nil {
		t.Errorf("DecodeLocalOps(invalid opcode) succeeded, want an error")
	}
}
