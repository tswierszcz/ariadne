package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

const springDataFixture = "../../testdata/springdata/"

// TestSpringDataOracle (framework:proxies Spring Data repositories; ROADMAP
// B4): testdata/springdata (scripts/springdata-oracle.sh) calls save,
// findById and count on a CrudRepository interface implemented by Spring
// Data Commons 2.4.8 over a minimal store module (a RepositoryFactorySupport
// subclass and its base repository class, which records the whole stack).
// The JVM runs main → the JDK proxy → JdkDynamicAopProxy#invoke →
// ReflectiveMethodInvocation, Spring Data's interceptors →
// RepositoryMethodInvoker → Method#invoke → the base class's method.
//
// Every observed body of the input must be reachable. The catalog links the
// repository call to JdkDynamicAopProxy#invoke and — since x2
// (spring-data:base-repository; b4-5 measured the reflective last hop
// unreached) — to the base repository's save, findById and count; a
// concrete Repository class nothing creates or names (OrphanRepository) gets
// no edge.
func TestSpringDataOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", springDataFixture + "app.jar",
		"-deps", springDataFixture + "deps", "-profile", "generic", "-roots", "profile",
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

	const proxyInvoke = "org.springframework.aop.framework.JdkDynamicAopProxy#invoke(Ljava/lang/Object;Ljava/lang/reflect/Method;[Ljava/lang/Object;)Ljava/lang/Object;"
	reach := reachableFrom(ex, "fx.app.Main#main([Ljava/lang/String;)V")
	all := (benchFixture{observed: springDataFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	obs := map[string]bool{}
	for b := range all {
		if strings.HasPrefix(b, "fx.") || strings.HasPrefix(b, "org.springframework.") {
			obs[b] = true
		}
	}
	if len(obs) < 14 || !obs[proxyInvoke] {
		t.Fatalf("observed.txt has %d input bodies; regenerate with scripts/springdata-oracle.sh", len(obs))
	}
	var missed []string
	for b := range obs {
		if !reach[b] {
			missed = append(missed, b)
		}
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
	}

	const query = "fx.app.Main#query()Ljava/lang/String;"
	proxyEdge := false
	base := map[string]bool{}
	for id := range ex.Edges {
		if id.Basis != string(callgraph.BasisModel) {
			continue
		}
		if id.From == query && id.To == proxyInvoke {
			proxyEdge = true
		}
		if strings.HasPrefix(id.To, "fx.store.") {
			if id.From != query {
				t.Errorf("base-repository edge from %s (want only the repository call site): -> %s", id.From, id.To)
			}
			base[id.To] = true
		}
	}
	if !proxyEdge {
		t.Error("no framework:proxies edge from the repository call to JdkDynamicAopProxy#invoke")
	}
	want := map[string]bool{
		"fx.store.FxRepository#save(Ljava/lang/Object;)Ljava/lang/Object;":       true,
		"fx.store.FxRepository#findById(Ljava/lang/Object;)Ljava/util/Optional;": true,
		"fx.store.FxRepository#count()J":                                         true,
	}
	if !reflect.DeepEqual(base, want) {
		t.Errorf("base-repository edges to %v, want %v (never OrphanRepository)", base, want)
	}
}
