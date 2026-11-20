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

// TestThreadModels (ASYNC_CALLBACKS a4-1) on testdata/async-threads
// (scripts/async-threads-oracle.sh):
//   - Worker (a Thread subclass overriding run) started: start → Worker#run;
//     a Special held in a Worker variable: start → Special#run only (the
//     receiver resolves to the `new`); Idle, never started — nothing;
//   - a one-thread pool with no queue and a Reject handler: both execute
//     calls → Reject#rejectedExecution, gated on main (the second task was
//     rejected: Reject ran, Task did not); Unused, registered only in an
//     uncalled method on a pool nobody passes — nothing.
func TestThreadModels(t *testing.T) {
	const fx = "../../testdata/async-threads/"
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
		"th.Worker#run()V": true, "th.Special#run()V": true,
		"th.Reject#rejectedExecution(Ljava/lang/Runnable;Ljava/util/concurrent/ThreadPoolExecutor;)V": true,
		"th.Idle#run()V": false, "th.Task#run()V": false,
		"th.Unused#rejectedExecution(Ljava/lang/Runnable;Ljava/util/concurrent/ThreadPoolExecutor;)V": false,
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
	starts, gated, st := FindThreadModels(h)
	var got []string
	for _, e := range starts {
		got = append(got, "start "+e.TargetFQCN+"#"+e.TargetMethod+" ["+e.Condition+"]")
	}
	for _, e := range gated {
		got = append(got, "gated "+e.TargetFQCN+"#"+e.TargetMethod+" ["+e.Condition+"] "+e.Creators[0].FQCN+"#"+e.Creators[0].Name)
	}
	sort.Strings(got)
	want := []string{
		"gated th.Reject#rejectedExecution [rejected] th.Main#main",
		"gated th.Reject#rejectedExecution [rejected] th.Main#main",
		"start th.Special#run [started]",
		"start th.Worker#run [started]",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("edges\n got %q\nwant %q", got, want)
	}
	if st != (ThreadStats{Starts: 2, StartEdges: 2, Handlers: 1, Submissions: 2}) {
		t.Errorf("stats %+v", st)
	}

	g := callgraph.Build(h, nil)
	var pe []callgraph.CallbackEdge
	for _, e := range starts {
		pe = append(pe, callgraph.CallbackEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC,
			TargetFQCN: e.TargetFQCN, TargetMethod: e.TargetMethod, TargetMethodDesc: e.TargetMethodDesc, Condition: e.Condition})
	}
	callgraph.ApplyCallbackEdges(g, pe)
	var cand []callgraph.SAMEdge
	for _, e := range gated {
		var cs []callgraph.MethodRef
		for _, c := range e.Creators {
			cs = append(cs, callgraph.MethodRef{FQCN: c.FQCN, Name: c.Name, Desc: c.Desc})
		}
		cand = append(cand, callgraph.SAMEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC, Kind: e.Kind,
			TargetFQCN: e.TargetFQCN, TargetName: e.TargetMethod, TargetDesc: e.TargetMethodDesc, Creators: cs, Condition: e.Condition})
	}
	root, _ := g.ID("th.Main", "main", "([Ljava/lang/String;)V")
	callgraph.NewCreatorGate(g, h, cand, callgraph.BasisCallback).Apply(g, []uint32{root})
	reach := reachable(g, []uint32{root})
	for m, want := range ran {
		if m == "th.Task#run()V" {
			// Rejected on this run; the may-run edge from execute (A2)
			// stays — saturation is not decided.
			continue
		}
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
