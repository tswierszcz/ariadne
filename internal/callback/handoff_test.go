package callback

import (
	"bufio"
	"os"
	"reflect"
	"sort"
	"testing"

	"ariadne/internal/hierarchy"
)

// TestHandoffs (ASYNC_CALLBACKS a3-3) on testdata/async-handoff:
//   - Loop (a queue drained by poll) and Swap (double buffering: submit
//     adds to one deque, process swaps, iterates the other — a catch
//     inside the loop — and clears it) give edges from where each object
//     enters: Api#go → A#work, the lambda, B#work; Api#other → C#work only
//     (other() never ran, so C#work did not either — the edge is Api#other's,
//     not Api#go's);
//   - Listeners (a list walked without clearing) is a broadcast, not a
//     drain: no edge to L#work (never fired, never ran);
//   - Unused is never stored: no edge;
//   - an NIO attachment: Sel#register → Sel#loop (reads the attachment
//     of selectedKeys), not Sel#dumpKeys (walks every key, ran too).
//
// observed.txt (scripts/async-handoff-oracle.sh: the drains on a second
// thread, a real Pipe and Selector) is checked for each claim.
func TestHandoffs(t *testing.T) {
	const fx = "../../testdata/async-handoff/"
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
		"ah.A#work()V": true, "ah.B#work()V": true, "ah.Api#lambda$go$0()V": true,
		"ah.Handler#ready(Ljava/nio/channels/SelectionKey;)V": true, "ah.Sel#loop(Ljava/nio/channels/Selector;)V": true,
		"ah.Sel#dumpKeys(Ljava/nio/channels/Selector;)V": true,
		"ah.C#work()V": false, "ah.L#work()V": false, "ah.Unused#work()V": false, "ah.Api#other(Lah/Loop;)V": false,
	} {
		if obs[m] != want {
			t.Errorf("observed.txt: %s ran=%v, want %v", m, obs[m], want)
		}
	}

	h, _, err := hierarchy.Build(fx+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	edges, stats, warnings := FindHandoffs(h)
	if len(warnings) != 0 {
		t.Errorf("warnings: %v", warnings)
	}
	var got []string
	for _, e := range edges {
		got = append(got, e.CallerFQCN+"#"+e.CallerName+" -> "+e.TargetFQCN+"#"+e.TargetMethod+" ["+e.Condition+"]")
	}
	sort.Strings(got)
	want := []string{
		"ah.Api#go -> ah.A#work [queued]",
		"ah.Api#go -> ah.Api#lambda$go$0 [queued]",
		"ah.Api#go -> ah.B#work [queued]",
		"ah.Api#other -> ah.C#work [queued]",
		"ah.Sel#register -> ah.Sel#loop [selected]",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("edges\n got %q\nwant %q", got, want)
	}
	// Loop#submit, Swap#submit, Listeners#add store; Listeners has no
	// drain. Drains: Loop#drain, Swap#process, Sel#loop.
	if stats.QueueStores != 3 || stats.QueueStoresDrained != 2 || stats.SelectorStores != 1 || stats.SelectorStoresDrained != 1 || stats.Drains != 3 {
		t.Errorf("stats %+v", stats)
	}
}
