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
	streamTerminals := map[siteKey]bool{}
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
			methodKey := siteKey{caller: fqcn, desc: method.Name + method.Descriptor}
			streamTerminals[methodKey] = hasStreamTerminal(method.Calls)
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
					origins = serviceLoadOrigins(receiver, loadBySite, h)
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
					streamLoadSites:  origins.LoadSites,
					streamTraceSites: origins.StreamTraceSites,
				}
				if aborted {
					site.UnresolvedReason = "local-value analysis for this method aborted: " + abortReason
				}
				sites = append(sites, site)
			}
		}
	}
	for i := range sites {
		stream := &sites[i]
		if stream.CallOwner != serviceLoaderOwner || stream.CallName != "stream" {
			continue
		}
		methodKey := siteKey{caller: stream.CallerFQCN, desc: stream.CallerName + stream.CallerDesc}
		stream.streamTerminalSeen = streamTerminals[methodKey]
		if !stream.streamTerminalSeen {
			continue
		}
		for j := range sites {
			providerGet := &sites[j]
			if providerGet.CallerFQCN != stream.CallerFQCN || providerGet.CallerName != stream.CallerName || providerGet.CallerDesc != stream.CallerDesc || providerGet.CallOwner != "java.util.ServiceLoader$Provider" || providerGet.CallName != "get" || providerGet.UnresolvedReason != "" {
				continue
			}
			if containsTraceSite(providerGet.streamTraceSites, bytecodeir.Site{
				Method: stream.CallerFQCN + "#" + stream.CallerName, Descriptor: stream.CallerDesc,
				PC: stream.PC, Op: bytecodeir.OpCall,
			}) && intersectsLoadSites(stream.streamLoadSites, providerGet.streamLoadSites) {
				stream.streamTerminalSupported = true
				break
			}
		}
	}
	return sites, warnings, nil
}

func containsTraceSite(sites []bytecodeir.Site, target bytecodeir.Site) bool {
	for _, site := range sites {
		if site == target {
			return true
		}
	}
	return false
}

func hasStreamTerminal(calls []classfile.Call) bool {
	for _, call := range calls {
		if !strings.HasPrefix(call.Owner, "java.util.stream.") && call.Owner != "java.util.stream.Stream" {
			continue
		}
		switch call.Name {
		case "findFirst", "findAny", "forEach", "forEachOrdered", "toArray", "reduce", "collect", "min", "max", "count":
			return true
		}
	}
	return false
}

func intersectsLoadSites(a, b []siteKey) bool {
	for _, left := range a {
		for _, right := range b {
			if left == right {
				return true
			}
		}
	}
	return false
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
	LoadSites          []siteKey
	StreamTraceSites   []bytecodeir.Site
}

func serviceLoadOrigins(v valueflow.Value, loads map[siteKey]ServiceLoadSite, h *hierarchy.Hierarchy) serviceOrigins {
	var out serviceOrigins
	interfaces := map[string]bool{}
	loaderKinds := map[LoaderKind]bool{}
	loadSites := map[siteKey]bool{}
	streamTraceSites := map[bytecodeir.Site]bool{}
	var reasons []string
	resolvedValues := map[string]bool{}
	var substituteParameters func(valueflow.Value, valueflow.CallResultRef) (valueflow.Value, []string)
	substituteParameters = func(current valueflow.Value, call valueflow.CallResultRef) (valueflow.Value, []string) {
		switch current.Kind {
		case valueflow.KindTyped:
			if current.Parameter == nil {
				return current, nil
			}
			return dataflow.ResolveCallArgumentAtCall(h, call, current.Parameter.Index)
		case valueflow.KindSet:
			var result valueflow.Value
			var warnings []string
			for i, candidate := range current.Candidates {
				substituted, candidateWarnings := substituteParameters(candidate, call)
				warnings = append(warnings, candidateWarnings...)
				if i == 0 {
					result = substituted
				} else {
					result = valueflow.Merge(result, substituted)
				}
			}
			return result, warnings
		default:
			return current, nil
		}
	}
	// via is the most recent exact helper call this trace came through:
	// a load site whose Class argument is that helper's own parameter is
	// narrowed to the argument at that call site (context-sensitive), so
	// two callers of one helper do not share each other's services.
	var visit func(valueflow.Value, int, *valueflow.CallResultRef)
	visit = func(current valueflow.Value, depth int, via *valueflow.CallResultRef) {
		if depth > valueflow.MaxCallTraceDepth+2 {
			reasons = append(reasons, "ServiceLoader receiver provenance exceeded the call-trace depth limit")
			return
		}
		switch current.Kind {
		case valueflow.KindCallResult:
			if current.Call == nil {
				return
			}
			if current.Call.Owner == serviceLoaderOwner {
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
				candidates := load.Candidates()
				unresolved := load.UnresolvedReason
				if iface, ok := narrowToCallSite(h, load, via); ok {
					candidates, unresolved = []string{iface}, ""
				}
				if len(candidates) == 0 {
					reasons = append(reasons, unresolved)
					return
				}
				if unresolved != "" {
					reasons = append(reasons, unresolved) // partial caller-derived set
				}
				loadSites[key] = true
				for _, iface := range candidates {
					interfaces[iface] = true
				}
				loaderKinds[load.LoaderKind] = true
				return
			}

			// DATA_FLOW D3 resolves only exact invokestatic/invokespecial
			// results. Follow a helper's returned value to a ServiceLoader
			// factory result, preserving that factory's own bytecode site.
			key := callResultSiteKey(*current.Call)
			if resolvedValues[key] {
				reasons = append(reasons, "service provenance revisits the same helper call result")
				return
			}
			resolvedValues[key] = true
			if h == nil {
				reasons = append(reasons, "helper call result cannot be resolved without a class hierarchy")
				return
			}
			resolved, warnings := dataflow.ResolveCallResultOneHopWithParameterRefs(h, *current.Call)
			reasons = append(reasons, warnings...)
			resolved, substitutionWarnings := substituteParameters(resolved, *current.Call)
			reasons = append(reasons, substitutionWarnings...)
			if resolved.Kind == valueflow.KindUnknown {
				if resolved.Reason != "" {
					reasons = append(reasons, "helper call result is unresolved: "+resolved.Reason)
				}
				return
			}
			visit(resolved, depth+1, current.Call)
		case valueflow.KindField:
			if current.Field == nil {
				return
			}
			ref := *current.Field
			key := "field:" + ref.Owner + "#" + ref.Name + ":" + ref.Desc
			if resolvedValues[key] {
				reasons = append(reasons, "service provenance revisits the same field")
				return
			}
			resolvedValues[key] = true
			if h == nil {
				reasons = append(reasons, "field value cannot be resolved without a class hierarchy")
				return
			}
			resolved, warnings := dataflow.ResolveFieldOneHop(h, ref)
			reasons = append(reasons, warnings...)
			if resolved.Kind == valueflow.KindUnknown {
				if resolved.Reason != "" {
					reasons = append(reasons, "field value is unresolved: "+resolved.Reason)
				}
				return
			}
			visit(resolved, depth+1, via)
		case valueflow.KindCallTrace:
			if current.CallTrace != nil {
				if current.CallTrace.Owner == serviceLoaderOwner && current.CallTrace.Name == "stream" {
					streamTraceSites[current.CallTrace.Site] = true
				}
				if current.CallTrace.Receiver != nil {
					visit(*current.CallTrace.Receiver, depth+1, via)
				}
			}
		case valueflow.KindSet:
			for _, candidate := range current.Candidates {
				visit(candidate, depth+1, via)
			}
		}
	}
	visit(v, 0, nil)
	for iface := range interfaces {
		out.Interfaces = append(out.Interfaces, iface)
	}
	for loader := range loaderKinds {
		out.LoaderKinds = append(out.LoaderKinds, loader)
	}
	for key := range loadSites {
		out.LoadSites = append(out.LoadSites, key)
	}
	for site := range streamTraceSites {
		out.StreamTraceSites = append(out.StreamTraceSites, site)
	}
	sort.Slice(out.StreamTraceSites, func(i, j int) bool {
		if out.StreamTraceSites[i].Method != out.StreamTraceSites[j].Method {
			return out.StreamTraceSites[i].Method < out.StreamTraceSites[j].Method
		}
		if out.StreamTraceSites[i].Descriptor != out.StreamTraceSites[j].Descriptor {
			return out.StreamTraceSites[i].Descriptor < out.StreamTraceSites[j].Descriptor
		}
		return out.StreamTraceSites[i].PC < out.StreamTraceSites[j].PC
	})
	sort.Slice(out.LoadSites, func(i, j int) bool {
		if out.LoadSites[i].caller != out.LoadSites[j].caller {
			return out.LoadSites[i].caller < out.LoadSites[j].caller
		}
		if out.LoadSites[i].desc != out.LoadSites[j].desc {
			return out.LoadSites[i].desc < out.LoadSites[j].desc
		}
		return out.LoadSites[i].pc < out.LoadSites[j].pc
	})
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

func callResultSiteKey(ref valueflow.CallResultRef) string {
	return ref.Owner + "#" + ref.Name + ref.Desc + "@" + ref.Site.Method + ref.Site.Descriptor + ":" + fmt.Sprint(ref.Site.PC)
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

// narrowToCallSite resolves a load site's Class parameter from the exact
// helper call the receiver trace came through, when the load site sits
// directly in that helper and the argument there is a Class literal.
func narrowToCallSite(h *hierarchy.Hierarchy, load ServiceLoadSite, via *valueflow.CallResultRef) (string, bool) {
	if via == nil || h == nil || !load.ClassParamSet {
		return "", false
	}
	decl, ok := h.ResolveExact(via.Owner, via.Name, via.Desc)
	if !ok || decl != load.CallerFQCN || via.Name != load.CallerName || via.Desc != load.CallerDesc {
		return "", false
	}
	arg, _ := dataflow.ResolveCallArgumentAtCall(h, *via, load.ClassParamIndex)
	if arg.Kind == valueflow.KindConst && arg.Type == "java.lang.Class" {
		return arg.Const.(string), true
	}
	return "", false
}
