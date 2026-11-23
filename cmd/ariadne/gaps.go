package main

import (
	"fmt"
	"sort"

	"ariadne/internal/alloc"
	"ariadne/internal/callback"
	"ariadne/internal/component"
	"ariadne/internal/deserialize"
	"ariadne/internal/hierarchy"
	"ariadne/internal/libapi"
	"ariadne/internal/methodhandle"
	"ariadne/internal/nativebound"
	"ariadne/internal/reach"
	"ariadne/internal/reflectinvoke"
	"ariadne/internal/report"
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

// clientGapSites turns library mode's client-code sites (LIBRARY_MODE
// LB2) into gap sites of kind "client". Their candidate set is bounded
// and empty: the client's classes are not in the input, and client code
// reaches the library only through its API, which is already the roots —
// so they never make a finding MODEL_INCOMPLETE; they are counted
// (coverage.gaps, mechanism client_code) and reachable ones reported.
func clientGapSites(sites []libapi.ClientSite) []reach.GapSite {
	out := make([]reach.GapSite, 0, len(sites))
	for _, s := range sites {
		out = append(out, reach.GapSite{Kind: "client", CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, Owners: []string{}})
	}
	return out
}

// suppliedGapSites turns library mode's caller-supplied sites (lb2-7)
// into gap sites of kind "supplied" (mechanism client_code): the code
// that runs there is the implementation the library's caller or container
// passes. Like client sites their candidate set is bounded and empty, so
// they never make a finding MODEL_INCOMPLETE.
func suppliedGapSites(sites []libapi.SuppliedSite) []reach.GapSite {
	out := make([]reach.GapSite, 0, len(sites))
	for _, s := range sites {
		out = append(out, reach.GapSite{Kind: "supplied", CallerFQCN: s.CallerFQCN, CallerName: s.CallerName, CallerDesc: s.CallerDesc, Owners: []string{}})
	}
	return out
}

// objectsExist: the classes whose objects exist without client code — the
// ones the input itself creates (the created-type catalog, every origin
// but an assumed entry) and the ones a client can construct or subclass
// without overriding (libapi.Constructible). A client-code site's dispatch
// to one of them runs library code on an ordinary object, so it is not a
// stand-in for client code (LIBRARY_MODE LB2).
func objectsExist(c alloc.Catalog, constructible func(string) bool) func(string) bool {
	made := map[string]bool{}
	for _, t := range c.Types {
		for _, o := range t.Origins {
			if o.Kind != alloc.KindEntry {
				made[t.Class] = true
			}
		}
	}
	return func(fqcn string) bool { return made[fqcn] || (constructible != nil && constructible(fqcn)) }
}

// nativeGapSites turns the input's native methods into gap sites of kind
// "native" (RUNTIME_NATIVE N5 n5-1): the method itself is the site — its
// body is native code, which may call any Java method through JNI — so the
// candidate set is unbounded. Only the reachable ones take part in the
// model-gap rule.
func nativeGapSites(ms []nativebound.Method) []reach.GapSite {
	out := make([]reach.GapSite, 0, len(ms))
	for _, m := range ms {
		out = append(out, reach.GapSite{Kind: "native", CallerFQCN: m.FQCN, CallerName: m.Name, CallerDesc: m.Desc})
	}
	return out
}

// maxNativeRows bounds the report's native-method table (v2 keeps all).
const maxNativeRows = 50

// nativeReportRows renders the native summary for report.md.
func nativeReportRows(s nativebound.Summary) (int, []report.NativeRow, []report.NativeLoadRow) {
	var rows []report.NativeRow
	for i, m := range s.Reachable {
		if i == maxNativeRows {
			rows = append(rows, report.NativeRow{Method: fmt.Sprintf("… %d more (v2 coverage.native)", len(s.Reachable)-maxNativeRows)})
			break
		}
		where := "dependency"
		if m.App {
			where = "application"
		}
		rows = append(rows, report.NativeRow{Method: m.FQCN + "#" + m.Name + m.Desc, Where: where})
	}
	var loads []report.NativeLoadRow
	for _, l := range s.Loads {
		loads = append(loads, report.NativeLoadRow{Caller: l.CallerFQCN + "#" + l.CallerName + l.CallerDesc, API: l.API, Library: l.Library, Reachable: l.Reachable})
	}
	return s.Methods, rows, loads
}

// maxShadowRows bounds the report's table of components defining the same
// classes (v2 components[].shadowed keeps the per-instance counts).
const maxShadowRows = 30

// shadowReportRows groups the run's class-name collisions by the pair of
// components: kept copy, shadowed copy (JVM_GRAPH G1, CVE:F1). Rows with
// differing content come first.
func shadowReportRows(inv *component.Inventory) (total, differing int, rows []report.ShadowRow) {
	if inv == nil {
		return 0, 0, nil
	}
	label := map[string]string{}
	for _, in := range inv.Instances {
		l := in.Label()
		if l == "" {
			l = inv.InstanceLocation(in.ID)
		}
		label[in.ID] = l
	}
	type key struct{ kept, shadowed string }
	byPair := map[key]*report.ShadowRow{}
	var order []key
	for _, c := range inv.Conflicts {
		total++
		k := key{c.KeptInstance, c.ShadowedInstance}
		r := byPair[k]
		if r == nil {
			r = &report.ShadowRow{Kept: orDash(label[c.KeptInstance]), Shadowed: orDash(label[c.ShadowedInstance]), Example: c.FQCN}
			byPair[k] = r
			order = append(order, k)
		}
		r.Classes++
		if !c.SameContent {
			r.Differing++
			differing++
			if r.Differing == 1 {
				r.Example = c.FQCN
			}
		}
	}
	for _, k := range order {
		rows = append(rows, *byPair[k])
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Differing != rows[j].Differing {
			return rows[i].Differing > rows[j].Differing
		}
		if rows[i].Classes != rows[j].Classes {
			return rows[i].Classes > rows[j].Classes
		}
		return rows[i].Kept+rows[i].Shadowed < rows[j].Kept+rows[j].Shadowed
	})
	if len(rows) > maxShadowRows {
		rest := len(rows) - maxShadowRows
		rows = append(rows[:maxShadowRows], report.ShadowRow{Kept: fmt.Sprintf("… %d more pairs (v2 components[].shadowed)", rest)})
	}
	return total, differing, rows
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
