package reach

import (
	"reflect"
	"testing"

	"ariadne/internal/findings"
	"ariadne/internal/vulnmap"
)

// TestFixDiffReasons (CVE:F5): a diff mapping read on the post-fix side
// only says so; one read on both sides does not; a CVE whose fix commits
// could not be read is UNKNOWN with SOURCE_UNAVAILABLE.
func TestFixDiffReasons(t *testing.T) {
	e, _, _ := buildEngine(t)
	notReached := []string{"reachfixture.VulnerableLib#unreachable"}

	both := vulnmap.Entry{Methods: notReached, Source: vulnmap.SourceDiff,
		FixCommits: []vulnmap.FixCommit{{Base: vulnmap.BaseParent, Parent: "p"}}}
	r := analyzeEntry(t, e, both)
	if r.Status != StatusNotReachable || hasReason(r.Reasons, findings.MappingNewSideOnly) {
		t.Errorf("both sides: %s %v", r.Status, r.Reasons)
	}

	newOnly := both
	newOnly.FixCommits = []vulnmap.FixCommit{{Base: vulnmap.BaseNewSideOnly, BaseReason: "merge commit"}}
	r = analyzeEntry(t, e, newOnly)
	if r.Status != StatusNotReachable || !hasReason(r.Reasons, findings.MappingNewSideOnly) {
		t.Errorf("new side only: %s %v, want NOT_REACHABLE qualified by MAPPING_NEW_SIDE_ONLY", r.Status, r.Reasons)
	}
	for _, f := range r.Findings {
		if hasReason(f.Reasons, findings.MappingNewSideOnly) {
			t.Errorf("instance finding carries a mapping qualifier: %v", f.Reasons)
		}
	}

	r = analyzeEntry(t, e, vulnmap.Entry{SourceError: "o/r@abc: fetching diff: 500"})
	if r.Status != StatusUnknown || !reflect.DeepEqual(r.Reasons, []findings.ReasonCode{findings.SourceUnavailable}) {
		t.Errorf("unreadable fix commits: %s %v, want UNKNOWN [SOURCE_UNAVAILABLE]", r.Status, r.Reasons)
	}
}

func hasReason(rs []findings.ReasonCode, c findings.ReasonCode) bool {
	for _, r := range rs {
		if r == c {
			return true
		}
	}
	return false
}
