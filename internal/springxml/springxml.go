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
)

// Bean is one `<bean>` declaration's own identity and target class —
// scoped to exactly what FRAMEWORKS F1's own bean-name-conflict check
// needs; `<property>`/`<constructor-arg>` wiring is not modeled here
// (nothing in this increment consumes it, and this codebase's own
// "don't build what nothing uses yet" rule applies).
type Bean struct {
	// ID is the bean's own `id` attribute, or, when absent, its first
	// `name` attribute value (Spring treats the two as interchangeable
	// primary identifiers; only the first of a space/comma-separated
	// `name` list is kept, since this package does not model bean
	// aliases as a separate concept). "" when neither attribute is
	// given (an anonymous inner bean, usually nested inside a
	// `<property>` — never seen here, since this package only looks at
	// `<beans>`' own direct `<bean>` children).
	ID    string
	Class string
	Scope string // "" means Spring's own default ("singleton")
	// Primary is this bean's own `primary="true"` attribute — Spring's
	// own tie-breaker when more than one bean matches an injection
	// point by type.
	Primary bool
}

// BeansXML is a limited, decoded Spring bean-definitions document.
type BeansXML struct {
	Beans []Bean
	// Skipped is every top-level <beans> child element name this
	// package does not model (e.g. `<import>`, `<alias>`, or any
	// namespaced element like `<context:component-scan>` — its own
	// Local name only, the namespace prefix is not modeled), in
	// first-seen, deduplicated order — never silently dropped without
	// a trace, same discipline as internal/webxml's own Skipped.
	Skipped []string
}

type rawBean struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Class   string `xml:"class,attr"`
	Scope   string `xml:"scope,attr"`
	Primary string `xml:"primary,attr"`
}

// Parse decodes a Spring bean-definitions document's bytes. Only the
// top-level children of the outermost <beans> element are inspected.
func Parse(r io.Reader) (BeansXML, error) {
	dec := xml.NewDecoder(r)
	// dec.Entity is deliberately left nil — see the package doc comment.

	var out BeansXML
	skippedSeen := map[string]bool{}
	enteredRoot := false
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
		if !enteredRoot {
			// The outermost element itself (<beans>) — see
			// webxml.Parse's own doc comment for why every subsequent
			// StartElement is necessarily a direct child.
			enteredRoot = true
			continue
		}

		if start.Name.Local == "bean" {
			var b rawBean
			if err := dec.DecodeElement(&b, &start); err != nil {
				return BeansXML{}, fmt.Errorf("decoding <bean>: %w", err)
			}
			id := b.ID
			if id == "" {
				id = firstName(b.Name)
			}
			out.Beans = append(out.Beans, Bean{
				ID:      id,
				Class:   b.Class,
				Scope:   b.Scope,
				Primary: b.Primary == "true",
			})
			continue
		}

		if !skippedSeen[start.Name.Local] {
			skippedSeen[start.Name.Local] = true
			out.Skipped = append(out.Skipped, start.Name.Local)
		}
		if err := dec.Skip(); err != nil {
			return BeansXML{}, fmt.Errorf("skipping <%s>: %w", start.Name.Local, err)
		}
	}
	return out, nil
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
