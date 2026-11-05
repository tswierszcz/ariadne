package report

import (
	"testing"

	"ariadne/internal/findings"
	"ariadne/internal/reach"
)

func TestJoinReasonsWithGaps(t *testing.T) {
	got := joinReasons([]findings.ReasonCode{findings.MappingPartial, findings.ModelIncomplete}, map[string]int{"reflection": 12, "callback": 1})
	want := "`MAPPING_PARTIAL`, `MODEL_INCOMPLETE` (callback 1, reflection 12)"
	if got != want {
		t.Errorf("joinReasons = %q, want %q", got, want)
	}
}

// A single instance whose reasons the advisory row already shows is not
// repeated in the per-instance table.
func TestShowPerInstance(t *testing.T) {
	row := []findings.ReasonCode{findings.MappingPartial, findings.ModelIncomplete}
	same := reach.Result{Reasons: row, Findings: []reach.InstanceFinding{{Reasons: []findings.ReasonCode{findings.ModelIncomplete}}}}
	if showPerInstance(same) {
		t.Error("single instance with the row's reasons shown")
	}
	other := reach.Result{Reasons: row, Findings: []reach.InstanceFinding{{Reasons: []findings.ReasonCode{findings.ClasspathConflict}}}}
	if !showPerInstance(other) {
		t.Error("instance with a reason not on the row hidden")
	}
}
