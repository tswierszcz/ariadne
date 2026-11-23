package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/nativebound"
)

// TestNativeBoundaries (RUNTIME_NATIVE N5 n5-1): testdata/native
// (scripts/native-fixture.sh) declares a native method main calls, one
// nothing calls, a System.loadLibrary of a literal and a
// Runtime#loadLibrary of a parameter. The reachable native method is a
// model gap of kind "native"; the report and v2 list the boundaries.
func TestNativeBoundaries(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "r.json")
	md := filepath.Join(dir, "r.md")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", "../../testdata/native/app.jar",
		"-profile", "generic", "-roots", "profile", "-json-format", "v2", "-out", md, "-json", out}
	if code, msg := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, msg)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Coverage struct {
			Native *nativebound.Summary `json:"native"`
			Gaps   []struct {
				Mechanism, Source, Detail string
				Count                     int
				Reachable                 *int
			} `json:"gaps"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	n := doc.Coverage.Native
	if n == nil || n.Version != nativebound.Version || n.Methods != 2 || len(n.Reachable) != 1 || n.Reachable[0].Name != "encode" || !n.Reachable[0].App {
		t.Fatalf("coverage.native = %+v", n)
	}
	if len(n.Loads) != 2 {
		t.Fatalf("loads = %+v", n.Loads)
	}
	lit, param := n.Loads[0], n.Loads[1]
	if lit.API != "System#loadLibrary" || lit.Library != "fxcodec" || lit.CallerName != "<clinit>" || !lit.Reachable {
		t.Errorf("literal load = %+v", lit)
	}
	if param.API != "Runtime#loadLibrary" || param.Library != "" || param.CallerName != "loadByName" || !param.Reachable {
		t.Errorf("parameter load = %+v", param)
	}
	found := false
	for _, g := range doc.Coverage.Gaps {
		if g.Mechanism == "native" {
			found = true
			if g.Source != "site" || g.Detail != "native" || g.Count != 2 || g.Reachable == nil || *g.Reachable != 1 {
				t.Errorf("native gap = %+v", g)
			}
		}
	}
	if !found {
		t.Error("no native gap in coverage.gaps")
	}
	rep, _ := os.ReadFile(md)
	for _, want := range []string{"## Native code boundaries", "`fx.nat.Main#encode([B)I` | application", "System#loadLibrary | `fxcodec` | yes", "(not a literal)"} {
		if !strings.Contains(string(rep), want) {
			t.Errorf("report lacks %q", want)
		}
	}
	if strings.Contains(string(rep), "unused") {
		t.Error("the unreachable native method is listed as reachable")
	}
}
