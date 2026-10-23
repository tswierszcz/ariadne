// Package webxml is FRAMEWORKS F1's own "ograniczony odczyt web.xml"
// slice: a limited, safe reader for a servlet deployment descriptor —
// servlets, servlet mappings, filters, filter mappings, listeners and
// context parameters — not a general DTD-validating XML processor.
//
// Safety: Go's encoding/xml never fetches an external DTD subset or
// resolves an external entity at all (it has no such capability to
// begin with, not a flag that happens to be off), so this package is
// XXE-safe by construction, matching F1's own explicit requirement
// ("XML nie pobiera zewnętrznych definicji ani nie uruchamia kodu").
package webxml

import (
	"encoding/xml"
	"fmt"
	"io"
)

// Servlet is one <servlet> declaration.
type Servlet struct {
	Name, Class string
}

// ServletMapping is one <servlet-mapping> declaration.
type ServletMapping struct {
	Name, URLPattern string
}

// Filter is one <filter> declaration.
type Filter struct {
	Name, Class string
}

// FilterMapping is one <filter-mapping> declaration — ServletName is ""
// when the mapping is by <url-pattern> instead (the two are mutually
// exclusive per the servlet spec, but both fields are kept so a caller
// can tell which shape a given entry actually used).
type FilterMapping struct {
	Name, URLPattern, ServletName string
}

// Listener is one <listener> declaration.
type Listener struct {
	Class string
}

// ContextParam is one <context-param> declaration.
type ContextParam struct {
	Name, Value string
}

// WebXML is a limited, decoded web.xml: every element this package
// recognizes, plus the FQCNs recognized here converted from their own
// binary-name-free, already-dotted form (web.xml class names are
// already written with dots, unlike a classfile Methodref owner, so no
// conversion is actually needed — Class is kept verbatim).
type WebXML struct {
	Servlets        []Servlet
	ServletMappings []ServletMapping
	Filters         []Filter
	FilterMappings  []FilterMapping
	Listeners       []Listener
	ContextParams   []ContextParam
	// Skipped is every top-level <web-app> child element name this
	// package does not model (e.g. <session-config>, <error-page>,
	// <security-constraint>, <jsp-config>, <welcome-file-list>) — F1's
	// own explicit requirement ("odnotować pominięte elementy"): a
	// present-but-unmodeled element is named here, never silently
	// dropped without a trace. Deduplicated, in first-seen order.
	Skipped []string
}

type rawServlet struct {
	Name  string `xml:"servlet-name"`
	Class string `xml:"servlet-class"`
}

type rawServletMapping struct {
	Name       string `xml:"servlet-name"`
	URLPattern string `xml:"url-pattern"`
}

type rawFilter struct {
	Name  string `xml:"filter-name"`
	Class string `xml:"filter-class"`
}

type rawFilterMapping struct {
	Name        string `xml:"filter-name"`
	URLPattern  string `xml:"url-pattern"`
	ServletName string `xml:"servlet-name"`
}

type rawListener struct {
	Class string `xml:"listener-class"`
}

type rawContextParam struct {
	Name  string `xml:"param-name"`
	Value string `xml:"param-value"`
}

// Parse decodes a web.xml document's bytes. Only the top-level
// children of the outermost <web-app> element are inspected — a
// malformed or truncated document is a real error (unlike an
// unmodeled-but-well-formed element, which is Skipped, not an error).
func Parse(r io.Reader) (WebXML, error) {
	dec := xml.NewDecoder(r)
	// dec.Entity is deliberately left nil: an undefined entity
	// reference in the document becomes a decode error rather than
	// being silently substituted or fetched — this package never
	// resolves an external entity, by construction, not by a flag.

	var out WebXML
	skippedSeen := map[string]bool{}
	enteredRoot := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return WebXML{}, fmt.Errorf("decoding web.xml: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !enteredRoot {
			// The outermost element itself (<web-app>, whatever its
			// namespace) — not inspected further, just entered. Every
			// subsequent StartElement dec.Token() hands back is
			// necessarily a direct child of it: each case below fully
			// consumes its own element's subtree (DecodeElement/Skip),
			// including its own EndElement, before the loop asks for
			// another token.
			enteredRoot = true
			continue
		}

		switch start.Name.Local {
		case "servlet":
			var s rawServlet
			if err := dec.DecodeElement(&s, &start); err != nil {
				return WebXML{}, fmt.Errorf("decoding <servlet>: %w", err)
			}
			out.Servlets = append(out.Servlets, Servlet{Name: s.Name, Class: s.Class})
		case "servlet-mapping":
			var m rawServletMapping
			if err := dec.DecodeElement(&m, &start); err != nil {
				return WebXML{}, fmt.Errorf("decoding <servlet-mapping>: %w", err)
			}
			out.ServletMappings = append(out.ServletMappings, ServletMapping{Name: m.Name, URLPattern: m.URLPattern})
		case "filter":
			var f rawFilter
			if err := dec.DecodeElement(&f, &start); err != nil {
				return WebXML{}, fmt.Errorf("decoding <filter>: %w", err)
			}
			out.Filters = append(out.Filters, Filter{Name: f.Name, Class: f.Class})
		case "filter-mapping":
			var m rawFilterMapping
			if err := dec.DecodeElement(&m, &start); err != nil {
				return WebXML{}, fmt.Errorf("decoding <filter-mapping>: %w", err)
			}
			out.FilterMappings = append(out.FilterMappings, FilterMapping{Name: m.Name, URLPattern: m.URLPattern, ServletName: m.ServletName})
		case "listener":
			var l rawListener
			if err := dec.DecodeElement(&l, &start); err != nil {
				return WebXML{}, fmt.Errorf("decoding <listener>: %w", err)
			}
			out.Listeners = append(out.Listeners, Listener{Class: l.Class})
		case "context-param":
			var p rawContextParam
			if err := dec.DecodeElement(&p, &start); err != nil {
				return WebXML{}, fmt.Errorf("decoding <context-param>: %w", err)
			}
			out.ContextParams = append(out.ContextParams, ContextParam{Name: p.Name, Value: p.Value})
		default:
			if !skippedSeen[start.Name.Local] {
				skippedSeen[start.Name.Local] = true
				out.Skipped = append(out.Skipped, start.Name.Local)
			}
			if err := dec.Skip(); err != nil {
				return WebXML{}, fmt.Errorf("skipping <%s>: %w", start.Name.Local, err)
			}
		}
	}
	return out, nil
}
