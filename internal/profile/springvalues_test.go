package profile

import (
	"reflect"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func strArrayElement(vals ...string) classfile.ElementValue {
	arr := make([]classfile.ElementValue, len(vals))
	for i, v := range vals {
		arr[i] = classfile.ElementValue{Kind: classfile.ElementConst, Const: v}
	}
	return classfile.ElementValue{Kind: classfile.ElementArray, Array: arr}
}

func enumArrayElement(enumType string, consts ...string) classfile.ElementValue {
	arr := make([]classfile.ElementValue, len(consts))
	for i, c := range consts {
		arr[i] = classfile.ElementValue{Kind: classfile.ElementEnum, EnumType: enumType, EnumConst: c}
	}
	return classfile.ElementValue{Kind: classfile.ElementArray, Array: arr}
}

// TestRoutes_getMappingWithClassPrefix confirms the real-world shape:
// a class-level @RequestMapping("/api") prefix combined with a method-
// level @GetMapping("/items") composed annotation, whose own verb
// comes from the annotation type itself (GetMapping never carries a
// "method" element).
func TestRoutes_getMappingWithClassPrefix(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.ItemController",
			IsApp: true,
			AnnotationDetails: []classfile.Annotation{
				{Type: "org.springframework.web.bind.annotation.RequestMapping", Values: map[string]classfile.ElementValue{
					"value": strArrayElement("/api"),
				}},
			},
			Methods: []hierarchy.Method{
				{
					Method: classfile.Method{Name: "list", Descriptor: "()V", AnnotationDetails: []classfile.Annotation{
						{Type: "org.springframework.web.bind.annotation.GetMapping", Values: map[string]classfile.ElementValue{
							"value": strArrayElement("/items"),
						}},
					}},
				},
			},
		},
	})

	routes := Routes(h)
	if len(routes) != 1 {
		t.Fatalf("Routes = %+v, want 1 route", routes)
	}
	r := routes[0]
	if !reflect.DeepEqual(r.Paths, []string{"/api/items"}) {
		t.Errorf("Paths = %v, want [/api/items]", r.Paths)
	}
	if !reflect.DeepEqual(r.HTTPMethods, []string{"GET"}) {
		t.Errorf("HTTPMethods = %v, want [GET]", r.HTTPMethods)
	}
}

// TestRoutes_requestMappingWithExplicitMethod confirms the base
// @RequestMapping's own "method" element (an enum array) is read when
// present, instead of defaulting to "every verb".
func TestRoutes_requestMappingWithExplicitMethod(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.Controller",
			IsApp: true,
			Methods: []hierarchy.Method{
				{
					Method: classfile.Method{Name: "save", Descriptor: "()V", AnnotationDetails: []classfile.Annotation{
						{Type: "org.springframework.web.bind.annotation.RequestMapping", Values: map[string]classfile.ElementValue{
							"path":   strArrayElement("/save"),
							"method": enumArrayElement("org.springframework.web.bind.annotation.RequestMethod", "POST"),
						}},
					}},
				},
			},
		},
	})

	routes := Routes(h)
	if len(routes) != 1 {
		t.Fatalf("Routes = %+v, want 1 route", routes)
	}
	if !reflect.DeepEqual(routes[0].HTTPMethods, []string{"POST"}) {
		t.Errorf("HTTPMethods = %v, want [POST]", routes[0].HTTPMethods)
	}
	if !reflect.DeepEqual(routes[0].Paths, []string{"/save"}) {
		t.Errorf("Paths = %v, want [/save]", routes[0].Paths)
	}
}

// TestRoutes_unspecifiedMethodIsNilNotGuessed confirms a bare
// @RequestMapping with no "method" element reports HTTPMethods == nil
// (every verb), never guessed as GET.
func TestRoutes_unspecifiedMethodIsNilNotGuessed(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.Controller",
			IsApp: true,
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "any", Descriptor: "()V", AnnotationDetails: []classfile.Annotation{
					{Type: "org.springframework.web.bind.annotation.RequestMapping", Values: map[string]classfile.ElementValue{
						"value": strArrayElement("/any"),
					}},
				}}},
			},
		},
	})

	routes := Routes(h)
	if len(routes) != 1 {
		t.Fatalf("Routes = %+v, want 1 route", routes)
	}
	if routes[0].HTTPMethods != nil {
		t.Errorf("HTTPMethods = %v, want nil (unspecified)", routes[0].HTTPMethods)
	}
}

func TestBeanNames_explicitValueAndPrimary(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:        "test.FooService",
			IsApp:       true,
			Annotations: []string{"org.springframework.context.annotation.Primary"},
			AnnotationDetails: []classfile.Annotation{
				{Type: "org.springframework.stereotype.Service", Values: map[string]classfile.ElementValue{
					"value": {Kind: classfile.ElementConst, Const: "fooService"},
				}},
			},
		},
	})

	names := BeanNames(h)
	if len(names) != 1 {
		t.Fatalf("BeanNames = %+v, want 1 entry", names)
	}
	if names[0].Name != "fooService" || !names[0].Primary {
		t.Errorf("BeanNames[0] = %+v, want Name=fooService Primary=true", names[0])
	}
}

func TestBeanNames_noExplicitNameIsEmpty(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:              "test.BarService",
			IsApp:             true,
			AnnotationDetails: []classfile.Annotation{{Type: "org.springframework.stereotype.Service"}},
		},
	})

	names := BeanNames(h)
	if len(names) != 1 || names[0].Name != "" || names[0].Primary {
		t.Errorf("BeanNames = %+v, want one entry with empty Name and Primary=false", names)
	}
}

func TestProfileConditions_noActiveListIsUnknown(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:              "test.DevConfig",
			IsApp:             true,
			AnnotationDetails: []classfile.Annotation{{Type: "org.springframework.context.annotation.Profile", Values: map[string]classfile.ElementValue{"value": strArrayElement("dev")}}},
		},
	})

	conds := ProfileConditions(h, nil)
	if len(conds) != 1 {
		t.Fatalf("ProfileConditions = %+v, want 1 entry", conds)
	}
	if conds[0].Active != ConditionUnknown {
		t.Errorf("Active = %v, want ConditionUnknown (no active-profile list given)", conds[0].Active)
	}
}

func TestProfileConditions_activeListResolvesTrueFalse(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:              "test.DevConfig",
			IsApp:             true,
			AnnotationDetails: []classfile.Annotation{{Type: "org.springframework.context.annotation.Profile", Values: map[string]classfile.ElementValue{"value": strArrayElement("dev")}}},
		},
		{
			FQCN:              "test.ProdConfig",
			IsApp:             true,
			AnnotationDetails: []classfile.Annotation{{Type: "org.springframework.context.annotation.Profile", Values: map[string]classfile.ElementValue{"value": strArrayElement("prod")}}},
		},
	})

	conds := ProfileConditions(h, []string{"dev"})
	if len(conds) != 2 {
		t.Fatalf("ProfileConditions = %+v, want 2 entries", conds)
	}
	byFQCN := map[string]Condition{}
	for _, c := range conds {
		byFQCN[c.FQCN] = c.Active
	}
	if byFQCN["test.DevConfig"] != ConditionTrue {
		t.Errorf("DevConfig Active = %v, want ConditionTrue", byFQCN["test.DevConfig"])
	}
	if byFQCN["test.ProdConfig"] != ConditionFalse {
		t.Errorf("ProdConfig Active = %v, want ConditionFalse", byFQCN["test.ProdConfig"])
	}
}

// TestProfileConditions_negatedProfile confirms Spring's own "!x"
// negation syntax: @Profile("!dev") is active exactly when "dev" is
// not in the active list.
func TestProfileConditions_negatedProfile(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:              "test.NotDevConfig",
			IsApp:             true,
			AnnotationDetails: []classfile.Annotation{{Type: "org.springframework.context.annotation.Profile", Values: map[string]classfile.ElementValue{"value": strArrayElement("!dev")}}},
		},
	})

	conds := ProfileConditions(h, []string{"prod"})
	if len(conds) != 1 || conds[0].Active != ConditionTrue {
		t.Errorf("ProfileConditions(!dev, active=[prod]) = %+v, want ConditionTrue", conds)
	}

	conds = ProfileConditions(h, []string{"dev"})
	if len(conds) != 1 || conds[0].Active != ConditionFalse {
		t.Errorf("ProfileConditions(!dev, active=[dev]) = %+v, want ConditionFalse", conds)
	}
}
