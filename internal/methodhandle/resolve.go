package methodhandle

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ResolvedInvoke is one InvokeSite with its receiver resolved, where
// possible, to the MethodHandles.Lookup find* call that produced it.
type ResolvedInvoke struct {
	Site InvokeSite
	// RequestedOwner/Method are HandleValue's own fields — the class
	// and method name literally passed to the find* call that produced
	// this handle, not yet resolved to the class that actually declares
	// it (FindHandleEdges does that, the same reasoning
	// ASYNC_CALLBACKS A2 already applies to a direct `new`).
	RequestedOwner, Method string
	// UnresolvedReason explains why RequestedOwner/Method are "" — the
	// receiver wasn't a provable valueflow.KindHandle (a parameter, a
	// field read, or a find* call this package's own dataflow
	// recognition couldn't resolve).
	UnresolvedReason string
}

// ResolveInvokes is DYNAMIC_LINKAGE L2's own receiver-resolution step:
// for every InvokeSite, read back its receiver via DATA_FLOW D2's
// Facts.ReceiverAt. Same re-read-the-class-file-for-its-pool approach
// SERVICE_LOADING S1/ASYNC_CALLBACKS A1 already use (classfile.ClassFile's
// constant pool is not retained past hierarchy.Build).
func ResolveInvokes(h *hierarchy.Hierarchy, sites []InvokeSite) ([]ResolvedInvoke, []string) {
	var out []ResolvedInvoke
	var warnings []string
	var jarReader classfile.JarClassReader
	defer jarReader.Close()

	byClass := map[string][]InvokeSite{}
	for _, s := range sites {
		byClass[s.CallerFQCN] = append(byClass[s.CallerFQCN], s)
	}

	for _, fqcn := range h.SortedFQCNs() {
		classSites, ok := byClass[fqcn]
		if !ok {
			continue
		}
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf(
				"re-reading %s (%s) for MethodHandle invoke-site analysis: %v", ci.SourceJar, ci.SourceEntry, err))
			continue
		}

		factsByMethod := map[[2]string]*dataflow.MethodFacts{}
		for _, s := range classSites {
			key := [2]string{s.CallerName, s.CallerDesc}
			facts, ok := factsByMethod[key]
			if !ok {
				m, found := ci.MethodByNameAndDesc(s.CallerName, s.CallerDesc)
				if !found {
					warnings = append(warnings, fmt.Sprintf(
						"%s#%s%s: method not found in the re-read class file", fqcn, s.CallerName, s.CallerDesc))
					continue
				}
				f, err := dataflow.Analyze(cf, fqcn, m.Method)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis: %v", fqcn, s.CallerName, s.CallerDesc, err))
					continue
				}
				factsByMethod[key] = f
				facts = f
			}
			out = append(out, resolveOne(fqcn, s, facts))
		}
	}
	return out, warnings
}

func resolveOne(fqcn string, s InvokeSite, facts *dataflow.MethodFacts) ResolvedInvoke {
	ri := ResolvedInvoke{Site: s}
	recvSite := bytecodeir.Site{Method: fqcn + "#" + s.CallerName, Descriptor: s.CallerDesc, PC: s.PC, Op: bytecodeir.OpCall}
	val := facts.ReceiverAt(recvSite)

	if aborted, reason := facts.Aborted(); aborted {
		ri.UnresolvedReason = "local-value analysis for this method aborted: " + reason
		return ri
	}
	switch val.Kind {
	case valueflow.KindHandle:
		ri.RequestedOwner, ri.Method = val.Handle.RequestedOwner, val.Handle.Name
	case valueflow.KindUnknown:
		ri.UnresolvedReason = "the receiver is unknown: " + val.Reason
	default:
		ri.UnresolvedReason = fmt.Sprintf("the receiver is %s, not a resolved MethodHandles.Lookup find* result", val.Kind)
	}
	return ri
}
