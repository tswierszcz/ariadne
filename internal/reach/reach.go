// Package reach assigns each CVE a reachability status (section 4/5) by
// running a single multi-source BFS from every application method over
// the CHA-expanded call graph, then checking whether each CVE's mapped
// vulnerable methods (internal/vulnmap) were reached.
package reach

import (
	"fmt"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/findings"
	"ariadne/internal/hierarchy"
	"ariadne/internal/vulnmap"
)

// Status is one of the outcomes section 5 of PLAN.md defines.
type Status string

const (
	StatusReachable    Status = "REACHABLE"
	StatusFramework    Status = "FRAMEWORK"
	StatusNotReachable Status = "NOT_REACHABLE"
	StatusUnknown      Status = "UNKNOWN"
	// StatusNotApplicable: every affected dependency is a WebJar (client-
	// side JavaScript/CSS, no JVM bytecode), so a call graph cannot answer
	// the question at all. It is not a verdict that the app is safe.
	StatusNotApplicable Status = "NOT_APPLICABLE"
)

// EntryEdge is the first app -> dependency edge on a REACHABLE/FRAMEWORK
// path: the one call where execution actually leaves the application's
// own code, which is usually the most useful single line for a human
// reviewing the result to look at. Kind/Basis/PC are JVM_GRAPH G0's
// "resolution rule" and "concrete call site" for that one edge: Basis
// says whether this is a direct call or one CHA fabricated to a subtype
// (a reader should trust a "direct" entry a lot more than a "cha" one),
// and PC is the bytecode offset of the call site that produced it (the
// first one, if several call sites collapsed into this edge) — enough to
// find the exact instruction with `javap -c -p`.
type EntryEdge struct {
	From  string `json:"from"`            // app method key, "fqcn#name(desc)"
	To    string `json:"to"`              // dependency method key it calls
	Kind  string `json:"kind,omitempty"`  // classfile.CallKind: virtual, interface, static, special, dynamic, synthetic
	Basis string `json:"basis,omitempty"` // callgraph.EdgeBasis: direct, cha, synthetic
	PC    *int   `json:"pc,omitempty"`    // bytecode offset of the call site, nil for a synthetic edge with none
}

// PCString renders PC for the report template, which can't cleanly
// dereference a *int itself: "" when there is no call site (a synthetic
// edge), otherwise "pc=<offset>".
func (e EntryEdge) PCString() string {
	if e.PC == nil {
		return ""
	}
	return fmt.Sprintf("pc=%d", *e.PC)
}

// Result is one CVE's reachability verdict.
type Result struct {
	CVE     string   `json:"cve"`
	Status  Status   `json:"status"`
	Methods []string `json:"methods,omitempty"` // from vulnmap, carried through for the report
	Source  string   `json:"source,omitempty"`  // vulnmap source: "manual", "diff" or "llm"
	// Model/Evidence are set for Source "llm": the model id, and whether
	// it saw a fix diff ("diff") or only the advisory text ("advisory").
	Model    string `json:"model,omitempty"`
	Evidence string `json:"evidence,omitempty"`
	// Note explains a verdict a reader might not expect, e.g. an LLM
	// advisory-only NOT_REACHABLE downgraded to UNKNOWN.
	Note  string     `json:"note,omitempty"`
	Path  []string   `json:"path,omitempty"` // root -> ... -> vulnerable method, rendered node keys
	Entry *EntryEdge `json:"entry,omitempty"`

	// Exploitability/ExploitReason/ExploitModel are set only when
	// -exploit-check asked an LLM about this REACHABLE CVE's path
	// (internal/exploit); this package never sets them itself.
	Exploitability string `json:"exploitability,omitempty"` // "likely", "unlikely" or "unknown"
	ExploitReason  string `json:"exploit_reason,omitempty"`
	ExploitModel   string `json:"exploit_model,omitempty"`

	// PathPlausibility/PathPlausibilityReason/PathPlausibilityModel are set
	// only when -framework-check asked an LLM about this FRAMEWORK CVE's
	// path (internal/exploit.FrameworkChecker); this package never sets
	// them itself.
	PathPlausibility       string `json:"path_plausibility,omitempty"` // "plausible", "artifact" or "unknown"
	PathPlausibilityReason string `json:"path_plausibility_reason,omitempty"`
	PathPlausibilityModel  string `json:"path_plausibility_model,omitempty"`

	// Reasons are stable reason codes (internal/findings, CVE:F1b) that
	// qualify Status; additive to the v1 schema.
	Reasons []findings.ReasonCode `json:"reasons,omitempty"`
	// ModelGaps counts, by kind, the reachable unresolved sites relevant
	// to a NOT_REACHABLE result (contract §4, report-only: the verdict is
	// kept, Reasons carries MODEL_INCOMPLETE); additive to v1.
	ModelGaps map[string]int `json:"model_gaps,omitempty"`
	// Findings are this advisory's per-component-instance results
	// (advisory × instance), from which Status is aggregated. Not part of
	// v1 JSON; the report shows them, v2 (CVE:F1c) will export them.
	Findings []InstanceFinding `json:"-"`
}

// InstanceFinding is one advisory's result for one affected component
// instance (docs/RESULTS_CONTRACT.MD §5).
type InstanceFinding struct {
	Instance  string // component instance ID
	GAV       string
	Status    Status
	Reasons   []findings.ReasonCode
	Path      []string
	Note      string
	ModelGaps map[string]int // as Result.ModelGaps, for this instance
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

	// webAssetPkgs holds "group:artifact" of dependencies that are WebJars
	// (see input.DepJar.WebAssetsOnly); set with SetWebAssetPackages.
	webAssetPkgs map[string]bool

	// components, when set (SetComponents), makes findings per affected
	// component instance (CVE:F1b); nil keeps whole-graph matching.
	components *component.Inventory
	// gaps are the reachable unresolved sites (SetGaps, contract §4).
	gaps []gap
}

// SetComponents enables per-instance findings over inv.
func (e *Engine) SetComponents(inv *component.Inventory) { e.components = inv }

// SetWebAssetPackages marks dependencies ("group:artifact") that contain
// only client-side web assets. A CVE whose every affected package is one
// of them gets StatusNotApplicable.
func (e *Engine) SetWebAssetPackages(pkgs map[string]bool) {
	e.webAssetPkgs = pkgs
}

// onlyWebAssets reports whether cve affects WebJars and nothing else.
func (e *Engine) onlyWebAssets(cve cves.CVE) bool {
	if len(cve.Affected) == 0 {
		return false
	}
	for _, a := range cve.Affected {
		if !e.webAssetPkgs[a.Package] {
			return false
		}
	}
	return true
}

// New builds an Engine with the default root set (AppRoots: every
// application method) and immediately runs both BFS passes, so
// constructing one over a very large graph does the expensive part once
// regardless of how many CVEs Analyze is later asked about.
func New(g *callgraph.Graph, h *hierarchy.Hierarchy, framework *FrameworkMatcher) *Engine {
	return NewWithRoots(g, h, framework, AppRoots(g, h))
}

// NewWithRoots is New with an explicit root set instead of every
// application method — --roots profile (internal/profile.Profile.AppRoots
// computes the profile-filtered set a caller passes in here).
func NewWithRoots(g *callgraph.Graph, h *hierarchy.Hierarchy, framework *FrameworkMatcher, roots []uint32) *Engine {
	if framework == nil {
		framework = NewFrameworkMatcher(DefaultFrameworkPackages)
	}
	e := &Engine{g: g, h: h, framework: framework, roots: roots}

	isFramework := e.isFrameworkNode
	e.appOnly = bfs(g, e.roots, isFramework)
	e.full = bfs(g, e.roots, func(uint32) bool { return false })
	return e
}

// AppRoots returns every method node belonging to an application class —
// the default root set section 3.4 describes.
func AppRoots(g *callgraph.Graph, h *hierarchy.Hierarchy) []uint32 {
	var roots []uint32
	// SortedFQCNs, not h.All() directly: root order feeds the multi-source
	// BFS's queue, and a tie between two roots that reach the same node at
	// the same distance is broken by which one the BFS happened to
	// enqueue first — which must not depend on Go's randomized map
	// iteration order if results.json is to report the same path on every
	// run of the same analysis.
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		for _, m := range ci.Methods {
			if id, ok := g.ID(fqcn, m.Name, m.Descriptor); ok {
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
	if e.onlyWebAssets(cve) {
		// Checked first: whatever a mapping says, a WebJar has no bytecode
		// for the vulnerable code to live in.
		pkgs := make([]string, len(cve.Affected))
		for i, a := range cve.Affected {
			pkgs[i] = a.Package + " " + a.Version
		}
		return Result{
			CVE: cve.ID, Status: StatusNotApplicable,
			Methods: entry.Methods, Source: entry.Source, Model: entry.Model, Evidence: entry.Evidence,
			Note: fmt.Sprintf("%s contains only client-side web assets (WebJar: JavaScript/CSS), no JVM bytecode; "+
				"check whether the app's pages use the affected component", strings.Join(pkgs, ", ")),
		}
	}
	if !ok {
		return Result{CVE: cve.ID, Status: StatusUnknown}
	}

	entry = entry.WithDefaults()
	var matched []uint32
	for _, spec := range entry.Methods {
		matched = append(matched, matchVulnerableMethod(e.g, e.h, spec)...)
	}
	matched = dedupeIDs(matched)

	result := Result{
		CVE: cve.ID, Status: StatusNotReachable, Methods: entry.Methods, Source: entry.Source,
		Model: entry.Model, Evidence: entry.Evidence,
	}
	var reasons []findings.ReasonCode
	if entry.Role == vulnmap.RoleEntrypoint {
		reasons = append(reasons, findings.MappingEntrypointOnly)
	}

	if instances := e.affectedInstances(cve); len(instances) > 0 {
		e.analyzePerInstance(&result, entry, matched, instances)
	} else {
		e.classify(&result, matched)
		switch {
		case result.Status == StatusNotReachable && len(matched) == 0:
			result.Status = StatusUnknown
			result.Reasons = append(result.Reasons, findings.MethodNotFound)
		case result.Status == StatusNotReachable:
			// No instance to scope by: only an unbounded gap is relevant.
			e.annotateGaps(&result.Reasons, &result.ModelGaps, nil)
		}
	}
	reasons = append(reasons, result.Reasons...)

	switch {
	case result.Status == StatusNotReachable && entry.Source == vulnmap.SourceLLM && entry.Evidence == vulnmap.EvidenceAdvisory:
		// An LLM that saw only the advisory text may have named a method
		// that exists but is not the vulnerable one; bytecode grounding
		// cannot catch that. "Not reachable" on such weak evidence would
		// read as "safe", so report it as UNKNOWN.
		result.Status = StatusUnknown
		result.Note = fmt.Sprintf("LLM (%s) suggested %s from the advisory text only (no fix diff); not reachable, but unverified",
			entry.Model, strings.Join(entry.Methods, ", "))
	case result.Status == StatusNotReachable && entry.Coverage != vulnmap.CoverageComplete:
		// Option B of CVE:F1b: v1 keeps NOT_REACHABLE for a partial
		// mapping but says so; v2 (F1c) carries the coverage itself.
		reasons = append(reasons, findings.MappingPartial)
	}
	result.Reasons = orderedReasons(reasons)
	return result
}

// classify sets result's status and path from the matched nodes over the
// whole graph (no component scoping).
func (e *Engine) classify(result *Result, matched []uint32) {
	// Prefer the shortest REACHABLE path among every matched node; fall
	// back to the shortest FRAMEWORK path only if no matched node is
	// REACHABLE at all.
	if _, path, ok := shortestOf(e.appOnly, matched); ok {
		result.Status = StatusReachable
		result.Path = renderPath(e.g, path)
		result.Entry = e.entryEdge(path)
		return
	}
	if _, path, ok := shortestOf(e.full, matched); ok {
		result.Status = StatusFramework
		result.Path = renderPath(e.g, path)
		result.Entry = e.entryEdge(path)
		return
	}
	result.Status = StatusNotReachable
}

// affectedInstances returns the component instances whose identity is one
// of cve's affected package/version pairs, in inventory order.
func (e *Engine) affectedInstances(cve cves.CVE) []component.Instance {
	if e.components == nil {
		return nil
	}
	want := map[string]bool{}
	for _, a := range cve.Affected {
		want[a.Package+"@"+a.Version] = true
	}
	var out []component.Instance
	for _, in := range e.components.Instances {
		if in.GAV != nil && want[in.GAV.PackageName()+"@"+in.GAV.Version] {
			out = append(out, in)
		}
	}
	return out
}

// nodeInstance attributes a matched node to the component instance whose
// code it runs: the instance of the class that declares the method (an
// inherited method matched through a subtype runs the supertype's body).
func (e *Engine) nodeInstance(id uint32) string {
	fqcn, rest, _ := strings.Cut(e.g.Key(id), "#")
	name := rest
	if i := strings.IndexByte(rest, '('); i >= 0 {
		name = rest[:i]
	}
	if decl, ok := e.h.ResolveMethod(fqcn, name); ok {
		fqcn = decl
	}
	if ci, ok := e.h.Lookup(fqcn); ok {
		return e.components.InstanceForClass(ci)
	}
	return ""
}

// analyzePerInstance computes one finding per affected instance and
// aggregates them into result (contract §5). A matched node is excluded
// only when it is "borrowed": the selected copy of a class whose affected
// instance holds its own, non-selected copy (another jar won under the
// classpath order) — that path does not run the affected instance's
// code. Every other matched node counts, including code in a sibling
// artifact of the affected one (an advisory filed against one Maven
// package whose fix touched another artifact of the same project); such
// a finding is annotated. The aggregate is the shortest path over the
// allowed nodes, exactly as before CVE:F1b; with none allowed it is
// UNKNOWN with CLASSPATH_CONFLICT (only borrowed nodes) or
// METHOD_NOT_FOUND (none at all).
func (e *Engine) analyzePerInstance(result *Result, entry vulnmap.Entry, matched []uint32, instances []component.Instance) {
	affected := map[string]bool{}
	for _, in := range instances {
		affected[in.ID] = true
	}
	// FQCNs whose affected instance's own copy was shadowed by a copy
	// from outside the affected set.
	borrowedFQCN := map[string]bool{}
	for _, c := range e.components.Conflicts {
		if affected[c.ShadowedInstance] && !affected[c.KeptInstance] {
			borrowedFQCN[c.FQCN] = true
		}
	}
	var allowed, borrowed []uint32
	byInstance := map[string][]uint32{}
	for _, id := range matched {
		fqcn, _, _ := strings.Cut(e.g.Key(id), "#")
		inst := e.nodeInstance(id)
		if borrowedFQCN[fqcn] && !affected[inst] {
			borrowed = append(borrowed, id)
			continue
		}
		allowed = append(allowed, id)
		byInstance[inst] = append(byInstance[inst], id)
	}

	matchedFQCN := map[string]bool{}
	for _, id := range matched {
		fqcn, _, _ := strings.Cut(e.g.Key(id), "#")
		matchedFQCN[fqcn] = true
	}
	shadowedOwn := map[string]bool{} // instance whose own copy of a matched class lost
	for _, c := range e.components.Conflicts {
		if matchedFQCN[c.FQCN] {
			shadowedOwn[c.ShadowedInstance] = true
		}
	}

	for _, in := range instances {
		f := InstanceFinding{Instance: in.ID, GAV: in.GAV.String()}
		var r Result
		switch own := byInstance[in.ID]; {
		case len(own) > 0:
			e.classify(&r, own)
			f.Status, f.Path = r.Status, r.Path
		case shadowedOwn[in.ID]:
			// Its own copy of a mapped class is not in the graph; another
			// jar's copy (affected or not) is not this instance's code.
			f.Status, f.Reasons = StatusUnknown, []findings.ReasonCode{findings.ClasspathConflict}
		case len(allowed) > 0:
			e.classify(&r, allowed)
			f.Status, f.Path = r.Status, r.Path
			f.Note = "mapped methods are in a related artifact, not in this instance"
		case len(borrowed) > 0:
			f.Status, f.Reasons = StatusUnknown, []findings.ReasonCode{findings.ClasspathConflict}
		default:
			f.Status, f.Reasons = StatusUnknown, []findings.ReasonCode{findings.MethodNotFound}
		}
		result.Findings = append(result.Findings, f)
	}

	for i := range result.Findings {
		f := &result.Findings[i]
		if f.Status == StatusNotReachable {
			e.annotateGaps(&f.Reasons, &f.ModelGaps, map[string]bool{f.Instance: true})
		}
	}

	// Aggregate (contract §5): a positive path wins; otherwise any
	// unresolved instance makes the advisory UNKNOWN with the union of
	// the instances' reasons, never NOT_REACHABLE.
	e.classify(result, allowed)
	if result.Status != StatusNotReachable {
		return
	}
	var unresolved []findings.ReasonCode
	for _, f := range result.Findings {
		if f.Status == StatusUnknown {
			unresolved = append(unresolved, f.Reasons...)
		}
	}
	if len(unresolved) > 0 {
		result.Status, result.Reasons = StatusUnknown, orderedReasons(unresolved)
		return
	}
	e.annotateGaps(&result.Reasons, &result.ModelGaps, affected)
}

// GapSite is one unresolved dynamic site of the run (reflection,
// MethodHandle, callback registration, ServiceLoader), for the model-gap
// rule of docs/RESULTS_CONTRACT.MD §4.
type GapSite struct {
	Kind                               string // "reflection", "methodhandle", "callback", "serviceloader"
	CallerFQCN, CallerName, CallerDesc string
	// Owners is the bounded set of classes the site could dispatch into;
	// nil when the site has no bounded candidate set at all.
	Owners []string
}

// gap is a GapSite whose caller is reachable from the roots.
type gap struct {
	site      GapSite
	unbounded bool
	owners    map[string]bool // owner instance IDs, with components set
}

// SetGaps records the run's unresolved sites; only those whose calling
// method the roots reach (FRAMEWORK pass) can make a finding UNKNOWN.
// Call after SetComponents.
func (e *Engine) SetGaps(sites []GapSite) {
	e.gaps = nil
	for _, s := range sites {
		id, ok := e.g.ID(s.CallerFQCN, s.CallerName, s.CallerDesc)
		if !ok {
			continue
		}
		if _, visited := e.full.path(id); !visited {
			continue
		}
		gp := gap{site: s, unbounded: s.Owners == nil, owners: map[string]bool{}}
		if e.components != nil {
			for _, o := range s.Owners {
				if ci, ok := e.h.Lookup(o); ok {
					if inst := e.components.InstanceForClass(ci); inst != "" {
						gp.owners[inst] = true
					}
				}
			}
		}
		e.gaps = append(e.gaps, gp)
	}
}

// ReachableGaps returns how many recorded gap sites are reachable from
// the roots, and how many of those have no bounded candidate set.
func (e *Engine) ReachableGaps() (reachable, unbounded int) {
	for _, g := range e.gaps {
		if g.unbounded {
			unbounded++
		}
	}
	return len(e.gaps), unbounded
}

// relevantGaps counts, by kind, the reachable unresolved sites that could
// dispatch into one of the instances (contract §4): those with no bounded
// candidate set, and those with a candidate in one of the instances. With
// instances nil only unbounded sites count.
func (e *Engine) relevantGaps(instances map[string]bool) map[string]int {
	var out map[string]int
	for _, g := range e.gaps {
		relevant := g.unbounded
		for inst := range g.owners {
			relevant = relevant || instances[inst]
		}
		if relevant {
			if out == nil {
				out = map[string]int{}
			}
			out[g.site.Kind]++
		}
	}
	return out
}

// annotateGaps applies the report-only form of the gap rule (contract §4
// as amended by CVE:F1b): a NOT_REACHABLE finding with relevant gaps
// keeps its verdict but carries MODEL_INCOMPLETE and the gap counts.
func (e *Engine) annotateGaps(reasons *[]findings.ReasonCode, gaps *map[string]int, instances map[string]bool) {
	if g := e.relevantGaps(instances); g != nil {
		*reasons = append(*reasons, findings.ModelIncomplete)
		*gaps = g
	}
}

// orderedReasons de-duplicates codes into findings.All order.
func orderedReasons(in []findings.ReasonCode) []findings.ReasonCode {
	if len(in) == 0 {
		return nil
	}
	seen := map[findings.ReasonCode]bool{}
	for _, r := range in {
		seen[r] = true
	}
	var out []findings.ReasonCode
	for _, r := range findings.All {
		if seen[r] {
			out = append(out, r)
		}
	}
	return out
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
			entry := &EntryEdge{From: e.g.Key(path[i]), To: e.g.Key(path[i+1])}
			if edge, ok := e.g.EdgeBetween(path[i], path[i+1]); ok {
				entry.Kind = string(edge.Kind)
				entry.Basis = string(edge.Basis)
				if len(edge.Sites) > 0 {
					pc := edge.Sites[0].PC
					entry.PC = &pc
				}
			}
			return entry
		}
	}
	return nil
}
