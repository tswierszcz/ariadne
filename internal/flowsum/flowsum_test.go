package flowsum

import (
	"sort"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
)

const fxJar = "../../testdata/flowsum/fs.jar"

func fixture(t *testing.T, opts Options) (*Engine, *hierarchy.Hierarchy) {
	t.Helper()
	h, _, err := hierarchy.Build(fxJar, nil)
	if err != nil {
		t.Fatal(err)
	}
	e := New(h, opts)
	t.Cleanup(func() { e.Close() })
	var roots []MethodKey
	ci, _ := h.Lookup("fs.Cases")
	for _, m := range ci.Methods {
		roots = append(roots, MethodKey{"fs.Cases", m.Name, m.Descriptor})
	}
	e.Summarize(roots)
	return e, h
}

func sum(t *testing.T, e *Engine, owner, name string) *Summary {
	t.Helper()
	for k, s := range e.sums {
		if k.Owner == owner && k.Name == name {
			return s
		}
	}
	t.Fatalf("no summary of %s#%s", owner, name)
	return nil
}

func deps(d Dep) string { return d.String() }

// callArg returns the deps of argument i of the n-th call (in PC order)
// of method s.
func callArg(t *testing.T, s *Summary, n, i int) Dep {
	t.Helper()
	var pcs []int
	for pc := range s.Calls {
		pcs = append(pcs, pc)
	}
	sort.Ints(pcs)
	if n >= len(pcs) || i >= len(s.Calls[pcs[n]].Args) {
		t.Fatalf("%s: no call %d arg %d", s.Method, n, i)
	}
	return s.Calls[pcs[n]].Args[i]
}

// TestSummaries: the D3 shapes on real bytecode (testdata/flowsum,
// scripts/build-flowsum-fixture.sh). Positive: identity, catalog
// transformation, constructor + getter, CHA single target, stream
// wrapping, mutation through a user method, recursion, an exception
// handler, a static field. Negative: a constant depends on nothing; two
// call sites of one helper do not mix. Conservative: an external method
// and two CHA targets are widened, with an unmodeled-call source.
func TestSummaries(t *testing.T) {
	e, _ := fixture(t, DefaultOptions)
	for _, c := range []struct{ name, ret string }{
		{"id", "param:0"},
		{"decode", "param:0"},         // Base64 decoder: the argument, no new source
		{"constant", ""},              // a literal
		{"concat", "param:0 param:1"}, // invokedynamic string concatenation
		{"holder", "field:fs.Holder.v param:0"},
		{"single", "param:1"}, // the one Codec implementation, through CHA
		{"read", "param:0"},   // ObjectInputStream(ByteArrayInputStream(data)).readObject()
		{"readDecoded", "param:0"},
		{"filled", "param:0"}, // fill(sb, p) appends p into sb
		{"rec", "param:0"},    // recursion converged
		{"guarded", "param:0"},
		{"load", "field:fs.Cases.shared"},
		{"external", "call:ext.Lib#thing(Ljava/lang/Object;)Ljava/lang/Object; param:0"},
		{"multi", "call:fs.Shape#name(Ljava/lang/String;)Ljava/lang/String; param:0 param:1"},
	} {
		s := sum(t, e, "fs.Cases", c.name)
		if deps(s.Return) != c.ret {
			t.Errorf("%s: return depends on %q, want %q", c.name, deps(s.Return), c.ret)
		}
		if len(s.Incomplete) > 0 {
			t.Errorf("%s: incomplete %v", c.name, s.Incomplete)
		}
	}
	// Widening is recorded with its reason, only where it happened.
	if w := sum(t, e, "fs.Cases", "external").Widened; len(w) != 1 || !strings.Contains(w[0], "target not in the input") {
		t.Errorf("external widened %v", w)
	}
	if w := sum(t, e, "fs.Cases", "multi").Widened; len(w) != 1 || !strings.Contains(w[0], "2 implementations (CHA)") {
		t.Errorf("multi widened %v", w)
	}
	for _, n := range []string{"id", "decode", "single", "read", "filled", "rec", "holder"} {
		if w := sum(t, e, "fs.Cases", n).Widened; len(w) > 0 {
			t.Errorf("%s widened %v", n, w)
		}
	}
	// Two call sites of id: sink(y) gets b only, sink(x) a only.
	two := sum(t, e, "fs.Cases", "twoSites")
	if got := deps(callArg(t, two, 2, 0)); got != "param:1" {
		t.Errorf("sink(id(b)) depends on %q, want param:1 only", got)
	}
	if got := deps(callArg(t, two, 3, 0)); got != "param:0" {
		t.Errorf("sink(id(a)) depends on %q, want param:0 only", got)
	}
	// The stream read at the entry depends on the parameter.
	rd := sum(t, e, "fs.Cases", "read")
	var pcs []int
	for pc := range rd.Calls {
		pcs = append(pcs, pc)
	}
	sort.Ints(pcs)
	if got := deps(rd.Calls[pcs[len(pcs)-1]].Receiver); got != "param:0" {
		t.Errorf("readObject receiver depends on %q", got)
	}
	// Field and object effects.
	if got := deps(sum(t, e, "fs.Cases", "store").Fields["fs.Cases.shared"]); got != "param:0" {
		t.Errorf("store writes %q", got)
	}
	if got := deps(sum(t, e, "fs.Holder", "<init>").IntoThis); got != "param:0" {
		t.Errorf("Holder(v) into this %q", got)
	}
	if got := deps(sum(t, e, "fs.Cases", "fill").IntoParams[0]); got != "param:1" {
		t.Errorf("fill into sb %q", got)
	}
}

// TestBudget: past MaxMethods a callee is answered conservatively and
// the widening names the budget — never dropped.
func TestBudget(t *testing.T) {
	h, _, err := hierarchy.Build(fxJar, nil)
	if err != nil {
		t.Fatal(err)
	}
	e := New(h, Options{MaxMethods: 1, MaxRounds: 5})
	defer e.Close()
	st := e.Summarize([]MethodKey{{"fs.Cases", "readDecoded", "(Ljava/lang/String;)Ljava/lang/Object;"}})
	s := e.Summary(MethodKey{"fs.Cases", "readDecoded", "(Ljava/lang/String;)Ljava/lang/Object;"})
	if st.Methods != 1 || st.BudgetCut != 2 || len(s.Widened) != 2 || !strings.Contains(s.Widened[0], "summary budget") {
		t.Fatalf("stats %+v, widened %v", st, s.Widened)
	}
	if !s.Return.has("param:0") || len(s.Return.Calls()) != 2 {
		t.Errorf("budget-cut result %v: still depends on the argument, plus both unmodeled calls", s.Return)
	}
}

// TestRecursionLimit: a fixed point not reached within MaxRounds marks
// every summary incomplete.
func TestRecursionLimit(t *testing.T) {
	h, _, err := hierarchy.Build(fxJar, nil)
	if err != nil {
		t.Fatal(err)
	}
	e := New(h, Options{MaxMethods: 100, MaxRounds: 1})
	defer e.Close()
	st := e.Summarize([]MethodKey{{"fs.Cases", "rec", "(Ljava/lang/String;I)Ljava/lang/String;"}})
	s := e.Summary(MethodKey{"fs.Cases", "rec", "(Ljava/lang/String;I)Ljava/lang/String;"})
	if st.Converged || len(s.Incomplete) == 0 || !strings.Contains(s.Incomplete[0], "did not converge") {
		t.Errorf("stats %+v incomplete %v", st, s.Incomplete)
	}
}
