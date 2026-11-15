package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

const jpaFixture = "../../testdata/jpa/"

// TestJPAOracle (framework:proxies spring-orm:*; ROADMAP B4 b4-6):
// testdata/jpa (scripts/jpa-oracle.sh) uses the proxies Spring ORM 5.3.6
// hands an application — the shared entity manager and the query it
// defers, the managed entity manager factory and the extended entity
// managers it creates — over a minimal provider (fx.store) whose bodies
// record the whole stack. The JVM runs each call through the JDK proxy to
// the Spring invocation handler, which calls the provider's method through
// Method#invoke.
//
// Every observed body of the input must be reachable from main; an entity
// manager nothing creates (OrphanEntityManager) must not be. The proxy
// catalog links main's calls to the handlers, but the handlers' last hop
// is reflection: the provider's bodies are reached by jpa:objects (e1;
// before it, b4-6 measured all four unreached) — model edges straight from
// main's call sites.
func TestJPAOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", jpaFixture + "app.jar",
		"-deps", jpaFixture + "deps", "-profile", "generic", "-roots", "profile",
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

	reach := reachableFrom(ex, "fx.app.Main#main([Ljava/lang/String;)V")
	all := (benchFixture{observed: jpaFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	obs := map[string]bool{}
	for b := range all {
		if strings.HasPrefix(b, "fx.") || strings.HasPrefix(b, "org.springframework.") {
			obs[b] = true
		}
	}
	if len(obs) < 12 {
		t.Fatalf("observed.txt has %d input bodies; regenerate with scripts/jpa-oracle.sh", len(obs))
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
	const orphan = "fx.store.OrphanEntityManager#find(Ljava/lang/Class;Ljava/lang/Object;)Ljava/lang/Object;"
	if reach[orphan] {
		t.Errorf("the graph reaches %s, which nothing creates", orphan)
	}

	const (
		shared  = "fx.app.Main#shared(Ljavax/persistence/EntityManagerFactory;)Ljava/lang/String;"
		managed = "fx.app.Main#managed()Ljava/lang/String;"
		find    = "fx.store.FxEntityManager#find(Ljava/lang/Class;Ljava/lang/Object;)Ljava/lang/Object;"
	)
	want := map[[2]string]bool{
		{shared, find}: true,
		{shared, "fx.store.FxEntityManager#createQuery(Ljava/lang/String;)Ljavax/persistence/Query;"}:       true,
		{shared, "fx.store.FxQuery#getResultList()Ljava/util/List;"}:                                        true,
		{managed, "fx.store.FxEntityManagerFactory#createEntityManager()Ljavax/persistence/EntityManager;"}: true,
		{managed, find}: true,
	}
	for id := range ex.Edges {
		if strings.HasPrefix(id.To, "fx.store.OrphanEntityManager#") {
			t.Errorf("edge to an entity manager nothing creates: %s -> %s", id.From, id.To)
		}
		if id.Basis == string(callgraph.BasisModel) {
			delete(want, [2]string{id.From, id.To})
		}
	}
	for e := range want {
		t.Errorf("no model edge %s -> %s", e[0], e[1])
	}
}
