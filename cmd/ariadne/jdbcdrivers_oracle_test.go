package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
)

const jdbcDriversFixture = "../../testdata/jdbcdrivers/"

// TestJDBCDriversOracle (jdbc:drivers): testdata/jdbcdrivers
// (scripts/jdbcdrivers-oracle.sh) opens jdbc:fx:oracle through
// DriverManager with pgjdbc 42.2.18 and MySQL Connector/J 8.0.23 on the
// class path. The JVM (traced through JDI) enters each one's connect,
// which checks the URL and returns null; nothing past the check runs.
//
// Every pgjdbc/Connector/J body the JVM ran is reached from main (the
// model never cuts what runs), and the bodies past the URL check —
// URL parsing, the logger setup of CVE-2022-26520, the connection — are
// not: the application's only URL is jdbc:fx:.
func TestJDBCDriversOracle(t *testing.T) {
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", jdbcDriversFixture + "app.jar",
		"-deps", jdbcDriversFixture + "deps", "-profile", "generic", "-roots", "profile",
		"-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-graph-out", graph}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr, func(string) string { return "" }); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "cut after the URL check org.postgresql.Driver, com.mysql.cj.jdbc.NonRegisteringDriver") {
		t.Errorf("log does not report both drivers cut:\n%s", grepLines(out, "JDBC drivers"))
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
	obs := (benchFixture{observed: jdbcDriversFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	if len(obs) < 40 {
		t.Fatalf("observed.txt has %d bodies; regenerate with scripts/jdbcdrivers-oracle.sh", len(obs))
	}
	for _, connect := range []string{
		"org.postgresql.Driver#connect(Ljava/lang/String;Ljava/util/Properties;)Ljava/sql/Connection;",
		"com.mysql.cj.jdbc.NonRegisteringDriver#connect(Ljava/lang/String;Ljava/util/Properties;)Ljava/sql/Connection;",
		"com.mysql.cj.conf.ConnectionUrl#acceptsUrl(Ljava/lang/String;)Z",
	} {
		if !obs[connect] {
			t.Errorf("the JVM did not run %s; the fixture no longer shows the URL check", connect)
		}
	}
	// The cut must keep everything the URL check runs. Bodies the JVM ran
	// only at class initialization or in the cleanup thread a static
	// initializer starts are outside what this model touches; they are
	// checked where the graph reaches the connect that follows them.
	for b := range obs {
		if strings.Contains(b, "#connect(") || strings.HasPrefix(b, "com.mysql.cj.conf.ConnectionUrl#acceptsUrl") {
			if !reach[b] {
				t.Errorf("the JVM ran %s, the graph does not reach it from main", b)
			}
		}
	}
	for _, past := range []string{
		"org.postgresql.Driver#getDefaultProperties()Ljava/util/Properties;",
		"org.postgresql.Driver#parseURL(Ljava/lang/String;Ljava/util/Properties;)Ljava/util/Properties;",
		"org.postgresql.Driver#setupLoggerFromProperties(Ljava/util/Properties;)V",
		"org.postgresql.Driver#makeConnection(Ljava/lang/String;Ljava/util/Properties;)Ljava/sql/Connection;",
		"com.mysql.cj.conf.ConnectionUrl#getConnectionUrlInstance(Ljava/lang/String;Ljava/util/Properties;)Lcom/mysql/cj/conf/ConnectionUrl;",
		"com.mysql.cj.jdbc.ConnectionImpl#getInstance(Lcom/mysql/cj/conf/HostInfo;)Lcom/mysql/cj/jdbc/JdbcConnection;",
	} {
		if obs[past] {
			t.Errorf("the JVM ran %s past the URL check; the fixture is wrong", past)
		}
		if reach[past] {
			t.Errorf("%s is reached past the URL check", past)
		}
	}
	// Nothing of either driver past its URL check: no class of pgjdbc's
	// connection or Connector/J's is reached.
	for n := range reach {
		if strings.HasPrefix(n, "org.postgresql.jdbc.PgConnection#") || strings.HasPrefix(n, "com.mysql.cj.jdbc.ConnectionImpl#") {
			t.Errorf("%s is reached", n)
		}
	}
}

func grepLines(s, sub string) string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.Contains(l, sub) {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
