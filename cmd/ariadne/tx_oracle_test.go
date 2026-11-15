package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/profile"
)

const txFixture = "../../testdata/tx/"

// TestTxOracle (framework:proxies spring-tx:transactional, x3):
// testdata/tx (scripts/tx-oracle.sh) calls @Transactional services — on
// the class (a CGLIB proxy), on one method, behind an interface (a JDK
// proxy), one throwing — and a plain service, through the beans a Spring
// 5.3.6 context with @EnableTransactionManagement hands out. The JVM runs
// TransactionInterceptor#invoke → TransactionAspectSupport →
// PlatformTransactionManager#getTransaction/commit/rollback around the
// transactional calls only.
//
// Every observed body of the input (not the generated proxy classes) must
// be reachable from the profile's roots. The model edge comes from the
// call sites of transactional methods only: none from the call of the
// non-transactional OrderService#count or PlainService#hello.
func TestTxOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", txFixture + "app.jar",
		"-deps", txFixture + "deps", "-profile", "spring", "-roots", "profile",
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
	h, _, err := hierarchy.Build(txFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	starts := []string{"fx.app.Main#main([Ljava/lang/String;)V", "fx.app.TxConfig#transactionManager()Lorg/springframework/transaction/PlatformTransactionManager;"}
	for _, r := range profile.BeanLifecycleRoots(h, nil) {
		starts = append(starts, r.FQCN+"#"+r.Name+r.Desc)
	}
	reach := map[string]bool{}
	for _, s := range starts {
		for b := range reachableFrom(ex, s) {
			reach[b] = true
		}
	}
	all := (benchFixture{observed: txFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	var missed []string
	n := 0
	for b := range all {
		if strings.Contains(b, "$$") || strings.HasPrefix(b, "jdk.proxy") {
			continue // generated at run time, not in the input
		}
		if strings.HasPrefix(b, "fx.") || strings.HasPrefix(b, "org.springframework.transaction.") {
			n++
			if !reach[b] {
				missed = append(missed, b)
			}
		}
	}
	if n < 15 {
		t.Fatalf("observed.txt has %d input bodies; regenerate with scripts/tx-oracle.sh", n)
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("the JVM ran %s, the graph does not reach it from the roots", b)
	}

	const run = "fx.app.Main#run(Lorg/springframework/context/annotation/AnnotationConfigApplicationContext;)Ljava/lang/String;"
	const interceptor = "org.springframework.transaction.interceptor.TransactionInterceptor#invoke(Lorg/aopalliance/intercept/MethodInvocation;)Ljava/lang/Object;"
	found := false
	for id := range ex.Edges {
		if id.From == run && id.To == interceptor && id.Basis == string(callgraph.BasisModel) {
			found = true
		}
	}
	if !found {
		t.Error("no model edge from the transactional call sites in Main#run to TransactionInterceptor#invoke")
	}
}
