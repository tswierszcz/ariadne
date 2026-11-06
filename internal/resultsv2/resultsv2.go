// Package resultsv2 builds the opt-in v2 results document
// (`-json-format v2`, CVE:F1c, docs/RESULTS_CONTRACT.MD §5): components,
// one finding per advisory × affected component instance with every axis
// of §1 kept separate, coverage, and a reference to the analyzed inputs.
// v1 stays the default; ExportV1 derives v1 verdicts from a document with
// the same aggregation the engine uses (reach.Aggregate).
package resultsv2

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/findings"
	"ariadne/internal/httpcache"
	"ariadne/internal/mavenversion"
	"ariadne/internal/reach"
	"ariadne/internal/relocate"
	"ariadne/internal/snapshot"
	"ariadne/internal/vulnmap"
)

// SchemaVersion is the v2 document's schema_version.
const SchemaVersion = "2"

// Document is the whole v2 output.
type Document struct {
	SchemaVersion string      `json:"schema_version"`
	Tool          Tool        `json:"tool"`
	Snapshot      Snapshot    `json:"snapshot"`
	Components    []Component `json:"components"`
	Advisories    []Advisory  `json:"advisories"`
	Findings      []Finding   `json:"findings"`
	Coverage      Coverage    `json:"coverage"`
}

// Tool identifies the analyzer build.
type Tool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Snapshot references what was analyzed. CVE:F3 adds a reproducible data
// snapshot (advisory and fix-commit responses); until then the reference
// covers the inputs only: ID is a SHA-256 over every artifact's stable
// location and content hash, so it is the same for the same input files
// wherever they were unpacked.
type Snapshot struct {
	Kind           string `json:"kind"` // "inputs"
	ID             string `json:"id"`
	Offline        bool   `json:"offline"`
	ClasspathOrder string `json:"classpath_order"`
	// DataSet is the data snapshot of the run (CVE:F3), when one was
	// recorded (-snapshot-out) or replayed (-snapshot-in).
	DataSet *DataSet `json:"data_set,omitempty"`
}

// DataSet references a data snapshot by its logical data-set ID.
type DataSet struct {
	ID            string `json:"id"`
	Replayed      bool   `json:"replayed"`
	ConfigMatches bool   `json:"config_matches"`
}

// Component is one component instance together with its artifact.
type Component struct {
	ID         string     `json:"id"`
	ArtifactID string     `json:"artifact_id"`
	Location   string     `json:"location"`
	Role       string     `json:"role"`
	SHA256     string     `json:"sha256"`
	Bytes      int64      `json:"bytes"`
	GAV        string     `json:"gav,omitempty"`
	Identity   string     `json:"identity"` // identified | unidentified | conflict
	Detection  string     `json:"detection"`
	Evidence   []Evidence `json:"evidence,omitempty"`
	Classes    int        `json:"classes"`
	// Shadowed counts this artifact's classes that another copy won under
	// the classpath order; ShadowedDiffering those whose content differs.
	Shadowed          int `json:"shadowed,omitempty"`
	ShadowedDiffering int `json:"shadowed_differing,omitempty"`
	// EmbeddedIn is, for an embedded instance (CVE:F6a, detection
	// "embedded"), the instance of the jar it is inside.
	EmbeddedIn string `json:"embedded_in,omitempty"`
	// Relocations: the package relocations under which an embedded
	// instance's classes sit in its host (CVE:F7a) — original prefix,
	// prefix in the artifact, and where the rule came from.
	Relocations []Relocation `json:"relocations,omitempty"`
	// VersionCandidates (CVE:F6b): an embedded instance whose classes fit
	// several published versions — its version is one of these (identity
	// "version_candidates", gav "group:artifact:{v1,v2}").
	VersionCandidates []string `json:"version_candidates,omitempty"`
}

// Relocation is one package relocation (CVE:F7a).
type Relocation struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Source string `json:"source"`
}

func relocations(rs []relocate.Rule) []Relocation {
	var out []Relocation
	for _, r := range rs {
		out = append(out, Relocation{From: r.From, To: r.To, Source: r.Source})
	}
	return out
}

// EmbeddedCandidate is one pom.properties inside a jar other than the
// jar's own identity, and what was concluded (CVE:F6a).
type EmbeddedCandidate struct {
	Host             string `json:"host"` // host instance ID
	GAV              string `json:"gav"`
	Path             string `json:"path"`
	Kind             string `json:"kind"`             // module | library
	Status           string `json:"status,omitempty"` // library: confirmed | content_differs | partial | classes_absent | reference_unavailable
	Reason           string `json:"reason,omitempty"`
	ReferenceURL     string `json:"reference_url,omitempty"`
	ReferenceSHA256  string `json:"reference_sha256,omitempty"`
	ReferenceClasses int    `json:"reference_classes,omitempty"`
	Present          int    `json:"present,omitempty"`
	Identical        int    `json:"identical,omitempty"`
	Instance         string `json:"instance,omitempty"` // the instance created when confirmed
	// Declared: named or relocated by -embedded-components (CVE:F7a);
	// Relocations: the rules its classes were found under.
	Declared    bool         `json:"declared,omitempty"`
	Relocations []Relocation `json:"relocations,omitempty"`
	// CVE:F6b: where the candidate was named (pom.properties | declared |
	// build-config | sbom) and how; the version its classes confirm
	// (when not the declared one, or none was declared), the candidate
	// versions they fit, every version compared, and how that set was
	// chosen.
	Source            string         `json:"source,omitempty"`
	SourceDetail      string         `json:"source_detail,omitempty"`
	Version           string         `json:"version,omitempty"`
	VersionCandidates []string       `json:"version_candidates,omitempty"`
	VersionsChecked   []VersionCheck `json:"versions_checked,omitempty"`
	VersionsPublished int            `json:"versions_published,omitempty"`
	VersionSearch     string         `json:"version_search,omitempty"`
}

// VersionCheck is one published version compared with a candidate's
// classes (CVE:F6b).
type VersionCheck struct {
	Version          string `json:"version"`
	ReferenceSHA256  string `json:"reference_sha256,omitempty"`
	ReferenceClasses int    `json:"reference_classes"`
	Present          int    `json:"present"`
	Identical        int    `json:"identical"`
	Fits             bool   `json:"fits"`
	Declared         bool   `json:"declared,omitempty"`
	Error            string `json:"error,omitempty"`
}

// SBOM is what -sbom contributed (CVE:F6b).
type SBOM struct {
	Path        string   `json:"path"`
	SHA256      string   `json:"sha256"`
	Format      string   `json:"format"`
	SpecVersion string   `json:"spec_version"`
	Serial      string   `json:"serial,omitempty"`
	Components  int      `json:"components"`
	Linked      int      `json:"linked"`
	Candidates  int      `json:"candidates"`
	Unlinked    int      `json:"unlinked"`
	Issues      []string `json:"issues,omitempty"`
}

// Evidence is one identity fact of a component.
type Evidence struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
	GAV    string `json:"gav,omitempty"`
}

// Advisory is one deduplicated advisory and its method mapping.
type Advisory struct {
	ID       string            `json:"id"`
	Aliases  []string          `json:"aliases,omitempty"`
	Summary  string            `json:"summary,omitempty"`
	Severity string            `json:"severity,omitempty"`
	Affected []AffectedPackage `json:"affected"`
	Mapping  *Mapping          `json:"mapping"` // nil: no mapping from any source
}

// AffectedPackage is one package/version the advisory was reported for.
type AffectedPackage struct {
	Package string `json:"package"`
	Version string `json:"version"`
}

// Mapping is the mapping-precision axis (contract §1).
type Mapping struct {
	Methods []string `json:"methods"`
	// MatchPrecision is how selectors match graph methods: "name" — class
	// and method name, every overload (no descriptor); CVE:F4 adds
	// descriptor-exact selectors.
	MatchPrecision string `json:"match_precision"`
	Source         string `json:"source"` // manual | diff | llm (the evidence kind)
	Role           string `json:"role,omitempty"`
	Coverage       string `json:"coverage"`
	Review         string `json:"review"`
	Model          string `json:"model,omitempty"`
	Evidence       string `json:"evidence,omitempty"` // LLM: fix-diff or advisory text
	// Variants: a versioned mapping's method lists (CVE:F4); each finding
	// uses the one covering its instance's version.
	Variants []MappingVariant `json:"variants,omitempty"`
	// MethodRoles: an LLM mapping's per-method roles (vulnerable /
	// entry_point); bytecode presence does not confirm either.
	MethodRoles map[string]string `json:"method_roles,omitempty"`
	// FixCommits: a diff mapping's per-commit evidence (CVE:F5) — base
	// revision ("parent" or "new_side_only" with its reason) and, per
	// file, the methods the fix changed, removed or added. Added methods
	// are not in Methods: they are not code of the vulnerable release.
	FixCommits []vulnmap.FixCommit `json:"fix_commits,omitempty"`
	// NewSideOnly: some fix commit or file was read on the post-fix side
	// only (reason MAPPING_NEW_SIDE_ONLY on the v1 row).
	NewSideOnly bool `json:"new_side_only,omitempty"`
}

// MappingVariant is one versioned method list.
type MappingVariant struct {
	Package  string   `json:"package,omitempty"`
	Versions string   `json:"versions"`
	Methods  []string `json:"methods"`
}

// ApplicabilityCheck is CVE:F4's own evaluation of the advisory's OSV
// ranges for the instance's version (mavenversion, Maven 3.9 ordering),
// next to OSV's query answer: "in", "out", "undetermined", or "none" when
// the records carry no range for the package.
type ApplicabilityCheck struct {
	OSVRanges string `json:"osv_ranges"`
	Agrees    bool   `json:"agrees"`
}

// Finding is one advisory × component instance.
type Finding struct {
	AdvisoryID          string `json:"advisory_id"`
	ComponentInstanceID string `json:"component_instance_id"`
	Identity            string `json:"identity"`
	// Applicability: OSV is queried with the instance's exact version, so
	// a finding exists only for a version reported as affected; the
	// version-range check of CVE:F4 adds not_affected/undetermined.
	Applicability      string                `json:"applicability"`
	ApplicabilityCheck *ApplicabilityCheck   `json:"applicability_check,omitempty"`
	MatchPrecision     string                `json:"match_precision,omitempty"`
	Reachability       string                `json:"reachability"`
	ReasonCodes        []findings.ReasonCode `json:"reason_codes,omitempty"`
	ModelGaps          map[string]int        `json:"model_gaps,omitempty"`
	Path               []string              `json:"path,omitempty"`
	Entry              *reach.EntryEdge      `json:"entry,omitempty"`
	Note               string                `json:"note,omitempty"`
}

// Coverage is what the findings do not say per instance.
type Coverage struct {
	GlobalGaps GlobalGaps `json:"global_gaps"`
	// UnidentifiedComponents are instances without a GAV: no advisory can
	// be looked up for them, so they have no findings.
	UnidentifiedComponents []string `json:"unidentified_components,omitempty"`
	// EmbeddedCandidates are all embedded-component candidates (CVE:F6a).
	// A library candidate that is not confirmed has no instance and no
	// advisory lookup: its advisories are not in this document.
	EmbeddedCandidates []EmbeddedCandidate `json:"embedded_candidates,omitempty"`
	// EmbeddedDeclarationIssues are -embedded-components entries that did
	// not apply (no analyzed jar with that hash, hash mismatch).
	EmbeddedDeclarationIssues []string `json:"embedded_declaration_issues,omitempty"`
	// SBOM (CVE:F6b): the -sbom file and what was taken from it.
	SBOM *SBOM `json:"sbom,omitempty"`
	// AdvisoriesWithoutInstance were reported for a package/version no
	// component instance carries; their verdict is the engine's
	// whole-graph one (the v1 row).
	AdvisoriesWithoutInstance []V1Row       `json:"advisories_without_instance,omitempty"`
	Counts                    Counts        `json:"counts"`
	DataFreshness             DataFreshness `json:"data_freshness"`
	// WithdrawnAdvisories are not analyzed: no finding is made for them.
	WithdrawnAdvisories []WithdrawnAdvisory `json:"withdrawn_advisories,omitempty"`
	// SnapshotChanges: with -snapshot-compare, why derived results were
	// invalidated or verdicts changed since that snapshot.
	SnapshotChanges []snapshot.Change `json:"snapshot_changes,omitempty"`
}

// GlobalGaps are the run's model gaps (contract §4), reported once.
type GlobalGaps struct {
	UnresolvedSites int      `json:"unresolved_sites"`
	Reachable       int      `json:"reachable"`
	Unbounded       int      `json:"unbounded"`
	ModelAnalysis   []string `json:"model_analysis,omitempty"`
}

// Counts reports unique advisories and per-component findings separately
// (contract §5).
type Counts struct {
	Advisories             int            `json:"advisories"`
	Findings               int            `json:"findings"`
	AdvisoriesByV1Status   map[string]int `json:"advisories_by_v1_status"`
	FindingsByReachability map[string]int `json:"findings_by_reachability"`
}

// V1Row is one advisory's v1 verdict.
type V1Row struct {
	CVE     string                `json:"cve"`
	Status  string                `json:"status"`
	Reasons []findings.ReasonCode `json:"reasons,omitempty"`
}

// Inputs is everything Build needs.
type Inputs struct {
	ToolVersion string
	Offline     bool
	Inventory   *component.Inventory
	CVEs        []cves.CVE
	Mappings    map[string]vulnmap.Entry
	Results     []reach.Result
	Gaps        GlobalGaps
	Freshness   DataFreshness
	Withdrawn   []cves.CVE
	DataSet     *DataSet
	Changes     []snapshot.Change
}

// DataFreshness is the run's data-freshness state (CVE:F2): State is
// report.Freshness* ("complete", "stale", "unknown"), Kinds the per-kind
// counts of what the cache served.
type DataFreshness struct {
	State  string                `json:"state"`
	TTL    string                `json:"ttl"`
	MaxAge string                `json:"max_data_age,omitempty"`
	Kinds  []httpcache.Freshness `json:"kinds"`
}

// WithdrawnAdvisory is an advisory whose every source record is withdrawn.
type WithdrawnAdvisory struct {
	ID       string            `json:"id"`
	Aliases  []string          `json:"aliases,omitempty"`
	Affected []AffectedPackage `json:"affected"`
	Sources  []cves.Source     `json:"sources"`
}

// Build assembles the document. Components keep inventory order;
// advisories are sorted by ID and findings by advisory, then inventory
// order — the output is byte-identical across runs on the same inputs.
func Build(in Inputs) Document {
	doc := Document{
		SchemaVersion: SchemaVersion,
		Tool:          Tool{Name: "ariadne", Version: in.ToolVersion},
		Components:    []Component{},
		Advisories:    []Advisory{},
		Findings:      []Finding{},
	}
	doc.Coverage.GlobalGaps = in.Gaps
	doc.Coverage.DataFreshness = in.Freshness
	doc.Coverage.SnapshotChanges = in.Changes
	if doc.Coverage.DataFreshness.Kinds == nil {
		doc.Coverage.DataFreshness.Kinds = []httpcache.Freshness{}
	}
	for _, w := range in.Withdrawn {
		wa := WithdrawnAdvisory{ID: w.ID, Aliases: w.Aliases, Affected: []AffectedPackage{}, Sources: w.Sources}
		for _, a := range w.Affected {
			wa.Affected = append(wa.Affected, AffectedPackage{Package: a.Package, Version: a.Version})
		}
		doc.Coverage.WithdrawnAdvisories = append(doc.Coverage.WithdrawnAdvisories, wa)
	}

	identity := map[string]string{}
	gavOf := map[string][2]string{} // instance -> package, version
	candidatesOf := map[string][]string{}
	if inv := in.Inventory; inv != nil {
		doc.Snapshot = Snapshot{Kind: "inputs", ID: inputsID(inv), Offline: in.Offline, ClasspathOrder: inv.ClasspathOrder, DataSet: in.DataSet}
		differing := map[string]int{}
		for _, c := range inv.Conflicts {
			if !c.SameContent {
				differing[c.ShadowedInstance]++
			}
		}
		arts := map[string]component.Artifact{}
		for _, a := range inv.Artifacts {
			arts[a.ID] = a
		}
		for _, inst := range inv.Instances {
			a := arts[inst.ArtifactID]
			c := Component{
				ID: inst.ID, ArtifactID: a.ID, Location: inv.InstanceLocation(inst.ID), Role: string(a.Role), SHA256: a.SHA256, Bytes: a.Bytes,
				Identity: "identified", Detection: string(inst.Detection), Classes: inst.Classes,
				Shadowed: inst.Shadowed, ShadowedDiffering: differing[inst.ID], EmbeddedIn: inst.HostInstance,
				Relocations: relocations(inst.Relocations), VersionCandidates: inst.VersionCandidates,
			}
			if inst.GAV != nil {
				c.GAV = inst.Label()
				if len(inst.VersionCandidates) > 0 {
					c.Identity = "version_candidates"
				}
			} else if a.Role != component.RoleApp {
				c.Identity = "unidentified"
				doc.Coverage.UnidentifiedComponents = append(doc.Coverage.UnidentifiedComponents, inst.ID)
			}
			if differing[inst.ID] > 0 {
				c.Identity = "conflict"
			}
			for _, ev := range inst.Evidence {
				e := Evidence{Kind: ev.Kind, Detail: ev.Detail}
				if ev.GAV != nil {
					e.GAV = ev.GAV.String()
				}
				c.Evidence = append(c.Evidence, e)
			}
			identity[inst.ID] = c.Identity
			if inst.GAV != nil {
				gavOf[inst.ID] = [2]string{inst.GAV.PackageName(), inst.GAV.Version}
			}
			if len(inst.VersionCandidates) > 0 {
				candidatesOf[inst.ID] = inst.VersionCandidates
			}
			doc.Components = append(doc.Components, c)
		}
		for _, e := range inv.Embedded {
			ec := EmbeddedCandidate{
				Host: e.HostInstance, GAV: e.GAV.String(), Path: e.POMPath, Kind: e.Kind, Status: e.Status, Reason: e.Reason,
				ReferenceURL: e.ReferenceURL, ReferenceSHA256: e.ReferenceSHA256, ReferenceClasses: e.ReferenceClasses,
				Present: e.Present, Identical: e.Identical, Instance: e.Instance,
				Declared: e.Declared, Relocations: relocations(e.Relocations),
				Source: e.Source, SourceDetail: e.SourceDetail, VersionCandidates: e.VersionCandidates,
				VersionsPublished: e.VersionsPublished, VersionSearch: e.VersionSearch,
			}
			if e.GAV.Version == "" {
				ec.GAV = e.GAV.PackageName()
			}
			if e.Version != e.GAV.Version {
				ec.Version = e.Version
			}
			if len(e.VersionsChecked) > 1 || (len(e.VersionsChecked) == 1 && !e.VersionsChecked[0].Declared) {
				for _, vc := range e.VersionsChecked {
					ec.VersionsChecked = append(ec.VersionsChecked, VersionCheck{Version: vc.Version, ReferenceSHA256: vc.ReferenceSHA256,
						ReferenceClasses: vc.ReferenceClasses, Present: vc.Present, Identical: vc.Identical, Fits: vc.Fits, Declared: vc.Declared, Error: vc.Error})
				}
			}
			doc.Coverage.EmbeddedCandidates = append(doc.Coverage.EmbeddedCandidates, ec)
		}
		doc.Coverage.EmbeddedDeclarationIssues = inv.DeclarationIssues
		if sb := inv.SBOM; sb != nil {
			doc.Coverage.SBOM = &SBOM{Path: sb.Path, SHA256: sb.SHA256, Format: sb.Format, SpecVersion: sb.SpecVersion, Serial: sb.Serial,
				Components: sb.Components, Linked: sb.Linked, Candidates: sb.Candidates, Unlinked: sb.Unlinked, Issues: sb.Issues}
		}
	}

	byID := map[string]reach.Result{}
	for _, r := range in.Results {
		byID[r.CVE] = r
	}
	list := append([]cves.CVE(nil), in.CVEs...)
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	counts := Counts{AdvisoriesByV1Status: map[string]int{}, FindingsByReachability: map[string]int{}}
	for _, cve := range list {
		r, ok := byID[cve.ID]
		if !ok {
			continue
		}
		adv := Advisory{ID: cve.ID, Aliases: cve.Aliases, Summary: cve.Summary, Severity: cve.Severity, Affected: []AffectedPackage{}}
		for _, a := range cve.Affected {
			adv.Affected = append(adv.Affected, AffectedPackage{Package: a.Package, Version: a.Version})
		}
		if e, ok := in.Mappings[cve.ID]; ok && len(e.Methods) > 0 {
			e = e.WithDefaults()
			methods := append([]string(nil), e.Methods...)
			sort.Strings(methods)
			adv.Mapping = &Mapping{
				Methods: methods, MatchPrecision: vulnmap.MappingPrecision(methods), Source: e.Source,
				Role: e.Role, Coverage: e.Coverage, Review: e.Review, Model: e.Model, Evidence: e.Evidence,
				MethodRoles: e.MethodRoles, FixCommits: e.FixCommits, NewSideOnly: e.Source == vulnmap.SourceDiff && e.NewSideOnly(),
			}
			for _, v := range e.Variants {
				adv.Mapping.Variants = append(adv.Mapping.Variants, MappingVariant{Package: v.Package, Versions: v.Versions, Methods: v.Methods})
			}
		}
		doc.Advisories = append(doc.Advisories, adv)
		counts.Advisories++
		counts.AdvisoriesByV1Status[string(r.Status)]++

		if len(r.Findings) == 0 {
			doc.Coverage.AdvisoriesWithoutInstance = append(doc.Coverage.AdvisoriesWithoutInstance,
				V1Row{CVE: r.CVE, Status: string(r.Status), Reasons: instanceReasons(r.Reasons)})
			continue
		}
		for _, f := range r.Findings {
			applicability, check := "affected", checkApplicability(cve, gavOf[f.Instance])
			if vs := candidatesOf[f.Instance]; len(vs) > 0 {
				applicability, check = candidateApplicability(cve, gavOf[f.Instance][0], vs), nil
			}
			doc.Findings = append(doc.Findings, Finding{
				AdvisoryID: cve.ID, ComponentInstanceID: f.Instance, Identity: identity[f.Instance],
				Applicability: applicability, ApplicabilityCheck: check,
				MatchPrecision: f.MatchPrecision, Reachability: string(f.Status), ReasonCodes: f.Reasons,
				ModelGaps: f.ModelGaps, Path: f.Path, Entry: f.Entry, Note: f.Note,
			})
			counts.Findings++
			counts.FindingsByReachability[string(f.Status)]++
		}
	}
	doc.Coverage.Counts = counts
	return doc
}

// candidateApplicability (CVE:F6b): a version-candidates instance is
// affected only when OSV reports every candidate version as affected;
// when only some are, it is undetermined (which version it is, is not
// known).
func candidateApplicability(c cves.CVE, pkg string, versions []string) string {
	affected := map[string]bool{}
	for _, a := range c.Affected {
		if a.Package == pkg {
			affected[a.Version] = true
		}
	}
	for _, v := range versions {
		if !affected[v] {
			return "undetermined"
		}
	}
	return "affected"
}

// checkApplicability evaluates every range of the advisory's active source
// records for the instance's package: in when any says in, out when every
// range says out, otherwise undetermined; nil without an instance version.
func checkApplicability(c cves.CVE, gav [2]string) *ApplicabilityCheck {
	if gav[1] == "" {
		return nil
	}
	n, in, out := 0, false, true
	for _, s := range c.Sources {
		if s.Withdrawn != "" {
			continue
		}
		for _, r := range s.Ranges {
			if r.Package != gav[0] {
				continue
			}
			n++
			switch mavenversion.EvalOSV(r.Type, r.Events, gav[1]) {
			case mavenversion.In:
				in = true
			case mavenversion.Undetermined:
				out = false
			}
		}
	}
	switch {
	case n == 0:
		return &ApplicabilityCheck{OSVRanges: "none", Agrees: true}
	case in:
		return &ApplicabilityCheck{OSVRanges: "in", Agrees: true}
	case out:
		return &ApplicabilityCheck{OSVRanges: "out", Agrees: false}
	}
	return &ApplicabilityCheck{OSVRanges: "undetermined", Agrees: false}
}

// instanceReasons drops the mapping-axis codes: in v2 the mapping's role
// and coverage are fields of Mapping, not reasons of a reachability
// verdict.
func instanceReasons(rs []findings.ReasonCode) []findings.ReasonCode {
	var out []findings.ReasonCode
	for _, c := range rs {
		if !findings.MappingQualifier(c) {
			out = append(out, c)
		}
	}
	return out
}

// inputsID hashes every artifact's location and content, in sorted order.
func inputsID(inv *component.Inventory) string {
	lines := make([]string, 0, len(inv.Artifacts))
	for _, a := range inv.Artifacts {
		lines = append(lines, a.Location+"\t"+a.SHA256)
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ExportV1 derives every advisory's v1 verdict from the document (contract
// §5): reach.Aggregate over its findings, or the recorded verdict of an
// advisory without an instance. Reasons are the instance-axis codes;
// the v1 row additionally carries the mapping-axis codes.
func ExportV1(doc Document) []V1Row {
	byAdvisory := map[string][]reach.InstanceFinding{}
	for _, f := range doc.Findings {
		byAdvisory[f.AdvisoryID] = append(byAdvisory[f.AdvisoryID], reach.InstanceFinding{
			Instance: f.ComponentInstanceID, Status: reach.Status(f.Reachability), Reasons: f.ReasonCodes,
		})
	}
	without := map[string]V1Row{}
	for _, r := range doc.Coverage.AdvisoriesWithoutInstance {
		without[r.CVE] = r
	}
	rows := make([]V1Row, 0, len(doc.Advisories))
	for _, a := range doc.Advisories {
		if fs, ok := byAdvisory[a.ID]; ok {
			st, rs := reach.Aggregate(fs)
			rows = append(rows, V1Row{CVE: a.ID, Status: string(st), Reasons: rs})
			continue
		}
		rows = append(rows, without[a.ID])
	}
	return rows
}

// CheckV1 compares ExportV1(doc) with the engine's own v1 results and
// returns one line per advisory whose status differs (nil when none).
func CheckV1(doc Document, results []reach.Result) []string {
	want := map[string]reach.Result{}
	for _, r := range results {
		want[r.CVE] = r
	}
	var out []string
	for _, row := range ExportV1(doc) {
		if r := want[row.CVE]; string(r.Status) != row.Status {
			out = append(out, fmt.Sprintf("%s: v1 %s, exported from v2 %s", row.CVE, r.Status, row.Status))
		}
	}
	return out
}

// Write writes doc as indented JSON.
func Write(path string, doc Document) error {
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling v2 results: %w", err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
