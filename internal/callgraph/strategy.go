package callgraph

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// GRAPH_PRECISION P0: the choice of dispatch targets (a Strategy) is kept
// apart from building, interning and storing the graph (Build). The only
// strategy today is CHA — the builder's historical behavior, unchanged;
// RTA and points-to (P2–P4) plug in here as further strategies, never by
// silently narrowing the default.

// Modes (the -callgraph-mode values).
const (
	ModeCHA = "cha"
)

// Modes lists the implemented modes (ModeRTA: rta.go).
var Modes = []string{ModeCHA, ModeRTA}

// Unknown-allocation policies (GRAPH_PRECISION P0 contract). A strategy
// that narrows dispatch by the types it knows can be created (RTA) must
// say what it does at a call whose receiver type may come from a
// mechanism it does not model.
const (
	// PolicyAllSubtypes: dispatch is never narrowed by allocations, so
	// nothing depends on knowing them (CHA).
	PolicyAllSubtypes = "all-subtypes"
	// PolicyCHAFallbackWithGap (RTA, P2): a class an unmodeled mechanism
	// may create keeps the CHA targets it dispatches to (the mechanism's
	// fallback classes become live when its site is reachable), and the
	// site is recorded as an active gap that limits a negative
	// (NOT_REACHABLE) result — never drop the targets silently.
	PolicyCHAFallbackWithGap = "cha-fallback-with-gap"
)

// Strategy chooses the implementations a virtual/interface call may run
// beyond its direct (resolved declared) target.
type Strategy interface {
	// Mode is the strategy's -callgraph-mode name.
	Mode() string
	// DispatchTargets returns, in a deterministic order, the classes
	// whose own declaration of call.Name+call.Desc the call may run as a
	// CHA-basis edge. Called only for invokevirtual/invokeinterface.
	// The result must depend only on h and call.Owner/Name/Desc: the
	// builder computes it once per such triple and reuses it for every
	// call site (BuildWith), and callers must not modify it.
	DispatchTargets(h *hierarchy.Hierarchy, call classfile.Call) []string
	// UnknownAllocationPolicy names the strategy's policy for receivers
	// created by unmodeled mechanisms.
	UnknownAllocationPolicy() string
	// SkipRule describes which declared receiver types are never
	// expanded (for the graph's metadata).
	SkipRule() string
}

// CHA expands a virtual/interface call to every transitive subtype of its
// declared receiver type that declares the exact descriptor, unless Skip
// exempts the receiver type or the subtype (section 3.6). Nil Skip uses
// DefaultSkip.
type CHA struct {
	Skip SkipFunc
	// Exclude, when set, removes classes from the targets (never from the
	// receivers): a model's decision that they do not run in this input
	// (internal/bootfeatures — Undertow features a Spring Boot application
	// leaves off). ExcludeRule describes it for the graph's settings.
	Exclude     func(class string) bool
	ExcludeRule string
}

// Mode implements Strategy.
func (CHA) Mode() string { return ModeCHA }

// UnknownAllocationPolicy implements Strategy.
func (CHA) UnknownAllocationPolicy() string { return PolicyAllSubtypes }

// SkipRule implements Strategy.
func (c CHA) SkipRule() string {
	rule := "custom skip function"
	if c.Skip == nil {
		rule = "receiver types under " + strings.Join(jdkPrefixes, ", ") + " are not expanded"
	}
	if c.Exclude != nil && c.ExcludeRule != "" {
		rule += "; " + c.ExcludeRule
	}
	return rule
}

func (c CHA) skip() SkipFunc {
	if c.Skip == nil {
		return DefaultSkip
	}
	return c.Skip
}

// DispatchTargets implements Strategy.
func (c CHA) DispatchTargets(h *hierarchy.Hierarchy, call classfile.Call) []string {
	skip := c.skip()
	// CHA expands from the call's own declared receiver type (call.Owner,
	// deliberately not the resolved declaring class): any subtype of
	// *that* static type could be the runtime receiver.
	if skip(call.Owner) {
		return nil
	}
	var out []string
	seen := map[string]bool{}
	excluded := func(string) bool { return false }
	if c.Exclude != nil {
		excluded = c.Exclude
	}
	for _, sub := range h.Subtypes(call.Owner) {
		if skip(sub) || excluded(sub) {
			continue
		}
		subInfo, ok := h.Lookup(sub)
		if !ok {
			continue
		}
		// Exact descriptor match (JVM_GRAPH G2), not a name-only match: a
		// subtype overriding a *different* overload must not be wired up.
		// A bridge method already carries the exact erased descriptor, so
		// it resolves like any declared method and its own Calls carry the
		// analysis to the covariant implementation.
		if _, ok := subInfo.MethodByNameAndDesc(call.Name, call.Desc); ok {
			if c.Exclude != nil && onlyExcludedInherit(h, sub, subInfo, call, skip, excluded) {
				continue
			}
			if !seen[sub] {
				seen[sub] = true
				out = append(out, sub)
			}
			continue
		}
		// A concrete subtype that does not declare the method runs the
		// body JVM method selection finds above it — possibly in a
		// superclass that is not itself a subtype of the call's type
		// (DefaultListableBeanFactory implements
		// ConfigurableListableBeanFactory but inherits getBean from
		// AbstractBeanFactory). That body is a target too (measured
		// against a runtime trace: testdata/graphbench/TRACE_GAPS.MD).
		if subInfo.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) != 0 {
			continue
		}
		decl, ok := h.ResolveExact(sub, call.Name, call.Desc)
		// The declaring class may itself be excluded (log4j's WriterAppender,
		// logback's OutputStreamAppender): an exclusion is about the runtime
		// type, and sub — kept — runs this body (builder 6).
		if !ok || decl == call.Owner || seen[decl] || skip(decl) {
			continue
		}
		declInfo, ok := h.Lookup(decl)
		if !ok {
			continue
		}
		// A default method of an interface outside the call's type is the
		// body selected the same way; an abstract declaration is not.
		if m, ok := declInfo.MethodByNameAndDesc(call.Name, call.Desc); !ok || m.IsAbstract() {
			continue
		}
		seen[decl] = true
		out = append(out, decl)
	}
	return out
}

// onlyExcludedInherit reports whether the body an abstract class or an
// interface declares runs only for excluded classes: concrete subtypes
// select it, and every one of them is excluded (or skipped). Such a body
// is not a target — the exclusion says those classes do not run here
// (logback's SMTPAppenderBase#append, inherited only by the SMTPAppender
// the configuration leaves out). A body no concrete subtype in the input
// selects keeps CHA's usual treatment; only a strategy with an Exclude
// asks.
func onlyExcludedInherit(h *hierarchy.Hierarchy, decl string, info *hierarchy.ClassInfo, call classfile.Call, skip SkipFunc, excluded func(string) bool) bool {
	if info.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) == 0 {
		return false
	}
	inheritors := 0
	for _, s := range h.Subtypes(decl) {
		si, ok := h.Lookup(s)
		if !ok || si.AccessFlags&(classfile.AccInterface|classfile.AccAbstract) != 0 {
			continue
		}
		if d, ok := h.ResolveExact(s, call.Name, call.Desc); !ok || d != decl {
			continue
		}
		if !skip(s) && !excluded(s) {
			return false
		}
		inheritors++
	}
	return inheritors > 0
}

// StrategyFor returns the strategy of a -callgraph-mode value.
func StrategyFor(mode string) (Strategy, error) {
	switch mode {
	case "", ModeCHA:
		return CHA{}, nil
	case ModeRTA:
		return RTA{}, nil
	}
	return nil, fmt.Errorf("unknown call-graph mode %q (implemented: %s)", mode, strings.Join(Modes, ", "))
}

// Config is the graph's construction metadata (GRAPH_PRECISION P0): what
// a comparison of two graphs must hold equal, or disclose.
type Config struct {
	Mode                    string `json:"mode"`
	Builder                 string `json:"builder"`
	SkipRule                string `json:"skip_rule"`
	UnknownAllocationPolicy string `json:"unknown_allocation_policy"`
	// Set by the caller (they are not the builder's): the roots mode and
	// the target JDK the analysis used.
	Roots      string `json:"roots,omitempty"`
	TargetJava int    `json:"target_java,omitempty"`
	// Participants are the models and creator gates the coordinator ran,
	// as "id@version", sorted (ROADMAP B3 b3-3): a graph built with
	// another selection is a different derivation.
	Participants []string `json:"participants,omitempty"`
	// DisabledParticipants are the ones the profile turned off.
	DisabledParticipants []string `json:"disabled_participants,omitempty"`
}

// Differences lists the fields in which two configs differ, sorted.
func (c Config) Differences(o Config) []string {
	var out []string
	add := func(name, a, b string) {
		if a != b {
			out = append(out, fmt.Sprintf("%s: %q vs %q", name, a, b))
		}
	}
	add("mode", c.Mode, o.Mode)
	add("builder", c.Builder, o.Builder)
	add("skip_rule", c.SkipRule, o.SkipRule)
	add("unknown_allocation_policy", c.UnknownAllocationPolicy, o.UnknownAllocationPolicy)
	add("roots", c.Roots, o.Roots)
	add("target_java", fmt.Sprint(c.TargetJava), fmt.Sprint(o.TargetJava))
	add("participants", strings.Join(c.Participants, ","), strings.Join(o.Participants, ","))
	sort.Strings(out)
	return out
}
