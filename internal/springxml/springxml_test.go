package springxml

import (
	"reflect"
	"strings"
	"testing"
)

const sampleBeansXML = `<?xml version="1.0" encoding="UTF-8"?>
<beans xmlns="http://www.springframework.org/schema/beans">
  <bean id="fooService" class="demo.FooServiceImpl" scope="singleton" primary="true">
    <property name="name" value="foo"/>
  </bean>
  <bean name="barService, barAlias" class="demo.BarServiceImpl"/>
  <import resource="other.xml"/>
  <context:component-scan base-package="demo"/>
</beans>
`

func TestParse_fullShape(t *testing.T) {
	got, err := Parse(strings.NewReader(sampleBeansXML))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	want := []Bean{
		{ID: "fooService", Class: "demo.FooServiceImpl", Scope: "singleton", Primary: true, Properties: []string{"name"}},
		{ID: "barService", Class: "demo.BarServiceImpl"},
	}
	if !reflect.DeepEqual(got.Beans, want) {
		t.Errorf("Beans = %+v, want %+v", got.Beans, want)
	}

	wantSkipped := []string{"import", "component-scan"}
	if !reflect.DeepEqual(got.Skipped, wantSkipped) {
		t.Errorf("Skipped = %v, want %v", got.Skipped, wantSkipped)
	}
}

func TestParse_malformedDocumentIsError(t *testing.T) {
	_, err := Parse(strings.NewReader("<beans><bean id=\"x\" class=\"demo.X\"></beans>"))
	if err == nil {
		t.Fatal("Parse(malformed) = nil error, want an error")
	}
}

func TestParse_undefinedEntityIsError(t *testing.T) {
	doc := `<!DOCTYPE beans [<!ENTITY xxe SYSTEM "file:///etc/passwd">]>
<beans><bean id="&xxe;" class="demo.X"/></beans>`
	_, err := Parse(strings.NewReader(doc))
	if err == nil {
		t.Fatal("Parse(undefined entity) = nil error, want an error (entity never resolved)")
	}
}

func TestParse_emptyBeans(t *testing.T) {
	got, err := Parse(strings.NewReader("<beans></beans>"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got.Beans) != 0 || len(got.Skipped) != 0 {
		t.Errorf("Parse(empty beans) = %+v, want everything empty", got)
	}
}

const lifecycleXML = `<?xml version="1.0" encoding="UTF-8"?>
<beans xmlns="http://www.springframework.org/schema/beans"
       xmlns:p="http://www.springframework.org/schema/p"
       xmlns:c="http://www.springframework.org/schema/c"
       default-init-method="setup">
  <bean id="service" class="demo.Service" init-method="start" destroy-method="stop" lazy-init="true">
    <constructor-arg ref="repo"/>
    <constructor-arg value="2"/>
    <property name="name" value="x"/>
    <property name="helper"><bean class="demo.Helper"/></property>
    <property name="list"><list><bean class="demo.Item" p:size="3"/></list></property>
  </bean>
  <bean id="pbean" class="demo.PBean" p:label="y" p:repo-ref="repo" c:_0="z"/>
  <bean id="base" class="demo.Base" abstract="true"/>
  <bean id="child" parent="base"/>
  <bean id="made" class="demo.Factory" factory-method="create"/>
  <beans profile="dev" default-destroy-method="dispose">
    <bean id="devOnly" class="demo.DevOnly"/>
  </beans>
</beans>
`

// TestParse_lifecycle: constructor arguments (elements and c:), properties
// (elements and p:, -ref stripped), init/destroy methods and the enclosing
// defaults (a nested section adds its own), parent/abstract/lazy/factory
// attributes, nested <beans> definitions, and inner beans at any depth.
func TestParse_lifecycle(t *testing.T) {
	got, err := Parse(strings.NewReader(lifecycleXML))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []Bean{
		{ID: "service", Class: "demo.Service", LazyInit: true, InitMethod: "start", DestroyMethod: "stop", DefaultInitMethod: "setup",
			Properties: []string{"name", "helper", "list"}, ConstructorArgs: 2},
		{ID: "pbean", Class: "demo.PBean", DefaultInitMethod: "setup", Properties: []string{"label", "repo"}, ConstructorArgs: 1},
		{ID: "base", Class: "demo.Base", Abstract: true, DefaultInitMethod: "setup"},
		{ID: "child", Parent: "base", DefaultInitMethod: "setup"},
		{ID: "made", Class: "demo.Factory", FactoryMethod: "create", DefaultInitMethod: "setup"},
		{ID: "devOnly", Class: "demo.DevOnly", DefaultInitMethod: "setup", DefaultDestroyMethod: "dispose"},
	}
	if !reflect.DeepEqual(got.Beans, want) {
		t.Errorf("Beans =\n %+v\nwant\n %+v", got.Beans, want)
	}
	wantInner := []Bean{
		{Class: "demo.Helper", DefaultInitMethod: "setup"},
		{Class: "demo.Item", DefaultInitMethod: "setup", Properties: []string{"size"}},
	}
	if !reflect.DeepEqual(got.Inner, wantInner) {
		t.Errorf("Inner =\n %+v\nwant\n %+v", got.Inner, wantInner)
	}
	if len(got.Skipped) != 0 {
		t.Errorf("Skipped = %v, want none (nested <beans> is modeled)", got.Skipped)
	}
}
