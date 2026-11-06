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
	c := &classifier{h: h, cache: map[string]int{}}
	sum := CandidateSummary{ByKind: map[string]int{}, UnknownBases: map[string]int{}}
	seen := map[Target]bool{}
	var out []Target
	add := func(t Target) {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
			sum.ByKind[t.Kind]++
		}
	}
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
		switch c.serializable(fqcn) {
		case serUnknown:
			sum.Unknown++
			sum.UnknownBases[c.exitClass(fqcn)]++
			continue
		case serNo:
			continue
		}
		if ci.AccessFlags&accEnum != 0 || ci.SuperClass == "java.lang.Enum" {
			sum.Enums++
			continue
		}
		if ci.SuperClass == "java.lang.Record" {
			sum.Records++
			continue
		}
		sum.Classes++
		if c.implements(fqcn, externalizableIface, map[string]bool{}) {
			sum.Externalizable++
			if m, ok := declared(ci, "<init>", "()V"); ok && m.AccessFlags&classfile.AccPublic != 0 {
				add(Target{FQCN: fqcn, Name: "<init>", Desc: "()V", Kind: "constructor"})
			}
			if decl, ok := h.ResolveExact(fqcn, "readExternal", "(Ljava/io/ObjectInput;)V"); ok {
				add(Target{FQCN: decl, Name: "readExternal", Desc: "(Ljava/io/ObjectInput;)V", Kind: "readExternal"})
			}
			continue
		}
		// The serializable chain: fqcn and its serializable superclasses
		// in the input; the first non-serializable superclass's no-arg
		// constructor creates the object.
		cur := ci
		for depth := 0; cur != nil && depth < 64; depth++ {
			for _, cb := range []struct{ name, desc string }{{"readObject", "(Ljava/io/ObjectInputStream;)V"}, {"readObjectNoData", "()V"}} {
				if m, ok := declared(cur, cb.name, cb.desc); ok && m.AccessFlags&accPrivate != 0 && m.AccessFlags&accStatic == 0 {
					add(Target{FQCN: cur.FQCN, Name: cb.name, Desc: cb.desc, Kind: cb.name})
				}
			}
			sup := cur.SuperClass
			if c.serializable(sup) != serYes {
				if sci, ok := h.Lookup(sup); ok {
					if _, ok := declared(sci, "<init>", "()V"); ok {
						add(Target{FQCN: sup, Name: "<init>", Desc: "()V", Kind: "constructor"})
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
		// readResolve: resolved like a virtual call with serialization's
		// access rules.
		for cur, depth := ci, 0; cur != nil && depth < 64; depth++ {
			if m, ok := declared(cur, "readResolve", "()Ljava/lang/Object;"); ok && m.AccessFlags&accStatic == 0 {
				private := m.AccessFlags&accPrivate != 0
				pkgPrivate := m.AccessFlags&(classfile.AccPublic|0x0004|accPrivate) == 0
				if cur.FQCN == fqcn || (!private && (!pkgPrivate || pkg(cur.FQCN) == pkg(fqcn))) {
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
	}
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
	return out, sum
}

// Edge is one "may run" edge from an entry site to a target.
type Edge struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	TargetFQCN, TargetName, TargetDesc string
}

// Edges connects every site of a model with modeled callbacks to every
// target. Sites of an entry-only model (XStream) get none.
func Edges(sites []Site, targets []Target) []Edge {
	var out []Edge
	for _, s := range sites {
		if s.Coverage != CallbacksModeled {
			continue
		}
		for _, t := range targets {
			out = append(out, Edge{CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, PC: s.PC,
				TargetFQCN: t.FQCN, TargetName: t.Name, TargetDesc: t.Desc})
		}
	}
	return out
}
