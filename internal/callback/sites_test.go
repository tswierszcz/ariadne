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
