package main

import (
	"sort"

	"ariadne/internal/callback"
	"ariadne/internal/deserialize"
	"ariadne/internal/hierarchy"
	"ariadne/internal/methodhandle"
	"ariadne/internal/reach"
	"ariadne/internal/reflectinvoke"
	"ariadne/internal/serviceload"
)

// gapInputs are the run's dynamic-site analyses whose unresolved sites
// are model gaps (docs/RESULTS_CONTRACT.MD §4).
type gapInputs struct {
	reflectInvokes []reflectinvoke.ResolvedInvoke
	newInstances   []reflectinvoke.ResolvedNewInstance
	handles        []methodhandle.ResolvedInvoke
	callbackSites  []callback.RegistrationSite
	callbacks      []callback.ResolvedCallback
	loadSites      []serviceload.ServiceLoadSite
	consumptions   []serviceload.ServiceConsumptionSite
	registrations  []serviceload.ServiceRegistration
	// deserialization entries (DESERIALIZATION Z1): unbounded — the
	// format's own JDK callbacks (e.g. HashMap#readObject hashing its
	// keys) can call into any class, which Z1 does not model.
	deserialization []deserialize.Site
}

// collectGapSites turns every unresolved site into a reach.GapSite with
// the bounded candidate owners the site's own model gives:
//   - reflection Method.invoke / Constructor.newInstance and
//     MethodHandle.invoke with an unresolved target: unbounded;
//   - an unresolved callback argument: subtypes of the cataloged
//     callback interface (a known API's uncataloged overload: subtypes
//     of Runnable and Callable, the types that make it one);
//   - an unresolved ServiceLoader load or consumption site: every
//     registered provider class (only a registered provider can be
//     loaded).
func collectGapSites(h *hierarchy.Hierarchy, in gapInputs) []reach.GapSite {
	var out []reach.GapSite
	for _, r := range in.reflectInvokes {
		if r.UnresolvedReason != "" {
			out = append(out, reach.GapSite{Kind: "reflection", CallerFQCN: r.Site.CallerFQCN, CallerName: r.Site.CallerName, CallerDesc: r.Site.CallerDesc})
		}
	}
	for _, r := range in.newInstances {
		if r.UnresolvedReason != "" {
			out = append(out, reach.GapSite{Kind: "reflection", CallerFQCN: r.Site.CallerFQCN, CallerName: r.Site.CallerName, CallerDesc: r.Site.CallerDesc})
		}
	}
	for _, r := range in.handles {
		if r.UnresolvedReason != "" {
			out = append(out, reach.GapSite{Kind: "methodhandle", CallerFQCN: r.Site.CallerFQCN, CallerName: r.Site.CallerName, CallerDesc: r.Site.CallerDesc})
		}
	}

	subtypes := func(ifaces ...string) []string {
		owners := []string{}
		for _, i := range ifaces {
			owners = append(owners, h.Subtypes(i)...)
		}
		sort.Strings(owners)
		return owners
	}
	for _, r := range in.callbacks {
		if r.UnresolvedReason != "" {
			s := r.Site
			out = append(out, reach.GapSite{Kind: "callback", CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, Owners: subtypes(s.Model.CallbackInterface)})
		}
	}
	for _, s := range in.callbackSites {
		if s.Outcome == callback.OutcomeUnsupportedAPI {
			out = append(out, reach.GapSite{Kind: "callback", CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, Owners: subtypes("java.lang.Runnable", "java.util.concurrent.Callable")})
		}
	}

	providers := []string{}
	for _, r := range in.registrations {
		providers = append(providers, r.ProviderClass)
	}
	sort.Strings(providers)
	for _, s := range in.loadSites {
		if s.UnresolvedReason != "" {
			out = append(out, reach.GapSite{Kind: "serviceloader", CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, Owners: providers})
		}
	}
	for _, s := range in.consumptions {
		if s.UnresolvedReason != "" {
			out = append(out, reach.GapSite{Kind: "serviceloader", CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, Owners: providers})
		}
	}
	for _, s := range in.deserialization {
		out = append(out, reach.GapSite{Kind: "deserialization", CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc})
	}
	return out
}
