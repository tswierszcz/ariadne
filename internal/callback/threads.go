package callback

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// ASYNC_CALLBACKS A4 (a4-1): two execution models A0–A3 left out.
//
// Thread subclasses: `new Worker().start()` where Worker extends Thread
// and overrides run() — no Runnable anywhere, so A0's catalog (Thread's
// Runnable-taking constructors) never sees it, and Thread#start is a JDK
// method CHA does not expand. A start() call whose receiver resolves
// (returnedResolver.receiverObjs), or is declared on a Thread subclass of
// the input (then: it and its concrete subclasses), gets an edge to the
// run() each such class runs, when that run() is the input's own (not
// Thread#run, which runs the target Runnable — A2's). Condition
// "started"; a plain edge (the object exists where start is called on it).
//
// Rejected execution: a ThreadPoolExecutor given a RejectedExecutionHandler
// (a constructor taking one, or setRejectedExecutionHandler) calls
// handler.rejectedExecution(task, executor) when execute/submit cannot
// queue the task (saturated, or shut down). An execute/submit call whose
// executor shares a `new` site with the registration gets an edge to the
// handler's rejectedExecution, condition "rejected", gated on the method
// registering the handler. Whether the pool ever saturates is not
// decided (a may-run edge); the JDK's own policies (AbortPolicy,
// CallerRunsPolicy…) are outside the input and need no edge.
const (
	// ThreadsVersion names the rules; a snapshot records it.
	ThreadsVersion = "a4-1"
	// ConditionStarted: the target runs once the thread is started.
	ConditionStarted = "started"
	// ConditionRejected: the target runs when the executor rejects a task.
	ConditionRejected = "rejected"
)

const (
	fqcnTPE        = "java.util.concurrent.ThreadPoolExecutor"
	rejectedDesc   = "(Ljava/lang/Runnable;Ljava/util/concurrent/ThreadPoolExecutor;)V"
	handlerDescArg = "Ljava/util/concurrent/RejectedExecutionHandler;"
)

// ThreadStats summarizes FindThreadModels for the run log.
type ThreadStats struct {
	// Starts: start() calls on a Thread subclass of the input (or a
	// receiver resolving to one); StartEdges: edges to their run().
	Starts, StartEdges int
	// Handlers: RejectedExecutionHandler registrations with a resolved
	// executor and handler; Submissions: execute/submit calls paired
	// with one.
	Handlers, Submissions int
}

// FindThreadModels returns the plain Thread-subclass edges and the gated
// rejection candidates.
func FindThreadModels(h *hierarchy.Hierarchy) ([]CallbackEdge, []GatedEdge, ThreadStats) {
	var (
		stats ThreadStats
		jar   classfile.JarClassReader
	)
	defer jar.Close()
	rs := ReturnedStats{}
	r := &returnedResolver{h: h, fw: newForwarder(h, &jar), stats: &rs, returns: map[string][]taskObj{}, returners: map[string]bool{}}

	isThread := func(owner string) bool {
		if owner == fqcnThread {
			return true
		}
		for _, c := range h.SuperclassChain(owner) {
			if ci, ok := h.Lookup(c); ok && ci.SuperClass == fqcnThread {
				return true
			}
		}
		return false
	}
	isTPE := func(owner string) bool {
		if owner == fqcnTPE {
			return true
		}
		for _, c := range h.SuperclassChain(owner) {
			if ci, ok := h.Lookup(c); ok && ci.SuperClass == fqcnTPE {
				return true
			}
		}
		return false
	}

	var starts []CallbackEdge
	seenStart := map[string]bool{}
	type reg struct {
		at   handoffMethod
		pcs  []objID
		objs []queued
	}
	var regs []reg
	type submission struct {
		at   handoffMethod
		pc   int
		kind classfile.CallKind
		objs []objID
	}
	var subs []submission

	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if ci.SourceJar == "" {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				at := handoffMethod{fqcn, m.Name, m.Descriptor}
				site := bytecodeir.Site{Method: fqcn + "#" + m.Name, Descriptor: m.Descriptor, PC: c.PC, Op: bytecodeir.OpCall}
				switch {
				case c.Name == "start" && c.Desc == "()V" && c.Kind == classfile.CallVirtual && isThread(c.Owner):
					objs, ok := r.siteReceiver(at, m, c.PC)
					var types []string
					if ok {
						for _, o := range objs {
							types = append(types, o.typ)
						}
					} else if c.Owner != fqcnThread {
						for _, o := range r.thisObjs(c.Owner) {
							types = append(types, o.typ)
						}
					} else {
						continue
					}
					stats.Starts++
					for _, t := range types {
						decl, ok := h.ResolveExact(t, "run", "()V")
						if !ok || decl == fqcnThread {
							continue
						}
						if dci, ok := h.Lookup(decl); !ok || dci.SourceJar == "" {
							continue
						}
						k := fmt.Sprintf("%s@%d>%s", at.key(), c.PC, decl)
						if seenStart[k] {
							continue
						}
						seenStart[k] = true
						stats.StartEdges++
						starts = append(starts, CallbackEdge{CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC,
							TargetFQCN: decl, TargetMethod: "run", TargetMethodDesc: "()V", Condition: ConditionStarted})
					}
				case isTPE(c.Owner) && (c.Name == "<init>" && strings.Contains(c.Desc, handlerDescArg) ||
					c.Name == "setRejectedExecutionHandler" && c.Desc == "("+handlerDescArg+")V"):
					facts := r.fw.analyze(fqcn, m)
					if facts == nil {
						continue
					}
					objs, ok := r.siteReceiver(at, m, c.PC)
					if !ok || len(objs) == 0 {
						continue
					}
					arg, ok := argumentAt(facts, site, handlerIndex(c.Desc))
					if !ok {
						continue
					}
					hs := queuedObjects(r.fw, queueStore{at: at, pc: c.PC, value: arg})
					if len(hs) == 0 {
						continue
					}
					stats.Handlers++
					regs = append(regs, reg{at: at, pcs: objs, objs: hs})
				case (c.Name == "execute" && c.Desc == "(Ljava/lang/Runnable;)V" || c.Name == "submit") &&
					(isTPE(c.Owner) || c.Owner == fqcnExecutor || c.Owner == fqcnExecutorService || c.Owner == "java.util.concurrent.AbstractExecutorService"):
					objs, ok := r.siteReceiver(at, m, c.PC)
					if !ok || len(objs) == 0 {
						continue
					}
					subs = append(subs, submission{at: at, pc: c.PC, kind: c.Kind, objs: objs})
				}
			}
		}
	}

	var gated []GatedEdge
	idx := map[string]int{}
	call := classfile.Call{Kind: classfile.CallInterface, Owner: "java.util.concurrent.RejectedExecutionHandler", Name: "rejectedExecution", Desc: rejectedDesc}
	paired := map[string]bool{}
	for _, rg := range regs {
		for _, s := range subs {
			if !samePublisher(rg.pcs, s.objs) {
				continue
			}
			paired[fmt.Sprintf("%s@%d", s.at.key(), s.pc)] = true
			for _, o := range rg.objs {
				t, ok := runsListener(h, o.rc, call)
				if !ok {
					continue
				}
				k := fmt.Sprintf("%s@%d>%s", s.at.key(), s.pc, t.key())
				c := MethodID{o.from.fqcn, o.from.name, o.from.desc}
				if i, ok := idx[k]; ok {
					gated[i].Creators = appendMethod(gated[i].Creators, c)
					continue
				}
				idx[k] = len(gated)
				gated = append(gated, GatedEdge{CallerFQCN: s.at.fqcn, CallerName: s.at.name, CallerDesc: s.at.desc, PC: s.pc, Kind: s.kind,
					TargetFQCN: t.fqcn, TargetMethod: t.name, TargetMethodDesc: t.desc, Creators: []MethodID{c}, Condition: ConditionRejected})
			}
		}
	}
	stats.Submissions = len(paired)
	sort.SliceStable(gated, func(i, j int) bool {
		a, b := gated[i], gated[j]
		return fmt.Sprintf("%s#%s%s@%06d>%s", a.CallerFQCN, a.CallerName, a.CallerDesc, a.PC, a.TargetFQCN) <
			fmt.Sprintf("%s#%s%s@%06d>%s", b.CallerFQCN, b.CallerName, b.CallerDesc, b.PC, b.TargetFQCN)
	})
	return starts, gated, stats
}

// handlerIndex: the RejectedExecutionHandler argument of a
// ThreadPoolExecutor constructor (always the last parameter) or of
// setRejectedExecutionHandler (the only one).
func handlerIndex(desc string) int {
	slots, err := classfile.ArgumentSlots(desc, true)
	if err != nil || len(slots) == 0 {
		return 0
	}
	return len(slots) - 1
}
