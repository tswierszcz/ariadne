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

const diFixture = "../../testdata/di/"

// diRun analyzes testdata/di under -profile spring -roots profile
// -active-profiles default (the oracle's JVM ran with no profile set,
// which Spring calls "default") and returns the graph and the bodies
// reachable from main, the @Bean method, the @EventListener method and
// the bean lifecycle roots (profile.BeanLifecycleRoots).
func diRun(t *testing.T) (callgraph.Export, map[string]bool) {
	t.Helper()
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", diFixture + "app.jar",
		"-deps", diFixture + "deps", "-profile", "spring", "-roots", "profile", "-active-profiles", "default",
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
	h, _, err := hierarchy.Build(diFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	starts := []string{
		"fx.app.Main#main([Ljava/lang/String;)V",
		"fx.app.AppConfig#engine()Lfx/beans/Engine;",
		"fx.beans.TickListener#onTick(Lfx/beans/Tick;)V",
	}
	for _, r := range profile.BeanLifecycleRoots(h, []string{"default"}) {
		starts = append(starts, r.FQCN+"#"+r.Name+r.Desc)
	}
	reach := map[string]bool{}
	for _, root := range starts {
		for b := range reachableFrom(ex, root) {
			reach[b] = true
		}
	}
	return ex, reach
}

// TestDIOracle (Spring dependency injection and bean lifecycle; ROADMAP B4
// b4-7): testdata/di (scripts/di-oracle.sh) starts a Spring 5.3.6
// annotation-config context. The JVM runs the beans' constructors, @Autowired
// setters, @PostConstruct/@PreDestroy methods, @Bean init/destroy methods,
// the callback interfaces (Aware, InitializingBean, DisposableBean,
// SmartInitializingSingleton, FactoryBean, BeanPostProcessor,
// BeanFactoryPostProcessor, ApplicationListener) and a Condition. Every
// observed body of the input must be reachable from the profile's roots;
// the constructors of beans the container never creates — an inactive
// @Profile (known from -active-profiles), a class with no stereotype —
// must not be. Before the bean lifecycle roots (b4-7) the graph missed 7
// bodies: both constructors, the @Autowired setter, @PostConstruct,
// @PreDestroy and the @Bean's initMethod/destroyMethod — all called
// reflectively; the callback interfaces were already reached through CHA.
//
// A failed @Conditional and a @Lazy bean nothing injects are not created
// on this run either, but whether the container creates them is decided at
// run time (a condition, a later getBean): the rule counts them as "may be
// created" rather than guess false, so their constructors are reached.
func TestDIOracle(t *testing.T) {
	_, reach := diRun(t)
	all := (benchFixture{observed: diFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	obs := map[string]bool{}
	for b := range all {
		if strings.HasPrefix(b, "fx.") {
			obs[b] = true
		}
	}
	if len(obs) < 18 {
		t.Fatalf("observed.txt has %d application bodies; regenerate with scripts/di-oracle.sh", len(obs))
	}
	var missed []string
	for b := range obs {
		if !reach[b] {
			missed = append(missed, b)
		}
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("the JVM ran %s, the graph does not reach it from the roots", b)
	}
	for never, wantReach := range map[string]bool{
		"fx.beans.Dormant#<init>()V": false,
		"fx.beans.Plain#<init>()V":   false,
		"fx.beans.Unmet#<init>()V":   true,
		"fx.beans.Sleepy#<init>()V":  true,
	} {
		if obs[never] {
			t.Fatalf("observed.txt has %s: the fixture's negative case ran", never)
		}
		if reach[never] != wantReach {
			t.Errorf("the graph reaches %s: %v, want %v", never, reach[never], wantReach)
		}
	}
}
