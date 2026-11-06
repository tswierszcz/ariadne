// Package deserialize is DESERIALIZATION Z0–Z1: a catalog of
// deserializer models (which API reads objects, which callbacks the
// format runs, which protections exist), and — for Java serialization —
// the entry sites in the analyzed bytecode and the callbacks they may
// run.
//
// What a deserialization entry is NOT: a proof that a gadget chain
// exists or that the CVE's code runs. The model separates five kinds of
// evidence (Evidence*): an API use, a candidate type, a callback that may
// run, a filter decision, and a chain candidate. Z1 produces the first
// three; filters are detected and reported as not evaluated (Z2); chain
// candidates are Z3. The existing manual XStream mappings keep their
// entry-point role — XStream is cataloged here as an entry API whose
// callbacks are not modeled.
package deserialize

import "ariadne/internal/models"

// EvidenceKind is one kind of deserialization evidence (Z0).
type EvidenceKind string

const (
	// EvidenceAPIUse: a call to a cataloged deserialization entry API.
	EvidenceAPIUse EvidenceKind = "api-use"
	// EvidenceTypeCandidate: a class the model says may be created at an
	// entry (here: every concrete serializable class of the input — the
	// stream decides which, and a cast after readObject does not bound
	// it: the callbacks have run by then).
	EvidenceTypeCandidate EvidenceKind = "type-candidate"
	// EvidenceCallback: a method the format calls on a created object
	// (exact signature and access rules, JLS/Java Object Serialization
	// Specification), connected to the entry by a "may run" edge.
	EvidenceCallback EvidenceKind = "callback"
	// EvidenceFilter: a filter or type check the entry is subject to.
	// Z1 only detects one (ObjectInputFilter set in the same method, or a
	// resolveClass-overriding stream); its decision is Z2.
	EvidenceFilter EvidenceKind = "filter"
	// EvidenceChainCandidate: a candidate gadget chain from an entry to
	// the advisory's code — Z3, never produced here.
	EvidenceChainCandidate EvidenceKind = "chain-candidate"
)

// Coverage of a model's callbacks.
const (
	CallbacksModeled = "modeled"    // entries get callback edges
	CallbacksEntry   = "entry-only" // entries are found; callbacks are not modeled
)

// CallbackSpec is one callback the format runs on a created object.
type CallbackSpec struct {
	Name, Desc string
	// Access: "private" (must be private and non-static: the format looks
	// it up on each class of the serializable chain), "inheritable"
	// (resolved like a virtual call, private only on the class itself,
	// package-private only within its package) or "public".
	Access string
	// Applies: "serializable" or "externalizable".
	Applies string
	Note    string
}

// Model is one deserializer (Z0's DeserializerModel).
type Model struct {
	ID string
	// Component: Maven group:artifact, "" for the JDK.
	Component string
	Versions  models.VersionRange
	// Entries are the exact call shapes that read an object.
	Entries   []models.Signature
	TypeRule  string
	Callbacks []CallbackSpec
	Coverage  string
	// Format selects the callback rules (ClassTargets); "" when the
	// callbacks are not modeled.
	Format string
	// Protections are the format's own type restrictions, by name.
	Protections []string
	Source      string
}

// Java serialization callbacks (Java Object Serialization Specification
// §3.4–3.7, §3.3 for readResolve and §3.6 for Externalizable).
var javaCallbacks = []CallbackSpec{
	{Name: "readObject", Desc: "(Ljava/io/ObjectInputStream;)V", Access: "private", Applies: "serializable", Note: "per class of the serializable chain, superclass first"},
	{Name: "readObjectNoData", Desc: "()V", Access: "private", Applies: "serializable", Note: "a class of the chain absent from the stream"},
	{Name: "readResolve", Desc: "()Ljava/lang/Object;", Access: "inheritable", Applies: "serializable", Note: "after the object is read"},
	{Name: "readExternal", Desc: "(Ljava/io/ObjectInput;)V", Access: "public", Applies: "externalizable", Note: "after the public no-arg constructor"},
}

// Catalog is the deserializer model catalog.
var Catalog = []Model{
	{
		ID:       "jdk:java-serialization",
		TypeRule: "any concrete Serializable/Externalizable class visible to the stream's class loader, unless a filter or resolveClass restricts it; enums are resolved by name (no callbacks), records through their canonical constructor (not modeled)",
		Entries: []models.Signature{
			{Owner: "java.io.ObjectInputStream", Name: "readObject", Desc: "()Ljava/lang/Object;"},
			{Owner: "java.io.ObjectInputStream", Name: "readUnshared", Desc: "()Ljava/lang/Object;"},
			{Owner: "java.io.ObjectInput", Name: "readObject", Desc: "()Ljava/lang/Object;"},
		},
		Callbacks:   javaCallbacks,
		Coverage:    CallbacksModeled,
		Format:      FormatJava,
		Protections: []string{"ObjectInputFilter (setObjectInputFilter, jdk.serialFilter)", "ObjectInputStream#resolveClass override"},
		Source:      "Java Object Serialization Specification (docs.oracle.com/en/java/javase/21/docs/specs/serialization), java.io.ObjectInputStream javadoc",
	},
	{
		ID:        "xstream:fromxml",
		Component: "com.thoughtworks.xstream:xstream",
		Versions:  models.VersionRange{Min: "1.4", Max: "1.4.21"},
		TypeRule:  "any class (XStream's reflection converters), restricted by its security framework: none before 1.4.7; any type by default in 1.4.7–1.4.17; from 1.4.18 a default allowlist of JDK basic types; allowTypes*/denyTypes*/addPermission/setupDefaultSecurity on the same instance, the last added permission deciding first",
		Entries: []models.Signature{
			{Owner: "com.thoughtworks.xstream.XStream", Name: "fromXML", Desc: "(Ljava/lang/String;)Ljava/lang/Object;"},
			{Owner: "com.thoughtworks.xstream.XStream", Name: "fromXML", Desc: "(Ljava/io/Reader;)Ljava/lang/Object;"},
			{Owner: "com.thoughtworks.xstream.XStream", Name: "fromXML", Desc: "(Ljava/io/InputStream;)Ljava/lang/Object;"},
			{Owner: "com.thoughtworks.xstream.XStream", Name: "fromXML", Desc: "(Ljava/net/URL;)Ljava/lang/Object;"},
			{Owner: "com.thoughtworks.xstream.XStream", Name: "fromXML", Desc: "(Ljava/io/File;)Ljava/lang/Object;"},
			{Owner: "com.thoughtworks.xstream.XStream", Name: "fromXML", Desc: "(Ljava/lang/String;Ljava/lang/Object;)Ljava/lang/Object;"},
			{Owner: "com.thoughtworks.xstream.XStream", Name: "fromXML", Desc: "(Ljava/io/Reader;Ljava/lang/Object;)Ljava/lang/Object;"},
		},
		Callbacks: []CallbackSpec{
			{Name: "readResolve", Desc: "()Ljava/lang/Object;", Access: "any", Applies: "any", Note: "declared in the class or a superclass, any access"},
			{Name: "readObject", Desc: "(Ljava/io/ObjectInputStream;)V", Access: "private", Applies: "serializable", Note: "SerializableConverter, per class of the serializable chain"},
			{Name: "readExternal", Desc: "(Ljava/io/ObjectInput;)V", Access: "public", Applies: "externalizable", Note: "ExternalizableConverter, after the public no-arg constructor"},
		},
		Coverage:    CallbacksModeled,
		Format:      FormatXStream,
		Protections: []string{"XStream security framework (allowTypes*/denyTypes*/addPermission, setupDefaultSecurity; default allowlist from 1.4.18)"},
		Source:      "x-stream.github.io/security.html and converters.html; XStream 1.4.x SecurityMapper, SerializationMembers; observed on the real JVM with 1.4.5 and 1.4.20 (scripts/prepare-xstream-demo.sh)",
	},
}

// lookupEntry returns the model whose entry is exactly (owner, name, desc).
func lookupEntry(owner, name, desc string) (Model, bool) {
	for _, m := range Catalog {
		for _, s := range m.Entries {
			if s.Owner == owner && s.Name == name && s.Desc == desc {
				return m, true
			}
		}
	}
	return Model{}, false
}
