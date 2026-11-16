package callgraph

// CutSitesAfter records that, in this analysis, method id never runs
// past bytecode offset after: its call sites at a greater pc are dropped
// now, from edges of every basis, and every edge a model adds later at
// such a site is refused (siteCut). An edge left with no site is
// removed; an edge that carries no site at all (it names no call) is
// kept. It returns how many sites and edges were removed.
//
// The JDBC driver model (internal/jdbcurl) uses it: a registered driver
// whose URL prefix matches no URL the application can use runs only its
// URL check when DriverManager offers it a connection, and returns null.
func (g *Graph) CutSitesAfter(id uint32, after int) (sitesRemoved, edgesRemoved int) {
	if g.siteCuts == nil {
		g.siteCuts = map[uint32]int{}
	}
	if old, ok := g.siteCuts[id]; ok && old <= after {
		return 0, 0
	}
	g.siteCuts[id] = after
	adj := g.adjacency[id]
	kept := adj[:0]
	for _, e := range adj {
		if len(e.Sites) == 0 {
			kept = append(kept, e)
			continue
		}
		sites := make([]CallSite, 0, len(e.Sites))
		for _, s := range e.Sites {
			if s.PC > after {
				sitesRemoved++
				continue
			}
			sites = append(sites, s)
		}
		if len(sites) == 0 {
			edgesRemoved++
			g.uncount(e.Basis)
			continue
		}
		e.Sites = sites
		kept = append(kept, e)
	}
	for i := len(kept); i < len(adj); i++ {
		adj[i] = Edge{}
	}
	g.adjacency[id] = kept
	return sitesRemoved, edgesRemoved
}

// SiteCut reports the offset CutSitesAfter recorded for id, if any.
func (g *Graph) SiteCut(id uint32) (int, bool) {
	after, ok := g.siteCuts[id]
	return after, ok
}

// siteCut reports whether a call at pc in from was cut by CutSitesAfter.
func (g *Graph) siteCut(from uint32, pc int) bool {
	after, ok := g.siteCuts[from]
	return ok && pc > after
}

// uncount keeps the report's running totals in step with a removed edge.
func (g *Graph) uncount(b EdgeBasis) {
	switch b {
	case BasisCHA:
		g.chaEdges--
	case BasisInit:
		g.initEdges--
	case BasisModel:
		g.modelEdges--
	case BasisSAM:
		g.samEdges--
	case BasisDirect, BasisSynthetic:
		g.preCHAEdges--
	}
}
