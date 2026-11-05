package httpcache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Kind classifies a cached resource by how it may change (CVE:F2).
type Kind string

const (
	// KindOSV: OSV query/querybatch/vuln responses — mutable upstream
	// (new advisories, withdrawals, edits); subject to the TTL. An empty
	// querybatch result is as mutable as a non-empty one.
	KindOSV Kind = "osv"
	// KindPinned: content addressed by a full commit SHA (a GitHub
	// commit diff or raw file at a 40-hex SHA) — immutable, never stale.
	KindPinned Kind = "pinned"
	// KindMutable: any other web resource (e.g. a commit URL with an
	// abbreviated SHA); treated like OSV data.
	KindMutable Kind = "mutable"
	// KindLLM: an LLM answer — a recorded, paid result, never refreshed
	// by the data-freshness policy (retries are the LLM source's own).
	KindLLM Kind = "llm"
)

// expires reports whether entries of k are subject to the TTL.
func (k Kind) expires() bool { return k == KindOSV || k == KindMutable }

// Meta is the metadata stored beside a cached body as <key>.meta.json.
// SHA256 binds it to the body: a sidecar whose hash does not match the
// body (a crash between the two renames, or a concurrent writer of the
// same key) is ignored, and the entry is of unknown age. Never contains
// request headers — authorization stays out of the cache.
type Meta struct {
	Key          string    `json:"key"`
	Kind         Kind      `json:"kind"`
	Method       string    `json:"method,omitempty"`
	URL          string    `json:"url,omitempty"`
	FetchedAt    time.Time `json:"fetched_at"`
	ValidatedAt  time.Time `json:"validated_at"`
	SHA256       string    `json:"sha256"`
	Bytes        int       `json:"bytes"`
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"last_modified,omitempty"`
}

// Policy is the data-freshness policy for expiring kinds.
type Policy struct {
	// TTL is how long a validated entry counts as fresh (default 24h).
	TTL time.Duration
	// MaxAge, when set, is a hard limit: no entry older than it (or of
	// unknown age) is ever served, offline included.
	MaxAge time.Duration
	// Refresh revalidates every expiring entry regardless of age.
	Refresh bool
	// AllowStale serves the cached copy when an online refresh fails
	// (never past MaxAge); the run is then marked as using stale data.
	AllowStale bool
}

// DefaultTTL is the default freshness window for OSV data.
const DefaultTTL = 24 * time.Hour

// ErrTooOld is returned when MaxAge forbids serving the only copy.
var ErrTooOld = errors.New("httpcache: cached data older than -max-data-age")

// Freshness summarizes, per kind, what a run served (CVE:F2).
type Freshness struct {
	Kind Kind `json:"kind"`
	// Fresh: served from cache within the TTL. Fetched: not cached,
	// downloaded. Refreshed: stale or unknown-age entry downloaded again.
	// Revalidated: a conditional request answered 304. Stale: served
	// past the TTL (offline, or -allow-stale after a failed refresh).
	// UnknownAge: served without metadata (cached before CVE:F2).
	Fresh       int `json:"fresh"`
	Fetched     int `json:"fetched"`
	Refreshed   int `json:"refreshed"`
	Revalidated int `json:"revalidated"`
	Stale       int `json:"stale"`
	UnknownAge  int `json:"unknown_age"`
	// Replayed: served from a data snapshot (CVE:F3 -snapshot-in).
	Replayed int `json:"replayed,omitempty"`
	// OldestValidated is the oldest validation time among the entries of
	// known age served from cache (a timestamp, not an age, so reruns on
	// the same cache report the same value); zero when none.
	OldestValidated time.Time `json:"oldest_validated_at,omitempty"`
	// RefreshFailures are the errors of failed refreshes that -allow-stale
	// answered from the cache.
	RefreshFailures []string `json:"refresh_failures,omitempty"`
}

func (f *Freshness) noteValidated(t time.Time) {
	if f.OldestValidated.IsZero() || t.Before(f.OldestValidated) {
		f.OldestValidated = t.UTC()
	}
}

type stats struct {
	mu     sync.Mutex
	byKind map[Kind]*Freshness
}

func (s *stats) record(k Kind, f func(*Freshness)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byKind == nil {
		s.byKind = map[Kind]*Freshness{}
	}
	fr := s.byKind[k]
	if fr == nil {
		fr = &Freshness{Kind: k}
		s.byKind[k] = fr
	}
	f(fr)
}

// Freshness returns the per-kind summary of this run, sorted by kind.
func (c *Cache) Freshness() []Freshness {
	c.stats.mu.Lock()
	defer c.stats.mu.Unlock()
	out := make([]Freshness, 0, len(c.stats.byKind))
	for _, f := range c.stats.byKind {
		cp := *f
		cp.RefreshFailures = append([]string(nil), f.RefreshFailures...)
		sort.Strings(cp.RefreshFailures)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}

func (c *Cache) metaPath(key string) string { return filepath.Join(c.dir, key+".meta.json") }

// readMeta returns the entry's metadata when present and bound to body.
func (c *Cache) readMeta(key string, body []byte) (Meta, bool) {
	b, err := os.ReadFile(c.metaPath(key))
	if err != nil {
		return Meta{}, false
	}
	var m Meta
	if json.Unmarshal(b, &m) != nil || m.SHA256 != sha256Hex(body) {
		return Meta{}, false
	}
	return m, true
}

func (c *Cache) writeMeta(m Meta) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache metadata %s: %w", m.Key, err)
	}
	return c.writeAtomic(c.metaPath(m.Key), m.Key+".meta", b)
}

func sha256Hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// SetPolicy replaces the freshness policy; a zero TTL means DefaultTTL.
func (c *Cache) SetPolicy(p Policy) {
	if p.TTL <= 0 {
		p.TTL = DefaultTTL
	}
	c.policy = p
}

// SetClock replaces the clock used for ages (tests).
func (c *Cache) SetClock(now func() time.Time) { c.now = now }

// effectiveTTL is the TTL, tightened by MaxAge when that is shorter.
func (p Policy) effectiveTTL() time.Duration {
	if p.MaxAge > 0 && p.MaxAge < p.TTL {
		return p.MaxAge
	}
	return p.TTL
}

// allowed reports whether MaxAge permits serving an entry of this age.
func (p Policy) allowed(age time.Duration, known bool) bool {
	return p.MaxAge <= 0 || (known && age <= p.MaxAge)
}
