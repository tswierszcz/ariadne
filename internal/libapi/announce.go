package libapi

import (
	"archive/zip"
	"bufio"
	"io"
	"path"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Announcement is one class a library announces to a framework in its
// own jar (LIBRARY_MODE LB2, lb2-5): the framework creates it and calls
// into it, whatever its access.
type Announcement struct {
	File  string `json:"file"`  // the jar entry, e.g. META-INF/spring.factories
	Key   string `json:"key"`   // the factory type or imports name
	Class string `json:"class"` // the announced class
}

// SpringAnnouncement reports an announcement read from Spring's own
// files (spring.factories, META-INF/spring/*.imports), not a CDI bean.
func SpringAnnouncement(a Announcement) bool {
	return a.File == springFactories || strings.HasPrefix(a.File, springImportsDir)
}

const (
	springFactories     = "META-INF/spring.factories"
	springImportsDir    = "META-INF/spring/"
	enableAutoConfigKey = "org.springframework.boot.autoconfigure.EnableAutoConfiguration"
	autoConfigImports   = "org.springframework.boot.autoconfigure.AutoConfiguration"
)

// Announcements reads the Spring entries a library jar announces:
// every class of every key in META-INF/spring.factories, and every class
// listed in a META-INF/spring/*.imports file (Spring Boot 2.7+; the key is
// the file name without .imports). Sorted by file, key, class.
func Announcements(jar string) ([]Announcement, error) {
	zr, err := zip.OpenReader(jar)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var out []Announcement
	for _, f := range zr.File {
		isFactories := f.Name == springFactories
		isImports := strings.HasPrefix(f.Name, springImportsDir) && strings.HasSuffix(f.Name, ".imports") && !strings.Contains(strings.TrimPrefix(f.Name, springImportsDir), "/")
		if !isFactories && !isImports {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		if isFactories {
			for key, classes := range parseFactories(string(b)) {
				for _, c := range classes {
					out = append(out, Announcement{File: f.Name, Key: key, Class: c})
				}
			}
			continue
		}
		key := strings.TrimSuffix(path.Base(f.Name), ".imports")
		sc := bufio.NewScanner(strings.NewReader(string(b)))
		for sc.Scan() {
			line := sc.Text()
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = line[:i]
			}
			if line = strings.TrimSpace(line); line != "" {
				out = append(out, Announcement{File: f.Name, Key: key, Class: line})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Key != b.Key {
			return a.Key < b.Key
		}
		return a.Class < b.Class
	})
	return out, nil
}

// parseFactories reads a java.util.Properties file of comma-separated
// class lists: comment lines (# or !), key=value or key:value, a trailing
// backslash continuing the line.
func parseFactories(text string) map[string][]string {
	out := map[string][]string{}
	var logical []string
	cur := ""
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line := strings.TrimLeft(raw, " \t\f")
		if cur == "" && (line == "" || line[0] == '#' || line[0] == '!') {
			continue
		}
		if strings.HasSuffix(line, `\`) && !strings.HasSuffix(line, `\\`) {
			cur += strings.TrimSuffix(line, `\`)
			continue
		}
		logical = append(logical, cur+line)
		cur = ""
	}
	if cur != "" {
		logical = append(logical, cur)
	}
	for _, l := range logical {
		i := strings.IndexAny(l, "=:")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(l[:i])
		for _, c := range strings.Split(l[i+1:], ",") {
			if c = strings.TrimSpace(c); c != "" {
				out[key] = append(out[key], c)
			}
		}
	}
	return out
}

var (
	configurationAnnotations = map[string]bool{
		"org.springframework.context.annotation.Configuration":         true,
		"org.springframework.boot.autoconfigure.AutoConfiguration":     true,
		"org.springframework.boot.SpringBootConfiguration":             true,
		"org.springframework.boot.autoconfigure.SpringBootApplication": true,
	}
	beanAnnotation   = "org.springframework.context.annotation.Bean"
	importAnnotation = "org.springframework.context.annotation.Import"
	// Methods the container calls on a configuration class it creates.
	containerMethodAnnotations = map[string]bool{
		beanAnnotation: true,
		"org.springframework.beans.factory.annotation.Autowired": true,
		"org.springframework.beans.factory.annotation.Value":     true,
		"javax.inject.Inject":                             true,
		"jakarta.inject.Inject":                           true,
		"javax.annotation.Resource":                       true,
		"jakarta.annotation.Resource":                     true,
		"javax.annotation.PostConstruct":                  true,
		"jakarta.annotation.PostConstruct":                true,
		"javax.annotation.PreDestroy":                     true,
		"jakarta.annotation.PreDestroy":                   true,
		"org.springframework.context.event.EventListener": true,
	}
	// Container callback interfaces of a configuration class, and the
	// interfaces of an @Import that make it a selector or registrar
	// rather than configuration: their public methods are called.
	callbackInterfaces = map[string]bool{
		"org.springframework.beans.factory.InitializingBean":                            true,
		"org.springframework.beans.factory.DisposableBean":                              true,
		"org.springframework.context.annotation.ImportSelector":                         true,
		"org.springframework.context.annotation.DeferredImportSelector":                 true,
		"org.springframework.context.annotation.ImportBeanDefinitionRegistrar":          true,
		"org.springframework.beans.factory.config.BeanFactoryPostProcessor":             true,
		"org.springframework.beans.factory.support.BeanDefinitionRegistryPostProcessor": true,
		"org.springframework.beans.factory.config.BeanPostProcessor":                    true,
	}
)

func hasAnnotation(anns []string, set map[string]bool) bool {
	for _, a := range anns {
		if set[a] {
			return true
		}
	}
	return false
}

// AnnouncedRoots turns announcements into roots (lb2-5), each with its
// provenance ("<file> <key>: <class>", then " > member <class>" or
// " > @Import <class>" for each step). An entry under EnableAutoConfiguration
// or an imports file, or one annotated as configuration, is a
// configuration class the container processes: its constructors and
// static initializer, its @Bean methods (any access, static too), the
// methods it injects or calls back (@Autowired, @Value, @Inject,
// @Resource, @PostConstruct, @PreDestroy, @EventListener, and the public
// methods of InitializingBean-style callbacks), its member configuration
// classes and its @Import targets, transitively. Any other spring.factories
// entry is a factory the framework instantiates and calls through the
// key's type: its constructors, static initializer and public instance
// methods. Classes outside the input or in an excluded package are left
// out, as providers are. Roots are sorted by key; classes counts the
// classes made entries.
func AnnouncedRoots(g *callgraph.Graph, h *hierarchy.Hierarchy, anns []Announcement, opt Options) (roots []uint32, provenance map[uint32][]string, classes int) {
	sel := newSelector(h, nil, nil, opt)
	provenance = map[uint32][]string{}
	add := func(fqcn, name, desc, prov string) {
		id, ok := g.ID(fqcn, name, desc)
		if !ok {
			return
		}
		if _, seen := provenance[id]; !seen {
			roots = append(roots, id)
		}
		for _, p := range provenance[id] {
			if p == prov {
				return
			}
		}
		provenance[id] = append(provenance[id], prov)
	}
	done, counted := map[string]bool{}, map[string]bool{}
	var visit func(fqcn, prov string, config bool)
	visit = func(fqcn, prov string, config bool) {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp || !sel.inScope(fqcn) || done[fqcn+"\x00"+prov] {
			return
		}
		done[fqcn+"\x00"+prov] = true
		if !counted[fqcn] {
			counted[fqcn] = true
			classes++
		}
		callbacks := implementsAny(h, fqcn, callbackInterfaces)
		config = config || hasAnnotation(ci.Annotations, configurationAnnotations)
		for _, m := range ci.Methods {
			f := m.AccessFlags
			if f&(accSynthetic|accBridge) != 0 {
				continue
			}
			public := f&classfile.AccPublic != 0 && f&accStatic == 0 && m.Name != "<init>"
			switch {
			case m.Name == "<init>" || m.Name == "<clinit>":
				add(fqcn, m.Name, m.Descriptor, prov)
			case config && hasAnnotation(m.Annotations, containerMethodAnnotations):
				add(fqcn, m.Name, m.Descriptor, prov)
			case (!config || callbacks) && public:
				add(fqcn, m.Name, m.Descriptor, prov)
			}
		}
		if !config {
			return
		}
		// Member configuration classes the container registers with
		// their enclosing one.
		prefix := fqcn + "$"
		for _, sub := range h.SortedFQCNs() {
			if !strings.HasPrefix(sub, prefix) || strings.Contains(sub[len(prefix):], "$") {
				continue
			}
			if sci, ok := h.Lookup(sub); ok && hasAnnotation(sci.Annotations, configurationAnnotations) {
				visit(sub, prov+" > member "+sub, true)
			}
		}
		for _, a := range ci.AnnotationDetails {
			if a.Type != importAnnotation {
				continue
			}
			v, ok := a.Value("value")
			if !ok {
				continue
			}
			vals := []classfile.ElementValue{v}
			if v.Kind == classfile.ElementArray {
				vals = v.Array
			}
			for _, x := range vals {
				if x.Kind == classfile.ElementClass {
					visit(x.ClassFQCN, prov+" > @Import "+x.ClassFQCN, true)
				}
			}
		}
	}
	// visitCDI: a CDI bean (CDIAnnouncements). The container constructs
	// it and calls its producers, observers, disposers, initializers,
	// lifecycle callbacks and interceptor methods whatever their access
	// (static producers too); a client calls its public instance methods
	// through the bean's types.
	visitCDI := func(fqcn, prov string) {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp || !sel.inScope(fqcn) {
			return
		}
		if !counted[fqcn] {
			counted[fqcn] = true
			classes++
		}
		for _, m := range ci.Methods {
			f := m.AccessFlags
			if f&(accSynthetic|accBridge) != 0 {
				continue
			}
			public := f&classfile.AccPublic != 0 && f&accStatic == 0
			if m.Name == "<init>" || m.Name == "<clinit>" || cdiEntry(m) || public {
				add(fqcn, m.Name, m.Descriptor, prov)
			}
		}
	}
	for _, a := range anns {
		if a.isCDI() {
			visitCDI(a.Class, a.File+" "+a.Key+": "+a.Class)
			continue
		}
		config := a.Key == enableAutoConfigKey || a.Key == autoConfigImports
		visit(a.Class, a.File+" "+a.Key+": "+a.Class, config)
	}
	sort.Slice(roots, func(i, j int) bool { return g.Key(roots[i]) < g.Key(roots[j]) })
	return roots, provenance, classes
}

// implementsAny reports whether fqcn or a superclass names one of set
// among its interfaces (the interfaces themselves may be outside the
// input: Spring is usually provided).
func implementsAny(h *hierarchy.Hierarchy, fqcn string, set map[string]bool) bool {
	for cur := fqcn; cur != ""; {
		ci, ok := h.Lookup(cur)
		if !ok {
			return false
		}
		for _, i := range ci.Interfaces {
			if set[i] {
				return true
			}
		}
		cur = ci.SuperClass
	}
	return false
}
