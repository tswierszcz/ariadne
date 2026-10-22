// Package methodhandle is DYNAMIC_LINKAGE L2's own consumer: it finds
// every `MethodHandle#invoke`/`invokeExact` call site, reads back the
// receiver DATA_FLOW D2 already resolved (valueflow.KindHandle, from a
// MethodHandles.Lookup find* call — see internal/dataflow's own
// methodhandle.go), and — like SERVICE_LOADING S1/S2 and
// ASYNC_CALLBACKS A1/A2 before it — turns a resolved identity into a
// real call-graph edge, never a fabricated one.
package methodhandle

import "ariadne/internal/hierarchy"

const methodHandleOwner = "java.lang.invoke.MethodHandle"

// InvokeSite is one call to MethodHandle#invoke/invokeExact found in the
// analyzed hierarchy. Matched by owner+name alone, not descriptor:
// invoke/invokeExact are polymorphic-signature methods, so javac emits
// a different descriptor at every call site depending on the actual
// argument/return types used there — there is no fixed overload to
// match against the way SERVICE_LOADING/ASYNC_CALLBACKS's catalogs do.
type InvokeSite struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Exact is true for invokeExact (strict MethodType match required
	// at the real JVM level) and false for invoke (adapting). Neither
	// distinction is enforced by this phase — recorded for the report,
	// not yet used to flag a type mismatch (DYNAMIC_LINKAGE L2's own
	// "niezgodny typ" bullet remains open, see this package's own
	// top-level doc note in RECOMMENDATION's own status record).
	Exact bool
}

// FindInvokeSites scans every method's already-extracted Calls for an
// invoke/invokeExact call on java.lang.invoke.MethodHandle.
func FindInvokeSites(h *hierarchy.Hierarchy) []InvokeSite {
	var sites []InvokeSite
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Owner != methodHandleOwner {
					continue
				}
				switch c.Name {
				case "invoke", "invokeExact":
					sites = append(sites, InvokeSite{
						CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor,
						PC: c.PC, Exact: c.Name == "invokeExact",
					})
				}
			}
		}
	}
	return sites
}
