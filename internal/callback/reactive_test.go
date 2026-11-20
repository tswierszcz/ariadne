package callback

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// TestReactive (ASYNC_CALLBACKS a3-6) on testdata/async-reactive — JDK
// Flow with real SubmissionPublishers (scripts/async-reactive-oracle.sh):
//   - pub: Printer, the Upper processor and Thrower subscribed, a consume
//     lambda; submit + close: their onNext/onComplete run, Upper's submit
//     on itself reaches Sink#onNext (processor chain), Thrower's onNext
//     throws and its onError runs (an emission also has an onError edge);
//   - quiet: subscribed and closed, never emits — Quiet#onNext has no
//     edge and did not run;
//   - bad: closeExceptionally — Failing gets onError only;
//   - Idle (never subscribed) and Later (subscribed only in an uncalled
//     method, its publisher a parameter nobody passes) — nothing.
//
// Checked: the edges that matter, and reachability from main after the
// gate against observed.txt for every subscriber's onNext.
func TestReactive(t *testing.T) {
	const fx = "../../testdata/async-reactive/"
	obs := map[string]bool{}
	f, err := os.Open(fx + "observed.txt")
	if err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		obs[sc.Text()] = true
	}
	f.Close()
	ran := map[string]bool{
		"rx.Printer#onNext(Ljava/lang/String;)V": true, "rx.Sink#onNext(Ljava/lang/String;)V": true,
		"rx.Thrower#onNext(Ljava/lang/String;)V": true, "rx.Thrower#onError(Ljava/lang/Throwable;)V": true,
		"rx.Upper#onNext(Ljava/lang/String;)V": true, "rx.Main#lambda$main$0(Ljava/lang/String;)V": true,
		"rx.Base#onComplete()V": true, "rx.Base#onSubscribe(Ljava/util/concurrent/Flow$Subscription;)V": true,
		"rx.Quiet#onNext(Ljava/lang/String;)V": false, "rx.Failing#onNext(Ljava/lang/String;)V": false,
		"rx.Idle#onNext(Ljava/lang/String;)V": false, "rx.Later#onNext(Ljava/lang/String;)V": false,
	}
	for m, want := range ran {
		if obs[m] != want {
			t.Errorf("observed.txt: %s ran=%v, want %v", m, obs[m], want)
		}
	}

	h, _, err := hierarchy.Build(fx+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	edges, stats, _ := FindReactive(h)
	got := map[string]bool{}
	for _, e := range edges {
		got[e.CallerFQCN+"#"+e.CallerName+" -> "+e.TargetFQCN+"#"+e.TargetMethod+" ["+e.Condition+"]"] = true
	}
	for _, want := range []string{
		"rx.Main#main -> rx.Printer#onNext [published]",
		"rx.Main#main -> rx.Upper#onNext [published]",
		"rx.Upper#onNext -> rx.Sink#onNext [published]",
		"rx.Main#main -> rx.Thrower#onError [exceptional]",
		"rx.Main#main -> rx.Main#lambda$main$0 [published]",
		"rx.Main#main -> rx.Upper#onComplete [completed]",
		"rx.Upper#onComplete -> rx.Base#onComplete [completed]",
	} {
		if !got[want] {
			t.Errorf("missing edge %s", want)
		}
	}
	for e := range got {
		for _, never := range []string{"Quiet#onNext", "Failing#onNext", "Idle#", "Later#"} {
			if strings.Contains(e, never) {
				t.Errorf("unexpected edge %s", e)
			}
		}
	}
	// Later's subscription: its publisher is a parameter nobody passes.
	if stats.Subscriptions != 7 || stats.Emissions != 7 || stats.Unresolved != 1 {
		t.Errorf("stats %+v", stats)
	}

	var cand []callgraph.SAMEdge
	for _, e := range edges {
		var cs []callgraph.MethodRef
		for _, c := range e.Creators {
			cs = append(cs, callgraph.MethodRef{FQCN: c.FQCN, Name: c.Name, Desc: c.Desc})
		}
		cand = append(cand, callgraph.SAMEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC, Kind: e.Kind,
			TargetFQCN: e.TargetFQCN, TargetName: e.TargetMethod, TargetDesc: e.TargetMethodDesc, Creators: cs, Condition: e.Condition})
	}
	g := callgraph.Build(h, nil)
	root, ok := g.ID("rx.Main", "main", "([Ljava/lang/String;)V")
	if !ok {
		t.Fatal("no main")
	}
	callgraph.NewCreatorGate(g, h, cand, callgraph.BasisCallback).Apply(g, []uint32{root})
	reach := reachable(g, []uint32{root})
	for m, want := range ran {
		i, j := strings.IndexByte(m, '#'), strings.IndexByte(m, '(')
		id, ok := g.ID(m[:i], m[i+1:j], m[j:])
		if !ok {
			t.Fatalf("no node %s", m)
		}
		if reach[id] != want {
			t.Errorf("%s reachable=%v, want %v (as on the JVM)", m, reach[id], want)
		}
	}
}
