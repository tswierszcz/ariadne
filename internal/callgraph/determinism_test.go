package callgraph

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
)

// TestBuild_deterministicAcrossRuns is a regression test for a real bug
// found running ariadne end-to-end: two separate `ariadne` processes
// analyzing the exact same inputs reported different REACHABLE paths for
// several CVEs whenever more than one path of the same (shortest) length
// existed, because several places iterated a Go map (hierarchy.Build's
// subtype inversion, both of callgraph.Build's own passes, and
// internal/reach's root collection) without sorting first. Go
// deliberately randomizes map iteration order per `range`, even within
// one process, so this is reproducible by calling hierarchy.Build and
// Build repeatedly in a single test run rather than needing to spawn
// separate processes.
func TestBuild_deterministicAcrossRuns(t *testing.T) {
	const runs = 20
	var renderings []string

	for i := 0; i < runs; i++ {
		h, _, err := hierarchy.Build("../../testdata/fixture/fixture-8.jar", nil)
		if err != nil {
			t.Fatalf("run %d: hierarchy.Build: %v", i, err)
		}
		g := Build(h, nil)
		renderings = append(renderings, renderGraph(g))
	}

	for i := 1; i < runs; i++ {
		if renderings[i] != renderings[0] {
			t.Fatalf("run %d's graph differs from run 0's — graph construction is not deterministic:\n--- run 0 ---\n%s\n--- run %d ---\n%s",
				i, renderings[0], i, renderings[i])
		}
	}
}

// renderGraph produces a stable, sorted-by-node-key text rendering of
// every node's outgoing edges (in the order Build actually stored them —
// edge order matters here too, since a BFS's tie-breaking depends on it),
// so two renderings can be compared with a plain string/byte comparison.
func renderGraph(g *Graph) string {
	type nodeEdges struct {
		key   string
		edges []string
	}
	nodes := make([]nodeEdges, g.NodeCount())
	for id := 0; id < g.NodeCount(); id++ {
		edges := g.Edges(uint32(id))
		edgeStrs := make([]string, len(edges))
		for i, e := range edges {
			edgeStrs[i] = fmt.Sprintf("%s(%s)", g.Key(e.To), e.Kind)
		}
		nodes[id] = nodeEdges{key: g.Key(uint32(id)), edges: edgeStrs}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].key < nodes[j].key })

	var sb strings.Builder
	for _, n := range nodes {
		sb.WriteString(n.key)
		sb.WriteString(" -> [")
		sb.WriteString(strings.Join(n.edges, ", "))
		sb.WriteString("]\n")
	}
	return sb.String()
}
