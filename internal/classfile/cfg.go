package classfile

import (
	"fmt"
	"sort"
)

// BasicBlock is a maximal straight-line run of instructions with one
// entry point: no instruction other than PCs[0] is ever the target of a
// jump or an exception handler, and control leaves the block only by
// falling off the end of its last instruction.
type BasicBlock struct {
	StartPC int
	PCs     []int // every instruction's PC in this block, in order
}

// EdgeKind says why a CFGEdge exists — DATA_FLOW D1's own requirement
// to tell a normal edge apart from an exceptional one, so a consumer
// never mistakes "this block can throw into a handler" for "this is
// where the method normally continues" (D1: "Nie mylić rzuconego
// wyjątku z normalnym wynikiem metody").
type EdgeKind string

const (
	EdgeFallThrough EdgeKind = "fallthrough" // the next instruction, no branch taken
	EdgeBranch      EdgeKind = "branch"      // goto, a taken if*, or a switch target
	EdgeException   EdgeKind = "exception"   // this block's PC range is guarded by a handler
)

// CFGEdge is one edge between two BasicBlocks, indexed into CFG.Blocks.
type CFGEdge struct {
	From, To int
	Kind     EdgeKind
	// CatchType is set only for EdgeException: "" for a catch-all
	// handler (see ExceptionHandler's own doc comment).
	CatchType string
}

// CFG is one method's control-flow graph (DATA_FLOW D1).
type CFG struct {
	Blocks []BasicBlock
	Edges  []CFGEdge
	// Unsupported lists the PC of every jsr/ret instruction found — the
	// plan's own named example of an instruction this builder
	// recognizes but does not model the control-flow semantics of. A
	// jsr/ret method still gets a CFG (a conservative fall-through edge
	// takes the place of the real subroutine-call semantics), but a
	// consumer must treat any fact near an Unsupported PC as unreliable,
	// not as a confirmed absence of a path.
	Unsupported []int
}

// BuildCFG decodes code and partitions it into basic blocks and edges,
// including the exceptional edges exceptionTable describes. It never
// panics or fails outright on code decodeInstructions itself accepts;
// an unknown/invalid opcode is still caught by decodeInstructions's own
// error return (DATA_FLOW D1: "niepoprawne wejścia nie powodują paniki
// ani fikcyjnego braku przepływu").
func BuildCFG(code []byte, exceptionTable []ExceptionHandler) (*CFG, error) {
	instrs, err := decodeInstructions(code)
	if err != nil {
		return nil, err
	}
	if len(instrs) == 0 {
		return &CFG{}, nil
	}

	pcIndex := make(map[int]int, len(instrs))
	for i, in := range instrs {
		pcIndex[in.pc] = i
	}

	// Step 1: every PC that starts a basic block: the method's own
	// first instruction, every branch/switch target, the instruction
	// right after any branch/return/throw/unsupported instruction (even
	// if some of those are unreachable in practice — that costs an
	// extra, harmless empty-in-edges block, never a missed edge), and
	// every exception handler's start.
	starts := map[int]bool{instrs[0].pc: true}
	for i, in := range instrs {
		targets, kind, err := controlTransfer(in)
		if err != nil {
			return nil, fmt.Errorf("pc=%d: %w", in.pc, err)
		}
		switch kind {
		case ctBranch, ctReturnOrThrow, ctUnsupported:
			for _, t := range targets {
				starts[t] = true
			}
			if i+1 < len(instrs) {
				starts[instrs[i+1].pc] = true
			}
		}
	}
	for _, h := range exceptionTable {
		starts[h.HandlerPC] = true
	}

	var blockStartPCs []int
	for pc := range starts {
		blockStartPCs = append(blockStartPCs, pc)
	}
	sort.Ints(blockStartPCs)

	blockIndexByPC := make(map[int]int, len(blockStartPCs))
	blocks := make([]BasicBlock, len(blockStartPCs))
	for bi, startPC := range blockStartPCs {
		blockIndexByPC[startPC] = bi
		blocks[bi] = BasicBlock{StartPC: startPC}
	}

	// Step 2: assign every instruction to the block whose StartPC is the
	// greatest one at or before it.
	cur := 0
	for _, in := range instrs {
		if bi, ok := blockIndexByPC[in.pc]; ok {
			cur = bi
		}
		blocks[cur].PCs = append(blocks[cur].PCs, in.pc)
	}

	// Step 3: edges from each block's last instruction.
	var edges []CFGEdge
	var unsupported []int
	for bi, b := range blocks {
		if len(b.PCs) == 0 {
			continue
		}
		lastIdx := pcIndex[b.PCs[len(b.PCs)-1]]
		in := instrs[lastIdx]
		targets, kind, err := controlTransfer(in)
		if err != nil {
			return nil, fmt.Errorf("pc=%d: %w", in.pc, err)
		}
		fallsThrough := func() {
			if lastIdx+1 < len(instrs) {
				edges = append(edges, CFGEdge{From: bi, To: blockIndexByPC[instrs[lastIdx+1].pc], Kind: EdgeFallThrough})
			}
		}
		switch kind {
		case ctFallThrough:
			fallsThrough()
		case ctBranch:
			if isConditionalBranch(in.opcode) {
				fallsThrough()
			}
			for _, t := range targets {
				edges = append(edges, CFGEdge{From: bi, To: blockIndexByPC[t], Kind: EdgeBranch})
			}
		case ctReturnOrThrow:
			// no successor: a thrown exception or a return both end the
			// block with no normal-flow edge out of it.
		case ctUnsupported:
			unsupported = append(unsupported, in.pc)
			// Conservative: still fall through, so one jsr/ret doesn't
			// dead-end analysis of the rest of the method — but see
			// CFG.Unsupported's own doc comment on why this isn't a
			// real modeling of jsr/ret's actual semantics.
			fallsThrough()
		}
	}

	// Step 4: exceptional edges. A block with any instruction inside
	// [h.StartPC, h.EndPC) gets an edge to h's handler block.
	for _, h := range exceptionTable {
		handlerBI, ok := blockIndexByPC[h.HandlerPC]
		if !ok {
			continue // malformed handler_pc outside the method; skip rather than fail the whole CFG
		}
		for bi, b := range blocks {
			for _, pc := range b.PCs {
				if pc >= h.StartPC && pc < h.EndPC {
					edges = append(edges, CFGEdge{From: bi, To: handlerBI, Kind: EdgeException, CatchType: h.CatchType})
					break
				}
			}
		}
	}

	return &CFG{Blocks: blocks, Edges: edges, Unsupported: unsupported}, nil
}

type controlKind int

const (
	ctFallThrough controlKind = iota
	ctBranch
	ctReturnOrThrow
	ctUnsupported
)

// controlTransfer reports how in affects control flow and, for a
// branch, every absolute target PC it can jump to.
func controlTransfer(in instruction) (targets []int, kind controlKind, err error) {
	op := in.opcode
	switch {
	case op == opJsr || op == opJsrW || op == opRet:
		return nil, ctUnsupported, nil
	case op == opGoto:
		t, err := branchTarget16(in)
		return []int{t}, ctBranch, err
	case op == opGotoW:
		t, err := branchTarget32(in)
		return []int{t}, ctBranch, err
	case op >= opIfeqFirst && op <= opIfAcmpneLast, op == opIfnull, op == opIfnonnull:
		t, err := branchTarget16(in)
		return []int{t}, ctBranch, err
	case op == opTableSwitch || op == opLookupSwitch:
		return switchTargets(in)
	case op >= opIreturn && op <= opReturnLast, op == opAthrow:
		return nil, ctReturnOrThrow, nil
	default:
		return nil, ctFallThrough, nil
	}
}

// isConditionalBranch reports whether op branches without ending the
// block outright — i.e. it also has a fall-through successor when not
// taken, unlike goto/goto_w which unconditionally transfer control.
func isConditionalBranch(op byte) bool {
	switch {
	case op >= opIfeqFirst && op <= opIfAcmpneLast, op == opIfnull, op == opIfnonnull:
		return true
	default:
		return false
	}
}

// branchTarget16 resolves a 2-byte signed branch offset (JVMS 6.5: if*,
// if_icmp*, if_acmp*, ifnull, ifnonnull, goto), relative to the
// instruction's own pc.
func branchTarget16(in instruction) (int, error) {
	if len(in.operand) < 2 {
		return 0, fmt.Errorf("branch instruction %#x has %d operand bytes, want 2", in.opcode, len(in.operand))
	}
	offset := int16(uint16(in.operand[0])<<8 | uint16(in.operand[1]))
	return in.pc + int(offset), nil
}

// branchTarget32 resolves a 4-byte signed branch offset (goto_w,
// jsr_w — jsr_w itself is routed through ctUnsupported before this is
// called, but the helper stays general).
func branchTarget32(in instruction) (int, error) {
	if len(in.operand) < 4 {
		return 0, fmt.Errorf("branch instruction %#x has %d operand bytes, want 4", in.opcode, len(in.operand))
	}
	offset := int32(uint32(in.operand[0])<<24 | uint32(in.operand[1])<<16 | uint32(in.operand[2])<<8 | uint32(in.operand[3]))
	return in.pc + int(offset), nil
}

// switchTargets resolves a tableswitch/lookupswitch's default target
// and every case target (JVMS 6.5 tableswitch/lookupswitch), all
// relative to the instruction's own pc per the spec.
//
// in.operand is code[in.pc+1 : end] (decodeInstructions's own slicing),
// which for these two opcodes *includes* the 0-3 padding bytes
// skipSwitch itself accounts for when computing instruction length —
// the padding is computed from in.pc+1 (the first byte after the
// opcode) up to the next multiple of 4, exactly mirroring skipSwitch's
// own math, so the actual defaultbyte/low/high/npairs fields start
// after that same padding, not at b[0].
func switchTargets(in instruction) ([]int, controlKind, error) {
	b := in.operand
	pad := (4 - ((in.pc + 1) % 4)) % 4
	if len(b) < pad+8 {
		return nil, ctBranch, fmt.Errorf("switch instruction %#x has only %d operand bytes", in.opcode, len(b))
	}
	b = b[pad:]
	defaultOffset := int32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
	targets := []int{in.pc + int(defaultOffset)}

	if in.opcode == opTableSwitch {
		low := int32(uint32(b[4])<<24 | uint32(b[5])<<16 | uint32(b[6])<<8 | uint32(b[7]))
		high := int32(uint32(b[8])<<24 | uint32(b[9])<<16 | uint32(b[10])<<8 | uint32(b[11]))
		n := int(high) - int(low) + 1
		for i := 0; i < n; i++ {
			at := 12 + i*4
			if at+4 > len(b) {
				return nil, ctBranch, fmt.Errorf("tableswitch at pc=%d: jump table runs past operand bytes", in.pc)
			}
			off := int32(uint32(b[at])<<24 | uint32(b[at+1])<<16 | uint32(b[at+2])<<8 | uint32(b[at+3]))
			targets = append(targets, in.pc+int(off))
		}
	} else { // lookupswitch
		npairs := int32(uint32(b[4])<<24 | uint32(b[5])<<16 | uint32(b[6])<<8 | uint32(b[7]))
		for i := 0; i < int(npairs); i++ {
			at := 8 + i*8
			if at+8 > len(b) {
				return nil, ctBranch, fmt.Errorf("lookupswitch at pc=%d: match-offset pairs run past operand bytes", in.pc)
			}
			off := int32(uint32(b[at+4])<<24 | uint32(b[at+5])<<16 | uint32(b[at+6])<<8 | uint32(b[at+7]))
			targets = append(targets, in.pc+int(off))
		}
	}
	return targets, ctBranch, nil
}
