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

// Set writes b under key. It writes to a temp file first and renames, so a
// crash mid-write never leaves a corrupt cache entry.
func (c *Cache) Set(key string, b []byte) error {
	final := c.path(key)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("writing cache entry %s: %w", key, err)
	}
	if err := os.Rename(tmp, final); err != nil {
		return fmt.Errorf("committing cache entry %s: %w", key, err)
	}
	return nil
}

// Offline reports whether this cache was constructed in offline mode.
func (c *Cache) Offline() bool {
	return c.offline
}
