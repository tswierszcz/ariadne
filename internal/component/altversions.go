package component

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/input"
	"ariadne/internal/mavenversion"
)

// LIBRARY_MODE LB3 (lb3-2): the other versions of a library's dependency
// that a user of the library can get, chosen by a named strategy
// (-alt-versions) from the dependency's published versions
// (maven-metadata.xml through the run's cache). The analysis input holds
// one version; a range declaration, a provided dependency or the user's own
// dependency management can give a user another one. Choosing never
// guesses: a version list that cannot be read is a note, not "no other
// version".

// Selection strategies (-alt-versions).
const (
	AltVersionsOff = "off"
	// AltVersionsNewest: the highest published version inside a declared
	// range, and the newest published version of the analyzed version's
	// major.minor line.
	AltVersionsNewest = "newest"
)

// Why a version was chosen (AltVersion.ChosenBy).
const (
	ChosenRangeMax   = "range_max"   // highest published version in the declared range
	ChosenLineLatest = "line_latest" // newest published version of the analyzed major.minor line
)

// AltVersionsVersion names the selection rules; a snapshot records it.
const AltVersionsVersion = "lb3-2"

// AltVersionSet is one dependency of the library and the other versions
// chosen for it.
type AltVersionSet struct {
	Package         string `json:"package"` // group:artifact
	SuppliedBy      string `json:"supplied_by"`
	Instance        string `json:"instance"`
	InstanceVersion string `json:"instance_version"`
	// Declared is the declared version or range (DeclaredDependency.Version).
	Declared string       `json:"declared,omitempty"`
	Versions []AltVersion `json:"versions,omitempty"`
	// Note says why no version was chosen or why the choice is partial.
	Note string `json:"note,omitempty"`
}

// AltVersion is one chosen version and, once assessed, its findings.
type AltVersion struct {
	Version  string   `json:"version"`
	ChosenBy []string `json:"chosen_by"`
	// Mode is how the version was assessed (-alt-versions-mode); Basis
	// says what the assessment rests on.
	Mode     string              `json:"mode,omitempty"`
	Basis    string              `json:"basis,omitempty"`
	Findings []AltVersionFinding `json:"findings"`
	// NotAssessed says why the version was chosen but not assessed (its
	// advisories or mappings could not be read); Findings is then empty and
	// is no evidence of anything.
	NotAssessed string `json:"not_assessed,omitempty"`
}

// AltVersionFinding is one advisory affecting an alternative version.
type AltVersionFinding struct {
	Advisory string   `json:"advisory"`
	Status   string   `json:"status"`
	Reasons  []string `json:"reasons,omitempty"`
	Path     []string `json:"path,omitempty"`
	Note     string   `json:"note,omitempty"`
}

// AltVersionCandidates applies strategy to every dependency of d that has
// an identified input instance: its published versions are fetched with
// fetch from base. Dependencies without an instance (absent, unidentified)
// have no code to assess and are left out.
func (inv *Inventory) AltVersionCandidates(ctx context.Context, fetch ReferenceFetcher, base string, d *LibraryDependencies, strategy string) []AltVersionSet {
	if d == nil || strategy == "" || strategy == AltVersionsOff {
		return nil
	}
	if base == "" {
		base = CentralBaseURL
	}
	var out []AltVersionSet
	for _, dd := range d.Declared {
		if dd.Instance == "" || dd.SuppliedBy == SuppliedUndeclared {
			continue
		}
		in, ok := inv.instance(dd.Instance)
		if !ok || in.GAV == nil || len(in.VersionCandidates) > 0 {
			continue
		}
		set := AltVersionSet{Package: dd.Package, SuppliedBy: dd.SuppliedBy, Instance: dd.Instance,
			InstanceVersion: in.GAV.Version, Declared: dd.Version}
		published, err := publishedVersions(ctx, fetch, base, *in.GAV)
		if err != nil {
			set.Note = fmt.Sprintf("published versions unavailable (%s: %v): no other version chosen", MetadataURL(base, *in.GAV), err)
			out = append(out, set)
			continue
		}
		set.Versions = chooseNewest(published, in.GAV.Version, dd.Version)
		if len(set.Versions) == 0 {
			set.Note = "no published version other than the analyzed one fits the strategy"
		}
		out = append(out, set)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Package < out[j].Package })
	return out
}

func (inv *Inventory) instance(id string) (Instance, bool) {
	for _, in := range inv.Instances {
		if in.ID == id {
			return in, true
		}
	}
	return Instance{}, false
}

// chooseNewest is the "newest" strategy over published (Maven order):
// the highest version inside declared when it is a range, and the newest
// version of have's major.minor line newer than have. have itself is never
// chosen; a version chosen for both reasons is listed once.
func chooseNewest(published []string, have, declared string) []AltVersion {
	chosen := map[string][]string{}
	if isRange(declared) {
		if r, err := mavenversion.ParseRange(declared); err == nil {
			for i := len(published) - 1; i >= 0; i-- {
				if r.Contains(published[i]) == mavenversion.In {
					if published[i] != have {
						chosen[published[i]] = append(chosen[published[i]], ChosenRangeMax)
					}
					break
				}
			}
		}
	}
	if line, ok := majorMinor(have); ok {
		for i := len(published) - 1; i >= 0; i-- {
			if l, ok := majorMinor(published[i]); ok && l == line && mavenversion.Compare(published[i], have) > 0 {
				chosen[published[i]] = append(chosen[published[i]], ChosenLineLatest)
				break
			}
		}
	}
	var out []AltVersion
	for v, by := range chosen {
		out = append(out, AltVersion{Version: v, ChosenBy: by, Findings: []AltVersionFinding{}})
	}
	sort.Slice(out, func(i, j int) bool { return mavenversion.Compare(out[i].Version, out[j].Version) < 0 })
	return out
}

// majorMinor is v's "major.minor" line; ok is false for a version without
// two leading numeric components.
func majorMinor(v string) (string, bool) {
	parts := strings.FieldsFunc(v, func(r rune) bool { return r == '.' || r == '-' })
	if len(parts) < 2 || !numeric(parts[0]) || !numeric(parts[1]) {
		return "", false
	}
	return parts[0] + "." + parts[1], true
}

func numeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// WithInstanceVersion returns a copy of inv in which instance id has the
// given version: the same classes and jar (assessment mode "approx" — the
// analyzed version's code stands in for the other version's).
func (inv *Inventory) WithInstanceVersion(id, version string) *Inventory {
	cp := *inv
	cp.Instances = append([]Instance(nil), inv.Instances...)
	for i := range cp.Instances {
		if cp.Instances[i].ID == id && cp.Instances[i].GAV != nil {
			g := *cp.Instances[i].GAV
			g.Version = version
			cp.Instances[i].GAV = &g
		}
	}
	return &cp
}

// AltGAV is the coordinate of an alternative version of set's package.
func AltGAV(set AltVersionSet, version string) (input.GAV, bool) {
	g, a, ok := strings.Cut(set.Package, ":")
	if !ok {
		return input.GAV{}, false
	}
	return input.GAV{Group: g, Artifact: a, Version: version}, true
}
