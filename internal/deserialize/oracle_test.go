package deserialize

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const oracleDir = "../../testdata/deserialize/oracle/"

// readTSV reads the non-comment rows of a tab-separated file.
func readTSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var rows [][]string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows
}

// TestDeserializationOracle (DESERIALIZATION Z4, ROADMAP B4): ariadne's
// model of the oracle fixture against what the real JVM ran
// (scripts/deserialize-oracle.sh → observed.tsv). Per entry, a callback
// the JVM ran but the model does not connect must be a documented gap
// (model-gaps.tsv), and one the model connects but the JVM did not run a
// documented extra (model-extras.tsv); both lists must be exact, so a
// change of the model or of the observation fails until it is explained.
// Covered: private callbacks (readObject, readResolve), a public
// readObject (no callback), Externalizable (constructor + readExternal),
// a record (special type, not covered), a class of unknown
// serializability, two instances of one library class (plain and
// relocated), and an evaluated filter.
func TestDeserializationOracle(t *testing.T) {
	h, _, err := hierarchy.Build(oracleDir+"scenarios.jar", []input.DepJar{{Path: oracleDir + "lib.jar"}, {Path: oracleDir + "shaded.jar"}})
	if err != nil {
		t.Fatal(err)
	}
	sites := FindSites(h)
	if w := ResolvePolicies(h, sites); len(w) > 0 {
		t.Fatalf("warnings %v", w)
	}
	java, sum := ClassTargets(h, FormatJava)
	edges := Connect(sites, map[string]map[string][]Target{FormatJava: java})
	model := map[string]map[string]bool{}
	for _, s := range sites {
		model[s.CallerFQCN+"#"+s.CallerName] = map[string]bool{}
		if s.CallerName == "filtered" && s.Filter != FilterEvaluated {
			t.Errorf("filtered entry: policy %s, want %s", s.Filter, FilterEvaluated)
		}
	}
	if len(model) != 2 {
		t.Fatalf("entries %v, want sc.Entries#plain and #filtered", model)
	}
	for _, e := range edges {
		model[e.CallerFQCN+"#"+e.CallerName][e.TargetFQCN+"#"+e.TargetName] = true
	}
	observed := map[string]map[string]bool{}
	for _, r := range readTSV(t, oracleDir+"observed.tsv") {
		if observed[r[0]] == nil {
			observed[r[0]] = map[string]bool{}
		}
		if r[2] == "-" {
			continue
		}
		for _, cb := range strings.Split(r[2], ",") {
			observed[r[0]][cb] = true
		}
	}
	declared := func(file string) map[string]bool {
		out := map[string]bool{}
		for _, r := range readTSV(t, oracleDir+file) {
			if len(r) < 3 || r[2] == "" {
				t.Errorf("%s: row %v without a reason", file, r)
				continue
			}
			out[r[0]+"\t"+r[1]] = true
		}
		return out
	}
	gaps, extras := declared("model-gaps.tsv"), declared("model-extras.tsv")
	gotGaps, gotExtras := map[string]bool{}, map[string]bool{}
	for entry, obs := range observed {
		for cb := range obs {
			if !model[entry][cb] {
				gotGaps[entry+"\t"+cb] = true
			}
		}
		for cb := range model[entry] {
			if !obs[cb] {
				gotExtras[entry+"\t"+cb] = true
			}
		}
	}
	diff := func(name string, got, want map[string]bool) {
		var missing, unexpected []string
		for k := range got {
			if !want[k] {
				unexpected = append(unexpected, k)
			}
		}
		for k := range want {
			if !got[k] {
				missing = append(missing, k)
			}
		}
		sort.Strings(missing)
		sort.Strings(unexpected)
		if len(missing)+len(unexpected) > 0 {
			t.Errorf("%s: undeclared %q; declared but not found %q", name, unexpected, missing)
		}
	}
	diff("model gaps (JVM ran, model does not connect)", gotGaps, gaps)
	diff("model extras (model connects, JVM did not run)", gotExtras, extras)
	// The two scenarios the model declares as not covered are counted,
	// never silently dropped.
	if sum.Records != 1 || sum.Unknown != 1 || sum.UnknownBases["java.util.concurrent.LinkedBlockingQueue"] != 1 {
		t.Errorf("summary: records %d, unknown %d %v; want the record and the queue counted", sum.Records, sum.Unknown, sum.UnknownBases)
	}
}
