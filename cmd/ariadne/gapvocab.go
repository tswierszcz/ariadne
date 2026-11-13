package main

import (
	"strings"

	"ariadne/internal/gaps"
	"ariadne/internal/reach"
	"ariadne/internal/runtimeevents"
)

// runGaps is the run's gaps in the shared vocabulary (ROADMAP B0 b0-3):
// the model-gap rule's sites (with how many are reachable), the
// created-type catalog's unknown allocations, the SAM-call gaps, the B3
// coordinator's incomplete facts (by model id) and, with a trace, the
// causes of measured calls missing from the graph.
func runGaps(sites []reach.GapSite, reachableByKind, allocGaps, lambdaGaps map[string]int, coordinatorReasons []string, rt *runtimeevents.Result) []gaps.Gap {
	total := map[string]int{}
	for _, s := range sites {
		total[s.Kind]++
	}
	var out []gaps.Gap
	for _, g := range gaps.FromCounts(gaps.SourceSite, total) {
		r := reachableByKind[g.Detail]
		g.Reachable = &r
		out = append(out, g)
	}
	out = append(out, gaps.FromCounts(gaps.SourceAllocation, allocGaps)...)
	out = append(out, gaps.FromCounts(gaps.SourceLambda, lambdaGaps)...)
	byModel := map[string]int{}
	for _, r := range coordinatorReasons {
		id, _, _ := strings.Cut(r, ": ")
		byModel[id]++
	}
	out = append(out, gaps.FromCounts(gaps.SourceCoordinator, byModel)...)
	if rt != nil {
		causes := map[string]int{}
		for _, c := range rt.GapCauses {
			causes[c.Cause] += c.Calls
		}
		out = append(out, gaps.FromCounts(gaps.SourceRuntime, causes)...)
	}
	return gaps.Sort(out)
}
