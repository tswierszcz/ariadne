package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ariadne/internal/runtimevuln"

	"ariadne/internal/report"
	"ariadne/internal/resultsv2"
	"ariadne/internal/runtimeinfo"
)

// loadTargetRuntime reads -jdk-home or -runtime-file (CVE:F8a); nil when
// neither is given.
func loadTargetRuntime(cfg config) (*runtimeinfo.Runtime, error) {
	var (
		r   runtimeinfo.Runtime
		err error
	)
	switch {
	case cfg.jdkHome != "":
		r, err = runtimeinfo.FromJDKHome(cfg.jdkHome)
	case cfg.runtimeFile != "":
		r, err = runtimeinfo.FromFile(cfg.runtimeFile)
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// targetRuntimeKey is the runtime's identity in the snapshot
// configuration: the normalized fields, not the path (a replay on another
// machine names the same runtime).
func targetRuntimeKey(r *runtimeinfo.Runtime) string {
	if r == nil {
		return ""
	}
	fields := []string{r.Source, r.Vendor, r.Distribution, r.VersionString, r.OS, r.Arch, r.Libc}
	if r.Package != nil {
		fields = append(fields, r.Package.Name, r.Package.Version, r.Package.Distro)
	}
	return strings.Join(fields, "|")
}

const runtimeNotDescribed = "no target runtime described (-jdk-home or -runtime-file)"

// assessTargetRuntime runs the runtime advisory provider (CVE:F8b); nil
// without a runtime.
// A catalogFile replaces the embedded catalog; its version is returned.
func assessTargetRuntime(r *runtimeinfo.Runtime, catalogFile string) (*runtimevuln.Assessment, string, error) {
	p, err := runtimevuln.NewOpenJDK()
	if err != nil {
		return nil, "", err
	}
	if catalogFile != "" {
		b, err := os.ReadFile(catalogFile)
		if err != nil {
			return nil, "", fmt.Errorf("-openjdk-catalog: %w", err)
		}
		if p.C, err = runtimevuln.ParseCatalog(b); err != nil {
			return nil, "", fmt.Errorf("-openjdk-catalog %s: %w", catalogFile, err)
		}
	}
	if r == nil {
		return nil, p.C.Version, nil
	}
	a := p.Assess(*r, time.Now())
	return &a, p.C.Version, nil
}

// targetRuntimeResult is the v2 target_runtime section. Without a
// description or a provider that covers the runtime the assessment is
// "not_performed" with the reason: an empty list of runtime advisories is
// never presented as evidence of safety.
func targetRuntimeResult(r *runtimeinfo.Runtime, a *runtimevuln.Assessment) *resultsv2.TargetRuntime {
	out := &resultsv2.TargetRuntime{RulesVersion: runtimeinfo.Version, Assessment: resultsv2.AssessmentNotPerformed, Runtime: r, Advisories: a}
	switch {
	case r == nil:
		out.Reason = runtimeNotDescribed
	case a.State == runtimevuln.Performed:
		out.Assessment = runtimevuln.Performed
	default:
		out.Reason = a.Reason
	}
	return out
}

// targetRuntimeLine is the report header's line.
func targetRuntimeLine(r *runtimeinfo.Runtime, a *runtimevuln.Assessment) string {
	if r == nil {
		return "not described (-jdk-home or -runtime-file); runtime assessment: not performed"
	}
	line := r.Describe()
	if a.State == runtimevuln.Performed {
		line += fmt.Sprintf("; runtime assessment (%s, catalog %s, advisories through %s): %d affected, %d undetermined, %d not affected — see \"Target runtime advisories\"",
			a.Provider, a.CatalogVersion, a.CuratedThrough, a.Counts[runtimevuln.Affected], a.Counts[runtimevuln.Undetermined], a.Counts[runtimevuln.NotAffected])
	} else {
		line += "; runtime assessment: not performed (" + a.Reason + ")"
	}
	for _, n := range a.Notes {
		line += "; " + n
	}
	if len(r.Unresolved) > 0 {
		line += "; unresolved: " + strings.Join(r.Unresolved, "; ")
	}
	return line
}

// runtimeAdvisoryRows are the report's runtime advisory rows: affected and
// undetermined ones.
func runtimeAdvisoryRows(a *runtimevuln.Assessment) []report.RuntimeAdvisory {
	if a == nil {
		return nil
	}
	var out []report.RuntimeAdvisory
	for _, f := range a.Findings {
		if f.Applicability == runtimevuln.NotAffected {
			continue
		}
		out = append(out, report.RuntimeAdvisory{CVE: f.CVE, Advisory: f.Advisory, URL: f.URL, Component: f.Component, CVSS: f.CVSS,
			Applicability: f.Applicability, LastAffected: f.LastAffected, Reason: f.Reason})
	}
	return out
}
