package samcall

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

const fixtureJar = "../../testdata/samcall/sc.jar"

func fixtureGraph(t *testing.T) (*callgraph.Graph, []Result, callgraph.SAMStats) {
	t.Helper()
	h, _, err := hierarchy.Build(fixtureJar, nil)
	if err != nil {
		t.Fatal(err)
	}
	g := callgraph.Build(h, nil)
	r := NewResolver(h)
	defer r.Close()
	results := r.Resolve(Find(h))
	var edges []callgraph.SAMEdge
	for _, e := range Edges(results) {
		var cs []callgraph.MethodRef
		for _, c := range e.Creators {
			cs = append(cs, callgraph.MethodRef{FQCN: c.Owner, Name: c.Name, Desc: c.Desc})
		}
		edges = append(edges, callgraph.SAMEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC,
			Kind: e.CallKind, TargetFQCN: e.TargetFQCN, TargetName: e.TargetName, TargetDesc: e.TargetDesc, Virtual: e.Virtual, Creators: cs})
	}
	main, ok := g.ID("sc.Main", "main", "([Ljava/lang/String;)V")
	if !ok {
		t.Fatal("no sc.Main#main")
	}
	st := callgraph.NewSAMGate(g, h, edges).Apply(g, []uint32{main})
	return g, results, st
}

func node(t *testing.T, g *callgraph.Graph, key string) uint32 {
	t.Helper()
	owner, rest, _ := strings.Cut(key, "#")
	i := strings.IndexByte(rest, '(')
	id, ok := g.ID(owner, rest[:i], rest[i:])
	if !ok {
		t.Fatalf("no node %s", key)
	}
	return id
}

// TestSAMFixture (DYNAMIC_LINKAGE L1): every caller -> body relation the
// real JVM ran in testdata/samcall (scripts/samcall-oracle.sh) is a SAM-call
// edge at the measured bytecode index — a lambda or method reference
// created in one method and run through its functional interface in
// another: passed as a parameter (also to a helper called through its
// interface), returned, stored by a setter, held in a final field (a bound
// reference to an overridden method: the override is the target),
// a constructor reference, handed to List#forEach, captured by another
// lambda. A lambda nobody runs and one whose creator is never called get
// no SAM edge.
func TestSAMFixture(t *testing.T) {
	g, results, st := fixtureGraph(t)
	f, err := os.Open("../../testdata/samcall/observed.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		p := strings.Split(line, "\t")
		bci, _ := strconv.Atoi(p[1])
		from, to := node(t, g, p[0]), node(t, g, p[2])
		rows++
		found := false
		for _, e := range g.Edges(from) {
			if e.To != to || e.Basis != callgraph.BasisSAM {
				continue
			}
			for _, s := range e.Sites {
				found = found || s.PC == bci
			}
		}
		if !found {
			t.Errorf("no SAM edge %s -> %s at pc %d", p[0], p[2], bci)
		}
	}
	if rows != 8 {
		t.Errorf("%d observed relations, want 8", rows)
	}

	for _, body := range []string{"sc.Main#lambda$unused$0(I)I", "sc.Main#lambda$neverCalled$0(I)I"} {
		to := node(t, g, body)
		for id := 0; id < g.NodeCount(); id++ {
			for _, e := range g.Edges(uint32(id)) {
				if e.To == to && e.Basis == callgraph.BasisSAM {
					t.Errorf("SAM edge %s -> %s: no execution runs it", g.Key(uint32(id)), body)
				}
			}
		}
	}
	if st.Gated == 0 {
		t.Errorf("stats %+v: the never-called creator's lambda should be a gated candidate", st)
	}
	for _, r := range results {
		if r.Site.CallerName == "run" && r.Site.Kind == KindCall && len(r.Targets) != 2 {
			t.Errorf("run(Op): targets %v, want the two lambdas passed to it (one gated)", r.Targets)
		}
	}
}

// TestSAMGapsDisclosed: a SAM call whose receiver the analysis cannot
// follow gets no target and names why — never every lambda of the
// interface.
func TestSAMGapsDisclosed(t *testing.T) {
	_, results, _ := fixtureGraph(t)
	for _, r := range results {
		if r.Site.Kind != KindCall {
			continue
		}
		if r.Site.CallerFQCN == "sc.Oracle" && len(r.Targets) != 0 {
			t.Errorf("%s#%s: targets %v", r.Site.CallerFQCN, r.Site.CallerName, r.Targets)
		}
		if len(r.Targets) == 0 && len(r.Gaps) == 0 && r.Objects == 0 {
			t.Errorf("%s#%s@%d: no target and no reason", r.Site.CallerFQCN, r.Site.CallerName, r.Site.PC)
		}
	}
}
