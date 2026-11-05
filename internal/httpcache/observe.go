package httpcache

import (
	"errors"
	"fmt"
)

// Observation is the outcome of one cache lookup, recorded for a data
// snapshot (CVE:F3): a body served (hit), a plain miss, or an error that
// the caller saw (an HTTP 404 for an alias record, an offline miss) —
// negative observations matter, they shape the result too.
type Observation struct {
	Namespace string `json:"namespace"` // which cache: "main", "llm", "exploit", "framework-check"
	Key       string `json:"key"`
	Kind      Kind   `json:"kind,omitempty"`
	Outcome   string `json:"outcome"` // OutcomeHit, OutcomeMiss, OutcomeError
	SHA256    string `json:"sha256,omitempty"`
	Bytes     int    `json:"bytes,omitempty"`
	Meta      *Meta  `json:"meta,omitempty"`
	ErrClass  string `json:"error_class,omitempty"` // "offline", "too_old", "other"
	Error     string `json:"error,omitempty"`
}

// Observation outcomes.
const (
	OutcomeHit   = "hit"
	OutcomeMiss  = "miss"
	OutcomeError = "error"
)

// Recorder receives every lookup outcome of a cache (and every Set, as a
// hit with the stored body: the final state of a key is what a replay
// must reproduce).
type Recorder interface {
	Observe(o Observation, body []byte)
}

// Replayer answers lookups from a snapshot instead of the cache
// directory and the network.
type Replayer interface {
	// Lookup returns the recorded observation and, for a hit, its body
	// (verified against the recorded hash). ok is false when the
	// snapshot has no observation for the key.
	Lookup(namespace, key string) (o Observation, body []byte, ok bool, err error)
}

// ErrNotInSnapshot is returned in replay for a lookup the snapshot did
// not record: the replayed analysis asked for data the original did not.
var ErrNotInSnapshot = errors.New("httpcache: not in the replayed snapshot")

// SetNamespace names this cache in observations ("main" by default).
func (c *Cache) SetNamespace(ns string) { c.namespace = ns }

// SetRecorder makes the cache report every lookup outcome to r.
func (c *Cache) SetRecorder(r Recorder) { c.recorder = r }

// SetReplay makes the cache answer only from r: no directory read, no
// write, no network (CVE:F3 -snapshot-in).
func (c *Cache) SetReplay(r Replayer) { c.replay = r; c.offline = true }

func (c *Cache) ns() string {
	if c.namespace == "" {
		return "main"
	}
	return c.namespace
}

func errClass(err error) string {
	switch {
	case errors.Is(err, ErrOffline):
		return "offline"
	case errors.Is(err, ErrTooOld):
		return "too_old"
	}
	return "other"
}

func (c *Cache) observeHit(key string, kind Kind, body []byte, meta *Meta) {
	if c.recorder != nil {
		c.recorder.Observe(Observation{Namespace: c.ns(), Key: key, Kind: kind, Outcome: OutcomeHit, SHA256: sha256Hex(body), Bytes: len(body), Meta: meta}, body)
	}
}

func (c *Cache) observeMiss(key string, kind Kind) {
	if c.recorder != nil {
		c.recorder.Observe(Observation{Namespace: c.ns(), Key: key, Kind: kind, Outcome: OutcomeMiss}, nil)
	}
}

func (c *Cache) observeErr(key string, kind Kind, err error) {
	if c.recorder != nil && err != nil {
		c.recorder.Observe(Observation{Namespace: c.ns(), Key: key, Kind: kind, Outcome: OutcomeError, ErrClass: errClass(err), Error: err.Error()}, nil)
	}
}

// replayedError reproduces a recorded error: same text, same class for
// errors.Is.
type replayedError struct {
	msg  string
	base error
}

func (e replayedError) Error() string { return e.msg }
func (e replayedError) Unwrap() error { return e.base }

// fromReplay answers a lookup in replay mode: (body, hit, err).
func (c *Cache) fromReplay(key string, kind Kind) ([]byte, bool, error) {
	o, body, ok, err := c.replay.Lookup(c.ns(), key)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, fmt.Errorf("%w: %s/%s", ErrNotInSnapshot, c.ns(), key)
	}
	switch o.Outcome {
	case OutcomeHit:
		c.stats.record(kind, func(f *Freshness) { f.Replayed++ })
		return body, true, nil
	case OutcomeMiss:
		return nil, false, nil
	}
	var base error
	switch o.ErrClass {
	case "offline":
		base = ErrOffline
	case "too_old":
		base = ErrTooOld
	}
	return nil, false, replayedError{msg: o.Error, base: base}
}
