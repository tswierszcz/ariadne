package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"ariadne/internal/httpcache"
)

// TestCentralClient: the local repository serves downloaded files only;
// requests carry ariadne's User-Agent; a failed URL is not asked again
// and after a 429 nothing more is sent.
func TestCentralClient(t *testing.T) {
	var hits atomic.Int32
	var ua atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		ua.Store(r.UserAgent())
		switch {
		case strings.HasSuffix(r.URL.Path, "/ok.pom"):
			w.Write([]byte("<project/>"))
		case strings.HasSuffix(r.URL.Path, "/missing.pom"), strings.HasSuffix(r.URL.Path, ".jar"):
			http.NotFound(w, r)
		default:
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}))
	defer srv.Close()
	local := t.TempDir()
	dir := filepath.Join(local, "g", "a", "1")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "a-1.pom"), []byte("<local/>"), 0o644)
	os.WriteFile(filepath.Join(dir, "a-1.jar"), []byte("installed"), 0o644)
	os.WriteFile(filepath.Join(dir, "_remote.repositories"), []byte("#x\na-1.pom>central=\na-1.jar>=\n"), 0o644)
	cache, err := httpcache.New(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	c := &centralClient{cache: cache, base: srv.URL, local: local, userAgent: "ariadne/test"}
	ctx := context.Background()

	if b, err := c.fetch(ctx, srv.URL+"/g/a/1/a-1.pom"); err != nil || string(b) != "<local/>" || hits.Load() != 0 {
		t.Fatalf("downloaded pom: %q %v, %d requests", b, err, hits.Load())
	}
	if _, err := c.fetch(ctx, srv.URL+"/g/a/1/a-1.jar"); err == nil || hits.Load() != 1 {
		t.Fatalf("installed jar must not be served locally: %v, %d requests", err, hits.Load())
	}
	if ua.Load() != "ariadne/test" {
		t.Errorf("User-Agent %v", ua.Load())
	}
	if _, err := c.fetch(ctx, srv.URL+"/x/ok.pom"); err != nil {
		t.Fatal(err)
	}
	c.fetch(ctx, srv.URL+"/x/missing.pom")
	n := hits.Load()
	if _, err := c.fetch(ctx, srv.URL+"/x/missing.pom"); err == nil || hits.Load() != n {
		t.Errorf("failed URL asked again: %d -> %d", n, hits.Load())
	}
	if _, err := c.fetch(ctx, srv.URL+"/z/limited.pom"); err == nil || hits.Load() != n+1 {
		t.Fatalf("429: %v", err)
	}
	n = hits.Load()
	if _, err := c.fetch(ctx, srv.URL+"/y/new.pom"); err == nil || !strings.Contains(err.Error(), "429") || hits.Load() != n {
		t.Errorf("request after 429: %v, %d -> %d", err, n, hits.Load())
	}
	if st := c.stats; st.Local != 1 || st.Network != 1 || st.Failed != 3 || st.Skipped != 2 {
		t.Errorf("stats %+v", st)
	}
}
