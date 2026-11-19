package libapi

import (
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// SuppliedSite is a call in the library's own code on an object the
// library's caller supplies (LIBRARY_MODE LB2 lb2-7,
// docs/RESULTS_CONTRACT.MD §8 "Caller-supplied objects"): a value of a
// dependency type that reaches the library as a parameter of one of its
// roots (an API method, an announced entry the container calls), directly
// or through a field written only from such parameters. Which
// implementation runs there is chosen by whoever supplies the object — the
// application, or the container it plugs the library into — so a dispatch
// from the site to every implementation in the dependencies (CHA) stands
// in for the caller's choice.
type SuppliedSite struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Called is the method the call names (owner#name+desc).
	Called string
	// Field: the receiver is read from a field set only by roots the
	// library never calls itself, from their parameters, so it is
	// caller-supplied on every path; otherwise it is the method's own
	// parameter, caller-supplied only on a path that enters the library at
	// this method (a root the library calls itself gets the library's
	// value there).
	Field bool
}

// SuppliedStats counts the analysis behind SuppliedSites.
type SuppliedStats struct {
	// Fields are the library fields of a dependency type written only
	// from the parameters of roots the library never calls itself.
	Fields int `json:"fields"`
	// Unanalyzed are library methods whose local values could not be had
	// (no site is reported in them).
	Unanalyzed int `json:"unanalyzed_methods"`
}

// SuppliedCoverage is v2 coverage.api.supplied.
type SuppliedCoverage struct {
	// Sites are the caller-supplied call sites, FieldSites of them on a
	// field's value; Edges the graph edges from them.
	Sites      int `json:"sites"`
	FieldSites int `json:"field_sites"`
	Edges      int `json:"edges"`
	SuppliedStats
}

// SuppliedSites lists, in deterministic order, the virtual and interface
// calls on a dependency type in the library's own methods whose receiver
// is a caller-supplied value. A value is caller-supplied when it is a
// parameter of the method itself and the method is a root (isRoot), or
// when it is read from a library instance field of a dependency type
// every write of which in the library stores a parameter of a root that no
// library method calls (so only a caller can have set it). Dependency
// types are classes and interfaces outside the library and the JDK that a
// caller can subtype (an interface, or a class that is not final). Not
// followed: a supplied value handed to a dependency method, or stored
// inside a dependency object, and called there.
func SuppliedSites(h *hierarchy.Hierarchy, isRoot func(owner, name, desc string) bool) ([]SuppliedSite, SuppliedStats) {
	var st SuppliedStats
	reader := &classfile.JarClassReader{}
	defer reader.Close()

	dep := map[string]bool{}
	isDep := func(fqcn string) bool {
		if v, ok := dep[fqcn]; ok {
			return v
		}
		v := false
		if !jdkClass(fqcn) {
			if ci, ok := h.Lookup(fqcn); ok && !ci.IsApp {
				v = ci.IsInterface() || ci.AccessFlags&classfile.AccFinal == 0
			}
		}
		dep[fqcn] = v
		return v
	}
	depType := func(desc string) bool {
		if !strings.HasPrefix(desc, "L") || !strings.HasSuffix(desc, ";") {
			return false
		}
		return isDep(strings.ReplaceAll(desc[1:len(desc)-1], "/", "."))
	}

	type method struct {
		fqcn string
		m    hierarchy.Method
		f    *dataflow.MethodFacts
		root bool
	}
	var methods []method
	// calledInLibrary: methods some library method calls (as resolved to
	// their declaring class), constructors included (new, super(…)).
	calledInLibrary := map[string]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp || ci.SourceJar == "" {
			continue
		}
		cf, err := reader.ReadClass(ci.SourceJar, ci.SourceEntry)
		if err != nil {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				if c.Kind == classfile.CallDynamic {
					continue
				}
				owner := c.Owner
				if d, ok := h.ResolveExact(c.Owner, c.Name, c.Desc); ok {
					owner = d
				}
				calledInLibrary[owner+"#"+c.Name+c.Desc] = true
			}
			if len(m.Code) == 0 {
				continue
			}
			f, err := dataflow.AnalyzeWithParameterRefs(cf, fqcn, m.Method)
			if err != nil {
				st.Unanalyzed++
				continue
			}
			if aborted, _ := f.Aborted(); aborted {
				st.Unanalyzed++
				continue
			}
			methods = append(methods, method{fqcn: fqcn, m: m, f: f, root: isRoot(fqcn, m.Name, m.Descriptor)})
		}
	}

	// param: v is a parameter of a dependency type of mt, a root.
	param := func(mt method, v valueflow.Value) bool {
		if !mt.root || v.Kind != valueflow.KindTyped || v.Parameter == nil || v.Parameter.Index < 0 {
			return false
		}
		args, err := classfile.ArgumentSlots(mt.m.Descriptor, mt.m.IsStatic())
		if err != nil || v.Parameter.Index >= len(args) {
			return false
		}
		return depType(args[v.Parameter.Index].Type)
	}

	// Fields of library classes, of a dependency type, every write of
	// which stores a parameter of a root no library method calls.
	type fieldKey struct{ owner, name string }
	fieldOK := map[fieldKey]bool{}
	for _, mt := range methods {
		external := !calledInLibrary[mt.fqcn+"#"+mt.m.Name+mt.m.Descriptor]
		for _, w := range mt.f.FieldWrites() {
			if w.Static {
				continue
			}
			k := fieldKey{w.Field.Owner, w.Field.Name}
			if ok, seen := fieldOK[k]; seen && !ok {
				continue
			}
			oci, found := h.Lookup(w.Field.Owner)
			fieldOK[k] = found && oci.IsApp && depType(w.Field.Desc) && external && param(mt, w.Value)
		}
	}
	for _, ok := range fieldOK {
		if ok {
			st.Fields++
		}
	}

	var out []SuppliedSite
	for _, mt := range methods {
		for _, c := range mt.m.Calls {
			if c.Kind != classfile.CallVirtual && c.Kind != classfile.CallInterface || !isDep(c.Owner) {
				continue
			}
			site := bytecodeir.Site{Method: mt.fqcn + "#" + mt.m.Name, Descriptor: mt.m.Descriptor, PC: c.PC, Op: bytecodeir.OpCall}
			v, ok := receiverAt(mt.f, site)
			if !ok {
				continue
			}
			s := SuppliedSite{CallerFQCN: mt.fqcn, CallerName: mt.m.Name, CallerDesc: mt.m.Descriptor, PC: c.PC, Called: c.Owner + "#" + c.Name + c.Desc}
			switch {
			case param(mt, v):
			case v.Kind == valueflow.KindField && v.Field != nil && fieldOK[fieldKey{v.Field.Owner, v.Field.Name}]:
				s.Field = true
			default:
				continue
			}
			out = append(out, s)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.CallerFQCN != b.CallerFQCN {
			return a.CallerFQCN < b.CallerFQCN
		}
		if a.CallerName != b.CallerName {
			return a.CallerName < b.CallerName
		}
		if a.CallerDesc != b.CallerDesc {
			return a.CallerDesc < b.CallerDesc
		}
		return a.PC < b.PC
	})
	return out, st
}

func jdkClass(fqcn string) bool {
	for _, p := range []string{"java.", "javax.", "jdk.", "sun."} {
		if strings.HasPrefix(fqcn, p) {
			return true
		}
	}
	return false
}

func receiverAt(f *dataflow.MethodFacts, at bytecodeir.Site) (v valueflow.Value, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return f.ReceiverAt(at), true
}
