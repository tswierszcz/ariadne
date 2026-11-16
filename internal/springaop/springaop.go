// Package springaop links calls of Spring beans' methods to the advice
// Spring AOP runs around them (FRAMEWORKS F3, model spring:aop).
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
//     stereotypes, @Bean products, XML beans), or the bean an
//     <aop:aspect ref> of an <aop:config> names; an @Aspect class that is
//     no bean advises nothing;
//   - advice is a method of an aspect annotated @Before, @After,
//     @AfterReturning, @AfterThrowing or @Around (its pointcut is the
//     annotation's pointcut or value element), or the method an
//     <aop:before|after|after-returning|after-throwing|around> names
//     (pointcut or pointcut-ref; and/or/not allowed);
//   - the advice's variables are bound as Spring's parameter name
//     discovery does: argNames/arg-names, else the names in the class file
//     (-parameters, -g), else inference from the expression when only one
//     parameter can fit (methodScope);
//   - the pointcut language evaluated, three-valued: execution(...)
//     (method annotations, modifiers, return, declaring type, name and
//     parameter patterns with *, .., +), within, @annotation, @within,
//     @target, this, target, args (types or bound variables; a value whose
//     declared type may be an instance of the type matches), bean (the
//     bean's names), references to @Pointcut methods (parameterized ones
//     bound by their own parameters), &&, ||, !. Anything else (@args,
//     cflow, annotation patterns on parameters, an unbound variable, …) is
//     unknown: an advice whose match depends on it gets no edge for that
//     join point and is counted as undecided — never a guessed match;
//   - @AfterReturning with a returning variable: the method's return type
//     must be compatible with the variable's type (Spring's
//     AspectJAfterReturningAdvice);
//   - a join point is a method a proxy can intercept on a bean class: an
//     application or bean class related to a bean type, its own and
//     inherited methods (not java.lang.Object's), neither static, private,
//     final, abstract, a constructor nor synthetic; the aspect classes
//     themselves are not advised;
//   - a call site matches when it is a virtual or interface call of the
//     join point's name and descriptor whose owner the bean class is
//     assignable to (the call may dispatch to that bean). A call on `this`
//     from the bean's own code (self-invocation) bypasses the proxy and
//     gets no edge;
//   - gate: for @Aspect advice a reachable method creating one of Spring's
//     AspectJ advice objects (AbstractAspectJAdvice subtypes, made by
//     ReflectiveAspectJAdvisorFactory#getAdvice); for <aop:config> advice
//     (whose advice objects the container creates reflectively) a
//     reachable ConfigBeanDefinitionParser#parse.
//
// @AfterReturning edges carry the condition "returned" and
// @AfterThrowing edges "exceptional" (path evidence and the report's
// arrow). For @Around the target still runs only if the advice calls
// proceed(); the existing edge to the target is kept (an
// over-approximation). Advice order (@Order), the throwing type
// restriction, <aop:advisor>, the AspectJ compiler or load-time weaving,
// the difference between this() on a JDK proxy and on the target, and
// JDK-proxy interception of final methods reached through an interface are
// not modeled.
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
	"ariadne/internal/springxml"
	"ariadne/internal/valueflow"
)

// ModelID identifies the model; its edges carry callgraph.BasisModel.
const ModelID = "spring:aop"

// Version versions the rules. Snapshots record it.
const Version = "a2"

const (
	aspectAnn      = "org.aspectj.lang.annotation.Aspect"
	pointcutAnn    = "org.aspectj.lang.annotation.Pointcut"
	abstractAdvice = "org.springframework.aop.aspectj.AbstractAspectJAdvice"
	configParser   = "org.springframework.aop.config.ConfigBeanDefinitionParser"
	objectFQCN     = "java.lang.Object"
	accPrivate     = 0x0002
	accBridge      = 0x0040
	accSynthetic   = 0x1000
	accAnnotation  = 0x2000
	// maxEdges bounds the candidate edges of one input; past it the
	// model stops adding and reports it (Result.Truncated).
	maxEdges = 50000
)

// adviceKinds maps the advice annotations and <aop:…> elements to the
// advice kind.
var adviceKinds = map[string]string{
	"org.aspectj.lang.annotation.Before":         "Before",
	"org.aspectj.lang.annotation.After":          "After",
	"org.aspectj.lang.annotation.Around":         "Around",
	"org.aspectj.lang.annotation.AfterReturning": "AfterReturning",
	"org.aspectj.lang.annotation.AfterThrowing":  "AfterThrowing",
	"before":          "Before",
	"after":           "After",
	"around":          "Around",
	"after-returning": "AfterReturning",
	"after-throwing":  "AfterThrowing",
}

// conditions are the edge conditions by advice kind.
var conditions = map[string]string{"AfterReturning": "returned", "AfterThrowing": "exceptional"}

var joinPointTypes = map[string]bool{
	"org.aspectj.lang.JoinPoint":            true,
	"org.aspectj.lang.ProceedingJoinPoint":  true,
	"org.aspectj.lang.JoinPoint$StaticPart": true,
}

// Input is what the spring profile knows about the application's beans.
type Input struct {
	// Beans are the bean types (stereotypes, @Bean products, XML beans).
	Beans []string
	// Names are the bean names by bean type.
	Names map[string][]string
	// XML are the <aop:config> sections of the input's bean XML.
	XML []springxml.AOPConfig
}

// Advice is one advice method.
type Advice struct {
	Aspect, Name, Desc string
	// Kind is the advice kind (Before, Around, …); XML is set for
	// <aop:config> advice.
	Kind       string
	XML        bool
	Expression string
	// Registered: the aspect class is a bean.
	Registered bool
	// Returning is the returning variable's type (@AfterReturning), ""
	// when not restricted.
	Returning string
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

// Pair is one advice matched with one join point.
type Pair struct {
	Aspect, Advice string
	JoinPoint
}

// Result is the model applied to one input.
type Result struct {
	Aspects []string
	Advice  []Advice
	// Unresolved lists <aop:config> advice whose aspect bean or method
	// was not found.
	Unresolved []string
	// JoinPoints are the bean methods matched by at least one advice.
	JoinPoints []JoinPoint
	// Pairs are the advice × join point matches, in match order.
	Pairs []Pair
	// SelfCalls counts call sites left out as self-invocation.
	SelfCalls int
	// Edges are the candidates for callgraph.NewCreatorGate.
	Edges []callgraph.SAMEdge
	// NoCreator: advice exists but nothing creates Spring's advice
	// objects (nor parses <aop:config>).
	NoCreator bool
	// Truncated: the edge budget was hit.
	Truncated bool
}

type adviceMatch struct {
	adv int
	jp  joinPoint
}

// Find builds the aspect catalog and the candidate edges.
func Find(h *hierarchy.Hierarchy, in Input) Result {
	var r Result
	type parsed struct {
		n   node
		res *resolver
	}
	var nodes []parsed
	aspects := map[string]bool{}
	add := func(adv Advice, n node, res *resolver) {
		why := map[string]bool{}
		unknownReasons(n, res, why, 0)
		adv.Unknown = sortedKeys(why)
		r.Advice = append(r.Advice, adv)
		nodes = append(nodes, parsed{n, res})
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp || ci.IsInterface() || !hasAnn(ci.Annotations, aspectAnn) {
			continue
		}
		aspects[fqcn] = true
		r.Aspects = append(r.Aspects, fqcn)
		reg := registered(h, fqcn, in.Beans)
		res := newResolver(h, fqcn)
		for _, m := range ci.Methods {
			for _, a := range m.AnnotationDetails {
				kind, ok := adviceKinds[a.Type]
				if !ok {
					continue
				}
				expr, _ := stringElement(a, "pointcut")
				if strings.TrimSpace(expr) == "" {
					expr, _ = stringElement(a, "value")
				}
				argNames, _ := stringElement(a, "argNames")
				returning, _ := stringElement(a, "returning")
				throwing, _ := stringElement(a, "throwing")
				p := parse(expr)
				sc := methodScope(h, p, m.Method, argNames, returning, throwing)
				adv := Advice{Aspect: fqcn, Name: m.Name, Desc: m.Descriptor, Kind: kind, Expression: expr, Registered: reg}
				if returning != "" {
					adv.Returning = sc[returning]
				}
				add(adv, bind(p, sc), res)
				break
			}
		}
	}
	// <aop:config> advice.
	byName := map[string]string{}
	for t, names := range in.Names {
		for _, n := range names {
			byName[n] = t
		}
	}
	xmlAdvice := false
	for _, cfg := range in.XML {
		for _, asp := range cfg.Aspects {
			cls, ok := byName[asp.Ref]
			for _, a := range asp.Advice {
				kind := adviceKinds[a.Kind]
				expr := a.Pointcut
				if expr == "" {
					expr = asp.Pointcuts[a.PointcutRef]
					if expr == "" {
						expr = cfg.Pointcuts[a.PointcutRef]
					}
				}
				m, found := hierarchy.Method{}, false
				if ok {
					m, found = adviceMethod(h, cls, a.Method)
				}
				if !found || expr == "" {
					r.Unresolved = append(r.Unresolved, asp.Ref+"."+a.Method)
					continue
				}
				if !aspects[cls] {
					aspects[cls] = true
					r.Aspects = append(r.Aspects, cls)
				}
				p := parseXML(expr)
				sc := methodScope(h, p, m.Method, a.ArgNames, a.Returning, a.Throwing)
				adv := Advice{Aspect: cls, Name: m.Name, Desc: m.Descriptor, Kind: kind, XML: true, Expression: expr, Registered: true}
				if a.Returning != "" {
					adv.Returning = sc[a.Returning]
				}
				add(adv, bind(p, sc), newResolver(h, cls))
				xmlAdvice = true
			}
		}
	}
	if len(r.Advice) == 0 {
		return r
	}
	gate := creators(h, func(ci *hierarchy.ClassInfo, m hierarchy.Method, advices map[string]bool) bool {
		for _, ref := range m.InitRefs {
			if ref.Kind == classfile.InitNew && advices[ref.Owner] {
				return true
			}
		}
		return false
	})
	var xmlGate []callgraph.MethodRef
	if xmlAdvice {
		if ci, ok := h.Lookup(configParser); ok {
			for _, m := range ci.Methods {
				if m.Name == "parse" {
					xmlGate = append(xmlGate, callgraph.MethodRef{FQCN: configParser, Name: m.Name, Desc: m.Descriptor})
				}
			}
		}
	}
	if len(gate) == 0 && len(xmlGate) == 0 {
		r.NoCreator = true
	}

	// Match every registered advice against every join point.
	var matches []adviceMatch
	matchedJP := map[JoinPoint]bool{}
	for _, c := range beanClasses(h, in.Beans, aspects) {
		names := beanNames(h, c, in.Names)
		for _, jp := range joinPoints(h, c) {
			jp := jp
			jp.names = names
			for i := range r.Advice {
				a := &r.Advice[i]
				if !a.Registered {
					continue
				}
				res := nodes[i].n.eval(&jp, nodes[i].res)
				if res != no && a.Returning != "" {
					res = and3(res, returnMatches(h, jp.m.Descriptor, a.Returning))
				}
				switch res {
				case yes:
					a.Matched++
					matches = append(matches, adviceMatch{adv: i, jp: jp})
					j := JoinPoint{Bean: jp.C, Declaring: jp.D, Name: jp.m.Name, Desc: jp.m.Descriptor}
					matchedJP[j] = true
					r.Pairs = append(r.Pairs, Pair{Aspect: a.Aspect, Advice: a.Name, JoinPoint: j})
				case unknown:
					a.Undecided++
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
	if len(matches) == 0 || r.NoCreator {
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
		g := gate
		if adv.XML {
			g = xmlGate
		}
		if len(g) == 0 {
			continue
		}
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
				Creators: g, Condition: conditions[adv.Kind],
			})
		}
	}
	return r
}

// creators lists the methods creating an AbstractAspectJAdvice subtype.
func creators(h *hierarchy.Hierarchy, creates func(*hierarchy.ClassInfo, hierarchy.Method, map[string]bool) bool) []callgraph.MethodRef {
	advices := map[string]bool{abstractAdvice: true}
	for _, s := range h.Subtypes(abstractAdvice) {
		advices[s] = true
	}
	var out []callgraph.MethodRef
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			if creates(ci, m, advices) {
				out = append(out, callgraph.MethodRef{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor})
			}
		}
	}
	return out
}

// adviceMethod finds an <aop:…> advice method by name on the aspect class
// or its superclasses: Spring's BeanUtils.resolveSignature takes the
// overload with the fewest parameters.
func adviceMethod(h *hierarchy.Hierarchy, cls, name string) (hierarchy.Method, bool) {
	for cur := cls; cur != ""; {
		ci, ok := h.Lookup(cur)
		if !ok {
			break
		}
		var best hierarchy.Method
		found := false
		for _, m := range ci.Methods {
			if m.Name != name {
				continue
			}
			p, _ := descParts(m.Descriptor)
			bp, _ := descParts(best.Descriptor)
			if !found || len(p) < len(bp) {
				best, found = m, true
			}
		}
		if found {
			return best, true
		}
		cur = ci.SuperClass
	}
	return hierarchy.Method{}, false
}

// returnMatches: Spring runs @AfterReturning only for a return value of
// the returning variable's type — Object accepts every method, void
// included; otherwise the declared return type must be compatible.
func returnMatches(h *hierarchy.Hierarchy, desc, typ string) tri {
	_, ret := descParts(desc)
	if typ == objectFQCN {
		return yes
	}
	if ret == "void" {
		return no
	}
	return instanceMayMatch(h, ret, typ)
}

// beanNames are the names of the beans whose type is related to c.
func beanNames(h *hierarchy.Hierarchy, c string, names map[string][]string) []string {
	var out []string
	for t, ns := range names {
		if t == c || assignable(h, t, c) || assignable(h, c, t) && t != objectFQCN {
			out = append(out, ns...)
		}
	}
	sort.Strings(out)
	return out
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
	if h == nil {
		return false
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

// XMLCount counts the <aop:config> advice.
func (r Result) XMLCount() int {
	n := 0
	for _, a := range r.Advice {
		if a.XML {
			n++
		}
	}
	return n
}
