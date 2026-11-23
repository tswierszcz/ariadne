package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/findings"
	"ariadne/internal/input"
	"ariadne/internal/reach"
	"ariadne/internal/vulnmap"
)

// LIBRARY_MODE LB3 lb3-2: the advisories of the other versions of a
// library's dependencies that a user of the library can get
// (-alt-versions picks them, -alt-versions-mode says how they are
// assessed). Separate from the run's findings: never part of a v1 row or a
// verdict.

// Assessment modes (-alt-versions-mode).
const (
	// altModeApprox: OSV for the other version, the mapping variant for
	// that version, reachability on the analyzed version's code (its jar is
	// the input's). Approximate: the other version's code may differ.
	altModeApprox = "approx"
	// altModeRerun (reserved): a second graph run with the other version's
	// jar. Not built yet.
	altModeRerun = "rerun"
)

const altApproxBasis = "approximate: the advisories and mapping variant of %s, reachability on the code of the analyzed %s"

func validateAltVersions(cfg config) error {
	switch cfg.altVersions {
	case "", component.AltVersionsOff, component.AltVersionsNewest:
	default:
		return fmt.Errorf("-alt-versions must be %q or %q, got %q", component.AltVersionsOff, component.AltVersionsNewest, cfg.altVersions)
	}
	switch cfg.altVersionsMode {
	case altModeApprox:
	case altModeRerun:
		return fmt.Errorf("-alt-versions-mode %s (a second graph run with the other version's jar) is not built yet; use %s", altModeRerun, altModeApprox)
	default:
		return fmt.Errorf("-alt-versions-mode must be %q, got %q", altModeApprox, cfg.altVersionsMode)
	}
	if altVersionsOn(cfg) && cfg.roots != "api" {
		return fmt.Errorf("-alt-versions needs -roots api (the library's dependency declarations, LIBRARY_MODE LB3)")
	}
	return nil
}

func altVersionsOn(cfg config) bool {
	return cfg.altVersions != "" && cfg.altVersions != component.AltVersionsOff
}

// altVersionDeps is what assessAltVersions needs from the run.
type altVersionDeps struct {
	inv      *component.Inventory
	libDeps  *component.LibraryDependencies
	fetch    component.ReferenceFetcher
	base     string
	osv      *cves.Client
	resolver *vulnmap.Resolver
	vm       map[string]vulnmap.Entry
	engine   *reach.Engine
	mode     string
	strategy string
}

// assessAltVersions chooses the other versions and assesses each in the
// given mode. The assessment is supplementary: a version list that cannot
// be read is a note on its set, and advisories or mappings that cannot be
// read mark every chosen version not_assessed with the reason — the run's
// own findings are unaffected either way.
func assessAltVersions(ctx context.Context, d altVersionDeps) ([]component.AltVersionSet, error) {
	sets := d.inv.AltVersionCandidates(ctx, d.fetch, d.base, d.libDeps, d.strategy)
	var gavs []input.GAV
	for _, s := range sets {
		for _, v := range s.Versions {
			if g, ok := component.AltGAV(s, v.Version); ok {
				gavs = append(gavs, g)
			}
		}
	}
	if len(gavs) == 0 {
		return sets, nil
	}
	res, err := d.osv.ResolveAll(ctx, gavs)
	if err != nil {
		return notAssessed(sets, fmt.Sprintf("advisories unavailable: %v", err)), nil
	}
	vm := map[string]vulnmap.Entry{}
	for k, e := range d.vm {
		vm[k] = e
	}
	var missing []cves.CVE
	for _, c := range res.CVEs {
		if _, ok := vm[c.ID]; !ok {
			missing = append(missing, c)
		}
	}
	if len(missing) > 0 {
		extra, err := d.resolver.ResolveAll(ctx, missing)
		if err != nil {
			return notAssessed(sets, fmt.Sprintf("mappings unavailable: %v", err)), nil
		}
		for k, e := range extra {
			vm[k] = e
		}
	}
	for i := range sets {
		s := &sets[i]
		for j := range s.Versions {
			v := &s.Versions[j]
			v.Mode = d.mode
			v.Basis = fmt.Sprintf(altApproxBasis, s.Package+":"+v.Version, s.InstanceVersion)
			var affecting []cves.CVE
			for _, c := range res.CVEs {
				for _, a := range c.Affected {
					if a.Package == s.Package && a.Version == v.Version {
						affecting = append(affecting, c)
						break
					}
				}
			}
			if len(affecting) == 0 {
				continue
			}
			eng := d.engine.WithComponents(d.inv.WithInstanceVersion(s.Instance, v.Version))
			for _, r := range eng.Analyze(affecting, vm) {
				for _, f := range r.Findings {
					if f.Instance != s.Instance {
						continue
					}
					v.Findings = append(v.Findings, component.AltVersionFinding{
						Advisory: r.CVE, Status: string(f.Status), Reasons: reasonStrings(f.Reasons), Path: f.Path, Note: f.Note,
					})
				}
			}
			sort.Slice(v.Findings, func(a, b int) bool { return v.Findings[a].Advisory < v.Findings[b].Advisory })
		}
	}
	return sets, nil
}

func notAssessed(sets []component.AltVersionSet, why string) []component.AltVersionSet {
	for i := range sets {
		for j := range sets[i].Versions {
			sets[i].Versions[j].NotAssessed = why
		}
	}
	return sets
}

func reasonStrings(rs []findings.ReasonCode) []string {
	var out []string
	for _, r := range rs {
		out = append(out, string(r))
	}
	return out
}

// altVersionsLine summarizes the sets for the log.
func altVersionsLine(sets []component.AltVersionSet) string {
	versions, affected, notes, skipped := 0, 0, 0, 0
	byStatus := map[string]int{}
	for _, s := range sets {
		if s.Note != "" {
			notes++
		}
		for _, v := range s.Versions {
			versions++
			if v.NotAssessed != "" {
				skipped++
			}
			if len(v.Findings) > 0 {
				affected++
			}
			for _, f := range v.Findings {
				byStatus[f.Status]++
			}
		}
	}
	var st []string
	for k, n := range byStatus {
		st = append(st, fmt.Sprintf("%s %d", k, n))
	}
	sort.Strings(st)
	return fmt.Sprintf("%d dependencies, %d other versions (%d with advisories, %d not assessed; findings: %s), %d notes",
		len(sets), versions, affected, skipped, orNoneStr(strings.Join(st, ", ")), notes)
}

func orNoneStr(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
