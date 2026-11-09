// Package bootfeatures decides which optional features of an embedded
// Undertow server a Spring Boot application leaves off, so that CHA does
// not dispatch into their classes.
//
// CHA (and RTA, whose candidates are CHA's) expands a call inside the
// container to every implementation on the class path: Undertow's
// `AbstractFramedStreamSourceChannel#close` reaches every
// `ChannelListener` (AJP, HTTP/2, PROXY protocol, the HTTP client's),
// `HttpServerExchange#endExchange` reaches `Http2ServerConnection`, a
// stream write reaches the TLS and write-timeout conduits. Each of these
// classes is created somewhere in Undertow by code CHA also considers
// reachable, so a creator gate does not remove them; what decides whether
// they run is configuration. Spring Boot's Undertow factory enables
// HTTP/2 only with `server.http2.enabled=true`, TLS only with `server.ssl.*`
// (and `server.ssl.enabled` not false), never configures AJP, the PROXY
// protocol, Undertow's own authentication mechanisms or its HTTP client,
// and sets a socket write timeout only through `server.undertow.options`.
//
// The catalog (Version) lists, per feature, the classes it owns and what
// turns it on: a property in the packaged `application*.properties` /
// `application*.yml` (every profile counts — a feature any of them enables
// is on), or a reference in an application class (a call, `new` or
// static field read of a listed symbol). A feature nothing turns on is
// off: its classes are not CHA/RTA targets. A switch whose value is a
// placeholder (`server.ssl.enabled=${WEBGOAT_SSLENABLED:false}`) is the
// environment's to decide, so that feature stays on; configuration outside the packaged
// files (environment variables, command line, an external config
// directory) is not read — the same limit as any packaged-input analysis.
// Direct calls to a feature class are untouched. The model applies when
// undertow-core is in the input together with Spring Boot's Undertow
// factory (the code that configures Undertow from those properties) or
// the application is a Spring Boot application (an application class
// carries @SpringBootApplication/@EnableAutoConfiguration) — a launcher
// outside the analyzed input (WebGoat's) leaves only the factory to go by.
package bootfeatures

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Version versions the catalog and the enabling rules.
const Version = "u1"

// Feature is one optional Undertow feature.
type Feature struct {
	Name string
	// Prefixes are class-name prefixes the feature owns; Classes are
	// single classes (an inner class included by prefix "C$").
	Prefixes, Classes []string
	// Enabled reports whether the packaged configuration turns it on.
	Enabled func(c Config) bool
	// Symbols: "owner#member" — an application class calling, creating
	// (member "<init>") or reading the static field member of owner turns
	// the feature on. An owner alone ("owner#") matches any reference.
	Symbols []string
	// Switches: the properties that turn it on or off; a placeholder value
	// (`${VAR:default}`) leaves the decision to the environment, so the
	// feature counts as on.
	Switches []string
}

var catalog = []Feature{
	{Name: "http2", Prefixes: []string{"io.undertow.server.protocol.http2.", "io.undertow.protocols.http2.", "io.undertow.client.http2."},
		Enabled: func(c Config) bool { return c.isTrue("server.http2.enabled") },
		Symbols: []string{"io.undertow.UndertowOptions#ENABLE_HTTP2"}, Switches: []string{"server.http2.enabled"}},
	{Name: "tls", Prefixes: []string{"io.undertow.protocols.ssl."},
		Enabled: func(c Config) bool { return c.hasPrefix("server.ssl.") && !c.isFalse("server.ssl.enabled") },
		Symbols: []string{"io.undertow.Undertow$Builder#addHttpsListener"}, Switches: []string{"server.ssl.enabled"}},
	{Name: "ajp", Prefixes: []string{"io.undertow.server.protocol.ajp.", "io.undertow.client.ajp."},
		Symbols: []string{"io.undertow.Undertow$Builder#addAjpListener", "io.undertow.Undertow$ListenerType#AJP"}},
	{Name: "proxy-protocol", Prefixes: []string{"io.undertow.server.protocol.proxy."},
		Symbols: []string{"io.undertow.Undertow$ListenerBuilder#setUseProxyProtocol"}},
	{Name: "write-timeout", Classes: []string{"io.undertow.conduits.WriteTimeoutStreamSinkConduit"},
		Enabled: func(c Config) bool { return c.keyContains("server.undertow.options.", "write_timeout") },
		Symbols: []string{"org.xnio.Options#WRITE_TIMEOUT"}},
	{Name: "undertow-authentication", Classes: []string{
		"io.undertow.security.impl.FormAuthenticationMechanism", "io.undertow.security.impl.BasicAuthenticationMechanism",
		"io.undertow.security.impl.DigestAuthenticationMechanism", "io.undertow.security.impl.ClientCertAuthenticationMechanism",
		"io.undertow.security.impl.GSSAPIAuthenticationMechanism", "io.undertow.security.impl.ExternalAuthenticationMechanism",
		"io.undertow.security.impl.GenericHeaderAuthenticationMechanism", "io.undertow.security.impl.SingleSignOnAuthenticationMechanism",
		"io.undertow.servlet.handlers.security.ServletFormAuthenticationMechanism", "io.undertow.servlet.handlers.security.ServletSingleSignOnAuthenticationMechanism"},
		Enabled: func(c Config) bool { return c.WebXMLLoginConfig },
		Symbols: []string{"io.undertow.servlet.api.LoginConfig#", "io.undertow.servlet.api.DeploymentInfo#setLoginConfig"}},
	{Name: "http-client", Prefixes: []string{"io.undertow.client."},
		Symbols: []string{"io.undertow.client.", "io.undertow.server.handlers.proxy."}},
}

// Config is the packaged configuration the model reads.
type Config struct {
	// Values maps a property key (lower case; YAML flattened to dotted
	// keys) to every value it has across the files.
	Values map[string][]string
	// Overridable: keys whose value is a placeholder with a default.
	Overridable map[string]bool
	// Files read, in order.
	Files []string
	// WebXMLLoginConfig: a web.xml declares a <login-config>.
	WebXMLLoginConfig bool
}

func (c Config) isTrue(key string) bool {
	for _, v := range c.Values[key] {
		if strings.EqualFold(v, "true") {
			return true
		}
	}
	return false
}

func (c Config) isFalse(key string) bool {
	vs := c.Values[key]
	if len(vs) == 0 {
		return false
	}
	for _, v := range vs {
		if !strings.EqualFold(v, "false") {
			return false
		}
	}
	return true
}

func (c Config) hasPrefix(prefix string) bool {
	for k := range c.Values {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

func (c Config) keyContains(prefix, part string) bool {
	for k := range c.Values {
		if strings.HasPrefix(k, prefix) && strings.Contains(strings.ReplaceAll(k, "-", "_"), part) {
			return true
		}
	}
	return false
}

var configName = regexp.MustCompile(`^(BOOT-INF/classes/|WEB-INF/classes/)?(config/)?application(-[^/]+)?\.(properties|ya?ml)$`)

// ReadConfig reads the Spring Boot configuration files packaged in the
// application jar, and whether any web.xml declares a login-config.
func ReadConfig(appJar string, webXMLs map[string][]byte) (Config, error) {
	c := Config{Values: map[string][]string{}, Overridable: map[string]bool{}}
	for _, x := range webXMLs {
		if bytes.Contains(x, []byte("<login-config")) {
			c.WebXMLLoginConfig = true
		}
	}
	r, err := zip.OpenReader(appJar)
	if err != nil {
		return c, err
	}
	defer r.Close()
	var files []*zip.File
	for _, f := range r.File {
		if configName.MatchString(f.Name) {
			files = append(files, f)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	for _, f := range files {
		rc, err := f.Open()
		if err != nil {
			return c, err
		}
		data, err := io.ReadAll(io.LimitReader(rc, 4<<20))
		rc.Close()
		if err != nil {
			return c, err
		}
		c.Files = append(c.Files, f.Name)
		if ext := path.Ext(f.Name); ext == ".properties" {
			parseProperties(data, c.add)
		} else {
			parseYAML(data, c.add)
		}
	}
	return c, nil
}

var placeholder = regexp.MustCompile(`^\$\{[^:}]+:([^}]*)\}$`)

func (c *Config) add(key, value string) {
	key = strings.ToLower(strings.TrimSpace(key))
	value = strings.TrimSpace(value)
	if m := placeholder.FindStringSubmatch(value); m != nil {
		value = m[1]
		c.Overridable[key] = true
	}
	c.Values[key] = append(c.Values[key], value)
}

func parseProperties(data []byte, add func(k, v string)) {
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' || line[0] == '!' {
			continue
		}
		i := strings.IndexAny(line, "=:")
		if i < 0 {
			add(line, "")
			continue
		}
		add(line[:i], line[i+1:])
	}
}

// parseYAML flattens the block-mapping subset Spring configuration uses
// (indented "key:" / "key: value" lines, "---" document separators);
// lists, flow style and multi-line scalars are skipped.
func parseYAML(data []byte, add func(k, v string)) {
	type level struct {
		indent int
		key    string
	}
	var stack []level
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		raw := sc.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "---" {
			stack = nil
			continue
		}
		if trimmed == "" || trimmed[0] == '#' || trimmed[0] == '-' {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		i := strings.Index(trimmed, ":")
		if i <= 0 {
			continue
		}
		key, value := strings.Trim(trimmed[:i], `"'`), strings.TrimSpace(trimmed[i+1:])
		if j := strings.Index(value, " #"); j >= 0 {
			value = strings.TrimSpace(value[:j])
		}
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		full := key
		if len(stack) > 0 {
			full = stack[len(stack)-1].key + "." + key
		}
		if value == "" {
			stack = append(stack, level{indent, full})
			continue
		}
		add(full, strings.Trim(value, `"'`))
	}
}

// State is one feature's decision.
type State struct {
	Feature string
	On      bool
	// Why: the property or application class that turned it on, or why it
	// is off.
	Why string
	// Overridable: on only because its switch is a placeholder the
	// environment may set.
	Overridable bool
}

// Result is the model's decision for an input.
type Result struct {
	// Applies: a Spring Boot application with undertow-core in the input.
	Applies bool
	States  []State
	off     []Feature
}

// bootUndertowFactories configure an embedded Undertow from Spring Boot's
// server.* properties (Boot 2.x and 3.x).
var bootUndertowFactories = []string{
	"org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory",
	"org.springframework.boot.web.embedded.undertow.UndertowReactiveWebServerFactory",
}

// Decide applies the catalog to the hierarchy and configuration. boot
// reports whether the input is a Spring Boot application.
func Decide(h *hierarchy.Hierarchy, c Config, boot bool) Result {
	var r Result
	if _, ok := h.Lookup("io.undertow.Undertow"); !ok {
		return r
	}
	factory := false
	for _, f := range bootUndertowFactories {
		if _, ok := h.Lookup(f); ok {
			factory = true
		}
	}
	if !boot && !factory {
		return r
	}
	r.Applies = true
	refs := appReferences(h)
	for _, f := range catalog {
		st := State{Feature: f.Name}
		if f.Enabled != nil && f.Enabled(c) {
			st.On, st.Why = true, "packaged configuration"
		}
		if !st.On {
			for _, s := range f.Symbols {
				if by, ok := refs.match(s); ok {
					st.On, st.Why = true, "referenced by "+by
					break
				}
			}
		}
		if !st.On {
			// A switch whose packaged value is a placeholder default is the
			// environment's to decide: the feature stays on (a negative
			// result must not rest on a value the deployment may change).
			for _, k := range f.Switches {
				if c.Overridable[k] {
					st.On, st.Overridable = true, true
					st.Why = k + " is a placeholder the environment may set"
					break
				}
			}
		}
		if !st.On {
			st.Why = "not enabled by the packaged configuration or application code"
			r.off = append(r.off, f)
		}
		r.States = append(r.States, st)
	}
	return r
}

// Excluded reports whether class belongs to a feature that is off.
func (r Result) Excluded(class string) bool {
	for _, f := range r.off {
		for _, p := range f.Prefixes {
			if strings.HasPrefix(class, p) {
				return true
			}
		}
		for _, c := range f.Classes {
			if class == c || strings.HasPrefix(class, c+"$") {
				return true
			}
		}
	}
	return false
}

// Off names the features that are off, in catalog order.
func (r Result) Off() []string {
	var out []string
	for _, f := range r.off {
		out = append(out, f.Name)
	}
	return out
}

type references map[string]string // "owner#member" -> first referencing class

func (rf references) match(symbol string) (string, bool) {
	owner, member, _ := strings.Cut(symbol, "#")
	if member != "" {
		by, ok := rf[symbol]
		return by, ok
	}
	// An owner or a package prefix: any member of a matching owner.
	var best string
	for k, by := range rf {
		o, _, _ := strings.Cut(k, "#")
		if o == owner || (strings.HasSuffix(owner, ".") && strings.HasPrefix(o, owner)) {
			if best == "" || by < best {
				best = by
			}
		}
	}
	return best, best != ""
}

// appReferences collects every symbol application classes call, create or
// read statically.
func appReferences(h *hierarchy.Hierarchy) references {
	out := references{}
	note := func(k, by string) {
		if prev, ok := out[k]; !ok || by < prev {
			out[k] = by
		}
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		if !ci.IsApp {
			continue
		}
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				note(c.Owner+"#"+c.Name, fqcn)
			}
			for _, ref := range m.InitRefs {
				switch ref.Kind {
				case classfile.InitNew:
					note(ref.Owner+"#<init>", fqcn)
				case classfile.InitGetStatic:
					note(ref.Owner+"#"+ref.Name, fqcn)
				}
			}
		}
	}
	return out
}
