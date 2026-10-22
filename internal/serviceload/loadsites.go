package serviceload

import (
	"fmt"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

const serviceLoaderOwner = "java.util.ServiceLoader"

// overloadKey is one java.util.ServiceLoader factory method's exact
// name+descriptor — needed together, not descriptor alone, since
// load(Class) and loadInstalled(Class) share a descriptor but differ in
// which class loader they use (LoaderKind).
type overloadKey struct{ name, desc string }

// serviceLoaderOverloads maps every ServiceLoader factory method
// ariadne knows about to which argument (0-based, matching
// valueflow.Facts.ArgumentAt's own indexing) holds the service
// interface's Class, and which LoaderKind it is.
var serviceLoaderOverloads = map[overloadKey]struct {
	argIndex int
	kind     LoaderKind
}{
	{"load", "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"}:                        {0, LoaderClasspath},
	{"load", "(Ljava/lang/Class;Ljava/lang/ClassLoader;)Ljava/util/ServiceLoader;"}: {0, LoaderExplicit},
	{"load", "(Ljava/lang/ModuleLayer;Ljava/lang/Class;)Ljava/util/ServiceLoader;"}: {1, LoaderModuleLayer},
	{"loadInstalled", "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"}:               {0, LoaderInstalled},
}

// FindLoadSites is SERVICE_LOADING S1: it finds every
// ServiceLoader.load(...)/loadInstalled(...) call site anywhere in h
// (app and dependencies both — a dependency can register its own
// providers with itself, same as S0's own inventory doesn't special-case
// the app) and resolves each one's service-interface argument via
// DATA_FLOW D0-D2's local-value analysis. It never creates a call-graph
// edge (same discipline as S0's Inventory): a ServiceLoadSite is a fact
// about what a method asks for, not a claim that any provider's
// constructor runs — that is S2.
//
// h's own Method.Calls (already extracted while building the hierarchy)
// is used to cheaply find candidate classes without re-parsing anything;
// a class file is only re-read — once per class, however many candidate
// call sites it has — when it actually contains one, since
// classfile.ClassFile's constant pool (which DecodeLocalOps needs to
// resolve ldc/field/call references) is deliberately not retained past
// hierarchy.Build (see hierarchy.Method's own doc comment).
func FindLoadSites(h *hierarchy.Hierarchy) ([]ServiceLoadSite, []string, error) {
	var sites []ServiceLoadSite
	var warnings []string
	var jarReader classfile.JarClassReader
	defer jarReader.Close()

	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			continue
		}
		if !anyServiceLoaderCall(ci.Methods) {
			continue
		}

		cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf(
				"re-reading %s (%s) for ServiceLoader call-site analysis: %v", ci.SourceJar, ci.SourceEntry, err))
			continue
		}

		for _, m := range ci.Methods {
			if !anyServiceLoaderCall([]hierarchy.Method{m}) {
				continue
			}
			found, err := loadSitesInMethod(cf, fqcn, m)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf(
					"%s#%s%s: local-value analysis: %v", fqcn, m.Name, m.Descriptor, err))
				continue
			}
			sites = append(sites, found...)
		}
	}
	return sites, warnings, nil
}

func anyServiceLoaderCall(methods []hierarchy.Method) bool {
	for _, m := range methods {
		for _, c := range m.Calls {
			if c.Owner != serviceLoaderOwner {
				continue
			}
			if _, ok := serviceLoaderOverloads[overloadKey{c.Name, c.Desc}]; ok {
				return true
			}
		}
	}
	return false
}

func loadSitesInMethod(cf *classfile.ClassFile, fqcn string, m hierarchy.Method) ([]ServiceLoadSite, error) {
	facts, err := dataflow.Analyze(cf, fqcn, m.Method)
	if err != nil {
		return nil, err
	}
	aborted, abortReason := facts.Aborted()

	var sites []ServiceLoadSite
	for _, c := range m.Calls {
		if c.Owner != serviceLoaderOwner {
			continue
		}
		ov, ok := serviceLoaderOverloads[overloadKey{c.Name, c.Desc}]
		if !ok {
			continue
		}

		site := ServiceLoadSite{
			CallerMethod: fmt.Sprintf("%s#%s%s", fqcn, m.Name, m.Descriptor),
			CallerFQCN:   fqcn,
			CallerName:   m.Name,
			CallerDesc:   m.Descriptor,
			LoaderKind:   ov.kind,
			PC:           c.PC,
		}
		argSite := bytecodeir.Site{Method: fqcn + "#" + m.Name, Descriptor: m.Descriptor, PC: c.PC, Op: bytecodeir.OpCall}
		val := facts.ArgumentAt(argSite, ov.argIndex)
		switch {
		case val.Kind == valueflow.KindConst && val.Type == "java.lang.Class":
			site.Interface = val.Const.(string)
		case aborted:
			site.UnresolvedReason = "local-value analysis for this method aborted: " + abortReason
		case val.Kind == valueflow.KindUnknown:
			site.UnresolvedReason = "the argument is unknown: " + val.Reason
		default:
			site.UnresolvedReason = fmt.Sprintf("the argument is %s, not a resolvable Class literal", val.Kind)
		}
		sites = append(sites, site)
	}
	return sites, nil
}
