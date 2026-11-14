package main

import (
	"os"
	"path/filepath"
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
// The catalog's edge from the repository call to JdkDynamicAopProxy#invoke
// must exist, and every observed body of the input must be reachable, with
// one documented exception and one documented weakness, both from the last
// hop being reflection (RepositoryMethodInvoker → Method#invoke, a site no
// literal resolves):
//   - fx.store.FxRepository#count is not reached;
//   - save and findById are reached only by accident — through CHA edges
//     from CrudRepositoryInvoker (Spring Data's web/REST helper, which the
//     JVM does not run here), never from Spring Data's invocation path
//     (org.springframework.data.repository.core.support).
// The test fails once either changes, so a model that closes the hop
// updates this oracle.
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
	const count = "fx.store.FxRepository#count()J"
	var missed []string
	for b := range obs {
		switch {
		case b == count && reach[b]:
			t.Errorf("%s is reached: the reflective hop to the base repository closed — update this oracle", b)
		case b != count && !reach[b]:
			missed = append(missed, b)
		}
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
	}

	proxyEdge := false
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisModel) && id.From == "fx.app.Main#query()Ljava/lang/String;" && id.To == proxyInvoke {
			proxyEdge = true
		}
		if strings.HasPrefix(id.To, "fx.store.FxRepository#") && strings.HasPrefix(id.From, "org.springframework.data.repository.core.support.") {
			t.Errorf("edge %s -> %s: Spring Data's invocation path now reaches the base repository — update this oracle", id.From, id.To)
		}
	}
	if !proxyEdge {
		t.Error("no framework:proxies edge from the repository call to JdkDynamicAopProxy#invoke")
	}
}
