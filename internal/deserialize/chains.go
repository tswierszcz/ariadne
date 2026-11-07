package deserialize

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/hierarchy"
	"ariadne/internal/relocate"
)

// DESERIALIZATION Z3a: candidate gadget chains. A small, versioned
// catalog of documented chains, each with the classes it needs, the
// bytecode that blocks it in fixed releases (a guard, read from the
// class actually analyzed — not inferred from a version string), the
// JDK it needs and the conditions no bytecode check decides. For every
// entry site whose format a chain uses, the chain is evaluated against
// the site's policy (Z2) and the identity of the classes present.
//
// What this is NOT: a proof of exploitability, or a search for chains.
// A site with no matching catalog chain is not free of chains — the
// catalog is a fixed list (CatalogVersion), never exhaustive. Whether
// the data read at the entry is attacker-controlled is not analyzed
// (DATA_FLOW:D4): every result carries Flow = FlowNotAnalyzed.

// ChainCatalogVersion versions the chain catalog; bump it when a chain,
// a guard or a condition changes.
const ChainCatalogVersion = "z3a-1"

// Chain states (per site × chain, and the chain-wide Presence).
const (
	// ChainCandidate: every required class is present with a confirmed
	// identity in a form no catalog guard blocks, the site's policy does
	// not reject any of them, and no modeled premise is undecided.
	ChainCandidate = "candidate_chain"
	// ChainUnresolved: nothing blocks the chain, but at least one
	// premise is undecided (an unidentified copy of a class, a policy
	// not evaluated, the target JDK not given).
	ChainUnresolved = "conditions_unresolved"
	// ChainBlocked: a guard in the analyzed bytecode, the target JDK or
	// the site's evaluated policy rules this chain out — this chain
	// only; other chains and the advisory itself are not decided by it.
	ChainBlocked = "candidate_blocked"
	// ChainAbsent (chain-wide only): a required class is not in the
	// input, so no site is evaluated for it.
	ChainAbsent = "classes_absent"
	// ChainNoEntry (chain-wide only): the classes are there but the input
	// has no entry site of the chain's format.
	ChainNoEntry = "no_entry"
)

// FlowNotAnalyzed: the data flow into the entry's argument is not
// analyzed (DATA_FLOW:D4); a chain result never says the stream is
// attacker-controlled.
const FlowNotAnalyzed = "not_analyzed"

// Guard is a bytecode fact of a fixed release that blocks a chain.
type Guard struct {
	// Kind: "declares-method" (Class declares Name+Desc; if Calls is set,
	// that method must call it — "owner#name"), "not-serializable"
	// (Class is not Serializable, so Java serialization cannot create
	// it) or "declares-field" (Class declares field Name).
	Kind        string
	Class       string
	Name, Desc  string
	Calls       string
	Description string
}

// ChainSpec is one catalog chain.
type ChainSpec struct {
	ID, Name string
	Format   string // FormatJava or FormatXStream
	// Components are the Maven group:artifact names the required classes
	// may belong to ("" when every required class is the JDK's).
	Components []string
	// Required are the component's classes the chain instantiates, by
	// their original (unrelocated) names.
	Required []string
	// JDKClasses are JDK classes the chain instantiates (the trigger
	// first); they are assumed present in the runtime and checked only
	// against the site's policy.
	JDKClasses []string
	Trigger    string
	Sink       string
	// Affected and Fixed name the releases, for the reader; the decision
	// is the guard's, read from the analyzed bytecode.
	Affected, Fixed string
	Guards          []Guard
	// MaxJava: the chain needs a JDK feature version <= MaxJava (0 = any).
	MaxJava  int
	JavaNote string
	// JavaUpdateDependent: on JDK MaxJava itself the chain depends on the
	// update release, which -target-java does not give (unresolved).
	JavaUpdateDependent bool
	// Conditions are premises no check here decides, for the reader;
	// they do not change the state (every candidate is conditional on
	// the data and the runtime).
	Conditions []string
	Advisories []string
	Source     string
}

// ChainCatalog is the Z3a chain catalog (ChainCatalogVersion).
var ChainCatalog = []ChainSpec{
	{
		ID: "java:commons-collections3:lazymap", Name: "CommonsCollections6 (HashMap → TiedMapEntry → LazyMap → InvokerTransformer)",
		Format:     FormatJava,
		Components: []string{"commons-collections:commons-collections"},
		Required: []string{
			"org.apache.commons.collections.keyvalue.TiedMapEntry",
			"org.apache.commons.collections.map.LazyMap",
			"org.apache.commons.collections.functors.ChainedTransformer",
			"org.apache.commons.collections.functors.ConstantTransformer",
			"org.apache.commons.collections.functors.InvokerTransformer",
		},
		JDKClasses: []string{"java.util.HashMap"},
		Trigger:    "java.util.HashMap#readObject (hashes its keys)",
		Sink:       "InvokerTransformer#transform — reflective call of any method (Runtime#exec)",
		Affected:   "3.0–3.2.1", Fixed: "3.2.2",
		Guards: []Guard{{Kind: "declares-method", Class: "org.apache.commons.collections.functors.InvokerTransformer",
			Name: "readObject", Desc: readObjectDesc, Calls: "org.apache.commons.collections.functors.FunctorUtils#checkUnsafeSerialization",
			Description: "InvokerTransformer#readObject refuses to deserialize (FunctorUtils#checkUnsafeSerialization) unless the system property org.apache.commons.collections.enableUnsafeSerialization=true"}},
		Conditions: []string{"the stream is attacker-controlled", "no process-wide jdk.serialFilter rejects the classes"},
		Advisories: []string{"CVE-2015-7501", "CVE-2015-4852"},
		Source:     "ysoserial CommonsCollections6; COLLECTIONS-580; commons-collections 3.2.2 release notes; bytecode of 3.2.1 and 3.2.2",
	},
	{
		ID: "java:commons-collections4:transformingcomparator", Name: "CommonsCollections2/4 (PriorityQueue → TransformingComparator → InvokerTransformer)",
		Format:     FormatJava,
		Components: []string{"org.apache.commons:commons-collections4"},
		Required: []string{
			"org.apache.commons.collections4.comparators.TransformingComparator",
			"org.apache.commons.collections4.functors.ChainedTransformer",
			"org.apache.commons.collections4.functors.ConstantTransformer",
			"org.apache.commons.collections4.functors.InvokerTransformer",
		},
		JDKClasses: []string{"java.util.PriorityQueue"},
		Trigger:    "java.util.PriorityQueue#readObject (re-heapifies with its comparator)",
		Sink:       "InvokerTransformer#transform — reflective call of any method",
		Affected:   "4.0", Fixed: "4.1",
		Guards: []Guard{{Kind: "not-serializable", Class: "org.apache.commons.collections4.functors.InvokerTransformer",
			Description: "InvokerTransformer is not Serializable (4.1+), so Java serialization cannot create it"}},
		Conditions: []string{"the stream is attacker-controlled", "no process-wide jdk.serialFilter rejects the classes"},
		Advisories: []string{"CVE-2015-7501"},
		Source:     "ysoserial CommonsCollections2/4; COLLECTIONS-580; bytecode of 4.0 and 4.1",
	},
	{
		ID: "java:commons-beanutils:beancomparator", Name: "CommonsBeanutils1 (PriorityQueue → BeanComparator → TemplatesImpl)",
		Format:     FormatJava,
		Components: []string{"commons-beanutils:commons-beanutils"},
		Required:   []string{"org.apache.commons.beanutils.BeanComparator"},
		JDKClasses: []string{"java.util.PriorityQueue", "com.sun.org.apache.xalan.internal.xsltc.trax.TemplatesImpl"},
		Trigger:    "java.util.PriorityQueue#readObject",
		Sink:       "BeanComparator#compare → PropertyUtils#getProperty(\"outputProperties\") → TemplatesImpl#getOutputProperties (loads attacker bytecode)",
		Affected:   "every 1.x release in the catalog's knowledge", Fixed: "none known (the gadget is BeanComparator itself)",
		MaxJava:  15,
		JavaNote: "from JDK 16 the strong encapsulation of com.sun.org.apache.xalan.internal stops PropertyUtils' reflective call unless the package is opened",
		Conditions: []string{"the stream is attacker-controlled", "the comparator the stream carries is serializable: BeanComparator's default ComparableComparator comes from commons-collections, a variant uses String.CASE_INSENSITIVE_ORDER",
			"no process-wide jdk.serialFilter rejects the classes"},
		Source: "ysoserial CommonsBeanutils1; bytecode of commons-beanutils 1.9.4",
	},
	{
		ID: "java:groovy:convertedclosure", Name: "Groovy1 (AnnotationInvocationHandler → ConvertedClosure → MethodClosure)",
		Format:     FormatJava,
		Components: []string{"org.codehaus.groovy:groovy", "org.codehaus.groovy:groovy-all"},
		Required:   []string{"org.codehaus.groovy.runtime.ConvertedClosure", "org.codehaus.groovy.runtime.MethodClosure"},
		JDKClasses: []string{"sun.reflect.annotation.AnnotationInvocationHandler"},
		Trigger:    "sun.reflect.annotation.AnnotationInvocationHandler#readObject (calls the proxied Map)",
		Sink:       "MethodClosure#call — any method (String#execute)",
		Affected:   "1.7.0–2.4.3", Fixed: "2.4.4",
		Guards: []Guard{{Kind: "declares-method", Class: "org.codehaus.groovy.runtime.MethodClosure", Name: "readResolve", Desc: "()Ljava/lang/Object;",
			Description: "MethodClosure#readResolve refuses to deserialize unless MethodClosure.ALLOW_RESOLVE is set (2.4.4+)"}},
		MaxJava:             8,
		JavaUpdateDependent: true,
		JavaNote:            "AnnotationInvocationHandler#readObject no longer calls the proxied map from JDK 8u72 (and every JDK 9+); a JDK 8 target is not decided (the update is unknown)",
		Conditions:          []string{"the stream is attacker-controlled", "no process-wide jdk.serialFilter rejects the classes"},
		Advisories:          []string{"CVE-2015-3253"},
		Source:              "ysoserial Groovy1; GROOVY-7504; bytecode of groovy 2.4.3 and 2.4.4",
	},
	{
		ID: "xstream:eventhandler", Name: "XStream dynamic proxy with java.beans.EventHandler",
		Format:     FormatXStream,
		JDKClasses: []string{"java.util.TreeSet", "java.beans.EventHandler"},
		Trigger:    "a sorted set (or map) comparing an element that is a dynamic proxy",
		Sink:       "java.beans.EventHandler#invoke — any public method of a deserialized object (ProcessBuilder#start)",
		Affected:   "XStream before 1.4.7", Fixed: "1.4.7",
		Guards: []Guard{{Kind: "declares-field", Class: "com.thoughtworks.xstream.converters.reflection.ReflectionConverter", Name: "eventHandlerType",
			Description: "XStream's ReflectionConverter refuses java.beans.EventHandler (1.4.7+)"}},
		Conditions: []string{"the XML is attacker-controlled", "the runtime has the java.desktop module (java.beans)"},
		Advisories: []string{"CVE-2013-7285"},
		Source:     "XStream security advisory CVE-2013-7285 (x-stream.github.io/CVE-2013-7285.html); bytecode of XStream 1.4.6 and 1.4.7",
	},
}

const readObjectDesc = "(Ljava/io/ObjectInputStream;)V"

// ClassOwner is the component instance a class of the input belongs to
// (CVE:F1a/F6a/F7a), as the caller's inventory states it.
type ClassOwner struct {
	Instance string
	// Component is "group:artifact", "" when the instance is unidentified
	// or the application.
	Component string
	// Version is the instance's version, or the candidates joined by ","
	// (CVE:F6b); "" when unknown.
	Version string
	// Relocations under which the instance's classes sit (CVE:F7a).
	Relocations []relocate.Rule
}

// ChainInputs is what chain evaluation needs besides the sites.
type ChainInputs struct {
	H *hierarchy.Hierarchy
	// Owner returns the owner of a selected class; nil means no
	// inventory (every copy is unidentified).
	Owner func(ci *hierarchy.ClassInfo) ClassOwner
	// Relocated are the instances with relocations, to find a chain's
	// classes under their relocated names.
	Relocated []ClassOwner
	// TargetJava is -target-java (0 = not given).
	TargetJava int
}

// ChainClass is one required class as found in the input.
type ChainClass struct {
	Name  string // original name
	Found string // the name in the input (relocated or not); "" = absent
	Owner ClassOwner
	// Confirmed: the owner is one of the chain's components.
	Confirmed bool
}

// ChainPresence is a chain's evaluation independent of a site.
type ChainPresence struct {
	Spec ChainSpec
	// Copy is "" for the classes under their original names, else the
	// ID of the instance whose relocations name them.
	Copy string
	// Instance owns the chain's (first) class; "" when absent.
	Instance string
	State    string // ChainAbsent, ChainNoEntry, ChainBlocked, ChainUnresolved, ChainCandidate
	Classes  []ChainClass
	Missing  []string
	// Blocks and Unresolved are the reasons for the state.
	Blocks, Unresolved []string
	// Sites are the per-site results (none when absent).
	Sites []ChainSite
	// Counts per state over Sites.
	Counts map[string]int
}

// ChainSite is one site × chain result.
type ChainSite struct {
	Site               *Site
	State              string
	Blocks, Unresolved []string
	// Own are the reasons specific to this site (its policy); the rest
	// are the chain's.
	Own  []string
	Flow string
}

// SiteReason is the site-specific reason, for the report ("—" when the
// state is the chain's own).
func (cs ChainSite) SiteReason() string {
	if len(cs.Own) == 0 {
		return "—"
	}
	return strings.Join(cs.Own, "; ")
}

// ShownSites are the sites the report lists one by one: the
// application's own entries and any site whose state is its own (its
// policy); HiddenSites counts the others (library entries sharing the
// chain's state).
func (p ChainPresence) ShownSites() []ChainSite {
	var out []ChainSite
	for _, s := range p.Sites {
		if s.Site.IsApp || len(s.Own) > 0 {
			out = append(out, s)
		}
	}
	return out
}

// HiddenSites: see ShownSites.
func (p ChainPresence) HiddenSites() int { return len(p.Sites) - len(p.ShownSites()) }

// CountsText is "state n, …" over the chain's sites, or "none".
func (p ChainPresence) CountsText() string {
	if len(p.Sites) == 0 {
		return "none"
	}
	var parts []string
	for _, k := range SortedStates(p.Counts) {
		parts = append(parts, k+" "+strconv.Itoa(p.Counts[k]))
	}
	return strings.Join(parts, ", ")
}

// chainCopy is one naming of a chain's classes: the original names
// (rules empty) or a confirmed instance's relocations.
type chainCopy struct {
	label string // "" for the original names, else the instance ID
	rules relocate.Set
}

// EvaluateChains evaluates every catalog chain against the input and
// every site of its format, once per copy of the chain's classes: the
// classes under their original names, and each confirmed instance of the
// chain's component that carries them under its relocations (CVE:F7a).
// Two copies are two results: one may be blocked by its guard while the
// other is a candidate. Results are in catalog order, the original copy
// first; sites in the order given.
func EvaluateChains(in ChainInputs, sites []Site) []ChainPresence {
	c := &classifier{h: in.H, cache: map[string]int{}}
	var out []ChainPresence
	for _, spec := range ChainCatalog {
		comps := map[string]bool{}
		for _, comp := range spec.Components {
			comps[comp] = true
		}
		copies := []chainCopy{{}}
		if len(spec.Required) > 0 {
			for _, r := range in.Relocated {
				if !comps[r.Component] {
					continue
				}
				set, err := relocate.New(r.Relocations)
				if err != nil || !set.Changes(spec.Required[0]) {
					continue
				}
				copies = append(copies, chainCopy{label: r.Instance, rules: set})
			}
		}
		var ps []ChainPresence
		for _, cp := range copies {
			p := evaluateCopy(in, c, spec, comps, cp, sites)
			if cp.label != "" && len(p.Missing) == len(spec.Required) {
				continue // this instance does not carry the chain at all
			}
			ps = append(ps, p)
		}
		// The original names are reported as absent only when no copy
		// exists at all.
		if len(ps) > 1 && ps[0].State == ChainAbsent && len(ps[0].Missing) == len(spec.Required) {
			ps = ps[1:]
		}
		out = append(out, ps...)
	}
	return out
}

func evaluateCopy(in ChainInputs, c *classifier, spec ChainSpec, comps map[string]bool, cp chainCopy, sites []Site) ChainPresence {
	p := ChainPresence{Spec: spec, Copy: cp.label, Counts: map[string]int{}}
	name := func(n string) string { return cp.rules.Name(n) }
	for _, orig := range spec.Required {
		cc := ChainClass{Name: orig}
		ci, ok := in.H.Lookup(name(orig))
		if !ok {
			p.Missing = append(p.Missing, name(orig))
			p.Classes = append(p.Classes, cc)
			continue
		}
		cc.Found = ci.FQCN
		if in.Owner != nil {
			cc.Owner = in.Owner(ci)
		}
		cc.Confirmed = comps[cc.Owner.Component]
		if !cc.Confirmed {
			who := "an unidentified component"
			if cc.Owner.Component != "" {
				who = cc.Owner.Component
			}
			p.Unresolved = append(p.Unresolved, fmt.Sprintf("%s found in %s, not identified as %s", cc.Found, who, strings.Join(spec.Components, " or ")))
		}
		if p.Instance == "" {
			p.Instance = cc.Owner.Instance
		}
		p.Classes = append(p.Classes, cc)
	}
	if len(p.Missing) > 0 {
		p.State = ChainAbsent
		return p
	}
	for _, g := range spec.Guards {
		gname := name(g.Class)
		blocked, decided := guardHolds(in.H, c, g, gname, cp.rules)
		switch {
		case !decided:
			p.Unresolved = append(p.Unresolved, "guard class "+gname+" not in the input: "+g.Description+" — not checked")
		case blocked:
			p.Blocks = append(p.Blocks, g.Description)
		}
		// A chain of JDK classes only (XStream's) belongs to the
		// deserializer whose guard decides it.
		if p.Instance == "" && in.Owner != nil {
			if ci, ok := in.H.Lookup(gname); ok {
				p.Instance = in.Owner(ci).Instance
			}
		}
	}
	if spec.MaxJava > 0 {
		switch {
		case in.TargetJava == 0:
			p.Unresolved = append(p.Unresolved, "needs JDK <= "+strconv.Itoa(spec.MaxJava)+"; the target JDK is not given (-target-java): "+spec.JavaNote)
		case in.TargetJava > spec.MaxJava:
			p.Blocks = append(p.Blocks, "target JDK "+strconv.Itoa(in.TargetJava)+" > "+strconv.Itoa(spec.MaxJava)+": "+spec.JavaNote)
		case in.TargetJava == spec.MaxJava && spec.JavaUpdateDependent:
			p.Unresolved = append(p.Unresolved, "target JDK "+strconv.Itoa(in.TargetJava)+": "+spec.JavaNote)
		}
	}
	p.State = stateOf(p.Blocks, p.Unresolved)
	classes := make([]string, 0, len(p.Classes)+len(spec.JDKClasses))
	for _, cc := range p.Classes {
		classes = append(classes, cc.Found)
	}
	classes = append(classes, spec.JDKClasses...)
	for i := range sites {
		s := &sites[i]
		if s.Format != spec.Format {
			continue
		}
		cs := ChainSite{Site: s, Flow: FlowNotAnalyzed}
		cs.Blocks = append(cs.Blocks, p.Blocks...)
		cs.Unresolved = append(cs.Unresolved, p.Unresolved...)
		switch s.Filter {
		case FilterUnknown, FilterVersionUnknown:
			r := "site policy " + s.Filter + " (not evaluated)"
			cs.Unresolved = append(cs.Unresolved, r)
			cs.Own = append(cs.Own, r)
		}
		if s.allow != nil {
			var rejected []string
			for _, cl := range classes {
				if !s.allow(cl) {
					rejected = append(rejected, cl)
				}
			}
			if len(rejected) > 0 {
				r := "site policy (" + s.Filter + ") rejects " + strings.Join(rejected, ", ")
				cs.Blocks = append(cs.Blocks, r)
				cs.Own = append(cs.Own, r)
			}
		}
		cs.State = stateOf(cs.Blocks, cs.Unresolved)
		p.Counts[cs.State]++
		p.Sites = append(p.Sites, cs)
	}
	if len(p.Sites) == 0 {
		p.State = ChainNoEntry
	}
	return p
}

func stateOf(blocks, unresolved []string) string {
	switch {
	case len(blocks) > 0:
		return ChainBlocked
	case len(unresolved) > 0:
		return ChainUnresolved
	}
	return ChainCandidate
}

// guardHolds reads guard g on class name: blocked (the fixed form is
// present) and decided (the class is in the input).
func guardHolds(h *hierarchy.Hierarchy, c *classifier, g Guard, name string, rules relocate.Set) (blocked, decided bool) {
	ci, ok := h.Lookup(name)
	if !ok {
		return false, false
	}
	switch g.Kind {
	case "not-serializable":
		return c.serializable(name) == serNo, true
	case "declares-field":
		for _, f := range ci.Fields {
			if f.Name == g.Name {
				return true, true
			}
		}
		return false, true
	case "declares-method":
		m, ok := ci.MethodByNameAndDesc(g.Name, g.Desc)
		if !ok {
			return false, true
		}
		if g.Calls == "" {
			return true, true
		}
		owner, mname, _ := strings.Cut(g.Calls, "#")
		owner = rules.Name(owner) // relocated with the guard's class
		for _, call := range m.Calls {
			if call.Name == mname && call.Owner == owner {
				return true, true
			}
		}
		return false, true
	}
	return false, false
}

// ChainStateCounts sums site states over every chain present (for the
// log line), keyed by state; absent chains count once as ChainAbsent.
func ChainStateCounts(ps []ChainPresence) map[string]int {
	out := map[string]int{}
	for _, p := range ps {
		if p.State == ChainAbsent || p.State == ChainNoEntry {
			out[p.State]++
			continue
		}
		for k, v := range p.Counts {
			out[k] += v
		}
	}
	return out
}

// SortedStates returns the keys of a state count map in a fixed order.
func SortedStates(m map[string]int) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
