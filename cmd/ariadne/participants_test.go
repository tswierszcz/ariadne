package main

import (
	"testing"

	"ariadne/internal/analysis"
	"ariadne/internal/models"
	"ariadne/internal/serviceload"
	"ariadne/internal/servletdispatch"
)

// TestParticipantCatalog (ROADMAP B3 b3-3): IDs are unique and versioned,
// and the models the CLI registers carry IDs the catalog names, so a
// profile can disable each of them.
func TestParticipantCatalog(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range participantCatalog() {
		if seen[p.ID] || p.Version == "" || (p.Kind != analysis.KindModel && p.Kind != analysis.KindGate) {
			t.Errorf("bad participant %+v", p)
		}
		seen[p.ID] = true
	}
	r := models.NewRegistry()
	serviceload.RegisterModel(r)
	servletdispatch.Register(r)
	for _, id := range r.IDs() {
		if !seen[id] {
			t.Errorf("registered model %s is not in the participant catalog", id)
		}
	}
	if len(r.IDs()) < 3 {
		t.Errorf("registered models = %v", r.IDs())
	}
}
