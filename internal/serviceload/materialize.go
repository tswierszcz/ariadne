package serviceload

import (
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// FindMaterializations is SERVICE_LOADING S2's own "which provider
// constructors are actually candidate call-graph targets" step. It
// deliberately materializes only registrations whose Interface was
// actually resolved at some ServiceLoadSite — S2's own declared negative
// case ("ujemny przypadek deklaracji bez użycia... mają oczekiwane
// wyniki"): a provider nobody's resolved code ever asks for gets no
// materialization and no edge, which is the correct outcome, not a gap.
//
// The same (Interface, ProviderClass) pair is materialized at most once
// even if registered more than once (S2's own "powtórzony wpis") — a
// provider's constructor is one call-graph target regardless of how many
// META-INF/services files named it.
func FindMaterializations(h *hierarchy.Hierarchy, sites []ServiceLoadSite, regs []ServiceRegistration) []ProviderMaterialization {
	usedInterfaces := map[string]bool{}
	for _, s := range sites {
		if s.Interface != "" {
			usedInterfaces[s.Interface] = true
		}
	}

	type key struct{ iface, provider string }
	seen := map[key]bool{}
	var out []ProviderMaterialization
	for _, reg := range regs {
		if !reg.ProviderFound || !usedInterfaces[reg.Interface] {
			continue
		}
		k := key{reg.Interface, reg.ProviderClass}
		if seen[k] {
			continue
		}
		seen[k] = true

		out = append(out, materialize(h, reg))
	}
	return out
}

func materialize(h *hierarchy.Hierarchy, reg ServiceRegistration) ProviderMaterialization {
	m := ProviderMaterialization{Registration: reg}

	ci, ok := h.Lookup(reg.ProviderClass)
	switch {
	case !ok:
		// Defensive: reg.ProviderFound was already true (S0), so this
		// should not happen — recorded as a gap rather than assumed
		// impossible.
		m.Gap = "provider class not found in the analyzed hierarchy"
	case ci.IsInterface() || ci.AccessFlags&classfile.AccAbstract != 0:
		m.Gap = "provider class is an interface or abstract and cannot be instantiated directly"
	default:
		ctor, found := ci.MethodByNameAndDesc("<init>", "()V")
		switch {
		case !found:
			m.Gap = "provider class has no no-argument constructor"
		case ctor.AccessFlags&classfile.AccPublic == 0:
			m.Gap = "provider class's no-argument constructor is not public"
		default:
			m.ConstructorMethod = reg.ProviderClass + "#<init>()V"
		}
	}
	return m
}

// MaterializedEdge is one (caller, provider constructor) pair
// SERVICE_LOADING S2 wants as a call-graph edge — internal/callgraph
// does the actual graph insertion (only it can touch Graph's unexported
// adjacency), so this package hands over plain identity, not a Graph
// reference.
type MaterializedEdge struct {
	CallerFQCN, CallerName, CallerDesc string
	// PC is the originating ServiceLoadSite's own call-site offset —
	// real bytecode evidence for why this edge exists, carried through
	// so the edge can cite it the same way a direct call's own Sites
	// would (JVM_GRAPH G0's own reasoning).
	PC           int
	ProviderFQCN string
}

// EdgeTargets joins sites with materializations into the concrete pairs
// EdgeTargets' caller asked for — one pair per (site, provider) when a
// site's resolved Interface matches a materialization with a resolved
// ConstructorMethod. A site with an unresolved Interface, or a
// materialization with a Gap, contributes nothing: never a fabricated
// edge standing in for missing information.
func EdgeTargets(sites []ServiceLoadSite, materializations []ProviderMaterialization) []MaterializedEdge {
	providersByInterface := map[string][]string{}
	for _, m := range materializations {
		if m.ConstructorMethod == "" {
			continue
		}
		providersByInterface[m.Registration.Interface] = append(providersByInterface[m.Registration.Interface], m.Registration.ProviderClass)
	}

	var out []MaterializedEdge
	for _, s := range sites {
		if s.Interface == "" {
			continue
		}
		for _, providerFQCN := range providersByInterface[s.Interface] {
			out = append(out, MaterializedEdge{
				CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc,
				PC: s.PC, ProviderFQCN: providerFQCN,
			})
		}
	}
	return out
}
