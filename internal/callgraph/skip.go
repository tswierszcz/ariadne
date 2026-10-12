package callgraph

import "strings"

// jdkPrefixes are the package prefixes CHA must never expand (section
// 3.6): a virtual/interface call declared on a JDK type like
// Object#toString, Runnable#run, Map#get or Iterator#next would otherwise
// fan out to a huge fraction of the whole classpath once every app class
// that happens to override one of these common methods is pulled in,
// making nearly every CVE look REACHABLE regardless of whether the app
// code actually has anything to do with it.
var jdkPrefixes = []string{"java.", "javax.", "jdk.", "sun."}

// SkipFunc decides whether a virtual/interface call's declared receiver
// type should be exempted from CHA expansion (its direct edge is still
// added either way; only the fan-out to every subtype is skipped).
type SkipFunc func(fqcn string) bool

// DefaultSkip implements the JDK-prefix rule above. Build's caller may
// combine it with additional entries (a configurable "too generic" list,
// e.g. from cha_skip.yaml) by wrapping it in their own SkipFunc.
func DefaultSkip(fqcn string) bool {
	for _, p := range jdkPrefixes {
		if strings.HasPrefix(fqcn, p) {
			return true
		}
	}
	return false
}
