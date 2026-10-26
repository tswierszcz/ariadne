package serviceload

import (
	"strings"
	"testing"

	"ariadne/internal/analysis"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/cves"
	"ariadne/internal/hierarchy"
	"ariadne/internal/models"
	"ariadne/internal/reach"
	"ariadne/internal/vulnmap"
)

func TestModelEvaluator_doesNotEagerlyMaterializeAtLoadSite(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{{
		FQCN:    "provider.One",
		Methods: []hierarchy.Method{{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}}, {Method: classfile.Method{Name: "<clinit>", Descriptor: "()V"}}},
	}})
	site := ServiceLoadSite{Interface: "api.Service", CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 12, LoaderKind: LoaderClasspath}
	mat := ProviderMaterialization{Registration: ServiceRegistration{Interface: "api.Service", ProviderClass: "provider.One", ProviderFound: true}, ConstructorMethod: "provider.One#<init>()V"}
	e := NewModelEvaluator(h, []ServiceLoadSite{site}, []ProviderMaterialization{mat})
	registry := models.NewRegistry()
	RegisterModel(registry)
	model := registry.Lookup(models.Signature{Owner: "java.util.ServiceLoader", Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"}, nil)[0]
	fact := e.Evaluate(model, models.CallSite{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, Call: models.Signature{Owner: "java.util.ServiceLoader", Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"}, PC: 12})
	if fact.Incomplete || len(fact.Targets) != 0 {
		t.Fatalf("fact=%+v, load() must not eagerly materialize a lazy provider", fact)
	}
	wrongPC := e.Evaluate(model, models.CallSite{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, PC: 13})
	if !wrongPC.Incomplete || len(wrongPC.Targets) != 0 {
		t.Fatalf("wrong-site fact=%+v, want explicit gap and no targets", wrongPC)
	}
}

func TestModelEvaluator_materializesAtLinkedIteratorStepAndProviderGet(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{{
		FQCN:    "provider.One",
		Methods: []hierarchy.Method{{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}}, {Method: classfile.Method{Name: "<clinit>", Descriptor: "()V"}}},
	}})
	load := ServiceLoadSite{Interface: "api.Service", CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 12, LoaderKind: LoaderClasspath}
	mat := ProviderMaterialization{Registration: ServiceRegistration{Interface: "api.Service", ProviderClass: "provider.One", ProviderFound: true}, ConstructorMethod: "provider.One#<init>()V"}
	consumptions := []ServiceConsumptionSite{
		{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 15, CallOwner: "java.util.Iterator", CallName: "hasNext", CallDesc: "()Z", Interfaces: []string{"api.Service"}, LoaderKinds: []LoaderKind{LoaderClasspath}},
		{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 20, CallOwner: "java.util.Iterator", CallName: "next", CallDesc: "()Ljava/lang/Object;", Interfaces: []string{"api.Service"}, LoaderKinds: []LoaderKind{LoaderClasspath}},
		{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 30, CallOwner: "java.util.ServiceLoader$Provider", CallName: "get", CallDesc: "()Ljava/lang/Object;", Interfaces: []string{"api.Service"}, LoaderKinds: []LoaderKind{LoaderClasspath}},
	}
	e := NewModelEvaluator(h, []ServiceLoadSite{load}, []ProviderMaterialization{mat}, consumptions)
	registry := models.NewRegistry()
	RegisterModel(registry)
	hasNextCall := models.CallSite{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, Call: models.Signature{Owner: "java.util.Iterator", Name: "hasNext", Desc: "()Z"}, PC: 15}
	hasNextModel := registry.Lookup(hasNextCall.Call, nil)
	if len(hasNextModel) != 1 {
		t.Fatalf("Lookup(hasNext) = %d models, want one", len(hasNextModel))
	}
	if fact := e.Evaluate(hasNextModel[0], hasNextCall); fact.Incomplete || len(fact.Targets) != 0 {
		t.Fatalf("Evaluate(hasNext) = %+v, want recognized non-materializing check", fact)
	}
	for _, call := range []models.CallSite{
		{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, Call: models.Signature{Owner: "java.util.Iterator", Name: "next", Desc: "()Ljava/lang/Object;"}, PC: 20},
		{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, Call: models.Signature{Owner: "java.util.ServiceLoader$Provider", Name: "get", Desc: "()Ljava/lang/Object;"}, PC: 30},
	} {
		registered := registry.Lookup(call.Call, nil)
		if len(registered) != 1 {
			t.Fatalf("Lookup(%+v) = %d models, want one", call.Call, len(registered))
		}
		fact := e.Evaluate(registered[0], call)
		if fact.Incomplete || len(fact.Targets) != 2 {
			t.Fatalf("Evaluate(%+v) = %+v, want provider constructor and class initializer", call, fact)
		}
		if fact.Targets[0].Method != (models.Signature{Owner: "provider.One", Name: "<init>", Desc: "()V"}) || fact.Targets[1].Kind != classfile.CallInit {
			t.Fatalf("targets=%+v, want provider constructor and init", fact.Targets)
		}
	}
}

func TestS2_linkedNextReachesProviderConstructorAndCHAServiceMethod(t *testing.T) {
	entry := &hierarchy.ClassInfo{
		FQCN: "app.Entry", IsApp: true,
		Methods: []hierarchy.Method{{
			Method: classfile.Method{Name: "run", Descriptor: "()V", AccessFlags: classfile.AccPublic | classfile.AccStatic},
			Calls: []classfile.Call{
				{Kind: classfile.CallStatic, Owner: serviceLoaderOwner, Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;", PC: 12},
				{Kind: classfile.CallVirtual, Owner: serviceLoaderOwner, Name: "iterator", Desc: "()Ljava/util/Iterator;", PC: 15},
				{Kind: classfile.CallInterface, Owner: "java.util.Iterator", Name: "next", Desc: "()Ljava/lang/Object;", PC: 20},
				{Kind: classfile.CallInterface, Owner: "api.Service", Name: "decode", Desc: "()V", PC: 25},
			},
		}},
	}
	service := &hierarchy.ClassInfo{FQCN: "api.Service", AccessFlags: classfile.AccPublic | classfile.AccInterface | classfile.AccAbstract, Methods: []hierarchy.Method{{Method: classfile.Method{Name: "decode", Descriptor: "()V", AccessFlags: classfile.AccPublic | classfile.AccAbstract}}}}
	provider := &hierarchy.ClassInfo{
		FQCN: "provider.One", SuperClass: "java.lang.Object", Interfaces: []string{"api.Service"},
		Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "<init>", Descriptor: "()V", AccessFlags: classfile.AccPublic}},
			{Method: classfile.Method{Name: "<clinit>", Descriptor: "()V", AccessFlags: classfile.AccStatic}},
			{Method: classfile.Method{Name: "decode", Descriptor: "()V", AccessFlags: classfile.AccPublic}},
			{Method: classfile.Method{Name: "decode", Descriptor: "(Ljava/lang/String;)V", AccessFlags: classfile.AccPublic}},
		},
	}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{entry, service, provider})
	g := callgraph.Build(h, nil)
	root, ok := g.ID("app.Entry", "run", "()V")
	if !ok {
		t.Fatal("missing app entry graph node")
	}
	load := ServiceLoadSite{Interface: "api.Service", CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 12, LoaderKind: LoaderClasspath}
	mat := ProviderMaterialization{Registration: ServiceRegistration{Interface: "api.Service", ProviderClass: "provider.One", ProviderFound: true}, ConstructorMethod: "provider.One#<init>()V"}
	consumptions := []ServiceConsumptionSite{
		{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 15, CallOwner: serviceLoaderOwner, CallName: "iterator", CallDesc: "()Ljava/util/Iterator;", Interfaces: []string{"api.Service"}, LoaderKinds: []LoaderKind{LoaderClasspath}},
		{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 20, CallOwner: "java.util.Iterator", CallName: "next", CallDesc: "()Ljava/lang/Object;", Interfaces: []string{"api.Service"}, LoaderKinds: []LoaderKind{LoaderClasspath}},
		{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 25, CallOwner: "api.Service", CallName: "decode", CallDesc: "()V", ProviderMethodCall: true, Interfaces: []string{"api.Service"}, LoaderKinds: []LoaderKind{LoaderClasspath}},
	}
	registry := models.NewRegistry()
	RegisterModel(registry)
	RegisterDispatchModels(registry, consumptions)
	modelResult, err := analysis.Run(h, g, registry, NewModelEvaluator(h, []ServiceLoadSite{load}, []ProviderMaterialization{mat}, consumptions), []uint32{root}, analysis.Options{MaxIterations: 3})
	if err != nil {
		t.Fatalf("analysis.Run: %v", err)
	}
	if modelResult.Incomplete || modelResult.AddedEdges != 3 {
		t.Fatalf("model result=%+v, want constructor, clinit and exact provider method edges", modelResult)
	}
	ctor, _ := g.ID("provider.One", "<init>", "()V")
	ctorEdge, ok := g.EdgeBetween(root, ctor)
	if !ok || ctorEdge.Basis != callgraph.BasisModel || len(ctorEdge.Sites) != 1 || ctorEdge.Sites[0].PC != 20 {
		t.Fatalf("constructor edge=%+v, found=%t; want model edge at Iterator.next pc 20", ctorEdge, ok)
	}
	decode, _ := g.ID("provider.One", "decode", "()V")
	var exactDispatch bool
	for _, edge := range g.Edges(root) {
		if edge.To != decode || edge.Basis != callgraph.BasisModel {
			continue
		}
		for _, callSite := range edge.Sites {
			if callSite.PC == 25 {
				exactDispatch = true
			}
		}
	}
	if !exactDispatch {
		t.Fatal("registered provider's exact decode()V model edge missing at service invocation pc 25")
	}
	wrongOverload, _ := g.ID("provider.One", "decode", "(Ljava/lang/String;)V")
	for _, edge := range g.Edges(root) {
		if edge.To == wrongOverload && edge.Basis == callgraph.BasisModel {
			t.Fatal("model added a descriptor-mismatched provider overload")
		}
	}
	engine := reach.NewWithRoots(g, h, nil, []uint32{root})
	cve := cves.CVE{ID: "CVE-S2-TEST"}
	results := engine.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: {Methods: []string{"provider.One#decode"}, Source: vulnmap.SourceManual}})
	if len(results) != 1 || results[0].Status != reach.StatusReachable {
		t.Fatalf("provider method result=%+v, want reachable through existing CHA dispatch", results)
	}
}

func TestModelEvaluator_marksUnresolvedServiceAndProviderGaps(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	site := ServiceLoadSite{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 5, UnresolvedReason: "class argument was unknown"}
	e := NewModelEvaluator(h, []ServiceLoadSite{site}, nil)
	fact := e.Evaluate(&models.Model{ID: serviceLoaderModelID}, models.CallSite{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, PC: 5})
	if !fact.Incomplete || fact.Reason != "class argument was unknown" {
		t.Fatalf("fact=%+v, want unresolved argument reason", fact)
	}

	resolved := ServiceLoadSite{Interface: "api.Service", CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 6}
	e = NewModelEvaluator(h, []ServiceLoadSite{resolved}, nil)
	fact = e.Evaluate(&models.Model{ID: serviceLoaderModelID}, models.CallSite{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, PC: 6})
	if fact.Incomplete || len(fact.Targets) != 0 {
		t.Fatalf("fact=%+v, a lazy load without provider consumption should not report materialization", fact)
	}
	consumption := ServiceConsumptionSite{CallerFQCN: "app.Entry", CallerName: "run", CallerDesc: "()V", PC: 7, CallOwner: "java.util.Iterator", CallName: "next", CallDesc: "()Ljava/lang/Object;", Interfaces: []string{"api.Service"}}
	e = NewModelEvaluator(h, []ServiceLoadSite{resolved}, nil, []ServiceConsumptionSite{consumption})
	fact = e.Evaluate(&models.Model{ID: serviceLoaderConsumptionModelID}, models.CallSite{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, Call: models.Signature{Owner: consumption.CallOwner, Name: consumption.CallName, Desc: consumption.CallDesc}, PC: consumption.PC})
	if !fact.Incomplete || !strings.Contains(fact.Reason, "no classpath provider") {
		t.Fatalf("fact=%+v, want missing provider gap at actual consumption", fact)
	}
}

func TestModelEvaluator_dispatchIgnoresSameSignatureWithoutProvenance(t *testing.T) {
	e := NewModelEvaluator(hierarchy.NewForTesting(nil), nil, nil)
	fact := e.Evaluate(&models.Model{ID: serviceLoaderDispatchModelID}, models.CallSite{
		Caller: models.Signature{Owner: "app.Other", Name: "run", Desc: "()V"},
		Call:   models.Signature{Owner: "api.Service", Name: "decode", Desc: "()V"},
		PC:     18,
	})
	if fact.Incomplete || len(fact.Targets) != 0 {
		t.Fatalf("unlinked fact=%+v, want irrelevant call ignored without a false gap", fact)
	}
}

func TestModelEvaluator_reportsReachableConsumptionGapsWithoutGuessingTargets(t *testing.T) {
	registry := models.NewRegistry()
	RegisterModel(registry)
	calls := []models.Signature{
		{Owner: serviceLoaderOwner, Name: "iterator", Desc: "()Ljava/util/Iterator;"},
		{Owner: serviceLoaderOwner, Name: "stream", Desc: "()Ljava/util/stream/Stream;"},
		{Owner: "java.util.ServiceLoader$Provider", Name: "get", Desc: "()Ljava/lang/Object;"},
	}
	wantFragments := []string{"ServiceLoader.iterator() receiver", "ServiceLoader.stream() receiver", "specific classpath registration"}
	e := NewModelEvaluator(hierarchy.NewForTesting(nil), nil, nil)
	for i, call := range calls {
		modelsForCall := registry.Lookup(call, nil)
		if len(modelsForCall) != 1 || modelsForCall[0].ID != serviceLoaderConsumptionModelID {
			t.Fatalf("Lookup(%+v) = %+v, want the consumption model", call, modelsForCall)
		}
		fact := e.Evaluate(modelsForCall[0], models.CallSite{Caller: models.Signature{Owner: "app.Entry", Name: "run", Desc: "()V"}, Call: call, PC: i + 1})
		if !fact.Incomplete || fact.Reason == "" || len(fact.Targets) != 0 {
			t.Fatalf("Evaluate(%+v) = %+v, want an explicit gap without guessed targets", call, fact)
		}
		if !strings.Contains(fact.Reason, wantFragments[i]) {
			t.Errorf("Evaluate(%+v) reason %q does not explain %q", call, fact.Reason, wantFragments[i])
		}
	}
}
