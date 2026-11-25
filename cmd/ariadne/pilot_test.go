package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/research"
)

// RESEARCH_HARNESS RH-A3: the pilot control target testdata/pilot
// (scripts/pilot-fixture.sh, EXPECTED.MD there). Both variants of the
// library reach the same write method by the same path — static
// reachability cannot tell the defect from the fix, the property assertion
// of the experiment does — and the guard that tells them apart is visible
// in the corrected variant's calls. Every method the JVM ran is found;
// the controls (dead method, unresolved reflection, an API allowed to
// write anywhere) give their expected states.
func TestPilotTarget(t *testing.T) {
	type answer struct {
		state, model string
		roots        []string
		chars        string
	}
	answers := map[string]map[string]answer{}
	guards := map[string][]string{}
	for _, v := range []string{"defective", "fixed"} {
		dir := t.TempDir()
		out := filepath.Join(dir, "bundle")
		args := []string{"research", "index", "-app", "../../testdata/pilot/" + v + ".jar", "-roots", "api", "-profile", "generic",
			"-cache-dir", filepath.Join(dir, "empty-cache"), "-out", out}
		if code, msg := runWithEnv(args, nil); code != 0 {
			t.Fatalf("%s: research index: exit %d: %s", v, code, msg)
		}
		b, err := research.Load(out)
		if err != nil {
			t.Fatal(err)
		}
		lib := instanceOf(b, "app:"+v+".jar")
		names := map[string]string{}
		for _, n := range b.Nodes {
			if n.Method != nil {
				names[n.NodeID] = n.Method.Class + "#" + n.Method.Name + n.Method.Descriptor
			}
		}
		// Targets: every px.archive method the JVM ran, plus the controls.
		targets := []any{}
		ids := []string{}
		add := func(id, key string) {
			tid := fmt.Sprintf("t%d", len(ids))
			class, rest, _ := strings.Cut(key, "#")
			i := strings.Index(rest, "(")
			targets = append(targets, map[string]any{"id": tid, "selector": map[string]any{"instance_id": lib, "class": class, "name": rest[:i], "descriptor": rest[i:]}})
			ids = append(ids, id)
		}
		for _, m := range observed(t, "../../testdata/pilot/observed-"+v+".txt") {
			add("ran:"+m, m)
		}
		add("dead", "px.archive.internal.Store#overwrite(Ljava/nio/file/Path;[B)V")
		_, doc := researchQuery(t, out, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": v,
			"operation": "reachability", "roots": map[string]any{"selection": "indexed"}, "targets": targets})
		answers[v] = map[string]answer{}
		for i, r := range doc["results"].([]any) {
			rm := r.(map[string]any)
			a := answer{state: rm["search"].(map[string]any)["state"].(string), model: rm["model"].(map[string]any)["state"].(string)}
			chars := map[string]bool{}
			for _, e := range rm["entries"].([]any) {
				em := e.(map[string]any)
				a.roots = append(a.roots, names[em["root_node_id"].(string)])
				for _, c := range em["character"].([]any) {
					chars[c.(string)] = true
				}
			}
			sort.Strings(a.roots)
			for c := range chars {
				a.chars += c + " "
			}
			answers[v][ids[i]] = a
			if strings.HasPrefix(ids[i], "ran:") && a.state != "found" {
				t.Errorf("%s: the JVM ran %s, the bundle says %s", v, ids[i], a.state)
			}
		}
		if a := answers[v]["dead"]; a.state != "not_found" {
			t.Errorf("%s: dead control %s", v, a.state)
		}
		// Unresolved reflection (Unpacker#filter) is a relevant gap of every target.
		if a := answers[v]["ran:px.archive.internal.Store#write(Ljava/nio/file/Path;[B)V"]; a.model != "incomplete" {
			t.Errorf("%s: Store#write model %s, want incomplete (reflection control)", v, a.model)
		}
		// The guard: Path calls in Entries#target.
		_, doc = researchQuery(t, out, map[string]any{"schema_version": research.QuerySchema, "analysis_id": b.AnalysisID, "request_id": v + "-calls",
			"operation": "calls", "filters": map[string]any{"class_prefix": "px.archive.Entries", "declared_class_prefix": "java.nio.file.Path"}})
		for _, r := range doc["records"].([]any) {
			guards[v] = append(guards[v], r.(map[string]any)["declared_name"].(string))
		}
		sort.Strings(guards[v])
	}
	// Same write method, same entries, same evidence character in both
	// variants: the static axis does not decide the property.
	w := "ran:px.archive.internal.Store#write(Ljava/nio/file/Path;[B)V"
	d, f := answers["defective"][w], answers["fixed"][w]
	if strings.Join(d.roots, ",") != strings.Join(f.roots, ",") || d.chars != f.chars || d.state != "found" {
		t.Errorf("Store#write differs between variants: %+v vs %+v", d, f)
	}
	// Store is a public class, so Store#write is API itself (a zero-edge
	// entry); -api-exclude px.archive.internal would leave it out.
	want := []string{"px.archive.Unpacker#unpack(Ljava/io/InputStream;)I", "px.archive.Unpacker#writeFile(Ljava/nio/file/Path;[B)V",
		"px.archive.internal.Store#write(Ljava/nio/file/Path;[B)V"}
	if strings.Join(d.roots, ",") != strings.Join(want, ",") {
		t.Errorf("Store#write entries %v, want %v (unpack, the allowed writeFile control and Store#write itself)", d.roots, want)
	}
	if strings.Join(guards["defective"], ",") != "resolve" || !strings.Contains(strings.Join(guards["fixed"], ","), "startsWith") {
		t.Errorf("Path calls in Entries#target: defective %v, fixed %v", guards["defective"], guards["fixed"])
	}
	// The JVM: same methods, different effect.
	effects := func(v string) string { b, _ := os.ReadFile("../../testdata/pilot/effects-" + v + ".txt"); return string(b) }
	if !strings.Contains(effects("defective"), "traversal unpacked 1, file outside dest true") ||
		!strings.Contains(effects("fixed"), "traversal refused, file outside dest false") {
		t.Errorf("effects:\n%s\n%s", effects("defective"), effects("fixed"))
	}
}

// observed lists the owner#name+desc lines of a MethodTrace file.
func observed(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if l := sc.Text(); l != "" && !strings.HasPrefix(l, "#") && !strings.Contains(l, "#<init>") {
			out = append(out, l)
		}
	}
	return out
}

// RH-A3 extension: a binary with conflicting sources. defective.jar indexed
// with the corrected variant's sources gets that file's hash under the
// binding "declared": Ariadne matches sources by path and cannot see that
// they were not the build's. Deciding correspondence (rebuild, compare)
// is the consumer's; the binding says no more was checked.
func TestPilotConflictingSources(t *testing.T) {
	src := t.TempDir()
	for _, d := range []string{"common", "fixed"} {
		copyTree(t, filepath.Join("../../testdata/pilot", d), src)
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "bundle")
	args := []string{"research", "index", "-app", "../../testdata/pilot/defective.jar", "-roots", "api", "-profile", "generic",
		"-cache-dir", filepath.Join(dir, "empty-cache"), "-sources", src, "-out", out}
	if code, msg := runWithEnv(args, nil); code != 0 {
		t.Fatalf("research index: exit %d: %s", code, msg)
	}
	b, err := research.Load(out)
	if err != nil {
		t.Fatal(err)
	}
	fixedSrc, _ := os.ReadFile("../../testdata/pilot/fixed/px/archive/Entries.java")
	want := fmt.Sprintf("sha256:%x", sha256.Sum256(fixedSrc))
	found := false
	for _, n := range b.Nodes {
		if n.Method != nil && n.Method.Class == "px.archive.Entries" && n.Method.Name == "target" {
			found = true
			if n.Source.SourceSHA256 == nil || *n.Source.SourceSHA256 != want || n.Source.Binding != "declared" {
				t.Errorf("Entries#target source %+v, want the given (fixed) file's hash, binding declared", n.Source)
			}
		}
	}
	if !found {
		t.Fatal("no Entries#target node")
	}
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(to, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(to, rel), b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}
