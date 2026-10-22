package callback

import "ariadne/internal/hierarchy"

// Outcome says what FindRegistrationSites could determine about one
// call — ASYNC_CALLBACKS A0's own "diagnostykę ... nieobsługiwanego
// API". OutcomeUnknownReceiver and OutcomeMissingTarget (A0's other two
// named diagnostics) need the callback argument's concrete type
// resolved first, which is ASYNC_CALLBACKS A1's own scope, not A0's —
// they are not produced here, so as not to claim a resolution step this
// phase does not perform.
type Outcome string

const (
	// OutcomeCataloged: the call matches DefaultCatalog exactly.
	OutcomeCataloged Outcome = "cataloged"
	// OutcomeUnsupportedAPI: the call's owner+name matches a cataloged
	// API and its descriptor mentions a known callback interface
	// (Runnable/Callable) as a parameter, but this specific overload is
	// not in DefaultCatalog — a real registration ariadne simply
	// doesn't model yet, reported rather than silently skipped. See
	// LooksLikeUncatalogedRegistration's own doc comment for why the
	// descriptor check matters (Thread has overloads with no callback
	// argument at all).
	OutcomeUnsupportedAPI Outcome = "unsupported-api"
)

// RegistrationSite is one call found in the analyzed hierarchy (app or
// dependency) whose owner+name matches a cataloged callback-registration
// API. It is not a call-graph edge and does not yet say what the
// callback actually is — see this package's own doc comment for why
// that is A1/A2's job.
type RegistrationSite struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Model is the matched catalog entry — the zero value when Outcome
	// is OutcomeUnsupportedAPI (nothing matched exactly).
	Model   RegistrationModel
	Outcome Outcome
}

// FindRegistrationSites scans every method's already-extracted Calls
// (hierarchy.Build's own ExtractCalls pass — no bytecode re-read needed,
// unlike SERVICE_LOADING S1's argument resolution, since this phase
// resolves nothing about the callback's identity yet) for a call
// matching, or nearly matching, DefaultCatalog.
func FindRegistrationSites(h *hierarchy.Hierarchy) []RegistrationSite {
	var sites []RegistrationSite
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if model, ok := Lookup(c.Owner, c.Name, c.Desc); ok {
					sites = append(sites, RegistrationSite{
						CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor,
						PC: c.PC, Model: model, Outcome: OutcomeCataloged,
					})
				} else if LooksLikeUncatalogedRegistration(c.Owner, c.Name, c.Desc) {
					sites = append(sites, RegistrationSite{
						CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor,
						PC: c.PC, Outcome: OutcomeUnsupportedAPI,
					})
				}
			}
		}
	}
	return sites
}
