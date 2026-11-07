package deserialize

import (
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/flowsum"
)

// DATA_FLOW D3's first consumer: what the data an entry reads depends on
// within its own method — the stream object for Java serialization
// (readObject's receiver, with what was wrapped into it), the first
// argument for XStream's fromXML. Stated in the method's own inputs
// (parameters, the receiver, fields, results of unmodeled calls); where
// those inputs come from in turn is DATA_FLOW:D4's taint analysis, not
// this. Report-only: no edge, no verdict.

// SummaryRoots are the methods containing the sites, for flowsum.
func SummaryRoots(sites []Site) []flowsum.MethodKey {
	seen := map[flowsum.MethodKey]bool{}
	var out []flowsum.MethodKey
	for _, s := range sites {
		k := flowsum.MethodKey{Owner: s.CallerFQCN, Name: s.CallerName, Desc: s.CallerDesc}
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}

// AttachArgumentFlow sets each site's DataFrom/DataFlowIncomplete from the
// summaries of e (after Summarize on SummaryRoots).
func AttachArgumentFlow(sites []Site, e *flowsum.Engine) {
	for i := range sites {
		s := &sites[i]
		sum := e.Summary(flowsum.MethodKey{Owner: s.CallerFQCN, Name: s.CallerName, Desc: s.CallerDesc})
		if sum == nil {
			s.DataFlowIncomplete = []string{"method not summarized"}
			continue
		}
		s.DataFlowIncomplete = append([]string(nil), sum.Incomplete...)
		c, ok := sum.Calls[s.PC]
		if !ok {
			s.DataFlowIncomplete = append(s.DataFlowIncomplete, "entry call not in the summary")
			continue
		}
		var d flowsum.Dep
		switch {
		case s.Format == FormatXStream || s.Model == "xstream:fromxml":
			if len(c.Args) > 0 {
				d = c.Args[0]
			}
		default:
			d = c.Receiver
		}
		s.DataFrom = describeDep(d, s.CallerDesc)
		s.DataFromRaw = d.Strings()
		if len(sum.Widened) > 0 {
			s.DataFlowWidened = len(sum.Widened)
		}
	}
}

// typeName renders a field descriptor as a Java type name.
func typeName(d string) string {
	dims := 0
	for strings.HasPrefix(d, "[") {
		dims++
		d = d[1:]
	}
	prim := map[string]string{"B": "byte", "C": "char", "D": "double", "F": "float", "I": "int", "J": "long", "S": "short", "Z": "boolean"}
	n, ok := prim[d]
	if !ok {
		n = strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(d, "L"), ";"), "/", ".")
	}
	return n + strings.Repeat("[]", dims)
}

// describeDep renders a Dep for the report: "parameter 0 (java.io.InputStream)",
// "this", "field a.B.f", "result of a.B#m(…)".
func describeDep(d flowsum.Dep, desc string) []string {
	var types []string
	if args, err := classfile.ArgumentSlots(desc, true); err == nil {
		for _, a := range args {
			types = append(types, typeName(a.Type))
		}
	}
	var out []string
	for _, p := range d.Params() {
		label := "parameter " + strconv.Itoa(p)
		if p < len(types) {
			label += " (" + types[p] + ")"
		}
		out = append(out, label)
	}
	if d.This() {
		out = append(out, "this")
	}
	for _, f := range d.Fields() {
		out = append(out, "field "+f)
	}
	calls := d.Calls()
	sort.Strings(calls)
	for _, c := range calls {
		if c == "*" {
			out = append(out, "results of many unmodeled calls")
			continue
		}
		if i := strings.IndexByte(c, '('); i > 0 {
			c = c[:i]
		}
		out = append(out, "result of "+c)
	}
	return out
}
