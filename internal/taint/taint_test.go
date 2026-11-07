package taint_test

import (
	"strings"
	"testing"

	"ariadne/internal/deserialize"
	"ariadne/internal/flowsum"
	"ariadne/internal/hierarchy"
	"ariadne/internal/taint"
)

const fxJar = "../../testdata/taint/app.jar"

// run is the CLI's pipeline on the fixture: entries, policies, edges (for
// candidate/rejected counts), summaries with opaque sources, taint.
func run(t *testing.T, opts taint.Options) map[string]*deserialize.Site {
	t.Helper()
	h, _, err := hierarchy.Build(fxJar, nil)
	if err != nil {
		t.Fatal(err)
	}
	sites := deserialize.FindSites(h)
	deserialize.ResolvePolicies(h, sites)
	java, _ := deserialize.ClassTargets(h, deserialize.FormatJava)
	xs, _ := deserialize.ClassTargets(h, deserialize.FormatXStream)
	deserialize.Connect(sites, map[string]map[string][]deserialize.Target{deserialize.FormatJava: java, deserialize.FormatXStream: xs})
	a := taint.New(h, opts)
	fo := flowsum.DefaultOptions
	fo.Opaque = taint.Opaque
	e := flowsum.New(h, fo)
	t.Cleanup(func() { e.Close() })
	e.Summarize(deserialize.TaintRoots(sites, a))
	deserialize.AttachTaint(sites, a, e)
	out := map[string]*deserialize.Site{}
	for i := range sites {
		s := &sites[i]
		out[s.CallerFQCN+"#"+s.CallerName] = s
	}
	return out
}

func get(t *testing.T, sites map[string]*deserialize.Site, k string) *taint.Result {
	t.Helper()
	s, ok := sites[k]
	if !ok || s.Taint == nil {
		t.Fatalf("no taint result for %s", k)
	}
	return s.Taint
}

// TestPilotCases: D4's "Gotowe" — user data, a constant, an effective
// guard, an unrelated guard and an unknown fragment each get their own
// expected result; plus the other sources and the per-site separation.
func TestPilotCases(t *testing.T) {
	sites := run(t, taint.DefaultOptions)
	cases := []struct {
		entry, state, guard, rule string
	}{
		{"demo.Codec#read", taint.FlowFound, taint.GuardNone, taint.RuleSpringParam},
		{"demo.Endpoints#constant", taint.NoFlow, "", ""},
		{"demo.Endpoints#twoSites", taint.NoFlow, "", ""},
		{"demo.Endpoints#filtered", taint.FlowFound, taint.GuardEffective, taint.RuleSpringParam},
		{"demo.Endpoints#escaped", taint.FlowFound, taint.GuardOnlyOther, taint.RuleSpringParam},
		{"demo.Endpoints#unknown", taint.Undetermined, "", ""},
		{"demo.Endpoints#servlet", taint.FlowFound, taint.GuardNone, taint.RuleServletCall},
		{"demo.Endpoints#implicit", taint.FlowFound, taint.GuardNone, taint.RuleSpringImplicit},
		{"demo.Endpoints#fromCli", taint.NoFlow, "", ""},
	}
	for _, c := range cases {
		r := get(t, sites, c.entry)
		if r.State != c.state || r.GuardState != c.guard {
			t.Errorf("%s: state %s guard %q, want %s %q (gaps %v)", c.entry, r.State, r.GuardState, c.state, c.guard, r.Gaps)
			continue
		}
		if c.rule != "" && (len(r.Flows) == 0 || r.Flows[0].Rule != c.rule) {
			t.Errorf("%s: flows %+v, want rule %s", c.entry, r.Flows, c.rule)
		}
		if r.Catalog != taint.CatalogVersion || r.Rule != deserialize.SinkJavaStream {
			t.Errorf("%s: catalog %s rule %s", c.entry, r.Catalog, r.Rule)
		}
	}
}

// TestFlowPath: the path names each hop from the source to the sink,
// with the call that passes the data on.
func TestFlowPath(t *testing.T) {
	r := get(t, run(t, taint.DefaultOptions), "demo.Codec#read")
	if len(r.Flows) != 1 || r.FlowCount != 1 {
		t.Fatalf("flows %+v", r.Flows)
	}
	f := r.Flows[0]
	if f.Source != "@RequestParam" || len(f.Steps) != 2 {
		t.Fatalf("flow %+v", f)
	}
	if !strings.HasPrefix(f.Steps[0].Method, "demo.Endpoints#userData") || f.Steps[0].CallPC < 0 || f.Steps[0].Input != "parameter 0 (java.lang.String)" {
		t.Errorf("first step %+v", f.Steps[0])
	}
	if !strings.HasPrefix(f.Steps[1].Method, "demo.Codec#read([B)") || f.Steps[1].CallPC != -1 || f.Steps[1].Input != "parameter 0 (byte[])" {
		t.Errorf("sink step %+v", f.Steps[1])
	}
}

// TestGuards: the policy is the only effective guard; a branch check and
// an escaper are listed, not effective.
func TestGuards(t *testing.T) {
	sites := run(t, taint.DefaultOptions)
	f := get(t, sites, "demo.Endpoints#filtered")
	if len(f.Guards) == 0 || f.Guards[0].Kind != taint.GuardPolicy || !f.Guards[0].Effective {
		t.Errorf("filtered guards %+v", f.Guards)
	}
	e := get(t, sites, "demo.Endpoints#escaped")
	kinds := map[string]bool{}
	for _, g := range e.Guards {
		if g.Effective {
			t.Errorf("escaped: effective guard %+v", g)
		}
		kinds[g.Kind] = true
	}
	if !kinds[taint.GuardControlCheck] || !kinds[taint.GuardUnrelated] || kinds[taint.GuardPolicy] {
		t.Errorf("escaped guards %+v", e.Guards)
	}
}

// TestGaps: a field, a lambda caller and the depth limit are gaps
// (undetermined), never "no flow"; a constant caller of a shared helper
// does not hide the request caller, nor add a flow.
func TestGaps(t *testing.T) {
	sites := run(t, taint.DefaultOptions)
	want := map[string]string{
		"demo.Codec#readOther":  "flows through fields are not traced",
		"demo.Extra#readString": "lambda or method reference",
		"demo.Codec#readDeep":   "depth limit",
	}
	for entry, gap := range want {
		var r *taint.Result
		if entry == "demo.Extra#readString" {
			// readString's sink is in Codec.read (shared with userData);
			// its own caller chain ends at a lambda.
			r = get(t, sites, "demo.Codec#read")
			if !strings.Contains(strings.Join(r.Gaps, "\n"), gap) {
				t.Errorf("Codec#read gaps %v, want %q", r.Gaps, gap)
			}
			continue
		}
		r = get(t, sites, entry)
		if r.State != taint.Undetermined || !strings.Contains(strings.Join(r.Gaps, "\n"), gap) {
			t.Errorf("%s: state %s gaps %v, want undetermined with %q", entry, r.State, r.Gaps, gap)
		}
	}
	shared := get(t, sites, "demo.Codec#readShared")
	if shared.State != taint.FlowFound || shared.FlowCount != 1 || !strings.HasPrefix(shared.Flows[0].Steps[0].Method, "demo.Extra#mixed") {
		t.Errorf("readShared %+v", shared)
	}
	// literalOnly passes a constant: nothing to follow there, so it adds
	// neither a flow nor an outside input (Outside lists inputs only).
	if len(shared.Outside) != 0 || len(shared.Gaps) != 0 {
		t.Errorf("readShared outside %v gaps %v", shared.Outside, shared.Gaps)
	}
}

// TestDepthBound: a larger depth follows the whole chain to the source.
func TestDepthBound(t *testing.T) {
	o := taint.DefaultOptions
	o.MaxDepth = 12
	r := get(t, run(t, o), "demo.Codec#readDeep")
	if r.State != taint.FlowFound || len(r.Flows[0].Steps) != 11 {
		t.Errorf("readDeep with depth 12: %+v", r)
	}
}
