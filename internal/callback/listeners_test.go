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

// TestListeners (ASYNC_CALLBACKS a3-5) on testdata/async-listeners:
//   - Bus (Consumer listeners — a JDK type CHA never expands — walked by
//     a for loop on a second thread): publish → A and the lambda, gated on
//     Main#main (both ran); → D, registered only in Api#later, never
//     called — gated out (did not run);
//   - IndexedBus (get(i) over a CopyOnWriteArrayList of the input's own
//     Listener type): publish → B (ran; CHA already has that edge — and
//     every other Listener's);
//   - SilentBus: registered, never published — the edge leaves an
//     unreachable method, C is not reached (did not run);
//   - U (a Consumer): created, never registered — nothing;
//   - Model (PropertyChangeSupport): setX's fire → View#propertyChange
//     (ran); QuietModel registers Unseen and never fires — nothing.
//
// observed.txt (scripts/async-listeners-oracle.sh) is checked for each
// claim, and reachability after the gate from main and the thread lambda.
func TestListeners(t *testing.T) {
	const fx = "../../testdata/async-listeners/"
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
		"al.A#accept(Ljava/lang/String;)V": true, "al.B#onEvent(Ljava/lang/String;)V": true,
		"al.Main#lambda$main$0(Ljava/lang/String;)V": true, "al.View#propertyChange(Ljava/beans/PropertyChangeEvent;)V": true,
		"al.C#accept(Ljava/lang/String;)V": false, "al.D#accept(Ljava/lang/String;)V": false,
		"al.U#accept(Ljava/lang/String;)V": false, "al.Unseen#propertyChange(Ljava/beans/PropertyChangeEvent;)V": false,
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
	r := ScanHandoffs(h)
	if len(r.Warnings) != 0 || len(r.Edges) != 0 {
		t.Errorf("hand-off edges %v, warnings %v (want none: nothing here is drained)", r.Edges, r.Warnings)
	}
	var got []string
	for _, e := range r.Listeners {
		var cs []string
		for _, c := range e.Creators {
			cs = append(cs, c.FQCN+"#"+c.Name)
		}
		got = append(got, e.CallerFQCN+"#"+e.CallerName+" -> "+e.TargetFQCN+"#"+e.TargetMethod+" ["+e.Condition+"] "+strings.Join(cs, ","))
	}
	sort.Strings(got)
	want := []string{
		"al.Bus#publish -> al.A#accept [notified] al.Main#main",
		"al.Bus#publish -> al.D#accept [notified] al.Api#later",
		"al.Bus#publish -> al.Main#lambda$main$0 [notified] al.Main#main",
		"al.IndexedBus#publish -> al.B#onEvent [notified] al.Main#main",
		"al.Model#setX -> al.View#propertyChange [notified] al.Main#main",
		"al.SilentBus#publish -> al.C#accept [notified] al.Main#main",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("edges\n got %q\nwant %q", got, want)
	}
	if s := r.ListenerStats; s.Broadcasts != 3 || s.Registrations != 4 || s.Fires != 1 || s.Unresolved != 0 {
		t.Errorf("stats %+v", s)
	}

	var cand []callgraph.SAMEdge
	for _, e := range r.Listeners {
		var cs []callgraph.MethodRef
		for _, c := range e.Creators {
			cs = append(cs, callgraph.MethodRef{FQCN: c.FQCN, Name: c.Name, Desc: c.Desc})
		}
		cand = append(cand, callgraph.SAMEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC, Kind: e.Kind,
			TargetFQCN: e.TargetFQCN, TargetName: e.TargetMethod, TargetDesc: e.TargetMethodDesc, Creators: cs, Condition: e.Condition})
	}
	g := callgraph.Build(h, nil)
	var roots []uint32
	for _, m := range [][3]string{{"al.Main", "main", "([Ljava/lang/String;)V"}, {"al.Main", "lambda$main$1", "(Lal/Bus;)V"}} {
		id, ok := g.ID(m[0], m[1], m[2])
		if !ok {
			t.Fatalf("no node %v", m)
		}
		roots = append(roots, id)
	}
	st := callgraph.NewCreatorGate(g, h, cand, callgraph.BasisCallback).Apply(g, roots)
	// B's edge is CHA's already (Listener is the input's own interface);
	// the JDK-typed ones (Consumer, PropertyChangeListener) are the
	// model's alone.
	if st.Added != 4 || st.Covered != 1 || st.Gated != 1 {
		t.Errorf("gate %+v, want 4 added, 1 covered (B), 1 gated (D)", st)
	}
	reach := reachable(g, roots)
	for m, want := range ran {
		i := strings.IndexByte(m, '#')
		j := strings.IndexByte(m, '(')
		id, ok := g.ID(m[:i], m[i+1:j], m[j:])
		if !ok {
			t.Fatalf("no node %s", m)
		}
		if reach[id] != want {
			t.Errorf("%s reachable=%v, want %v (as on the JVM)", m, reach[id], want)
		}
	}
}

func reachable(g *callgraph.Graph, roots []uint32) map[uint32]bool {
	seen := map[uint32]bool{}
	queue := append([]uint32(nil), roots...)
	for _, r := range roots {
		seen[r] = true
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, e := range g.Edges(u) {
			if !seen[e.To] {
				seen[e.To] = true
				queue = append(queue, e.To)
			}
		}
	}
	return seen
}
