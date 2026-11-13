package main

import (
	"ariadne/internal/analysis"
	"ariadne/internal/jdbcdispatch"
	"ariadne/internal/proxycatalog"
	"ariadne/internal/samcall"
	"ariadne/internal/serviceload"
	"ariadne/internal/servletdispatch"
	"ariadne/internal/spidispatch"
)

// Gate IDs of the coordinator's steps.
const (
	samGateID     = "dynamic-linkage:sam-calls"
	servletGateID = "servlet-container:objects"
)

// participantCatalog is every fixed-point participant the analyzer has
// (ROADMAP B3 b3-3): the registered models and the creator gates, with
// the version of their rules. A profile's models.disable names IDs from
// this list; the catalog does not depend on the input, so a profile is
// valid or not regardless of what an application happens to contain.
func participantCatalog() []analysis.Participant {
	return []analysis.Participant{
		{ID: serviceload.LoadModelID, Version: serviceload.ModelVersion, Kind: analysis.KindModel},
		{ID: serviceload.ConsumptionModelID, Version: serviceload.ModelVersion, Kind: analysis.KindModel},
		{ID: serviceload.DispatchModelID, Version: serviceload.ModelVersion, Kind: analysis.KindModel},
		{ID: servletdispatch.ModelID, Version: servletdispatch.Version, Kind: analysis.KindModel},
		{ID: spidispatch.ModelID, Version: spidispatch.Version, Kind: analysis.KindModel},
		{ID: samGateID, Version: samcall.Version, Kind: analysis.KindGate},
		{ID: servletGateID, Version: servletdispatch.Version, Kind: analysis.KindGate},
		{ID: jdbcdispatch.ModelID, Version: jdbcdispatch.Version, Kind: analysis.KindGate},
		{ID: proxycatalog.ModelID, Version: proxycatalog.Version, Kind: analysis.KindGate},
	}
}
