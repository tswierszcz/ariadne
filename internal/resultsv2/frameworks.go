package resultsv2

import (
	"sort"

	"ariadne/internal/profile"
)

// Frameworks (FRAMEWORKS F5 follow-up, additive) is what the run's profile
// knows about the application's framework: beans, @Profile conditions,
// injection points and their resolution, external entry points and the
// framework roots it added. Until now these facts were only in the log and
// report.md. Facts, not findings: nothing here is part of reachability or
// the v1 row.
type Frameworks struct {
	// Profile is the ariadne profile the run used (e.g. "spring").
	Profile string `json:"profile"`
	// RulesVersion is profile.RulesVersion: the rules these facts come from.
	RulesVersion string `json:"rules_version"`
	// ActiveProfiles is -active-profiles as given; omitted when not given,
	// in which case every @Profile condition is "unknown".
	ActiveProfiles []string             `json:"active_profiles,omitempty"`
	Beans          []FrameworkBean      `json:"beans"`
	Conditions     []FrameworkCondition `json:"conditions"`
	Injections     []FrameworkInjection `json:"injections"`
	Entries        []FrameworkEntry     `json:"entries"`
	Roots          []FrameworkRoot      `json:"roots"`
	Summary        FrameworksSummary    `json:"summary"`
}

// FrameworkBean is one bean the profile found: a stereotype class, a @Bean
// method's product, or a Spring Data repository.
type FrameworkBean struct {
	Class string `json:"class"`
	// Name is the bean's name in the container (explicit, else the default
	// Spring derives); Explicit says whether the annotation gave it.
	Name      string `json:"name"`
	Explicit  bool   `json:"explicit"`
	Source    string `json:"source"` // stereotype | bean_method | repository
	Primary   bool   `json:"primary,omitempty"`
	Qualifier string `json:"qualifier,omitempty"`
}

// FrameworkCondition is one @Profile usage and its state under the run's
// active profiles.
type FrameworkCondition struct {
	Class  string   `json:"class"`
	Method string   `json:"method,omitempty"`
	Names  []string `json:"names"`
	Active string   `json:"active"` // true | false | unknown
}

// FrameworkInjection is one injection point and what the container would
// wire there.
type FrameworkInjection struct {
	Class      string `json:"class"`
	Member     string `json:"member"`
	ParamIndex int    `json:"param_index"` // -1 for a field
	Type       string `json:"type"`
	Qualifier  string `json:"qualifier,omitempty"`
	// Value is the @Value expression of a config-value point (never a bean
	// lookup).
	Value      string   `json:"value,omitempty"`
	Resolved   string   `json:"resolved,omitempty"`
	Candidates []string `json:"candidates,omitempty"`
	Reason     string   `json:"reason,omitempty"`
}

// FrameworkEntry is one external entry point (an HTTP mapping, a web.xml
// servlet/filter/listener method) and its @Profile state.
type FrameworkEntry struct {
	Class  string `json:"class"`
	Method string `json:"method"`
	Basis  string `json:"basis"`
	Active string `json:"active"`
}

// FrameworkRoot is one method the profile made a root because the
// framework calls it reflectively.
type FrameworkRoot struct {
	Class  string `json:"class"`
	Method string `json:"method"` // name+descriptor
	Kind   string `json:"kind"`   // e.g. "constructor", "response body getter"
	// Origin is the bean or handler the root belongs to.
	Origin string `json:"origin,omitempty"`
}

// FrameworksSummary counts the lists above.
type FrameworksSummary struct {
	Beans                int            `json:"beans"`
	ConditionsByState    map[string]int `json:"conditions_by_state"`
	Injections           int            `json:"injections"`
	InjectionsResolved   int            `json:"injections_resolved"`
	InjectionsValue      int            `json:"injections_value"`
	InjectionsUnresolved int            `json:"injections_unresolved"`
	EntriesByBasis       map[string]int `json:"entries_by_basis"`
	RootsByKind          map[string]int `json:"roots_by_kind"`
}

// FrameworksInputs are the profile facts main.go already computes.
type FrameworksInputs struct {
	Profile        string
	ActiveProfiles []string
	Stereotypes    []profile.BeanName
	BeanMethods    []profile.BeanName
	Repositories   []profile.BeanName
	Conditions     []profile.ProfileCondition
	Injections     []profile.ResolvedInjection
	Entries        []profile.ExternalEntry
	LifecycleRoots []profile.BeanRoot
	WebBodyRoots   []profile.WebBodyRoot
}

// BuildFrameworks renders the profile facts; every list is sorted so the
// section is deterministic.
func BuildFrameworks(in FrameworksInputs) *Frameworks {
	f := &Frameworks{
		Profile: in.Profile, RulesVersion: profile.RulesVersion, ActiveProfiles: in.ActiveProfiles,
		Beans: []FrameworkBean{}, Conditions: []FrameworkCondition{}, Injections: []FrameworkInjection{},
		Entries: []FrameworkEntry{}, Roots: []FrameworkRoot{},
		Summary: FrameworksSummary{ConditionsByState: map[string]int{}, EntriesByBasis: map[string]int{}, RootsByKind: map[string]int{}},
	}
	for _, src := range []struct {
		name  string
		beans []profile.BeanName
	}{{"stereotype", in.Stereotypes}, {"bean_method", in.BeanMethods}, {"repository", in.Repositories}} {
		for _, b := range src.beans {
			f.Beans = append(f.Beans, FrameworkBean{Class: b.FQCN, Name: b.EffectiveName(), Explicit: b.Name != "",
				Source: src.name, Primary: b.Primary, Qualifier: b.Qualifier})
		}
	}
	sort.SliceStable(f.Beans, func(i, j int) bool {
		a, b := f.Beans[i], f.Beans[j]
		if a.Class != b.Class {
			return a.Class < b.Class
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		return a.Name < b.Name
	})
	for _, c := range in.Conditions {
		f.Conditions = append(f.Conditions, FrameworkCondition{Class: c.FQCN, Method: c.Method, Names: c.Names, Active: string(c.Active)})
		f.Summary.ConditionsByState[string(c.Active)]++
	}
	sort.SliceStable(f.Conditions, func(i, j int) bool {
		a, b := f.Conditions[i], f.Conditions[j]
		return a.Class+"#"+a.Method < b.Class+"#"+b.Method
	})
	for _, ri := range in.Injections {
		f.Injections = append(f.Injections, FrameworkInjection{Class: ri.FQCN, Member: ri.Member, ParamIndex: ri.ParamIndex,
			Type: ri.Type, Qualifier: ri.Qualifier, Value: ri.ValueExpression, Resolved: ri.Resolved,
			Candidates: ri.Candidates, Reason: ri.Reason})
		switch {
		case ri.ValueExpression != "":
			f.Summary.InjectionsValue++
		case ri.Resolved != "":
			f.Summary.InjectionsResolved++
		default:
			f.Summary.InjectionsUnresolved++
		}
	}
	sort.SliceStable(f.Injections, func(i, j int) bool {
		a, b := f.Injections[i], f.Injections[j]
		if a.Class != b.Class {
			return a.Class < b.Class
		}
		if a.Member != b.Member {
			return a.Member < b.Member
		}
		return a.ParamIndex < b.ParamIndex
	})
	for _, e := range in.Entries {
		f.Entries = append(f.Entries, FrameworkEntry{Class: e.FQCN, Method: e.Method, Basis: string(e.Basis), Active: string(e.Active)})
		f.Summary.EntriesByBasis[string(e.Basis)]++
	}
	sort.SliceStable(f.Entries, func(i, j int) bool {
		a, b := f.Entries[i], f.Entries[j]
		if a.Class+"#"+a.Method != b.Class+"#"+b.Method {
			return a.Class+"#"+a.Method < b.Class+"#"+b.Method
		}
		return a.Basis < b.Basis
	})
	for _, r := range in.LifecycleRoots {
		f.Roots = append(f.Roots, FrameworkRoot{Class: r.FQCN, Method: r.Name + r.Desc, Kind: string(r.Kind), Origin: r.Bean})
	}
	for _, r := range in.WebBodyRoots {
		f.Roots = append(f.Roots, FrameworkRoot{Class: r.FQCN, Method: r.Name + r.Desc, Kind: string(r.Kind), Origin: r.Handler})
	}
	sort.SliceStable(f.Roots, func(i, j int) bool {
		a, b := f.Roots[i], f.Roots[j]
		if a.Class+"#"+a.Method != b.Class+"#"+b.Method {
			return a.Class+"#"+a.Method < b.Class+"#"+b.Method
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Origin < b.Origin
	})
	for _, r := range f.Roots {
		f.Summary.RootsByKind[r.Kind]++
	}
	f.Summary.Beans = len(f.Beans)
	f.Summary.Injections = len(f.Injections)
	return f
}
