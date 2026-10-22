package valueflow

import (
	"testing"

	"ariadne/internal/bytecodeir"
)

// fakeFacts is a trivial Facts implementation, demonstrating that the
// contract is usable by a future consumer (SERVICE_LOADING S1, say)
// today, without waiting for D1/D2's real implementation.
type fakeFacts struct {
	args map[int]Value
}

func (f fakeFacts) ArgumentAt(_ bytecodeir.Site, i int) Value {
	if v, ok := f.args[i]; ok {
		return v
	}
	return Unknown("fakeFacts has no value for this argument index")
}

func (f fakeFacts) ReceiverAt(_ bytecodeir.Site) Value {
	return Unknown("fakeFacts has no receiver")
}

func TestFacts_contractIsImplementable(t *testing.T) {
	var f Facts = fakeFacts{args: map[int]Value{
		0: Const("java.lang.Class", "java.sql.Driver"),
	}}

	site := bytecodeir.Site{Method: "java.util.ServiceLoader#load", Descriptor: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;", PC: 5, Op: bytecodeir.OpCall}
	got := f.ArgumentAt(site, 0)
	if got.Kind != KindConst || got.Const != "java.sql.Driver" {
		t.Errorf("ArgumentAt(0) = %+v, want the known Class constant", got)
	}

	missing := f.ArgumentAt(site, 1)
	if missing.Kind != KindUnknown {
		t.Errorf("ArgumentAt(1) = %+v, want KindUnknown", missing)
	}
}
