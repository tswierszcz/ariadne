// Package taint is DATA_FLOW D4's first taint analysis: whether data from
// a cataloged source (request data of a Spring MVC / JAX-RS handler, a
// ServletRequest accessor) reaches a sink's argument. The pilot's sinks
// are the data read by deserialization entries (DESERIALIZATION Z1); the
// caller states, per sink, which value of the call is the data.
//
// The search runs backwards from the sink over DATA_FLOW D3's dependency
// summaries (internal/flowsum): the sink's data in terms of its method's
// inputs, then, per call site of that method found in application code,
// the argument passed there in the caller's terms, until a source, a
// constant, an input outside the catalog or a gap. Substituting per call
// site keeps two calls of one helper apart.
//
// A result is report-only: it never changes a reachability verdict and is
// never proof that a CVE is exploitable. A gap (an unmodeled call's
// result, a field, a limit, an incomplete summary) makes the result
// undetermined unless a flow was found anyway — never "no flow".
package taint

import (
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/flowsum"
	"ariadne/internal/hierarchy"
)

// States of a result.
const (
	FlowFound    = "flow_found"
	NoFlow       = "no_flow_in_model"
	Undetermined = "undetermined"
	NotAssessed  = "not_assessed"
)

// Options bound the search.
type Options struct {
	// MaxDepth bounds the call sites followed backwards from a sink.
	MaxDepth int
	// MaxNodes bounds the (method, input) pairs visited per sink.
	MaxNodes int
	// MaxFlows bounds the flows kept per result (the count is exact).
	MaxFlows int
}

// DefaultOptions are the CLI's bounds.
var DefaultOptions = Options{MaxDepth: 8, MaxNodes: 2000, MaxFlows: 5}

// Sink is one value to assess: Data of the call at PC in Method.
type Sink struct {
	Method flowsum.MethodKey
	PC     int
	// Data selects the call's value: "receiver" or "arg:<i>".
	Data string
	// Rule names the sink model (e.g. "sink:deser-java-stream").
	Rule string
	// Policy is the entry's own guard, already evaluated by its model.
	Policy PolicyGuard
}

// PolicyGuard is the sink's own protection, as its model evaluated it.
type PolicyGuard struct {
	State  string // GuardEffective, GuardPartial, GuardUnknown or GuardNone
	Detail string
}

// Step is one hop of a flow, from the source towards the sink: the input
// of Method the data is in, and the PC of the call that passes it on
// (-1 at the sink's method, where PC is the sink's own call).
type Step struct {
	Method string `json:"method"`
	Input  string `json:"input"`
	CallPC int    `json:"call_pc"`
}

// Flow is one path from a source to the sink.
type Flow struct {
	Rule   string `json:"rule"`
	Source string `json:"source"`
	Steps  []Step `json:"steps"`
}

// Guard is one protection seen on a flow (or the sink's policy).
type Guard struct {
	Kind      string `json:"kind"`
	Effective bool   `json:"effective"`
	Detail    string `json:"detail"`
	Method    string `json:"method,omitempty"`
	// PCs are the calls in Method (one guard per method and call kind).
	PCs []int `json:"pcs,omitempty"`
}

// Result is one sink's assessment.
type Result struct {
	State      string  `json:"state"`
	Catalog    string  `json:"catalog"`
	Rule       string  `json:"rule"`
	Flows      []Flow  `json:"flows,omitempty"`
	FlowCount  int     `json:"flow_count,omitempty"`
	GuardState string  `json:"guard_state,omitempty"`
	Guards     []Guard `json:"guards,omitempty"`
	// Gaps: why the search could not follow some input (always listed,
	// also when a flow was found).
	Gaps []string `json:"gaps,omitempty"`
	// GapCount is the number of distinct gaps (Gaps keeps the first 20).
	GapCount int `json:"gap_count,omitempty"`
	// Outside: inputs the search ended at that are not cataloged
	// sources (a parameter of a method no application code calls).
	Outside []string `json:"outside_inputs,omitempty"`
	// Reason says why a sink was not assessed.
	Reason string `json:"reason,omitempty"`
}

type callRef struct {
	caller flowsum.MethodKey
	call   classfile.Call
}

// Analyzer holds the reverse call index of the application classes.
type Analyzer struct {
	h    *hierarchy.Hierarchy
	opts Options
	// byNameDesc: application call sites by callee name+desc.
	byNameDesc map[string][]callRef
	// dynamic: methods referenced by an invokedynamic (a lambda or a
	// method reference): called with arguments the search cannot see.
	dynamic  map[flowsum.MethodKey]bool
	callers  map[flowsum.MethodKey][]callRef
	subtypes map[string]map[string]bool
}

// New indexes the calls of h's application classes.
func New(h *hierarchy.Hierarchy, opts Options) *Analyzer {
	a := &Analyzer{h: h, opts: opts, byNameDesc: map[string][]callRef{}, dynamic: map[flowsum.MethodKey]bool{},
		callers: map[flowsum.MethodKey][]callRef{}, subtypes: map[string]map[string]bool{}}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		for _, m := range ci.Methods {
			k := flowsum.MethodKey{Owner: fqcn, Name: m.Name, Desc: m.Descriptor}
			for _, c := range m.Calls {
				if c.Kind == classfile.CallDynamic {
					a.dynamic[flowsum.MethodKey{Owner: c.Owner, Name: c.Name, Desc: c.Desc}] = true
					continue
				}
				nd := c.Name + c.Desc
				a.byNameDesc[nd] = append(a.byNameDesc[nd], callRef{caller: k, call: c})
			}
		}
	}
	return a
}

func (a *Analyzer) isSubtype(sub, super string) bool {
	if sub == super {
		return true
	}
	set, ok := a.subtypes[super]
	if !ok {
		set = map[string]bool{}
		for _, s := range a.h.Subtypes(super) {
			set[s] = true
		}
		a.subtypes[super] = set
	}
	return set[sub]
}

// Callers returns the application call sites that may run m, sorted.
func (a *Analyzer) Callers(m flowsum.MethodKey) []callRef {
	if cs, ok := a.callers[m]; ok {
		return cs
	}
	var out []callRef
	for _, r := range a.byNameDesc[m.Name+m.Desc] {
		c := r.call
		decl, ok := a.h.ResolveExact(c.Owner, c.Name, c.Desc)
		switch {
		case ok && decl == m.Owner:
		case c.Kind == classfile.CallVirtual || c.Kind == classfile.CallInterface:
			// Dispatch may reach m when m's class is a subtype of the
			// named owner and m is what that class runs.
			if d, ok := a.h.ResolveExact(m.Owner, m.Name, m.Desc); !ok || d != m.Owner || !a.isSubtype(m.Owner, c.Owner) {
				continue
			}
		default:
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].caller != out[j].caller {
			return out[i].caller.String() < out[j].caller.String()
		}
		return out[i].call.PC < out[j].call.PC
	})
	a.callers[m] = out
	return out
}

// Closure returns the methods the search may need summaries of: the sink
// methods and their application callers, transitively, within MaxDepth.
func (a *Analyzer) Closure(sinks []flowsum.MethodKey) []flowsum.MethodKey {
	seen := map[flowsum.MethodKey]bool{}
	var out []flowsum.MethodKey
	level := sinks
	for d := 0; d <= a.opts.MaxDepth && len(level) > 0; d++ {
		var next []flowsum.MethodKey
		for _, m := range level {
			if seen[m] {
				continue
			}
			seen[m] = true
			out = append(out, m)
			for _, r := range a.Callers(m) {
				if !seen[r.caller] {
					next = append(next, r.caller)
				}
			}
		}
		level = next
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}

// sourceParam returns the source rule and label of parameter i of m, or
// "" when it is not a cataloged source.
func (a *Analyzer) sourceParam(m flowsum.MethodKey, i int) (rule, label string) {
	ci, ok := a.h.Lookup(m.Owner)
	if !ok || !ci.IsApp {
		return "", ""
	}
	meth, ok := ci.MethodByNameAndDesc(m.Name, m.Desc)
	if !ok {
		return "", ""
	}
	var anns []string
	if i < len(meth.ParameterAnnotations) {
		anns = meth.ParameterAnnotations[i]
	}
	for _, an := range anns {
		if l, ok := springParamAnnotations[an]; ok {
			return RuleSpringParam, l
		}
		if l, ok := jaxrsParamAnnotations[an]; ok {
			return RuleJaxRSParam, l
		}
	}
	handler := false
	for _, an := range meth.Annotations {
		if springMappings[an] {
			handler = true
		}
	}
	if !handler || len(anns) > 0 {
		return "", ""
	}
	args, err := classfile.ArgumentSlots(m.Desc, true)
	if err != nil || i >= len(args) || !springImplicitTypes[args[i].Type] {
		return "", ""
	}
	return RuleSpringImplicit, "implicitly bound handler parameter"
}

// typeName renders a field descriptor as a Java type name.
func typeName(d string) string {
	dims := 0
	for strings.HasPrefix(d, "[") {
		dims++
		d = d[1:]
	}
	prim := map[string]string{"B": "byte", "C": "char", "D": "double", "F": "float", "I": "int", "J": "long", "S": "short", "Z": "boolean"}
	n, ok := prim[d]
	if !ok {
		n = strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(d, "L"), ";"), "/", ".")
	}
	return n + strings.Repeat("[]", dims)
}

// describe renders one flowsum source of m for the report.
func describe(m flowsum.MethodKey, src string) string {
	switch {
	case strings.HasPrefix(src, "param:"):
		i, _ := strconv.Atoi(src[len("param:"):])
		label := "parameter " + strconv.Itoa(i)
		if args, err := classfile.ArgumentSlots(m.Desc, true); err == nil && i < len(args) {
			label += " (" + typeName(args[i].Type) + ")"
		}
		return label
	case src == "this":
		return "this"
	case strings.HasPrefix(src, "field:"):
		return "field " + src[len("field:"):]
	case src == flowsum.CallMany:
		return "results of many unmodeled calls"
	case strings.HasPrefix(src, "call:"):
		c := src[len("call:"):]
		if i := strings.IndexByte(c, '('); i > 0 {
			c = c[:i]
		}
		return "result of " + c
	}
	return src
}

// search is one sink's backward search.
type search struct {
	a       *Analyzer
	e       *flowsum.Engine
	visited map[string]bool
	flows   []Flow
	count   int
	gaps    map[string]bool
	outside map[string]bool
	guards  map[string]Guard
	nodes   int
}

func (s *search) gap(g string) { s.gaps[g] = true }

// path is the chain from the sink outwards: path[0] is the sink's method.
type hop struct {
	m     flowsum.MethodKey
	src   string // the flowsum source in m's terms
	pcOut int    // the call in m that passes the data towards the sink
}

func (s *search) found(rule, source string, path []hop) {
	s.count++
	s.checkGuards(path)
	if len(s.flows) >= s.a.opts.MaxFlows {
		return
	}
	f := Flow{Rule: rule, Source: source}
	for i := len(path) - 1; i >= 0; i-- {
		f.Steps = append(f.Steps, Step{Method: path[i].m.String(), Input: describe(path[i].m, path[i].src), CallPC: path[i].pcOut})
	}
	s.flows = append(s.flows, f)
}

// checkGuards records the guard calls on a found flow: in each method of
// the path, a cataloged call whose receiver or arguments depend on the
// input the data is in there.
func (s *search) checkGuards(path []hop) {
	for _, hp := range path {
		sum := s.e.Summary(hp.m)
		ci, ok := s.a.h.Lookup(hp.m.Owner)
		if sum == nil || !ok {
			continue
		}
		meth, ok := ci.MethodByNameAndDesc(hp.m.Name, hp.m.Desc)
		if !ok {
			continue
		}
		for _, c := range meth.Calls {
			spec, ok := guardCalls[c.Owner+"#"+c.Name]
			if !ok {
				continue
			}
			cd, ok := sum.Calls[c.PC]
			if !ok || !dependsOn(cd, hp.src) {
				continue
			}
			detail := c.Owner + "#" + c.Name + ": " + spec.what
			if spec.kind == GuardControlCheck {
				detail += " — decides a branch only, the data itself is not cleaned"
			} else {
				detail += " — does not restrict the types the entry creates"
			}
			key := hp.m.String() + "|" + detail
			g, seen := s.guards[key]
			if !seen {
				g = Guard{Kind: spec.kind, Detail: detail, Method: hp.m.String()}
			}
			if !containsInt(g.PCs, c.PC) {
				g.PCs = append(g.PCs, c.PC)
				sort.Ints(g.PCs)
			}
			s.guards[key] = g
		}
	}
}

func containsInt(xs []int, x int) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

func dependsOn(cd flowsum.CallDeps, src string) bool {
	for _, x := range cd.Receiver.Strings() {
		if x == src {
			return true
		}
	}
	for _, d := range cd.Args {
		for _, x := range d.Strings() {
			if x == src {
				return true
			}
		}
	}
	return false
}

// trace follows src, a flowsum source of path[0].m... (the last hop).
func (s *search) trace(path []hop, depth int) {
	cur := path[len(path)-1]
	key := cur.m.String() + "|" + cur.src
	if s.visited[key] {
		return
	}
	s.visited[key] = true
	s.nodes++
	if s.nodes > s.a.opts.MaxNodes {
		s.gap("search limit: more than " + strconv.Itoa(s.a.opts.MaxNodes) + " inputs visited")
		return
	}
	src := cur.src
	switch {
	case strings.HasPrefix(src, "field:"):
		s.gap("value of " + describe(cur.m, src) + " (in the summary of " + cur.m.String() + "): flows through fields are not traced (" + CatalogVersion + ")")
		return
	case src == flowsum.CallMany:
		s.gap("results of many unmodeled calls (in the summary of " + cur.m.String() + ")")
		return
	case strings.HasPrefix(src, "call:"):
		call := src[len("call:"):]
		if servletSource(call) {
			s.found(RuleServletCall, strings.TrimPrefix(describe(cur.m, src), "result of "), path)
			return
		}
		s.gap(describe(cur.m, src) + " (in the summary of " + cur.m.String() + "): an unmodeled call")
		return
	case strings.HasPrefix(src, "new@"):
		return // an allocation: what it holds is in the other sources
	}
	idx := -1
	if strings.HasPrefix(src, "param:") {
		idx, _ = strconv.Atoi(src[len("param:"):])
		if rule, label := s.a.sourceParam(cur.m, idx); rule != "" {
			s.found(rule, label, path)
			return
		}
	} else if src != "this" {
		s.gap("unknown source " + src)
		return
	}
	callers := s.a.Callers(cur.m)
	if s.a.dynamic[cur.m] {
		s.gap(cur.m.String() + " is called through a lambda or method reference: its arguments are not traced")
	}
	if len(callers) == 0 {
		if !s.a.dynamic[cur.m] {
			s.outside[describe(cur.m, src)+" of "+cur.m.String()+" (no caller in the application; not a cataloged source)"] = true
		}
		return
	}
	if depth >= s.a.opts.MaxDepth {
		s.gap("depth limit (" + strconv.Itoa(s.a.opts.MaxDepth) + " calls) at " + cur.m.String())
		return
	}
	for _, r := range callers {
		sum := s.e.Summary(r.caller)
		if sum == nil {
			s.gap(r.caller.String() + ": not summarized")
			continue
		}
		for _, inc := range sum.Incomplete {
			s.gap(r.caller.String() + ": summary incomplete: " + inc)
		}
		cd, ok := sum.Calls[r.call.PC]
		if !ok {
			s.gap(r.caller.String() + "@" + strconv.Itoa(r.call.PC) + ": call not in the summary")
			continue
		}
		var d flowsum.Dep
		if idx >= 0 {
			if idx >= len(cd.Args) {
				s.gap(r.caller.String() + "@" + strconv.Itoa(r.call.PC) + ": argument count mismatch")
				continue
			}
			d = cd.Args[idx]
		} else {
			d = cd.Receiver
		}
		for _, x := range d.Strings() {
			next := append(append([]hop(nil), path...), hop{m: r.caller, src: x, pcOut: r.call.PC})
			s.trace(next, depth+1)
		}
	}
}

// Analyze assesses one sink with e's summaries (Summarize must have run
// on Closure of the sinks' methods).
func (a *Analyzer) Analyze(e *flowsum.Engine, sk Sink) Result {
	res := Result{Catalog: CatalogVersion, Rule: sk.Rule}
	sum := e.Summary(sk.Method)
	if sum == nil {
		res.State, res.Reason = NotAssessed, "the sink's method was not summarized"
		return res
	}
	cd, ok := sum.Calls[sk.PC]
	if !ok {
		res.State, res.Reason = Undetermined, "the sink's call is not in its method's summary"
		res.Gaps = append(res.Gaps, sum.Incomplete...)
		return res
	}
	var d flowsum.Dep
	switch {
	case sk.Data == "receiver":
		d = cd.Receiver
	case strings.HasPrefix(sk.Data, "arg:"):
		i, _ := strconv.Atoi(sk.Data[len("arg:"):])
		if i < len(cd.Args) {
			d = cd.Args[i]
		}
	}
	s := &search{a: a, e: e, visited: map[string]bool{}, gaps: map[string]bool{}, outside: map[string]bool{}, guards: map[string]Guard{}}
	for _, inc := range sum.Incomplete {
		s.gap(sk.Method.String() + ": summary incomplete: " + inc)
	}
	for _, x := range d.Strings() {
		s.trace([]hop{{m: sk.Method, src: x, pcOut: -1}}, 0)
	}
	res.Flows, res.FlowCount = s.flows, s.count
	res.Gaps, res.GapCount = sorted(s.gaps, 20), len(s.gaps)
	res.Outside = sorted(s.outside, 20)
	switch {
	case s.count > 0:
		res.State = FlowFound
	case len(s.gaps) > 0:
		res.State = Undetermined
	default:
		res.State = NoFlow
	}
	// Guards: the sink's own policy first, then what the flows pass.
	if sk.Policy.State != "" && sk.Policy.State != GuardNone {
		res.Guards = append(res.Guards, Guard{Kind: GuardPolicy, Effective: sk.Policy.State == GuardEffective, Detail: sk.Policy.Detail})
	}
	keys := make([]string, 0, len(s.guards))
	for k := range s.guards {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		res.Guards = append(res.Guards, s.guards[k])
	}
	if res.State == FlowFound {
		switch {
		case sk.Policy.State == GuardEffective || sk.Policy.State == GuardPartial || sk.Policy.State == GuardUnknown:
			res.GuardState = sk.Policy.State
		case len(s.guards) > 0:
			res.GuardState = GuardOnlyOther
		default:
			res.GuardState = GuardNone
		}
	}
	return res
}

func sorted(m map[string]bool, max int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	if len(out) > max {
		out = append(out[:max], "… "+strconv.Itoa(len(m)-max)+" more")
	}
	return out
}
