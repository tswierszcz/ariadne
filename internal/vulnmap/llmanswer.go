package vulnmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const maxLLMMethods = 10

var (
	llmClassRe  = regexp.MustCompile(`^[A-Za-z_$][\w$]*(\.[A-Za-z_$][\w$]*)+$`)
	llmMethodRe = regexp.MustCompile(`^([A-Za-z_$][\w$]*|<init>|<clinit>)$`)
)

type llmMethod struct {
	Class  string `json:"class"`
	Method string `json:"method"`
	Role   string `json:"role"`
	Reason string `json:"reason,omitempty"`
}

type llmAnswer struct {
	Methods []llmMethod `json:"methods"`
	Note    string      `json:"note,omitempty"`
}

// parseLLMAnswer extracts the first JSON object from content (tolerating
// code fences and prose around it) and validates it against the schema the
// system prompt asks for. The returned error text is sent back to the model
// verbatim on the corrective retry, so it names the exact violation.
func parseLLMAnswer(content, finishReason string) (llmAnswer, error) {
	if finishReason == "length" {
		return llmAnswer{}, fmt.Errorf("the reply was cut off by the token limit (finish_reason \"length\")")
	}

	start := strings.IndexByte(content, '{')
	if start < 0 {
		return llmAnswer{}, fmt.Errorf("the reply contains no JSON object")
	}
	// A Decoder reads exactly one value, so closing fences or trailing
	// prose after the object are ignored.
	dec := json.NewDecoder(strings.NewReader(content[start:]))
	var raw struct {
		Methods *[]llmMethod `json:"methods"`
		Note    string       `json:"note"`
	}
	if err := dec.Decode(&raw); err != nil {
		return llmAnswer{}, fmt.Errorf("the reply is not valid JSON (%v)", err)
	}
	if raw.Methods == nil {
		return llmAnswer{}, fmt.Errorf(`the JSON object has no "methods" array`)
	}
	methods := *raw.Methods
	if len(methods) > maxLLMMethods {
		return llmAnswer{}, fmt.Errorf("methods has %d elements, the maximum is %d", len(methods), maxLLMMethods)
	}
	for i, m := range methods {
		switch {
		case !llmClassRe.MatchString(m.Class):
			return llmAnswer{}, fmt.Errorf("methods[%d].class %q is not a fully qualified class name", i, m.Class)
		case !llmMethodRe.MatchString(m.Method):
			return llmAnswer{}, fmt.Errorf("methods[%d].method %q is not a bare method name", i, m.Method)
		case m.Role != "vulnerable" && m.Role != "entry_point":
			return llmAnswer{}, fmt.Errorf(`methods[%d].role %q is neither "vulnerable" nor "entry_point"`, i, m.Role)
		}
	}
	return llmAnswer{Methods: methods, Note: raw.Note}, nil
}

// groundMethods keeps only the methods exist reports as present in the
// analyzed bytecode, returned as sorted, deduplicated "Class#method"
// strings, along with how many answers were dropped. A nested class written
// with dots ("a.Outer.Inner") is retried as "a.Outer$Inner". exists == nil
// disables the check.
func groundMethods(ans llmAnswer, exists func(fqcn, method string) bool) (methods []string, dropped int) {
	seen := map[string]bool{}
	for _, m := range ans.Methods {
		class, ok := m.Class, true
		if exists != nil {
			class, ok = findClass(m.Class, m.Method, exists)
		}
		if !ok {
			dropped++
			continue
		}
		spec := class + "#" + m.Method
		if !seen[spec] {
			seen[spec] = true
			methods = append(methods, spec)
		}
	}
	sort.Strings(methods)
	return methods, dropped
}

// findClass tries class as written, then with its trailing dot-separated
// segments progressively rewritten as nested-class "$" separators.
func findClass(class, method string, exists func(fqcn, method string) bool) (string, bool) {
	cand := class
	for {
		if exists(cand, method) {
			return cand, true
		}
		i := strings.LastIndexByte(cand, '.')
		if i < 0 {
			return "", false
		}
		// Rewriting the last '.' that is still a '.' turns
		// a.b.Outer.Inner into a.b.Outer$Inner, then a.b$Outer$Inner.
		cand = cand[:i] + "$" + cand[i+1:]
		if !strings.Contains(cand, ".") {
			return "", false
		}
	}
}

// mustJSON is for building cache records and test fixtures from values that
// cannot fail to marshal.
func mustJSON(v any) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return bytes.TrimRight(b.Bytes(), "\n")
}
