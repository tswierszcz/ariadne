package dataflow

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

// Analyze runs DATA_FLOW D2's local-value analysis over one method of cf
// (fqcn is cf.ThisClass, passed separately so a caller that already has
// it doesn't need a second lookup). It returns an error only for a
// genuinely malformed method (the same contract classfile.BuildCFG and
// DecodeLocalOps already have) — anything this analysis itself can't
// handle (jsr/ret, an internal inconsistency) is reported through
// MethodFacts.Aborted, never as an error or a panic.
func Analyze(cf *classfile.ClassFile, fqcn string, m classfile.Method) (*MethodFacts, error) {
	f := &MethodFacts{calls: map[bytecodeir.Site][]valueflow.Value{}}
	if len(m.Code) == 0 {
		return f, nil
	}

	cfg, err := classfile.BuildCFG(m.Code, m.ExceptionTable)
	if err != nil {
		return nil, err
	}
	ops, err := cf.DecodeLocalOps(m.Code)
	if err != nil {
		return nil, err
	}

	site := func(pc int, kind bytecodeir.OpKind) bytecodeir.Site {
		return bytecodeir.Site{Method: fqcn + "#" + m.Name, Descriptor: m.Descriptor, PC: pc, Op: kind}
	}

	opByPC := make(map[int]classfile.LocalOp, len(ops))
	for _, op := range ops {
		opByPC[op.PC] = op
		if op.Kind != classfile.LocalOpCall {
			continue
		}
		n, err := callArgCount(op.CallDesc)
		if err != nil {
			continue // malformed descriptor: this Site stays unregistered, so a lookup there is a genuine caller error, not a false "aborted" claim
		}
		args := make([]valueflow.Value, n)
		for i := range args {
			args[i] = valueflow.Unknown("call argument not yet analyzed")
		}
		f.calls[site(op.PC, bytecodeir.OpCall)] = args
	}

	it := &interp{fqcn: fqcn, m: m, opByPC: opByPC, site: site, facts: f}
	f.reason = it.run(cfg)
	f.aborted = f.reason != ""
	return f, nil
}

func callArgCount(desc string) (int, error) {
	args, err := classfile.ArgumentSlots(desc, true)
	if err != nil {
		return 0, err
	}
	return len(args), nil
}

type interp struct {
	fqcn   string
	m      classfile.Method
	opByPC map[int]classfile.LocalOp
	site   func(pc int, kind bytecodeir.OpKind) bytecodeir.Site
	facts  *MethodFacts
}

// stackEntry is one operand-stack slot: its abstract value and its JVM
// computational-type width (1 word, or 2 for a long/double) — needed to
// get dup2/pop2/swap's width-dependent behavior right (JVMS 6.5).
type stackEntry struct {
	v     valueflow.Value
	width int
}

type absState struct {
	locals map[int]valueflow.Value
	stack  []stackEntry
}

func (s absState) clone() absState {
	locals := make(map[int]valueflow.Value, len(s.locals))
	for k, v := range s.locals {
		locals[k] = v
	}
	return absState{locals: locals, stack: append([]stackEntry(nil), s.stack...)}
}

func (s *absState) push(v valueflow.Value, width int) {
	s.stack = append(s.stack, stackEntry{v: v, width: width})
}

func (s *absState) pushEntry(e stackEntry) {
	s.stack = append(s.stack, e)
}

func (s *absState) pop() (stackEntry, bool) {
	if len(s.stack) == 0 {
		return stackEntry{}, false
	}
	e := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return e, true
}

// popN pops n entries, returned deepest-first: result[0] was pushed
// earliest of the group, result[n-1] is whatever was on top.
func (s *absState) popN(n int) ([]stackEntry, bool) {
	if n == 0 {
		return nil, true
	}
	if len(s.stack) < n {
		return nil, false
	}
	out := append([]stackEntry(nil), s.stack[len(s.stack)-n:]...)
	s.stack = s.stack[:len(s.stack)-n]
	return out, true
}

// maxIterations bounds the worklist in case of a genuine bug; it is far
// above anything a real method should need — valueflow.Merge's own
// widening-to-Unknown-on-budget-exceeded guarantees every local/stack
// slot's value chain stabilizes within a small, bounded number of merges
// (DefaultBudget-many distinct candidates, then Unknown forever after),
// so a real method's block count dominates the true iteration count, not
// this constant. D1 measured up to 451 blocks in one real method.
const maxIterations = 50000

// run is the fixed-point worklist over cfg's blocks. It returns "" on a
// clean pass over every reachable block, or a reason string the instant
// it hits something this analysis increment does not model (jsr/ret) or
// does not trust itself to have modeled correctly (a stack-shape
// inconsistency at a control-flow join — never guessed past, per DATA_FLOW
// D0's own widen-don't-guess contract).
func (it *interp) run(cfg *classfile.CFG) string {
	if len(cfg.Blocks) == 0 {
		return ""
	}

	maxLocals := it.m.MaxLocals
	initial := absState{locals: make(map[int]valueflow.Value, maxLocals)}
	for i := 0; i < maxLocals; i++ {
		initial.locals[i] = valueflow.Unknown("local variable not yet assigned by this analysis")
	}
	if !it.m.IsStatic() && maxLocals > 0 {
		initial.locals[0] = valueflow.Typed(it.fqcn)
	}
	args, err := classfile.ArgumentSlots(it.m.Descriptor, it.m.IsStatic())
	if err != nil {
		return fmt.Sprintf("invalid method descriptor: %v", err)
	}
	for _, a := range args {
		if a.Slot < maxLocals {
			initial.locals[a.Slot] = valueflow.Typed(a.Type)
		}
	}

	entry := make([]*absState, len(cfg.Blocks))
	entry[0] = &initial
	queue := []int{0}
	queued := make([]bool, len(cfg.Blocks))
	queued[0] = true

	iterations := 0
	for len(queue) > 0 {
		iterations++
		if iterations > maxIterations {
			return "exceeded the fixed-point iteration budget (possible non-terminating merge)"
		}

		bi := queue[0]
		queue = queue[1:]
		queued[bi] = false

		out, reason := it.interpretBlock(cfg.Blocks[bi], entry[bi].clone())
		if reason != "" {
			return reason
		}

		for _, e := range cfg.Edges {
			if e.From != bi {
				continue
			}
			succIn := out
			if e.Kind == classfile.EdgeException {
				succIn = exceptionEntryState(maxLocals)
			}
			if entry[e.To] == nil {
				c := succIn.clone()
				entry[e.To] = &c
				if !queued[e.To] {
					queue = append(queue, e.To)
					queued[e.To] = true
				}
				continue
			}
			merged, changed, reason := mergeStates(*entry[e.To], succIn)
			if reason != "" {
				return reason
			}
			if changed {
				entry[e.To] = &merged
				if !queued[e.To] {
					queue = append(queue, e.To)
					queued[e.To] = true
				}
			}
		}
	}
	return ""
}

func (it *interp) interpretBlock(b classfile.BasicBlock, st absState) (absState, string) {
	for _, pc := range b.PCs {
		op, ok := it.opByPC[pc]
		if !ok {
			return st, fmt.Sprintf("pc=%d: no decoded instruction (internal inconsistency between CFG and LocalOps)", pc)
		}
		if reason := it.step(&st, op); reason != "" {
			return st, fmt.Sprintf("pc=%d: %s", pc, reason)
		}
	}
	return st, ""
}

func exceptionEntryState(maxLocals int) absState {
	locals := make(map[int]valueflow.Value, maxLocals)
	for i := 0; i < maxLocals; i++ {
		locals[i] = valueflow.Unknown("local variable value before the exception that entered this handler is not tracked by this analysis increment")
	}
	return absState{
		locals: locals,
		stack:  []stackEntry{{v: valueflow.Unknown("exception object"), width: 1}},
	}
}

// mergeStates joins two control-flow predecessors' exit states into one
// entry state. A mismatched operand-stack depth or per-slot width is an
// internal inconsistency this analysis refuses to paper over by
// guessing; it aborts the whole method instead (DATA_FLOW D0's own
// widen-or-abort, never a disguised fabrication).
func mergeStates(a, b absState) (merged absState, changed bool, abortReason string) {
	if len(a.stack) != len(b.stack) {
		return absState{}, false, fmt.Sprintf(
			"control-flow join sees mismatched operand-stack depths (%d vs %d)", len(a.stack), len(b.stack))
	}
	stack := make([]stackEntry, len(a.stack))
	for i := range a.stack {
		if a.stack[i].width != b.stack[i].width {
			return absState{}, false, fmt.Sprintf(
				"control-flow join sees mismatched operand-stack widths at depth %d", i)
		}
		mv := valueflow.Merge(a.stack[i].v, b.stack[i].v)
		if !stableEqual(mv, a.stack[i].v) {
			changed = true
		}
		stack[i] = stackEntry{v: mv, width: a.stack[i].width}
	}

	locals := make(map[int]valueflow.Value, len(a.locals))
	for slot, av := range a.locals {
		bv, ok := b.locals[slot]
		if !ok {
			bv = valueflow.Unknown("local variable absent on one control-flow path")
		}
		mv := valueflow.Merge(av, bv)
		if !stableEqual(mv, av) {
			changed = true
		}
		locals[slot] = mv
	}
	return absState{locals: locals, stack: stack}, changed, ""
}

// stableEqual is Value.Equal except two KindUnknown values always count
// as equal, and two KindSet values count as equal when their candidates
// pairwise do (regardless of Reason text). Value.Equal itself
// deliberately treats every KindUnknown as distinct (each occurrence is
// its own fact, per its own doc comment) — exactly right for Merge's own
// semantics, but wrong for this package's fixed-point convergence check:
// using Value.Equal here would make every Unknown-into-Unknown merge
// report "changed", so the worklist would never converge on any method
// with a loop touching an already-widened slot.
func stableEqual(a, b valueflow.Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case valueflow.KindUnknown:
		return true
	case valueflow.KindSet:
		if len(a.Candidates) != len(b.Candidates) {
			return false
		}
		for i := range a.Candidates {
			if !stableEqual(a.Candidates[i], b.Candidates[i]) {
				return false
			}
		}
		return true
	default:
		return a.Equal(b)
	}
}
