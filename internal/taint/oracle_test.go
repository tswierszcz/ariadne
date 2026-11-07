package taint_test

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/taint"
)

const observedTSV = "../../testdata/taint/oracle/observed.tsv"

// readObserved returns the entries the real JVM saw reading the Canary
// sent as request data (scripts/taint-oracle.sh), with "read" or
// "rejected".
func readObserved(t *testing.T) map[string]string {
	t.Helper()
	f, err := os.Open(observedTSV)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "\t") {
			continue
		}
		k, v, _ := strings.Cut(line, "\t")
		out[k] = v
	}
	if len(out) == 0 {
		t.Fatal("observed.tsv has no entries")
	}
	return out
}

// TestTaintOracle (DATA_FLOW D5, ROADMAP B4): the model against the real
// JVM. Every entry the JVM saw reading request data must be flow_found
// or undetermined — never no_flow_in_model (a miss); every flow_found
// must have been observed (none unconfirmed on this fixture); an entry
// the oracle never reached is not proven safe, only not contradicted.
// The undetermined-but-observed entries are the measured gaps, pinned
// here so a change in them is seen.
func TestTaintOracle(t *testing.T) {
	observed := readObserved(t)
	byKey := map[string]string{}
	for _, s := range runSites(t, taint.DefaultOptions) {
		byKey[s.CallerFQCN+"#"+s.CallerName+s.CallerDesc] = s.Taint.State
	}
	var misses, unconfirmed, gaps []string
	for k, how := range observed {
		st, ok := byKey[k]
		if !ok {
			t.Errorf("observed entry %s has no model result", k)
			continue
		}
		switch st {
		case taint.NoFlow, taint.NotAssessed:
			misses = append(misses, k)
		case taint.Undetermined:
			gaps = append(gaps, k)
		case taint.FlowFound:
		}
		_ = how
	}
	for k, st := range byKey {
		if st == taint.FlowFound && observed[k] == "" {
			unconfirmed = append(unconfirmed, k)
		}
	}
	sort.Strings(gaps)
	if len(misses) > 0 {
		t.Errorf("model says no flow where the JVM read request data: %v", misses)
	}
	if len(unconfirmed) > 0 {
		t.Errorf("flow_found never observed by the oracle: %v", unconfirmed)
	}
	wantGaps := []string{
		"demo.Codec#readDeep([B)Ljava/lang/Object;",                          // depth limit (9 helper levels)
		"demo.Codec#readOther([B)Ljava/lang/Object;",                         // through an instance field
		"demo.Loader#inCallback(Ljava/lang/String;)Ljava/lang/Object;",       // called from a lambda
		"demo.Shapes#exceptionMessage(Ljava/lang/String;)Ljava/lang/Object;", // Throwable#getMessage not modeled
	}
	if strings.Join(gaps, "\n") != strings.Join(wantGaps, "\n") {
		t.Errorf("observed but undetermined:\n%s\nwant:\n%s", strings.Join(gaps, "\n"), strings.Join(wantGaps, "\n"))
	}
	if observed["demo.Endpoints#filtered(Ljava/lang/String;)Ljava/lang/Object;"] != "rejected" {
		t.Errorf("the filtered entry's own filter should reject the Canary")
	}
}

// TestShapes (DATA_FLOW D5): the plan's shapes each get their expected
// state on the fixture.
func TestShapes(t *testing.T) {
	byKey := map[string]string{}
	for _, s := range runSites(t, taint.DefaultOptions) {
		byKey[s.CallerFQCN+"#"+s.CallerName+s.CallerDesc] = s.Taint.State
	}
	want := map[string]string{
		"demo.Shapes#aliasing(Ljava/lang/String;)Ljava/lang/Object;":         taint.FlowFound,
		"demo.Shapes#arrays(Ljava/lang/String;)Ljava/lang/Object;":           taint.FlowFound,
		"demo.Shapes#exceptionLocal(Ljava/lang/String;)Ljava/lang/Object;":   taint.FlowFound,
		"demo.Shapes#exceptionMessage(Ljava/lang/String;)Ljava/lang/Object;": taint.Undetermined,
		"demo.Shapes#twoInstances(Ljava/lang/String;)Ljava/lang/Object;":     taint.NoFlow,
		"demo.Loader#load(Ljava/lang/String;)Ljava/lang/Object;":             taint.FlowFound,
		"demo.Loader#load([B)Ljava/lang/Object;":                             taint.NoFlow,
		"demo.Loader#inCallback(Ljava/lang/String;)Ljava/lang/Object;":       taint.Undetermined,
		"demo.Batch#run(Ljava/lang/String;)Ljava/lang/Object;":               taint.NoFlow,
	}
	for k, st := range want {
		if byKey[k] != st {
			t.Errorf("%s: %s, want %s", k, byKey[k], st)
		}
	}
}
