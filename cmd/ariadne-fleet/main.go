// Command ariadne-fleet is the prototype of the fleet report
// (RECOMMENDATION/FLEET_CI_PLAN.MD, point 3): it reads the ariadne results
// collected from many CI pipelines — one directory per artifact, holding a
// v2 results file and an optional meta.json — and writes one Markdown
// report with separate sections for libraries and applications.
//
// Usage: ariadne-fleet -in <dir> [-out fleet.md]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	in := flag.String("in", "", "directory with one subdirectory per artifact (ariadne-results/ of each pipeline)")
	out := flag.String("out", "", "Markdown report to write (default: standard output)")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "ariadne-fleet: -in is required")
		os.Exit(2)
	}
	arts, err := Load(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ariadne-fleet:", err)
		os.Exit(1)
	}
	w := io.Writer(os.Stdout)
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ariadne-fleet:", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}
	Render(w, arts)
}

// Meta is ariadne-results/meta.json, written by the CI step.
type Meta struct {
	Artifact string `json:"artifact"`
	Commit   string `json:"commit"`
	Branch   string `json:"branch"`
	Status   string `json:"status"` // ok | failed
	Reason   string `json:"reason"`
}

// doc is the part of a v2 results document the report reads.
type doc struct {
	SchemaVersion string `json:"schema_version"`
	Tool          struct {
		Version string `json:"version"`
	} `json:"tool"`
	Advisories []struct {
		ID       string `json:"id"`
		Severity string `json:"severity"`
	} `json:"advisories"`
	Findings []struct {
		Advisory   string            `json:"advisory_id"`
		Instance   string            `json:"component_instance_id"`
		Reach      string            `json:"reachability"`
		SuppliedBy string            `json:"supplied_by"`
		Reasons    []string          `json:"reason_codes"`
		ExposedAPI []json.RawMessage `json:"exposed_api"`
	} `json:"findings"`
	CallGraph struct {
		Roots string `json:"roots"`
	} `json:"callgraph"`
	Coverage struct {
		Freshness struct {
			Kinds []struct {
				Kind   string `json:"kind"`
				Oldest string `json:"oldest_validated_at"`
			} `json:"kinds"`
		} `json:"data_freshness"`
		API *struct {
			FrameworkAdded []string `json:"framework_packages_added"`
			Alt            *struct {
				Deps []struct {
					Instance string `json:"instance"`
					Versions []struct {
						NotAssessed string `json:"not_assessed"`
						Findings    []struct {
							Advisory string `json:"advisory"`
						} `json:"findings"`
					} `json:"versions"`
				} `json:"dependencies"`
			} `json:"alternate_versions"`
		} `json:"api"`
	} `json:"coverage"`
}

// Artifact is one analyzed artifact (or a failed analysis).
type Artifact struct {
	Name     string
	Meta     Meta
	Library  bool
	Failed   string // reason; "" when the analysis succeeded
	Tool     string
	OSVDate  string // oldest validation of the OSV data, "" when unknown
	Counts   map[string]int
	Reach    map[string]bool // REACHABLE advisories
	Severity map[string]string
	// Library-only figures.
	ReachLibrary, ReachOther, SuppliedOnly, MaxExposed int
	SpringFramework                                    bool
	FixedByLine, AltNotAssessed                        int
}

const reachable = "REACHABLE"

// Load reads every artifact directory under dir.
func Load(dir string) ([]Artifact, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []Artifact
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		a, ok, err := loadOne(filepath.Join(dir, e.Name()), e.Name())
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func loadOne(dir, name string) (Artifact, bool, error) {
	a := Artifact{Name: name, Counts: map[string]int{}, Reach: map[string]bool{}, Severity: map[string]string{}}
	if b, err := os.ReadFile(filepath.Join(dir, "meta.json")); err == nil {
		if err := json.Unmarshal(b, &a.Meta); err != nil {
			return a, false, fmt.Errorf("%s/meta.json: %w", name, err)
		}
		if a.Meta.Artifact != "" {
			a.Name = a.Meta.Artifact
		}
	}
	if a.Meta.Status == "failed" {
		a.Failed = orDefault(a.Meta.Reason, "no reason given")
		return a, true, nil
	}
	d, found, err := findDoc(dir)
	if err != nil {
		return a, false, fmt.Errorf("%s: %w", name, err)
	}
	if !found {
		if a.Meta.Status == "" {
			return a, false, nil // not an artifact directory
		}
		a.Failed = "no v2 results file"
		return a, true, nil
	}
	a.Library = d.CallGraph.Roots == "api"
	a.Tool = d.Tool.Version
	for _, k := range d.Coverage.Freshness.Kinds {
		// A zero time means no OSV answer had a validation time.
		if k.Kind == "osv" && k.Oldest != "" && !strings.HasPrefix(k.Oldest, "0001-") {
			a.OSVDate = k.Oldest
		}
	}
	for _, adv := range d.Advisories {
		a.Severity[adv.ID] = adv.Severity
	}
	base := map[string]map[string]bool{}
	for _, f := range d.Findings {
		a.Counts[f.Reach]++
		if base[f.Instance] == nil {
			base[f.Instance] = map[string]bool{}
		}
		base[f.Instance][f.Advisory] = true
		if f.Reach != reachable {
			continue
		}
		a.Reach[f.Advisory] = true
		if !a.Library {
			continue
		}
		if f.SuppliedBy == "library" {
			a.ReachLibrary++
		} else {
			a.ReachOther++
		}
		for _, r := range f.Reasons {
			if r == "SUPPLIED_DISPATCH_ONLY" {
				a.SuppliedOnly++
			}
		}
		if len(f.ExposedAPI) > a.MaxExposed {
			a.MaxExposed = len(f.ExposedAPI)
		}
	}
	if api := d.Coverage.API; api != nil {
		a.SpringFramework = len(api.FrameworkAdded) > 0
		if api.Alt != nil {
			for _, dep := range api.Alt.Deps {
				for _, v := range dep.Versions {
					if v.NotAssessed != "" {
						a.AltNotAssessed++
						continue
					}
					still := map[string]bool{}
					for _, f := range v.Findings {
						still[f.Advisory] = true
					}
					for adv := range base[dep.Instance] {
						if !still[adv] {
							a.FixedByLine++
						}
					}
				}
			}
		}
	}
	return a, true, nil
}

// findDoc returns the v2 document of an artifact directory: results.json,
// else the first *.json file with schema_version "2".
func findDoc(dir string) (doc, bool, error) {
	names := []string{filepath.Join(dir, "results.json")}
	m, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	sort.Strings(m)
	names = append(names, m...)
	for _, n := range names {
		if filepath.Base(n) == "meta.json" {
			continue
		}
		b, err := os.ReadFile(n)
		if err != nil {
			continue
		}
		var d doc
		if json.Unmarshal(b, &d) != nil || d.SchemaVersion != "2" {
			continue
		}
		return d, true, nil
	}
	return doc{}, false, nil
}

var statuses = []string{"REACHABLE", "FRAMEWORK", "NOT_REACHABLE", "UNKNOWN"}

var sevRank = map[string]int{"CRITICAL": 4, "HIGH": 3, "MODERATE": 2, "MEDIUM": 2, "LOW": 1}

// Render writes the fleet report.
func Render(w io.Writer, arts []Artifact) {
	fmt.Fprintln(w, "# Fleet report (ariadne)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "> **How to read the verdicts.** `REACHABLE`: a static path exists — a priority, not proof of exploitation. "+
		"`FRAMEWORK`: the path runs through framework internals — review by hand. `NOT_REACHABLE`: no path in the model — not proof of safety (docs/SUPPORT.MD). "+
		"`UNKNOWN`: no method mapping. For a **library**, `REACHABLE` means a user of its public API can reach the vulnerable code, not that an application does; "+
		"asynchronous hand-offs are not execution traces, public internal classes count as API, and Spring starter findings are mostly `FRAMEWORK`.")
	fmt.Fprintln(w)
	var libs, apps, failed []Artifact
	for _, a := range arts {
		switch {
		case a.Failed != "":
			failed = append(failed, a)
		case a.Library:
			libs = append(libs, a)
		default:
			apps = append(apps, a)
		}
	}
	fmt.Fprintf(w, "%d artifacts: %d libraries, %d applications, %d failed analyses.\n\n", len(arts), len(libs), len(apps), len(failed))

	fmt.Fprintln(w, "## Libraries")
	fmt.Fprintln(w)
	if len(libs) == 0 {
		fmt.Fprintln(w, "None.")
	} else {
		fmt.Fprintln(w, "| Library | REACHABLE | FRAMEWORK | NOT_REACHABLE | UNKNOWN | REACHABLE: library supplies the version / the user does | supplied-object only | max API methods per finding | Spring as framework | advisories fixed by the newest in the line |")
		fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|---|")
		for _, a := range libs {
			alt := "—"
			if a.FixedByLine > 0 || a.AltNotAssessed > 0 {
				alt = fmt.Sprint(a.FixedByLine)
				if a.AltNotAssessed > 0 {
					alt += fmt.Sprintf(" (%d versions not assessed)", a.AltNotAssessed)
				}
			}
			fmt.Fprintf(w, "| %s | %s | %d / %d | %d | %d | %s | %s |\n", a.Name, countCells(a), a.ReachLibrary, a.ReachOther,
				a.SuppliedOnly, a.MaxExposed, yesNo(a.SpringFramework), alt)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## Applications")
	fmt.Fprintln(w)
	if len(apps) == 0 {
		fmt.Fprintln(w, "None.")
	} else {
		fmt.Fprintln(w, "| Application | REACHABLE | FRAMEWORK | NOT_REACHABLE | UNKNOWN |")
		fmt.Fprintln(w, "|---|---|---|---|---|")
		for _, a := range apps {
			fmt.Fprintf(w, "| %s | %s |\n", a.Name, countCells(a))
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## REACHABLE advisories across the fleet")
	fmt.Fprintln(w)
	type row struct {
		adv, sev string
		arts     []string
	}
	byAdv := map[string]*row{}
	for _, a := range append(append([]Artifact{}, libs...), apps...) {
		for adv := range a.Reach {
			r := byAdv[adv]
			if r == nil {
				r = &row{adv: adv}
				byAdv[adv] = r
			}
			r.arts = append(r.arts, a.Name)
			if sevRank[a.Severity[adv]] > sevRank[r.sev] {
				r.sev = a.Severity[adv]
			}
		}
	}
	var rows []*row
	for _, r := range byAdv {
		sort.Strings(r.arts)
		rows = append(rows, r)
	}
	sort.Slice(rows, func(i, j int) bool {
		if len(rows[i].arts) != len(rows[j].arts) {
			return len(rows[i].arts) > len(rows[j].arts)
		}
		if sevRank[rows[i].sev] != sevRank[rows[j].sev] {
			return sevRank[rows[i].sev] > sevRank[rows[j].sev]
		}
		return rows[i].adv < rows[j].adv
	})
	if len(rows) == 0 {
		fmt.Fprintln(w, "None.")
	} else {
		fmt.Fprintln(w, "| Advisory | Severity | Artifacts | Where |")
		fmt.Fprintln(w, "|---|---|---|---|")
		for _, r := range rows {
			fmt.Fprintf(w, "| %s | %s | %d | %s |\n", r.adv, orDefault(r.sev, "—"), len(r.arts), strings.Join(r.arts, ", "))
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## Failed analyses")
	fmt.Fprintln(w)
	if len(failed) == 0 {
		fmt.Fprintln(w, "None.")
	} else {
		for _, a := range failed {
			fmt.Fprintf(w, "- **%s**: %s\n", a.Name, a.Failed)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## Tool versions and data")
	fmt.Fprintln(w)
	tools := map[string][]string{}
	dates := map[string]int{}
	for _, a := range arts {
		if a.Failed != "" {
			continue
		}
		tools[orDefault(a.Tool, "unknown")] = append(tools[orDefault(a.Tool, "unknown")], a.Name)
		if a.OSVDate != "" {
			dates[a.OSVDate[:10]]++
		}
	}
	var tv []string
	for t, names := range tools {
		tv = append(tv, fmt.Sprintf("`%s` (%d)", t, len(names)))
	}
	sort.Strings(tv)
	var dv []string
	for d, n := range dates {
		dv = append(dv, fmt.Sprintf("%s (%d)", d, n))
	}
	sort.Strings(dv)
	fmt.Fprintf(w, "- Tool versions: %s\n", orDefault(strings.Join(tv, ", "), "—"))
	fmt.Fprintf(w, "- Oldest OSV data per artifact (date: artifacts): %s\n", orDefault(strings.Join(dv, ", "), "—"))
	if len(tools) > 1 || len(dates) > 1 {
		fmt.Fprintln(w, "- **The artifacts were assessed with different tool versions or OSV data dates; compare their results with care.**")
	}
}

func countCells(a Artifact) string {
	var c []string
	for _, s := range statuses {
		c = append(c, fmt.Sprint(a.Counts[s]))
	}
	return strings.Join(c, " | ")
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
