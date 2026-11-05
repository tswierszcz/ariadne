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
	"ariadne/internal/reach"
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
}

// Finding is one advisory × component instance.
type Finding struct {
	AdvisoryID          string `json:"advisory_id"`
	ComponentInstanceID string `json:"component_instance_id"`
	Identity            string `json:"identity"`
	// Applicability: OSV is queried with the instance's exact version, so
	// a finding exists only for a version reported as affected; the
	// version-range check of CVE:F4 adds not_affected/undetermined.
	Applicability string                `json:"applicability"`
	Reachability  string                `json:"reachability"`
	ReasonCodes   []findings.ReasonCode `json:"reason_codes,omitempty"`
	ModelGaps     map[string]int        `json:"model_gaps,omitempty"`
	Path          []string              `json:"path,omitempty"`
	Entry         *reach.EntryEdge      `json:"entry,omitempty"`
	Note          string                `json:"note,omitempty"`
}

// Coverage is what the findings do not say per instance.
type Coverage struct {
	GlobalGaps GlobalGaps `json:"global_gaps"`
	// UnidentifiedComponents are instances without a GAV: no advisory can
	// be looked up for them, so they have no findings.
	UnidentifiedComponents []string `json:"unidentified_components,omitempty"`
	// AdvisoriesWithoutInstance were reported for a package/version no
	// component instance carries; their verdict is the engine's
	// whole-graph one (the v1 row).
	AdvisoriesWithoutInstance []V1Row       `json:"advisories_without_instance,omitempty"`
	Counts                    Counts        `json:"counts"`
	DataFreshness             DataFreshness `json:"data_freshness"`
	// WithdrawnAdvisories are not analyzed: no finding is made for them.
	WithdrawnAdvisories []WithdrawnAdvisory `json:"withdrawn_advisories,omitempty"`
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
	if inv := in.Inventory; inv != nil {
		doc.Snapshot = Snapshot{Kind: "inputs", ID: inputsID(inv), Offline: in.Offline, ClasspathOrder: inv.ClasspathOrder}
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
				ID: inst.ID, ArtifactID: a.ID, Location: a.Location, Role: string(a.Role), SHA256: a.SHA256, Bytes: a.Bytes,
				Identity: "identified", Detection: string(inst.Detection), Classes: inst.Classes,
				Shadowed: inst.Shadowed, ShadowedDiffering: differing[inst.ID],
			}
			if inst.GAV != nil {
				c.GAV = inst.GAV.String()
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
			doc.Components = append(doc.Components, c)
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
				Methods: methods, MatchPrecision: "name", Source: e.Source,
				Role: e.Role, Coverage: e.Coverage, Review: e.Review, Model: e.Model, Evidence: e.Evidence,
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
			doc.Findings = append(doc.Findings, Finding{
				AdvisoryID: cve.ID, ComponentInstanceID: f.Instance, Identity: identity[f.Instance],
				Applicability: "affected", Reachability: string(f.Status), ReasonCodes: f.Reasons,
				ModelGaps: f.ModelGaps, Path: f.Path, Entry: f.Entry, Note: f.Note,
			})
			counts.Findings++
			counts.FindingsByReachability[string(f.Status)]++
		}
	}
	doc.Coverage.Counts = counts
	return doc
}

// instanceReasons drops the mapping-axis codes: in v2 the mapping's role
// and coverage are fields of Mapping, not reasons of a reachability
// verdict.
func instanceReasons(rs []findings.ReasonCode) []findings.ReasonCode {
	var out []findings.ReasonCode
	for _, c := range rs {
		if c != findings.MappingPartial && c != findings.MappingEntrypointOnly {
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
