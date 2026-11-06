package main

import (
	"bytes"
	"strings"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

func runWithEnv(args []string, env map[string]string) (code int, stderr string) {
	var out, errBuf bytes.Buffer
	code = run(args, &out, &errBuf, func(k string) string { return env[k] })
	return code, errBuf.String()
}

var baseArgs = []string{"-app", "/nonexistent/app", "-out", "r.md", "-json", "r.json"}

func TestRootsFlagRejectsUnknownValue(t *testing.T) {
	args := append([]string{"-roots", "annotated"}, baseArgs...)
	code, stderr := runWithEnv(args, nil)
	if code != 2 || !strings.Contains(stderr, "-roots must be") {
		t.Errorf("code=%d stderr=%q, want 2 and a -roots message", code, stderr)
	}
}

func TestRootsFlagAcceptsAllAndProfile(t *testing.T) {
	for _, v := range []string{"all", "profile"} {
		args := append([]string{"-roots", v}, baseArgs...)
		// -app /nonexistent/app always fails past validation, inside
		// analyze(); what this checks is that validate() itself accepts
		// the value (exit code 2 means a flag-validation failure, which
		// this must not be).
		code, stderr := runWithEnv(args, nil)
		if code == 2 {
			t.Errorf("-roots %s: code=%d stderr=%q, want validation to accept this value", v, code, stderr)
		}
	}
}

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

func TestJPMSScenarioFlags(t *testing.T) {
	for _, tc := range []struct {
		args    []string
		invalid bool
		msg     string
	}{
		{[]string{"-jpms-scenario", "modulepath"}, true, "unknown JPMS scenario"},
		{[]string{"-jpms-scenario", "mixed"}, true, "needs -jpms-module-path"},
		{[]string{"-jpms-module-path", "a.jar"}, true, "only valid with -jpms-scenario mixed"},
		{[]string{"-jpms-scenario", "mixed", "-jpms-module-path", "[bad"}, true, "bad module-path jar pattern"},
		{[]string{"-jpms-scenario", "mixed", "-jpms-module-path", "app,commons-*.jar"}, false, ""},
		{[]string{"-jpms-scenario", "module-path"}, false, ""},
	} {
		code, stderr := runWithEnv(append(tc.args, baseArgs...), nil)
		if tc.invalid && (code != 2 || !strings.Contains(stderr, tc.msg)) {
			t.Errorf("%v: code=%d stderr=%q, want 2 and %q", tc.args, code, stderr, tc.msg)
		}
		if !tc.invalid && code == 2 {
			t.Errorf("%v: rejected by validation: %q", tc.args, stderr)
		}
	}
}

func TestJSONFormatFlag(t *testing.T) {
	code, stderr := runWithEnv(append([]string{"-json-format", "v3"}, baseArgs...), nil)
	if code != 2 || !strings.Contains(stderr, "-json-format must be") {
		t.Errorf("v3: code=%d stderr=%q, want 2 and a -json-format message", code, stderr)
	}
	for _, v := range []string{"v1", "v2"} {
		if code, stderr := runWithEnv(append([]string{"-json-format", v}, baseArgs...), nil); code == 2 {
			t.Errorf("-json-format %s: code=%d stderr=%q, want validation to accept it", v, code, stderr)
		}
	}
}

func TestFreshnessFlags(t *testing.T) {
	for _, c := range []struct {
		args []string
		msg  string
	}{
		{[]string{"-refresh", "-offline"}, "-refresh and -offline cannot be combined"},
		{[]string{"-data-ttl", "0s"}, "-data-ttl must be positive"},
		{[]string{"-max-data-age", "-1h"}, "-max-data-age must not be negative"},
	} {
		if code, stderr := runWithEnv(append(c.args, baseArgs...), nil); code != 2 || !strings.Contains(stderr, c.msg) {
			t.Errorf("%v: code=%d stderr=%q", c.args, code, stderr)
		}
	}
	for _, args := range [][]string{{"-refresh"}, {"-allow-stale", "-max-data-age", "48h"}, {"-offline", "-max-data-age", "1h"}} {
		if code, stderr := runWithEnv(append(args, baseArgs...), nil); code == 2 {
			t.Errorf("%v rejected: %q", args, stderr)
		}
	}
}

func TestSnapshotFlags(t *testing.T) {
	for _, args := range [][]string{{"-snapshot-in", "a", "-snapshot-out", "b"}, {"-snapshot-in", "a", "-refresh"}} {
		if code, stderr := runWithEnv(append(args, baseArgs...), nil); code != 2 || !strings.Contains(stderr, "-snapshot-in replays recorded data") {
			t.Errorf("%v: code=%d stderr=%q", args, code, stderr)
		}
	}
}

// TestDeclaredOverloads (CVE:F5b): bridge and synthetic methods are not
// overloads of the source method; an absent class is said so.
func TestDeclaredOverloads(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{{
		FQCN: "p.Box",
		Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "get", Descriptor: "()Ljava/lang/String;"}},
			{Method: classfile.Method{Name: "get", Descriptor: "()Ljava/lang/Object;", AccessFlags: 0x0040 | 0x1000}},
			{Method: classfile.Method{Name: "lambda$get$0", Descriptor: "()V", AccessFlags: 0x1000}},
		},
	}})
	f := declaredOverloads(h)
	if d, ok := f("p.Box", "get"); !ok || len(d) != 1 || d[0] != "()Ljava/lang/String;" {
		t.Errorf("get = %v %v, want the one non-bridge overload", d, ok)
	}
	if _, ok := f("p.Missing", "get"); ok {
		t.Error("absent class reported present")
	}
}
