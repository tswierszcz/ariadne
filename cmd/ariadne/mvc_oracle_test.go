package main

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/profile"
)

const mvcFixture = "../../testdata/mvc/"

// mvcDeps puts testdata/mvc's jars and the Spring Framework jars
// testdata/tx holds into one directory (-deps takes one), as
// scripts/mvc-oracle.sh's classpath does.
func mvcDeps(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "deps")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var jars []string
	m, _ := filepath.Glob(mvcFixture + "deps/*.jar")
	jars = append(jars, m...)
	for _, n := range []string{"jcl", "core", "beans", "aop", "context", "expression"} {
		jars = append(jars, txFixture+"deps/spring-"+n+"-5.3.6.jar")
	}
	for _, j := range jars {
		in, err := os.Open(j)
		if err != nil {
			t.Fatal(err)
		}
		out, err := os.Create(filepath.Join(dir, filepath.Base(j)))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(out, in); err != nil {
			t.Fatal(err)
		}
		in.Close()
		out.Close()
	}
	return dir
}

// mvcRun analyzes testdata/mvc under -profile spring -roots profile and
// returns the bodies reachable from the profile's own roots: the
// annotation and main roots (Profile.AppRoots) and the bean lifecycle
// roots, as -roots profile starts from them.
func mvcRun(t *testing.T) map[string]bool {
	t.Helper()
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", mvcFixture + "app.jar",
		"-deps", mvcDeps(t), "-profile", "spring", "-roots", "profile", "-active-profiles", "default",
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
	h, _, err := hierarchy.Build(mvcFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	cat, err := profile.LoadDefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	prof, _ := cat.Get("spring")
	g := callgraph.Build(h, nil)
	var starts []string
	for _, id := range prof.AppRoots(g, h) {
		starts = append(starts, g.Key(id))
	}
	for _, r := range profile.BeanLifecycleRoots(h, []string{"default"}) {
		starts = append(starts, r.FQCN+"#"+r.Name+r.Desc)
	}
	for _, r := range profile.WebBodyRoots(h) {
		starts = append(starts, r.FQCN+"#"+r.Name+r.Desc)
	}
	reach := map[string]bool{}
	for _, root := range starts {
		reach[root] = true
		for b := range reachableFrom(ex, root) {
			reach[b] = true
		}
	}
	return reach
}

// TestMVCOracle (FRAMEWORKS F5 f5-2, the container before the controller):
// testdata/mvc (scripts/mvc-oracle.sh) runs a real embedded Tomcat 9 with a
// Spring MVC 5.3.6 DispatcherServlet and sends real HTTP requests. Every
// application body the JVM ran must be reachable from the spring profile's
// roots under -roots profile. Before f5-2 the graph missed 14 of the 26:
// the @ExceptionHandler, @ModelAttribute and @InitBinder methods (invoked
// reflectively like a handler), the WebMvcConfigurer callbacks and what
// they create, the interceptor's methods, and Jackson's calls into the
// response and request body types. The fix: those three annotations and
// the WebMvcConfigurer/HandlerInterceptor methods as spring profile roots,
// and the web body roots (profile.WebBodyRoots).
//
// An unmapped public method nothing calls must stay unreached. A mapping
// never requested, and a controller outside the component scan (the JVM
// answers its request with a 404), are reached by design: the profile
// roots every mapping of an annotated class without knowing the scan.
func TestMVCOracle(t *testing.T) {
	reach := mvcRun(t)
	all := (benchFixture{observed: mvcFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	obs := map[string]bool{}
	for b := range all {
		if strings.HasPrefix(b, "fx.") {
			obs[b] = true
		}
	}
	h, _, err := hierarchy.Build(mvcFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	var bodies []string
	for _, r := range profile.WebBodyRoots(h) {
		bodies = append(bodies, string(r.Kind)+" "+r.FQCN+"#"+r.Name+r.Desc)
	}
	wantBodies := []string{
		"response body getter fx.web.Item#getId()Ljava/lang/String;",
		"response body getter fx.web.Item#getName()Ljava/lang/String;",
		"request body constructor fx.web.NewItem#<init>()V",
		"request body setter fx.web.NewItem#setName(Ljava/lang/String;)V",
	}
	if strings.Join(bodies, "\n") != strings.Join(wantBodies, "\n") {
		t.Errorf("web body roots:\n%s\nwant:\n%s", strings.Join(bodies, "\n"), strings.Join(wantBodies, "\n"))
	}
	if len(obs) < 20 {
		t.Fatalf("observed.txt has %d application bodies; regenerate with scripts/mvc-oracle.sh", len(obs))
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
		"fx.web.ShopController#unmapped()Ljava/lang/String;": false,
		"fx.web.ShopController#never()Ljava/lang/String;":    true,
		"fx.other.GhostController#ghost()Ljava/lang/String;": true,
		"fx.other.GhostController#<init>()V":                 true,
	} {
		if obs[never] {
			t.Fatalf("observed.txt has %s: the fixture's case that never runs ran", never)
		}
		if reach[never] != wantReach {
			t.Errorf("the graph reaches %s: %v, want %v", never, reach[never], wantReach)
		}
	}
}
