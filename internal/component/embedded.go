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
)

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
// (and those decls adds), confirms library candidates against their
// reference jars, and adds an instance for each confirmed one: the host
// jar's classes found in the reference are reassigned to it (class
// counts, packages and class conflicts follow).
//
// CVE:F7a: a candidate's classes are looked up under their relocated
// names when relocation rules apply — the host jar's own maven-shade
// configuration (its embedded pom.xml) and the user's declarations
// (-embedded-components, pinned to the jar's SHA-256). Host-level rules
// also normalize references: a class that only refers to a relocated
// dependency compares equal to its original. A reference that cannot be
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
	type hostData struct {
		classes map[string]*zip.File
		zr      *zip.ReadCloser
		rules   []relocate.Rule
	}
	hosts := map[string]*hostData{}
	defer func() {
		for _, hd := range hosts {
			hd.zr.Close()
		}
	}()
	claimed := map[string]string{} // host path + "\x00" + class entry -> claiming candidate GAV
	var added []Instance
	for i := range inv.Embedded {
		e := &inv.Embedded[i]
		host := inv.Instances[byID[e.HostInstance]]
		if !e.Declared && host.GAV != nil && isModuleOf(e.GAV, *host.GAV) {
			e.Kind = EmbeddedModule
			e.Reason = fmt.Sprintf("module of %s (same version, related group): part of the host, no separate lookup", host.GAV)
			continue
		}
		e.Kind = EmbeddedLibrary
		a := inv.artifact(host.ArtifactID)
		hd := hosts[a.ID]
		if hd == nil {
			zr, err := zip.OpenReader(a.Path)
			if err != nil {
				return fmt.Errorf("reading %s for embedded components: %w", a.Path, err)
			}
			hd = &hostData{classes: classEntries(zr.File), zr: zr}
			hd.rules, inv.BuildConfigNotes = shadeRelocations(zr.File, host.GAV, inv.BuildConfigNotes)
			hosts[a.ID] = hd
		}
		rules, err := relocate.New(append(append([]relocate.Rule(nil), hd.rules...), declared[host.ID]...))
		if err != nil {
			e.Status, e.Reason = EmbeddedRelocationConflict, err.Error()
			continue
		}
		inverse := rules.Inverse()

		e.ReferenceURL = ReferenceURL(base, e.GAV)
		body, err := fetch(ctx, e.ReferenceURL)
		if err != nil {
			e.Status, e.Reason = EmbeddedReferenceUnavailable, "reference jar unavailable: "+err.Error()
			continue
		}
		sum := sha256.Sum256(body)
		e.ReferenceSHA256 = hex.EncodeToString(sum[:])
		rz, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			e.Status, e.Reason = EmbeddedReferenceUnavailable, "reference jar unreadable: "+err.Error()
			continue
		}
		ref := classEntries(rz.File)
		e.ReferenceClasses = len(ref)
		own := map[relocate.Rule]bool{}
		var images []string
		collisions := 0
		for n, re := range ref {
			orig := strings.TrimSuffix(n, ".class")
			image := rules.Name(orig) + ".class"
			he, ok := hd.classes[image]
			if !ok {
				continue
			}
			if r, under := rules.Under(rules.Name(orig)); under && image != n {
				own[r] = true
			}
			if g, taken := claimed[a.Path+"\x00"+image]; taken && g != e.GAV.String() {
				collisions++
				continue
			}
			e.Present++
			images = append(images, image)
			if sameClass(he, re, inverse) {
				e.Identical++
			}
		}
		for r := range own {
			e.Relocations = append(e.Relocations, r)
		}
		sort.Slice(e.Relocations, func(i, j int) bool { return e.Relocations[i].From < e.Relocations[j].From })
		sort.Strings(images)
		switch {
		case collisions > 0:
			e.Status, e.Reason = EmbeddedCollision, fmt.Sprintf("%d of its classes are already attributed to another embedded component", collisions)
		case e.ReferenceClasses == 0:
			e.Status, e.Reason = EmbeddedClassesAbsent, "the reference jar has no classes"
		case e.Present == 0 && rules.Empty():
			e.Status, e.Reason = EmbeddedClassesAbsent, "none of the reference classes is in the jar: stale metadata or a relocated copy (declare the relocation, CVE:F7a)"
		case e.Present == 0:
			e.Status, e.Reason = EmbeddedClassesAbsent, "none of the reference classes is in the jar, also under the known relocations"
		case e.Present*2 < e.ReferenceClasses:
			e.Status, e.Reason = EmbeddedPartial, fmt.Sprintf("only %d of %d reference classes present (a minimized copy is not proof the rest is absent or safe)", e.Present, e.ReferenceClasses)
		case e.Identical*10 < e.Present*9:
			e.Status, e.Reason = EmbeddedContentDiffers, fmt.Sprintf("%d of %d present classes differ from version %s", e.Present-e.Identical, e.Present, e.GAV.Version)
		default:
			e.Status = EmbeddedConfirmed
			e.Reason = fmt.Sprintf("%d of %d reference classes present, %d identical", e.Present, e.ReferenceClasses, e.Identical)
			if len(e.Relocations) > 0 {
				e.Reason += ", relocated"
			}
		}
		if e.Status != EmbeddedConfirmed {
			continue
		}
		for _, img := range images {
			claimed[a.Path+"\x00"+img] = e.GAV.String()
		}
		g := e.GAV
		path := strings.TrimSuffix(e.POMPath, "/pom.properties")
		var evidence []IdentityEvidence
		if e.POMPath != "" {
			evidence = append(evidence, IdentityEvidence{Kind: "pom.properties", Detail: e.POMPath, GAV: &g})
		} else {
			path = "declared/" + g.Group + "/" + g.Artifact
		}
		if e.Declared {
			evidence = append(evidence, IdentityEvidence{Kind: "declared", Detail: "-embedded-components entry for this jar's SHA-256", GAV: &g})
		}
		evidence = append(evidence, IdentityEvidence{Kind: "reference-classes", Detail: e.Reason + " (" + e.ReferenceURL + ")"})
		for _, r := range e.Relocations {
			evidence = append(evidence, IdentityEvidence{Kind: "relocation", Detail: r.From + " -> " + r.To + " (" + r.Source + ")"})
		}
		inst := Instance{
			ID: "c-" + shortHash(host.ArtifactID, "embedded:"+g.String()), ArtifactID: host.ArtifactID,
			GAV: &g, Detection: DetectedEmbedded, HostInstance: host.ID, EmbeddedPath: path,
			Evidence: evidence, Relocations: e.Relocations,
		}
		e.Instance = inst.ID
		if inv.byClass == nil {
			inv.byClass = map[string]string{}
		}
		for _, n := range images {
			inv.byClass[a.Path+"\x00"+classFQCN(n)] = inst.ID
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

// EmbeddedCounts returns confirmed library candidates, unconfirmed
// library candidates and modules (CLI log line).
func (inv *Inventory) EmbeddedCounts() (confirmed, unconfirmed, modules int) {
	for _, e := range inv.Embedded {
		switch {
		case e.Kind == EmbeddedModule:
			modules++
		case e.Status == EmbeddedConfirmed:
			confirmed++
		default:
			unconfirmed++
		}
	}
	return
}

// ConfirmedEmbeddedGAVs are the GAVs of confirmed embedded instances, for
// the advisory lookup.
func (inv *Inventory) ConfirmedEmbeddedGAVs() []input.GAV {
	var out []input.GAV
	for _, in := range inv.Instances {
		if in.Detection == DetectedEmbedded && in.GAV != nil {
			out = append(out, *in.GAV)
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
