package log4jconfig

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func writeJar(t *testing.T, name string, files map[string]string) Jar {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for n, body := range files {
		e, _ := w.Create(n)
		e.Write([]byte(body))
	}
	w.Close()
	f.Close()
	return Jar{Path: p, Name: name}
}

func TestReadConfig(t *testing.T) {
	app := writeJar(t, "app.jar", map[string]string{
		"log4j.properties": "# c\nlog4j.rootLogger=WARN, stdout, db\n" +
			"log4j.appender.stdout=org.apache.log4j.ConsoleAppender\n" +
			"log4j.appender.stdout.layout=org.apache.log4j.PatternLayout\n" +
			"log4j.appender.db = org.apache.log4j.jdbc.JDBCAppender\n" +
			"log4j.appender.db.sql=INSERT INTO log VALUES('%m')\n",
		"conf/log4j.properties": "log4j.appender.x=org.apache.log4j.net.JMSAppender\n",
	})
	dep := writeJar(t, "lib.jar", map[string]string{
		"log4j.xml": `<?xml version="1.0"?><!DOCTYPE log4j:configuration SYSTEM "log4j.dtd">
<log4j:configuration xmlns:log4j="http://jakarta.apache.org/log4j/">
  <appender name="async" class="org.apache.log4j.AsyncAppender"><appender-ref ref="sock"/></appender>
  <appender name="sock" class=" org.apache.log4j.net.SocketAppender "/>
</log4j:configuration>`,
	})
	c, err := ReadConfig([]Jar{app, dep}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Files, []string{"app.jar!/log4j.properties", "lib.jar!/log4j.xml"}) {
		t.Errorf("files %v (a file below the root is not log4j's default)", c.Files)
	}
	want := map[string]string{
		"org.apache.log4j.ConsoleAppender":    "app.jar!/log4j.properties",
		"org.apache.log4j.jdbc.JDBCAppender":  "app.jar!/log4j.properties",
		"org.apache.log4j.AsyncAppender":      "lib.jar!/log4j.xml",
		"org.apache.log4j.net.SocketAppender": "lib.jar!/log4j.xml",
	}
	if !reflect.DeepEqual(c.Classes, want) || len(c.Unreadable) != 0 {
		t.Errorf("classes %v, unreadable %v", c.Classes, c.Unreadable)
	}

	// A placeholder class, broken XML and a log4jConfigLocation outside
	// the class-path root make the configuration unreadable.
	bad := writeJar(t, "bad.jar", map[string]string{
		"log4j.properties": "log4j.appender.a=${APPENDER}\n",
		"log4j.xml":        "<log4j:configuration><appender class=",
	})
	c, err = ReadConfig([]Jar{bad}, map[string][]byte{"": []byte(
		"<web-app><context-param><param-name>log4jConfigLocation</param-name><param-value>/WEB-INF/log4j.properties</param-value></context-param></web-app>")})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Unreadable) != 3 {
		t.Errorf("unreadable %v, want web.xml location, placeholder and XML", c.Unreadable)
	}

	// classpath:<root file> is read; a missing one is unreadable.
	custom := writeJar(t, "app.jar", map[string]string{"log4j-prod.properties": "log4j.appender.f=org.apache.log4j.FileAppender\n"})
	webXML := func(loc string) map[string][]byte {
		return map[string][]byte{"": []byte("<web-app><context-param><param-name>log4jConfigLocation</param-name><param-value>" + loc + "</param-value></context-param></web-app>")}
	}
	c, _ = ReadConfig([]Jar{custom}, webXML("classpath:log4j-prod.properties"))
	if c.Classes["org.apache.log4j.FileAppender"] == "" || len(c.Unreadable) != 0 {
		t.Errorf("classpath location: %v %v", c.Classes, c.Unreadable)
	}
	c, _ = ReadConfig([]Jar{custom}, webXML("classpath:log4j-test.properties"))
	if len(c.Unreadable) != 1 {
		t.Errorf("missing classpath location: %v", c.Unreadable)
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

const (
	appenderIface = "org.apache.log4j.Appender"
	skeleton      = "org.apache.log4j.AppenderSkeleton"
	console       = "org.apache.log4j.ConsoleAppender"
	jdbc          = "org.apache.log4j.jdbc.JDBCAppender"
	jms           = "org.apache.log4j.net.JMSAppender"
	doAppend      = "(Lorg/apache/log4j/spi/LoggingEvent;)V"
)

// fixture: log4j's dispatch shape — AppenderAttachableImpl calls
// Appender#doAppend, AppenderSkeleton#doAppend (inherited by every
// appender) calls this.append, three appenders declare append.
func fixture(extra ...*hierarchy.ClassInfo) *hierarchy.Hierarchy {
	classes := []*hierarchy.ClassInfo{
		{FQCN: appenderIface, SuperClass: "java.lang.Object", AccessFlags: classfile.AccInterface | classfile.AccAbstract, Methods: []hierarchy.Method{method("doAppend", doAppend, true)}},
		{FQCN: skeleton, SuperClass: "java.lang.Object", Interfaces: []string{appenderIface}, AccessFlags: classfile.AccAbstract, Methods: []hierarchy.Method{
			method("doAppend", doAppend, false, classfile.Call{Kind: classfile.CallVirtual, Owner: skeleton, Name: "append", Desc: doAppend, PC: 1}),
			method("append", doAppend, true),
		}},
		class(console, skeleton, nil, false, method("append", doAppend, false)),
		class(jdbc, skeleton, nil, false, method("append", doAppend, false)),
		class(jms, skeleton, nil, false, method("append", doAppend, false)),
		class(marker, "java.lang.Object", nil, false,
			method("appendLoopOnAppenders", doAppend, false, classfile.Call{Kind: classfile.CallInterface, Owner: appenderIface, Name: "doAppend", Desc: doAppend, PC: 1})),
	}
	return hierarchy.NewForTesting(append(classes, extra...))
}

func TestDecide(t *testing.T) {
	h := fixture()
	cfg := Config{Files: []string{"app.jar!/log4j.properties"}, Classes: map[string]string{console: "app.jar!/log4j.properties"}}
	r := Decide(h, cfg)
	if !r.Applies || !reflect.DeepEqual(r.On(), []string{"ConsoleAppender"}) || !reflect.DeepEqual(r.Off(), []string{"JDBCAppender", "JMSAppender"}) {
		t.Fatalf("decision %+v", r)
	}
	if !r.Excluded(jdbc) || !r.Excluded(jdbc+"$1") || r.Excluded(console) || r.Excluded(skeleton) {
		t.Error("exclusions")
	}

	// The graph: the interface call and AppenderSkeleton's own
	// this.append reach ConsoleAppender only.
	g := callgraph.BuildWith(h, callgraph.CHA{Exclude: r.Excluded, ExcludeRule: "test"})
	edge := func(fromOwner, fromName, toOwner, toName string) bool {
		from, ok1 := g.ID(fromOwner, fromName, doAppend)
		to, ok2 := g.ID(toOwner, toName, doAppend)
		if !ok1 || !ok2 {
			return false
		}
		_, ok := g.EdgeBetween(from, to)
		return ok
	}
	if !edge(marker, "appendLoopOnAppenders", skeleton, "doAppend") {
		t.Error("inherited doAppend not reached through ConsoleAppender")
	}
	if !edge(skeleton, "doAppend", console, "append") || edge(skeleton, "doAppend", jdbc, "append") || edge(skeleton, "doAppend", jms, "append") {
		t.Error("this.append targets do not follow the configuration")
	}

	// Application code creating JDBCAppender turns it on.
	app := class("app.Logs", "java.lang.Object", nil, true, hierarchy.Method{
		Method:   classfile.Method{Name: "setup", Descriptor: "()V"},
		InitRefs: []classfile.InitRef{{Kind: classfile.InitNew, Owner: jdbc}},
	})
	r = Decide(fixture(app), cfg)
	if r.Excluded(jdbc) || !r.Excluded(jms) {
		t.Errorf("referenced appender: %+v", r.States)
	}

	// No decision: no log4j 1.2, no packaged file, unreadable file, a
	// configurator called by the application.
	confApp := class("app.Boot", "java.lang.Object", nil, true,
		method("main", "()V", false, classfile.Call{Kind: classfile.CallStatic, Owner: "org.apache.log4j.PropertyConfigurator", Name: "configure", Desc: "(Ljava/lang/String;)V", PC: 1}))
	for name, tc := range map[string]struct {
		h   *hierarchy.Hierarchy
		cfg Config
		why string
	}{
		"no log4j":     {hierarchy.NewForTesting([]*hierarchy.ClassInfo{class(console, "java.lang.Object", nil, false)}), cfg, "not in the input"},
		"no file":      {h, Config{Classes: map[string]string{}}, "external"},
		"unreadable":   {h, Config{Files: cfg.Files, Unreadable: []string{"x: placeholder"}}, "placeholder"},
		"configurator": {fixture(confApp), cfg, "app.Boot calls org.apache.log4j.PropertyConfigurator#configure"},
	} {
		r := Decide(tc.h, tc.cfg)
		if r.Applies || r.Excluded(jdbc) || !strings.Contains(r.Why, tc.why) {
			t.Errorf("%s: %+v", name, r)
		}
	}
}
