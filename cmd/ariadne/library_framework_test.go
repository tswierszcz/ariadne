package main

import (
	"slices"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/libapi"
	"ariadne/internal/profile"
)

// TestSpringFrameworkPatterns (LIBRARY_MODE lb2-8): a library announcing
// classes to Spring gets the spring profile's framework packages the
// selected profile lacks; a CDI bean archive or no announcement gets none,
// and nothing is added twice when the spring profile is already selected.
func TestSpringFrameworkPatterns(t *testing.T) {
	catalog, err := profile.LoadDefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	h, _, err := hierarchy.Build("../../testdata/library-autoconfig/lib.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	generic, _ := catalog.Get("generic")
	spring, _ := catalog.Get("spring")
	factories := []libapi.Announcement{{File: "META-INF/spring.factories", Key: "org.springframework.boot.autoconfigure.EnableAutoConfiguration", Class: "a.B"}}
	imports := []libapi.Announcement{{File: "META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports", Class: "a.B"}}
	cdi := []libapi.Announcement{{File: "META-INF/beans.xml", Key: "bean-discovery-mode=annotated", Class: "a.B"}}

	for name, anns := range map[string][]libapi.Announcement{"spring.factories": factories, "imports": imports} {
		got := springFrameworkPatterns(anns, catalog, generic, h)
		if !slices.Contains(got, "org.springframework.boot:*") || !slices.Contains(got, "org.springframework:spring-beans") || !slices.Contains(got, "ch.qos.logback:*") {
			t.Errorf("%s under generic: %v, want the spring profile's packages", name, got)
		}
	}
	if got := springFrameworkPatterns(cdi, catalog, generic, h); got != nil {
		t.Errorf("CDI only: %v, want none", got)
	}
	if got := springFrameworkPatterns(nil, catalog, generic, h); got != nil {
		t.Errorf("no announcements: %v, want none", got)
	}
	if got := springFrameworkPatterns(factories, catalog, spring, h); got != nil {
		t.Errorf("spring profile already selected: %v, want none added", got)
	}
}
