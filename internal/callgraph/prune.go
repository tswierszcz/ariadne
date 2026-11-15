package callgraph

// ModePTA is the -callgraph-mode value of GRAPH_PRECISION P3's receiver
// allocation narrowing: RTA, then internal/pta drops the CHA call sites
// whose receiver's allocated types prove another target.
const ModePTA = "pta"

// PruneCHASites drops, from every CHA-basis edge, the call sites drop
// rejects; an edge left with no site is removed. It returns how many
// edges and sites were removed. Every other edge is left as it is.
func (g *Graph) PruneCHASites(drop func(from uint32, e Edge, pc int) bool) (edgesRemoved, sitesRemoved int) {
	for id := range g.adjacency {
		adj := g.adjacency[id]
		kept := adj[:0]
		for _, e := range adj {
			if e.Basis != BasisCHA {
				kept = append(kept, e)
				continue
			}
			sites := make([]CallSite, 0, len(e.Sites))
			for _, s := range e.Sites {
				if drop(uint32(id), e, s.PC) {
					sitesRemoved++
					continue
				}
				sites = append(sites, s)
			}
			if len(sites) == 0 {
				edgesRemoved++
				continue
			}
			e.Sites = sites
			kept = append(kept, e)
		}
		for i := len(kept); i < len(adj); i++ {
			adj[i] = Edge{}
		}
		g.adjacency[id] = kept
	}
	g.chaEdges -= edgesRemoved
	return edgesRemoved, sitesRemoved
}

// PTA is the strategy recorded for a pta graph: built and narrowed as
// RTA, then narrowed again per call site by internal/pta.
type PTA struct{ RTA }

// Mode implements Strategy.
func (PTA) Mode() string { return ModePTA }
