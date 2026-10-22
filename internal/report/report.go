// Package report turns a finished analysis run into report.md (via an
// embedded text/template) and results.json (deterministic: sorted by CVE
// id, with each CVE's own method list sorted too).
package report

import (
	"sort"
	"time"

	"ariadne/internal/callback"
	"ariadne/internal/cves"
	"ariadne/internal/methodhandle"
	"ariadne/internal/reach"
	"ariadne/internal/serviceload"
)

// Header carries everything section 6's report header asks for.
type Header struct {
	App             string
	GeneratedAt     time.Time
	AriadneVersion  string
	Profile         string
	MaxBytecodeVer  int
	GraphNodes      int
	GraphEdges      int
	PreCHAEdges     int
	CHAEdges        int
	InitEdges       int
	UnresolvedDyn   int
	UnresolvedOwner int
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
}

// Data is everything the Markdown/JSON templates need, already computed
// and sorted — the template itself does no logic beyond formatting.
type Data struct {
	Header
	StatusCounts    map[string]int
	SeverityCounts  map[string]int
	Rows            []Row
	Notes           []reach.Result // results carrying a Note, e.g. downgraded LLM verdicts
	Reachable       []reach.Result
	Framework       []reach.Result
	TotalCVEs       int
	WarningMessages []string

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
}

// Build assembles Data from a finished run's pieces. severityOf/packageOf
// come from the matching cves.CVE (Analyze's Result doesn't carry them,
// since internal/reach only needs the vulnerable methods, not the
// advisory metadata).
func Build(h Header, list []cves.CVE, results []reach.Result, warnings []string, svcRegs []serviceload.ServiceRegistration, svcLoadSites []serviceload.ServiceLoadSite, svcMaterializations []serviceload.ProviderMaterialization, callbackSites []callback.RegistrationSite, resolvedCallbacks []callback.ResolvedCallback, callbackEdges []callback.CallbackEdgeResult, handleInvokeSites []methodhandle.ResolvedInvoke, handleEdges []methodhandle.HandleEdgeResult) Data {
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
			})
		}
		if len(cve.Affected) == 0 {
			data.Rows = append(data.Rows, Row{
				CVE: r.CVE, Severity: sev, Status: r.Status,
				Methods: joinMethods(r.Methods), Source: sourceLabel(r),
			})
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
