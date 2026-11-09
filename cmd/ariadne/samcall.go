package main

import (
	"fmt"

	"ariadne/internal/callgraph"
	"ariadne/internal/report"
	"ariadne/internal/samcall"
)

// samSiteSummary renders DYNAMIC_LINKAGE L1's site counts.
func samSiteSummary(s samcall.Stats) string {
	c, e := samcall.KindCall, samcall.KindEscape
	gaps := "none"
	if len(s.Gaps) > 0 {
		gaps = report.GapCounts(s.Gaps)
	}
	return fmt.Sprintf("%d SAM call sites (%d resolved, %d partly, %d unresolved, %d with only ordinary objects/null; gaps by cause: %s), "+
		"%d sites handing a functional value outside the input (%d resolved, %d partly)",
		s.Sites[c], s.Resolved[c], s.Partial[c], s.Unresolved[c], s.NoLambda[c], gaps, s.Sites[e], s.Resolved[e], s.Partial[e])
}

// samSummary adds the edges to samSiteSummary, for the report header.
func samSummary(s samcall.Stats, applied callgraph.SAMStats) string {
	return fmt.Sprintf("%s; %d edges added, %d already in the graph, %d left out (no method creating the value reachable from the roots)",
		samSiteSummary(s), applied.Added, applied.Covered, applied.Gated)
}
