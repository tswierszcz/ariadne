package reach

import (
	"strings"
	"testing"

	"ariadne/internal/findings"
	"ariadne/internal/vulnmap"
)

// TestInitializerReasons (2026-11-17): a diff mapping that names only
// static initializers carries MAPPING_INITIALIZER_ONLY on the row, never
// on an instance finding; a redirected mapping (initializers replaced by
// entry points) carries the entry-point qualifier and the redirect note.
func TestInitializerReasons(t *testing.T) {
	e, _, _ := buildEngine(t)

	clinit := vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#<clinit>"}, Source: vulnmap.SourceDiff}
	r := analyzeEntry(t, e, clinit)
	if !hasReason(r.Reasons, findings.MappingInitializerOnly) {
		t.Errorf("<clinit> only: %s %v, want MAPPING_INITIALIZER_ONLY", r.Status, r.Reasons)
	}
	for _, f := range r.Findings {
		if hasReason(f.Reasons, findings.MappingInitializerOnly) {
			t.Errorf("instance finding carries the mapping qualifier: %v", f.Reasons)
		}
	}

	method := vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#unreachable"}, Source: vulnmap.SourceDiff}
	if r := analyzeEntry(t, e, method); hasReason(r.Reasons, findings.MappingInitializerOnly) {
		t.Errorf("method mapping: %v, want no MAPPING_INITIALIZER_ONLY", r.Reasons)
	}

	redirected := method
	redirected.Role = vulnmap.RoleEntrypoint
	redirected.InitializerRedirect = &vulnmap.InitializerRedirect{Initializers: []string{"p.Blocklist#<clinit>"}, Reason: "blocklist only"}
	r = analyzeEntry(t, e, redirected)
	if r.Status != StatusNotReachable || !hasReason(r.Reasons, findings.MappingEntrypointOnly) ||
		hasReason(r.Reasons, findings.MappingInitializerOnly) || !strings.Contains(r.Note, "blocklist only") || !strings.Contains(r.Note, "p.Blocklist#<clinit>") {
		t.Errorf("redirect: %s %v note=%q", r.Status, r.Reasons, r.Note)
	}
}
