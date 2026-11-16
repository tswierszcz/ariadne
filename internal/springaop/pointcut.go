package springaop

import (
	"regexp"
	"strings"
	"sync"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// tri is a three-valued match: an expression the model does not
// understand makes the result unknown instead of a guess.
type tri int8

const (
	no tri = iota
	yes
	unknown
)

func and3(a, b tri) tri {
	switch {
	case a == no || b == no:
		return no
	case a == yes && b == yes:
		return yes
	}
	return unknown
}

func or3(a, b tri) tri {
	switch {
	case a == yes || b == yes:
		return yes
	case a == no && b == no:
		return no
	}
	return unknown
}

func not3(a tri) tri {
	switch a {
	case yes:
		return no
	case no:
		return yes
	}
	return unknown
}

// joinPoint is one method execution Spring AOP can intercept: method m,
// declared in class D, called on a bean of class C (C is D or a subtype).
type joinPoint struct {
	h    *hierarchy.Hierarchy
	C, D string
	m    hierarchy.Method
}

// node is a parsed pointcut expression.
type node interface {
	eval(jp *joinPoint, r *resolver) tri
}

type andNode struct{ l, r node }
type orNode struct{ l, r node }
type notNode struct{ x node }

func (n andNode) eval(jp *joinPoint, r *resolver) tri {
	l := n.l.eval(jp, r)
	if l == no {
		return no
	}
	return and3(l, n.r.eval(jp, r))
}

func (n orNode) eval(jp *joinPoint, r *resolver) tri {
	l := n.l.eval(jp, r)
	if l == yes {
		return yes
	}
	return or3(l, n.r.eval(jp, r))
}

func (n notNode) eval(jp *joinPoint, r *resolver) tri { return not3(n.x.eval(jp, r)) }

// unknownNode is a designator or form the model does not evaluate.
type unknownNode struct{ why string }

func (n unknownNode) eval(*joinPoint, *resolver) tri { return unknown }

// withinNode: within(TypePattern) — the method is declared in a matching
// type.
type withinNode struct{ pat string }

func (n withinNode) eval(jp *joinPoint, _ *resolver) tri {
	return typeMatches(jp.h, n.pat, jp.D)
}

// annNode: @annotation(Type) on the executed method, @within(Type) on its
// declaring type.
type annNode struct {
	ann    string
	onType bool
}

func (n annNode) eval(jp *joinPoint, _ *resolver) tri {
	var anns []string
	if n.onType {
		ci, ok := jp.h.Lookup(jp.D)
		if !ok {
			return unknown
		}
		anns = ci.Annotations
	} else {
		anns = jp.m.Annotations
	}
	for _, a := range anns {
		if a == n.ann {
			return yes
		}
	}
	return no
}

// refNode is a reference to a named @Pointcut method.
type refNode struct{ qualifier, name string }

func (n refNode) eval(jp *joinPoint, r *resolver) tri {
	x, ok := r.lookup(n.qualifier, n.name)
	if !ok {
		return unknown
	}
	return x.eval(jp, r)
}

// execNode: execution(modifiers? ret declaring.name(params) throws?).
type execNode struct {
	mods   []string
	ret    string
	decl   string // "" = any declaring type
	name   string
	params []string
}

func (n execNode) eval(jp *joinPoint, _ *resolver) tri {
	m := jp.m
	if !nameMatches(n.name, m.Name) {
		return no
	}
	res := yes
	for _, mod := range n.mods {
		neg := strings.HasPrefix(mod, "!")
		flag, ok := modifierFlags[strings.TrimPrefix(mod, "!")]
		if !ok {
			return unknown
		}
		has := m.AccessFlags&flag != 0
		if has == neg {
			return no
		}
	}
	params, ret := descParts(m.Descriptor)
	res = and3(res, typeMatches(jp.h, n.ret, ret))
	if res == no {
		return no
	}
	res = and3(res, paramsMatch(jp.h, n.params, params))
	if res == no {
		return no
	}
	if n.decl == "" {
		return res
	}
	// The declaring-type pattern matches any type of the target's
	// hierarchy declaring the method (AspectJ's signature matching), the
	// target class itself included.
	decl := no
	for _, t := range declaringTypes(jp) {
		decl = or3(decl, typeMatches(jp.h, n.decl, t))
		if decl == yes {
			break
		}
	}
	return and3(res, decl)
}

var modifierFlags = map[string]uint16{
	"public": classfile.AccPublic, "private": 0x0002, "protected": 0x0004,
	"static": classfile.AccStatic, "final": classfile.AccFinal, "synchronized": 0x0020,
}

func declaringTypes(jp *joinPoint) []string {
	out := []string{jp.C}
	seen := map[string]bool{jp.C: true}
	var up func(string)
	up = func(c string) {
		ci, ok := jp.h.Lookup(c)
		if !ok {
			return
		}
		for _, s := range append([]string{ci.SuperClass}, ci.Interfaces...) {
			if s == "" || seen[s] {
				continue
			}
			seen[s] = true
			if sci, ok := jp.h.Lookup(s); ok {
				if _, ok := sci.MethodByNameAndDesc(jp.m.Name, jp.m.Descriptor); ok {
					out = append(out, s)
				}
			}
			up(s)
		}
	}
	up(jp.C)
	return out
}

// parse parses a pointcut expression. A syntax the parser does not
// accept becomes an unknownNode, never an error that drops the advice
// silently.
func parse(expr string) node {
	p := &parser{s: expr}
	n := p.or()
	p.ws()
	if p.err != "" {
		return unknownNode{p.err}
	}
	if p.i < len(p.s) {
		return unknownNode{"unparsed: " + p.s[p.i:]}
	}
	return n
}

type parser struct {
	s   string
	i   int
	err string
}

func (p *parser) ws() {
	for p.i < len(p.s) && (p.s[p.i] == ' ' || p.s[p.i] == '\t' || p.s[p.i] == '\n' || p.s[p.i] == '\r') {
		p.i++
	}
}

func (p *parser) eat(tok string) bool {
	p.ws()
	if strings.HasPrefix(p.s[p.i:], tok) {
		p.i += len(tok)
		return true
	}
	return false
}

func (p *parser) fail(msg string) node {
	if p.err == "" {
		p.err = msg
	}
	p.i = len(p.s)
	return unknownNode{msg}
}

func (p *parser) or() node {
	n := p.and()
	for p.err == "" && p.eat("||") {
		n = orNode{n, p.and()}
	}
	return n
}

func (p *parser) and() node {
	n := p.unary()
	for p.err == "" && p.eat("&&") {
		n = andNode{n, p.unary()}
	}
	return n
}

func (p *parser) unary() node {
	if p.eat("!") {
		return notNode{p.unary()}
	}
	if p.eat("(") {
		n := p.or()
		if !p.eat(")") {
			return p.fail("unbalanced parenthesis")
		}
		return n
	}
	return p.primary()
}

// primary: designator(args) or a named pointcut reference.
func (p *parser) primary() node {
	p.ws()
	start := p.i
	for p.i < len(p.s) && p.s[p.i] != '(' && p.s[p.i] != ' ' && p.s[p.i] != ')' && p.s[p.i] != '&' && p.s[p.i] != '|' {
		p.i++
	}
	head := p.s[start:p.i]
	p.ws()
	if head == "" || p.i >= len(p.s) || p.s[p.i] != '(' {
		return p.fail("expected a designator at: " + p.s[start:])
	}
	p.i++ // (
	depth, argStart := 1, p.i
	for p.i < len(p.s) && depth > 0 {
		switch p.s[p.i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		p.i++
	}
	if depth != 0 {
		return p.fail("unbalanced parenthesis")
	}
	args := strings.TrimSpace(p.s[argStart : p.i-1])
	switch head {
	case "execution":
		return parseExecution(args)
	case "within":
		if !isTypePattern(args) {
			return unknownNode{"within(" + args + ")"}
		}
		return withinNode{args}
	case "@annotation", "@within":
		if !isQualifiedName(args) {
			// A binding form (@annotation(name)): the advice method
			// resolves it (bindAnnotations); left unknown here.
			return bindingNode{designator: head, name: args}
		}
		return annNode{ann: args, onType: head == "@within"}
	case "args", "this", "target", "@args", "@target", "bean", "call", "get", "set",
		"initialization", "preinitialization", "staticinitialization", "handler",
		"adviceexecution", "withincode", "@withincode", "cflow", "cflowbelow", "if":
		return unknownNode{head + "(" + args + ")"}
	}
	if args != "" {
		return unknownNode{"named pointcut with arguments: " + head + "(" + args + ")"}
	}
	q, name := "", head
	if i := strings.LastIndexByte(head, '.'); i >= 0 {
		q, name = head[:i], head[i+1:]
	}
	return refNode{qualifier: q, name: name}
}

// bindingNode is @annotation(param)/@within(param): the annotation type is
// the advice parameter's type. Advice.bind replaces it with an annNode
// when the type is known.
type bindingNode struct{ designator, name string }

func (n bindingNode) eval(*joinPoint, *resolver) tri { return unknown }

func parseExecution(args string) node {
	s := strings.TrimSpace(args)
	if strings.ContainsAny(s, "@") {
		return unknownNode{"annotation pattern in execution(" + args + ")"}
	}
	// The parameter list: the last top-level parenthesis group.
	close := strings.LastIndexByte(s, ')')
	if close < 0 {
		return unknownNode{"execution(" + args + ") without parameters"}
	}
	open := strings.LastIndexByte(s[:close], '(')
	if open < 0 {
		return unknownNode{"execution(" + args + ")"}
	}
	rest := strings.TrimSpace(s[close+1:])
	if rest != "" && !strings.HasPrefix(rest, "throws") {
		return unknownNode{"execution(" + args + ")"}
	}
	head := strings.Fields(s[:open])
	if len(head) < 2 {
		return unknownNode{"execution(" + args + ") without a return type"}
	}
	n := execNode{mods: head[:len(head)-2], ret: head[len(head)-2]}
	dn := head[len(head)-1]
	if i := strings.LastIndexByte(dn, '.'); i >= 0 {
		n.decl, n.name = dn[:i], dn[i+1:]
		if strings.HasSuffix(n.decl, ".") {
			// com.x..*(..): any type below com.x.
			n.decl += ".*"
		}
	} else {
		n.name = dn
	}
	if !isTypePattern(n.ret) || n.decl != "" && !isTypePattern(n.decl) || !isNamePattern(n.name) {
		return unknownNode{"execution(" + args + ")"}
	}
	ps := strings.TrimSpace(s[open+1 : close])
	if ps != "" {
		for _, x := range strings.Split(ps, ",") {
			x = strings.TrimSpace(x)
			if x != ".." && !isTypePattern(x) {
				return unknownNode{"execution(" + args + ")"}
			}
			n.params = append(n.params, x)
		}
	}
	return n
}

var (
	typePatternRe = regexp.MustCompile(`^[A-Za-z0-9_$.*]+\+?(\[\])*$`)
	namePatternRe = regexp.MustCompile(`^[A-Za-z0-9_$*]+$`)
	qualifiedRe   = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*(\.[A-Za-z_$][A-Za-z0-9_$]*)+$`)
)

func isTypePattern(s string) bool   { return typePatternRe.MatchString(s) }
func isNamePattern(s string) bool   { return namePatternRe.MatchString(s) }
func isQualifiedName(s string) bool { return qualifiedRe.MatchString(s) }

func nameMatches(pat, name string) bool {
	if pat == "*" {
		return true
	}
	return globRegexp(pat).MatchString(name)
}

var (
	reMu    sync.Mutex
	reCache = map[string]*regexp.Regexp{}
)

// globRegexp compiles a type or name pattern: * is any run of characters
// but a dot, .. any run of packages.
func globRegexp(pat string) *regexp.Regexp {
	reMu.Lock()
	defer reMu.Unlock()
	if re, ok := reCache[pat]; ok {
		return re
	}
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pat); {
		switch {
		case strings.HasPrefix(pat[i:], ".."):
			b.WriteString(`\.(?:[^.]+\.)*`)
			i += 2
		case pat[i] == '*':
			b.WriteString(`[^.]*`)
			i++
		default:
			b.WriteString(regexp.QuoteMeta(pat[i : i+1]))
			i++
		}
	}
	b.WriteString("$")
	re := regexp.MustCompile(b.String())
	reCache[pat] = re
	return re
}

var primitives = map[string]bool{"void": true, "boolean": true, "byte": true, "char": true, "short": true, "int": true, "long": true, "float": true, "double": true}

// typeMatches matches a type pattern against a type in source form
// (java.lang.String, int, a.B[]; nested classes with $).
func typeMatches(h *hierarchy.Hierarchy, pat, typ string) tri {
	for strings.HasSuffix(pat, "[]") {
		if !strings.HasSuffix(typ, "[]") {
			return no
		}
		pat, typ = strings.TrimSuffix(pat, "[]"), strings.TrimSuffix(typ, "[]")
	}
	plus := strings.HasSuffix(pat, "+")
	pat = strings.TrimSuffix(pat, "+")
	if pat == "*" {
		return yes
	}
	if strings.HasSuffix(typ, "[]") {
		return no
	}
	if primitives[pat] || primitives[typ] {
		if pat == typ {
			return yes
		}
		return no
	}
	pats := []string{pat}
	if !strings.ContainsAny(pat, ".*") {
		// Annotation-style pointcuts name types fully qualified; only
		// java.lang resolves without an import.
		pats = []string{"java.lang." + pat}
	}
	match := func(t string) bool {
		t = strings.ReplaceAll(t, "$", ".")
		for _, p := range pats {
			if globRegexp(p).MatchString(t) {
				return true
			}
		}
		return false
	}
	if !plus {
		if match(typ) {
			return yes
		}
		return no
	}
	for _, t := range supertypes(h, typ) {
		if match(t) {
			return yes
		}
	}
	return no
}

func supertypes(h *hierarchy.Hierarchy, t string) []string {
	out := []string{}
	seen := map[string]bool{}
	var up func(string)
	up = func(c string) {
		if c == "" || seen[c] {
			return
		}
		seen[c] = true
		out = append(out, c)
		if ci, ok := h.Lookup(c); ok {
			up(ci.SuperClass)
			for _, i := range ci.Interfaces {
				up(i)
			}
		}
	}
	up(t)
	return out
}

// paramsMatch matches a parameter pattern list (with .. for any number of
// parameters) against the method's parameter types.
func paramsMatch(h *hierarchy.Hierarchy, pats, types []string) tri {
	if len(pats) == 0 {
		if len(types) == 0 {
			return yes
		}
		return no
	}
	if pats[0] == ".." {
		res := no
		for k := 0; k <= len(types); k++ {
			res = or3(res, paramsMatch(h, pats[1:], types[k:]))
			if res == yes {
				return yes
			}
		}
		return res
	}
	if len(types) == 0 {
		return no
	}
	first := typeMatches(h, pats[0], types[0])
	if first == no {
		return no
	}
	return and3(first, paramsMatch(h, pats[1:], types[1:]))
}

// descParts renders a method descriptor's parameter and return types in
// source form.
func descParts(desc string) ([]string, string) {
	end := strings.IndexByte(desc, ')')
	if !strings.HasPrefix(desc, "(") || end < 0 {
		return nil, ""
	}
	var params []string
	s := desc[1:end]
	for len(s) > 0 {
		t, n := sourceType(s)
		if n == 0 {
			break
		}
		params = append(params, t)
		s = s[n:]
	}
	ret, _ := sourceType(desc[end+1:])
	return params, ret
}

var primitiveDesc = map[byte]string{'V': "void", 'Z': "boolean", 'B': "byte", 'C': "char", 'S': "short", 'I': "int", 'J': "long", 'F': "float", 'D': "double"}

// sourceType renders the field descriptor at the start of s and returns
// how many bytes it took.
func sourceType(s string) (string, int) {
	dims := 0
	for dims < len(s) && s[dims] == '[' {
		dims++
	}
	if dims >= len(s) {
		return "", 0
	}
	var base string
	n := dims + 1
	if s[dims] == 'L' {
		j := strings.IndexByte(s, ';')
		if j < 0 {
			return "", 0
		}
		base = strings.ReplaceAll(s[dims+1:j], "/", ".")
		n = j + 1
	} else {
		base = primitiveDesc[s[dims]]
	}
	return base + strings.Repeat("[]", dims), n
}

// resolver resolves named pointcut references from one aspect.
type resolver struct {
	h      *hierarchy.Hierarchy
	aspect string
	// cache holds parsed @Pointcut methods by class and name; active
	// guards recursion.
	cache  map[string]node
	active map[string]bool
}

func newResolver(h *hierarchy.Hierarchy, aspect string) *resolver {
	return &resolver{h: h, aspect: aspect, cache: map[string]node{}, active: map[string]bool{}}
}

func (r *resolver) lookup(qualifier, name string) (node, bool) {
	var classes []string
	if qualifier == "" {
		classes = supertypes(r.h, r.aspect)
	} else {
		classes = []string{qualifier, strings.ReplaceAll(qualifier, ".", "$")}
		if i := strings.LastIndexByte(r.aspect, '.'); i >= 0 && !strings.Contains(qualifier, ".") {
			pkg := r.aspect[:i]
			classes = append(classes, pkg+"."+qualifier, r.aspect+"$"+qualifier)
		}
	}
	for _, c := range classes {
		key := c + "#" + name
		if n, ok := r.cache[key]; ok {
			return wrapActive{n, key}, true
		}
		ci, ok := r.h.Lookup(c)
		if !ok {
			continue
		}
		for _, m := range ci.Methods {
			if m.Name != name {
				continue
			}
			for _, a := range m.AnnotationDetails {
				if a.Type != pointcutAnn {
					continue
				}
				if r.active[key] {
					return nil, false
				}
				expr, _ := stringElement(a, "value")
				n := parse(expr)
				r.cache[key] = n
				return wrapActive{n, key}, true
			}
		}
	}
	return nil, false
}

// wrapActive guards a named pointcut against referring to itself.
type wrapActive struct {
	n   node
	key string
}

func (w wrapActive) eval(jp *joinPoint, r *resolver) tri {
	if r.active[w.key] {
		return unknown
	}
	r.active[w.key] = true
	defer delete(r.active, w.key)
	return w.n.eval(jp, r)
}

func stringElement(a classfile.Annotation, name string) (string, bool) {
	v, ok := a.Value(name)
	if !ok || v.Kind != classfile.ElementConst {
		return "", false
	}
	s, ok := v.Const.(string)
	return s, ok
}

// unknownReasons collects why an expression is not fully understood.
func unknownReasons(n node, r *resolver, out map[string]bool, depth int) {
	if depth > 8 {
		return
	}
	switch x := n.(type) {
	case andNode:
		unknownReasons(x.l, r, out, depth+1)
		unknownReasons(x.r, r, out, depth+1)
	case orNode:
		unknownReasons(x.l, r, out, depth+1)
		unknownReasons(x.r, r, out, depth+1)
	case notNode:
		unknownReasons(x.x, r, out, depth+1)
	case unknownNode:
		out[x.why] = true
	case bindingNode:
		out[x.designator+"("+x.name+") bound to an advice parameter of unknown type"] = true
	case refNode:
		if y, ok := r.lookup(x.qualifier, x.name); ok {
			unknownReasons(y, r, out, depth+1)
		} else {
			out["unresolved pointcut reference "+strings.TrimPrefix(x.qualifier+"."+x.name, ".")+"()"] = true
		}
	case wrapActive:
		unknownReasons(x.n, r, out, depth+1)
	}
}
