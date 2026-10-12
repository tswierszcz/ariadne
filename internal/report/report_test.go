package report

import (
	"os"
	"strings"
	"testing"
	"time"

	"ariadne/internal/cves"
	"ariadne/internal/reach"
)

func TestBuildAndRenderMarkdown(t *testing.T) {
	list := []cves.CVE{
		{ID: "CVE-2021-39144", Severity: "CVSS_V3:9.8", Affected: []cves.Affected{{Package: "com.thoughtworks.xstream:xstream", Version: "1.4.5"}}},
		{ID: "CVE-2099-0002", Severity: "CVSS_V3:5.0", Affected: []cves.Affected{{Package: "org.example:thing", Version: "1.0"}}},
		{ID: "CVE-2099-0003"},
		{ID: "CVE-2099-0004"},
	}
	results := []reach.Result{
		{
			CVE: "CVE-2021-39144", Status: reach.StatusReachable,
			Methods: []string{"com.thoughtworks.xstream.XStream#fromXML"}, Source: "manual",
			Path:  []string{"org.owasp.webgoat.VulnerableComponentsLesson#completed(Ljava/lang/String;)...", "com.thoughtworks.xstream.XStream#fromXML(...)..."},
			Entry: &reach.EntryEdge{From: "org.owasp.webgoat.VulnerableComponentsLesson#completed(...)...", To: "com.thoughtworks.xstream.XStream#fromXML(...)..."},
		},
		{CVE: "CVE-2099-0002", Status: reach.StatusFramework, Methods: []string{"org.example.Thing#risky"}, Source: "diff"},
		{CVE: "CVE-2099-0003", Status: reach.StatusNotReachable, Methods: []string{"org.example.Other#m"}, Source: "manual"},
		{CVE: "CVE-2099-0004", Status: reach.StatusUnknown},
	}

	h := Header{
		App: "test-app", GeneratedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		AriadneVersion: "v0.0.0-test", Profile: "generic", MaxBytecodeVer: 65,
		GraphNodes: 100, GraphEdges: 200, PreCHAEdges: 150, CHAEdges: 50,
		UnresolvedDyn: 1, DepsTotal: 10, DepsIdentified: 9,
		UnidentifiedJar: []string{"mystery-1.0.jar"},
	}
	data := Build(h, list, results, []string{"some warning"})

	if data.TotalCVEs != 4 {
		t.Fatalf("TotalCVEs = %d, want 4", data.TotalCVEs)
	}
	if data.StatusCounts["REACHABLE"] != 1 || data.StatusCounts["FRAMEWORK"] != 1 ||
		data.StatusCounts["NOT_REACHABLE"] != 1 || data.StatusCounts["UNKNOWN"] != 1 {
		t.Fatalf("StatusCounts = %+v", data.StatusCounts)
	}
	if len(data.Reachable) != 1 || len(data.Framework) != 1 {
		t.Fatalf("Reachable=%d Framework=%d, want 1 and 1", len(data.Reachable), len(data.Framework))
	}

	md, err := RenderMarkdown(data)
	if err != nil {
		t.Fatalf("RenderMarkdown: %v", err)
	}

	for _, want := range []string{
		"test-app",
		"CVE-2021-39144",
		"XStream#fromXML",
		"mystery-1.0.jar",
		"some warning",
		"## Limitations",
		"CHA over-approximates",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("rendered report missing %q\n--- report ---\n%s", want, md)
		}
	}
}

func TestBuildAndRenderMarkdown_emptyResultsDoNotPanic(t *testing.T) {
	data := Build(Header{App: "empty"}, nil, nil, nil)
	md, err := RenderMarkdown(data)
	if err != nil {
		t.Fatalf("RenderMarkdown: %v", err)
	}
	if !strings.Contains(md, "0 CVEs analyzed") {
		t.Errorf("expected a 0-CVE summary, got:\n%s", md)
	}
}

func TestWriteResultsJSON_deterministicOrder(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/results.json"
	results := []reach.Result{
		{CVE: "CVE-2", Methods: []string{"z", "a"}},
		{CVE: "CVE-1", Methods: []string{"b", "a"}},
	}
	if err := WriteResultsJSON(path, results); err != nil {
		t.Fatalf("WriteResultsJSON: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading result: %v", err)
	}
	s := string(b)
	// CVE-1 must appear before CVE-2, and within each, methods sorted.
	if strings.Index(s, `"CVE-1"`) > strings.Index(s, `"CVE-2"`) {
		t.Errorf("results not sorted by CVE id:\n%s", s)
	}
	if strings.Index(s, `"a"`) > strings.Index(s, `"b"`) {
		t.Errorf("methods not sorted within a result:\n%s", s)
	}
}
