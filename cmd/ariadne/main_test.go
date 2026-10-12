package main

import (
	"bytes"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

func runWithEnv(args []string, env map[string]string) (code int, stderr string) {
	var out, errBuf bytes.Buffer
	code = run(args, &out, &errBuf, func(k string) string { return env[k] })
	return code, errBuf.String()
}

var baseArgs = []string{"-app", "/nonexistent/app", "-out", "r.md", "-json", "r.json"}

func TestLLMFlagNeedsModelAndKeyAtStartup(t *testing.T) {
	args := append([]string{"-llm"}, baseArgs...)

	code, stderr := runWithEnv(args, map[string]string{"ARIADNE_LLM_API_KEY": "k"})
	if code != 2 || !strings.Contains(stderr, "-llm needs a model") {
		t.Errorf("no model: code=%d stderr=%q, want 2 and a model message", code, stderr)
	}

	code, stderr = runWithEnv(args, map[string]string{"ARIADNE_LLM_MODEL": "m"})
	if code != 2 || !strings.Contains(stderr, "ARIADNE_LLM_API_KEY") {
		t.Errorf("no key: code=%d stderr=%q, want 2 and a key message", code, stderr)
	}
}

func TestLLMFlagsOverrideEnvironment(t *testing.T) {
	// -llm-model supplies the model, so only the key is missing from env.
	code, stderr := runWithEnv(append([]string{"-llm", "-llm-model", "m"}, baseArgs...), map[string]string{"ARIADNE_LLM_API_KEY": "k"})
	if code == 2 {
		t.Errorf("config should validate with -llm-model and a key in env, got code 2: %s", stderr)
	}
	// Past validation the run fails on the nonexistent -app (exit 1), not on usage.
	if code != 1 {
		t.Errorf("code = %d, want 1 (the nonexistent -app)", code)
	}
}

func TestLLMOfflineDoesNotNeedKey(t *testing.T) {
	code, stderr := runWithEnv(append([]string{"-llm", "-offline"}, baseArgs...), map[string]string{"ARIADNE_LLM_MODEL": "m"})
	if code == 2 {
		t.Errorf("-llm -offline should not require a key (answers come from the cache): %s", stderr)
	}
}

func TestWithoutLLMFlagNothingAboutLLMIsRequired(t *testing.T) {
	code, stderr := runWithEnv(baseArgs, nil)
	if code == 2 {
		t.Errorf("code = 2 without -llm: %s", stderr)
	}
}

func TestMethodExists(t *testing.T) {
	const dir = "../../testdata/reachfixture/"
	h, _, err := hierarchy.Build(dir+"appentry.jar", []input.DepJar{
		{Path: dir + "vulnerablelib.jar", GAV: &input.GAV{Group: "org.vulnerable", Artifact: "lib", Version: "1.0"}},
	})
	if err != nil {
		t.Fatalf("hierarchy.Build: %v", err)
	}
	exists := methodExists(h)
	for _, tt := range []struct {
		class, method string
		want          bool
	}{
		{"reachfixture.VulnerableLib", "vulnerable", true},
		{"reachfixture.VulnerableLib", "invented", false},
		{"reachfixture.NoSuchClass", "vulnerable", false},
		{"reachfixture.VulnerableLib", "<init>", true},
		{"reachfixture.VulnerableLibSub", "vulnerable", true}, // inherited
	} {
		if got := exists(tt.class, tt.method); got != tt.want {
			t.Errorf("exists(%s, %s) = %v, want %v", tt.class, tt.method, got, tt.want)
		}
	}
}
