// Package input turns a built application (a classes directory, a plain
// jar, or a Spring Boot fat jar) plus its dependency jars into a uniform
// model: the application's class files and a resolved list of dependency
// coordinates (GAVs).
package input

import "fmt"

// GAV identifies a Maven artifact: groupId:artifactId[:classifier]:version.
type GAV struct {
	Group      string
	Artifact   string
	Version    string
	Classifier string // may be empty
}

// String renders the GAV as "group:artifact[:classifier]:version".
func (g GAV) String() string {
	if g.Classifier == "" {
		return fmt.Sprintf("%s:%s:%s", g.Group, g.Artifact, g.Version)
	}
	return fmt.Sprintf("%s:%s:%s:%s", g.Group, g.Artifact, g.Classifier, g.Version)
}

// PackageName is the "group:artifact" form OSV's Maven ecosystem expects as
// a package name.
func (g GAV) PackageName() string {
	return g.Group + ":" + g.Artifact
}

// Source says where a GAV came from, kept for the report's provenance and
// for resolving disagreements between sources.
type Source string

const (
	SourcePOMProperties Source = "pom.properties" // read from inside the jar; highest priority
	SourceDepsList      Source = "deps-list"      // from `mvn dependency:list`
)

// DepJar is one dependency jar together with its resolved coordinates (if
// any could be found) and where they came from.
type DepJar struct {
	Path     string // filesystem path to the jar
	GAV      *GAV   // nil if unidentified
	Source   Source
	Warnings []string // e.g. "multiple pom.properties found"
}
