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
		{ID: "fooService", Class: "demo.FooServiceImpl", Scope: "singleton", Primary: true},
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
