package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestFleetReport (FLEET_CI_PLAN.MD point 3, prototype) on testdata/fleet.
func TestFleetReport(t *testing.T) {
	arts, err := Load("../../testdata/fleet")
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 5 {
		t.Fatalf("%d artifacts, want 5", len(arts))
	}
	var b bytes.Buffer
	Render(&b, arts)
	out := b.String()
	for _, want := range []string{
		"5 artifacts: 3 libraries, 1 applications, 1 failed analyses.",
		// lib-a: 2 REACHABLE, 0 FRAMEWORK, 1 NOT_REACHABLE, 0 UNKNOWN; 1 library / 1 user;
		// 1 supplied-object only; max 3 API methods; Spring yes; CVE-B fixed by 1.2.
		"| com.example:lib-a | 2 | 0 | 1 | 0 | 1 / 1 | 1 | 3 | yes | 1 |",
		"| lib-d | 0 | 0 | 1 | 0 | 0 / 0 | 0 | 0 | no | 0 (1 versions not assessed) |",
		"| app-b | 1 | 1 | 0 | 1 |",
		// CVE-A in two artifacts first, then CVE-B (one artifact, CRITICAL).
		"| CVE-A | HIGH | 2 | app-b, com.example:lib-a |\n| CVE-B | CRITICAL | 1 | com.example:lib-a |",
		"- **com.example:app-c**: mvn dependency:copy-dependencies failed (exit 1)",
		"`v0.9.0` (1), `v1.0.0` (3)",
		"| lib-e | 0 | 0 | 0 | 0 | 0 / 0 | 0 | 0 | no | — |",
		"2026-11-20 (2), 2026-11-22 (1)",
		"different tool versions or OSV data dates",
		"not proof of safety",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, "0001-01-01") {
		t.Error("a zero OSV validation time is reported as a date")
	}
}
