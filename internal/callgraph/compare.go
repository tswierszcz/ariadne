package callgraph

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// GRAPH_PRECISION P0: a graph is written as stable, sorted text (-graph-out)
// and two of them are compared edge by edge (-graph-compare). An edge's
// identity is its source and target method keys, its call kind and its
// basis — never a node number — so the same graph built twice, or by two
// strategies, compares by meaning. The findings' paths are written with
// it, so a comparison says which verdicts changed and through which edges.

// ExportFormat versions the -graph-out format.
const ExportFormat = "ariadne-callgraph-1"

// EdgeID is an edge's stable identity.
type EdgeID struct {
	From, To, Kind, Basis string
}

func (e EdgeID) String() string {
	return e.From + " -> " + e.To + " [" + e.Kind + "/" + e.Basis + "]"
}

// FindingPath is one finding's verdict and path (node keys) for the export.
type FindingPath struct {
	ID     string // advisory (and instance) identity
	Status string
	Path   []string
}

// Export is a graph as written by Write and read by Read.
type Export struct {
	Config   Config
	Edges    map[EdgeID][]int // PCs (evidence), sorted
	Findings map[string]FindingPath
}

// Snapshot captures g (after every model edge was added) with findings.
func Snapshot(g *Graph, findings []FindingPath) Export {
	ex := Export{Config: g.Config(), Edges: map[EdgeID][]int{}, Findings: map[string]FindingPath{}}
	for from := range g.adjacency {
		for _, e := range g.adjacency[from] {
			id := EdgeID{From: g.Key(uint32(from)), To: g.Key(e.To), Kind: string(e.Kind), Basis: string(e.Basis)}
			pcs := ex.Edges[id]
			for _, s := range e.Sites {
				pcs = append(pcs, s.PC)
			}
			sort.Ints(pcs)
			ex.Edges[id] = dedupInts(pcs)
		}
	}
	for _, f := range findings {
		ex.Findings[f.ID] = f
	}
	return ex
}

func dedupInts(xs []int) []int {
	out := xs[:0]
	for i, x := range xs {
		if i == 0 || x != xs[i-1] {
			out = append(out, x)
		}
	}
	if out == nil {
		return []int{}
	}
	return out
}

// Write writes ex as sorted, tab-separated text: a format line, the
// config (JSON), one "edge" line per edge and one "finding" line per
// finding. Two runs of the same analysis write identical bytes.
func Write(w io.Writer, ex Export) error {
	bw := bufio.NewWriter(w)
	cfg, err := json.Marshal(ex.Config)
	if err != nil {
		return err
	}
	fmt.Fprintf(bw, "# %s\n", ExportFormat)
	fmt.Fprintf(bw, "config\t%s\n", cfg)
	ids := make([]EdgeID, 0, len(ex.Edges))
	for id := range ex.Edges {
		ids = append(ids, id)
	}
	sortEdgeIDs(ids)
	for _, id := range ids {
		pcs := make([]string, len(ex.Edges[id]))
		for i, pc := range ex.Edges[id] {
			pcs[i] = strconv.Itoa(pc)
		}
		fmt.Fprintf(bw, "edge\t%s\t%s\t%s\t%s\t%s\n", id.From, id.To, id.Kind, id.Basis, strings.Join(pcs, ","))
	}
	fids := make([]string, 0, len(ex.Findings))
	for id := range ex.Findings {
		fids = append(fids, id)
	}
	sort.Strings(fids)
	for _, id := range fids {
		f := ex.Findings[id]
		fmt.Fprintf(bw, "finding\t%s\t%s\t%s\n", f.ID, f.Status, strings.Join(f.Path, " > "))
	}
	return bw.Flush()
}

// Read parses what Write wrote.
func Read(r io.Reader) (Export, error) {
	ex := Export{Edges: map[EdgeID][]int{}, Findings: map[string]FindingPath{}}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	line := 0
	sawFormat := false
	for sc.Scan() {
		line++
		t := sc.Text()
		if strings.HasPrefix(t, "# ") {
			if strings.TrimPrefix(t, "# ") != ExportFormat {
				return ex, fmt.Errorf("line %d: format %q, want %q", line, strings.TrimPrefix(t, "# "), ExportFormat)
			}
			sawFormat = true
			continue
		}
		f := strings.Split(t, "\t")
		switch f[0] {
		case "config":
			if len(f) != 2 || json.Unmarshal([]byte(f[1]), &ex.Config) != nil {
				return ex, fmt.Errorf("line %d: bad config", line)
			}
		case "edge":
			if len(f) != 6 {
				return ex, fmt.Errorf("line %d: bad edge", line)
			}
			var pcs []int
			if f[5] != "" {
				for _, p := range strings.Split(f[5], ",") {
					n, err := strconv.Atoi(p)
					if err != nil {
						return ex, fmt.Errorf("line %d: bad pc %q", line, p)
					}
					pcs = append(pcs, n)
				}
			}
			if pcs == nil {
				pcs = []int{}
			}
			ex.Edges[EdgeID{From: f[1], To: f[2], Kind: f[3], Basis: f[4]}] = pcs
		case "finding":
			if len(f) != 4 {
				return ex, fmt.Errorf("line %d: bad finding", line)
			}
			fp := FindingPath{ID: f[1], Status: f[2]}
			if f[3] != "" {
				fp.Path = strings.Split(f[3], " > ")
			}
			ex.Findings[fp.ID] = fp
		default:
			return ex, fmt.Errorf("line %d: unknown record %q", line, f[0])
		}
	}
	if err := sc.Err(); err != nil {
		return ex, err
	}
	if !sawFormat {
		return ex, fmt.Errorf("not an %s file", ExportFormat)
	}
	return ex, nil
}

func sortEdgeIDs(ids []EdgeID) {
	sort.Slice(ids, func(i, j int) bool {
		a, b := ids[i], ids[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Basis < b.Basis
	})
}

// FindingChange is one finding whose verdict or path differs.
type FindingChange struct {
	ID          string `json:"id"`
	Old         string `json:"old_status"`
	New         string `json:"new_status"`
	PathChanged bool   `json:"path_changed"`
	// Because lists the added edges on the new path and the removed
	// edges of the old path (as "from -> to"): what explains the change.
	Because []string `json:"because,omitempty"`
}

// Comparison explains how a graph differs from an earlier one.
type Comparison struct {
	// ConfigDifferences: the comparison is not between equal settings
	// unless this is empty; it is always disclosed.
	ConfigDifferences []string       `json:"config_differences,omitempty"`
	EdgesBefore       int            `json:"edges_before"`
	EdgesAfter        int            `json:"edges_after"`
	Added             []EdgeID       `json:"-"`
	Removed           []EdgeID       `json:"-"`
	AddedByBasis      map[string]int `json:"added_by_basis"`
	RemovedByBasis    map[string]int `json:"removed_by_basis"`
	// EvidenceChanged counts edges present in both with other PCs.
	EvidenceChanged int             `json:"evidence_changed"`
	Findings        []FindingChange `json:"findings_changed,omitempty"`
	// FindingsOnlyIn: findings present in one export only (another
	// advisory set or input), by side.
	FindingsOnlyBefore []string `json:"findings_only_before,omitempty"`
	FindingsOnlyAfter  []string `json:"findings_only_after,omitempty"`
}

// Empty reports whether nothing differs.
func (c Comparison) Empty() bool {
	return len(c.ConfigDifferences) == 0 && len(c.Added) == 0 && len(c.Removed) == 0 && c.EvidenceChanged == 0 &&
		len(c.Findings) == 0 && len(c.FindingsOnlyBefore) == 0 && len(c.FindingsOnlyAfter) == 0
}

// Compare explains after relative to before.
func Compare(before, after Export) Comparison {
	c := Comparison{ConfigDifferences: before.Config.Differences(after.Config), EdgesBefore: len(before.Edges), EdgesAfter: len(after.Edges),
		AddedByBasis: map[string]int{}, RemovedByBasis: map[string]int{}}
	pair := func(id EdgeID) string { return id.From + " -> " + id.To }
	addedPairs, removedPairs := map[string]bool{}, map[string]bool{}
	for id, pcs := range after.Edges {
		old, ok := before.Edges[id]
		if !ok {
			c.Added = append(c.Added, id)
			c.AddedByBasis[id.Basis]++
			addedPairs[pair(id)] = true
		} else if fmt.Sprint(old) != fmt.Sprint(pcs) {
			c.EvidenceChanged++
		}
	}
	for id := range before.Edges {
		if _, ok := after.Edges[id]; !ok {
			c.Removed = append(c.Removed, id)
			c.RemovedByBasis[id.Basis]++
			removedPairs[pair(id)] = true
		}
	}
	sortEdgeIDs(c.Added)
	sortEdgeIDs(c.Removed)
	for id, nf := range after.Findings {
		of, ok := before.Findings[id]
		if !ok {
			c.FindingsOnlyAfter = append(c.FindingsOnlyAfter, id)
			continue
		}
		pathChanged := strings.Join(of.Path, "\x00") != strings.Join(nf.Path, "\x00")
		if of.Status == nf.Status && !pathChanged {
			continue
		}
		fc := FindingChange{ID: id, Old: of.Status, New: nf.Status, PathChanged: pathChanged}
		seen := map[string]bool{}
		for i := 0; i+1 < len(nf.Path); i++ {
			p := nf.Path[i] + " -> " + nf.Path[i+1]
			if addedPairs[p] && !seen[p] {
				seen[p] = true
				fc.Because = append(fc.Because, "added "+p)
			}
		}
		for i := 0; i+1 < len(of.Path); i++ {
			p := of.Path[i] + " -> " + of.Path[i+1]
			if removedPairs[p] && !seen[p] {
				seen[p] = true
				fc.Because = append(fc.Because, "removed "+p)
			}
		}
		c.Findings = append(c.Findings, fc)
	}
	for id := range before.Findings {
		if _, ok := after.Findings[id]; !ok {
			c.FindingsOnlyBefore = append(c.FindingsOnlyBefore, id)
		}
	}
	sort.Slice(c.Findings, func(i, j int) bool { return c.Findings[i].ID < c.Findings[j].ID })
	sort.Strings(c.FindingsOnlyBefore)
	sort.Strings(c.FindingsOnlyAfter)
	return c
}

// SampleLimit bounds the edges a report lists per side.
const SampleLimit = 30

// AddedSample and RemovedSample are the first SampleLimit edges of each
// side, rendered, for the report.
func (c Comparison) AddedSample() []string   { return sample(c.Added) }
func (c Comparison) RemovedSample() []string { return sample(c.Removed) }

func sample(ids []EdgeID) []string {
	var out []string
	for i, id := range ids {
		if i == SampleLimit {
			break
		}
		out = append(out, id.String())
	}
	return out
}

// BasisText renders a by-basis count map ("cha 12, direct 3"), "none" when empty.
func BasisText(m map[string]int) string {
	if len(m) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + " " + strconv.Itoa(m[k])
	}
	return strings.Join(parts, ", ")
}

// AddedText and RemovedText are BasisText of each side.
func (c Comparison) AddedText() string   { return BasisText(c.AddedByBasis) }
func (c Comparison) RemovedText() string { return BasisText(c.RemovedByBasis) }
