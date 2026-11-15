package reflectinvoke

import (
	"reflect"
	"sort"
	"testing"

	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
)

// TestResolveThroughParameters (REFLECTION R3 r3-3) on the oracle's own
// jar (testdata/reflectparam): Ops#apply's Method parameter resolves from
// two static fields and a lookup at the call (Minus, Plus, Times), Ops#call
// from a lookup of an instance method — both replaced, no gap left;
// Ops#mixed resolves Known#check from one caller, while the other passes a
// Method found by a run-time name: the resolved entry is added and the
// site stays a gap, marked partial. The census then sees that one site,
// not convertible.
func TestResolveThroughParameters(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/reflectparam/app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	resolved, warnings := ResolveInvokes(h, FindInvokeSites(h))
	if len(warnings) > 0 {
		t.Fatalf("warnings: %v", warnings)
	}
	idx := dataflow.NewCallerIndex(h)
	defer idx.Close()
	out, stats, warnings := ResolveThroughParameters(h, idx, resolved)
	if len(warnings) > 0 {
		t.Fatalf("warnings: %v", warnings)
	}
	if want := (ParameterStats{ParameterSites: 3, ResolvedSites: 3, PartialSites: 1, Targets: 5}); stats != want {
		t.Errorf("stats = %+v, want %+v", stats, want)
	}
	got := map[string][]string{}
	var gaps []string
	for _, r := range out {
		if r.UnresolvedReason != "" {
			gaps = append(gaps, r.Site.CallerName)
			if !r.PartlyThroughParameter {
				t.Errorf("gap at %s not marked partial: %s", r.Site.CallerName, r.UnresolvedReason)
			}
			continue
		}
		if !r.ThroughParameter {
			t.Errorf("entry at %s not from a parameter: %+v", r.Site.CallerName, r)
		}
		got[r.Site.CallerName] = append(got[r.Site.CallerName], r.OwnerFQCN+"#"+r.MethodName)
	}
	for _, v := range got {
		sort.Strings(v)
	}
	want := map[string][]string{
		"apply": {"fx.app.Minus#check", "fx.app.Plus#check", "fx.app.Times#check"},
		"call":  {"fx.app.Greeter#hello"},
		"mixed": {"fx.app.Known#check"},
	}
	if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(gaps, []string{"mixed"}) {
		t.Errorf("resolved = %v, gaps = %v; want %v and [mixed]", got, gaps, want)
	}
	for _, r := range out {
		if r.Site.CallerName == "apply" && !r.ObjIsNull {
			t.Errorf("Ops#apply invokes a static method (null target): %+v", r)
		}
	}

	census, _ := Provenance(h, idx, out, nil, func(string, string, string) bool { return true })
	if len(census) != 1 || census[0].Category != ProvParameter || census[0].Convertible || census[0].CallerName != "mixed" {
		t.Errorf("census = %+v, want the one partial parameter site", census)
	}
}
