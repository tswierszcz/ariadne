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
	"ariadne/internal/mavenversion"
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

// candidatesInstance builds org.host:fat:5.0 embedding the vulnerable
// library with no pom.properties of its own: the host's pom.xml names it
// (shade includes ∩ dependencies) without a resolvable version, and both
// published versions 1.0 and 1.1 hold exactly its classes — a
// version-candidates instance {1.0,1.1} (CVE:F6b).
func candidatesInstance(t *testing.T) (*hierarchy.Hierarchy, *component.Inventory, string) {
	t.Helper()
	dir := t.TempDir()
	fat := filepath.Join(dir, "fat-5.0.jar")
	relocateJar(t, fixtureDir+"vulnerablelib.jar", fat, "unused-prefix", "unused-prefix", map[string]string{
		"META-INF/maven/org.host/fat/pom.properties": "groupId=org.host\nartifactId=fat\nversion=5.0\n",
		"META-INF/maven/org.host/fat/pom.xml": `<project><groupId>org.host</groupId><artifactId>fat</artifactId><version>5.0</version>
<dependencies><dependency><groupId>org.vulnerable</groupId><artifactId>lib</artifactId><version>${lib.version}</version></dependency></dependencies>
<build><plugins><plugin><artifactId>maven-shade-plugin</artifactId><configuration><artifactSet><includes><include>org.vulnerable:lib</include></includes></artifactSet></configuration></plugin></plugins></build></project>`,
	})
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
	ref, _ := os.ReadFile(fixtureDir + "vulnerablelib.jar")
	fetch := func(_ context.Context, url string) ([]byte, error) {
		switch {
		case strings.HasSuffix(url, "/org/vulnerable/lib/maven-metadata.xml"):
			return []byte("<metadata><versioning><versions><version>1.0</version><version>1.1</version></versions></versioning></metadata>"), nil
		case strings.HasSuffix(url, ".jar"):
			return ref, nil
		}
		return nil, os.ErrNotExist
	}
	if err := inv.ResolveEmbedded(context.Background(), fetch, "https://central.test", h, nil); err != nil {
		t.Fatal(err)
	}
	for _, in := range inv.Instances {
		if in.Detection == component.DetectedEmbedded && len(in.VersionCandidates) == 2 {
			return h, inv, in.ID
		}
	}
	t.Fatalf("no version-candidates instance: %+v", inv.Embedded)
	return nil, nil, ""
}

// TestVersionCandidatesFinding (CVE:F6b): an advisory affecting one of
// the candidate versions gets a finding on the instance, qualified by
// VERSION_CANDIDATES; one affecting neither gets none.
func TestVersionCandidatesFinding(t *testing.T) {
	h, inv, emb := candidatesInstance(t)
	e := New(callgraph.Build(h, nil), h, NewFrameworkMatcher([]string{"org.framework:*"}))
	e.SetComponents(inv)
	m := vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#vulnerable"}, Source: vulnmap.SourceManual}
	hit := cves.CVE{ID: "ARIADNE-TEST-F6B", Affected: []cves.Affected{{Package: "org.vulnerable:lib", Version: "1.0"}}}
	miss := cves.CVE{ID: "ARIADNE-TEST-F6B-MISS", Affected: []cves.Affected{{Package: "org.vulnerable:lib", Version: "0.9"}}}
	rs := e.Analyze([]cves.CVE{hit, miss}, map[string]vulnmap.Entry{hit.ID: m, miss.ID: m})
	r := rs[0]
	if r.CVE != hit.ID {
		r, rs[1] = rs[1], rs[0]
	}
	if r.Status != StatusReachable || len(r.Findings) != 1 || r.Findings[0].Instance != emb ||
		len(r.Findings[0].Reasons) == 0 || r.Findings[0].Reasons[0] != "VERSION_CANDIDATES" || r.Findings[0].GAV != "org.vulnerable:lib:{1.0,1.1}" {
		t.Fatalf("hit: %s %+v", r.Status, r.Findings)
	}
	if len(rs[1].Findings) != 0 {
		t.Errorf("miss has findings %+v", rs[1].Findings)
	}
}

// TestVersionCandidatesVariants (CVE:F6b): a versioned mapping is used
// for a version-candidates instance only when every candidate selects
// the same variant; otherwise VERSION_UNKNOWN, never a pick.
func TestVersionCandidatesVariants(t *testing.T) {
	h, inv, emb := candidatesInstance(t)
	e := New(callgraph.Build(h, nil), h, NewFrameworkMatcher([]string{"org.framework:*"}))
	e.SetComponents(inv)
	cve := cves.CVE{ID: "ARIADNE-TEST-F6B-V", Affected: []cves.Affected{{Package: "org.vulnerable:lib", Version: "1.1"}}}
	variants := func(specs ...string) vulnmap.Entry {
		en := vulnmap.Entry{Methods: []string{"reachfixture.VulnerableLib#vulnerable"}, Source: vulnmap.SourceManual}
		for _, s := range specs {
			r, err := mavenversion.ParseRange(s)
			if err != nil {
				t.Fatal(err)
			}
			en.Variants = append(en.Variants, vulnmap.Variant{Versions: s, Range: r, Methods: []string{"reachfixture.VulnerableLib#vulnerable"}})
		}
		return en
	}
	both := e.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: variants("[1.0,2.0)")})[0]
	if len(both.Findings) != 1 || both.Findings[0].Instance != emb || both.Findings[0].Status != StatusReachable {
		t.Errorf("one variant for both: %+v", both.Findings)
	}
	split := e.Analyze([]cves.CVE{cve}, map[string]vulnmap.Entry{cve.ID: variants("[1.1,2.0)")})[0]
	if len(split.Findings) != 1 || split.Findings[0].Status != StatusUnknown || split.Findings[0].Reasons[0] != "VERSION_CANDIDATES" || split.Findings[0].Reasons[1] != "VERSION_UNKNOWN" {
		t.Errorf("variant for one candidate: %+v", split.Findings)
	}
}
