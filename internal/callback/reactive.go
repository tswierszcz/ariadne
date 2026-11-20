package callback

import (
	"fmt"
	"sort"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Reactive streams (ASYNC_CALLBACKS a3-6): one library, one version — the
// JDK's java.util.concurrent.Flow with its implementation
// SubmissionPublisher (Java 9+; the analyzer's own JDK 25 is what the
// oracle ran). A pipeline that is only built does nothing; a subscriber
// runs when a publisher it subscribed to emits. The catalog, per the
// SubmissionPublisher contract and checked on the JVM
// (testdata/async-reactive):
//
//   - subscribe(s): s.onSubscribe(subscription) — condition "subscribed";
//   - submit(item)/offer(item, …) on that publisher: s.onNext(item) —
//     "published" (demand — Subscription#request — is not modeled: a
//     subscriber that never requests still has the edge);
//   - close(): s.onComplete() — "completed";
//   - closeExceptionally(t): s.onError(t) — "exceptional"; an onNext that
//     throws also ends in onError (the JVM runs it), so submit/offer get
//     that edge too;
//   - consume(c): the Consumer as a subscriber's onNext (c.accept(item))
//     on submit/offer.
//
// A publisher is identified by the objects its receiver may be (a `new`
// site, `this` in a SubmissionPublisher subclass — a Flow.Processor that
// republishes with submit —, a parameter's callers, a private/final
// field's writers: returnedResolver.receiverObjs); a subscription and an
// emission pair when they share one. A receiver that resolves to nothing
// of the kind is a gap (counted), never "every publisher". The
// subscriber is the subscribe argument: a `new`, or a parameter its
// callers fill with one (a3-2's forwarder); a lambda for consume.
//
// Edges go from the subscribe call (onSubscribe) or the emitting call
// (onNext, onComplete, onError) to the subscriber's method, gated
// (callgraph.NewCreatorGate) on the method where the subscriber enters —
// a subscriber created but never subscribed, or subscribed only in code
// that never runs, gets nothing; a publisher that never emits gives only
// onSubscribe (and onComplete/onError if it is closed).
//
// Not modeled: publishers of other libraries and Flow.Publisher
// implementations of the input (their subscribe is ordinary code calling
// the subscriber — a JDK-typed parameter CHA does not expand), demand and
// cancellation, the executor's thread (asynchronous by default).
const (
	// ReactiveVersion names the rules; a snapshot records it.
	ReactiveVersion = "a3-6"
	// ConditionSubscribed: the target runs once the subscription is set up.
	ConditionSubscribed = "subscribed"
	// ConditionPublished: the target runs when the publisher emits an item.
	ConditionPublished = "published"
	// ConditionCompleted: the target runs when the publisher is closed.
	ConditionCompleted = "completed"
)

const (
	submissionPublisher = "java.util.concurrent.SubmissionPublisher"
	flowPublisher       = "java.util.concurrent.Flow$Publisher"
	subscribeDesc       = "(Ljava/util/concurrent/Flow$Subscriber;)V"
	consumeDesc         = "(Ljava/util/function/Consumer;)Ljava/util/concurrent/CompletableFuture;"
	onSubscribeDesc     = "(Ljava/util/concurrent/Flow$Subscription;)V"
	onNextDesc          = "(Ljava/lang/Object;)V"
	onErrorDesc         = "(Ljava/lang/Throwable;)V"
)

// ReactiveStats summarizes FindReactive for the run log.
type ReactiveStats struct {
	// Subscriptions: subscribe/consume calls on a SubmissionPublisher
	// with a resolved subscriber; Emissions: submit/offer/close/
	// closeExceptionally calls; Unresolved: subscriptions or emissions
	// whose publisher, or subscriptions whose subscriber, did not resolve.
	Subscriptions, Emissions, Unresolved int
}

type emitKind int

const (
	emitItem emitKind = iota
	emitClose
	emitError
)

type rxSite struct {
	at   handoffMethod
	pc   int
	kind classfile.CallKind
	pubs []objID
}

// FindReactive finds Flow subscriptions to SubmissionPublishers and the
// calls that make them emit, and returns the gated edges (see above).
func FindReactive(h *hierarchy.Hierarchy) ([]GatedEdge, ReactiveStats, []string) {
	var (
		stats ReactiveStats
		jar   classfile.JarClassReader
	)
	defer jar.Close()
	rstats := ReturnedStats{}
	r := &returnedResolver{h: h, fw: newForwarder(h, &jar), stats: &rstats, returns: map[string][]taskObj{}, returners: map[string]bool{}}

	// isPub: SubmissionPublisher or a subclass in the input (the chain
	// ends at the first class the hierarchy has not read — the JDK's).
	isPub := func(owner string) bool {
		if owner == submissionPublisher {
			return true
		}
		for _, c := range h.SuperclassChain(owner) {
			if ci, ok := h.Lookup(c); ok && ci.SuperClass == submissionPublisher {
				return true
			}
		}
		return false
	}

	type sub struct {
		rxSite
		consume bool
		objs    []queued
	}
	var subs []sub
	emits := map[emitKind][]rxSite{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if ci.SourceJar == "" {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind != classfile.CallVirtual && c.Kind != classfile.CallInterface {
					continue
				}
				var ek emitKind
				isSub, isConsume, isEmit := false, false, false
				switch {
				case c.Name == "subscribe" && c.Desc == subscribeDesc && (c.Owner == flowPublisher || isPub(c.Owner)):
					isSub = true
				case c.Name == "consume" && c.Desc == consumeDesc && isPub(c.Owner):
					isSub, isConsume = true, true
				case isPub(c.Owner) && (c.Name == "submit" && c.Desc == "(Ljava/lang/Object;)I" ||
					c.Name == "offer" && (c.Desc == "(Ljava/lang/Object;Ljava/util/function/BiPredicate;)I" ||
						c.Desc == "(Ljava/lang/Object;JLjava/util/concurrent/TimeUnit;Ljava/util/function/BiPredicate;)I")):
					isEmit, ek = true, emitItem
				case isPub(c.Owner) && c.Name == "close" && c.Desc == "()V":
					isEmit, ek = true, emitClose
				case isPub(c.Owner) && c.Name == "closeExceptionally" && c.Desc == onErrorDesc:
					isEmit, ek = true, emitError
				}
				if !isSub && !isEmit {
					continue
				}
				at := handoffMethod{fqcn, m.Name, m.Descriptor}
				facts := r.fw.analyze(fqcn, m)
				if facts == nil {
					stats.Unresolved++
					continue
				}
				site := bytecodeir.Site{Method: fqcn + "#" + m.Name, Descriptor: m.Descriptor, PC: c.PC, Op: bytecodeir.OpCall}
				recv := receiverAt(facts, site)
				r.work, r.exhausted = 0, false
				pubs, ok := r.receiverObjs(&recv, at, 0)
				pubs = publishers(h, pubs, isPub)
				if !ok || len(pubs) == 0 {
					stats.Unresolved++
					continue
				}
				s := rxSite{at: at, pc: c.PC, kind: c.Kind, pubs: pubs}
				if isEmit {
					stats.Emissions++
					emits[ek] = append(emits[ek], s)
					continue
				}
				arg, ok := argumentAt(facts, site, 0)
				if !ok {
					stats.Unresolved++
					continue
				}
				objs := queuedObjects(r.fw, queueStore{at: at, pc: c.PC, value: arg})
				if len(objs) == 0 {
					stats.Unresolved++
					continue
				}
				stats.Subscriptions++
				subs = append(subs, sub{rxSite: s, consume: isConsume, objs: objs})
			}
		}
	}

	var out []GatedEdge
	idx := map[string]int{}
	add := func(from rxSite, t handoffMethod, creator handoffMethod, cond string) {
		k := fmt.Sprintf("%s@%d>%s|%s", from.at.key(), from.pc, t.key(), cond)
		c := MethodID{creator.fqcn, creator.name, creator.desc}
		if i, ok := idx[k]; ok {
			if !hasCreator(out[i].Creators, c) {
				out[i].Creators = append(out[i].Creators, c)
			}
			return
		}
		idx[k] = len(out)
		out = append(out, GatedEdge{CallerFQCN: from.at.fqcn, CallerName: from.at.name, CallerDesc: from.at.desc, PC: from.pc, Kind: from.kind,
			TargetFQCN: t.fqcn, TargetMethod: t.name, TargetMethodDesc: t.desc, Creators: []MethodID{c}, Condition: cond})
	}
	subscriber := func(name, desc string) classfile.Call {
		return classfile.Call{Kind: classfile.CallInterface, Owner: "java.util.concurrent.Flow$Subscriber", Name: name, Desc: desc}
	}
	consumer := classfile.Call{Kind: classfile.CallInterface, Owner: "java.util.function.Consumer", Name: "accept", Desc: onNextDesc}
	for _, s := range subs {
		for _, o := range s.objs {
			if s.consume {
				if t, ok := runsListener(h, o.rc, consumer); ok {
					for _, e := range emits[emitItem] {
						if samePublisher(s.pubs, e.pubs) {
							add(e, t, o.from, ConditionPublished)
						}
					}
				}
				continue
			}
			if t, ok := runsListener(h, o.rc, subscriber("onSubscribe", onSubscribeDesc)); ok {
				add(s.rxSite, t, o.from, ConditionSubscribed)
			}
			for _, e := range emits[emitItem] {
				if !samePublisher(s.pubs, e.pubs) {
					continue
				}
				if t, ok := runsListener(h, o.rc, subscriber("onNext", onNextDesc)); ok {
					add(e, t, o.from, ConditionPublished)
				}
				if t, ok := runsListener(h, o.rc, subscriber("onError", onErrorDesc)); ok {
					add(e, t, o.from, ConditionExceptional)
				}
			}
			for _, e := range emits[emitClose] {
				if t, ok := runsListener(h, o.rc, subscriber("onComplete", "()V")); ok && samePublisher(s.pubs, e.pubs) {
					add(e, t, o.from, ConditionCompleted)
				}
			}
			for _, e := range emits[emitError] {
				if t, ok := runsListener(h, o.rc, subscriber("onError", onErrorDesc)); ok && samePublisher(s.pubs, e.pubs) {
					add(e, t, o.from, ConditionExceptional)
				}
			}
		}
	}
	for i := range out {
		sort.Slice(out[i].Creators, func(a, b int) bool {
			x, y := out[i].Creators[a], out[i].Creators[b]
			return x.FQCN+"#"+x.Name+x.Desc < y.FQCN+"#"+y.Name+y.Desc
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		ka := a.CallerFQCN + "#" + a.CallerName + a.CallerDesc + fmt.Sprintf("@%06d", a.PC) + a.TargetFQCN + "#" + a.TargetMethod + a.TargetMethodDesc + a.Condition
		kb := b.CallerFQCN + "#" + b.CallerName + b.CallerDesc + fmt.Sprintf("@%06d", b.PC) + b.TargetFQCN + "#" + b.TargetMethod + b.TargetMethodDesc + b.Condition
		return ka < kb
	})
	return out, stats, nil
}

// publishers keeps the objects that are SubmissionPublishers.
func publishers(h *hierarchy.Hierarchy, objs []objID, isPub func(string) bool) []objID {
	var out []objID
	for _, o := range objs {
		if isPub(o.typ) {
			out = append(out, o)
		}
	}
	return out
}

// samePublisher: a subscription and an emission share an object — the
// same `new` site, or for `this` (no site) the same class.
func samePublisher(a, b []objID) bool {
	for _, x := range a {
		for _, y := range b {
			if x.site != "" && y.site != "" {
				if x.site == y.site {
					return true
				}
				continue
			}
			if x.typ == y.typ {
				return true
			}
		}
	}
	return false
}
