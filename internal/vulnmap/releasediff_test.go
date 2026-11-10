package vulnmap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

// releaseDiffServer serves a fake Maven Central with testdata/runtime's
// rtm.jar as fix:rtm 1.0.0 and rtm-v2.jar (one method of rtm.Square
// changed) as 1.0.1.
func releaseDiffServer(t *testing.T) *Resolver {
	t.Helper()
	v1, err := os.ReadFile("../../testdata/runtime/rtm.jar")
	if err != nil {
		t.Fatal(err)
	}
	v2, err := os.ReadFile("../../testdata/runtime/rtm-v2.jar")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/fix/rtm/maven-metadata.xml", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "ariadne") {
			w.WriteHeader(http.StatusTeapot)
			return
		}
		w.Write([]byte(`<metadata><versioning><versions><version>0.9</version><version>1.0.0</version><version>1.0.1-RC1</version><version>1.0.1</version><version>1.2</version><version>2.0.2</version><version>2.0.3.Final</version><version>3.0.0</version></versions></versioning></metadata>`))
	})
	mux.HandleFunc("/fix/rtm/1.0.0/rtm-1.0.0.jar", func(w http.ResponseWriter, r *http.Request) { w.Write(v1) })
	mux.HandleFunc("/fix/rtm/1.0.1/rtm-1.0.1.jar", func(w http.ResponseWriter, r *http.Request) { w.Write(v2) })
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	r := NewResolver(nil, mustCache(t))
	r.HTTPClient = srv.Client()
	r.CentralBaseURL = srv.URL
	r.LocalMavenRepo = filepath.Join(t.TempDir(), "none")
	r.ReleaseDiff = true
	old := centralDelay
	centralDelay = time.Millisecond
	t.Cleanup(func() { centralDelay = old })
	return r
}

func releaseDiffCVE(text string) cves.CVE {
	return cves.CVE{
		ID:       "CVE-F10-RD",
		Summary:  text,
		Affected: []cves.Affected{{Package: "fix:rtm", Version: "0.9"}},
		Sources: []cves.Source{{ID: "GHSA-x", Ranges: []cves.AffectedRange{{Package: "fix:rtm", Type: "ECOSYSTEM",
			Events: []map[string]string{{"introduced": "0"}, {"fixed": "1.0.1"}, {"introduced": "2.0"}, {"fixed": "2.0.3"}}}}}},
	}
}

// The methods whose code changed between 1.0.0 (the release before the
// fix) and 1.0.1, in the class the advisory names; unchanged classes and
// unnamed ones give nothing.
func TestReleaseDiff_changedMethodsOfNamedClass(t *testing.T) {
	r := releaseDiffServer(t)
	e, ok, err := r.Resolve(context.Background(), releaseDiffCVE("Area computation in the SquareShape is wrong; see `rtm.Square#area` and LoudGreeter. A plain Square is not code."))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Source != SourceReleaseDiff || e.Review != ReviewUnreviewed || e.Role != RoleVulnerable || e.Coverage != CoveragePartial {
		t.Errorf("entry meta = %+v", e)
	}
	if len(e.Methods) == 0 {
		t.Fatal("no methods")
	}
	for _, m := range e.Methods {
		if !strings.HasPrefix(m, "rtm.Square#") || !strings.Contains(m, "(") {
			t.Errorf("method %q: want an exact selector of rtm.Square", m)
		}
	}
	rd := e.ReleaseDiffs[0]
	if rd.Before != "1.0.0" || rd.Fixed != "1.0.1" || strings.Join(rd.Classes, ",") != "rtm.LoudGreeter,rtm.Square" || rd.Changed != len(e.Methods) {
		t.Errorf("evidence = %+v (LoudGreeter is named but unchanged)", rd)
	}
}

func TestReleaseDiff_nothingToSay(t *testing.T) {
	r := releaseDiffServer(t)
	for name, text := range map[string]string{
		"no class named":        "A denial of service in the shapes library; a Square is fine.",
		"named class unchanged": "The LoudGreeter shouts.",
	} {
		_, ok, err := r.Resolve(context.Background(), releaseDiffCVE(text))
		if err != nil || ok {
			t.Errorf("%s: ok=%v err=%v, want no mapping", name, ok, err)
		}
	}
	// Disabled: the source is not consulted.
	r.ReleaseDiff = false
	if _, ok, _ := r.Resolve(context.Background(), releaseDiffCVE("see `Square`")); ok {
		t.Error("release diff ran while disabled")
	}
}

// The fixed release is matched to its published name, and a fix first
// released in a new minor line gets no predecessor.
func TestReleaseBefore(t *testing.T) {
	r := releaseDiffServer(t)
	ctx := context.Background()
	if pub, before, err := r.releaseBefore(ctx, "fix:rtm", "2.0.3"); err != nil || pub != "2.0.3.Final" || before != "2.0.2" {
		t.Errorf("2.0.3: %q %q %v", pub, before, err)
	}
	if pub, before, err := r.releaseBefore(ctx, "fix:rtm", "1.0.1"); err != nil || pub != "1.0.1" || before != "1.0.0" {
		t.Errorf("1.0.1: %q %q %v (the RC is not a base)", pub, before, err)
	}
	if _, _, err := r.releaseBefore(ctx, "fix:rtm", "3.0.0"); err == nil || !strings.Contains(err.Error(), "new minor release") {
		t.Errorf("3.0.0: %v", err)
	}
}

func TestFixedAfterAndNames(t *testing.T) {
	c := releaseDiffCVE("UriComponentsBuilder and HttpURLConnection; not Spring or JSON, but `Yaml#load` and `org.x.Constructor`.")
	if got := fixedAfter(c, "fix:rtm", "0.9"); got != "1.0.1" {
		t.Errorf("fixedAfter 0.9 = %q", got)
	}
	if got := fixedAfter(c, "fix:rtm", "2.0.1"); got != "2.0.3" {
		t.Errorf("fixedAfter 2.0.1 = %q", got)
	}
	if got := fixedAfter(c, "fix:rtm", "2.0.3"); got != "" {
		t.Errorf("fixedAfter 2.0.3 = %q, want none", got)
	}
	if got := strings.Join(advisoryClassNames(c), ","); got != "Constructor,HttpURLConnection,UriComponentsBuilder,Yaml" {
		t.Errorf("names = %q", got)
	}
}

// Offline, a jar the cache lacks comes from the local Maven repository;
// metadata never does (a local repository lists only what it resolved).
func TestReleaseDiff_offlineLocalRepository(t *testing.T) {
	online := releaseDiffServer(t)
	// Warm only the metadata into a cache, then go offline.
	dir := t.TempDir()
	warm, err := httpcache.New(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	online.Cache = warm
	if _, _, err := online.releaseBefore(context.Background(), "fix:rtm", "1.0.1"); err != nil {
		t.Fatal(err)
	}
	off, err := httpcache.New(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	local := t.TempDir()
	for v, src := range map[string]string{"1.0.0": "rtm.jar", "1.0.1": "rtm-v2.jar"} {
		b, _ := os.ReadFile("../../testdata/runtime/" + src)
		p := filepath.Join(local, "fix", "rtm", v, "rtm-"+v+".jar")
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, b, 0o644)
	}
	r := NewResolver(nil, off)
	r.ReleaseDiff, r.LocalMavenRepo, r.CentralBaseURL = true, local, online.CentralBaseURL
	e, ok, err := r.Resolve(context.Background(), releaseDiffCVE("see `rtm.Square`"))
	if err != nil || !ok || e.Source != SourceReleaseDiff {
		t.Fatalf("offline with a local repository: ok=%v err=%v %+v", ok, err, e)
	}
	// Without the local jars, offline is the usual contract error.
	r.LocalMavenRepo = filepath.Join(t.TempDir(), "none")
	if _, _, err := r.Resolve(context.Background(), releaseDiffCVE("see `rtm.Square`")); err == nil || !errors.Is(err, httpcache.ErrOffline) {
		t.Errorf("offline without jars: %v, want ErrOffline", err)
	}
}
