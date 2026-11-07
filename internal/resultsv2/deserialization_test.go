package resultsv2

import (
	"reflect"
	"testing"

	"ariadne/internal/cves"
	"ariadne/internal/deserialize"
)

// TestChainRefs: a finding links to the chains naming its advisory (by
// ID or alias) on its own instance only; another instance's copy, or a
// chain naming another advisory, is not linked.
func TestChainRefs(t *testing.T) {
	spec := deserialize.ChainSpec{ID: "java:x", Advisories: []string{"CVE-1"}}
	other := deserialize.ChainSpec{ID: "java:y", Advisories: []string{"CVE-2"}}
	chains := []deserialize.ChainPresence{
		{Spec: spec, Instance: "a", State: deserialize.ChainCandidate},
		{Spec: spec, Copy: "b", Instance: "b", State: deserialize.ChainBlocked},
		{Spec: other, Instance: "a", State: deserialize.ChainCandidate},
	}
	got := chainRefs(chains, cves.CVE{ID: "GHSA-1", Aliases: []string{"CVE-1"}}, "a")
	want := []ChainRef{{Chain: "java:x", State: deserialize.ChainCandidate}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("instance a: %+v, want %+v", got, want)
	}
	got = chainRefs(chains, cves.CVE{ID: "CVE-1"}, "b")
	want = []ChainRef{{Chain: "java:x", Copy: "b", State: deserialize.ChainBlocked}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("instance b: %+v, want %+v", got, want)
	}
	if got := chainRefs(chains, cves.CVE{ID: "CVE-1"}, ""); got != nil {
		t.Errorf("no instance: %+v", got)
	}
}

// TestBuildDeserialization: absent without catalog versions; entries and
// chains rendered with their site states, and the summary counts them.
func TestBuildDeserialization(t *testing.T) {
	if d := buildDeserialization(Inputs{}); d != nil {
		t.Fatalf("no model: %+v", d)
	}
	sites := []deserialize.Site{{CallerFQCN: "app.M", CallerName: "read", CallerDesc: "()V", PC: 7, Owner: "java.io.ObjectInputStream",
		Name: "readObject", Desc: "()Ljava/lang/Object;", Model: "jdk:java-serialization", Filter: deserialize.FilterNoneSeen, IsApp: true}}
	chains := []deserialize.ChainPresence{{Spec: deserialize.ChainSpec{ID: "java:x", Format: deserialize.FormatJava}, State: deserialize.ChainCandidate,
		Sites: []deserialize.ChainSite{{Site: &sites[0], State: deserialize.ChainCandidate}}, Counts: map[string]int{deserialize.ChainCandidate: 1}}}
	d := buildDeserialization(Inputs{DeserModelCatalog: "m", DeserChainCatalog: "c", DeserSites: sites, DeserChains: chains})
	if len(d.Entries) != 1 || d.Entries[0].ID != "app.M#read()V@7" || !d.Entries[0].App {
		t.Errorf("entries %+v", d.Entries)
	}
	if len(d.Chains) != 1 || d.Chains[0].Flow != deserialize.FlowNotAnalyzed || len(d.Chains[0].Sites) != 1 || d.Chains[0].Sites[0].Entry != "app.M#read()V@7" {
		t.Errorf("chains %+v", d.Chains)
	}
	if d.Summary.EntriesByPolicy[deserialize.FilterNoneSeen] != 1 || d.Summary.ChainsByState[deserialize.ChainCandidate] != 1 {
		t.Errorf("summary %+v", d.Summary)
	}
}
