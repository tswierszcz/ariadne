package vulnmap

import "strings"

// Static-initializer mappings (2026-11-17). When a fix changed only field
// initializers or static blocks, the diff and release-diff sources map the
// advisory to the class's <clinit> (CVE:F10). A static initializer runs
// the first time the class is used, so reaching it says "the class was
// loaded", not "the vulnerable code ran". Two consequences:
//
//   - A mapping that names only static initializers carries the
//     qualifier MAPPING_INITIALIZER_ONLY (Entry.InitializerOnly): its
//     verdict stands, but a reader knows how weak the anchor is.
//   - For a known family whose initializer is a data table guarding a
//     configuration-gated feature, the initializers are replaced by that
//     feature's entry points (initializerRedirects).

// InitializerRedirect is the evidence of a mapping whose static
// initializers were replaced by entry points (initializerRedirects).
type InitializerRedirect struct {
	// Initializers are the static initializers the fix changed — what
	// the source mapped before the redirect.
	Initializers []string `json:"initializers"`
	// Reason says what the initializers hold and which entry points
	// replace them.
	Reason string `json:"reason"`
}

// initializerRedirect is one curated family: a set of classes whose
// <clinit> holds a data table, and the entry points that enable the code
// the table guards.
type initializerRedirect struct {
	classes     []string
	entrypoints []string
	reason      string
}

// initializerRedirects are checked by hand like vuln_methods.yaml.
//
// jackson-databind polymorphic-typing gadgets (CVE-2017-7525 and the ~60
// advisories after it): each fix only added class names to the
// deserialization blocklist — SubTypeValidator's static set from 2.9.x
// (and the 2.8/2.7 backports), BeanDeserializerFactory's before it. That
// initializer runs on any bean deserialization, so every Jackson user
// would be REACHABLE. The advisories apply only when polymorphic typing
// with class-name type ids is enabled for a broad nominal type; the entry
// points below are the API that enables it globally (default typing).
// Checked in the bytecode of databind 2.0.0, 2.9.0, 2.9.10.4, 2.10.0,
// 2.11.4, 2.12.0 and 2.13.5: ObjectMapper declares enableDefaultTyping
// (3 overloads), enableDefaultTypingAsProperty and setDefaultTyping in
// every one; activateDefaultTyping (3), activateDefaultTypingAsProperty
// and the MapperBuilder methods exist from 2.10.0 (MapperBuilder
// #setDefaultTyping from 2.11). Not covered: @JsonTypeInfo(use = CLASS or
// MINIMAL_CLASS) on an application type, which enables the same
// resolution per property without any of these calls — the mapping stays
// partial.
var initializerRedirects = []initializerRedirect{{
	classes: []string{
		"com.fasterxml.jackson.databind.jsontype.impl.SubTypeValidator",
		"com.fasterxml.jackson.databind.deser.BeanDeserializerFactory",
	},
	entrypoints: []string{
		"com.fasterxml.jackson.databind.ObjectMapper#activateDefaultTyping",
		"com.fasterxml.jackson.databind.ObjectMapper#activateDefaultTypingAsProperty",
		"com.fasterxml.jackson.databind.ObjectMapper#enableDefaultTyping",
		"com.fasterxml.jackson.databind.ObjectMapper#enableDefaultTypingAsProperty",
		"com.fasterxml.jackson.databind.ObjectMapper#setDefaultTyping",
		"com.fasterxml.jackson.databind.cfg.MapperBuilder#activateDefaultTyping",
		"com.fasterxml.jackson.databind.cfg.MapperBuilder#activateDefaultTypingAsProperty",
		"com.fasterxml.jackson.databind.cfg.MapperBuilder#setDefaultTyping",
	},
	reason: "the fix only extended jackson-databind's polymorphic-typing blocklist (a static initializer that runs on any " +
		"deserialization); mapped to the API that enables default typing instead. @JsonTypeInfo(use = CLASS) on an " +
		"application type is not covered",
}}

// staticInitializerClass returns the class of a "C#<clinit>" selector
// (with or without a descriptor).
func staticInitializerClass(sel string) (string, bool) {
	cls, name, ok := strings.Cut(sel, "#")
	if !ok {
		return "", false
	}
	if i := strings.IndexByte(name, '('); i >= 0 {
		name = name[:i]
	}
	return cls, name == "<clinit>"
}

// redirectInitializers replaces a mapping that names only static
// initializers of one curated family by that family's entry points. ok is
// false (methods returned unchanged) when any method is something else.
func redirectInitializers(methods []string) ([]string, *InitializerRedirect, bool) {
	if len(methods) == 0 {
		return methods, nil, false
	}
	for _, r := range initializerRedirects {
		in := map[string]bool{}
		for _, c := range r.classes {
			in[c] = true
		}
		all := true
		for _, m := range methods {
			if cls, ok := staticInitializerClass(m); !ok || !in[cls] {
				all = false
				break
			}
		}
		if all {
			from := append([]string(nil), methods...)
			return append([]string(nil), r.entrypoints...), &InitializerRedirect{Initializers: from, Reason: r.reason}, true
		}
	}
	return methods, nil, false
}

// InitializerOnly reports a diff or release-diff mapping that names only
// static initializers (reason MAPPING_INITIALIZER_ONLY): reaching one
// means the class was used, not that the changed code matters on that
// path. Instance initializers ("C#<init>" for every constructor) are not
// counted — the selector cannot tell them from an explicit constructor.
func (e Entry) InitializerOnly() bool {
	if e.Source != SourceDiff && e.Source != SourceReleaseDiff || len(e.Methods) == 0 {
		return false
	}
	for _, m := range e.Methods {
		if _, ok := staticInitializerClass(m); !ok {
			return false
		}
	}
	return true
}
