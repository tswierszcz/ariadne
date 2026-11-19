package libapi

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strconv"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// CDI bean archives (LIBRARY_MODE LB2): a library jar with
// META-INF/beans.xml is a bean archive — any CDI container (Weld in
// Jakarta EE/SE, ArC in Quarkus) discovers its beans and calls their
// producers, observers, initializers and lifecycle callbacks whatever
// their access, as Spring calls an announced configuration (lb2-5). A jar
// with only a Jandex index (META-INF/jandex.idx) is discovered by Quarkus
// the same way, in "annotated" mode.

const (
	beansXML  = "META-INF/beans.xml"
	jandexIdx = "META-INF/jandex.idx"
	// cdiKeyPrefix starts the Key of a CDI announcement:
	// "bean-discovery-mode=<mode>", then " @<annotation>" in annotated mode.
	cdiKeyPrefix = "bean-discovery-mode="
)

var (
	// beanDefining are the annotations that make a class a bean in an
	// "annotated" archive: the built-in scopes, @Dependent, interceptors
	// and decorators — plus @Singleton, which Quarkus treats as one (a
	// bean that "may run").
	beanDefining = map[string]bool{}
	// cdiMethodAnnotations mark methods the container calls.
	cdiMethodAnnotations = map[string]bool{}
	// cdiParamAnnotations mark a method the container calls by one of its
	// parameters: observers and disposers.
	cdiParamAnnotations = map[string]bool{}
	// Meta-annotations that make a library's own annotation bean-defining.
	cdiMeta = map[string]bool{}
)

func init() {
	for _, ns := range []string{"javax", "jakarta"} {
		for _, a := range []string{"enterprise.context.ApplicationScoped", "enterprise.context.RequestScoped",
			"enterprise.context.SessionScoped", "enterprise.context.ConversationScoped", "enterprise.context.Dependent",
			"inject.Singleton", "interceptor.Interceptor", "decorator.Decorator"} {
			beanDefining[ns+"."+a] = true
		}
		for _, a := range []string{"inject.Inject", "enterprise.inject.Produces", "annotation.PostConstruct",
			"annotation.PreDestroy", "interceptor.AroundInvoke", "interceptor.AroundConstruct"} {
			cdiMethodAnnotations[ns+"."+a] = true
		}
		for _, a := range []string{"enterprise.event.Observes", "enterprise.event.ObservesAsync", "enterprise.inject.Disposes"} {
			cdiParamAnnotations[ns+"."+a] = true
		}
		for _, a := range []string{"enterprise.inject.Stereotype", "enterprise.context.NormalScope", "inject.Scope"} {
			cdiMeta[ns+"."+a] = true
		}
	}
}

// CDIAnnouncements lists the beans of a library jar that is a CDI bean
// archive, as announcements (File: META-INF/beans.xml or
// META-INF/jandex.idx; Key: "bean-discovery-mode=<mode>", plus
// " @<annotation>" for the bean-defining annotation in annotated mode).
// The discovery mode is beans.xml's bean-discovery-mode; without one, an
// empty or pre-4.0 beans.xml means "all" (CDI 1.1-3.0; CDI 4.0 changed the
// default to "annotated", so "all" is what some container may do) and a
// version 4.0+ one "annotated"; "none" discovers nothing. In "all" mode a
// bean is every concrete class with a no-argument or @Inject constructor;
// in "annotated" mode a class with a bean-defining annotation (or a
// library annotation meta-annotated @Stereotype, @NormalScope or @Scope).
// Classes of other jars are not looked at. Sorted by class.
func CDIAnnouncements(jar string, h *hierarchy.Hierarchy) ([]Announcement, error) {
	zr, err := zip.OpenReader(jar)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var beans []byte
	hasBeans, hasJandex := false, false
	for _, f := range zr.File {
		switch f.Name {
		case beansXML:
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			beans, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
			hasBeans = true
		case jandexIdx:
			hasJandex = true
		}
	}
	file, mode := "", ""
	switch {
	case hasBeans:
		file, mode = beansXML, discoveryMode(beans)
	case hasJandex:
		file, mode = jandexIdx, "annotated"
	default:
		return nil, nil
	}
	if mode == "none" {
		return nil, nil
	}
	var out []Announcement
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp || !concreteClass(ci) || localOrAnonymous(fqcn) {
			continue
		}
		key := cdiKeyPrefix + mode
		if mode == "annotated" {
			a := beanDefiningOf(h, ci)
			if a == "" {
				continue
			}
			key += " @" + a
		} else if !hasBeanConstructor(ci) {
			continue
		}
		out = append(out, Announcement{File: file, Key: key, Class: fqcn})
	}
	return out, nil
}

// discoveryMode reads beans.xml's mode (see CDIAnnouncements).
func discoveryMode(b []byte) string {
	if len(bytes.TrimSpace(b)) == 0 {
		return "all"
	}
	dec := xml.NewDecoder(bytes.NewReader(b))
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err != nil {
			return "all" // unreadable: what an older container assumes
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "beans" {
			continue
		}
		mode, version := "", ""
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "bean-discovery-mode":
				mode = strings.ToLower(strings.TrimSpace(a.Value))
			case "version":
				version = strings.TrimSpace(a.Value)
			}
		}
		if mode == "all" || mode == "annotated" || mode == "none" {
			return mode
		}
		if major, err := strconv.Atoi(strings.SplitN(version, ".", 2)[0]); err == nil && major >= 4 {
			return "annotated"
		}
		return "all"
	}
}

func concreteClass(ci *hierarchy.ClassInfo) bool {
	const accAnnotation, accEnum = 0x2000, 0x4000
	return ci.AccessFlags&(classfile.AccInterface|classfile.AccAbstract|accAnnotation|accEnum) == 0
}

// localOrAnonymous: Outer$1, Outer$1Local.
func localOrAnonymous(fqcn string) bool {
	i := strings.LastIndexByte(fqcn, '$')
	return i >= 0 && i+1 < len(fqcn) && fqcn[i+1] >= '0' && fqcn[i+1] <= '9'
}

func hasBeanConstructor(ci *hierarchy.ClassInfo) bool {
	for _, m := range ci.Methods {
		if m.Name == "<init>" && (m.Descriptor == "()V" || hasAnnotation(m.Annotations, cdiMethodAnnotations)) {
			return true
		}
	}
	return false
}

// beanDefiningOf returns the class's bean-defining annotation, "" if none.
func beanDefiningOf(h *hierarchy.Hierarchy, ci *hierarchy.ClassInfo) string {
	for _, a := range ci.Annotations {
		if beanDefining[a] {
			return a
		}
	}
	for _, a := range ci.Annotations {
		if ai, ok := h.Lookup(a); ok && hasAnnotation(ai.Annotations, cdiMeta) {
			return a
		}
	}
	return ""
}

// isCDI reports whether an announcement is a CDI bean.
func (a Announcement) isCDI() bool { return strings.HasPrefix(a.Key, cdiKeyPrefix) }

// cdiEntry reports whether the container calls m on a bean: an
// @Inject/@Produces/@PostConstruct/@PreDestroy/@AroundInvoke/
// @AroundConstruct method, or one with an @Observes/@ObservesAsync/
// @Disposes parameter.
func cdiEntry(m hierarchy.Method) bool {
	if hasAnnotation(m.Annotations, cdiMethodAnnotations) {
		return true
	}
	for _, p := range m.ParameterAnnotations {
		if hasAnnotation(p, cdiParamAnnotations) {
			return true
		}
	}
	return false
}
