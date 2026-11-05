package httpcache

import (
	"errors"
	"net/http"
	"testing"
)

// memRecorder is a minimal Recorder/Replayer pair for testing the hooks.
type memRecorder struct {
	obs    map[string]Observation
	bodies map[string][]byte
}

func (m *memRecorder) Observe(o Observation, body []byte) {
	id := o.Namespace + "/" + o.Key
	if prev, ok := m.obs[id]; ok && prev.Outcome == OutcomeHit && o.Outcome != OutcomeHit {
		return
	}
	m.obs[id] = o
	if body != nil {
		m.bodies[o.SHA256] = body
	}
}

func (m *memRecorder) Lookup(ns, key string) (Observation, []byte, bool, error) {
	o, ok := m.obs[ns+"/"+key]
	return o, m.bodies[o.SHA256], ok, nil
}

func TestRecordThenReplayWithoutNetwork(t *testing.T) {
	s := newServer(t, "osv-body", "")
	rec := &memRecorder{obs: map[string]Observation{}, bodies: map[string][]byte{}}
	c, _ := New(t.TempDir(), false)
	c.SetRecorder(rec)
	get(t, c, s, KindOSV)
	c.Get("llm-miss")
	c.Set("llm-answer", []byte("answer"))
	s.fail.Store(true)
	req, _ := http.NewRequest("GET", s.URL+"/broken", nil)
	if _, err := Fetch(s.Client(), c, Key("GET", s.URL+"/broken"), KindOSV, req); err == nil {
		t.Fatal("want an error")
	}
	off, _ := New(t.TempDir(), true)
	off.SetRecorder(rec)
	off.Get("offline-miss")
	calls := s.calls.Load()

	r, _ := New(t.TempDir(), false)
	r.SetReplay(rec)
	if b, err := get(t, r, s, KindOSV); b != "osv-body" || err != nil {
		t.Errorf("replayed hit: %q %v", b, err)
	}
	if _, hit, err := r.Get("llm-miss"); hit || err != nil {
		t.Errorf("replayed miss: %v %v", hit, err)
	}
	if b, hit, _ := r.Get("llm-answer"); !hit || string(b) != "answer" {
		t.Errorf("replayed Set: %q %v", b, hit)
	}
	if _, err := Fetch(s.Client(), r, Key("GET", s.URL+"/broken"), KindOSV, req); err == nil || err.Error() == "" {
		t.Errorf("replayed error: %v", err)
	}
	if _, _, err := r.Get("offline-miss"); !errors.Is(err, ErrOffline) {
		t.Errorf("replayed offline miss keeps its class: %v", err)
	}
	if _, _, err := r.Get("never-seen"); !errors.Is(err, ErrNotInSnapshot) {
		t.Errorf("unrecorded key: %v", err)
	}
	if err := r.Set("new", []byte("x")); err != nil {
		t.Error(err)
	}
	if s.calls.Load() != calls {
		t.Errorf("replay touched the network")
	}
	if f := r.Freshness(); len(f) == 0 {
		t.Error("no replay stats")
	}
}
