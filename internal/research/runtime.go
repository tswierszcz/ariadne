package research

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/runtimeevents"
)

// RESEARCH_INTERFACE RI6 (experimental, outside contract 1.0.0-draft.1;
// docs/research-interface/RUNTIME.md): runtime evidence for arbitrary
// research targets. TraceConfig turns reachability targets into a JFR
// jdk.MethodTrace filter; Correlate attaches an ariadne-runtime-trace-1
// trace to the bundle's nodes by class content, never by name alone.

const (
	TraceRequestSchema = "ariadne-research-trace-request-1"
	TraceConfigSchema  = "ariadne-research-trace-config-1"
	CorrelationSchema  = "ariadne-research-correlation-1"
	RuntimeRule        = "research-runtime-1"
)

// TraceRequest names the targets to trace.
type TraceRequest struct {
	SchemaVersion string   `json:"schema_version"`
	AnalysisID    string   `json:"analysis_id"`
	RequestID     string   `json:"request_id"`
	Targets       []Target `json:"targets"`
}

// TracedTarget is one target of a trace configuration.
type TracedTarget struct {
	TargetID       string   `json:"target_id"`
	Resolution     string   `json:"resolution"`
	Reason         *string  `json:"reason"`
	MatchedNodeIDs []string `json:"matched_node_ids"`
	// Filter is the jdk.MethodTrace filter entry ("Class::method"), null
	// when the target is not traceable. JFR filters by name: every
	// overload of the method is traced too (AlsoTraced).
	Filter     *string  `json:"filter"`
	AlsoTraced []string `json:"also_traced_node_ids"`
}

// TraceConfig is the trace-config document.
type TraceConfig struct {
	SchemaVersion    string         `json:"schema_version"`
	ContractRevision string         `json:"contract_revision"`
	RuleVersion      string         `json:"rule_version"`
	AnalysisID       string         `json:"analysis_id"`
	RequestID        string         `json:"request_id"`
	Filters          []string       `json:"filters"`
	MinimumJDK       int            `json:"minimum_jdk"`
	Targets          []TracedTarget `json:"targets"`
}

// ParseTraceRequest decodes and checks a trace request against b.
func ParseTraceRequest(b *Bundle, raw []byte) (*TraceRequest, *Error) {
	var r TraceRequest
	if err := decodeStrict(raw, &r); err != nil {
		return nil, invalid("/", err.Error())
	}
	if r.SchemaVersion != TraceRequestSchema {
		return nil, &Error{Code: "UNSUPPORTED_VERSION", Message: fmt.Sprintf("schema_version %q, want %q", r.SchemaVersion, TraceRequestSchema), Location: "/schema_version"}
	}
	if r.AnalysisID != b.AnalysisID {
		return nil, &Error{Code: "ANALYSIS_MISMATCH", Message: "the request names another analysis", Location: "/analysis_id"}
	}
	if !identifierRE.MatchString(r.RequestID) {
		return nil, invalid("/request_id", "not an identifier")
	}
	if len(r.Targets) == 0 || len(r.Targets) > 256 {
		return nil, invalid("/targets", "1 to 256 targets")
	}
	if e := checkTargets(&Request{Operation: "reachability", Targets: r.Targets, Roots: &RootSelection{Selection: "indexed"}}); e != nil {
		return nil, e
	}
	return &r, nil
}

// TraceConfigFor resolves the targets as reachability does and derives the
// JFR filter.
func (b *Bundle) TraceConfigFor(r *TraceRequest) TraceConfig {
	c := TraceConfig{SchemaVersion: TraceConfigSchema, ContractRevision: ContractRevision, RuleVersion: RuntimeRule,
		AnalysisID: b.AnalysisID, RequestID: r.RequestID, Filters: []string{}, MinimumJDK: 25}
	set := map[string]bool{}
	for _, t := range r.Targets {
		res, nodes, why := b.resolve(t)
		tt := TracedTarget{TargetID: t.ID, Resolution: res, MatchedNodeIDs: []string{}, AlsoTraced: []string{}}
		if why != "" {
			tt.Reason = strPtr(why)
		}
		for _, p := range nodes {
			tt.MatchedNodeIDs = append(tt.MatchedNodeIDs, b.Nodes[p].NodeID)
		}
		if res == "matched" {
			m := b.Nodes[nodes[0]].Method
			f := m.Class + "::" + m.Name
			tt.Filter = &f
			set[f] = true
			matched := map[string]bool{}
			for _, id := range tt.MatchedNodeIDs {
				matched[id] = true
			}
			for _, n := range b.Nodes {
				if n.Method != nil && n.CanonicalNodeID == n.NodeID && n.Method.Class == m.Class && n.Method.Name == m.Name && !matched[n.NodeID] {
					tt.AlsoTraced = append(tt.AlsoTraced, n.NodeID)
				}
			}
			sort.Strings(tt.AlsoTraced)
		} else if tt.Reason == nil {
			tt.Reason = strPtr("unresolved targets are not traced")
		}
		c.Targets = append(c.Targets, tt)
	}
	for f := range set {
		c.Filters = append(c.Filters, f)
	}
	sort.Strings(c.Filters)
	return c
}

// JFC renders a trace configuration as JFR settings, given next to the
// usual ones: -XX:StartFlightRecording=settings=profile,settings=FILE.jfc.
func (c TraceConfig) JFC() []byte {
	esc := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, `<configuration version="2.0" label="ariadne research targets" description="%s">`+"\n",
		esc(fmt.Sprintf("ariadne research trace-config %s for analysis %s: exact entries (jdk.MethodTrace, JDK 25+) of %d methods", c.RequestID, c.AnalysisID, len(c.Filters))))
	b.WriteString(`  <event name="jdk.MethodTrace">` + "\n")
	b.WriteString(`    <setting name="enabled">` + fmt.Sprint(len(c.Filters) > 0) + "</setting>\n")
	b.WriteString(`    <setting name="stackTrace">true</setting>` + "\n")
	b.WriteString(`    <setting name="threshold">0 ms</setting>` + "\n")
	b.WriteString(`    <setting name="filter">` + esc(strings.Join(c.Filters, ";")) + "</setting>\n")
	b.WriteString("  </event>\n</configuration>\n")
	return []byte(b.String())
}

// TraceRef identifies the correlated trace.
type TraceRef struct {
	SHA256        string   `json:"sha256"`
	RunID         string   `json:"run_id"`
	Scenario      string   `json:"scenario"`
	Runtime       string   `json:"runtime"`
	Recorder      string   `json:"recorder"`
	Sampling      string   `json:"sampling"`
	Complete      bool     `json:"complete"`
	LostEvents    int      `json:"lost_events"`
	Truncated     int      `json:"truncated_stacks"`
	Reasons       []string `json:"incomplete_reasons"`
	MethodFilters []string `json:"method_filters"`
}

// ClassCounts summarizes the trace's classes against the bundle.
type ClassCounts struct {
	Matched     int `json:"matched"`      // name and content equal to a bundle class
	HashDiffers int `json:"hash_differs"` // a bundle class of that name, other bytes
	NotInBundle int `json:"not_in_bundle"`
}

// NodeObservation is execution evidence attached to a bundle node.
type NodeObservation struct {
	NodeID  string `json:"node_id"`
	Traced  int    `json:"traced"`
	Sampled int    `json:"sampled"`
}

// CallerObservation is a measured call into a target.
type CallerObservation struct {
	CallerNodeID *string `json:"caller_node_id"` // null: the caller's class is not matched
	CallerMethod string  `json:"caller_method"`  // Class#name+desc as recorded
	CalleeNodeID string  `json:"callee_node_id"`
	EdgeID       *string `json:"edge_id"` // the bundle's edge caller -> callee, null if none
	Basis        string  `json:"basis"`
	Source       string  `json:"source"`
	Count        int     `json:"count"`
	BCI          *int    `json:"bci"`
	Line         *int    `json:"line"`
}

// TargetObservation is the runtime evidence for one configured target.
type TargetObservation struct {
	TargetID       string   `json:"target_id"`
	MatchedNodeIDs []string `json:"matched_node_ids"`
	// ClassIdentity: matched (the run loaded these bytes), differs (the
	// run's class of that name has other bytes: nothing is attached),
	// not_seen (no class of that name in the trace).
	ClassIdentity string `json:"class_identity"`
	Traced        bool   `json:"traced"` // a method filter of the run covers the target
	// Observation: entered (traced entries), sampled (only on sampled
	// stacks), traced_not_entered, not_seen. None of them is evidence of
	// unreachability.
	Observation string              `json:"observation"`
	Entries     NodeObservation     `json:"entries"`
	Callers     []CallerObservation `json:"callers"`
}

// Correlation is the correlate document.
type Correlation struct {
	SchemaVersion    string              `json:"schema_version"`
	ContractRevision string              `json:"contract_revision"`
	RuleVersion      string              `json:"rule_version"`
	AnalysisID       string              `json:"analysis_id"`
	RequestID        *string             `json:"request_id"`
	Trace            TraceRef            `json:"trace"`
	Classes          ClassCounts         `json:"classes"`
	Calls            map[string]int      `json:"calls"` // in_bundle, not_in_bundle, unmatched
	Executed         []NodeObservation   `json:"executed"`
	Targets          []TargetObservation `json:"targets"`
	Warnings         []string            `json:"warnings"`
}

// Correlate attaches a trace to the bundle. cfg (optional) names targets.
func (b *Bundle) Correlate(t *runtimeevents.Trace, cfg *TraceConfig) (Correlation, *Error) {
	if cfg != nil && cfg.AnalysisID != b.AnalysisID {
		return Correlation{}, &Error{Code: "ANALYSIS_MISMATCH", Message: "the trace configuration names another analysis", Location: "-config"}
	}
	h := t.Header
	out := Correlation{SchemaVersion: CorrelationSchema, ContractRevision: ContractRevision, RuleVersion: RuntimeRule,
		AnalysisID: b.AnalysisID, Calls: map[string]int{"in_bundle": 0, "not_in_bundle": 0, "unmatched": 0},
		Executed: []NodeObservation{}, Targets: []TargetObservation{}, Warnings: []string{},
		Trace: TraceRef{SHA256: "sha256:" + t.SHA256, RunID: h.RunID, Scenario: h.Scenario, Runtime: strings.TrimSpace(h.Runtime.Name + " " + h.Runtime.Version),
			Recorder: strings.TrimSpace(h.Recorder.Name + " " + h.Recorder.Version), Sampling: h.Sampling.Kind, Complete: h.Completeness.Complete,
			LostEvents: h.Completeness.LostEvents, Truncated: h.Completeness.TruncatedStacks, Reasons: nonNil(h.Completeness.Reasons),
			MethodFilters: nonNil(h.Scope.MethodFilters)}}
	if h.Sampling.Kind == "periodic" {
		out.Trace.Sampling = fmt.Sprintf("periodic (%d ms)", h.Sampling.IntervalMS)
	}
	if cfg != nil {
		out.RequestID = strPtr(cfg.RequestID)
	}
	// Class identity: a trace class attaches to the bundle's class record
	// of the same name and bytes; its instance then decides the nodes.
	byName := map[string][]Class{}
	for _, c := range b.Classes {
		byName[c.Class] = append(byName[c.Class], c)
	}
	classInst := map[string]string{} // trace class id -> instance
	identity := map[string]string{}  // class name -> matched|differs (seen in the trace)
	for id, c := range t.Classes {
		cands := byName[c.Name]
		hit := ""
		for _, k := range cands {
			if k.ClassSHA256 == "sha256:"+c.SHA256 {
				hit = k.InstanceID
			}
		}
		switch {
		case hit != "":
			classInst[id] = hit
			out.Classes.Matched++
			identity[c.Name] = "matched"
		case len(cands) > 0:
			out.Classes.HashDiffers++
			if identity[c.Name] != "matched" {
				identity[c.Name] = "differs"
			}
		default:
			out.Classes.NotInBundle++
		}
	}
	if out.Classes.HashDiffers > 0 {
		out.Warnings = append(out.Warnings, fmt.Sprintf("%d classes of the run have the name of a bundle class but other bytes (another build or a transformed class): their evidence is not attached", out.Classes.HashDiffers))
	}
	// Method -> canonical node, through the matched class's instance.
	nodeOf := map[string]string{} // instance + "\x00" + class#name+desc -> node id
	for _, n := range b.Nodes {
		if n.Method == nil || n.Method.InstanceID == nil || n.CanonicalNodeID != n.NodeID {
			continue
		}
		nodeOf[*n.Method.InstanceID+"\x00"+n.Method.Class+"#"+n.Method.Name+n.Method.Descriptor] = n.NodeID
	}
	methodNode := func(mid string) (string, string) {
		m := t.Methods[mid]
		c := t.Classes[m.Class]
		key := c.Name + "#" + m.Name + m.Desc
		inst, ok := classInst[m.Class]
		if !ok {
			return "", key
		}
		return nodeOf[inst+"\x00"+key], key
	}
	executed := map[string]*NodeObservation{}
	for _, e := range t.Executed {
		id, _ := methodNode(e.Method)
		if id == "" {
			continue
		}
		o := executed[id]
		if o == nil {
			o = &NodeObservation{NodeID: id}
			executed[id] = o
		}
		if e.Basis == "traced" {
			o.Traced += e.Count
		} else {
			o.Sampled += e.Count
		}
	}
	for _, o := range executed {
		out.Executed = append(out.Executed, *o)
	}
	sort.Slice(out.Executed, func(i, j int) bool { return out.Executed[i].NodeID < out.Executed[j].NodeID })
	edgeBetween := map[string]string{}
	for _, e := range b.Edges {
		k := e.From + "\x00" + e.To
		if _, ok := edgeBetween[k]; !ok || e.EdgeID < edgeBetween[k] {
			edgeBetween[k] = e.EdgeID
		}
	}
	// The receiver specializations of a node are the same method at run
	// time: an edge from a specialization counts.
	specs := map[string][]string{}
	for _, n := range b.Nodes {
		if n.CanonicalNodeID != n.NodeID {
			specs[n.CanonicalNodeID] = append(specs[n.CanonicalNodeID], n.NodeID)
		}
	}
	edgeFor := func(from, to string) *string {
		for _, f := range append([]string{from}, specs[from]...) {
			for _, tt := range append([]string{to}, specs[to]...) {
				if id, ok := edgeBetween[f+"\x00"+tt]; ok {
					return &id
				}
			}
		}
		return nil
	}
	callersOf := map[string][]CallerObservation{}
	for _, c := range t.Calls {
		callee, _ := methodNode(c.Callee)
		caller, callerKey := methodNode(c.Caller)
		if callee == "" || caller == "" {
			out.Calls["unmatched"]++
		} else if edgeFor(caller, callee) != nil {
			out.Calls["in_bundle"]++
		} else {
			out.Calls["not_in_bundle"]++
		}
		if callee == "" {
			continue
		}
		o := CallerObservation{CallerMethod: callerKey, CalleeNodeID: callee, Basis: c.Basis, Source: c.Source, Count: c.Count}
		if caller != "" {
			o.CallerNodeID = strPtr(caller)
			o.EdgeID = edgeFor(caller, callee)
		}
		if c.BCI >= 0 {
			o.BCI = intPtr(c.BCI)
		}
		if c.Line >= 0 {
			o.Line = intPtr(c.Line)
		}
		callersOf[callee] = append(callersOf[callee], o)
	}
	if cfg == nil {
		return out, nil
	}
	for _, tt := range cfg.Targets {
		o := TargetObservation{TargetID: tt.TargetID, MatchedNodeIDs: tt.MatchedNodeIDs, Callers: []CallerObservation{}, ClassIdentity: "not_seen", Observation: "not_seen"}
		if tt.Resolution != "matched" || len(tt.MatchedNodeIDs) == 0 {
			o.Observation = "not_evaluated"
			out.Targets = append(out.Targets, o)
			continue
		}
		p := b.idx.nodePos[tt.MatchedNodeIDs[0]]
		canon := b.Nodes[p].CanonicalNodeID
		m := b.Nodes[p].Method
		if id, ok := identity[m.Class]; ok {
			o.ClassIdentity = id
		}
		if tt.Filter != nil {
			for _, f := range h.Scope.MethodFilters {
				if f == *tt.Filter || f == m.Class {
					o.Traced = true
				}
			}
		}
		o.Entries.NodeID = canon
		if e := executed[canon]; e != nil {
			o.Entries = *e
		}
		o.Callers = append(o.Callers, callersOf[canon]...)
		sort.Slice(o.Callers, func(i, j int) bool {
			a, c := o.Callers[i], o.Callers[j]
			if a.CallerMethod != c.CallerMethod {
				return a.CallerMethod < c.CallerMethod
			}
			return deref(a.BCI) < deref(c.BCI)
		})
		switch {
		case o.ClassIdentity == "differs":
			o.Observation = "class_differs"
		case o.Entries.Traced > 0:
			o.Observation = "entered"
		case o.Entries.Sampled > 0:
			o.Observation = "sampled"
		case o.Traced:
			o.Observation = "traced_not_entered"
		}
		out.Targets = append(out.Targets, o)
	}
	return out, nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func deref(p *int) int {
	if p == nil {
		return -1
	}
	return *p
}
