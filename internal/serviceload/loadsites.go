package serviceload

import (
	"fmt"
	"strings"

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
	callers := &lazyCallers{h: h}
	defer callers.close()

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
			found, err := loadSitesInMethod(cf, fqcn, m, callers)
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

// classLoaderArgument recognizes the ClassLoader argument shapes whose
// provenance is exact in the calling method: the current/another
// thread's context loader, the system loader (explicitly or as null),
// or some class's defining loader. Anything else — a parameter, a field,
// a constructed loader — returns "" and keeps the visibility gap.
func classLoaderArgument(v valueflow.Value) (LoaderIdentity, string) {
	switch v.Kind {
	case valueflow.KindNull:
		return LoaderIdentSystem, "null (ServiceLoader uses the system class loader)"
	case valueflow.KindCallResult:
		if v.Call != nil && v.Call.Owner == "java.lang.ClassLoader" && v.Call.Name == "getSystemClassLoader" && v.Call.Desc == "()Ljava/lang/ClassLoader;" {
			return LoaderIdentSystem, "ClassLoader.getSystemClassLoader()"
		}
	case valueflow.KindCallTrace:
		t := v.CallTrace
		if t == nil {
			break
		}
		if t.Owner == "java.lang.Thread" && t.Name == "getContextClassLoader" {
			if t.Receiver != nil && t.Receiver.Kind == valueflow.KindCallResult && t.Receiver.Call != nil &&
				t.Receiver.Call.Owner == "java.lang.Thread" && t.Receiver.Call.Name == "currentThread" {
				return LoaderIdentContext, "Thread.currentThread().getContextClassLoader()"
			}
			return LoaderIdentContext, "a thread's getContextClassLoader() (receiver not provably the current thread)"
		}
		if t.Owner == "java.lang.Class" && t.Name == "getClassLoader" {
			if t.Receiver != nil && t.Receiver.Kind == valueflow.KindConst && t.Receiver.Type == "java.lang.Class" {
				return LoaderIdentDefining, fmt.Sprintf("%s.class.getClassLoader()", t.Receiver.Const)
			}
			return LoaderIdentDefining, "getClassLoader() of a class not known statically"
		}
	}
	return "", ""
}

// moduleLayerArgument recognizes the one layer a static analysis can
// name exactly: the result of invokestatic java.lang.ModuleLayer#boot().
// Any other value — a parameter, a field, a layer built at run time —
// stays LayerUnknown and keeps the existing visibility gap.
func moduleLayerArgument(v valueflow.Value, aborted bool) (LayerKind, string) {
	if !aborted && v.Kind == valueflow.KindCallResult && v.Call != nil &&
		v.Call.Owner == "java.lang.ModuleLayer" && v.Call.Name == "boot" && v.Call.Desc == "()Ljava/lang/ModuleLayer;" {
		return LayerBoot, ""
	}
	if aborted {
		return LayerUnknown, "local-value analysis for this method aborted"
	}
	return LayerUnknown, fmt.Sprintf("the ModuleLayer argument is %s, not ModuleLayer.boot()", v.Kind)
}

// lazyCallers builds DATA_FLOW D3's caller index only when a load-site
// argument is actually a parameter.
type lazyCallers struct {
	h   *hierarchy.Hierarchy
	idx *dataflow.CallerIndex
}

func (l *lazyCallers) get() *dataflow.CallerIndex {
	if l.idx == nil {
		l.idx = dataflow.NewCallerIndex(l.h)
	}
	return l.idx
}

func (l *lazyCallers) close() {
	if l.idx != nil {
		l.idx.Close()
	}
}

// resolveParameterArg resolves a formal-parameter argument through the
// method's callers (DATA_FLOW D3).
func (l *lazyCallers) resolveParameterArg(v valueflow.Value, fqcn string, m hierarchy.Method) (dataflow.ParameterValues, bool) {
	if v.Kind != valueflow.KindTyped || v.Parameter == nil {
		return dataflow.ParameterValues{}, false
	}
	return l.get().ResolveParameter(dataflow.MethodID{Owner: fqcn, Name: m.Name, Desc: m.Descriptor}, v.Parameter.Index), true
}

// applyParameterInterfaces fills the site's service interface(s) from
// the caller-derived values of its Class parameter.
func applyParameterInterfaces(site *ServiceLoadSite, pv dataflow.ParameterValues) {
	var ifaces, other []string
	for _, v := range pv.Values {
		if v.Kind == valueflow.KindConst && v.Type == "java.lang.Class" {
			ifaces = appendUnique(ifaces, v.Const.(string))
		} else {
			other = append(other, fmt.Sprintf("a caller passes %s, not a Class literal", v.Kind))
		}
	}
	missing := append(append([]string{}, pv.Unknown...), other...)
	site.ParamSource = "Class parameter resolved through callers (DATA_FLOW D3): " + pv.Describe()
	switch {
	case len(ifaces) == 1 && len(missing) == 0:
		site.Interface = ifaces[0]
	case len(ifaces) > 0:
		site.Interfaces = ifaces
		if len(missing) > 0 {
			site.UnresolvedReason = "some callers' Class arguments are not resolved: " + strings.Join(missing, "; ")
		}
	default:
		site.UnresolvedReason = "the argument is a parameter whose callers pass no resolvable Class literal: " + strings.Join(missing, "; ")
	}
}

func appendUnique(list []string, s string) []string {
	for _, x := range list {
		if x == s {
			return list
		}
	}
	return append(list, s)
}

// parameterLoaderIdentity recognizes a ClassLoader parameter only when
// every caller-derived value is a recognized loader shape.
func parameterLoaderIdentity(pv dataflow.ParameterValues) (LoaderIdentity, string) {
	if !pv.Complete() {
		return "", ""
	}
	var ident LoaderIdentity
	var details []string
	for _, v := range pv.Values {
		id, detail := classLoaderArgument(v)
		if id == "" {
			return "", ""
		}
		if ident == "" {
			ident = id
		} else if ident != id {
			ident = LoaderIdentMixed
		}
		details = append(details, detail)
	}
	return ident, "parameter; callers pass " + strings.Join(details, ", ")
}

func loadSitesInMethod(cf *classfile.ClassFile, fqcn string, m hierarchy.Method, callers *lazyCallers) ([]ServiceLoadSite, error) {
	facts, err := dataflow.AnalyzeWithParameterRefs(cf, fqcn, m.Method)
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
		pv, isParam := callers.resolveParameterArg(val, fqcn, m)
		switch {
		case val.Kind == valueflow.KindConst && val.Type == "java.lang.Class":
			site.Interface = val.Const.(string)
		case !aborted && isParam:
			site.ClassParamSet, site.ClassParamIndex = true, val.Parameter.Index
			applyParameterInterfaces(&site, pv)
		case aborted:
			site.UnresolvedReason = "local-value analysis for this method aborted: " + abortReason
		case val.Kind == valueflow.KindUnknown:
			site.UnresolvedReason = "the argument is unknown: " + val.Reason
		default:
			site.UnresolvedReason = fmt.Sprintf("the argument is %s, not a resolvable Class literal", val.Kind)
		}
		if ov.kind == LoaderExplicit && !aborted {
			loaderArg := facts.ArgumentAt(argSite, 1)
			site.LoaderIdentity, site.LoaderDetail = classLoaderArgument(loaderArg)
			if lpv, ok := callers.resolveParameterArg(loaderArg, fqcn, m); ok && site.LoaderIdentity == "" {
				site.LoaderIdentity, site.LoaderDetail = parameterLoaderIdentity(lpv)
			}
		}
		if ov.kind == LoaderModuleLayer {
			site.Layer, site.LayerReason = moduleLayerArgument(facts.ArgumentAt(argSite, 0), aborted)
		}
		sites = append(sites, site)
	}
	return sites, nil
}
