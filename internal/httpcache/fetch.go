package httpcache

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// fullSHA matches a 40-hex commit SHA as a URL path segment (optionally
// with a ".diff"/".patch" suffix): GitHub content at such a SHA is
// immutable.
var fullSHA = regexp.MustCompile(`/[0-9a-f]{40}(\.diff|\.patch)?(/|$)`)

// centralRelease matches a release (non-SNAPSHOT) artifact on Maven
// Central: Central never changes a published release file (CVE:F6a
// reference jars, CVE:F6b release poms). maven-metadata.xml (the version
// list) is not matched: it grows with every release.
var centralRelease = regexp.MustCompile(`^https://repo1\.maven\.org/maven2/.+/[^/]+/[^/]+/[^/]+\.(jar|pom)$`)

// KindForURL is KindPinned for a URL addressed by a full commit SHA or a
// Maven Central release jar or pom, and KindMutable otherwise.
func KindForURL(url string) Kind {
	if fullSHA.MatchString(url) || (centralRelease.MatchString(url) && !strings.Contains(url, "-SNAPSHOT")) {
		return KindPinned
	}
	return KindMutable
}

// Fetch returns the body for key, applying the cache's freshness policy
// to kind (CVE:F2):
//   - a pinned entry is served whenever present;
//   - an expiring entry is served while it is fresh (validated within the
//     TTL, no -refresh); otherwise it is fetched again — conditionally when
//     it carries HTTP validators — and the new copy replaces it;
//   - offline, a present entry is served whatever its age (counted as
//     stale or unknown-age), unless MaxAge forbids it; a missing one is
//     ErrOffline;
//   - online, a failed refresh is an error unless AllowStale, which
//     serves the cached copy and records the failure.
//
// req must already be built, body included. A non-2xx response other than
// 304 is an error and is not cached. req's headers are never stored.
func Fetch(client *http.Client, cache *Cache, key string, kind Kind, req *http.Request) ([]byte, error) {
	if cache.replay != nil {
		b, _, err := cache.fromReplay(key, kind)
		return b, err
	}
	b, meta, err := fetch(client, cache, key, kind, req)
	if err != nil {
		cache.observeErr(key, kind, err)
		return nil, err
	}
	cache.observeHit(key, kind, b, meta)
	return b, nil
}

// fetch is Fetch without observation; meta is the entry's metadata when
// known.
func fetch(client *http.Client, cache *Cache, key string, kind Kind, req *http.Request) ([]byte, *Meta, error) {
	c := cache
	body, err := os.ReadFile(c.path(key))
	present := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, nil, fmt.Errorf("reading cache entry %s: %w", key, err)
	}
	var meta Meta
	known := false
	if present {
		meta, known = c.readMeta(key, body)
	}
	metaPtr := func() *Meta {
		if known {
			m := meta
			return &m
		}
		return nil
	}
	now := c.now()
	age := now.Sub(meta.ValidatedAt)
	p := c.policy

	if present && (!kind.expires() || (known && !p.Refresh && age <= p.effectiveTTL())) {
		traceHit(key)
		c.stats.record(kind, func(f *Freshness) {
			f.Fresh++
			if kind.expires() {
				f.noteValidated(meta.ValidatedAt)
			}
		})
		return body, metaPtr(), nil
	}

	serveStale := func(cause error) ([]byte, *Meta, error) {
		if !p.allowed(age, known) {
			if known {
				return nil, nil, fmt.Errorf("%w: %s is %s old (limit %s)", ErrTooOld, key, age.Round(1e9), p.MaxAge)
			}
			return nil, nil, fmt.Errorf("%w: %s has no recorded fetch time (cached before CVE:F2)", ErrTooOld, key)
		}
		traceHit(key)
		c.stats.record(kind, func(f *Freshness) {
			if known {
				f.Stale++
				f.noteValidated(meta.ValidatedAt)
			} else {
				f.UnknownAge++
			}
			if cause != nil {
				f.RefreshFailures = append(f.RefreshFailures, cause.Error())
			}
		})
		return body, metaPtr(), nil
	}

	if c.offline {
		if !present {
			return nil, nil, fmt.Errorf("%w: %s", ErrOffline, key)
		}
		return serveStale(nil)
	}

	if present && known {
		if meta.ETag != "" {
			req.Header.Set("If-None-Match", meta.ETag)
		}
		if meta.LastModified != "" {
			req.Header.Set("If-Modified-Since", meta.LastModified)
		}
	}
	fresh, resp, ferr := do(client, req)
	if ferr != nil {
		if present && p.AllowStale {
			return serveStale(ferr)
		}
		if present {
			return nil, nil, fmt.Errorf("refreshing cached %s (pass -allow-stale to use the cached copy): %w", key, ferr)
		}
		return nil, nil, ferr
	}
	if resp.StatusCode == http.StatusNotModified && present && known {
		meta.ValidatedAt = now.UTC()
		if err := c.writeMeta(meta); err != nil {
			return nil, nil, err
		}
		c.stats.record(kind, func(f *Freshness) { f.Revalidated++ })
		return body, &meta, nil
	}
	m := Meta{
		Key: key, Kind: kind, Method: req.Method, URL: req.URL.String(),
		FetchedAt: now.UTC(), ValidatedAt: now.UTC(),
		ETag: resp.Header.Get("ETag"), LastModified: resp.Header.Get("Last-Modified"),
	}
	if err := c.store(m, fresh); err != nil {
		return nil, nil, err
	}
	m.SHA256, m.Bytes = sha256Hex(fresh), len(fresh)
	c.stats.record(kind, func(f *Freshness) {
		if present {
			f.Refreshed++
		} else {
			f.Fetched++
		}
	})
	return fresh, &m, nil
}

// do performs req; a 304 is returned as a response with a nil body.
func do(client *http.Client, req *http.Request) ([]byte, *http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%s %s: %w", req.Method, req.URL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response body for %s %s: %w", req.Method, req.URL, err)
	}
	if resp.StatusCode == http.StatusNotModified {
		return nil, resp, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("%s %s: HTTP %d: %s", req.Method, req.URL, resp.StatusCode, truncate(body, 500))
	}
	return body, resp, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
