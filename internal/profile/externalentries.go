// FRAMEWORKS F2's own "katalog punktów wejścia" slice: a single,
// labeled inventory of every entry point already known to be reached
// through framework/container dispatch before any application code
// runs — Spring MVC HTTP mappings (F1's own Routes) and web.xml-
// declared servlet/filter/listener lifecycle methods (F2's own
// WebXMLRoots) — each carrying its own Basis (what kind of dispatch)
// and Active condition (F1's own ProfileConditions tri-state, cross-
// referenced rather than reinvented).
//
// Deliberately only a labeling/reporting pass over facts these other
// functions already compute correctly: it adds no new root-selection
// rule (`-roots profile`'s own entry-point set is unchanged) and no
// new call-graph edge — FRAMEWORKS.MD's own scenario 4 warns against
// fabricating a call from an external parser/dispatcher into
// application code; this only shows that such a dispatch exists ahead
// of the listed method, never a traced call into it.
package profile

import "ariadne/internal/hierarchy"

// EntryBasis names which kind of framework/container dispatch an
// ExternalEntry is known through.
type EntryBasis string

const (
	BasisHTTPMapping EntryBasis = "http-mapping"
	BasisServlet     EntryBasis = "servlet"
	BasisFilter      EntryBasis = "filter"
	BasisListener    EntryBasis = "listener"
)

// ExternalEntry is one entry point reached through framework/container
// dispatch before any application code runs, with its own basis and
// activation condition.
type ExternalEntry struct {
	FQCN, Method string
	Basis        EntryBasis
	// Active is F1's own ProfileConditions tri-state for this
	// FQCN/Method: ConditionTrue when nothing gates it (or an active
	// `@Profile` match is known), ConditionFalse for a known inactive
	// `@Profile`, ConditionUnknown when a `@Profile` gates it but
	// -active-profiles was never given — never a guessed true/false.
	Active Condition
}

// ExternalEntries builds the catalog from F1's Routes and F2's
// WebXMLRoots, resolving each entry's own Active condition against
// ProfileConditions(h, activeProfiles). servletClasses/filterClasses/
// listenerClasses are the same deduplicated FQCNs WebXMLRoots itself
// takes.
func ExternalEntries(h *hierarchy.Hierarchy, activeProfiles []string, servletClasses, filterClasses, listenerClasses []string) []ExternalEntry {
	conditions := ProfileConditions(h, activeProfiles)

	var out []ExternalEntry
	for _, r := range Routes(h) {
		out = append(out, ExternalEntry{FQCN: r.FQCN, Method: r.Method, Basis: BasisHTTPMapping, Active: activeCondition(conditions, r.FQCN, r.Method)})
	}
	for _, r := range WebXMLRoots(h, servletClasses, filterClasses, listenerClasses) {
		out = append(out, ExternalEntry{FQCN: r.FQCN, Method: r.Method, Basis: EntryBasis(r.Kind), Active: activeCondition(conditions, r.FQCN, r.Method)})
	}
	return out
}

// activeCondition is fqcn#method's own activation condition: the
// method-level `@Profile` if one is declared on it, else the class-
// level one, else ConditionTrue (nothing gates it at all — not
// "unknown", since there is no unresolved condition to be unsure
// about).
func activeCondition(conditions []ProfileCondition, fqcn, method string) Condition {
	for _, c := range conditions {
		if c.FQCN == fqcn && c.Method == method {
			return c.Active
		}
	}
	for _, c := range conditions {
		if c.FQCN == fqcn && c.Method == "" {
			return c.Active
		}
	}
	return ConditionTrue
}
