package research

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Default and maximum limits (CONTRACT.md §10).
const (
	DefaultMaxRecords   = 100
	DefaultMaxEntries   = 20
	DefaultMaxPathNodes = 100
	DefaultMaxWorkUnits = 1000000
)

// FrameworkSetting is the identity setting listing the component
// instances whose methods the app_only pass does not expand (Ariadne's
// REACHABLE/FRAMEWORK split): comma-separated instance ids, sorted.
const FrameworkSetting = "framework_instance_ids"

// Limits are request limits; nil fields take defaults.
type Limits struct {
	MaxRecords   *int `json:"max_records,omitempty"`
	MaxEntries   *int `json:"max_entries,omitempty"`
	MaxPathNodes *int `json:"max_path_nodes,omitempty"`
	MaxWorkUnits *int `json:"max_work_units,omitempty"`
}

// EffectiveLimits are the limits a response applied.
type EffectiveLimits struct {
	MaxRecords   int `json:"max_records"`
	MaxEntries   int `json:"max_entries"`
	MaxPathNodes int `json:"max_path_nodes"`
	MaxWorkUnits int `json:"max_work_units"`
}

// Selector names a method by identity.
type Selector struct {
	InstanceID string `json:"instance_id"`
	Class      string `json:"class"`
	Name       string `json:"name"`
	Descriptor string `json:"descriptor"`
}

// Target is one reachability target.
type Target struct {
	ID       string    `json:"id"`
	MethodID *string   `json:"method_id,omitempty"`
	Selector *Selector `json:"selector,omitempty"`
}

// RootSelection selects the query roots.
type RootSelection struct {
	Selection string   `json:"selection"`
	NodeIDs   []string `json:"node_ids,omitempty"`
}

// Filters are discovery filters.
type Filters struct {
	InstanceID          *string  `json:"instance_id,omitempty"`
	ClassPrefix         *string  `json:"class_prefix,omitempty"`
	MethodName          *string  `json:"method_name,omitempty"`
	Descriptor          *string  `json:"descriptor,omitempty"`
	NodeIDs             []string `json:"node_ids,omitempty"`
	DeclaredClassPrefix *string  `json:"declared_class_prefix,omitempty"`
	ReachableOnly       *bool    `json:"reachable_only,omitempty"`
}

// Request is any query document (the union of the three request shapes;
// which fields an operation admits is checked after decoding).
type Request struct {
	SchemaVersion string          `json:"schema_version"`
	AnalysisID    string          `json:"analysis_id"`
	RequestID     string          `json:"request_id"`
	Operation     string          `json:"operation"`
	Limits        *Limits         `json:"limits,omitempty"`
	Targets       []Target        `json:"targets,omitempty"`
	Roots         *RootSelection  `json:"roots,omitempty"`
	Filters       *Filters        `json:"filters,omitempty"`
	Cursor        json.RawMessage `json:"cursor,omitempty"`
	NodeID        *string         `json:"node_id,omitempty"`
}

// Path is one entry's path.
type Path struct {
	NodeIDs       []string `json:"node_ids"`
	EdgeIDs       []string `json:"edge_ids"`
	Truncated     bool     `json:"truncated"`
	FullNodeCount *int     `json:"full_node_count"`
}

// EntryEvidence is one reaching entry.
type EntryEvidence struct {
	RootNodeID       string   `json:"root_node_id"`
	Pass             string   `json:"pass"`
	Path             Path     `json:"path"`
	Character        []string `json:"character"`
	ClientDispatch   bool     `json:"client_dispatch"`
	SuppliedDispatch bool     `json:"supplied_dispatch"`
}

// Search is a target's search outcome.
type Search struct {
	State        string  `json:"state"`
	WorkComplete bool    `json:"work_complete"`
	WorkUnits    int     `json:"work_units"`
	Reason       *string `json:"reason"`
}

// ModelAssessment is a target's model completeness.
type ModelAssessment struct {
	State                string   `json:"state"`
	RelevantGapIDs       []string `json:"relevant_gap_ids"`
	GlobalGapCount       int      `json:"global_gap_count"`
	RelevanceRuleVersion string   `json:"relevance_rule_version"`
	ScopeUncertain       bool     `json:"scope_uncertain"`
}

// TargetResult is one target's result.
type TargetResult struct {
	TargetID         string          `json:"target_id"`
	Resolution       string          `json:"resolution"`
	MatchedNodeIDs   []string        `json:"matched_node_ids"`
	Search           Search          `json:"search"`
	Model            ModelAssessment `json:"model"`
	Entries          []EntryEvidence `json:"entries"`
	TotalEntryCount  *int            `json:"total_entry_count"`
	EntriesTruncated bool            `json:"entries_truncated"`
	RootBasis        string          `json:"root_basis"`
}

// Page is a discovery page.
type Page struct {
	ReturnedCount int     `json:"returned_count"`
	TotalCount    *int    `json:"total_count"`
	NextCursor    *string `json:"next_cursor"`
	Truncated     bool    `json:"truncated"`
	WorkComplete  bool    `json:"work_complete"`
}

// ReachabilityResponse answers a reachability request.
type ReachabilityResponse struct {
	SchemaVersion    string          `json:"schema_version"`
	ContractRevision string          `json:"contract_revision"`
	Status           string          `json:"status"`
	AnalysisID       string          `json:"analysis_id"`
	RequestID        string          `json:"request_id"`
	Operation        string          `json:"operation"`
	Provenance       Provenance      `json:"provenance"`
	EffectiveLimits  EffectiveLimits `json:"effective_limits"`
	Warnings         []string        `json:"warnings"`
	Results          []TargetResult  `json:"results"`
	Evidence         []any           `json:"evidence"`
}

// DiscoveryResponse answers entries/methods/calls/callers/callees.
type DiscoveryResponse struct {
	SchemaVersion    string          `json:"schema_version"`
	ContractRevision string          `json:"contract_revision"`
	Status           string          `json:"status"`
	AnalysisID       string          `json:"analysis_id"`
	RequestID        string          `json:"request_id"`
	Operation        string          `json:"operation"`
	Provenance       Provenance      `json:"provenance"`
	EffectiveLimits  EffectiveLimits `json:"effective_limits"`
	Warnings         []string        `json:"warnings"`
	Records          []any           `json:"records"`
	Page             Page            `json:"page"`
}

// ErrorDetail locates an error.
type ErrorDetail struct {
	Location string `json:"location"`
	Reason   string `json:"reason"`
}

// ErrorBody is the error member of an error response.
type ErrorBody struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []ErrorDetail `json:"details"`
	Retryable bool          `json:"retryable"`
}

// ErrorResponse is the error envelope (CONTRACT.md §11).
type ErrorResponse struct {
	SchemaVersion    string    `json:"schema_version"`
	ContractRevision string    `json:"contract_revision"`
	Status           string    `json:"status"`
	AnalysisID       *string   `json:"analysis_id"`
	RequestID        *string   `json:"request_id"`
	Operation        *string   `json:"operation"`
	Error            ErrorBody `json:"error"`
}

// ErrorEnvelope builds the error response for err, with whatever
// correlation fields of the request are valid (raw may be nil).
func ErrorEnvelope(raw []byte, err *Error) ErrorResponse {
	r := ErrorResponse{SchemaVersion: ResponseSchema, ContractRevision: ContractRevision, Status: "error",
		Error: ErrorBody{Code: err.Code, Message: err.Message, Details: []ErrorDetail{}, Retryable: err.Code == "IO_ERROR"}}
	if err.Location != "" {
		r.Error.Details = append(r.Error.Details, ErrorDetail{Location: err.Location, Reason: orDefault(err.Reason, err.Message)})
	}
	var head struct {
		AnalysisID any `json:"analysis_id"`
		RequestID  any `json:"request_id"`
		Operation  any `json:"operation"`
	}
	if raw != nil && json.Unmarshal(raw, &head) == nil {
		if s, ok := head.AnalysisID.(string); ok && hashRE.MatchString(s) {
			r.AnalysisID = &s
		}
		if s, ok := head.RequestID.(string); ok && identifierRE.MatchString(s) {
			r.RequestID = &s
		}
		if s, ok := head.Operation.(string); ok && s != "" {
			r.Operation = &s
		}
	}
	return r
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func invalid(loc, reason string) *Error {
	return &Error{Code: "INVALID_REQUEST", Message: "invalid request: " + reason, Location: loc, Reason: reason}
}

var acceptedFilters = map[string][]string{
	"entries": {"instance_id", "class_prefix", "method_name", "descriptor", "node_ids"},
	"methods": {"instance_id", "class_prefix", "method_name", "descriptor", "node_ids", "reachable_only"},
	"calls":   {"instance_id", "class_prefix", "method_name", "descriptor", "node_ids", "declared_class_prefix", "reachable_only"},
}

// ParseRequest decodes and validates a request against bundle b.
func ParseRequest(b *Bundle, raw []byte) (*Request, *Error) {
	if err := checkDuplicates(raw); err != nil {
		return nil, invalid("/", err.Error())
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return nil, invalid("/", "the request is not a JSON object")
	}
	var sv string
	if json.Unmarshal(members["schema_version"], &sv) != nil || sv != QuerySchema {
		return nil, &Error{Code: "UNSUPPORTED_SCHEMA", Message: fmt.Sprintf("schema_version must be %q", QuerySchema), Location: "/schema_version", Reason: "unsupported request schema"}
	}
	var req Request
	if err := decodeStrict(raw, &req); err != nil {
		return nil, invalid("/", err.Error())
	}
	for _, f := range []string{"analysis_id", "request_id", "operation"} {
		if _, ok := members[f]; !ok {
			return nil, invalid("/"+f, "required field is missing")
		}
	}
	if !identifierRE.MatchString(req.RequestID) {
		return nil, invalid("/request_id", "not a valid identifier")
	}
	if !hashRE.MatchString(req.AnalysisID) {
		return nil, invalid("/analysis_id", "not a sha256: hash")
	}
	if req.AnalysisID != b.AnalysisID {
		return nil, &Error{Code: "ANALYSIS_MISMATCH", Message: fmt.Sprintf("the request is for %s, the bundle is %s", req.AnalysisID, b.AnalysisID), Location: "/analysis_id", Reason: "analysis differs from the bundle"}
	}
	if !contains(Capabilities, req.Operation) {
		return nil, &Error{Code: "UNSUPPORTED_OPERATION", Message: fmt.Sprintf("operation %q is not supported", req.Operation), Location: "/operation", Reason: "unknown operation"}
	}
	if !contains(b.Identity.Capabilities, req.Operation) {
		return nil, &Error{Code: "UNSUPPORTED_OPERATION", Message: fmt.Sprintf("the bundle does not offer %q", req.Operation), Location: "/operation", Reason: "not a capability of this bundle"}
	}
	// Fields per operation.
	allowed := map[string]bool{"schema_version": true, "analysis_id": true, "request_id": true, "operation": true, "limits": true}
	switch req.Operation {
	case "reachability":
		allowed["targets"], allowed["roots"] = true, true
	case "entries", "methods", "calls":
		allowed["filters"], allowed["cursor"] = true, true
	case "callers", "callees":
		allowed["node_id"], allowed["cursor"] = true, true
	}
	for k := range members {
		if !allowed[k] {
			return nil, invalid("/"+k, fmt.Sprintf("field not accepted by operation %s", req.Operation))
		}
	}
	if err := checkLimits(req.Limits); err != nil {
		return nil, err
	}
	switch req.Operation {
	case "reachability":
		if err := checkTargets(&req); err != nil {
			return nil, err
		}
		if err := checkRoots(b, &req); err != nil {
			return nil, err
		}
	case "callers", "callees":
		if req.NodeID == nil {
			return nil, invalid("/node_id", "required field is missing")
		}
		if _, ok := b.idx.nodePos[*req.NodeID]; !ok {
			return nil, invalid("/node_id", "unknown node")
		}
	default:
		if req.Filters != nil {
			if err := checkFilters(b, req.Operation, members["filters"], req.Filters); err != nil {
				return nil, err
			}
		}
	}
	if len(req.Cursor) > 0 && string(req.Cursor) != "null" {
		var s string
		if json.Unmarshal(req.Cursor, &s) != nil || s == "" {
			return nil, invalid("/cursor", "must be a nonempty string or null")
		}
	}
	return &req, nil
}

func checkLimits(l *Limits) *Error {
	if l == nil {
		return nil
	}
	check := func(v *int, name string, max int) *Error {
		if v != nil && (*v < 1 || *v > max) {
			return invalid("/limits/"+name, fmt.Sprintf("must be between 1 and %d", max))
		}
		return nil
	}
	for _, c := range []struct {
		v    *int
		name string
		max  int
	}{{l.MaxRecords, "max_records", 1000}, {l.MaxEntries, "max_entries", 1000}, {l.MaxPathNodes, "max_path_nodes", 10000}, {l.MaxWorkUnits, "max_work_units", 1000000000}} {
		if err := check(c.v, c.name, c.max); err != nil {
			return err
		}
	}
	return nil
}

func effective(l *Limits) EffectiveLimits {
	e := EffectiveLimits{DefaultMaxRecords, DefaultMaxEntries, DefaultMaxPathNodes, DefaultMaxWorkUnits}
	if l != nil {
		if l.MaxRecords != nil {
			e.MaxRecords = *l.MaxRecords
		}
		if l.MaxEntries != nil {
			e.MaxEntries = *l.MaxEntries
		}
		if l.MaxPathNodes != nil {
			e.MaxPathNodes = *l.MaxPathNodes
		}
		if l.MaxWorkUnits != nil {
			e.MaxWorkUnits = *l.MaxWorkUnits
		}
	}
	return e
}

func checkTargets(req *Request) *Error {
	if len(req.Targets) < 1 || len(req.Targets) > 256 {
		return invalid("/targets", "a reachability request takes 1 to 256 targets")
	}
	if req.Roots == nil {
		return invalid("/roots", "required field is missing")
	}
	seen := map[string]bool{}
	for i, t := range req.Targets {
		loc := fmt.Sprintf("/targets/%d", i)
		if !identifierRE.MatchString(t.ID) {
			return invalid(loc+"/id", "not a valid identifier")
		}
		if seen[t.ID] {
			return invalid(loc+"/id", "duplicate target id")
		}
		seen[t.ID] = true
		if (t.MethodID == nil) == (t.Selector == nil) {
			return invalid(loc, "a target has exactly one of method_id or selector")
		}
		if t.MethodID != nil && !identifierRE.MatchString(*t.MethodID) {
			return invalid(loc+"/method_id", "not a valid identifier")
		}
		if s := t.Selector; s != nil {
			if !identifierRE.MatchString(s.InstanceID) || s.Class == "" || s.Name == "" {
				return invalid(loc+"/selector", "instance_id, class and name are required")
			}
			if !ValidMethodDescriptor(s.Descriptor) {
				return invalid(loc+"/selector/descriptor", "not a JVM method descriptor")
			}
		}
	}
	return nil
}

func checkRoots(b *Bundle, req *Request) *Error {
	r := req.Roots
	switch r.Selection {
	case "indexed":
		if r.NodeIDs != nil {
			return invalid("/roots/node_ids", "indexed selection takes no node_ids")
		}
	case "subset":
		if len(r.NodeIDs) == 0 {
			return invalid("/roots/node_ids", "a subset needs at least one root")
		}
		seen := map[string]bool{}
		for i, id := range r.NodeIDs {
			if seen[id] {
				return invalid(fmt.Sprintf("/roots/node_ids/%d", i), "duplicate root")
			}
			seen[id] = true
			if !b.idx.rootByNode[id] {
				return &Error{Code: "ROOTS_OUTSIDE_INDEX", Message: fmt.Sprintf("%s is not an indexed root; another root scenario needs another index", id),
					Location: fmt.Sprintf("/roots/node_ids/%d", i), Reason: "not an indexed root"}
			}
		}
	default:
		return invalid("/roots/selection", `must be "indexed" or "subset"`)
	}
	return nil
}

func checkFilters(b *Bundle, op string, raw json.RawMessage, f *Filters) *Error {
	var present map[string]json.RawMessage
	json.Unmarshal(raw, &present)
	for k := range present {
		if !contains(acceptedFilters[op], k) {
			return &Error{Code: "UNSUPPORTED_FILTER", Message: fmt.Sprintf("filter %s is not accepted by %s", k, op), Location: "/filters/" + k, Reason: "unsupported filter for this operation"}
		}
	}
	if f.Descriptor != nil && !ValidMethodDescriptor(*f.Descriptor) {
		return invalid("/filters/descriptor", "not a JVM method descriptor")
	}
	if f.InstanceID != nil && !identifierRE.MatchString(*f.InstanceID) {
		return invalid("/filters/instance_id", "not a valid identifier")
	}
	for k, v := range map[string]*string{"class_prefix": f.ClassPrefix, "method_name": f.MethodName, "declared_class_prefix": f.DeclaredClassPrefix} {
		if v != nil && *v == "" {
			return invalid("/filters/"+k, "must be nonempty")
		}
	}
	if present["node_ids"] != nil && len(f.NodeIDs) == 0 {
		return invalid("/filters/node_ids", "must list at least one node")
	}
	seen := map[string]bool{}
	for i, id := range f.NodeIDs {
		if seen[id] {
			return invalid(fmt.Sprintf("/filters/node_ids/%d", i), "duplicate node")
		}
		seen[id] = true
	}
	return nil
}

// Query answers a validated request.
func Query(b *Bundle, req *Request) (any, *Error) {
	lim := effective(req.Limits)
	switch req.Operation {
	case "reachability":
		return b.reachability(req, lim), nil
	default:
		return b.discover(req, lim)
	}
}

// ---- discovery ----

type cursorState struct {
	Analysis string `json:"a"`
	Binding  string `json:"b"`
	After    string `json:"after"`
}

func binding(req *Request, lim EffectiveLimits) string {
	f, _ := json.Marshal(req.Filters)
	node := ""
	if req.NodeID != nil {
		node = *req.NodeID
	}
	s := sha256.Sum256([]byte(strings.Join([]string{QuerySchema, ContractRevision, req.Operation, string(f), node, fmt.Sprint(lim.MaxRecords)}, "\x00")))
	return hex.EncodeToString(s[:12])
}

func encodeCursor(c cursorState) string {
	b, _ := json.Marshal(c)
	return "c1." + base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (cursorState, bool) {
	var c cursorState
	rest, ok := strings.CutPrefix(s, "c1.")
	if !ok {
		return c, false
	}
	b, err := base64.RawURLEncoding.DecodeString(rest)
	if err != nil || json.Unmarshal(b, &c) != nil || c.After == "" {
		return c, false
	}
	return c, true
}

func (b *Bundle) provenance() Provenance { return b.Identity.Provenance }

func (b *Bundle) discover(req *Request, lim EffectiveLimits) (any, *Error) {
	bind := binding(req, lim)
	after := ""
	if len(req.Cursor) > 0 && string(req.Cursor) != "null" {
		var s string
		json.Unmarshal(req.Cursor, &s)
		c, ok := decodeCursor(s)
		if !ok || c.Analysis != b.AnalysisID || c.Binding != bind {
			return nil, &Error{Code: "INVALID_CURSOR", Message: "the cursor is invalid or bound to another analysis, operation, filter set or record limit", Location: "/cursor", Reason: "cursor binding differs"}
		}
		after = c.After
	}
	type item struct {
		id  string
		rec any
	}
	budget := lim.MaxWorkUnits
	work := 0
	exhausted := false
	charge := func(n int) bool {
		if work+n > budget {
			exhausted = true
			return false
		}
		work += n
		return true
	}
	var reach []bool
	f := req.Filters
	if f == nil {
		f = &Filters{}
	}
	if f.ReachableOnly != nil && *f.ReachableOnly {
		var ok bool
		reach, ok = b.forward(b.idx.rootNodePos, false, charge)
		if !ok {
			reach = nil
		}
	}
	nodeIDs := map[string]bool{}
	for _, id := range f.NodeIDs {
		nodeIDs[id] = true
	}
	matchNode := func(p int32) bool {
		n := b.Nodes[p]
		if len(nodeIDs) > 0 && !nodeIDs[n.NodeID] {
			return false
		}
		if f.ReachableOnly != nil && *f.ReachableOnly && (reach == nil || !reach[p]) {
			return false
		}
		if f.InstanceID == nil && f.ClassPrefix == nil && f.MethodName == nil && f.Descriptor == nil {
			return true
		}
		m := n.Method
		if m == nil {
			return false
		}
		if f.InstanceID != nil && (m.InstanceID == nil || *m.InstanceID != *f.InstanceID) {
			return false
		}
		if f.ClassPrefix != nil && !strings.HasPrefix(m.Class, *f.ClassPrefix) {
			return false
		}
		if f.MethodName != nil && m.Name != *f.MethodName {
			return false
		}
		if f.Descriptor != nil && m.Descriptor != *f.Descriptor {
			return false
		}
		return true
	}
	var all []item
	total := 0
	scan := func(id string, keep func() bool, rec any) {
		if exhausted || !charge(1) {
			return
		}
		if keep() {
			total++
			if id > after {
				all = append(all, item{id, rec})
			}
		}
	}
	switch req.Operation {
	case "entries":
		for _, r := range b.Roots {
			r := r
			scan(r.RootID, func() bool { return matchNode(b.idx.nodePos[r.NodeID]) }, r)
		}
	case "methods":
		for i, n := range b.Nodes {
			i, n := i, n
			scan(n.NodeID, func() bool { return n.NodeKind != "synthetic_event" && matchNode(int32(i)) }, n)
		}
	case "calls":
		for _, c := range b.Calls {
			c := c
			scan(c.CallID, func() bool {
				if f.DeclaredClassPrefix != nil && !strings.HasPrefix(c.DeclaredClass, *f.DeclaredClassPrefix) {
					return false
				}
				return matchNode(b.idx.nodePos[c.Site.CallerNodeID])
			}, c)
		}
	case "callers", "callees":
		p := b.idx.nodePos[*req.NodeID]
		list := b.idx.in[p]
		if req.Operation == "callees" {
			list = b.idx.out[p]
		}
		for _, e := range list {
			ed := b.Edges[e]
			scan(ed.EdgeID, func() bool { return true }, ed)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].id < all[j].id })
	resp := DiscoveryResponse{SchemaVersion: ResponseSchema, ContractRevision: ContractRevision, Status: "ok", AnalysisID: b.AnalysisID,
		RequestID: req.RequestID, Operation: req.Operation, Provenance: b.provenance(), EffectiveLimits: lim, Warnings: []string{}, Records: []any{}}
	shown := all
	if len(shown) > lim.MaxRecords {
		shown = shown[:lim.MaxRecords]
	}
	for _, it := range shown {
		resp.Records = append(resp.Records, it.rec)
	}
	resp.Page.ReturnedCount = len(shown)
	resp.Page.WorkComplete = !exhausted
	resp.Page.Truncated = exhausted || len(all) > len(shown)
	if !exhausted {
		resp.Page.TotalCount = &total // every matching record, all pages
	}
	if resp.Page.Truncated && len(shown) > 0 {
		c := encodeCursor(cursorState{Analysis: b.AnalysisID, Binding: bind, After: shown[len(shown)-1].id})
		resp.Page.NextCursor = &c
	}
	if exhausted {
		resp.Warnings = append(resp.Warnings, fmt.Sprintf("max_work_units (%d) ended the scan; the page is incomplete", budget))
	}
	return resp, nil
}

// forward is a multi-source BFS from roots (positions, visited in the
// given order). appOnly stops at framework nodes. charge accounts each
// node visit and edge inspection; ok is false when the budget ended.
func (b *Bundle) forward(roots []int32, appOnly bool, charge func(int) bool) ([]bool, bool) {
	fw := b.frameworkNodes()
	vis := make([]bool, len(b.Nodes))
	queue := make([]int32, 0, len(roots))
	for _, r := range roots {
		if !vis[r] {
			vis[r] = true
			queue = append(queue, r)
		}
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		if !charge(1) {
			return vis, false
		}
		if appOnly && fw[u] {
			continue
		}
		for _, e := range b.idx.out[u] {
			if !charge(1) {
				return vis, false
			}
			t := b.idx.nodePos[b.Edges[e].To]
			if !vis[t] {
				vis[t] = true
				queue = append(queue, t)
			}
		}
	}
	return vis, true
}

func (b *Bundle) frameworkNodes() []bool {
	fw := make([]bool, len(b.Nodes))
	s, _ := b.Identity.Configuration.Settings[FrameworkSetting].(string)
	if s == "" {
		return fw
	}
	set := map[string]bool{}
	for _, id := range strings.Split(s, ",") {
		set[id] = true
	}
	for i := range b.Nodes {
		if set[b.nodeInstance(int32(i))] {
			fw[i] = true
		}
	}
	return fw
}

// ---- reachability ----

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// resolve maps a target to matched node positions.
func (b *Bundle) resolve(t Target) (resolution string, nodes []int32, reason string) {
	var canon []int32
	if t.MethodID != nil {
		canon = b.idx.byMethodID[*t.MethodID]
		if len(canon) == 0 {
			return "missing", nil, "Target method id is absent from this bundle."
		}
	} else {
		s := t.Selector
		var selectedHere, shadowedHere bool
		for _, ci := range b.idx.classByName[s.Class] {
			c := b.Classes[ci]
			if c.InstanceID != s.InstanceID {
				continue
			}
			if c.Selected {
				selectedHere = true
			} else {
				shadowedHere = true
			}
		}
		if shadowedHere && !selectedHere {
			return "shadowed", nil, "The requested instance's copy of the class lost selection to another copy; its methods are not in the graph."
		}
		for i, n := range b.Nodes {
			m := n.Method
			if n.NodeKind != "method" || m == nil || m.InstanceID == nil || *m.InstanceID != s.InstanceID {
				continue
			}
			if m.Class == s.Class && m.Name == s.Name && m.Descriptor == s.Descriptor {
				canon = append(canon, int32(i))
			}
		}
		if len(canon) == 0 {
			return "missing", nil, "Target method is absent from this bundle."
		}
	}
	if len(canon) > 1 {
		return "ambiguous", nil, "The selector matches more than one method node."
	}
	nodes = append(nodes, canon...)
	nodes = append(nodes, b.idx.specsOf[b.Nodes[canon[0]].NodeID]...)
	sort.Slice(nodes, func(i, j int) bool { return b.Nodes[nodes[i]].NodeID < b.Nodes[nodes[j]].NodeID })
	return "matched", nodes, ""
}

var charOrder = []string{"jvm_rule", "library_model", "input_declaration", "heuristic", "unclassified"}

func (b *Bundle) reachability(req *Request, lim EffectiveLimits) ReachabilityResponse {
	resp := ReachabilityResponse{SchemaVersion: ResponseSchema, ContractRevision: ContractRevision, Status: "ok", AnalysisID: b.AnalysisID,
		RequestID: req.RequestID, Operation: req.Operation, Provenance: b.provenance(), EffectiveLimits: lim, Warnings: []string{}}
	rootBasis := "indexed_graph"
	roots := b.idx.rootNodePos
	if req.Roots.Selection == "subset" {
		rootBasis = "subset_on_indexed_graph"
		roots = nil
		for _, id := range req.Roots.NodeIDs {
			roots = append(roots, b.idx.nodePos[id])
		}
		sort.Slice(roots, func(i, j int) bool { return b.Nodes[roots[i]].NodeID < b.Nodes[roots[j]].NodeID })
	}
	ev := newEvidence(b)
	remaining := lim.MaxWorkUnits
	var reachFull []bool // shared: forward full-pass reachability from the query roots
	sharedDone, sharedOK := false, false
	for _, t := range req.Targets {
		res := TargetResult{TargetID: t.ID, MatchedNodeIDs: []string{}, Entries: []EntryEvidence{}, RootBasis: rootBasis,
			Model: ModelAssessment{RelevantGapIDs: []string{}, GlobalGapCount: len(b.Gaps), RelevanceRuleVersion: RelevanceRule}}
		resolution, matched, why := b.resolve(t)
		res.Resolution = resolution
		if resolution != "matched" {
			res.Search = Search{State: "not_evaluated", WorkComplete: true, Reason: strPtr(why)}
			res.Model.State, res.Model.ScopeUncertain = "undetermined", true
			res.TotalEntryCount = intPtr(0)
			resp.Results = append(resp.Results, res)
			continue
		}
		for _, p := range matched {
			res.MatchedNodeIDs = append(res.MatchedNodeIDs, b.Nodes[p].NodeID)
			ev.node(p)
		}
		used := 0
		charge := func(n int) bool {
			if n > remaining {
				remaining = 0
				return false
			}
			remaining -= n
			used += n
			return true
		}
		if !sharedDone {
			sharedDone = true
			reachFull, sharedOK = b.forward(roots, false, charge)
		}
		b.assessModel(&res.Model, matched, reachFull, sharedOK, ev)
		if !sharedOK {
			res.Search = Search{State: "incomplete", WorkComplete: false, WorkUnits: used, Reason: strPtr("max_work_units exhausted before a path was discovered.")}
			res.EntriesTruncated = true
			resp.Results = append(resp.Results, res)
			continue
		}
		anyReached := false
		for _, p := range matched {
			if reachFull[p] {
				anyReached = true
			}
		}
		if !anyReached {
			res.Search = Search{State: "not_found", WorkComplete: true, WorkUnits: used}
			res.TotalEntryCount = intPtr(0)
			resp.Results = append(resp.Results, res)
			continue
		}
		entries, total, complete, reason := b.entries(matched, roots, reachFull, lim, charge)
		res.Search.WorkUnits = used
		res.Search.WorkComplete = complete
		if reason != "" {
			res.Search.Reason = strPtr(reason)
		}
		switch {
		case len(entries) > 0:
			res.Search.State = "found"
		case complete:
			// Reached in the shared pass yet no root path: cannot happen
			// on a consistent bundle; reported, not hidden.
			res.Search.State = "not_found"
			res.TotalEntryCount = intPtr(0)
		default:
			res.Search.State = "incomplete"
			res.EntriesTruncated = true
		}
		if len(entries) > 0 {
			if complete {
				res.TotalEntryCount = intPtr(total)
			} else {
				res.EntriesTruncated = true
			}
			if len(entries) > lim.MaxEntries {
				entries = entries[:lim.MaxEntries]
				res.EntriesTruncated = true
			}
			for _, e := range entries {
				for _, id := range e.Path.NodeIDs {
					ev.node(b.idx.nodePos[id])
				}
				for _, id := range e.Path.EdgeIDs {
					ev.edge(b.idx.edgePos[id])
				}
			}
			res.Entries = entries
		}
		resp.Results = append(resp.Results, res)
	}
	resp.Evidence = ev.records()
	return resp
}

// assessModel applies research-relevance-1 (CONTRACT.md §8).
func (b *Bundle) assessModel(m *ModelAssessment, matched []int32, reach []bool, reachOK bool, ev *evidence) {
	insts := map[string]bool{}
	methods := map[string]bool{}
	for _, p := range matched {
		if in := b.nodeInstance(p); in != "" {
			insts[in] = true
		}
		if mm := b.Nodes[p].Method; mm != nil {
			methods[mm.MethodID] = true
		}
	}
	for i, g := range b.Gaps {
		relevant := false
		switch {
		case g.Site == nil:
			relevant = true
		case !reachOK:
			// Query-root reachability unknown: the indexed hint stands in,
			// and the scope is uncertain.
			relevant = g.ReachableFromIndexedRoot
			m.ScopeUncertain = true
		case reach[b.idx.nodePos[g.Site.CallerNodeID]]:
			if g.CandidateInstanceIDs == nil && g.CandidateMethodIDs == nil {
				relevant = true
				m.ScopeUncertain = true
				break
			}
			for _, id := range g.CandidateInstanceIDs {
				relevant = relevant || insts[id]
			}
			for _, id := range g.CandidateMethodIDs {
				relevant = relevant || methods[id]
			}
		}
		if relevant {
			m.RelevantGapIDs = append(m.RelevantGapIDs, g.GapID)
			ev.gap(i)
		}
	}
	sort.Strings(m.RelevantGapIDs)
	switch {
	case len(m.RelevantGapIDs) > 0:
		m.State = "incomplete"
	case !reachOK:
		m.State = "undetermined"
	default:
		m.State = "complete_under_declared_model"
	}
}

// entries finds, for every query root that reaches a matched node, one
// witness path under research-paths-1: the app_only pass (framework nodes
// not expanded) before the full pass; within a pass, a path without
// CHA-only steps before one with them; then the shortest, ties by node and
// edge id. It is a reverse BFS from the targets per (pass, CHA) variant.
// Only nodes the roots reach (reach, the shared forward pass) can lie on
// a root's path, so the reverse search does not leave them.
func (b *Bundle) entries(matched []int32, roots []int32, reach []bool, lim EffectiveLimits, charge func(int) bool) ([]EntryEvidence, int, bool, string) {
	fw := b.frameworkNodes()
	isTarget := make([]bool, len(b.Nodes))
	for _, p := range matched {
		isTarget[p] = true
	}
	type variant struct {
		pass  string
		noCHA bool
	}
	variants := []variant{{"app_only", true}, {"app_only", false}, {"full", true}, {"full", false}}
	chosen := map[int32]EntryEvidence{}
	complete := true
	reason := ""
	for _, v := range variants {
		keep := func(e int) bool {
			ed := b.Edges[e]
			if !reach[b.idx.nodePos[ed.From]] {
				return false
			}
			if v.noCHA && ed.Basis == "cha" {
				return false
			}
			if v.pass == "app_only" && fw[b.idx.nodePos[ed.From]] {
				return false
			}
			return true
		}
		next, vis, ok := b.reverse(matched, keep, charge)
		for _, r := range roots {
			if _, done := chosen[r]; done || !vis[r] {
				continue
			}
			chosen[r] = b.witness(r, next, isTarget, v.pass, lim)
		}
		if !ok {
			complete = false
			reason = "max_work_units ended the search; the entries found so far are valid witnesses."
			break
		}
	}
	var out []EntryEvidence
	for _, r := range roots {
		if e, ok := chosen[r]; ok {
			out = append(out, e)
		}
	}
	if complete && b.idx.hasDispatch && len(out) > 0 {
		for _, flag := range []string{"client", "supplied"} {
			keep := func(e int) bool {
				ed := b.Edges[e]
				if !reach[b.idx.nodePos[ed.From]] {
					return false
				}
				if flag == "client" {
					return !ed.ClientDispatch
				}
				return !ed.SuppliedDispatch
			}
			_, vis, ok := b.reverse(matched, keep, charge)
			for i := range out {
				p := b.idx.nodePos[out[i].RootNodeID]
				needs := !vis[p] || !ok
				if flag == "client" {
					out[i].ClientDispatch = needs
				} else {
					out[i].SuppliedDispatch = needs
				}
			}
			if !ok {
				complete = false
				reason = "max_work_units ended while classifying client/supplied dispatch; entries not shown independent of it are marked as requiring it."
				break
			}
		}
	}
	return out, len(out), complete, reason
}

// reverse is a BFS from the targets over incoming edges keep admits;
// next[u] is the edge u takes one step closer to a target (-1 at a target).
func (b *Bundle) reverse(targets []int32, keep func(int) bool, charge func(int) bool) ([]int, []bool, bool) {
	next := make([]int, len(b.Nodes))
	vis := make([]bool, len(b.Nodes))
	queue := make([]int32, 0, len(targets))
	for _, t := range targets {
		vis[t] = true
		next[t] = -1
		queue = append(queue, t)
	}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		if !charge(1) {
			return next, vis, false
		}
		for _, e := range b.idx.in[v] {
			if !charge(1) {
				return next, vis, false
			}
			if !keep(e) {
				continue
			}
			u := b.idx.nodePos[b.Edges[e].From]
			if !vis[u] {
				vis[u] = true
				next[u] = e
				queue = append(queue, u)
			}
		}
	}
	return next, vis, true
}

func (b *Bundle) witness(r int32, next []int, isTarget []bool, pass string, lim EffectiveLimits) EntryEvidence {
	nodes := []string{b.Nodes[r].NodeID}
	var edges []string
	chars := map[string]bool{}
	cur := r
	for !isTarget[cur] {
		e := next[cur]
		ed := b.Edges[e]
		edges = append(edges, ed.EdgeID)
		chars[ed.Character] = true
		cur = b.idx.nodePos[ed.To]
		nodes = append(nodes, ed.To)
	}
	full := len(nodes)
	p := Path{NodeIDs: nodes, EdgeIDs: edges, FullNodeCount: intPtr(full)}
	if p.EdgeIDs == nil {
		p.EdgeIDs = []string{}
	}
	if full > lim.MaxPathNodes {
		p.NodeIDs = nodes[:lim.MaxPathNodes]
		p.EdgeIDs = edges[:lim.MaxPathNodes-1]
		p.Truncated = true
	}
	ch := []string{}
	for _, c := range charOrder {
		if chars[c] {
			ch = append(ch, c)
		}
	}
	return EntryEvidence{RootNodeID: b.Nodes[r].NodeID, Pass: pass, Path: p, Character: ch}
}

// evidence collects the records a reachability response must carry.
type evidence struct {
	b                   *Bundle
	nodes, edges, gaps  map[int]bool
	classes, components map[int]bool
}

func newEvidence(b *Bundle) *evidence {
	return &evidence{b: b, nodes: map[int]bool{}, edges: map[int]bool{}, gaps: map[int]bool{}, classes: map[int]bool{}, components: map[int]bool{}}
}

func (e *evidence) node(p int32) {
	if e.nodes[int(p)] {
		return
	}
	e.nodes[int(p)] = true
	n := e.b.Nodes[p]
	if n.Method == nil || n.Method.InstanceID == nil {
		return
	}
	if ci, ok := e.b.idx.compByID[*n.Method.InstanceID]; ok {
		e.components[ci] = true
	}
	for _, ci := range e.b.idx.classByName[n.Method.Class] {
		if e.b.Classes[ci].InstanceID == *n.Method.InstanceID {
			e.classes[ci] = true
		}
	}
}

func (e *evidence) edge(i int) { e.edges[i] = true }
func (e *evidence) gap(i int)  { e.gaps[i] = true }

func (e *evidence) records() []any {
	type rec struct {
		typ, id string
		r       any
	}
	var all []rec
	for i := range e.components {
		c := e.b.Components[i]
		all = append(all, rec{TypeComponent, c.InstanceID, c})
	}
	for i := range e.classes {
		c := e.b.Classes[i]
		all = append(all, rec{TypeClass, c.ClassID, c})
	}
	for i := range e.nodes {
		n := e.b.Nodes[i]
		all = append(all, rec{TypeNode, n.NodeID, n})
	}
	for i := range e.edges {
		x := e.b.Edges[i]
		all = append(all, rec{TypeEdge, x.EdgeID, x})
	}
	for i := range e.gaps {
		g := e.b.Gaps[i]
		all = append(all, rec{TypeGap, g.GapID, g})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].typ != all[j].typ {
			return all[i].typ < all[j].typ
		}
		return all[i].id < all[j].id
	})
	out := make([]any, 0, len(all))
	for _, r := range all {
		out = append(out, r.r)
	}
	return out
}

// MarshalResponse renders a response document with a trailing newline.
func MarshalResponse(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
