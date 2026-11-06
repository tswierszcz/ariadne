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
// classfile.Fingerprint — a class rewritten by a shading tool without
// renaming it keeps its fingerprint. An unparsable entry is not the same.
func sameClass(a, b *zip.File) bool {
	if a.CRC32 == b.CRC32 && a.UncompressedSize64 == b.UncompressedSize64 {
		return true
	}
	fa, err := fingerprintEntry(a)
	if err != nil {
		return false
	}
	fb, err := fingerprintEntry(b)
	return err == nil && fa == fb
}

func fingerprintEntry(f *zip.File) (string, error) {
	r, err := f.Open()
	if err != nil {
		return "", err
	}
	defer r.Close()
	cf, err := classfile.Parse(r)
	if err != nil {
		return "", err
	}
	return classfile.Fingerprint(cf)
}

// classFQCN turns "a/b/C$D.class" into "a.b.C$D".
func classFQCN(entry string) string {
	return strings.ReplaceAll(strings.TrimSuffix(entry, ".class"), "/", ".")
}

// ResolveEmbedded examines every embedded candidate recorded by Build,
// confirms library candidates against their reference jars, and adds an
// instance for each confirmed one: the host jar's classes found in the
// reference are reassigned to it (class counts, packages and class
// conflicts follow). A reference that cannot be read is a per-candidate
// outcome, never an error; only an unreadable host jar is.
func (inv *Inventory) ResolveEmbedded(ctx context.Context, fetch ReferenceFetcher, base string, h *hierarchy.Hierarchy) error {
	if base == "" {
		base = CentralBaseURL
	}
	byID := map[string]int{}
	for i, in := range inv.Instances {
		byID[in.ID] = i
	}
	var added []Instance
	for i := range inv.Embedded {
		e := &inv.Embedded[i]
		host := inv.Instances[byID[e.HostInstance]]
		if host.GAV != nil && isModuleOf(e.GAV, *host.GAV) {
			e.Kind = EmbeddedModule
			e.Reason = fmt.Sprintf("module of %s (same version, related group): part of the host, no separate lookup", host.GAV)
			continue
		}
		e.Kind = EmbeddedLibrary
		a := inv.artifact(host.ArtifactID)
		hz, err := zip.OpenReader(a.Path)
		if err != nil {
			return fmt.Errorf("reading %s for embedded components: %w", a.Path, err)
		}
		hostClasses := classEntries(hz.File)

		e.ReferenceURL = ReferenceURL(base, e.GAV)
		body, err := fetch(ctx, e.ReferenceURL)
		if err != nil {
			hz.Close()
			e.Status, e.Reason = EmbeddedReferenceUnavailable, "reference jar unavailable: "+err.Error()
			continue
		}
		s := sha256.Sum256(body)
		e.ReferenceSHA256 = hex.EncodeToString(s[:])
		rz, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			hz.Close()
			e.Status, e.Reason = EmbeddedReferenceUnavailable, "reference jar unreadable: "+err.Error()
			continue
		}
		ref := classEntries(rz.File)
		e.ReferenceClasses = len(ref)
		var names []string
		for n, re := range ref {
			he, ok := hostClasses[n]
			if !ok {
				continue
			}
			e.Present++
			names = append(names, n)
			if sameClass(he, re) {
				e.Identical++
			}
		}
		hz.Close()
		sort.Strings(names)
		switch {
		case e.ReferenceClasses == 0:
			e.Status, e.Reason = EmbeddedClassesAbsent, "the reference jar has no classes"
		case e.Present == 0:
			e.Status, e.Reason = EmbeddedClassesAbsent, "none of the reference classes is in the jar: stale metadata or a relocated copy (CVE:F7)"
		case e.Present*2 < e.ReferenceClasses:
			e.Status, e.Reason = EmbeddedPartial, fmt.Sprintf("only %d of %d reference classes present", e.Present, e.ReferenceClasses)
		case e.Identical*10 < e.Present*9:
			e.Status, e.Reason = EmbeddedContentDiffers, fmt.Sprintf("%d of %d present classes differ from version %s", e.Present-e.Identical, e.Present, e.GAV.Version)
		default:
			e.Status = EmbeddedConfirmed
			e.Reason = fmt.Sprintf("%d of %d reference classes present, %d identical", e.Present, e.ReferenceClasses, e.Identical)
		}
		if e.Status != EmbeddedConfirmed {
			continue
		}
		g := e.GAV
		inst := Instance{
			ID: "c-" + shortHash(host.ArtifactID, "embedded:"+g.String()), ArtifactID: host.ArtifactID,
			GAV: &g, Detection: DetectedEmbedded, HostInstance: host.ID, EmbeddedPath: strings.TrimSuffix(e.POMPath, "/pom.properties"),
			Evidence: []IdentityEvidence{
				{Kind: "pom.properties", Detail: e.POMPath, GAV: &g},
				{Kind: "reference-classes", Detail: e.Reason + " (" + e.ReferenceURL + ")"},
			},
		}
		e.Instance = inst.ID
		if inv.byClass == nil {
			inv.byClass = map[string]string{}
		}
		for _, n := range names {
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
