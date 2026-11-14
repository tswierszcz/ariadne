package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

const mybatisFixture = "../../testdata/mybatis/"

// TestMyBatisOracle (framework:proxies MyBatis mapper, with jdbc:objects and
// service-provider:dispatch p2; ROADMAP B4): testdata/mybatis
// (scripts/mybatis-oracle.sh) calls an @Mapper method through the real
// MyBatis 3.4.1 over an UnpooledDataSource; the driver's bodies record every
// frame on the stack, so observed.txt holds the bodies the JVM ran from
// main through the JDK proxy, MapperProxy, MapperMethod, the executor, the
// statement and result-set handlers, DriverManager, down to the driver.
//
// Every one of them in the input (fx.*, org.apache.ibatis.*) must be
// reachable from main — including the application's own InvocationHandler
// lambda behind a JDK proxy of an interface that is not a mapper (reached
// through the lambda's creation edge), which the catalog must not link to
// MapperProxy (negative case), and fx.drv.FxStatement#close: MyBatis closes
// through java.sql.Statement, the driver's statement implements only
// PreparedStatement, and that PreparedStatement extends Statement is a JDK
// fact the hierarchy does not read — internal/jdkapi supplies it
// (jdbc:objects j2; b4-4 measured it missing before).
func TestMyBatisOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", mybatisFixture + "app.jar",
		"-deps", mybatisFixture + "deps", "-profile", "generic", "-roots", "profile",
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
	all := (benchFixture{observed: mybatisFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	obs := map[string]bool{}
	for b := range all {
		if strings.HasPrefix(b, "fx.") || strings.HasPrefix(b, "org.apache.ibatis.") {
			obs[b] = true
		}
	}
	if len(obs) < 45 || !obs["org.apache.ibatis.binding.MapperProxy#invoke(Ljava/lang/Object;Ljava/lang/reflect/Method;[Ljava/lang/Object;)Ljava/lang/Object;"] {
		t.Fatalf("observed.txt has %d input bodies; regenerate with scripts/mybatis-oracle.sh", len(obs))
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

	// The mapper call reaches MapperProxy through the catalog; the
	// non-mapper proxy call does not.
	mapperEdge, ownEdge := false, false
	for id := range ex.Edges {
		if id.Basis != string(callgraph.BasisModel) || !strings.HasPrefix(id.To, "org.apache.ibatis.binding.MapperProxy#invoke") {
			continue
		}
		switch id.From {
		case "fx.app.Main#query()Ljava/lang/String;":
			mapperEdge = true
		case "fx.app.Main#ownProxy()Ljava/lang/String;":
			ownEdge = true
		}
	}
	if !mapperEdge {
		t.Error("no framework:proxies edge from the mapper call to MapperProxy#invoke")
	}
	if ownEdge {
		t.Error("the call on a JDK proxy of a non-mapper interface is linked to MapperProxy#invoke")
	}
}
