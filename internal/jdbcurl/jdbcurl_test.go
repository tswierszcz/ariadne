package jdbcurl

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

func jar(t *testing.T, files map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "app.jar")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	for name, body := range files {
		w, _ := z.Create(name)
		w.Write([]byte(body))
	}
	z.Close()
	f.Close()
	return p
}

func prefixes(u URLs) []string {
	var out []string
	for p := range u.Prefixes {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// Collect reads URLs from every configuration file (all profiles),
// cut at a placeholder; a placeholder the packaged configuration does
// not set, or a JNDI DataSource, makes the URLs unbounded.
func TestCollect(t *testing.T) {
	u, err := Collect(jar(t, map[string]string{
		"application.properties":       "spring.datasource.url=jdbc:hsqldb:hsql://${server.address}:${port}/db\n",
		"application-mysql.properties": "spring.datasource.url=jdbc:mysql://${MYSQL_URL:localhost}/petclinic\n",
		"config/app.yml":               "db:\n  url: ${DB_URL:jdbc:h2:mem:x}\n",
		"README.txt":                   "jdbc:postgresql: is not read from a .txt",
	}), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := prefixes(u), []string{"jdbc:h2:mem", "jdbc:hsqldb:hsql", "jdbc:mysql:"}; !reflect.DeepEqual(got, want) || len(u.Unbounded) != 0 {
		t.Errorf("prefixes %v unbounded %v, want %v and none", got, u.Unbounded, want)
	}

	// A URL from a key the packaged configuration sets is bounded by
	// that value; an unset one is not.
	u, _ = Collect(jar(t, map[string]string{
		"jdbc.properties": "jdbc.url=jdbc:hsqldb:mem:db\n",
		"beans.xml":       `<property name="url" value="${jdbc.url}"/>`,
	}), nil, nil)
	if len(u.Unbounded) != 0 || u.Prefixes["jdbc:hsqldb:mem"] == "" {
		t.Errorf("set placeholder: %v %v", prefixes(u), u.Unbounded)
	}
	u, _ = Collect(jar(t, map[string]string{"application.yml": "spring:\n  datasource:\n    url: ${DATABASE_URL}\n"}), nil, nil)
	if len(u.Unbounded) != 1 || !strings.Contains(u.Unbounded[0], "${DATABASE_URL}") {
		t.Errorf("unset placeholder: %v", u.Unbounded)
	}
	u, _ = Collect(jar(t, map[string]string{
		"application.yml": "spring:\n  datasource:\n    url: ${db.url}\ndb:\n  url: jdbc:postgresql://x/y\n",
	}), nil, nil)
	if len(u.Unbounded) != 0 || u.Prefixes["jdbc:postgresql:"] == "" {
		t.Errorf("YAML-set placeholder: %v %v", prefixes(u), u.Unbounded)
	}
	u, _ = Collect(jar(t, map[string]string{"application.properties": "spring.datasource.jndi-name=java:comp/env/jdbc/x\n"}), nil, nil)
	if len(u.Unbounded) == 0 {
		t.Error("a JNDI DataSource left the URLs bounded")
	}
	u, _ = Collect(jar(t, map[string]string{}), map[string][]byte{"WEB-INF/web.xml": []byte("<param-value>jdbc:derby:memory:x</param-value>")}, nil)
	if u.Prefixes["jdbc:derby:memory"] == "" {
		t.Errorf("web.xml: %v", prefixes(u))
	}
}

// Decide against the real drivers of testdata/jdbcdrivers: a driver
// some URL beginning reaches is kept (also a beginning shorter than its
// prefix, "jdbc:"), every other catalog driver is cut after its URL
// check, and unbounded URLs keep every driver.
func TestDecide(t *testing.T) {
	deps := []input.DepJar{
		{Path: "../../testdata/jdbcdrivers/deps/postgresql-42.2.18.jar"},
		{Path: "../../testdata/jdbcdrivers/deps/mysql-connector-java-8.0.23.jar"},
	}
	h, _, err := hierarchy.Build("../../testdata/jdbcdrivers/app.jar", deps)
	if err != nil {
		t.Fatal(err)
	}
	d := Decide(h, URLs{Prefixes: map[string]string{"jdbc:fx:oracle": "Main"}})
	if !d.Applies || len(d.Kept) != 0 || len(d.Cuts) != 2 || d.Cuts[0].Driver != "org.postgresql.Driver" || d.Cuts[0].PC != 17 || d.Cuts[1].PC != 1 {
		t.Errorf("fx only: %+v", d)
	}
	d = Decide(h, URLs{Prefixes: map[string]string{"jdbc:mysql:": "x"}})
	if !reflect.DeepEqual(d.Kept, []string{"com.mysql.cj.jdbc.NonRegisteringDriver"}) || len(d.Cuts) != 1 || d.Cuts[0].Driver != "org.postgresql.Driver" {
		t.Errorf("mysql: %+v", d)
	}
	d = Decide(h, URLs{Prefixes: map[string]string{"jdbc:": "x"}})
	if len(d.Kept) != 2 || len(d.Cuts) != 0 {
		t.Errorf("bare jdbc: %+v", d)
	}
	d = Decide(h, URLs{Prefixes: map[string]string{}, Unbounded: []string{"x"}})
	if d.Applies || !strings.Contains(d.Why, "not bounded") {
		t.Errorf("unbounded: %+v", d)
	}
}
