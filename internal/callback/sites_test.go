package callback

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func callerMethod(name string, calls ...classfile.Call) hierarchy.Method {
	return hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: "()V"}, Calls: calls}
}

// TestFindRegistrationSites_threadConstructor is A0's own positive
// example for the Thread(Runnable) overload.
func TestFindRegistrationSites_threadConstructor(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			callerMethod("m", classfile.Call{
				Kind: classfile.CallSpecial, Owner: fqcnThread, Name: "<init>",
				Desc: "(Ljava/lang/Runnable;)V", PC: 3,
			}),
		}},
	})

	sites := FindRegistrationSites(h)
	if len(sites) != 1 {
		t.Fatalf("sites = %+v, want exactly 1", sites)
	}
	s := sites[0]
	if s.Outcome != OutcomeCataloged {
		t.Errorf("Outcome = %q, want %q", s.Outcome, OutcomeCataloged)
	}
	if s.Model.CallbackInterface != fqcnRunnable || s.Model.CallbackMethod != "run" {
		t.Errorf("Model = %+v, want CallbackInterface=%q CallbackMethod=run", s.Model, fqcnRunnable)
	}
	if s.CallerFQCN != "test.Caller" || s.CallerName != "m" || s.PC != 3 {
		t.Errorf("site identity = %+v, want test.Caller#m pc=3", s)
	}
}

// TestFindRegistrationSites_executorSubmitCallable is A0's own positive
// example for the Callable overload (distinct CallbackInterface/Method
// from the Runnable overloads).
func TestFindRegistrationSites_executorSubmitCallable(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			callerMethod("m", classfile.Call{
				Kind: classfile.CallInterface, Owner: fqcnExecutorService, Name: "submit",
				Desc: "(Ljava/util/concurrent/Callable;)Ljava/util/concurrent/Future;", PC: 10,
			}),
		}},
	})

	sites := FindRegistrationSites(h)
	if len(sites) != 1 || sites[0].Model.CallbackInterface != fqcnCallable || sites[0].Model.CallbackMethod != "call" {
		t.Fatalf("sites = %+v, want one site with CallbackInterface=%q CallbackMethod=call", sites, fqcnCallable)
	}
}

// TestFindRegistrationSites_uncatalogedOverloadIsFlagged is A0's own
// negative example: a call to a known API's *name*, with a descriptor
// DefaultCatalog does not cover, must be reported as a diagnostic — not
// silently ignored, and not misidentified as a resolved registration.
func TestFindRegistrationSites_uncatalogedOverloadIsFlagged(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			callerMethod("m", classfile.Call{
				// ExecutorService#submit(Runnable, Object) — a real
				// overload, deliberately not in DefaultCatalog.
				Kind: classfile.CallInterface, Owner: fqcnExecutorService, Name: "submit",
				Desc: "(Ljava/lang/Runnable;Ljava/lang/Object;)Ljava/util/concurrent/Future;", PC: 7,
			}),
		}},
	})

	sites := FindRegistrationSites(h)
	if len(sites) != 1 {
		t.Fatalf("sites = %+v, want exactly 1", sites)
	}
	if sites[0].Outcome != OutcomeUnsupportedAPI {
		t.Errorf("Outcome = %q, want %q", sites[0].Outcome, OutcomeUnsupportedAPI)
	}
	if sites[0].Model != (RegistrationModel{}) {
		t.Errorf("Model = %+v, want the zero value for an unsupported overload", sites[0].Model)
	}
}

// TestFindRegistrationSites_unrelatedCallIsIgnored: a call that matches
// neither the exact catalog nor any known owner+name must not appear at
// all — never a false positive diagnostic.
func TestFindRegistrationSites_unrelatedCallIsIgnored(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			callerMethod("m", classfile.Call{
				Kind: classfile.CallVirtual, Owner: "java.lang.String", Name: "valueOf", Desc: "(I)Ljava/lang/String;", PC: 1,
			}),
		}},
	})

	if sites := FindRegistrationSites(h); len(sites) != 0 {
		t.Errorf("sites = %+v, want none", sites)
	}
}

// TestFindRegistrationSites_methodReferenceImplIsNotACallSite is a
// regression test for a real false positive found running against
// petclinic: a method reference used as a value (e.g. `Thread::new`
// passed as a ThreadFactory) is recorded as a classfile.CallDynamic
// entry whose Owner/Name/Desc name the reference's *resolved
// implementation* — here, coincidentally identical to this package's
// own Thread(Runnable) catalog entry — not an actual invocation
// matching that shape. Must never be treated as a registration call.
func TestFindRegistrationSites_methodReferenceImplIsNotACallSite(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			callerMethod("m", classfile.Call{
				Kind: classfile.CallDynamic, Owner: fqcnThread, Name: "<init>",
				Desc: "(Ljava/lang/Runnable;)V", PC: 2,
			}),
		}},
	})
	if sites := FindRegistrationSites(h); len(sites) != 0 {
		t.Errorf("sites = %+v, want none (a CallDynamic entry names a lambda impl, not a call site)", sites)
	}
}

func TestLookup_andLooksLikeUncatalogedRegistration(t *testing.T) {
	if _, ok := Lookup(fqcnThread, "<init>", "(Ljava/lang/Runnable;)V"); !ok {
		t.Errorf("Lookup(Thread(Runnable)) = not found, want found")
	}
	if !LooksLikeUncatalogedRegistration(fqcnExecutorService, "submit", "(Ljava/lang/Runnable;Ljava/lang/Object;)Ljava/util/concurrent/Future;") {
		t.Errorf("LooksLikeUncatalogedRegistration(submit(Runnable,Object)) = false, want true")
	}
	if LooksLikeUncatalogedRegistration("java.lang.String", "valueOf", "(I)Ljava/lang/String;") {
		t.Errorf("LooksLikeUncatalogedRegistration(String#valueOf) = true, want false")
	}
	// Thread has several constructor overloads that take no Runnable at
	// all (used by a Thread subclass's own super() chain, among other
	// things) — these must not be flagged, or every Thread subclass's
	// constructor would get a misleading diagnostic.
	for _, desc := range []string{"()V", "(Ljava/lang/String;)V", "(Ljava/lang/ThreadGroup;Ljava/lang/String;)V"} {
		if LooksLikeUncatalogedRegistration(fqcnThread, "<init>", desc) {
			t.Errorf("LooksLikeUncatalogedRegistration(Thread%s) = true, want false (no Runnable parameter)", desc)
		}
	}
}

// TestFindRegistrationSites_threadSubclassSuperCallIsNotFlagged is the
// same false-positive case TestLookup_andLooksLikeUncatalogedRegistration
// checks directly, exercised through the actual inventory pass.
func TestFindRegistrationSites_threadSubclassSuperCallIsNotFlagged(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.MyThread", SuperClass: fqcnThread, Methods: []hierarchy.Method{
			callerMethod("<init>", classfile.Call{
				Kind: classfile.CallSpecial, Owner: fqcnThread, Name: "<init>", Desc: "(Ljava/lang/String;)V", PC: 1,
			}),
		}},
	})
	if sites := FindRegistrationSites(h); len(sites) != 0 {
		t.Errorf("sites = %+v, want none (super(String), no Runnable involved)", sites)
	}
}

// TestCompletionCatalog (ASYNC_CALLBACKS A3): the CompletionStage
// continuations are cataloged on both owners a call may name, with the
// callback's index, interface method, execution and condition.
func TestCompletionCatalog(t *testing.T) {
	for _, c := range []struct {
		owner, name, desc string
		idx               int
		method, mdesc     string
		exec              ExecutionKind
		cond              string
	}{
		{"java.util.concurrent.CompletableFuture", "thenApply", "(Ljava/util/function/Function;)Ljava/util/concurrent/CompletableFuture;", 0, "apply", "(Ljava/lang/Object;)Ljava/lang/Object;", ExecutionImmediateOrAsync, ""},
		{"java.util.concurrent.CompletionStage", "thenAcceptAsync", "(Ljava/util/function/Consumer;Ljava/util/concurrent/Executor;)Ljava/util/concurrent/CompletionStage;", 0, "accept", "(Ljava/lang/Object;)V", ExecutionAsyncMaybe, ""},
		{"java.util.concurrent.CompletableFuture", "thenCombine", "(Ljava/util/concurrent/CompletionStage;Ljava/util/function/BiFunction;)Ljava/util/concurrent/CompletableFuture;", 1, "apply", "(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;", ExecutionImmediateOrAsync, ""},
		{"java.util.concurrent.CompletionStage", "exceptionally", "(Ljava/util/function/Function;)Ljava/util/concurrent/CompletionStage;", 0, "apply", "(Ljava/lang/Object;)Ljava/lang/Object;", ExecutionImmediateOrAsync, ConditionExceptional},
		{"java.util.concurrent.CompletableFuture", "exceptionallyAsync", "(Ljava/util/function/Function;Ljava/util/concurrent/Executor;)Ljava/util/concurrent/CompletableFuture;", 0, "apply", "(Ljava/lang/Object;)Ljava/lang/Object;", ExecutionAsyncMaybe, ConditionExceptional},
		{"java.util.concurrent.CompletableFuture", "whenComplete", "(Ljava/util/function/BiConsumer;)Ljava/util/concurrent/CompletableFuture;", 0, "accept", "(Ljava/lang/Object;Ljava/lang/Object;)V", ExecutionImmediateOrAsync, ""},
		{"java.util.concurrent.CompletableFuture", "supplyAsync", "(Ljava/util/function/Supplier;)Ljava/util/concurrent/CompletableFuture;", 0, "get", "()Ljava/lang/Object;", ExecutionAsyncMaybe, ""},
		{"java.util.concurrent.CompletableFuture", "runAsync", "(Ljava/lang/Runnable;Ljava/util/concurrent/Executor;)Ljava/util/concurrent/CompletableFuture;", 0, "run", "()V", ExecutionAsyncMaybe, ""},
	} {
		m, ok := Lookup(c.owner, c.name, c.desc)
		if !ok {
			t.Errorf("%s#%s%s not cataloged", c.owner, c.name, c.desc)
			continue
		}
		if m.CallbackArgIndex != c.idx || m.CallbackMethod != c.method || m.CallbackMethodDesc != c.mdesc || m.Execution != c.exec || m.Condition != c.cond {
			t.Errorf("%s#%s%s = %+v", c.owner, c.name, c.desc, m)
		}
	}
	// exceptionallyAsync is CompletableFuture's own (Java 12+), not the interface's.
	if _, ok := Lookup("java.util.concurrent.CompletionStage", "exceptionallyAsync", "(Ljava/util/function/Function;)Ljava/util/concurrent/CompletionStage;"); ok {
		t.Error("CompletionStage#exceptionallyAsync cataloged")
	}
	n := 0
	for _, m := range DefaultCatalog {
		if m.Owner == "java.util.concurrent.CompletableFuture" || m.Owner == "java.util.concurrent.CompletionStage" {
			n++
		}
	}
	if n != len(completionCatalog) || n != 2*(12*3+1)+2+3+4 {
		t.Errorf("%d completion entries in DefaultCatalog, catalog has %d", n, len(completionCatalog))
	}
}
