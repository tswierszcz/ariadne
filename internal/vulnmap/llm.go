package vulnmap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultLLMBaseURL is the only LLM setting with a built-in default.
	DefaultLLMBaseURL = "https://openrouter.ai/api/v1"

	envLLMBaseURL = "ARIADNE_LLM_BASE_URL"
	envLLMModel   = "ARIADNE_LLM_MODEL"
	envLLMAPIKey  = "ARIADNE_LLM_API_KEY"

	llmMaxAttempts    = 3
	llmMaxRetryAfter  = 30 * time.Second
	llmMaxTokens      = 16384
	llmMaxErrorBodyKB = 2
)

// LLMConfig is everything needed to talk to an OpenAI-compatible
// chat-completions endpoint. APIKey is only ever put in the request's
// Authorization header: never in an error message, a log line, or the cache.
type LLMConfig struct {
	BaseURL string
	Model   string
	APIKey  string
}

// LLMConfigFromEnv reads the three ARIADNE_LLM_* variables through getenv
// (os.Getenv in production; a fake in tests). An unset base URL falls back
// to DefaultLLMBaseURL.
func LLMConfigFromEnv(getenv func(string) string) LLMConfig {
	cfg := LLMConfig{
		BaseURL: getenv(envLLMBaseURL),
		Model:   getenv(envLLMModel),
		APIKey:  getenv(envLLMAPIKey),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultLLMBaseURL
	}
	return cfg
}

// Validate reports a missing model or key, so ariadne can fail at startup
// with a clear message instead of in the middle of a run. A key is not
// needed in offline mode, where every LLM answer must already be cached.
func (c LLMConfig) Validate(offline bool) error {
	if c.Model == "" {
		return fmt.Errorf("-llm needs a model: set %s or pass -llm-model", envLLMModel)
	}
	if c.APIKey == "" && !offline {
		return fmt.Errorf("-llm needs %s in the environment", envLLMAPIKey)
	}
	return nil
}

// fatalLLMError is an LLM failure that makes every further call pointless
// (a rejected key), so the whole run should stop rather than skip one CVE.
type fatalLLMError struct{ msg string }

func (e *fatalLLMError) Error() string { return e.msg }

func isFatalLLM(err error) bool {
	var f *fatalLLMError
	return errors.As(err, &f)
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string            `json:"model"`
	Temperature    int               `json:"temperature"`
	MaxTokens      int               `json:"max_tokens"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
	// Reasoning caps a reasoning model's hidden thinking: GLM 5.3 spent its
	// whole token budget on it once and returned an empty reply.
	Reasoning *chatReasoning `json:"reasoning,omitempty"`
	Messages  []chatMessage  `json:"messages"`
}

type chatReasoning struct {
	Effort string `json:"effort"`
}

type chatResponse struct {
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage json.RawMessage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// chatReply is one successful completion.
type chatReply struct {
	Content      string
	FinishReason string
	Usage        json.RawMessage
}

type llmClient struct {
	cfg  LLMConfig
	http *http.Client
	// sleep is replaced in tests so retry backoff does not slow them down.
	sleep func(context.Context, time.Duration) error
}

func newLLMClient(cfg LLMConfig, httpClient *http.Client) *llmClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &llmClient{cfg: cfg, http: httpClient, sleep: sleepCtx}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// chat sends one chat-completions request, retrying transient failures:
//   - network errors, HTTP 429 and 5xx: up to llmMaxAttempts attempts in
//     total, backing off 2s then 4s (or Retry-After, capped at 30s);
//   - HTTP 400: one extra attempt without response_format, in case the
//     model rejects it;
//   - HTTP 401/403: a *fatalLLMError;
//   - any other failure: an ordinary error, and the caller skips this CVE.
func (c *llmClient) chat(ctx context.Context, msgs []chatMessage) (chatReply, error) {
	jsonMode := true
	backoff := 2 * time.Second
	var lastErr error

	for attempt := 1; attempt <= llmMaxAttempts; attempt++ {
		reply, status, retryAfter, err := c.chatOnce(ctx, msgs, jsonMode)
		if err == nil {
			return reply, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return chatReply{}, ctx.Err()
		}

		switch {
		case status == http.StatusUnauthorized || status == http.StatusForbidden:
			return chatReply{}, &fatalLLMError{msg: fmt.Sprintf("LLM endpoint rejected the credentials (HTTP %d): check %s", status, envLLMAPIKey)}
		case status == http.StatusBadRequest && jsonMode:
			jsonMode = false
			attempt-- // the no-response_format retry does not use up a transient-failure attempt
			continue
		case status == 0 || status == http.StatusTooManyRequests || status >= 500:
			if attempt == llmMaxAttempts {
				break
			}
			wait := backoff
			if retryAfter > 0 {
				wait = min(retryAfter, llmMaxRetryAfter)
			}
			backoff *= 2
			if err := c.sleep(ctx, wait); err != nil {
				return chatReply{}, err
			}
		default:
			return chatReply{}, err
		}
	}
	return chatReply{}, fmt.Errorf("giving up after %d attempts: %w", llmMaxAttempts, lastErr)
}

// chatOnce performs a single HTTP exchange. status is 0 for a transport
// error or an unusable 200 body (treated as transient by the caller).
func (c *llmClient) chatOnce(ctx context.Context, msgs []chatMessage, jsonMode bool) (reply chatReply, status int, retryAfter time.Duration, err error) {
	body := chatRequest{Model: c.cfg.Model, MaxTokens: llmMaxTokens, Messages: msgs, Reasoning: &chatReasoning{Effort: "low"}}
	if jsonMode {
		body.ResponseFormat = map[string]string{"type": "json_object"}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return chatReply{}, 0, 0, fmt.Errorf("encoding LLM request: %w", err)
	}

	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return chatReply{}, 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("X-Title", "ariadne")

	resp, err := c.http.Do(req)
	if err != nil {
		// err is a *url.Error carrying the URL, never the headers.
		return chatReply{}, 0, 0, fmt.Errorf("POST %s: %w", url, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return chatReply{}, 0, 0, fmt.Errorf("reading LLM response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var ra time.Duration
		if secs, convErr := strconv.Atoi(resp.Header.Get("Retry-After")); convErr == nil && secs > 0 {
			ra = time.Duration(secs) * time.Second
		}
		return chatReply{}, resp.StatusCode, ra, fmt.Errorf("POST %s: HTTP %d: %s", url, resp.StatusCode, clip(raw, llmMaxErrorBodyKB*1024))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return chatReply{}, 0, 0, fmt.Errorf("decoding LLM response: %w", err)
	}
	// OpenRouter can answer 200 with an {"error":...} body (an upstream
	// provider failed after the stream started); treat it like a 5xx.
	if parsed.Error != nil || len(parsed.Choices) == 0 {
		msg := "no choices in response"
		if parsed.Error != nil {
			msg = parsed.Error.Message
		}
		return chatReply{}, 0, 0, fmt.Errorf("LLM response unusable: %s", msg)
	}
	ch := parsed.Choices[0]
	return chatReply{Content: ch.Message.Content, FinishReason: ch.FinishReason, Usage: parsed.Usage}, resp.StatusCode, 0, nil
}

func clip(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
