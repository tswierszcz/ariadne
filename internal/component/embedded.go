package component

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/relocate"
)

// CVE:F6a: components embedded in a dependency jar without relocation.
// Every pom.properties of a jar other than the jar's own identity is a
// candidate. A candidate of the host's own project (same group or a
// group nested in it, same version) is a module of the host, not a
// separate library. A library candidate is confirmed only by its classes:
// the reference jar of exactly that version (Maven Central) is compared
// class by class with the host jar (same name; same bytes or the same
// serialization-independent fingerprint). Only a confirmed
// candidate becomes a component instance with its own classes and its
// own advisory lookup; every other one is listed with its status, as a
// coverage gap.

// DetectedEmbedded: an instance confirmed inside another jar.
const DetectedEmbedded Detection = "embedded"

// Embedded candidate kinds.
const (
	EmbeddedModule  = "module"  // a module of the host's own project
	EmbeddedLibrary = "library" // a separately published library
)

// Embedded candidate statuses (library candidates).
const (
	// EmbeddedConfirmed: at least half of the reference classes are in
	// the host jar and at least 90% of those are the same class
	// (byte-identical, or the same classfile.Fingerprint when a shading
	// tool rewrote the file): the version is confirmed; an instance is
	// created.
	EmbeddedConfirmed = "confirmed"
	// EmbeddedContentDiffers: the classes are there, their content is not
	// that of the declared version — the version is not confirmed.
	EmbeddedContentDiffers = "content_differs"
	// EmbeddedPartial: fewer than half of the reference classes are
	// present (a minimized copy, or unrelated classes sharing names).
	EmbeddedPartial = "partial"
	// EmbeddedClassesAbsent: metadata without classes — stale metadata,
	// or a relocated copy (explicit relocation is CVE:F7).
	EmbeddedClassesAbsent = "classes_absent"
	// EmbeddedReferenceUnavailable: the reference jar could not be read
	// (offline cache without it, not on Maven Central): no conclusion.
	EmbeddedReferenceUnavailable = "reference_unavailable"
	// EmbeddedCollision: some of its classes (under their relocated
	// names) were already attributed to another embedded component.
	EmbeddedCollision = "collision"
	// EmbeddedRelocationConflict: the applicable relocation rules
	// contradict each other (one package relocated to two places).
	EmbeddedRelocationConflict = "relocation_conflict"
	// EmbeddedVersionCandidates (CVE:F6b): the classes fit several of the
	// published versions that were checked and none is singled out — an
	// instance is created with that bounded set of candidate versions;
	// advisories of any of them are reported, not as confirmed affected.
	EmbeddedVersionCandidates = "version_candidates"
	// EmbeddedVersionUnknown (CVE:F6b): the classes are there but no
	// checked version fits (or no version could be checked): no instance,
	// no advisory lookup — a coverage gap, never an "no CVE" result.
	EmbeddedVersionUnknown = "version_unknown"
)

// Candidate sources (CVE:F6b): where a candidate was named.
const (
	SourcePOMProperties = "pom.properties" // a pom.properties inside the jar
	SourceDeclared      = "declared"       // only the -embedded-components file
	SourceBuildConfig   = "build-config"   // the jar's embedded pom.xml: shade includes ∩ dependencies
	SourceSBOM          = "sbom"           // a CycloneDX component nested in the jar's SBOM component
)

// MaxVersionSearch bounds how many other published versions are compared
// for one candidate (CVE:F6b): around the declared version, or — with no
// version at all — every published version when there are at most this
// many.
const MaxVersionSearch = 8

// VersionCheck is one published version compared with the host's
// classes (CVE:F6b).
type VersionCheck struct {
	Version          string
	ReferenceURL     string
	ReferenceSHA256  string
	ReferenceClasses int
	Present          int
	Identical        int
	Fits             bool   // meets the confirmation thresholds
	Declared         bool   // the candidate's own declared version
	Error            string // reference unavailable
}

// Embedded is one embedded-component candidate and what was concluded.
type Embedded struct {
	HostInstance string
	GAV          input.GAV
	// POMPath is the candidate's pom.properties entry in the host jar.
	POMPath string
	Kind    string // EmbeddedModule | EmbeddedLibrary
	Status  string // library candidates only
	Reason  string
	// Reference: where the reference jar was read and its hash; class
	// counts: reference classes, present in the host by name, identical.
	ReferenceURL     string
	ReferenceSHA256  string
	ReferenceClasses int
	Present          int
	Identical        int
	// Instance is the created instance (EmbeddedConfirmed only).
	Instance string
	// Declared: the candidate (or its relocations) comes from the user's
	// -embedded-components file (CVE:F7a); Relocations are the rules
	// under which its own classes were found (with their source).
	Declared    bool
	Relocations []relocate.Rule
	// CVE:F6b. Source: where the candidate was named (Source*), with a
	// detail (the pom.xml path and dependency chain, the SBOM bom-ref).
	Source       string
	SourceDetail string
	// Version is the version the classes confirm (status confirmed; it
	// may differ from GAV.Version when the declared one did not fit);
	// VersionCandidates the fitting versions (status version_candidates).
	Version           string
	VersionCandidates []string
	// VersionsChecked are every version compared, declared one first;
	// VersionsPublished how many releases the version list had (0 = not
	// read); VersionSearch says how the checked set was chosen.
	VersionsChecked   []VersionCheck
	VersionsPublished int
	VersionSearch     string
}

// ReferenceFetcher returns the body at url (a Maven Central jar), through
// the run's cache.
type ReferenceFetcher func(ctx context.Context, url string) ([]byte, error)

// CentralBaseURL is the Maven Central repository root used for
// reference jars.
const CentralBaseURL = "https://repo1.maven.org/maven2"

// ReferenceURL is the Central URL of g's main jar.
func ReferenceURL(base string, g input.GAV) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s-%s.jar", base, strings.ReplaceAll(g.Group, ".", "/"), g.Artifact, g.Version, g.Artifact, g.Version)
}

// isModuleOf: same version, and the same group or one nested in the other.
func isModuleOf(c, host input.GAV) bool {
	if c.Version != host.Version {
		return false
	}
	return c.Group == host.Group || strings.HasPrefix(c.Group, host.Group+".") || strings.HasPrefix(host.Group, c.Group+".")
}

// classEntries lists a jar's class entries by path, without
// multi-release variants and module descriptors (they belong to the jar,
// not to a component's API).
func classEntries(files []*zip.File) map[string]*zip.File {
	out := map[string]*zip.File{}
	for _, f := range files {
		n := f.Name
		if !strings.HasSuffix(n, ".class") || strings.HasPrefix(n, "META-INF/") || strings.HasSuffix(n, "module-info.class") {
			continue
		}
		out[n] = f
	}
	return out
}

// sameClass: byte-identical entries (CRC-32 and size), or the same
// classfile.Fingerprint — a class rewritten by a shading tool keeps its
// fingerprint once the host side's relocated names are mapped back
// (inverse). An unparsable entry is not the same.
func sameClass(host, ref *zip.File, inverse relocate.Set) bool {
	if host.CRC32 == ref.CRC32 && host.UncompressedSize64 == ref.UncompressedSize64 {
		return true
	}
	fa, err := fingerprintEntry(host, inverse)
	if err != nil {
		return false
	}
	fb, err := fingerprintEntry(ref, relocate.Set{})
	return err == nil && fa == fb
}

func fingerprintEntry(f *zip.File, rs relocate.Set) (string, error) {
	r, err := f.Open()
	if err != nil {
		return "", err
	}
	defer r.Close()
	cf, err := classfile.Parse(r)
	if err != nil {
		return "", err
	}
	return classfile.FingerprintWith(cf, rs)
}

// classFQCN turns "a/b/C$D.class" into "a.b.C$D".
func classFQCN(entry string) string {
	return strings.ReplaceAll(strings.TrimSuffix(entry, ".class"), "/", ".")
}

// ResolveEmbedded examines every embedded candidate recorded by Build
// (and those decls, an SBOM and the host's build configuration add),
// confirms library candidates against their reference jars, and adds an
// instance for each confirmed one: the host jar's classes found in the
// reference are reassigned to it (class counts, packages and class
// conflicts follow).
//
// CVE:F7a: a candidate's classes are looked up under their relocated
// names when relocation rules apply — the host jar's own maven-shade
// configuration (its embedded pom.xml) and the user's declarations
// (-embedded-components, pinned to the jar's SHA-256). Host-level rules
// also normalize references: a class that only refers to a relocated
// dependency compares equal to its original.
//
// CVE:F6b: a candidate whose declared version is not confirmed by its
// classes (content_differs), or that has no version at all, is compared
// with a bounded set of other published versions (searchVersions): one
// fitting version confirms it, several make a version-candidates
// instance, none leaves it version_unknown. A reference that cannot be
// read is a per-candidate outcome, never an error; only an unreadable
// host jar is.
func (inv *Inventory) ResolveEmbedded(ctx context.Context, fetch ReferenceFetcher, base string, h *hierarchy.Hierarchy, decls *Declarations) error {
	if base == "" {
		base = CentralBaseURL
	}
	declared := inv.applyDeclarations(decls)
	byID := map[string]int{}
	for i, in := range inv.Instances {
		byID[in.ID] = i
	}
	hosts := map[string]*hostData{}
	defer func() {
		for _, hd := range hosts {
			hd.zr.Close()
		}
	}()
	openHost := func(host Instance) (*hostData, error) {
		a := inv.artifact(host.ArtifactID)
		if hd := hosts[a.ID]; hd != nil {
			return hd, nil
		}
		zr, err := zip.OpenReader(a.Path)
		if err != nil {
			return nil, fmt.Errorf("reading %s for embedded components: %w", a.Path, err)
		}
		hd := &hostData{path: a.Path, classes: classEntries(zr.File), zr: zr}
		hd.rules, inv.BuildConfigNotes = shadeRelocations(zr.File, host.GAV, inv.BuildConfigNotes)
		hosts[a.ID] = hd
		return hd, nil
	}
	// CVE:F6b: candidates the host's build configuration names (shade
	// artifactSet includes ∩ the pom's dependencies) — after the
	// pom.properties and declared ones, so those win for the same package.
	for i := 0; i < len(inv.Instances); i++ {
		host := inv.Instances[i]
		if host.Detection == DetectedApp || host.Detection == DetectedEmbedded || host.GAV == nil {
			continue
		}
		hd, err := openHost(host)
		if err != nil {
			return err
		}
		inv.addBuildConfigCandidates(ctx, fetch, base, host, hd.zr.File)
	}

	cl := claims{}
	var added []Instance
	for i := range inv.Embedded {
		e := &inv.Embedded[i]
		host := inv.Instances[byID[e.HostInstance]]
		if e.Source == "" {
			e.Source = SourcePOMProperties
		}
		if e.Source == SourcePOMProperties && !e.Declared && host.GAV != nil && isModuleOf(e.GAV, *host.GAV) {
			e.Kind = EmbeddedModule
			e.Reason = fmt.Sprintf("module of %s (same version, related group): part of the host, no separate lookup", host.GAV)
			continue
		}
		e.Kind = EmbeddedLibrary
		hd, err := openHost(host)
		if err != nil {
			return err
		}
		rules, err := relocate.New(append(append([]relocate.Rule(nil), hd.rules...), declared[host.ID]...))
		if err != nil {
			e.Status, e.Reason = EmbeddedRelocationConflict, err.Error()
			continue
		}
		var best *versionCheck
		if e.GAV.Version != "" {
			c := compareVersion(ctx, fetch, base, e.GAV, hd, rules, cl)
			e.recordCheck(c, true)
			e.ReferenceURL, e.ReferenceSHA256 = c.url, c.sha
			e.ReferenceClasses, e.Present, e.Identical = c.refClasses, c.present, c.identical
			e.Status, e.Reason = c.status(e.GAV.Version, rules.Empty())
			if e.Status == EmbeddedConfirmed {
				best, e.Version = c, e.GAV.Version
			}
		} else {
			e.Status, e.Reason = EmbeddedVersionUnknown, "no version declared for this candidate"
		}
		if best == nil && (e.Status == EmbeddedContentDiffers || e.GAV.Version == "") {
			best = inv.searchVersions(ctx, fetch, base, e, hd, rules, cl)
		}
		if best == nil {
			continue
		}
		for r := range best.own {
			e.Relocations = append(e.Relocations, r)
		}
		sort.Slice(e.Relocations, func(i, j int) bool { return e.Relocations[i].From < e.Relocations[j].From })
		if len(e.Relocations) > 0 && e.Status == EmbeddedConfirmed {
			e.Reason += ", relocated"
		}
		cl.claim(hd.path, best.images, e.GAV.PackageName())
		inst := inv.embeddedInstance(e, host, best)
		e.Instance = inst.ID
		if inv.byClass == nil {
			inv.byClass = map[string]string{}
		}
		for _, n := range best.images {
			inv.byClass[hd.path+"\x00"+classFQCN(n)] = inst.ID
		}
		added = append(added, inst)
	}
	if len(added) == 0 {
		return nil
	}
	inv.Instances = append(inv.Instances, added...)
	inv.recount(h)
	return nil
}

// hostData is one host jar opened for embedded-component checks.
type hostData struct {
	path    string
	classes map[string]*zip.File
	zr      *zip.ReadCloser
	rules   []relocate.Rule
}

// claims records which candidate (package name) a confirmed instance's
// classes went to, per host jar: a class is never given to two.
type claims map[string]string

func (c claims) owner(jar, entry string) (string, bool) {
	g, ok := c[jar+"\x00"+entry]
	return g, ok
}

func (c claims) claim(jar string, entries []string, pkg string) {
	for _, n := range entries {
		c[jar+"\x00"+n] = pkg
	}
}

// versionCheck is one version's reference jar compared with the host.
type versionCheck struct {
	version                        string
	url, sha                       string
	err                            error
	refClasses, present, identical int
	collisions                     int
	images                         []string
	own                            map[relocate.Rule]bool
}

// fits: the F6a confirmation thresholds — at least half of the reference
// classes present, at least 90% of those the same class, no collision.
func (c *versionCheck) fits() bool {
	return c.err == nil && c.collisions == 0 && c.refClasses > 0 && c.present > 0 &&
		c.present*2 >= c.refClasses && c.identical*10 >= c.present*9
}

// status classifies one version's check (F6a statuses).
func (c *versionCheck) status(version string, noRules bool) (string, string) {
	switch {
	case c.err != nil:
		return EmbeddedReferenceUnavailable, c.err.Error()
	case c.collisions > 0:
		return EmbeddedCollision, fmt.Sprintf("%d of its classes are already attributed to another embedded component", c.collisions)
	case c.refClasses == 0:
		return EmbeddedClassesAbsent, "the reference jar has no classes"
	case c.present == 0 && noRules:
		return EmbeddedClassesAbsent, "none of the reference classes is in the jar: stale metadata or a relocated copy (declare the relocation, CVE:F7a)"
	case c.present == 0:
		return EmbeddedClassesAbsent, "none of the reference classes is in the jar, also under the known relocations"
	case c.present*2 < c.refClasses:
		return EmbeddedPartial, fmt.Sprintf("only %d of %d reference classes present (a minimized copy is not proof the rest is absent or safe)", c.present, c.refClasses)
	case c.identical*10 < c.present*9:
		return EmbeddedContentDiffers, fmt.Sprintf("%d of %d present classes differ from version %s", c.present-c.identical, c.present, version)
	}
	return EmbeddedConfirmed, fmt.Sprintf("%d of %d reference classes present, %d identical", c.present, c.refClasses, c.identical)
}

// compareVersion reads g's reference jar and compares it class by class
// with the host jar, under rules.
func compareVersion(ctx context.Context, fetch ReferenceFetcher, base string, g input.GAV, hd *hostData, rules relocate.Set, cl claims) *versionCheck {
	c := &versionCheck{version: g.Version, url: ReferenceURL(base, g), own: map[relocate.Rule]bool{}}
	body, err := fetch(ctx, c.url)
	if err != nil {
		c.err = fmt.Errorf("reference jar unavailable: %w", err)
		return c
	}
	sum := sha256.Sum256(body)
	c.sha = hex.EncodeToString(sum[:])
	rz, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		c.err = fmt.Errorf("reference jar unreadable: %w", err)
		return c
	}
	inverse := rules.Inverse()
	ref := classEntries(rz.File)
	c.refClasses = len(ref)
	for n, re := range ref {
		orig := strings.TrimSuffix(n, ".class")
		image := rules.Name(orig) + ".class"
		he, ok := hd.classes[image]
		if !ok {
			continue
		}
		if r, under := rules.Under(rules.Name(orig)); under && image != n {
			c.own[r] = true
		}
		if owner, taken := cl.owner(hd.path, image); taken && owner != g.PackageName() {
			c.collisions++
			continue
		}
		c.present++
		c.images = append(c.images, image)
		if sameClass(he, re, inverse) {
			c.identical++
		}
	}
	sort.Strings(c.images)
	return c
}

// recordCheck appends c to the candidate's checked versions.
func (e *Embedded) recordCheck(c *versionCheck, declared bool) {
	vc := VersionCheck{Version: c.version, ReferenceURL: c.url, ReferenceSHA256: c.sha,
		ReferenceClasses: c.refClasses, Present: c.present, Identical: c.identical, Fits: c.fits(), Declared: declared}
	if c.err != nil {
		vc.Error = c.err.Error()
	}
	e.VersionsChecked = append(e.VersionsChecked, vc)
}

// embeddedInstance builds the instance of a confirmed or
// version-candidates candidate.
func (inv *Inventory) embeddedInstance(e *Embedded, host Instance, best *versionCheck) Instance {
	g := e.GAV
	g.Version = e.Version
	var path string
	var evidence []IdentityEvidence
	switch e.Source {
	case SourcePOMProperties:
		path = strings.TrimSuffix(e.POMPath, "/pom.properties")
		pg := e.GAV
		evidence = append(evidence, IdentityEvidence{Kind: "pom.properties", Detail: e.POMPath, GAV: &pg})
	case SourceBuildConfig:
		path = "build-config/" + g.Group + "/" + g.Artifact
		pg := e.GAV
		evidence = append(evidence, IdentityEvidence{Kind: "build-config", Detail: e.SourceDetail, GAV: &pg})
	case SourceSBOM:
		path = "sbom/" + g.Group + "/" + g.Artifact
		pg := e.GAV
		evidence = append(evidence, IdentityEvidence{Kind: "sbom", Detail: e.SourceDetail, GAV: &pg})
	default:
		path = "declared/" + g.Group + "/" + g.Artifact
	}
	if e.Declared {
		dg := e.GAV
		evidence = append(evidence, IdentityEvidence{Kind: "declared", Detail: "-embedded-components entry for this jar's SHA-256", GAV: &dg})
	}
	if len(e.VersionCandidates) > 0 {
		evidence = append(evidence, IdentityEvidence{Kind: "version-candidates", Detail: e.Reason})
	} else {
		evidence = append(evidence, IdentityEvidence{Kind: "reference-classes", Detail: e.Reason + " (" + best.url + ")"})
	}
	for _, r := range e.Relocations {
		evidence = append(evidence, IdentityEvidence{Kind: "relocation", Detail: r.From + " -> " + r.To + " (" + r.Source + ")"})
	}
	key := "embedded:" + g.String()
	if len(e.VersionCandidates) > 0 {
		key = "embedded:" + g.PackageName() + ":{" + strings.Join(e.VersionCandidates, ",") + "}"
	}
	return Instance{
		ID: "c-" + shortHash(host.ArtifactID, key), ArtifactID: host.ArtifactID,
		GAV: &g, Detection: DetectedEmbedded, HostInstance: host.ID, EmbeddedPath: path,
		Evidence: evidence, Relocations: e.Relocations, VersionCandidates: e.VersionCandidates,
	}
}

// EmbeddedCounts returns confirmed library candidates, unconfirmed
// library candidates and modules (CLI log line).
func (inv *Inventory) EmbeddedCounts() (confirmed, unconfirmed, modules int) {
	for _, e := range inv.Embedded {
		switch {
		case e.Kind == EmbeddedModule:
			modules++
		case e.Status == EmbeddedConfirmed || e.Status == EmbeddedVersionCandidates:
			confirmed++
		default:
			unconfirmed++
		}
	}
	return
}

// ConfirmedEmbeddedGAVs are the GAVs of confirmed embedded instances, for
// the advisory lookup — and, for a version-candidates instance, one GAV
// per candidate version (CVE:F6b).
func (inv *Inventory) ConfirmedEmbeddedGAVs() []input.GAV {
	var out []input.GAV
	for _, in := range inv.Instances {
		if in.Detection != DetectedEmbedded || in.GAV == nil {
			continue
		}
		if len(in.VersionCandidates) == 0 {
			out = append(out, *in.GAV)
			continue
		}
		for _, v := range in.VersionCandidates {
			g := *in.GAV
			g.Version = v
			out = append(out, g)
		}
	}
	return out
}

// InstanceLocation is an instance's stable location: its artifact's, and
// for an embedded instance the host location followed by
// "!/META-INF/maven/<group>/<artifact>".
func (inv *Inventory) InstanceLocation(id string) string {
	for _, in := range inv.Instances {
		if in.ID != id {
			continue
		}
		loc := inv.ArtifactLocation(in.ArtifactID)
		if in.EmbeddedPath != "" {
			loc += "!/" + in.EmbeddedPath
		}
		return loc
	}
	return ""
}

// recount recomputes per-instance class counts, packages, conflict
// owners and shadowed counts at class granularity.
func (inv *Inventory) recount(h *hierarchy.Hierarchy) {
	byID := map[string]*Instance{}
	for i := range inv.Instances {
		in := &inv.Instances[i]
		in.Classes, in.Packages, in.Shadowed = 0, nil, 0
		byID[in.ID] = in
	}
	pkgs := map[string]map[string]bool{}
	for _, ci := range h.All() {
		id := inv.InstanceForClass(ci)
		if id == "" {
			continue
		}
		byID[id].Classes++
		if pkgs[id] == nil {
			pkgs[id] = map[string]bool{}
		}
		if dot := strings.LastIndexByte(ci.FQCN, '.'); dot > 0 {
			pkgs[id][ci.FQCN[:dot]] = true
		} else {
			pkgs[id][""] = true
		}
	}
	for id, set := range pkgs {
		for p := range set {
			byID[id].Packages = append(byID[id].Packages, p)
		}
		sort.Strings(byID[id].Packages)
	}
	for i := range inv.Conflicts {
		c := &inv.Conflicts[i]
		c.KeptInstance = inv.instanceFor(c.keptJar, c.FQCN)
		c.ShadowedInstance = inv.instanceFor(c.shadowJar, c.FQCN)
		if in := byID[c.ShadowedInstance]; in != nil {
			in.Shadowed++
		}
	}
	sort.Slice(inv.Conflicts, func(i, j int) bool {
		if inv.Conflicts[i].FQCN != inv.Conflicts[j].FQCN {
			return inv.Conflicts[i].FQCN < inv.Conflicts[j].FQCN
		}
		return inv.Conflicts[i].ShadowedInstance < inv.Conflicts[j].ShadowedInstance
	})
}

// instanceFor is the instance owning class fqcn read from jar path.
func (inv *Inventory) instanceFor(jar, fqcn string) string {
	if id, ok := inv.byClass[jar+"\x00"+fqcn]; ok {
		return id
	}
	return inv.byJar[jar]
}
