// Package reach assigns each CVE a reachability status (section 4/5) by
// running a single multi-source BFS from every application method over
// the CHA-expanded call graph, then checking whether each CVE's mapped
// vulnerable methods (internal/vulnmap) were reached.
package reach

import (
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/cves"
	"ariadne/internal/hierarchy"
	"ariadne/internal/vulnmap"
)

// Status is one of the four outcomes section 4 defines.
type Status string

const (
	StatusReachable    Status = "REACHABLE"
	StatusFramework    Status = "FRAMEWORK"
	StatusNotReachable Status = "NOT_REACHABLE"
	StatusUnknown      Status = "UNKNOWN"
)

// EntryEdge is the first app -> dependency edge on a REACHABLE/FRAMEWORK
// path: the one call where execution actually leaves the application's
// own code, which is usually the most useful single line for a human
// reviewing the result to look at.
type EntryEdge struct {
	From string `json:"from"` // app method key, "fqcn#name(desc)"
	To   string `json:"to"`   // dependency method key it calls
}

// Result is one CVE's reachability verdict.
type Result struct {
	CVE     string     `json:"cve"`
	Status  Status     `json:"status"`
	Methods []string   `json:"methods,omitempty"` // from vulnmap, carried through for the report
	Source  string     `json:"source,omitempty"`  // vulnmap source: "manual" or "diff"
	Path    []string   `json:"path,omitempty"`    // root -> ... -> vulnerable method, rendered node keys
	Entry   *EntryEdge `json:"entry,omitempty"`
}

// Engine holds everything needed to answer "is this CVE reachable" for
// every CVE in one run: the call graph, the class hierarchy (for roots,
// inherited-method matching, and the app/dependency distinction), and
// which dependency packages count as framework code.
type Engine struct {
	g         *callgraph.Graph
	h         *hierarchy.Hierarchy
	framework *FrameworkMatcher

	roots []uint32
	// appOnly is the REACHABLE-pass BFS: starts from every app method,
	// treating every framework-package node as a dead end so execution
	// can't be traced *through* one. full is the FRAMEWORK-pass BFS from
	// the same roots with no such restriction; computed once, up front,
	// since both are independent of which CVE is being checked.
	appOnly bfsResult
	full    bfsResult
}

// New builds an Engine and immediately runs both BFS passes, so
// constructing one over a very large graph does the expensive part once
// regardless of how many CVEs Analyze is later asked about.
func New(g *callgraph.Graph, h *hierarchy.Hierarchy, framework *FrameworkMatcher) *Engine {
	if framework == nil {
		framework = NewFrameworkMatcher(DefaultFrameworkPackages)
	}
	e := &Engine{g: g, h: h, framework: framework}
	e.roots = e.appRoots()

	isFramework := e.isFrameworkNode
	e.appOnly = bfs(g, e.roots, isFramework)
	e.full = bfs(g, e.roots, func(uint32) bool { return false })
	return e
}

// appRoots returns every method node belonging to an application class —
// the default root set section 3.4 describes. (An annotation-filtered
// "--roots annotated" mode is CLI/profile territory, not implemented
// here.)
func (e *Engine) appRoots() []uint32 {
	var roots []uint32
	for fqcn, ci := range e.h.All() {
		if !ci.IsApp {
			continue
		}
		for _, m := range ci.Methods {
			if id, ok := e.g.ID(fqcn, m.Name, m.Descriptor); ok {
				roots = append(roots, id)
			}
		}
	}
	return roots
}

func (e *Engine) isFrameworkNode(id uint32) bool {
	ci, ok := e.h.Lookup(nodeFQCN(e.g, id))
	if !ok {
		return false
	}
	return e.framework.Match(ci.GAV)
}

func (e *Engine) isAppNode(id uint32) bool {
	ci, ok := e.h.Lookup(nodeFQCN(e.g, id))
	return ok && ci.IsApp
}

func nodeFQCN(g *callgraph.Graph, id uint32) string {
	fqcn, _, _ := strings.Cut(g.Key(id), "#")
	return fqcn
}

// Analyze assigns a status to every CVE in list. A CVE vulnmap has no
// entry for is UNKNOWN; a CVE whose methods were never found, or found
// but unvisited by either BFS pass, is NOT_REACHABLE.
func (e *Engine) Analyze(list []cves.CVE, vm map[string]vulnmap.Entry) []Result {
	results := make([]Result, len(list))
	for i, cve := range list {
		results[i] = e.analyzeOne(cve, vm)
	}
	return results
}

func (e *Engine) analyzeOne(cve cves.CVE, vm map[string]vulnmap.Entry) Result {
	entry, ok := vm[cve.ID]
	if !ok {
		return Result{CVE: cve.ID, Status: StatusUnknown}
	}

	var matched []uint32
	for _, spec := range entry.Methods {
		matched = append(matched, matchVulnerableMethod(e.g, e.h, spec)...)
	}
	matched = dedupeIDs(matched)

	result := Result{CVE: cve.ID, Status: StatusNotReachable, Methods: entry.Methods, Source: entry.Source}

	// Prefer the shortest REACHABLE path among every matched node; fall
	// back to the shortest FRAMEWORK path only if no matched node is
	// REACHABLE at all.
	if _, path, ok := shortestOf(e.appOnly, matched); ok {
		result.Status = StatusReachable
		result.Path = renderPath(e.g, path)
		result.Entry = e.entryEdge(path)
		return result
	}
	if _, path, ok := shortestOf(e.full, matched); ok {
		result.Status = StatusFramework
		result.Path = renderPath(e.g, path)
		result.Entry = e.entryEdge(path)
		return result
	}
	return result
}

// shortestOf returns the shortest path, among every node in candidates
// that r actually visited, along with which candidate it ends at.
func shortestOf(r bfsResult, candidates []uint32) (target uint32, path []uint32, ok bool) {
	bestLen := -1
	for _, c := range candidates {
		p, visited := r.path(c)
		if !visited {
			continue
		}
		if bestLen == -1 || len(p) < bestLen {
			bestLen = len(p)
			target = c
			path = p
			ok = true
		}
	}
	return target, path, ok
}

func renderPath(g *callgraph.Graph, path []uint32) []string {
	out := make([]string, len(path))
	for i, id := range path {
		out[i] = g.Key(id)
	}
	return out
}

// entryEdge finds the first app -> non-app edge on path: the call where
// execution actually leaves the application's own code.
func (e *Engine) entryEdge(path []uint32) *EntryEdge {
	for i := 0; i+1 < len(path); i++ {
		if e.isAppNode(path[i]) && !e.isAppNode(path[i+1]) {
			return &EntryEdge{From: e.g.Key(path[i]), To: e.g.Key(path[i+1])}
		}
	}
	return nil
}
