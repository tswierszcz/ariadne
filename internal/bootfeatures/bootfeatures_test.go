package bootfeatures

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func writeJar(t *testing.T, files map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "app.jar")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for name, body := range files {
		e, _ := w.Create(name)
		e.Write([]byte(body))
	}
	w.Close()
	f.Close()
	return p
}

func TestReadConfig(t *testing.T) {
	jar := writeJar(t, map[string]string{
		"application.properties":                "# comment\nserver.http2.enabled = true\nserver.port: 8080\n",
		"BOOT-INF/classes/application-prod.yml": "server:\n  ssl:\n    enabled: ${SSL:false} # env\n  undertow:\n    options:\n      socket:\n        WRITE_TIMEOUT: 5000\n",
		"other.properties":                      "server.http2.enabled=false\n",
	})
	c, err := ReadConfig(jar, map[string][]byte{"": []byte("<web-app><login-config/></web-app>")})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Files, []string{"BOOT-INF/classes/application-prod.yml", "application.properties"}) {
		t.Errorf("files %v", c.Files)
	}
	if !c.isTrue("server.http2.enabled") || c.Values["server.port"][0] != "8080" {
		t.Errorf("properties %v", c.Values)
	}
	if !c.Overridable["server.ssl.enabled"] || c.Values["server.ssl.enabled"][0] != "false" {
		t.Errorf("placeholder: %v %v", c.Values["server.ssl.enabled"], c.Overridable)
	}
	if !c.keyContains("server.undertow.options.", "write_timeout") || !c.WebXMLLoginConfig {
		t.Errorf("yaml/web.xml: %v %t", c.Values, c.WebXMLLoginConfig)
	}
}

func class(fqcn, super string, ifaces []string, app bool, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: super, Interfaces: ifaces, Methods: methods, IsApp: app}
}

func method(name, desc string, abstract bool, calls ...classfile.Call) hierarchy.Method {
	m := hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc}, Calls: calls}
	if abstract {
		m.AccessFlags = classfile.AccAbstract
	}
	return m
}

// fixture: Undertow with a listener interface, an HTTP/1 listener and an
// AJP and HTTP/2 listener, a channel invoking the listener, Spring Boot's
// Undertow factory.
func fixture(extra ...*hierarchy.ClassInfo) *hierarchy.Hierarchy {
	listener := "org.xnio.ChannelListener"
	classes := []*hierarchy.ClassInfo{
		class("io.undertow.Undertow", "java.lang.Object", nil, false),
		class("org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory", "java.lang.Object", nil, false),
		{FQCN: listener, SuperClass: "java.lang.Object", AccessFlags: classfile.AccInterface | classfile.AccAbstract, Methods: []hierarchy.Method{method("handleEvent", "()V", true)}},
		class("io.undertow.server.protocol.http.HttpReadListener", "java.lang.Object", []string{listener}, false, method("handleEvent", "()V", false)),
		class("io.undertow.server.protocol.ajp.AjpReadListener", "java.lang.Object", []string{listener}, false, method("handleEvent", "()V", false)),
		class("io.undertow.server.protocol.http2.Http2ReceiveListener", "java.lang.Object", []string{listener}, false, method("handleEvent", "()V", false)),
		class("io.undertow.Channel", "java.lang.Object", nil, false,
			method("close", "()V", false, classfile.Call{Kind: classfile.CallInterface, Owner: listener, Name: "handleEvent", Desc: "()V", PC: 4})),
	}
	return hierarchy.NewForTesting(append(classes, extra...))
}

func TestDecideAndExclude(t *testing.T) {
	h := fixture()
	r := Decide(h, Config{Values: map[string][]string{"server.http2.enabled": {"true"}}}, false)
	if !r.Applies {
		t.Fatal("not applied with Boot's Undertow factory in the input")
	}
	if r.Excluded("io.undertow.server.protocol.http2.Http2ReceiveListener") || !r.Excluded("io.undertow.server.protocol.ajp.AjpReadListener") ||
		r.Excluded("io.undertow.server.protocol.http.HttpReadListener") {
		t.Errorf("exclusions: off %v", r.Off())
	}
	g := callgraph.BuildWith(h, callgraph.CHA{Exclude: r.Excluded, ExcludeRule: "test"})
	from, _ := g.ID("io.undertow.Channel", "close", "()V")
	has := func(owner string) bool {
		to, ok := g.ID(owner, "handleEvent", "()V")
		if !ok {
			return false
		}
		_, ok = g.EdgeBetween(from, to)
		return ok
	}
	if !has("io.undertow.server.protocol.http.HttpReadListener") || !has("io.undertow.server.protocol.http2.Http2ReceiveListener") || has("io.undertow.server.protocol.ajp.AjpReadListener") {
		t.Error("CHA targets do not follow the feature decision")
	}

	// A placeholder switch leaves the decision to the environment: on.
	r = Decide(h, Config{Values: map[string][]string{"server.ssl.enabled": {"false"}}, Overridable: map[string]bool{"server.ssl.enabled": true}}, false)
	for _, s := range r.States {
		if s.Feature == "tls" && (!s.On || !s.Overridable) {
			t.Errorf("tls %+v", s)
		}
	}

	// An application class referencing the AJP listener API turns AJP on.
	app := class("app.Server", "java.lang.Object", nil, true,
		method("start", "()V", false, classfile.Call{Kind: classfile.CallVirtual, Owner: "io.undertow.Undertow$Builder", Name: "addAjpListener", Desc: "(ILjava/lang/String;)Lio/undertow/Undertow$Builder;", PC: 1}))
	r = Decide(fixture(app), Config{}, false)
	if r.Excluded("io.undertow.server.protocol.ajp.AjpReadListener") {
		t.Error("AJP off although the application adds an AJP listener")
	}

	// Without Undertow, or without Boot (no factory, no annotation): no-op.
	plain := hierarchy.NewForTesting([]*hierarchy.ClassInfo{class("io.undertow.Undertow", "java.lang.Object", nil, false)})
	if Decide(plain, Config{}, false).Applies {
		t.Error("applied without Spring Boot")
	}
}
