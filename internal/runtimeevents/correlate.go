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
}

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
			}
		}
		r.CallCounts[o.Status]++
		r.Calls = append(r.Calls, o)
	}
	sort.Slice(r.Calls, func(i, j int) bool { return r.Calls[i].Evidence < r.Calls[j].Evidence })
	return r
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
