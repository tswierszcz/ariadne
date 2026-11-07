package resultsv2

import (
	"fmt"
	"sort"

	"ariadne/internal/cves"
	"ariadne/internal/deserialize"
	"ariadne/internal/taint"
)

// DESERIALIZATION Z4: the deserialization model of the run in v2, as
// additive fields (contract §5). Nothing here changes a finding's
// reachability: chain states are evidence of kind chain-candidate.

// Deserialization is the run's deserialization entries and catalog
// chains.
type Deserialization struct {
	ModelCatalog string `json:"model_catalog"`
	ChainCatalog string `json:"chain_catalog"`
	// TaintCatalog (DATA_FLOW D4, additive): the taint catalog when
	// -taint ran; absent otherwise (no entry carries a taint result).
	TaintCatalog string             `json:"taint_catalog,omitempty"`
	Entries      []DeserEntry       `json:"entries"`
	Chains       []DeserChain       `json:"chains"`
	Summary      DeserializationSum `json:"summary"`
}

// DeserializationSum counts entries by policy and chain results by state.
type DeserializationSum struct {
	EntriesByPolicy map[string]int `json:"entries_by_policy"`
	ChainsByState   map[string]int `json:"chains_by_state"`
	// EntriesByTaint (DATA_FLOW D4, -taint): entries by taint state.
	EntriesByTaint map[string]int `json:"entries_by_taint,omitempty"`
}

// DeserEntry is one entry site.
type DeserEntry struct {
	ID           string `json:"id"` // caller#name+desc@pc
	Call         string `json:"call"`
	Model        string `json:"model"`
	Version      string `json:"version,omitempty"`
	App          bool   `json:"app"`
	Policy       string `json:"policy"`
	PolicyDetail string `json:"policy_detail,omitempty"`
	Candidates   int    `json:"candidate_classes"`
	Rejected     int    `json:"rejected_classes"`
	// DataDependsOn (DATA_FLOW D3, additive): what the data the entry
	// reads depends on within its method, as dependency-summary sources
	// (param:i, this, field:…, call:… for an unmodeled call's result);
	// DataFlowIncomplete why that may be incomplete; DataFlowWidened how
	// many calls of the method were answered conservatively.
	DataDependsOn      []string `json:"data_depends_on,omitempty"`
	DataFlowIncomplete []string `json:"data_flow_incomplete,omitempty"`
	DataFlowWidened    int      `json:"data_flow_widened,omitempty"`
	// Taint (DATA_FLOW D4, -taint, additive): whether request data
	// reaches the data the entry reads — a separate flow assessment,
	// never a reachability verdict and never proof of exploitability.
	Taint *taint.Result `json:"taint,omitempty"`
}

// DeserChain is one catalog chain evaluated on one copy of its classes.
type DeserChain struct {
	ID string `json:"id"`
	// Copy is "" for the original class names, else the instance whose
	// relocations name the classes.
	Copy                string            `json:"copy,omitempty"`
	ComponentInstanceID string            `json:"component_instance_id,omitempty"`
	Format              string            `json:"format"`
	State               string            `json:"state"`
	Advisories          []string          `json:"advisories,omitempty"`
	Classes             []DeserChainClass `json:"classes,omitempty"`
	Missing             []string          `json:"missing,omitempty"`
	Blocks              []string          `json:"blocks,omitempty"`
	Unresolved          []string          `json:"unresolved,omitempty"`
	UncheckedConditions []string          `json:"unchecked_conditions,omitempty"`
	Flow                string            `json:"flow"`
	Sites               []DeserChainSite  `json:"sites,omitempty"`
	SitesByState        map[string]int    `json:"sites_by_state,omitempty"`
}

// DeserChainClass is one required class as found.
type DeserChainClass struct {
	Name                string `json:"name"`
	Found               string `json:"found"`
	ComponentInstanceID string `json:"component_instance_id,omitempty"`
	Confirmed           bool   `json:"confirmed"`
}

// DeserChainSite is a chain's result at one entry; Reasons are the
// site's own (its policy), the chain's are on the chain.
type DeserChainSite struct {
	Entry   string   `json:"entry"`
	State   string   `json:"state"`
	Reasons []string `json:"reasons,omitempty"`
}

// ChainRef links a finding to a catalog chain naming its advisory on the
// same component instance.
type ChainRef struct {
	Chain string `json:"chain"`
	Copy  string `json:"copy,omitempty"`
	State string `json:"state"`
}

func entryID(s *deserialize.Site) string {
	return fmt.Sprintf("%s#%s%s@%d", s.CallerFQCN, s.CallerName, s.CallerDesc, s.PC)
}

// buildDeserialization renders the model; nil when the run has none.
func buildDeserialization(in Inputs) *Deserialization {
	if in.DeserModelCatalog == "" {
		return nil
	}
	d := &Deserialization{ModelCatalog: in.DeserModelCatalog, ChainCatalog: in.DeserChainCatalog, TaintCatalog: in.TaintCatalog,
		Entries: []DeserEntry{}, Chains: []DeserChain{},
		Summary: DeserializationSum{EntriesByPolicy: map[string]int{}, ChainsByState: map[string]int{}}}
	for i := range in.DeserSites {
		s := &in.DeserSites[i]
		d.Entries = append(d.Entries, DeserEntry{ID: entryID(s), Call: s.Owner + "#" + s.Name + s.Desc, Model: s.Model, Version: s.Version,
			App: s.IsApp, Policy: s.Filter, PolicyDetail: s.FilterDetail, Candidates: s.Candidates, Rejected: s.Rejected,
			DataDependsOn: s.DataFromRaw, DataFlowIncomplete: s.DataFlowIncomplete, DataFlowWidened: s.DataFlowWidened})
		d.Summary.EntriesByPolicy[s.Filter]++
		if in.TaintCatalog != "" && s.Taint != nil {
			d.Entries[len(d.Entries)-1].Taint = s.Taint
			if d.Summary.EntriesByTaint == nil {
				d.Summary.EntriesByTaint = map[string]int{}
			}
			d.Summary.EntriesByTaint[s.Taint.State]++
		}
	}
	for _, p := range in.DeserChains {
		c := DeserChain{ID: p.Spec.ID, Copy: p.Copy, ComponentInstanceID: p.Instance, Format: p.Spec.Format, State: p.State,
			Advisories: p.Spec.Advisories, Missing: p.Missing, Blocks: p.Blocks, Unresolved: p.Unresolved,
			UncheckedConditions: p.Spec.Conditions, Flow: deserialize.FlowNotAnalyzed}
		for _, cc := range p.Classes {
			if cc.Found == "" {
				continue
			}
			c.Classes = append(c.Classes, DeserChainClass{Name: cc.Name, Found: cc.Found, ComponentInstanceID: cc.Owner.Instance, Confirmed: cc.Confirmed})
		}
		for _, cs := range p.Sites {
			c.Sites = append(c.Sites, DeserChainSite{Entry: entryID(cs.Site), State: cs.State, Reasons: cs.Own})
		}
		if len(p.Counts) > 0 {
			c.SitesByState = p.Counts
		}
		d.Chains = append(d.Chains, c)
		d.Summary.ChainsByState[p.State]++
	}
	return d
}

// chainRefs returns the chains that name the advisory (by ID or alias)
// and whose classes belong to instance.
func chainRefs(chains []deserialize.ChainPresence, cve cves.CVE, instance string) []ChainRef {
	if instance == "" {
		return nil
	}
	names := map[string]bool{cve.ID: true}
	for _, a := range cve.Aliases {
		names[a] = true
	}
	var out []ChainRef
	for _, p := range chains {
		if p.Instance != instance {
			continue
		}
		for _, a := range p.Spec.Advisories {
			if names[a] {
				out = append(out, ChainRef{Chain: p.Spec.ID, Copy: p.Copy, State: p.State})
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Chain != out[j].Chain {
			return out[i].Chain < out[j].Chain
		}
		return out[i].Copy < out[j].Copy
	})
	return out
}
