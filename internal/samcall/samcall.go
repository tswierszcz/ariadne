// Package samcall is DYNAMIC_LINKAGE L1's SAM-call materialization: a
// lambda or method reference (invokedynamic through LambdaMetafactory) is
// created in one method and run in another — passed as an argument,
// returned, stored in a field — through a call of its functional
// interface's method (Supplier#get, Runnable#run, a library callback). The
// call graph already links the creating method to the implementation (the
// invokedynamic's own edge); this package links the *calling* site.
//
// Two kinds of site:
//
//   - sam-call: an invokeinterface/invokevirtual of a functional
//     interface's abstract method (an interface some lambda of the input
//     implements, or a supertype of one) whose receiver DATA_FLOW
//     resolves to functional values — in the same method, through the
//     method's callers (dataflow.CallerIndex.ResolveParameter, with the
//     values a lambda body captured from its creator), a field
//     (ResolveFinalField / ResolveField) or the return of a static/private
//     helper. The edge goes to each value's implementation.
//   - escape: a functional value passed as an argument to a method outside
//     the analyzed input (List#forEach, Map#computeIfAbsent,
//     Optional#orElseGet…): the code that may run it is not in the input,
//     so the edge goes from the call site that hands it over. Only values
//     created in *another* method add anything — the creating method
//     already has the invokedynamic edge.
//
// Nothing falls back to "every lambda of that interface": a receiver the
// analysis cannot resolve contributes no edge and is counted, with its
// reason, as a gap (Stats). Possible, not certain: like a CHA edge, an
// edge here says the body may run at that site.
package samcall

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// Version names the rules (site selection, resolution limits); a snapshot
// records it.
const Version = "l1-1"

// Resolution limits. Context-insensitive: a parameter gets the union of
// what every caller passes, so a generic helper called from many places
// (ReflectionUtils#doWithMethods, Lazy#of) gets every lambda any of them
// passes — bounded by MaxCallers per method, past which the site is a gap.
const (
	MaxCallers = 64
	// MaxChain bounds the nesting of links (a caller level, a field, a
	// helper's return) chased for one value — a recursion guard: every
	// link's outcome is computed once and memoized, so the cost does not
	// depend on it. Past it the site has a depth gap.
	MaxChain = 64
	// MaxWork bounds the outcomes (parameters, fields, helper returns)
	// computed for one site; past it the site has a budget gap.
	MaxWork = 20000
)

// Kind of a site.
type Kind string

const (
	// KindCall is an interface/virtual call of a functional interface's method.
	KindCall Kind = "sam-call"
	// KindEscape is a functional value handed to a method outside the input.
	KindEscape Kind = "escape"
)

// Site is one candidate call site.
type Site struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Owner/Name/Desc: the call as the bytecode names it.
	Owner, Name, Desc string
	CallKind          classfile.CallKind
	Kind              Kind
	// Arg is, for KindEscape, the argument (0-based, receiver excluded)
	// whose declared type is a functional interface; -1 for KindCall.
	Arg int
}

// Target is one implementation a site may run.
type Target struct {
	Owner, Name, Desc string
	RefKind           byte
	// Creators are the methods whose invokedynamic created the values
	// reaching the site with this implementation: the edge is real only
	// once one of them runs (callgraph.SAMGate).
	Creators []dataflow.MethodID
}

type targetKey struct {
	owner, name, desc string
	refKind           byte
}

// Virtual reports whether the implementation is selected by the receiver
// at run time (an unbound or bound instance-method reference to a method
// that may be overridden), so the edge must be dispatch-expanded.
func (t Target) Virtual() bool {
	return t.RefKind == classfile.RefInvokeVirtual || t.RefKind == classfile.RefInvokeInterface
}

// Result is one Site with what it resolved to.
type Result struct {
	Site    Site
	Targets []Target
	// Gaps: why the target set may be incomplete, by category (Gap*),
	// each with one example reason.
	Gaps map[string]string
	// Objects counts values that are ordinary objects (an anonymous or
	// named class implementing the interface): CHA's dispatch covers them
	// where the receiver type is expanded; they are not this model's.
	Objects int
}

// Gap categories.
const (
	GapParameter  = "parameter"     // a caller of the method could not be seen or resolved
	GapSAMArg     = "sam-argument"  // the value is itself an argument of a functional-interface call
	GapField      = "field"         // a field whose writers could not be resolved
	GapCallResult = "call-result"   // the return of a static/private helper could not be resolved
	GapInstance   = "instance-call" // the result of an instance call (not followed)
	GapValue      = "unknown-value" // the local analysis has no value (array element, cast of unknown…)
	GapDepth      = "depth"         // a chain longer than the limits
	GapBootstrap  = "bootstrap"     // a lambda whose implementation was not resolved
	GapBudget     = "budget"        // the site's resolution exceeded MaxWork
)

// Find lists the candidate sites of h: calls of a SAM method on an
// interface some lambda/method reference of the input implements (or a
// supertype of one), and calls to a method outside the input with an
// argument of such an interface type.
func Find(h *hierarchy.Hierarchy) []Site {
	sams := samMethods(h)
	if len(sams) == 0 {
		return nil
	}
	// (owner, name) pairs a SAM call can name: each functional interface
	// with its abstract method, and every supertype interface the method
	// may be declared on (Function#apply for a UnaryOperator).
	type ckey struct{ owner, name string }
	samPairs := map[ckey]bool{}
	for i, names := range sams {
		supers := superInterfaces(h, i)
		for n := range names {
			samPairs[ckey{i, n}] = true
			for _, sup := range supers {
				samPairs[ckey{sup, n}] = true
			}
		}
	}
	samCall := func(owner, name string) bool { return samPairs[ckey{owner, name}] }
	outside := map[string]bool{} // owner#name+desc -> target not in the input
	escapes := func(c classfile.Call) bool {
		k := c.Owner + "#" + c.Name + c.Desc
		if v, ok := outside[k]; ok {
			return v
		}
		v := false
		if _, in := h.Lookup(c.Owner); !in {
			v = true
		} else if _, ok := h.ResolveExact(c.Owner, c.Name, c.Desc); !ok {
			v = true // inherited from a class outside the input
		}
		outside[k] = v
		return v
	}

	var out []Site
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind == classfile.CallDynamic || c.Kind == classfile.CallInit {
					continue
				}
				base := Site{CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC,
					Owner: c.Owner, Name: c.Name, Desc: c.Desc, CallKind: c.Kind, Arg: -1}
				if (c.Kind == classfile.CallInterface || c.Kind == classfile.CallVirtual) && samCall(c.Owner, c.Name) {
					s := base
					s.Kind = KindCall
					out = append(out, s)
				}
				params := functionalParams(c.Desc, sams)
				if len(params) == 0 || !escapes(c) {
					continue
				}
				for _, i := range params {
					s := base
					s.Kind, s.Arg = KindEscape, i
					out = append(out, s)
				}
			}
		}
	}
	return out
}

// samMethods maps each functional interface some lambda/method reference
// of h implements to its abstract method names.
func samMethods(h *hierarchy.Hierarchy) map[string]map[string]bool {
	sams := map[string]map[string]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind == classfile.CallDynamic && c.SamInterface != "" {
					if sams[c.SamInterface] == nil {
						sams[c.SamInterface] = map[string]bool{}
					}
					sams[c.SamInterface][c.SamMethod] = true
				}
			}
		}
	}
	return sams
}

// holderTypes are the declared types a lambda of h can be stored as: its
// functional interfaces, their superinterfaces, and java.lang.Object. A
// parameter or field of any other type never holds one — the verifier
// guarantees it — so it is not followed (and is not a gap).
func holderTypes(h *hierarchy.Hierarchy, sams map[string]map[string]bool) map[string]bool {
	out := map[string]bool{"java.lang.Object": true}
	for i := range sams {
		out[i] = true
		for _, s := range superInterfaces(h, i) {
			out[s] = true
		}
	}
	return out
}

// jdkFunctionalSupers: functional interfaces of the JDK (not in the input,
// so not in the hierarchy) whose abstract method is inherited from a
// supertype.
var jdkFunctionalSupers = map[string]string{
	"java.util.function.UnaryOperator":  "java.util.function.Function",
	"java.util.function.BinaryOperator": "java.util.function.BiFunction",
}

// superInterfaces lists iface's transitive superinterfaces: through the
// hierarchy for an interface of the input, jdkFunctionalSupers otherwise.
func superInterfaces(h *hierarchy.Hierarchy, iface string) []string {
	var out []string
	seen := map[string]bool{iface: true}
	queue := []string{iface}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		next := []string{}
		if ci, ok := h.Lookup(cur); ok {
			next = append(next, ci.Interfaces...)
		} else if sup, ok := jdkFunctionalSupers[cur]; ok {
			next = append(next, sup)
		}
		for _, n := range next {
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
				queue = append(queue, n)
			}
		}
	}
	return out
}

// functionalParams returns the indexes of desc's parameters whose type is
// one of the functional interfaces of sams.
func functionalParams(desc string, sams map[string]map[string]bool) []int {
	if !strings.Contains(desc, "L") {
		return nil
	}
	types, err := classfile.ParamDescriptors(desc)
	if err != nil {
		return nil
	}
	var out []int
	for i, t := range types {
		if len(t) > 2 && t[0] == 'L' {
			if _, ok := sams[strings.ReplaceAll(t[1:len(t)-1], "/", ".")]; ok {
				out = append(out, i)
			}
		}
	}
	return out
}

// Resolver resolves sites' functional values, memoizing parameter, field
// and helper-return resolutions across sites (the same method's parameter
// is asked for by every SAM call in it and by every caller chain through
// it).
type Resolver struct {
	h      *hierarchy.Hierarchy
	idx    *dataflow.CallerIndex
	fields map[valueflow.FieldRef]fieldValues
	calls  map[string]callValues
	nests  map[string][]string
	// outcomes memoizes what a parameter, a field or a helper's returns
	// resolve to (lambdas and gaps); running marks the ones being
	// computed, to cut cycles.
	outcomes map[string]*acc
	running  map[string]bool
	// provisional lists outcomes memoized although cut by a cycle (or the
	// budget): valid for the rest of the current site only (the cycle's
	// head adds what they miss), dropped when the site is done —
	// recomputing them at every visit is exponential in recursive code.
	provisional []string
	holders     map[string]bool
	// work counts the outcomes computed for the current site (MaxWork).
	work int
}

type fieldValues struct {
	values  []ctxValue
	unknown []string
}

// ctxValue is a value with the method it was read in (a formal
// parameter in it is resolved through that method's callers).
type ctxValue struct {
	v   valueflow.Value
	ctx dataflow.MethodID
}

type callValues struct {
	values  []valueflow.Value // in the callee's terms (parameter refs kept)
	callee  dataflow.MethodID // the declaring method
	unknown string
}

// NewResolver builds the caller index once for h.
func NewResolver(h *hierarchy.Hierarchy) *Resolver {
	idx := dataflow.NewCallerIndex(h)
	idx.SetOptions(dataflow.CallerOptions{MaxCallers: MaxCallers, FollowCaptures: true, VirtualCallers: true, Direct: true, MaxFacts: 4096})
	return &Resolver{h: h, idx: idx, fields: map[valueflow.FieldRef]fieldValues{}, calls: map[string]callValues{},
		outcomes: map[string]*acc{}, running: map[string]bool{}, holders: holderTypes(h, samMethods(h))}
}

// Close releases the jars read.
func (r *Resolver) Close() error { return r.idx.Close() }

// resolveBatch is how many distinct site methods are analyzed (in
// parallel) and resolved together before their facts are dropped: the
// facts of every method at once cost hundreds of MB on a real
// application.
const resolveBatch = 1024

// Resolve resolves every site, in order.
func (r *Resolver) Resolve(sites []Site) []Result {
	out := make([]Result, 0, len(sites))
	for start := 0; start < len(sites); {
		var methods []dataflow.MethodID
		seen := map[dataflow.MethodID]bool{}
		end := start
		for ; end < len(sites); end++ {
			s := sites[end]
			m := dataflow.MethodID{Owner: s.CallerFQCN, Name: s.CallerName, Desc: s.CallerDesc}
			if !seen[m] {
				if len(methods) == resolveBatch {
					break
				}
				seen[m] = true
				methods = append(methods, m)
			}
		}
		r.idx.Prefetch(methods, runtime.GOMAXPROCS(0))
		for _, s := range sites[start:end] {
			out = append(out, r.resolveSite(s))
		}
		r.idx.Forget(methods)
		start = end
	}
	return out
}

type acc struct {
	targets map[targetKey]map[dataflow.MethodID]bool
	gaps    map[string]string
	objects int
	// cut names the outcomes still being computed that this one stopped
	// at (a cycle): it is complete only once they are.
	cut map[string]bool
}

const budgetCut = "#budget"

func newAcc() *acc {
	return &acc{targets: map[targetKey]map[dataflow.MethodID]bool{}, gaps: map[string]string{}, cut: map[string]bool{}}
}

func (a *acc) gap(cat, reason string) {
	if _, ok := a.gaps[cat]; !ok {
		a.gaps[cat] = reason
	}
}

func (a *acc) merge(b *acc) {
	for k, cs := range b.targets {
		if a.targets[k] == nil {
			a.targets[k] = map[dataflow.MethodID]bool{}
		}
		for c := range cs {
			a.targets[k][c] = true
		}
	}
	for c, r := range b.gaps {
		a.gap(c, r)
	}
	a.objects += b.objects
	for k := range b.cut {
		a.cut[k] = true
	}
}

// memo adds the outcome named key, computing it once. A key met again
// while it is computed contributes nothing there (its own computation
// adds those values: the union is the fixed point).
func (r *Resolver) memo(key string, depth int, a *acc, compute func(b *acc, depth int)) {
	if out, ok := r.outcomes[key]; ok {
		a.merge(out)
		return
	}
	if r.running[key] {
		a.cut[key] = true
		return
	}
	if depth > MaxChain {
		a.gap(GapDepth, fmt.Sprintf("value chain longer than %d parameters/fields/helper returns", MaxChain))
		return
	}
	if r.work >= MaxWork {
		a.gap(GapBudget, fmt.Sprintf("resolution stopped after %d parameters/fields/helper returns", MaxWork))
		a.cut[budgetCut] = true // never memoized past this site
		return
	}
	r.work++
	r.running[key] = true
	b := newAcc()
	compute(b, depth+1)
	delete(r.running, key)
	delete(b.cut, key)
	r.outcomes[key] = b
	if len(b.cut) > 0 {
		r.provisional = append(r.provisional, key)
	}
	a.merge(b)
}

func (r *Resolver) resolveSite(s Site) Result {
	res := Result{Site: s, Gaps: map[string]string{}}
	caller := dataflow.MethodID{Owner: s.CallerFQCN, Name: s.CallerName, Desc: s.CallerDesc}
	facts, reason := r.idx.Facts(caller)
	if reason != "" {
		res.Gaps[GapValue] = reason
		return res
	}
	if aborted, why := facts.Aborted(); aborted {
		res.Gaps[GapValue] = "local-value analysis aborted: " + why
		return res
	}
	site := bytecodeir.Site{Method: s.CallerFQCN + "#" + s.CallerName, Descriptor: s.CallerDesc, PC: s.PC, Op: bytecodeir.OpCall}
	var v valueflow.Value
	if s.Kind == KindCall {
		v = facts.ReceiverAt(site)
	} else {
		v = facts.ArgumentAt(site, s.Arg)
	}
	a := newAcc()
	r.work = 0
	r.value(v, caller, 0, a)
	for _, k := range r.provisional {
		delete(r.outcomes, k)
	}
	r.provisional = r.provisional[:0]
	res.Gaps = a.gaps
	for k, cs := range a.targets {
		t := Target{Owner: k.owner, Name: k.name, Desc: k.desc, RefKind: k.refKind}
		for c := range cs {
			t.Creators = append(t.Creators, c)
		}
		sort.Slice(t.Creators, func(i, j int) bool { return t.Creators[i].String() < t.Creators[j].String() })
		res.Targets = append(res.Targets, t)
	}
	sort.Slice(res.Targets, func(i, j int) bool {
		x, y := res.Targets[i], res.Targets[j]
		if x.Owner != y.Owner {
			return x.Owner < y.Owner
		}
		if x.Name != y.Name {
			return x.Name < y.Name
		}
		return x.Desc < y.Desc
	})
	res.Objects = a.objects
	return res
}

// value adds what v (a value read in method ctx) may be: lambdas to
// a.targets, unresolved contributions to a.gaps.
func (r *Resolver) value(v valueflow.Value, ctx dataflow.MethodID, depth int, a *acc) {
	switch v.Kind {
	case valueflow.KindFunctional:
		f := v.Func
		if f.ImplOwner == "" {
			a.gap(GapBootstrap, "a "+f.SamInterface+" created by an unresolved bootstrap")
			return
		}
		k := targetKey{f.ImplOwner, f.ImplName, f.ImplDesc, f.RefKind}
		if a.targets[k] == nil {
			a.targets[k] = map[dataflow.MethodID]bool{}
		}
		if creator, ok := siteMethod(f.Site); ok {
			a.targets[k][creator] = true
		}
	case valueflow.KindSet:
		for _, c := range v.Candidates {
			r.value(c, ctx, depth, a)
		}
	case valueflow.KindNull:
	case valueflow.KindAlloc:
		a.objects++
	case valueflow.KindTyped:
		if v.Parameter != nil && !r.holders[descType(v.Type)] {
			a.objects++ // a parameter whose declared type cannot hold a lambda
			return
		}
		if v.Parameter == nil {
			a.gap(GapValue, "a value of type "+v.Type+" with no known origin")
			return
		}
		index := v.Parameter.Index
		r.memo(fmt.Sprintf("p:%s/%d", ctx, index), depth, a, func(b *acc, depth int) {
			// The direct callers' (and creators') arguments, each read in
			// its own method: a parameter of that method is the next level.
			pv := r.idx.ResolveParameter(ctx, index)
			for _, u := range pv.Unknown {
				if strings.Contains(u, "argument of the functional-interface call") {
					b.gap(GapSAMArg, u)
				} else {
					b.gap(GapParameter, u)
				}
			}
			if len(pv.Values) == 0 && len(pv.Unknown) == 0 {
				b.gap(GapParameter, fmt.Sprintf("no value for parameter %d of %s", index, ctx))
			}
			for i, c := range pv.Values {
				r.value(c, pv.In[i], depth, b)
			}
		})
	case valueflow.KindField:
		ref := *v.Field
		if t := descType(ref.Desc); !r.holders[t] {
			a.objects++
			return
		}
		r.memo("f:"+ref.Owner+"."+ref.Name+":"+ref.Desc, depth, a, func(b *acc, depth int) {
			fv := r.field(ref)
			for _, u := range fv.unknown {
				b.gap(GapField, u)
			}
			for _, c := range fv.values {
				r.value(c.v, c.ctx, depth, b)
			}
		})
	case valueflow.KindCallResult:
		r.callResult(*v.Call, depth, a)
	case valueflow.KindCallTrace:
		a.gap(GapInstance, fmt.Sprintf("the result of the instance call %s#%s is not followed", v.CallTrace.Owner, v.CallTrace.Name))
	case valueflow.KindUnknown:
		a.gap(GapValue, v.Reason)
	default:
		a.gap(GapValue, fmt.Sprintf("a %s value, not a functional object", v.Kind))
	}
}

// field resolves what a field may hold: every value its writers store
// (instance-insensitive — any object's field). A private field is
// written only in its class's nest (the class and its nested classes),
// a final one only in its class: both are complete; for any other field
// writers outside the nest are not searched, which is a disclosed gap.
func (r *Resolver) field(ref valueflow.FieldRef) fieldValues {
	if fv, ok := r.fields[ref]; ok {
		return fv
	}
	fv := r.fieldUncached(ref)
	r.fields[ref] = fv
	return fv
}

func (r *Resolver) fieldUncached(ref valueflow.FieldRef) fieldValues {
	var fv fieldValues
	decl, ok := r.h.ResolveField(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		fv.unknown = []string{fmt.Sprintf("field %s.%s is not declared in the analyzed input", ref.Owner, ref.Name)}
		return fv
	}
	ci, _ := r.h.Lookup(decl)
	f, ok := ci.FieldByNameAndDesc(ref.Name, ref.Desc)
	if !ok {
		fv.unknown = []string{fmt.Sprintf("field %s.%s is not declared in the analyzed input", decl, ref.Name)}
		return fv
	}
	final := f.AccessFlags&classfile.AccFinal != 0
	writers := []string{decl}
	if !final {
		writers = r.nest(decl)
	}
	if !final && f.AccessFlags&accPrivate == 0 {
		fv.unknown = append(fv.unknown, fmt.Sprintf("field %s.%s is neither private nor final: writers outside %s are not searched", decl, ref.Name, nestHost(decl)))
	}
	writes := 0
	for _, w := range writers {
		wci, ok := r.h.Lookup(w)
		if !ok {
			continue
		}
		for _, m := range wci.Methods {
			if len(m.Code) == 0 || (final && m.Name != "<init>" && m.Name != "<clinit>") {
				continue
			}
			id := dataflow.MethodID{Owner: w, Name: m.Name, Desc: m.Descriptor}
			facts, reason := r.idx.Facts(id)
			if reason != "" {
				fv.unknown = append(fv.unknown, fmt.Sprintf("writer %s: %s", id, reason))
				continue
			}
			for _, fw := range facts.FieldWrites() {
				if fw.Field.Name != ref.Name || fw.Field.Desc != ref.Desc {
					continue
				}
				if d, ok := r.h.ResolveField(fw.Field.Owner, fw.Field.Name, fw.Field.Desc); !ok || d != decl {
					continue
				}
				writes++
				fv.values = append(fv.values, ctxValue{v: fw.Value, ctx: id})
			}
		}
	}
	if writes == 0 {
		fv.unknown = append(fv.unknown, fmt.Sprintf("no write to field %s.%s found in %s", decl, ref.Name, nestHost(decl)))
	}
	return fv
}

const accPrivate = 0x0002

// nest returns the classes of decl's nest: its outermost class and every
// class nested in it (by binary name).
func (r *Resolver) nest(decl string) []string {
	host := nestHost(decl)
	if r.nests == nil {
		r.nests = map[string][]string{}
		for _, c := range r.h.SortedFQCNs() {
			h := nestHost(c)
			r.nests[h] = append(r.nests[h], c)
		}
	}
	return r.nests[host]
}

func nestHost(c string) string {
	dot := strings.LastIndexByte(c, '.')
	if i := strings.IndexByte(c[dot+1:], '$'); i > 0 {
		return c[:dot+1+i]
	}
	return c
}

// callResult resolves the return of an invokestatic/invokespecial call:
// the callee's returns, a returned parameter substituted by the call's
// argument in the caller.
func (r *Resolver) callResult(ref valueflow.CallResultRef, depth int, a *acc) {
	k := ref.Owner + "#" + ref.Name + ref.Desc
	cv, ok := r.calls[k]
	if !ok {
		cv = r.calleeReturns(ref)
		r.calls[k] = cv
	}
	if cv.unknown != "" {
		a.gap(GapCallResult, cv.unknown)
		return
	}
	caller, okCaller := siteMethod(ref.Site)
	callee := cv.callee
	// What the helper returns of its own (lambdas, fields, other helpers)
	// is the same for every call site; a returned parameter is the
	// argument of this one.
	var params []valueflow.Value
	r.memo("r:"+k, depth, a, func(b *acc, depth int) {
		for _, ret := range flatten(cv.values) {
			if ret.Kind != valueflow.KindTyped || ret.Parameter == nil {
				r.value(ret, callee, depth, b)
			}
		}
	})
	for _, ret := range flatten(cv.values) {
		if ret.Kind == valueflow.KindTyped && ret.Parameter != nil {
			params = append(params, ret)
		}
	}
	for _, ret := range params {
		r.returnedParam(ret, ref, caller, okCaller, depth, a)
	}
}

// descType is the class a field descriptor names ("" for a primitive or
// an array).
func descType(desc string) string {
	if len(desc) > 2 && desc[0] == 'L' {
		return strings.ReplaceAll(desc[1:len(desc)-1], "/", ".")
	}
	return ""
}

func flatten(vs []valueflow.Value) []valueflow.Value {
	var out []valueflow.Value
	for _, v := range vs {
		if v.Kind == valueflow.KindSet {
			out = append(out, flatten(v.Candidates)...)
		} else {
			out = append(out, v)
		}
	}
	return out
}

func (r *Resolver) returnedParam(ret valueflow.Value, ref valueflow.CallResultRef, caller dataflow.MethodID, okCaller bool, depth int, a *acc) {
	if !okCaller {
		a.gap(GapCallResult, "helper call site has no usable caller identity")
		return
	}
	facts, reason := r.idx.Facts(caller)
	if reason != "" {
		a.gap(GapCallResult, reason)
		return
	}
	arg, ok := argumentAt(facts, ref.Site, ret.Parameter.Index)
	if !ok {
		a.gap(GapCallResult, fmt.Sprintf("no argument %d recorded at %s pc=%d", ret.Parameter.Index, caller, ref.Site.PC))
		return
	}
	r.value(arg, caller, depth+1, a)
}

func (r *Resolver) calleeReturns(ref valueflow.CallResultRef) callValues {
	decl, ok := r.h.ResolveExact(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		return callValues{unknown: fmt.Sprintf("%s#%s%s is not in the analyzed input", ref.Owner, ref.Name, ref.Desc)}
	}
	facts, reason := r.idx.Facts(dataflow.MethodID{Owner: decl, Name: ref.Name, Desc: ref.Desc})
	if reason != "" {
		return callValues{unknown: reason}
	}
	if aborted, why := facts.Aborted(); aborted {
		return callValues{unknown: fmt.Sprintf("%s#%s%s: local-value analysis aborted: %s", decl, ref.Name, ref.Desc, why)}
	}
	vals := facts.ReturnValues()
	if len(vals) == 0 {
		return callValues{unknown: fmt.Sprintf("%s#%s%s has no value-returning return", decl, ref.Name, ref.Desc)}
	}
	return callValues{values: vals, callee: dataflow.MethodID{Owner: decl, Name: ref.Name, Desc: ref.Desc}}
}

func argumentAt(f *dataflow.MethodFacts, s bytecodeir.Site, i int) (v valueflow.Value, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return f.ArgumentAt(s, i), true
}

func siteMethod(s bytecodeir.Site) (dataflow.MethodID, bool) {
	i := strings.LastIndexByte(s.Method, '#')
	if i <= 0 || s.Descriptor == "" {
		return dataflow.MethodID{}, false
	}
	return dataflow.MethodID{Owner: s.Method[:i], Name: s.Method[i+1:], Desc: s.Descriptor}, true
}

// Edge is one call-graph edge to add: from the site's method at PC to the
// target (dispatch-expanded by internal/callgraph when Virtual).
type Edge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	CallKind                           classfile.CallKind
	TargetFQCN, TargetName, TargetDesc string
	Virtual                            bool
	Kind                               Kind
	Creators                           []dataflow.MethodID
}

// Edges turns results into edges. A sam-call runs only values whose
// interface method is the one called — guaranteed by the bytecode's types
// for a lambda reaching the site, so every target counts.
func Edges(results []Result) []Edge {
	var out []Edge
	for _, r := range results {
		for _, t := range r.Targets {
			out = append(out, Edge{CallerFQCN: r.Site.CallerFQCN, CallerName: r.Site.CallerName, CallerDesc: r.Site.CallerDesc,
				PC: r.Site.PC, CallKind: r.Site.CallKind, TargetFQCN: t.Owner, TargetName: t.Name, TargetDesc: t.Desc,
				Virtual: t.Virtual(), Kind: r.Site.Kind, Creators: t.Creators})
		}
	}
	return out
}

// Stats summarize results for the log and the report.
type Stats struct {
	// Sites by kind; Resolved: at least one target and no gap; Partial:
	// targets and a gap; Unresolved: a gap and no target; NoLambda: only
	// ordinary objects / null.
	Sites, Resolved, Partial, Unresolved, NoLambda map[Kind]int
	// Gaps: sites with each gap category (a site may have several).
	Gaps map[string]int
}

// Summarize counts results.
func Summarize(results []Result) Stats {
	s := Stats{Sites: map[Kind]int{}, Resolved: map[Kind]int{}, Partial: map[Kind]int{}, Unresolved: map[Kind]int{},
		NoLambda: map[Kind]int{}, Gaps: map[string]int{}}
	for _, r := range results {
		k := r.Site.Kind
		s.Sites[k]++
		switch {
		case len(r.Targets) > 0 && len(r.Gaps) == 0:
			s.Resolved[k]++
		case len(r.Targets) > 0:
			s.Partial[k]++
		case len(r.Gaps) > 0:
			s.Unresolved[k]++
		default:
			s.NoLambda[k]++
		}
		for c := range r.Gaps {
			if k == KindCall {
				s.Gaps[c]++
			}
		}
	}
	return s
}
