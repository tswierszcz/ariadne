// newinstance_sites.go is REFLECTION R2's own second half: finding
// every `java.lang.reflect.Constructor#newInstance` call site. The
// mirror of sites.go's own FindInvokeSites, but for allocation rather
// than invocation — see newinstance_resolve.go/newinstance_edges.go
// for the rest of this half.
package reflectinvoke

import "ariadne/internal/hierarchy"

const constructorOwner = "java.lang.reflect.Constructor"

// newInstanceDesc is Constructor#newInstance's own single descriptor —
// like Method#invoke, not a polymorphic-signature method, so every
// call site shares this exact descriptor.
const newInstanceDesc = "([Ljava/lang/Object;)Ljava/lang/Object;"

// NewInstanceSite is one call to Constructor#newInstance found in the
// analyzed hierarchy, matched by owner+name+descriptor.
type NewInstanceSite struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
}

// FindNewInstanceSites scans every method's already-extracted Calls
// for a newInstance call on java.lang.reflect.Constructor.
func FindNewInstanceSites(h *hierarchy.Hierarchy) []NewInstanceSite {
	var sites []NewInstanceSite
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Owner != constructorOwner || c.Name != "newInstance" || c.Desc != newInstanceDesc {
					continue
				}
				sites = append(sites, NewInstanceSite{
					CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC,
				})
			}
		}
	}
	return sites
}
