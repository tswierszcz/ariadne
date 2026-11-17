package libapi

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
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
// provider, the dependency methods API classes inherit (lb2-3); never
// private or package-private code, a private member class, or a
// synthetic method.
func TestRoots(t *testing.T) {
	g, h := build(t)
	providers := []Provider{{Service: "lib.api.Codec", Class: "lib.internal.ExpandCodec"}}
	roots, cov, inherited := Roots(g, h, nil, providers, Options{})
	want := []string{
		"api.Codec#decode", "api.Codecs#<init>", "api.Codecs#decodeAll",
		"api.Expander#<init>",
		"api.Formatter#<init>", "api.Formatter#format",
		"api.Listener#onEvent", "api.Pipeline#<init>", "api.Pipeline#run",
		"api.SafeExpander#<init>", "api.SafeExpander#replace",
		"api.StrictExpander#<init>", "api.StrictExpander#substitute",
		"api.Templates$Options#<init>", "api.Templates$Options#strict",
		"api.Templates#<init>", "api.Templates#prepare", "api.Templates#render",
		"internal.ExpandCodec#<init>", "internal.ExpandCodec#decode",
		"internal.Util#<init>", "internal.Util#expand",
	}
	sort.Strings(want)
	var own, inh []uint32
	for _, r := range roots {
		if inherited[r] != nil {
			inh = append(inh, r)
		} else {
			own = append(own, r)
		}
	}
	if got := names(g, own); !reflect.DeepEqual(got, want) {
		t.Errorf("roots\n got %v\nwant %v", got, want)
	}
	if cov.Classes != 12 || cov.Methods != len(roots) || cov.Inherited != len(inh) || cov.Providers != 1 || cov.JPMS != "classpath" {
		t.Errorf("coverage %+v", cov)
	}
	// lb2-3: (g) Expander inherits StringSubstitutor's public and
	// protected instance methods; (h) SafeExpander overrides replace(String),
	// so only Expander inherits that one; (i) StrictExpander inherits it
	// and overrides substitute. Static methods and constructors are never
	// inherited roots.
	by := map[string][]string{}
	for _, r := range inh {
		by[g.Key(r)] = inherited[r]
	}
	both := []string{"lib.api.Expander", "lib.api.SafeExpander"}
	all := []string{"lib.api.Expander", "lib.api.SafeExpander", "lib.api.StrictExpander"}
	for key, want := range map[string][]string{
		"org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;":              {"lib.api.Expander", "lib.api.StrictExpander"},
		"org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/Object;)Ljava/lang/String;":              all,
		"org.apache.commons.text.StringSubstitutor#substitute(Lorg/apache/commons/text/TextStringBuilder;II)Z": both, // protected
	} {
		if got := by[key]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s inherited by %v, want %v", key, got, want)
		}
	}
	for key := range by {
		if strings.Contains(key, "#<init>") || strings.Contains(key, "#replace(Ljava/lang/Object;Ljava/util/Map;)") {
			t.Errorf("constructor or static method %s is an inherited root", key)
		}
	}

	// An excluded package takes its classes and its providers out.
	roots, cov, _ = Roots(g, h, nil, providers, Options{Exclude: []string{"lib.internal"}})
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

// TestSpecialize (lb2-4): the copy of an inherited root runs on its API
// class. On a StrictExpander (case i) the inherited replace(String)
// calls the class's own substitute on this and never gets to
// StringSubstitutor's substitute or resolveVariable — what the JVM ran
// (observed-strict.txt) — while on an Expander (g) it does. Calls on
// another object keep every target.
func TestSpecialize(t *testing.T) {
	g, h := build(t)
	providers := []Provider{{Service: "lib.api.Codec", Class: "lib.internal.ExpandCodec"}}
	roots, _, inherited := Roots(g, h, nil, providers, Options{})
	nodes := g.NodeCount()
	roots, byClone, stats := Specialize(g, h, roots, inherited, func(c classfile.Call) []string {
		return callgraph.CHA{}.DispatchTargets(h, c)
	})
	if stats.Clones == 0 || stats.Dropped == 0 || stats.Unanalyzed != 0 || g.NodeCount() != nodes+stats.Clones {
		t.Fatalf("stats %+v", stats)
	}
	replace := "org.apache.commons.text.StringSubstitutor#replace(Ljava/lang/String;)Ljava/lang/String;"
	copies := map[string]uint32{}
	for _, r := range roots {
		if g.Key(r) == replace {
			if g.Origin(r) == r || len(byClone[r]) != 1 {
				t.Fatalf("root %d for %s is not a single-class copy", r, replace)
			}
			copies[byClone[r][0]] = r
		}
		if inherited[r] != nil {
			t.Errorf("root %s is still the unspecialized inherited method", g.Key(r))
		}
	}
	if len(copies) != 2 {
		t.Fatalf("replace(String) copies for %v, want Expander and StrictExpander", copies)
	}
	reached := func(from uint32) map[string]bool {
		seen, out, queue := map[uint32]bool{from: true}, map[string]bool{}, []uint32{from}
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			out[g.Key(u)] = true
			for _, e := range g.Edges(u) {
				if !seen[e.To] {
					seen[e.To] = true
					queue = append(queue, e.To)
				}
			}
		}
		return out
	}
	const (
		own     = "lib.api.StrictExpander#substitute(Lorg/apache/commons/text/TextStringBuilder;II)Z"
		dep     = "org.apache.commons.text.StringSubstitutor#substitute(Lorg/apache/commons/text/TextStringBuilder;II)Z"
		resolve = "org.apache.commons.text.StringSubstitutor#resolveVariable(Ljava/lang/String;Lorg/apache/commons/text/TextStringBuilder;II)Ljava/lang/String;"
	)
	strict := reached(copies["lib.api.StrictExpander"])
	if !strict[own] || strict[dep] || strict[resolve] {
		t.Errorf("StrictExpander copy reaches own %v, StringSubstitutor#substitute %v, resolveVariable %v; want true, false, false", strict[own], strict[dep], strict[resolve])
	}
	plain := reached(copies["lib.api.Expander"])
	if plain[own] || !plain[dep] || !plain[resolve] {
		t.Errorf("Expander copy reaches StrictExpander#substitute %v, StringSubstitutor#substitute %v, resolveVariable %v; want false, true, true", plain[own], plain[dep], plain[resolve])
	}
	// The unspecialized method keeps every CHA target.
	orig, _ := g.IDOfKey(replace)
	if all := reached(orig); !all[own] || !all[dep] {
		t.Errorf("original replace(String) lost a target: own %v, dep %v", all[own], all[dep])
	}
	// The oracle: on a StrictExpander the JVM ran its own substitute and
	// neither of the other two.
	b, err := os.ReadFile(fixture + "observed-strict.txt")
	if err != nil {
		t.Fatal(err)
	}
	ran := map[string]bool{}
	for _, l := range strings.Split(string(b), "\n") {
		ran[strings.TrimSpace(l)] = true
	}
	if !ran[own] || !ran[replace] || ran[dep] || ran[resolve] {
		t.Error("observed-strict.txt disagrees with the case; regenerate with scripts/library-oracle.sh")
	}
}
