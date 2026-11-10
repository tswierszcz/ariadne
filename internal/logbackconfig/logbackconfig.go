// Package logbackconfig decides which of logback's own appenders an
// application's configuration leaves out, so that CHA does not dispatch
// into their classes (FRAMEWORKS:F2, l2) — the logback counterpart of
// internal/log4jconfig (l1).
//
// logback creates appenders reflectively from its configuration (Joran),
// or Spring Boot creates them in code (DefaultLogbackConfiguration), so
// every call logback makes on an appender — Appender#doAppend from
// AppenderAttachableImpl, AppenderBase/UnsynchronizedAppenderBase's own
// this.append, an OutputStreamAppender's subAppend — is expanded by CHA
// to every appender logback ships: DBAppender's SQL, SMTPAppender's
// evaluator, RollingFileAppender's rollover. Measured on petclinic and
// WebGoat (Spring Boot defaults, console only): three FRAMEWORK paths
// each through appenders neither configures (testdata/graphbench).
//
// The catalog (Appenders, Version) lists logback-core/logback-classic's
// concrete appenders (1.1.11–1.5.12). One of them is on when:
//   - a packaged configuration names it: logback.xml, logback-test.xml,
//     logback-spring.xml or logback-test-spring.xml at a class-path root
//     of the application or a dependency jar (every one found counts),
//     `<appender class="...">` anywhere in it (springProfile and if
//     blocks are unioned), `<include resource="...">` followed into the
//     input's jars (Spring Boot's base.xml includes its file appender),
//     a short class name resolved through `<import class="...">`;
//   - or, with no packaged file, Spring Boot's logging system is in the
//     input: its default configuration — ConsoleAppender, and
//     RollingFileAppender when a packaged application*.properties/yml
//     sets logging.file.name, logging.file.path, logging.file or
//     logging.path (any profile; a placeholder value counts as set);
//   - or application code references it (a call, `new`, a static read).
//
// Every other catalog appender is off: not a CHA/RTA target. Direct calls
// are untouched; logback-access's appenders are not in the catalog.
//
// The model decides nothing — every appender stays on — when the
// configuration may come from elsewhere: no packaged file and no Spring
// Boot (-Dlogback.configurationFile, BasicConfigurator), Boot's
// logging.config set, a logback*.groovy file, a Configurator registered
// in META-INF/services, a ${...} or unresolvable class name, an
// include by file or URL (or a resource not in the input), an
// unparsable file, or application code running Joran itself.
package logbackconfig

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"ariadne/internal/appref"
	"ariadne/internal/bootfeatures"
	"ariadne/internal/hierarchy"
)

// Version versions the appender catalog and the rules above.
const Version = "l2-1"

// Appenders are logback's own concrete appenders: every class of
// logback-core and logback-classic implementing
// ch.qos.logback.core.Appender, checked in the bytecode of 1.1.11, 1.2.3,
// 1.2.13, 1.3.14, 1.4.14 and 1.5.12 (DBAppender up to 1.2.3, the JMS
// appenders in 1.1.x only, WithLayoutListAppender from 1.3; the
// testUtil appenders of 1.3+ are left out — nothing configures them).
var Appenders = []string{
	"ch.qos.logback.classic.AsyncAppender",
	"ch.qos.logback.classic.db.DBAppender",
	"ch.qos.logback.classic.helpers.WithLayoutListAppender",
	"ch.qos.logback.classic.net.JMSQueueAppender",
	"ch.qos.logback.classic.net.JMSTopicAppender",
	"ch.qos.logback.classic.net.SMTPAppender",
	"ch.qos.logback.classic.net.SSLSocketAppender",
	"ch.qos.logback.classic.net.SocketAppender",
	"ch.qos.logback.classic.net.SyslogAppender",
	"ch.qos.logback.classic.net.server.SSLServerSocketAppender",
	"ch.qos.logback.classic.net.server.ServerSocketAppender",
	"ch.qos.logback.classic.sift.SiftingAppender",
	"ch.qos.logback.core.AsyncAppenderBase",
	"ch.qos.logback.core.ConsoleAppender",
	"ch.qos.logback.core.FileAppender",
	"ch.qos.logback.core.OutputStreamAppender",
	"ch.qos.logback.core.helpers.NOPAppender",
	"ch.qos.logback.core.read.CyclicBufferAppender",
	"ch.qos.logback.core.read.ListAppender",
	"ch.qos.logback.core.rolling.RollingFileAppender",
}

const (
	// marker: logback-classic is in the input.
	marker = "ch.qos.logback.classic.Logger"
	// bootMarker: Spring Boot's logback logging system is in the input.
	bootMarker = "org.springframework.boot.logging.logback.LogbackLoggingSystem"
	console    = "ch.qos.logback.core.ConsoleAppender"
	rolling    = "ch.qos.logback.core.rolling.RollingFileAppender"
)

// configFiles are the class-path root files logback (and, with -spring,
// Spring Boot) looks for.
var configFiles = map[string]bool{
	"logback.xml": true, "logback-test.xml": true, "logback-spring.xml": true, "logback-test-spring.xml": true,
}

// groovyFiles are the Groovy configurations logback 1.2 and Boot read; a
// Groovy file is code, not a list of classes.
var groovyFiles = map[string]bool{
	"logback.groovy": true, "logback-test.groovy": true, "logback-spring.groovy": true, "logback-test-spring.groovy": true,
}

// bootFileKeys turn on Spring Boot's default file appender.
var bootFileKeys = []string{"logging.file.name", "logging.file.path", "logging.file", "logging.path"}

// configurators run Joran on a configuration of the caller's choosing.
var configurators = []string{
	"ch.qos.logback.classic.joran.JoranConfigurator#doConfigure",
	"ch.qos.logback.core.joran.GenericConfigurator#doConfigure",
	"ch.qos.logback.core.joran.GenericXMLConfigurator#doConfigure",
	"ch.qos.logback.core.joran.JoranConfiguratorBase#doConfigure",
	"ch.qos.logback.classic.util.ContextInitializer#configureByResource",
}

// Config is what the packaged configuration says.
type Config struct {
	// Files are the configuration files read, as "<jar>!/<entry>"
	// (included resources too).
	Files []string
	// Classes maps every appender class a file names to the first file
	// naming it.
	Classes map[string]string
	// Unreadable explains why the configuration cannot be relied on.
	Unreadable []string
}

// ReadConfig reads the logback configuration files at the root of every
// jar (the application's classes first) and the resources they include.
func ReadConfig(jars []appref.Jar) (Config, error) {
	c := Config{Classes: map[string]string{}}
	idx, err := indexJars(jars)
	defer idx.close()
	if err != nil {
		return c, err
	}
	for _, e := range idx.entries {
		switch {
		case groovyFiles[e.name]:
			c.Unreadable = append(c.Unreadable, e.where()+": a Groovy configuration")
		case strings.HasPrefix(e.name, "META-INF/services/ch.qos.logback.classic.spi.Configurator"):
			c.Unreadable = append(c.Unreadable, e.where()+": a Configurator registered as a service")
		}
	}
	seen := map[string]bool{}
	for _, e := range idx.entries {
		if configFiles[e.name] {
			c.readFile(idx, e, seen, 0)
		}
	}
	return c, nil
}

type entry struct {
	jar  string
	name string
	f    *zip.File
}

func (e entry) where() string { return e.jar + "!/" + e.name }

type index struct {
	entries []entry
	byName  map[string][]entry
	closers []io.Closer
}

func (idx *index) close() {
	for _, c := range idx.closers {
		c.Close()
	}
}

func indexJars(jars []appref.Jar) (*index, error) {
	idx := &index{byName: map[string][]entry{}}
	for _, j := range jars {
		r, err := zip.OpenReader(j.Path)
		if err != nil {
			return idx, err
		}
		idx.closers = append(idx.closers, r)
		var es []entry
		for _, f := range r.File {
			if configFiles[f.Name] || groovyFiles[f.Name] || strings.HasPrefix(f.Name, "META-INF/services/ch.qos.logback.classic.spi.Configurator") || strings.HasSuffix(f.Name, ".xml") {
				es = append(es, entry{jar: j.Name, name: f.Name, f: f})
			}
		}
		sort.Slice(es, func(a, b int) bool { return es[a].name < es[b].name })
		for _, e := range es {
			idx.entries = append(idx.entries, e)
			idx.byName[e.name] = append(idx.byName[e.name], e)
		}
	}
	return idx, nil
}

func (c *Config) readFile(idx *index, e entry, seen map[string]bool, depth int) {
	if seen[e.where()] {
		return
	}
	seen[e.where()] = true
	if depth > 8 {
		c.Unreadable = append(c.Unreadable, e.where()+": includes nested too deep")
		return
	}
	rc, err := e.f.Open()
	if err != nil {
		c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: %v", e.where(), err))
		return
	}
	data, err := io.ReadAll(io.LimitReader(rc, 4<<20))
	rc.Close()
	if err != nil {
		c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: %v", e.where(), err))
		return
	}
	c.Files = append(c.Files, e.where())
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	imports := map[string]string{}
	var classes []string
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: not parsable XML: %v", e.where(), err))
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		attr := func(name string) (string, bool) {
			for _, a := range se.Attr {
				if a.Name.Local == name {
					return strings.TrimSpace(a.Value), true
				}
			}
			return "", false
		}
		switch se.Name.Local {
		case "import":
			if cl, ok := attr("class"); ok {
				imports[cl[strings.LastIndex(cl, ".")+1:]] = cl
			}
		case "appender":
			if cl, ok := attr("class"); ok {
				classes = append(classes, cl)
			}
		case "include":
			if res, ok := attr("resource"); ok && !strings.Contains(res, "${") {
				inc := idx.byName[strings.TrimPrefix(res, "/")]
				if len(inc) == 0 {
					c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: included resource %q not in the input", e.where(), res))
				}
				for _, ie := range inc {
					c.readFile(idx, ie, seen, depth+1)
				}
				continue
			}
			c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: an include by file, URL or placeholder", e.where()))
		}
	}
	for _, cl := range classes {
		switch {
		case strings.Contains(cl, "${"):
			c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: appender class %q is a placeholder", e.where(), cl))
			continue
		case !strings.Contains(cl, "."):
			full, ok := imports[cl]
			if !ok {
				c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: appender class %q not imported", e.where(), cl))
				continue
			}
			cl = full
		}
		if _, ok := c.Classes[cl]; !ok {
			c.Classes[cl] = e.where()
		}
	}
}

// State is one catalog appender's decision.
type State struct {
	Class string
	On    bool
	Why   string
}

// Result is the model's decision for one input.
type Result struct {
	// Applies is false when the model leaves every appender on; Why then
	// says why.
	Applies bool
	Why     string
	// Source names what decided: the packaged files, or Spring Boot's
	// default configuration.
	Source string
	States []State
	off    map[string]bool
}

// NotInInput is Result.Why when logback-classic is absent.
const NotInInput = "logback not in the input"

// Decide applies the rules of the package comment. boot is the packaged
// Spring Boot configuration (bootfeatures.ReadConfig).
func Decide(h *hierarchy.Hierarchy, c Config, boot bootfeatures.Config) Result {
	if _, ok := h.Lookup(marker); !ok {
		return Result{Why: NotInInput}
	}
	if len(c.Unreadable) > 0 {
		return Result{Why: strings.Join(c.Unreadable, "; ")}
	}
	_, isBoot := h.Lookup(bootMarker)
	if isBoot && len(boot.Values["logging.config"]) > 0 {
		return Result{Why: "Spring Boot's logging.config names the configuration (" + strings.Join(boot.Values["logging.config"], ", ") + ")"}
	}
	on := map[string]string{}
	var source string
	switch {
	case len(c.Files) > 0:
		source = "config " + strings.Join(c.Files, ", ")
		for cl, file := range c.Classes {
			on[cl] = "configured in " + file
		}
	case isBoot:
		source = "Spring Boot default configuration"
		on[console] = "Spring Boot default configuration"
		for _, k := range bootFileKeys {
			if len(boot.Values[k]) > 0 {
				on[rolling] = "Spring Boot default configuration, " + k + " set"
				break
			}
		}
	default:
		return Result{Why: "no packaged logback configuration and no Spring Boot (the configuration may be external)"}
	}
	refs := appref.Collect(h)
	for _, cfg := range configurators {
		if by, ok := refs.Member(cfg); ok {
			return Result{Why: fmt.Sprintf("application code runs Joran itself (%s calls %s)", by, cfg)}
		}
	}
	r := Result{Applies: true, Source: source, off: map[string]bool{}}
	for _, class := range Appenders {
		if _, ok := h.Lookup(class); !ok {
			continue
		}
		st := State{Class: class}
		if why, ok := on[class]; ok {
			st.On, st.Why = true, why
		} else if by, ok := refs.Owner(class); ok {
			st.On, st.Why = true, "referenced by "+by
		} else {
			st.Why = "not configured or referenced"
			r.off[class] = true
		}
		r.States = append(r.States, st)
	}
	return r
}

// Excluded reports whether class is an appender the configuration leaves
// out (or a nested class of one).
func (r Result) Excluded(class string) bool {
	if r.off[class] {
		return true
	}
	if outer, _, ok := strings.Cut(class, "$"); ok {
		return r.off[outer]
	}
	return false
}

func (r Result) names(on bool) []string {
	var out []string
	for _, st := range r.States {
		if st.On == on {
			out = append(out, st.Class[strings.LastIndex(st.Class, ".")+1:])
		}
	}
	return out
}

// Off lists the excluded appenders, in catalog order, by simple name.
func (r Result) Off() []string { return r.names(false) }

// On lists the appenders left on, in catalog order, by simple name.
func (r Result) On() []string { return r.names(true) }
