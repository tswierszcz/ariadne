// Package models defines ROADMAP B3's registered model, exact call-site,
// fact, and evaluator contracts. internal/analysis consumes these to
// coordinate model-derived call edges with the call graph. Model
// semantics remain supplied by evaluator implementations; this package
// does not load executable code from profiles.
//
// Package-dependency decision (ROADMAP.MD's own explicit prerequisite,
// "ustalić kierunek zależności pakietów przed refaktorem, aby uniknąć
// cyklu profile → callgraph → models → profile"): this package may
// depend only on leaf domain packages that have no dependency back on
// internal/profile or internal/callgraph — internal/input is such a
// leaf. It must never import internal/profile or internal/callgraph.
// internal/analysis depends on both this package and internal/callgraph;
// internal/profile may depend on internal/models (to select which
// model IDs apply to the active profile) but never the other way, and
// never on internal/analysis — profile stays "a configuration/model-
// selection provider, not a second coordinator" (ROADMAP.MD's own
// words). That keeps the dependency graph a DAG: models ← analysis →
// callgraph, and profile → models, with nothing pointing back into
// profile.
package models

import (
	"ariadne/internal/classfile"
	"ariadne/internal/input"
	"sort"
)

// Signature identifies one library API call a Model knows how to
// model — the exact Owner/Name/Desc a Methodref names in the bytecode
// itself, the same "owner as named in bytecode, before resolution"
// convention valueflow.CallResultRef/FieldRef already use elsewhere in
// ariadne. Never a bare method name: B3's own explicit requirement is
// "przypinać do konkretnego wywołania/obiektu, nie do wszystkich
// metod o tej samej nazwie w całej aplikacji" — matching by Signature
// alone already excludes every other class's own same-named method, by
// construction, not by a separate check a caller could forget.
type Signature struct {
	Owner, Name, Desc string
}

// VersionRange is the library version span a Model's own semantics
// were actually verified against — Min/Max inclusive, "" meaning
// unbounded on that side. This type only records what a Model claims;
// it deliberately does not yet define how two version strings compare
// (real semver — pre-release tags, build metadata, non-numeric
// segments — is its own unsettled scope, not papered over here with a
// naive comparator pretending to be complete). Registry.Lookup's own
// doc comment explains what happens when a comparison can't be made.
type VersionRange struct {
	Min, Max string
}

// Model is one library API's own, independently verified modeled
// behavior — B3's own "rejestr modeli z identyfikatorem, obsługiwanymi
// sygnaturami, zakresem wersji i źródłem semantyki", one registry
// entry.
type Model struct {
	// ID is this model's own stable identifier (e.g.
	// "spring-core:beanutils-copyproperties") — versioned independently
	// of the artifact/version it targets, so correcting this model's
	// own semantics later doesn't need a new ID.
	ID string
	// Artifact is the Maven "group:artifact" this model's semantics
	// were written against (input.GAV's own PackageName() shape) — a
	// Signature's Owner is a class, never an artifact id, so matching
	// against the declaring class's own resolved GAV is Lookup's job,
	// not this field's. "" means this model applies regardless of
	// artifact (e.g. a JDK-only API with no Maven coordinate of its
	// own) — never a guessed artifact name.
	Artifact string
	Versions VersionRange
	// Signatures is every call shape this one Model knows how to
	// model — more than one when an overload or a closely related
	// method shares the same semantics (e.g. getMethod/
	// getDeclaredMethod), never a wildcard across unrelated methods.
	Signatures []Signature
	// Source documents where this model's semantics actually came from
	// — a javadoc URL, a cited source read, a WALA/SootUp reference —
	// B3's own explicit "źródło semantyki", never left implicit or
	// asserted without a citation.
	Source string
}

// Fact is the shape an evaluator reports for one matched call site — B3's own
// "budżety i stan incomplete z przyczyną": a model that could not
// fully resolve something (a budget exceeded, an argument this
// package's own value domain couldn't prove) reports why, rather than
// silently guessing or silently doing nothing.
type Fact struct {
	ModelID string
	// Caller and PC identify the exact bytecode call site that produced
	// this fact. The coordinator fills these from the evaluated site.
	Caller Signature
	PC     int
	// Targets are exact method identities that may be invoked from this
	// site according to the model; they do not imply runtime execution.
	Targets []Target
	// Incomplete is true when this Model could not fully resolve this
	// call site — Reason is then required, non-empty, and specific
	// (ROADMAP.MD's own B0: "Każdy limit zbioru, głębokości lub czasu
	// powoduje poszerzenie do nieznanej wartości z diagnostyką. Nie
	// wolno po prostu odrzucać nadmiarowych kandydatów.").
	Incomplete bool
	Reason     string
}

// Target is an exact method target and the call-graph edge kind needed
// to represent the relationship. Zero Kind defaults to a modeled call.
type Target struct {
	Method Signature
	Kind   classfile.CallKind
}

// CallSite is an exact call instruction offered to a model evaluator.
type CallSite struct {
	Caller Signature
	Call   Signature
	PC     int
	GAV    *input.GAV
}

// Evaluator applies one registered model at one matched call site.
// Unresolved relevant cases should be returned as incomplete Facts.
type Evaluator interface {
	Evaluate(model *Model, site CallSite) Fact
}

// Registry is a Model-by-Signature lookup — the one place a future
// coordinator asks "does any registered model know this call" without
// scanning every Model linearly itself. Not safe for concurrent
// Register calls (construction is expected to happen once, up front,
// same convention as callgraph.Graph's own construction).
type Registry struct {
	byKey map[Signature][]*Model
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{byKey: map[Signature][]*Model{}}
}

// Register adds m under every one of its own Signatures. A second
// Model registered under a Signature another Model already claims is
// kept alongside it, not replacing it — Lookup returns every match.
// Consumers decide whether and how to apply multiple matching models.
func (r *Registry) Register(m *Model) {
	for _, s := range m.Signatures {
		r.byKey[s] = append(r.byKey[s], m)
	}
}

// IDs lists the IDs of the registered models, sorted, each once.
func (r *Registry) IDs() []string {
	seen := map[string]bool{}
	var out []string
	for _, ms := range r.byKey {
		for _, m := range ms {
			if !seen[m.ID] {
				seen[m.ID] = true
				out = append(out, m.ID)
			}
		}
	}
	sort.Strings(out)
	return out
}

// Remove drops every registration of the model with this ID (ROADMAP B3
// b3-3: a profile disabled it) and reports how many signatures it held.
func (r *Registry) Remove(id string) int {
	n := 0
	for sig, ms := range r.byKey {
		kept := ms[:0]
		for _, m := range ms {
			if m.ID == id {
				n++
				continue
			}
			kept = append(kept, m)
		}
		if len(kept) == 0 {
			delete(r.byKey, sig)
		} else {
			r.byKey[sig] = kept
		}
	}
	return n
}

// Lookup returns every registered Model matching sig, restricted by
// gav when gav is non-nil: a Model declaring a non-empty Artifact only
// matches when gav's own PackageName() equals it and gav's own Version
// falls inside Versions (inclusiveVersionInRange — a minimal, honestly
// narrow dotted-numeric comparator; see its own doc comment for what
// it does not attempt). A Model declaring Artifact == "" always
// matches regardless of gav (a JDK-only API model). gav == nil (an
// application class, or a dependency ResolveDepJars could not
// identify) matches only an Artifact == "" Model — never guessed
// against an unknown dependency's identity.
func (r *Registry) Lookup(sig Signature, gav *input.GAV) []*Model {
	var out []*Model
	for _, m := range r.byKey[sig] {
		if m.Artifact == "" {
			out = append(out, m)
			continue
		}
		if gav == nil || gav.PackageName() != m.Artifact {
			continue
		}
		if inclusiveVersionInRange(gav.Version, m.Versions) {
			out = append(out, m)
		}
	}
	return out
}

// inclusiveVersionInRange compares dotted-numeric version segments
// only (e.g. "5.3.21") — B0's own "every limit widens to unknown
// rather than discarding a candidate" is honored by treating a version
// this comparator cannot parse (a qualifier like "5.3.21-RELEASE", a
// non-numeric segment, a range endpoint left unparseable) as *inside*
// the range, never excluded by a guess this package cannot actually
// prove either way — a Model consumer that cares about that
// uncertainty should surface it via Fact.Incomplete once a real
// coordinator exists, not have this function silently exclude the
// Model instead.
func inclusiveVersionInRange(version string, r VersionRange) bool {
	v, ok := parseDottedVersion(version)
	if !ok {
		return true
	}
	if r.Min != "" {
		if min, ok := parseDottedVersion(r.Min); ok && compareDottedVersion(v, min) < 0 {
			return false
		}
	}
	if r.Max != "" {
		if max, ok := parseDottedVersion(r.Max); ok && compareDottedVersion(v, max) > 0 {
			return false
		}
	}
	return true
}

func parseDottedVersion(s string) ([]int, bool) {
	parts := splitDots(s)
	out := make([]int, len(parts))
	for i, p := range parts {
		n, ok := parseNonNegativeInt(p)
		if !ok {
			return nil, false
		}
		out[i] = n
	}
	return out, true
}

func splitDots(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func parseNonNegativeInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}

// compareDottedVersion compares two parsed dotted-numeric versions of
// possibly different lengths — a missing trailing segment reads as 0
// (so "5.3" == "5.3.0"), the same convention Maven's own version
// comparator uses for this case.
func compareDottedVersion(a, b []int) int {
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		var av, bv int
		if i < len(a) {
			av = a[i]
		}
		if i < len(b) {
			bv = b[i]
		}
		if av != bv {
			if av < bv {
				return -1
			}
			return 1
		}
	}
	return 0
}

// Router sends each model to the evaluator registered for its ID, and
// every other model to Default — so independent models (ServiceLoader,
// servlet dispatch) share one coordinator run without knowing each other.
type Router struct {
	ByID    map[string]Evaluator
	Default Evaluator
}

// Evaluate implements Evaluator.
func (r Router) Evaluate(model *Model, site CallSite) Fact {
	if e, ok := r.ByID[model.ID]; ok {
		return e.Evaluate(model, site)
	}
	if r.Default != nil {
		return r.Default.Evaluate(model, site)
	}
	return Fact{ModelID: model.ID, Incomplete: true, Reason: "no evaluator for model " + model.ID}
}
