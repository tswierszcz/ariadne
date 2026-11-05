package serviceload

import (
	"testing"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/valueflow"
)

func TestContainsTraceSite_requiresExactStreamCallsite(t *testing.T) {
	want := bytecodeir.Site{Method: "app.Entry#run", Descriptor: "()V", PC: 24, Op: bytecodeir.OpCall}
	if !containsTraceSite([]bytecodeir.Site{want}, want) {
		t.Fatal("expected the exact stream() call site to match")
	}
	otherStream := want
	otherStream.PC++
	if containsTraceSite([]bytecodeir.Site{otherStream}, want) {
		t.Fatal("a different stream() call site must not be associated with this terminal")
	}
}

func TestServiceLoadOrigins_followsIteratorToExactLoadSite(t *testing.T) {
	loadSite := bytecodeir.Site{Method: "app.Entry#run", Descriptor: "()V", PC: 12, Op: bytecodeir.OpCall}
	load := valueflow.CallResult(valueflow.CallResultRef{
		Owner: serviceLoaderOwner, Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;", Site: loadSite,
	})
	loaderIterator := valueflow.TracedCall(valueflow.CallTraceValue{
		Owner: serviceLoaderOwner, Name: "iterator", Desc: "()Ljava/util/Iterator;",
		Receiver: &load,
	})
	next := valueflow.TracedCall(valueflow.CallTraceValue{
		Owner: "java.util.Iterator", Name: "next", Desc: "()Ljava/lang/Object;",
		Receiver: &loaderIterator,
	})
	loads := map[siteKey]ServiceLoadSite{
		{caller: "app.Entry", desc: "run()V", pc: 12}: {
			Interface: "api.Service", LoaderKind: LoaderClasspath,
		},
	}

	origins := serviceLoadOrigins(next, loads, nil)
	if !origins.ServiceLoaderValue || origins.Reason != "" {
		t.Fatalf("origins=%+v, want a fully resolved ServiceLoader origin", origins)
	}
	if len(origins.Interfaces) != 1 || origins.Interfaces[0] != "api.Service" {
		t.Fatalf("interfaces=%v, want [api.Service]", origins.Interfaces)
	}
	if len(origins.LoaderKinds) != 1 || origins.LoaderKinds[0] != LoaderClasspath {
		t.Fatalf("loader kinds=%v, want [classpath]", origins.LoaderKinds)
	}

	wrongSite := loadSite
	wrongSite.PC++
	wrongLoad := valueflow.CallResult(valueflow.CallResultRef{
		Owner: serviceLoaderOwner, Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;", Site: wrongSite,
	})
	wrongIterator := valueflow.TracedCall(valueflow.CallTraceValue{Owner: serviceLoaderOwner, Name: "iterator", Desc: "()Ljava/util/Iterator;", Receiver: &wrongLoad})
	wrongNext := valueflow.TracedCall(valueflow.CallTraceValue{Owner: "java.util.Iterator", Name: "next", Desc: "()Ljava/lang/Object;", Receiver: &wrongIterator})
	origins = serviceLoadOrigins(wrongNext, loads, nil)
	if !origins.ServiceLoaderValue || len(origins.Interfaces) != 0 || origins.Reason == "" {
		t.Fatalf("wrong-site origins=%+v, want explicit no-match gap", origins)
	}
}

func TestLinkedProviderMethodCall_requiresMaterializedReceiverAndExactServiceOwner(t *testing.T) {
	loadSite := bytecodeir.Site{Method: "app.Entry#run", Descriptor: "()V", PC: 12, Op: bytecodeir.OpCall}
	load := valueflow.CallResult(valueflow.CallResultRef{Owner: serviceLoaderOwner, Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;", Site: loadSite})
	iterator := valueflow.TracedCall(valueflow.CallTraceValue{Owner: serviceLoaderOwner, Name: "iterator", Desc: "()Ljava/util/Iterator;", Receiver: &load})
	next := valueflow.TracedCall(valueflow.CallTraceValue{Owner: "java.util.Iterator", Name: "next", Desc: "()Ljava/lang/Object;", Receiver: &iterator})
	loads := map[siteKey]ServiceLoadSite{{caller: "app.Entry", desc: "run()V", pc: 12}: {Interface: "api.Service", LoaderKind: LoaderClasspath}}
	origins := serviceLoadOrigins(next, loads, nil)
	if !linkedProviderMethodCall(next, "api.Service", origins) {
		t.Fatal("expected a provider instance call through its exact service interface to be linked")
	}
	if linkedProviderMethodCall(next, "other.Service", origins) {
		t.Fatal("must not link a call through an unrelated service owner")
	}
	if linkedProviderMethodCall(iterator, "api.Service", serviceLoadOrigins(iterator, loads, nil)) {
		t.Fatal("an Iterator value is not a materialized service instance")
	}
}
