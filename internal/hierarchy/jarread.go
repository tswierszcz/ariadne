package hierarchy

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/input"
)

// readJarClasses reads every selected .class entry from a jar (per
// selectClassEntries — base entries, or the right multi-release variant
// for targetJava, JVM_GRAPH G4), parses it, and returns one ClassInfo
// per class, plus any warnings about malformed multi-release entries.
func readJarClasses(jarPath string, gav *input.GAV, isApp bool, targetJava int) ([]*ClassInfo, []string, error) {
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, nil, fmt.Errorf("opening %s: %w", jarPath, err)
	}
	defer zr.Close()

	entries, warnings := selectClassEntries(&zr.Reader, targetJava)

	var out []*ClassInfo
	for _, sel := range entries {
		rc, err := sel.file.Open()
		if err != nil {
			return nil, warnings, fmt.Errorf("opening %s in %s: %w", sel.file.Name, jarPath, err)
		}
		cf, err := classfile.Parse(rc)
		rc.Close()
		if err != nil {
			return nil, warnings, fmt.Errorf("parsing %s in %s: %w", sel.file.Name, jarPath, err)
		}

		// ExtractCalls needs cf's constant pool and bootstrap methods,
		// which classfile.ClassFile deliberately doesn't expose, so this
		// is the only point where it can run; the result (not cf itself)
		// is what ClassInfo keeps.
		methods := make([]Method, len(cf.Methods))
		for i, m := range cf.Methods {
			methods[i].Method = m
			if m.Code == nil {
				continue
			}
			calls, initRefs, unresolved, err := cf.ExtractCalls(m.Code)
			if err != nil {
				return nil, warnings, fmt.Errorf("extracting calls from %s %s#%s%s: %w", jarPath, cf.ThisClass, m.Name, m.Descriptor, err)
			}
			methods[i].Calls = calls
			methods[i].InitRefs = initRefs
			methods[i].UnresolvedDynamic = unresolved
		}

		out = append(out, &ClassInfo{
			FQCN:           cf.ThisClass,
			SuperClass:     cf.SuperClass,
			Interfaces:     cf.Interfaces,
			AccessFlags:    cf.AccessFlags,
			MajorVersion:   cf.MajorVersion,
			Methods:        methods,
			Fields:         cf.Fields,
			Annotations:    cf.Annotations,
			SourceJar:      jarPath,
			SourceEntry:    sel.file.Name,
			VariantVersion: sel.variant,
			GAV:            gav,
			IsApp:          isApp,
		})
	}
	return out, warnings, nil
}

// selectedEntry is one class path's chosen zip entry: file is exactly
// which bytes to parse, variant is 0 for an ordinary base entry or the
// multi-release version N it was selected from.
type selectedEntry struct {
	file    *zip.File
	variant int
}

// selectClassEntries decides, for every logical class in zr, exactly
// one zip entry to read (JVM_GRAPH G4):
//   - a plain jar, or any jar when targetJava is unset (0) — every base
//     .class entry, exactly as before this phase: META-INF/versions/* is
//     never treated as active, which is both the documented fallback for
//     "no target given" and (independently) the correct behavior for a
//     jar that isn't actually declared multi-release at all.
//   - a jar whose manifest declares "Multi-Release: true", with
//     targetJava set — for each class path, the highest-numbered
//     META-INF/versions/N/<path> with 9 <= N <= targetJava, falling back
//     to the base entry (or to nothing, if neither exists) when no
//     variant qualifies. Never picks the *highest* bytecode version seen
//     as a guess; an unset target simply does not activate any variant.
//
// module-info.class is always skipped (ariadne analyzes classes, not
// module declarations). A malformed multi-release path (non-numeric or
// <9 version directory, or no class path after it) is skipped with a
// warning rather than guessed at.
func selectClassEntries(zr *zip.Reader, targetJava int) ([]selectedEntry, []string) {
	multiRelease := targetJava > 0 && isMultiReleaseJar(zr)

	base := make(map[string]*zip.File)
	var basePaths []string // insertion order, for deterministic output
	variants := make(map[string]map[int]*zip.File)
	var warnings []string

	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".class") {
			continue
		}
		if f.Name == "module-info.class" {
			continue
		}
		if strings.HasPrefix(f.Name, "META-INF/versions/") {
			if !multiRelease {
				continue // inert resource entry in a non-multi-release (or untargeted) jar
			}
			rest := strings.TrimPrefix(f.Name, "META-INF/versions/")
			verStr, classPath, ok := strings.Cut(rest, "/")
			if !ok || classPath == "" || classPath == "module-info.class" {
				if ok {
					warnings = append(warnings, fmt.Sprintf("malformed multi-release entry %q: skipped", f.Name))
				}
				continue
			}
			n, err := parseVersionDir(verStr)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("invalid multi-release version directory %q: skipped", f.Name))
				continue
			}
			if variants[classPath] == nil {
				variants[classPath] = make(map[int]*zip.File)
			}
			variants[classPath][n] = f
			continue
		}
		if _, dup := base[f.Name]; !dup {
			basePaths = append(basePaths, f.Name)
		}
		base[f.Name] = f
	}

	chosen := make(map[string]selectedEntry, len(basePaths))
	for _, path := range basePaths {
		chosen[path] = selectedEntry{file: base[path]}
	}
	// Paths with a qualifying variant but no base entry at all still need
	// a slot in the output order; append them after the base ones.
	var variantOnlyPaths []string
	for path := range variants {
		if _, hasBase := base[path]; !hasBase {
			variantOnlyPaths = append(variantOnlyPaths, path)
		}
	}

	pickVariant := func(path string) {
		best := -1
		for n, f := range variants[path] {
			if n <= targetJava && n > best {
				best = n
				chosen[path] = selectedEntry{file: f, variant: n}
			}
		}
	}
	for _, path := range basePaths {
		pickVariant(path)
	}
	for _, path := range variantOnlyPaths {
		pickVariant(path)
		if _, ok := chosen[path]; !ok {
			continue // no base and targetJava is below every variant: nothing to select
		}
	}

	out := make([]selectedEntry, 0, len(chosen))
	for _, path := range basePaths {
		out = append(out, chosen[path])
	}
	for _, path := range variantOnlyPaths {
		if sel, ok := chosen[path]; ok {
			out = append(out, sel)
		}
	}
	return out, warnings
}

// parseVersionDir parses a META-INF/versions/<N> segment: N must be a
// plain non-negative integer >= 9 (multi-release JARs were introduced in
// Java 9; JAR_FILE_SPEC rejects anything else as not a version directory
// at all).
func parseVersionDir(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty version directory")
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number: %q", s)
		}
		n = n*10 + int(c-'0')
	}
	if n < 9 {
		return 0, fmt.Errorf("version %d is below the minimum multi-release version (9)", n)
	}
	return n, nil
}

// isMultiReleaseJar reports whether zr's manifest declares
// "Multi-Release: true" — the JAR File Specification's own switch for
// treating META-INF/versions/* as active variants at all, independent
// of whether a target was given.
func isMultiReleaseJar(zr *zip.Reader) bool {
	for _, f := range zr.File {
		if f.Name != "META-INF/MANIFEST.MF" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return false
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return false
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimRight(line, "\r")
			name, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(name), "Multi-Release") &&
				strings.EqualFold(strings.TrimSpace(value), "true") {
				return true
			}
		}
		return false
	}
	return false
}
