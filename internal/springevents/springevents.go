// Package springevents links the calls that publish a Spring application
// event to the @EventListener methods Spring runs for it (FRAMEWORKS F4,
// slice 1: model spring:events).
//
// ApplicationEventPublisher#publishEvent goes through the context's
// multicaster (SimpleApplicationEventMulticaster#multicastEvent →
// invokeListener) to every listener whose event type matches. A listener
// written as a class implementing ApplicationListener is called through
// that interface, and CHA already reaches it. A listener written as an
// annotated method is not: Spring wraps it in an
// ApplicationListenerMethodAdapter, whose doInvoke calls the method
// through Method#invoke, with nothing a reflection model can resolve. A
// real-JVM oracle (testdata/events) measured every such method body
// unreached from main: an @EventListener on the published type, on a
// supertype of it, on a @Bean-registered class, with classes= and no
// parameter, and a @TransactionalEventListener. Under -roots profile an
// @EventListener is a root of the spring profile by itself, but a
// @TransactionalEventListener (a meta-annotation) is not, and no path
// showed which code publishes the event.
//
// The model adds an edge (basis model) from a publishing call site to each
// listener method whose event type may match the published object:
//
//   - a publishing call site is a call of publishEvent(Object) or
//     publishEvent(ApplicationEvent) on ApplicationEventPublisher or a
//     subtype (ApplicationContext, …), anywhere in the input (the
//     context itself publishes ContextRefreshedEvent, ContextClosedEvent);
//     a publisher implementation forwarding its own parameter (a child
//     context to its parent) is not one — the original call is;
//   - the published type comes from the argument's local value (DATA_FLOW
//     D2): `new T(...)` is exactly T; a parameter is followed to its
//     callers (dataflow.CallerIndex); anything else is its static type,
//     or the publishEvent parameter's type (Object: every listener) when
//     unknown — never a guessed narrower type;
//   - a listener method is a method of an application class annotated
//     @EventListener or with an annotation itself meta-annotated
//     @EventListener (@TransactionalEventListener); its event types are
//     the annotation's classes/value element, else its one parameter's
//     type;
//   - it matches when Spring could deliver the event to it: the listener
//     type is a supertype of the published type (a payload object that is
//     not an ApplicationEvent is wrapped in PayloadApplicationEvent, which
//     a listener on that type or a supertype also receives); for a static
//     (upper-bound) published type, also when the listener type is a
//     subtype of it;
//   - the listener's class must be registered as a bean: related (equal,
//     subtype or supertype) to a bean type the caller passes in — the
//     stereotypes, @Bean products and XML beans the spring profile finds.
//     A listener on a class that is no bean gets nothing;
//   - gate: a reachable method creating an ApplicationListenerMethodAdapter
//     (the EventListenerFactory implementations), so an input that never
//     processes listener annotations gets nothing.
//
// The edge says the listener may run: its condition (SpEL), a
// @TransactionalEventListener's phase (it runs only for an event published
// in a transaction, after commit by default), an @Async executor and
// generic type arguments of the event are not evaluated.
package springevents

import (
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ModelID identifies the model; its edges carry callgraph.BasisModel.
const ModelID = "spring:events"

// Version versions the rules. Snapshots record it.
const Version = "ev1"

const (
	eventListener      = "org.springframework.context.event.EventListener"
	publisher          = "org.springframework.context.ApplicationEventPublisher"
	applicationEvent   = "org.springframework.context.ApplicationEvent"
	payloadEvent       = "org.springframework.context.PayloadApplicationEvent"
	methodAdapter      = "org.springframework.context.event.ApplicationListenerMethodAdapter"
	objectFQCN         = "java.lang.Object"
	publishObjectDesc  = "(Ljava/lang/Object;)V"
	publishAppEvtDesc  = "(Lorg/springframework/context/ApplicationEvent;)V"
	publishMethodName  = "publishEvent"
	maxTypesPerSite    = 16
	maxAnnotationDepth = 3
)

// Listener is one annotated listener method.
type Listener struct {
	FQCN, Name, Desc string
	// Annotation is the listener annotation on the method (@EventListener
	// or one meta-annotated with it).
	Annotation string
	// Events are the event types the listener declares.
	Events []string
	// Registered: the class is related to a bean type.
	Registered bool
}

// Site is one publishing call site and the types it may publish.
type Site struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Exact types are allocated at the site (or by the callers passing
	// the parameter); Bounds are static upper bounds.
	Exact, Bounds []string
	// Unbounded: nothing narrower than the publishEvent parameter's type
	// is known (Bounds holds that type); Reason says why.
	Unbounded bool
	Reason    string
}

// Result is the model applied to one input.
type Result struct {
	Listeners []Listener
	Sites     []Site
	// Forwarding counts publisher implementations' own forwarding calls,
	// left out.
	Forwarding int
	// Edges are the candidates for callgraph.NewCreatorGate.
	Edges []callgraph.SAMEdge
	// NoCreator: listeners exist but nothing creates the adapter.
	NoCreator bool
}

// Find builds the listener catalog, the publishing sites and the
// candidate edges. beans are the bean types the spring profile found.
func Find(h *hierarchy.Hierarchy, beans []string) Result {
	var r Result
	if _, ok := h.Lookup(publisher); !ok {
		return r
	}
	r.Listeners = listeners(h, beans)
	if len(r.Listeners) == 0 {
		return r
	}
	var gate []callgraph.MethodRef
	adapters := map[string]bool{methodAdapter: true}
	for _, s := range h.Subtypes(methodAdapter) {
		adapters[s] = true
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, ref := range m.InitRefs {
				if ref.Kind == classfile.InitNew && adapters[ref.Owner] {
					gate = append(gate, callgraph.MethodRef{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor})
					break
				}
			}
		}
	}
	if len(gate) == 0 {
		r.NoCreator = true
		return r
	}
	idx := dataflow.NewCallerIndex(h)
	defer idx.Close()
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if !isPublishCall(h, c) {
					continue
				}
				var site Site
				forwarding := false
				if assignable(h, fqcn, methodAdapter) {
					site = resultSite(h, r.Listeners, ci, m, c)
				} else {
					site, forwarding = publishedTypes(h, idx, ci, m, c)
				}
				if forwarding {
					r.Forwarding++
					continue
				}
				r.Sites = append(r.Sites, site)
				for _, l := range r.Listeners {
					if !l.Registered || !matches(h, site, l.Events) {
						continue
					}
					r.Edges = append(r.Edges, callgraph.SAMEdge{
						CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC, Kind: c.Kind,
						TargetFQCN: l.FQCN, TargetName: l.Name, TargetDesc: l.Desc,
						Creators: gate,
					})
				}
			}
		}
	}
	return r
}

func isPublishCall(h *hierarchy.Hierarchy, c classfile.Call) bool {
	if c.Name != publishMethodName || (c.Desc != publishObjectDesc && c.Desc != publishAppEvtDesc) {
		return false
	}
	if c.Kind != classfile.CallInterface && c.Kind != classfile.CallVirtual {
		return false
	}
	return assignable(h, c.Owner, publisher)
}

// resultSite: the adapter publishes what a listener method returns
// (ApplicationListenerMethodAdapter#handleResult: a non-null result is
// published, an array's or a collection's elements one by one). The
// published types are the registered listeners' return types.
func resultSite(h *hierarchy.Hierarchy, ls []Listener, ci *hierarchy.ClassInfo, m hierarchy.Method, c classfile.Call) Site {
	site := Site{CallerFQCN: ci.FQCN, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC, Reason: "listener results"}
	bounds := map[string]bool{}
	for _, l := range ls {
		if !l.Registered {
			continue
		}
		ret := l.Desc[strings.IndexByte(l.Desc, ')')+1:]
		if ret == "V" || len(ret) == 1 {
			continue // void or primitive: nothing published
		}
		t := descType(ret)
		switch {
		case strings.HasPrefix(t, "["):
			t = descType(strings.TrimPrefix(t, "["))
		case assignable(h, t, "java.util.Collection") || t == "java.util.Collection" || t == "java.util.List" || t == "java.util.Set" ||
			t == "java.util.concurrent.CompletionStage" || t == "java.util.concurrent.CompletableFuture" || t == "org.springframework.util.concurrent.ListenableFuture":
			t = objectFQCN // elements or an asynchronous result: unknown
		}
		if len(t) == 1 {
			continue
		}
		bounds[t] = true
	}
	site.Bounds = sortedKeys(bounds)
	if bounds[objectFQCN] {
		site.Unbounded = true
	}
	return site
}

// descType renders a field descriptor as an FQCN (object types) or keeps
// it (primitives, arrays).
func descType(d string) string {
	if strings.HasPrefix(d, "L") && strings.HasSuffix(d, ";") {
		return strings.ReplaceAll(d[1:len(d)-1], "/", ".")
	}
	return d
}

// publishedTypes resolves the argument of a publishing call.
func publishedTypes(h *hierarchy.Hierarchy, idx *dataflow.CallerIndex, ci *hierarchy.ClassInfo, m hierarchy.Method, c classfile.Call) (Site, bool) {
	site := Site{CallerFQCN: ci.FQCN, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC}
	declared := objectFQCN
	if c.Desc == publishAppEvtDesc {
		declared = applicationEvent
	}
	unbounded := func(reason string) {
		site.Exact, site.Bounds, site.Unbounded, site.Reason = nil, []string{declared}, true, reason
	}
	caller := dataflow.MethodID{Owner: ci.FQCN, Name: m.Name, Desc: m.Descriptor}
	facts, reason := idx.Facts(caller)
	if facts == nil {
		unbounded("no local-value facts: " + reason)
		return site, false
	}
	if aborted, why := facts.Aborted(); aborted {
		unbounded("local-value analysis aborted: " + why)
		return site, false
	}
	at := bytecodeir.Site{Method: ci.FQCN + "#" + m.Name, Descriptor: m.Descriptor, PC: c.PC, Op: bytecodeir.OpCall}
	v := facts.ArgumentAt(at, 0)
	// A publisher implementation forwarding its own parameter (a child
	// context publishing to its parent) is not where the event is made.
	if v.Kind == valueflow.KindTyped && v.Parameter != nil && (m.Name == publishMethodName || m.Name == "multicastEvent") && assignable(h, ci.FQCN, publisher) {
		return site, true
	}
	exact, bounds := map[string]bool{}, map[string]bool{}
	var why []string
	var walk func(v valueflow.Value, from dataflow.MethodID, depth int)
	walk = func(v valueflow.Value, from dataflow.MethodID, depth int) {
		switch v.Kind {
		case valueflow.KindAlloc, valueflow.KindConst:
			exact[v.Type] = true
		case valueflow.KindNull:
		case valueflow.KindSet:
			for _, c := range v.Candidates {
				walk(c, from, depth)
			}
		case valueflow.KindTyped:
			if v.Parameter != nil && depth == 0 {
				pv := idx.ResolveParameter(from, v.Parameter.Index)
				for _, a := range pv.Values {
					walk(a, from, depth+1)
				}
				if pv.Complete() {
					return
				}
				why = append(why, pv.Unknown...)
			}
			if v.Type != "" {
				bounds[v.Type] = true
			} else {
				bounds[declared] = true
			}
		default:
			if v.Type != "" && v.Kind != valueflow.KindUnknown {
				bounds[v.Type] = true
				return
			}
			why = append(why, string(v.Kind)+" "+v.Reason)
			bounds[declared] = true
		}
	}
	walk(v, caller, 0)
	if len(exact)+len(bounds) > maxTypesPerSite {
		unbounded("more than the per-site type budget")
		return site, false
	}
	site.Exact, site.Bounds = sortedKeys(exact), sortedKeys(bounds)
	if bounds[declared] {
		site.Unbounded = true
		site.Reason = strings.Join(why, "; ")
	}
	return site, false
}

// matches reports whether Spring could deliver an event the site
// publishes to a listener declaring one of events.
func matches(h *hierarchy.Hierarchy, s Site, events []string) bool {
	for _, p := range events {
		for _, t := range s.Exact {
			if delivers(h, t, p) {
				return true
			}
		}
		for _, b := range s.Bounds {
			// The runtime type is b or a subtype of it.
			if delivers(h, b, p) || assignable(h, p, b) {
				return true
			}
			if b == objectFQCN || !isApplicationEvent(h, b) && isInterfaceOrUnknown(h, b) {
				// An unknown payload: anything, wrapped when not an event.
				if assignable(h, payloadEvent, p) {
					return true
				}
			}
		}
	}
	return false
}

// delivers: an event of runtime type t reaches a listener on type p.
func delivers(h *hierarchy.Hierarchy, t, p string) bool {
	if assignable(h, t, p) {
		return true
	}
	// A payload object is wrapped in a PayloadApplicationEvent.
	return !isApplicationEvent(h, t) && assignable(h, payloadEvent, p)
}

func isApplicationEvent(h *hierarchy.Hierarchy, t string) bool {
	return assignable(h, t, applicationEvent)
}

func isInterfaceOrUnknown(h *hierarchy.Hierarchy, t string) bool {
	ci, ok := h.Lookup(t)
	return !ok || ci.IsInterface()
}

// assignable reports whether t is p or a subtype of it, walking up from t
// (supertypes outside the input are compared by name).
func assignable(h *hierarchy.Hierarchy, t, p string) bool {
	if p == objectFQCN || t == p {
		return true
	}
	seen := map[string]bool{}
	var up func(string) bool
	up = func(c string) bool {
		if c == "" || seen[c] {
			return false
		}
		seen[c] = true
		if c == p {
			return true
		}
		ci, ok := h.Lookup(c)
		if !ok {
			return false
		}
		if up(ci.SuperClass) {
			return true
		}
		for _, i := range ci.Interfaces {
			if up(i) {
				return true
			}
		}
		return false
	}
	return up(t)
}

// listeners finds the annotated listener methods of the application.
func listeners(h *hierarchy.Hierarchy, beans []string) []Listener {
	meta := map[string]bool{}
	isListenerAnnotation := func(a string) bool {
		if v, ok := meta[a]; ok {
			return v
		}
		v := metaAnnotated(h, a, 0, map[string]bool{})
		meta[a] = v
		return v
	}
	var out []Listener
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp || ci.IsInterface() {
			continue
		}
		for _, m := range ci.Methods {
			for _, ann := range m.AnnotationDetails {
				if !isListenerAnnotation(ann.Type) {
					continue
				}
				events := classesElement(ann)
				if len(events) == 0 {
					params := paramTypes(m.Descriptor)
					if len(params) != 1 {
						continue // Spring rejects it
					}
					events = params
				}
				out = append(out, Listener{
					FQCN: fqcn, Name: m.Name, Desc: m.Descriptor, Annotation: ann.Type, Events: events,
					Registered: registered(h, fqcn, beans),
				})
				break
			}
		}
	}
	return out
}

func metaAnnotated(h *hierarchy.Hierarchy, a string, depth int, seen map[string]bool) bool {
	if a == eventListener {
		return true
	}
	if depth >= maxAnnotationDepth || seen[a] {
		return false
	}
	seen[a] = true
	ci, ok := h.Lookup(a)
	if !ok {
		return false
	}
	for _, x := range ci.Annotations {
		if metaAnnotated(h, x, depth+1, seen) {
			return true
		}
	}
	return false
}

// classesElement reads the classes (or its alias value) element.
func classesElement(a classfile.Annotation) []string {
	var out []string
	for _, name := range []string{"classes", "value"} {
		v, ok := a.Value(name)
		if !ok {
			continue
		}
		vals := []classfile.ElementValue{v}
		if v.Kind == classfile.ElementArray {
			vals = v.Array
		}
		for _, e := range vals {
			if e.Kind == classfile.ElementClass && e.ClassFQCN != "" {
				out = append(out, e.ClassFQCN)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return out
}

func registered(h *hierarchy.Hierarchy, fqcn string, beans []string) bool {
	for _, b := range beans {
		if assignable(h, b, fqcn) || assignable(h, fqcn, b) && b != objectFQCN {
			return true
		}
	}
	return false
}

// paramTypes returns a method descriptor's parameter types as FQCNs
// (primitives and arrays as their descriptors).
func paramTypes(desc string) []string {
	end := strings.IndexByte(desc, ')')
	if !strings.HasPrefix(desc, "(") || end < 0 {
		return nil
	}
	s := desc[1:end]
	var out []string
	for len(s) > 0 {
		i := 0
		for s[i] == '[' {
			i++
		}
		if s[i] == 'L' {
			j := strings.IndexByte(s, ';')
			if i == 0 {
				out = append(out, strings.ReplaceAll(s[1:j], "/", "."))
			} else {
				out = append(out, s[:j+1])
			}
			s = s[j+1:]
			continue
		}
		out = append(out, s[:i+1])
		s = s[i+1:]
	}
	return out
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RegisteredCount counts the listeners on bean classes.
func (r Result) RegisteredCount() int {
	n := 0
	for _, l := range r.Listeners {
		if l.Registered {
			n++
		}
	}
	return n
}

// UnboundedCount counts the sites publishing an unknown type.
func (r Result) UnboundedCount() int {
	n := 0
	for _, s := range r.Sites {
		if s.Unbounded {
			n++
		}
	}
	return n
}
