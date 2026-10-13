package vulnmap

import (
	"path"
	"strings"
)

// isTestPath reports whether p looks like test code rather than main
// source: a path segment literally named "test"/"tests", or a file name
// ending "Test.java"/"Tests.java" (but not merely containing "Test", so
// e.g. "Contest.java" is not a test).
func isTestPath(p string) bool {
	for _, seg := range strings.Split(path.Dir(p), "/") {
		if seg == "test" || seg == "tests" {
			return true
		}
	}
	base := path.Base(p)
	return strings.HasSuffix(base, "Test.java") || strings.HasSuffix(base, "Tests.java")
}

// isNonTestJavaFile keeps every non-test .java file, regardless of build
// layout. It does not require a Maven-style src/main/ prefix: repositories
// that lay out code differently (notably Apache Tomcat, which uses
// java/org/apache/... for code and test/... for tests) would otherwise
// lose every file to a stricter filter, hiding their fix-commit diffs from
// both the diff source and the LLM source.
func isNonTestJavaFile(p string) bool {
	return strings.HasSuffix(p, ".java") && !isTestPath(p)
}
