// Package nativebound inventories the boundaries between the analyzed Java
// code and native code (RUNTIME_NATIVE N5, first slice `n5-1`): the native
// methods the input declares, and the places that load a native library.
// Nothing here models native code. A native method has no bytecode, so a
// path ends at it; its native body may call back into any Java method
// through JNI, which is why the CLI reports a reachable native method as an
// unbounded model gap (kind "native") rather than a dead end.
package nativebound

import (
	"fmt"
	"sort"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// Version names the rules; a snapshot records it.
const Version = "n5-1"

// Method is one native method of the input.
type Method struct {
	FQCN string `json:"class"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	App  bool   `json:"app"`
}

// Load is one call that loads a native library.
type Load struct {
	CallerFQCN string `json:"caller_class"`
	CallerName string `json:"caller_name"`
	CallerDesc string `json:"caller_desc"`
	PC         int    `json:"pc"`
	// API is the loading method: System#loadLibrary, System#load,
	// Runtime#loadLibrary or Runtime#load.
	API string `json:"api"`
	// Library is the library name or path when it is a provable literal,
	// "" otherwise.
	Library string `json:"library,omitempty"`
}

var loadAPIs = map[string]map[string]bool{
	"java.lang.System":  {"loadLibrary": true, "load": true},
	"java.lang.Runtime": {"loadLibrary": true, "load": true},
}

// Methods lists every native method of the input's classes, sorted.
func Methods(h *hierarchy.Hierarchy) []Method {
	var out []Method
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			if m.IsNative() {
				out = append(out, Method{FQCN: fqcn, Name: m.Name, Desc: m.Descriptor, App: ci.IsApp})
			}
		}
	}
	return out
}

// Loads lists every call to a native-library loading API in the input,
// with the library name when the argument is a provable string literal.
// Warnings name classes that could not be re-read for the argument.
func Loads(h *hierarchy.Hierarchy) ([]Load, []string) {
	var out []Load
	var warnings []string
	var reader classfile.JarClassReader
	defer reader.Close()
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		var cf *classfile.ClassFile
		for _, m := range ci.Methods {
			var sites []classfile.Call
			for _, c := range m.Calls {
				if loadAPIs[c.Owner][c.Name] && c.Desc == "(Ljava/lang/String;)V" {
					sites = append(sites, c)
				}
			}
			if len(sites) == 0 {
				continue
			}
			var f *dataflow.MethodFacts
			if cf == nil {
				var err error
				if cf, err = reader.ReadClass(ci.SourceJar, ci.SourceEntry); err != nil {
					warnings = append(warnings, fmt.Sprintf("re-reading %s (%s) for native library loads: %v", ci.SourceJar, ci.SourceEntry, err))
				}
			}
			if cf != nil {
				if facts, err := dataflow.Analyze(cf, fqcn, m.Method); err == nil {
					f = facts
				}
			}
			for _, c := range sites {
				l := Load{CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC,
					API: shortOwner(c.Owner) + "#" + c.Name}
				if f != nil {
					site := bytecodeir.Site{Method: fqcn + "#" + m.Name, Descriptor: m.Descriptor, PC: c.PC, Op: bytecodeir.OpCall}
					l.Library = literal(f, site)
				}
				out = append(out, l)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.CallerFQCN != b.CallerFQCN {
			return a.CallerFQCN < b.CallerFQCN
		}
		if a.CallerName+a.CallerDesc != b.CallerName+b.CallerDesc {
			return a.CallerName+a.CallerDesc < b.CallerName+b.CallerDesc
		}
		return a.PC < b.PC
	})
	return out, warnings
}

func literal(f *dataflow.MethodFacts, site bytecodeir.Site) (s string) {
	defer func() {
		if recover() != nil {
			s = ""
		}
	}()
	v := f.ArgumentAt(site, 0)
	if v.Kind != valueflow.KindConst || v.Type != "java.lang.String" {
		return ""
	}
	s, _ = v.Const.(string)
	return s
}

func shortOwner(owner string) string {
	switch owner {
	case "java.lang.System":
		return "System"
	case "java.lang.Runtime":
		return "Runtime"
	}
	return owner
}

// LoadRow is a Load and whether the roots reach its caller.
type LoadRow struct {
	Load
	Reachable bool `json:"reachable"`
}

// Summary is the run's native boundaries (v2 coverage.native).
type Summary struct {
	Version string `json:"rules_version"`
	// Methods counts every native method of the input; Reachable lists the
	// ones the roots reach — each a model gap of kind "native".
	Methods   int       `json:"native_methods"`
	Reachable []Method  `json:"reachable_native_methods"`
	Loads     []LoadRow `json:"library_loads"`
}

// Summarize builds the summary with reachable (the engine's FRAMEWORK-pass
// reachability of a method).
func Summarize(ms []Method, ls []Load, reachable func(fqcn, name, desc string) bool) Summary {
	s := Summary{Version: Version, Methods: len(ms), Reachable: []Method{}, Loads: []LoadRow{}}
	for _, m := range ms {
		if reachable(m.FQCN, m.Name, m.Desc) {
			s.Reachable = append(s.Reachable, m)
		}
	}
	for _, l := range ls {
		s.Loads = append(s.Loads, LoadRow{Load: l, Reachable: reachable(l.CallerFQCN, l.CallerName, l.CallerDesc)})
	}
	return s
}

// ReachableLoads counts the loads whose caller is reachable.
func (s Summary) ReachableLoads() int {
	n := 0
	for _, l := range s.Loads {
		if l.Reachable {
			n++
		}
	}
	return n
}
