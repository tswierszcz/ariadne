package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	// altModeRerun (lb3-3): a second run of the static pipeline with the
	// other version's jar in place of the analyzed one, the advisories
	// assessed on that graph — the other version's own code.
	altModeRerun = "rerun"
)

const (
	altApproxBasis = "approximate: the advisories and mapping variant of %s, reachability on the code of the analyzed %s"
	altRerunBasis  = "rerun: the graph rebuilt with the jar of %s (sha256 %s…) in place of %s, reachability on its own code"
)

func validateAltVersions(cfg config) error {
	switch cfg.altVersions {
	case "", component.AltVersionsOff, component.AltVersionsNewest:
	default:
		return fmt.Errorf("-alt-versions must be %q or %q, got %q", component.AltVersionsOff, component.AltVersionsNewest, cfg.altVersions)
	}
	switch cfg.altVersionsMode {
	case altModeApprox, altModeRerun:
	default:
		return fmt.Errorf("-alt-versions-mode must be %q or %q, got %q", altModeApprox, altModeRerun, cfg.altVersionsMode)
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
	// rerun runs the static pipeline again with jar in place of the -deps
	// file named file (mode rerun).
	rerun func(ctx context.Context, gav input.GAV, file, oldVersion string, jar []byte) (*staticRun, error)
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
			if d.mode == altModeRerun {
				v.Basis = fmt.Sprintf("rerun: no advisory affects %s, so no second run was needed", s.Package+":"+v.Version)
			}
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
			if d.mode == altModeRerun {
				d.rerunVersion(ctx, s, v, affecting, vm)
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

// rerunVersion assesses one other version on its own code: its jar from
// the reference repository (cache, ARIADNE_MAVEN_LOCAL, Central unless
// offline) replaces the analyzed jar in a copy of -deps, the static pipeline
// runs again, and the advisories are assessed on that graph for the
// instance of that version. Anything missing leaves the version
// not_assessed with the reason; the run's own findings are unaffected.
func (d altVersionDeps) rerunVersion(ctx context.Context, s *component.AltVersionSet, v *component.AltVersion, affecting []cves.CVE, vm map[string]vulnmap.Entry) {
	gav, _ := component.AltGAV(*s, v.Version)
	file := ""
	for _, in := range d.inv.Instances {
		if in.ID != s.Instance {
			continue
		}
		for _, a := range d.inv.Artifacts {
			if a.ID == in.ArtifactID && strings.HasPrefix(a.Location, "deps:") && in.EmbeddedPath == "" {
				file = strings.TrimPrefix(a.Location, "deps:")
			}
		}
	}
	if file == "" {
		v.NotAssessed = fmt.Sprintf("rerun: %s is not a jar of -deps (embedded or inside the application), so it cannot be replaced", s.Package+":"+s.InstanceVersion)
		return
	}
	if d.rerun == nil {
		v.NotAssessed = "rerun: not available in this run"
		return
	}
	v.Basis = fmt.Sprintf("rerun of %s (not assessed)", gav)
	base := d.base
	if base == "" {
		base = component.CentralBaseURL
	}
	jar, err := d.fetch(ctx, component.ReferenceURL(base, gav))
	if err != nil {
		v.NotAssessed = fmt.Sprintf("rerun: the jar of %s is unavailable: %v", gav, err)
		return
	}
	st, err := d.rerun(ctx, gav, file, s.InstanceVersion, jar)
	if err != nil {
		v.NotAssessed = fmt.Sprintf("rerun: the second run of the static pipeline failed: %v", err)
		return
	}
	defer st.cleanup()
	inst := ""
	for _, in := range st.components.Instances {
		if in.GAV != nil && *in.GAV == gav && in.EmbeddedPath == "" {
			inst = in.ID
		}
	}
	if inst == "" {
		v.NotAssessed = fmt.Sprintf("rerun: the second run did not identify %s in the replaced jar", gav)
		return
	}
	sum := sha256.Sum256(jar)
	v.Basis = fmt.Sprintf(altRerunBasis, gav, fmt.Sprintf("%x", sum[:6]), s.Package+":"+s.InstanceVersion)
	for _, r := range st.engine.Analyze(affecting, vm) {
		for _, f := range r.Findings {
			if f.Instance != inst {
				continue
			}
			v.Findings = append(v.Findings, component.AltVersionFinding{
				Advisory: r.CVE, Status: string(f.Status), Reasons: reasonStrings(f.Reasons), Path: f.Path, Note: f.Note,
			})
		}
	}
	sort.Slice(v.Findings, func(a, b int) bool { return v.Findings[a].Advisory < v.Findings[b].Advisory })
}

// rerunWithJar runs prepare again on a copy of -deps where jar replaces
// the file named file (and, with -deps-list, its line names the new
// version). Outputs, snapshots, traces and -alt-versions are off in the
// second run; its log is discarded. The returned cleanup also removes the
// copy.
func rerunWithJar(ctx context.Context, cfg config, gav input.GAV, file, oldVersion string, jar []byte) (*staticRun, error) {
	dir, err := os.MkdirTemp("", "ariadne-rerun-*")
	if err != nil {
		return nil, err
	}
	fail := func(err error) (*staticRun, error) {
		os.RemoveAll(dir)
		return nil, err
	}
	entries, err := os.ReadDir(cfg.deps)
	if err != nil {
		return fail(err)
	}
	for _, e := range entries {
		if e.Name() == file {
			continue
		}
		src, err := filepath.Abs(filepath.Join(cfg.deps, e.Name()))
		if err != nil {
			return fail(err)
		}
		if err := os.Symlink(src, filepath.Join(dir, e.Name())); err != nil {
			return fail(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, gav.Artifact+"-"+gav.Version+".jar"), jar, 0o644); err != nil {
		return fail(err)
	}
	cfg2 := cfg
	cfg2.deps = dir
	if cfg.depsList != "" {
		b, err := os.ReadFile(cfg.depsList)
		if err != nil {
			return fail(err)
		}
		old := gav.Group + ":" + gav.Artifact + ":jar:" + oldVersion + ":"
		list := strings.ReplaceAll(string(b), old, gav.Group+":"+gav.Artifact+":jar:"+gav.Version+":")
		cfg2.depsList = filepath.Join(dir, ".deps-list.txt")
		if err := os.WriteFile(cfg2.depsList, []byte(list), 0o644); err != nil {
			return fail(err)
		}
	}
	cfg2.altVersions = component.AltVersionsOff
	cfg2.snapshotIn, cfg2.snapshotOut, cfg2.snapshotCompare = "", "", ""
	cfg2.graphOut, cfg2.graphCompare, cfg2.runtimeTrace, cfg2.runtimeFilter = "", "", "", ""
	cfg2.research = nil
	st, err := prepare(ctx, cfg2, io.Discard)
	if err != nil {
		return fail(err)
	}
	inner := st.cleanup
	st.cleanup = func() error {
		err := inner()
		os.RemoveAll(dir)
		return err
	}
	return st, nil
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
