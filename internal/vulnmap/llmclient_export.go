package vulnmap

import (
	"context"
	"encoding/json"
	"net/http"
)

// ChatMessage is one message in a chat-completions exchange, exported so a
// caller outside this package (internal/exploit) can share this package's
// HTTP client — retries, backoff, and fatal-auth-error detection — instead
// of reimplementing it for its own, differently-shaped prompt and answer.
type ChatMessage struct {
	Role    string
	Content string
}

// LLMClient is a thin exported wrapper around the package-private
// llmClient, for callers that need the same OpenAI-compatible endpoint but
// have their own prompt and answer schema.
type LLMClient struct{ c *llmClient }

// NewLLMClient builds an LLMClient talking to cfg's endpoint. httpClient
// may be nil to use http.DefaultClient.
func NewLLMClient(cfg LLMConfig, httpClient *http.Client) *LLMClient {
	return &LLMClient{c: newLLMClient(cfg, httpClient)}
}

// Chat sends one chat-completions request and returns the first choice's
// content, finish reason, and usage, retrying transient failures the same
// way the CVE-mapping LLM source does.
func (l *LLMClient) Chat(ctx context.Context, msgs []ChatMessage) (content, finishReason string, usage json.RawMessage, err error) {
	internal := make([]chatMessage, len(msgs))
	for i, m := range msgs {
		internal[i] = chatMessage{Role: m.Role, Content: m.Content}
	}
	reply, err := l.c.chat(ctx, internal)
	if err != nil {
		return "", "", nil, err
	}
	return reply.Content, reply.FinishReason, reply.Usage, nil
}

// IsFatalLLMError reports whether err means every further LLM call would
// fail too (e.g. a rejected API key), so a caller should stop instead of
// skipping just the one request.
func IsFatalLLMError(err error) bool { return isFatalLLM(err) }
