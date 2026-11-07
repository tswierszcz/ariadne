package deserialize

import (
	"fmt"

	"ariadne/internal/flowsum"
	"ariadne/internal/taint"
)

// DATA_FLOW D4's pilot sinks: the data each entry reads — the stream
// (readObject's receiver) for Java serialization, fromXML's first
// argument for XStream — assessed by internal/taint, with the entry's
// own Z2 policy as the only guard that matters for it. Report-only.

// Sink rule IDs.
const (
	SinkJavaStream = "sink:deser-java-stream"
	SinkXStreamXML = "sink:xstream-fromxml"
)

func sinkOf(s *Site) taint.Sink {
	sk := taint.Sink{Method: flowsum.MethodKey{Owner: s.CallerFQCN, Name: s.CallerName, Desc: s.CallerDesc}, PC: s.PC,
		Data: "receiver", Rule: SinkJavaStream}
	if s.Format == FormatXStream || s.Model == "xstream:fromxml" {
		sk.Data, sk.Rule = "arg:0", SinkXStreamXML
	}
	sk.Policy = policyGuard(s)
	return sk
}

// policyGuard states the entry's Z2 policy as a guard.
func policyGuard(s *Site) taint.PolicyGuard {
	switch s.Filter {
	case FilterDefaultAllowlist:
		return taint.PolicyGuard{State: taint.GuardEffective, Detail: "XStream default allowlist (no class of the input admitted)"}
	case FilterEvaluated:
		switch {
		case s.Candidates == 0 || s.Rejected >= s.Candidates:
			return taint.PolicyGuard{State: taint.GuardEffective, Detail: fmt.Sprintf("evaluated policy rejects every candidate class (%d): %s", s.Candidates, s.FilterDetail)}
		default:
			return taint.PolicyGuard{State: taint.GuardPartial, Detail: fmt.Sprintf("evaluated policy rejects %d of %d candidate classes: %s", s.Rejected, s.Candidates, s.FilterDetail)}
		}
	case FilterUnknown, FilterVersionUnknown:
		return taint.PolicyGuard{State: taint.GuardUnknown, Detail: s.Filter + ": " + s.FilterDetail}
	}
	return taint.PolicyGuard{State: taint.GuardNone}
}

// TaintRoots are the methods the taint search needs summaries of.
func TaintRoots(sites []Site, a *taint.Analyzer) []flowsum.MethodKey {
	return a.Closure(SummaryRoots(sites))
}

// AttachTaint assesses every entry. An entry in library code is assessed
// only when application code calls its method (the callers searched are
// the application's); otherwise it is not_assessed.
func AttachTaint(sites []Site, a *taint.Analyzer, e *flowsum.Engine) {
	for i := range sites {
		s := &sites[i]
		sk := sinkOf(s)
		if !s.IsApp && len(a.Callers(sk.Method)) == 0 {
			s.Taint = &taint.Result{State: taint.NotAssessed, Catalog: taint.CatalogVersion, Rule: sk.Rule,
				Reason: "entry in library code not called from application code (only application callers are searched)"}
			continue
		}
		r := a.Analyze(e, sk)
		s.Taint = &r
	}
}

// TaintStateCounts counts entries by taint state ("" entries skipped).
func TaintStateCounts(sites []Site) map[string]int {
	out := map[string]int{}
	for _, s := range sites {
		if s.Taint != nil {
			out[s.Taint.State]++
		}
	}
	return out
}
