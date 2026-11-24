package research

import (
	"fmt"
	"sort"
)

// index is the in-memory view query answers from.
type index struct {
	nodePos      map[string]int32 // node id -> position in Nodes
	edgePos      map[string]int
	out, in      [][]int // node position -> edge positions, sorted by edge id
	classByID    map[string]int
	compByID     map[string]int
	gapByID      map[string]int
	rootByNode   map[string]bool
	rootNodePos  []int32 // sorted by node id
	byMethodID   map[string][]int32
	specsOf      map[string][]int32 // canonical node id -> its specialization nodes
	classByName  map[string][]int   // binary class name -> class records
	hasDispatch  bool               // some edge is a client or supplied dispatch edge
	globalGapIDs []string
}

func (b *Bundle) buildIndex() error {
	x := &b.idx
	x.compByID = map[string]int{}
	for i, c := range b.Components {
		if !identifierRE.MatchString(c.InstanceID) || !hashRE.MatchString(c.ArtifactSHA256) {
			return corrupt("component %q has an invalid id or hash", c.InstanceID)
		}
		if _, dup := x.compByID[c.InstanceID]; dup {
			return corrupt("component %s repeated", c.InstanceID)
		}
		x.compByID[c.InstanceID] = i
	}
	x.classByID = map[string]int{}
	x.classByName = map[string][]int{}
	for i, c := range b.Classes {
		if _, dup := x.classByID[c.ClassID]; dup || !identifierRE.MatchString(c.ClassID) {
			return corrupt("class id %q invalid or repeated", c.ClassID)
		}
		if _, ok := x.compByID[c.InstanceID]; !ok {
			return corrupt("class %s names unknown instance %s", c.ClassID, c.InstanceID)
		}
		x.classByID[c.ClassID] = i
		x.classByName[c.Class] = append(x.classByName[c.Class], i)
	}
	for _, c := range b.Classes {
		sel, ok := x.classByID[c.SelectedClassID]
		if !ok || b.Classes[sel].Class != c.Class || !b.Classes[sel].Selected || c.Selected != (c.SelectedClassID == c.ClassID) {
			return corrupt("class %s has an invalid selected-copy reference %s", c.ClassID, c.SelectedClassID)
		}
	}
	x.nodePos = make(map[string]int32, len(b.Nodes))
	x.byMethodID = map[string][]int32{}
	x.specsOf = map[string][]int32{}
	for i, n := range b.Nodes {
		if _, dup := x.nodePos[n.NodeID]; dup || !identifierRE.MatchString(n.NodeID) {
			return corrupt("node id %q invalid or repeated", n.NodeID)
		}
		x.nodePos[n.NodeID] = int32(i)
	}
	for i, n := range b.Nodes {
		if _, ok := x.nodePos[n.CanonicalNodeID]; !ok {
			return corrupt("node %s names unknown canonical node %s", n.NodeID, n.CanonicalNodeID)
		}
		switch n.NodeKind {
		case "method":
			if n.CanonicalNodeID != n.NodeID || n.ReceiverType != nil || n.Method == nil {
				return corrupt("method node %s must be its own canonical node, with a method and no receiver type", n.NodeID)
			}
		case "receiver_specialization":
			if n.CanonicalNodeID == n.NodeID || n.ReceiverType == nil || n.Method == nil {
				return corrupt("specialization %s must name its canonical node and receiver type", n.NodeID)
			}
			x.specsOf[n.CanonicalNodeID] = append(x.specsOf[n.CanonicalNodeID], int32(i))
		case "synthetic_event":
			if n.Method != nil || n.Body != "synthetic" {
				return corrupt("synthetic event %s must have no method identity", n.NodeID)
			}
		default:
			return corrupt("node %s has unknown kind %q", n.NodeID, n.NodeKind)
		}
		if n.Method != nil {
			if !ValidMethodDescriptor(n.Method.Descriptor) {
				return corrupt("node %s has an invalid descriptor %q", n.NodeID, n.Method.Descriptor)
			}
			if n.Method.InstanceID != nil {
				if _, ok := x.compByID[*n.Method.InstanceID]; !ok {
					return corrupt("node %s names unknown instance %s", n.NodeID, *n.Method.InstanceID)
				}
			}
			if n.NodeKind == "method" {
				x.byMethodID[n.Method.MethodID] = append(x.byMethodID[n.Method.MethodID], int32(i))
			}
		}
	}
	x.out = make([][]int, len(b.Nodes))
	x.in = make([][]int, len(b.Nodes))
	x.edgePos = make(map[string]int, len(b.Edges))
	for i, e := range b.Edges {
		if _, dup := x.edgePos[e.EdgeID]; dup || !identifierRE.MatchString(e.EdgeID) {
			return corrupt("edge id %q invalid or repeated", e.EdgeID)
		}
		x.edgePos[e.EdgeID] = i
		f, ok1 := x.nodePos[e.From]
		t, ok2 := x.nodePos[e.To]
		if !ok1 || !ok2 {
			return corrupt("edge %s joins unknown nodes", e.EdgeID)
		}
		for _, s := range e.Sites {
			if s.CallerNodeID != e.From {
				return corrupt("edge %s has a site in another caller %s", e.EdgeID, s.CallerNodeID)
			}
		}
		x.out[f] = append(x.out[f], i)
		x.in[t] = append(x.in[t], i)
		if e.ClientDispatch || e.SuppliedDispatch {
			x.hasDispatch = true
		}
	}
	byID := func(l []int) {
		sort.Slice(l, func(i, j int) bool { return b.Edges[l[i]].EdgeID < b.Edges[l[j]].EdgeID })
	}
	for i := range x.out {
		byID(x.out[i])
		byID(x.in[i])
	}
	x.rootByNode = map[string]bool{}
	seenRoot := map[string]bool{}
	for _, r := range b.Roots {
		if seenRoot[r.RootID] || !identifierRE.MatchString(r.RootID) {
			return corrupt("root id %q invalid or repeated", r.RootID)
		}
		seenRoot[r.RootID] = true
		p, ok := x.nodePos[r.NodeID]
		if !ok {
			return corrupt("root %s names unknown node %s", r.RootID, r.NodeID)
		}
		if !x.rootByNode[r.NodeID] {
			x.rootByNode[r.NodeID] = true
			x.rootNodePos = append(x.rootNodePos, p)
		}
	}
	sort.Slice(x.rootNodePos, func(i, j int) bool { return b.Nodes[x.rootNodePos[i]].NodeID < b.Nodes[x.rootNodePos[j]].NodeID })
	x.gapByID = map[string]int{}
	for i, g := range b.Gaps {
		if _, dup := x.gapByID[g.GapID]; dup || !identifierRE.MatchString(g.GapID) {
			return corrupt("gap id %q invalid or repeated", g.GapID)
		}
		x.gapByID[g.GapID] = i
		if g.Site != nil {
			if _, ok := x.nodePos[g.Site.CallerNodeID]; !ok {
				return corrupt("gap %s names unknown caller %s", g.GapID, g.Site.CallerNodeID)
			}
		} else {
			x.globalGapIDs = append(x.globalGapIDs, g.GapID)
		}
		for _, id := range g.CandidateInstanceIDs {
			if _, ok := x.compByID[id]; !ok {
				return corrupt("gap %s names unknown instance %s", g.GapID, id)
			}
		}
	}
	seenCall := map[string]bool{}
	for _, c := range b.Calls {
		if seenCall[c.CallID] || !identifierRE.MatchString(c.CallID) {
			return corrupt("call id %q invalid or repeated", c.CallID)
		}
		seenCall[c.CallID] = true
		if _, ok := x.nodePos[c.Site.CallerNodeID]; !ok {
			return corrupt("call %s names unknown caller %s", c.CallID, c.Site.CallerNodeID)
		}
		for _, t := range c.TargetNodeIDs {
			if _, ok := x.nodePos[t]; !ok {
				return corrupt("call %s names unknown target %s", c.CallID, t)
			}
		}
	}
	return nil
}

// nodeInstance is the instance of a node's method, "" when unknown.
func (b *Bundle) nodeInstance(p int32) string {
	n := b.Nodes[p]
	if n.Method == nil || n.Method.InstanceID == nil {
		return ""
	}
	return *n.Method.InstanceID
}

func (b *Bundle) String() string {
	return fmt.Sprintf("bundle %s (%d nodes, %d edges, %d roots, %d gaps, %d calls)", b.AnalysisID, len(b.Nodes), len(b.Edges), len(b.Roots), len(b.Gaps), len(b.Calls))
}
