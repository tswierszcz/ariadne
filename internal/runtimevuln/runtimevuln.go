// Package runtimevuln assesses the target runtime's own advisories
// (CVE:F8b). A Provider answers for the runtimes it covers and says so
// explicitly for the ones it does not: an empty list is never evidence
// that a runtime is unaffected.
//
// The first provider is the OpenJDK Vulnerability Group's advisories as a
// curated catalog embedded in the binary (openjdk_advisories.yaml). The
// OpenJDK web site's terms of use do not allow automated access, so
// ariadne never fetches openjdk.org: a maintainer saves the advisory
// pages by hand, converts them (cmd/openjdk-catalog), reviews the result
// and commits it. Every entry links to its official advisory.
package runtimevuln

import (
	_ "embed"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"ariadne/internal/runtimeinfo"

	"gopkg.in/yaml.v3"
)

// Applicability of one advisory to the runtime (docs/RESULTS_CONTRACT.MD).
const (
	Affected     = "affected"
	NotAffected  = "not_affected"
	Undetermined = "undetermined"
)

// Assessment states of a runtime.
const (
	Performed    = "performed"
	NotPerformed = "not_performed"
)

// Finding is one advisory's applicability to the runtime. Reachability is
// never assessed here (no model of the JDK's code: CVE:F8c).
type Finding struct {
	CVE           string  `json:"cve"`
	Advisory      string  `json:"advisory"`
	URL           string  `json:"url"`
	Component     string  `json:"component,omitempty"`
	CVSS          float64 `json:"cvss,omitempty"`
	Applicability string  `json:"applicability"`
	// LastAffected is the line's last affected release the advisory
	// states; the runtime is compared with it.
	LastAffected string `json:"last_affected"`
	Reachability string `json:"reachability"`
	Reason       string `json:"reason"`
}

// Assessment is a provider's answer for one runtime.
type Assessment struct {
	Provider string `json:"provider"`
	// CatalogVersion and CuratedThrough identify the data: the catalog's
	// own version and the date of its newest advisory.
	CatalogVersion string    `json:"catalog_version,omitempty"`
	CuratedThrough string    `json:"curated_through,omitempty"`
	State          string    `json:"state"`
	Reason         string    `json:"reason,omitempty"`
	Findings       []Finding `json:"findings,omitempty"`
	// Counts by applicability.
	Counts map[string]int `json:"counts,omitempty"`
	// Notes are coverage limits that apply to this runtime (an advisory
	// that does not cover its line, the catalog's age).
	Notes []string `json:"notes,omitempty"`
}

// Provider assesses runtimes it covers.
type Provider interface {
	Name() string
	Assess(r runtimeinfo.Runtime, now time.Time) Assessment
}

// Catalog is the curated OpenJDK advisory catalog.
type Catalog struct {
	Schema     int        `yaml:"schema"`
	Version    string     `yaml:"version"`
	Source     string     `yaml:"source"`
	Note       string     `yaml:"note"`
	Advisories []Advisory `yaml:"advisories"`
}

// Advisory is one OpenJDK Vulnerability Group advisory.
type Advisory struct {
	Date string `yaml:"date"`
	URL  string `yaml:"url"`
	// LastAffected maps a feature line ("8", "11", "17", …) to the last
	// affected release the advisory names for it ("8u461", "17.0.16",
	// "25"). A line absent here is not covered by the advisory.
	LastAffected map[string]string `yaml:"last_affected"`
	// EarlierAffected records whether the advisory states that releases
	// before LastAffected are affected too ("… and earlier"). Without
	// that statement an earlier release is undetermined: a fix list does
	// not say when the flaw was introduced.
	EarlierAffected bool  `yaml:"earlier_affected"`
	CVEs            []CVE `yaml:"cves"`
}

// CVE is one row of an advisory's risk matrix.
type CVE struct {
	ID        string  `yaml:"id"`
	Component string  `yaml:"component"`
	CVSS      float64 `yaml:"cvss"`
	// Lines are the feature lines the matrix marks as affected.
	Lines []int `yaml:"lines"`
}

//go:embed openjdk_advisories.yaml
var embeddedCatalog []byte

var (
	cveRe     = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)
	dateRe    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	urlRe     = regexp.MustCompile(`^https://openjdk\.org/groups/vulnerability/advisories/\d{4}-\d{2}-\d{2}$`)
	jdk8uRe   = regexp.MustCompile(`^8u(\d+)$`)
	lineKeyRe = regexp.MustCompile(`^\d+$`)
)

// ParseCatalog reads and validates a catalog. Unknown fields, malformed
// ids, dates, URLs and versions are errors.
func ParseCatalog(b []byte) (Catalog, error) {
	var c Catalog
	dec := yaml.NewDecoder(strings.NewReader(string(b)))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return Catalog{}, fmt.Errorf("openjdk catalog: %w", err)
	}
	if c.Schema != 1 {
		return Catalog{}, fmt.Errorf("openjdk catalog: schema %d, want 1", c.Schema)
	}
	seen := map[string]bool{}
	for _, a := range c.Advisories {
		if !dateRe.MatchString(a.Date) || !urlRe.MatchString(a.URL) || !strings.HasSuffix(a.URL, a.Date) {
			return Catalog{}, fmt.Errorf("openjdk catalog: advisory %q: bad date or URL %q", a.Date, a.URL)
		}
		if seen[a.Date] {
			return Catalog{}, fmt.Errorf("openjdk catalog: advisory %s twice", a.Date)
		}
		seen[a.Date] = true
		for line, v := range a.LastAffected {
			if !lineKeyRe.MatchString(line) {
				return Catalog{}, fmt.Errorf("openjdk catalog: advisory %s: line %q", a.Date, line)
			}
			if _, ok := ParseUpstream(v); !ok {
				return Catalog{}, fmt.Errorf("openjdk catalog: advisory %s: line %s version %q", a.Date, line, v)
			}
		}
		for _, cve := range a.CVEs {
			if !cveRe.MatchString(cve.ID) {
				return Catalog{}, fmt.Errorf("openjdk catalog: advisory %s: CVE id %q", a.Date, cve.ID)
			}
			for _, l := range cve.Lines {
				if _, ok := a.LastAffected[strconv.Itoa(l)]; !ok {
					return Catalog{}, fmt.Errorf("openjdk catalog: advisory %s: %s marks line %d, which the advisory names no affected release for", a.Date, cve.ID, l)
				}
			}
		}
	}
	sort.Slice(c.Advisories, func(i, j int) bool { return c.Advisories[i].Date < c.Advisories[j].Date })
	return c, nil
}

// ParseUpstream parses an upstream release name: "25", "17.0.16",
// "11.0.21.1", "8u461".
func ParseUpstream(s string) (runtimeinfo.Release, bool) {
	if m := jdk8uRe.FindStringSubmatch(s); m != nil {
		u, _ := strconv.Atoi(m[1])
		return runtimeinfo.Release{Feature: 8, Update: u}, true
	}
	return runtimeinfo.ParseVersion(s)
}

// compare orders releases of one line by interim, update and patch; the
// build number is not part of a release's identity here.
func compare(a, b runtimeinfo.Release) int {
	for _, d := range [][2]int{{a.Interim, b.Interim}, {a.Update, b.Update}, {a.Patch, b.Patch}} {
		if d[0] != d[1] {
			if d[0] < d[1] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// OpenJDK is the provider over a catalog.
type OpenJDK struct {
	C Catalog
	// Distributions it answers for: builds that use upstream's update
	// numbering and ship upstream's security fixes in the same update.
	Distributions map[string]bool
}

// NewOpenJDK returns the provider over the embedded catalog.
func NewOpenJDK() (*OpenJDK, error) {
	c, err := ParseCatalog(embeddedCatalog)
	if err != nil {
		return nil, err
	}
	return &OpenJDK{C: c, Distributions: map[string]bool{"Temurin": true}}, nil
}

// EmbeddedCatalogVersion is the embedded catalog's version ("" when it
// does not parse — NewOpenJDK reports the error).
func EmbeddedCatalogVersion() string {
	c, err := ParseCatalog(embeddedCatalog)
	if err != nil {
		return ""
	}
	return c.Version
}

// Name implements Provider.
func (p *OpenJDK) Name() string { return "openjdk-vulnerability-group" }

// StaleAfter is how old the newest catalog advisory may be before the
// assessment says the catalog is behind (advisories are quarterly).
const StaleAfter = 100 * 24 * time.Hour

// Assess implements Provider.
func (p *OpenJDK) Assess(r runtimeinfo.Runtime, now time.Time) Assessment {
	out := Assessment{Provider: p.Name(), CatalogVersion: p.C.Version, State: NotPerformed}
	if n := len(p.C.Advisories); n > 0 {
		out.CuratedThrough = p.C.Advisories[n-1].Date
	}
	switch {
	case !p.Distributions[r.Distribution]:
		dist := strings.TrimSpace(r.Vendor + " " + r.Distribution)
		if r.Distribution == "" {
			dist = "an unidentified distribution"
		}
		out.Reason = fmt.Sprintf("%s is not covered: the upstream OpenJDK advisories are applied only to %s", dist, strings.Join(keys(p.Distributions), ", "))
		return out
	case r.Version == nil:
		out.Reason = "the runtime version is unknown"
		return out
	case len(p.C.Advisories) == 0:
		out.Reason = "the curated OpenJDK advisory catalog has no advisories yet"
		return out
	}
	v := *r.Version
	line := strconv.Itoa(v.Feature)
	out.State = Performed
	out.Counts = map[string]int{}
	uncovered := 0
	for _, a := range p.C.Advisories {
		last, ok := a.LastAffected[line]
		if !ok {
			uncovered++
			continue
		}
		lv, _ := ParseUpstream(last)
		c := compare(v, lv)
		for _, cve := range a.CVEs {
			f := Finding{CVE: cve.ID, Advisory: a.Date, URL: a.URL, Component: cve.Component, CVSS: cve.CVSS, LastAffected: last, Reachability: "not_assessed",
				Reason: "no model of the runtime's own code (CVE:F8c)"}
			marked := false
			for _, l := range cve.Lines {
				if l == v.Feature {
					marked = true
				}
			}
			switch {
			case !marked:
				f.Applicability = NotAffected
				f.Reason = "the advisory does not mark line " + line + " as affected"
			case c > 0:
				f.Applicability = NotAffected
				f.Reason = "fixed: the runtime is later than the last affected release " + last
			case c == 0:
				f.Applicability = Affected
			case a.EarlierAffected:
				f.Applicability = Affected
				f.Reason = "earlier than " + last + ", which the advisory states are affected too; " + f.Reason
			default:
				f.Applicability = Undetermined
				f.Reason = "earlier than the last affected release " + last + ", and the advisory does not say whether earlier releases are affected"
			}
			out.Counts[f.Applicability]++
			out.Findings = append(out.Findings, f)
		}
	}
	if uncovered > 0 {
		out.Notes = append(out.Notes, fmt.Sprintf("%d of %d catalog advisories do not cover line %s (no data, not unaffected)", uncovered, len(p.C.Advisories), line))
	}
	if t, err := time.Parse("2006-01-02", out.CuratedThrough); err == nil && now.Sub(t) > StaleAfter {
		out.Notes = append(out.Notes, fmt.Sprintf("the catalog's newest advisory is from %s: advisories published since are not included", out.CuratedThrough))
	}
	sort.Slice(out.Findings, func(i, j int) bool {
		if out.Findings[i].Advisory != out.Findings[j].Advisory {
			return out.Findings[i].Advisory > out.Findings[j].Advisory
		}
		return out.Findings[i].CVE < out.Findings[j].CVE
	})
	return out
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
