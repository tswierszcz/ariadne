package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/profile"
	"ariadne/internal/springxml"
)

const xmlBeansFixture = "../../testdata/xmlbeans/"

// TestXMLBeansOracle (Spring XML bean definitions; ROADMAP B4 b4-8):
// testdata/xmlbeans (scripts/xmlbeans-oracle.sh) starts a Spring 5.3.6
// ClassPathXmlApplicationContext from fx/app/beans.xml. The JVM runs the
// constructors the definitions select (a constructor-arg, a c: argument),
// the setters of <property> and p: properties (an inherited one through a
// parent definition), an inner bean, explicit and default init/destroy
// methods — for application classes and a library class (deps/fx-lib.jar,
// like JPetStore's SqlSessionFactoryBean). Before the XML lifecycle roots
// the graph reached 1 of the 18 bodies the JVM ran (main).
//
// Every observed body must be reachable from main and the spring
// profile's bean lifecycle roots; the class only an abstract definition
// names (Ghost), a class no definition names (Plain) and the constructor
// no definition selects (Service's no-argument one) must not be. The
// lazy-init bean nothing asks for is not created on this run, but the rule
// counts lazy beans (created on first use), so its constructor is reached.
func TestXMLBeansOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", xmlBeansFixture + "app.jar",
		"-deps", xmlBeansFixture + "deps", "-profile", "spring", "-roots", "profile",
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
	h, _, err := hierarchy.Build(xmlBeansFixture+"app.jar", []input.DepJar{{Path: xmlBeansFixture + "deps/fx-lib.jar"}})
	if err != nil {
		t.Fatal(err)
	}
	xf, err := os.Open(xmlBeansFixture + "app/fx/app/beans.xml")
	if err != nil {
		t.Fatal(err)
	}
	defs, err := springxml.Parse(xf)
	xf.Close()
	if err != nil {
		t.Fatal(err)
	}
	starts := []string{"fx.app.Main#main([Ljava/lang/String;)V"}
	for _, r := range profile.XMLBeanLifecycleRoots(h, defs.Beans, defs.Inner) {
		starts = append(starts, r.FQCN+"#"+r.Name+r.Desc)
	}
	reach := map[string]bool{}
	for _, root := range starts {
		for b := range reachableFrom(ex, root) {
			reach[b] = true
		}
	}

	all := (benchFixture{observed: xmlBeansFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	obs := map[string]bool{}
	for b := range all {
		if strings.HasPrefix(b, "fx.app.") || strings.HasPrefix(b, "fx.lib.") {
			obs[b] = true
		}
	}
	if len(obs) < 18 {
		t.Fatalf("observed.txt has %d input bodies; regenerate with scripts/xmlbeans-oracle.sh", len(obs))
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
	for body, wantReach := range map[string]bool{
		"fx.app.Ghost#<init>()V":                   false,
		"fx.app.Ghost#setTag(Ljava/lang/String;)V": false,
		"fx.app.Plain#<init>()V":                   false,
		"fx.app.Service#<init>()V":                 false,
		"fx.app.Sleepy#<init>()V":                  true,
	} {
		if obs[body] {
			t.Fatalf("observed.txt has %s: the fixture's negative case ran", body)
		}
		if reach[body] != wantReach {
			t.Errorf("the graph reaches %s: %v, want %v", body, reach[body], wantReach)
		}
	}
}
