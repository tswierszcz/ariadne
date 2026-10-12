package javadiff

import (
	"path"
	"regexp"
	"strings"
)

var packageRe = regexp.MustCompile(`(?m)^\s*package\s+([\w.]+)\s*;`)

// FQCN derives a .java file's top-level public type's fully-qualified
// name from its own "package ...;" declaration and file name — the file
// name IS the type's simple name, by Java's own rule that a public
// top-level type must live in a file of the same name. Inner/nested
// types reachable from it are found separately (classifyHeader tracks
// them while walking the file), so this only needs to get the outer
// name right.
func FQCN(filePath, src string) (string, bool) {
	if !strings.HasSuffix(filePath, ".java") {
		return "", false
	}
	simpleName := strings.TrimSuffix(path.Base(filePath), ".java")
	if simpleName == "" {
		return "", false
	}
	m := packageRe.FindStringSubmatch(src)
	if m == nil {
		return simpleName, true // the default package
	}
	return m[1] + "." + simpleName, true
}
