package profile

import "ariadne/internal/hierarchy"

// Condition is a tri-state fact FRAMEWORKS F0 onward uses instead of a
// plain bool: whether some specific configuration/condition is active
// in the analyzed app. "Unknown" is a first-class result, not an
// implementation gap papered over as false — F0's own requirement
// ("warunek ma stan true/false/unknown") and consistent with
// ROADMAP.MD B0's "nieznany cel... może uniemożliwić negatywne
// rozstrzygnięcie": an unsupported or unrecognized condition must never
// silently read as "not active".
type Condition string

const (
	ConditionTrue    Condition = "true"
	ConditionFalse   Condition = "false"
	ConditionUnknown Condition = "unknown"
)

// Detection is FRAMEWORKS F0's first separated fact: whether one
// catalog profile's own Detect markers are present anywhere in the
// analyzed app's dependency closure — independent of whether that
// profile ends up contributing to the actually active merged Profile
// (Select already merges every matched, non-generic profile together;
// before F0, nothing recorded which of several simultaneously-present
// frameworks actually matched, only the merged result's own Name
// string, e.g. "spring+jakarta" — not inspectable as individual facts).
type Detection struct {
	Name    string
	Present bool
}

// DetectAll reports, for every non-generic profile in catalog,
// Detection{Present: true} when its own Detect markers are found in h.
// Always returns one entry per non-generic catalog profile, in the
// catalog's own fixed order — so a caller can always tell "checked,
// absent" apart from "never checked".
func DetectAll(catalog Catalog, h *hierarchy.Hierarchy) []Detection {
	out := make([]Detection, 0, len(catalog.order))
	for _, name := range catalog.order {
		if name == "generic" {
			continue
		}
		p := catalog.profiles[name]
		out = append(out, Detection{Name: name, Present: classPresent(h, p.Detect)})
	}
	return out
}

// FrameworkScenario is FRAMEWORKS F0's own proposed, not-yet-wired
// contract for a named, versioned "shape" of framework usage (e.g.
// "spring-mvc-rest" vs "spring-data-jpa" within the same detected
// Spring presence) — distinct from a Detection, which only says a
// framework's marker classes exist somewhere. F1/F2 are what actually
// construct these; F0 only settles the shape, same reasoning DATA_FLOW
// D0 settled Value's shape before D1/D2 produced any. Nothing in
// cmd/ariadne constructs a FrameworkScenario yet.
type FrameworkScenario struct {
	Name      string
	Framework string // the Detection.Name this scenario builds on
	Active    Condition
}

// BeanDefinition is FRAMEWORKS F0's own proposed contract for one DI
// component — a class (or @Bean factory method) a container constructs
// and wires, with its declared qualifier/primary status. FRAMEWORKS F2's
// own scope to construct; not built here.
type BeanDefinition struct {
	FQCN      string
	Qualifier string
	Primary   bool
}

// ConditionalBinding is FRAMEWORKS F0's own proposed contract for one
// configuration fact gating a bean, route, or profile (e.g. Spring's
// @ConditionalOnProperty, a `@Profile` tag) — Active is a Condition,
// never a bare bool, so an unrecognized or unsupported condition
// expression reads as ConditionUnknown rather than a guessed true/false.
type ConditionalBinding struct {
	Description string
	Active      Condition
}

// ExternalEntry is FRAMEWORKS F0's own proposed contract for an entry
// point reached through framework/container dispatch before any
// application code runs (PLAN.md's own "roots"), carrying its own basis
// and activation condition instead of being silently folded into the
// same Roots.Annotations/Overrides list every match already uses.
type ExternalEntry struct {
	FQCN, Method string
	Basis        string
	Active       Condition
}
