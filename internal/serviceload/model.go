package serviceload

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/models"
)

const serviceLoaderModelID = "jdk:service-loader-classpath-provider"
const serviceLoaderConsumptionModelID = "jdk:service-loader-consumption"
const serviceLoaderDispatchModelID = "jdk:service-loader-provider-dispatch"

// Exported model IDs and the version of their rules (labelled by the
// SERVICE_LOADING phase that closed the area, S4), for the coordinator's participant
// catalog (ROADMAP B3 b3-3).
const (
	LoadModelID        = serviceLoaderModelID
	ConsumptionModelID = serviceLoaderConsumptionModelID
	DispatchModelID    = serviceLoaderDispatchModelID
	ModelVersion       = "s4"
)

// RegisterModel adds the classpath ServiceLoader semantics implemented
// by the S0-S2 inventory and resolver. Its exact API signatures are kept
// in sync with FindLoadSites; modules and custom classloaders remain gaps.
func RegisterModel(registry *models.Registry) {
	registry.Register(&models.Model{
		ID:     serviceLoaderModelID,
		Source: "JDK ServiceLoader API and classpath META-INF/services metadata (https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/ServiceLoader.html)",
		Signatures: []models.Signature{
			{Owner: serviceLoaderOwner, Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"},
			{Owner: serviceLoaderOwner, Name: "load", Desc: "(Ljava/lang/Class;Ljava/lang/ClassLoader;)Ljava/util/ServiceLoader;"},
			{Owner: serviceLoaderOwner, Name: "load", Desc: "(Ljava/lang/ModuleLayer;Ljava/lang/Class;)Ljava/util/ServiceLoader;"},
			{Owner: serviceLoaderOwner, Name: "loadInstalled", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"},
		},
	})
	// Consumption APIs are registered separately from load(): exact
	// call-site evaluation can link a supported local receiver chain to a
	// load result, while unsupported flows remain explicit gaps.
	registry.Register(&models.Model{
		ID:     serviceLoaderConsumptionModelID,
		Source: "JDK ServiceLoader API (https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/ServiceLoader.html)",
		Signatures: []models.Signature{
			{Owner: serviceLoaderOwner, Name: "iterator", Desc: "()Ljava/util/Iterator;"},
			{Owner: serviceLoaderOwner, Name: "stream", Desc: "()Ljava/util/stream/Stream;"},
			{Owner: "java.util.ServiceLoader$Provider", Name: "get", Desc: "()Ljava/lang/Object;"},
			{Owner: "java.util.Iterator", Name: "hasNext", Desc: "()Z"},
			{Owner: "java.util.Iterator", Name: "next", Desc: "()Ljava/lang/Object;"},
		},
	})
}

// RegisterDispatchModels adds exact interface-method signatures that were
// found with a receiver trace rooted at Iterator.next() or Provider.get().
// Registration is per analyzed hierarchy; no general method name is guessed.
func RegisterDispatchModels(registry *models.Registry, sites []ServiceConsumptionSite) {
	seen := map[models.Signature]bool{}
	var signatures []models.Signature
	for _, site := range sites {
		if !site.ProviderMethodCall || len(site.Interfaces) == 0 {
			continue
		}
		sig := models.Signature{Owner: site.CallOwner, Name: site.CallName, Desc: site.CallDesc}
		if seen[sig] {
			continue
		}
		seen[sig] = true
		signatures = append(signatures, sig)
	}
	if len(signatures) == 0 {
		return
	}
	sort.Slice(signatures, func(i, j int) bool {
		if signatures[i].Owner != signatures[j].Owner {
			return signatures[i].Owner < signatures[j].Owner
		}
		if signatures[i].Name != signatures[j].Name {
			return signatures[i].Name < signatures[j].Name
		}
		return signatures[i].Desc < signatures[j].Desc
	})
	registry.Register(&models.Model{
		ID:         serviceLoaderDispatchModelID,
		Source:     "bounded local receiver provenance from ServiceLoader consumption and provider registrations",
		Signatures: signatures,
	})
}

// ModelEvaluator turns the already-resolved S1/S2 facts into coordinator
// facts. Provider edges are keyed by exact caller method and bytecode PC,
// so a declaration cannot create an edge at an unrelated load site.
// Lazy materialization is attached to iterator steps or Provider.get only
// when bounded local receiver provenance reaches the exact load site.
type ModelEvaluator struct {
	h                    *hierarchy.Hierarchy
	sites                map[siteKey]ServiceLoadSite
	consumptions         map[siteKey]ServiceConsumptionSite
	dispatches           map[siteKey]ServiceConsumptionSite
	providersByInterface map[string][]ProviderMaterialization
	scenario             ScenarioView
}

// SetScenario selects the JPMS launch scenario (SERVICE_LOADING S3).
// The materializations passed to NewModelEvaluator must already come
// from the same view's Registrations; this adds the scenario's
// per-caller `uses` check. The zero value is the class-path scenario.
func (e *ModelEvaluator) SetScenario(view ScenarioView) { e.scenario = view }

func (e *ModelEvaluator) noProviderReason(iface string) string {
	if len(e.scenario.moduleByJar) > 0 {
		return fmt.Sprintf("no provider is modeled for service %s in the %s scenario (module-path explicit modules' provides, automatic modules' and class-path jars' META-INF/services)", iface, e.scenario.Scenario)
	}
	return fmt.Sprintf("no classpath provider is modeled for service %s; JPMS providers are outside this model", iface)
}

// linkedLoadSites returns the load sites a consumption's receiver trace
// reaches that name iface.
func (e *ModelEvaluator) linkedLoadSites(c ServiceConsumptionSite, iface string) []ServiceLoadSite {
	var out []ServiceLoadSite
	for _, key := range c.streamLoadSites {
		if load, ok := e.sites[key]; ok && containsString(load.Candidates(), iface) {
			out = append(out, load)
		}
	}
	return out
}

// bootLayerOnly reports whether every load site linked to c is
// load(ModuleLayer.boot(), ...) — the only layer modeled exactly.
func (e *ModelEvaluator) bootLayerOnly(c ServiceConsumptionSite) bool {
	found := false
	for _, key := range c.streamLoadSites {
		load, ok := e.sites[key]
		if !ok {
			continue
		}
		if load.LoaderKind != LoaderModuleLayer || load.Layer != LayerBoot {
			return false
		}
		found = true
	}
	return found
}

// explicitLoaderRecognized reports whether every load site linked to c
// that passes an explicit ClassLoader passes a recognized one
// (LoaderIdentity), so the class-path registrations are the providers
// it sees under the single application class loader model.
func (e *ModelEvaluator) explicitLoaderRecognized(c ServiceConsumptionSite) bool {
	found := false
	for _, key := range c.streamLoadSites {
		load, ok := e.sites[key]
		if !ok || load.LoaderKind != LoaderExplicit {
			continue
		}
		if load.LoaderIdentity == "" {
			return false
		}
		found = true
	}
	return found
}

// visibleProviders narrows the providers for iface by the linked load
// sites' layer (SERVICE_LOADING S3): load(ModuleLayer.boot(), S) sees
// only providers in named modules on the module path, never the class
// path's unnamed module. Any other linked load site keeps the full set.
// reason explains an empty result caused by the layer.
func (e *ModelEvaluator) visibleProviders(c ServiceConsumptionSite, iface string) ([]ProviderMaterialization, string) {
	all := e.providersByInterface[iface]
	linked := e.linkedLoadSites(c, iface)
	if len(linked) == 0 {
		return all, ""
	}
	for _, load := range linked {
		if load.LoaderKind != LoaderModuleLayer || load.Layer != LayerBoot {
			return all, ""
		}
	}
	var named []ProviderMaterialization
	for _, m := range all {
		if m.Registration.Module != "" {
			named = append(named, m)
		}
	}
	if len(named) > 0 {
		return named, ""
	}
	if len(e.scenario.moduleByJar) == 0 {
		return nil, fmt.Sprintf("load(ModuleLayer.boot(), %s) in the %s scenario: the boot layer holds only JDK modules, so no provider from the input jars is visible (JDK providers are not analyzed)", iface, e.scenario.scenarioName())
	}
	return nil, fmt.Sprintf("load(ModuleLayer.boot(), %s): no provider in a named module on the module path; class-path providers are in no layer (JDK providers are not analyzed)", iface)
}

// usesAllowed applies the scenario's `uses` check to the load sites a
// consumption is linked to: the service is loadable if any linked load
// site for it passes. With no recorded load site the consumption's own
// caller is checked.
func (e *ModelEvaluator) usesAllowed(consumption ServiceConsumptionSite, iface string) (bool, string) {
	var reason string
	checked := false
	for _, key := range consumption.streamLoadSites {
		load, ok := e.sites[key]
		if !ok || !containsString(load.Candidates(), iface) {
			continue
		}
		checked = true
		ok, why := e.scenario.usesCheck(e.h, load.CallerFQCN, iface)
		if ok {
			return true, ""
		}
		reason = why
	}
	if !checked {
		return e.scenario.usesCheck(e.h, consumption.CallerFQCN, iface)
	}
	return false, reason
}

type siteKey struct {
	caller, desc string
	pc           int
}

func NewModelEvaluator(h *hierarchy.Hierarchy, sites []ServiceLoadSite, materializations []ProviderMaterialization, consumptionSets ...[]ServiceConsumptionSite) *ModelEvaluator {
	e := &ModelEvaluator{
		h: h, sites: make(map[siteKey]ServiceLoadSite), consumptions: make(map[siteKey]ServiceConsumptionSite), dispatches: make(map[siteKey]ServiceConsumptionSite),
		providersByInterface: make(map[string][]ProviderMaterialization),
	}
	for _, site := range sites {
		e.sites[siteKey{site.CallerFQCN, site.CallerName + site.CallerDesc, site.PC}] = site
	}
	for _, set := range consumptionSets {
		for _, site := range set {
			key := siteKey{site.CallerFQCN, site.CallerName + site.CallerDesc, site.PC}
			if site.ProviderMethodCall {
				e.dispatches[key] = site
			} else {
				e.consumptions[key] = site
			}
		}
	}
	for _, m := range materializations {
		e.providersByInterface[m.Registration.Interface] = append(e.providersByInterface[m.Registration.Interface], m)
	}
	return e
}

func (e *ModelEvaluator) Evaluate(model *models.Model, site models.CallSite) models.Fact {
	fact := models.Fact{ModelID: model.ID}
	if model.ID == serviceLoaderConsumptionModelID {
		return e.evaluateConsumption(model, site)
	}
	if model.ID == serviceLoaderDispatchModelID {
		return e.evaluateDispatch(model, site)
	}
	key := siteKey{site.Caller.Owner, site.Caller.Name + site.Caller.Desc, site.PC}
	resolved, ok := e.sites[key]
	if !ok {
		return models.Fact{Incomplete: true, Reason: "ServiceLoader call site has no matching resolved S1 inventory record"}
	}
	if len(resolved.Candidates()) == 0 || resolved.UnresolvedReason != "" {
		return models.Fact{Incomplete: true, Reason: resolved.UnresolvedReason}
	}
	// Creating a ServiceLoader is lazy: provider constructors are only
	// materialized at a recognized iterator step or Provider.get() site.
	// The exact source site remains in the inventory for provenance.
	return fact
}

func (e *ModelEvaluator) evaluateDispatch(model *models.Model, site models.CallSite) models.Fact {
	fact := models.Fact{ModelID: model.ID}
	key := siteKey{site.Caller.Owner, site.Caller.Name + site.Caller.Desc, site.PC}
	dispatch, ok := e.dispatches[key]
	if !ok {
		// A registered descriptor may also be called on values unrelated
		// to ServiceLoader. The dynamic model is relevant only at the exact
		// call sites whose receiver provenance was discovered.
		return fact
	}
	if dispatch.UnresolvedReason != "" {
		fact.Incomplete = true
		fact.Reason = dispatch.UnresolvedReason
	}
	for _, iface := range dispatch.Interfaces {
		// Receiver provenance may contain several candidates at a control-
		// flow join. A call through one exact service interface can only
		// dispatch to registrations for that interface, not the union's
		// unrelated services.
		if iface != dispatch.CallOwner {
			continue
		}
		mats, layerReason := e.visibleProviders(dispatch, iface)
		if len(mats) == 0 {
			fact.Incomplete = true
			if layerReason == "" {
				layerReason = e.noProviderReason(iface)
			}
			fact.Reason = joinReasons(fact.Reason, layerReason)
			continue
		}
		if ok, why := e.usesAllowed(dispatch, iface); !ok {
			fact.Incomplete = true
			fact.Reason = joinReasons(fact.Reason, why)
			continue
		}
		for _, m := range mats {
			if m.Gap != "" {
				fact.Incomplete = true
				fact.Reason = joinReasons(fact.Reason, m.Gap)
				continue
			}
			if m.FactoryDesc != "" {
				// provider() may return any implementation; its runtime
				// type is not tracked, so only the CHA edges remain.
				fact.Incomplete = true
				fact.Reason = joinReasons(fact.Reason, fmt.Sprintf("provider %s is created by its provider() method; the returned object's type is not tracked, so no exact dispatch edge is modeled", m.Registration.ProviderClass))
				continue
			}
			if _, knownInterface := e.h.Lookup(iface); !knownInterface {
				fact.Incomplete = true
				fact.Reason = joinReasons(fact.Reason, fmt.Sprintf("service interface %s is outside the analyzed hierarchy; provider assignability cannot be verified", iface))
				continue
			}
			if !containsString(e.h.Subtypes(iface), m.Registration.ProviderClass) {
				fact.Incomplete = true
				fact.Reason = joinReasons(fact.Reason, fmt.Sprintf("registered provider %s is not a known subtype of service %s", m.Registration.ProviderClass, iface))
				continue
			}
			owner, found := e.h.ResolveExact(m.Registration.ProviderClass, dispatch.CallName, dispatch.CallDesc)
			if !found {
				fact.Incomplete = true
				fact.Reason = joinReasons(fact.Reason, fmt.Sprintf("provider %s does not resolve exact method %s%s", m.Registration.ProviderClass, dispatch.CallName, dispatch.CallDesc))
				continue
			}
			fact.Targets = appendModelTarget(fact.Targets, models.Target{Method: models.Signature{Owner: owner, Name: dispatch.CallName, Desc: dispatch.CallDesc}, Kind: classfile.CallInterface})
		}
	}
	return fact
}

func (e *ModelEvaluator) evaluateConsumption(model *models.Model, site models.CallSite) models.Fact {
	fact := models.Fact{ModelID: model.ID}
	key := siteKey{site.Caller.Owner, site.Caller.Name + site.Caller.Desc, site.PC}
	consumption, ok := e.consumptions[key]
	if !ok {
		// Generic Iterator calls that have no trace to ServiceLoader are
		// outside this model and do not create noise or false gaps.
		if site.Call.Owner == "java.util.Iterator" {
			return fact
		}
		fact.Incomplete = true
		fact.Reason = consumptionSiteGap(site.Call)
		return fact
	}
	if consumption.UnresolvedReason != "" {
		fact.Incomplete = true
		fact.Reason = consumption.UnresolvedReason
	}
	if site.Call.Owner == serviceLoaderOwner && site.Call.Name == "stream" {
		if consumption.streamTerminalSeen && !consumption.streamTerminalSupported {
			fact.Incomplete = true
			fact.Reason = joinReasons(fact.Reason, "ServiceLoader.stream() terminal or callback flow is not modeled")
		}
		return fact
	}
	if site.Call.Owner == serviceLoaderOwner && site.Call.Name == "iterator" {
		return fact
	}
	// ServiceLoader's iterator checks for another provider in hasNext(),
	// but materializes the instance only when next() is consumed.
	if site.Call.Owner == "java.util.Iterator" && site.Call.Name == "hasNext" {
		return fact
	}
	if len(consumption.Interfaces) == 0 {
		if !fact.Incomplete {
			fact.Incomplete = true
			fact.Reason = consumptionSiteGap(site.Call)
		}
		return fact
	}
	bootOnly := e.bootLayerOnly(consumption)
	explicitKnown := e.explicitLoaderRecognized(consumption)
	for _, loader := range consumption.LoaderKinds {
		if loader == LoaderModuleLayer && bootOnly {
			continue // ModuleLayer.boot() is modeled exactly below
		}
		if loader == LoaderExplicit && explicitKnown {
			continue // context/system/defining loader: the modeled class path
		}
		if loader != LoaderClasspath {
			fact.Incomplete = true
			fact.Reason = joinReasons(fact.Reason, fmt.Sprintf("provider visibility for %s loader is not verified against classpath registrations", loader))
		}
	}
	for _, iface := range consumption.Interfaces {
		mats, layerReason := e.visibleProviders(consumption, iface)
		if len(mats) == 0 {
			fact.Incomplete = true
			if layerReason == "" {
				layerReason = e.noProviderReason(iface)
			}
			fact.Reason = joinReasons(fact.Reason, layerReason)
			continue
		}
		if ok, why := e.usesAllowed(consumption, iface); !ok {
			fact.Incomplete = true
			fact.Reason = joinReasons(fact.Reason, why)
			continue
		}
		for _, m := range mats {
			if m.Gap != "" {
				fact.Incomplete = true
				fact.Reason = joinReasons(fact.Reason, m.Gap)
				continue
			}
			owner := m.Registration.ProviderClass
			if m.FactoryDesc != "" {
				fact.Targets = appendModelTarget(fact.Targets, models.Target{Method: models.Signature{Owner: owner, Name: "provider", Desc: m.FactoryDesc}, Kind: classfile.CallStatic})
			} else {
				fact.Targets = appendModelTarget(fact.Targets, models.Target{Method: models.Signature{Owner: owner, Name: "<init>", Desc: "()V"}})
			}
			if ci, found := e.h.Lookup(owner); found {
				if _, hasInit := ci.MethodByNameAndDesc("<clinit>", "()V"); hasInit {
					fact.Targets = appendModelTarget(fact.Targets, models.Target{Method: models.Signature{Owner: owner, Name: "<clinit>", Desc: "()V"}, Kind: classfile.CallInit})
				}
			}
		}
	}
	return fact
}

func consumptionSiteGap(call models.Signature) string {
	switch {
	case call.Owner == serviceLoaderOwner && call.Name == "iterator":
		return "ServiceLoader.iterator() receiver is not traceably linked to a resolved load site"
	case call.Owner == serviceLoaderOwner && call.Name == "stream":
		return "ServiceLoader.stream() receiver is not traceably linked to a resolved load site"
	case call.Owner == "java.util.ServiceLoader$Provider" && call.Name == "get":
		return "ServiceLoader.Provider.get() receiver is not traceably linked to a specific classpath registration"
	default:
		return "iterator receiver is not traceably linked to a ServiceLoader load site"
	}
}

func appendModelTarget(targets []models.Target, target models.Target) []models.Target {
	for _, existing := range targets {
		if existing == target {
			return targets
		}
	}
	return append(targets, target)
}

func joinReasons(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" || strings.Contains(a, b) {
		return a
	}
	return a + "; " + b
}
