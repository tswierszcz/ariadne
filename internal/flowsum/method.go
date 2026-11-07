package flowsum

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"ariadne/internal/classfile"
)

// MethodKey is one exact method: declaring class, name, descriptor.
type MethodKey struct{ Owner, Name, Desc string }

func (k MethodKey) String() string { return k.Owner + "#" + k.Name + k.Desc }

// CallDeps are the dependences of one call site's receiver and arguments.
type CallDeps struct {
	Receiver Dep
	Args     []Dep
}

// Summary is one method's dependency summary, in the method's own terms.
type Summary struct {
	Method MethodKey
	// Return: what the returned value depends on (nil for void, or a
	// value that depends on nothing — a constant).
	Return Dep
	// Fields: per written field ("owner.name"), what the written values
	// depend on.
	Fields map[string]Dep
	// Calls: per call PC, the receiver's and arguments' dependences.
	Calls map[int]CallDeps
	// IntoThis and IntoParams: what the method stores into its receiver
	// object or into the objects passed as arguments (a constructor, a
	// setter, an append) — applied to the caller's objects at each call.
	IntoThis   Dep
	IntoParams map[int]Dep
	// Widened lists the calls answered by the conservative policy (no
	// single analyzed target and no catalog entry), with the reason.
	Widened []string
	// Incomplete is non-empty when the summary is not to be trusted as
	// complete: the analysis aborted, the summary budget was exceeded, or
	// recursion did not converge. A consumer reports it, never treats the
	// missing dependence as absent.
	Incomplete []string
}

func (s *Summary) equal(o *Summary) bool {
	if o == nil || !equal(s.Return, o.Return) || !equal(s.IntoThis, o.IntoThis) || len(s.IntoParams) != len(o.IntoParams) ||
		len(s.Fields) != len(o.Fields) || len(s.Calls) != len(o.Calls) {
		return false
	}
	for k, v := range s.IntoParams {
		if !equal(v, o.IntoParams[k]) {
			return false
		}
	}
	for k, v := range s.Fields {
		if !equal(v, o.Fields[k]) {
			return false
		}
	}
	for pc, c := range s.Calls {
		oc, ok := o.Calls[pc]
		if !ok || !equal(c.Receiver, oc.Receiver) || len(c.Args) != len(oc.Args) {
			return false
		}
		for i := range c.Args {
			if !equal(c.Args[i], oc.Args[i]) {
				return false
			}
		}
	}
	return true
}

// callEffect is how a call site is answered: the result's dependence and
// what flows into the receiver object (a mutator such as append, a
// constructor), in the caller's terms.
type callEffect struct {
	result, intoReceiver Dep
	intoArgs             map[int]Dep
	widened              string
}

// callAnswer answers one call given its receiver and argument deps.
type callAnswer func(op classfile.LocalOp, recv Dep, args []Dep) callEffect

// decoded is one method's code, decoded once and reused every round.
type decoded struct {
	key   MethodKey
	m     classfile.Method
	cfg   *classfile.CFG
	byPC  map[int]classfile.LocalOp
	order []int
}

type entry struct {
	d     Dep
	width int
}

type state struct {
	stack  []entry
	locals []Dep
}

func (s state) clone() state {
	return state{stack: append([]entry(nil), s.stack...), locals: append([]Dep(nil), s.locals...)}
}

func (s *state) setLocal(i int, d Dep) {
	for i >= len(s.locals) {
		s.locals = append(s.locals, nil)
	}
	s.locals[i] = d
}

func (s *state) local(i int) Dep {
	if i < len(s.locals) {
		return s.locals[i]
	}
	return nil
}

func (s *state) push(d Dep, w int) { s.stack = append(s.stack, entry{d, w}) }

func (s *state) pop() (entry, bool) {
	if len(s.stack) == 0 {
		return entry{}, false
	}
	e := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return e, true
}

func (s *state) popN(n int) ([]entry, bool) {
	if n > len(s.stack) {
		return nil, false
	}
	out := append([]entry(nil), s.stack[len(s.stack)-n:]...)
	s.stack = s.stack[:len(s.stack)-n]
	return out, true
}

// run computes the summary of one decoded method under answer.
func run(dm *decoded, answer callAnswer) *Summary {
	sum := &Summary{Method: dm.key, Fields: map[string]Dep{}, Calls: map[int]CallDeps{}, IntoParams: map[int]Dep{}}
	r := &runner{dm: dm, answer: answer, sum: sum, objs: map[int32]Dep{}, widened: map[string]bool{}, labels: map[int]Dep{}}
	if reason := r.fixpoint(); reason != "" {
		sum.Incomplete = append(sum.Incomplete, dm.key.String()+": "+reason)
	}
	// Allocation labels are internal: expand them to what flowed into
	// the object.
	sum.Return = r.expand(sum.Return)
	sum.IntoThis = r.expand(sum.IntoThis)
	for k, v := range sum.IntoParams {
		sum.IntoParams[k] = r.expand(v)
	}
	for k, v := range sum.Fields {
		sum.Fields[k] = r.expand(v)
	}
	for pc, c := range sum.Calls {
		args := make([]Dep, len(c.Args))
		for i, a := range c.Args {
			args[i] = r.expand(a)
		}
		sum.Calls[pc] = CallDeps{Receiver: r.expand(c.Receiver), Args: args}
	}
	for w := range r.widened {
		sum.Widened = append(sum.Widened, w)
	}
	sort.Strings(sum.Widened)
	return sum
}

type runner struct {
	dm      *decoded
	answer  callAnswer
	sum     *Summary
	objs    map[int32]Dep // allocation label → what flowed into the object
	widened map[string]bool
	labels  map[int]Dep // per-PC constant deps (allocation, field), built once
}

// label returns the constant Dep built by mk for pc, once per run.
func (r *runner) label(pc int, mk func() Dep) Dep {
	if d, ok := r.labels[pc]; ok {
		return d
	}
	d := mk()
	r.labels[pc] = d
	return d
}

var argCountCache sync.Map // descriptor → int

func argCount(desc string) (int, error) {
	if n, ok := argCountCache.Load(desc); ok {
		return n.(int), nil
	}
	a, err := classfile.ArgumentSlots(desc, true)
	if err != nil {
		return 0, err
	}
	argCountCache.Store(desc, len(a))
	return len(a), nil
}

const maxBlockVisits = 50000

func (r *runner) fixpoint() string {
	cfg := r.dm.cfg
	if cfg == nil || len(cfg.Blocks) == 0 {
		return ""
	}
	if len(cfg.Unsupported) > 0 {
		return "jsr/ret not modeled"
	}
	m := r.dm.m
	init := state{locals: make([]Dep, m.MaxLocals)}
	if !m.IsStatic() {
		init.setLocal(0, of(srcThis))
	}
	args, err := classfile.ArgumentSlots(m.Descriptor, m.IsStatic())
	if err != nil {
		return "invalid descriptor: " + err.Error()
	}
	for i, a := range args {
		init.setLocal(a.Slot, of(param(i)))
	}
	in := make([]*state, len(cfg.Blocks))
	in[0] = &init
	queue, queued := []int{0}, make([]bool, len(cfg.Blocks))
	queued[0] = true
	succs := make([][]classfile.CFGEdge, len(cfg.Blocks))
	for _, e := range cfg.Edges {
		succs[e.From] = append(succs[e.From], e)
	}
	// Objects change flow-insensitively; a block is revisited until its
	// entry state and the object map are stable.
	for visits := 0; len(queue) > 0; visits++ {
		if visits > maxBlockVisits {
			return "block-visit budget exceeded"
		}
		bi := queue[0]
		queue = queue[1:]
		queued[bi] = false
		st := in[bi].clone()
		start := *in[bi] // read-only: entry states are replaced, never mutated
		for _, pc := range cfg.Blocks[bi].PCs {
			op, ok := r.dm.byPC[pc]
			if !ok {
				return fmt.Sprintf("pc=%d: no decoded instruction", pc)
			}
			if reason := r.step(&st, op); reason != "" {
				return fmt.Sprintf("pc=%d: %s", pc, reason)
			}
		}
		for _, e := range succs[bi] {
			next := st
			if e.Kind == classfile.EdgeException {
				// The handler may start at any instruction of the block:
				// locals as at entry or exit, the exception on the stack.
				if h := in[e.To]; h != nil && len(h.stack) == 1 && localsCover(h.locals, start.locals) && localsCover(h.locals, st.locals) {
					continue
				}
				next = state{locals: mergeLocals(start.locals, st.locals), stack: []entry{{nil, 1}}}
			}
			if in[e.To] == nil {
				c := next.clone()
				in[e.To] = &c
			} else {
				merged, changed, reason := mergeState(*in[e.To], next)
				if reason != "" {
					return reason
				}
				if !changed {
					continue
				}
				in[e.To] = &merged
			}
			if !queued[e.To] {
				queue = append(queue, e.To)
				queued[e.To] = true
			}
		}
	}
	return ""
}

// localsCover reports whether every local of b is contained in a's.
func localsCover(a, b []Dep) bool {
	if len(b) > len(a) {
		for _, d := range b[len(a):] {
			if len(d) > 0 {
				return false
			}
		}
		b = b[:len(a)]
	}
	for i, d := range b {
		if !subset(d, a[i]) {
			return false
		}
	}
	return true
}

func mergeLocals(a, b []Dep) []Dep {
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	out := make([]Dep, n)
	for i := range out {
		var x, y Dep
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		out[i] = union(x, y)
	}
	return out
}

func mergeState(a, b state) (state, bool, string) {
	if len(a.stack) != len(b.stack) {
		return state{}, false, fmt.Sprintf("control-flow join with stack depths %d and %d", len(a.stack), len(b.stack))
	}
	// A union only grows: a longer set is a changed one. Nothing is
	// allocated unless something changed.
	changed := len(b.locals) > len(a.locals)
	for i := 0; !changed && i < len(a.stack); i++ {
		changed = !subset(b.stack[i].d, a.stack[i].d)
	}
	for i := 0; !changed && i < len(b.locals); i++ {
		changed = !subset(b.locals[i], a.locals[i])
	}
	if !changed {
		return a, false, ""
	}
	st := state{stack: make([]entry, len(a.stack)), locals: mergeLocals(a.locals, b.locals)}
	for i := range a.stack {
		st.stack[i] = entry{union(a.stack[i].d, b.stack[i].d), a.stack[i].width}
	}
	return st, true, ""
}

// absorb records that d flows into every object labelled in into.
func (r *runner) absorb(into, d Dep) {
	if len(d) == 0 {
		return
	}
	for _, id := range into {
		_, k, p := info(id)
		switch k {
		case kAlloc:
			r.objs[id] = union(r.objs[id], d)
		case kThis:
			r.sum.IntoThis = union(r.sum.IntoThis, d)
		case kParam:
			r.sum.IntoParams[p] = union(r.sum.IntoParams[p], d)
		}
	}
}

// expand replaces allocation labels with what flowed into the objects,
// transitively.
func (r *runner) expand(d Dep) Dep {
	var out Dep
	seen := map[int32]bool{}
	var walk func(Dep)
	walk = func(d Dep) {
		var plain []int32
		for _, id := range d {
			if kindOf(id) == kAlloc {
				if !seen[id] {
					seen[id] = true
					walk(r.objs[id])
				}
				continue
			}
			plain = append(plain, id)
		}
		out = union(out, fromIDs(plain))
	}
	walk(d)
	return out
}

func depsOf(es []entry) Dep {
	var d Dep
	for _, e := range es {
		d = union(d, e.d)
	}
	return d
}

func (r *runner) step(st *state, op classfile.LocalOp) string {
	switch op.Kind {
	case classfile.LocalOpNop, classfile.LocalOpIinc:
	case classfile.LocalOpConst:
		st.push(nil, op.Width)
	case classfile.LocalOpLoad:
		st.push(st.local(op.Slot), op.Width)
	case classfile.LocalOpStore:
		e, ok := st.pop()
		if !ok {
			return "stack underflow at store"
		}
		st.setLocal(op.Slot, e.d)
	case classfile.LocalOpArrayLoad:
		es, ok := st.popN(2)
		if !ok {
			return "stack underflow at array load"
		}
		st.push(es[0].d, op.Width) // an element depends on the array (which absorbed its stores)
	case classfile.LocalOpArrayStore:
		es, ok := st.popN(3)
		if !ok {
			return "stack underflow at array store"
		}
		r.absorb(es[0].d, es[2].d)
	case classfile.LocalOpFieldGet:
		f := r.label(op.PC, func() Dep { return of(srcField + op.FieldOwner + "." + op.FieldName) })
		if !op.IsStatic {
			e, ok := st.pop()
			if !ok {
				return "stack underflow at getfield"
			}
			// A field of an object depends on the object too (what
			// flowed into it, e.g. through its constructor).
			f = union(f, e.d)
		}
		st.push(f, op.Width)
	case classfile.LocalOpFieldPut:
		n := 1
		if !op.IsStatic {
			n = 2
		}
		es, ok := st.popN(n)
		if !ok {
			return "stack underflow at putfield"
		}
		v := es[len(es)-1].d
		k := op.FieldOwner + "." + op.FieldName
		r.sum.Fields[k] = union(r.sum.Fields[k], v)
		if !op.IsStatic {
			r.absorb(es[0].d, v)
		}
	case classfile.LocalOpNew:
		st.push(r.label(op.PC, func() Dep { return of(srcAlloc + strconv.Itoa(op.PC)) }), 1)
	case classfile.LocalOpNewArray:
		if _, ok := st.pop(); !ok {
			return "stack underflow at newarray"
		}
		st.push(r.label(op.PC, func() Dep { return of(srcAlloc + strconv.Itoa(op.PC)) }), 1)
	case classfile.LocalOpMultiNewArray:
		if _, ok := st.popN(op.Dims); !ok {
			return "stack underflow at multianewarray"
		}
		st.push(r.label(op.PC, func() Dep { return of(srcAlloc + strconv.Itoa(op.PC)) }), 1)
	case classfile.LocalOpArrayLength, classfile.LocalOpCheckCast, classfile.LocalOpInstanceOf, classfile.LocalOpConvert, classfile.LocalOpUnaryNumeric:
		e, ok := st.pop()
		if !ok {
			return "stack underflow"
		}
		w := op.Width
		if op.Kind == classfile.LocalOpArrayLength || op.Kind == classfile.LocalOpInstanceOf || op.Kind == classfile.LocalOpCheckCast {
			w = 1
		}
		st.push(e.d, w)
	case classfile.LocalOpBinaryNumeric, classfile.LocalOpCompare:
		es, ok := st.popN(2)
		if !ok {
			return "stack underflow at a binary operation"
		}
		w := op.Width
		if op.Kind == classfile.LocalOpCompare {
			w = 1
		}
		st.push(depsOf(es), w)
	case classfile.LocalOpMonitor, classfile.LocalOpThrow:
		if _, ok := st.pop(); !ok {
			return "stack underflow"
		}
	case classfile.LocalOpReturn:
		if op.Width > 0 {
			e, ok := st.pop()
			if !ok {
				return "stack underflow at return"
			}
			r.sum.Return = union(r.sum.Return, e.d)
		}
	case classfile.LocalOpBranch:
		if _, ok := st.popN(op.Width); !ok {
			return "stack underflow at branch"
		}
	case classfile.LocalOpCall:
		return r.stepCall(st, op)
	case classfile.LocalOpPop:
		if _, ok := st.pop(); !ok {
			return "stack underflow at pop"
		}
	case classfile.LocalOpPop2:
		e, ok := st.pop()
		if !ok {
			return "stack underflow at pop2"
		}
		if e.width == 1 {
			if _, ok := st.pop(); !ok {
				return "stack underflow at pop2"
			}
		}
	case classfile.LocalOpDup:
		e, ok := st.pop()
		if !ok {
			return "stack underflow at dup"
		}
		st.stack = append(st.stack, e, e)
	case classfile.LocalOpSwap:
		es, ok := st.popN(2)
		if !ok {
			return "stack underflow at swap"
		}
		st.stack = append(st.stack, es[1], es[0])
	case classfile.LocalOpDupX1, classfile.LocalOpDupX2, classfile.LocalOpDup2, classfile.LocalOpDup2X1, classfile.LocalOpDup2X2:
		return stepDupWide(st, op.Kind)
	default:
		return "unsupported instruction " + string(op.Kind)
	}
	return ""
}

// stepDupWide applies the dup_x/dup2 forms by category (JVMS §6.5).
func stepDupWide(st *state, k classfile.LocalOpKind) string {
	// Gather the top words as entries, apply the form on word counts.
	words := map[classfile.LocalOpKind][2]int{ // [words duplicated, words skipped beneath]
		classfile.LocalOpDupX1: {1, 1}, classfile.LocalOpDupX2: {1, 2},
		classfile.LocalOpDup2: {2, 0}, classfile.LocalOpDup2X1: {2, 1}, classfile.LocalOpDup2X2: {2, 2},
	}[k]
	take := func(n int) ([]entry, bool) {
		var out []entry
		for n > 0 {
			e, ok := st.pop()
			if !ok || e.width > n {
				return nil, false
			}
			out = append([]entry{e}, out...)
			n -= e.width
		}
		return out, true
	}
	top, ok := take(words[0])
	if !ok {
		return "stack shape does not fit " + string(k)
	}
	below, ok := take(words[1])
	if !ok {
		return "stack shape does not fit " + string(k)
	}
	st.stack = append(st.stack, top...)
	st.stack = append(st.stack, below...)
	st.stack = append(st.stack, top...)
	return ""
}

func (r *runner) stepCall(st *state, op classfile.LocalOp) string {
	n, err := argCount(op.CallDesc)
	if err != nil {
		return "invalid call descriptor " + op.CallDesc
	}
	if op.CallHasReceiver {
		n++
	}
	es, ok := st.popN(n)
	if !ok {
		return "stack underflow at a call"
	}
	var recv Dep
	argEs := es
	if op.CallHasReceiver {
		recv, argEs = es[0].d, es[1:]
	}
	args := make([]Dep, len(argEs))
	for i, e := range argEs {
		args[i] = e.d
	}
	prev, seen := r.sum.Calls[op.PC]
	grow := !seen || !subset(recv, prev.Receiver)
	for i := 0; !grow && i < len(args); i++ {
		grow = i >= len(prev.Args) || !subset(args[i], prev.Args[i])
	}
	if grow {
		merged := CallDeps{Receiver: union(prev.Receiver, recv), Args: make([]Dep, len(args))}
		for i := range args {
			var p Dep
			if i < len(prev.Args) {
				p = prev.Args[i]
			}
			merged.Args[i] = union(p, args[i])
		}
		r.sum.Calls[op.PC] = merged
	}
	eff := r.answer(op, recv, args)
	if eff.widened != "" {
		r.widened[eff.widened] = true
	}
	r.absorb(recv, eff.intoReceiver)
	for i, d := range eff.intoArgs {
		if i < len(args) {
			r.absorb(args[i], d)
		}
	}
	if op.Width > 0 {
		st.push(eff.result, op.Width)
	}
	return ""
}
