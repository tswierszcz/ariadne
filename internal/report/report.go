// Package report turns a finished analysis run into report.md (via an
// embedded text/template) and results.json (deterministic: sorted by CVE
// id, with each CVE's own method list sorted too).
package report

import (
	"ariadne/internal/alloc"
	"ariadne/internal/callgraph"
	"fmt"
	"sort"
	"strings"
	"time"

	"ariadne/internal/aop"
	"ariadne/internal/callback"
	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/deserialize"
	"ariadne/internal/findings"
	"ariadne/internal/methodhandle"
	"ariadne/internal/profile"
	"ariadne/internal/reach"
	"ariadne/internal/reflectinvoke"
	"ariadne/internal/serviceload"
	"ariadne/internal/snapshot"
)

// Header carries everything section 6's report header asks for.
type Header struct {
	App            string
	GeneratedAt    time.Time
	AriadneVersion string
	// AnalysisID and the results file it was written with (CVE:F9.9):
	// the v2 document's analysis_id and the -json file's name, format and
	// SHA-256 — a report and a results file belong together only when
	// these match.
	AnalysisID    string
	ResultsFile   string
	ResultsFormat string
	ResultsSHA256 string
	Profile       string
	// FrameworksDetected is FRAMEWORKS F0's own separated fact: a
	// space-joined list of every profile whose own marker classes were
	// found, independent of which one(s) Profile above actually merges
	// (Select may combine several — before F0 there was no way to see
	// the individual detections that produced a merged name like
	// "spring+jakarta").
	FrameworksDetected string
	MaxBytecodeVer     int
	// JPMSScenario is the -jpms-scenario the ServiceLoader model used
	// (SERVICE_LOADING S3); "" renders as the classpath default.
	JPMSScenario string
	// GraphConfig is the call graph's construction metadata
	// (GRAPH_PRECISION P0): mode, builder, skip rule, allocation policy,
	// roots, target JDK.
	GraphConfig     callgraph.Config
	GraphNodes      int
	GraphEdges      int
	PreCHAEdges     int
	CHAEdges        int
	InitEdges       int
	ModelEdges      int
	UnresolvedDyn   int
	UnresolvedOwner int
	// Model gaps (contract §4): unresolved dynamic sites, those reachable
	// from the roots, and those of them with no bounded candidate set.
	GapSites, GapsReachable, GapsUnbounded int
	// DataFreshness is the run's freshness state (FreshnessComplete,
	// FreshnessStale, FreshnessUnknown) and FreshnessLines its per-kind
	// detail (CVE:F2).
	DataFreshness   string
	FreshnessLines  []string
	DepsTotal       int
	DepsIdentified  int
	UnidentifiedJar []string

	// LLM* are set only when -llm was used.
	LLMModel   string
	LLMMapped  int
	LLMSkipped int // requests that failed or replies that stayed invalid

	// ExploitModel is set only when -exploit-check was used.
	ExploitModel   string
	ExploitChecked int
	ExploitSkipped int // requests that failed, were invalid, or had no path to ask about

	// FrameworkCheckModel is set only when -framework-check was used.
	FrameworkCheckModel   string
	FrameworkChecked      int
	FrameworkCheckSkipped int
}

// Row is one line of the CVE table.
type Row struct {
	CVE      string
	Package  string
	Version  string
	Severity string
	Status   reach.Status
	Methods  string
	Source   string
	Reasons  string // CVE:F1b reason codes, comma-separated
	// Chains (DESERIALIZATION Z4): catalog gadget chains linked to the
	// advisory's findings, with their states; evidence only.
	Chains string
}

// Data is everything the Markdown/JSON templates need, already computed
// and sorted — the template itself does no logic beyond formatting.
type Data struct {
	Header
	StatusCounts   map[string]int
	SeverityCounts map[string]int
	Rows           []Row
	Notes          []reach.Result // results carrying a Note, e.g. downgraded LLM verdicts
	// PerInstance are results whose advisory affects several component
	// instances or whose instance findings carry reasons/notes (CVE:F1b).
	PerInstance []reach.Result
	// Withdrawn are advisories whose every OSV record is withdrawn: listed,
	// not analyzed (CVE:F2).
	Withdrawn []cves.CVE
	// SnapshotChanges explain, per advisory, what invalidated its derived
	// result or changed its verdict since the snapshot SnapshotComparedWith
	// (CVE:F3 -snapshot-compare).
	SnapshotChanges      []snapshot.Change
	SnapshotComparedWith string
	Reachable            []reach.Result
	Framework            []reach.Result
	TotalCVEs            int
	WarningMessages      []string

	// ServiceRegistrations is SERVICE_LOADING S0's own inventory: every
	// META-INF/services/<Interface> provider declaration found across the
	// app and its dependencies. These are declarative facts, not
	// call-graph edges — see serviceload.ServiceRegistration's own doc
	// comment for why S0 deliberately stops there.
	ServiceRegistrations []serviceload.ServiceRegistration

	// ServiceLoadSites is SERVICE_LOADING S1's own result: every
	// ServiceLoader.load(...)/loadInstalled(...) call site found, with
	// its service interface resolved where DATA_FLOW D0-D2's
	// local-value analysis could — still not a call-graph edge, same
	// reasoning as ServiceRegistrations.
	ServiceLoadSites []serviceload.ServiceLoadSite

	// ServiceMaterializations is SERVICE_LOADING S2's own result: every
	// registered provider whose interface was actually resolved at some
	// ServiceLoadSite, with the constructor modeled as a call-graph
	// target or, if that wasn't possible, why (Gap) — "brak
	// implementacji daje lukę, nie brak podatności", never silently
	// dropped.
	ServiceMaterializations []serviceload.ProviderMaterialization

	// ServiceConsumptionSites records recognized ServiceLoader iterator,
	// stream, Provider calls and linked service-method dispatches plus any
	// bounded local load-site link.
	// Discovery alone does not claim that a site is reachable or that it
	// materialized a provider.
	ServiceConsumptionSites []serviceload.ServiceConsumptionSite

	// Modules is SERVICE_LOADING S3's first slice: module descriptors and
	// `provides` declarations under an explicit launch scenario
	// (class path only) — facts, never registrations or edges.
	Modules serviceload.ModuleInventory

	// Components is CVE:F1a's artifact/component-instance inventory
	// (nil in reports built without it, e.g. tests).
	Components *component.Inventory

	// Deserialization is DESERIALIZATION Z1's entry sites, candidate
	// summary and edge counts (nil in reports built without it).
	// GraphComparison (GRAPH_PRECISION P0, -graph-compare): how this
	// run's call graph differs from GraphCompareFile; nil without it.
	GraphComparison  *callgraph.Comparison
	GraphCompareFile string
	// CreatedTypes/CreatedTypeGaps (GRAPH_PRECISION P1): the catalog of
	// created classes, summarized, and its gaps.
	CreatedTypes    *alloc.Summary
	CreatedTypeGaps []alloc.Gap
	Deserialization *Deserialization

	// CallbackRegistrationSites is ASYNC_CALLBACKS A0's own inventory:
	// every call site matching (or nearly matching) a cataloged
	// callback-registration API (Thread/Executor so far) — not yet a
	// call-graph edge, and not yet resolving which concrete object the
	// callback argument is (that's A1/A2).
	CallbackRegistrationSites []callback.RegistrationSite

	// ResolvedCallbacks is ASYNC_CALLBACKS A1's own result: each
	// cataloged registration site's callback argument resolved to a
	// concrete class, where it was a direct `new` allocation — still
	// not a call-graph edge (that's A2).
	ResolvedCallbacks []callback.ResolvedCallback

	// CallbackEdges is ASYNC_CALLBACKS A2's own result: each resolved
	// callback identity connected to the class that actually declares
	// its run/call method, as a real call-graph edge — or a Gap
	// explaining why not.
	CallbackEdges []callback.CallbackEdgeResult

	// HandleInvokeSites/HandleEdges are DYNAMIC_LINKAGE L2's own
	// result: every MethodHandle#invoke/invokeExact call site found,
	// and each one's resolved receiver connected to its target method
	// as a real call-graph edge (or a Gap explaining why not).
	HandleInvokeSites []methodhandle.ResolvedInvoke
	HandleEdges       []methodhandle.HandleEdgeResult

	// SpringRoutes/SpringBeanNames/SpringProfileConditions are
	// FRAMEWORKS F1's own result: real Spring MVC route paths/verbs,
	// @Component-family bean names, and @Profile conditions, sourced
	// from annotation element values rather than just the bare
	// presence of the annotation. Routes/ProfileConditions are the
	// rendering-friendly form (joined strings) since text/template's
	// own Parse here registers no custom join function.
	SpringRoutes            []RouteRow
	SpringBeanNames         []profile.BeanName
	SpringProfileConditions []ProfileConditionRow
	// ActiveProfiles echoes the -active-profiles flag (comma-joined),
	// "" when unset — the report's own reminder that an unset flag is
	// why every SpringProfileConditions entry reports "unknown", not a
	// guessed "false".
	ActiveProfiles string

	// WebXML is FRAMEWORKS F1's own ograniczony odczyt web.xml result.
	WebXML WebXMLInput

	// SpringXMLBeans is FRAMEWORKS F1's own XML bean definitions
	// result: every `<bean>` found in a Spring classic XML config
	// resource inside the app's own classpath, keyed by that
	// resource's own path.
	SpringXMLBeans []XMLBeanRow

	// ConfigConflicts is FRAMEWORKS F1's own "priorytet/konflikt"
	// result: every bean name declared with a different class by more
	// than one source (an annotation stereotype and/or an XML
	// `<bean>`) — reported as distinct candidates, never silently
	// merged into one assumed winner. Spring's own real
	// bean-definition-override behavior is configurable
	// (allow-bean-definition-overriding) and not modeled here; this is
	// detection, not resolution.
	ConfigConflicts []ConfigConflictRow

	// Injections is FRAMEWORKS F2's own first DI result: every
	// `@Autowired` field/constructor injection point found, resolved
	// against F1's own BeanNames catalog by type, `@Qualifier`, and
	// `@Primary` — the same disambiguation order Spring's own container
	// uses. An injection point with no resolvable candidate, or more
	// than one surviving candidate with nothing to disambiguate them,
	// is reported Unknown with its own reason, never guessed.
	Injections []InjectionRow

	// WebXMLRoots is FRAMEWORKS F2's own marked-roots result: every
	// lifecycle method a web.xml-declared servlet/filter/listener
	// class actually declares — the set `-roots profile` adds to its
	// own entry points (shown regardless of which `-roots` mode this
	// run used, same as every other inventory section).
	WebXMLRoots []WebXMLRootRow

	// ExternalEntries is FRAMEWORKS F2's own entry-point catalog: every
	// known Spring MVC HTTP mapping and web.xml-declared servlet/
	// filter/listener lifecycle method, labeled with its own basis
	// (what kind of framework/container dispatch reaches it) and
	// activation condition (cross-referenced against
	// SpringProfileConditions, never a second, separate condition
	// model). Purely a labeling/reporting pass over facts the sections
	// above already compute — adds no new call-graph edge and no new
	// `-roots profile` entry point of its own.
	ExternalEntries []ExternalEntryRow

	// ProxyHandlers is FRAMEWORKS F3's own first slice: every
	// application class implementing `InvocationHandler`, with its own
	// candidate target field(s) (a `final` instance field of a
	// reference type) — identity only, no new call-graph edge; wiring
	// one of these up to REFLECTION R2's own ReflectInvokeEdges remains
	// FRAMEWORKS F3's own next slice, not done here.
	ProxyHandlers []ProxyHandlerRow

	// ReflectInvokeSites/ReflectInvokeEdges are REFLECTION R2's own
	// result: every java.lang.reflect.Method#invoke call site found,
	// and each one's resolved Method receiver connected to its real
	// dispatch target as a call-graph edge (or a Gap explaining why
	// not — an unresolved receiver, an unsupported parameter type, or a
	// proven argument-count mismatch).
	ReflectInvokeSites []reflectinvoke.ResolvedInvoke
	ReflectInvokeEdges []reflectinvoke.InvokeEdgeResult

	// ReflectNewInstanceSites/ReflectNewInstanceEdges are REFLECTION
	// R2's own second half: every java.lang.reflect.Constructor#
	// newInstance call site found, and each one's resolved Constructor
	// receiver connected to the exact "<init>" its resolved parameter
	// types declared (or a Gap explaining why not).
	ReflectNewInstanceSites []reflectinvoke.ResolvedNewInstance
	ReflectNewInstanceEdges []reflectinvoke.NewInstanceEdgeResult
}

// ProxyHandlerRow is Data's own rendering-friendly form of
// aop.ProxyHandler — TargetFields joined to a plain string.
type ProxyHandlerRow struct {
	FQCN, TargetFields string
}

// InjectionRow is Data's own rendering-friendly form of
// profile.ResolvedInjection — Candidates joined to a plain string,
// since this package's own template.Parse registers no custom join
// function.
type InjectionRow struct {
	FQCN, Member, Type, Qualifier, Resolved, Reason, Candidates, ValueExpression string
}

// WebXMLRootRow is Data's own rendering-friendly form of
// profile.WebXMLRoot.
type WebXMLRootRow struct {
	FQCN, Method, Kind string
}

// ExternalEntryRow is Data's own rendering-friendly form of
// profile.ExternalEntry.
type ExternalEntryRow struct {
	FQCN, Method, Basis string
	Active              profile.Condition
}

// XMLBeanRow is Data's own rendering-friendly form of springxml.Bean,
// with its source resource's own path attached (springxml.Bean itself
// doesn't know which resource it came from — that's
// input.FindBeansXML's own concern, one layer up, same split
// internal/webxml's Module tagging already uses).
type XMLBeanRow struct {
	Source, ID, Class, Scope string
	Primary                  bool
}

// NamedBeanCandidate is one source's own claim to a bean name —
// ConfigConflictRow's own building block.
type NamedBeanCandidate struct {
	// Kind is "annotation" or the XML resource's own path.
	Kind, Class string
}

// ConfigConflictRow is one bean name claimed, with a different class,
// by more than one source.
type ConfigConflictRow struct {
	Name       string
	Candidates []NamedBeanCandidate
}

// Build assembles Data from a finished run's pieces. severityOf/packageOf
// come from the matching cves.CVE (Analyze's Result doesn't carry them,
// since internal/reach only needs the vulnerable methods, not the
// advisory metadata).
// ServletRow/ServletMappingRow/FilterRow/FilterMappingRow/ListenerRow/
// ContextParamRow are webxml's own types plus Module — "" for a
// standalone WAR, or the owning nested WAR's own entry name for an
// EAR's module(s) — since webxml.Parse itself has no notion of which
// archive module it was reading (that's PreparedApp.WebXMLs' own
// concern, one layer up).
type ServletRow struct{ Module, Name, Class string }
type ServletMappingRow struct{ Module, Name, URLPattern string }
type FilterRow struct{ Module, Name, Class string }
type FilterMappingRow struct {
	Module, Name, URLPattern, ServletName string
}
type ListenerRow struct{ Module, Class string }
type ContextParamRow struct{ Module, Name, Value string }

// WebXMLInput is FRAMEWORKS F1's own "ograniczony odczyt web.xml"
// result, bundled into one Build parameter rather than six more
// positional ones — every row across every module (PreparedApp.WebXMLs
// can hold more than one for an EAR), plus every unmodeled element
// name seen in any of them (deduplicated).
type WebXMLInput struct {
	Servlets        []ServletRow
	ServletMappings []ServletMappingRow
	Filters         []FilterRow
	FilterMappings  []FilterMappingRow
	Listeners       []ListenerRow
	ContextParams   []ContextParamRow
	Skipped         []string
}

// RouteRow is Data's own rendering-friendly form of profile.Route —
// Paths/HTTPMethods joined to plain strings, since this package's own
// template.Parse registers no custom join function (joinMethods's own
// precedent).
type RouteRow struct {
	FQCN, Method, Paths, HTTPMethods string
}

// ProfileConditionRow is Data's own rendering-friendly form of
// profile.ProfileCondition.
type ProfileConditionRow struct {
	FQCN, Method, Names string
	Active              profile.Condition
}

func Build(h Header, list []cves.CVE, results []reach.Result, warnings []string, svcRegs []serviceload.ServiceRegistration, svcLoadSites []serviceload.ServiceLoadSite, svcMaterializations []serviceload.ProviderMaterialization, callbackSites []callback.RegistrationSite, resolvedCallbacks []callback.ResolvedCallback, callbackEdges []callback.CallbackEdgeResult, handleInvokeSites []methodhandle.ResolvedInvoke, handleEdges []methodhandle.HandleEdgeResult, springRoutes []profile.Route, springBeanNames []profile.BeanName, springProfileConditions []profile.ProfileCondition, activeProfiles string, webXML WebXMLInput, springXMLBeans []XMLBeanRow, injections []profile.ResolvedInjection, webXMLRoots []profile.WebXMLRoot, externalEntries []profile.ExternalEntry, proxyHandlers []aop.ProxyHandler, reflectInvokeSites []reflectinvoke.ResolvedInvoke, reflectInvokeEdges []reflectinvoke.InvokeEdgeResult, reflectNewInstanceSites []reflectinvoke.ResolvedNewInstance, reflectNewInstanceEdges []reflectinvoke.NewInstanceEdgeResult, serviceConsumptionSets ...[]serviceload.ServiceConsumptionSite) Data {
	byID := make(map[string]cves.CVE, len(list))
	for _, c := range list {
		byID[c.ID] = c
	}

	sorted := append([]reach.Result(nil), results...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CVE < sorted[j].CVE })
	for i := range sorted {
		sort.Strings(sorted[i].Methods)
	}

	data := Data{
		Header:                    h,
		StatusCounts:              map[string]int{},
		SeverityCounts:            map[string]int{},
		TotalCVEs:                 len(sorted),
		WarningMessages:           warnings,
		ServiceRegistrations:      svcRegs,
		ServiceLoadSites:          svcLoadSites,
		ServiceMaterializations:   svcMaterializations,
		CallbackRegistrationSites: callbackSites,
		ResolvedCallbacks:         resolvedCallbacks,
		CallbackEdges:             callbackEdges,
		HandleInvokeSites:         handleInvokeSites,
		HandleEdges:               handleEdges,
		SpringRoutes:              routeRows(springRoutes),
		SpringBeanNames:           springBeanNames,
		SpringProfileConditions:   profileConditionRows(springProfileConditions),
		ActiveProfiles:            activeProfiles,
		WebXML:                    webXML,
		SpringXMLBeans:            springXMLBeans,
		Injections:                injectionRows(injections),
		WebXMLRoots:               webXMLRootRows(webXMLRoots),
		ExternalEntries:           externalEntryRows(externalEntries),
		ProxyHandlers:             proxyHandlerRows(proxyHandlers),
		ConfigConflicts:           buildConfigConflicts(springBeanNames, springXMLBeans),
		ReflectInvokeSites:        reflectInvokeSites,
		ReflectInvokeEdges:        reflectInvokeEdges,
		ReflectNewInstanceSites:   reflectNewInstanceSites,
		ReflectNewInstanceEdges:   reflectNewInstanceEdges,
	}
	for _, set := range serviceConsumptionSets {
		data.ServiceConsumptionSites = append(data.ServiceConsumptionSites, set...)
	}

	for _, r := range sorted {
		data.StatusCounts[string(r.Status)]++
		cve := byID[r.CVE]
		sev := severityLabel(cve.Severity)
		data.SeverityCounts[sev]++

		for _, pkg := range cve.Affected {
			data.Rows = append(data.Rows, Row{
				CVE:      r.CVE,
				Package:  pkg.Package,
				Version:  pkg.Version,
				Severity: sev,
				Status:   r.Status,
				Methods:  joinMethods(r.Methods),
				Source:   sourceLabel(r),
				Reasons:  joinReasons(r.Reasons, r.ModelGaps),
			})
		}
		if len(cve.Affected) == 0 {
			data.Rows = append(data.Rows, Row{
				CVE: r.CVE, Severity: sev, Status: r.Status,
				Methods: joinMethods(r.Methods), Source: sourceLabel(r), Reasons: joinReasons(r.Reasons, r.ModelGaps),
			})
		}
		if showPerInstance(r) {
			data.PerInstance = append(data.PerInstance, r)
		}

		if r.Note != "" {
			data.Notes = append(data.Notes, r)
		}
		switch r.Status {
		case reach.StatusReachable:
			data.Reachable = append(data.Reachable, r)
		case reach.StatusFramework:
			data.Framework = append(data.Framework, r)
		}
	}
	sort.Slice(data.Rows, func(i, j int) bool { return data.Rows[i].CVE < data.Rows[j].CVE })

	return data
}

// sourceLabel shows an LLM mapping's evidence next to its source, e.g.
// "llm (diff)", so a reader can tell a diff-backed mapping from a guess.
func sourceLabel(r reach.Result) string {
	if r.Source == "llm" && r.Evidence != "" {
		return "llm (" + r.Evidence + ")"
	}
	return r.Source
}

// routeRows converts profile.Route into Data's own rendering-friendly
// RouteRow, joining Paths/HTTPMethods — HTTPMethods nil (unspecified)
// renders as "any", never a guessed single verb.
// buildConfigConflicts groups every named bean claim (an annotation's
// own explicit name, an XML <bean>'s own id) by name, and keeps only
// the names where more than one *distinct class* claims it — the same
// name with the same class from two sources is not a conflict (e.g. an
// XML bean re-declaring exactly the bean an annotation already names,
// a common no-op override pattern). Order is first-seen across
// annotationBeans then springXMLBeans, both of which are already
// deterministically ordered by their own callers — never a map
// iteration, which this codebase's own byte-identical-results.json
// contract cannot tolerate.
// injectionRows converts profile.ResolvedInjection into Data's own
// rendering-friendly InjectionRow, joining Candidates — "none" when
// empty, not a blank cell, so an empty-candidate row reads as a real
// fact rather than a rendering glitch.
func webXMLRootRows(roots []profile.WebXMLRoot) []WebXMLRootRow {
	out := make([]WebXMLRootRow, len(roots))
	for i, r := range roots {
		out[i] = WebXMLRootRow{FQCN: r.FQCN, Method: r.Method, Kind: string(r.Kind)}
	}
	return out
}

func proxyHandlerRows(handlers []aop.ProxyHandler) []ProxyHandlerRow {
	out := make([]ProxyHandlerRow, len(handlers))
	for i, p := range handlers {
		fields := make([]string, len(p.TargetFields))
		for j, f := range p.TargetFields {
			fields[j] = f.Name + ": " + f.Type
		}
		out[i] = ProxyHandlerRow{FQCN: p.FQCN, TargetFields: strings.Join(fields, ", ")}
	}
	return out
}

func externalEntryRows(entries []profile.ExternalEntry) []ExternalEntryRow {
	out := make([]ExternalEntryRow, len(entries))
	for i, e := range entries {
		out[i] = ExternalEntryRow{FQCN: e.FQCN, Method: e.Method, Basis: string(e.Basis), Active: e.Active}
	}
	return out
}

func injectionRows(injections []profile.ResolvedInjection) []InjectionRow {
	out := make([]InjectionRow, len(injections))
	for i, inj := range injections {
		candidates := "none"
		if len(inj.Candidates) > 0 {
			candidates = strings.Join(inj.Candidates, ", ")
		}
		out[i] = InjectionRow{
			FQCN:            inj.FQCN,
			Member:          inj.Member,
			Type:            inj.Type,
			Qualifier:       inj.Qualifier,
			Resolved:        inj.Resolved,
			Reason:          inj.Reason,
			Candidates:      candidates,
			ValueExpression: inj.ValueExpression,
		}
	}
	return out
}

func buildConfigConflicts(annotationBeans []profile.BeanName, xmlBeans []XMLBeanRow) []ConfigConflictRow {
	byName := map[string][]NamedBeanCandidate{}
	var order []string
	add := func(name string, c NamedBeanCandidate) {
		if name == "" {
			return
		}
		if _, ok := byName[name]; !ok {
			order = append(order, name)
		}
		byName[name] = append(byName[name], c)
	}
	for _, b := range annotationBeans {
		add(b.Name, NamedBeanCandidate{Kind: "annotation", Class: b.FQCN})
	}
	for _, b := range xmlBeans {
		add(b.ID, NamedBeanCandidate{Kind: b.Source, Class: b.Class})
	}

	var out []ConfigConflictRow
	for _, name := range order {
		candidates := byName[name]
		distinctClasses := map[string]bool{}
		for _, c := range candidates {
			distinctClasses[c.Class] = true
		}
		if len(distinctClasses) > 1 {
			out = append(out, ConfigConflictRow{Name: name, Candidates: candidates})
		}
	}
	return out
}

func routeRows(routes []profile.Route) []RouteRow {
	out := make([]RouteRow, len(routes))
	for i, r := range routes {
		methods := "any"
		if len(r.HTTPMethods) > 0 {
			methods = strings.Join(r.HTTPMethods, ", ")
		}
		out[i] = RouteRow{
			FQCN:        r.FQCN,
			Method:      r.Method,
			Paths:       strings.Join(r.Paths, ", "),
			HTTPMethods: methods,
		}
	}
	return out
}

func profileConditionRows(conds []profile.ProfileCondition) []ProfileConditionRow {
	out := make([]ProfileConditionRow, len(conds))
	for i, c := range conds {
		out[i] = ProfileConditionRow{FQCN: c.FQCN, Method: c.Method, Names: strings.Join(c.Names, ", "), Active: c.Active}
	}
	return out
}

func joinMethods(methods []string) string {
	s := ""
	for i, m := range methods {
		if i > 0 {
			s += ", "
		}
		s += "`" + m + "`"
	}
	return s
}

// severityLabel extracts a short label ("CRITICAL", "9.8", ...) from
// OSV's raw "TYPE:score" severity string, or "unknown" when absent.
func severityLabel(raw string) string {
	if raw == "" {
		return "unknown"
	}
	return raw
}

// joinReasons renders reason codes; MODEL_INCOMPLETE is followed by the
// relevant gap counts by kind, e.g. "(reflection 12, callback 1)".
func joinReasons(rs []findings.ReasonCode, gaps map[string]int) string {
	parts := make([]string, len(rs))
	for i, r := range rs {
		parts[i] = "`" + string(r) + "`"
		if r == findings.ModelIncomplete && len(gaps) > 0 {
			parts[i] += " (" + GapCounts(gaps) + ")"
		}
	}
	return strings.Join(parts, ", ")
}

// GapCounts renders model-gap counts by kind in sorted kind order.
func GapCounts(gaps map[string]int) string {
	kinds := make([]string, 0, len(gaps))
	for k := range gaps {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	parts := make([]string, len(kinds))
	for i, k := range kinds {
		parts[i] = fmt.Sprintf("%s %d", k, gaps[k])
	}
	return strings.Join(parts, ", ")
}

// Deserialization is the report's view of DESERIALIZATION Z0–Z1.
type Deserialization struct {
	Models     []deserialize.Model
	Sites      []deserialize.Site
	Summary    deserialize.CandidateSummary
	Targets    int
	EdgesAdded int
	// Chains are DESERIALIZATION Z3a's catalog chains, evaluated.
	Chains              []deserialize.ChainPresence
	ChainCatalogVersion string
	ModelCatalogVersion string
	// TaintCatalog (DATA_FLOW D4): the taint catalog when -taint ran,
	// else "" (the report then says taint was not assessed).
	TaintCatalog string
}

// TaintNotAssessed counts the entries -taint did not assess (library
// entries not called from application code); the report lists only the
// assessed ones.
func (d *Deserialization) TaintNotAssessed() int {
	n := 0
	for _, s := range d.Sites {
		if s.Taint != nil && s.Taint.State == "not_assessed" {
			n++
		}
	}
	return n
}

// showPerInstance: several instances, or an instance finding that says
// more than the advisory row already does (a note, or another reason).
func showPerInstance(r reach.Result) bool {
	if len(r.Findings) > 1 {
		return true
	}
	onRow := map[findings.ReasonCode]bool{}
	for _, c := range r.Reasons {
		onRow[c] = true
	}
	for _, f := range r.Findings {
		if f.Note != "" {
			return true
		}
		for _, c := range f.Reasons {
			if !onRow[c] {
				return true
			}
		}
	}
	return false
}
