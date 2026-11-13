// Package analysis coordinates registered library models with the
// hierarchy-built call graph. It owns fixed-point iteration; models and
// profiles remain data/evaluation providers.
//
// Two kinds of participant share one fixed point (ROADMAP B3): registered
// models, evaluated at reachable call sites matched by exact signature,
// and steps (Options.Steps) — passes with their own rule that add edges
// given the graph and the roots, such as the creator gates
// (callgraph.NewCreatorGate: SAM calls, servlet and JDBC objects,
// framework proxies). Each iteration runs every step, then one model pass;
// the analysis stops when an iteration adds no edge. Every participant
// only adds edges and decides from reachability, which only grows, so the
// fixed point does not depend on the order of the participants.
package analysis

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/models"
)

// CoordinatorVersion versions the fixed-point rules: b3-1 iterated the
// registered models alone; b3-2 iterates models and steps (the creator
// gates) together; b3-3 runs only the participants the profile selects
// (Select). Snapshots record it.
const CoordinatorVersion = "b3-3"

// Participant kinds.
const (
	KindModel = "model" // a registered model (models.Registry)
	KindGate  = "gate"  // a step: a creator gate
)

// Participant is one entry of the analyzer's catalog of fixed-point
// participants: a stable ID a profile can name, the version of its
// rules, and its kind.
type Participant struct {
	ID      string
	Version string
	Kind    string
}

// Label is "id@version".
func (p Participant) Label() string { return p.ID + "@" + p.Version }

// Select splits the catalog into the participants that run and the ones
// disable names (ROADMAP B3 b3-3). An ID the catalog does not know is an
// error, never ignored: a typo in a profile would otherwise silently
// leave a model on. Both lists keep the catalog's order.
func Select(catalog []Participant, disable []string) (active, disabled []Participant, err error) {
	known := map[string]bool{}
	for _, p := range catalog {
		known[p.ID] = true
	}
	off := map[string]bool{}
	var unknown []string
	for _, id := range disable {
		if !known[id] {
			unknown = append(unknown, fmt.Sprintf("%q", id))
			continue
		}
		off[id] = true
	}
	if len(unknown) > 0 {
		var ids []string
		for _, p := range catalog {
			ids = append(ids, p.ID)
		}
		return nil, nil, fmt.Errorf("unknown model or gate %s in the profile's models.disable (known: %s)", strings.Join(unknown, ", "), strings.Join(ids, ", "))
	}
	for _, p := range catalog {
		if off[p.ID] {
			disabled = append(disabled, p)
		} else {
			active = append(active, p)
		}
	}
	return active, disabled, nil
}

// Labels returns the participants' labels, sorted.
func Labels(ps []Participant) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Label())
	}
	sort.Strings(out)
	return out
}

// Options bounds work. MaxIterations must be positive; each iteration
// runs every step, then processes the currently reachable call sites.
type Options struct {
	MaxIterations int
	Steps         []Step
}

// Step is a fixed-point participant other than a registered model. Apply
// adds the edges its rule admits for the graph and roots as they are now
// and returns how many it added in this call (0 when it has nothing new).
type Step struct {
	Name  string
	Apply func(g *callgraph.Graph, roots []uint32) int
}

// Result reports the mutated graph and the completeness of model analysis.
type Result struct {
	Graph      *callgraph.Graph
	Facts      []models.Fact
	Iterations int
	// AddedEdges counts the model edges; StepEdges the edges each step
	// added, by name.
	AddedEdges int
	StepEdges  map[string]int
	// LastAdders names the participants (step names, "models") that still
	// added edges in the final iteration — empty at a fixed point reached
	// before the limit, the culprits when the limit stopped the analysis.
	LastAdders []string
	Incomplete bool
	Reasons    []string
}

// Run evaluates matching models only at reachable bytecode call sites.
// Newly added edges can expose further modeled sites on the next pass.
// All model targets must be methods already present in the analyzed
// hierarchy; this first coordinator slice does not synthesize classfiles.
func Run(h *hierarchy.Hierarchy, g *callgraph.Graph, registry *models.Registry, evaluator models.Evaluator, roots []uint32, opts Options) (Result, error) {
	if h == nil || g == nil || registry == nil || evaluator == nil {
		return Result{}, fmt.Errorf("analysis requires hierarchy, graph, registry, and evaluator")
	}
	if opts.MaxIterations <= 0 {
		return Result{}, fmt.Errorf("MaxIterations must be positive")
	}
	methods := make(map[uint32]hierarchy.Method)
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, method := range ci.Methods {
			if id, ok := g.ID(fqcn, method.Name, method.Descriptor); ok {
				methods[id] = method
			}
		}
	}
	result := Result{Graph: g, StepEdges: map[string]int{}}
	evaluated := make(map[string]bool)
	for iteration := 1; iteration <= opts.MaxIterations; iteration++ {
		result.Iterations = iteration
		var adders []string
		stepAdded := 0
		for _, st := range opts.Steps {
			if n := st.Apply(g, roots); n > 0 {
				result.StepEdges[st.Name] += n
				stepAdded += n
				adders = append(adders, st.Name)
			}
		}
		reachable := reachableNodes(g, roots)
		var edges []callgraph.ModelEdge
		for _, id := range reachable {
			method, ok := methods[id]
			if !ok {
				continue
			}
			callerFQCN, callerName, callerDesc, ok := splitMethodKey(g.Key(id))
			if !ok {
				continue
			}
			for _, call := range method.Calls {
				callee := models.Signature{Owner: call.Owner, Name: call.Name, Desc: call.Desc}
				ci, _ := h.Lookup(call.Owner)
				var depGAV *input.GAV
				if ci != nil {
					depGAV = ci.GAV
				}
				matched := registry.Lookup(callee, depGAV)
				for _, model := range matched {
					site := models.CallSite{Caller: models.Signature{Owner: callerFQCN, Name: callerName, Desc: callerDesc}, Call: callee, PC: call.PC, GAV: depGAV}
					key := fmt.Sprintf("%s|%s|%s|%s|%d", model.ID, callerFQCN, callerName+callerDesc, call.Owner+"#"+call.Name+call.Desc, call.PC)
					if evaluated[key] {
						continue
					}
					evaluated[key] = true
					fact := evaluator.Evaluate(model, site)
					fact.ModelID, fact.Caller, fact.PC = model.ID, site.Caller, call.PC
					if fact.Incomplete {
						result.Incomplete = true
						if fact.Reason == "" {
							fact.Reason = "model evaluation incomplete without a reason"
						}
						result.Reasons = appendUnique(result.Reasons, model.ID+": "+fact.Reason)
					}
					for _, target := range fact.Targets {
						if _, ok := g.ID(target.Method.Owner, target.Method.Name, target.Method.Desc); !ok {
							fact.Incomplete, result.Incomplete = true, true
							reason := "modeled target is outside the analyzed hierarchy: " + target.Method.Owner + "#" + target.Method.Name + target.Method.Desc
							if fact.Reason == "" {
								fact.Reason = reason
							}
							result.Reasons = appendUnique(result.Reasons, model.ID+": "+reason)
							continue
						}
						edges = append(edges, callgraph.ModelEdge{CallerFQCN: callerFQCN, CallerName: callerName, CallerDesc: callerDesc, TargetFQCN: target.Method.Owner, TargetName: target.Method.Name, TargetDesc: target.Method.Desc, Kind: target.Kind, PC: call.PC})
					}
					result.Facts = append(result.Facts, fact)
				}
			}
		}
		added := callgraph.ApplyModelEdges(g, edges)
		result.AddedEdges += added
		if added > 0 {
			adders = append(adders, "models")
		}
		result.LastAdders = adders
		if added == 0 && stepAdded == 0 {
			return result, nil
		}
	}
	result.Incomplete = true
	result.Reasons = appendUnique(result.Reasons, fmt.Sprintf("model analysis reached iteration limit (%d) before fixed point (still adding: %s)", opts.MaxIterations, joinOr(result.LastAdders)))
	return result, nil
}

// reachableNodes returns a deterministic breadth-first traversal.
func reachableNodes(g *callgraph.Graph, roots []uint32) []uint32 {
	seen := make(map[uint32]bool)
	queue := append([]uint32(nil), roots...)
	out := make([]uint32, 0, len(queue))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
		for _, e := range g.Edges(id) {
			if !seen[e.To] {
				queue = append(queue, e.To)
			}
		}
	}
	return out
}

func splitMethodKey(key string) (string, string, string, bool) {
	hash := -1
	for i := 0; i < len(key); i++ {
		if key[i] == '#' {
			hash = i
			break
		}
	}
	if hash < 0 {
		return "", "", "", false
	}
	open := -1
	for i := hash + 1; i < len(key); i++ {
		if key[i] == '(' {
			open = i
			break
		}
	}
	if open < 0 {
		return "", "", "", false
	}
	return key[:hash], key[hash+1 : open], key[open:], true
}

func joinOr(names []string) string {
	if len(names) == 0 {
		return "nothing"
	}
	out := names[0]
	for _, n := range names[1:] {
		out += ", " + n
	}
	return out
}

func appendUnique(in []string, value string) []string {
	for _, s := range in {
		if s == value {
			return in
		}
	}
	return append(in, value)
}
