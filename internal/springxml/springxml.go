// Package springxml is FRAMEWORKS F1's own "ograniczony... XML bean
// definitions" slice: a limited reader for Spring's classic
// XML-based bean wiring (`<beans><bean id="..." class="...">`), not a
// general Spring XML config processor (no `<aop:config>`, no
// `<context:component-scan>`, no property placeholder resolution).
//
// Safety: same as internal/webxml — Go's encoding/xml never fetches an
// external DTD subset or resolves an external entity at all, so this
// package is XXE-safe by construction, not by a disabled flag.
package springxml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Bean is one `<bean>` declaration. ID/Class/Scope/Primary serve
// FRAMEWORKS F1's bean-name-conflict check; the rest is what the
// container does with the bean when it creates it (ROADMAP B4 b4-8: the
// spring profile's bean lifecycle roots) — which constructor arity it
// calls, which properties it sets, which init/destroy methods it runs.
type Bean struct {
	// ID is the bean's own `id` attribute, or, when absent, its first
	// `name` attribute value (Spring treats the two as interchangeable
	// primary identifiers; only the first of a space/comma-separated
	// `name` list is kept, since this package does not model bean
	// aliases as a separate concept). "" for an anonymous bean (an inner
	// bean, see BeansXML.Inner).
	ID    string
	Class string
	Scope string // "" means Spring's own default ("singleton")
	// Primary is this bean's own `primary="true"` attribute — Spring's
	// own tie-breaker when more than one bean matches an injection
	// point by type.
	Primary bool

	// Abstract (`abstract="true"`): a template other definitions name as
	// `parent`; the container never instantiates it.
	Abstract bool
	// LazyInit (`lazy-init="true"`): created on first use only.
	LazyInit bool
	// Parent is the `parent` definition's name: the child inherits its
	// class, properties, constructor arguments and init/destroy methods.
	Parent string
	// InitMethod/DestroyMethod are the bean's own attributes ("" when
	// absent); DefaultInitMethod/DefaultDestroyMethod are the enclosing
	// <beans>' default-init-method/default-destroy-method, which Spring
	// applies only when the bean has such a method.
	InitMethod, DestroyMethod               string
	DefaultInitMethod, DefaultDestroyMethod string
	// Properties are the names set through <property name> and p:name /
	// p:name-ref attributes, in document order, without duplicates.
	Properties []string
	// ConstructorArgs counts <constructor-arg> elements and c: attributes.
	ConstructorArgs int
	// FactoryMethod/FactoryBean: the bean is made by a factory method
	// (static on Class, or on the FactoryBean bean), not a constructor.
	FactoryMethod, FactoryBean string
}

// BeansXML is a limited, decoded Spring bean-definitions document.
type BeansXML struct {
	// Beans are the named definitions: <bean> children of the root
	// <beans> and of nested <beans> sections (a nested section's profile
	// attribute is not evaluated — its beans count).
	Beans []Bean
	// Inner are the anonymous beans declared inside another bean's
	// <property>, <constructor-arg> or collection elements.
	Inner []Bean
	// Skipped is every top-level <beans> child element name this
	// package does not model (e.g. `<import>`, `<alias>`, or any
	// namespaced element like `<context:component-scan>` — its own
	// Local name only, the namespace prefix is not modeled), in
	// first-seen, deduplicated order — never silently dropped without
	// a trace, same discipline as internal/webxml's own Skipped.
	Skipped []string
	// AOP are the <aop:config> sections (FRAMEWORKS F3, spring:aop).
	AOP []AOPConfig
}

// AOPConfig is one <aop:config> section: its named pointcuts and aspects.
// <aop:advisor> elements (advice beans such as <tx:advice>) are counted,
// not modeled here.
type AOPConfig struct {
	// Pointcuts are the section's <aop:pointcut id expression>.
	Pointcuts map[string]string
	Aspects   []AOPAspect
	Advisors  int
}

// AOPAspect is an <aop:aspect ref>: advice methods of the referenced bean.
type AOPAspect struct {
	ID, Ref string
	// Pointcuts are the aspect's own <aop:pointcut id expression>.
	Pointcuts map[string]string
	Advice    []AOPAdvice
}

// AOPAdvice is one <aop:before|after|after-returning|after-throwing|around>.
type AOPAdvice struct {
	// Kind is the element name (before, after-returning, …).
	Kind, Method string
	// Pointcut is the inline expression; PointcutRef names a pointcut.
	Pointcut, PointcutRef string
	Returning, Throwing   string
	ArgNames              string
}

const aopNamespace = "http://www.springframework.org/schema/aop"

func aopConfig(n node) AOPConfig {
	cfg := AOPConfig{Pointcuts: map[string]string{}}
	for _, c := range n.Children {
		switch c.XMLName.Local {
		case "pointcut":
			cfg.Pointcuts[c.attr("id")] = c.attr("expression")
		case "advisor":
			cfg.Advisors++
		case "aspect":
			a := AOPAspect{ID: c.attr("id"), Ref: c.attr("ref"), Pointcuts: map[string]string{}}
			for _, d := range c.Children {
				switch d.XMLName.Local {
				case "pointcut":
					a.Pointcuts[d.attr("id")] = d.attr("expression")
				case "before", "after", "after-returning", "after-throwing", "around":
					a.Advice = append(a.Advice, AOPAdvice{
						Kind: d.XMLName.Local, Method: d.attr("method"),
						Pointcut: d.attr("pointcut"), PointcutRef: d.attr("pointcut-ref"),
						Returning: d.attr("returning"), Throwing: d.attr("throwing"), ArgNames: d.attr("arg-names"),
					})
				}
			}
			cfg.Aspects = append(cfg.Aspects, a)
		}
	}
	return cfg
}

const (
	pNamespace = "http://www.springframework.org/schema/p"
	cNamespace = "http://www.springframework.org/schema/c"
)

// node is any element, with its attributes and children, decoded whole.
type node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Children []node     `xml:",any"`
}

func (n node) attr(local string) string {
	for _, a := range n.Attrs {
		if a.Name.Local == local && (a.Name.Space == "" || a.Name.Space == n.XMLName.Space) {
			return a.Value
		}
	}
	return ""
}

type defaults struct{ init, destroy string }

func (d defaults) within(n node) defaults {
	if v := n.attr("default-init-method"); v != "" {
		d.init = v
	}
	if v := n.attr("default-destroy-method"); v != "" {
		d.destroy = v
	}
	return d
}

// Parse decodes a Spring bean-definitions document's bytes: the
// definitions under the outermost <beans> element and nested <beans>
// sections, with their inner beans.
func Parse(r io.Reader) (BeansXML, error) {
	dec := xml.NewDecoder(r)
	// dec.Entity is deliberately left nil — see the package doc comment.

	var out BeansXML
	skippedSeen := map[string]bool{}
	var root *xml.StartElement
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return BeansXML{}, fmt.Errorf("decoding beans XML: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if root == nil {
			// The outermost element itself (<beans>) — see
			// webxml.Parse's own doc comment for why every subsequent
			// StartElement is necessarily a direct child.
			s := start.Copy()
			root = &s
			continue
		}
		var n node
		if err := dec.DecodeElement(&n, &start); err != nil {
			return BeansXML{}, fmt.Errorf("decoding <%s>: %w", start.Name.Local, err)
		}
		d := defaults{}.within(node{XMLName: root.Name, Attrs: root.Attr})
		switch n.XMLName.Local {
		case "bean":
			out.addBean(n, d)
		case "beans":
			out.addSection(n, d)
		case "config":
			if n.XMLName.Space == aopNamespace {
				out.AOP = append(out.AOP, aopConfig(n))
				continue
			}
			fallthrough
		default:
			if !skippedSeen[n.XMLName.Local] {
				skippedSeen[n.XMLName.Local] = true
				out.Skipped = append(out.Skipped, n.XMLName.Local)
			}
		}
	}
	return out, nil
}

// addSection adds a nested <beans> section's definitions.
func (out *BeansXML) addSection(n node, d defaults) {
	d = d.within(n)
	for _, c := range n.Children {
		switch c.XMLName.Local {
		case "bean":
			out.addBean(c, d)
		case "beans":
			out.addSection(c, d)
		case "config":
			if c.XMLName.Space == aopNamespace {
				out.AOP = append(out.AOP, aopConfig(c))
			}
		}
	}
}

func (out *BeansXML) addBean(n node, d defaults) {
	out.Beans = append(out.Beans, out.bean(n, d))
}

// bean interprets one <bean> element and collects its inner beans.
func (out *BeansXML) bean(n node, d defaults) Bean {
	id := n.attr("id")
	if id == "" {
		id = firstName(n.attr("name"))
	}
	b := Bean{
		ID:                   id,
		Class:                n.attr("class"),
		Scope:                n.attr("scope"),
		Primary:              n.attr("primary") == "true",
		Abstract:             n.attr("abstract") == "true",
		LazyInit:             n.attr("lazy-init") == "true",
		Parent:               n.attr("parent"),
		InitMethod:           n.attr("init-method"),
		DestroyMethod:        n.attr("destroy-method"),
		DefaultInitMethod:    d.init,
		DefaultDestroyMethod: d.destroy,
		FactoryMethod:        n.attr("factory-method"),
		FactoryBean:          n.attr("factory-bean"),
	}
	seen := map[string]bool{}
	addProp := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			b.Properties = append(b.Properties, name)
		}
	}
	for _, a := range n.Attrs {
		switch a.Name.Space {
		case pNamespace:
			addProp(strings.TrimSuffix(a.Name.Local, "-ref"))
		case cNamespace:
			b.ConstructorArgs++
		}
	}
	for _, c := range n.Children {
		switch c.XMLName.Local {
		case "property":
			addProp(c.attr("name"))
		case "constructor-arg":
			b.ConstructorArgs++
		}
		out.collectInner(c, d)
	}
	return b
}

// collectInner finds the anonymous <bean> elements anywhere under n (a
// <property>, a <constructor-arg>, a <list> or <map> inside them).
func (out *BeansXML) collectInner(n node, d defaults) {
	if n.XMLName.Local == "bean" {
		out.Inner = append(out.Inner, out.bean(n, d))
		return
	}
	for _, c := range n.Children {
		out.collectInner(c, d)
	}
}

// firstName returns the first of a bean's own (possibly
// space/comma-separated) `name` attribute's aliases — Spring's own
// shape for "one or more additional names", not modeled as a separate
// alias list here, only as a fallback identifier when `id` is absent.
func firstName(name string) string {
	for i, c := range name {
		if c == ' ' || c == ',' {
			return name[:i]
		}
	}
	return name
}
