// Package callback is ASYNC_CALLBACKS A0's own contract and initial
// catalog: a callback-registration API is not an ordinary call — the
// method it actually invokes (Runnable#run, Callable#call, ...) is
// named by an argument's runtime type, not by the registration call's
// own symbolic target, and may run later, on another thread, or never
// at all. A0 defines the model and a small, hand-picked initial catalog
// (JDK Thread/Executor only — "wybrać początkowy katalog JDK i
// konkretne overloads", not a generic reflection-discovered list) plus
// the inventory pass that finds matching call sites. Resolving *which*
// concrete type the callback argument actually is (ASYNC_CALLBACKS A1)
// and turning a resolved site into a real call-graph edge (A2) are
// later, separate phases — same staged approach SERVICE_LOADING's
// S0/S1/S2 already used.
package callback

import "strings"

// ExecutionKind says how certainly/synchronously a registered callback
// actually runs — ASYNC_CALLBACKS A0's own "charakter wykonania", kept
// explicit so a later consumer never mistakes a callback edge for an
// ordinary synchronous call (A2's own requirement not to invent a
// thread identity the static analysis can't see).
type ExecutionKind string

const (
	// ExecutionAsyncMaybe: the callback may run later, possibly on
	// another thread, and may never run at all (e.g. the thread is
	// never started, or a submitted task is cancelled first).
	ExecutionAsyncMaybe ExecutionKind = "async-maybe"
	// ExecutionImmediateOrAsync: some implementations of this API (most
	// notably a "direct"/same-thread Executor) may run the callback
	// synchronously inside the registration call itself, so "possibly
	// immediate" must stay part of the model, not just "possibly later".
	ExecutionImmediateOrAsync ExecutionKind = "immediate-or-async"
	// ExecutionSynchronous: the API runs the callback on the calling
	// thread before it returns, and returns the callback's own result
	// (AccessController.doPrivileged and its run() action).
	ExecutionSynchronous ExecutionKind = "synchronous"
)

// RegistrationModel is one callback-registration API ariadne knows
// about: which call registers a callback, which argument holds it, and
// which method on which interface that callback object will eventually
// be invoked through.
type RegistrationModel struct {
	// Owner/Name/Desc identify the registration call itself exactly as
	// classfile.Call would report it (e.g.
	// "java.util.concurrent.ExecutorService"/"execute"/
	// "(Ljava/lang/Runnable;)V").
	Owner, Name, Desc string
	// CallbackArgIndex is the 0-based argument index (not counting the
	// receiver) holding the callback object.
	CallbackArgIndex int
	// CallbackInterface is the functional/callback interface's FQCN
	// whose method the callback argument's real object dispatches
	// through (e.g. "java.lang.Runnable").
	CallbackInterface string
	// CallbackMethod/CallbackMethodDesc identify which method on
	// CallbackInterface actually gets invoked (e.g. "run", "()V").
	CallbackMethod, CallbackMethodDesc string
	Execution                          ExecutionKind
	// Condition is when the callback runs at all, beyond being
	// registered: "" for every completion (or not modeled),
	// ConditionExceptional for a CompletionStage branch that runs only
	// when the previous stage fails (exceptionally/exceptionallyCompose).
	// It is carried into the edge's path evidence.
	Condition string
}

// ConditionExceptional: the callback runs only when the stage it is
// registered on completes exceptionally (ASYNC_CALLBACKS A3).
const ConditionExceptional = "exceptional"

const (
	fqcnThread          = "java.lang.Thread"
	fqcnExecutor        = "java.util.concurrent.Executor"
	fqcnExecutorService = "java.util.concurrent.ExecutorService"
	fqcnRunnable        = "java.lang.Runnable"
	fqcnCallable        = "java.util.concurrent.Callable"
	fqcnAccessCtl       = "java.security.AccessController"
	fqcnPrivAction      = "java.security.PrivilegedAction"
	fqcnPrivExcAction   = "java.security.PrivilegedExceptionAction"
	// ASYNC_CALLBACKS A3.
	fqcnCompletableFuture = "java.util.concurrent.CompletableFuture"
	fqcnCompletionStage   = "java.util.concurrent.CompletionStage"
)

// DefaultCatalog is ASYNC_CALLBACKS A0's own initial JDK catalog: the
// Thread constructors that accept a Runnable, and Executor/
// ExecutorService's execute/submit — deliberately not exhaustive (no
// ScheduledExecutorService). ASYNC_CALLBACKS A3 adds completionCatalog
// (CompletionStage continuations).
var DefaultCatalog = append([]RegistrationModel{
	{Owner: fqcnThread, Name: "<init>", Desc: "(Ljava/lang/Runnable;)V",
		CallbackArgIndex: 0, CallbackInterface: fqcnRunnable, CallbackMethod: "run", CallbackMethodDesc: "()V",
		Execution: ExecutionAsyncMaybe},
	{Owner: fqcnThread, Name: "<init>", Desc: "(Ljava/lang/Runnable;Ljava/lang/String;)V",
		CallbackArgIndex: 0, CallbackInterface: fqcnRunnable, CallbackMethod: "run", CallbackMethodDesc: "()V",
		Execution: ExecutionAsyncMaybe},
	{Owner: fqcnThread, Name: "<init>", Desc: "(Ljava/lang/ThreadGroup;Ljava/lang/Runnable;)V",
		CallbackArgIndex: 1, CallbackInterface: fqcnRunnable, CallbackMethod: "run", CallbackMethodDesc: "()V",
		Execution: ExecutionAsyncMaybe},
	{Owner: fqcnThread, Name: "<init>", Desc: "(Ljava/lang/ThreadGroup;Ljava/lang/Runnable;Ljava/lang/String;)V",
		CallbackArgIndex: 1, CallbackInterface: fqcnRunnable, CallbackMethod: "run", CallbackMethodDesc: "()V",
		Execution: ExecutionAsyncMaybe},
	{Owner: fqcnExecutor, Name: "execute", Desc: "(Ljava/lang/Runnable;)V",
		CallbackArgIndex: 0, CallbackInterface: fqcnRunnable, CallbackMethod: "run", CallbackMethodDesc: "()V",
		Execution: ExecutionImmediateOrAsync},
	{Owner: fqcnExecutorService, Name: "execute", Desc: "(Ljava/lang/Runnable;)V",
		CallbackArgIndex: 0, CallbackInterface: fqcnRunnable, CallbackMethod: "run", CallbackMethodDesc: "()V",
		Execution: ExecutionImmediateOrAsync},
	{Owner: fqcnExecutorService, Name: "submit", Desc: "(Ljava/lang/Runnable;)Ljava/util/concurrent/Future;",
		CallbackArgIndex: 0, CallbackInterface: fqcnRunnable, CallbackMethod: "run", CallbackMethodDesc: "()V",
		Execution: ExecutionAsyncMaybe},
	{Owner: fqcnExecutorService, Name: "submit", Desc: "(Ljava/util/concurrent/Callable;)Ljava/util/concurrent/Future;",
		CallbackArgIndex: 0, CallbackInterface: fqcnCallable, CallbackMethod: "call", CallbackMethodDesc: "()Ljava/lang/Object;",
		Execution: ExecutionAsyncMaybe},
	// java.security.AccessController: every doPrivileged overload runs
	// action.run() synchronously and returns its result.
	{Owner: fqcnAccessCtl, Name: "doPrivileged", Desc: "(Ljava/security/PrivilegedAction;)Ljava/lang/Object;",
		CallbackArgIndex: 0, CallbackInterface: fqcnPrivAction, CallbackMethod: "run", CallbackMethodDesc: "()Ljava/lang/Object;",
		Execution: ExecutionSynchronous},
	{Owner: fqcnAccessCtl, Name: "doPrivileged", Desc: "(Ljava/security/PrivilegedAction;Ljava/security/AccessControlContext;)Ljava/lang/Object;",
		CallbackArgIndex: 0, CallbackInterface: fqcnPrivAction, CallbackMethod: "run", CallbackMethodDesc: "()Ljava/lang/Object;",
		Execution: ExecutionSynchronous},
	{Owner: fqcnAccessCtl, Name: "doPrivileged", Desc: "(Ljava/security/PrivilegedAction;Ljava/security/AccessControlContext;[Ljava/security/Permission;)Ljava/lang/Object;",
		CallbackArgIndex: 0, CallbackInterface: fqcnPrivAction, CallbackMethod: "run", CallbackMethodDesc: "()Ljava/lang/Object;",
		Execution: ExecutionSynchronous},
	{Owner: fqcnAccessCtl, Name: "doPrivileged", Desc: "(Ljava/security/PrivilegedExceptionAction;)Ljava/lang/Object;",
		CallbackArgIndex: 0, CallbackInterface: fqcnPrivExcAction, CallbackMethod: "run", CallbackMethodDesc: "()Ljava/lang/Object;",
		Execution: ExecutionSynchronous},
	{Owner: fqcnAccessCtl, Name: "doPrivileged", Desc: "(Ljava/security/PrivilegedExceptionAction;Ljava/security/AccessControlContext;)Ljava/lang/Object;",
		CallbackArgIndex: 0, CallbackInterface: fqcnPrivExcAction, CallbackMethod: "run", CallbackMethodDesc: "()Ljava/lang/Object;",
		Execution: ExecutionSynchronous},
	{Owner: fqcnAccessCtl, Name: "doPrivileged", Desc: "(Ljava/security/PrivilegedExceptionAction;Ljava/security/AccessControlContext;[Ljava/security/Permission;)Ljava/lang/Object;",
		CallbackArgIndex: 0, CallbackInterface: fqcnPrivExcAction, CallbackMethod: "run", CallbackMethodDesc: "()Ljava/lang/Object;",
		Execution: ExecutionSynchronous},
}, completionCatalog...)

// ASYNC_CALLBACKS A3 slice 1: CompletionStage continuations and
// CompletableFuture's async sources. Every method of CompletionStage that
// takes a callback, as declared on the interface and as overridden by
// CompletableFuture (the call's owner is whichever the bytecode names),
// plus CompletableFuture's exceptionally*Async/exceptionallyCompose* (Java
// 12+) and supplyAsync/runAsync. A plain continuation may run on the
// registering thread (the stage already complete) or on the completing
// one; an *Async one, and a source, on an executor.
var completionCatalog = func() []RegistrationModel {
	type cb struct{ iface, method, desc string }
	var (
		function   = cb{"java.util.function.Function", "apply", "(Ljava/lang/Object;)Ljava/lang/Object;"}
		consumer   = cb{"java.util.function.Consumer", "accept", "(Ljava/lang/Object;)V"}
		runnable   = cb{fqcnRunnable, "run", "()V"}
		biFunction = cb{"java.util.function.BiFunction", "apply", "(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;"}
		biConsumer = cb{"java.util.function.BiConsumer", "accept", "(Ljava/lang/Object;Ljava/lang/Object;)V"}
		supplier   = cb{"java.util.function.Supplier", "get", "()Ljava/lang/Object;"}
	)
	const stage = "Ljava/util/concurrent/CompletionStage;"
	const executor = "Ljava/util/concurrent/Executor;"
	type op struct {
		name       string
		other      bool // a CompletionStage argument before the callback
		cb         cb
		async      bool // also has *Async(cb) and *Async(cb, Executor)
		condition  string
		futureOnly bool // declared on CompletableFuture only
	}
	ops := []op{
		{name: "thenApply", cb: function, async: true},
		{name: "thenAccept", cb: consumer, async: true},
		{name: "thenRun", cb: runnable, async: true},
		{name: "thenCombine", other: true, cb: biFunction, async: true},
		{name: "thenAcceptBoth", other: true, cb: biConsumer, async: true},
		{name: "runAfterBoth", other: true, cb: runnable, async: true},
		{name: "applyToEither", other: true, cb: function, async: true},
		{name: "acceptEither", other: true, cb: consumer, async: true},
		{name: "runAfterEither", other: true, cb: runnable, async: true},
		{name: "thenCompose", cb: function, async: true},
		{name: "handle", cb: biFunction, async: true},
		{name: "whenComplete", cb: biConsumer, async: true},
		{name: "exceptionally", cb: function, condition: ConditionExceptional},
		{name: "exceptionallyAsync", cb: function, condition: ConditionExceptional, futureOnly: true},
		{name: "exceptionallyCompose", cb: function, async: true, condition: ConditionExceptional, futureOnly: true},
	}
	var out []RegistrationModel
	add := func(owner, ret, name string, params []string, idx int, c cb, exec ExecutionKind, cond string) {
		out = append(out, RegistrationModel{
			Owner: owner, Name: name, Desc: "(" + strings.Join(params, "") + ")" + ret,
			CallbackArgIndex: idx, CallbackInterface: c.iface, CallbackMethod: c.method, CallbackMethodDesc: c.desc,
			Execution: exec, Condition: cond,
		})
	}
	for _, o := range ops {
		cbDesc := "L" + strings.ReplaceAll(o.cb.iface, ".", "/") + ";"
		params := []string{cbDesc}
		idx := 0
		if o.other {
			params = []string{stage, cbDesc}
			idx = 1
		}
		owners := [][2]string{{fqcnCompletableFuture, "Ljava/util/concurrent/CompletableFuture;"}}
		if !o.futureOnly {
			owners = append(owners, [2]string{fqcnCompletionStage, stage})
		}
		for _, ow := range owners {
			if o.name == "exceptionallyAsync" {
				add(ow[0], ow[1], o.name, params, idx, o.cb, ExecutionAsyncMaybe, o.condition)
				add(ow[0], ow[1], o.name, append(append([]string{}, params...), executor), idx, o.cb, ExecutionAsyncMaybe, o.condition)
				continue
			}
			add(ow[0], ow[1], o.name, params, idx, o.cb, ExecutionImmediateOrAsync, o.condition)
			if o.async {
				add(ow[0], ow[1], o.name+"Async", params, idx, o.cb, ExecutionAsyncMaybe, o.condition)
				add(ow[0], ow[1], o.name+"Async", append(append([]string{}, params...), executor), idx, o.cb, ExecutionAsyncMaybe, o.condition)
			}
		}
	}
	const cf = "Ljava/util/concurrent/CompletableFuture;"
	sup := "Ljava/util/function/Supplier;"
	run := "Ljava/lang/Runnable;"
	add(fqcnCompletableFuture, cf, "supplyAsync", []string{sup}, 0, supplier, ExecutionAsyncMaybe, "")
	add(fqcnCompletableFuture, cf, "supplyAsync", []string{sup, executor}, 0, supplier, ExecutionAsyncMaybe, "")
	add(fqcnCompletableFuture, cf, "runAsync", []string{run}, 0, runnable, ExecutionAsyncMaybe, "")
	add(fqcnCompletableFuture, cf, "runAsync", []string{run, executor}, 0, runnable, ExecutionAsyncMaybe, "")
	return out
}()

var byKey = func() map[[3]string]RegistrationModel {
	m := make(map[[3]string]RegistrationModel, len(DefaultCatalog))
	for _, model := range DefaultCatalog {
		m[[3]string{model.Owner, model.Name, model.Desc}] = model
	}
	return m
}()

var knownOwnerName = func() map[[2]string]bool {
	m := make(map[[2]string]bool, len(DefaultCatalog))
	for _, model := range DefaultCatalog {
		m[[2]string{model.Owner, model.Name}] = true
	}
	return m
}()

// Lookup returns the cataloged model for an exact (owner, name, desc)
// call, if any.
func Lookup(owner, name, desc string) (RegistrationModel, bool) {
	m, ok := byKey[[3]string{owner, name, desc}]
	return m, ok
}

// LooksLikeUncatalogedRegistration reports whether a call to owner+name
// (which does not match DefaultCatalog exactly — callers check Lookup
// first) is still worth flagging as a probable, uncataloged callback
// registration: owner+name matches some cataloged API, *and* desc's own
// parameter list mentions one of the callback interfaces that API's
// cataloged overloads use. Checking the descriptor text, not just
// owner+name, matters in practice: java.lang.Thread has several
// constructor overloads that take no Runnable at all (Thread(),
// Thread(String), Thread(ThreadGroup, String), and a Thread subclass's
// own constructor chaining to one of those via super()) — flagging
// every one of those as "a registration API we don't model" would be a
// real, misleading false positive, not a cautious diagnostic.
func LooksLikeUncatalogedRegistration(owner, name, desc string) bool {
	if !knownOwnerName[[2]string{owner, name}] {
		return false
	}
	for _, iface := range []string{fqcnRunnable, fqcnCallable} {
		if callbackParamMentioned(desc, iface) {
			return true
		}
	}
	return false
}

// callbackParamMentioned reports whether desc's parameter section
// mentions ifaceFQCN as a field descriptor (e.g. "Ljava/lang/Runnable;")
// — a cheap substring check, not a full descriptor parse, which is
// precise enough here since false descriptor collisions (another type
// whose name happens to contain "Runnable;") are not a real concern for
// the handful of interfaces this package cares about.
func callbackParamMentioned(desc, ifaceFQCN string) bool {
	descriptor := "L" + strings.ReplaceAll(ifaceFQCN, ".", "/") + ";"
	return strings.Contains(desc, descriptor)
}
