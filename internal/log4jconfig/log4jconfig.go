// Package log4jconfig decides which of log4j 1.2's own appenders an
// application's packaged log4j configuration leaves out, so that CHA does
// not dispatch into their classes (FRAMEWORKS:F2, l1).
//
// log4j 1.2 creates appenders reflectively from its configuration file
// (PropertyConfigurator / DOMConfigurator → OptionConverter
// #instantiateByClassName), outside the call graph's models. Every call
// log4j makes on an appender — Appender#doAppend from
// AppenderAttachableImpl, OptionHandler#activateOptions from
// PropertySetter, AppenderSkeleton's own this.append(...) — is therefore
// expanded by CHA (and RTA's fallback) to every appender on the class
// path: JDBCAppender's SQL, JMSAppender's JNDI lookup, SocketAppender's
// socket. Measured on JPetStore (a ConsoleAppender-only configuration):
// three false REACHABLE verdicts (testdata/graphbench/RESULTS.MD).
//
// The catalog (Appenders, Version) lists the concrete appenders log4j
// 1.2.8–1.2.17 ships. One of them is on when the packaged configuration
// names it (log4j.properties `log4j.appender.<name>=<class>`, log4j.xml
// `<appender class="...">` — every such file at a class-path root of the
// application or of a dependency jar counts, so a file log4j would not
// pick still turns its appenders on) or when application code references
// it (a call, `new` or static field read). Every other catalog appender
// is off: not a CHA/RTA target. Direct calls are untouched.
//
// The model decides nothing — every appender stays on — unless log4j 1.2
// is in the input and a packaged configuration file is found, because the
// configuration may come from elsewhere (-Dlog4j.configuration, a file
// outside the archive). It also decides nothing when that configuration
// is not fully readable: a class name with a ${...} placeholder, an XML
// file that does not parse, web.xml's log4jConfigLocation (Spring's
// Log4jConfigListener) pointing at anything but a class-path root file it
// read, or application code calling a configurator itself
// (PropertyConfigurator/DOMConfigurator configure*, Spring's
// Log4jConfigurer#initLogging), which may load any other file.
package log4jconfig

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"ariadne/internal/appref"
	"ariadne/internal/hierarchy"
)

// Version versions the appender catalog and the rules above. It is part
// of a snapshot's algorithm versions.
const Version = "l1-1"

// Appenders are log4j 1.2's own concrete appenders (checked in the
// bytecode of 1.2.8, 1.2.9, 1.2.12, 1.2.14, 1.2.15, 1.2.16 and 1.2.17:
// every class implementing org.apache.log4j.Appender; RewriteAppender
// only in 1.2.17). AppenderSkeleton is abstract and never a runtime type.
var Appenders = []string{
	"org.apache.log4j.AsyncAppender",
	"org.apache.log4j.ConsoleAppender",
	"org.apache.log4j.DailyRollingFileAppender",
	"org.apache.log4j.FileAppender",
	"org.apache.log4j.RollingFileAppender",
	"org.apache.log4j.WriterAppender",
	"org.apache.log4j.jdbc.JDBCAppender",
	"org.apache.log4j.lf5.LF5Appender",
	"org.apache.log4j.net.JMSAppender",
	"org.apache.log4j.net.SMTPAppender",
	"org.apache.log4j.net.SocketAppender",
	"org.apache.log4j.net.SocketHubAppender",
	"org.apache.log4j.net.SyslogAppender",
	"org.apache.log4j.net.TelnetAppender",
	"org.apache.log4j.nt.NTEventLogAppender",
	"org.apache.log4j.rewrite.RewriteAppender",
	"org.apache.log4j.varia.ExternallyRolledFileAppender",
	"org.apache.log4j.varia.NullAppender",
}

// marker is a class only log4j 1.2 itself declares (not the
// log4j-over-slf4j bridge, whose AppenderSkeleton runs nothing): the
// model applies only when it is in the input.
const marker = "org.apache.log4j.helpers.AppenderAttachableImpl"

// configurators are the calls through which application code can load a
// configuration of its own choosing.
var configurators = []string{
	"org.apache.log4j.PropertyConfigurator#configure",
	"org.apache.log4j.PropertyConfigurator#configureAndWatch",
	"org.apache.log4j.xml.DOMConfigurator#configure",
	"org.apache.log4j.xml.DOMConfigurator#configureAndWatch",
	"org.springframework.util.Log4jConfigurer#initLogging",
}

// Config is what the packaged configuration says.
type Config struct {
	// Files are the configuration files read, as "<jar>!/<entry>".
	Files []string
	// Classes maps every appender class a file names to the first file
	// naming it.
	Classes map[string]string
	// Unreadable explains why the configuration cannot be relied on (a
	// placeholder, an unparsable file, an unread log4jConfigLocation);
	// empty when it can.
	Unreadable []string
}

// Jar is a jar to look in, with the name to report it by.
type Jar = appref.Jar

// ReadConfig reads log4j.properties and log4j.xml at the root of every
// jar (the application's classes first), plus the class-path file a
// web.xml log4jConfigLocation names.
func ReadConfig(jars []Jar, webXMLs map[string][]byte) (Config, error) {
	c := Config{Classes: map[string]string{}}
	names := map[string]bool{"log4j.properties": true, "log4j.xml": true}
	var locations []string
	for _, module := range sortedKeys(webXMLs) {
		for _, loc := range configLocations(webXMLs[module]) {
			if rest, ok := strings.CutPrefix(loc, "classpath:"); ok && !strings.ContainsAny(rest, "/*${") {
				names[rest] = true
				locations = append(locations, rest)
				continue
			}
			c.Unreadable = append(c.Unreadable, fmt.Sprintf("web.xml log4jConfigLocation %q is not a class-path root file", loc))
		}
	}
	found := map[string]bool{}
	for _, j := range jars {
		if err := c.readJar(j, names, found); err != nil {
			return c, err
		}
	}
	for _, loc := range locations {
		if !found[loc] {
			c.Unreadable = append(c.Unreadable, fmt.Sprintf("web.xml log4jConfigLocation classpath:%s not found in the input", loc))
		}
	}
	return c, nil
}

func (c *Config) readJar(j Jar, names, found map[string]bool) error {
	r, err := zip.OpenReader(j.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	var files []*zip.File
	for _, f := range r.File {
		if names[f.Name] {
			files = append(files, f)
		}
	}
	sort.Slice(files, func(a, b int) bool { return files[a].Name < files[b].Name })
	for _, f := range files {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(rc, 4<<20))
		rc.Close()
		if err != nil {
			return err
		}
		where := j.Name + "!/" + f.Name
		found[f.Name] = true
		c.Files = append(c.Files, where)
		add := func(class string) {
			class = strings.TrimSpace(class)
			if strings.Contains(class, "${") {
				c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: appender class %q is a placeholder", where, class))
				return
			}
			if _, ok := c.Classes[class]; !ok && class != "" {
				c.Classes[class] = where
			}
		}
		if strings.HasSuffix(f.Name, ".xml") {
			if err := xmlAppenders(data, add); err != nil {
				c.Unreadable = append(c.Unreadable, fmt.Sprintf("%s: %v", where, err))
			}
		} else {
			propertiesAppenders(data, add)
		}
	}
	return nil
}

// propertiesAppenders reports the value of every log4j.appender.<name>
// key (the appender's class; deeper keys are its options). Continuation
// lines and escapes are not interpreted.
func propertiesAppenders(data []byte, add func(string)) {
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' || line[0] == '!' {
			continue
		}
		i := strings.IndexAny(line, "=:")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		if name, ok := strings.CutPrefix(key, "log4j.appender."); ok && name != "" && !strings.Contains(name, ".") {
			add(line[i+1:])
		}
	}
}

// xmlAppenders reports the class attribute of every <appender> element
// (with or without the log4j: prefix). The DOCTYPE is not resolved.
func xmlAppenders(data []byte, add func(string)) error {
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("not parsable XML: %v", err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "appender" {
			for _, a := range se.Attr {
				if a.Name.Local == "class" {
					add(a.Value)
				}
			}
		}
	}
}

// configLocations returns the log4jConfigLocation context-param values
// of a web.xml (comma/whitespace separated, as Spring splits them).
func configLocations(webXML []byte) []string {
	var doc struct {
		Params []struct {
			Name  string `xml:"param-name"`
			Value string `xml:"param-value"`
		} `xml:"context-param"`
	}
	d := xml.NewDecoder(bytes.NewReader(webXML))
	d.Strict = false
	if d.Decode(&doc) != nil {
		return nil
	}
	var out []string
	for _, p := range doc.Params {
		if strings.TrimSpace(p.Name) != "log4jConfigLocation" {
			continue
		}
		out = append(out, strings.FieldsFunc(p.Value, func(r rune) bool {
			return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
		})...)
	}
	return out
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
	States  []State
	off     map[string]bool
}

// Decide applies the rules of the package comment to a hierarchy and the
// configuration read from the same input.
func Decide(h *hierarchy.Hierarchy, c Config) Result {
	if _, ok := h.Lookup(marker); !ok {
		return Result{Why: "log4j 1.2 not in the input"}
	}
	if len(c.Files) == 0 {
		return Result{Why: "no packaged log4j.properties/log4j.xml (the configuration may be external)"}
	}
	if len(c.Unreadable) > 0 {
		return Result{Why: strings.Join(c.Unreadable, "; ")}
	}
	refs := appref.Collect(h)
	for _, cfg := range configurators {
		if by, ok := refs.Member(cfg); ok {
			return Result{Why: fmt.Sprintf("application code loads a configuration itself (%s calls %s)", by, cfg)}
		}
	}
	r := Result{Applies: true, off: map[string]bool{}}
	for _, class := range Appenders {
		if _, ok := h.Lookup(class); !ok {
			continue
		}
		st := State{Class: class}
		if file, ok := c.Classes[class]; ok {
			st.On, st.Why = true, "configured in "+file
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

// Off lists the excluded appenders, in catalog order, by simple name.
func (r Result) Off() []string {
	var out []string
	for _, st := range r.States {
		if !st.On {
			out = append(out, st.Class[strings.LastIndex(st.Class, ".")+1:])
		}
	}
	return out
}

// On lists the appenders left on, in catalog order, by simple name.
func (r Result) On() []string {
	var out []string
	for _, st := range r.States {
		if st.On {
			out = append(out, st.Class[strings.LastIndex(st.Class, ".")+1:])
		}
	}
	return out
}

func sortedKeys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
