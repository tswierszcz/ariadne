package javadiff

import "regexp"

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
var newExprRe = regexp.MustCompile(`(?:^|\W)new\s+\w+\s*(?:<[^>]*>)?\s*\([^()]*\)\s*$`)

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
	frameOther  frameKind = iota // a plain block: if/for/lambda body, static/instance initializer, anonymous class, ...
	frameClass                   // a class/interface/enum/record body
	frameMethod                  // a method or constructor body
)

type frame struct {
	kind frameKind
	name string
}

// FindMethodsForLines returns the name of the method/constructor
// enclosing each of the given 1-based line numbers in src (a single
// top-level — possibly with nested types — .java source file), as
// "<init>" for a constructor and the plain method name otherwise.
// Overloads are not distinguished (no descriptor), consistent with how
// ariadne matches methods everywhere else.
//
// A line outside any method (a field initializer, a class-body line, an
// import) contributes nothing, not an error: fix commits routinely touch
// such lines too, and the caller simply gets fewer methods out of it.
func FindMethodsForLines(src string, lines []int) []string {
	cleaned := stripCommentsAndStrings(src)
	lineOwner := computeLineOwners(cleaned)

	seen := make(map[string]bool)
	var out []string
	for _, l := range lines {
		if l < 1 || l >= len(lineOwner) {
			continue
		}
		name := lineOwner[l]
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

// computeLineOwners returns, for each 1-based line number, the name of
// the innermost enclosing method/constructor active at that line ("" if
// none).
func computeLineOwners(cleaned string) []string {
	totalLines := 1
	for i := 0; i < len(cleaned); i++ {
		if cleaned[i] == '\n' {
			totalLines++
		}
	}
	lineOwner := make([]string, totalLines+1)

	var stack []frame
	currentLine := 1
	headerStart := 0

	topMethodName := func() string {
		for i := len(stack) - 1; i >= 0; i-- {
			if stack[i].kind == frameMethod {
				return stack[i].name
			}
		}
		return ""
	}
	currentClassName := func() string {
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
			lineOwner[currentLine] = topMethodName()
			currentLine++
		case '{':
			header := cleaned[headerStart:i]
			stack = append(stack, classifyHeader(header, currentClassName()))
			headerStart = i + 1
		case '}':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			headerStart = i + 1
		case ';':
			headerStart = i + 1
		}
	}
	lineOwner[currentLine] = topMethodName()
	return lineOwner
}

func classifyHeader(header, enclosingClass string) frame {
	h := trimSpace(header)
	if h == "" {
		return frame{kind: frameOther} // static/instance initializer, or a bare nested block
	}
	if newExprRe.MatchString(h) {
		return frame{kind: frameOther} // anonymous class body, not a method
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
		return frame{kind: frameMethod, name: name}
	}
	return frame{kind: frameOther}
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
