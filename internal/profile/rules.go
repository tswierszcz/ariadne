package profile

import (
	"crypto/sha256"
	"encoding/hex"
)

// RulesVersion names the rules that turn a profile into roots and
// framework facts (FRAMEWORKS F5): the root kinds in this package
// (annotations, main, overrides, web.xml, roots.beans with its constructor
// and lifecycle choice, Spring XML through internal/springxml; f5-2: the
// Spring MVC JSON body roots of WebBodyRoots), @Profile evaluation (f5-3:
// Spring 5.1+ expressions), the entry catalog, injection points and their
// resolution (f5-3: a qualifier matches a default bean name or a class
// @Qualifier) and @Value facts. A snapshot records it next to
// BuiltinDigest, so a comparison can say a verdict moved because these
// rules changed; bump it with any change to them —
// TestRulesVersion_tracksRuleSource fails until you do.
const RulesVersion = "f5-3"

// BuiltinDigest is a short SHA-256 of the built-in profile YAMLs, in
// catalog order: a profile's roots, synthetic edges and models change with
// these files without any Go change, so the snapshot records their
// content, not a hand-kept version.
func BuiltinDigest() string {
	h := sha256.New()
	for _, raw := range builtinYAMLs() {
		h.Write(raw)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func builtinYAMLs() [][]byte {
	return [][]byte{springYAML, servletYAML, jakartaYAML, micronautYAML, quarkusYAML, stripesYAML, genericYAML}
}
