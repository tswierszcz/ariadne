package cves

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"ariadne/internal/httpcache"
	"ariadne/internal/input"
)

// TestUnavailableSourceIsAnError (CVE:F9.8): OSV answering HTTP 429 (rate
// limit) or 503 makes the lookup fail — never an empty, "no CVE" answer —
// and the failed answer is not cached: once OSV answers again, the
// advisories are there. The CLI turns this error into a run without any
// output (cmd/ariadne TestGlobalDataErrorWritesNothing).
func TestUnavailableSourceIsAnError(t *testing.T) {
	good := fakeOSV(t)
	defer good.Close()
	for _, code := range []int{http.StatusTooManyRequests, http.StatusServiceUnavailable} {
		var failing atomic.Bool
		failing.Store(true)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if failing.Load() {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "slow down", code)
				return
			}
			req, _ := http.NewRequestWithContext(r.Context(), r.Method, good.URL+r.URL.Path, r.Body)
			req.Header = r.Header
			resp, err := good.Client().Do(req)
			if err != nil {
				http.Error(w, err.Error(), 502)
				return
			}
			defer resp.Body.Close()
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		}))
		cache, err := httpcache.New(t.TempDir(), false)
		if err != nil {
			t.Fatal(err)
		}
		client := &Client{BaseURL: srv.URL, HTTPClient: srv.Client(), Cache: cache}
		gavs := []input.GAV{{Group: "org.yaml", Artifact: "snakeyaml", Version: "1.27"}}
		res, err := client.ResolveAll(context.Background(), gavs)
		if err == nil || !strings.Contains(err.Error(), "HTTP "+strconv.Itoa(code)) || len(res.CVEs) != 0 {
			t.Errorf("HTTP %d: err %v, %d CVEs — want an error and no answer", code, err, len(res.CVEs))
		}
		failing.Store(false)
		res, err = client.ResolveAll(context.Background(), gavs)
		if err != nil || len(res.CVEs) != 1 || res.CVEs[0].ID != "CVE-2022-1471" {
			t.Errorf("HTTP %d then recovery: %v %+v — the failed answer must not have been cached", code, err, res.CVEs)
		}
		srv.Close()
	}
}
