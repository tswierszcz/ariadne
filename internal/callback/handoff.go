package callback

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// Hand-offs (ASYNC_CALLBACKS a3-3): an object a method stores for later
// and the code that takes it back and acts on it, possibly on another
// thread — a library's own work queue (Jetty's ManagedSelector#submit:
// `_updates.offer(update)`, drained by SelectorProducer#processUpdates,
// which calls update.update(selector)) or an NIO selection key's
// attachment (`channel.register(selector, ops, this)`, read back with
// `key.attachment()` when the selector reports the key ready). Neither is
// a call, and no JDK executor is involved, so A2 has no site for them.
//
// Queue: a store puts an object into a JDK collection held in a field; a
// drain takes elements out of a collection field — removing them
// (poll/take/pop…) or reading them (iterating, peek) and clearing the
// field (a walk without clearing is a broadcast to listeners, A3's own
// scope; a peek, a stack's current element) — and calls
// methods on them. Store and drain pair on the same field, or on two
// collection fields of one class when the stored type is known and the
// drain calls its methods (double buffering). The edge goes from where
// the object enters — the store for a direct `new`/lambda there, or the
// caller passing it when it is the storing method's parameter (a3-2's
// forwarding) — to the method the drain's call runs on that object, so a
// caller reaches only what it hands in (ManagedSelector#dump's DumpKeys,
// not every update).
//
// Selection key: the edge goes from the register/attach call to each
// method of the same class nest that reads an attachment and dispatches
// on readiness (isReadable…/readyOps, or takes selectedKeys() itself) —
// the select loop, not code walking every key (dump, close).
//
// Both are "may run" edges, like A2's, conditioned on what has to happen
// first; neither claims the drain runs.
const (
	// ConditionQueued: the target runs when its class drains the queue
	// the object was put into.
	ConditionQueued = "queued"
	// ConditionSelected: the target runs when the NIO selector reports
	// the key the object is attached to as ready.
	ConditionSelected = "selected"
)

// MaxDrains bounds the drain methods one store connects to; a store with
// more gets no edge and is counted (HandoffStats.TooManyDrains).
const MaxDrains = 8

// HandoffStats summarizes FindHandoffs for the run log.
type HandoffStats struct {
	// QueueStores: calls putting an object into a JDK collection held
	// in a field (add/offer/push/put…); QueueStoresDrained: those with
	// at least one drain of the same class's collection fields.
	QueueStores, QueueStoresDrained int
	// SelectorStores: NIO register(…, attachment)/attach calls;
	// SelectorStoresDrained: those whose class nest reads an attachment.
	SelectorStores, SelectorStoresDrained int
	// Drains: distinct methods found draining a collection field or
	// reading a selection key's attachment.
	Drains        int
	TooManyDrains int
	// QueueUnresolved: drained stores whose object is neither a direct
	// `new`/lambda nor a parameter its callers fill with one — no edge.
	QueueUnresolved int
}

type handoffMethod struct{ fqcn, name, desc string }

func (m handoffMethod) key() string { return m.fqcn + "#" + m.name + m.desc }

type queueStore struct {
	at handoffMethod
	pc int
	// field: the collection field (collectionField); holder: the class
	// declaring it; elem: the stored value's static type ("" unknown).
	field, holder, elem string
	value               valueflow.Value
}

type queueDrain struct {
	at            handoffMethod
	field, holder string
	// calls: the methods called on a taken element.
	calls []classfile.Call
}

// queued is one object a queue store hands over: the call site where it
// enters (the store itself, or for a parameter the caller passing it —
// a3-2's forwarding) and what it is.
type queued struct {
	from handoffMethod
	pc   int
	rc   ResolvedCallback
}

// FindHandoffs finds hand-off stores and their drains and returns their
// edges (see above).
func FindHandoffs(h *hierarchy.Hierarchy) ([]CallbackEdge, HandoffStats, []string) {
	r := ScanHandoffs(h)
	return r.Edges, r.Stats, r.Warnings
}

// HandoffResult is ScanHandoffs' result: the hand-off edges (a3-3) and
// the listener candidates (a3-5), from one scan of the input.
type HandoffResult struct {
	Edges         []CallbackEdge
	Stats         HandoffStats
	Listeners     []GatedEdge
	ListenerStats ListenerStats
	Warnings      []string
}

// ScanHandoffs finds hand-offs (a3-3) and listener notifications (a3-5).
func ScanHandoffs(h *hierarchy.Hierarchy) HandoffResult {
	var (
		stats     HandoffStats
		warnings  []string
		stores    []queueStore
		drains    []queueDrain
		bcasts    []queueDrain
		pcs       pcsSites
		selStores []queueStore // holder: the store's class nest
		selDrains = map[string][]handoffMethod{}
		jar       classfile.JarClassReader
	)
	defer jar.Close()

	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || ci.SourceJar == "" {
			continue
		}
		var cf *classfile.ClassFile
		for _, m := range ci.Methods {
			if !candidateMethod(m.Calls) {
				continue
			}
			if cf == nil {
				var err error
				if cf, err = jar.ReadClass(ci.SourceJar, ci.SourceEntry); err != nil {
					warnings = append(warnings, fmt.Sprintf("re-reading %s (%s) for hand-off analysis: %v", ci.SourceJar, ci.SourceEntry, err))
					break
				}
			}
			facts, err := dataflow.AnalyzeWithParameterRefs(cf, fqcn, m.Method)
			if err != nil {
				continue
			}
			if aborted, _ := facts.Aborted(); aborted {
				continue
			}
			at := handoffMethod{fqcn, m.Name, m.Descriptor}
			s, d, ss, sd, b, p := scanMethod(h, at, m.Calls, facts)
			stores = append(stores, s...)
			drains = append(drains, d...)
			bcasts = append(bcasts, b...)
			pcs.adds = append(pcs.adds, p.adds...)
			pcs.fires = append(pcs.fires, p.fires...)
			selStores = append(selStores, ss...)
			if sd {
				n := nest(fqcn)
				selDrains[n] = append(selDrains[n], at)
			}
		}
	}

	drainMethods := map[string]bool{}
	var edges []CallbackEdge
	add := func(s queueStore, targets []handoffMethod, cond string) {
		if len(targets) > MaxDrains {
			stats.TooManyDrains++
			return
		}
		for _, t := range targets {
			if t.key() == s.at.key() {
				continue
			}
			drainMethods[t.key()] = true
			edges = append(edges, CallbackEdge{
				CallerFQCN: s.at.fqcn, CallerName: s.at.name, CallerDesc: s.at.desc, PC: s.pc,
				TargetFQCN: t.fqcn, TargetMethod: t.name, TargetMethodDesc: t.desc, Condition: cond,
			})
		}
	}

	stats.QueueStores = len(stores)
	fw := newForwarder(h, &jar)
	seenEdge := map[string]bool{}
	for _, s := range stores {
		var matched []queueDrain
		for _, d := range drains {
			if d.at.key() != s.at.key() && pairs(h, s, d) {
				matched = append(matched, d)
			}
		}
		if len(matched) == 0 {
			continue
		}
		stats.QueueStoresDrained++
		if len(matched) > MaxDrains {
			stats.TooManyDrains++
			continue
		}
		objs := queuedObjects(fw, s)
		if len(objs) == 0 {
			stats.QueueUnresolved++
			continue
		}
		for _, o := range objs {
			for _, d := range matched {
				for _, c := range d.calls {
					t, ok := runs(h, o.rc, c)
					if !ok {
						continue
					}
					drainMethods[d.at.key()] = true
					e := CallbackEdge{
						CallerFQCN: o.from.fqcn, CallerName: o.from.name, CallerDesc: o.from.desc, PC: o.pc,
						TargetFQCN: t.fqcn, TargetMethod: t.name, TargetMethodDesc: t.desc, Condition: ConditionQueued,
					}
					k := fmt.Sprintf("%s@%d>%s", o.from.key(), o.pc, t.key())
					if !seenEdge[k] {
						seenEdge[k] = true
						edges = append(edges, e)
					}
				}
			}
		}
	}
	stats.SelectorStores = len(selStores)
	for _, s := range selStores {
		if targets := selDrains[s.holder]; len(targets) > 0 {
			stats.SelectorStoresDrained++
			add(s, targets, ConditionSelected)
		}
	}
	stats.Drains = len(drainMethods)
	sort.SliceStable(edges, func(i, j int) bool {
		a, b := edges[i], edges[j]
		ka := a.CallerFQCN + "#" + a.CallerName + a.CallerDesc + fmt.Sprintf("@%06d", a.PC) + a.TargetFQCN + "#" + a.TargetMethod + a.TargetMethodDesc
		kb := b.CallerFQCN + "#" + b.CallerName + b.CallerDesc + fmt.Sprintf("@%06d", b.PC) + b.TargetFQCN + "#" + b.TargetMethod + b.TargetMethodDesc
		return ka < kb
	})
	lis, lstats := listenerEdges(h, fw, stores, bcasts, pcs)
	return HandoffResult{Edges: edges, Stats: stats, Listeners: lis, ListenerStats: lstats, Warnings: warnings}
}

const (
	selectionKey     = "java.nio.channels.SelectionKey"
	registerDesc     = "(Ljava/nio/channels/Selector;ILjava/lang/Object;)Ljava/nio/channels/SelectionKey;"
	attachDesc       = "(Ljava/lang/Object;)Ljava/lang/Object;"
	attachmentDesc   = "()Ljava/lang/Object;"
	iteratorOwner    = "java.util.Iterator"
	iteratorNextDesc = "()Ljava/lang/Object;"
)

func isQueueStore(c classfile.Call) bool {
	if !dataflow.IsJDKCollection(c.Owner) {
		return false
	}
	switch c.Name {
	case "add", "offer", "addLast", "addFirst", "offerLast", "offerFirst", "push", "put":
		return c.Desc == "(Ljava/lang/Object;)Z" || c.Desc == "(Ljava/lang/Object;)V"
	}
	return false
}

func isSelectorStore(c classfile.Call) bool {
	return (strings.HasPrefix(c.Owner, "java.nio.channels.") && c.Name == "register" && c.Desc == registerDesc) ||
		(c.Owner == selectionKey && c.Name == "attach" && c.Desc == attachDesc)
}

func isAttachmentRead(c classfile.Call) bool {
	return c.Owner == selectionKey && c.Name == "attachment" && c.Desc == attachmentDesc
}

// candidateMethod: the method has a call that can store or take back —
// only those are analyzed.
func candidateMethod(calls []classfile.Call) bool {
	for _, c := range calls {
		if c.Kind == classfile.CallDynamic {
			continue
		}
		if isQueueStore(c) || isSelectorStore(c) || isAttachmentRead(c) || dataflow.IsCollectionRetrieval(c.Owner, c.Name, c.Desc) || isPCSAdd(c) || isPCSFire(c) {
			return true
		}
	}
	return false
}

func scanMethod(h *hierarchy.Hierarchy, at handoffMethod, calls []classfile.Call, facts *dataflow.MethodFacts) (stores []queueStore, drains []queueDrain, selStores []queueStore, selDrain bool, bcasts []queueDrain, pcs pcsSites) {
	site := func(pc int) bytecodeir.Site {
		return bytecodeir.Site{Method: at.fqcn + "#" + at.name, Descriptor: at.desc, PC: pc, Op: bytecodeir.OpCall}
	}
	// Per collection field: owners called on a taken element; whether
	// an element was removed (not only iterated); whether this method
	// clears the field's collection. Iterating counts only with a clear
	// of the same field.
	taken := map[string][]classfile.Call{}
	removed, cleared := map[string]bool{}, map[string]bool{}
	attachment, ready := false, false
	// The collection fields this method iterates: an Iterator#next whose
	// iterator the analysis lost (a catch handler inside the loop jumps
	// back to its head without the locals) is attributed to the one
	// field, if there is exactly one.
	iterated := map[string]bool{}
	for _, c := range calls {
		if c.Kind != classfile.CallDynamic && dataflow.IsJDKCollection(c.Owner) && c.Name == "iterator" {
			if field, ok := collectionField(h, receiverAt(facts, site(c.PC))); ok {
				iterated[field] = true
			}
		}
	}
	for _, c := range calls {
		if c.Kind == classfile.CallDynamic {
			continue
		}
		switch {
		case isQueueStore(c):
			field, ok := collectionField(h, receiverAt(facts, site(c.PC)))
			if !ok {
				continue
			}
			v, _ := argumentAt(facts, site(c.PC), 0)
			if v.Kind == valueflow.KindNull || v.Kind == valueflow.KindConst {
				continue
			}
			stores = append(stores, queueStore{at: at, pc: c.PC, field: field, holder: fieldClass(field), elem: staticType(v), value: v})
		case isSelectorStore(c):
			arg := 0
			if c.Name == "register" {
				arg = 2
			}
			if v, ok := argumentAt(facts, site(c.PC), arg); ok && v.Kind != valueflow.KindNull {
				selStores = append(selStores, queueStore{at: at, pc: c.PC, holder: nest(at.fqcn)})
			}
		case isPCSAdd(c):
			field, ok := anyField(h, receiverAt(facts, site(c.PC)))
			if !ok {
				continue
			}
			v, _ := argumentAt(facts, site(c.PC), pcsListenerArg(c))
			if v.Kind == valueflow.KindNull || v.Kind == valueflow.KindConst {
				continue
			}
			pcs.adds = append(pcs.adds, queueStore{at: at, pc: c.PC, field: field, holder: fieldClass(field), value: v})
		case isPCSFire(c):
			if field, ok := anyField(h, receiverAt(facts, site(c.PC))); ok {
				pcs.fires = append(pcs.fires, pcsFire{at: at, pc: c.PC, field: field, kind: c.Kind})
			}
		case isAttachmentRead(c):
			attachment = true
		case isReadinessCall(c):
			ready = true
		case dataflow.IsJDKCollection(c.Owner) && c.Name == "clear" && c.Desc == "()V":
			if field, ok := collectionField(h, receiverAt(facts, site(c.PC))); ok {
				cleared[field] = true
			}
		case c.Kind == classfile.CallVirtual || c.Kind == classfile.CallInterface:
			if !elementCall(c) {
				continue
			}
			field, viaIterator, ok := takenFrom(h, receiverAt(facts, site(c.PC)))
			if !ok {
				continue
			}
			if field == lostIterator {
				if len(iterated) != 1 {
					continue
				}
				for f := range iterated {
					field = f
				}
			}
			taken[field] = append(taken[field], c)
			if !viaIterator {
				removed[field] = true
			}
		}
	}
	fields := make([]string, 0, len(taken))
	for k := range taken {
		fields = append(fields, k)
	}
	sort.Strings(fields)
	for _, field := range fields {
		// Iterating without clearing is a broadcast (listeners), not a
		// drain: the elements stay for the next event.
		if removed[field] || cleared[field] {
			drains = append(drains, queueDrain{at: at, field: field, holder: fieldClass(field), calls: taken[field]})
		} else {
			// A walk that keeps the elements: a broadcast to listeners
			// (a3-5).
			bcasts = append(bcasts, queueDrain{at: at, field: field, holder: fieldClass(field), calls: taken[field]})
		}
	}
	// A selection-key drain dispatches on readiness: it reads the
	// attachment and checks the key's ready operations, or takes the
	// keys from selectedKeys() itself — not code that walks every
	// registered key (dump, close, shutdown).
	selDrain = attachment && ready
	return stores, drains, selStores, selDrain, bcasts, pcs
}

func isReadinessCall(c classfile.Call) bool {
	if c.Owner == selectionKey {
		switch c.Name {
		case "readyOps", "isReadable", "isWritable", "isConnectable", "isAcceptable":
			return true
		}
	}
	return strings.HasPrefix(c.Owner, "java.nio.channels.") && strings.HasSuffix(c.Owner, "Selector") && c.Name == "selectedKeys" && c.Desc == "()Ljava/util/Set;"
}

// elementCall: a call that acts on an element, not Object/String
// plumbing (logging, equality).
func elementCall(c classfile.Call) bool {
	switch c.Name {
	case "equals", "hashCode", "toString", "getClass":
		return false
	}
	return !strings.HasPrefix(c.Owner, "java.lang.") || c.Owner == "java.lang.Runnable"
}

// takenFrom: v is an element taken out of a collection field (a
// retrieval, or Iterator#next on its iterator); the field (see
// collectionField), and whether it was only read (iterated or peeked),
// not removed.
//
// A set (a loop whose catch handler jumps back to its head loses the
// locals) counts when its traceable candidates agree on one class.
func takenFrom(h *hierarchy.Hierarchy, v valueflow.Value) (string, bool, bool) {
	if v.Kind == valueflow.KindSet {
		holder, iterated, found := "", false, false
		for _, c := range v.Candidates {
			ch, it, ok := takenFrom(h, c)
			if !ok || (ch == lostIterator && found) {
				continue
			}
			if found && holder == lostIterator {
				holder, iterated = ch, it
				continue
			}
			if found && ch != holder {
				return "", false, false
			}
			holder, iterated, found = ch, iterated || it, true
		}
		return holder, iterated, found
	}
	if v.Kind != valueflow.KindCallTrace || v.CallTrace == nil || v.CallTrace.Receiver == nil {
		return "", false, false
	}
	t := v.CallTrace
	if t.Owner == iteratorOwner && t.Name == "next" && t.Desc == iteratorNextDesc {
		it := t.Receiver
		if it.Kind == valueflow.KindUnknown {
			return lostIterator, true, true
		}
		if it.Kind != valueflow.KindCallTrace || it.CallTrace == nil || it.CallTrace.Name != "iterator" || it.CallTrace.Receiver == nil {
			return "", false, false
		}
		holder, ok := collectionField(h, *it.CallTrace.Receiver)
		return holder, true, ok
	}
	if t.Name == "iterator" || !dataflow.IsCollectionRetrieval(t.Owner, t.Name, t.Desc) {
		return "", false, false
	}
	holder, ok := collectionField(h, *t.Receiver)
	// peek/element/getFirst/getLast read the head without taking it (a
	// stack's current scope): like iterating, they need a clear.
	return holder, !removes(t.Name), ok
}

// collectionField: v is read from a field holding a JDK collection,
// directly or through a static accessor returning it (javac's access$NNN
// for a nested class); the field as "declaring class#name".
func collectionField(h *hierarchy.Hierarchy, v valueflow.Value) (string, bool) {
	if v.Kind == valueflow.KindCallResult && v.Call != nil {
		v, _ = dataflow.ResolveCallResultOneHop(h, *v.Call)
	}
	if v.Kind != valueflow.KindField || v.Field == nil {
		return "", false
	}
	d := v.Field.Desc
	if !strings.HasPrefix(d, "L") || !strings.HasSuffix(d, ";") || !dataflow.IsJDKCollection(strings.ReplaceAll(d[1:len(d)-1], "/", ".")) {
		return "", false
	}
	decl, ok := h.ResolveField(v.Field.Owner, v.Field.Name, v.Field.Desc)
	if !ok {
		return "", false
	}
	return decl + "#" + v.Field.Name, true
}

func removes(name string) bool {
	switch name {
	case "peek", "peekFirst", "peekLast", "element", "getFirst", "getLast", "get":
		return false
	}
	return true
}

// lostIterator: takenFrom's field for Iterator#next on an iterator the
// analysis lost track of; scanMethod resolves it.
const lostIterator = "?"

// fieldClass: the declaring class of a collectionField key. Stores and
// drains are paired by it, not by the field: a class may swap two
// collections (double buffering, Jetty's _updates/_updateable).
func fieldClass(field string) string { return field[:strings.IndexByte(field, '#')] }

func staticType(v valueflow.Value) string {
	switch v.Kind {
	case valueflow.KindAlloc, valueflow.KindTyped:
		// A parameter's type is its descriptor (Lpkg/Name;).
		if t := v.Type; strings.HasPrefix(t, "L") && strings.HasSuffix(t, ";") {
			return strings.ReplaceAll(t[1:len(t)-1], "/", ".")
		} else if strings.HasPrefix(t, "[") || len(t) == 1 {
			return ""
		}
		return v.Type
	case valueflow.KindFunctional:
		if v.Func != nil {
			return v.Func.SamInterface
		}
	}
	return ""
}

// pairs: the drain takes from the store's own field, or — a class
// swapping its collections (double buffering) — from another collection
// field of the same class, acting on elements of a type the stored value
// can have (the stored type or a supertype/subtype of it; it has to be
// known).
func pairs(h *hierarchy.Hierarchy, s queueStore, d queueDrain) bool {
	if d.field == s.field {
		return true
	}
	return d.holder == s.holder && elementMatches(h, s.elem, d.calls)
}

func elementMatches(h *hierarchy.Hierarchy, elem string, calls []classfile.Call) bool {
	if elem == "" || elem == "java.lang.Object" {
		return false
	}
	for _, c := range calls {
		t := c.Owner
		if t == elem || subtypeOf(h, elem, t) || subtypeOf(h, t, elem) {
			return true
		}
	}
	return false
}

func subtypeOf(h *hierarchy.Hierarchy, sub, sup string) bool {
	for _, s := range h.Subtypes(sup) {
		if s == sub {
			return true
		}
	}
	return false
}

// queuedObjects: what a store puts in — a direct `new` or lambda at the
// store, or, for the storing method's parameter, what its callers pass
// (a3-2's forwarder), each from the call site where it enters.
func queuedObjects(fw *forwarder, s queueStore) []queued {
	v := s.value
	switch {
	case v.Kind == valueflow.KindAlloc:
		return []queued{{from: s.at, pc: s.pc, rc: ResolvedCallback{TargetFQCN: v.Type}}}
	case v.Kind == valueflow.KindFunctional && v.Func != nil && v.Func.ImplOwner != "":
		return []queued{{from: s.at, pc: s.pc, rc: ResolvedCallback{ImplOwner: v.Func.ImplOwner, ImplName: v.Func.ImplName, ImplDesc: v.Func.ImplDesc,
			SamInterface: v.Func.SamInterface, SamMethod: v.Func.SamMethod}}}
	case v.Kind == valueflow.KindTyped && v.Parameter != nil && v.Parameter.Index >= 0:
		fw.site = RegistrationSite{CallerFQCN: s.at.fqcn, CallerName: s.at.name, CallerDesc: s.at.desc, PC: s.pc}
		var out []queued
		for _, rc := range fw.resolve(s.at.key(), v.Parameter.Index, nil, 0) {
			if rc.Origin == nil {
				continue
			}
			out = append(out, queued{from: handoffMethod{rc.Origin.FQCN, rc.Origin.Name, rc.Origin.Desc}, pc: rc.Origin.PC, rc: rc})
		}
		return out
	}
	return nil
}

// runs: the method a drain's call c runs on the queued object, if the
// object can be c's receiver — the class's own resolution of c for a
// `new`, the implementation for a lambda of c's interface method.
func runs(h *hierarchy.Hierarchy, rc ResolvedCallback, c classfile.Call) (handoffMethod, bool) {
	switch {
	case rc.TargetFQCN != "":
		if rc.TargetFQCN != c.Owner && !subtypeOf(h, rc.TargetFQCN, c.Owner) {
			return handoffMethod{}, false
		}
		decl, ok := h.ResolveExact(rc.TargetFQCN, c.Name, c.Desc)
		if !ok {
			return handoffMethod{}, false
		}
		return handoffMethod{decl, c.Name, c.Desc}, true
	case rc.ImplOwner != "":
		if rc.SamInterface != c.Owner || rc.SamMethod != c.Name {
			return handoffMethod{}, false
		}
		return handoffMethod{rc.ImplOwner, rc.ImplName, rc.ImplDesc}, true
	}
	return handoffMethod{}, false
}

// nest: the top-level class of fqcn (Outer for Outer$Inner).
func nest(fqcn string) string {
	if i := strings.IndexByte(fqcn, '$'); i > 0 {
		return fqcn[:i]
	}
	return fqcn
}

func receiverAt(f *dataflow.MethodFacts, at bytecodeir.Site) (v valueflow.Value) {
	defer func() {
		if recover() != nil {
			v = valueflow.Unknown("no receiver fact")
		}
	}()
	return f.ReceiverAt(at)
}
