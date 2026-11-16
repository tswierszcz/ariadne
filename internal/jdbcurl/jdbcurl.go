// Package jdbcurl narrows the JDBC drivers a connection can run to those
// whose URL prefix matches a URL the application can use (model
// jdbc:drivers).
//
// DriverManager#getConnection(url) offers the URL to every registered
// driver in turn, calling its connect; a driver that does not accept the
// URL checks it first thing and returns null. The service-provider
// dispatch model links java.sql.Driver#connect to every driver registered
// in META-INF/services, and the JDBC objects model then reaches every
// such driver's connection, statements and result sets — so an
// application on HSQLDB "reaches" pgjdbc's and MySQL's code because their
// jars are on the class path.
//
// For a driver in Catalog (verified on the real JVM, scripts/
// jdbcdrivers-oracle.sh), when no URL the application can use starts with
// one of its prefixes, the model cuts its connect after the URL check
// (callgraph.Graph.CutSitesAfter): the check still runs, nothing after it
// does. The URLs come from the application jar itself — string constants
// of its classes, its configuration files (every Spring profile's, a
// web.xml) — plus the embedded databases Spring can start without a URL
// (H2, HSQLDB, Derby) when Spring's embedded-database support is in the
// input. The model assumes that packaged configuration is what runs: a
// URL set only from outside (an environment variable, a JNDI DataSource)
// cannot be seen, so a placeholder with no packaged value, a JNDI lookup
// or a URL with no literal "jdbc:<subprotocol>:" prefix turns the model
// off for the run (every driver kept).
package jdbcurl

import (
	"archive/zip"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// ModelID identifies the model in logs and snapshots.
const ModelID = "jdbc:drivers"

// Version versions the model's semantics (catalog, URL sources).
const Version = "u1"

// Driver is one catalog entry: a driver class, the URL prefixes it
// accepts, and the calls its connect makes to check the URL (the first
// one found is the cut point).
type Driver struct {
	Class    string
	Prefixes []string
	Guards   []string // "owner#name"
}

const connectDesc = "(Ljava/lang/String;Ljava/util/Properties;)Ljava/sql/Connection;"

// Catalog lists the drivers whose connect was checked to return null
// right after its URL check for a URL it does not accept (bytecode of the
// listed versions, and the JVM oracle).
var Catalog = []Driver{
	// pgjdbc 42.0–42.7: if (!url.startsWith("jdbc:postgresql:")) return null.
	{Class: "org.postgresql.Driver", Prefixes: []string{"jdbc:postgresql:"}, Guards: []string{"java.lang.String#startsWith"}},
	// MySQL Connector/J 8.0.23+ (com.mysql.cj.jdbc.Driver extends it):
	// if (!ConnectionUrl.acceptsUrl(url)) return null. 8.0.11 parses
	// first: no guard, no cut.
	{Class: "com.mysql.cj.jdbc.NonRegisteringDriver", Prefixes: []string{"jdbc:mysql", "mysqlx:"}, Guards: []string{"com.mysql.cj.conf.ConnectionUrl#acceptsUrl"}},
	// H2 1.4: if (!acceptsURL(url)) return null; 2.x: url.startsWith(...).
	{Class: "org.h2.Driver", Prefixes: []string{"jdbc:h2:", "jdbc:default:connection"}, Guards: []string{"org.h2.Driver#acceptsURL", "java.lang.String#startsWith"}},
}

// embeddedPrefixes are the databases Spring starts without a configured
// URL (EmbeddedDatabaseConnection, EmbeddedDatabaseBuilder).
var embeddedPrefixes = []string{"jdbc:h2:", "jdbc:hsqldb:", "jdbc:derby:"}

var embeddedSupport = []string{
	"org.springframework.boot.jdbc.EmbeddedDatabaseConnection",
	"org.springframework.boot.autoconfigure.jdbc.EmbeddedDatabaseConnection",
	"org.springframework.jdbc.datasource.embedded.EmbeddedDatabaseFactory",
}

// URLs is what the application jar says about the JDBC URLs it can use.
type URLs struct {
	// Prefixes are the literal beginnings of the URLs found (up to a
	// placeholder), each with the first place it was found.
	Prefixes map[string]string
	// Unbounded explains why the URLs cannot be bounded; empty when they
	// can.
	Unbounded []string
}

var (
	urlInText = regexp.MustCompile(`jdbc:[A-Za-z0-9+_.-]*(:[A-Za-z0-9+_.-]*)?`)
	// A URL-valued setting whose value is a placeholder: url=${KEY},
	// url: ${KEY}, name="url" value="${KEY}", jdbcUrl="${KEY}".
	urlPlaceholder = regexp.MustCompile(`(?i)(?:url|jdbc-url)["']?\s*(?:[=:]|value\s*=)\s*["']?\$\{([^}:]+)(:[^}]*)?\}`)
	jndiMarkers    = []string{"jndi-name", "jee:jndi-lookup", "JndiObjectFactoryBean", "java:comp/env/jdbc"}
)

// Collect reads the application jar's string constants and text files
// (and the web.xml files given) for JDBC URLs. h decides the embedded
// databases.
func Collect(appJar string, webXMLs map[string][]byte, h *hierarchy.Hierarchy) (URLs, error) {
	u := URLs{Prefixes: map[string]string{}}
	keys := map[string]bool{}
	var placeholders []placeholder
	text := func(where string, data []byte) {
		s := string(data)
		for _, m := range urlInText.FindAllString(s, -1) {
			u.add(m, where)
		}
		for _, m := range urlPlaceholder.FindAllStringSubmatch(s, -1) {
			placeholders = append(placeholders, placeholder{key: m[1], hasDefault: m[2] != "", where: where})
		}
		for _, j := range jndiMarkers {
			if strings.Contains(s, j) {
				u.Unbounded = append(u.Unbounded, where+": JNDI DataSource ("+j+")")
			}
		}
		switch path.Ext(where) {
		case ".properties":
			propertiesKeys(s, keys)
		case ".yml", ".yaml":
			yamlKeys(s, keys)
		}
	}
	r, err := zip.OpenReader(appJar)
	if err != nil {
		return u, err
	}
	defer r.Close()
	for _, f := range r.File {
		if f.FileInfo().IsDir() || f.UncompressedSize64 > 4<<20 {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		if strings.HasSuffix(f.Name, ".class") {
			cf, err := classfile.Parse(strings.NewReader(string(data)))
			if err != nil {
				continue
			}
			for _, s := range cf.Strings() {
				if strings.HasPrefix(s, "jdbc:") {
					u.add(literalPrefix(s), f.Name)
				}
				for _, j := range jndiMarkers {
					if strings.Contains(s, j) {
						u.Unbounded = append(u.Unbounded, f.Name+": JNDI DataSource ("+j+")")
					}
				}
			}
			continue
		}
		switch path.Ext(f.Name) {
		case ".properties", ".yml", ".yaml", ".xml", ".conf", ".json":
			text(f.Name, data)
		}
	}
	for _, name := range sortedKeys(webXMLs) {
		text(name, webXMLs[name])
	}
	for _, p := range placeholders {
		if !p.hasDefault && !keys[p.key] {
			u.Unbounded = append(u.Unbounded, p.where+": URL from ${"+p.key+"}, not set in the packaged configuration")
		}
	}
	if h != nil {
		for _, c := range embeddedSupport {
			if _, ok := h.Lookup(c); ok {
				for _, p := range embeddedPrefixes {
					u.add(p, c+" (embedded database without a URL)")
				}
				break
			}
		}
	}
	sort.Strings(u.Unbounded)
	return u, nil
}

type placeholder struct {
	key        string
	hasDefault bool
	where      string
}

// add records a URL's literal beginning (a URL text is cut at its first
// placeholder).
func (u *URLs) add(url, where string) {
	url = literalPrefix(url)
	if _, ok := u.Prefixes[url]; !ok {
		u.Prefixes[url] = where
	}
}

func literalPrefix(s string) string {
	if i := strings.Index(s, "${"); i >= 0 {
		s = s[:i]
	}
	return s
}

// accepts reports whether a URL beginning with lit can be one prefix
// accepts: either starts with the other (a literal shorter than the
// driver's prefix, e.g. "jdbc:", may still become it).
func accepts(lit, prefix string) bool {
	return strings.HasPrefix(lit, prefix) || strings.HasPrefix(prefix, lit)
}

func propertiesKeys(s string, keys map[string]bool) {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' || line[0] == '!' {
			continue
		}
		if i := strings.IndexAny(line, "=:"); i > 0 {
			keys[strings.TrimSpace(line[:i])] = true
		}
	}
}

// yamlKeys records the dotted keys of a block-style YAML file (nesting by
// indentation; enough for Spring configuration files).
func yamlKeys(s string, keys map[string]bool) {
	type level struct {
		indent int
		key    string
	}
	var stack []level
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed[0] == '#' || trimmed == "---" || strings.HasPrefix(trimmed, "- ") {
			if trimmed == "---" {
				stack = nil
			}
			continue
		}
		i := strings.Index(trimmed, ":")
		if i <= 0 {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		key := strings.Trim(trimmed[:i], `"' `)
		full := key
		if len(stack) > 0 {
			full = stack[len(stack)-1].key + "." + key
		}
		keys[full] = true
		stack = append(stack, level{indent, full})
	}
}

// Cut is one driver connect the model cut after its URL check.
type Cut struct {
	Driver string // the catalog class
	Method string // the graph key of its connect
	PC     int    // the URL check's call site; later sites are cut
}

// Decision is the model's outcome for one run.
type Decision struct {
	// Applies is false when the URLs cannot be bounded (Why says why) or
	// no catalog driver is in the input.
	Applies bool
	Why     string
	// Kept are the catalog drivers in the input some URL can reach;
	// Cuts the others, cut after their URL check; Unguarded the drivers
	// whose connect has none of the catalog's checks (left as they are).
	Kept, Unguarded []string
	Cuts            []Cut
}

// Decide chooses, for every catalog driver in h, whether some URL in u
// can reach it; it changes nothing.
func Decide(h *hierarchy.Hierarchy, u URLs) Decision {
	var d Decision
	present := false
	for _, drv := range Catalog {
		if _, ok := h.Lookup(drv.Class); ok {
			present = true
		}
	}
	switch {
	case !present:
		d.Why = "no catalog driver in the input"
		return d
	case len(u.Unbounded) > 0:
		d.Why = "URLs not bounded: " + strings.Join(u.Unbounded, "; ")
		return d
	}
	d.Applies = true
	for _, drv := range Catalog {
		ci, ok := h.Lookup(drv.Class)
		if !ok {
			continue
		}
		reached := false
		for lit := range u.Prefixes {
			for _, p := range drv.Prefixes {
				if accepts(lit, p) {
					reached = true
				}
			}
		}
		if reached {
			d.Kept = append(d.Kept, drv.Class)
			continue
		}
		var m *hierarchy.Method
		for i := range ci.Methods {
			if ci.Methods[i].Name == "connect" && ci.Methods[i].Descriptor == connectDesc {
				m = &ci.Methods[i]
			}
		}
		pc := -1
		if m != nil {
			for _, c := range m.Calls {
				for _, gd := range drv.Guards {
					if c.Owner+"#"+c.Name == gd && (pc < 0 || c.PC < pc) {
						pc = c.PC
					}
				}
			}
		}
		if pc < 0 {
			d.Unguarded = append(d.Unguarded, drv.Class)
			continue
		}
		d.Cuts = append(d.Cuts, Cut{Driver: drv.Class, Method: drv.Class + "#connect" + connectDesc, PC: pc})
	}
	return d
}

// Apply cuts every connect the decision chose in g; it returns how many
// call sites were removed.
func (d Decision) Apply(g *callgraph.Graph) int {
	n := 0
	for _, c := range d.Cuts {
		if id, ok := g.IDOfKey(c.Method); ok {
			s, _ := g.CutSitesAfter(id, c.PC)
			n += s
		}
	}
	return n
}

func sortedKeys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
