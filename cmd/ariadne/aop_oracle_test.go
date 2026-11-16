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

const aopFixture = "../../testdata/aop/"

// aopGraph runs the spring profile over testdata/aop and returns the
// exported graph and the input's hierarchy.
func aopGraph(t *testing.T, extra ...string) (callgraph.Export, *hierarchy.Hierarchy) {
	t.Helper()
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", aopFixture + "app.jar",
		"-deps", txFixture + "deps", "-profile", "spring", "-roots", "profile",
		"-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-graph-out", graph}
	args = append(args, extra...)
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
	h, _, err := hierarchy.Build(aopFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	return ex, h
}

// TestAOPOracle (spring:aop): testdata/aop (scripts/aop-oracle.sh) calls
// beans advised by TracingAspect on a Spring 5.3.6 context with
// @EnableAspectJAutoProxy — OrderService through CGLIB, JdbcRepo through a
// JDK proxy, the @Monitored ReportService. Spring runs the advice methods
// through AbstractAspectJAdvice#invokeAdviceMethodWithGivenArgs →
// Method#invoke, so without the model nothing reaches them.
//
// Every advice body the JVM ran is reached from main and the bean
// lifecycle roots, except withArgs: its pointcut uses args(...), which a1
// does not evaluate (the declared gap). The advice that never ran — a
// pattern matching no method, an @Aspect that is no bean — gets no edge;
// the self-invocation place → internal gets none, and the advice record
// shows the JVM ran no advice for it, nor for the final method.
func TestAOPOracle(t *testing.T) {
	ex, h := aopGraph(t)
	starts := []string{"fx.app.Main#main([Ljava/lang/String;)V"}
	for _, r := range profile.BeanLifecycleRoots(h, nil) {
		starts = append(starts, r.FQCN+"#"+r.Name+r.Desc)
	}
	reach := map[string]bool{}
	for _, s := range starts {
		for b := range reachableFrom(ex, s) {
			reach[b] = true
		}
	}
	const withArgs = "fx.aop.aspect.TracingAspect#withArgs(J)V"
	all := (benchFixture{observed: aopFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	var missed []string
	n := 0
	for b := range all {
		if !strings.HasPrefix(b, "fx.") || strings.Contains(b, "$$") {
			continue // generated at run time
		}
		n++
		if !reach[b] && b != withArgs {
			missed = append(missed, b)
		}
	}
	if n < 12 || !all[withArgs] {
		t.Fatalf("observed.txt has %d input bodies; regenerate with scripts/aop-oracle.sh", n)
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("the JVM ran %s, the graph does not reach it", b)
	}
	if reach[withArgs] {
		t.Errorf("%s reached: args(...) is not evaluated, update the declared gap", withArgs)
	}

	// The advice record: no advice for the self-invocation or the final
	// method.
	for _, a := range []string{"advised before <- fx.aop.svc.OrderService#place", "advised before <- fx.aop.svc.OrderService#cancel"} {
		if !all[a] {
			t.Fatalf("observed.txt lacks %q; regenerate with scripts/aop-oracle.sh", a)
		}
	}
	for _, a := range []string{"advised before <- fx.aop.svc.OrderService#internal", "advised before <- fx.aop.svc.OrderService#fixed"} {
		if all[a] {
			t.Errorf("observed.txt has %q: the model's self-invocation/final rule disagrees with the JVM", a)
		}
	}

	model := map[string]map[string]bool{} // caller → model targets
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisModel) && strings.HasPrefix(id.To, "fx.aop.aspect.") {
			if model[id.From] == nil {
				model[id.From] = map[string]bool{}
			}
			model[id.From][id.To] = true
		}
	}
	const (
		main      = "fx.app.Main#main([Ljava/lang/String;)V"
		before    = "fx.aop.aspect.TracingAspect#before(Lorg/aspectj/lang/JoinPoint;)V"
		afterSave = "fx.aop.aspect.TracingAspect#afterSave(Ljava/lang/Object;)V"
		onError   = "fx.aop.aspect.TracingAspect#onError(Ljava/lang/Exception;)V"
		audit     = "fx.aop.aspect.TracingAspect#audit(Lorg/aspectj/lang/ProceedingJoinPoint;Lfx/aop/svc/Audited;)Ljava/lang/Object;"
		monitor   = "fx.aop.aspect.TracingAspect#monitor()V"
	)
	for _, to := range []string{before, afterSave, onError, audit, monitor} {
		if !model[main][to] {
			t.Errorf("no model edge %s → %s", main, to)
		}
	}
	for from, tos := range model {
		if from != main {
			t.Errorf("model edges from %s → %v: only main calls advised beans through the proxy", from, tos)
		}
		for to := range tos {
			if strings.Contains(to, "#ghost(") || strings.Contains(to, "#withArgs(") || strings.HasPrefix(to, "fx.aop.aspect.Unregistered#") {
				t.Errorf("model edge %s → %s: the JVM never runs it", from, to)
			}
		}
	}
}
