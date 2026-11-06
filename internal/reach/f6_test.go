package reach

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/vulnmap"
)

// TestEmbeddedInstanceFinding (CVE:F6a): the vulnerable library is not a
// jar of its own but embedded, with its packages, in org.host:fat:5.0.
// Once its classes confirm it, an advisory against org.vulnerable:lib:1.0
// gets a finding on the embedded instance — not on the host, and not
// "no instance".
func TestEmbeddedInstanceFinding(t *testing.T) {
	dir := t.TempDir()
	fat := filepath.Join(dir, "fat-5.0.jar")
	zr, err := zip.OpenReader(fixtureDir + "vulnerablelib.jar")
	if err != nil {
		t.Fatal(err)
	}
	f, _ := os.Create(fat)
	zw := zip.NewWriter(f)
	for _, e := range zr.File {
		if e.FileInfo().IsDir() {
			continue
		}
		if err := zw.Copy(e); err != nil {
			t.Fatal(err)
		}
	}
	for path, body := range map[string]string{
		"META-INF/maven/org.host/fat/pom.properties":       "groupId=org.host\nartifactId=fat\nversion=5.0\n",
		"META-INF/maven/org.vulnerable/lib/pom.properties": "groupId=org.vulnerable\nartifactId=lib\nversion=1.0\n",
	} {
		w, _ := zw.Create(path)
		w.Write([]byte(body))
	}
	zw.Close()
	f.Close()
	zr.Close()

	deps := []input.DepJar{
		{Path: fixtureDir + "frameworkthing.jar", GAV: &input.GAV{Group: "org.framework", Artifact: "thing", Version: "1.0"}, Source: input.SourceDepsList},
		{Path: fat, GAV: &input.GAV{Group: "org.host", Artifact: "fat", Version: "5.0"}, Source: input.SourcePOMProperties},
	}
	h, _, err := hierarchy.Build(fixtureDir+"appentry.jar", deps)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := component.Build(component.Inputs{AppInput: fixtureDir + "appentry.jar", AppJar: fixtureDir + "appentry.jar", DepJars: deps}, h)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := os.ReadFile(fixtureDir + "vulnerablelib.jar")
	if err != nil {
		t.Fatal(err)
	}
	fetch := func(_ context.Context, url string) ([]byte, error) { return ref, nil }
	if err := inv.ResolveEmbedded(context.Background(), fetch, "https://central.test", h); err != nil {
		t.Fatal(err)
	}
	var emb, host string
	for _, in := range inv.Instances {
		switch {
		case in.Detection == component.DetectedEmbedded:
			emb = in.ID
		case in.GAV != nil && in.GAV.Artifact == "fat":
			host = in.ID
		}
	}
	if emb == "" {
		t.Fatalf("no embedded instance: %+v", inv.Embedded)
	}

	e := New(callgraph.Build(h, nil), h, NewFrameworkMatcher([]string{"org.framework:*"}))
	e.SetComponents(inv)
	cve := cves.CVE{ID: "ARIADNE-TEST-F6", Affected: []cves.Affected{{Package: "org.vulnerable:lib", Version: "1.0"}}}
	r := e.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: {Methods: []string{"reachfixture.VulnerableLib#vulnerable"}, Source: vulnmap.SourceManual}})[0]
	if r.Status != StatusReachable || len(r.Findings) != 1 || r.Findings[0].Instance != emb || r.Findings[0].Status != StatusReachable {
		t.Fatalf("result %s findings %+v, want one REACHABLE finding on %s (host %s)", r.Status, r.Findings, emb, host)
	}
}
