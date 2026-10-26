// Package analysis coordinates registered library models with the
// hierarchy-built call graph. It owns fixed-point iteration; models and
// profiles remain data/evaluation providers.
package analysis

import (
	"fmt"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/models"
)

// Options bounds work. MaxIterations must be positive; each iteration
// processes the currently reachable call sites, then propagates new edges.
type Options struct{ MaxIterations int }

// Result reports the mutated graph and the completeness of model analysis.
type Result struct {
	Graph      *callgraph.Graph
	Facts      []models.Fact
	Iterations int
	AddedEdges int
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
	result := Result{Graph: g}
	evaluated := make(map[string]bool)
	for iteration := 1; iteration <= opts.MaxIterations; iteration++ {
		result.Iterations = iteration
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
		if added == 0 {
			return result, nil
		}
	}
	result.Incomplete = true
	result.Reasons = appendUnique(result.Reasons, fmt.Sprintf("model analysis reached iteration limit (%d) before fixed point", opts.MaxIterations))
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

func appendUnique(in []string, value string) []string {
	for _, s := range in {
		if s == value {
			return in
		}
	}
	return append(in, value)
}
