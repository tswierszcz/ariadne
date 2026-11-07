package main

import (
	"strings"

	"ariadne/internal/component"
	"ariadne/internal/deserialize"
	"ariadne/internal/hierarchy"
)

// chainInputs adapts the component inventory (CVE:F1a/F6a/F7a) to
// DESERIALIZATION Z3a's chain evaluation: who owns a class, and which
// instances sit under relocations.
func chainInputs(h *hierarchy.Hierarchy, inv *component.Inventory, targetJava int) deserialize.ChainInputs {
	in := deserialize.ChainInputs{H: h, TargetJava: targetJava}
	if inv == nil {
		return in
	}
	byID := map[string]deserialize.ClassOwner{}
	for _, inst := range inv.Instances {
		o := deserialize.ClassOwner{Instance: inst.ID, Relocations: inst.Relocations}
		if inst.GAV != nil {
			o.Component = inst.GAV.PackageName()
			o.Version = inst.GAV.Version
			if len(inst.VersionCandidates) > 0 {
				o.Version = strings.Join(inst.VersionCandidates, ",")
			}
		}
		byID[inst.ID] = o
		if len(inst.Relocations) > 0 {
			in.Relocated = append(in.Relocated, o)
		}
	}
	in.Owner = func(ci *hierarchy.ClassInfo) deserialize.ClassOwner {
		id := inv.InstanceForClass(ci)
		if o, ok := byID[id]; ok {
			return o
		}
		return deserialize.ClassOwner{Instance: id}
	}
	return in
}
