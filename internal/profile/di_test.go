package profile

import (
	"reflect"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func autowiredField(name, desc string, qualifier string) classfile.Field {
	anns := []classfile.Annotation{{Type: autowiredFQCN}}
	if qualifier != "" {
		anns = append(anns, classfile.Annotation{Type: qualifierFQCN, Values: map[string]classfile.ElementValue{
			"value": {Kind: classfile.ElementConst, Const: qualifier},
		}})
	}
	return classfile.Field{Name: name, Descriptor: desc, AnnotationDetails: anns}
}

func TestInjectionPoints_autowiredField(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.Consumer",
			IsApp: true,
			Fields: []classfile.Field{
				autowiredField("service", "Ltest/FooService;", ""),
				{Name: "notInjected", Descriptor: "Ljava/lang/String;"},
			},
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}},
			},
		},
	})

	points := InjectionPoints(h)
	want := []InjectionPoint{{FQCN: "test.Consumer", Member: "service", ParamIndex: -1, Type: "test.FooService"}}
	if !reflect.DeepEqual(points, want) {
		t.Errorf("InjectionPoints = %+v, want %+v", points, want)
	}
}

func TestInjectionPoints_soleConstructorImplicitlyAutowired(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.Consumer",
			IsApp: true,
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "<init>", Descriptor: "(Ltest/FooService;)V"}},
			},
		},
	})

	points := InjectionPoints(h)
	want := []InjectionPoint{{FQCN: "test.Consumer", Member: "<init>", ParamIndex: 0, Type: "test.FooService"}}
	if !reflect.DeepEqual(points, want) {
		t.Errorf("InjectionPoints = %+v, want %+v", points, want)
	}
}

func TestInjectionPoints_multipleConstructorsWithoutAutowiredIsSkipped(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.Consumer",
			IsApp: true,
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}},
				{Method: classfile.Method{Name: "<init>", Descriptor: "(Ltest/FooService;)V"}},
			},
		},
	})

	points := InjectionPoints(h)
	if len(points) != 0 {
		t.Errorf("InjectionPoints(ambiguous constructors, none @Autowired) = %+v, want none", points)
	}
}

func TestInjectionPoints_autowiredConstructorAmongSeveralIsChosen(t *testing.T) {
	autowiredCtor := classfile.Method{
		Name: "<init>", Descriptor: "(Ltest/FooService;)V",
		AnnotationDetails: []classfile.Annotation{{Type: autowiredFQCN}},
	}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.Consumer",
			IsApp: true,
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}},
				{Method: autowiredCtor},
			},
		},
	})

	points := InjectionPoints(h)
	want := []InjectionPoint{{FQCN: "test.Consumer", Member: "<init>", ParamIndex: 0, Type: "test.FooService"}}
	if !reflect.DeepEqual(points, want) {
		t.Errorf("InjectionPoints = %+v, want %+v", points, want)
	}
}

func TestResolveInjections_singleCandidateResolves(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	points := []InjectionPoint{{FQCN: "test.Consumer", Member: "service", ParamIndex: -1, Type: "test.FooService"}}
	beans := []BeanName{{FQCN: "test.FooService"}}

	got := ResolveInjections(h, points, beans)
	if len(got) != 1 || got[0].Resolved != "test.FooService" {
		t.Errorf("ResolveInjections = %+v, want Resolved=test.FooService", got)
	}
}

func TestResolveInjections_noKnownBeanIsUnresolvedWithReason(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	points := []InjectionPoint{{FQCN: "test.Consumer", Member: "service", ParamIndex: -1, Type: "test.FooService"}}

	got := ResolveInjections(h, points, nil)
	if len(got) != 1 || got[0].Resolved != "" || got[0].Reason == "" {
		t.Errorf("ResolveInjections(no beans) = %+v, want unresolved with a reason", got)
	}
}

func TestResolveInjections_ambiguousWithoutQualifierOrPrimaryIsUnresolved(t *testing.T) {
	// Two distinct implementations, neither named nor primary.
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.FooService"},
		{FQCN: "test.FooServiceImplA", Interfaces: []string{"test.FooService"}},
		{FQCN: "test.FooServiceImplB", Interfaces: []string{"test.FooService"}},
	})
	points := []InjectionPoint{{FQCN: "test.Consumer", Member: "service", ParamIndex: -1, Type: "test.FooService"}}
	beans := []BeanName{
		{FQCN: "test.FooServiceImplA"},
		{FQCN: "test.FooServiceImplB"},
	}

	got := ResolveInjections(h, points, beans)
	if len(got) != 1 || got[0].Resolved != "" || len(got[0].Candidates) != 2 {
		t.Errorf("ResolveInjections(ambiguous) = %+v, want unresolved with 2 candidates", got)
	}
}

func TestResolveInjections_qualifierDisambiguates(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.FooService"},
		{FQCN: "test.FooServiceImplA", Interfaces: []string{"test.FooService"}},
		{FQCN: "test.FooServiceImplB", Interfaces: []string{"test.FooService"}},
	})
	points := []InjectionPoint{{FQCN: "test.Consumer", Member: "service", ParamIndex: -1, Type: "test.FooService", Qualifier: "implB"}}
	beans := []BeanName{
		{FQCN: "test.FooServiceImplA", Name: "implA"},
		{FQCN: "test.FooServiceImplB", Name: "implB"},
	}

	got := ResolveInjections(h, points, beans)
	if len(got) != 1 || got[0].Resolved != "test.FooServiceImplB" {
		t.Errorf("ResolveInjections(qualifier) = %+v, want Resolved=test.FooServiceImplB", got)
	}
}

func TestResolveInjections_primaryDisambiguates(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.FooService"},
		{FQCN: "test.FooServiceImplA", Interfaces: []string{"test.FooService"}},
		{FQCN: "test.FooServiceImplB", Interfaces: []string{"test.FooService"}},
	})
	points := []InjectionPoint{{FQCN: "test.Consumer", Member: "service", ParamIndex: -1, Type: "test.FooService"}}
	beans := []BeanName{
		{FQCN: "test.FooServiceImplA"},
		{FQCN: "test.FooServiceImplB", Primary: true},
	}

	got := ResolveInjections(h, points, beans)
	if len(got) != 1 || got[0].Resolved != "test.FooServiceImplB" {
		t.Errorf("ResolveInjections(primary) = %+v, want Resolved=test.FooServiceImplB", got)
	}
}
