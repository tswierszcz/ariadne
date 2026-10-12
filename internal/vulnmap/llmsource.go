package vulnmap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

// llmConcurrency bounds simultaneous LLM calls; ResolveAll's 8 workers
// would otherwise all hit the endpoint at once and trip rate limits.
const llmConcurrency = 4

// LLMStats counts what source 3 did across a run, for the log and report.
type LLMStats struct {
	Mapped  int // CVEs given at least one grounded method
	Empty   int // the model (or grounding) left no method: UNKNOWN
	Skipped int // the request failed; no answer was obtained
	Invalid int // the model's reply failed validation twice
	Dropped int // individual methods discarded as absent from the bytecode
}

// LLMSource is source 3: an LLM asked to name the vulnerable methods from
// the advisory and, when there is one, the fix diff.
type LLMSource struct {
	Config LLMConfig
	// Cache holds one entry per (model, prompt); it is separate from the
	// OSV/GitHub cache so the key and its contents stay easy to inspect.
	Cache *httpcache.Cache

	client *llmClient
	sem    chan struct{}

	mu       sync.Mutex
	stats    LLMStats
	warnings []string
}

// NewLLMSource builds a source that talks to cfg's endpoint and caches in
// cache (typically <cache-dir>/llm).
func NewLLMSource(cfg LLMConfig, cache *httpcache.Cache, httpClient *http.Client) *LLMSource {
	return &LLMSource{
		Config: cfg,
		Cache:  cache,
		client: newLLMClient(cfg, httpClient),
		sem:    make(chan struct{}, llmConcurrency),
	}
}

// Stats returns a snapshot of the counters.
func (s *LLMSource) Stats() LLMStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats
}

// Warnings returns a copy of the per-CVE problems recorded so far.
func (s *LLMSource) Warnings() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.warnings...)
}

func (s *LLMSource) record(update func(*LLMStats), warning string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	update(&s.stats)
	if warning != "" {
		s.warnings = append(s.warnings, warning)
	}
}

// llmRecord is one cache entry: the full exchange, so a run can be audited
// and replayed without the network. It never contains the API key.
type llmRecord struct {
	Model         string          `json:"model"`
	PromptVersion int             `json:"prompt_version"`
	Messages      []chatMessage   `json:"messages"`
	Responses     []string        `json:"responses"`
	Valid         bool            `json:"valid"`
	Error         string          `json:"error,omitempty"`
	Answer        *llmAnswer      `json:"answer,omitempty"`
	Usage         json.RawMessage `json:"usage,omitempty"`
}

// resolveFromLLM is called by Resolver.Resolve once sources 1 and 2 found
// nothing. It returns an error only for fatal conditions (rejected key,
// offline cache miss, cancelled context); every other failure is recorded
// as a warning and yields (Entry{}, false, nil), i.e. UNKNOWN.
func (r *Resolver) resolveFromLLM(ctx context.Context, cve cves.CVE) (Entry, bool, error) {
	s := r.LLM

	commitURLs, commits, err := r.llmCommitDiffs(ctx, cve)
	if err != nil {
		return Entry{}, false, err
	}
	user, evidence, err := buildUserPrompt(cve, commitURLs, commits)
	if err != nil {
		return Entry{}, false, err
	}
	system := llmSystemPrompt()

	rec, err := s.exchange(ctx, cve.ID, system, user)
	if err != nil {
		if isFatalLLM(err) || errors.Is(err, httpcache.ErrOffline) || ctx.Err() != nil {
			return Entry{}, false, err
		}
		s.record(func(st *LLMStats) { st.Skipped++ }, fmt.Sprintf("%s: LLM request failed, left UNKNOWN: %v", cve.ID, err))
		return Entry{}, false, nil
	}
	if !rec.Valid {
		s.record(func(st *LLMStats) { st.Invalid++ }, fmt.Sprintf("%s: LLM reply invalid after one corrective retry, left UNKNOWN: %s", cve.ID, rec.Error))
		return Entry{}, false, nil
	}

	methods, dropped := groundMethods(*rec.Answer, r.MethodExists)
	if len(methods) == 0 {
		s.record(func(st *LLMStats) { st.Empty++; st.Dropped += dropped }, "")
		return Entry{}, false, nil
	}
	s.record(func(st *LLMStats) { st.Mapped++; st.Dropped += dropped }, "")
	return Entry{Methods: methods, Source: SourceLLM, Model: s.Config.Model, Evidence: evidence}, true, nil
}

// llmCommitDiffs fetches (through the shared HTTP cache, so normally a hit
// after source 2) every fix commit's diff and keeps its non-test .java files.
func (r *Resolver) llmCommitDiffs(ctx context.Context, cve cves.CVE) (urls []string, commits []commitDiff, err error) {
	for _, c := range fixCommits(cve) {
		url := fmt.Sprintf("%s/%s/%s/commit/%s", r.GithubBaseURL, c.owner, c.repo, c.sha)
		// The prompt names the commit by its public GitHub URL, not by
		// whatever GithubBaseURL a test points at, so the golden prompt
		// text stays stable.
		public := fmt.Sprintf("https://github.com/%s/%s/commit/%s", c.owner, c.repo, c.sha)
		urls = append(urls, public)

		body, err := r.fetch(ctx, url+".diff")
		if err != nil {
			return nil, nil, fmt.Errorf("fetching fix commit diff %s.diff: %w", public, err)
		}
		cd := commitDiff{url: public}
		for _, f := range splitDiffFiles(string(body)) {
			if keepForLLM(f.path) {
				cd.files = append(cd.files, f)
			}
		}
		commits = append(commits, cd)
	}
	return urls, commits, nil
}

// reasoningTokens extracts completion_tokens_details.reasoning_tokens from
// a usage object for the truncation warning, or "unknown".
func reasoningTokens(usage json.RawMessage) string {
	var u struct {
		Details struct {
			Reasoning *int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	}
	if json.Unmarshal(usage, &u) != nil || u.Details.Reasoning == nil {
		return "unknown"
	}
	return strconv.Itoa(*u.Details.Reasoning)
}

// exchange returns the cached record for (model, prompt) or performs the
// call, validates the reply, and on a validation failure makes the single
// corrective retry. Transport failures are returned as errors and never
// cached; a reply that stays invalid is cached with Valid=false so the
// next run reproduces the same outcome without another request.
func (s *LLMSource) exchange(ctx context.Context, cveID, system, user string) (llmRecord, error) {
	key := httpcache.Key("llm", s.Config.Model, strconv.Itoa(promptVersion), system, user)
	if b, hit, err := s.Cache.Get(key); err != nil {
		return llmRecord{}, fmt.Errorf("%s: %w", cveID, err)
	} else if hit {
		var rec llmRecord
		if err := json.Unmarshal(b, &rec); err != nil {
			return llmRecord{}, fmt.Errorf("decoding cached LLM answer %s: %w", key, err)
		}
		return rec, nil
	}

	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return llmRecord{}, ctx.Err()
	}

	rec := llmRecord{
		Model:         s.Config.Model,
		PromptVersion: promptVersion,
		Messages:      []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
	}

	for attempt := 0; attempt < 2; attempt++ {
		reply, err := s.client.chat(ctx, rec.Messages)
		if err != nil {
			return llmRecord{}, err
		}
		rec.Responses = append(rec.Responses, reply.Content)
		rec.Usage = reply.Usage

		ans, perr := parseLLMAnswer(reply.Content, reply.FinishReason)
		if errors.Is(perr, errTruncated) {
			// Not cached and not retried: report it as a failed request
			// so the CVE is skipped with a clear warning.
			return llmRecord{}, fmt.Errorf("%w (reasoning tokens: %s)", perr, reasoningTokens(reply.Usage))
		}
		if perr == nil {
			rec.Valid, rec.Answer, rec.Error = true, &ans, ""
			break
		}
		rec.Error = perr.Error()
		if attempt == 1 {
			break // still invalid after the corrective retry
		}
		rec.Messages = append(rec.Messages,
			chatMessage{Role: "assistant", Content: reply.Content},
			chatMessage{Role: "user", Content: correctionMessage(perr.Error())},
		)
	}

	if err := s.Cache.Set(key, mustJSON(rec)); err != nil {
		return llmRecord{}, err
	}
	return rec, nil
}
