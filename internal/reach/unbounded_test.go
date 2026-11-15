package reach

import (
	"reflect"
	"testing"
)

// TestUnboundedSummary (ROADMAP B0 b0-5): only reachable sites with no
// bounded candidate set count; they are grouped by the calling class's
// component (its GAV; NoInstance for the application) and by calling
// method, busiest first, with TopN bounding the caller list.
func TestUnboundedSummary(t *testing.T) {
	site := func(kind, fqcn, name string, owners ...string) GapSite {
		return GapSite{Kind: kind, CallerFQCN: fqcn, CallerName: name, CallerDesc: "()V", Owners: owners}
	}
	e := buildInstanceEngine(t, false)
	if e.UnboundedSummary(5) != nil {
		t.Fatal("summary without gaps, want nil")
	}
	e.SetGaps([]GapSite{
		site("reflection", "reachfixture.AppEntry", "callsVulnerableDirectly"),
		site("reflection", "reachfixture.AppEntry", "callsVulnerableDirectly"),
		site("methodhandle", "reachfixture.AppEntry", "callsVulnerableDirectly"),
		site("deserialization", "reachfixture.FrameworkThing", "doWork"),
		site("reflection", "reachfixture.AppEntry", "callsFrameworkWhichCallsVulnerable", "reachfixture.VulnerableLib"), // bounded
		site("reflection", "reachfixture.VulnerableLib", "unreachable"),                                                 // not reachable
	})
	got := e.UnboundedSummary(2)
	app := "reachfixture.AppEntry#callsVulnerableDirectly()V"
	want := &UnboundedSummary{
		Version: UnboundedVersion, Sites: 4, Callers: 2,
		ByKind: map[string]int{"reflection": 2, "methodhandle": 1, "deserialization": 1},
		TopN:   2, TopCallerSites: 3,
		ByComponent: []UnboundedGroup{
			{Component: got.ByComponent[0].Component, Sites: 3, Callers: 1},
			{Component: "org.framework:thing:1.0", Sites: 1, Callers: 1},
		},
		TopCallers: []UnboundedCaller{
			{Method: app, Kind: "reflection", Sites: 2},
			{Method: app, Kind: "methodhandle", Sites: 1},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("summary =\n %+v\nwant\n %+v", got, want)
	}
	if c := got.ByComponent[0].Component; c == "org.framework:thing:1.0" || c == "" {
		t.Errorf("application sites grouped under %q", c)
	}
}
