package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

const jdbcFixture = "../../testdata/jdbc/"

// TestJDBCOracle (jdbc:objects and service-provider:dispatch for
// java.sql.Driver; ROADMAP B4): testdata/jdbc (scripts/jdbc-oracle.sh) runs
// one query through DriverManager and one through a pool's DataSource on
// the real JVM and records the bodies that ran. Every one of them must be
// reachable from main:
//   - the pool's bodies — its getConnection, its proxy connection, its proxy
//     statement's executeQuery inherited from an abstract class (Hikari's
//     shape) — through jdbc:objects edges, gated by their creators;
//   - the driver's bodies, which the JVM reaches from
//     java.sql.DriverManager#getConnection (JDK code outside the input):
//     service-provider:dispatch p2 models that facade as a call on
//     Driver#connect of the registered drivers (b4-3 measured them all
//     missing before).
//
// Statements outside the driver's and the pool's jars (never created, or
// created only by an uncalled method) are not reached. Inside the
// registered driver's jar the product expansion is not creator-gated, by
// design (as testdata/spidispatch's UnusedCodec): its never-created
// UnusedStatement, and the stub body it inherits, are targets the JVM never
// ran — listed exactly here.
func TestJDBCOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", jdbcFixture + "app.jar",
		"-deps", jdbcFixture + "deps", "-profile", "generic", "-roots", "profile",
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
	obs := (benchFixture{observed: jdbcFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	if len(obs) != 13 {
		t.Fatalf("observed.txt has %d bodies (%v); regenerate with scripts/jdbc-oracle.sh", len(obs), obs)
	}
	for b := range obs {
		if !reach[b] {
			t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
		}
	}

	for n := range reach {
		if strings.HasPrefix(n, "fx.other.") {
			t.Errorf("%s is reached: nothing reachable creates it and the JVM never ran it", n)
		}
	}
	// Every model edge ends at a body the JVM ran, except the documented
	// over-approximation of the driver jar's never-created statement.
	overApprox := map[string]bool{
		"fx.drv.UnusedStatement#executeQuery()Ljava/sql/ResultSet;": true,
		"fx.stub.StubPreparedStatement#close()V":                    true, // inherited by UnusedStatement
	}
	extra := map[string]bool{}
	for id := range ex.Edges {
		if id.Basis != string(callgraph.BasisModel) || !strings.HasPrefix(id.To, "fx.") || obs[id.To] {
			continue
		}
		if !overApprox[id.To] {
			t.Errorf("model edge to a body the JVM never ran: %s -> %s", id.From, id.To)
		}
		extra[id.To] = true
	}
	if len(extra) != len(overApprox) {
		t.Errorf("over-approximated targets %v, want exactly %v", extra, overApprox)
	}
}
