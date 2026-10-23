// FRAMEWORKS F3's own first, narrowest slice: a proxy/target/
// interceptor *identity* inventory — which application classes
// implement `java.lang.reflect.InvocationHandler`, and which of their
// own fields is a plausible target (the real object `invoke` forwards
// to), nothing more.
//
// Deliberately stops short of F3's own larger goal ("powiązać handler
// i argument Method z właściwym wywołaniem interfejsu" — binding the
// handler's own `invoke(Object, Method, Object[])` argument back to
// the real interface call site): that needs REFLECTION R2's own
// `Method.invoke` receiver/target reconstruction, which does not
// exist yet (confirmed: REFLECTION is entirely UNASSESSED). Without
// it, there is no sound way to turn this identity fact into a real
// call-graph edge — doing so here would fabricate a dispatch this
// package cannot actually prove. CGLIB-generated proxy subclasses are
// a separate, structurally unreachable case: they never exist as
// `.class` files at all (the JVM generates their bytecode at runtime),
// so no static analysis of app+dependency jars can see one — not a
// gap this package's own scope could close even in principle.
package aop

import (
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

const invocationHandlerFQCN = "java.lang.reflect.InvocationHandler"

// TargetField is one of a ProxyHandler's own fields that could
// plausibly be the real object its `invoke` forwards to: a `final`
// instance field of a reference type (the idiom javac/Lombok and
// hand-written handlers alike produce — the handler's constructor
// assigns it once and `invoke` reads it, never reassigned). A field
// failing this shape (static, mutable, or primitive/array) is not
// reported — report-only candidates, never asserted as the one real
// target without tracing `invoke`'s own body, which this slice does
// not do.
type TargetField struct {
	Name, Type string
}

// ProxyHandler is one application class implementing
// `InvocationHandler`, with every one of its own fields shaped like a
// plausible proxy target.
type ProxyHandler struct {
	FQCN         string
	TargetFields []TargetField
}

// FindProxyHandlers finds every application class that is a
// (transitive) subtype of `InvocationHandler`, with its own candidate
// target field(s). A handler with no field matching TargetField's own
// shape is still reported, with TargetFields empty — the identity
// fact ("this class is a proxy handler") and the target-field guess
// are independently honest; one being absent must not hide the other.
func FindProxyHandlers(h *hierarchy.Hierarchy) []ProxyHandler {
	var out []ProxyHandler
	for _, fqcn := range h.Subtypes(invocationHandlerFQCN) {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp {
			continue
		}
		var fields []TargetField
		for _, f := range ci.Fields {
			if f.AccessFlags&classfile.AccStatic != 0 || f.AccessFlags&classfile.AccFinal == 0 {
				continue
			}
			typ, ok := classfile.FieldTypeFQCN(f.Descriptor)
			if !ok {
				continue
			}
			fields = append(fields, TargetField{Name: f.Name, Type: typ})
		}
		out = append(out, ProxyHandler{FQCN: fqcn, TargetFields: fields})
	}
	return out
}
