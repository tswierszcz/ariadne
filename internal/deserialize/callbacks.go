package deserialize

import (
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

const (
	accPrivate = 0x0002
	accStatic  = 0x0008
	accEnum    = 0x4000

	serializableIface   = "java.io.Serializable"
	externalizableIface = "java.io.Externalizable"
)

// jdkSerializable are JDK classes outside the input known to be
// Serializable (a class of the input whose superclass chain leaves the
// input at one of them is serializable). The list is incomplete by
// design; a chain that leaves the input at another JDK class is counted
// as unknown, never assumed either way.
var jdkSerializable = map[string]bool{
	"java.lang.Throwable": true, "java.lang.Exception": true, "java.lang.RuntimeException": true, "java.lang.Error": true,
	"java.io.IOException": true, "java.lang.IllegalArgumentException": true, "java.lang.IllegalStateException": true,
	"java.lang.Number": true, "java.lang.Enum": true,
	"java.util.HashMap": true, "java.util.LinkedHashMap": true, "java.util.TreeMap": true, "java.util.Hashtable": true,
	"java.util.Properties": true, "java.util.ArrayList": true, "java.util.LinkedList": true, "java.util.HashSet": true,
	"java.util.LinkedHashSet": true, "java.util.TreeSet": true, "java.util.Vector": true, "java.util.Stack": true,
	"java.util.EventObject": true, "java.util.Date": true, "java.util.concurrent.ConcurrentHashMap": true,
	"java.util.concurrent.atomic.AtomicInteger": true, "java.util.concurrent.atomic.AtomicLong": true,
	"java.util.concurrent.atomic.AtomicReference": true, "java.io.File": true, "java.util.Random": true,
	"java.security.Permission": true, "java.security.BasicPermission": true, "java.util.AbstractMap$SimpleEntry": true,
	"java.text.Format": true, "java.util.Locale": true, "java.net.URI": true, "java.net.URL": true,
	"javax.servlet.jsp.tagext.TagSupport": true, "javax.servlet.jsp.tagext.BodyTagSupport": true,
	"javax.servlet.GenericServlet": true, "javax.servlet.http.HttpServlet": true,
	"jakarta.servlet.GenericServlet": true, "jakarta.servlet.http.HttpServlet": true,
	"kotlin.jvm.internal.Lambda": true, "kotlin.jvm.internal.CallableReference": true,
}

// jdkNotSerializable are JDK classes known not to be Serializable.
var jdkNotSerializable = map[string]bool{
	"java.lang.Object": true, "java.lang.Record": true, "java.util.AbstractMap": true, "java.util.AbstractList": true,
	"java.util.AbstractCollection": true, "java.util.AbstractSet": true, "java.lang.Thread": true,
	"java.io.InputStream": true, "java.io.OutputStream": true, "java.io.Reader": true, "java.io.Writer": true,
	"java.lang.ClassLoader": true, "java.util.TimerTask": true,
	"java.lang.ThreadLocal": true, "java.lang.InheritableThreadLocal": true, "java.lang.ref.WeakReference": true,
	"java.lang.ref.SoftReference": true, "java.lang.ref.PhantomReference": true, "java.lang.ref.Reference": true,
	"java.io.FilterReader": true, "java.io.FilterWriter": true, "java.io.FilterInputStream": true,
	"java.io.FilterOutputStream": true, "java.io.ByteArrayOutputStream": true, "java.io.ByteArrayInputStream": true,
	"java.io.ObjectInputStream": true, "java.io.ObjectOutputStream": true, "java.lang.invoke.MutableCallSite": true,
	"java.lang.ClassValue": true, "java.beans.PropertyEditorSupport": true,
	"org.xml.sax.helpers.DefaultHandler": true, "org.xml.sax.helpers.XMLFilterImpl": true, "org.xml.sax.HandlerBase": true,
}

// tri-state serializability.
const (
	serNo = iota
	serYes
	serUnknown
)

type classifier struct {
	h     *hierarchy.Hierarchy
	cache map[string]int
	ext   map[string]bool
}

func (c *classifier) implements(fqcn, iface string, seen map[string]bool) bool {
	if fqcn == iface {
		return true
	}
	if seen[fqcn] {
		return false
	}
	seen[fqcn] = true
	ci, ok := c.h.Lookup(fqcn)
	if !ok {
		return false
	}
	for _, i := range ci.Interfaces {
		if c.implements(i, iface, seen) {
			return true
		}
	}
	return ci.SuperClass != "" && c.implements(ci.SuperClass, iface, seen)
}

// serializable classifies fqcn (a class of the input).
func (c *classifier) serializable(fqcn string) int {
	if v, ok := c.cache[fqcn]; ok {
		return v
	}
	c.cache[fqcn] = serNo // cycle guard
	v := serNo
	ci, ok := c.h.Lookup(fqcn)
	switch {
	case !ok && jdkSerializable[fqcn]:
		v = serYes
	case !ok && (strings.HasPrefix(fqcn, "java.") || strings.HasPrefix(fqcn, "javax.")) && (strings.HasSuffix(fqcn, "Exception") || strings.HasSuffix(fqcn, "Error")):
		// A JDK exception or error: a Throwable, which is Serializable.
		v = serYes
	case !ok && (jdkNotSerializable[fqcn] || fqcn == ""):
		v = serNo
	case !ok:
		v = serUnknown
	case c.implements(fqcn, serializableIface, map[string]bool{}) || c.implements(fqcn, externalizableIface, map[string]bool{}):
		v = serYes
	default:
		// Its own interfaces (in or out of the input) do not name
		// Serializable; the superclass decides.
		v = c.serializable(ci.SuperClass)
		for _, i := range ci.Interfaces {
			if _, in := c.h.Lookup(i); !in && v == serNo && !strings.HasPrefix(i, "java.") {
				v = serUnknown // an interface outside the input could extend Serializable
			}
		}
	}
	c.cache[fqcn] = v
	return v
}

// exitClass is the first class of fqcn's superclass chain outside the
// input.
func (c *classifier) exitClass(fqcn string) string {
	for i := 0; i < 64; i++ {
		ci, ok := c.h.Lookup(fqcn)
		if !ok {
			return fqcn
		}
		fqcn = ci.SuperClass
	}
	return fqcn
}

// Target is one method the format may run at an entry.
type Target struct {
	FQCN, Name, Desc string
	Kind             string // callback name, or "constructor"
}

// CandidateSummary counts what the candidate set was built from.
type CandidateSummary struct {
	// Concrete serializable classes of the input (type candidates) and,
	// of those, Externalizable ones.
	Classes, Externalizable int
	// Excluded: enums (resolved by name, no callbacks) and records
	// (canonical constructor, not modeled).
	Enums, Records int
	// Unknown: classes whose chain leaves the input at a JDK class not in
	// the built-in table — not counted as candidates.
	Unknown int
	// UnknownBases counts, per JDK class where such a chain leaves the
	// input, the unknown classes (to extend the table with evidence).
	UnknownBases map[string]int
	// Callbacks by kind.
	ByKind map[string]int
}

// Candidates computes every method Java serialization may run at an
// entry, given the input's classes. A callback must match the
// specification exactly: a readObject with another descriptor, or not
// private, or static, is not one.
func Candidates(h *hierarchy.Hierarchy) ([]Target, CandidateSummary) {
	byClass, sum := ClassTargets(h, FormatJava)
	return union(byClass), sum
}

// Formats whose callbacks are modeled.
const (
	FormatJava    = "java"
	FormatXStream = "xstream"
)

// ClassTargets returns, per concrete class of the input the format may
// create, the methods it runs on it (a target may serve several
// classes: a superclass's readObject). Java serialization (FormatJava)
// creates only serializable classes; XStream (FormatXStream, 1.4.x
// reflection converters) creates any class, calls readResolve declared
// in its class or any superclass (any access), the private readObject
// of each serializable class of the chain (SerializableConverter) and,
// for Externalizable, the public no-arg constructor and readExternal; it
// bypasses other constructors (its default reflection provider
// allocates without one), so it gets no constructor targets.
func ClassTargets(h *hierarchy.Hierarchy, format string) (map[string][]Target, CandidateSummary) {
	c := &classifier{h: h, cache: map[string]int{}}
	sum := CandidateSummary{ByKind: map[string]int{}, UnknownBases: map[string]int{}}
	out := map[string][]Target{}
	counted := map[Target]bool{}
	declared := func(ci *hierarchy.ClassInfo, name, desc string) (*hierarchy.Method, bool) {
		for i := range ci.Methods {
			m := &ci.Methods[i]
			if m.Name == name && m.Descriptor == desc {
				return m, true
			}
		}
		return nil, false
	}
	pkg := func(fqcn string) string {
		if i := strings.LastIndexByte(fqcn, '.'); i >= 0 {
			return fqcn[:i]
		}
		return ""
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) != 0 {
			continue
		}
		var ts []Target
		seen := map[Target]bool{}
		add := func(t Target) {
			if !seen[t] {
				seen[t] = true
				ts = append(ts, t)
				if !counted[t] {
					counted[t] = true
					sum.ByKind[t.Kind]++
				}
			}
		}
		state := c.serializable(fqcn)
		if format == FormatJava {
			switch state {
			case serUnknown:
				sum.Unknown++
				sum.UnknownBases[c.exitClass(fqcn)]++
				continue
			case serNo:
				continue
			}
		}
		if ci.AccessFlags&accEnum != 0 || ci.SuperClass == "java.lang.Enum" {
			if format == FormatJava || state == serYes {
				sum.Enums++
			}
			continue
		}
		if ci.SuperClass == "java.lang.Record" {
			sum.Records++
			continue
		}
		if format == FormatXStream && state == serUnknown {
			sum.Unknown++
			sum.UnknownBases[c.exitClass(fqcn)]++
		}
		sum.Classes++
		if state == serYes && c.implements(fqcn, externalizableIface, map[string]bool{}) {
			sum.Externalizable++
			if m, ok := declared(ci, "<init>", "()V"); ok && m.AccessFlags&classfile.AccPublic != 0 {
				add(Target{FQCN: fqcn, Name: "<init>", Desc: "()V", Kind: "constructor"})
			}
			if decl, ok := h.ResolveExact(fqcn, "readExternal", "(Ljava/io/ObjectInput;)V"); ok {
				add(Target{FQCN: decl, Name: "readExternal", Desc: "(Ljava/io/ObjectInput;)V", Kind: "readExternal"})
			}
			out[fqcn] = ts
			continue
		}
		if state == serYes {
			// The serializable chain: fqcn and its serializable
			// superclasses in the input; for Java serialization the first
			// non-serializable superclass's no-arg constructor creates
			// the object.
			cur := ci
			for depth := 0; cur != nil && depth < 64; depth++ {
				for _, cb := range []struct{ name, desc string }{{"readObject", "(Ljava/io/ObjectInputStream;)V"}, {"readObjectNoData", "()V"}} {
					if format == FormatXStream && cb.name == "readObjectNoData" {
						continue
					}
					if m, ok := declared(cur, cb.name, cb.desc); ok && m.AccessFlags&accPrivate != 0 && m.AccessFlags&accStatic == 0 {
						add(Target{FQCN: cur.FQCN, Name: cb.name, Desc: cb.desc, Kind: cb.name})
					}
				}
				sup := cur.SuperClass
				if c.serializable(sup) != serYes {
					if format == FormatJava {
						if sci, ok := h.Lookup(sup); ok {
							if _, ok := declared(sci, "<init>", "()V"); ok {
								add(Target{FQCN: sup, Name: "<init>", Desc: "()V", Kind: "constructor"})
							}
						}
					}
					break
				}
				next, ok := h.Lookup(sup)
				if !ok {
					break
				}
				cur = next
			}
		}
		// readResolve: Java serialization resolves it like a virtual
		// call with serialization's access rules; XStream takes the first
		// one declared in the class or a superclass, whatever its access.
		for cur, depth := ci, 0; cur != nil && depth < 64; depth++ {
			if m, ok := declared(cur, "readResolve", "()Ljava/lang/Object;"); ok && m.AccessFlags&accStatic == 0 {
				private := m.AccessFlags&accPrivate != 0
				pkgPrivate := m.AccessFlags&(classfile.AccPublic|0x0004|accPrivate) == 0
				if format == FormatXStream || cur.FQCN == fqcn || (!private && (!pkgPrivate || pkg(cur.FQCN) == pkg(fqcn))) {
					add(Target{FQCN: cur.FQCN, Name: "readResolve", Desc: "()Ljava/lang/Object;", Kind: "readResolve"})
				}
				break
			}
			next, ok := h.Lookup(cur.SuperClass)
			if !ok {
				break
			}
			cur = next
		}
		if len(ts) > 0 {
			out[fqcn] = ts
		}
	}
	return out, sum
}

// union is every distinct target of byClass, sorted.
func union(byClass map[string][]Target) []Target {
	seen := map[Target]bool{}
	var out []Target
	for _, ts := range byClass {
		for _, t := range ts {
			if !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	sortTargets(out)
	return out
}

func sortTargets(out []Target) {
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.FQCN != b.FQCN {
			return a.FQCN < b.FQCN
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Desc < b.Desc
	})
}

// Edge is one "may run" edge from an entry site to a target.
type Edge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetName, TargetDesc string
}

// Connect returns the edges of every site whose model has modeled
// callbacks: to the targets of every class the format may create there
// and the site's evaluated policy (if any) does not reject. It records
// on each site how many classes were candidates and how many were
// rejected.
func Connect(sites []Site, byFormat map[string]map[string][]Target) []Edge {
	var out []Edge
	for i := range sites {
		s := &sites[i]
		byClass := byFormat[s.Format]
		if s.Format == "" || byClass == nil {
			continue
		}
		classes := make([]string, 0, len(byClass))
		for c := range byClass {
			classes = append(classes, c)
		}
		sort.Strings(classes)
		var ts []Target
		seen := map[Target]bool{}
		for _, c := range classes {
			s.Candidates++
			if s.allow != nil && !s.allow(c) {
				s.Rejected++
				if len(s.RejectedSample) < 5 {
					s.RejectedSample = append(s.RejectedSample, c)
				}
				continue
			}
			for _, t := range byClass[c] {
				if !seen[t] {
					seen[t] = true
					ts = append(ts, t)
				}
			}
		}
		sortTargets(ts)
		for _, t := range ts {
			out = append(out, Edge{CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, PC: s.PC,
				TargetFQCN: t.FQCN, TargetName: t.Name, TargetDesc: t.Desc})
		}
	}
	return out
}

// Edges connects every Java serialization site to every target, with no
// policy (Z1 behaviour; kept for callers that have no per-class targets).
func Edges(sites []Site, targets []Target) []Edge {
	var out []Edge
	for _, s := range sites {
		if s.Format != FormatJava {
			continue
		}
		for _, t := range targets {
			out = append(out, Edge{CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, PC: s.PC,
				TargetFQCN: t.FQCN, TargetName: t.Name, TargetDesc: t.Desc})
		}
	}
	return out
}
