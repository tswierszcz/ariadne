package vulnmap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestLLM_realEndpoint makes one real call with the real prompts, using the
// ARIADNE_LLM_* environment. It is skipped unless ARIADNE_LLM_E2E=1, so
// `go test ./...` never needs the network or a key.
func TestLLM_realEndpoint(t *testing.T) {
	if os.Getenv("ARIADNE_LLM_E2E") != "1" {
		t.Skip("set ARIADNE_LLM_E2E=1 (and the ARIADNE_LLM_* variables) to run")
	}
	cfg := LLMConfigFromEnv(os.Getenv)
	if err := cfg.Validate(false); err != nil {
		t.Fatalf("config: %v", err)
	}

	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/example/lib/commit/abc1234.diff" {
			w.Write([]byte(fakeDiff))
			return
		}
		http.NotFound(w, r)
	}))
	defer gh.Close()

	r := NewResolver(nil, mustCache(t))
	r.GithubBaseURL, r.RawBaseURL = gh.URL, gh.URL
	r.LLM = NewLLMSource(cfg, mustCache(t), nil)

	entry, ok, err := r.Resolve(context.Background(), testCVE())
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	t.Logf("ok=%v entry=%+v stats=%+v warnings=%v", ok, entry, r.LLM.Stats(), r.LLM.Warnings())
	if !ok || entry.Source != SourceLLM || entry.Evidence != EvidenceDiff {
		t.Fatalf("expected an llm/diff mapping, got ok=%v entry=%+v", ok, entry)
	}
}
