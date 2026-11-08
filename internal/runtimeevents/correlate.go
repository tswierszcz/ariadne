package runtimeevents

import (
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/callgraph"
)

// Class correlation statuses (docs/RUNTIME_TRACE.MD).
const (
	StatusMatched          = "matched"
	StatusMatchedArtifact  = "matched_artifact"
	StatusDifferentVersion = "different_version"
	StatusNameOnly         = "name_only"
	StatusNotInInput       = "not_in_input"
)

// Usable reports whether evidence of a class with this status may be
// attached to the analyzed class.
func Usable(status string) bool {
	return status == StatusMatched || status == StatusMatchedArtifact
}

// BasisThroughRuntime is the call basis with frames outside the recorded
// scope between caller and callee.
const BasisThroughRuntime = "stack-through-runtime"

// Call statuses against the static graph.
const (
	CallInGraph    = "in_graph"
	CallNotInGraph = "not_in_graph"
	CallUnusable   = "unusable" // an end's class is not usable evidence, or not a graph node
)

// Inputs is what correlation needs from the analysis.
type Inputs struct {
	G *callgraph.Graph
	// Lookup returns, for an analyzed class of that binary name, the
	// SHA-256 of its class file and of the artifact it was read from;
	// ok is false when no analyzed class has that name.
	Lookup func(fqcn string) (classSHA, artifactSHA string, ok bool)
	// Classify, when set, names why the static graph has no edge for a
	// measured call (caller and callee "owner#name+desc", the caller's
	// bytecode index, -1 when unknown, and the call's basis): a cause from
	// the Gap* constants and a detail (e.g. the call's receiver type).
	Classify func(caller, callee string, bci int, basis string) (cause, detail string)
}

// Causes of a measured call the static graph has no edge for, decided
// from the caller's bytecode at the measured index (GapCauseVersion).
const (
	GapCauseVersion    = "g1"
	GapClassInit       = "class-init"       // the callee is a <clinit> the JVM ran implicitly
	GapInvokeDynamic   = "invokedynamic"    // an invokedynamic site (lambda creation, string concatenation, …)
	GapReflection      = "reflection"       // Method#invoke, Constructor#newInstance, Class#newInstance
	GapMethodHandle    = "method-handle"    // MethodHandle#invoke*
	GapProxy           = "proxy"            // a call through a generated proxy into an InvocationHandler
	GapAsync           = "async"            // a task handed to a thread or executor
	GapSkippedReceiver = "skipped-receiver" // a call on a java./javax./jdk./sun. type CHA never expands
	GapRTAPruned       = "rta-pruned"       // a CHA target RTA removed
	GapInherited       = "inherited-impl"   // the receiver's class inherits the body from a superclass outside the call's type: CHA keeps only declaring subtypes
	GapSAMCall         = "sam-call"         // a functional-interface call running a lambda or method-reference body
	GapDispatchMissing = "dispatch-missing" // a non-skipped virtual call whose subtype target has no edge
	GapIndirect        = "indirect"         // the callee is not the site's target: frames in between were dropped
	GapNoCallAtSite    = "no-call-at-site"  // no call instruction at the measured index (inlining, line tables)
	GapUnclassified    = "unclassified"
)

// ClassResult is one recorded class definition and its status.
type ClassResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Loader     string `json:"loader"`
	LoaderName string `json:"loader_name,omitempty"`
	Status     string `json:"status"`
}

// MethodObservation is one executed record, correlated.
type MethodObservation struct {
	Evidence    string `json:"evidence"`
	Method      string `json:"method"` // owner#name+desc
	Loader      string `json:"loader"`
	Basis       string `json:"basis"`
	Count       int    `json:"count"`
	ClassStatus string `json:"class_status"`
	InGraph     bool   `json:"in_graph"`
}

// CallObservation is one call record, compared with the static graph.
type CallObservation struct {
	Evidence string   `json:"evidence"`
	Caller   string   `json:"caller"`
	Callee   string   `json:"callee"`
	Loader   string   `json:"callee_loader"` // loader id of the callee's class (several definitions stay apart)
	Basis    string   `json:"basis"`
	Source   string   `json:"source"`
	Count    int      `json:"count"`
	BCI      int      `json:"bci,omitempty"`
	Line     int      `json:"line,omitempty"`
	Status   string   `json:"status"`
	Bases    []string `json:"graph_bases,omitempty"` // bases of the static edges caller → callee
	// Cause and CauseDetail (with Inputs.Classify): why a not_in_graph
	// call has no static edge.
	Cause       string `json:"cause,omitempty"`
	CauseDetail string `json:"cause_detail,omitempty"`
}

// GapCause is one cause with its count and most frequent details.
type GapCause struct {
	Cause   string        `json:"cause"`
	Calls   int           `json:"calls"`
	Details []GapCauseTop `json:"top_details,omitempty"`
}

// GapCauseTop is one detail (e.g. a receiver type) and its count.
type GapCauseTop struct {
	Detail string `json:"detail"`
	Calls  int    `json:"calls"`
}

// Result is a correlated trace. It never changes the graph or a verdict.
type Result struct {
	ImportVersion string   `json:"import_version"`
	Format        string   `json:"format"`
	SHA256        string   `json:"trace_sha256"`
	RunID         string   `json:"run_id"`
	Scenario      string   `json:"scenario"`
	Started       string   `json:"started"`
	Ended         string   `json:"ended"`
	Runtime       string   `json:"runtime"`
	Recorder      string   `json:"recorder"`
	EventKinds    []string `json:"event_kinds"`
	MethodFilters []string `json:"method_filters,omitempty"`
	Sampling      string   `json:"sampling"`
	Complete      bool     `json:"complete"`
	LostEvents    int      `json:"lost_events"`
	Truncated     int      `json:"truncated_stacks"`
	Incomplete    []string `json:"incomplete_reasons,omitempty"`

	Classes      []ClassResult  `json:"-"`
	ClassCounts  map[string]int `json:"classes_by_status"`
	MultiDefined []string       `json:"classes_with_several_definitions,omitempty"`
	LoadedCounts map[string]int `json:"loaded_by_status"`

	Executed       []MethodObservation `json:"-"`
	ExecutedCounts map[string]int      `json:"executed_by_basis_usable"`
	ExecutedOther  int                 `json:"executed_not_usable"`

	Calls      []CallObservation `json:"-"`
	CallCounts map[string]int    `json:"calls_by_status"`
	Handoffs   int               `json:"handoffs"`
	// GapCauses (with Inputs.Classify): not_in_graph calls by cause, most
	// frequent first, each with its top 10 details.
	GapCauseVersion string     `json:"gap_cause_version,omitempty"`
	GapCauses       []GapCause `json:"calls_not_in_graph_by_cause,omitempty"`
}

// Correlate matches t against the analyzed input.
func Correlate(t *Trace, in Inputs) Result {
	h := t.Header
	r := Result{
		ImportVersion: ImportVersion, Format: h.Format, SHA256: t.SHA256, RunID: h.RunID, Scenario: h.Scenario,
		Started: h.Started, Ended: h.Ended,
		Runtime:    strings.TrimSpace(h.Runtime.Name + " " + h.Runtime.Version),
		Recorder:   h.Recorder.Name + " " + h.Recorder.Version,
		EventKinds: append([]string(nil), h.Scope.EventKinds...), MethodFilters: h.Scope.MethodFilters,
		Sampling: h.Sampling.Kind, Complete: h.Completeness.Complete, LostEvents: h.Completeness.LostEvents,
		Truncated: h.Completeness.TruncatedStacks, Incomplete: h.Completeness.Reasons,
		ClassCounts: map[string]int{}, LoadedCounts: map[string]int{}, ExecutedCounts: map[string]int{},
		CallCounts: map[string]int{}, Handoffs: len(t.Handoffs),
	}
	if h.Sampling.Kind == "periodic" {
		r.Sampling = "periodic"
		if h.Sampling.IntervalMS > 0 {
			r.Sampling += " (" + strconv.Itoa(h.Sampling.IntervalMS) + " ms)"
		}
	}

	status := map[string]string{}
	defs := map[string]int{}
	for _, id := range sortedIDs(t.Classes) {
		c := t.Classes[id]
		st := classStatus(t, c, in.Lookup)
		status[id] = st
		defs[c.Name]++
		r.Classes = append(r.Classes, ClassResult{ID: id, Name: c.Name, Loader: c.Loader, LoaderName: t.Loaders[c.Loader].Name, Status: st})
		r.ClassCounts[st]++
	}
	for name, n := range defs {
		if n > 1 {
			r.MultiDefined = append(r.MultiDefined, name)
		}
	}
	sort.Strings(r.MultiDefined)
	sort.Slice(r.Classes, func(i, j int) bool {
		a, b := r.Classes[i], r.Classes[j]
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Loader < b.Loader
	})
	for _, l := range t.Loaded {
		r.LoadedCounts[status[l.Class]]++
	}

	key := func(mid string) (string, string, string, string) {
		m := t.Methods[mid]
		c := t.Classes[m.Class]
		return c.Name + "#" + m.Name + m.Desc, c.Loader, status[m.Class], c.Name
	}
	node := func(mid string) (uint32, bool) {
		m := t.Methods[mid]
		if in.G == nil {
			return 0, false
		}
		return in.G.ID(t.Classes[m.Class].Name, m.Name, m.Desc)
	}
	for _, e := range t.Executed {
		k, loader, st, _ := key(e.Method)
		_, inGraph := node(e.Method)
		r.Executed = append(r.Executed, MethodObservation{Evidence: e.ID, Method: k, Loader: loader, Basis: e.Basis, Count: e.Count, ClassStatus: st, InGraph: inGraph})
		if Usable(st) && inGraph {
			r.ExecutedCounts[e.Basis]++
		} else {
			r.ExecutedOther++
		}
	}
	sort.Slice(r.Executed, func(i, j int) bool { return r.Executed[i].Evidence < r.Executed[j].Evidence })
	for _, c := range t.Calls {
		ck, _, cst, _ := key(c.Caller)
		ek, eloader, est, _ := key(c.Callee)
		o := CallObservation{Evidence: c.ID, Caller: ck, Callee: ek, Loader: eloader, Basis: c.Basis, Source: c.Source, Count: c.Count, Status: CallUnusable}
		if c.BCI >= 0 {
			o.BCI = c.BCI
		}
		if c.Line >= 0 {
			o.Line = c.Line
		}
		from, okFrom := node(c.Caller)
		to, okTo := node(c.Callee)
		if Usable(cst) && Usable(est) && okFrom && okTo {
			o.Status = CallNotInGraph
			seen := map[string]bool{}
			for _, edge := range in.G.Edges(from) {
				if edge.To == to && !seen[string(edge.Basis)] {
					seen[string(edge.Basis)] = true
					o.Bases = append(o.Bases, string(edge.Basis))
				}
			}
			if len(o.Bases) > 0 {
				o.Status = CallInGraph
				sort.Strings(o.Bases)
			} else if in.Classify != nil {
				o.Cause, o.CauseDetail = in.Classify(ck, ek, c.BCI, c.Basis)
				if o.Cause == "" {
					o.Cause = GapUnclassified
				}
			}
		}
		r.CallCounts[o.Status]++
		r.Calls = append(r.Calls, o)
	}
	sort.Slice(r.Calls, func(i, j int) bool { return r.Calls[i].Evidence < r.Calls[j].Evidence })
	if in.Classify != nil {
		r.GapCauseVersion = GapCauseVersion
		r.GapCauses = gapCauses(r.Calls)
	}
	return r
}

func gapCauses(calls []CallObservation) []GapCause {
	by := map[string]map[string]int{}
	total := map[string]int{}
	for _, c := range calls {
		if c.Status != CallNotInGraph {
			continue
		}
		if by[c.Cause] == nil {
			by[c.Cause] = map[string]int{}
		}
		by[c.Cause][c.CauseDetail]++
		total[c.Cause]++
	}
	var out []GapCause
	for cause, details := range by {
		g := GapCause{Cause: cause, Calls: total[cause]}
		for d, n := range details {
			if d != "" {
				g.Details = append(g.Details, GapCauseTop{Detail: d, Calls: n})
			}
		}
		sort.Slice(g.Details, func(i, j int) bool {
			if g.Details[i].Calls != g.Details[j].Calls {
				return g.Details[i].Calls > g.Details[j].Calls
			}
			return g.Details[i].Detail < g.Details[j].Detail
		})
		if len(g.Details) > 10 {
			g.Details = g.Details[:10]
		}
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Calls != out[j].Calls {
			return out[i].Calls > out[j].Calls
		}
		return out[i].Cause < out[j].Cause
	})
	return out
}

func classStatus(t *Trace, c Class, lookup func(string) (string, string, bool)) string {
	if lookup == nil {
		return StatusNotInInput
	}
	classSHA, artifactSHA, ok := lookup(c.Name)
	if !ok {
		return StatusNotInInput
	}
	if c.SHA256 != "" {
		if c.SHA256 == classSHA {
			return StatusMatched
		}
		return StatusDifferentVersion
	}
	if a, ok := t.Artifacts[c.Artifact]; ok && c.Artifact != "" {
		if a.SHA256 == artifactSHA {
			return StatusMatchedArtifact
		}
		return StatusDifferentVersion
	}
	return StatusNameOnly
}

// ObservedVulnerable returns the usable executed observations of methods
// mapped as vulnerable: a mapping "Class#method" matches any descriptor,
// "Class#method(desc)" only that one.
func (r Result) ObservedVulnerable(mapped []string) []MethodObservation {
	var out []MethodObservation
	for _, o := range r.Executed {
		if !Usable(o.ClassStatus) || !o.InGraph {
			continue
		}
		for _, m := range mapped {
			if o.Method == m || (!strings.Contains(m, "(") && strings.HasPrefix(o.Method, m+"(")) {
				out = append(out, o)
				break
			}
		}
	}
	return out
}

// NotInGraph returns the measured calls between usable methods that the
// static graph has no edge for.
func (r Result) NotInGraph() []CallObservation {
	var out []CallObservation
	for _, c := range r.Calls {
		if c.Status == CallNotInGraph {
			out = append(out, c)
		}
	}
	return out
}
