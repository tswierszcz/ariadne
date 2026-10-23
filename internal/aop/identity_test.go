package aop

import (
	"reflect"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestFindProxyHandlers_finalFieldIsCandidateTarget(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: invocationHandlerFQCN, AccessFlags: classfile.AccInterface},
		{
			FQCN:       "test.LessonConnectionInvocationHandler",
			IsApp:      true,
			Interfaces: []string{invocationHandlerFQCN},
			Fields: []classfile.Field{
				{Name: "targetConnection", Descriptor: "Ljava/sql/Connection;", AccessFlags: classfile.AccFinal},
			},
		},
	})

	handlers := FindProxyHandlers(h)
	want := []ProxyHandler{{
		FQCN:         "test.LessonConnectionInvocationHandler",
		TargetFields: []TargetField{{Name: "targetConnection", Type: "java.sql.Connection"}},
	}}
	if !reflect.DeepEqual(handlers, want) {
		t.Errorf("FindProxyHandlers = %+v, want %+v", handlers, want)
	}
}

func TestFindProxyHandlers_nonFinalFieldSkipped(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: invocationHandlerFQCN, AccessFlags: classfile.AccInterface},
		{
			FQCN:       "test.MutableHandler",
			IsApp:      true,
			Interfaces: []string{invocationHandlerFQCN},
			Fields: []classfile.Field{
				{Name: "target", Descriptor: "Ljava/sql/Connection;"},
			},
		},
	})

	handlers := FindProxyHandlers(h)
	want := []ProxyHandler{{FQCN: "test.MutableHandler"}}
	if !reflect.DeepEqual(handlers, want) {
		t.Errorf("FindProxyHandlers = %+v, want %+v (non-final field is not a candidate target)", handlers, want)
	}
}

func TestFindProxyHandlers_staticFieldSkipped(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: invocationHandlerFQCN, AccessFlags: classfile.AccInterface},
		{
			FQCN:       "test.StaticFieldHandler",
			IsApp:      true,
			Interfaces: []string{invocationHandlerFQCN},
			Fields: []classfile.Field{
				{Name: "shared", Descriptor: "Ljava/sql/Connection;", AccessFlags: classfile.AccFinal | classfile.AccStatic},
			},
		},
	})

	handlers := FindProxyHandlers(h)
	want := []ProxyHandler{{FQCN: "test.StaticFieldHandler"}}
	if !reflect.DeepEqual(handlers, want) {
		t.Errorf("FindProxyHandlers = %+v, want %+v (static field is never an instance target)", handlers, want)
	}
}

func TestFindProxyHandlers_primitiveFieldSkipped(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: invocationHandlerFQCN, AccessFlags: classfile.AccInterface},
		{
			FQCN:       "test.CounterHandler",
			IsApp:      true,
			Interfaces: []string{invocationHandlerFQCN},
			Fields: []classfile.Field{
				{Name: "count", Descriptor: "I", AccessFlags: classfile.AccFinal},
			},
		},
	})

	handlers := FindProxyHandlers(h)
	want := []ProxyHandler{{FQCN: "test.CounterHandler"}}
	if !reflect.DeepEqual(handlers, want) {
		t.Errorf("FindProxyHandlers = %+v, want %+v (a primitive field is never a proxy target)", handlers, want)
	}
}

func TestFindProxyHandlers_nonHandlerClassNotIncluded(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: invocationHandlerFQCN, AccessFlags: classfile.AccInterface},
		{FQCN: "test.PlainService", IsApp: true},
	})

	if handlers := FindProxyHandlers(h); len(handlers) != 0 {
		t.Errorf("FindProxyHandlers = %+v, want none (not an InvocationHandler subtype)", handlers)
	}
}

func TestFindProxyHandlers_dependencyJarHandlerNotIncluded(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: invocationHandlerFQCN, AccessFlags: classfile.AccInterface},
		{
			FQCN:       "lib.LibHandler",
			IsApp:      false,
			Interfaces: []string{invocationHandlerFQCN},
		},
	})

	if handlers := FindProxyHandlers(h); len(handlers) != 0 {
		t.Errorf("FindProxyHandlers = %+v, want none (dependency-jar class, not application code)", handlers)
	}
}
