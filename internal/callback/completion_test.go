package callback

import (
	"bufio"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// TestCompletion (ASYNC_CALLBACKS a3-7) on testdata/async-completion —
// futures created with new CompletableFuture and completed by hand
// (scripts/async-completion-oracle.sh):
//   - ok (complete): thenApply(Upper) gated on Main#main (ran);
//     exceptionally(Recover) dropped — nothing fails ok (did not run);
//   - never: thenAccept(Never) dropped (did not run);
//   - failing (completeExceptionally): exceptionally(Handled) and
//     whenComplete(Both) gated (ran); thenApply(Skipped) dropped (did not);
//   - later, completed by Api#finish it is handed to: thenRun(Later)
//     gated on Api#finish (ran);
//   - Holder's final-field future, completed only by the uncalled
//     Holder#done: thenRun(Pending) gated on it — left out (did not run);
//   - runAsync(Source): a source, a plain edge as before (ran).
func TestCompletion(t *testing.T) {
	const fx = "../../testdata/async-completion/"
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
		"cf.Upper#apply(Ljava/lang/Object;)Ljava/lang/Object;": true, "cf.Handled#apply(Ljava/lang/Throwable;)Ljava/lang/Object;": true,
		"cf.Both#accept(Ljava/lang/Object;Ljava/lang/Throwable;)V": true, "cf.Later#run()V": true, "cf.Source#run()V": true,
		"cf.Recover#apply(Ljava/lang/Throwable;)Ljava/lang/Object;": false, "cf.Never#accept(Ljava/lang/Object;)V": false,
		"cf.Skipped#apply(Ljava/lang/Object;)Ljava/lang/Object;": false, "cf.Pending#run()V": false,
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
	resolved, _ := ResolveCallbacks(h, FindRegistrationSites(h))
	plain, gated, stats := GateCompletions(h, FindCallbackEdges(h, resolved))
	var got []string
	for _, e := range plain {
		got = append(got, "plain "+e.TargetFQCN+"#"+e.TargetMethod)
	}
	for _, e := range gated {
		var cs []string
		for _, c := range e.Creators {
			cs = append(cs, c.FQCN+"#"+c.Name)
		}
		sort.Strings(cs)
		got = append(got, "gated "+e.TargetFQCN+"#"+e.TargetMethod+" ["+e.Condition+"] "+strings.Join(cs, ","))
	}
	sort.Strings(got)
	want := []string{
		"gated cf.Both#accept [] cf.Main#main",
		"gated cf.Handled#apply [exceptional] cf.Main#main",
		"gated cf.Later#run [] cf.Api#finish",
		"gated cf.Pending#run [] cf.Holder#done",
		"gated cf.Upper#apply [] cf.Main#main",
		"plain cf.Source#run",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("edges\n got %q\nwant %q", got, want)
	}
	if stats.Continuations != 8 || stats.Manual != 8 || stats.Gated != 5 || stats.Dropped != 3 || stats.Unresolved != 0 {
		t.Errorf("stats %+v", stats)
	}

	var cand []callgraph.SAMEdge
	for _, e := range gated {
		var cs []callgraph.MethodRef
		for _, c := range e.Creators {
			cs = append(cs, callgraph.MethodRef{FQCN: c.FQCN, Name: c.Name, Desc: c.Desc})
		}
		cand = append(cand, callgraph.SAMEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC, Kind: e.Kind,
			TargetFQCN: e.TargetFQCN, TargetName: e.TargetMethod, TargetDesc: e.TargetMethodDesc, Creators: cs, Condition: e.Condition})
	}
	g := callgraph.Build(h, nil)
	var pe []callgraph.CallbackEdge
	for _, e := range plain {
		pe = append(pe, callgraph.CallbackEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC,
			TargetFQCN: e.TargetFQCN, TargetMethod: e.TargetMethod, TargetMethodDesc: e.TargetMethodDesc, Condition: e.Condition})
	}
	callgraph.ApplyCallbackEdges(g, pe)
	root, _ := g.ID("cf.Main", "main", "([Ljava/lang/String;)V")
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
