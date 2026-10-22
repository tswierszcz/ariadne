package classfile

import "testing"

func edgeExists(edges []CFGEdge, from, to int, kind EdgeKind) bool {
	for _, e := range edges {
		if e.From == from && e.To == to && e.Kind == kind {
			return true
		}
	}
	return false
}

func blockIndexAt(blocks []BasicBlock, startPC int) int {
	for i, b := range blocks {
		if b.StartPC == startPC {
			return i
		}
	}
	return -1
}

// TestBuildCFG_branch is DATA_FLOW D1's own "branch" example: an `if`
// with a taken path and a fall-through path that both converge at a
// join point. Hand-assembled since no fixture source has a real branch
// and this environment has no javac.
//
// Bytecode:
//
//	pc=0: iconst_0                    (then-branch condition pushed)
//	pc=1: ifeq target=8   (offset 7)  conditional: branch to 8, else fall through to pc=4
//	pc=4: iconst_1                    (the "then" path's own work)
//	pc=5: goto target=9   (offset 4)  unconditional join
//	pc=8: iconst_2                    (the "else" path's own work)
//	pc=9: return                      (join point)
func TestBuildCFG_branch(t *testing.T) {
	code := []byte{
		0x03,             // pc=0 iconst_0
		0x99, 0x00, 0x07, // pc=1 ifeq +7 -> pc=8
		0x04,             // pc=4 iconst_1
		0xa7, 0x00, 0x04, // pc=5 goto +4 -> pc=9
		0x05, // pc=8 iconst_2
		0xb1, // pc=9 return
	}
	cfg, err := BuildCFG(code, nil)
	if err != nil {
		t.Fatalf("BuildCFG: %v", err)
	}

	b0 := blockIndexAt(cfg.Blocks, 0)
	b4 := blockIndexAt(cfg.Blocks, 4)
	b8 := blockIndexAt(cfg.Blocks, 8)
	b9 := blockIndexAt(cfg.Blocks, 9)
	if b0 < 0 || b4 < 0 || b8 < 0 || b9 < 0 {
		t.Fatalf("missing expected block starts; got %+v", cfg.Blocks)
	}

	if !edgeExists(cfg.Edges, b0, b8, EdgeBranch) {
		t.Errorf("missing branch edge (ifeq taken) %d->%d; edges=%+v", b0, b8, cfg.Edges)
	}
	if !edgeExists(cfg.Edges, b0, b4, EdgeFallThrough) {
		t.Errorf("missing fall-through edge (ifeq not taken) %d->%d; edges=%+v", b0, b4, cfg.Edges)
	}
	if !edgeExists(cfg.Edges, b4, b9, EdgeBranch) {
		t.Errorf("missing branch edge (goto) %d->%d; edges=%+v", b4, b9, cfg.Edges)
	}
	if !edgeExists(cfg.Edges, b8, b9, EdgeFallThrough) {
		t.Errorf("missing fall-through edge (into join point) %d->%d; edges=%+v", b8, b9, cfg.Edges)
	}
	// The join block (return) must have no outgoing edge at all.
	for _, e := range cfg.Edges {
		if e.From == b9 {
			t.Errorf("return block has an outgoing edge, want none: %+v", e)
		}
	}
}

// TestBuildCFG_loop is D1's own "loop" example: a backward branch
// creating a cycle in the graph.
//
// Bytecode:
//
//	pc=0: iconst_0
//	pc=1: ifeq target=8    (offset 7)  loop exit
//	pc=4: iconst_1                     loop body
//	pc=5: goto target=1    (offset -4) back-edge
//	pc=8: return
func TestBuildCFG_loop(t *testing.T) {
	code := []byte{
		0x03,             // pc=0 iconst_0
		0x99, 0x00, 0x07, // pc=1 ifeq +7 -> pc=8
		0x04,             // pc=4 iconst_1
		0xa7, 0xff, 0xfc, // pc=5 goto -4 -> pc=1
		0xb1, // pc=8 return
	}
	cfg, err := BuildCFG(code, nil)
	if err != nil {
		t.Fatalf("BuildCFG: %v", err)
	}

	b1 := blockIndexAt(cfg.Blocks, 1)
	b4 := blockIndexAt(cfg.Blocks, 4)
	if b1 < 0 || b4 < 0 {
		t.Fatalf("missing expected block starts; got %+v", cfg.Blocks)
	}
	if !edgeExists(cfg.Edges, b4, b1, EdgeBranch) {
		t.Errorf("missing the loop's own back-edge %d->%d; edges=%+v", b4, b1, cfg.Edges)
	}
	if !edgeExists(cfg.Edges, b1, b4, EdgeFallThrough) {
		t.Errorf("missing edge into the loop body %d->%d; edges=%+v", b1, b4, cfg.Edges)
	}
}

// TestBuildCFG_exception is D1's own "exception" example, and its own
// explicit requirement "nie mylić rzuconego wyjątku z normalnym
// wynikiem metody": a try range ending in `return` must get an
// exceptional edge to the handler, and must NOT also get a normal
// fall-through edge there (a return has no normal successor at all).
//
// Bytecode:
//
//	pc=0: iconst_0        \ try range [0,3)
//	pc=1: iconst_1        /
//	pc=2: return          (ends the try range; no normal successor)
//	pc=3: astore_1        handler start
//	pc=4: return
func TestBuildCFG_exception(t *testing.T) {
	code := []byte{
		0x03, // pc=0 iconst_0
		0x04, // pc=1 iconst_1
		0xb1, // pc=2 return
		0x4c, // pc=3 astore_1 (handler)
		0xb1, // pc=4 return
	}
	handlers := []ExceptionHandler{
		{StartPC: 0, EndPC: 3, HandlerPC: 3, CatchType: "test.MyExc"},
	}
	cfg, err := BuildCFG(code, handlers)
	if err != nil {
		t.Fatalf("BuildCFG: %v", err)
	}

	b0 := blockIndexAt(cfg.Blocks, 0)
	b3 := blockIndexAt(cfg.Blocks, 3)
	if b0 < 0 || b3 < 0 {
		t.Fatalf("missing expected block starts; got %+v", cfg.Blocks)
	}

	var found *CFGEdge
	for i := range cfg.Edges {
		if cfg.Edges[i].From == b0 && cfg.Edges[i].To == b3 {
			found = &cfg.Edges[i]
		}
	}
	if found == nil {
		t.Fatalf("missing exceptional edge %d->%d; edges=%+v", b0, b3, cfg.Edges)
	}
	if found.Kind != EdgeException {
		t.Errorf("edge %d->%d has Kind=%q, want EdgeException", b0, b3, found.Kind)
	}
	if found.CatchType != "test.MyExc" {
		t.Errorf("edge CatchType = %q, want test.MyExc", found.CatchType)
	}

	// Exactly one edge between these blocks: the exceptional one. A
	// second, fall-through edge here would be exactly the bug D1 warns
	// against (a thrown exception mistaken for a normal result).
	count := 0
	for _, e := range cfg.Edges {
		if e.From == b0 && e.To == b3 {
			count++
		}
	}
	if count != 1 {
		t.Errorf("found %d edges %d->%d, want exactly 1 (exception only, no fall-through)", count, b0, b3)
	}
}

// TestBuildCFG_jsrRetIsUnsupportedNotCrashing is D1's own named example
// of an instruction to mark unsupported rather than model or crash on.
func TestBuildCFG_jsrRetIsUnsupportedNotCrashing(t *testing.T) {
	code := []byte{
		0xa8, 0x00, 0x03, // pc=0 jsr +3 -> pc=3 (never actually modeled as a branch)
		0x4c,       // pc=3 astore_1
		0xa9, 0x01, // pc=4 ret 1
	}
	cfg, err := BuildCFG(code, nil)
	if err != nil {
		t.Fatalf("BuildCFG: %v", err)
	}
	if len(cfg.Unsupported) != 2 {
		t.Fatalf("Unsupported = %v, want exactly 2 PCs (jsr at 0, ret at 4)", cfg.Unsupported)
	}
	foundJsr, foundRet := false, false
	for _, pc := range cfg.Unsupported {
		if pc == 0 {
			foundJsr = true
		}
		if pc == 4 {
			foundRet = true
		}
	}
	if !foundJsr || !foundRet {
		t.Errorf("Unsupported = %v, want PCs 0 and 4", cfg.Unsupported)
	}
}

func TestBuildCFG_emptyCode(t *testing.T) {
	cfg, err := BuildCFG(nil, nil)
	if err != nil {
		t.Fatalf("BuildCFG(nil): %v", err)
	}
	if len(cfg.Blocks) != 0 || len(cfg.Edges) != 0 {
		t.Errorf("BuildCFG(nil) = %+v, want an empty CFG", cfg)
	}
}

// TestBuildCFG_invalidOpcodeErrorsNotPanics is D1's own requirement:
// an invalid input must not panic.
func TestBuildCFG_invalidOpcodeErrorsNotPanics(t *testing.T) {
	code := []byte{0xca} // 0xca (breakpoint) is reserved, never a real instruction
	if _, err := BuildCFG(code, nil); err == nil {
		t.Errorf("BuildCFG(invalid opcode) succeeded, want an error")
	}
}

// TestBuildCFG_realFixtureMethod runs BuildCFG against real, javac-
// compiled bytecode (App#handle, which has invokeinterface,
// invokestatic and invokedynamic call sites — see App.java's own doc
// comment) to confirm it handles real-world instructions without
// error, even though this particular method has no branch of its own:
// every invoke* instruction must fall through normally (it doesn't
// transfer control by itself in this CFG model), landing in one single
// block with no internal edges.
func TestBuildCFG_realFixtureMethod(t *testing.T) {
	for _, release := range []int{8, 21} {
		t.Run(releaseName(release), func(t *testing.T) {
			cf := readFixtureClass(t, release, "fixture/App.class")
			m := methodByName(cf, "handle")
			if m == nil {
				t.Fatalf("no handle method found")
			}
			cfg, err := BuildCFG(m.Code, m.ExceptionTable)
			if err != nil {
				t.Fatalf("BuildCFG: %v", err)
			}
			if len(cfg.Blocks) != 1 {
				t.Errorf("handle() has no branches of its own, want exactly 1 block, got %d: %+v", len(cfg.Blocks), cfg.Blocks)
			}
			if len(cfg.Edges) != 0 {
				t.Errorf("a single straight-line block should have no internal edges, got %+v", cfg.Edges)
			}
			if len(cfg.Unsupported) != 0 {
				t.Errorf("handle() has no jsr/ret, got Unsupported=%v", cfg.Unsupported)
			}
		})
	}
}
