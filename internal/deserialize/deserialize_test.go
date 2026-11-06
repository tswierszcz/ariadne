package deserialize

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

const (
	pub    = classfile.AccPublic
	priv   = accPrivate
	abstr  = classfile.AccAbstract
	static = accStatic
)

func meth(name, desc string, flags uint16, calls ...classfile.Call) hierarchy.Method {
	return hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc, AccessFlags: flags}, Calls: calls}
}

func class(fqcn, super string, flags uint16, ifaces []string, ms ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: super, AccessFlags: flags, Interfaces: ifaces, Methods: ms}
}

var ser = []string{"java.io.Serializable"}

const readObj = "(Ljava/io/ObjectInputStream;)V"

func targetSet(ts []Target) []string {
	var out []string
	for _, t := range ts {
		out = append(out, t.Kind+" "+t.FQCN+"#"+t.Name+t.Desc)
	}
	sort.Strings(out)
	return out
}

// TestCandidates: callbacks follow the specification exactly — private
// non-static readObject per class of the serializable chain, the first
// non-serializable superclass's no-arg constructor, readResolve resolved
// with serialization's access rules, Externalizable's public constructor
// and readExternal — and nothing else (wrong descriptor, public or static
// readObject, enums, records, abstract classes on their own).
func TestCandidates(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		class("p.NonSer", "java.lang.Object", pub, nil, meth("<init>", "()V", pub)),
		class("p.Base", "p.NonSer", pub|abstr, ser,
			meth("readObject", readObj, priv),
			meth("readResolve", "()Ljava/lang/Object;", 0)), // package-private
		class("p.Sub", "p.Base", pub, nil,
			meth("readObject", readObj, priv),
			meth("readObjectNoData", "()V", priv)),
		class("q.Other", "p.Base", pub, nil), // other package: no p.Base#readResolve
		class("p.WrongDesc", "java.lang.Object", pub, ser,
			meth("readObject", "(Ljava/io/ObjectInput;)V", priv),
			meth("readResolve", "()Ljava/lang/Object;", priv)),
		class("p.PublicRO", "java.lang.Object", pub, ser, meth("readObject", readObj, pub)),
		class("p.StaticRO", "java.lang.Object", pub, ser, meth("readObject", readObj, priv|static)),
		class("p.Ext", "java.lang.Object", pub, []string{"java.io.Externalizable"},
			meth("<init>", "()V", pub), meth("readExternal", "(Ljava/io/ObjectInput;)V", pub)),
		class("p.E", "java.lang.Enum", pub|accEnum, nil, meth("readObject", readObj, priv)),
		class("p.R", "java.lang.Record", pub, ser, meth("readObject", readObj, priv)),
		class("p.Ex", "java.lang.IllegalStateException", pub, nil, meth("readObject", readObj, priv)),
		class("p.Odd", "java.awt.Point2", pub, nil, meth("readObject", readObj, priv)),
		class("p.Plain", "java.lang.Object", pub, nil, meth("readObject", readObj, priv)),
	})
	ts, sum := Candidates(h)
	want := []string{
		"constructor p.Ext#<init>()V",
		"constructor p.NonSer#<init>()V",
		"readExternal p.Ext#readExternal(Ljava/io/ObjectInput;)V",
		"readObject p.Base#readObject(Ljava/io/ObjectInputStream;)V",
		"readObject p.Ex#readObject(Ljava/io/ObjectInputStream;)V",
		"readObject p.Sub#readObject(Ljava/io/ObjectInputStream;)V",
		"readObjectNoData p.Sub#readObjectNoData()V",
		"readResolve p.Base#readResolve()Ljava/lang/Object;",
		"readResolve p.WrongDesc#readResolve()Ljava/lang/Object;",
	}
	if got := targetSet(ts); !reflect.DeepEqual(got, want) {
		t.Errorf("targets:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if sum.Enums != 1 || sum.Records != 1 || sum.Unknown != 1 || sum.UnknownBases["java.awt.Point2"] != 1 || sum.Externalizable != 1 {
		t.Errorf("summary %+v", sum)
	}
	// Sub, Other, WrongDesc, PublicRO, StaticRO, Ext, Ex
	if sum.Classes != 7 {
		t.Errorf("classes = %d", sum.Classes)
	}
}

// TestSites: entries are found on ObjectInputStream, ObjectInput and a
// subclass; a filter in the same method or a resolveClass override is
// detected (not evaluated); XStream is an entry-only model; a method
// that only shares the name is not an entry.
func TestSites(t *testing.T) {
	read := func(owner string, pc int) classfile.Call {
		return classfile.Call{Kind: classfile.CallVirtual, Owner: owner, Name: "readObject", Desc: "()Ljava/lang/Object;", PC: pc}
	}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		class("app.Safe", "java.io.ObjectInputStream", pub, nil,
			meth("resolveClass", "(Ljava/io/ObjectStreamClass;)Ljava/lang/Class;", 0x0004)),
		class("app.Plain", "java.io.ObjectInputStream", pub, nil),
		class("app.Main", "java.lang.Object", pub, nil,
			meth("a", "()V", pub, read("java.io.ObjectInputStream", 1)),
			meth("b", "()V", pub, read("app.Safe", 2)),
			meth("c", "()V", pub, read("app.Plain", 3)),
			meth("d", "()V", pub, classfile.Call{Kind: classfile.CallVirtual, Owner: "java.io.ObjectInputStream", Name: "setObjectInputFilter", Desc: "(Ljava/io/ObjectInputFilter;)V", PC: 4}, read("java.io.ObjectInputStream", 5)),
			meth("e", "()V", pub, classfile.Call{Kind: classfile.CallInterface, Owner: "java.io.ObjectInput", Name: "readObject", Desc: "()Ljava/lang/Object;", PC: 6}),
			meth("f", "()V", pub, classfile.Call{Kind: classfile.CallVirtual, Owner: "com.thoughtworks.xstream.XStream", Name: "fromXML", Desc: "(Ljava/lang/String;)Ljava/lang/Object;", PC: 7}),
			meth("g", "()V", pub, classfile.Call{Kind: classfile.CallVirtual, Owner: "app.Codec", Name: "readObject", Desc: "()Ljava/lang/Object;", PC: 8}),
			meth("h", "()V", pub, classfile.Call{Kind: classfile.CallVirtual, Owner: "java.io.ObjectInputStream", Name: "readObject", Desc: "(I)Ljava/lang/Object;", PC: 9})),
	})
	got := map[string]Site{}
	for _, s := range FindSites(h) {
		got[s.CallerName] = s
	}
	if len(got) != 6 {
		t.Fatalf("sites %v", got)
	}
	for name, want := range map[string]string{"a": FilterNoneSeen, "b": FilterNotEvaluated, "c": FilterNoneSeen, "d": FilterNotEvaluated, "e": FilterNoneSeen} {
		if s := got[name]; s.Model != "jdk:java-serialization" || s.Filter != want {
			t.Errorf("%s = %+v, want filter %s", name, s, want)
		}
	}
	if s := got["f"]; s.Model != "xstream:fromxml" || s.Coverage != CallbacksEntry {
		t.Errorf("xstream = %+v", s)
	}
	edges := Edges(FindSites(h), []Target{{FQCN: "p.X", Name: "readObject", Desc: readObj, Kind: "readObject"}})
	if len(edges) != 5 {
		t.Errorf("edges = %d, want one per Java serialization site", len(edges))
	}
	for _, e := range edges {
		if e.CallerName == "f" {
			t.Errorf("entry-only XStream site got an edge")
		}
	}
}

func TestCatalogEntriesExact(t *testing.T) {
	for _, m := range Catalog {
		if m.ID == "" || m.Source == "" || len(m.Entries) == 0 || (m.Coverage != CallbacksModeled && m.Coverage != CallbacksEntry) {
			t.Errorf("model %+v", m)
		}
		for _, s := range m.Entries {
			if s.Owner == "" || s.Name == "" || !strings.HasPrefix(s.Desc, "(") {
				t.Errorf("%s: entry %+v is not an exact call shape", m.ID, s)
			}
		}
	}
}
