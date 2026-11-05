package httpcache

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestCacheSetGet(t *testing.T) {
	c, err := New(t.TempDir(), false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	key := Key("GET", "https://example/x")

	if _, hit, err := c.Get(key); hit || err != nil {
		t.Fatalf("expected a plain miss, got hit=%v err=%v", hit, err)
	}
	if err := c.Set(key, []byte(`{"a":1}`)); err != nil {
		t.Fatalf("Set: %v", err)
	}
	b, hit, err := c.Get(key)
	if err != nil || !hit {
		t.Fatalf("Get after Set: hit=%v err=%v", hit, err)
	}
	if string(b) != `{"a":1}` {
		t.Fatalf("Get returned %q", b)
	}
}

func TestCacheOfflineMiss(t *testing.T) {
	c, err := New(t.TempDir(), true)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, hit, err := c.Get(Key("GET", "https://example/missing"))
	if hit {
		t.Fatalf("expected a miss")
	}
	if !errors.Is(err, ErrOffline) {
		t.Fatalf("got err=%v, want ErrOffline", err)
	}
}

func TestFetchCachesResponse(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	c, err := New(dir, false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	key := Key("GET", srv.URL)

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
		b, err := Fetch(srv.Client(), c, key, KindOSV, req)
		if err != nil {
			t.Fatalf("Fetch[%d]: %v", i, err)
		}
		if string(b) != `{"ok":true}` {
			t.Fatalf("Fetch[%d] = %q", i, b)
		}
	}
	if calls != 1 {
		t.Fatalf("server called %d times, want 1 (second Fetch should hit the cache)", calls)
	}
}

func TestFetchOfflineMissFailsWithoutNetworkCall(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer srv.Close()

	c, err := New(t.TempDir(), true)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	_, err = Fetch(srv.Client(), c, Key("GET", srv.URL), KindOSV, req)
	if !errors.Is(err, ErrOffline) {
		t.Fatalf("got err=%v, want ErrOffline", err)
	}
	if calls != 0 {
		t.Fatalf("server was called %d times, want 0 in offline mode", calls)
	}
}

func TestCacheSetWritesUnderDir(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir, false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	key := Key("a", "b")
	if err := c.Set(key, []byte("x")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 2 || entries[0].Name() != key+".json" || entries[1].Name() != key+".meta.json" {
		t.Fatalf("dir entries = %v, want exactly the body and its metadata sidecar", entries)
	}
}

// TestCacheConcurrentSetSameKey reproduces a real failure found running
// ariadne end-to-end: two CVEs can share the same GitHub fix-commit URL
// (one commit fixing several CVEs), so two of ResolveAll's worker
// goroutines legitimately call Set for the *same* key concurrently. A
// shared "key+.tmp" temp file name let one goroutine's rename steal the
// other's temp file out from under it, failing with "no such file or
// directory" — run with -race to also catch a data race on the file
// itself, not just the renamed-away error.
func TestCacheConcurrentSetSameKey(t *testing.T) {
	c, err := New(t.TempDir(), false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	key := Key("GET", "https://example/shared")

	const n = 20
	errs := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs <- c.Set(key, []byte(fmt.Sprintf(`{"n":%d}`, i)))
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Errorf("concurrent Set: %v", err)
		}
	}

	b, hit, err := c.Get(key)
	if err != nil || !hit {
		t.Fatalf("Get after concurrent Set: hit=%v err=%v", hit, err)
	}
	if len(b) == 0 {
		t.Errorf("final cache entry is empty")
	}
}

func TestGetTracesHitsWhenRequested(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Set("k1", []byte("{}")); err != nil {
		t.Fatal(err)
	}
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv("ARIADNE_CACHE_TRACE", trace)
	if _, ok, err := c.Get("k1"); err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if _, _, err := c.Get("missing"); err == nil {
		t.Fatal("offline miss must still error")
	}
	b, err := os.ReadFile(trace)
	if err != nil || string(b) != "k1\n" {
		t.Errorf("trace = %q (%v), want only the hit", b, err)
	}
}
