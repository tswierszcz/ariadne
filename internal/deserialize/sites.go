package deserialize

import (
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/taint"
)

// Site is one call of a cataloged deserialization entry.
type Site struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Owner/Name/Desc are the call as named in the bytecode (Owner may
	// be a subclass of ObjectInputStream).
	Owner, Name, Desc string
	Model             string // Model.ID
	Format            string // Model.Format ("" when callbacks are not modeled)
	Coverage          string // Model.Coverage
	Filter            string
	FilterDetail      string
	IsApp             bool
	// Version is the deserializer library's release (XStream), when read.
	Version string
	// Candidates/Rejected: classes the format may create here, and those
	// the evaluated policy rejects (no edges for them); set by Connect.
	Candidates, Rejected int
	// RejectedSample lists up to 5 rejected classes, for the report.
	RejectedSample []string
	// DataFrom (DATA_FLOW D3): what the data this entry reads depends on
	// within its method (AttachArgumentFlow); DataFromRaw the same as
	// flowsum sources; DataFlowIncomplete why it may be incomplete;
	// DataFlowWidened how many calls of the method were answered by the
	// conservative policy.
	DataFrom           []string
	DataFromRaw        []string
	DataFlowIncomplete []string
	DataFlowWidened    int
	// Taint (DATA_FLOW D4, -taint): whether request data reaches the
	// data this entry reads; nil when not run.
	Taint *taint.Result
	allow policy
}

const (
	ois         = "java.io.ObjectInputStream"
	objectInput = "java.io.ObjectInput"
)

// isOISSubtype: fqcn is ObjectInputStream or a class of the input
// extending it.
func isOISSubtype(h *hierarchy.Hierarchy, fqcn string) bool {
	for i := 0; i < 64 && fqcn != ""; i++ {
		if fqcn == ois {
			return true
		}
		ci, ok := h.Lookup(fqcn)
		if !ok {
			return false
		}
		fqcn = ci.SuperClass
	}
	return false
}

// overridesResolveClass: a class of the input between fqcn and
// ObjectInputStream declares resolveClass(ObjectStreamClass).
func overridesResolveClass(h *hierarchy.Hierarchy, fqcn string) bool {
	for i := 0; i < 64 && fqcn != "" && fqcn != ois; i++ {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			return false
		}
		for _, m := range ci.Methods {
			if m.Name == "resolveClass" && m.Descriptor == "(Ljava/io/ObjectStreamClass;)Ljava/lang/Class;" {
				return true
			}
		}
		fqcn = ci.SuperClass
	}
	return false
}

// matchEntry maps a call to a catalog entry: exact for every model, and
// for Java serialization also a readObject/readUnshared invoked on a
// subclass of ObjectInputStream (the owner a compiler writes is the
// static receiver type).
func matchEntry(h *hierarchy.Hierarchy, c classfile.Call) (Model, bool) {
	if c.Kind == classfile.CallDynamic {
		return Model{}, false
	}
	if m, ok := lookupEntry(c.Owner, c.Name, c.Desc); ok {
		return m, true
	}
	if (c.Name == "readObject" || c.Name == "readUnshared") && c.Desc == "()Ljava/lang/Object;" && c.Owner != ois && isOISSubtype(h, c.Owner) {
		return lookupEntry(ois, c.Name, c.Desc)
	}
	return Model{}, false
}

// FindSites lists every call of a cataloged entry in the input, in a
// stable order. A call inside a class that implements the format itself
// (ObjectInputStream subclasses' own readObject overrides) still counts:
// it reads from a stream.
func FindSites(h *hierarchy.Hierarchy) []Site {
	var out []Site
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			var sites []Site
			for _, c := range m.Calls {
				model, ok := matchEntry(h, c)
				if !ok {
					continue
				}
				sites = append(sites, Site{CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC,
					Owner: c.Owner, Name: c.Name, Desc: c.Desc, Model: model.ID, Format: model.Format, Coverage: model.Coverage, Filter: FilterNoneSeen, IsApp: ci.IsApp})
			}
			if len(sites) == 0 {
				continue
			}
			var details []string
			for _, c := range m.Calls {
				if c.Name == "setObjectInputFilter" && c.Desc == "(Ljava/io/ObjectInputFilter;)V" && isOISSubtype(h, c.Owner) {
					details = append(details, "setObjectInputFilter in the same method")
					break
				}
			}
			for _, r := range m.InitRefs {
				if r.Kind == classfile.InitNew && r.Owner != ois && isOISSubtype(h, r.Owner) && overridesResolveClass(h, r.Owner) {
					details = append(details, "creates "+r.Owner+" (overrides resolveClass)")
				}
			}
			for i := range sites {
				s := &sites[i]
				d := append([]string(nil), details...)
				if s.Owner != ois && s.Owner != objectInput && overridesResolveClass(h, s.Owner) {
					d = append(d, "receiver "+s.Owner+" overrides resolveClass")
				}
				if len(d) > 0 && s.Format == FormatJava {
					sort.Strings(d)
					s.Filter, s.FilterDetail = FilterNotEvaluated, strings.Join(d, "; ")
				}
			}
			out = append(out, sites...)
		}
	}
	return out
}
