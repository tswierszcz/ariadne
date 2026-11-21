package reflectinvoke

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// ForNameInit is one Class.forName call site that initializes a class of
// the input (REFLECTION R4 r4-1): the class named by a provable string
// literal, loaded with initialization — forName(String) always
// initializes, forName(String, boolean, ClassLoader) when its boolean is
// true or not a provable constant. InitFQCN is the class whose <clinit>
// runs first: the named class when it declares one, otherwise its nearest
// superclass that does (JVMS 5.5 initializes the superclass before the
// class; a class's own <clinit> already has its JVM_GRAPH G3 edge to its
// superclass's). Interfaces initialize no superinterface here.
type ForNameInit struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	// Named is the class the literal names, InitFQCN the class whose
	// <clinit> the edge goes to.
	Named, InitFQCN string
	// MaybeInit: the initialize argument is not a provable constant — the
	// class may or may not be initialized here.
	MaybeInit bool
}

// ForNameStats counts what FindForNameInits saw.
type ForNameStats struct {
	// Sites are the Class.forName calls (both overloads), Literal those
	// whose name is a provable string literal.
	Sites, Literal int
	// NoInit: literal sites with a constant initialize=false; Outside:
	// literal sites naming a class not in the input (the JDK, or absent);
	// NoInitializer: named classes with no <clinit> in themselves or a
	// superclass; Inits: the sites returned, MaybeInit of them with an
	// initialize argument that is not a constant.
	NoInit, Outside, NoInitializer, Inits, MaybeInit int
}

// forNameIndex: overload descriptor → index of the initialize argument
// (-1: none, the class is always initialized). The name is argument 0.
var forNameIndex = map[string]int{
	"(Ljava/lang/String;)Ljava/lang/Class;":                         -1,
	"(Ljava/lang/String;ZLjava/lang/ClassLoader;)Ljava/lang/Class;": 1,
}

// FindForNameInits lists, in hierarchy order, the Class.forName sites that
// initialize a class of the input with a <clinit> (see ForNameInit). A
// literal naming a class outside the input, initialize=false, or a class
// with no initializer adds nothing; a non-literal name stays the
// reflection gap it already is.
func FindForNameInits(h *hierarchy.Hierarchy) ([]ForNameInit, ForNameStats, []string) {
	var out []ForNameInit
	var st ForNameStats
	var warnings []string
	var reader classfile.JarClassReader
	defer reader.Close()
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		var cf *classfile.ClassFile
		for _, m := range ci.Methods {
			var sites []classfile.Call
			for _, c := range m.Calls {
				if c.Owner != "java.lang.Class" || c.Name != "forName" {
					continue
				}
				if _, ok := forNameIndex[c.Desc]; ok {
					sites = append(sites, c)
				}
			}
			if len(sites) == 0 {
				continue
			}
			st.Sites += len(sites)
			if cf == nil {
				var err error
				if cf, err = reader.ReadClass(ci.SourceJar, ci.SourceEntry); err != nil {
					warnings = append(warnings, fmt.Sprintf("re-reading %s (%s) for Class.forName sites: %v", ci.SourceJar, ci.SourceEntry, err))
					break
				}
			}
			f, err := dataflow.Analyze(cf, fqcn, m.Method)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s#%s%s: local-value analysis: %v", fqcn, m.Name, m.Descriptor, err))
				continue
			}
			for _, c := range sites {
				site := bytecodeir.Site{Method: fqcn + "#" + m.Name, Descriptor: m.Descriptor, PC: c.PC, Op: bytecodeir.OpCall}
				name, ok := literalArg(f, site, 0)
				if !ok {
					continue
				}
				st.Literal++
				maybe := false
				if i := forNameIndex[c.Desc]; i >= 0 {
					switch init, known := intArg(f, site, i); {
					case known && init == 0:
						st.NoInit++
						continue
					case !known:
						maybe = true
					}
				}
				named, ok := h.Lookup(name)
				if !ok {
					st.Outside++
					continue
				}
				initFQCN, ok := initializerOf(h, named)
				if !ok {
					st.NoInitializer++
					continue
				}
				st.Inits++
				if maybe {
					st.MaybeInit++
				}
				out = append(out, ForNameInit{CallerFQCN: fqcn, CallerName: m.Name, CallerDesc: m.Descriptor, PC: c.PC,
					Named: name, InitFQCN: initFQCN, MaybeInit: maybe})
			}
		}
	}
	return out, st, warnings
}

// initializerOf: the class itself when it declares <clinit>, else (for a
// class) its nearest superclass in the input that does.
func initializerOf(h *hierarchy.Hierarchy, ci *hierarchy.ClassInfo) (string, bool) {
	for cur := ci; cur != nil; {
		if _, ok := cur.MethodByNameAndDesc("<clinit>", "()V"); ok {
			return cur.FQCN, true
		}
		if cur.IsInterface() || cur.SuperClass == "" {
			return "", false
		}
		next, ok := h.Lookup(cur.SuperClass)
		if !ok {
			return "", false
		}
		cur = next
	}
	return "", false
}

// literalArg: argument i at site is a provable string literal.
func literalArg(f *dataflow.MethodFacts, site bytecodeir.Site, i int) (s string, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	v := f.ArgumentAt(site, i)
	if v.Kind != valueflow.KindConst || v.Type != "java.lang.String" {
		return "", false
	}
	s, ok = v.Const.(string)
	return s, ok
}

// intArg: argument i at site is a provable int (boolean) constant.
func intArg(f *dataflow.MethodFacts, site bytecodeir.Site, i int) (n int, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	v := f.ArgumentAt(site, i)
	if v.Kind != valueflow.KindConst || v.Type != "int" {
		return 0, false
	}
	n, ok = v.Const.(int)
	return n, ok
}
