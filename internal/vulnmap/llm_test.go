package vulnmap

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
)

var update = flag.Bool("update", false, "rewrite testdata/llm/*.golden")

const fakeDiff = `diff --git a/java/org/example/Parser.java b/java/org/example/Parser.java
index 1111111..2222222 100644
--- a/java/org/example/Parser.java
+++ b/java/org/example/Parser.java
@@ -10,3 +10,4 @@ public class Parser {
     void parse(String s) {
+        check(s);
     }
diff --git a/test/org/example/TestParser.java b/test/org/example/TestParser.java
index 3333333..4444444 100644
--- a/test/org/example/TestParser.java
+++ b/test/org/example/TestParser.java
@@ -1,1 +1,2 @@
+// a new regression test
diff --git a/java/org/example/LocalStrings.properties b/java/org/example/LocalStrings.properties
index 5555555..6666666 100644
--- a/java/org/example/LocalStrings.properties
+++ b/java/org/example/LocalStrings.properties
@@ -1,1 +1,2 @@
+parser.bad=bad
`

const jsDiff = `diff --git a/js/bootstrap.js b/js/bootstrap.js
index 1111111..2222222 100644
--- a/js/bootstrap.js
+++ b/js/bootstrap.js
@@ -1,1 +1,2 @@
+// fix
`

func testCVE() cves.CVE {
	return cves.CVE{
		ID:      "CVE-2099-0001",
		Aliases: []string{"GHSA-aaaa-bbbb-cccc"},
		Summary: "Parser bypass in example-lib",
		Details: "The Parser.parse method accepts crafted input.",
		Affected: []cves.Affected{
			{Package: "org.example:example-lib", Version: "1.2.3"},
		},
		References: []cves.Reference{{Type: "FIX", URL: "https://github.com/example/lib/commit/abc1234"}},
	}
}

func readGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "llm", name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test ./internal/vulnmap -update to create it)", err)
	}
	if string(want) != got {
		t.Errorf("%s differs from the rendered prompt; rerun with -update and review the diff.\n--- want\n%s\n--- got\n%s", name, want, got)
	}
}

func TestPromptGolden(t *testing.T) {
	cve := testCVE()
	commitURL := "https://github.com/example/lib/commit/abc1234"

	var keptFiles []fileDiff
	for _, f := range splitDiffFiles(fakeDiff) {
		if keepForLLM(f.path) {
			keptFiles = append(keptFiles, f)
		}
	}

	a, ev, err := buildUserPrompt(cve, []string{commitURL}, []commitDiff{{url: commitURL, files: keptFiles}})
	if err != nil || ev != EvidenceDiff {
		t.Fatalf("variant A: ev=%q err=%v", ev, err)
	}
	b, ev, err := buildUserPrompt(cve, []string{commitURL}, []commitDiff{{url: commitURL}})
	if err != nil || ev != EvidenceAdvisory {
		t.Fatalf("variant B: ev=%q err=%v", ev, err)
	}
	noRef := cve
	noRef.References = nil
	noRef.Details = ""
	c, ev, err := buildUserPrompt(noRef, nil, nil)
	if err != nil || ev != EvidenceAdvisory {
		t.Fatalf("variant C: ev=%q err=%v", ev, err)
	}

	readGolden(t, "system.golden", llmSystemPrompt())
	readGolden(t, "user_a.golden", a)
	readGolden(t, "user_b.golden", b)
	readGolden(t, "user_c.golden", c)
	readGolden(t, "correction.golden", correctionMessage(`methods[2].class "Request" is not a fully qualified class name`))
}

func TestSystemPromptMatchesApprovedDoc(t *testing.T) {
	doc, err := os.ReadFile(filepath.Join("..", "..", "docs", "phase4-llm.md"))
	if err != nil {
		t.Skipf("design doc not available: %v", err)
	}
	if !strings.Contains(string(doc), llmSystemPrompt()) {
		t.Errorf("the embedded system prompt no longer matches docs/phase4-llm.md; update both together")
	}
}

func TestIsTestPathAndKeep(t *testing.T) {
	cases := map[string]bool{
		"java/org/apache/catalina/Request.java":   true,
		"src/main/java/a/B.java":                  true,
		"src/test/java/a/BTest.java":              false,
		"test/org/apache/TestRequest.java":        false,
		"modules/tests/java/a/B.java":             false,
		"src/main/java/a/FooTests.java":           false,
		"java/org/apache/LocalStrings.properties": false,
		"src/main/resources/application.yml":      false,
		"src/main/java/a/Contest.java":            true, // "Test" must be a suffix of the file name, "Contest" is not a test
	}
	for p, want := range cases {
		if got := keepForLLM(p); got != want {
			t.Errorf("keepForLLM(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestAssembleDiffTruncatesAtFileBoundary(t *testing.T) {
	f := func(name string, n int) fileDiff {
		return fileDiff{path: name, text: "diff --git a/" + name + " b/" + name + "\n" + strings.Repeat("x", n) + "\n"}
	}
	commits := []commitDiff{{url: "https://github.com/o/r/commit/abc", files: []fileDiff{f("A.java", 100), f("B.java", 100), f("C.java", 100)}}}

	text, truncated := assembleDiff(commits, 10000)
	if truncated || !strings.Contains(text, "B.java") || strings.Contains(text, "[truncated") {
		t.Fatalf("expected everything kept, got truncated=%v:\n%s", truncated, text)
	}

	text, truncated = assembleDiff(commits, 220)
	if !truncated {
		t.Fatalf("expected truncation")
	}
	if !strings.Contains(text, "A.java") || strings.Contains(text, "diff --git a/C.java") {
		t.Errorf("wrong files kept:\n%s", text)
	}
	if !strings.Contains(text, "# [truncated: ") || !strings.Contains(text, "C.java") {
		t.Errorf("omitted files not listed:\n%s", text)
	}

	// A single file bigger than the limit is cut, not dropped.
	text, truncated = assembleDiff([]commitDiff{{url: "u", files: []fileDiff{f("Big.java", 5000)}}}, 300)
	if !truncated || !strings.Contains(text, "diff --git a/Big.java") {
		t.Errorf("oversized first file should be cut, not dropped:\n%s", text)
	}
}

func TestParseLLMAnswer(t *testing.T) {
	valid := `{"methods":[{"class":"org.example.Parser","method":"parse","role":"vulnerable","reason":"r"}],"note":""}`
	tests := []struct {
		name, content, finish string
		wantErr               string // substring; "" means valid
	}{
		{"plain", valid, "stop", ""},
		{"fenced", "```json\n" + valid + "\n```", "stop", ""},
		{"prose around", "Here you go: " + valid + "\nHope that helps.", "stop", ""},
		{"empty methods", `{"methods":[],"note":"none"}`, "stop", ""},
		{"constructor", `{"methods":[{"class":"a.B$C","method":"<init>","role":"entry_point"}]}`, "stop", ""},
		{"no object", "I cannot help", "stop", "no JSON object"},
		{"bad json", `{"methods": [`, "stop", "not valid JSON"},
		{"no methods key", `{"note":"x"}`, "stop", `no "methods" array`},
		{"length", valid, "length", "token limit"},
		{"class without package", `{"methods":[{"class":"Request","method":"m","role":"vulnerable"}]}`, "stop", "not a fully qualified class name"},
		{"method with parens", `{"methods":[{"class":"a.B","method":"m()","role":"vulnerable"}]}`, "stop", "not a bare method name"},
		{"bad role", `{"methods":[{"class":"a.B","method":"m","role":"maybe"}]}`, "stop", "neither"},
		{"too many", `{"methods":[` + strings.TrimSuffix(strings.Repeat(`{"class":"a.B","method":"m","role":"vulnerable"},`, 11), ",") + `]}`, "stop", "maximum is 10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseLLMAnswer(tt.content, tt.finish)
			switch {
			case tt.wantErr == "" && err != nil:
				t.Errorf("unexpected error: %v", err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Errorf("error = %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestGroundMethods(t *testing.T) {
	present := map[string]bool{
		"org.example.Parser#parse":    true,
		"org.example.Outer$Inner#run": true,
		"org.example.Parser#<init>":   true,
	}
	exists := func(c, m string) bool { return present[c+"#"+m] }
	ans := llmAnswer{Methods: []llmMethod{
		{Class: "org.example.Parser", Method: "parse"},
		{Class: "org.example.Parser", Method: "parse"}, // duplicate
		{Class: "org.example.Outer.Inner", Method: "run"},
		{Class: "org.example.Parser", Method: "<init>"},
		{Class: "org.example.Parser", Method: "invented"},
		{Class: "org.example.Nope", Method: "parse"},
	}}
	got, dropped := groundMethods(ans, exists)
	want := []string{"org.example.Outer$Inner#run", "org.example.Parser#<init>", "org.example.Parser#parse"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("methods = %v, want %v", got, want)
	}
	if dropped != 2 {
		t.Errorf("dropped = %d, want 2", dropped)
	}
}

// fakeLLM is an OpenAI-compatible endpoint that answers from a script and
// records what it was sent.
type fakeLLM struct {
	t     *testing.T
	srv   *httptest.Server
	mu    sync.Mutex
	calls int
	reqs  []chatRequest
	// reply decides call n's (1-based) response.
	reply func(n int, req chatRequest) (status int, body string, header http.Header)
}

func newFakeLLM(t *testing.T, reply func(n int, req chatRequest) (int, string, http.Header)) *fakeLLM {
	f := &fakeLLM{t: t, reply: reply}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("X-Title"); got != "ariadne" {
			t.Errorf("X-Title = %q", got)
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		f.mu.Lock()
		f.calls++
		n := f.calls
		f.reqs = append(f.reqs, req)
		f.mu.Unlock()

		status, body, hdr := f.reply(n, req)
		for k, v := range hdr {
			w.Header()[k] = v
		}
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeLLM) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func completion(content, finish string) string {
	return string(mustJSON(map[string]any{
		"choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": content}, "finish_reason": finish}},
		"usage":   map[string]int{"prompt_tokens": 10, "completion_tokens": 5},
	}))
}

const goodAnswer = `{"methods":[{"class":"org.example.Parser","method":"parse","role":"vulnerable","reason":"fix adds check"}],"note":""}`

// llmFixture wires a Resolver to a fake GitHub (serving diff) and a fake LLM.
func llmFixture(t *testing.T, diff string, llm *fakeLLM) (*Resolver, *LLMSource) {
	t.Helper()
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/example/lib/commit/abc1234.diff" {
			w.Write([]byte(diff))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(gh.Close)

	r := NewResolver(nil, mustCache(t))
	r.GithubBaseURL, r.RawBaseURL = gh.URL, gh.URL
	r.MethodExists = func(c, m string) bool { return c == "org.example.Parser" && (m == "parse") }

	src := NewLLMSource(LLMConfig{BaseURL: llm.srv.URL, Model: "test/model", APIKey: "test-key"}, mustCache(t), nil)
	src.client.sleep = func(context.Context, time.Duration) error { return nil }
	r.LLM = src
	return r, src
}

func TestLLM_mapsFromDiffWithGrounding(t *testing.T) {
	llm := newFakeLLM(t, func(n int, req chatRequest) (int, string, http.Header) {
		if req.Model != "test/model" || req.Temperature != 0 || req.ResponseFormat["type"] != "json_object" {
			t.Errorf("unexpected request params: %+v", req)
		}
		if len(req.Messages) != 2 || req.Messages[0].Role != "system" || !strings.Contains(req.Messages[1].Content, "<fix_diff>") {
			t.Errorf("unexpected messages: %+v", req.Messages)
		}
		if strings.Contains(req.Messages[1].Content, "TestParser") || strings.Contains(req.Messages[1].Content, "LocalStrings") {
			t.Errorf("test/properties files leaked into the diff:\n%s", req.Messages[1].Content)
		}
		// One real method, one invented one: grounding must drop the latter.
		return 200, completion(`{"methods":[{"class":"org.example.Parser","method":"parse","role":"vulnerable"},{"class":"org.example.Parser","method":"invented","role":"entry_point"}]}`, "stop"), nil
	})
	r, src := llmFixture(t, fakeDiff, llm)

	entry, ok, err := r.Resolve(context.Background(), testCVE())
	if err != nil || !ok {
		t.Fatalf("Resolve = (%+v, %v, %v)", entry, ok, err)
	}
	if entry.Source != SourceLLM || entry.Model != "test/model" || entry.Evidence != EvidenceDiff {
		t.Errorf("entry = %+v", entry)
	}
	if len(entry.Methods) != 1 || entry.Methods[0] != "org.example.Parser#parse" {
		t.Errorf("methods = %v", entry.Methods)
	}
	if st := src.Stats(); st.Mapped != 1 || st.Dropped != 1 {
		t.Errorf("stats = %+v", st)
	}
}

func TestLLM_advisoryOnlyEvidence(t *testing.T) {
	llm := newFakeLLM(t, func(n int, req chatRequest) (int, string, http.Header) {
		if strings.Contains(req.Messages[1].Content, "<fix_diff>") {
			t.Errorf("no diff expected for a CVE with only a Javascript fix")
		}
		return 200, completion(goodAnswer, "stop"), nil
	})
	r, _ := llmFixture(t, jsDiff, llm)
	entry, ok, err := r.Resolve(context.Background(), testCVE())
	if err != nil || !ok || entry.Evidence != EvidenceAdvisory {
		t.Fatalf("Resolve = (%+v, %v, %v), want advisory evidence", entry, ok, err)
	}
}

func TestLLM_retriesTransientErrors(t *testing.T) {
	llm := newFakeLLM(t, func(n int, _ chatRequest) (int, string, http.Header) {
		if n == 1 {
			return 429, `{"error":"slow down"}`, http.Header{"Retry-After": {"1"}}
		}
		if n == 2 {
			return 503, "bad gateway", nil
		}
		return 200, completion(goodAnswer, "stop"), nil
	})
	r, _ := llmFixture(t, fakeDiff, llm)
	if _, ok, err := r.Resolve(context.Background(), testCVE()); err != nil || !ok {
		t.Fatalf("Resolve ok=%v err=%v", ok, err)
	}
	if llm.callCount() != 3 {
		t.Errorf("calls = %d, want 3", llm.callCount())
	}
}

func TestLLM_persistentServerErrorSkipsCVEAndIsNotCached(t *testing.T) {
	llm := newFakeLLM(t, func(int, chatRequest) (int, string, http.Header) { return 500, "boom", nil })
	r, src := llmFixture(t, fakeDiff, llm)

	_, ok, err := r.Resolve(context.Background(), testCVE())
	if err != nil || ok {
		t.Fatalf("Resolve ok=%v err=%v, want no entry and no error", ok, err)
	}
	if llm.callCount() != 3 {
		t.Errorf("calls = %d, want 3 attempts", llm.callCount())
	}
	if st := src.Stats(); st.Skipped != 1 {
		t.Errorf("stats = %+v", st)
	}
	if len(src.Warnings()) != 1 {
		t.Errorf("warnings = %v", src.Warnings())
	}
	// Not cached: a second run asks again.
	r.Resolve(context.Background(), testCVE())
	if llm.callCount() != 6 {
		t.Errorf("calls after second run = %d, want 6 (failure must not be cached)", llm.callCount())
	}
}

func TestLLM_unauthorizedIsFatal(t *testing.T) {
	llm := newFakeLLM(t, func(int, chatRequest) (int, string, http.Header) { return 401, `{"error":"no"}`, nil })
	r, _ := llmFixture(t, fakeDiff, llm)
	_, _, err := r.Resolve(context.Background(), testCVE())
	if err == nil || !isFatalLLM(err) {
		t.Fatalf("err = %v, want a fatal LLM error", err)
	}
	if strings.Contains(err.Error(), "test-key") {
		t.Errorf("error leaks the API key: %v", err)
	}
	if llm.callCount() != 1 {
		t.Errorf("calls = %d, want 1 (no retry on 401)", llm.callCount())
	}
}

func TestLLM_badRequestRetriesWithoutResponseFormat(t *testing.T) {
	llm := newFakeLLM(t, func(n int, req chatRequest) (int, string, http.Header) {
		if req.ResponseFormat != nil {
			return 400, `{"error":"response_format unsupported"}`, nil
		}
		return 200, completion(goodAnswer, "stop"), nil
	})
	r, _ := llmFixture(t, fakeDiff, llm)
	if _, ok, err := r.Resolve(context.Background(), testCVE()); err != nil || !ok {
		t.Fatalf("Resolve ok=%v err=%v", ok, err)
	}
	if llm.callCount() != 2 {
		t.Errorf("calls = %d, want 2", llm.callCount())
	}
}

func TestLLM_invalidReplyGetsOneCorrectiveRetry(t *testing.T) {
	llm := newFakeLLM(t, func(n int, req chatRequest) (int, string, http.Header) {
		if n == 1 {
			return 200, completion(`{"methods":[{"class":"Parser","method":"parse","role":"vulnerable"}]}`, "stop"), nil
		}
		last := req.Messages[len(req.Messages)-1]
		if len(req.Messages) != 4 || req.Messages[2].Role != "assistant" || last.Role != "user" || !strings.Contains(last.Content, "not a fully qualified class name") {
			t.Errorf("corrective turn missing or wrong: %+v", req.Messages)
		}
		return 200, completion(goodAnswer, "stop"), nil
	})
	r, _ := llmFixture(t, fakeDiff, llm)
	entry, ok, err := r.Resolve(context.Background(), testCVE())
	if err != nil || !ok || entry.Methods[0] != "org.example.Parser#parse" {
		t.Fatalf("Resolve = (%+v, %v, %v)", entry, ok, err)
	}
	if llm.callCount() != 2 {
		t.Errorf("calls = %d, want 2", llm.callCount())
	}
}

func TestLLM_truncatedReplyIsNotRetriedOrCached(t *testing.T) {
	llm := newFakeLLM(t, func(_ int, req chatRequest) (int, string, http.Header) {
		if req.MaxTokens != 16384 || req.Reasoning == nil || req.Reasoning.Effort != "low" {
			t.Errorf("request must cap reasoning and allow 16384 tokens: %+v", req)
		}
		return 200, completion("", "length"), nil
	})
	r, src := llmFixture(t, fakeDiff, llm)

	for i := 0; i < 2; i++ {
		if _, ok, err := r.Resolve(context.Background(), testCVE()); err != nil || ok {
			t.Fatalf("run %d: ok=%v err=%v, want no entry and no error", i, ok, err)
		}
	}
	if llm.callCount() != 2 {
		t.Errorf("calls = %d, want 2 (one per run: no corrective retry, no caching)", llm.callCount())
	}
	if st := src.Stats(); st.Skipped != 2 || st.Invalid != 0 {
		t.Errorf("stats = %+v, want 2 skipped", st)
	}
	if w := src.Warnings(); len(w) != 2 || !strings.Contains(w[0], "token limit") {
		t.Errorf("warnings = %v", w)
	}
}

func TestLLM_twiceInvalidIsUnknownAndCachedAsInvalid(t *testing.T) {
	llm := newFakeLLM(t, func(int, chatRequest) (int, string, http.Header) {
		return 200, completion("sorry, I cannot do that", "stop"), nil
	})
	r, src := llmFixture(t, fakeDiff, llm)

	for i := 0; i < 2; i++ {
		_, ok, err := r.Resolve(context.Background(), testCVE())
		if err != nil || ok {
			t.Fatalf("run %d: ok=%v err=%v", i, ok, err)
		}
	}
	if llm.callCount() != 2 {
		t.Errorf("calls = %d, want 2 (one run's initial try + retry; the second run must read the cache)", llm.callCount())
	}
	if st := src.Stats(); st.Invalid != 2 {
		t.Errorf("stats = %+v", st)
	}

	entries, _ := os.ReadDir(src.Cache.Dir())
	var found bool
	for _, e := range entries {
		b, _ := os.ReadFile(filepath.Join(src.Cache.Dir(), e.Name()))
		var rec llmRecord
		if json.Unmarshal(b, &rec) == nil && !rec.Valid && len(rec.Responses) == 2 {
			found = true
		}
	}
	if !found {
		t.Errorf("no valid:false record with both replies in the cache")
	}
}

func TestLLM_secondRunReadsCache(t *testing.T) {
	llm := newFakeLLM(t, func(int, chatRequest) (int, string, http.Header) {
		return 200, completion(goodAnswer, "stop"), nil
	})
	r, src := llmFixture(t, fakeDiff, llm)
	for i := 0; i < 2; i++ {
		if _, ok, err := r.Resolve(context.Background(), testCVE()); err != nil || !ok {
			t.Fatalf("run %d: ok=%v err=%v", i, ok, err)
		}
	}
	if llm.callCount() != 1 {
		t.Errorf("calls = %d, want 1", llm.callCount())
	}

	// The cache must never contain the API key.
	entries, _ := os.ReadDir(src.Cache.Dir())
	for _, e := range entries {
		b, _ := os.ReadFile(filepath.Join(src.Cache.Dir(), e.Name()))
		if strings.Contains(string(b), "test-key") {
			t.Errorf("cache entry %s contains the API key", e.Name())
		}
	}
}

func TestLLM_offlineMissIsErrOffline(t *testing.T) {
	llm := newFakeLLM(t, func(int, chatRequest) (int, string, http.Header) { return 200, completion(goodAnswer, "stop"), nil })
	r, src := llmFixture(t, fakeDiff, llm)
	offline, err := httpcache.New(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	src.Cache = offline

	_, _, err = r.Resolve(context.Background(), testCVE())
	if !errors.Is(err, httpcache.ErrOffline) {
		t.Fatalf("err = %v, want ErrOffline", err)
	}
	if llm.callCount() != 0 {
		t.Errorf("calls = %d, want 0 offline", llm.callCount())
	}
}

func TestLLM_notCalledWhenEarlierSourceHits(t *testing.T) {
	llm := newFakeLLM(t, func(int, chatRequest) (int, string, http.Header) { return 200, completion(goodAnswer, "stop"), nil })
	r, _ := llmFixture(t, fakeDiff, llm)
	r.Manual = map[string][]string{"CVE-2099-0001": {"a.B#m"}}

	entry, ok, err := r.Resolve(context.Background(), testCVE())
	if err != nil || !ok || entry.Source != SourceManual {
		t.Fatalf("Resolve = (%+v, %v, %v)", entry, ok, err)
	}
	if llm.callCount() != 0 {
		t.Errorf("LLM called %d times although the manual table had an entry", llm.callCount())
	}
}

func TestLLM_notCalledWhenDisabled(t *testing.T) {
	r := NewResolver(nil, mustCache(t))
	_, ok, err := r.Resolve(context.Background(), cves.CVE{ID: "CVE-2099-0002"})
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestLLMConfig(t *testing.T) {
	env := map[string]string{"ARIADNE_LLM_MODEL": "m", "ARIADNE_LLM_API_KEY": "k"}
	cfg := LLMConfigFromEnv(func(k string) string { return env[k] })
	if cfg.BaseURL != DefaultLLMBaseURL || cfg.Model != "m" || cfg.APIKey != "k" {
		t.Errorf("cfg = %+v", cfg)
	}
	if err := cfg.Validate(false); err != nil {
		t.Errorf("Validate: %v", err)
	}
	if err := (LLMConfig{APIKey: "k"}).Validate(false); err == nil || !strings.Contains(err.Error(), "model") {
		t.Errorf("missing model: %v", err)
	}
	if err := (LLMConfig{Model: "m"}).Validate(false); err == nil || !strings.Contains(err.Error(), "ARIADNE_LLM_API_KEY") {
		t.Errorf("missing key: %v", err)
	}
	if err := (LLMConfig{Model: "m"}).Validate(true); err != nil {
		t.Errorf("offline should not need a key: %v", err)
	}
}
