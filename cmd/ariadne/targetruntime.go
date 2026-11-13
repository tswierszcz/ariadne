package main

import (
	"strings"

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

// Runtime assessment states (CVE:F8a: no advisory provider yet).
const (
	runtimeNotDescribed = "no target runtime described (-jdk-home or -runtime-file)"
	runtimeNoProvider   = "no runtime advisory provider in this build (CVE:F8b)"
)

// targetRuntimeResult is the v2 target_runtime section. The assessment is
// "not_performed" in every case until a provider exists: an empty list of
// runtime advisories is never presented as evidence of safety.
func targetRuntimeResult(r *runtimeinfo.Runtime) *resultsv2.TargetRuntime {
	out := &resultsv2.TargetRuntime{RulesVersion: runtimeinfo.Version, Assessment: resultsv2.AssessmentNotPerformed, Runtime: r}
	if r == nil {
		out.Reason = runtimeNotDescribed
	} else {
		out.Reason = runtimeNoProvider
	}
	return out
}

// targetRuntimeLine is the report header's line.
func targetRuntimeLine(r *runtimeinfo.Runtime) string {
	if r == nil {
		return "not described (-jdk-home or -runtime-file); runtime assessment: not performed"
	}
	line := r.Describe() + "; runtime assessment: not performed (" + runtimeNoProvider + ")"
	if len(r.Unresolved) > 0 {
		line += "; unresolved: " + strings.Join(r.Unresolved, "; ")
	}
	return line
}
