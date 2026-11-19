package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const jettyFixture = "../../testdata/library-lb4/jetty-client/"

// TestLibraryJetty (LIBRARY_MODE LB4): the callback-style library of the
// qualification set (testdata/library-lb4/EXPECTED.MD), jetty-client
// 9.4.36 with its own jars and the OSV/GitHub cache entries the run reads.
// A user who calls only HttpClient#GET runs both mapped vulnerable methods
// on the JVM (observed.txt, scripts/library-lb4-oracle.sh): the request
// URI through HttpURI#parse (CVE-2022-2047), the response's "+5"
// Content-Length through HttpParser#convertContentLength (CVE-2023-40167,
// after the NIO selector hands the read to the receiver). Both are
// REACHABLE with HttpClient#GET exposed; CVE-2024-6763 (Jetty 12 mapping,
// and by its advisory not Jetty's own use of HttpURI) NOT_REACHABLE; the
// two advisories without a fix commit UNKNOWN.
func TestLibraryJetty(t *testing.T) {
	const (
		get    = "org.eclipse.jetty.client.HttpClient#GET(Ljava/lang/String;)Lorg/eclipse/jetty/client/api/ContentResponse;"
		getURI = "org.eclipse.jetty.client.HttpClient#GET(Ljava/net/URI;)Lorg/eclipse/jetty/client/api/ContentResponse;"
	)
	obs := (benchFixture{observed: jettyFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	for _, m := range []string{get, getURI,
		"org.eclipse.jetty.http.HttpParser#convertContentLength(Ljava/lang/String;)J",
		"org.eclipse.jetty.http.HttpURI#parse(Lorg/eclipse/jetty/http/HttpURI$State;Ljava/lang/String;II)V"} {
		if !obs[m] {
			t.Errorf("observed.txt: %s never ran; regenerate with scripts/library-lb4-oracle.sh", m)
		}
	}
	want := map[string]string{
		"CVE-2022-2047":  "REACHABLE",
		"CVE-2023-40167": "REACHABLE",
		"CVE-2024-6763":  "NOT_REACHABLE",
		"CVE-2025-11143": "UNKNOWN",
		"CVE-2026-2332":  "UNKNOWN",
	}
	for _, mode := range []string{"cha", "rta"} {
		dir := t.TempDir()
		out := filepath.Join(dir, "r.json")
		cache := filepath.Join(dir, "cache")
		copyDir(t, jettyFixture+"cache", cache)
		args := []string{"-offline", "-cache-dir", cache, "-app", jettyFixture + "lib.jar",
			"-deps", jettyFixture + "deps", "-profile", "generic", "-roots", "api", "-callgraph-mode", mode,
			"-json-format", "v2", "-out", filepath.Join(dir, "r.md"), "-json", out}
		env := map[string]string{"ARIADNE_MAVEN_LOCAL": t.TempDir()}
		if code, stderr := runWithEnv(args, env); code != 0 {
			t.Fatalf("%s: exit %d: %s", mode, code, stderr)
		}
		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		var doc struct {
			Findings []struct {
				AdvisoryID   string `json:"advisory_id"`
				Reachability string `json:"reachability"`
				ExposedAPI   []struct {
					Method string `json:"method"`
				} `json:"exposed_api"`
			} `json:"findings"`
		}
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		got := map[string]string{}
		for _, f := range doc.Findings {
			got[f.AdvisoryID] = f.Reachability
			if f.Reachability != "REACHABLE" {
				continue
			}
			exposed := map[string]bool{}
			ran := 0
			for _, x := range f.ExposedAPI {
				exposed[x.Method] = true
				if obs[x.Method] {
					ran++
				}
			}
			if !exposed[get] || !exposed[getURI] {
				t.Errorf("%s: %s: HttpClient#GET not exposed (%d entries)", mode, f.AdvisoryID, len(f.ExposedAPI))
			}
			t.Logf("%s: %s: %d exposed API methods, %d ran in the GET scenario", mode, f.AdvisoryID, len(f.ExposedAPI), ran)
		}
		for id, w := range want {
			if got[id] != w {
				t.Errorf("%s: %s = %q, want %s", mode, id, got[id], w)
			}
		}
		if len(got) != len(want) {
			t.Errorf("%s: findings %v, want %v", mode, got, want)
		}
	}
}
