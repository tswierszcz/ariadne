package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

const jdbcFixture = "../../testdata/jdbc/"

// TestJDBCOracle (jdbc:objects and service-provider:dispatch for
// java.sql.Driver; ROADMAP B4): testdata/jdbc (scripts/jdbc-oracle.sh) runs
// one query through DriverManager and one through a pool's DataSource on
// the real JVM and records the bodies that ran.
//
// The pool's bodies — its getConnection, its proxy connection, its proxy
// statement's executeQuery inherited from an abstract class (Hikari's
// shape) — are reached through jdbc:objects edges, gated by their
// creators. Statements nothing reachable creates (outside the driver and
// the pool, never created, or created only by an uncalled method) are not
// reached.
//
// Known false negative, measured by this oracle: the driver's bodies. The
// JVM reaches FxDriver#connect from java.sql.DriverManager#getConnection,
// JDK code outside the input, so no call on java.sql.Driver is in the
// input; service-provider:dispatch has nothing to expand, the driver's
// connection is never created by a reachable method, and jdbc:objects
// gates out its connection, statement and result set. No gap is reported
// for it. The test fails when any of them becomes reachable, so a model
// change that closes this updates the oracle.
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
	var reached, missed []string
	for b := range obs {
		if reach[b] {
			reached = append(reached, b)
		} else {
			missed = append(missed, b)
		}
	}
	sort.Strings(reached)
	sort.Strings(missed)
	wantReached := []string{
		"fx.pool.AbstractProxyStatement#executeQuery()Ljava/sql/ResultSet;",
		"fx.pool.FxPool#getConnection()Ljava/sql/Connection;",
		"fx.pool.ProxyConnection#close()V",
		"fx.pool.ProxyConnection#prepareStatement(Ljava/lang/String;)Ljava/sql/PreparedStatement;",
		"fx.pool.ProxyStatement#close()V",
	}
	if strings.Join(reached, "\n") != strings.Join(wantReached, "\n") {
		t.Errorf("reached bodies the JVM ran:\n%s\nwant:\n%s", strings.Join(reached, "\n"), strings.Join(wantReached, "\n"))
	}
	for _, b := range missed {
		if !strings.HasPrefix(b, "fx.drv.") {
			t.Errorf("the JVM ran %s, the graph does not reach it from main (only the driver's bodies are a known miss)", b)
		}
	}
	if len(missed) != 8 {
		t.Errorf("missed %d driver bodies (%v), want all 8 — if the DriverManager → Driver#connect gap closed, update this oracle", len(missed), missed)
	}

	for n := range reach {
		if strings.HasPrefix(n, "fx.other.") || strings.HasPrefix(n, "fx.drv.UnusedStatement#") {
			t.Errorf("%s is reached: nothing reachable creates it and the JVM never ran it", n)
		}
	}
	// Every jdbc:objects edge ends at a pool body the JVM ran.
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisModel) && strings.HasPrefix(id.To, "fx.") && !obs[id.To] {
			t.Errorf("model edge to a body the JVM never ran: %s -> %s", id.From, id.To)
		}
	}
}
