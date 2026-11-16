// Package springaop links calls of Spring beans' methods to the
// @AspectJ-style advice Spring AOP runs around them (FRAMEWORKS F3,
// slice: model spring:aop).
//
// A bean advised by an aspect is replaced in the container by a proxy (JDK
// or CGLIB). A call through the bean reference goes to the proxy, which
// runs the matching advice (ReflectiveMethodInvocation#proceed →
// AspectJ*Advice → AbstractAspectJAdvice#invokeAdviceMethodWithGivenArgs →
// Method#invoke) before or around the target method. The advice method is
// called through reflection with nothing a reflection model resolves, so
// without this model no path reached it.
//
// The model adds an edge (basis model) from each call site that may
// invoke an advised method on a bean to the advice method:
//
//   - an aspect is an application class annotated @Aspect that is
//     registered as a bean (related to a bean type the caller passes in:
//     stereotypes, @Bean products, XML beans); an @Aspect class that is
//     no bean advises nothing;
//   - advice is a method of an aspect annotated @Before, @After,
//     @AfterReturning, @AfterThrowing or @Around; its pointcut is the
//     annotation's pointcut (or value) element;
//   - the pointcut language evaluated: execution(...) with modifier,
//     return, declaring-type, name and parameter patterns (*, .., +),
//     within(...), @annotation(Type), @within(Type) (and their binding
//     forms when the advice has exactly one annotation-typed parameter),
//     references to @Pointcut methods, &&, ||, !. Anything else (args,
//     this, target, bean, annotation patterns inside execution, …) is
//     unknown: an advice whose match depends on it gets no edge for that
//     join point and is counted as undecided — never a guessed match;
//   - a join point is a method a proxy can intercept on a bean class: an
//     application or bean class related to a bean type, its own and
//     inherited methods (not java.lang.Object's), neither static, private,
//     final, a constructor nor synthetic; the aspect classes themselves
//     are not advised;
//   - a call site matches when it is a virtual or interface call of the
//     join point's name and descriptor whose owner the bean class is
//     assignable to (the call may dispatch to that bean). A call on `this`
//     from the bean's own code (self-invocation) bypasses the proxy and
//     gets no edge;
//   - gate: a reachable method creating one of Spring's AspectJ advice
//     objects (AbstractAspectJAdvice subtypes, created by
//     ReflectiveAspectJAdvisorFactory#getAdvice), so an input that never
//     builds advisors from @Aspect classes gets nothing.
//
// @AfterReturning edges carry the condition "returned" and
// @AfterThrowing edges "exceptional" (path evidence and the report's
// arrow). For @Around the target still runs only if the advice calls
// proceed(); the existing edge to the target is kept (an
// over-approximation). Advice order (@Order), the returning/throwing
// type restrictions, XML <aop:config>, the AspectJ compiler or load-time
// weaving and JDK-vs-CGLIB differences for final methods reached through
// an interface are not modeled.
package springaop

import (
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ModelID identifies the model; its edges carry callgraph.BasisModel.
const ModelID = "spring:aop"

// Version versions the rules. Snapshots record it.
const Version = "a1"

const (
	aspectAnn      = "org.aspectj.lang.annotation.Aspect"
	pointcutAnn    = "org.aspectj.lang.annotation.Pointcut"
	abstractAdvice = "org.springframework.aop.aspectj.AbstractAspectJAdvice"
	objectFQCN     = "java.lang.Object"
	accPrivate     = 0x0002
	accBridge      = 0x0040
	accSynthetic   = 0x1000
	accAnnotation  = 0x2000
	// maxEdges bounds the candidate edges of one input; past it the
	// model stops adding and reports it (Result.Truncated).
	maxEdges = 50000
)

// adviceKinds maps the advice annotations to the edge condition.
var adviceKinds = map[string]string{
	"org.aspectj.lang.annotation.Before":         "",
	"org.aspectj.lang.annotation.After":          "",
	"org.aspectj.lang.annotation.Around":         "",
	"org.aspectj.lang.annotation.AfterReturning": "returned",
	"org.aspectj.lang.annotation.AfterThrowing":  "exceptional",
}

var joinPointTypes = map[string]bool{
	"org.aspectj.lang.JoinPoint":            true,
	"org.aspectj.lang.ProceedingJoinPoint":  true,
	"org.aspectj.lang.JoinPoint$StaticPart": true,
}

// Advice is one advice method.
type Advice struct {
	Aspect, Name, Desc string
	// Kind is the advice annotation's simple name (Before, Around, …).
	Kind       string
	Expression string
	// Registered: the aspect class is related to a bean type.
	Registered bool
	// Matched join points (bean class#method) and the undecided ones.
	Matched, Undecided int
	// Unknown lists the parts of the expression the model does not
	// evaluate.
	Unknown []string
}

// JoinPoint is one advised method of a bean class.
type JoinPoint struct {
	Bean, Declaring, Name, Desc string
}

// Result is the model applied to one input.
type Result struct {
	Aspects []string
	Advice  []Advice
	// JoinPoints are the bean methods matched by at least one advice.
	JoinPoints []JoinPoint
	// SelfCalls counts call sites left out as self-invocation.
	SelfCalls int
	// Edges are the candidates for callgraph.NewCreatorGate.
	Edges []callgraph.SAMEdge
	// NoCreator: advice exists but nothing creates Spring's advice
	// objects.
	NoCreator bool
	// Truncated: the edge budget was hit.
	Truncated bool
}

type adviceMatch struct {
	adv  int
	jp   joinPoint
	cond string
}

// Find builds the aspect catalog and the candidate edges. beans are the
// bean types the spring profile found.
func Find(h *hierarchy.Hierarchy, beans []string) Result {
	var r Result
	type parsed struct {
		n   node
		res *resolver
	}
	var nodes []parsed
	aspects := map[string]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp || ci.IsInterface() || !hasAnn(ci.Annotations, aspectAnn) {
			continue
		}
		aspects[fqcn] = true
		r.Aspects = append(r.Aspects, fqcn)
		reg := registered(h, fqcn, beans)
		res := newResolver(h, fqcn)
		for _, m := range ci.Methods {
			for _, a := range m.AnnotationDetails {
				if _, ok := adviceKinds[a.Type]; !ok {
					continue
				}
				expr, _ := stringElement(a, "pointcut")
				if strings.TrimSpace(expr) == "" {
					expr, _ = stringElement(a, "value")
				}
				n := bind(h, parse(expr), m.Descriptor)
				adv := Advice{Aspect: fqcn, Name: m.Name, Desc: m.Descriptor, Kind: a.Type[strings.LastIndexByte(a.Type, '.')+1:], Expression: expr, Registered: reg}
				why := map[string]bool{}
				unknownReasons(n, res, why, 0)
				adv.Unknown = sortedKeys(why)
				r.Advice = append(r.Advice, adv)
				nodes = append(nodes, parsed{n, res})
				break
			}
		}
	}
	if len(r.Advice) == 0 {
		return r
	}
	var gate []callgraph.MethodRef
	advices := map[string]bool{abstractAdvice: true}
	for _, s := range h.Subtypes(abstractAdvice) {
		advices[s] = true
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, ref := range m.InitRefs {
				if ref.Kind == classfile.InitNew && advices[ref.Owner] {
					gate = append(gate, callgraph.MethodRef{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor})
					break
				}
			}
		}
	}
	if len(gate) == 0 {
		r.NoCreator = true
	}

	// Match every registered advice against every join point.
	var matches []adviceMatch
	matchedJP := map[JoinPoint]bool{}
	for _, c := range beanClasses(h, beans, aspects) {
		for _, jp := range joinPoints(h, c) {
			jp := jp
			for i := range r.Advice {
				if !r.Advice[i].Registered {
					continue
				}
				switch nodes[i].n.eval(&jp, nodes[i].res) {
				case yes:
					r.Advice[i].Matched++
					matches = append(matches, adviceMatch{adv: i, jp: jp, cond: adviceKinds["org.aspectj.lang.annotation."+r.Advice[i].Kind]})
					matchedJP[JoinPoint{Bean: jp.C, Declaring: jp.D, Name: jp.m.Name, Desc: jp.m.Descriptor}] = true
				case unknown:
					r.Advice[i].Undecided++
				}
			}
		}
	}
	for jp := range matchedJP {
		r.JoinPoints = append(r.JoinPoints, jp)
	}
	sort.Slice(r.JoinPoints, func(i, j int) bool {
		a, b := r.JoinPoints[i], r.JoinPoints[j]
		if a.Bean != b.Bean {
			return a.Bean < b.Bean
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Desc < b.Desc
	})
	if len(matches) == 0 || len(gate) == 0 {
		return r
	}

	// Call sites by name and descriptor.
	type site struct {
		ci *hierarchy.ClassInfo
		m  hierarchy.Method
		c  classfile.Call
	}
	bySig := map[string][]site{}
	want := map[string]bool{}
	for _, mt := range matches {
		want[mt.jp.m.Name+mt.jp.m.Descriptor] = true
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if (c.Kind == classfile.CallVirtual || c.Kind == classfile.CallInterface) && want[c.Name+c.Desc] {
					bySig[c.Name+c.Desc] = append(bySig[c.Name+c.Desc], site{ci, m, c})
				}
			}
		}
	}
	idx := dataflow.NewCallerIndex(h)
	defer idx.Close()
	selfCache := map[string]bool{}
	isSelf := func(s site) bool {
		k := s.ci.FQCN + "#" + s.m.Name + s.m.Descriptor + "@" + strconv.Itoa(s.c.PC)
		if v, ok := selfCache[k]; ok {
			return v
		}
		v := selfInvocation(idx, s.ci, s.m, s.c)
		selfCache[k] = v
		if v {
			r.SelfCalls++
		}
		return v
	}
	type edgeKey struct {
		caller string
		pc     int
		target string
	}
	seen := map[edgeKey]bool{}
	for _, mt := range matches {
		adv := r.Advice[mt.adv]
		for _, s := range bySig[mt.jp.m.Name+mt.jp.m.Descriptor] {
			if !assignable(h, mt.jp.C, s.c.Owner) {
				continue
			}
			if !s.m.IsStatic() && (assignable(h, s.ci.FQCN, mt.jp.C) || assignable(h, mt.jp.C, s.ci.FQCN)) && isSelf(s) {
				continue
			}
			k := edgeKey{s.ci.FQCN + "#" + s.m.Name + s.m.Descriptor, s.c.PC, adv.Aspect + "#" + adv.Name + adv.Desc}
			if seen[k] {
				continue
			}
			seen[k] = true
			if len(r.Edges) >= maxEdges {
				r.Truncated = true
				return r
			}
			r.Edges = append(r.Edges, callgraph.SAMEdge{
				CallerFQCN: s.ci.FQCN, CallerName: s.m.Name, CallerDesc: s.m.Descriptor, PC: s.c.PC, Kind: s.c.Kind,
				TargetFQCN: adv.Aspect, TargetName: adv.Name, TargetDesc: adv.Desc,
				Creators: gate, Condition: mt.cond,
			})
		}
	}
	return r
}

// selfInvocation: the call's receiver is the calling method's own `this`
// (DATA_FLOW D2: local 0 of an instance method is the class's typed
// value, with no parameter, field or call provenance).
func selfInvocation(idx *dataflow.CallerIndex, ci *hierarchy.ClassInfo, m hierarchy.Method, c classfile.Call) bool {
	facts, _ := idx.Facts(dataflow.MethodID{Owner: ci.FQCN, Name: m.Name, Desc: m.Descriptor})
	if facts == nil {
		return false
	}
	if aborted, _ := facts.Aborted(); aborted {
		return false
	}
	at := bytecodeir.Site{Method: ci.FQCN + "#" + m.Name, Descriptor: m.Descriptor, PC: c.PC, Op: bytecodeir.OpCall}
	var v valueflow.Value
	ok := func() (ok bool) {
		defer func() {
			if recover() != nil {
				ok = false
			}
		}()
		v = facts.ReceiverAt(at)
		return true
	}()
	return ok && v.Kind == valueflow.KindTyped && v.Parameter == nil && v.Type == ci.FQCN
}

// bind resolves @annotation(name)/@within(name) to the advice method's one
// annotation-typed parameter.
func bind(h *hierarchy.Hierarchy, n node, desc string) node {
	switch x := n.(type) {
	case andNode:
		return andNode{bind(h, x.l, desc), bind(h, x.r, desc)}
	case orNode:
		return orNode{bind(h, x.l, desc), bind(h, x.r, desc)}
	case notNode:
		return notNode{bind(h, x.x, desc)}
	case bindingNode:
		params, _ := descParts(desc)
		var anns []string
		for _, p := range params {
			if joinPointTypes[p] {
				continue
			}
			if ci, ok := h.Lookup(p); ok && ci.AccessFlags&accAnnotation != 0 {
				anns = append(anns, p)
			}
		}
		if len(anns) == 1 {
			return annNode{ann: anns[0], onType: x.designator == "@within"}
		}
	}
	return n
}

// beanClasses are the classes whose methods may be advised: application
// classes and bean types related to a bean type, never an aspect,
// interfaces left out (a proxy runs the class's body).
func beanClasses(h *hierarchy.Hierarchy, beans []string, aspects map[string]bool) []string {
	isBean := map[string]bool{}
	for _, b := range beans {
		isBean[b] = true
	}
	var out []string
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if ci.IsInterface() || aspects[fqcn] || ci.AccessFlags&accAnnotation != 0 {
			continue
		}
		if !ci.IsApp && !isBean[fqcn] {
			continue
		}
		if registered(h, fqcn, beans) {
			out = append(out, fqcn)
		}
	}
	return out
}

// joinPoints lists the methods a proxy of class c can intercept.
func joinPoints(h *hierarchy.Hierarchy, c string) []joinPoint {
	var out []joinPoint
	seen := map[string]bool{}
	for cur := c; cur != "" && cur != objectFQCN; {
		ci, ok := h.Lookup(cur)
		if !ok {
			break
		}
		for _, m := range ci.Methods {
			k := m.Name + m.Descriptor
			if seen[k] {
				continue
			}
			seen[k] = true
			if m.Name == "<init>" || m.Name == "<clinit>" || m.IsStatic() || m.IsAbstract() ||
				m.AccessFlags&(accPrivate|classfile.AccFinal|accBridge|accSynthetic) != 0 {
				continue
			}
			out = append(out, joinPoint{h: h, C: c, D: cur, m: m})
		}
		cur = ci.SuperClass
	}
	return out
}

func hasAnn(anns []string, a string) bool {
	for _, x := range anns {
		if x == a {
			return true
		}
	}
	return false
}

// assignable reports whether t is p or a subtype of it.
func assignable(h *hierarchy.Hierarchy, t, p string) bool {
	if p == objectFQCN || t == p {
		return true
	}
	for _, s := range supertypes(h, t) {
		if s == p {
			return true
		}
	}
	return false
}

func registered(h *hierarchy.Hierarchy, fqcn string, beans []string) bool {
	for _, b := range beans {
		if assignable(h, b, fqcn) || assignable(h, fqcn, b) && b != objectFQCN {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RegisteredAspects counts the aspects that are beans.
func (r Result) RegisteredAspects() int {
	seen := map[string]bool{}
	for _, a := range r.Advice {
		if a.Registered {
			seen[a.Aspect] = true
		}
	}
	return len(seen)
}

// UndecidedCount sums the advice × join point pairs left undecided.
func (r Result) UndecidedCount() int {
	n := 0
	for _, a := range r.Advice {
		n += a.Undecided
	}
	return n
}
