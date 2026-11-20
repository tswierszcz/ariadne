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

// TestReturnedTasks (ASYNC_CALLBACKS a3-4) on testdata/async-returned —
// Jetty's shape: Strategy#loop runs (runTask: task.run()) what
// Producer#produce returns; SelProducer returns what Endpoint#onSelected
// returns; Conn#onSelected returns one of its final task fields (an
// anonymous class, a lambda):
//   - runTask → Conn$1#run and the write lambda, gated on Conn#onSelected
//     (both ran);
//   - runTask → IdleTask#run, gated on Idle#onSelected (Idle is never
//     created; IdleTask#run did not run — CHA still reaches
//     Idle#onSelected, the gate is as precise as the graph);
//   - Conn's CloseTask, returned only to code that never runs it: no edge
//     (did not run);
//   - OtherStrategy (Jetty's ProduceConsume: reached through ExecStrategy
//     by CHA, never constructed) runs what its producer field returns, but
//     the field is set only from its constructor's callers — none: no
//     edge (receiver narrowing; plain CHA on Producer#produce would give
//     it Conn's tasks);
//   - ExecutorService#submit(factory.job()) → the Callable lambda, gated
//     on Factory#job (ran).
//
// observed.txt (scripts/async-returned-oracle.sh) is checked for each
// claim; the gate adds nothing from roots that reach the run site but no
// returning method.
func TestReturnedTasks(t *testing.T) {
	const fx = "../../testdata/async-returned/"
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
	for m, want := range map[string]bool{
		"ar.Conn$1#run()V": true, "ar.Conn#lambda$new$0()V": true, "ar.Factory#lambda$job$0()Ljava/lang/String;": true,
		"ar.Strategy#runTask(Ljava/lang/Runnable;)V": true,
		"ar.CloseTask#run()V":                        false, "ar.IdleTask#run()V": false, "ar.Idle#onSelected()Ljava/lang/Runnable;": false,
	} {
		if obs[m] != want {
			t.Errorf("observed.txt: %s ran=%v, want %v", m, obs[m], want)
		}
	}

	h, _, err := hierarchy.Build(fx+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	edges, stats, warnings := FindReturnedTasks(h, FindRegistrationSites(h))
	if len(warnings) != 0 {
		t.Errorf("warnings: %v", warnings)
	}
	var got []string
	for _, e := range edges {
		var cs []string
		for _, c := range e.Creators {
			cs = append(cs, c.FQCN+"#"+c.Name)
		}
		got = append(got, e.CallerFQCN+"#"+e.CallerName+" -> "+e.TargetFQCN+"#"+e.TargetMethod+" ["+e.Condition+"] "+strings.Join(cs, ","))
	}
	sort.Strings(got)
	want := []string{
		"ar.Main#main -> ar.Factory#lambda$job$0 [produced] ar.Factory#job",
		"ar.Strategy#runTask -> ar.Conn#lambda$new$0 [produced] ar.Conn#onSelected",
		"ar.Strategy#runTask -> ar.Conn$1#run [produced] ar.Conn#onSelected",
		"ar.Strategy#runTask -> ar.IdleTask#run [produced] ar.Idle#onSelected",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("edges\n got %q\nwant %q", got, want)
	}
	// Sites: runTask's and OtherStrategy#loop's run(), Main's submit and
	// Thread(Runnable).
	if stats.Sites != 4 || stats.Resolved != 2 || stats.TooManyImpls != 0 || stats.Budget != 0 {
		t.Errorf("stats %+v", stats)
	}

	// The gate: from runTask alone no returning method is reachable.
	var cand []callgraph.SAMEdge
	for _, e := range edges {
		var cs []callgraph.MethodRef
		for _, c := range e.Creators {
			cs = append(cs, callgraph.MethodRef{FQCN: c.FQCN, Name: c.Name, Desc: c.Desc})
		}
		cand = append(cand, callgraph.SAMEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC, Kind: e.Kind,
			TargetFQCN: e.TargetFQCN, TargetName: e.TargetMethod, TargetDesc: e.TargetMethodDesc, Creators: cs, Condition: e.Condition})
	}
	id := func(g *callgraph.Graph, fqcn, name, desc string) uint32 {
		n, ok := g.ID(fqcn, name, desc)
		if !ok {
			t.Fatalf("no node %s#%s%s", fqcn, name, desc)
		}
		return n
	}
	g := callgraph.Build(h, nil)
	runTask := id(g, "ar.Strategy", "runTask", "(Ljava/lang/Runnable;)V")
	if st := callgraph.NewCreatorGate(g, h, cand, callgraph.BasisCallback).Apply(g, []uint32{runTask}); st.Added != 0 {
		t.Errorf("roots {runTask}: %+v, want nothing added", st)
	}
	// From main and the strategy thread's entry (the pipeline links
	// `new Thread(s::loop)` with A2/L1 edges this bare graph lacks).
	g = callgraph.Build(h, nil)
	st := callgraph.NewCreatorGate(g, h, cand, callgraph.BasisCallback).Apply(g, []uint32{
		id(g, "ar.Main", "main", "([Ljava/lang/String;)V"), id(g, "ar.Strategy", "loop", "()V")})
	if st.Added != 4 {
		t.Errorf("roots {main, Strategy#loop}: %+v, want 4 added", st)
	}
	if len(g.EdgesBetween(id(g, "ar.Strategy", "runTask", "(Ljava/lang/Runnable;)V"), id(g, "ar.Conn$1", "run", "()V"))) == 0 {
		t.Error("no edge runTask → Conn$1#run")
	}
}
