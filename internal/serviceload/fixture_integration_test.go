package serviceload

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/analysis"
	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/models"
)

// TestCompiledServiceLoaderFixture exercises the real classfile path from
// hierarchy extraction through load/consumption discovery and model edges.
// The checked-in .class files keep the regular Go test suite independent of
// a Java compiler.
func TestCompiledServiceLoaderFixture(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "testdata", "serviceloader", "classes")
	prepared, err := input.PrepareAppInput(fixtureDir)
	if err != nil {
		t.Fatalf("PrepareAppInput: %v", err)
	}
	defer prepared.Cleanup()

	h, hierarchyWarnings, err := hierarchy.Build(prepared.AppClassesJar, nil)
	if err != nil {
		t.Fatalf("hierarchy.Build: %v", err)
	}
	if len(hierarchyWarnings) != 0 {
		t.Fatalf("hierarchy warnings: %v", hierarchyWarnings)
	}

	loads, loadWarnings, err := FindLoadSites(h)
	if err != nil {
		t.Fatalf("FindLoadSites: %v", err)
	}
	if len(loadWarnings) != 0 {
		t.Fatalf("load warnings: %v", loadWarnings)
	}
	if len(loads) != 8 {
		t.Fatalf("load sites = %+v, want factory/helper, field, parameter, stream, and constructor load sites", loads)
	}
	var helperLoad, unusedLoad, fieldLoad, instanceLoad, parameterLoad, streamLoad, lazyStreamLoad, callbackLoad *ServiceLoadSite
	for i := range loads {
		load := &loads[i]
		if load.Interface != "fixture.service.Decoder" {
			t.Errorf("load site %s resolved interface %q, want fixture.service.Decoder", load.CallerMethod, load.Interface)
		}
		switch load.CallerName {
		case "createLoader":
			helperLoad = load
		case "loadWithoutConsuming":
			unusedLoad = load
		case "<clinit>":
			fieldLoad = load
		case "<init>":
			instanceLoad = load
		case "runThroughParameter":
			parameterLoad = load
		case "runThroughStream":
			streamLoad = load
		case "streamWithoutConsuming":
			lazyStreamLoad = load
		case "runThroughStreamCallback":
			callbackLoad = load
		}
	}
	if helperLoad == nil || unusedLoad == nil || fieldLoad == nil || instanceLoad == nil || parameterLoad == nil || streamLoad == nil || lazyStreamLoad == nil || callbackLoad == nil {
		t.Fatalf("load sites missing expected methods: %+v", loads)
	}

	registrations, inventoryWarnings, err := Inventory(h, prepared.AppClassesJar, nil)
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(inventoryWarnings) != 0 || len(registrations) != 1 || !registrations[0].ProviderFound {
		t.Fatalf("registrations=%+v warnings=%v, want one found provider", registrations, inventoryWarnings)
	}
	materializations := FindMaterializations(h, loads, registrations)
	if len(materializations) != 1 || materializations[0].Gap != "" {
		t.Fatalf("materializations=%+v, want one constructible provider", materializations)
	}

	consumptions, consumptionWarnings, err := FindConsumptionSites(h, loads)
	if err != nil {
		t.Fatalf("FindConsumptionSites: %v", err)
	}
	if len(consumptionWarnings) != 0 {
		t.Fatalf("consumption warnings: %v", consumptionWarnings)
	}
	var nextSite, decodeSite, fieldNextSite, fieldDecodeSite *ServiceConsumptionSite
	var instanceNextSite, instanceDecodeSite, cycleIterator *ServiceConsumptionSite
	var parameterNextSite, parameterDecodeSite, unknownParameterIterator *ServiceConsumptionSite
	var streamProviderGet, streamDecode, lazyStreamAnchor, callbackProviderGet, callbackDecode *ServiceConsumptionSite
	for i := range consumptions {
		site := &consumptions[i]
		switch {
		case site.CallerName == "run" && site.CallOwner == "java.util.Iterator" && site.CallName == "next":
			nextSite = site
		case site.CallerName == "run" && site.CallOwner == "fixture.service.Decoder" && site.CallName == "decode":
			decodeSite = site
		case site.CallerName == "runThroughStaticField" && site.CallOwner == "java.util.Iterator" && site.CallName == "next":
			fieldNextSite = site
		case site.CallerName == "runThroughStaticField" && site.CallOwner == "fixture.service.Decoder" && site.CallName == "decode":
			fieldDecodeSite = site
		case site.CallerName == "runThroughFinalInstanceField" && site.CallOwner == "java.util.Iterator" && site.CallName == "next":
			instanceNextSite = site
		case site.CallerName == "runThroughFinalInstanceField" && site.CallOwner == "fixture.service.Decoder" && site.CallName == "decode":
			instanceDecodeSite = site
		case site.CallerName == "runThroughCyclicField" && site.CallOwner == serviceLoaderOwner && site.CallName == "iterator":
			cycleIterator = site
		case site.CallerName == "runThroughParameter" && site.CallOwner == "java.util.Iterator" && site.CallName == "next":
			parameterNextSite = site
		case site.CallerName == "runThroughParameter" && site.CallOwner == "fixture.service.Decoder" && site.CallName == "decode":
			parameterDecodeSite = site
		case site.CallerName == "runWithUnknownParameter" && site.CallOwner == serviceLoaderOwner && site.CallName == "iterator":
			unknownParameterIterator = site
		case site.CallerName == "runThroughStream" && site.CallOwner == "java.util.ServiceLoader$Provider" && site.CallName == "get":
			streamProviderGet = site
		case site.CallerName == "runThroughStream" && site.CallOwner == "fixture.service.Decoder" && site.CallName == "decode":
			streamDecode = site
		case site.CallerName == "streamWithoutConsuming" && site.CallOwner == serviceLoaderOwner && site.CallName == "stream":
			lazyStreamAnchor = site
		case strings.HasPrefix(site.CallerName, "lambda$runThroughStreamCallback") && site.CallOwner == "java.util.ServiceLoader$Provider" && site.CallName == "get":
			callbackProviderGet = site
		case site.CallerName == "runThroughStreamCallback" && site.CallOwner == "fixture.service.Decoder" && site.CallName == "decode":
			callbackDecode = site
		}
	}
	if nextSite == nil || decodeSite == nil {
		t.Fatalf("consumption sites = %+v, want Iterator.next and Decoder.decode in run()", consumptions)
	}
	if decodeSite.ProviderMethodCall != true || len(decodeSite.Interfaces) != 1 || decodeSite.Interfaces[0] != "fixture.service.Decoder" {
		t.Fatalf("decode site = %+v, want linked provider method on exact service interface", decodeSite)
	}
	if fieldNextSite == nil || fieldDecodeSite == nil || fieldDecodeSite.ProviderMethodCall != true || len(fieldDecodeSite.Interfaces) != 1 || fieldDecodeSite.Interfaces[0] != "fixture.service.Decoder" {
		t.Fatalf("static-field consumptions = next:%+v decode:%+v, want resolved ServiceLoader provenance", fieldNextSite, fieldDecodeSite)
	}
	if instanceNextSite == nil || instanceDecodeSite == nil || instanceDecodeSite.ProviderMethodCall != true || len(instanceDecodeSite.Interfaces) != 1 || instanceDecodeSite.Interfaces[0] != "fixture.service.Decoder" {
		t.Fatalf("final-instance-field consumptions = next:%+v decode:%+v, want resolved ServiceLoader provenance", instanceNextSite, instanceDecodeSite)
	}
	if cycleIterator == nil || cycleIterator.UnresolvedReason == "" || len(cycleIterator.Interfaces) != 0 {
		t.Fatalf("cyclic-field consumption = %+v, want an explicit unresolved cycle", cycleIterator)
	}
	if parameterNextSite == nil || parameterDecodeSite == nil || parameterDecodeSite.ProviderMethodCall != true || len(parameterDecodeSite.Interfaces) != 1 || parameterDecodeSite.Interfaces[0] != "fixture.service.Decoder" {
		t.Fatalf("parameter-helper consumptions = next:%+v decode:%+v, want resolved ServiceLoader provenance", parameterNextSite, parameterDecodeSite)
	}
	if unknownParameterIterator == nil || unknownParameterIterator.UnresolvedReason == "" || len(unknownParameterIterator.Interfaces) != 0 {
		t.Fatalf("unknown-parameter consumption = %+v, want an explicit gap", unknownParameterIterator)
	}
	if streamProviderGet == nil || streamProviderGet.UnresolvedReason != "" || len(streamProviderGet.Interfaces) != 1 || streamProviderGet.Interfaces[0] != "fixture.service.Decoder" || streamDecode == nil || !streamDecode.ProviderMethodCall {
		t.Fatalf("linear stream provenance = Provider.get:%+v decode:%+v, want linked service/provider", streamProviderGet, streamDecode)
	}
	if lazyStreamAnchor == nil || lazyStreamAnchor.UnresolvedReason != "" || len(lazyStreamAnchor.Interfaces) != 1 || lazyStreamAnchor.Interfaces[0] != "fixture.service.Decoder" {
		t.Fatalf("lazy stream site = %+v, want recognized loader stream without materialization", lazyStreamAnchor)
	}
	if callbackProviderGet == nil || callbackProviderGet.UnresolvedReason == "" || len(callbackProviderGet.Interfaces) != 0 || callbackDecode != nil {
		t.Fatalf("callback stream sites = Provider.get:%+v decode:%+v, want explicit unresolved Provider.get and no fabricated linked dispatch", callbackProviderGet, callbackDecode)
	}
	var unknownIterator *ServiceConsumptionSite
	var mutableFieldIterator *ServiceConsumptionSite
	for _, site := range consumptions {
		if site.CallerName == "loadWithoutConsuming" {
			t.Fatalf("loadWithoutConsuming unexpectedly has a consumption: %+v", site)
		}
		if site.CallerName == "runWithUnknownLoader" && site.CallOwner == serviceLoaderOwner && site.CallName == "iterator" {
			unknownIterator = &site
		}
		if site.CallerName == "runThroughMutableField" && site.CallOwner == serviceLoaderOwner && site.CallName == "iterator" {
			mutableFieldIterator = &site
		}
	}
	if unknownIterator == nil || unknownIterator.UnresolvedReason == "" || len(unknownIterator.Interfaces) != 0 {
		t.Fatalf("unknown helper consumption = %+v, want an explicit unresolved gap", unknownIterator)
	}
	if mutableFieldIterator == nil || mutableFieldIterator.UnresolvedReason == "" || len(mutableFieldIterator.Interfaces) != 0 {
		t.Fatalf("mutable field consumption = %+v, want an explicit unresolved gap", mutableFieldIterator)
	}

	g := callgraph.Build(h, nil)
	root, ok := g.ID("fixture.app.Entry", "run", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing fixture entry method in graph")
	}
	registry := models.NewRegistry()
	RegisterModel(registry)
	RegisterDispatchModels(registry, consumptions)
	result, err := analysis.Run(h, g, registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{root}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run: %v", err)
	}
	if result.Incomplete {
		t.Fatalf("analysis result has gaps: %+v", result)
	}

	providerCtor, ok := g.ID("fixture.provider.DecoderProvider", "<init>", "()V")
	if !ok || !hasModelEdgeAt(g, root, providerCtor, nextSite.PC) {
		t.Fatalf("no provider-constructor model edge at Iterator.next pc %d", nextSite.PC)
	}
	providerDecode, ok := g.ID("fixture.provider.DecoderProvider", "decode", "()Ljava/lang/String;")
	if !ok || !hasModelEdgeAt(g, root, providerDecode, decodeSite.PC) {
		t.Fatalf("no descriptor-exact provider decode model edge at pc %d", decodeSite.PC)
	}
	fieldRoot, ok := g.ID("fixture.app.Entry", "runThroughStaticField", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing static-field fixture entry method in graph")
	}
	fieldGraph := callgraph.Build(h, nil)
	fieldAnalysis, err := analysis.Run(h, fieldGraph, registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{fieldRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run static field: %v", err)
	}
	if fieldAnalysis.Incomplete || !hasModelEdgeAt(fieldGraph, fieldRoot, providerCtor, fieldNextSite.PC) || !hasModelEdgeAt(fieldGraph, fieldRoot, providerDecode, fieldDecodeSite.PC) {
		t.Fatalf("static-field analysis = %+v; want complete provider constructor and dispatch edges", fieldAnalysis)
	}
	instanceRoot, ok := g.ID("fixture.app.Entry", "runThroughFinalInstanceField", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing final-instance-field fixture method in graph")
	}
	instanceGraph := callgraph.Build(h, nil)
	instanceAnalysis, err := analysis.Run(h, instanceGraph, registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{instanceRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run final instance field: %v", err)
	}
	if instanceAnalysis.Incomplete || !hasModelEdgeAt(instanceGraph, instanceRoot, providerCtor, instanceNextSite.PC) || !hasModelEdgeAt(instanceGraph, instanceRoot, providerDecode, instanceDecodeSite.PC) {
		t.Fatalf("final-instance-field analysis = %+v; want complete provider constructor and dispatch edges", instanceAnalysis)
	}
	parameterRoot, ok := g.ID("fixture.app.Entry", "runThroughParameter", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing parameter-helper fixture method in graph")
	}
	parameterGraph := callgraph.Build(h, nil)
	parameterAnalysis, err := analysis.Run(h, parameterGraph, registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{parameterRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run parameter helper: %v", err)
	}
	if parameterAnalysis.Incomplete || !hasModelEdgeAt(parameterGraph, parameterRoot, providerCtor, parameterNextSite.PC) || !hasModelEdgeAt(parameterGraph, parameterRoot, providerDecode, parameterDecodeSite.PC) {
		t.Fatalf("parameter-helper analysis = %+v; want provider constructor and dispatch edges", parameterAnalysis)
	}
	streamRoot, ok := g.ID("fixture.app.Entry", "runThroughStream", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing linear-stream fixture method in graph")
	}
	streamGraph := callgraph.Build(h, nil)
	streamAnalysis, err := analysis.Run(h, streamGraph, registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{streamRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run linear stream: %v", err)
	}
	if streamAnalysis.Incomplete || !hasModelEdgeAt(streamGraph, streamRoot, providerCtor, streamProviderGet.PC) || !hasModelEdgeAt(streamGraph, streamRoot, providerDecode, streamDecode.PC) {
		t.Fatalf("linear-stream analysis = %+v; want provider constructor and dispatch edges", streamAnalysis)
	}
	lazyStreamRoot, ok := g.ID("fixture.app.Entry", "streamWithoutConsuming", "()Ljava/util/stream/Stream;")
	if !ok {
		t.Fatal("missing lazy-stream fixture method in graph")
	}
	lazyStreamGraph := callgraph.Build(h, nil)
	lazyStreamAnalysis, err := analysis.Run(h, lazyStreamGraph, registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{lazyStreamRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run lazy stream: %v", err)
	}
	if lazyStreamAnalysis.Incomplete || lazyStreamAnalysis.AddedEdges != 0 {
		t.Fatalf("lazy-stream analysis = %+v, want no materialization edges and no gap", lazyStreamAnalysis)
	}
	callbackRoot, ok := g.ID("fixture.app.Entry", "runThroughStreamCallback", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing callback-stream fixture method in graph")
	}
	callbackGraph := callgraph.Build(h, nil)
	callbackAnalysis, err := analysis.Run(h, callbackGraph, registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{callbackRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run callback stream: %v", err)
	}
	if !callbackAnalysis.Incomplete || callbackAnalysis.AddedEdges != 0 {
		t.Fatalf("callback-stream analysis = %+v, want explicit incomplete gap and no provider edges", callbackAnalysis)
	}
	cycleRoot, ok := g.ID("fixture.app.Entry", "runThroughCyclicField", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing cyclic-field fixture method in graph")
	}
	cycleGraph := callgraph.Build(h, nil)
	cycleAnalysis, err := analysis.Run(h, cycleGraph, registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{cycleRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run cyclic field: %v", err)
	}
	if !cycleAnalysis.Incomplete || cycleAnalysis.AddedEdges != 0 {
		t.Fatalf("cyclic-field analysis = %+v, want explicit gap and no modeled edges", cycleAnalysis)
	}

	unusedCall := models.CallSite{
		Caller: models.Signature{Owner: unusedLoad.CallerFQCN, Name: unusedLoad.CallerName, Desc: unusedLoad.CallerDesc},
		Call:   models.Signature{Owner: serviceLoaderOwner, Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"},
		PC:     unusedLoad.PC,
	}
	loadModel := registry.Lookup(unusedCall.Call, nil)
	if len(loadModel) != 1 {
		t.Fatalf("load model lookup returned %d models, want one", len(loadModel))
	}
	unusedFact := NewModelEvaluator(h, loads, materializations, consumptions).Evaluate(loadModel[0], unusedCall)
	if unusedFact.Incomplete || len(unusedFact.Targets) != 0 {
		t.Fatalf("unused load fact = %+v, want no target and no gap", unusedFact)
	}

	// Replace the service declaration with a provider absent from the class
	// hierarchy. Inventory keeps it explicit; materialization reports a gap
	// and never fabricates a constructor target.
	missingFixture := t.TempDir()
	if err := copyFixtureTree(fixtureDir, missingFixture); err != nil {
		t.Fatalf("copy fixture for missing-provider case: %v", err)
	}
	serviceFile := filepath.Join(missingFixture, "META-INF", "services", "fixture.service.Decoder")
	if err := os.WriteFile(serviceFile, []byte("fixture.provider.MissingDecoder\n"), 0o600); err != nil {
		t.Fatalf("write missing-provider declaration: %v", err)
	}
	missingApp, err := input.PrepareAppInput(missingFixture)
	if err != nil {
		t.Fatalf("PrepareAppInput missing-provider fixture: %v", err)
	}
	defer missingApp.Cleanup()
	missingHierarchy, _, err := hierarchy.Build(missingApp.AppClassesJar, nil)
	if err != nil {
		t.Fatalf("hierarchy.Build missing-provider fixture: %v", err)
	}
	missingLoads, _, err := FindLoadSites(missingHierarchy)
	if err != nil {
		t.Fatalf("FindLoadSites missing-provider fixture: %v", err)
	}
	missingRegistrations, missingWarnings, err := Inventory(missingHierarchy, missingApp.AppClassesJar, nil)
	if err != nil {
		t.Fatalf("Inventory missing-provider fixture: %v", err)
	}
	if len(missingRegistrations) != 1 || missingRegistrations[0].ProviderFound || len(missingWarnings) == 0 {
		t.Fatalf("missing registrations=%+v warnings=%v, want one explicit unresolved provider", missingRegistrations, missingWarnings)
	}
	missingMats := FindMaterializations(missingHierarchy, missingLoads, missingRegistrations)
	if len(missingMats) != 0 {
		t.Fatalf("missing materializations=%+v, want no fabricated target", missingMats)
	}
	missingConsumptions, _, err := FindConsumptionSites(missingHierarchy, missingLoads)
	if err != nil {
		t.Fatalf("FindConsumptionSites missing-provider fixture: %v", err)
	}
	var missingDecode *ServiceConsumptionSite
	for i := range missingConsumptions {
		if missingConsumptions[i].CallOwner == "fixture.service.Decoder" && missingConsumptions[i].CallName == "decode" {
			missingDecode = &missingConsumptions[i]
		}
	}
	if missingDecode == nil {
		t.Fatalf("missing-provider fixture has no decode site: %+v", missingConsumptions)
	}
	missingFact := NewModelEvaluator(missingHierarchy, missingLoads, missingMats, missingConsumptions).Evaluate(
		&models.Model{ID: serviceLoaderDispatchModelID},
		models.CallSite{
			Caller: models.Signature{Owner: missingDecode.CallerFQCN, Name: missingDecode.CallerName, Desc: missingDecode.CallerDesc},
			Call:   models.Signature{Owner: missingDecode.CallOwner, Name: missingDecode.CallName, Desc: missingDecode.CallDesc},
			PC:     missingDecode.PC,
		},
	)
	if !missingFact.Incomplete || len(missingFact.Targets) != 0 || !strings.Contains(missingFact.Reason, "no classpath provider") {
		t.Fatalf("missing provider fact = %+v, want explicit gap and no targets", missingFact)
	}

	unknownRoot, ok := g.ID("fixture.app.Entry", "runWithUnknownLoader", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing unknown-loader fixture method in graph")
	}
	unknownAnalysis, err := analysis.Run(h, callgraph.Build(h, nil), registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{unknownRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run unknown loader: %v", err)
	}
	if !unknownAnalysis.Incomplete || unknownAnalysis.AddedEdges != 0 {
		t.Fatalf("unknown helper analysis = %+v, want an explicit gap and no modeled edges", unknownAnalysis)
	}
	mutableRoot, ok := g.ID("fixture.app.Entry", "runThroughMutableField", "()Ljava/lang/String;")
	if !ok {
		t.Fatal("missing mutable-field fixture method in graph")
	}
	mutableAnalysis, err := analysis.Run(h, callgraph.Build(h, nil), registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{mutableRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run mutable field: %v", err)
	}
	if !mutableAnalysis.Incomplete || mutableAnalysis.AddedEdges != 0 {
		t.Fatalf("mutable-field analysis = %+v, want an explicit gap and no modeled edges", mutableAnalysis)
	}
	unknownParameterRoot, ok := g.ID("fixture.app.Entry", "runWithUnknownParameter", "(Ljava/util/ServiceLoader;)Ljava/lang/String;")
	if !ok {
		t.Fatal("missing unknown-parameter fixture method in graph")
	}
	unknownParameterAnalysis, err := analysis.Run(h, callgraph.Build(h, nil), registry, NewModelEvaluator(h, loads, materializations, consumptions), []uint32{unknownParameterRoot}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatalf("analysis.Run unknown parameter: %v", err)
	}
	if !unknownParameterAnalysis.Incomplete || unknownParameterAnalysis.AddedEdges != 0 {
		t.Fatalf("unknown-parameter analysis = %+v, want explicit gap and no modeled edges", unknownParameterAnalysis)
	}
}

func hasModelEdgeAt(g *callgraph.Graph, from, to uint32, pc int) bool {
	for _, edge := range g.Edges(from) {
		if edge.To != to || edge.Basis != callgraph.BasisModel {
			continue
		}
		for _, site := range edge.Sites {
			if site.PC == pc {
				return true
			}
		}
	}
	return false
}

func copyFixtureTree(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}
