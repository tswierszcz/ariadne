package flowsum

import (
	"fmt"
	"runtime/debug"
	"sort"
	"strconv"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Options bounds the summarization.
type Options struct {
	// MaxMethods bounds how many methods are summarized from the roots
	// (callees beyond it get the conservative policy, recorded as
	// widened with the reason "summary budget").
	MaxMethods int
	// MaxRounds bounds the fixed point: at most MaxRounds evaluations per
	// method on average; recursion that has not converged by then marks
	// every summary incomplete.
	MaxRounds int
}

// DefaultOptions are the CLI's bounds.
var DefaultOptions = Options{MaxMethods: 20000, MaxRounds: 30}

// Stats describe one Summarize run.
type Stats struct {
	Methods, Evaluations, Widened, Incomplete, BudgetCut int
	Converged                                            bool
}

// Engine summarizes methods of one hierarchy.
type Engine struct {
	h        *hierarchy.Hierarchy
	opts     Options
	jr       classfile.JarClassReader
	decoded  map[MethodKey]*decoded
	failed   map[MethodKey]string
	sums     map[MethodKey]*Summary
	set      []MethodKey
	inSet    map[MethodKey]bool
	targets  map[string]target
	budgeted map[MethodKey]bool
}

type target struct {
	key    MethodKey
	ok     bool
	reason string
}

// New returns an engine over h.
func New(h *hierarchy.Hierarchy, opts Options) *Engine {
	return &Engine{h: h, opts: opts, decoded: map[MethodKey]*decoded{}, failed: map[MethodKey]string{},
		sums: map[MethodKey]*Summary{}, inSet: map[MethodKey]bool{}, targets: map[string]target{}, budgeted: map[MethodKey]bool{}}
}

// Close releases the jars read.
func (e *Engine) Close() error { return e.jr.Close() }

// Summary returns the summary of k after Summarize (nil if k was not
// summarized).
func (e *Engine) Summary(k MethodKey) *Summary { return e.sums[k] }

func (e *Engine) decode(k MethodKey) (*decoded, string) {
	if d, ok := e.decoded[k]; ok {
		return d, ""
	}
	if r, ok := e.failed[k]; ok {
		return nil, r
	}
	fail := func(r string) (*decoded, string) { e.failed[k] = r; return nil, r }
	ci, ok := e.h.Lookup(k.Owner)
	if !ok {
		return fail("class not in the input")
	}
	m, ok := ci.MethodByNameAndDesc(k.Name, k.Desc)
	if !ok || len(m.Code) == 0 {
		return fail("no code")
	}
	cf, err := e.jr.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		return fail("re-reading the class: " + err.Error())
	}
	// The re-read method carries the code attribute the decoder needs.
	var cm *classfile.Method
	for i := range cf.Methods {
		if cf.Methods[i].Name == k.Name && cf.Methods[i].Descriptor == k.Desc {
			cm = &cf.Methods[i]
		}
	}
	if cm == nil || len(cm.Code) == 0 {
		return fail("no code in the re-read class")
	}
	cfg, err := classfile.BuildCFG(cm.Code, cm.ExceptionTable)
	if err != nil {
		return fail("CFG: " + err.Error())
	}
	ops, err := cf.DecodeLocalOps(cm.Code)
	if err != nil {
		return fail("decoding: " + err.Error())
	}
	d := &decoded{key: k, m: *cm, cfg: cfg, byPC: make(map[int]classfile.LocalOp, len(ops))}
	for _, op := range ops {
		d.byPC[op.PC] = op
		d.order = append(d.order, op.PC)
	}
	e.decoded[k] = d
	return d, ""
}

// resolve finds the single analyzed method a call runs: the exact target
// of invokestatic/invokespecial, and for invokevirtual/invokeinterface the
// one implementation among the static type and its subtypes in the input
// (CHA); anything else is not resolved, with the reason.
func (e *Engine) resolve(op classfile.LocalOp) target {
	key := string(op.CallKind) + " " + op.CallOwner + "#" + op.CallName + op.CallDesc
	if t, ok := e.targets[key]; ok {
		return t
	}
	t := e.resolveUncached(op)
	e.targets[key] = t
	return t
}

func (e *Engine) hasCode(k MethodKey) bool {
	ci, ok := e.h.Lookup(k.Owner)
	if !ok {
		return false
	}
	m, ok := ci.MethodByNameAndDesc(k.Name, k.Desc)
	return ok && len(m.Code) > 0
}

func (e *Engine) resolveUncached(op classfile.LocalOp) target {
	switch op.CallKind {
	case classfile.CallDynamic:
		return target{reason: "invokedynamic"}
	case classfile.CallStatic, classfile.CallSpecial:
		decl, ok := e.h.ResolveExact(op.CallOwner, op.CallName, op.CallDesc)
		if !ok {
			return target{reason: "target not in the input"}
		}
		k := MethodKey{decl, op.CallName, op.CallDesc}
		if !e.hasCode(k) {
			return target{reason: "target has no code"}
		}
		return target{key: k, ok: true}
	}
	if _, ok := e.h.Lookup(op.CallOwner); !ok {
		return target{reason: "receiver type not in the input"}
	}
	found := map[MethodKey]bool{}
	for _, c := range append([]string{op.CallOwner}, e.h.Subtypes(op.CallOwner)...) {
		ci, ok := e.h.Lookup(c)
		if !ok || ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) != 0 {
			continue
		}
		decl, ok := e.h.ResolveExact(c, op.CallName, op.CallDesc)
		if !ok {
			return target{reason: "an implementation outside the input (" + c + ")"}
		}
		k := MethodKey{decl, op.CallName, op.CallDesc}
		if !e.hasCode(k) {
			return target{reason: "an implementation without code (" + k.String() + ")"}
		}
		found[k] = true
	}
	switch len(found) {
	case 0:
		return target{reason: "no implementation in the input"}
	case 1:
		for k := range found {
			return target{key: k, ok: true}
		}
	}
	return target{reason: strconv.Itoa(len(found)) + " implementations (CHA)"}
}

func (e *Engine) answer() callAnswer {
	return func(op classfile.LocalOp, recv Dep, args []Dep) callEffect {
		if op.CallKind == classfile.CallDynamic {
			// invokedynamic: a lambda captures its arguments; any other
			// bootstrap (string concatenation above all) is taken to
			// produce a value from its arguments only.
			eff := callEffect{result: flow{args: true}.apply(nil, args)}
			if op.SamInterface == "" {
				eff.widened = "invokedynamic " + op.CallDesc + ": non-lambda bootstrap, result taken to depend on its arguments"
			}
			return eff
		}
		if c, ok := lookupCatalog(op.CallOwner, op.CallName); ok {
			return callEffect{result: c.result.apply(recv, args), intoReceiver: c.intoRecv.apply(recv, args)}
		}
		name := op.CallOwner + "#" + op.CallName + op.CallDesc
		t := e.resolve(op)
		reason := t.reason
		if t.ok {
			if s := e.sums[t.key]; s != nil && e.inSet[t.key] {
				eff := callEffect{result: substitute(s.Return, recv, args), intoReceiver: substitute(s.IntoThis, recv, args)}
				if len(s.IntoParams) > 0 {
					eff.intoArgs = map[int]Dep{}
					for i, d := range s.IntoParams {
						eff.intoArgs[i] = substitute(d, recv, args)
					}
				}
				return eff
			}
			if e.inSet[t.key] {
				// Not summarized yet in this round: bottom (the fixed
				// point adds what it returns in a later round).
				return callEffect{}
			}
			reason = "summary budget"
		}
		// The conservative policy: the result depends on everything the
		// call sees and is itself an unmodeled source; the receiver may
		// keep its arguments.
		eff := callEffect{result: union(recv, of(srcCall+name)), widened: name + ": " + reason}
		for _, a := range args {
			eff.result = union(eff.result, a)
		}
		if op.CallHasReceiver {
			for _, a := range args {
				eff.intoReceiver = union(eff.intoReceiver, a)
			}
		}
		return eff
	}
}

// Summarize summarizes roots and every method they reach through
// resolved calls, within the budget, to a fixed point: a worklist that
// starts from the deepest callees and re-evaluates a method's callers
// only when its summary changed.
func (e *Engine) Summarize(roots []MethodKey) Stats {
	var st Stats
	queue := append([]MethodKey(nil), roots...)
	sort.Slice(queue, func(i, j int) bool { return queue[i].String() < queue[j].String() })
	callees := map[MethodKey][]MethodKey{}
	for len(queue) > 0 {
		k := queue[0]
		queue = queue[1:]
		if e.inSet[k] {
			continue
		}
		if e.opts.MaxMethods > 0 && len(e.set) >= e.opts.MaxMethods {
			e.budgeted[k] = true
			continue
		}
		dm, reason := e.decode(k)
		if dm == nil {
			e.sums[k] = &Summary{Method: k, Incomplete: []string{k.String() + ": " + reason}}
			continue
		}
		e.inSet[k] = true
		e.set = append(e.set, k)
		for _, pc := range dm.order {
			op := dm.byPC[pc]
			if op.Kind != classfile.LocalOpCall || op.CallKind == classfile.CallDynamic {
				continue
			}
			if _, ok := lookupCatalog(op.CallOwner, op.CallName); ok {
				continue
			}
			if t := e.resolve(op); t.ok {
				callees[k] = append(callees[k], t.key)
				if !e.inSet[t.key] {
					queue = append(queue, t.key)
				}
			}
		}
	}
	st.BudgetCut = len(e.budgeted)
	callers := map[MethodKey][]MethodKey{}
	for k, cs := range callees {
		for _, c := range cs {
			if e.inSet[c] {
				callers[c] = append(callers[c], k)
			}
		}
	}
	for _, cs := range callers {
		sort.Slice(cs, func(i, j int) bool { return cs[i].String() < cs[j].String() })
	}
	rounds := e.opts.MaxRounds
	if rounds <= 0 {
		rounds = 30
	}
	// The fixed point allocates many short-lived sets while the analyzed
	// hierarchy keeps a large heap alive: collect less often meanwhile.
	defer debug.SetGCPercent(debug.SetGCPercent(400))
	// Strongly connected components of the call relation, callees first
	// (Tarjan emits them in that order): a method outside recursion is
	// evaluated once, after everything it calls; a recursive component is
	// iterated until stable, at most MaxRounds evaluations per member.
	st.Converged = true
	for _, comp := range sccs(e.set, callees, e.inSet) {
		budget := rounds * len(comp)
		inComp := map[MethodKey]bool{}
		for _, k := range comp {
			inComp[k] = true
		}
		work := append([]MethodKey(nil), comp...)
		queued := map[MethodKey]bool{}
		for _, k := range comp {
			queued[k] = true
		}
		evals := 0
		for len(work) > 0 {
			if evals >= budget {
				st.Converged = false
				for _, k := range comp {
					e.sums[k].Incomplete = append(e.sums[k].Incomplete, fmt.Sprintf("recursion did not converge within %d evaluations", budget))
				}
				break
			}
			// Smallest method first: callees settle before a large caller
			// is re-evaluated (fewer evaluations of the costly ones).
			best := 0
			for i := 1; i < len(work); i++ {
				a, b := len(e.decoded[work[i]].order), len(e.decoded[work[best]].order)
				if a < b || a == b && work[i].String() < work[best].String() {
					best = i
				}
			}
			k := work[best]
			work = append(work[:best], work[best+1:]...)
			queued[k] = false
			evals++
			s := run(e.decoded[k], e.answer())
			old := e.sums[k]
			e.sums[k] = s
			if old != nil && s.equal(old) {
				continue
			}
			for _, c := range callers[k] {
				if inComp[c] && !queued[c] {
					queued[c] = true
					work = append(work, c)
				}
			}
		}
		st.Evaluations += evals
	}
	// An incomplete callee makes its callers incomplete, transitively.
	for changed := true; changed; {
		changed = false
		for _, k := range e.set {
			s := e.sums[k]
			for _, c := range callees[k] {
				cs := e.sums[c]
				note := "callee " + c.String() + " incomplete"
				if c != k && cs != nil && len(cs.Incomplete) > 0 && !contains(s.Incomplete, note) {
					s.Incomplete = append(s.Incomplete, note)
					sort.Strings(s.Incomplete)
					changed = true
				}
			}
		}
	}
	for _, k := range e.set {
		s := e.sums[k]
		st.Methods++
		st.Widened += len(s.Widened)
		if len(s.Incomplete) > 0 {
			st.Incomplete++
		}
	}
	return st
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// sccs returns the strongly connected components of the call relation
// over set, callees before callers (Tarjan), deterministically.
func sccs(set []MethodKey, callees map[MethodKey][]MethodKey, in map[MethodKey]bool) [][]MethodKey {
	index := map[MethodKey]int{}
	low := map[MethodKey]int{}
	onStack := map[MethodKey]bool{}
	var stack []MethodKey
	var out [][]MethodKey
	next := 0
	type frame struct {
		k MethodKey
		i int
	}
	for _, root := range set {
		if _, seen := index[root]; seen {
			continue
		}
		// Iterative Tarjan: deep call chains must not exhaust the stack.
		frames := []frame{{root, 0}}
		index[root], low[root] = next, next
		next++
		stack = append(stack, root)
		onStack[root] = true
		for len(frames) > 0 {
			f := &frames[len(frames)-1]
			cs := callees[f.k]
			if f.i < len(cs) {
				c := cs[f.i]
				f.i++
				if !in[c] {
					continue
				}
				if _, seen := index[c]; !seen {
					index[c], low[c] = next, next
					next++
					stack = append(stack, c)
					onStack[c] = true
					frames = append(frames, frame{c, 0})
				} else if onStack[c] && index[c] < low[f.k] {
					low[f.k] = index[c]
				}
				continue
			}
			k := f.k
			frames = frames[:len(frames)-1]
			if len(frames) > 0 {
				p := frames[len(frames)-1].k
				if low[k] < low[p] {
					low[p] = low[k]
				}
			}
			if low[k] == index[k] {
				var comp []MethodKey
				for {
					top := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					onStack[top] = false
					comp = append(comp, top)
					if top == k {
						break
					}
				}
				out = append(out, comp)
			}
		}
	}
	return out
}
