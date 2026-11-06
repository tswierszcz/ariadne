package component

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"ariadne/internal/input"
	"ariadne/internal/relocate"
)

// Declarations is the -embedded-components file (CVE:F7a): per jar,
// pinned to its SHA-256, the components the user says are embedded in it
// and the package relocations they sit under. A declaration is a claim:
// it is verified against the classes exactly like a pom.properties
// candidate, and marked as declared in the evidence.
//
//	version: 1
//	jars:
//	  - sha256: 3f1c…            # the jar's SHA-256 (required)
//	    name: ehcache-3.9.2.jar   # optional, only for diagnostics
//	    components:
//	      - gav: org.terracotta:offheap-store:2.5.1
//	        relocations:
//	          - from: org.terracotta.offheapstore
//	            to: org.ehcache.shadow.org.terracotta.offheapstore
type Declarations struct {
	Version int               `yaml:"version"`
	Jars    []DeclarationsJar `yaml:"jars"`
}

// DeclarationsJar is one jar's entry.
type DeclarationsJar struct {
	SHA256     string                  `yaml:"sha256"`
	Name       string                  `yaml:"name"`
	Components []DeclarationsComponent `yaml:"components"`
}

// DeclarationsComponent is one declared embedded component.
type DeclarationsComponent struct {
	GAV         string                   `yaml:"gav"`
	Relocations []DeclarationsRelocation `yaml:"relocations"`
}

// DeclarationsRelocation is one package relocation.
type DeclarationsRelocation struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// LoadDeclarations reads and validates path ("" = none).
func LoadDeclarations(path string) (*Declarations, error) {
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading -embedded-components %s: %w", path, err)
	}
	var d Declarations
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("parsing -embedded-components %s: %w", path, err)
	}
	if d.Version != 1 {
		return nil, fmt.Errorf("-embedded-components %s: version %d, want 1", path, d.Version)
	}
	for i, j := range d.Jars {
		if len(j.SHA256) != 64 || strings.Trim(strings.ToLower(j.SHA256), "0123456789abcdef") != "" {
			return nil, fmt.Errorf("-embedded-components %s: jars[%d].sha256 %q is not a SHA-256", path, i, j.SHA256)
		}
		for k, c := range j.Components {
			if _, err := parseGAV(c.GAV); err != nil {
				return nil, fmt.Errorf("-embedded-components %s: jars[%d].components[%d]: %w", path, i, k, err)
			}
			var rules []relocate.Rule
			for _, r := range c.Relocations {
				rules = append(rules, relocate.Rule{From: r.From, To: r.To})
			}
			if _, err := relocate.New(rules); err != nil {
				return nil, fmt.Errorf("-embedded-components %s: jars[%d].components[%d]: %w", path, i, k, err)
			}
		}
	}
	return &d, nil
}

func parseGAV(s string) (input.GAV, error) {
	p := strings.Split(s, ":")
	if len(p) != 3 || p[0] == "" || p[1] == "" || p[2] == "" {
		return input.GAV{}, fmt.Errorf("gav %q: want group:artifact:version", s)
	}
	return input.GAV{Group: p[0], Artifact: p[1], Version: p[2]}, nil
}

// applyDeclarations attaches declared components to the candidates of
// the jar with the declared SHA-256 — a known candidate (same GAV) gets
// the declared relocations, an unknown one is added as a declared
// candidate — and records declarations that match no analyzed jar. It
// returns the declared rules per host instance: a relocation is a
// property of the jar (shading renames every reference to the package),
// so it applies to every candidate of that jar.
func (inv *Inventory) applyDeclarations(d *Declarations) map[string][]relocate.Rule {
	out := map[string][]relocate.Rule{}
	if d == nil {
		return out
	}
	for _, j := range d.Jars {
		sha := strings.ToLower(j.SHA256)
		var host *Instance
		for i := range inv.Instances {
			in := &inv.Instances[i]
			if in.Detection == DetectedApp || in.Detection == DetectedEmbedded {
				continue
			}
			if a := inv.artifact(in.ArtifactID); a != nil && a.SHA256 == sha {
				host = in
				break
			}
		}
		if host == nil {
			issue := fmt.Sprintf("declaration for sha256 %s ignored: no analyzed dependency jar has that hash", sha)
			if j.Name != "" {
				for _, a := range inv.Artifacts {
					if filepath.Base(a.Path) == j.Name || strings.HasSuffix(a.Location, "/"+j.Name) || strings.HasSuffix(a.Location, ":"+j.Name) {
						issue = fmt.Sprintf("declaration for %s ignored: hash mismatch (declared %s, analyzed %s)", j.Name, sha, a.SHA256)
					}
				}
			}
			inv.DeclarationIssues = append(inv.DeclarationIssues, issue)
			continue
		}
		for _, c := range j.Components {
			g, _ := parseGAV(c.GAV)
			var rules []relocate.Rule
			for _, r := range c.Relocations {
				rules = append(rules, relocate.Rule{From: r.From, To: r.To, Source: "declared (-embedded-components)"})
			}
			idx := -1
			for i, e := range inv.Embedded {
				if e.HostInstance == host.ID && e.GAV == g {
					idx = i
				}
			}
			if idx < 0 {
				inv.Embedded = append(inv.Embedded, Embedded{HostInstance: host.ID, GAV: g})
				idx = len(inv.Embedded) - 1
			}
			inv.Embedded[idx].Declared = true
			out[host.ID] = append(out[host.ID], rules...)
		}
	}
	return out
}
