package deserialize

import (
	"reflect"
	"sort"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const fxDir = "../../testdata/deserialize/"

// fixture builds the fx cases with the XStream stub identified as
// version v ("" = unidentified) and resolves every site's policy.
func fixture(t *testing.T, v string) (map[string]Site, map[string][]string) {
	t.Helper()
	dep := input.DepJar{Path: fxDir + "xstream-stub.jar"}
	if v != "" {
		dep.GAV = &input.GAV{Group: "com.thoughtworks.xstream", Artifact: "xstream", Version: v}
	}
	h, _, err := hierarchy.Build(fxDir+"fx.jar", []input.DepJar{dep})
	if err != nil {
		t.Fatal(err)
	}
	sites := FindSites(h)
	if w := ResolvePolicies(h, sites); len(w) > 0 {
		t.Fatalf("warnings %v", w)
	}
	java, _ := ClassTargets(h, FormatJava)
	xs, _ := ClassTargets(h, FormatXStream)
	edges := Connect(sites, map[string]map[string][]Target{FormatJava: java, FormatXStream: xs})
	bySite := map[string]Site{}
	for _, s := range sites {
		if s.CallerFQCN == "fx.Cases" {
			bySite[s.CallerName] = s
		}
	}
	allowed := map[string][]string{} // site -> target classes reached
	for _, e := range edges {
		if e.CallerFQCN == "fx.Cases" {
			allowed[e.CallerName] = append(allowed[e.CallerName], e.TargetFQCN)
		}
	}
	for k := range allowed {
		sort.Strings(allowed[k])
	}
	return bySite, allowed
}

// TestJavaFilterPolicy: a literal createFilter on the same stream is
// evaluated (rejected classes lose their edges; UNDECIDED is allowed); a
// non-literal one is unknown (every edge kept); a filter on another
// stream does not apply.
func TestJavaFilterPolicy(t *testing.T) {
	sites, reached := fixture(t, "1.4.10")
	all := []string{"fx.Allowed", "fx.Bad"}
	for name, want := range map[string]struct {
		state   string
		reached []string
	}{
		"javaNone":              {FilterNoneSeen, all},
		"javaFilter":            {FilterEvaluated, []string{"fx.Allowed"}},
		"javaFilterDynamic":     {FilterUnknown, all},
		"javaFilterOtherStream": {FilterNoneSeen, all},
	} {
		s := sites[name]
		if s.Filter != want.state || !reflect.DeepEqual(reached[name], want.reached) {
			t.Errorf("%s: %s %q reached %v, want %s %v", name, s.Filter, s.FilterDetail, reached[name], want.state, want.reached)
		}
	}
	if s := sites["javaFilter"]; s.Rejected != 1 || s.RejectedSample[0] != "fx.Bad" {
		t.Errorf("javaFilter rejected %d %v", s.Rejected, s.RejectedSample)
	}
}

// TestXStreamPolicyByVersion: the same bytecode under three releases.
func TestXStreamPolicyByVersion(t *testing.T) {
	everything := []string{"fx.Allowed", "fx.Bad", "fx.Derived", "fx.allowed.Inner"}
	type exp struct {
		state   string
		reached []string
	}
	cases := map[string]map[string]exp{
		// no security framework: configuration calls do not exist / do nothing
		"1.4.5": {
			"xsPlain": {FilterUnrestricted, everything}, "xsAllow": {FilterUnrestricted, everything},
			"xsDefaultSecurity": {FilterUnrestricted, everything},
		},
		"1.4.10": {
			"xsPlain":           {FilterUnrestricted, everything},
			"xsAllow":           {FilterEvaluated, everything}, // allowTypes adds to ANY: still everything
			"xsWildcard":        {FilterEvaluated, []string{"fx.allowed.Inner"}},
			"xsHierarchy":       {FilterEvaluated, []string{"fx.Derived"}},
			"xsDefaultSecurity": {FilterEvaluated, nil},
			"xsAnyAfterNone":    {FilterEvaluated, everything},
			"xsDeny":            {FilterEvaluated, []string{"fx.Allowed", "fx.Derived", "fx.allowed.Inner"}}, // fx.** spans subpackages
			"xsField":           {FilterUnknown, everything},
			"xsEscape":          {FilterUnknown, everything},
		},
		"1.4.20": {
			"xsPlain":    {FilterDefaultAllowlist, nil},
			"xsAllow":    {FilterEvaluated, []string{"fx.Allowed"}},
			"xsWildcard": {FilterEvaluated, []string{"fx.allowed.Inner"}},
			"xsEscape":   {FilterUnknown, everything},
		},
		"": {"xsPlain": {FilterVersionUnknown, everything}},
	}
	for v, byName := range cases {
		sites, reached := fixture(t, v)
		for name, want := range byName {
			s := sites[name]
			if s.Filter != want.state || !reflect.DeepEqual(dedup(reached[name]), want.reached) {
				t.Errorf("XStream %q %s: %s %q reached %v, want %s %v", v, name, s.Filter, s.FilterDetail, dedup(reached[name]), want.state, want.reached)
			}
		}
	}
}

func dedup(in []string) []string {
	var out []string
	for i, s := range in {
		if i == 0 || s != in[i-1] {
			out = append(out, s)
		}
	}
	return out
}

func TestJavaFilterPatterns(t *testing.T) {
	pats, ok := parseJavaFilter("maxdepth=5;a.B;a.c.*;!a.**;x.Pre*;!*")
	if !ok {
		t.Fatal("rejected")
	}
	p := javaFilterPolicy(pats)
	for cls, want := range map[string]bool{"a.B": true, "a.c.D": true, "a.c.d.E": false, "a.Z": false, "x.PrefixThing": true, "y.Q": false} {
		if p(cls) != want {
			t.Errorf("%s = %v", cls, !want)
		}
	}
	if p := javaFilterPolicy(nil); !p("any.Class") {
		t.Errorf("UNDECIDED must allow")
	}
	if _, ok := parseJavaFilter("java.base/java.lang.*"); ok {
		t.Errorf("module-qualified pattern accepted")
	}
}

func TestXStreamWildcard(t *testing.T) {
	for pat, cases := range map[string]map[string]bool{
		"fx.**":  {"fx.A": true, "fx.a.B": true, "fy.A": false},
		"fx.*":   {"fx.A": true, "fx.a.B": false},
		"fx.?":   {"fx.A": true, "fx.AB": false},
		"fx.A$*": {"fx.A$1": true, "fx.A": false},
	} {
		re := xsWildcard(pat)
		for cls, want := range cases {
			if re.MatchString(cls) != want {
				t.Errorf("%s ~ %s = %v", pat, cls, !want)
			}
		}
	}
}
