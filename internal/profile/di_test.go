package profile

import (
	"reflect"
	"strings"
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

	points := InjectionPoints(h, nil)
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

	points := InjectionPoints(h, []BeanName{{FQCN: "test.Consumer"}})
	want := []InjectionPoint{{FQCN: "test.Consumer", Member: "<init>", ParamIndex: 0, Type: "test.FooService"}}
	if !reflect.DeepEqual(points, want) {
		t.Errorf("InjectionPoints = %+v, want %+v", points, want)
	}
}

// TestInjectionPoints_soleConstructorOnPlainClassIsNotFabricated is
// this session's own fix: a class with exactly one constructor but no
// stereotype annotation and no `@Configuration` is never container-
// managed by Spring at all (an enum's compiler-generated constructor,
// a Lombok-generated POJO, a manually-`new`'d helper — all confirmed
// real false-positive shapes found on WebGoat), so it must contribute
// no injection point, unlike TestInjectionPoints_soleConstructorImplicitlyAutowired
// above where the class *is* a known stereotype bean.
func TestInjectionPoints_soleConstructorOnPlainClassIsNotFabricated(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.PlainPojo",
			IsApp: true,
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "<init>", Descriptor: "(Ljava/lang/String;)V"}},
			},
		},
	})

	if points := InjectionPoints(h, nil); len(points) != 0 {
		t.Errorf("InjectionPoints(plain class, no stereotype/@Configuration) = %+v, want none", points)
	}
}

// TestInjectionPoints_soleConstructorOnConfigurationClass confirms the
// other half of the same fix: a `@Configuration` class is container-
// managed even though BeanNames (the @Component-family stereotype
// scan) never catalogs it — the real WebGoat shape this session found
// (`WebSecurityConfig`, `@Configuration` plus Lombok's own
// `@AllArgsConstructor`, no `@Autowired` anywhere).
func TestInjectionPoints_soleConstructorOnConfigurationClass(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:        "test.SecurityConfig",
			IsApp:       true,
			Annotations: []string{"org.springframework.context.annotation.Configuration"},
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "<init>", Descriptor: "(Ltest/FooService;)V"}},
			},
		},
	})

	points := InjectionPoints(h, nil)
	want := []InjectionPoint{{FQCN: "test.SecurityConfig", Member: "<init>", ParamIndex: 0, Type: "test.FooService"}}
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

	points := InjectionPoints(h, []BeanName{{FQCN: "test.Consumer"}})
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

	// Not a stereotype/@Configuration class — confirms the explicit
	// `@Autowired` path stays unrestricted by this session's own fix,
	// unlike the implicit sole-constructor rule.
	points := InjectionPoints(h, nil)
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

// TestInjectionPoints_valueFieldWithoutAutowired confirms a `@Value`
// field is captured even with no `@Autowired` alongside it — Spring
// treats `@Value` as its own, independent injection annotation, not
// something that needs `@Autowired` to take effect.
func TestInjectionPoints_valueFieldWithoutAutowired(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.Consumer",
			IsApp: true,
			Fields: []classfile.Field{
				{Name: "driverClassName", Descriptor: "Ljava/lang/String;", AnnotationDetails: []classfile.Annotation{
					{Type: valueFQCN, Values: map[string]classfile.ElementValue{
						"value": {Kind: classfile.ElementConst, Const: "${spring.datasource.driver-class-name}"},
					}},
				}},
			},
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "<init>", Descriptor: "()V"}},
			},
		},
	})

	points := InjectionPoints(h, nil)
	want := []InjectionPoint{{
		FQCN: "test.Consumer", Member: "driverClassName", ParamIndex: -1, Type: "java.lang.String",
		ValueExpression: "${spring.datasource.driver-class-name}",
	}}
	if !reflect.DeepEqual(points, want) {
		t.Errorf("InjectionPoints = %+v, want %+v", points, want)
	}
}

// TestInjectionPoints_valueConstructorParam confirms the real WebGoat
// shape this increment addresses: DatabaseConfiguration's own
// `@Value("${spring.datasource.driver-class-name}") String
// driverClassName` constructor parameter.
func TestInjectionPoints_valueConstructorParam(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:        "test.DatabaseConfiguration",
			IsApp:       true,
			Annotations: []string{"org.springframework.context.annotation.Configuration"},
			Methods: []hierarchy.Method{
				{Method: classfile.Method{
					Name: "<init>", Descriptor: "(Ljava/lang/String;)V",
					ParameterAnnotationDetails: [][]classfile.Annotation{
						{{Type: valueFQCN, Values: map[string]classfile.ElementValue{
							"value": {Kind: classfile.ElementConst, Const: "${spring.datasource.driver-class-name}"},
						}}},
					},
				}},
			},
		},
	})

	points := InjectionPoints(h, nil)
	want := []InjectionPoint{{
		FQCN: "test.DatabaseConfiguration", Member: "<init>", ParamIndex: 0, Type: "java.lang.String",
		ValueExpression: "${spring.datasource.driver-class-name}",
	}}
	if !reflect.DeepEqual(points, want) {
		t.Errorf("InjectionPoints = %+v, want %+v", points, want)
	}
}

// TestResolveInjections_valueExpressionIsNotUnresolved confirms a
// @Value point is reported as its own, already-explained fact — never
// run through the bean-candidate search, never reported with
// ResolveInjections' own "no known bean assignable to ..." Reason
// (which would misreport a config-value injection as a failed bean
// lookup).
func TestResolveInjections_valueExpressionIsNotUnresolved(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	points := []InjectionPoint{{
		FQCN: "test.DatabaseConfiguration", Member: "<init>", ParamIndex: 0, Type: "java.lang.String",
		ValueExpression: "${spring.datasource.driver-class-name}",
	}}

	got := ResolveInjections(h, points, nil)
	if len(got) != 1 {
		t.Fatalf("ResolveInjections = %+v, want exactly 1", got)
	}
	if got[0].Resolved != "" || got[0].Candidates != nil {
		t.Errorf("ResolveInjections(@Value) = %+v, want Resolved=\"\" and no Candidates", got[0])
	}
	if got[0].Reason == "" || strings.Contains(got[0].Reason, "no known bean") {
		t.Errorf("Reason = %q, want a @Value-specific explanation, not the generic no-known-bean one", got[0].Reason)
	}
}
