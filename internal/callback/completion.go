package callback

import (
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Completion (ASYNC_CALLBACKS a3-7): a continuation runs when the stage it
// is registered on completes — normally (then*, applyToEither…), with an
// exception (exceptionally*), or either way (handle, whenComplete). A3
// slice 1 left every continuation a "may run" edge whatever happens to
// the stage. This slice decides it for the one case the input fully
// shows: a future created with `new CompletableFuture<>()` and completed
// by hand. Its continuations run only if a matching completion call on
// that future is reachable:
//
//   - normally: complete, completeAsync, obtrudeValue, completeOnTimeout;
//   - exceptionally: completeExceptionally, obtrudeException, cancel,
//     orTimeout.
//
// A completion call pairs with a continuation when their receivers share
// the `new` site (returnedResolver.receiverObjs: a local, a parameter's
// callers, a private/final field's writers). A completion call whose
// receiver does not resolve may complete any future, so it is a creator
// of every manual future's edges of its kind — never a reason to drop
// one. The continuation's edge, from its registration site as before, is
// then gated (callgraph.NewCreatorGate) on the methods making those
// calls; a manual future with no such call of the right kind anywhere
// (never completed, or only failed for a then*) gets no edge (counted).
// Every other continuation — on a stage from supplyAsync/runAsync,
// another continuation, a library, or a receiver that does not resolve —
// keeps A3 slice 1's plain edge.
const (
	// CompletionVersion names the rules; a snapshot records it.
	CompletionVersion = "a3-7"
)

// CompletionStats summarizes GateCompletions for the run log.
type CompletionStats struct {
	// Continuations: continuation edges examined; Manual: those on a
	// `new CompletableFuture` only; Gated: manual ones turned into gated
	// candidates; Dropped: manual ones no completion of their kind can
	// trigger; Completers: completion calls found (Unresolved of them
	// with a receiver that did not resolve) — the input is scanned for
	// them only when some continuation is on a manual future.
	Continuations, Manual, Gated, Dropped, Completers, Unresolved int
}

type outcome int

const (
	completesNormally outcome = 1 << iota
	completesExceptionally
)

func completionCall(c classfile.Call) (outcome, bool) {
	if c.Owner != fqcnCompletableFuture || (c.Kind != classfile.CallVirtual && c.Kind != classfile.CallInterface) {
		return 0, false
	}
	switch c.Name {
	case "complete", "completeAsync", "obtrudeValue", "completeOnTimeout":
		return completesNormally, true
	case "completeExceptionally", "obtrudeException", "cancel", "orTimeout":
		return completesExceptionally, true
	}
	return 0, false
}

// needs: the completions under which a continuation model runs.
func needs(m RegistrationModel) outcome {
	name := strings.TrimSuffix(m.Name, "Async")
	switch {
	case m.Condition == ConditionExceptional:
		return completesExceptionally
	case name == "handle" || name == "whenComplete":
		return completesNormally | completesExceptionally
	}
	return completesNormally
}

func isContinuation(m RegistrationModel) bool {
	if m.Owner != fqcnCompletableFuture && m.Owner != fqcnCompletionStage {
		return false
	}
	return m.Name != "supplyAsync" && m.Name != "runAsync"
}

// GateCompletions splits A2's edge results: plain edges (everything but a
// continuation on a manual future) and gated candidates for continuations
// on manual futures (see above). Results with a gap contribute nothing,
// as in EdgeTargets.
func GateCompletions(h *hierarchy.Hierarchy, results []CallbackEdgeResult) ([]CallbackEdge, []GatedEdge, CompletionStats) {
	var (
		stats CompletionStats
		jar   classfile.JarClassReader
	)
	defer jar.Close()
	rs := ReturnedStats{}
	r := &returnedResolver{h: h, fw: newForwarder(h, &jar), stats: &rs, returns: map[string][]taskObj{}, returners: map[string]bool{}}

	// Completion calls: by `new` site and outcome, the methods making
	// them; unresolved receivers by outcome.
	type key struct {
		site string
		o    outcome
	}
	completers := map[key][]MethodID{}
	unresolved := map[outcome][]MethodID{}
	scanned := false
	scan := func() {
		scanned = true
		for _, fqcn := range h.SortedFQCNs() {
			ci, _ := h.Lookup(fqcn)
			if ci.SourceJar == "" {
				continue
			}
			for _, m := range ci.Methods {
				for _, c := range m.Calls {
					o, ok := completionCall(c)
					if !ok {
						continue
					}
					stats.Completers++
					id := MethodID{fqcn, m.Name, m.Descriptor}
					at := handoffMethod{fqcn, m.Name, m.Descriptor}
					objs, ok := r.siteReceiver(at, m, c.PC)
					if !ok {
						stats.Unresolved++
						unresolved[o] = appendMethod(unresolved[o], id)
						continue
					}
					for _, obj := range objs {
						if obj.site != "" {
							k := key{obj.site, o}
							completers[k] = appendMethod(completers[k], id)
						}
					}
				}
			}
		}
	}

	var plain []CallbackEdge
	var gated []GatedEdge
	for _, res := range results {
		if res.Gap != "" {
			continue
		}
		s := res.ResolvedCallback.Site
		e := CallbackEdge{
			CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, PC: s.PC,
			TargetFQCN: res.DeclaringFQCN, TargetMethod: res.DeclaringMethod, TargetMethodDesc: res.DeclaringDesc,
			Condition: s.Model.Condition,
		}
		if !isContinuation(s.Model) {
			plain = append(plain, e)
			continue
		}
		stats.Continuations++
		ci, ok := h.Lookup(s.CallerFQCN)
		var m hierarchy.Method
		if ok {
			m, ok = ci.MethodByNameAndDesc(s.CallerName, s.CallerDesc)
		}
		var objs []objID
		if ok {
			objs, ok = r.siteReceiver(handoffMethod{s.CallerFQCN, s.CallerName, s.CallerDesc}, m, s.PC)
		}
		if !ok || !allManual(objs) {
			plain = append(plain, e)
			continue
		}
		if !scanned {
			scan()
		}
		stats.Manual++
		need := needs(s.Model)
		var creators []MethodID
		for _, o := range []outcome{completesNormally, completesExceptionally} {
			if need&o == 0 {
				continue
			}
			for _, obj := range objs {
				for _, id := range completers[key{obj.site, o}] {
					creators = appendMethod(creators, id)
				}
			}
			for _, id := range unresolved[o] {
				creators = appendMethod(creators, id)
			}
		}
		if len(creators) == 0 {
			stats.Dropped++
			continue
		}
		stats.Gated++
		gated = append(gated, GatedEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC,
			Kind: classfile.CallSynthetic, TargetFQCN: e.TargetFQCN, TargetMethod: e.TargetMethod, TargetMethodDesc: e.TargetMethodDesc,
			Creators: creators, Condition: e.Condition})
	}
	return plain, gated, stats
}

// allManual: every object is a `new CompletableFuture` (a known site).
func allManual(objs []objID) bool {
	if len(objs) == 0 {
		return false
	}
	for _, o := range objs {
		if o.typ != fqcnCompletableFuture || o.site == "" {
			return false
		}
	}
	return true
}

func appendMethod(ms []MethodID, m MethodID) []MethodID {
	for _, x := range ms {
		if x == m {
			return ms
		}
	}
	return append(ms, m)
}

// siteReceiver: receiverObjs of the receiver of the call at pc in m.
func (r *returnedResolver) siteReceiver(at handoffMethod, m hierarchy.Method, pc int) ([]objID, bool) {
	facts := r.fw.analyze(at.fqcn, m)
	if facts == nil {
		return nil, false
	}
	v := receiverAt(facts, bytecodeir.Site{Method: at.fqcn + "#" + at.name, Descriptor: at.desc, PC: pc, Op: bytecodeir.OpCall})
	r.work, r.exhausted = 0, false
	return r.receiverObjs(&v, at, 0)
}
