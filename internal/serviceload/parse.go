// Package serviceload inventories ServiceLoader provider declarations
// (SERVICE_LOADING S0): parsing META-INF/services/<Interface> files and
// cross-referencing each declared provider class against the analyzed
// hierarchy. A ServiceRegistration is a declarative fact read from a
// jar's own metadata, never a call-graph edge — whether anything in the
// app actually calls ServiceLoader.load for that interface, and whether
// a provider is ever actually materialized, are SERVICE_LOADING S1/S2's
// own, later concerns (see ServiceLoadSite/ProviderMaterialization's own
// doc comments).
package serviceload

import (
	"fmt"
	"strings"
	"unicode"
)

// Entry is one line of a provider-configuration file, after comment
// stripping and whitespace trimming — JVMS's own parsing rule for these
// files (see java.util.ServiceLoader's "Deploying service providers on
// the class path" documentation): '#' starts a comment to end of line,
// surrounding whitespace is ignored, blank lines are ignored, and
// remaining content must be a legal fully qualified binary class name.
type Entry struct {
	Line          int    // 1-based line number in the raw file, for diagnostics
	Raw           string // the original line, unmodified
	ProviderClass string // "" if this line was blank/comment-only or malformed
	Valid         bool   // true if ProviderClass is a well-formed binary class name
	Duplicate     bool   // true if ProviderClass repeats an earlier *valid* entry in this same file
}

// ParseResult is one service file's parsed content.
type ParseResult struct {
	Interface string
	Entries   []Entry
	// Warnings are human-readable diagnostics for malformed names and
	// duplicates — S0's own "Gotowe" bar ("poprawna lista kandydatów i
	// diagnostyka") wants these surfaced, not silently dropped.
	Warnings []string
}

// Parse parses one META-INF/services/<interfaceFQCN> file's raw content.
// It never errors: a malformed file produces warnings and excludes the
// bad lines from ValidProviders, exactly as a real JVM would skip them
// (ServiceConfigurationError there, a diagnostic here) rather than
// fail the whole analysis over one bad line.
func Parse(interfaceFQCN string, content []byte) ParseResult {
	res := ParseResult{Interface: interfaceFQCN}
	seen := make(map[string]bool)

	lines := strings.Split(string(content), "\n")
	for i, raw := range lines {
		lineNo := i + 1
		line := raw
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		e := Entry{Line: lineNo, Raw: raw}
		if isBinaryClassName(line) {
			e.ProviderClass = line
			e.Valid = true
			if seen[line] {
				e.Duplicate = true
				res.Warnings = append(res.Warnings, fmt.Sprintf("line %d: duplicate provider %q (first occurrence wins; harmless but redundant)", lineNo, line))
			}
			seen[line] = true
		} else {
			res.Warnings = append(res.Warnings, fmt.Sprintf("line %d: %q is not a legal fully qualified class name, skipped", lineNo, line))
		}
		res.Entries = append(res.Entries, e)
	}
	return res
}

// ValidProviders returns every distinct, well-formed provider class name
// r declares, in file order, with duplicates collapsed to their first
// occurrence.
func (r ParseResult) ValidProviders() []string {
	var out []string
	seen := make(map[string]bool)
	for _, e := range r.Entries {
		if !e.Valid || e.Duplicate || seen[e.ProviderClass] {
			continue
		}
		seen[e.ProviderClass] = true
		out = append(out, e.ProviderClass)
	}
	return out
}

// isBinaryClassName reports whether s is a legal fully qualified Java
// binary class name: dot-separated identifiers, each starting with a
// letter/'_'/'$' and continuing with letters/digits/'_'/'$'. This is the
// same shape javac itself requires of an identifier (JLS 3.8), applied
// per dot-separated segment; it does not reject a segment that happens
// to be a Java keyword, since a binary class name is not restricted to
// source-level identifier rules that strictly.
func isBinaryClassName(s string) bool {
	if s == "" {
		return false
	}
	for _, part := range strings.Split(s, ".") {
		if !isJavaIdentifier(part) {
			return false
		}
	}
	return true
}

func isJavaIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' && r != '$' {
				return false
			}
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '$' {
			return false
		}
	}
	return true
}
