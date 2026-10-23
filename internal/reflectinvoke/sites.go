// Package reflectinvoke is REFLECTION R2's own consumer: it finds every
// `java.lang.reflect.Method#invoke` call site, reads back the Method
// receiver DATA_FLOW D2/REFLECTION R1 already resolved
// (valueflow.KindReflectiveMethod, from a `Class#getMethod`/
// `getDeclaredMethod` call — see internal/dataflow's own reflectapi.go),
// and — like DYNAMIC_LINKAGE L2's internal/methodhandle before it —
// turns a resolved identity into a real call-graph edge, never a
// fabricated one. `Constructor#newInstance` is this phase's own next
// slice, not built here (REFLECTION R1 never resolved a
// getConstructor/getDeclaredConstructor receiver in the first place).
package reflectinvoke

import "ariadne/internal/hierarchy"

const methodOwner = "java.lang.reflect.Method"

// invokeDesc is Method#invoke's own single descriptor — not a
// polymorphic-signature method like MethodHandle#invoke, so every call
// site shares this exact descriptor.
const invokeDesc = "(Ljava/lang/Object;[Ljava/lang/Object;)Ljava/lang/Object;"

// InvokeSite is one call to Method#invoke found in the analyzed
// hierarchy, matched by owner+name+descriptor — unlike
// MethodHandle#invoke, this one is fixed, so there is no ambiguity to
// record the way InvokeSite.Exact does for DYNAMIC_LINKAGE L2.
type InvokeSite struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
}

// FindInvokeSites scans every method's already-extracted Calls for an
// invoke call on java.lang.reflect.Method.
func FindInvokeSites(h *hierarchy.Hierarchy) []InvokeSite {
	var sites []InvokeSite
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Owner != methodOwner || c.Name != "invoke" || c.Desc != invokeDesc {
					continue
				}
				sites = append(sites, InvokeSite{
					CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC,
				})
			}
		}
	}
	return sites
}
