package profile

import (
	"reflect"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestExternalEntries_httpMappingUnconditionallyActive(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:  "test.ItemController",
			IsApp: true,
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "list", Descriptor: "()V", AnnotationDetails: []classfile.Annotation{
					{Type: "org.springframework.web.bind.annotation.GetMapping"},
				}}},
			},
		},
	})

	entries := ExternalEntries(h, nil, nil, nil, nil)
	want := []ExternalEntry{{FQCN: "test.ItemController", Method: "list", Basis: BasisHTTPMapping, Active: ConditionTrue}}
	if !reflect.DeepEqual(entries, want) {
		t.Errorf("ExternalEntries = %+v, want %+v", entries, want)
	}
}

func TestExternalEntries_profileGatedRouteUnknownWithoutActiveList(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:              "test.DevController",
			IsApp:             true,
			AnnotationDetails: []classfile.Annotation{{Type: "org.springframework.context.annotation.Profile", Values: map[string]classfile.ElementValue{"value": strArrayElement("dev")}}},
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "list", Descriptor: "()V", AnnotationDetails: []classfile.Annotation{
					{Type: "org.springframework.web.bind.annotation.GetMapping"},
				}}},
			},
		},
	})

	entries := ExternalEntries(h, nil, nil, nil, nil)
	if len(entries) != 1 || entries[0].Active != ConditionUnknown {
		t.Errorf("ExternalEntries = %+v, want one entry with Active=unknown (no -active-profiles given)", entries)
	}
}

func TestExternalEntries_profileGatedRouteResolvesWithActiveList(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:              "test.DevController",
			IsApp:             true,
			AnnotationDetails: []classfile.Annotation{{Type: "org.springframework.context.annotation.Profile", Values: map[string]classfile.ElementValue{"value": strArrayElement("dev")}}},
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "list", Descriptor: "()V", AnnotationDetails: []classfile.Annotation{
					{Type: "org.springframework.web.bind.annotation.GetMapping"},
				}}},
			},
		},
	})

	active := ExternalEntries(h, []string{"dev"}, nil, nil, nil)
	if len(active) != 1 || active[0].Active != ConditionTrue {
		t.Errorf("ExternalEntries(active=dev) = %+v, want Active=true", active)
	}

	inactive := ExternalEntries(h, []string{"prod"}, nil, nil, nil)
	if len(inactive) != 1 || inactive[0].Active != ConditionFalse {
		t.Errorf("ExternalEntries(active=prod) = %+v, want Active=false", inactive)
	}
}

func TestExternalEntries_webXMLServletIncluded(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN: "demo.MainServlet",
			Methods: []hierarchy.Method{
				{Method: classfile.Method{Name: "doGet", Descriptor: "()V"}},
			},
		},
	})

	entries := ExternalEntries(h, nil, []string{"demo.MainServlet"}, nil, nil)
	want := []ExternalEntry{{FQCN: "demo.MainServlet", Method: "doGet", Basis: BasisServlet, Active: ConditionTrue}}
	if !reflect.DeepEqual(entries, want) {
		t.Errorf("ExternalEntries = %+v, want %+v", entries, want)
	}
}

func TestExternalEntries_noEntriesIsEmptyNotNilPanic(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	if entries := ExternalEntries(h, nil, nil, nil, nil); len(entries) != 0 {
		t.Errorf("ExternalEntries(empty hierarchy) = %+v, want none", entries)
	}
}
