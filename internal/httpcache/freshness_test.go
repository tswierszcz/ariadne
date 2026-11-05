package httpcache

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// server answers with body() and counts calls; fail makes it return 500.
type server struct {
	*httptest.Server
	calls atomic.Int32
	fail  atomic.Bool
	body  atomic.Value // string
	etag  string
}

func newServer(t *testing.T, body, etag string) *server {
	s := &server{etag: etag}
	s.body.Store(body)
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.calls.Add(1)
		if s.fail.Load() {
			http.Error(w, "down", http.StatusInternalServerError)
			return
		}
		if s.etag != "" {
			if r.Header.Get("If-None-Match") == s.etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("ETag", s.etag)
		}
		w.Write([]byte(s.body.Load().(string)))
	}))
	t.Cleanup(s.Close)
	return s
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newCache(t *testing.T, dir string, offline bool, clk *clock, p Policy) *Cache {
	t.Helper()
	c, err := New(dir, offline)
	if err != nil {
		t.Fatal(err)
	}
	c.SetClock(clk.now)
	c.SetPolicy(p)
	return c
}

func get(t *testing.T, c *Cache, s *server, kind Kind) (string, error) {
	t.Helper()
	req, _ := http.NewRequest("GET", s.URL+"/q", nil)
	b, err := Fetch(s.Client(), c, Key("GET", s.URL+"/q"), kind, req)
	return string(b), err
}

// ARIADNE-TEST-0401: an empty OSV answer is refreshed after the TTL and
// the new advisory appears; within the TTL the cache answers.
func TestTTLRefreshesEmptyAnswer(t *testing.T) {
	dir, clk := t.TempDir(), &clock{time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)}
	s := newServer(t, `{"results":[{}]}`, "")
	c := newCache(t, dir, false, clk, Policy{})
	if b, _ := get(t, c, s, KindOSV); b != `{"results":[{}]}` {
		t.Fatalf("first fetch = %q", b)
	}
	s.body.Store(`{"results":[{"vulns":[{"id":"GHSA-new"}]}]}`)
	clk.t = clk.t.Add(DefaultTTL) // exactly at the boundary: still fresh
	if b, _ := get(t, c, s, KindOSV); !strings.Contains(b, `{}`) || s.calls.Load() != 1 {
		t.Fatalf("within TTL: %q, calls %d", b, s.calls.Load())
	}
	clk.t = clk.t.Add(time.Second)
	if b, _ := get(t, c, s, KindOSV); !strings.Contains(b, "GHSA-new") || s.calls.Load() != 2 {
		t.Fatalf("after TTL: %q, calls %d", b, s.calls.Load())
	}
	f := c.Freshness()
	if len(f) != 1 || f[0].Fetched != 1 || f[0].Fresh != 1 || f[0].Refreshed != 1 {
		t.Errorf("freshness = %+v", f)
	}
}

func TestPinnedAndLLMNeverExpire(t *testing.T) {
	dir, clk := t.TempDir(), &clock{time.Now()}
	s := newServer(t, "diff", "")
	c := newCache(t, dir, false, clk, Policy{Refresh: true})
	get(t, c, s, KindPinned)
	clk.t = clk.t.Add(365 * 24 * time.Hour)
	if _, err := get(t, c, s, KindPinned); err != nil || s.calls.Load() != 1 {
		t.Errorf("pinned refetched: calls %d err %v", s.calls.Load(), err)
	}
	if KindForURL("https://github.com/o/r/commit/0123456789abcdef0123456789abcdef01234567.diff") != KindPinned ||
		KindForURL("https://raw.githubusercontent.com/o/r/0123456789abcdef0123456789abcdef01234567/src/A.java") != KindPinned ||
		KindForURL("https://github.com/o/r/commit/0123456.diff") != KindMutable {
		t.Error("KindForURL")
	}
}

func TestRevalidateWithETag(t *testing.T) {
	dir, clk := t.TempDir(), &clock{time.Now()}
	s := newServer(t, "body", `"v1"`)
	c := newCache(t, dir, false, clk, Policy{})
	get(t, c, s, KindOSV)
	clk.t = clk.t.Add(2 * DefaultTTL)
	if b, err := get(t, c, s, KindOSV); b != "body" || err != nil {
		t.Fatalf("revalidate: %q %v", b, err)
	}
	clk.t = clk.t.Add(time.Hour) // validated_at moved: fresh again
	get(t, c, s, KindOSV)
	if f := c.Freshness()[0]; f.Revalidated != 1 || f.Fresh != 1 || s.calls.Load() != 2 {
		t.Errorf("freshness %+v calls %d", f, s.calls.Load())
	}
}

// ARIADNE-TEST-0403: a failed refresh is an error unless -allow-stale,
// which serves the copy and records the failure.
func TestFailedRefresh(t *testing.T) {
	dir, clk := t.TempDir(), &clock{time.Now()}
	s := newServer(t, "old", "")
	get(t, newCache(t, dir, false, clk, Policy{}), s, KindOSV)
	clk.t = clk.t.Add(48 * time.Hour)
	s.fail.Store(true)
	if _, err := get(t, newCache(t, dir, false, clk, Policy{}), s, KindOSV); err == nil || !strings.Contains(err.Error(), "-allow-stale") {
		t.Errorf("no -allow-stale: err = %v", err)
	}
	c := newCache(t, dir, false, clk, Policy{AllowStale: true})
	if b, err := get(t, c, s, KindOSV); b != "old" || err != nil {
		t.Fatalf("-allow-stale: %q %v", b, err)
	}
	if f := c.Freshness()[0]; f.Stale != 1 || len(f.RefreshFailures) != 1 || !f.OldestValidated.Equal(clk.t.Add(-48*time.Hour).UTC()) {
		t.Errorf("freshness %+v", f)
	}
	// -max-data-age stays hard under -allow-stale.
	c = newCache(t, dir, false, clk, Policy{AllowStale: true, MaxAge: 24 * time.Hour})
	if _, err := get(t, c, s, KindOSV); !errors.Is(err, ErrTooOld) {
		t.Errorf("max age with allow-stale: %v", err)
	}
}

func TestOfflineAgeRules(t *testing.T) {
	dir, clk := t.TempDir(), &clock{time.Now()}
	s := newServer(t, "x", "")
	get(t, newCache(t, dir, false, clk, Policy{}), s, KindOSV)
	clk.t = clk.t.Add(72 * time.Hour)
	c := newCache(t, dir, true, clk, Policy{})
	if b, err := get(t, c, s, KindOSV); b != "x" || err != nil || c.Freshness()[0].Stale != 1 {
		t.Errorf("offline stale: %q %v %+v", b, err, c.Freshness())
	}
	if _, err := get(t, newCache(t, dir, true, clk, Policy{MaxAge: 24 * time.Hour}), s, KindOSV); !errors.Is(err, ErrTooOld) {
		t.Errorf("offline max age: %v", err)
	}
	if s.calls.Load() != 1 {
		t.Errorf("offline touched the network: %d calls", s.calls.Load())
	}
}

// Entries cached before CVE:F2 (no sidecar), or whose sidecar does not
// match the body, are of unknown age: revalidated online, flagged offline,
// refused under -max-data-age. No migration date is invented.
func TestUnknownAgeEntries(t *testing.T) {
	dir, clk := t.TempDir(), &clock{time.Now()}
	s := newServer(t, "new", "")
	key := Key("GET", s.URL+"/q")
	os.WriteFile(filepath.Join(dir, key+".json"), []byte("legacy"), 0o644)

	off := newCache(t, dir, true, clk, Policy{})
	if b, _ := get(t, off, s, KindOSV); b != "legacy" || off.Freshness()[0].UnknownAge != 1 {
		t.Errorf("offline legacy: %q %+v", b, off.Freshness())
	}
	if _, err := get(t, newCache(t, dir, true, clk, Policy{MaxAge: time.Hour}), s, KindOSV); !errors.Is(err, ErrTooOld) {
		t.Errorf("legacy under max age: %v", err)
	}
	on := newCache(t, dir, false, clk, Policy{})
	if b, _ := get(t, on, s, KindOSV); b != "new" || on.Freshness()[0].Refreshed != 1 {
		t.Errorf("online legacy: %q %+v", b, on.Freshness())
	}
	// A sidecar that does not match its body is ignored.
	os.WriteFile(filepath.Join(dir, key+".json"), []byte("tampered"), 0o644)
	if b, _ := get(t, newCache(t, dir, true, clk, Policy{}), s, KindOSV); b != "tampered" {
		t.Errorf("mismatched sidecar body = %q", b)
	}
	if _, err := get(t, newCache(t, dir, true, clk, Policy{MaxAge: 100 * time.Hour}), s, KindOSV); !errors.Is(err, ErrTooOld) {
		t.Errorf("mismatched sidecar treated as known age: %v", err)
	}
}

func TestRefreshRevalidatesAndMetaHasNoSecrets(t *testing.T) {
	dir, clk := t.TempDir(), &clock{time.Now()}
	s := newServer(t, "x", "")
	get(t, newCache(t, dir, false, clk, Policy{}), s, KindOSV)
	c := newCache(t, dir, false, clk, Policy{Refresh: true})
	req, _ := http.NewRequest("GET", s.URL+"/q", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	if _, err := Fetch(s.Client(), c, Key("GET", s.URL+"/q"), KindOSV, req); err != nil || s.calls.Load() != 2 {
		t.Fatalf("-refresh: calls %d err %v", s.calls.Load(), err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, Key("GET", s.URL+"/q")+".meta.json"))
	if strings.Contains(string(b), "secret-token") || strings.Contains(string(b), "Authorization") || !strings.Contains(string(b), `"kind": "osv"`) {
		t.Errorf("meta = %s", b)
	}
}
