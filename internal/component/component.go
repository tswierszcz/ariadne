// Package component is CVE:F1a's model of what was analyzed: physical
// artifacts (jars, by content hash and stable location) and the logical
// component instances found in them, with the evidence for each identity
// (docs/RESULTS_CONTRACT.MD, "Component identity" axis). It is an
// inventory: nothing here changes reachability or verdicts. Findings per
// advisory × instance are CVE:F1b/F1c.
//
// Identifiers are derived from content and a stable location inside the
// input — never from a temp-directory path — so preparing the same input
// twice yields the same identifiers.
package component

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/relocate"
)

// Role is an artifact's role in the analyzed input.
type Role string

const (
	RoleApp        Role = "app"
	RoleDependency Role = "dependency"
)

// Detection names how an instance's identity was established.
type Detection string

const (
	DetectedPOMProperties Detection = "pom.properties"
	DetectedDepsList      Detection = "deps-list"
	DetectedWebJar        Detection = "webjar"
	DetectedApp           Detection = "app"
	DetectedUnidentified  Detection = "unidentified"
)

// Artifact is one physical jar that was read.
type Artifact struct {
	ID string
	// Location is stable across runs: "app:<input file name>" for the
	// application, "deps:<file name>" for a jar from -deps,
	// "app:<input file name>!/<entry>" for a jar extracted from the
	// application archive (BOOT-INF/lib, WEB-INF/lib, an EAR module).
	Location string
	Role     Role
	SHA256   string
	Bytes    int64
	// Path is where the jar was read from in this run — informational
	// only (it may be a temp file) and never part of an identifier.
	Path string
}

// IdentityEvidence is one fact supporting (or competing with) an
// instance's identity.
type IdentityEvidence struct {
	Kind   string // "pom.properties", "deps-list", "embedded-candidate", "webjar", "app-input"
	Detail string // e.g. "META-INF/maven/g/a/pom.properties", "deps.txt", the candidate's GAV
	GAV    *input.GAV
}

// Instance is one logical component instance. CVE:F1a assigns one per
// artifact; additional pom.properties found in the same jar are recorded
// as embedded-candidate evidence, and CVE:F6a adds an instance for each
// candidate confirmed by its classes (ResolveEmbedded), owning those
// classes.
type Instance struct {
	ID         string
	ArtifactID string
	GAV        *input.GAV // nil when unidentified or the app
	Detection  Detection
	Evidence   []IdentityEvidence
	// Classes is how many classes of this instance were selected into the
	// analyzed hierarchy; Packages their distinct packages, sorted.
	Classes  int
	Packages []string
	// Shadowed counts classes of this instance that another copy of the
	// same FQCN won under the declared classpath order (not in the graph).
	Shadowed int
	// HostInstance and EmbeddedPath are set on an embedded instance
	// (CVE:F6a): the instance of the jar it is inside, and its
	// "META-INF/maven/<group>/<artifact>" directory there.
	HostInstance string
	EmbeddedPath string
	// Relocations are the package relocations under which an embedded
	// instance's classes sit in its host (CVE:F7a): a method selector of
	// the original component is rewritten with them.
	Relocations []relocate.Rule
	// VersionCandidates (CVE:F6b) is set on an embedded instance whose
	// classes fit several published versions: GAV.Version is then empty
	// and the version is one of these, not known which.
	VersionCandidates []string
}

// Label is the instance's GAV for display: "group:artifact:version", or
// "group:artifact:{v1,v2}" for a version-candidates instance; "" when
// unidentified.
func (in Instance) Label() string {
	if in.GAV == nil {
		return ""
	}
	if len(in.VersionCandidates) > 0 {
		return in.GAV.PackageName() + ":{" + strings.Join(in.VersionCandidates, ",") + "}"
	}
	return in.GAV.String()
}

// Conflict is one FQCN present in several artifacts: the copy selected
// into the graph and one copy that was not (hierarchy.Duplicates). The
// non-selected copy is inventoried, never used as a reachability proof.
type Conflict struct {
	FQCN                     string
	KeptInstance             string
	ShadowedInstance         string
	KeptSHA256, ShadowSHA256 string
	SameContent              bool

	keptJar, shadowJar string // run paths, for class-level owners
}

// Inventory is the run's component model.
type Inventory struct {
	Artifacts []Artifact
	Instances []Instance
	Conflicts []Conflict
	// Embedded are the embedded-component candidates (CVE:F6a), in
	// artifact order; ResolveEmbedded sets their kind and status.
	Embedded []Embedded
	// DeclarationIssues are -embedded-components entries that did not
	// apply (unknown jar, hash mismatch); BuildConfigNotes the shade
	// relocations read from, or skipped in, embedded pom.xml files.
	DeclarationIssues []string
	BuildConfigNotes  []string
	// SBOM is what -sbom contributed (CVE:F6b), nil without one.
	SBOM *SBOMSummary
	// ClasspathOrder documents the order assumption conflicts follow.
	ClasspathOrder string
	byJar          map[string]string // run path -> instance ID
	byClass        map[string]string // run path + "\x00" + FQCN -> embedded instance ID
}

// InstanceForJar returns the instance ID of the artifact read from path
// in this run (a ClassInfo.SourceJar), or "".
func (inv *Inventory) InstanceForJar(path string) string { return inv.byJar[path] }

// InstanceForClass returns the instance a selected class belongs to.
func (inv *Inventory) InstanceForClass(ci *hierarchy.ClassInfo) string {
	return inv.instanceFor(ci.SourceJar, ci.FQCN)
}

// Counts returns artifacts, identified instances, unidentified instances
// and conflicts, for the CLI log line.
func (inv *Inventory) Counts() (artifacts, identified, unidentified, conflicts int) {
	for _, in := range inv.Instances {
		switch in.Detection {
		case DetectedUnidentified:
			unidentified++
		case DetectedApp:
		default:
			identified++
		}
	}
	return len(inv.Artifacts), identified, unidentified, len(inv.Conflicts)
}

// Inputs describes where the analyzed jars came from.
type Inputs struct {
	AppInput   string // the -app value as given
	AppJar     string // the jar hierarchy.Build read the app from
	DepsDir    string // directory the dependency jars were listed from
	Extracted  bool   // DepsDir is the app archive's extracted lib dir
	LibEntries map[string]string
	DepJars    []input.DepJar
}

// Build constructs the inventory from the prepared inputs and the built
// hierarchy (for class assignment and conflicts).
func Build(in Inputs, h *hierarchy.Hierarchy) (*Inventory, error) {
	inv := &Inventory{
		byJar:          map[string]string{},
		ClasspathOrder: "application first, then dependency jars in the order listed from the dependency directory (sorted by file name); the first copy of a class wins",
	}
	appName := filepath.Base(strings.TrimRight(in.AppInput, string(os.PathSeparator)))

	app, err := newArtifact(in.AppJar, "app:"+appName, RoleApp)
	if err != nil {
		return nil, err
	}
	inv.add(app, Instance{Detection: DetectedApp, Evidence: []IdentityEvidence{{Kind: "app-input", Detail: in.AppInput}}})

	for _, dj := range in.DepJars {
		base := filepath.Base(dj.Path)
		loc := "deps:" + base
		if in.Extracted {
			entry := in.LibEntries[base]
			if entry == "" {
				entry = base
			}
			loc = "app:" + appName + "!/" + entry
		}
		a, err := newArtifact(dj.Path, loc, RoleDependency)
		if err != nil {
			return nil, err
		}
		inst := Instance{GAV: dj.GAV}
		gavs, err := input.JarPOMProperties(dj.Path)
		if err != nil {
			return nil, err
		}
		var embedded []input.GAV
		switch {
		case dj.WebAssetsOnly:
			inst.Detection = DetectedWebJar
			inst.Evidence = append(inst.Evidence, IdentityEvidence{Kind: "webjar", Detail: "META-INF/resources/webjars/ only, no .class"})
		case dj.GAV == nil:
			inst.Detection = DetectedUnidentified
		case dj.Source == input.SourcePOMProperties:
			inst.Detection = DetectedPOMProperties
		default:
			inst.Detection = DetectedDepsList
		}
		if dj.GAV != nil {
			if dj.Source == input.SourcePOMProperties {
				g := *dj.GAV
				inst.Evidence = append(inst.Evidence, IdentityEvidence{Kind: "pom.properties", Detail: fmt.Sprintf("META-INF/maven/%s/%s/pom.properties", g.Group, g.Artifact), GAV: &g})
			} else {
				g := *dj.GAV
				inst.Evidence = append(inst.Evidence, IdentityEvidence{Kind: "deps-list", Detail: "deps.txt entry matched by file name", GAV: &g})
			}
		}
		for _, g := range gavs {
			g := g
			if dj.GAV != nil && g == *dj.GAV {
				continue
			}
			inst.Evidence = append(inst.Evidence, IdentityEvidence{Kind: "embedded-candidate", Detail: g.String(), GAV: &g})
			embedded = append(embedded, g)
		}
		inv.add(a, inst)
		host := inv.Instances[len(inv.Instances)-1].ID
		for _, g := range embedded {
			inv.Embedded = append(inv.Embedded, Embedded{HostInstance: host, GAV: g,
				POMPath: fmt.Sprintf("META-INF/maven/%s/%s/pom.properties", g.Group, g.Artifact)})
		}
	}

	for _, d := range h.Duplicates() {
		inv.Conflicts = append(inv.Conflicts, Conflict{
			FQCN: d.FQCN, KeptSHA256: d.Kept.SHA256, ShadowSHA256: d.Shadowed.SHA256,
			SameContent: d.Kept.SHA256 != "" && d.Kept.SHA256 == d.Shadowed.SHA256,
			keptJar:     d.Kept.Jar, shadowJar: d.Shadowed.Jar,
		})
	}
	inv.recount(h)
	return inv, nil
}

func (inv *Inventory) add(a Artifact, inst Instance) {
	identity := "unidentified"
	switch {
	case inst.Detection == DetectedApp:
		identity = "app"
	case inst.GAV != nil:
		identity = inst.GAV.String()
	}
	inst.ArtifactID = a.ID
	inst.ID = "c-" + shortHash(a.ID, identity)
	inv.Artifacts = append(inv.Artifacts, a)
	inv.Instances = append(inv.Instances, inst)
	inv.byJar[a.Path] = inst.ID
}

func newArtifact(path, location string, role Role) (Artifact, error) {
	f, err := os.Open(path)
	if err != nil {
		return Artifact{}, fmt.Errorf("hashing %s: %w", path, err)
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return Artifact{}, fmt.Errorf("hashing %s: %w", path, err)
	}
	sum := hex.EncodeToString(h.Sum(nil))
	return Artifact{ID: "a-" + shortHash(sum, location), Location: location, Role: role, SHA256: sum, Bytes: n, Path: path}, nil
}

func shortHash(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (inv *Inventory) artifact(id string) *Artifact {
	for i := range inv.Artifacts {
		if inv.Artifacts[i].ID == id {
			return &inv.Artifacts[i]
		}
	}
	return nil
}

// ArtifactLocation is the stable location of artifact id (report helper).
func (inv *Inventory) ArtifactLocation(id string) string {
	if a := inv.artifact(id); a != nil {
		return a.Location
	}
	return ""
}

// ArtifactShortHash is the first 12 hex digits of artifact id's SHA-256.
func (inv *Inventory) ArtifactShortHash(id string) string {
	if a := inv.artifact(id); a != nil && len(a.SHA256) >= 12 {
		return a.SHA256[:12]
	}
	return ""
}
