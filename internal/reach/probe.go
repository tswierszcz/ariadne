package reach

import (
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
)

// GapProbeVersion versions the gap-rule probe (ROADMAP B0 b0-4).
const GapProbeVersion = "b0-4"

// GapProbe measures, for one NOT_REACHABLE finding, which gaps a wider or
// a narrower model-gap rule would count (ROADMAP B0 b0-4). It is a
// measurement only: no verdict, reason code or model_gaps entry depends
// on it, and v1 does not carry it.
//
// The current rule (contract §4) counts Unbounded + Bounded. The probe
// adds two gap kinds the rule does not take, each bounded to the classes
// it could dispatch into, so it can be asked whether one of them lands in
// the finding's instance:
//   - Lambda: a reachable functional-interface call whose lambda or method
//     reference was not resolved (samcall gaps), relevant when a class of
//     the instance creates a lambda of that interface;
//   - SkippedReceiver: a reachable virtual/interface call on a java./
//     javax./jdk./sun. type that CHA never expands, relevant when the
//     expansion CHA skipped would select a body in the instance.
//     SkippedReceiverObject is the part of it on java.lang.Object methods
//     (toString, equals, hashCode, …).
type GapProbe struct {
	Unbounded int `json:"unbounded"`
	Bounded   int `json:"bounded"`
	// BoundedByKind splits Bounded by site kind (GapSite.Kind).
	BoundedByKind         map[string]int `json:"bounded_by_kind,omitempty"`
	Lambda                int            `json:"lambda"`
	SkippedReceiver       int            `json:"skipped_receiver"`
	SkippedReceiverObject int            `json:"skipped_receiver_object"`
}

// Probe rule variants, from the current rule to the widest scoped one.
const (
	ProbeCurrent        = "current"                      // Unbounded + Bounded (the rule as it is)
	ProbeScoped         = "scoped"                       // Bounded only: drop sites with no candidate set
	ProbeScopedLambda   = "scoped+lambda"                // + Lambda
	ProbeScopedSkipped  = "scoped+lambda+skipped"        // + SkippedReceiver
	ProbeScopedNoObject = "scoped+lambda+skipped-object" // + SkippedReceiver without java.lang.Object methods
)

// ProbeVariants lists the variants in reporting order.
var ProbeVariants = []string{ProbeCurrent, ProbeScoped, ProbeScopedLambda, ProbeScopedNoObject, ProbeScopedSkipped}

// Flags reports whether each variant would mark the finding.
func (p GapProbe) Flags() map[string]bool {
	return map[string]bool{
		ProbeCurrent:        p.Unbounded+p.Bounded > 0,
		ProbeScoped:         p.Bounded > 0,
		ProbeScopedLambda:   p.Bounded+p.Lambda > 0,
		ProbeScopedNoObject: p.Bounded+p.Lambda+p.SkippedReceiver-p.SkippedReceiverObject > 0,
		ProbeScopedSkipped:  p.Bounded+p.Lambda+p.SkippedReceiver > 0,
	}
}

// skippedTarget is one JDK method called virtually by reachable code,
// with the instances CHA would have dispatched into had it expanded it.
type skippedTarget struct {
	calls     int // reachable caller → target edges
	object    bool
	instances map[string]bool
}

// probeState holds the probe's inputs (SetProbe).
type probeState struct {
	lambda  []gap
	skipped []skippedTarget
}

// SetProbe records the probe's inputs (ROADMAP B0 b0-4): lambda are the
// unresolved functional-interface calls as GapSites of kind "lambda"
// (Owners: the classes creating a lambda of the called interface), skip
// the CHA skip rule of the graph. Skipped-receiver calls are read from the
// graph: every direct virtual/interface edge from a reachable method to a
// skipped type, with the targets CHA would have added without the skip.
// Call after SetComponents; without components the probe counts only
// unbounded sites.
func (e *Engine) SetProbe(lambda []GapSite, skip callgraph.SkipFunc) {
	if skip == nil {
		skip = callgraph.DefaultSkip
	}
	p := &probeState{lambda: e.reachableGaps(lambda)}
	type targetInfo struct {
		idx int
		ok  bool
	}
	byTarget := map[uint32]targetInfo{}
	for id := 0; id < e.g.NodeCount(); id++ {
		if !e.full.visited[id] {
			continue
		}
		for _, ed := range e.g.Edges(uint32(id)) {
			if ed.Basis != callgraph.BasisDirect || (ed.Kind != classfile.CallVirtual && ed.Kind != classfile.CallInterface) {
				continue
			}
			ti, seen := byTarget[ed.To]
			if !seen {
				ti = targetInfo{idx: -1}
				if st, ok := e.skippedTarget(ed.To, ed.Kind, skip); ok {
					ti = targetInfo{idx: len(p.skipped), ok: true}
					p.skipped = append(p.skipped, st)
				}
				byTarget[ed.To] = ti
			}
			if ti.ok {
				p.skipped[ti.idx].calls++
			}
		}
	}
	e.probe = p
}

// skippedTarget returns, for a target node on a skipped type, the
// instances of the bodies CHA's expansion would select were the type not
// skipped (subtypes that are themselves skipped stay out, as in CHA).
func (e *Engine) skippedTarget(to uint32, kind classfile.CallKind, skip callgraph.SkipFunc) (skippedTarget, bool) {
	key := e.g.Key(to)
	owner, rest, _ := strings.Cut(key, "#")
	if !skip(owner) {
		return skippedTarget{}, false
	}
	name, desc := rest, ""
	if i := strings.IndexByte(rest, '('); i >= 0 {
		name, desc = rest[:i], rest[i:]
	}
	st := skippedTarget{object: owner == "java.lang.Object", instances: map[string]bool{}}
	if e.components == nil {
		return st, true
	}
	expand := callgraph.CHA{Skip: func(fqcn string) bool { return fqcn != owner && skip(fqcn) }}
	for _, decl := range expand.DispatchTargets(e.h, classfile.Call{Kind: kind, Owner: owner, Name: name, Desc: desc}) {
		if ci, ok := e.h.Lookup(decl); ok {
			if inst := e.components.InstanceForClass(ci); inst != "" {
				st.instances[inst] = true
			}
		}
	}
	return st, true
}

// probeFor measures one finding's instances.
func (e *Engine) probeFor(instances map[string]bool) GapProbe {
	var p GapProbe
	inInstance := func(owners map[string]bool) bool {
		for inst := range owners {
			if instances[inst] {
				return true
			}
		}
		return false
	}
	for _, g := range e.gaps {
		switch {
		case g.unbounded:
			p.Unbounded++
		case inInstance(g.owners):
			p.Bounded++
			if p.BoundedByKind == nil {
				p.BoundedByKind = map[string]int{}
			}
			p.BoundedByKind[g.site.Kind]++
		}
	}
	if e.probe == nil {
		return p
	}
	for _, g := range e.probe.lambda {
		if inInstance(g.owners) {
			p.Lambda++
		}
	}
	for _, st := range e.probe.skipped {
		if inInstance(st.instances) {
			p.SkippedReceiver += st.calls
			if st.object {
				p.SkippedReceiverObject += st.calls
			}
		}
	}
	return p
}

// ApplyProbe sets GapProbe on every NOT_REACHABLE per-instance finding of
// results (ROADMAP B0 b0-4). It changes nothing else. A no-op before
// SetProbe.
func (e *Engine) ApplyProbe(results []Result) {
	if e.probe == nil {
		return
	}
	for i := range results {
		for j := range results[i].Findings {
			f := &results[i].Findings[j]
			if f.Status != StatusNotReachable {
				continue
			}
			p := e.probeFor(map[string]bool{f.Instance: true})
			f.GapProbe = &p
		}
	}
}

// ProbeInputs summarizes the probe's inputs for the log: reachable
// unresolved lambda calls, and skipped JDK targets with their reachable
// calls.
func (e *Engine) ProbeInputs() (lambda, skippedTargets, skippedCalls int) {
	if e.probe == nil {
		return 0, 0, 0
	}
	for _, st := range e.probe.skipped {
		skippedCalls += st.calls
	}
	return len(e.probe.lambda), len(e.probe.skipped), skippedCalls
}

// ProbeSummary totals the probe over findings (ROADMAP B0 b0-4).
type ProbeSummary struct {
	Version string `json:"version"`
	// NotReachable counts the NOT_REACHABLE findings measured.
	NotReachable int `json:"not_reachable"`
	// Marked: per variant, how many of them it would mark.
	Marked map[string]int `json:"marked"`
}

// ProbeRow is one measured finding, for the report.
type ProbeRow struct {
	CVE, GAV string
	Probe    GapProbe
}

// ProbeRows lists the measured findings in result order.
func ProbeRows(results []Result) []ProbeRow {
	var out []ProbeRow
	for _, r := range results {
		for _, f := range r.Findings {
			if f.GapProbe != nil {
				out = append(out, ProbeRow{CVE: r.CVE, GAV: f.GAV, Probe: *f.GapProbe})
			}
		}
	}
	return out
}

// Marks lists the variants that would mark the finding, in reporting
// order.
func (p GapProbe) Marks() []string {
	flags := p.Flags()
	var out []string
	for _, v := range ProbeVariants {
		if flags[v] {
			out = append(out, v)
		}
	}
	return out
}

// SummarizeProbe totals the findings' probes.
func SummarizeProbe(results []Result) *ProbeSummary {
	s := &ProbeSummary{Version: GapProbeVersion, Marked: map[string]int{}}
	measured := false
	for _, r := range results {
		for _, f := range r.Findings {
			if f.GapProbe == nil {
				continue
			}
			measured = true
			s.NotReachable++
			for v, on := range f.GapProbe.Flags() {
				if on {
					s.Marked[v]++
				}
			}
		}
	}
	if !measured {
		return nil
	}
	for _, v := range ProbeVariants {
		if _, ok := s.Marked[v]; !ok {
			s.Marked[v] = 0
		}
	}
	return s
}
