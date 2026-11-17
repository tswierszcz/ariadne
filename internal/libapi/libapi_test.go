package libapi

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const fixture = "../../testdata/library/"

func build(t *testing.T) (*callgraph.Graph, *hierarchy.Hierarchy) {
	t.Helper()
	h, _, err := hierarchy.Build(fixture+"lib.jar", []input.DepJar{{Path: fixture + "deps/commons-text-1.9.jar"}})
	if err != nil {
		t.Fatal(err)
	}
	return callgraph.Build(h, callgraph.DefaultSkip), h
}

func names(g *callgraph.Graph, roots []uint32) []string {
	var out []string
	for _, r := range roots {
		k := g.Key(r)
		out = append(out, strings.TrimPrefix(k[:strings.IndexByte(k, '(')], "lib."))
	}
	sort.Strings(out)
	return out
}

// Roots follow docs/RESULTS_CONTRACT.MD §8 on testdata/library: public
// and protected members of accessible classes (protected not in a final
// class), public member classes of public classes, the registered
// provider; never private or package-private code, a private member
// class, or a synthetic method.
func TestRoots(t *testing.T) {
	g, h := build(t)
	providers := []Provider{{Service: "lib.api.Codec", Class: "lib.internal.ExpandCodec"}}
	roots, cov := Roots(g, h, nil, providers, Options{})
	want := []string{
		"api.Codec#decode", "api.Codecs#<init>", "api.Codecs#decodeAll",
		"api.Formatter#<init>", "api.Formatter#format",
		"api.Listener#onEvent", "api.Pipeline#<init>", "api.Pipeline#run",
		"api.Templates$Options#<init>", "api.Templates$Options#strict",
		"api.Templates#<init>", "api.Templates#prepare", "api.Templates#render",
		"internal.ExpandCodec#<init>", "internal.ExpandCodec#decode",
		"internal.Util#<init>", "internal.Util#expand",
	}
	sort.Strings(want)
	if got := names(g, roots); !reflect.DeepEqual(got, want) {
		t.Errorf("roots\n got %v\nwant %v", got, want)
	}
	if cov.Classes != 9 || cov.Methods != len(want) || cov.Providers != 1 || cov.JPMS != "classpath" {
		t.Errorf("coverage %+v", cov)
	}

	// An excluded package takes its classes and its providers out.
	roots, cov = Roots(g, h, nil, providers, Options{Exclude: []string{"lib.internal"}})
	for _, n := range names(g, roots) {
		if strings.HasPrefix(n, "internal.") {
			t.Errorf("excluded root %s", n)
		}
	}
	if cov.Providers != 0 || !reflect.DeepEqual(cov.Excluded, []string{"lib.internal"}) {
		t.Errorf("excluded coverage %+v", cov)
	}
}

// ClientSites (LIBRARY_MODE LB2) on testdata/library: the library's calls
// on methods a client can implement or override — the client's Listener
// in Pipeline#run (case d), the Codec a ServiceLoader yields — and never a
// call into a final, inaccessible or dependency type.
func TestClientSites(t *testing.T) {
	_, h := build(t)
	site := func(opt Options) []string {
		var got []string
		for _, s := range ClientSites(h, nil, opt) {
			d := strings.TrimPrefix(s.Declared[:strings.IndexByte(s.Declared, '(')], "lib.")
			if s.Abstract {
				d += " (abstract)"
			}
			got = append(got, strings.TrimPrefix(s.CallerFQCN, "lib.")+"#"+s.CallerName+" -> "+d)
		}
		sort.Strings(got)
		return got
	}
	want := []string{"api.Codecs#decodeAll -> api.Codec#decode (abstract)", "api.Pipeline#run -> api.Listener#onEvent (abstract)",
		"api.Templates#render -> api.Templates#prepare"}
	if got := site(Options{}); !reflect.DeepEqual(got, want) {
		t.Errorf("client sites %v, want %v", got, want)
	}
	// An excluded package's types are not the client's to implement.
	if got := site(Options{Exclude: []string{"lib.api"}}); len(got) != 0 {
		t.Errorf("with lib.api excluded: %v", got)
	}
}

// Constructible (lb2-2): a client can get objects of public classes with a
// public constructor and of public interfaces, never of a package-private
// class — so only the latter's bodies stand in for client code.
func TestConstructible(t *testing.T) {
	_, h := build(t)
	c := Constructible(h, nil, Options{})
	for fqcn, want := range map[string]bool{
		"lib.api.Templates": true, "lib.api.Listener": true, "lib.api.Pipeline": true,
		"lib.api.ExpandingListener": false, "lib.api.Hidden": false, "lib.api.Engine": false,
		"org.apache.commons.text.StringSubstitutor": false, // a dependency's, not the library's
	} {
		if got := c(fqcn); got != want {
			t.Errorf("Constructible(%s) = %v, want %v", fqcn, got, want)
		}
	}
	if Constructible(h, nil, Options{Exclude: []string{"lib.api"}})("lib.api.Templates") {
		t.Error("an excluded package's class counted as constructible")
	}
}
