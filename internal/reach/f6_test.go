package reach

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
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
	if err := inv.ResolveEmbedded(context.Background(), fetch, "https://central.test", h, nil); err != nil {
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

// relocateJar copies src to dst with every class relocated from package
// prefix from to to (entry names and constant pools), plus extra entries.
func relocateJar(t *testing.T, src, dst, from, to string, extra map[string]string) {
	t.Helper()
	zr, err := zip.OpenReader(src)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	f, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for _, e := range zr.File {
		if e.FileInfo().IsDir() {
			continue
		}
		r, _ := e.Open()
		b, _ := io.ReadAll(r)
		r.Close()
		name := e.Name
		if strings.HasSuffix(name, ".class") {
			if b, err = classfile.RelocateClassForTesting(b, from, to); err != nil {
				t.Fatal(err)
			}
			name = strings.Replace(name, from+"/", to+"/", 1)
		}
		w, _ := zw.Create(name)
		w.Write(b)
	}
	for n, body := range extra {
		w, _ := zw.Create(n)
		w.Write([]byte(body))
	}
	zw.Close()
	f.Close()
}

// TestRelocatedInstanceFinding (CVE:F7a): the vulnerable library is
// embedded in org.host:fat:5.0 relocated from reachfixture to
// shaded/reachfixture (the application calls it under that name), as the
// host's maven-shade configuration says. The advisory's selector names
// the original method; it is matched under the relocated name and the
// finding is on the embedded instance. Without the rewrite there would be
// no method to find.
func TestRelocatedInstanceFinding(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "app.jar")
	fat := filepath.Join(dir, "fat-5.0.jar")
	relocateJar(t, fixtureDir+"appentry.jar", app, "reachfixture", "shaded/reachfixture", nil)
	relocateJar(t, fixtureDir+"vulnerablelib.jar", fat, "reachfixture", "shaded/reachfixture", map[string]string{
		"META-INF/maven/org.host/fat/pom.properties":       "groupId=org.host\nartifactId=fat\nversion=5.0\n",
		"META-INF/maven/org.vulnerable/lib/pom.properties": "groupId=org.vulnerable\nartifactId=lib\nversion=1.0\n",
		"META-INF/maven/org.host/fat/pom.xml": `<project><build><plugins><plugin><artifactId>maven-shade-plugin</artifactId><configuration><relocations>
<relocation><pattern>reachfixture</pattern><shadedPattern>shaded.reachfixture</shadedPattern></relocation></relocations></configuration></plugin></plugins></build></project>`,
	})
	deps := []input.DepJar{
		{Path: fixtureDir + "frameworkthing.jar", GAV: &input.GAV{Group: "org.framework", Artifact: "thing", Version: "1.0"}, Source: input.SourceDepsList},
		{Path: fat, GAV: &input.GAV{Group: "org.host", Artifact: "fat", Version: "5.0"}, Source: input.SourcePOMProperties},
	}
	h, _, err := hierarchy.Build(app, deps)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := component.Build(component.Inputs{AppInput: app, AppJar: app, DepJars: deps}, h)
	if err != nil {
		t.Fatal(err)
	}
	ref, _ := os.ReadFile(fixtureDir + "vulnerablelib.jar")
	fetch := func(context.Context, string) ([]byte, error) { return ref, nil }
	if err := inv.ResolveEmbedded(context.Background(), fetch, "https://central.test", h, nil); err != nil {
		t.Fatal(err)
	}
	var emb string
	for _, in := range inv.Instances {
		if in.Detection == component.DetectedEmbedded && len(in.Relocations) == 1 {
			emb = in.ID
		}
	}
	if emb == "" {
		t.Fatalf("no relocated embedded instance: %+v", inv.Embedded)
	}
	e := New(callgraph.Build(h, nil), h, NewFrameworkMatcher([]string{"org.framework:*"}))
	e.SetComponents(inv)
	cve := cves.CVE{ID: "ARIADNE-TEST-F7", Affected: []cves.Affected{{Package: "org.vulnerable:lib", Version: "1.0"}}}
	r := e.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: {Methods: []string{"reachfixture.VulnerableLib#vulnerable"}, Source: vulnmap.SourceManual}})[0]
	if r.Status != StatusReachable || len(r.Findings) != 1 || r.Findings[0].Instance != emb || !strings.Contains(strings.Join(r.Path, " "), "shaded.reachfixture.VulnerableLib#vulnerable") {
		t.Fatalf("result %s path %v findings %+v, want REACHABLE on %s via the relocated method", r.Status, r.Path, r.Findings, emb)
	}
}
