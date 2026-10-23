// FRAMEWORKS F2's own "oznaczone korzenie dla HTTP, servlet/filter/
// lifecycle" slice: a web.xml-declared servlet/filter/listener's own
// lifecycle method(s) become real roots under `-roots profile` —
// previously, a plain old servlet with no annotation at all (the
// scripts/prepare-servlet-demo.sh control app's own exact shape)
// contributed nothing to `-roots profile`'s entry-point set; only
// `-roots all` (every application method) happened to still cover it.
//
// Deliberately narrow: only the well-known container-callback method
// *names* for each of the three declared kinds become roots — never
// "every public method of a servlet/filter/listener class", which
// would repeat the plan's own explicit warning ("nie uznawać każdego
// publicznego helpera za endpoint") for a class whose other public
// methods are ordinary helpers the container never calls directly.
package profile

import (
	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// WebRootKind names which of the three web.xml declaration shapes
// produced a WebXMLRoot.
type WebRootKind string

const (
	WebRootServlet  WebRootKind = "servlet"
	WebRootFilter   WebRootKind = "filter"
	WebRootListener WebRootKind = "listener"
)

// WebXMLRoot is one container-callback method a web.xml declaration
// makes a real entry point.
type WebXMLRoot struct {
	FQCN, Method string
	Kind         WebRootKind
}

// servletLifecycleMethods/filterLifecycleMethods/
// listenerLifecycleMethods are javax/jakarta.servlet's own callback
// method names (overload-insensitive by name, the same matching
// discipline every root/vulnerable-method lookup in this codebase
// already uses) — not every public method a servlet/filter/listener
// class happens to declare.
var (
	servletLifecycleMethods = []string{
		"init", "destroy", "service",
		"doGet", "doPost", "doPut", "doDelete", "doHead", "doOptions", "doTrace",
	}
	filterLifecycleMethods   = []string{"init", "destroy", "doFilter"}
	listenerLifecycleMethods = []string{
		"contextInitialized", "contextDestroyed",
		"sessionCreated", "sessionDestroyed",
		"requestInitialized", "requestDestroyed",
		"attributeAdded", "attributeRemoved", "attributeReplaced",
	}
)

// WebXMLRoots finds, for each declared servlet/filter/listener class,
// which of its own lifecycle method names it actually declares —
// FindWebXMLRoots's own I/O-free inventory, independent of whether
// `-roots profile` or `-roots all` is active (an inventory fact, not a
// root-selection decision — see WebXMLAppRoots for the latter).
// servletClasses/filterClasses/listenerClasses are deduplicated FQCNs
// as found in every web.xml module read; a class named there but not
// found in the analyzed hierarchy (outside the app/dependency jars
// given) is honestly skipped, never fabricated.
func WebXMLRoots(h *hierarchy.Hierarchy, servletClasses, filterClasses, listenerClasses []string) []WebXMLRoot {
	var out []WebXMLRoot
	out = append(out, webRootsFor(h, servletClasses, servletLifecycleMethods, WebRootServlet)...)
	out = append(out, webRootsFor(h, filterClasses, filterLifecycleMethods, WebRootFilter)...)
	out = append(out, webRootsFor(h, listenerClasses, listenerLifecycleMethods, WebRootListener)...)
	return out
}

func webRootsFor(h *hierarchy.Hierarchy, classes, methodNames []string, kind WebRootKind) []WebXMLRoot {
	var out []WebXMLRoot
	seen := map[string]bool{}
	for _, fqcn := range classes {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		for _, name := range methodNames {
			for _, m := range ci.MethodsByName(name) {
				if m.IsAbstract() {
					continue
				}
				key := fqcn + "#" + m.Name + m.Descriptor
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, WebXMLRoot{FQCN: fqcn, Method: m.Name, Kind: kind})
			}
		}
	}
	return out
}

// WebXMLAppRoots is WebXMLRoots' own call-graph-node form — the actual
// root-selection decision `-roots profile` needs, mirroring
// Profile.overrideRoots' own shape (declaring class only, no subtype
// expansion: web.xml names the exact concrete class the container
// instantiates, unlike Roots.Overrides' own "fqcn or any subtype"
// rule for an abstract framework base class).
func WebXMLAppRoots(g *callgraph.Graph, h *hierarchy.Hierarchy, servletClasses, filterClasses, listenerClasses []string) []uint32 {
	var roots []uint32
	for _, r := range WebXMLRoots(h, servletClasses, filterClasses, listenerClasses) {
		ci, ok := h.Lookup(r.FQCN)
		if !ok {
			continue
		}
		for _, m := range ci.MethodsByName(r.Method) {
			if m.IsAbstract() {
				continue
			}
			if id, ok := g.ID(r.FQCN, m.Name, m.Descriptor); ok {
				roots = append(roots, id)
			}
		}
	}
	return roots
}
