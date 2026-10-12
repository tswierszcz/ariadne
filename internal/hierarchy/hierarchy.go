// Package hierarchy builds the class hierarchy (superclasses, interfaces,
// transitive subtypes) and per-class metadata (methods, annotations,
// source jar/GAV) that both CHA expansion (internal/callgraph) and
// profile-based root detection (internal/reach) need.
package hierarchy

import (
	"ariadne/internal/classfile"
	"ariadne/internal/input"
)

// ClassInfo is everything the rest of ariadne needs about one class.
type ClassInfo struct {
	FQCN        string
	SuperClass  string // "" only for java.lang.Object itself
	Interfaces  []string
	AccessFlags uint16
	Methods     []classfile.Method
	Annotations []string
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
func (c *ClassInfo) MethodByName(name string) (classfile.Method, bool) {
	for _, m := range c.Methods {
		if m.Name == name {
			return m, true
		}
	}
	return classfile.Method{}, false
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
