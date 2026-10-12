// Package hierarchy builds the class hierarchy (superclasses, interfaces,
// transitive subtypes) and per-class metadata (methods, annotations,
// source jar/GAV) that both CHA expansion (internal/callgraph) and
// profile-based root detection (internal/reach) need.
package hierarchy

import (
	"ariadne/internal/classfile"
	"ariadne/internal/input"
)

// Method is one method's static metadata plus the call-site edges found
// in its bytecode, extracted once while the class's constant pool and
// bootstrap methods were still in scope (classfile.ClassFile does not
// keep them, by design, so this is the only time ExtractCalls can run).
type Method struct {
	classfile.Method
	// Calls is nil for an abstract/native method (no Code), and otherwise
	// every call-site edge ExtractCalls found.
	Calls []classfile.Call
	// UnresolvedDynamic counts this method's invokedynamic sites whose
	// bootstrap wasn't LambdaMetafactory/altMetafactory (e.g. a
	// StringConcatFactory string concatenation), and so produced no edge.
	UnresolvedDynamic int
}

// ClassInfo is everything the rest of ariadne needs about one class.
type ClassInfo struct {
	FQCN        string
	SuperClass  string // "" only for java.lang.Object itself
	Interfaces  []string
	AccessFlags uint16
	// MajorVersion is the class file's own bytecode version (unrelated
	// to Hierarchy's Go code); the report surfaces the highest one seen
	// across the whole run.
	MajorVersion uint16
	Methods      []Method
	Annotations  []string
	// SourceJar is the jar this class was read from ("" for an app class
	// read from a classes directory/app-classes.jar rather than a
	// dependency jar).
	SourceJar string
	// GAV is the resolved Maven coordinate of SourceJar, or nil if it
	// came from the app, or from a dependency jar ResolveDepJars could
	// not identify.
	GAV *input.GAV
	// IsApp is true for a class that is part of the application being
	// analyzed, as opposed to one of its dependencies.
	IsApp bool
}

func (c *ClassInfo) IsInterface() bool { return c.AccessFlags&classfile.AccInterface != 0 }

// MethodByName returns the first method named name declared directly on
// this class, regardless of descriptor (ariadne matches overload-
// insensitively throughout, per the plan).
func (c *ClassInfo) MethodByName(name string) (Method, bool) {
	for _, m := range c.Methods {
		if m.Name == name {
			return m, true
		}
	}
	return Method{}, false
}

// MethodsByName returns every method named name declared directly on this
// class — there can be more than one, since Java allows overloads and
// ariadne matches by name only. Used where a caller needs every overload
// rather than just "does this class declare a method named X" (CHA
// expansion, which does not try to pick one overload over another).
func (c *ClassInfo) MethodsByName(name string) []Method {
	var out []Method
	for _, m := range c.Methods {
		if m.Name == name {
			out = append(out, m)
		}
	}
	return out
}

// Hierarchy is the whole class universe ariadne read: every app class and
// every dependency class, indexed by FQCN.
type Hierarchy struct {
	classes map[string]*ClassInfo
	// direct maps a supertype FQCN to the FQCNs that directly extend or
	// implement it. A class not in this universe at all (e.g. a JDK type
	// we never read, since we only read the app + its dependency jars)
	// simply has no entry and resolves to zero subtypes.
	direct map[string][]string

	subtypesCache map[string][]string // memoized transitive subtypes, see Subtypes

	// sorted is every FQCN in classes, sorted once in Build. Anything
	// that needs a deterministic iteration order over the whole
	// hierarchy (CHA edge insertion order in internal/callgraph, root
	// order in internal/reach, and buildDirectSubtypes below, all of
	// which affect which of several equally-short paths a BFS ends up
	// recording) must iterate this slice instead of ranging over the
	// classes map directly, since Go's map iteration order is randomized
	// per run and would otherwise make results.json's reported paths
	// nondeterministic between two runs of the very same analysis.
	sorted []string
}

// SortedFQCNs returns every class's FQCN, sorted, for callers that need a
// deterministic iteration order over the whole hierarchy (see the
// `sorted` field comment above for why this matters). The caller must
// not modify the returned slice.
func (h *Hierarchy) SortedFQCNs() []string {
	return h.sorted
}

// Lookup returns the ClassInfo for an FQCN, or (nil, false) if it's not
// part of this universe (typically a JDK class, since ariadne only reads
// the app's own classes and its declared dependency jars).
func (h *Hierarchy) Lookup(fqcn string) (*ClassInfo, bool) {
	c, ok := h.classes[fqcn]
	return c, ok
}

// All returns every class in the hierarchy. The caller must not modify
// the returned map.
func (h *Hierarchy) All() map[string]*ClassInfo {
	return h.classes
}

// Len returns the number of classes in the hierarchy.
func (h *Hierarchy) Len() int { return len(h.classes) }
