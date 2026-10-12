// Package httpcache is a small disk cache for HTTP responses, keyed by a
// caller-supplied key (typically a hash of the request). It backs
// internal/cves (OSV queries) and internal/vulnmap (fix-commit diffs, raw
// files, LLM answers), and supports an --offline mode that serves only
// from the cache and fails clearly on a miss.
package httpcache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrOffline is returned by Get when the cache is offline and the key is
// not present.
var ErrOffline = errors.New("httpcache: offline and not in cache")

// Cache is a directory of files, one per key.
type Cache struct {
	dir     string
	offline bool
}

// New returns a Cache rooted at dir, creating it if necessary.
func New(dir string, offline bool) (*Cache, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir %s: %w", dir, err)
	}
	return &Cache{dir: dir, offline: offline}, nil
}

// Key hashes an arbitrary set of strings (e.g. method, URL, body, model id)
// into a stable cache key.
func Key(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Cache) path(key string) string {
	return filepath.Join(c.dir, key+".json")
}

// Get returns the cached bytes for key, (nil, false, nil) on a plain miss
// while online, or (nil, false, ErrOffline) on a miss while offline.
func (c *Cache) Get(key string) ([]byte, bool, error) {
	b, err := os.ReadFile(c.path(key))
	if errors.Is(err, os.ErrNotExist) {
		if c.offline {
			return nil, false, fmt.Errorf("%w: %s", ErrOffline, key)
		}
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("reading cache entry %s: %w", key, err)
	}
	return b, true, nil
}

// Set writes b under key. It writes to a uniquely-named temp file first
// and renames it into place, so a crash mid-write never leaves a corrupt
// cache entry. The temp name must be unique per call, not just per key:
// two different CVEs sharing the same fix-commit URL (a single commit
// fixing more than one CVE) legitimately race to cache the same key from
// two goroutines, and a shared "key+.tmp" name would let one call's
// rename steal the other's temp file out from under it.
func (c *Cache) Set(key string, b []byte) error {
	final := c.path(key)
	tmp, err := os.CreateTemp(c.dir, key+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file for cache entry %s: %w", key, err)
	}
	tmpName := tmp.Name()
	_, writeErr := tmp.Write(b)
	closeErr := tmp.Close()
	if writeErr != nil {
		os.Remove(tmpName)
		return fmt.Errorf("writing cache entry %s: %w", key, writeErr)
	}
	if closeErr != nil {
		os.Remove(tmpName)
		return fmt.Errorf("writing cache entry %s: %w", key, closeErr)
	}
	if err := os.Rename(tmpName, final); err != nil {
		return fmt.Errorf("committing cache entry %s: %w", key, err)
	}
	return nil
}

// Dir returns the directory the cache stores its entries in.
func (c *Cache) Dir() string {
	return c.dir
}

// Offline reports whether this cache was constructed in offline mode.
func (c *Cache) Offline() bool {
	return c.offline
}
