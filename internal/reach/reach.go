// Package reach assigns each CVE a reachability status (section 4/5) by
// running a single multi-source BFS from every application method over
// the CHA-expanded call graph, then checking whether each CVE's mapped
// vulnerable methods (internal/vulnmap) were reached.
package reach

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/findings"
	"ariadne/internal/hierarchy"
	"ariadne/internal/mavenversion"
	"ariadne/internal/relocate"
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
	Source  string   `json:"source,omitempty"`  // vulnmap source: "manual", "diff", "release-diff" or "llm"
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
	// PathEdges and Character (ROADMAP B0/B1): the evidence of each step
	// of Path and the set of its characters. Not part of v1 JSON; v2
	// exports them per finding.
	PathEdges []PathEdge `json:"-"`
	Character []string   `json:"-"`
}

// InstanceFinding is one advisory's result for one affected component
// instance (docs/RESULTS_CONTRACT.MD §5).
type InstanceFinding struct {
	Instance  string // component instance ID
	GAV       string
	Status    Status
	Reasons   []findings.ReasonCode
	Path      []string
	Entry     *EntryEdge
	Note      string
	ModelGaps map[string]int // as Result.ModelGaps, for this instance
	// MatchPrecision is, for a positive finding, the precision of the
	// selector that matched the path's target (vulnmap.Precision*): a
	// path to a name-only match is weaker evidence than an exact one.
	MatchPrecision string
	// PathEdges and Character: as Result's, for this instance's path.
	PathEdges []PathEdge
	Character []string
}

// EngineVersion versions the reachability semantics (statuses, per-instance
// findings, aggregation, gap rule). Bump it on any change that can change
// a result for the same graph and mapping; snapshots record it (CVE:F3).
const EngineVersion = "2" // 2: CVE:F4 descriptor selectors, variant selection

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
		e.fillFindings(&results[i], cve)
		e.notePrecision(&results[i], vm[cve.ID])
		r := &results[i]
		r.PathEdges, r.Character = pathEvidence(e.g, r.Path)
		for j := range r.Findings {
			f := &r.Findings[j]
			f.PathEdges, f.Character = pathEvidence(e.g, f.Path)
		}
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
		return Result{CVE: cve.ID, Status: StatusUnknown, Reasons: []findings.ReasonCode{findings.MappingMissing}}
	}
	if !entry.Mapped() {
		// Fix-commit references existed but none could be read (CVE:F5).
		return Result{CVE: cve.ID, Status: StatusUnknown, Reasons: []findings.ReasonCode{findings.SourceUnavailable},
			Note: "mapping source unavailable: " + entry.SourceError}
	}

	entry = entry.WithDefaults()
	matched := e.matchSpecs(entry.Methods, e.affectedInstances(cve))

	result := Result{
		CVE: cve.ID, Status: StatusNotReachable, Methods: entry.Methods, Source: entry.Source,
		Model: entry.Model, Evidence: entry.Evidence,
	}
	var reasons []findings.ReasonCode
	if entry.Role == vulnmap.RoleEntrypoint {
		reasons = append(reasons, findings.MappingEntrypointOnly)
	}
	if entry.Source == vulnmap.SourceDiff && entry.NewSideOnly() {
		reasons = append(reasons, findings.MappingNewSideOnly)
	}

	if instances := e.affectedInstances(cve); len(instances) > 0 && len(entry.Variants) > 0 {
		e.analyzeVariants(&result, entry, instances)
	} else if len(instances) > 0 {
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
		for i := range result.Findings {
			if f := &result.Findings[i]; f.Status == StatusNotReachable {
				f.Status, f.Note = StatusUnknown, result.Note
			}
		}
	case result.Status == StatusNotReachable && entry.Coverage != vulnmap.CoverageComplete:
		// Option B of CVE:F1b: v1 keeps NOT_REACHABLE for a partial
		// mapping but says so; v2 (F1c) carries the coverage itself.
		reasons = append(reasons, findings.MappingPartial)
	}
	result.Reasons = orderedReasons(reasons)
	return result
}

// matchSpecs matches every selector in the graph, and for an affected
// instance whose classes were relocated (CVE:F7a) also the selector
// rewritten by that instance's relocations — the original component's
// method under the name it has in the artifact.
func (e *Engine) matchSpecs(specs []string, instances []component.Instance) []uint32 {
	var matched []uint32
	for _, spec := range specs {
		matched = append(matched, matchVulnerableMethod(e.g, e.h, spec)...)
	}
	for _, in := range instances {
		if len(in.Relocations) == 0 {
			continue
		}
		rs, err := relocate.New(in.Relocations)
		if err != nil {
			continue
		}
		for _, spec := range specs {
			if r := rs.Selector(spec); r != spec {
				matched = append(matched, matchVulnerableMethod(e.g, e.h, r)...)
			}
		}
	}
	return dedupeIDs(matched)
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
		if in.GAV == nil {
			continue
		}
		if want[in.GAV.PackageName()+"@"+in.GAV.Version] {
			out = append(out, in)
			continue
		}
		// CVE:F6b: a version-candidates instance is affected when any of
		// its candidate versions is (its finding says so).
		for _, v := range in.VersionCandidates {
			if want[in.GAV.PackageName()+"@"+v] {
				out = append(out, in)
				break
			}
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

	// related are allowed nodes outside every affected instance: code of
	// a sibling artifact the advisory's fix touched.
	var related []uint32
	for _, id := range allowed {
		if !affected[e.nodeInstance(id)] {
			related = append(related, id)
		}
	}
	const relatedNote = "mapped methods are in a related artifact, not in this instance"

	for _, in := range instances {
		f := InstanceFinding{Instance: in.ID, GAV: in.Label()}
		var r Result
		switch own := byInstance[in.ID]; {
		case len(own) > 0:
			e.classify(&r, own)
			f.Status, f.Path, f.Entry = r.Status, r.Path, r.Entry
			if r.Status != StatusReachable && len(related) > 0 {
				// Related-artifact evidence counts for every instance
				// (as for the advisory row); own code wins a tie.
				var rr Result
				e.classify(&rr, related)
				if statusRank(rr.Status) > statusRank(r.Status) {
					f.Status, f.Path, f.Entry, f.Note = rr.Status, rr.Path, rr.Entry, relatedNote
				}
			}
		case shadowedOwn[in.ID]:
			// Its own copy of a mapped class is not in the graph; another
			// affected jar's copy is not this instance's code. Only
			// related-artifact code can still give a path.
			f.Status, f.Reasons = StatusUnknown, []findings.ReasonCode{findings.ClasspathConflict}
			if len(related) > 0 {
				e.classify(&r, related)
				if r.Status != StatusNotReachable {
					f.Status, f.Path, f.Entry, f.Note = r.Status, r.Path, r.Entry, relatedNote
				}
			}
		case len(allowed) > 0:
			e.classify(&r, allowed)
			f.Status, f.Path, f.Entry = r.Status, r.Path, r.Entry
			f.Note = relatedNote
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
	markCandidates(result.Findings, instances)

	// Aggregate (contract §5) from the findings, so v1 is an export of
	// them. The row's path stays the shortest over every allowed node
	// (as before F1b); its status equals the best finding's by
	// construction of the cases above.
	result.Status, result.Reasons = Aggregate(result.Findings)
	switch result.Status {
	case StatusReachable, StatusFramework:
		var r Result
		e.classify(&r, allowed)
		result.Path, result.Entry = r.Path, r.Entry
	case StatusNotReachable:
		e.annotateGaps(&result.Reasons, &result.ModelGaps, affected)
	}
}

// markCandidates prefixes VERSION_CANDIDATES (CVE:F6b) to the findings of
// version-candidates instances, once: the advisory affects at least one
// candidate version, not necessarily the one present.
func markCandidates(fs []InstanceFinding, instances []component.Instance) {
	candidates := map[string]bool{}
	for _, in := range instances {
		if len(in.VersionCandidates) > 0 {
			candidates[in.ID] = true
		}
	}
	for i := range fs {
		f := &fs[i]
		if !candidates[f.Instance] || (len(f.Reasons) > 0 && f.Reasons[0] == findings.VersionCandidates) {
			continue
		}
		f.Reasons = append([]findings.ReasonCode{findings.VersionCandidates}, f.Reasons...)
	}
}

// analyzeVariants handles a versioned mapping (CVE:F4): every instance
// gets the method list of the one variant whose package and version range
// cover it. An irregular version is VERSION_UNKNOWN, no covering variant
// MAPPING_VERSION_UNCOVERED, several with different methods
// MAPPING_CONFLICT — never a guess. Instances that select the same list
// are analyzed together exactly like an unversioned mapping; the row is
// the aggregate of all findings, its path the shortest among the findings
// that carry the aggregate status.
func (e *Engine) analyzeVariants(result *Result, entry vulnmap.Entry, instances []component.Instance) {
	byInstance := map[string]InstanceFinding{}
	groups := map[string][]component.Instance{}
	methodsOf := map[string][]string{}
	var order []string
	for _, in := range instances {
		pkg := in.GAV.PackageName()
		versions := []string{in.GAV.Version}
		if len(in.VersionCandidates) > 0 {
			versions = in.VersionCandidates // CVE:F6b
		}
		f := InstanceFinding{Instance: in.ID, GAV: in.Label(), Status: StatusUnknown}
		irregular := ""
		for _, v := range versions {
			if !mavenversion.Regular(v) {
				irregular = v
			}
		}
		if irregular != "" {
			f.Reasons = []findings.ReasonCode{findings.VersionUnknown}
			f.Note = fmt.Sprintf("version %q cannot be placed in the mapping's version ranges", irregular)
			byInstance[in.ID] = f
			continue
		}
		// keysOf[v]: the distinct method lists of the variants covering v.
		var chosenAll []vulnmap.Variant
		keysOf := map[string]map[string]bool{}
		for _, version := range versions {
			keysOf[version] = map[string]bool{}
			for _, v := range entry.Variants {
				if (v.Package == "" || v.Package == pkg) && v.Range.Contains(version) == mavenversion.In {
					chosenAll = append(chosenAll, v)
					ms := append([]string(nil), v.Methods...)
					sort.Strings(ms)
					keysOf[version][strings.Join(ms, "\n")] = true
				}
			}
		}
		keys := map[string]bool{}
		uncovered, conflict := 0, false
		for _, version := range versions {
			if len(keysOf[version]) == 0 {
				uncovered++
			}
			if len(keysOf[version]) > 1 {
				conflict = true
			}
			for k := range keysOf[version] {
				keys[k] = true
			}
		}
		version := strings.Join(versions, ", ")
		switch {
		case uncovered == len(versions):
			f.Reasons = []findings.ReasonCode{findings.MappingVersionUncovered}
			f.Note = fmt.Sprintf("no variant of the mapping covers %s %s", pkg, version)
			byInstance[in.ID] = f
		case conflict:
			var ranges []string
			for _, v := range chosenAll {
				ranges = append(ranges, v.Versions)
			}
			f.Reasons = []findings.ReasonCode{findings.MappingConflict}
			f.Note = fmt.Sprintf("variants %s all cover %s with different methods", strings.Join(ranges, ", "), version)
			byInstance[in.ID] = f
		case uncovered > 0 || len(keys) > 1:
			f.Reasons = []findings.ReasonCode{findings.VersionUnknown}
			f.Note = fmt.Sprintf("the candidate versions %s select different variants of the mapping", version)
			byInstance[in.ID] = f
		default:
			var key string
			for k := range keys {
				key = k
			}
			if _, seen := groups[key]; !seen {
				order = append(order, key)
				methodsOf[key] = chosenAll[0].Methods
			}
			groups[key] = append(groups[key], in)
		}
	}

	var parts []Result
	for _, key := range order {
		matched := e.matchSpecs(methodsOf[key], groups[key])
		var r Result
		e.analyzePerInstance(&r, entry, matched, groups[key])
		for _, f := range r.Findings {
			byInstance[f.Instance] = f
		}
		parts = append(parts, r)
	}
	for _, in := range instances {
		result.Findings = append(result.Findings, byInstance[in.ID])
	}
	markCandidates(result.Findings, instances)
	result.Status, result.Reasons = Aggregate(result.Findings)
	switch result.Status {
	case StatusReachable, StatusFramework:
		for _, r := range parts {
			if r.Status == result.Status && (result.Path == nil || len(r.Path) < len(result.Path)) {
				result.Path, result.Entry = r.Path, r.Entry
			}
		}
	case StatusNotReachable:
		affected := map[string]bool{}
		for _, in := range instances {
			affected[in.ID] = true
		}
		e.annotateGaps(&result.Reasons, &result.ModelGaps, affected)
	}
}

// statusRank orders verdicts for "a positive finding wins".
func statusRank(s Status) int {
	switch s {
	case StatusReachable:
		return 3
	case StatusFramework:
		return 2
	}
	return 0
}

// Aggregate is the v1 export rule of docs/RESULTS_CONTRACT.MD §5 over one
// advisory's per-instance findings: the best positive status wins
// (REACHABLE over FRAMEWORK) with no reasons; otherwise any UNKNOWN
// finding makes it UNKNOWN with the union of their reasons; NOT_APPLICABLE
// only when every finding is; else NOT_REACHABLE.
func Aggregate(fs []InstanceFinding) (Status, []findings.ReasonCode) {
	best := StatusNotReachable
	var unresolved []findings.ReasonCode
	unknown, notApplicable := false, len(fs) > 0
	for _, f := range fs {
		if statusRank(f.Status) > statusRank(best) {
			best = f.Status
		}
		if f.Status == StatusUnknown {
			unknown = true
			unresolved = append(unresolved, f.Reasons...)
		}
		notApplicable = notApplicable && f.Status == StatusNotApplicable
	}
	switch {
	case best != StatusNotReachable:
		return best, nil
	case unknown:
		return StatusUnknown, orderedReasons(unresolved)
	case notApplicable:
		return StatusNotApplicable, nil
	}
	return StatusNotReachable, nil
}

// GapSite is one unresolved dynamic site of the run (reflection,
// MethodHandle, callback registration, ServiceLoader), for the model-gap
// rule of docs/RESULTS_CONTRACT.MD §4.
type GapSite struct {
	Kind                               string // "reflection", "methodhandle", "callback", "serviceloader", "deserialization"
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

// fillFindings gives a result decided without per-instance matching
// (WebJar, no mapping, no affected instance in the inventory) one finding
// per affected instance carrying the row's verdict, so every advisory ×
// instance has a finding. Mapping-axis codes stay on the row (v2 keeps
// mapping on its own axis).
func (e *Engine) fillFindings(r *Result, cve cves.CVE) {
	if len(r.Findings) > 0 {
		return
	}
	var reasons []findings.ReasonCode
	for _, c := range r.Reasons {
		if !findings.MappingQualifier(c) {
			reasons = append(reasons, c)
		}
	}
	for _, in := range e.affectedInstances(cve) {
		r.Findings = append(r.Findings, InstanceFinding{
			Instance: in.ID, GAV: in.GAV.String(), Status: r.Status, Reasons: reasons,
			Path: r.Path, Entry: r.Entry, Note: r.Note, ModelGaps: r.ModelGaps,
		})
	}
}

// notePrecision sets MatchPrecision on every positive finding: the best
// precision among the mapping's selectors that match the path's target.
func (e *Engine) notePrecision(r *Result, entry vulnmap.Entry) {
	rank := map[string]int{vulnmap.PrecisionName: 1, vulnmap.PrecisionParams: 2, vulnmap.PrecisionExact: 3}
	for i := range r.Findings {
		f := &r.Findings[i]
		if len(f.Path) == 0 || (f.Status != StatusReachable && f.Status != StatusFramework) {
			continue
		}
		fqcn, rest, _ := strings.Cut(f.Path[len(f.Path)-1], "#")
		name, desc := rest, ""
		if j := strings.IndexByte(rest, '('); j >= 0 {
			name, desc = rest[:j], rest[j:]
		}
		target, ok := e.g.ID(fqcn, name, desc)
		if !ok {
			continue
		}
		best := ""
		for _, spec := range entry.Methods {
			sel, err := vulnmap.ParseSelector(spec)
			if err != nil || rank[sel.Precision()] <= rank[best] {
				continue
			}
			for _, id := range matchVulnerableMethod(e.g, e.h, spec) {
				if id == target {
					best = sel.Precision()
					break
				}
			}
		}
		f.MatchPrecision = best
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
