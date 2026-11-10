package javadiff

import (
	"regexp"
	"strings"
)

// methodHeaderRe matches a method/constructor declaration's tail: an
// identifier directly followed by a balanced, paren-free argument list
// and an optional throws clause, with nothing else before the opening
// brace. Not anchored at the start, since the text before the name
// (modifiers, return type, generics, annotations) is irrelevant; anchored
// at the end, since that's what forces the match onto the *last*
// identifier+parens pair in the header rather than some unrelated one
// earlier in it (an annotation's own "(...)", say).
var methodHeaderRe = regexp.MustCompile(`(\w+)\s*\(([^()]*)\)\s*(?:throws\s+[\w.,\s<>\[\]]+)?\s*$`)

// newExprRe matches "new Foo(...)" / "new Foo<Bar>(...)" at the end of a
// header: an anonymous class body, not a method declaration, even though
// it has the same "identifier(args)" shape methodHeaderRe looks for.
var newExprRe = regexp.MustCompile(`(?:^|\W)new\s+[\w.]+\s*(?:<[^>]*>)?\s*\([^()]*\)\s*$`)

// classHeaderRe matches a class/interface/enum/record declaration and
// captures its simple name, so a same-named method further down (a
// constructor) can be recognized and renamed to "<init>" — the name the
// JVM (and classfile.Method) actually uses, so vulnmap's output lines up
// with the call graph's own keys.
var classHeaderRe = regexp.MustCompile(`\b(?:class|interface|enum|record)\s+(\w+)`)

var controlKeywords = map[string]bool{
	"if": true, "for": true, "while": true, "switch": true, "catch": true,
	"synchronized": true, "do": true, "else": true, "try": true, "finally": true,
}

type frameKind byte

const (
	frameOther  frameKind = iota // a plain block: if/for/lambda body, static/instance initializer, ...
	frameClass                   // a class/interface/enum/record body
	frameMethod                  // a method or constructor body
	frameAnon                    // an anonymous class body ("new Foo() {")
)

type frame struct {
	kind   frameKind
	name   string
	params []string
	line   int  // the line the body opened on
	iface  bool // a frameClass for an interface or annotation type (its fields are static)
}

// InitializerParams marks the "<init>" MethodRef of instance field
// initializers and initializer blocks: the compiler copies them into every
// constructor, so they belong to all overloads, not to one.
var InitializerParams = []string{"*"}

// MethodRef identifies a method declared in a source file (CVE:F5b).
// Class is its declaring class's binary name relative to the package —
// "Outer" for a top-level class, "Outer$Inner" for a member class — so
// package + "." + Class is the bytecode class name. Params are the source
// parameter types reduced to simple names with array dimensions ("String",
// "int[]", "Entry"); "?" marks a type whose erasure cannot be known from
// the source alone (a type variable). StartLine/EndLine bound the body.
//
// Methods of anonymous and local classes are not separate MethodRefs:
// their bytecode names (Outer$1, Outer$1Local) depend on compiler
// numbering, so their lines belong to the enclosing method instead.
type MethodRef struct {
	Class     string
	Name      string
	Params    []string
	StartLine int
	EndLine   int
}

// Key is "Class#name(p1,p2)", unique per overload within a file.
func (m MethodRef) Key() string {
	return m.Class + "#" + m.Name + "(" + strings.Join(m.Params, ",") + ")"
}

// FindMethodsForLines returns the name of the method/constructor
// enclosing each of the given 1-based line numbers in src, as "<init>"
// for a constructor and the plain method name otherwise (see
// MethodsForLines for the declaring class and parameters).
//
// A line outside any method (a field initializer, a class-body line, an
// import) contributes nothing, not an error: fix commits routinely touch
// such lines too, and the caller simply gets fewer methods out of it.
func FindMethodsForLines(src string, lines []int) []string {
	seen := make(map[string]bool)
	var out []string
	for _, m := range MethodsForLines(src, lines) {
		if !seen[m.Name] {
			seen[m.Name] = true
			out = append(out, m.Name)
		}
	}
	return out
}

// MethodsForLines returns the declared method owning each given 1-based
// line, deduplicated, in line order.
func MethodsForLines(src string, lines []int) []MethodRef {
	owners, decl := scanMethods(stripCommentsAndStrings(src))
	seen := make(map[int]bool)
	var out []MethodRef
	for _, l := range lines {
		if l < 1 || l >= len(owners) {
			continue
		}
		i := owners[l]
		if i < 0 || seen[i] {
			continue
		}
		seen[i] = true
		out = append(out, decl[i])
	}
	return out
}

// Declared returns every method/constructor with a body declared in src
// by a named (top-level or member) class, in declaration order.
func Declared(src string) []MethodRef {
	_, decl := scanMethods(stripCommentsAndStrings(src))
	return decl
}

// DeclaredMethods returns the name of every method/constructor with a
// body declared in src ("<init>" for constructors), the same names
// FindMethodsForLines returns. Abstract and interface methods without a
// body are not listed.
func DeclaredMethods(src string) map[string]bool {
	out := make(map[string]bool)
	for _, m := range Declared(src) {
		out[m.Name] = true
	}
	return out
}

// scanMethods returns, per 1-based line, the index into decl of the
// method owning it (-1 for none), and decl itself.
func scanMethods(cleaned string) ([]int, []MethodRef) {
	totalLines := strings.Count(cleaned, "\n") + 1
	owner := make([]int, totalLines+1)
	for i := range owner {
		owner[i] = -1
	}
	var decl []MethodRef
	var stack []frame
	var stackDecl []int // per frame: index into decl for a named method frame, else -1
	currentLine := 1
	headerStart := 0

	// innermost returns the innermost method frame that belongs to a
	// named class reachable through class frames only (not inside an
	// anonymous or local class).
	innermost := func() int {
		best := -1
		opaque := false
		inMethod := false
		for i, f := range stack {
			switch f.kind {
			case frameClass:
				if inMethod {
					opaque = true // a local class
				}
			case frameAnon:
				opaque = true
			case frameMethod:
				if !opaque {
					best = stackDecl[i]
				}
				inMethod = true
			}
		}
		return best
	}
	classPath := func() (string, bool) {
		var names []string
		for _, f := range stack {
			switch f.kind {
			case frameClass:
				names = append(names, f.name)
			case frameMethod, frameAnon:
				return "", false
			}
		}
		return strings.Join(names, "$"), len(names) > 0
	}
	headerLine := 1
	// Field initializers and initializer blocks (CVE:F10): their code is
	// the class's <clinit> (static) or every constructor (instance) in
	// bytecode. One synthetic MethodRef per class and kind.
	initDecl := map[string]int{}
	initializer := func(static bool) int {
		cls, ok := classPath()
		if !ok {
			return -1
		}
		name, params := "<init>", InitializerParams
		if static {
			name, params = "<clinit>", nil
		}
		k := cls + "#" + name
		if i, ok := initDecl[k]; ok {
			return i
		}
		initDecl[k] = len(decl)
		decl = append(decl, MethodRef{Class: cls, Name: name, Params: params, StartLine: currentLine})
		return initDecl[k]
	}
	topIsClass := func() (frame, bool) {
		if n := len(stack); n > 0 && stack[n-1].kind == frameClass {
			return stack[n-1], true
		}
		return frame{}, false
	}
	currentClass := func() string {
		for i := len(stack) - 1; i >= 0; i-- {
			if stack[i].kind == frameClass {
				return stack[i].name
			}
		}
		return ""
	}

	for i := 0; i < len(cleaned); i++ {
		switch cleaned[i] {
		case '\n':
			if owner[currentLine] < 0 {
				owner[currentLine] = innermost()
			}
			currentLine++
		case '{':
			header := cleaned[headerStart:i]
			f := classifyHeader(header, currentClass())
			f.line = currentLine
			idx := -1
			if f.kind == frameMethod {
				if cls, ok := classPath(); ok {
					idx = len(decl)
					decl = append(decl, MethodRef{Class: cls, Name: f.name, Params: f.params, StartLine: currentLine})
				}
			} else if f.kind == frameOther {
				// An initializer block directly in a class body.
				if _, ok := topIsClass(); ok {
					switch trimSpace(header) {
					case "static":
						if idx = initializer(true); idx >= 0 {
							f.kind = frameMethod
						}
					case "":
						if idx = initializer(false); idx >= 0 {
							f.kind = frameMethod
						}
					}
				}
			}
			if f.kind == frameClass {
				f.iface = isInterfaceHeader(header)
			}
			stack = append(stack, f)
			stackDecl = append(stackDecl, idx)
			// A method opened on this line owns it even if it also closes
			// on it ("void g() { bad(); }").
			if idx >= 0 && owner[currentLine] < 0 {
				owner[currentLine] = innermost()
			}
			headerStart, headerLine = i+1, currentLine
		case '}':
			if n := len(stack); n > 0 {
				if d := stackDecl[n-1]; d >= 0 && currentLine > decl[d].EndLine {
					decl[d].EndLine = currentLine
				}
				stack, stackDecl = stack[:n-1], stackDecl[:n-1]
			}
			headerStart, headerLine = i+1, currentLine
		case ';':
			// A field declaration with an initializer directly in a class
			// body: its lines belong to <clinit> or the constructors.
			if cls, ok := topIsClass(); ok {
				if static, isField := fieldInitializer(cleaned[headerStart:i], cls.iface); isField {
					if d := initializer(static); d >= 0 {
						start := headerLine
						if headerStart < i && cleaned[headerStart] == '\n' {
							start++
						}
						for l := start; l <= currentLine && l < len(owner); l++ {
							if owner[l] < 0 || l == currentLine {
								owner[l] = d
							}
						}
						if decl[d].StartLine > start || decl[d].StartLine == 0 {
							decl[d].StartLine = start
						}
						if currentLine > decl[d].EndLine {
							decl[d].EndLine = currentLine
						}
					}
				}
			}
			headerStart, headerLine = i+1, currentLine
		}
	}
	if owner[currentLine] < 0 {
		owner[currentLine] = innermost()
	}
	return owner, decl
}

func classifyHeader(header, enclosingClass string) frame {
	h := trimSpace(header)
	if h == "" {
		return frame{kind: frameOther} // static/instance initializer, or a bare nested block
	}
	if newExprRe.MatchString(h) {
		return frame{kind: frameAnon}
	}
	if m := classHeaderRe.FindStringSubmatch(h); m != nil {
		return frame{kind: frameClass, name: m[1]}
	}
	if m := methodHeaderRe.FindStringSubmatch(h); m != nil {
		name := m[1]
		if controlKeywords[name] {
			return frame{kind: frameOther}
		}
		if name == enclosingClass {
			name = "<init>"
		}
		if strings.HasSuffix(strings.TrimSpace(h[:len(h)-len(m[0])]), "->") || strings.HasSuffix(strings.TrimSpace(h[:len(h)-len(m[0])]), "new") {
			return frame{kind: frameOther}
		}
		return frame{kind: frameMethod, name: name, params: parseParams(m[2], typeVars(h))}
	}
	return frame{kind: frameOther}
}

var typeVarListRe = regexp.MustCompile(`(?:^|[\s>])<([^()]*?)>\s*[\w.\[\]<>?,\s]*\s+\w+\s*\(`)

// typeVars returns the method's own type parameters (the "<T, U extends
// X>" before its return type), so parameters of those types are "?".
func typeVars(header string) map[string]bool {
	out := map[string]bool{}
	m := typeVarListRe.FindStringSubmatch(header)
	if m == nil {
		return out
	}
	for _, part := range splitTopLevel(m[1]) {
		f := strings.Fields(part)
		if len(f) > 0 {
			out[f[0]] = true
		}
	}
	return out
}

// parseParams reduces "final Map.Entry<K,V> e, String... xs, int[] a" to
// ["Entry", "String[]", "int[]"]. Annotations and modifiers are dropped,
// generic arguments erased; a type variable of the method is "?". Class
// type variables are not known here either, so a bare single-uppercase-
// letter type is "?" too.
func parseParams(list string, methodTypeVars map[string]bool) []string {
	list = strings.TrimSpace(list)
	if list == "" {
		return nil
	}
	var out []string
	for _, p := range splitTopLevel(list) {
		p = eraseGenerics(p)
		var toks []string
		for _, t := range strings.Fields(strings.ReplaceAll(p, "...", " ... ")) {
			if strings.HasPrefix(t, "@") || t == "final" {
				continue
			}
			toks = append(toks, t)
		}
		if len(toks) < 2 {
			out = append(out, "?")
			continue
		}
		name := toks[len(toks)-1]
		typ := strings.Join(toks[:len(toks)-1], "")
		dims := strings.Count(name, "[]")
		if strings.HasSuffix(typ, "...") {
			typ, dims = strings.TrimSuffix(typ, "..."), dims+1
		}
		dims += strings.Count(typ, "[]")
		typ = strings.ReplaceAll(typ, "[]", "")
		if i := strings.LastIndexByte(typ, '.'); i >= 0 {
			typ = typ[i+1:]
		}
		if methodTypeVars[typ] || (len(typ) == 1 && typ[0] >= 'A' && typ[0] <= 'Z') {
			typ = "?"
		}
		out = append(out, typ+strings.Repeat("[]", dims))
	}
	return out
}

func splitTopLevel(s string) []string {
	var parts []string
	depth, start := 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, s[start:])
}

func eraseGenerics(s string) string {
	var b strings.Builder
	depth := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '<':
			depth++
		case c == '>':
			depth--
		case depth == 0:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && isSpace(s[start]) {
		start++
	}
	end := len(s)
	for end > start && isSpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

var wordInterface = regexp.MustCompile(`(^|\s)(@?interface)\s`)

func isInterfaceHeader(header string) bool {
	return wordInterface.MatchString(" " + header + " ")
}

var wordStatic = regexp.MustCompile(`(^|[\s>])static\s`)

// fieldInitializer reports whether a class-body statement (the text up
// to its ';') declares a field with an initializer — an '=' outside
// parentheses and brackets, with no parenthesis before it (a method
// declaration has one) — and whether that field is static (an
// interface's fields always are).
func fieldInitializer(stmt string, inInterface bool) (static, ok bool) {
	depth := 0
	for i := 0; i < len(stmt); i++ {
		switch stmt[i] {
		case '(', '[', '<':
			if stmt[i] == '(' && depth == 0 {
				// "@Annotation(x = 1) int f = 2;" keeps going; a method
				// header "int f(int a)" never has a top-level '=' after.
				if j := strings.LastIndex(stmt[:i], "@"); j < 0 || strings.ContainsAny(stmt[j:i], " \t\n") {
					return false, false
				}
			}
			depth++
		case ')', ']', '>':
			if depth > 0 {
				depth--
			}
		case '=':
			if depth == 0 && (i+1 >= len(stmt) || stmt[i+1] != '=') {
				head := stmt[:i]
				return inInterface || wordStatic.MatchString(" "+head), true
			}
		}
	}
	return false, false
}
