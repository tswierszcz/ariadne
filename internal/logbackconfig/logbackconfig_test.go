package logbackconfig

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/appref"
	"ariadne/internal/bootfeatures"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func writeJar(t *testing.T, name string, files map[string]string) appref.Jar {
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
	return appref.Jar{Path: p, Name: name}
}

// bootJar mimics spring-boot's own logback resources.
func bootJar(t *testing.T) appref.Jar {
	return writeJar(t, "spring-boot.jar", map[string]string{
		"org/springframework/boot/logging/logback/base.xml": `<included>
  <include resource="org/springframework/boot/logging/logback/defaults.xml"/>
  <include resource="org/springframework/boot/logging/logback/console-appender.xml"/>
  <include resource="org/springframework/boot/logging/logback/file-appender.xml"/>
  <root level="INFO"><appender-ref ref="CONSOLE"/><appender-ref ref="FILE"/></root>
</included>`,
		"org/springframework/boot/logging/logback/defaults.xml":         `<included><conversionRule conversionWord="clr" converterClass="x.Y"/></included>`,
		"org/springframework/boot/logging/logback/console-appender.xml": `<included><appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender"/></included>`,
		"org/springframework/boot/logging/logback/file-appender.xml":    `<included><appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender"/></included>`,
	})
}

func TestReadConfig(t *testing.T) {
	app := writeJar(t, "app.jar", map[string]string{
		"logback-spring.xml": `<configuration>
  <include resource="org/springframework/boot/logging/logback/base.xml"/>
  <import class="ch.qos.logback.classic.net.SMTPAppender"/>
  <springProfile name="prod">
    <appender name="MAIL" class="SMTPAppender">
      <evaluator class="ch.qos.logback.classic.boolex.JaninoEventEvaluator"><expression>level &gt;= ERROR</expression></evaluator>
    </appender>
  </springProfile>
  <appender name="ASYNC" class="ch.qos.logback.classic.AsyncAppender"><appender-ref ref="MAIL"/></appender>
</configuration>`,
		"conf/logback.xml": `<configuration><appender class="ch.qos.logback.classic.db.DBAppender"/></configuration>`,
	})
	c, err := ReadConfig([]appref.Jar{app, bootJar(t)})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Unreadable) != 0 {
		t.Fatalf("unreadable %v", c.Unreadable)
	}
	if len(c.Files) != 5 || c.Files[0] != "app.jar!/logback-spring.xml" {
		t.Errorf("files %v (the file and 4 included Boot resources; a file below the root is not logback's)", c.Files)
	}
	want := []string{
		"ch.qos.logback.classic.AsyncAppender",
		"ch.qos.logback.classic.boolex.JaninoEventEvaluator",
		"ch.qos.logback.classic.net.SMTPAppender",
		"ch.qos.logback.core.ConsoleAppender",
		"ch.qos.logback.core.rolling.RollingFileAppender",
	}
	var got []string
	for cl := range c.Classes {
		got = append(got, cl)
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("classes %v", got)
	}
	if !c.Bare["expression"] || !c.Bare["root"] || c.Bare["evaluator"] {
		t.Errorf("bare elements %v", c.Bare)
	}

	for name, files := range map[string]map[string]string{
		"placeholder":      {"logback.xml": `<configuration><appender class="${APPENDER}"/></configuration>`},
		"not imported":     {"logback.xml": `<configuration><appender class="SMTPAppender"/></configuration>`},
		"file include":     {"logback.xml": `<configuration><include file="/etc/app/logback-extra.xml"/></configuration>`},
		"missing resource": {"logback.xml": `<configuration><include resource="nope.xml"/></configuration>`},
		"groovy":           {"logback.groovy": `appender("X", ConsoleAppender)`},
		"service":          {"META-INF/services/ch.qos.logback.classic.spi.Configurator": "app.MyConfigurator\n"},
		"broken XML":       {"logback.xml": `<configuration><appender class=`},
	} {
		c, err := ReadConfig([]appref.Jar{writeJar(t, "app.jar", files)})
		if err != nil || len(c.Unreadable) != 1 {
			t.Errorf("%s: unreadable %v (%v)", name, c.Unreadable, err)
		}
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
	appenderIface = "ch.qos.logback.core.Appender"
	base          = "ch.qos.logback.core.UnsynchronizedAppenderBase"
	smtp          = "ch.qos.logback.classic.net.SMTPAppender"
	dbApp         = "ch.qos.logback.classic.db.DBAppender"
	attachable    = "ch.qos.logback.core.spi.AppenderAttachableImpl"
	desc          = "(Ljava/lang/Object;)V"
)

// fixture: logback's dispatch shape — AppenderAttachableImpl calls
// Appender#doAppend, the base class's doAppend calls this.append; console,
// rolling-file, SMTP and DB appenders declare append.
func fixture(boot bool, extra ...*hierarchy.ClassInfo) *hierarchy.Hierarchy {
	classes := []*hierarchy.ClassInfo{
		class(marker, "java.lang.Object", nil, false),
		{FQCN: appenderIface, SuperClass: "java.lang.Object", AccessFlags: classfile.AccInterface | classfile.AccAbstract, Methods: []hierarchy.Method{method("doAppend", desc, true)}},
		{FQCN: base, SuperClass: "java.lang.Object", Interfaces: []string{appenderIface}, AccessFlags: classfile.AccAbstract, Methods: []hierarchy.Method{
			method("doAppend", desc, false, classfile.Call{Kind: classfile.CallVirtual, Owner: base, Name: "append", Desc: desc, PC: 1}),
			method("append", desc, true),
		}},
		class(console, base, nil, false, method("append", desc, false)),
		class(rolling, console, nil, false, method("append", desc, false)),
		class(smtp, base, nil, false, method("append", desc, false)),
		class(dbApp, base, nil, false, method("append", desc, false)),
		class(attachable, "java.lang.Object", nil, false,
			method("appendLoopOnAppenders", desc, false, classfile.Call{Kind: classfile.CallInterface, Owner: appenderIface, Name: "doAppend", Desc: desc, PC: 1})),
	}
	if boot {
		classes = append(classes, class(bootMarker, "java.lang.Object", nil, false))
	}
	return hierarchy.NewForTesting(append(classes, extra...))
}

func TestDecide(t *testing.T) {
	noBoot := bootfeatures.Config{Values: map[string][]string{}}

	// Spring Boot defaults, no file property: console only.
	r := Decide(fixture(true), Config{Classes: map[string]string{}}, noBoot)
	if !r.Applies || r.Source != "Spring Boot default configuration" || !reflect.DeepEqual(r.On(), []string{"ConsoleAppender"}) ||
		!reflect.DeepEqual(r.Off(), []string{"DBAppender", "SMTPAppender", "RollingFileAppender"}) {
		t.Fatalf("boot defaults: %+v", r)
	}
	g := callgraph.BuildWith(fixture(true), callgraph.CHA{Exclude: r.Excluded, ExcludeRule: "test"})
	edge := func(fromOwner, fromName, toOwner string) bool {
		from, ok1 := g.ID(fromOwner, fromName, desc)
		to, ok2 := g.ID(toOwner, "append", desc)
		if !ok1 || !ok2 {
			return false
		}
		_, ok := g.EdgeBetween(from, to)
		return ok
	}
	if !edge(base, "doAppend", console) || edge(base, "doAppend", smtp) || edge(base, "doAppend", dbApp) || edge(base, "doAppend", rolling) {
		t.Error("this.append targets do not follow the configuration")
	}

	// A file property (a placeholder counts) turns on the file appender.
	withFile := bootfeatures.Config{Values: map[string][]string{"logging.file.name": {"${LOG_FILE}"}}}
	if r := Decide(fixture(true), Config{Classes: map[string]string{}}, withFile); r.Excluded(rolling) || !r.Excluded(smtp) {
		t.Errorf("boot with logging.file.name: %+v", r.States)
	}

	// A packaged file decides, Boot or not.
	cfg := Config{Files: []string{"app.jar!/logback.xml"}, Classes: map[string]string{smtp: "app.jar!/logback.xml"}}
	if r := Decide(fixture(false), cfg, noBoot); !r.Applies || r.Excluded(smtp) || !r.Excluded(console) {
		t.Errorf("packaged file: %+v", r.States)
	}

	// Application code creating DBAppender turns it on.
	app := class("app.Logs", "java.lang.Object", nil, true, hierarchy.Method{
		Method:   classfile.Method{Name: "setup", Descriptor: "()V"},
		InitRefs: []classfile.InitRef{{Kind: classfile.InitNew, Owner: dbApp}},
	})
	if r := Decide(fixture(true, app), Config{Classes: map[string]string{}}, noBoot); r.Excluded(dbApp) {
		t.Errorf("referenced appender: %+v", r.States)
	}

	// No decision.
	joran := class("app.Boot", "java.lang.Object", nil, true,
		method("main", "()V", false, classfile.Call{Kind: classfile.CallVirtual, Owner: "ch.qos.logback.classic.joran.JoranConfigurator", Name: "doConfigure", Desc: "(Ljava/lang/String;)V", PC: 1}))
	for name, tc := range map[string]struct {
		h    *hierarchy.Hierarchy
		cfg  Config
		boot bootfeatures.Config
		why  string
	}{
		"no logback":     {hierarchy.NewForTesting([]*hierarchy.ClassInfo{class(console, "java.lang.Object", nil, false)}), cfg, noBoot, NotInInput},
		"no file, plain": {fixture(false), Config{Classes: map[string]string{}}, noBoot, "may be external"},
		"logging.config": {fixture(true), Config{Classes: map[string]string{}}, bootfeatures.Config{Values: map[string][]string{"logging.config": {"classpath:custom.xml"}}}, "logging.config"},
		"unreadable":     {fixture(true), Config{Unreadable: []string{"x: a Groovy configuration"}}, noBoot, "Groovy"},
		"joran":          {fixture(true, joran), Config{Classes: map[string]string{}}, noBoot, "app.Boot calls ch.qos.logback.classic.joran.JoranConfigurator#doConfigure"},
	} {
		r := Decide(tc.h, tc.cfg, tc.boot)
		if r.Applies || r.Excluded(smtp) || !strings.Contains(r.Why, tc.why) {
			t.Errorf("%s: %+v", name, r)
		}
	}
}

// TestDecideComponents (l2-2): Joran-only components follow the
// configuration; converters, created classes, setter parameter types and
// appenders are never in the rule.
func TestDecideComponents(t *testing.T) {
	const (
		lc        = "ch.qos.logback.core.spi.LifeCycle"
		janino    = "ch.qos.logback.classic.boolex.JaninoEventEvaluator"
		janBase   = "ch.qos.logback.core.boolex.JaninoEventEvaluatorBase"
		dsSource  = "ch.qos.logback.core.db.DataSourceConnectionSource"
		threshold = "ch.qos.logback.classic.filter.ThresholdFilter"
		conv      = "ch.qos.logback.classic.pattern.LevelConverter"
		matcher   = "ch.qos.logback.core.boolex.Matcher"
		encoder   = "ch.qos.logback.classic.encoder.PatternLayoutEncoder"
		helper    = "ch.qos.logback.core.util.StatusListenerConfigHelper"
	)
	iface := func(fqcn string, ifs ...string) *hierarchy.ClassInfo {
		return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: "java.lang.Object", Interfaces: ifs, AccessFlags: classfile.AccInterface | classfile.AccAbstract,
			Methods: []hierarchy.Method{method("start", "()V", true)}}
	}
	extra := []*hierarchy.ClassInfo{
		iface(lc),
		iface("ch.qos.logback.core.pattern.Converter", lc),
		{FQCN: janBase, SuperClass: "java.lang.Object", Interfaces: []string{lc}, AccessFlags: classfile.AccAbstract, Methods: []hierarchy.Method{method("start", "()V", false)}},
		class(janino, janBase, nil, false),
		class(dsSource, "java.lang.Object", []string{lc}, false, method("start", "()V", false)),
		class(threshold, "java.lang.Object", []string{lc}, false, method("start", "()V", false)),
		class(conv, "java.lang.Object", []string{"ch.qos.logback.core.pattern.Converter"}, false, method("start", "()V", false)),
		class(matcher, "java.lang.Object", []string{lc}, false, method("start", "()V", false)),
		class(encoder, "java.lang.Object", []string{lc}, false, method("start", "()V", false)),
		// JaninoEventEvaluatorBase's host: an adder typed with Matcher, and
		// a `new PatternLayoutEncoder` (Spring Boot's defaults).
		class("ch.qos.logback.core.boolex.Host", "java.lang.Object", nil, false,
			method("addMatcher", "(L"+strings.ReplaceAll(matcher, ".", "/")+";)V", false),
			hierarchy.Method{Method: classfile.Method{Name: "boot", Descriptor: "()V"}, InitRefs: []classfile.InitRef{{Kind: classfile.InitNew, Owner: encoder}}}),
		// StatusListenerConfigHelper's shape: LifeCycle#start on a
		// reflectively created object.
		class(helper, "java.lang.Object", nil, false,
			method("initAndAddListener", "()V", false, classfile.Call{Kind: classfile.CallInterface, Owner: lc, Name: "start", Desc: "()V", PC: 1})),
	}
	h := fixture(true, extra...)
	noBoot := bootfeatures.Config{Values: map[string][]string{}}
	r := Decide(h, Config{Classes: map[string]string{}, Bare: map[string]bool{}}, noBoot)
	got := map[string]bool{}
	for _, st := range r.Components {
		got[st.Class] = st.On
	}
	want := map[string]bool{janino: false, dsSource: false, threshold: false, matcher: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("components %v, want %v (converter, created encoder and abstract base never in the rule)", got, want)
	}
	g := callgraph.BuildWith(h, callgraph.CHA{Exclude: r.Excluded, ExcludeRule: "test"})
	from, _ := g.ID(helper, "initAndAddListener", "()V")
	reaches := func(owner string) bool {
		to, ok := g.ID(owner, "start", "()V")
		if !ok {
			return false
		}
		_, ok = g.EdgeBetween(from, to)
		return ok
	}
	if reaches(janBase) || reaches(dsSource) || !reaches(matcher) || !reaches(conv) || !reaches(encoder) {
		t.Error("LifeCycle#start targets do not follow the component decision")
	}

	// A bare <evaluator> turns on its default class, a class attribute
	// the class it names.
	r = Decide(h, Config{Files: []string{"app.jar!/logback.xml"}, Classes: map[string]string{dsSource: "app.jar!/logback.xml"}, Bare: map[string]bool{"evaluator": true}}, noBoot)
	if r.Excluded(janino) || r.Excluded(dsSource) || !r.Excluded(threshold) {
		t.Errorf("configured components: %+v", r.Components)
	}
}
