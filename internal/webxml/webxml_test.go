package webxml

import (
	"reflect"
	"strings"
	"testing"
)

const sampleWebXML = `<?xml version="1.0" encoding="UTF-8"?>
<web-app xmlns="http://xmlns.jcp.org/xml/ns/javaee" version="4.0">
  <display-name>Demo</display-name>
  <context-param>
    <param-name>config.path</param-name>
    <param-value>/etc/demo.conf</param-value>
  </context-param>
  <listener>
    <listener-class>demo.StartupListener</listener-class>
  </listener>
  <filter>
    <filter-name>auth</filter-name>
    <filter-class>demo.AuthFilter</filter-class>
  </filter>
  <filter-mapping>
    <filter-name>auth</filter-name>
    <url-pattern>/secure/*</url-pattern>
  </filter-mapping>
  <servlet>
    <servlet-name>main</servlet-name>
    <servlet-class>demo.MainServlet</servlet-class>
  </servlet>
  <servlet-mapping>
    <servlet-name>main</servlet-name>
    <url-pattern>/main</url-pattern>
  </servlet-mapping>
  <session-config>
    <session-timeout>30</session-timeout>
  </session-config>
  <error-page>
    <error-code>404</error-code>
    <location>/404.html</location>
  </error-page>
</web-app>
`

func TestParse_fullShape(t *testing.T) {
	got, err := Parse(strings.NewReader(sampleWebXML))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	wantServlets := []Servlet{{Name: "main", Class: "demo.MainServlet"}}
	if !reflect.DeepEqual(got.Servlets, wantServlets) {
		t.Errorf("Servlets = %+v, want %+v", got.Servlets, wantServlets)
	}

	wantServletMappings := []ServletMapping{{Name: "main", URLPattern: "/main"}}
	if !reflect.DeepEqual(got.ServletMappings, wantServletMappings) {
		t.Errorf("ServletMappings = %+v, want %+v", got.ServletMappings, wantServletMappings)
	}

	wantFilters := []Filter{{Name: "auth", Class: "demo.AuthFilter"}}
	if !reflect.DeepEqual(got.Filters, wantFilters) {
		t.Errorf("Filters = %+v, want %+v", got.Filters, wantFilters)
	}

	wantFilterMappings := []FilterMapping{{Name: "auth", URLPattern: "/secure/*"}}
	if !reflect.DeepEqual(got.FilterMappings, wantFilterMappings) {
		t.Errorf("FilterMappings = %+v, want %+v", got.FilterMappings, wantFilterMappings)
	}

	wantListeners := []Listener{{Class: "demo.StartupListener"}}
	if !reflect.DeepEqual(got.Listeners, wantListeners) {
		t.Errorf("Listeners = %+v, want %+v", got.Listeners, wantListeners)
	}

	wantContextParams := []ContextParam{{Name: "config.path", Value: "/etc/demo.conf"}}
	if !reflect.DeepEqual(got.ContextParams, wantContextParams) {
		t.Errorf("ContextParams = %+v, want %+v", got.ContextParams, wantContextParams)
	}

	wantSkipped := []string{"display-name", "session-config", "error-page"}
	if !reflect.DeepEqual(got.Skipped, wantSkipped) {
		t.Errorf("Skipped = %v, want %v", got.Skipped, wantSkipped)
	}
}

// TestParse_filterMappingByServletName confirms the <servlet-name>-based
// filter mapping shape (the alternative to <url-pattern>) is read too.
func TestParse_filterMappingByServletName(t *testing.T) {
	doc := `<web-app>
  <filter-mapping>
    <filter-name>auth</filter-name>
    <servlet-name>main</servlet-name>
  </filter-mapping>
</web-app>`
	got, err := Parse(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []FilterMapping{{Name: "auth", ServletName: "main"}}
	if !reflect.DeepEqual(got.FilterMappings, want) {
		t.Errorf("FilterMappings = %+v, want %+v", got.FilterMappings, want)
	}
}

// TestParse_malformedDocumentIsError confirms a truncated/invalid XML
// document is a real decode error, distinct from an unmodeled-but-
// well-formed element (which is Skipped, not an error).
func TestParse_malformedDocumentIsError(t *testing.T) {
	_, err := Parse(strings.NewReader("<web-app><servlet><servlet-name>x</web-app>"))
	if err == nil {
		t.Fatal("Parse(malformed) = nil error, want an error")
	}
}

// TestParse_undefinedEntityIsError confirms this package never resolves
// an external/undefined entity — a document referencing one is a
// decode error, not a silently-expanded value (encoding/xml's own
// default behavior with no Entity map set; this test pins that down as
// this package's own contract, not an incidental library default).
func TestParse_undefinedEntityIsError(t *testing.T) {
	doc := `<!DOCTYPE web-app [<!ENTITY xxe SYSTEM "file:///etc/passwd">]>
<web-app><display-name>&xxe;</display-name></web-app>`
	_, err := Parse(strings.NewReader(doc))
	if err == nil {
		t.Fatal("Parse(undefined entity) = nil error, want an error (entity never resolved)")
	}
}

func TestParse_emptyWebApp(t *testing.T) {
	got, err := Parse(strings.NewReader("<web-app></web-app>"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got.Servlets)+len(got.ServletMappings)+len(got.Filters)+len(got.FilterMappings)+len(got.Listeners)+len(got.ContextParams)+len(got.Skipped) != 0 {
		t.Errorf("Parse(empty web-app) = %+v, want everything empty", got)
	}
}
