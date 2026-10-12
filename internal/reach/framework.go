package reach

import (
	"path"
	"strings"

	"ariadne/internal/input"
)

// DefaultFrameworkPackages is section 4's initial framework package list:
// code that's part of the runtime framework rather than the app itself,
// so a vulnerability in it is driven by the framework (Tomcat parsing an
// HTTP request, Spring dispatching to a controller, ...) and won't show
// up on a path from app code. Each entry is "groupId:artifactId", with
// artifactId "*" matching any artifact in that group.
//
// This is deliberately not yet the full profile system section 4
// describes (framework_packages.yaml, a conditional jackson-databind
// rule keyed on @RequestBody/@ResponseBody): that's CLI/profile wiring
// for a later phase. jackson-databind is included unconditionally here
// rather than left out, since Phase 5 needs *some* answer and an
// unconditional "maybe this is framework-driven" is more honest than
// silently treating it as pure app-code reachability.
var DefaultFrameworkPackages = []string{
	"org.apache.tomcat.embed:*",
	"org.springframework:spring-web",
	"org.springframework:spring-webmvc",
	"org.springframework.boot:*",
	"com.fasterxml.jackson.core:jackson-databind",
	"org.hibernate:*",
	"ch.qos.logback:*",
	"org.apache.logging.log4j:*",
}

// FrameworkMatcher decides whether a GAV belongs to the active profile's
// framework packages.
type FrameworkMatcher struct {
	patterns []pattern
}

type pattern struct {
	group    string
	artifact string // "*" matches any
}

// NewFrameworkMatcher builds a matcher from "groupId:artifactId" patterns
// (artifactId may be "*"); a malformed pattern (no colon) is ignored.
func NewFrameworkMatcher(patterns []string) *FrameworkMatcher {
	m := &FrameworkMatcher{}
	for _, p := range patterns {
		g, a, ok := strings.Cut(p, ":")
		if !ok {
			continue
		}
		m.patterns = append(m.patterns, pattern{group: g, artifact: a})
	}
	return m
}

// Match reports whether gav belongs to a framework package. A nil gav
// (an app class, or a dependency jar ResolveDepJars couldn't identify)
// never matches.
func (m *FrameworkMatcher) Match(gav *input.GAV) bool {
	if gav == nil {
		return false
	}
	for _, p := range m.patterns {
		if p.group != gav.Group {
			continue
		}
		if p.artifact == "*" {
			return true
		}
		if ok, _ := path.Match(p.artifact, gav.Artifact); ok {
			return true
		}
	}
	return false
}
