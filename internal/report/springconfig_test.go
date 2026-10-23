package report

import (
	"reflect"
	"testing"

	"ariadne/internal/profile"
)

func TestBuildConfigConflicts_differentClassSameNameIsConflict(t *testing.T) {
	annotationBeans := []profile.BeanName{{FQCN: "demo.FooServiceImpl", Name: "fooService"}}
	xmlBeans := []XMLBeanRow{{Source: "applicationContext.xml", ID: "fooService", Class: "demo.OtherFooService"}}

	got := buildConfigConflicts(annotationBeans, xmlBeans)
	want := []ConfigConflictRow{{
		Name: "fooService",
		Candidates: []NamedBeanCandidate{
			{Kind: "annotation", Class: "demo.FooServiceImpl"},
			{Kind: "applicationContext.xml", Class: "demo.OtherFooService"},
		},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildConfigConflicts = %+v, want %+v", got, want)
	}
}

func TestBuildConfigConflicts_sameClassSameNameIsNotAConflict(t *testing.T) {
	annotationBeans := []profile.BeanName{{FQCN: "demo.FooServiceImpl", Name: "fooService"}}
	xmlBeans := []XMLBeanRow{{Source: "applicationContext.xml", ID: "fooService", Class: "demo.FooServiceImpl"}}

	got := buildConfigConflicts(annotationBeans, xmlBeans)
	if len(got) != 0 {
		t.Errorf("buildConfigConflicts(same class) = %+v, want none (harmless re-declaration)", got)
	}
}

func TestBuildConfigConflicts_unnamedBeansAreIgnored(t *testing.T) {
	annotationBeans := []profile.BeanName{{FQCN: "demo.FooServiceImpl", Name: ""}}
	xmlBeans := []XMLBeanRow{{Source: "applicationContext.xml", ID: "", Class: "demo.OtherFooService"}}

	got := buildConfigConflicts(annotationBeans, xmlBeans)
	if len(got) != 0 {
		t.Errorf("buildConfigConflicts(unnamed) = %+v, want none", got)
	}
}

func TestBuildConfigConflicts_noOverlapIsEmpty(t *testing.T) {
	annotationBeans := []profile.BeanName{{FQCN: "demo.FooServiceImpl", Name: "fooService"}}
	xmlBeans := []XMLBeanRow{{Source: "applicationContext.xml", ID: "barService", Class: "demo.BarServiceImpl"}}

	got := buildConfigConflicts(annotationBeans, xmlBeans)
	if len(got) != 0 {
		t.Errorf("buildConfigConflicts(no overlap) = %+v, want none", got)
	}
}
