package callback

import (
	"bufio"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
)

// TestForwardedTasks (ASYNC_CALLBACKS a3-2) on testdata/async-forward: the
// Executor#execute site in Pool#submit(Runnable task) gets the task's
// parameter resolved at its callers — Task (a direct `new` in Api#go), the
// lambda Api#lambda$go$0, Deep (passed through Relay#relay → Relay#pass),
// each with the caller chain — and never Unused, which nobody submits.
// observed.txt (scripts/async-forward-oracle.sh, a real
// ThreadPoolExecutor) shows the three ran and Unused did not.
func TestForwardedTasks(t *testing.T) {
	const fx = "../../testdata/async-forward/"
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
	for _, m := range []string{"af.Task#run()V", "af.Deep#run()V", "af.Api#lambda$go$0()V"} {
		if !obs[m] {
			t.Errorf("observed.txt: %s did not run", m)
		}
	}
	if obs["af.Unused#run()V"] {
		t.Error("observed.txt: Unused#run ran")
	}

	h, _, err := hierarchy.Build(fx+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	var sites []RegistrationSite
	for _, s := range FindRegistrationSites(h) {
		if s.CallerFQCN == "af.Pool" && s.CallerName == "submit" {
			sites = append(sites, s)
		}
	}
	if len(sites) != 1 {
		t.Fatalf("%d registration sites in Pool#submit, want 1", len(sites))
	}
	resolved, warnings := ResolveCallbacks(h, sites)
	if len(warnings) != 0 {
		t.Errorf("warnings: %v", warnings)
	}
	var got []string
	for _, rc := range resolved {
		target := rc.TargetFQCN
		if rc.ImplOwner != "" {
			target = rc.ImplOwner + "#" + rc.ImplName
		}
		got = append(got, target+" via "+strings.Join(rc.Forwarded, " < "))
	}
	sort.Strings(got)
	want := []string{
		"af.Api#lambda$go$0 via af.Api#go(Laf/Pool;)V",
		"af.Deep via af.Relay#pass(Ljava/lang/Runnable;)V < af.Relay#relay(Ljava/lang/Runnable;)V < af.Api#go(Laf/Pool;)V",
		"af.Task via af.Api#go(Laf/Pool;)V",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resolved\n got %q\nwant %q", got, want)
	}
	edges := EdgeTargets(FindCallbackEdges(h, resolved))
	var targets []string
	for _, e := range edges {
		targets = append(targets, e.TargetFQCN+"#"+e.TargetMethod)
	}
	sort.Strings(targets)
	if !reflect.DeepEqual(targets, []string{"af.Api#lambda$go$0", "af.Deep#run", "af.Task#run"}) {
		t.Errorf("edge targets %v", targets)
	}
}
