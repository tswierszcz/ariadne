package reach

import "sort"

// UnboundedVersion versions the unbounded-site summary (ROADMAP B0 b0-5).
const UnboundedVersion = "b0-5"

// UnboundedSummary says where the reachable gap sites with no bounded
// candidate set are (ROADMAP B0 b0-5, report-only). Those sites are what
// makes the model-gap rule mark every NOT_REACHABLE finding (contract §4,
// b0-4); a site could be bounded only by a model of what its caller
// dispatches to, so the question is whether a few callers hold most of
// them. Measured on petclinic/WebGoat/JPetStore: no — hundreds of distinct
// callers across dozens of libraries.
type UnboundedSummary struct {
	Version string `json:"version"`
	// Sites are the reachable unbounded sites; Callers the distinct methods
	// holding them.
	Sites   int            `json:"sites"`
	Callers int            `json:"callers"`
	ByKind  map[string]int `json:"by_kind"`
	// TopCallerSites is how many sites the TopN callers with the most
	// sites hold together.
	TopN           int `json:"top_n"`
	TopCallerSites int `json:"top_caller_sites"`
	// ByComponent groups the sites by the component instance of the
	// calling class — its GAV, or its ID marked unidentified; NoInstance
	// when the class has none —, most sites first.
	ByComponent []UnboundedGroup `json:"by_component"`
	// TopCallers are the TopN callers with the most sites.
	TopCallers []UnboundedCaller `json:"top_callers"`
}

// NoInstance labels sites whose calling class belongs to no component
// instance (application classes, unidentified jars).
const NoInstance = "(no component instance)"

// UnboundedGroup is one component's share.
type UnboundedGroup struct {
	Component string `json:"component"`
	Sites     int    `json:"sites"`
	Callers   int    `json:"callers"`
}

// UnboundedCaller is one calling method and its sites.
type UnboundedCaller struct {
	Method string `json:"method"`
	Kind   string `json:"kind"`
	Sites  int    `json:"sites"`
}

// UnboundedSummary summarizes the reachable unbounded sites; nil when
// there are none. topN bounds TopCallers.
func (e *Engine) UnboundedSummary(topN int) *UnboundedSummary {
	sites := e.UnboundedGapSites()
	if len(sites) == 0 {
		return nil
	}
	s := &UnboundedSummary{Version: UnboundedVersion, Sites: len(sites), ByKind: map[string]int{}, TopN: topN}
	type callerKey struct{ method, kind string }
	perCaller := map[callerKey]int{}
	labels := map[string]string{}
	if e.components != nil {
		for _, in := range e.components.Instances {
			if l := in.Label(); l != "" {
				labels[in.ID] = l
			} else {
				labels[in.ID] = in.ID + " (unidentified)"
			}
		}
	}
	compSites := map[string]int{}
	compCallers := map[string]map[string]bool{}
	methods := map[string]bool{}
	for _, g := range sites {
		method := g.CallerFQCN + "#" + g.CallerName + g.CallerDesc
		methods[method] = true
		s.ByKind[g.Kind]++
		perCaller[callerKey{method, g.Kind}]++
		comp := NoInstance
		if e.components != nil {
			if ci, ok := e.h.Lookup(g.CallerFQCN); ok {
				if inst := e.components.InstanceForClass(ci); inst != "" {
					comp = labels[inst]
					if comp == "" {
						comp = inst
					}
				}
			}
		}
		compSites[comp]++
		if compCallers[comp] == nil {
			compCallers[comp] = map[string]bool{}
		}
		compCallers[comp][method] = true
	}
	s.Callers = len(methods)
	for c, n := range compSites {
		s.ByComponent = append(s.ByComponent, UnboundedGroup{Component: c, Sites: n, Callers: len(compCallers[c])})
	}
	sort.Slice(s.ByComponent, func(i, j int) bool {
		a, b := s.ByComponent[i], s.ByComponent[j]
		if a.Sites != b.Sites {
			return a.Sites > b.Sites
		}
		return a.Component < b.Component
	})
	for k, n := range perCaller {
		s.TopCallers = append(s.TopCallers, UnboundedCaller{Method: k.method, Kind: k.kind, Sites: n})
	}
	sort.Slice(s.TopCallers, func(i, j int) bool {
		a, b := s.TopCallers[i], s.TopCallers[j]
		if a.Sites != b.Sites {
			return a.Sites > b.Sites
		}
		if a.Method != b.Method {
			return a.Method < b.Method
		}
		return a.Kind < b.Kind
	})
	if len(s.TopCallers) > topN {
		s.TopCallers = s.TopCallers[:topN]
	}
	for _, c := range s.TopCallers {
		s.TopCallerSites += c.Sites
	}
	return s
}
