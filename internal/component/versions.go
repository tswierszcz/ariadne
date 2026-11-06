package component

import (
	"context"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/input"
	"ariadne/internal/mavenversion"
	"ariadne/internal/relocate"
)

// CVE:F6b: the version of an embedded candidate when its declared
// version is not confirmed by its classes, or when it has none. The
// classes are compared with a bounded set of other published versions
// of the same group:artifact (the Central version list,
// maven-metadata.xml, through the run's cache). The outcome is never a
// guess: exactly one fitting version confirms it, several fitting
// versions are a candidate set, none is version_unknown — and how the
// checked set was chosen is recorded.

// MetadataURL is the Central URL of g's version list.
func MetadataURL(base string, g input.GAV) string {
	return fmt.Sprintf("%s/%s/%s/maven-metadata.xml", base, strings.ReplaceAll(g.Group, ".", "/"), g.Artifact)
}

// publishedVersions reads the release versions of g from its
// maven-metadata.xml, in Maven order (SNAPSHOTs dropped).
func publishedVersions(ctx context.Context, fetch ReferenceFetcher, base string, g input.GAV) ([]string, error) {
	body, err := fetch(ctx, MetadataURL(base, g))
	if err != nil {
		return nil, err
	}
	var md struct {
		Versions []string `xml:"versioning>versions>version"`
	}
	if err := xml.Unmarshal(body, &md); err != nil {
		return nil, fmt.Errorf("maven-metadata.xml: %w", err)
	}
	seen := map[string]bool{}
	var out []string
	for _, v := range md.Versions {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] || strings.HasSuffix(v, "-SNAPSHOT") {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.SliceStable(out, func(i, j int) bool { return mavenversion.Compare(out[i], out[j]) < 0 })
	return out, nil
}

// versionWindow picks the versions to compare: with a declared version,
// the nearest published ones on each side of it (it was already checked),
// at most MaxVersionSearch; with none, every published version if there
// are at most MaxVersionSearch, else none (no anchor to choose a subset
// by). It returns the versions and how they were chosen.
func versionWindow(published []string, declared string) ([]string, string) {
	if declared == "" {
		if len(published) <= MaxVersionSearch {
			return published, fmt.Sprintf("no declared version: all %d published versions checked", len(published))
		}
		return nil, fmt.Sprintf("no declared version and %d published versions (more than %d): not searched — declare the version (-embedded-components) or provide an SBOM", len(published), MaxVersionSearch)
	}
	pos := sort.Search(len(published), func(i int) bool { return mavenversion.Compare(published[i], declared) >= 0 })
	lo, hi := pos-1, pos
	if hi < len(published) && published[hi] == declared {
		hi++
	}
	var out []string
	for len(out) < MaxVersionSearch && (lo >= 0 || hi < len(published)) {
		if hi < len(published) {
			out = append(out, published[hi])
			hi++
		}
		if len(out) < MaxVersionSearch && lo >= 0 {
			out = append(out, published[lo])
			lo--
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return mavenversion.Compare(out[i], out[j]) < 0 })
	return out, fmt.Sprintf("%d published versions nearest to the declared %s checked (of %d)", len(out), declared, len(published))
}

// searchVersions compares e's classes with the versions versionWindow
// picks and sets e's status: EmbeddedConfirmed (one fits; e.Version),
// EmbeddedVersionCandidates (several; e.VersionCandidates) or
// EmbeddedVersionUnknown. It returns the check whose classes the
// instance owns (the best fit), or nil when no instance is made.
func (inv *Inventory) searchVersions(ctx context.Context, fetch ReferenceFetcher, base string, e *Embedded, hd *hostData, rules relocate.Set, cl claims) *versionCheck {
	declared := e.GAV.Version
	what := "the declared version " + declared + " does not fit its classes"
	if declared == "" {
		what = "no version declared"
	}
	published, err := publishedVersions(ctx, fetch, base, e.GAV)
	if err != nil {
		// Not searched: a declared version keeps its F6a status
		// (content_differs), an undeclared one is version_unknown.
		e.VersionSearch = "version list unavailable: " + err.Error()
		if declared == "" {
			e.Status, e.Reason = EmbeddedVersionUnknown, what+"; "+e.VersionSearch
		} else {
			e.Reason += "; other versions not checked (" + e.VersionSearch + ")"
		}
		return nil
	}
	e.VersionsPublished = len(published)
	window, how := versionWindow(published, declared)
	e.VersionSearch = how
	if len(window) == 0 {
		if declared == "" {
			e.Status, e.Reason = EmbeddedVersionUnknown, what+"; "+how
		} else {
			e.Reason += "; no other published version to check"
		}
		return nil
	}
	var fits []*versionCheck
	unavailable := 0
	for _, v := range window {
		g := e.GAV
		g.Version = v
		c := compareVersion(ctx, fetch, base, g, hd, rules, cl)
		e.recordCheck(c, false)
		if c.err != nil {
			unavailable++
		}
		if c.fits() {
			fits = append(fits, c)
		}
	}
	switch len(fits) {
	case 0:
		reason := what + "; " + how + ", none fits"
		if unavailable > 0 {
			reason += fmt.Sprintf(" (%d reference jars unavailable)", unavailable)
		}
		e.Status, e.Reason = EmbeddedVersionUnknown, reason
		return nil
	case 1:
		c := fits[0]
		e.Status, e.Version = EmbeddedConfirmed, c.version
		e.ReferenceURL, e.ReferenceSHA256 = c.url, c.sha
		e.ReferenceClasses, e.Present, e.Identical = c.refClasses, c.present, c.identical
		e.Reason = fmt.Sprintf("version %s from its classes (%s; %s): %d of %d reference classes present, %d identical", c.version, what, how, c.present, c.refClasses, c.identical)
		return c
	}
	best := fits[0]
	for _, c := range fits {
		e.VersionCandidates = append(e.VersionCandidates, c.version)
		if c.identical*best.present > best.identical*c.present {
			best = c
		}
	}
	e.Status = EmbeddedVersionCandidates
	e.Reason = fmt.Sprintf("its classes fit %d versions (%s): %s; %s", len(fits), strings.Join(e.VersionCandidates, ", "), what, how)
	return best
}
