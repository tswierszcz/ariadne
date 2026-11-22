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
	return diRunProfiles(t, "default")
}

// diRunProfiles is diRun with the given -active-profiles (comma-separated)
// — FRAMEWORKS F5 f5-3's second scenario — and returns the graph and the bodies
// reachable from main, the @Bean method, the @EventListener method and
// the bean lifecycle roots (profile.BeanLifecycleRoots).
func diRunProfiles(t *testing.T, active string) (callgraph.Export, map[string]bool) {
	t.Helper()
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", diFixture + "app.jar",
		"-deps", diFixture + "deps", "-profile", "spring", "-roots", "profile", "-active-profiles", active,
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
	for _, r := range profile.BeanLifecycleRoots(h, strings.Split(active, ",")) {
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
		// FRAMEWORKS F5 f5-3: profile expressions, none holding with only
		// "default" active.
		"fx.beans.EuTax#<init>()V":      false,
		"fx.beans.CloudEuTax#<init>()V": false,
		"fx.beans.EitherTax#<init>()V":  false,
		"fx.beans.AnyTax#<init>()V":     false,
	} {
		if obs[never] {
			t.Fatalf("observed.txt has %s: the fixture's negative case ran", never)
		}
		if reach[never] != wantReach {
			t.Errorf("the graph reaches %s: %v, want %v", never, reach[never], wantReach)
		}
	}
}

// TestDIProfilesOracle (FRAMEWORKS F5 f5-3, a second scenario): the same
// application on the JVM with the profiles eu and cloud active
// (observed-eu-cloud.txt). Every body it ran must be reachable under
// -active-profiles eu,cloud — the beans of "eu", "eu & cloud", "eu | us"
// and {"us", "cloud"} among them — and the "!eu" bean's constructor must
// not be.
func TestDIProfilesOracle(t *testing.T) {
	_, reach := diRunProfiles(t, "eu,cloud")
	all := (benchFixture{observed: diFixture + "observed-eu-cloud.txt", bodyColumn: 0}).observedBodies(t)
	var missed []string
	n := 0
	for b := range all {
		if !strings.HasPrefix(b, "fx.") {
			continue
		}
		n++
		if !reach[b] {
			missed = append(missed, b)
		}
	}
	if n < 18 {
		t.Fatalf("observed-eu-cloud.txt has %d application bodies; regenerate with scripts/di-oracle.sh", n)
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("with eu,cloud active the JVM ran %s, the graph does not reach it from the roots", b)
	}
	for _, b := range []string{"fx.beans.EuTax#<init>()V", "fx.beans.CloudEuTax#<init>()V", "fx.beans.EitherTax#<init>()V", "fx.beans.AnyTax#<init>()V"} {
		if !all[b] {
			t.Fatalf("observed-eu-cloud.txt lacks %s; regenerate with scripts/di-oracle.sh", b)
		}
	}
	if us := "fx.beans.UsTax#<init>()V"; all[us] || reach[us] {
		t.Errorf("%s: ran %v, reached %v — want neither with eu active (@Profile(\"!eu\"))", us, all[us], reach[us])
	}
}

// TestDIInjectionOracle (FRAMEWORKS F5 f5-3, qualifier and @Primary): the
// JVM prints which bean the container injected into each of Shop's four
// Store fields ("# inject" lines, the same in both scenarios). The report's
// injection resolution (profile.ResolveInjections, as main.go runs it) must
// name the same class for each — an undecided point would be honest, a
// different class is a wrong fact.
func TestDIInjectionOracle(t *testing.T) {
	jvm := map[string]string{}
	for _, f := range []string{"observed.txt", "observed-eu-cloud.txt"} {
		raw, err := os.ReadFile(diFixture + f)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(raw), "\n") {
			if rest, ok := strings.CutPrefix(line, "# inject "); ok {
				k, v, _ := strings.Cut(rest, " = ")
				if prev, seen := jvm[k]; seen && prev != v {
					t.Fatalf("%s: %s in one scenario, %s in the other", k, prev, v)
				}
				jvm[k] = v
			}
		}
	}
	if len(jvm) != 4 {
		t.Fatalf("%d \"# inject\" lines (%v); regenerate with scripts/di-oracle.sh", len(jvm), jvm)
	}
	h, _, err := hierarchy.Build(diFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	names := profile.BeanNames(h)
	beans := append(append([]profile.BeanName{}, names...), profile.BeanMethods(h)...)
	got := map[string]profile.ResolvedInjection{}
	for _, ri := range profile.ResolveInjections(h, profile.InjectionPoints(h, names), beans) {
		if ri.FQCN == "fx.beans.Shop" && ri.ParamIndex == -1 {
			got[ri.FQCN+"#"+ri.Member] = ri
		}
	}
	for k, want := range jvm {
		ri, ok := got[k]
		switch {
		case !ok:
			t.Errorf("%s: no injection point found; the JVM injected %s", k, want)
		case ri.Resolved != want:
			t.Errorf("%s (qualifier %q): resolved %q (%s), the JVM injected %s", k, ri.Qualifier, ri.Resolved, ri.Reason, want)
		}
	}
}
