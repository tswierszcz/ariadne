package reflectinvoke

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ProvenanceVersion versions the census (REFLECTION R3, slice 3).
const ProvenanceVersion = "r3-3"

// Provenance categories: where the receiver of an unresolved
// Method#invoke / Constructor#newInstance comes from, read in the calling
// method with parameter provenance (DATA_FLOW D2/D3).
const (
	ProvParameter    = "parameter"
	ProvField        = "field"
	ProvStaticResult = "static call result"
	ProvCallResult   = "instance call result"
	ProvTyped        = "typed value"
	ProvUnknown      = "unknown"
	ProvAborted      = "analysis aborted"
	ProvOther        = "other"
)

// ProvenanceSite is one unresolved reflective call and where its receiver
// comes from. Convertible is REFLECTION R3's yield: the receiver is a
// parameter and a caller in the input passes a Method/Constructor that
// R1/R2 resolve to a class in the input — the one shape interprocedural
// propagation would turn into an edge.
type ProvenanceSite struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	API                                string // "Method#invoke" or "Constructor#newInstance"
	Category                           string
	// Detail: what a field or static call result resolves to, the
	// parameter's callers' outcome, the instance call, or the unknown
	// reason — normalized so that sites group.
	Detail      string
	Convertible bool
	Target      string // the resolved member, when Convertible
}

// unresolvedCall is one site the census looks at.
type unresolvedCall struct {
	fqcn, name, desc string
	pc               int
	api              string
	partly           bool
}

// Provenance classifies every unresolved site of invokes and
// newInstances whose caller keep accepts (the census measures reachable
// callers only). idx resolves parameters through the input's callers.
func Provenance(h *hierarchy.Hierarchy, idx *dataflow.CallerIndex, invokes []ResolvedInvoke, newInstances []ResolvedNewInstance, keep func(fqcn, name, desc string) bool) ([]ProvenanceSite, []string) {
	var calls []unresolvedCall
	for _, r := range invokes {
		if r.UnresolvedReason != "" && keep(r.Site.CallerFQCN, r.Site.CallerName, r.Site.CallerDesc) {
			calls = append(calls, unresolvedCall{r.Site.CallerFQCN, r.Site.CallerName, r.Site.CallerDesc, r.Site.PC, "Method#invoke", r.PartlyThroughParameter})
		}
	}
	for _, r := range newInstances {
		if r.UnresolvedReason != "" && keep(r.Site.CallerFQCN, r.Site.CallerName, r.Site.CallerDesc) {
			calls = append(calls, unresolvedCall{r.Site.CallerFQCN, r.Site.CallerName, r.Site.CallerDesc, r.Site.PC, "Constructor#newInstance", false})
		}
	}
	sort.SliceStable(calls, func(i, j int) bool {
		a, b := calls[i], calls[j]
		if a.fqcn != b.fqcn {
			return a.fqcn < b.fqcn
		}
		if a.name+a.desc != b.name+b.desc {
			return a.name+a.desc < b.name+b.desc
		}
		return a.pc < b.pc
	})

	var out []ProvenanceSite
	var warnings []string
	var reader classfile.JarClassReader
	defer reader.Close()
	var cf *classfile.ClassFile
	cfOf := ""
	facts := map[string]*dataflow.MethodFacts{}
	for _, c := range calls {
		ps := ProvenanceSite{CallerFQCN: c.fqcn, CallerName: c.name, CallerDesc: c.desc, PC: c.pc, API: c.api}
		ci, ok := h.Lookup(c.fqcn)
		if !ok {
			continue
		}
		if cfOf != c.fqcn {
			var err error
			cf, err = reader.ReadClass(ci.SourceJar, ci.SourceEntry)
			cfOf = c.fqcn
			if err != nil {
				cf = nil
				warnings = append(warnings, fmt.Sprintf("re-reading %s for the reflection census: %v", c.fqcn, err))
			}
		}
		if cf == nil {
			continue
		}
		key := c.fqcn + "#" + c.name + c.desc
		f, ok := facts[key]
		if !ok {
			m, found := ci.MethodByNameAndDesc(c.name, c.desc)
			if !found {
				continue
			}
			var err error
			f, err = dataflow.AnalyzeWithParameterRefs(cf, c.fqcn, m.Method)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s: local-value analysis: %v", key, err))
				continue
			}
			facts[key] = f
		}
		if aborted, reason := f.Aborted(); aborted {
			ps.Category, ps.Detail = ProvAborted, firstClause(reason)
			out = append(out, ps)
			continue
		}
		site := bytecodeir.Site{Method: c.fqcn + "#" + c.name, Descriptor: c.desc, PC: c.pc, Op: bytecodeir.OpCall}
		classify(h, idx, &ps, f.ReceiverAt(site))
		if c.partly {
			// Its resolved values are edges already (R3); the rest is not
			// resolvable by this step.
			ps.Convertible, ps.Target = false, ""
			ps.Detail = "partly resolved through callers; other values unresolved"
		}
		out = append(out, ps)
	}
	return out, warnings
}

func classify(h *hierarchy.Hierarchy, idx *dataflow.CallerIndex, ps *ProvenanceSite, v valueflow.Value) {
	switch v.Kind {
	case valueflow.KindTyped:
		if v.Parameter == nil {
			ps.Category, ps.Detail = ProvTyped, v.Type
			return
		}
		ps.Category = ProvParameter
		pv := idx.ResolveParameter(dataflow.MethodID{Owner: ps.CallerFQCN, Name: ps.CallerName, Desc: ps.CallerDesc}, v.Parameter.Index)
		var outside []string
		for _, cv := range pv.Values {
			// One more step through a field or a static helper's result
			// the caller passes — the "parameter + field" shape.
			rv := oneHop(h, cv)
			if target, ok := reflectiveTarget(h, rv); ok {
				ps.Convertible, ps.Target = true, target
				ps.Detail = "a caller passes a resolved member"
				return
			}
			if o, ok := outsideTarget(h, rv); ok {
				outside = append(outside, o)
			}
		}
		switch {
		case len(outside) > 0:
			ps.Detail = "callers pass members outside the input"
		case len(pv.Values) > 0:
			ps.Detail = "callers pass " + kinds(pv.Values)
		case len(pv.Unknown) > 0:
			ps.Detail = "callers not determined: " + firstClause(pv.Unknown[0])
		default:
			ps.Detail = "no caller in the input"
		}
	case valueflow.KindField:
		ps.Category, ps.Detail = ProvField, resolvedDetail(h, v, "holds")
	case valueflow.KindCallResult:
		ps.Category, ps.Detail = ProvStaticResult, resolvedDetail(h, v, "returns")
	case valueflow.KindCallTrace:
		ps.Category = ProvCallResult
		if v.CallTrace != nil {
			ps.Detail = v.CallTrace.Owner + "#" + v.CallTrace.Name
		}
	case valueflow.KindUnknown:
		ps.Category, ps.Detail = ProvUnknown, firstClause(v.Reason)
	default:
		ps.Category, ps.Detail = ProvOther, string(v.Kind)
	}
}

// oneHop resolves a field read or a static call result one step;
// anything else is returned as is.
func oneHop(h *hierarchy.Hierarchy, v valueflow.Value) valueflow.Value {
	switch v.Kind {
	case valueflow.KindField:
		r, _ := dataflow.ResolveField(h, *v.Field)
		return r
	case valueflow.KindCallResult:
		r, _ := dataflow.ResolveCallResult(h, *v.Call)
		return r
	}
	return v
}

// resolvedDetail says what a field or static call result resolves to:
// a member outside the input (the JDK: Java 9+ compatibility shims), or
// the kind of value it is.
func resolvedDetail(h *hierarchy.Hierarchy, v valueflow.Value, verb string) string {
	r := oneHop(h, v)
	if _, ok := outsideTarget(h, r); ok {
		return verb + " a member of a class outside the input"
	}
	if r.Kind == valueflow.KindUnknown {
		return verb + " an unknown value: " + firstClause(r.Reason)
	}
	return verb + " a " + string(r.Kind) + " value"
}

// outsideTarget: a resolved Method/Constructor whose class is not in the
// input (resolving it would give no edge).
func outsideTarget(h *hierarchy.Hierarchy, v valueflow.Value) (string, bool) {
	switch v.Kind {
	case valueflow.KindReflectiveMethod:
		if _, ok := h.Lookup(v.ReflectiveMethod.OwnerFQCN); !ok {
			return v.ReflectiveMethod.OwnerFQCN, true
		}
	case valueflow.KindReflectiveConstructor:
		if _, ok := h.Lookup(v.ReflectiveConstructor.OwnerFQCN); !ok {
			return v.ReflectiveConstructor.OwnerFQCN, true
		}
	}
	return "", false
}

// reflectiveTarget: a resolved Method/Constructor on a class in the input.
func reflectiveTarget(h *hierarchy.Hierarchy, v valueflow.Value) (string, bool) {
	switch v.Kind {
	case valueflow.KindReflectiveMethod:
		if _, ok := h.Lookup(v.ReflectiveMethod.OwnerFQCN); ok {
			return v.ReflectiveMethod.OwnerFQCN + "#" + v.ReflectiveMethod.Name, true
		}
	case valueflow.KindReflectiveConstructor:
		if _, ok := h.Lookup(v.ReflectiveConstructor.OwnerFQCN); ok {
			return v.ReflectiveConstructor.OwnerFQCN + "#<init>", true
		}
	}
	return "", false
}

func kinds(vs []valueflow.Value) string {
	seen := map[string]bool{}
	var out []string
	for _, v := range vs {
		k := string(v.Kind)
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}

// firstClause keeps a reason's leading clause, so the census groups
// reasons that differ only in their specifics.
func firstClause(s string) string {
	for _, sep := range []string{": ", " (", "; "} {
		if i := strings.Index(s, sep); i > 0 {
			s = s[:i]
		}
	}
	return s
}

// ProvenanceSummary totals the census for the report and v2.
type ProvenanceSummary struct {
	Version string `json:"version"`
	// Sites are the reachable unresolved reflective calls measured.
	Sites      int            `json:"sites"`
	ByCategory map[string]int `json:"by_category"`
	// Convertible: sites R3's parameter step would still resolve (0 once
	// ResolveThroughParameters ran).
	Convertible int `json:"convertible"`
	// Details: the most frequent category/detail pairs.
	Details []ProvenanceDetail `json:"details"`
}

// ProvenanceDetail is one category/detail pair and its site count.
type ProvenanceDetail struct {
	Category string `json:"category"`
	Detail   string `json:"detail"`
	Sites    int    `json:"sites"`
}

// SummarizeProvenance totals sites; topN bounds Details. Nil when empty.
func SummarizeProvenance(sites []ProvenanceSite, topN int) *ProvenanceSummary {
	if len(sites) == 0 {
		return nil
	}
	s := &ProvenanceSummary{Version: ProvenanceVersion, Sites: len(sites), ByCategory: map[string]int{}}
	pairs := map[[2]string]int{}
	for _, p := range sites {
		s.ByCategory[p.Category]++
		if p.Convertible {
			s.Convertible++
		}
		pairs[[2]string{p.Category, p.Detail}]++
	}
	for k, n := range pairs {
		s.Details = append(s.Details, ProvenanceDetail{Category: k[0], Detail: k[1], Sites: n})
	}
	sort.Slice(s.Details, func(i, j int) bool {
		a, b := s.Details[i], s.Details[j]
		if a.Sites != b.Sites {
			return a.Sites > b.Sites
		}
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		return a.Detail < b.Detail
	})
	if len(s.Details) > topN {
		s.Details = s.Details[:topN]
	}
	return s
}
