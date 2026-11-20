package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callback"
	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// asyncQualFixture is one ASYNC_CALLBACKS fixture with a real-JVM oracle.
type asyncQualFixture struct {
	dir, pkg, main, slice string
}

var asyncQualFixtures = []asyncQualFixture{
	{"async", "fx.a.", "fx.a.Main", "A0–A2"},
	{"continuations", "fx.c.", "fx.c.Main", "A3 slice 1, a3-7"},
	{"async-forward", "af.", "af.Main", "a3-2"},
	{"async-handoff", "ah.", "ah.Main", "a3-3"},
	{"async-returned", "ar.", "ar.Main", "a3-4"},
	{"async-listeners", "al.", "al.Main", "a3-5"},
	{"async-reactive", "rx.", "rx.Main", "a3-6"},
	{"async-completion", "cf.", "cf.Main", "a3-7"},
	{"async-threads", "th.", "th.Main", "a4-1"},
}

// asyncQualification is what TestAsyncQualification measures for one
// fixture: bodies the JVM ran (Ran) and how many of them the graph does
// not reach from main (Missed); `callback` edges from a reached method
// to a fixture body (Edges) and how many of them lead to a body the JVM
// did not run (FalseAssoc; a bridge counts as run when the method it
// delegates to ran) — each a "may run" the controlled run did not take.
type asyncQualification struct {
	Ran, Missed, Edges, FalseAssoc int
	missed, falseAssoc             []string
}

// TestAsyncQualification (ASYNC_CALLBACKS A4, a4-2) runs the whole
// pipeline (-profile generic -roots profile, CHA) on every async fixture
// and compares the graph with its real-JVM oracle (observed.txt): every
// body that ran must be reached from main (no misses), and the `callback`
// edges whose target did not run are listed — the expected ones are the
// models' documented may-run cases (a rejected task, an onError of a
// subscriber whose onNext does not throw, a listener on a source that
// never notifies…), not wrong associations. The table is written to
// testdata/async-qualify/RESULTS.MD when ARIADNE_WRITE_ASYNC_RESULTS=1.
func TestAsyncQualification(t *testing.T) {
	// async's one miss is DYNAMIC_LINKAGE's, not this area's: a
	// MethodHandle whose method name is chosen at run time.
	want := map[string]asyncQualification{
		"async":            {Ran: 10, Missed: 1, Edges: 6},
		"continuations":    {Ran: 12, Edges: 16, FalseAssoc: 4},
		"async-forward":    {Ran: 13, Edges: 3},
		"async-handoff":    {Ran: 24, Edges: 5},
		"async-returned":   {Ran: 20, Edges: 4, FalseAssoc: 1},
		"async-listeners":  {Ran: 28, Edges: 4},
		"async-reactive":   {Ran: 26, Edges: 14, FalseAssoc: 1},
		"async-completion": {Ran: 21, Edges: 5},
		"async-threads":    {Ran: 12, Edges: 3},
	}
	results := map[string]asyncQualification{}
	for _, fx := range asyncQualFixtures {
		fx := fx
		t.Run(fx.dir, func(t *testing.T) {
			q := qualifyAsync(t, fx)
			results[fx.dir] = q
			w := want[fx.dir]
			if q.Ran != w.Ran || q.Missed != w.Missed || q.Edges != w.Edges || q.FalseAssoc != w.FalseAssoc {
				t.Errorf("%s: ran %d, missed %d %v, callback edges %d, to bodies that did not run %d %v; want %+v",
					fx.dir, q.Ran, q.Missed, q.missed, q.Edges, q.FalseAssoc, q.falseAssoc, w)
			}
		})
	}
	if os.Getenv("ARIADNE_WRITE_ASYNC_RESULTS") == "1" {
		writeAsyncResults(t, results)
	}
}

func qualifyAsync(t *testing.T, fx asyncQualFixture) asyncQualification {
	t.Helper()
	base := "../../testdata/" + fx.dir + "/"
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", base + "app.jar",
		"-profile", "generic", "-roots", "profile",
		"-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-graph-out", graph}
	if code, out := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	f, err := os.Open(graph)
	if err != nil {
		t.Fatal(err)
	}
	ex, err := callgraph.Read(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	obs := (benchFixture{observed: base + "observed.txt", bodyColumn: 0}).observedBodies(t)
	var mainID string
	for id := range ex.Edges {
		if strings.HasPrefix(id.From, fx.main+"#main(") {
			mainID = id.From
			break
		}
	}
	if mainID == "" {
		t.Fatalf("no edges from %s#main", fx.main)
	}
	reached := reachableFrom(ex, mainID)
	var q asyncQualification
	for b := range obs {
		if !strings.HasPrefix(b, fx.pkg) {
			continue
		}
		q.Ran++
		if !reached[b] {
			q.Missed++
			q.missed = append(q.missed, b)
		}
	}
	h, _, err := hierarchy.Build(base+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	// ranAs: the body ran — itself, or, for a bridge (a Callable's
	// call()Object, a Consumer's accept(Object)), the method of the same
	// class and name it delegates to.
	ranAs := func(m string) bool {
		if obs[m] {
			return true
		}
		i := strings.IndexByte(m, '(')
		for b := range obs {
			if strings.HasPrefix(b, m[:i]+"(") {
				return true
			}
		}
		return false
	}
	hasBody := func(m string) bool {
		i, j := strings.IndexByte(m, '#'), strings.IndexByte(m, '(')
		ci, ok := h.Lookup(m[:i])
		if !ok {
			return false
		}
		mm, ok := ci.MethodByNameAndDesc(m[i+1:j], m[j:])
		return ok && len(mm.Code) > 0
	}
	seen := map[string]bool{}
	for id := range ex.Edges {
		if id.Basis != string(callgraph.BasisCallback) || !strings.HasPrefix(id.To, fx.pkg) {
			continue
		}
		// Only edges this run could have taken: from a reached method,
		// to a method with a body.
		if !reached[id.From] || !hasBody(id.To) {
			continue
		}
		k := id.From + ">" + id.To
		if seen[k] {
			continue
		}
		seen[k] = true
		q.Edges++
		if !ranAs(id.To) {
			q.FalseAssoc++
			q.falseAssoc = append(q.falseAssoc, short(id.From)+" -> "+short(id.To))
		}
	}
	sort.Strings(q.missed)
	sort.Strings(q.falseAssoc)
	return q
}

func short(m string) string {
	if i := strings.IndexByte(m, '('); i > 0 {
		return m[:i]
	}
	return m
}

// TestAsyncTwoTasks (ASYNC_CALLBACKS.MD scenario 3): two tasks handed to
// two registration sites do not mix — every site with a direct `new` or
// lambda argument resolves to exactly one target.
func TestAsyncTwoTasks(t *testing.T) {
	for _, fx := range asyncQualFixtures {
		h, _, err := hierarchy.Build("../../testdata/"+fx.dir+"/app.jar", nil)
		if err != nil {
			t.Fatal(err)
		}
		cbs, _ := callback.ResolveCallbacks(h, callback.FindRegistrationSites(h))
		per := map[string]map[string]bool{}
		for _, c := range cbs {
			if len(c.Forwarded) > 0 || (c.TargetFQCN == "" && c.ImplOwner == "") {
				continue // through callers (a3-2): one site, several callers
			}
			k := fmt.Sprintf("%s#%s%s@%d", c.Site.CallerFQCN, c.Site.CallerName, c.Site.CallerDesc, c.Site.PC)
			if per[k] == nil {
				per[k] = map[string]bool{}
			}
			per[k][c.TargetFQCN+c.ImplOwner+"#"+c.ImplName] = true
		}
		for site, targets := range per {
			if len(targets) != 1 {
				t.Errorf("%s: site %s resolves to %d targets %v, want 1", fx.dir, site, len(targets), targets)
			}
		}
	}
}

func writeAsyncResults(t *testing.T, results map[string]asyncQualification) {
	var b strings.Builder
	b.WriteString("# ASYNC_CALLBACKS qualification (A4, a4-2)\n\n")
	b.WriteString("Generated by `ARIADNE_WRITE_ASYNC_RESULTS=1 go test ./cmd/ariadne -run TestAsyncQualification`:\n")
	b.WriteString("the whole pipeline (`-profile generic -roots profile`, CHA) on every async fixture, against\n")
	b.WriteString("its real-JVM oracle (`observed.txt`). *Missed*: bodies the JVM ran that the graph does not\n")
	b.WriteString("reach from `main`. *To bodies that did not run*: `callback` edges whose target the controlled\n")
	b.WriteString("run did not execute — may-run edges, each explained below; none is a wrong association\n")
	b.WriteString("(a task, listener or subscriber linked to a site that does not hand it over).\n\n")
	b.WriteString("| Fixture | Slice | Bodies ran | Missed | `callback` edges | To bodies that did not run |\n|---|---|---|---|---|---|\n")
	var all []string
	for _, fx := range asyncQualFixtures {
		q := results[fx.dir]
		fmt.Fprintf(&b, "| `%s` | %s | %d | %d | %d | %d |\n", fx.dir, fx.slice, q.Ran, q.Missed, q.Edges, q.FalseAssoc)
		for _, e := range q.falseAssoc {
			all = append(all, fmt.Sprintf("- `%s`: %s", fx.dir, e))
		}
	}
	b.WriteString("\nMissed: `async` — `fx.a.Handled#chosenAtRunTime`, a `MethodHandle` whose method name is chosen at\n")
	b.WriteString("run time (DYNAMIC_LINKAGE L2's documented gap, not this area's).\n\n")
	b.WriteString("Edges to bodies that did not run, and why:\n\n")
	for i, e := range all {
		for k, why := range asyncWhy {
			if strings.Contains(e, k) {
				all[i] = e + " — " + why
				break
			}
		}
	}
	b.WriteString(strings.Join(all, "\n") + "\n")
	if err := os.WriteFile("../../testdata/async-qualify/RESULTS.MD", []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// asyncWhy explains each expected edge to a body the controlled run did
// not execute (RESULTS.MD).
var asyncWhy = map[string]string{
	"lambda$failure$0": "ran (the supplier throws), but `continuations`' oracle records only bodies calling `Ran.here()`",
	"lambda$failure$5": "the same: `exceptionally(t -> null)` has no marker",
	"lambda$failure$1": "`thenApply` on a `supplyAsync` stage that always fails — completion of derived stages is not decided (a3-7 decides hand-completed futures only)",
	"lambda$success$0": "`exceptionally` on `completedFuture` — the same",
	"IdleTask#run":     "a3-4's gate on `Idle#onSelected`, which CHA reaches through `Endpoint#onSelected` though nothing creates `Idle` — the gate is only as precise as the graph's reachability of the returning method (RTA keeps it too)",
	"Upper#onError":    "a3-6: every emission may end in `onError` (an `onNext` that throws); `Upper#onNext` does not throw on this run",
}
