package serviceload

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// FindConsumptionSites resolves selected iterator/stream operations back to
// their exact ServiceLoader.load call when DATA_FLOW has preserved a local
// receiver chain. Calls through fields, parameters, or unsupported operations
// remain explicit gaps for the ServiceLoader/Provider-specific APIs.
func FindConsumptionSites(h *hierarchy.Hierarchy, loads []ServiceLoadSite) ([]ServiceConsumptionSite, []string, error) {
	loadBySite := make(map[siteKey]ServiceLoadSite, len(loads))
	for _, load := range loads {
		loadBySite[siteKey{load.CallerFQCN, load.CallerName + load.CallerDesc, load.PC}] = load
	}

	var sites []ServiceConsumptionSite
	var warnings []string
	var jarReader classfile.JarClassReader
	defer jarReader.Close()

	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !anyServiceConsumptionCall(ci.Methods) {
			continue
		}
		cf, err := jarReader.ReadClass(ci.SourceJar, ci.SourceEntry)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("re-reading %s (%s) for ServiceLoader consumption analysis: %v", ci.SourceJar, ci.SourceEntry, err))
			continue
		}
		for _, method := range ci.Methods {
			if !anyServiceConsumptionCall([]hierarchy.Method{method}) {
				continue
			}
			facts, err := dataflow.Analyze(cf, fqcn, method.Method)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s#%s%s: ServiceLoader consumption analysis: %v", fqcn, method.Name, method.Descriptor, err))
				continue
			}
			aborted, abortReason := facts.Aborted()
			for _, call := range method.Calls {
				isConsumption := isServiceConsumptionCall(call.Owner, call.Name, call.Desc)
				if !isConsumption && call.Kind != classfile.CallVirtual && call.Kind != classfile.CallInterface {
					continue
				}
				at := bytecodeir.Site{Method: fqcn + "#" + method.Name, Descriptor: method.Descriptor, PC: call.PC, Op: bytecodeir.OpCall}
				var origins serviceOrigins
				var receiver valueflow.Value
				if aborted {
					if !isConsumption {
						continue
					}
					origins.Reason = "local-value analysis for this method aborted: " + abortReason
				} else {
					receiver = facts.ReceiverAt(at)
					origins = serviceLoadOrigins(receiver, loadBySite)
				}
				providerMethodCall := false
				if !isConsumption {
					if !linkedProviderMethodCall(receiver, call.Owner, origins) {
						continue
					}
					providerMethodCall = true
				}
				// Iterator calls are common outside ServiceLoader. Only retain
				// those for which the receiver trace actually names a load site.
				if isIteratorStep(call.Owner, call.Name, call.Desc) && !origins.ServiceLoaderValue {
					continue
				}
				site := ServiceConsumptionSite{
					CallerFQCN: fqcn, CallerName: method.Name, CallerDesc: method.Descriptor,
					PC: call.PC, CallOwner: call.Owner, CallName: call.Name, CallDesc: call.Desc,
					ProviderMethodCall: providerMethodCall,
					Interfaces:         origins.Interfaces, LoaderKinds: origins.LoaderKinds,
					UnresolvedReason: origins.Reason,
				}
				if aborted {
					site.UnresolvedReason = "local-value analysis for this method aborted: " + abortReason
				}
				sites = append(sites, site)
			}
		}
	}
	return sites, warnings, nil
}

func linkedProviderMethodCall(receiver valueflow.Value, callOwner string, origins serviceOrigins) bool {
	return hasMaterializedProviderOrigin(receiver) && origins.ServiceLoaderValue && len(origins.Interfaces) > 0 && containsString(origins.Interfaces, callOwner)
}

func hasMaterializedProviderOrigin(v valueflow.Value) bool {
	found := false
	var visit func(valueflow.Value, int)
	visit = func(current valueflow.Value, depth int) {
		if found || depth > valueflow.MaxCallTraceDepth+2 {
			return
		}
		switch current.Kind {
		case valueflow.KindCallTrace:
			if current.CallTrace == nil {
				return
			}
			trace := current.CallTrace
			if (trace.Owner == "java.util.Iterator" && trace.Name == "next" && trace.Desc == "()Ljava/lang/Object;") || (trace.Owner == "java.util.ServiceLoader$Provider" && trace.Name == "get" && trace.Desc == "()Ljava/lang/Object;") {
				found = true
				return
			}
			if trace.Receiver != nil {
				visit(*trace.Receiver, depth+1)
			}
		case valueflow.KindSet:
			for _, candidate := range current.Candidates {
				visit(candidate, depth+1)
			}
		}
	}
	visit(v, 0)
	return found
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func anyServiceConsumptionCall(methods []hierarchy.Method) bool {
	for _, method := range methods {
		for _, call := range method.Calls {
			if isServiceLoaderConsumptionAnchor(call.Owner, call.Name, call.Desc) {
				return true
			}
		}
	}
	return false
}

func isServiceLoaderConsumptionAnchor(owner, name, desc string) bool {
	return (owner == serviceLoaderOwner && ((name == "iterator" && desc == "()Ljava/util/Iterator;") || (name == "stream" && desc == "()Ljava/util/stream/Stream;"))) || (owner == "java.util.ServiceLoader$Provider" && name == "get" && desc == "()Ljava/lang/Object;")
}

func isServiceConsumptionCall(owner, name, desc string) bool {
	switch {
	case owner == serviceLoaderOwner && name == "iterator" && desc == "()Ljava/util/Iterator;":
		return true
	case owner == serviceLoaderOwner && name == "stream" && desc == "()Ljava/util/stream/Stream;":
		return true
	case owner == "java.util.ServiceLoader$Provider" && name == "get" && desc == "()Ljava/lang/Object;":
		return true
	case owner == "java.util.Iterator" && name == "hasNext" && desc == "()Z":
		return true
	case owner == "java.util.Iterator" && name == "next" && desc == "()Ljava/lang/Object;":
		return true
	default:
		return false
	}
}

func isIteratorStep(owner, name, desc string) bool {
	return owner == "java.util.Iterator" && ((name == "hasNext" && desc == "()Z") || (name == "next" && desc == "()Ljava/lang/Object;"))
}

type serviceOrigins struct {
	ServiceLoaderValue bool
	Interfaces         []string
	LoaderKinds        []LoaderKind
	Reason             string
}

func serviceLoadOrigins(v valueflow.Value, loads map[siteKey]ServiceLoadSite) serviceOrigins {
	var out serviceOrigins
	interfaces := map[string]bool{}
	loaderKinds := map[LoaderKind]bool{}
	var reasons []string
	var visit func(valueflow.Value, int)
	visit = func(current valueflow.Value, depth int) {
		if depth > valueflow.MaxCallTraceDepth+2 {
			reasons = append(reasons, "ServiceLoader receiver provenance exceeded the call-trace depth limit")
			return
		}
		switch current.Kind {
		case valueflow.KindCallResult:
			if current.Call == nil {
				return
			}
			if current.Call.Owner != serviceLoaderOwner {
				return
			}
			if _, isLoad := serviceLoaderOverloads[overloadKey{current.Call.Name, current.Call.Desc}]; !isLoad {
				return
			}
			out.ServiceLoaderValue = true
			key, ok := siteKeyFromIR(current.Call.Site)
			if !ok {
				reasons = append(reasons, "load call result has no usable bytecode-site identity")
				return
			}
			load, ok := loads[key]
			if !ok {
				reasons = append(reasons, "load call result has no matching resolved ServiceLoader load-site record")
				return
			}
			if load.Interface == "" {
				reasons = append(reasons, load.UnresolvedReason)
				return
			}
			interfaces[load.Interface] = true
			loaderKinds[load.LoaderKind] = true
		case valueflow.KindCallTrace:
			if current.CallTrace != nil && current.CallTrace.Receiver != nil {
				visit(*current.CallTrace.Receiver, depth+1)
			}
		case valueflow.KindSet:
			for _, candidate := range current.Candidates {
				visit(candidate, depth+1)
			}
		}
	}
	visit(v, 0)
	for iface := range interfaces {
		out.Interfaces = append(out.Interfaces, iface)
	}
	for loader := range loaderKinds {
		out.LoaderKinds = append(out.LoaderKinds, loader)
	}
	sort.Strings(out.Interfaces)
	sort.Slice(out.LoaderKinds, func(i, j int) bool { return out.LoaderKinds[i] < out.LoaderKinds[j] })
	if len(reasons) > 0 {
		out.Reason = strings.Join(uniqueStrings(reasons), "; ")
	} else if !out.ServiceLoaderValue {
		out.Reason = "receiver is not traceably derived from a ServiceLoader.load() call in this method"
	} else if len(out.Interfaces) == 0 {
		out.Reason = "ServiceLoader load-site identity was found, but its service interface is unresolved"
	}
	return out
}

func siteKeyFromIR(site bytecodeir.Site) (siteKey, bool) {
	i := strings.LastIndexByte(site.Method, '#')
	if i <= 0 || site.Descriptor == "" {
		return siteKey{}, false
	}
	return siteKey{caller: site.Method[:i], desc: site.Method[i+1:] + site.Descriptor, pc: site.PC}, true
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
