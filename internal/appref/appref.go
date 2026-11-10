// Package appref holds what the configuration models (internal/log4jconfig,
// internal/logbackconfig) share: the jars of the analyzed class path, by
// the name a report gives them, and the symbols application code
// references — a model treats a class the application calls, creates or
// reads statically as in use, whatever the configuration says.
package appref

import (
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Jar is a jar of the analyzed class path, with the name to report it by.
type Jar struct {
	Path string
	Name string
}

// References maps "owner#member" to the first application class (in
// sorted order) referencing it.
type References map[string]string

// Collect gathers every symbol application classes call, create (`new`)
// or read statically.
func Collect(h *hierarchy.Hierarchy) References {
	out := References{}
	note := func(k, by string) {
		if prev, ok := out[k]; !ok || by < prev {
			out[k] = by
		}
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				note(c.Owner+"#"+c.Name, fqcn)
			}
			for _, ref := range m.InitRefs {
				switch ref.Kind {
				case classfile.InitNew:
					note(ref.Owner+"#<init>", fqcn)
				case classfile.InitGetStatic:
					note(ref.Owner+"#"+ref.Name, fqcn)
				}
			}
		}
	}
	return out
}

// Member reports the application class referencing exactly owner#member.
func (r References) Member(symbol string) (string, bool) {
	by, ok := r[symbol]
	return by, ok
}

// Owner reports the first application class referencing any member of
// class.
func (r References) Owner(class string) (string, bool) {
	best := ""
	for k, by := range r {
		if o, _, _ := strings.Cut(k, "#"); o == class && (best == "" || by < best) {
			best = by
		}
	}
	return best, best != ""
}
