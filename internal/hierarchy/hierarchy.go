// Package hierarchy builds the class hierarchy (superclasses, interfaces,
// transitive subtypes) and per-class metadata (methods, annotations,
// source jar/GAV) that both CHA expansion (internal/callgraph) and
// profile-based root detection (internal/reach) need.
package hierarchy

import (
	"sort"

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
	// InitRefs is every new/getstatic/putstatic reference found alongside
	// Calls (JVM_GRAPH G3) — each one requires some class's <clinit> to
	// run before this instruction executes.
	InitRefs []classfile.InitRef
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
	// Fields are this class's own declared fields (JVM_GRAPH G3 needs
	// these to resolve which class actually declares a field a
	// getstatic/putstatic refers to — the field equivalent of Methods).
	Fields       []classfile.Field
	Annotations  []string
	// SourceJar is the path of the jar this class was read from — the app
	// jar (or the in-memory jar PrepareAppInput packed from a classes
	// directory) for an app class, or the dependency jar's own path
	// otherwise. Never "" in practice: Build always passes a jar path for
	// the app too.
	SourceJar string
	// SourceEntry is the exact zip entry path this class was read from
	// inside SourceJar (JVM_GRAPH G4) — ordinarily the same as the class's
	// own path (e.g. "com/example/Foo.class"), but
	// "META-INF/versions/N/com/example/Foo.class" when a multi-release
	// variant was selected for the active -target-java.
	SourceEntry string
	// VariantVersion is 0 for an ordinary base entry, or the
	// multi-release version N SourceEntry was selected from (JVM_GRAPH
	// G4).
	VariantVersion int
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

// FieldByNameAndDesc returns the field declared directly on this class
// matching both name and descriptor exactly (JVMS 5.4.3.2 field
// resolution matches by name+descriptor, same as a method).
func (c *ClassInfo) FieldByNameAndDesc(name, desc string) (classfile.Field, bool) {
	for _, f := range c.Fields {
		if f.Name == name && f.Descriptor == desc {
			return f, true
		}
	}
	return classfile.Field{}, false
}

// MethodByNameAndDesc returns the method declared directly on this class
// matching both name and descriptor exactly — the exact-signature
// counterpart to MethodByName's overload-insensitive lookup, used where
// ariadne needs to tell distinct overloads apart (JVM_GRAPH G2).
func (c *ClassInfo) MethodByNameAndDesc(name, desc string) (Method, bool) {
	for _, m := range c.Methods {
		if m.Name == name && m.Descriptor == desc {
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

	// duplicates records every FQCN collision Build resolved (see
	// Duplicate's own doc comment) — structured evidence behind the
	// plain-string warnings Build also returns.
	duplicates []Duplicate
}

// NewForTesting builds a Hierarchy directly from already-constructed
// ClassInfo values, skipping jar I/O entirely. Production code always
// goes through Build; this exists for tests elsewhere (internal/callgraph,
// internal/reach) that need a specific class/method/descriptor shape —
// e.g. two overloads of the same name, so CHA's overload-sensitivity
// (JVM_GRAPH G2) can be tested precisely — without compiling a new
// fixture, which this repo's own test suite deliberately avoids
// depending on (go test needs no JDK).
func NewForTesting(classes []*ClassInfo) *Hierarchy {
	m := make(map[string]*ClassInfo, len(classes))
	for _, c := range classes {
		m[c.FQCN] = c
	}
	sorted := make([]string, 0, len(m))
	for fqcn := range m {
		sorted = append(sorted, fqcn)
	}
	sort.Strings(sorted)
	return &Hierarchy{
		classes:       m,
		direct:        buildDirectSubtypes(m, sorted),
		subtypesCache: make(map[string][]string),
		sorted:        sorted,
	}
}

// Duplicates returns every class-name collision Build resolved while
// assembling this Hierarchy, in the order they were found (app, then
// dependency jars in explicit input order). Empty when nothing
// collided. The caller must not modify the returned slice.
func (h *Hierarchy) Duplicates() []Duplicate {
	return h.duplicates
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
