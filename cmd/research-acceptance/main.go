// Command research-acceptance builds the producer acceptance package of the
// research interface (RESEARCH_HARNESS RH-A0, docs/research-interface/
// ACCEPTANCE.md): it indexes the control library testdata/research with a
// given ariadne binary (-scope all and reachable), answers every request of
// testdata/research/acceptance/requests.json with that binary, requires each
// response to pass research.CheckResponse and replay, then derives mutated
// responses, each of which must break the rule it names. Everything is
// written under -out with a SHA256SUMS file; an existing -out is refused.
//
// Usage: go run ./cmd/research-acceptance -ariadne bin/ariadne -out DIR
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/research"
)

type spec struct {
	ID      string         `json:"id"`
	Scopes  []string       `json:"scopes"`
	Note    string         `json:"note"`
	Request map[string]any `json:"request"`
}

// Case is one entry of cases.json.
type Case struct {
	ID            string   `json:"id"`
	Kind          string   `json:"kind"` // response | mutation
	Base          string   `json:"base,omitempty"`
	Bundle        string   `json:"bundle"`
	Request       string   `json:"request"`
	Response      string   `json:"response"`
	ExpectedValid bool     `json:"expected_valid"`
	ExpectedRules []string `json:"expected_rules"`
	Note          string   `json:"note"`
}

// Package is cases.json.
type Package struct {
	Schema           string            `json:"schema"`
	ContractRevision string            `json:"contract_revision"`
	Producer         map[string]string `json:"producer"`
	Bundles          map[string]string `json:"bundles"` // scope -> analysis_id
	Rules            []string          `json:"rules"`
	Cases            []Case            `json:"cases"`
}

func main() {
	bin := flag.String("ariadne", "bin/ariadne", "ariadne binary that produces the bundles and responses")
	out := flag.String("out", "", "package directory to create")
	reqs := flag.String("requests", "testdata/research/acceptance/requests.json", "request specs")
	fx := flag.String("fixture", "testdata/research", "control library directory")
	flag.Parse()
	if *out == "" {
		fail("-out is required")
	}
	if abs, err := filepath.Abs(*bin); err == nil {
		*bin = abs
	}
	if _, err := os.Stat(*out); err == nil {
		fail("%s exists", *out)
	}
	raw, err := os.ReadFile(*reqs)
	check(err)
	var specs struct {
		Cases []spec `json:"cases"`
	}
	check(json.Unmarshal(raw, &specs))
	check(os.MkdirAll(filepath.Join(*out, "requests"), 0o755))
	check(os.MkdirAll(filepath.Join(*out, "responses"), 0o755))
	check(os.MkdirAll(filepath.Join(*out, "mutations"), 0o755))
	binSum := fileSHA(*bin)
	pkg := Package{Schema: "ariadne-research-acceptance-1", ContractRevision: research.ContractRevision,
		Producer: map[string]string{"binary_sha256": "sha256:" + binSum, "built_by": goVersion()},
		Bundles:  map[string]string{}, Rules: research.Rules}
	bundles := map[string]*research.Bundle{}
	empty := filepath.Join(*out, ".empty-cache")
	check(os.MkdirAll(empty, 0o755))
	for _, scope := range []string{"all", "reachable"} {
		dir := filepath.Join(*out, "bundle-"+scope)
		run(*bin, "research", "index", "-app", filepath.Join(*fx, "lib.jar"), "-deps", filepath.Join(*fx, "deps"),
			"-deps-list", filepath.Join(*fx, "deps.txt"), "-roots", "api", "-profile", "generic",
			"-cache-dir", empty, "-scope", scope, "-out", dir)
		b, err := research.Load(dir)
		check(err)
		bundles[scope] = b
		pkg.Bundles[scope] = b.AnalysisID
		if pkg.Producer["revision"] == "" {
			pkg.Producer["producer"] = b.Identity.Provenance.Producer
			pkg.Producer["revision"] = b.Identity.Provenance.Revision
		}
	}
	os.RemoveAll(empty)
	cursors := map[string]string{}
	envelopeDone := map[string]bool{}
	for _, s := range specs.Cases {
		for _, scope := range s.Scopes {
			b := bundles[scope]
			id := s.ID + "." + scope
			reqDoc := fill(s.Request, b.AnalysisID, cursors)
			var resp []byte
			reqPath := filepath.Join("requests", id+".json")
			respPath := filepath.Join("responses", id+".json")
			if v, ok := reqDoc["limits"].(map[string]any); ok && v["max_work_units"] == "@FOUND_INCOMPLETE@" {
				reqDoc, resp = foundIncomplete(*bin, *out, b, reqDoc, reqPath, respPath)
			} else {
				writeJSON(filepath.Join(*out, reqPath), reqDoc)
				resp = query(*bin, *out, "bundle-"+scope, reqPath, respPath)
			}
			reqRaw, _ := os.ReadFile(filepath.Join(*out, reqPath))
			if v := research.CheckResponse(b, reqRaw, resp); len(v) > 0 {
				fail("%s: the producer's own response breaks %v", id, v)
			}
			if v := research.Replay(b, reqRaw, resp); v != nil {
				fail("%s: replay differs", id)
			}
			var head struct {
				Page *struct {
					NextCursor *string `json:"next_cursor"`
				} `json:"page"`
			}
			json.Unmarshal(resp, &head)
			if head.Page != nil && head.Page.NextCursor != nil {
				cursors[s.ID] = *head.Page.NextCursor
			}
			pkg.Cases = append(pkg.Cases, Case{ID: id, Kind: "response", Bundle: "bundle-" + scope, Request: reqPath, Response: respPath,
				ExpectedValid: true, ExpectedRules: []string{}, Note: s.Note})
			kind := responseKind(resp)
			for _, m := range mutations(b, resp) {
				// Envelope mutations once per kind of response.
				if m.envelope {
					if envelopeDone[kind+m.id] {
						continue
					}
					envelopeDone[kind+m.id] = true
				}
				mid := id + "--" + m.id
				mp := filepath.Join("mutations", mid+".json")
				check(os.WriteFile(filepath.Join(*out, mp), m.raw, 0o644))
				got := research.CheckResponse(b, reqRaw, m.raw)
				if !hasRule(got, m.rule) {
					fail("%s: expected rule %s, the check reports %v", mid, m.rule, got)
				}
				pkg.Cases = append(pkg.Cases, Case{ID: mid, Kind: "mutation", Base: id, Bundle: "bundle-" + scope, Request: reqPath, Response: mp,
					ExpectedValid: false, ExpectedRules: []string{m.rule}, Note: m.note})
			}
		}
	}
	writeJSON(filepath.Join(*out, "cases.json"), pkg)
	writeSums(*out)
	nm := 0
	for _, c := range pkg.Cases {
		if c.Kind == "mutation" {
			nm++
		}
	}
	fmt.Printf("acceptance package %s: %d responses, %d mutations\n", *out, len(pkg.Cases)-nm, nm)
}

// fill substitutes @ANALYSIS_ID@ and @CURSOR:id@ placeholders.
func fill(tmpl map[string]any, analysisID string, cursors map[string]string) map[string]any {
	raw, _ := json.Marshal(tmpl)
	s := strings.ReplaceAll(string(raw), "@ANALYSIS_ID@", analysisID)
	for id, c := range cursors {
		s = strings.ReplaceAll(s, "@CURSOR:"+id+"@", c)
	}
	var m map[string]any
	d := json.NewDecoder(strings.NewReader(s))
	d.UseNumber()
	check(d.Decode(&m))
	return m
}

// foundIncomplete finds the smallest work budget at which some target is
// found while the search stays incomplete (a witness, then the budget ends).
func foundIncomplete(bin, out string, b *research.Bundle, req map[string]any, reqPath, respPath string) (map[string]any, []byte) {
	for w := 1; w <= 100000; w++ {
		req["limits"].(map[string]any)["max_work_units"] = w
		writeJSON(filepath.Join(out, reqPath), req)
		resp := query(bin, out, bundleDirOf(b, out), reqPath, respPath)
		var r research.ReachabilityResponse
		check(json.Unmarshal(resp, &r))
		for _, t := range r.Results {
			if t.Search.State == "found" && !t.Search.WorkComplete {
				return req, resp
			}
		}
		if allComplete(r) {
			break
		}
	}
	fail("no budget gives a found target with an incomplete search")
	return nil, nil
}

func allComplete(r research.ReachabilityResponse) bool {
	for _, t := range r.Results {
		if !t.Search.WorkComplete {
			return false
		}
	}
	return true
}

func bundleDirOf(b *research.Bundle, out string) string {
	rel, _ := filepath.Rel(out, b.Dir)
	return rel
}

func query(bin, out, bundle, reqPath, respPath string) []byte {
	cmd := exec.Command(bin, "research", "query", "-bundle", bundle, "-request", reqPath, "-out", respPath)
	cmd.Dir = out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run() // an error envelope exits non-zero; the response file decides
	resp, err := os.ReadFile(filepath.Join(out, respPath))
	if err != nil {
		fail("%s: no response: %s", reqPath, stderr.String())
	}
	return resp
}

type mutation struct {
	id, rule, note string
	raw            []byte
	envelope       bool
}

// responseKind: reachability, discovery or error.
func responseKind(resp []byte) string {
	d := decode(resp)
	switch {
	case d["status"] == "error":
		return "error"
	case d["operation"] == "reachability":
		return "reachability"
	}
	return "discovery"
}

// mutations derives broken variants of a valid response. Each must make
// research.CheckResponse report its rule; a consumer must reject them all.
func mutations(b *research.Bundle, resp []byte) []mutation {
	var out []mutation
	add := func(id, rule, note string, edit func(doc map[string]any) bool) {
		doc := decode(resp)
		if !edit(doc) {
			return
		}
		raw, err := json.MarshalIndent(doc, "", "  ")
		check(err)
		out = append(out, mutation{id: id, rule: rule, note: note, raw: append(raw, '\n')})
	}
	status := decode(resp)["status"]
	// Envelope mutations, any response.
	out = append(out, mutation{id: "duplicate-member", rule: research.RuleJSON, note: "a repeated top-level member",
		raw: append([]byte("{\n  \"status\": \"ok\","), resp[1:]...), envelope: true})
	add("unknown-member", research.RuleShape, "an extra top-level member", func(d map[string]any) bool { d["confidence"] = "high"; return true })
	if status == "error" {
		add("missing-error-code", research.RuleShape, "the error object without its code", func(d map[string]any) bool {
			delete(d["error"].(map[string]any), "code")
			return true
		})
		for i := range out {
			out[i].envelope = true
		}
		return out
	}
	add("wrong-analysis", research.RuleBinding, "analysis_id of another bundle", func(d map[string]any) bool {
		d["analysis_id"] = "sha256:" + strings.Repeat("ab", 32)
		return true
	})
	add("wrong-request-id", research.RuleBinding, "request_id of another request", func(d map[string]any) bool {
		d["request_id"] = "someone-else"
		return true
	})
	add("wrong-origin", research.RuleBinding, "provenance of another producer build", func(d map[string]any) bool {
		d["provenance"].(map[string]any)["revision"] = "another-build"
		return true
	})
	add("wrong-revision", research.RuleVersion, "another contract revision", func(d map[string]any) bool {
		d["contract_revision"] = "1.0.0-draft.0"
		return true
	})
	for i := range out {
		out[i].envelope = true
	}
	if d := decode(resp); d["operation"] != "reachability" {
		discoveryMutations(b, resp, add)
	} else {
		reachMutations(b, resp, add)
	}
	return out
}

func discoveryMutations(b *research.Bundle, resp []byte, add func(string, string, string, func(map[string]any) bool)) {
	records := func(d map[string]any) []any { return d["records"].([]any) }
	add("record-modified", research.RuleRecordOriginal, "a record member changed", func(d map[string]any) bool {
		rs := records(d)
		if len(rs) == 0 {
			return false
		}
		r := rs[0].(map[string]any)
		for _, k := range []string{"basis", "body", "declared_name", "character"} {
			if _, ok := r[k]; ok {
				r[k] = "changed"
				return true
			}
		}
		return false
	})
	add("record-dropped", research.RulePage, "returned_count no longer counts the records", func(d map[string]any) bool {
		rs := records(d)
		if len(rs) == 0 {
			return false
		}
		d["records"] = rs[1:]
		return true
	})
	add("cursor-on-last-page", research.RulePage, "a cursor on an untruncated page", func(d map[string]any) bool {
		p := d["page"].(map[string]any)
		if p["truncated"] == true {
			return false
		}
		p["next_cursor"] = "c1.bogus"
		return true
	})
	add("truncated-without-cursor", research.RulePage, "a truncated page with no cursor", func(d map[string]any) bool {
		p := d["page"].(map[string]any)
		if p["truncated"] != true {
			return false
		}
		p["next_cursor"] = nil
		return true
	})
	op := decode(resp)["operation"]
	if op == "callers" || op == "callees" {
		add("unrelated-edge", research.RuleNeighbor, "an edge of another node in place of the first record", func(d map[string]any) bool {
			rs := records(d)
			if len(rs) == 0 {
				return false
			}
			first := rs[0].(map[string]any)
			for i := range b.Edges {
				e := b.Edges[i]
				if e.To != first["to"] && e.From != first["from"] {
					rs[0] = generic(&e)
					return true
				}
			}
			return false
		})
	}
}

func reachMutations(b *research.Bundle, resp []byte, add func(string, string, string, func(map[string]any) bool)) {
	results := func(d map[string]any) []any { return d["results"].([]any) }
	// longest path of a found target: its result and entry index.
	pick := func(d map[string]any, minEdges int) (map[string]any, map[string]any) {
		var bestR, bestE map[string]any
		best := -1
		for _, r := range results(d) {
			rm := r.(map[string]any)
			for _, e := range rm["entries"].([]any) {
				em := e.(map[string]any)
				n := len(em["path"].(map[string]any)["edge_ids"].([]any))
				if n >= minEdges && n > best {
					best, bestR, bestE = n, rm, em
				}
			}
		}
		return bestR, bestE
	}
	firstWith := func(d map[string]any, pred func(map[string]any) bool) map[string]any {
		for _, r := range results(d) {
			if rm := r.(map[string]any); pred(rm) {
				return rm
			}
		}
		return nil
	}
	state := func(r map[string]any) string { return r["search"].(map[string]any)["state"].(string) }
	add("missing-with-path", research.RuleResolution, "a found target declared missing, its path kept", func(d map[string]any) bool {
		r := firstWith(d, func(r map[string]any) bool { return state(r) == "found" })
		if r == nil {
			return false
		}
		r["resolution"] = "missing"
		return true
	})
	add("found-without-entries", research.RuleSearchState, "found with the path evidence removed", func(d map[string]any) bool {
		r := firstWith(d, func(r map[string]any) bool { return state(r) == "found" })
		if r == nil {
			return false
		}
		r["entries"] = []any{}
		return true
	})
	add("not-found-to-found", research.RuleSearchState, "a not_found target claimed found without a path", func(d map[string]any) bool {
		r := firstWith(d, func(r map[string]any) bool { return state(r) == "not_found" })
		if r == nil {
			return false
		}
		r["search"].(map[string]any)["state"] = "found"
		return true
	})
	add("unknown-count-zeroed", research.RuleEntryCount, "an unknown total_entry_count (null) turned into 0", func(d map[string]any) bool {
		r := firstWith(d, func(r map[string]any) bool { return r["total_entry_count"] == nil })
		if r == nil {
			return false
		}
		r["total_entry_count"] = 0
		return true
	})
	add("budget-claimed-complete", research.RuleEntryCount, "a search the budget ended reported complete, count still unknown", func(d map[string]any) bool {
		r := firstWith(d, func(r map[string]any) bool {
			return state(r) == "found" && r["search"].(map[string]any)["work_complete"] == false
		})
		if r == nil {
			return false
		}
		r["search"].(map[string]any)["work_complete"] = true
		return true
	})
	add("work-over-budget", research.RuleLimits, "work units above max_work_units", func(d map[string]any) bool {
		r := results(d)[0].(map[string]any)
		lim := num(d["effective_limits"].(map[string]any)["max_work_units"])
		r["search"].(map[string]any)["work_units"] = lim + 1
		return true
	})
	add("model-complete-with-gaps", research.RuleModel, "complete_under_declared_model with relevant gaps", func(d map[string]any) bool {
		r := firstWith(d, func(r map[string]any) bool {
			return len(r["model"].(map[string]any)["relevant_gap_ids"].([]any)) > 0
		})
		if r == nil {
			return false
		}
		r["model"].(map[string]any)["state"] = "complete_under_declared_model"
		r["model"].(map[string]any)["scope_uncertain"] = false
		return true
	})
	add("path-gap", research.RulePathContinuity, "a middle node and its edge removed from a path", func(d map[string]any) bool {
		_, e := pick(d, 2)
		if e == nil {
			return false
		}
		p := e["path"].(map[string]any)
		nodes, edges := p["node_ids"].([]any), p["edge_ids"].([]any)
		p["node_ids"] = append(append([]any{}, nodes[:1]...), nodes[2:]...)
		p["edge_ids"] = edges[1:]
		if p["truncated"] == false {
			p["full_node_count"] = len(nodes) - 1
		}
		return true
	})
	add("path-edge-swapped", research.RulePathContinuity, "a path edge replaced by an edge between other nodes", func(d map[string]any) bool {
		_, e := pick(d, 1)
		if e == nil {
			return false
		}
		edges := e["path"].(map[string]any)["edge_ids"].([]any)
		for i := range b.Edges {
			if ed := b.Edges[i]; ed.EdgeID != edges[0] && ed.From != e["root_node_id"] {
				edges[0] = ed.EdgeID
				return true
			}
		}
		return false
	})
	add("path-cut-silently", research.RulePathEnd, "the last step removed, the path still marked complete", func(d map[string]any) bool {
		_, e := pick(d, 1)
		if e == nil {
			return false
		}
		p := e["path"].(map[string]any)
		if p["truncated"] == true {
			return false
		}
		nodes, edges := p["node_ids"].([]any), p["edge_ids"].([]any)
		p["node_ids"] = nodes[:len(nodes)-1]
		p["edge_ids"] = edges[:len(edges)-1]
		p["full_node_count"] = len(nodes) - 1
		return true
	})
	add("path-wrong-root", research.RulePathRoot, "an entry's root replaced by another node of its path", func(d map[string]any) bool {
		_, e := pick(d, 1)
		if e == nil {
			return false
		}
		e["root_node_id"] = e["path"].(map[string]any)["node_ids"].([]any)[1]
		return true
	})
	add("character-dropped", research.RulePathCharacter, "an edge character removed from an entry", func(d map[string]any) bool {
		_, e := pick(d, 1)
		if e == nil {
			return false
		}
		ch := e["character"].([]any)
		if len(ch) == 0 {
			return false
		}
		e["character"] = ch[1:]
		return true
	})
	add("pass-unknown", research.RulePathPass, "an entry's pass outside app_only/full", func(d map[string]any) bool {
		_, e := pick(d, 0)
		if e == nil {
			return false
		}
		e["pass"] = "fast"
		return true
	})
	add("evidence-modified", research.RuleEvidenceOriginal, "an evidence edge's character changed", func(d map[string]any) bool {
		for _, r := range d["evidence"].([]any) {
			if rm := r.(map[string]any); rm["record_type"] == "edge" {
				if rm["character"] == "heuristic" {
					rm["character"] = "jvm_rule"
				} else {
					rm["character"] = "heuristic"
				}
				return true
			}
		}
		return false
	})
	add("evidence-dropped", research.RuleEvidenceMissing, "an edge of a path missing from the evidence", func(d map[string]any) bool {
		ev := d["evidence"].([]any)
		for i, r := range ev {
			if r.(map[string]any)["record_type"] == "edge" {
				d["evidence"] = append(append([]any{}, ev[:i]...), ev[i+1:]...)
				return true
			}
		}
		return false
	})
	add("evidence-duplicated", research.RuleEvidenceExtra, "an evidence record repeated", func(d map[string]any) bool {
		ev := d["evidence"].([]any)
		if len(ev) == 0 {
			return false
		}
		d["evidence"] = append(ev, ev[0])
		return true
	})
	add("targets-reordered", research.RuleTargetOrder, "results not in the request's target order", func(d map[string]any) bool {
		rs := results(d)
		if len(rs) < 2 {
			return false
		}
		rs[0], rs[1] = rs[1], rs[0]
		return true
	})
	add("root-basis", research.RuleRootBasis, "root_basis not matching the request's roots", func(d map[string]any) bool {
		r := results(d)[0].(map[string]any)
		if r["root_basis"] == "indexed_graph" {
			r["root_basis"] = "subset_on_indexed_graph"
		} else {
			r["root_basis"] = "indexed_graph"
		}
		return true
	})
}

func decode(raw []byte) map[string]any {
	var m map[string]any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	check(d.Decode(&m))
	return m
}

func generic(v any) any {
	raw, _ := json.Marshal(v)
	var g any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.Decode(&g)
	return g
}

func num(v any) int {
	n, _ := strconv.Atoi(fmt.Sprint(v))
	return n
}

func hasRule(vs []research.Violation, rule string) bool {
	for _, v := range vs {
		if v.Rule == rule {
			return true
		}
	}
	return false
}

func writeJSON(path string, v any) {
	raw, err := json.MarshalIndent(v, "", "  ")
	check(err)
	check(os.WriteFile(path, append(raw, '\n'), 0o644))
}

// writeSums writes SHA256SUMS over every file of the package (sorted
// paths, sha256sum format).
func writeSums(dir string) {
	var lines []string
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		check(err)
		if d.IsDir() || d.Name() == "SHA256SUMS" {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		lines = append(lines, fileSHA(p)+"  "+filepath.ToSlash(rel))
		return nil
	})
	sort.Slice(lines, func(i, j int) bool { return lines[i][66:] < lines[j][66:] })
	check(os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(strings.Join(lines, "\n")+"\n"), 0o644))
}

func fileSHA(p string) string {
	raw, err := os.ReadFile(p)
	check(err)
	s := sha256.Sum256(raw)
	return hex.EncodeToString(s[:])
}

func goVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		return bi.GoVersion
	}
	return ""
}

func run(bin string, args ...string) {
	cmd := exec.Command(bin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fail("%s %s: %v\n%s", bin, strings.Join(args, " "), err, stderr.String())
	}
}

func check(err error) {
	if err != nil {
		fail("%v", err)
	}
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "research-acceptance: "+format+"\n", a...)
	os.Exit(1)
}
