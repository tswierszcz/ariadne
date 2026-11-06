package component

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"ariadne/internal/input"
)

// CVE:F6b: an optional SBOM (-sbom), one format: CycloneDX JSON, spec
// 1.4–1.6. An SBOM is a list of claims. It is tied to the analyzed input
// only through SHA-256 hashes: a CycloneDX component whose SHA-256 is the
// hash of an analyzed dependency jar describes that jar. Components
// nested inside such a component (CycloneDX "components" of a component:
// contained in it) are embedded-component candidates of that jar, and
// are verified by their classes like every other candidate. Nothing else
// is taken from the SBOM: a listed dependency (the "dependencies" graph,
// or a component not tied to a jar by its hash) does not prove that its
// code is in the input, and a hash-matched SBOM identity is evidence on
// the jar's instance, never a replacement of the identity read from the
// jar.

// SBOM is a loaded CycloneDX document.
type SBOM struct {
	Path        string
	SHA256      string
	SpecVersion string
	Serial      string
	root        cdxComponent // metadata.component
	components  []cdxComponent
}

// SBOMSummary is what the run took from the SBOM.
type SBOMSummary struct {
	Path        string
	SHA256      string
	Format      string
	SpecVersion string
	Serial      string
	// Components counts every component (nested included); Linked those
	// tied to an analyzed jar by SHA-256; Candidates the nested components
	// added as embedded candidates; Unlinked Maven components tied to
	// nothing (not used).
	Components, Linked, Candidates, Unlinked int
	Issues                                   []string
}

type cdxHash struct {
	Alg     string `json:"alg"`
	Content string `json:"content"`
}

type cdxComponent struct {
	Type       string         `json:"type"`
	BOMRef     string         `json:"bom-ref"`
	Group      string         `json:"group"`
	Name       string         `json:"name"`
	Version    string         `json:"version"`
	PURL       string         `json:"purl"`
	Hashes     []cdxHash      `json:"hashes"`
	Components []cdxComponent `json:"components"`
}

type cdxDocument struct {
	BOMFormat    string `json:"bomFormat"`
	SpecVersion  string `json:"specVersion"`
	SerialNumber string `json:"serialNumber"`
	Metadata     struct {
		Component *cdxComponent `json:"component"`
	} `json:"metadata"`
	Components []cdxComponent `json:"components"`
}

// LoadSBOM reads and validates a CycloneDX JSON SBOM ("" = none).
func LoadSBOM(path string) (*SBOM, error) {
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading -sbom %s: %w", path, err)
	}
	var d cdxDocument
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("-sbom %s: not CycloneDX JSON: %w", path, err)
	}
	if d.BOMFormat != "CycloneDX" {
		return nil, fmt.Errorf("-sbom %s: bomFormat %q, want CycloneDX (only CycloneDX JSON is supported)", path, d.BOMFormat)
	}
	switch d.SpecVersion {
	case "1.4", "1.5", "1.6":
	default:
		return nil, fmt.Errorf("-sbom %s: CycloneDX specVersion %q not supported (1.4, 1.5, 1.6)", path, d.SpecVersion)
	}
	sum := sha256.Sum256(b)
	s := &SBOM{Path: path, SHA256: hex.EncodeToString(sum[:]), SpecVersion: d.SpecVersion, Serial: d.SerialNumber, components: d.Components}
	if d.Metadata.Component != nil {
		s.root = *d.Metadata.Component
	}
	return s, nil
}

// mavenPURL parses "pkg:maven/<group>/<name>@<version>[?…][#…]".
func mavenPURL(p string) (input.GAV, bool) {
	if !strings.HasPrefix(p, "pkg:maven/") {
		return input.GAV{}, false
	}
	p = strings.TrimPrefix(p, "pkg:maven/")
	if i := strings.IndexAny(p, "?#"); i >= 0 {
		if q, err := url.ParseQuery(strings.TrimPrefix(p[i:], "?")); err == nil && q.Get("classifier") != "" {
			return input.GAV{}, false // a classified artifact is not the main jar
		}
		p = p[:i]
	}
	at := strings.LastIndexByte(p, '@')
	if at < 0 {
		return input.GAV{}, false
	}
	ga, v := p[:at], p[at+1:]
	slash := strings.LastIndexByte(ga, '/')
	if slash <= 0 {
		return input.GAV{}, false
	}
	unesc := func(s string) string {
		if u, err := url.PathUnescape(s); err == nil {
			return u
		}
		return s
	}
	g := input.GAV{Group: unesc(ga[:slash]), Artifact: unesc(ga[slash+1:]), Version: unesc(v)}
	if g.Group == "" || g.Artifact == "" || g.Version == "" {
		return input.GAV{}, false
	}
	return g, true
}

func (c cdxComponent) sha256() string {
	for _, h := range c.Hashes {
		if strings.EqualFold(strings.ReplaceAll(h.Alg, "-", ""), "SHA256") {
			return strings.ToLower(h.Content)
		}
	}
	return ""
}

func (c cdxComponent) label() string {
	if c.BOMRef != "" {
		return c.BOMRef
	}
	if c.PURL != "" {
		return c.PURL
	}
	return c.Group + ":" + c.Name + ":" + c.Version
}

// ApplySBOM ties the SBOM's components to the analyzed dependency jars by
// SHA-256, adds "sbom" identity evidence to those jars' instances, and
// adds their nested Maven components as embedded candidates (verified by
// ResolveEmbedded). Call it after Build and before ResolveEmbedded.
func (inv *Inventory) ApplySBOM(s *SBOM) {
	if s == nil {
		return
	}
	sum := &SBOMSummary{Path: s.Path, SHA256: s.SHA256, Format: "CycloneDX JSON", SpecVersion: s.SpecVersion, Serial: s.Serial}
	inv.SBOM = sum
	byHash := map[string]int{} // artifact SHA-256 -> host instance index
	for i, in := range inv.Instances {
		if in.Detection == DetectedApp || in.Detection == DetectedEmbedded {
			continue
		}
		if a := inv.artifact(in.ArtifactID); a != nil && a.Role == RoleDependency {
			if _, dup := byHash[a.SHA256]; !dup {
				byHash[a.SHA256] = i
			}
		}
	}
	var walk func(c cdxComponent, linkedParent int)
	walk = func(c cdxComponent, linkedParent int) {
		sum.Components++
		g, isMaven := mavenPURL(c.PURL)
		idx, linked := byHash[c.sha256()]
		switch {
		case linked:
			sum.Linked++
			host := &inv.Instances[idx]
			ev := IdentityEvidence{Kind: "sbom", Detail: fmt.Sprintf("SBOM component %s (SHA-256 match)", c.label())}
			if isMaven {
				gg := g
				ev.GAV = &gg
				switch {
				case host.GAV == nil:
					sum.Issues = append(sum.Issues, fmt.Sprintf("component %s names unidentified jar %s as %s: recorded as evidence, not used as its identity", c.label(), inv.ArtifactLocation(host.ArtifactID), g))
				case *host.GAV != g:
					sum.Issues = append(sum.Issues, fmt.Sprintf("component %s (%s) disagrees with the identity read from the jar (%s): the jar's identity is kept", c.label(), g, host.GAV))
				}
			}
			host.Evidence = append(host.Evidence, ev)
		case linkedParent >= 0 && isMaven:
			host := inv.Instances[linkedParent]
			parentRef := ""
			if a := inv.artifact(host.ArtifactID); a != nil {
				parentRef = inv.ArtifactLocation(a.ID)
			}
			if host.GAV != nil && host.GAV.PackageName() == g.PackageName() {
				break
			}
			dup := false
			for _, e := range inv.Embedded {
				if e.HostInstance == host.ID && e.GAV.PackageName() == g.PackageName() {
					dup = true
					if e.GAV.Version != g.Version {
						sum.Issues = append(sum.Issues, fmt.Sprintf("component %s nested in %s: version %s disagrees with the candidate %s already known for that jar (kept; the classes decide)", c.label(), parentRef, g.Version, e.GAV))
					}
				}
			}
			if dup {
				break
			}
			inv.Embedded = append(inv.Embedded, Embedded{HostInstance: host.ID, GAV: g, Source: SourceSBOM,
				SourceDetail: fmt.Sprintf("SBOM %s (sha256 %s): component %s nested in the component of %s", s.Path, s.SHA256[:12], c.label(), parentRef)})
			sum.Candidates++
		case isMaven:
			sum.Unlinked++
		}
		next := -1
		if linked {
			next = idx
		}
		for _, n := range c.Components {
			walk(n, next)
		}
	}
	if s.root.Name != "" || len(s.root.Components) > 0 {
		walk(s.root, -1)
	}
	for _, c := range s.components {
		walk(c, -1)
	}
	if sum.Unlinked > 0 {
		sum.Issues = append(sum.Issues, fmt.Sprintf("%d Maven component(s) are not tied to any analyzed jar by SHA-256 nor nested in one that is: not used (a listed dependency does not prove its code is in the input)", sum.Unlinked))
	}
	sort.Strings(sum.Issues)
}
