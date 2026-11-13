// Package gaps is the one vocabulary of model gaps (ROADMAP B0 b0-3): the
// mechanisms through which the static model may miss a call, shared by
// every analyzer that reports gaps. Each analyzer keeps its own codes —
// they are part of its output and are never renamed — and each code maps
// to exactly one mechanism here, so the same gap reads the same way
// wherever it is reported. The mapping is a reading aid, not a rule: it
// changes no verdict (the model-gap rule of docs/RESULTS_CONTRACT.MD §4 is
// untouched).
//
// The mechanisms and the mapping are listed in docs/RESULTS_CONTRACT.MD
// ("Gap vocabulary"); TestVocabularyDocumented keeps the two identical.
package gaps

import (
	"fmt"
	"sort"
)

// Version of the vocabulary and mapping.
const Version = "b0-3"

// Mechanism is a canonical gap mechanism.
type Mechanism string

// Mechanisms.
const (
	Reflection        Mechanism = "reflection"         // Method#invoke, Constructor#newInstance, Class.forName with an unresolved value
	MethodHandle      Mechanism = "method_handle"      // MethodHandle#invoke*, Lookup#find* with an unresolved value
	CallbackAsync     Mechanism = "callback_async"     // a task or callback handed to a thread, executor or framework
	ServiceLoader     Mechanism = "service_loader"     // ServiceLoader use whose service or providers are unknown
	Deserialization   Mechanism = "deserialization"    // a deserialization entry whose input classes are open
	DynamicProxy      Mechanism = "dynamic_proxy"      // a JDK Proxy / generated InvocationHandler dispatch
	ContainerProxy    Mechanism = "container_proxy"    // a container-generated implementation (repository, AOP proxy)
	UnknownAllocation Mechanism = "unknown_allocation" // an object created where the type is unknown (Unsafe, constructor handles)
	ExternalInput     Mechanism = "external_input"     // roots=all: a root's parameters may be any subtype
	Lambda            Mechanism = "lambda"             // a functional-interface call whose lambda / method reference is unresolved
	SkippedReceiver   Mechanism = "skipped_receiver"   // a call on java./javax./jdk./sun. types, never CHA-expanded
	VirtualDispatch   Mechanism = "virtual_dispatch"   // a virtual call target the graph has no edge for (RTA pruning, inherited bodies)
	ClassInit         Mechanism = "class_init"         // a class initializer the JVM runs implicitly
	ModelIncomplete   Mechanism = "model_incomplete"   // a library model that could not resolve a site (B3 coordinator)
	TraceAlignment    Mechanism = "trace_alignment"    // a measured call not attributable to one site (dropped frames, inlining)
	Unclassified      Mechanism = "unclassified"       // no mechanism known
)

// Mechanisms lists every mechanism, in documentation order.
var Mechanisms = []Mechanism{Reflection, MethodHandle, CallbackAsync, ServiceLoader, Deserialization, DynamicProxy, ContainerProxy,
	UnknownAllocation, ExternalInput, Lambda, SkippedReceiver, VirtualDispatch, ClassInit, ModelIncomplete, TraceAlignment, Unclassified}

// Sources: the analyzers whose own codes are mapped.
const (
	SourceSite        = "site"        // unresolved dynamic sites of the model-gap rule (reach.GapSite kinds)
	SourceAllocation  = "allocation"  // the created-type catalog's unknown allocations (internal/alloc)
	SourceLambda      = "lambda"      // SAM-call resolution gaps (internal/samcall)
	SourceRuntime     = "runtime"     // causes of measured calls missing from the graph (internal/runtimeevents)
	SourceCoordinator = "coordinator" // incomplete model facts of the B3 coordinator (detail: model id)
)

// Sources lists the sources in documentation order.
var Sources = []string{SourceSite, SourceAllocation, SourceLambda, SourceRuntime, SourceCoordinator}

// Mapping is each source's codes and their mechanism. The coordinator's
// detail is a model id; every coordinator gap is ModelIncomplete.
var Mapping = map[string]map[string]Mechanism{
	SourceSite: {
		"reflection": Reflection, "methodhandle": MethodHandle, "callback": CallbackAsync, "serviceloader": ServiceLoader,
		"deserialization": Deserialization, "allocation": UnknownAllocation,
	},
	SourceAllocation: {
		"reflection-unresolved": Reflection, "constructor-handle": MethodHandle, "service-unresolved": ServiceLoader,
		"dynamic-proxy": DynamicProxy, "container-proxy": ContainerProxy, "unsafe-allocation": UnknownAllocation,
		"external-input": ExternalInput,
	},
	SourceLambda: {
		"parameter": Lambda, "sam-argument": Lambda, "field": Lambda, "call-result": Lambda, "instance-call": Lambda,
		"unknown-value": Lambda, "depth": Lambda, "bootstrap": Lambda, "budget": Lambda,
	},
	SourceRuntime: {
		"reflection": Reflection, "method-handle": MethodHandle, "async": CallbackAsync, "proxy": DynamicProxy,
		"sam-call": Lambda, "invokedynamic": Lambda, "skipped-receiver": SkippedReceiver, "rta-pruned": VirtualDispatch,
		"inherited-impl": VirtualDispatch, "dispatch-missing": VirtualDispatch, "class-init": ClassInit,
		"indirect": TraceAlignment, "no-call-at-site": TraceAlignment, "unclassified": Unclassified,
	},
}

// Classify returns a source code's mechanism; Unclassified for a code the
// mapping does not know (a test keeps every analyzer's codes mapped).
func Classify(source, code string) Mechanism {
	if source == SourceCoordinator {
		return ModelIncomplete
	}
	if m, ok := Mapping[source][code]; ok {
		return m
	}
	return Unclassified
}

// Gap is a count of gaps of one source code, with its mechanism.
type Gap struct {
	Mechanism Mechanism `json:"mechanism"`
	Source    string    `json:"source"`
	// Detail is the source's own code (or, for the coordinator, the model
	// id), unchanged.
	Detail string `json:"detail"`
	Count  int    `json:"count"`
	// Reachable (source "site" only): how many of Count have a caller the
	// roots reach — the ones the model-gap rule considers.
	Reachable *int `json:"reachable,omitempty"`
}

// FromCounts turns a source's code → count map into gaps.
func FromCounts(source string, counts map[string]int) []Gap {
	var out []Gap
	for code, n := range counts {
		if n > 0 {
			out = append(out, Gap{Mechanism: Classify(source, code), Source: source, Detail: code, Count: n})
		}
	}
	return Sort(out)
}

// Sort orders gaps by mechanism (documentation order), source, detail.
func Sort(gs []Gap) []Gap {
	rank := map[Mechanism]int{}
	for i, m := range Mechanisms {
		rank[m] = i
	}
	srank := map[string]int{}
	for i, s := range Sources {
		srank[s] = i
	}
	sort.Slice(gs, func(i, j int) bool {
		a, b := gs[i], gs[j]
		if a.Mechanism != b.Mechanism {
			return rank[a.Mechanism] < rank[b.Mechanism]
		}
		if a.Source != b.Source {
			return srank[a.Source] < srank[b.Source]
		}
		return a.Detail < b.Detail
	})
	return gs
}

// ByMechanism totals gaps per mechanism, for one-line summaries.
func ByMechanism(gs []Gap) string {
	tot := map[Mechanism]int{}
	for _, g := range gs {
		tot[g.Mechanism] += g.Count
	}
	s := ""
	for _, m := range Mechanisms {
		if tot[m] > 0 {
			if s != "" {
				s += ", "
			}
			s += fmt.Sprintf("%s %d", m, tot[m])
		}
	}
	return s
}
