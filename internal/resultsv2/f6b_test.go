package resultsv2

import (
	"testing"

	"ariadne/internal/cves"
)

// TestCandidateApplicability (CVE:F6b): a version-candidates instance is
// "affected" only when every candidate version is; otherwise which
// version is present decides, and that is not known.
func TestCandidateApplicability(t *testing.T) {
	c := cves.CVE{Affected: []cves.Affected{{Package: "g:a", Version: "1.0"}, {Package: "g:a", Version: "1.1"}, {Package: "other:x", Version: "2.0"}}}
	for _, tc := range []struct {
		versions []string
		want     string
	}{
		{[]string{"1.0", "1.1"}, "affected"},
		{[]string{"1.0", "2.0"}, "undetermined"},
		{[]string{"2.0"}, "undetermined"},
	} {
		if got := candidateApplicability(c, "g:a", tc.versions); got != tc.want {
			t.Errorf("%v = %s, want %s", tc.versions, got, tc.want)
		}
	}
}
