package httpcache

import (
	"fmt"
	"io"
	"net/http"
)

// Fetch returns the cached body for key if present; otherwise it performs
// req (which the caller must have already built, including its body) and
// caches the response body before returning it. A non-2xx response is
// returned as an error and is not cached.
//
// Offline mode is handled by Get: a miss returns ErrOffline without ever
// touching the network.
func Fetch(client *http.Client, cache *Cache, key string, req *http.Request) ([]byte, error) {
	if b, hit, err := cache.Get(key); hit || err != nil {
		return b, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", req.Method, req.URL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body for %s %s: %w", req.Method, req.URL, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s: HTTP %d: %s", req.Method, req.URL, resp.StatusCode, truncate(body, 500))
	}

	if err := cache.Set(key, body); err != nil {
		return nil, err
	}
	return body, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
