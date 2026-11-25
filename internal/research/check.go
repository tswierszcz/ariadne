package research

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// RESEARCH_HARNESS RH-A0: the producer's verdict on a response. CheckResponse
// applies the rules a consumer holding the same bundle can apply itself
// (docs/research-interface/ACCEPTANCE.md lists them by name): exact shape,
// bindings, the state rules of every target, path continuity against the
// bundle, and that every record a response carries is the bundle's record
// unchanged. Replay is separate: it compares with what this producer
// answers now, which a consumer cannot do.

// Violation is one broken rule.
type Violation struct {
	Rule     string `json:"rule"`
	Location string `json:"location"`
	Detail   string `json:"detail"`
}

// Rule names (ACCEPTANCE.md §3).
const (
	RuleJSON             = "json"
	RuleShape            = "shape"
	RuleVersion          = "version"
	RuleBinding          = "binding"
	RuleErrorEnvelope    = "error-envelope"
	RulePage             = "page"
	RuleRecordType       = "record-type"
	RuleRecordOriginal   = "record-original"
	RuleNeighbor         = "neighbor"
	RuleTargetOrder      = "target-order"
	RuleResolution       = "resolution"
	RuleSearchState      = "search-state"
	RuleEntryCount       = "entry-count"
	RuleModel            = "model"
	RuleLimits           = "limits"
	RuleRootBasis        = "root-basis"
	RulePathRoot         = "path-root"
	RulePathContinuity   = "path-continuity"
	RulePathEnd          = "path-end"
	RulePathCharacter    = "path-character"
	RulePathPass         = "path-pass"
	RulePathDispatch     = "path-dispatch"
	RuleEvidenceOriginal = "evidence-original"
	RuleEvidenceMissing  = "evidence-missing"
	RuleEvidenceExtra    = "evidence-extra"
	RuleReplay           = "replay"
)

// Rules lists every rule CheckResponse and Replay can report.
var Rules = []string{RuleJSON, RuleShape, RuleVersion, RuleBinding, RuleErrorEnvelope, RulePage, RuleRecordType,
	RuleRecordOriginal, RuleNeighbor, RuleTargetOrder, RuleResolution, RuleSearchState, RuleEntryCount, RuleModel,
	RuleLimits, RuleRootBasis, RulePathRoot, RulePathContinuity, RulePathEnd, RulePathCharacter, RulePathPass,
	RulePathDispatch, RuleEvidenceOriginal, RuleEvidenceMissing, RuleEvidenceExtra, RuleReplay}

type checker struct {
	b      *Bundle
	req    *Request
	reqRaw []byte
	out    []Violation
	// originals: record type -> id -> the bundle record as a generic value
	originals map[string]map[string]any
}

func (c *checker) add(rule, loc, format string, a ...any) {
	c.out = append(c.out, Violation{Rule: rule, Location: loc, Detail: fmt.Sprintf(format, a...)})
}

// CheckResponse checks a response document against the request it answers
// and the bundle it names. b must be loaded with every dataset (Load). The
// request must be one ParseRequest accepts; a response to a request it
// rejects can only be an error envelope, checked for shape and binding.
func CheckResponse(b *Bundle, reqRaw, respRaw []byte) []Violation {
	c := &checker{b: b, reqRaw: reqRaw}
	if r, perr := ParseRequest(b, reqRaw); perr == nil {
		c.req = r
	}
	if err := checkDuplicates(respRaw); err != nil {
		c.add(RuleJSON, "/", "%v", err)
		return c.out
	}
	var gen any
	dec := json.NewDecoder(bytes.NewReader(respRaw))
	dec.UseNumber()
	if err := dec.Decode(&gen); err != nil {
		c.add(RuleJSON, "/", "%v", err)
		return c.out
	}
	top, ok := gen.(map[string]any)
	if !ok {
		c.add(RuleShape, "/", "the response is not a JSON object")
		return c.out
	}
	status, _ := top["status"].(string)
	op, _ := top["operation"].(string)
	switch {
	case status == "error":
		var r ErrorResponse
		if c.decodeShape(respRaw, gen, &r) {
			c.checkError(&r)
		}
	case status == "ok" && op == "reachability":
		var r ReachabilityResponse
		if c.decodeShape(respRaw, gen, &r) {
			c.checkEnvelope(r.SchemaVersion, r.ContractRevision, r.AnalysisID, r.RequestID, r.Operation, r.Provenance)
			c.checkReachability(&r)
		}
	case status == "ok":
		var r DiscoveryResponse
		if c.decodeShape(respRaw, gen, &r) {
			c.checkEnvelope(r.SchemaVersion, r.ContractRevision, r.AnalysisID, r.RequestID, r.Operation, r.Provenance)
			c.checkDiscovery(&r)
		}
	default:
		c.add(RuleShape, "/status", "status %q is neither ok nor error", status)
	}
	return c.out
}

// decodeShape decodes raw into v and requires the document to have exactly
// the members of v's type at every object level (no unknown, no missing;
// null where the contract writes null). Records inside records/evidence
// are compared whole later.
func (c *checker) decodeShape(raw []byte, gen any, v any) bool {
	if err := json.Unmarshal(raw, v); err != nil {
		c.add(RuleShape, "/", "%v", err)
		return false
	}
	back, _ := json.Marshal(v)
	var want any
	d := json.NewDecoder(bytes.NewReader(back))
	d.UseNumber()
	d.Decode(&want)
	before := len(c.out)
	c.shapeDiff(gen, want, "")
	return len(c.out) == before
}

func (c *checker) shapeDiff(got, want any, loc string) {
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			c.add(RuleShape, pointerOrRoot(loc), "expected an object")
			return
		}
		for k := range g {
			if _, ok := w[k]; !ok {
				c.add(RuleShape, loc+"/"+k, "unknown member")
			}
		}
		for _, k := range sortedKeys(w) {
			gv, ok := g[k]
			if !ok {
				c.add(RuleShape, loc+"/"+k, "missing member")
				continue
			}
			if k == "records" || k == "evidence" {
				continue // generic records, checked against the bundle
			}
			c.shapeDiff(gv, w[k], loc+"/"+k)
		}
	case []any:
		g, ok := got.([]any)
		if !ok {
			c.add(RuleShape, pointerOrRoot(loc), "expected an array")
			return
		}
		for i := range w {
			if i < len(g) {
				c.shapeDiff(g[i], w[i], fmt.Sprintf("%s/%d", loc, i))
			}
		}
	case nil:
		if got != nil {
			c.add(RuleShape, pointerOrRoot(loc), "expected null")
		}
	}
}

func sortedKeys(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func (c *checker) checkEnvelope(schema, revision, analysisID, requestID, op string, prov Provenance) {
	if schema != ResponseSchema {
		c.add(RuleVersion, "/schema_version", "%q, want %q", schema, ResponseSchema)
	}
	if revision != ContractRevision {
		c.add(RuleVersion, "/contract_revision", "%q, want %q", revision, ContractRevision)
	}
	if analysisID != c.b.AnalysisID {
		c.add(RuleBinding, "/analysis_id", "%s is not the bundle's analysis %s", analysisID, c.b.AnalysisID)
	}
	if !reflect.DeepEqual(prov, c.b.Identity.Provenance) {
		c.add(RuleBinding, "/provenance", "differs from the bundle identity's provenance")
	}
	if c.req == nil {
		c.add(RuleBinding, "/", "an ok response to a request the producer rejects")
		return
	}
	if requestID != c.req.RequestID {
		c.add(RuleBinding, "/request_id", "%q, the request says %q", requestID, c.req.RequestID)
	}
	if op != c.req.Operation {
		c.add(RuleBinding, "/operation", "%q, the request says %q", op, c.req.Operation)
	}
	if c.req.AnalysisID != c.b.AnalysisID {
		c.add(RuleBinding, "/analysis_id", "the request names analysis %s", c.req.AnalysisID)
	}
}

func (c *checker) checkError(r *ErrorResponse) {
	if r.SchemaVersion != ResponseSchema || r.ContractRevision != ContractRevision {
		c.add(RuleVersion, "/", "schema %q revision %q", r.SchemaVersion, r.ContractRevision)
	}
	if r.Error.Code == "" || r.Error.Message == "" {
		c.add(RuleErrorEnvelope, "/error", "code and message are required")
	}
	// A rejected request is not parsed; its request_id, when it has a
	// string one, is still what the envelope must echo.
	var head struct {
		RequestID any `json:"request_id"`
	}
	json.Unmarshal(c.reqRaw, &head)
	if id, ok := head.RequestID.(string); ok && r.RequestID != nil && *r.RequestID != id {
		c.add(RuleBinding, "/request_id", "%q, the request says %q", *r.RequestID, id)
	}
}

// original returns the bundle's record of type typ and id, generically.
func (c *checker) original(typ, id string) (any, bool) {
	if c.originals == nil {
		c.originals = map[string]map[string]any{}
		put := func(t, id string, r any) {
			if c.originals[t] == nil {
				c.originals[t] = map[string]any{}
			}
			c.originals[t][id] = r
		}
		for i := range c.b.Components {
			put(TypeComponent, c.b.Components[i].InstanceID, &c.b.Components[i])
		}
		for i := range c.b.Classes {
			put(TypeClass, c.b.Classes[i].ClassID, &c.b.Classes[i])
		}
		for i := range c.b.Nodes {
			put(TypeNode, c.b.Nodes[i].NodeID, &c.b.Nodes[i])
		}
		for i := range c.b.Edges {
			put(TypeEdge, c.b.Edges[i].EdgeID, &c.b.Edges[i])
		}
		for i := range c.b.Roots {
			put(TypeRoot, c.b.Roots[i].RootID, &c.b.Roots[i])
		}
		for i := range c.b.Gaps {
			put(TypeGap, c.b.Gaps[i].GapID, &c.b.Gaps[i])
		}
		for i := range c.b.Calls {
			put(TypeCall, c.b.Calls[i].CallID, &c.b.Calls[i])
		}
	}
	r, ok := c.originals[typ][id]
	if !ok {
		return nil, false
	}
	if g, isGen := r.(genRecord); isGen {
		return g.v, true
	}
	raw, _ := json.Marshal(r)
	var g any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.Decode(&g)
	c.originals[typ][id] = genRecord{g}
	return g, true
}

type genRecord struct{ v any }

var idField = map[string]string{TypeComponent: "instance_id", TypeClass: "class_id", TypeNode: "node_id",
	TypeEdge: "edge_id", TypeRoot: "root_id", TypeGap: "gap_id", TypeCall: "call_id"}

// recordKey names a generic record; ok is false for a record without a
// known type or id.
func recordKey(r any) (typ, id string, ok bool) {
	m, isMap := r.(map[string]any)
	if !isMap {
		return "", "", false
	}
	typ, _ = m["record_type"].(string)
	f, known := idField[typ]
	if !known {
		return typ, "", false
	}
	id, _ = m[f].(string)
	return typ, id, id != ""
}

// checkOriginal requires r to be the bundle's record, member for member.
func (c *checker) checkOriginal(rule string, r any, loc string) (typ, id string, ok bool) {
	typ, id, ok = recordKey(r)
	if !ok {
		c.add(rule, loc, "not a record with a known record_type and id")
		return
	}
	orig, found := c.original(typ, id)
	if !found {
		c.add(rule, loc, "%s %s is not in the bundle", typ, id)
		return typ, id, false
	}
	gen := normalize(r)
	if !reflect.DeepEqual(gen, orig) {
		c.add(rule, loc, "%s %s differs from the bundle's record", typ, id)
		return typ, id, false
	}
	return typ, id, true
}

// normalize re-decodes a generic value with json.Number, as originals are.
func normalize(v any) any {
	raw, _ := json.Marshal(v)
	var g any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.Decode(&g)
	return g
}

var discoveryType = map[string]string{"entries": TypeRoot, "methods": TypeNode, "calls": TypeCall, "callers": TypeEdge, "callees": TypeEdge}

func (c *checker) checkDiscovery(r *DiscoveryResponse) {
	want, known := discoveryType[r.Operation]
	if !known {
		c.add(RuleShape, "/operation", "unknown operation %q", r.Operation)
		return
	}
	if r.Page.ReturnedCount != len(r.Records) {
		c.add(RulePage, "/page/returned_count", "%d, the page holds %d records", r.Page.ReturnedCount, len(r.Records))
	}
	if len(r.Records) > r.EffectiveLimits.MaxRecords {
		c.add(RuleLimits, "/records", "%d records over max_records %d", len(r.Records), r.EffectiveLimits.MaxRecords)
	}
	if !r.Page.Truncated && r.Page.NextCursor != nil {
		c.add(RulePage, "/page/next_cursor", "a cursor on an untruncated page")
	}
	if r.Page.Truncated && r.Page.NextCursor == nil {
		c.add(RulePage, "/page/next_cursor", "a truncated page without a cursor")
	}
	if r.Page.TotalCount != nil && *r.Page.TotalCount < r.Page.ReturnedCount {
		c.add(RulePage, "/page/total_count", "%d below the %d returned", *r.Page.TotalCount, r.Page.ReturnedCount)
	}
	if c.req != nil && c.req.Limits != nil {
		c.checkEffective(r.EffectiveLimits)
	}
	prev := ""
	for i, rec := range r.Records {
		loc := fmt.Sprintf("/records/%d", i)
		typ, id, ok := c.checkOriginal(RuleRecordOriginal, rec, loc)
		if typ != "" && typ != want {
			c.add(RuleRecordType, loc, "%s record in a %s response", typ, r.Operation)
			continue
		}
		if !ok {
			continue
		}
		if id <= prev {
			c.add(RulePage, loc, "records are not in ascending id order")
		}
		prev = id
		if (r.Operation == "callers" || r.Operation == "callees") && c.req != nil && c.req.NodeID != nil {
			e := c.b.Edges[c.b.idx.edgePos[id]]
			if r.Operation == "callers" && e.To != *c.req.NodeID || r.Operation == "callees" && e.From != *c.req.NodeID {
				c.add(RuleNeighbor, loc, "edge %s does not %s node %s", id, map[string]string{"callers": "enter", "callees": "leave"}[r.Operation], *c.req.NodeID)
			}
		}
	}
}

// checkEffective: a limit the request set is the limit the response applied.
func (c *checker) checkEffective(e EffectiveLimits) {
	l := c.req.Limits
	for _, p := range []struct {
		name string
		req  *int
		got  int
	}{{"max_records", l.MaxRecords, e.MaxRecords}, {"max_entries", l.MaxEntries, e.MaxEntries},
		{"max_path_nodes", l.MaxPathNodes, e.MaxPathNodes}, {"max_work_units", l.MaxWorkUnits, e.MaxWorkUnits}} {
		if p.req != nil && *p.req != p.got {
			c.add(RuleLimits, "/effective_limits/"+p.name, "%d, the request asked for %d", p.got, *p.req)
		}
	}
}

func (c *checker) checkReachability(r *ReachabilityResponse) {
	if c.req == nil {
		return
	}
	if c.req.Limits != nil {
		c.checkEffective(r.EffectiveLimits)
	}
	lim := r.EffectiveLimits
	// Evidence: every record the bundle's, once.
	type key struct{ typ, id string }
	ev := map[key]bool{}
	for i, rec := range r.Evidence {
		loc := fmt.Sprintf("/evidence/%d", i)
		typ, id, ok := c.checkOriginal(RuleEvidenceOriginal, rec, loc)
		if !ok {
			continue
		}
		if ev[key{typ, id}] {
			c.add(RuleEvidenceExtra, loc, "%s %s appears twice", typ, id)
		}
		ev[key{typ, id}] = true
	}
	need := func(typ, id, loc string) {
		if !ev[key{typ, id}] {
			c.add(RuleEvidenceMissing, loc, "%s %s is not in the evidence", typ, id)
		}
	}
	needNode := func(id, loc string) {
		need(TypeNode, id, loc)
		p, ok := c.b.idx.nodePos[id]
		if !ok {
			return
		}
		n := c.b.Nodes[p]
		if n.Method == nil || n.Method.InstanceID == nil {
			return
		}
		if _, ok := c.b.idx.compByID[*n.Method.InstanceID]; ok {
			need(TypeComponent, *n.Method.InstanceID, loc)
		}
		for _, ci := range c.b.idx.classByName[n.Method.Class] {
			if cl := c.b.Classes[ci]; cl.InstanceID == *n.Method.InstanceID {
				need(TypeClass, cl.ClassID, loc)
			}
		}
	}
	if len(r.Results) != len(c.req.Targets) {
		c.add(RuleTargetOrder, "/results", "%d results for %d targets", len(r.Results), len(c.req.Targets))
	}
	rootSet := map[string]bool{}
	for _, p := range c.b.idx.rootNodePos {
		rootSet[c.b.Nodes[p].NodeID] = true
	}
	subset := map[string]bool{}
	wantBasis := "indexed_graph"
	if c.req.Roots != nil && c.req.Roots.Selection == "subset" {
		wantBasis = "subset_on_indexed_graph"
		for _, id := range c.req.Roots.NodeIDs {
			subset[id] = true
		}
	}
	fw := c.b.frameworkNodes()
	work := 0
	for i, res := range r.Results {
		loc := fmt.Sprintf("/results/%d", i)
		if i < len(c.req.Targets) && res.TargetID != c.req.Targets[i].ID {
			c.add(RuleTargetOrder, loc+"/target_id", "%q in place of target %q", res.TargetID, c.req.Targets[i].ID)
		}
		if res.RootBasis != wantBasis {
			c.add(RuleRootBasis, loc+"/root_basis", "%q, the request's roots make it %q", res.RootBasis, wantBasis)
		}
		work += res.Search.WorkUnits
		// Resolution: the bundle decides it.
		matchedSet := map[string]bool{}
		for _, id := range res.MatchedNodeIDs {
			matchedSet[id] = true
			needNode(id, loc+"/matched_node_ids")
		}
		if i < len(c.req.Targets) {
			wantRes, nodes, _ := c.b.resolve(c.req.Targets[i])
			if res.Resolution != wantRes {
				c.add(RuleResolution, loc+"/resolution", "%q, the bundle resolves the target as %q", res.Resolution, wantRes)
			} else {
				var want []string
				for _, p := range nodes {
					want = append(want, c.b.Nodes[p].NodeID)
				}
				if want == nil {
					want = []string{}
				}
				if !reflect.DeepEqual(want, res.MatchedNodeIDs) {
					c.add(RuleResolution, loc+"/matched_node_ids", "%v, the bundle matches %v", res.MatchedNodeIDs, want)
				}
			}
		}
		c.checkState(&res, loc)
		// Model.
		m := res.Model
		if m.GlobalGapCount != len(c.b.Gaps) {
			c.add(RuleModel, loc+"/model/global_gap_count", "%d, the bundle has %d gaps", m.GlobalGapCount, len(c.b.Gaps))
		}
		if m.RelevanceRuleVersion != RelevanceRule {
			c.add(RuleModel, loc+"/model/relevance_rule_version", "%q, want %q", m.RelevanceRuleVersion, RelevanceRule)
		}
		for _, g := range m.RelevantGapIDs {
			if _, ok := c.b.idx.gapByID[g]; !ok {
				c.add(RuleModel, loc+"/model/relevant_gap_ids", "gap %s is not in the bundle", g)
				continue
			}
			need(TypeGap, g, loc+"/model/relevant_gap_ids")
		}
		switch m.State {
		case "complete_under_declared_model":
			if len(m.RelevantGapIDs) > 0 || m.ScopeUncertain {
				c.add(RuleModel, loc+"/model", "complete with relevant gaps or an uncertain scope")
			}
		case "incomplete":
			if len(m.RelevantGapIDs) == 0 {
				c.add(RuleModel, loc+"/model", "incomplete without a relevant gap")
			}
		case "undetermined":
		default:
			c.add(RuleModel, loc+"/model/state", "unknown state %q", m.State)
		}
		if res.Resolution != "matched" && m.State != "undetermined" {
			c.add(RuleModel, loc+"/model/state", "%q for an unresolved target", m.State)
		}
		// Entries.
		if len(res.Entries) > lim.MaxEntries {
			c.add(RuleLimits, loc+"/entries", "%d entries over max_entries %d", len(res.Entries), lim.MaxEntries)
		}
		seenRoot := map[string]bool{}
		for j, e := range res.Entries {
			c.checkEntry(e, fmt.Sprintf("%s/entries/%d", loc, j), res.Search.WorkComplete, matchedSet, rootSet, subset, fw, lim, seenRoot, needNode, need)
		}
	}
	if work > lim.MaxWorkUnits {
		c.add(RuleLimits, "/results", "%d work units over max_work_units %d", work, lim.MaxWorkUnits)
	}
	// Records no result refers to are allowed (the contract's fixtures carry
	// them); replay pins this producer's exact set.
}

// checkState applies the per-state rules (CONTRACT.md §8).
func (c *checker) checkState(res *TargetResult, loc string) {
	s := res.Search
	n := len(res.Entries)
	bad := func(rule, field, format string, a ...any) { c.add(rule, loc+field, format, a...) }
	switch s.State {
	case "not_evaluated":
		if res.Resolution == "matched" {
			bad(RuleSearchState, "/search/state", "not_evaluated for a matched target")
		}
		if n > 0 || len(res.MatchedNodeIDs) > 0 && (res.Resolution == "missing" || res.Resolution == "shadowed") {
			bad(RuleSearchState, "/entries", "entries or matched nodes without an evaluation")
		}
	case "found":
		if res.Resolution != "matched" {
			bad(RuleSearchState, "/search/state", "found for an unresolved target")
		}
		if n == 0 {
			bad(RuleSearchState, "/entries", "found without path evidence")
		}
		if s.WorkComplete {
			if res.TotalEntryCount == nil {
				bad(RuleEntryCount, "/total_entry_count", "null after a complete search")
			} else if *res.TotalEntryCount < n {
				bad(RuleEntryCount, "/total_entry_count", "%d below the %d entries shown", *res.TotalEntryCount, n)
			} else if res.EntriesTruncated != (*res.TotalEntryCount > n) {
				bad(RuleEntryCount, "/entries_truncated", "%v with %d of %d entries shown", res.EntriesTruncated, n, *res.TotalEntryCount)
			}
		} else {
			if res.TotalEntryCount != nil {
				bad(RuleEntryCount, "/total_entry_count", "%d after an incomplete search (the count is unknown: null)", *res.TotalEntryCount)
			}
			if !res.EntriesTruncated {
				bad(RuleEntryCount, "/entries_truncated", "false after an incomplete search")
			}
			if s.Reason == nil {
				bad(RuleSearchState, "/search/reason", "an incomplete search needs a reason")
			}
		}
	case "not_found":
		if res.Resolution != "matched" {
			bad(RuleSearchState, "/search/state", "not_found for an unresolved target")
		}
		if n > 0 {
			bad(RuleSearchState, "/entries", "not_found with entries")
		}
		if !s.WorkComplete {
			bad(RuleSearchState, "/search/work_complete", "not_found needs a complete search")
		}
		if res.TotalEntryCount == nil || *res.TotalEntryCount != 0 {
			bad(RuleEntryCount, "/total_entry_count", "not_found needs 0")
		}
	case "incomplete":
		if n > 0 {
			bad(RuleSearchState, "/entries", "incomplete with entries (a witness makes it found)")
		}
		if s.WorkComplete {
			bad(RuleSearchState, "/search/work_complete", "incomplete with a complete search")
		}
		if res.TotalEntryCount != nil {
			bad(RuleEntryCount, "/total_entry_count", "%d for an incomplete search (the count is unknown: null)", *res.TotalEntryCount)
		}
		if s.Reason == nil {
			bad(RuleSearchState, "/search/reason", "an incomplete search needs a reason")
		}
	default:
		bad(RuleSearchState, "/search/state", "unknown state %q", s.State)
	}
	if s.State != "found" && s.State != "incomplete" && res.EntriesTruncated {
		bad(RuleEntryCount, "/entries_truncated", "true without entries to truncate")
	}
}

func (c *checker) checkEntry(e EntryEvidence, loc string, complete bool, matched, roots, subset map[string]bool, fw []bool, lim EffectiveLimits,
	seenRoot map[string]bool, needNode func(id, loc string), need func(typ, id, loc string)) {
	p := e.Path
	if !roots[e.RootNodeID] {
		c.add(RulePathRoot, loc+"/root_node_id", "%s is not an indexed root", e.RootNodeID)
	}
	if len(subset) > 0 && !subset[e.RootNodeID] {
		c.add(RulePathRoot, loc+"/root_node_id", "%s is outside the requested root subset", e.RootNodeID)
	}
	if seenRoot[e.RootNodeID] {
		c.add(RulePathRoot, loc+"/root_node_id", "a second entry for root %s", e.RootNodeID)
	}
	seenRoot[e.RootNodeID] = true
	if len(p.NodeIDs) == 0 {
		c.add(RulePathContinuity, loc+"/path/node_ids", "an empty path")
		return
	}
	if p.NodeIDs[0] != e.RootNodeID {
		c.add(RulePathRoot, loc+"/path/node_ids/0", "%s, the entry's root is %s", p.NodeIDs[0], e.RootNodeID)
	}
	if len(p.EdgeIDs) != len(p.NodeIDs)-1 {
		c.add(RulePathContinuity, loc+"/path/edge_ids", "%d edges for %d nodes", len(p.EdgeIDs), len(p.NodeIDs))
	}
	if len(p.NodeIDs) > lim.MaxPathNodes {
		c.add(RuleLimits, loc+"/path/node_ids", "%d nodes over max_path_nodes %d", len(p.NodeIDs), lim.MaxPathNodes)
	}
	for k, id := range p.NodeIDs {
		if _, ok := c.b.idx.nodePos[id]; !ok {
			c.add(RulePathContinuity, fmt.Sprintf("%s/path/node_ids/%d", loc, k), "node %s is not in the bundle", id)
			continue
		}
		needNode(id, fmt.Sprintf("%s/path/node_ids/%d", loc, k))
	}
	chars := map[string]bool{}
	for k, id := range p.EdgeIDs {
		eloc := fmt.Sprintf("%s/path/edge_ids/%d", loc, k)
		pos, ok := c.b.idx.edgePos[id]
		if !ok {
			c.add(RulePathContinuity, eloc, "edge %s is not in the bundle", id)
			continue
		}
		need(TypeEdge, id, eloc)
		ed := c.b.Edges[pos]
		if k+1 < len(p.NodeIDs) && (ed.From != p.NodeIDs[k] || ed.To != p.NodeIDs[k+1]) {
			c.add(RulePathContinuity, eloc, "edge %s runs %s → %s, the path steps %s → %s", id, ed.From, ed.To, p.NodeIDs[k], p.NodeIDs[k+1])
		}
		chars[ed.Character] = true
		if e.Pass == "app_only" {
			if fp, ok := c.b.idx.nodePos[ed.From]; ok && fw[fp] {
				c.add(RulePathPass, eloc, "an app_only path expands framework node %s", ed.From)
			}
		}
	}
	if e.Pass != "app_only" && e.Pass != "full" {
		c.add(RulePathPass, loc+"/pass", "unknown pass %q", e.Pass)
	}
	if p.FullNodeCount == nil {
		c.add(RulePathEnd, loc+"/path/full_node_count", "null")
	} else if !p.Truncated {
		if !matched[p.NodeIDs[len(p.NodeIDs)-1]] {
			c.add(RulePathEnd, loc+"/path/node_ids", "the path ends at %s, not at a matched node", p.NodeIDs[len(p.NodeIDs)-1])
		}
		if *p.FullNodeCount != len(p.NodeIDs) {
			c.add(RulePathEnd, loc+"/path/full_node_count", "%d for an untruncated path of %d nodes", *p.FullNodeCount, len(p.NodeIDs))
		}
		for _, id := range p.NodeIDs[:len(p.NodeIDs)-1] {
			if matched[id] {
				c.add(RulePathEnd, loc+"/path/node_ids", "the path passes matched node %s before its end", id)
				break
			}
		}
	} else {
		if *p.FullNodeCount <= len(p.NodeIDs) || len(p.NodeIDs) != lim.MaxPathNodes {
			c.add(RulePathEnd, loc+"/path", "truncated to %d of %d nodes under max_path_nodes %d", len(p.NodeIDs), *p.FullNodeCount, lim.MaxPathNodes)
		}
	}
	// Character: the ordered set over the path's edges; a truncated path
	// shows a prefix, so its edges' characters must be among them.
	got := map[string]bool{}
	order := map[string]int{}
	for i, ch := range charOrder {
		order[ch] = i
	}
	last := -1
	for _, ch := range e.Character {
		o, known := order[ch]
		if !known || got[ch] || o <= last {
			c.add(RulePathCharacter, loc+"/character", "%v is not an ordered set of characters", e.Character)
			break
		}
		got[ch] = true
		last = o
	}
	for ch := range chars {
		if !got[ch] {
			c.add(RulePathCharacter, loc+"/character", "the path has a %s edge the entry does not list", ch)
		}
	}
	if !p.Truncated {
		for ch := range got {
			if !chars[ch] {
				c.add(RulePathCharacter, loc+"/character", "%s listed, no edge of the path has it", ch)
			}
		}
	}
	// CONTRACT.md §9: an entry's flag says that every path from its root
	// to a matched node takes such an edge, not that the shown path does.
	// Recomputed from the bundle when the search was complete (an
	// incomplete classification may mark an entry as requiring it).
	if complete {
		if want := !c.reaches(e.RootNodeID, matched, func(ed Edge) bool { return ed.ClientDispatch }); want != e.ClientDispatch {
			c.add(RulePathDispatch, loc+"/client_dispatch", "%v, the bundle says %v (every path from the root takes a client dispatch edge: %v)", e.ClientDispatch, want, want)
		}
		if want := !c.reaches(e.RootNodeID, matched, func(ed Edge) bool { return ed.SuppliedDispatch }); want != e.SuppliedDispatch {
			c.add(RulePathDispatch, loc+"/supplied_dispatch", "%v, the bundle says %v (every path from the root takes a supplied dispatch edge: %v)", e.SuppliedDispatch, want, want)
		}
	}
}

// reaches: a matched node is reachable from root over edges drop rejects.
func (c *checker) reaches(root string, matched map[string]bool, drop func(Edge) bool) bool {
	start, ok := c.b.idx.nodePos[root]
	if !ok {
		return false
	}
	seen := map[int32]bool{start: true}
	queue := []int32{start}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		if matched[c.b.Nodes[v].NodeID] {
			return true
		}
		for _, ei := range c.b.idx.out[v] {
			ed := c.b.Edges[ei]
			if drop(ed) {
				continue
			}
			if u, ok := c.b.idx.nodePos[ed.To]; ok && !seen[u] {
				seen[u] = true
				queue = append(queue, u)
			}
		}
	}
	return false
}

// Replay compares a response with what this producer answers to the same
// request on the same bundle (RH-A0: real results are reproducible). nil
// when identical.
func Replay(b *Bundle, reqRaw, respRaw []byte) []Violation {
	var want any
	req, perr := ParseRequest(b, reqRaw)
	if perr != nil {
		want = ErrorEnvelope(reqRaw, perr)
	} else if resp, qerr := Query(b, req); qerr != nil {
		want = ErrorEnvelope(reqRaw, qerr)
	} else {
		want = resp
	}
	raw, err := MarshalResponse(want)
	if err != nil {
		return []Violation{{Rule: RuleReplay, Location: "/", Detail: err.Error()}}
	}
	if bytes.Equal(raw, respRaw) {
		return nil
	}
	return []Violation{{Rule: RuleReplay, Location: "/", Detail: "differs from this producer's answer to the request"}}
}
