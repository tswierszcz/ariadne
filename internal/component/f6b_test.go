package component

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ariadne/internal/input"
)

const testBase = "https://central.test"

func metadata(versions ...string) []byte {
	var b strings.Builder
	b.WriteString("<metadata><versioning><versions>")
	for _, v := range versions {
		fmt.Fprintf(&b, "<version>%s</version>", v)
	}
	b.WriteString("</versions></versioning></metadata>")
	return []byte(b.String())
}

func lib(g, a string) input.GAV { return input.GAV{Group: g, Artifact: a} }

// versionCentral: org.other:embedded with the given published versions;
// fitting versions hold the host's classes, the others different content.
func versionCentral(t *testing.T, published []string, fitting ...string) *fakeCentral {
	codec, hasher := classBytes(t)
	good := jarBytes(t, map[string][]byte{codecEntry: codec, hasherEntry: hasher})
	bad := jarBytes(t, map[string][]byte{codecEntry: hasher, hasherEntry: codec})
	fit := map[string]bool{}
	for _, v := range fitting {
		fit[v] = true
	}
	c := &fakeCentral{jars: map[string][]byte{}, files: map[string][]byte{
		MetadataURL(testBase, lib("org.other", "embedded")): metadata(published...),
	}}
	for _, v := range published {
		if fit[v] {
			c.jars["org.other:embedded:"+v] = good
		} else {
			c.jars["org.other:embedded:"+v] = bad
		}
	}
	return c
}

func resolveHost(t *testing.T, c *fakeCentral, declared string) (*Inventory, Embedded) {
	t.Helper()
	codec, hasher := classBytes(t)
	deps := t.TempDir()
	hostJar(t, filepath.Join(deps, "fat-1.0.jar"), map[string][]byte{codecEntry: padUtf8(t, codec), hasherEntry: hasher}, "org.other:embedded:"+declared)
	inv, h := buildWithH(t, deps)
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, testBase, h, nil); err != nil {
		t.Fatal(err)
	}
	return inv, embeddedByGAV(inv, "org.other:embedded:"+declared)
}

// TestVersionFromClasses: the declared version's classes differ; of the
// neighbouring published versions exactly one fits — the instance gets
// that version, read from its classes, and the record says so.
func TestVersionFromClasses(t *testing.T) {
	c := versionCentral(t, []string{"2.0", "2.2", "2.3", "2.4", "3.0"}, "2.4")
	inv, e := resolveHost(t, c, "2.3")
	if e.Status != EmbeddedConfirmed || e.Version != "2.4" || e.Instance == "" {
		t.Fatalf("candidate = %+v", e)
	}
	if !strings.Contains(e.Reason, "version 2.4 from its classes") || !strings.Contains(e.Reason, "declared version 2.3 does not fit") {
		t.Errorf("reason = %q", e.Reason)
	}
	if len(e.VersionsChecked) != 5 || !e.VersionsChecked[0].Declared || e.VersionsChecked[0].Fits || e.VersionsPublished != 5 {
		t.Errorf("checked = %+v published %d", e.VersionsChecked, e.VersionsPublished)
	}
	in := instanceByID(inv, e.Instance)
	if in.GAV.String() != "org.other:embedded:2.4" || len(in.VersionCandidates) != 0 || in.Classes != 2 {
		t.Errorf("instance = %+v", in)
	}
	if gs := inv.ConfirmedEmbeddedGAVs(); len(gs) != 1 || gs[0].Version != "2.4" {
		t.Errorf("lookup = %v", gs)
	}
}

// TestVersionCandidates: two published versions fit equally — neither is
// picked; the instance carries both, and the advisory lookup asks for
// each.
func TestVersionCandidates(t *testing.T) {
	c := versionCentral(t, []string{"2.2", "2.3", "2.4", "2.5"}, "2.2", "2.5")
	inv, e := resolveHost(t, c, "2.3")
	if e.Status != EmbeddedVersionCandidates || !reflect.DeepEqual(e.VersionCandidates, []string{"2.2", "2.5"}) || e.Instance == "" || e.Version != "" {
		t.Fatalf("candidate = %+v", e)
	}
	in := instanceByID(inv, e.Instance)
	if in.GAV.Version != "" || !reflect.DeepEqual(in.VersionCandidates, []string{"2.2", "2.5"}) || in.Label() != "org.other:embedded:{2.2,2.5}" {
		t.Errorf("instance = %+v label %q", in, in.Label())
	}
	var got []string
	for _, g := range inv.ConfirmedEmbeddedGAVs() {
		got = append(got, g.String())
	}
	if !reflect.DeepEqual(got, []string{"org.other:embedded:2.2", "org.other:embedded:2.5"}) {
		t.Errorf("lookup = %v", got)
	}
	if conf, unconf, _ := inv.EmbeddedCounts(); conf != 1 || unconf != 0 {
		t.Errorf("counts = %d %d", conf, unconf)
	}
}

// TestVersionNoneFits: no checked version fits — version_unknown, no
// instance, no lookup.
func TestVersionNoneFits(t *testing.T) {
	c := versionCentral(t, []string{"2.2", "2.3", "2.4"})
	inv, e := resolveHost(t, c, "2.3")
	if e.Status != EmbeddedVersionUnknown || e.Instance != "" || len(e.VersionsChecked) != 3 || !strings.Contains(e.Reason, "none fits") {
		t.Fatalf("candidate = %+v", e)
	}
	if len(inv.ConfirmedEmbeddedGAVs()) != 0 {
		t.Errorf("lookup = %v", inv.ConfirmedEmbeddedGAVs())
	}
}

// TestVersionDeclaredFitsNoSearch: the declared version fits — F6a
// behaviour, no version list read.
func TestVersionDeclaredFitsNoSearch(t *testing.T) {
	c := versionCentral(t, []string{"2.2", "2.3", "2.4"}, "2.3", "2.4")
	_, e := resolveHost(t, c, "2.3")
	if e.Status != EmbeddedConfirmed || e.Version != "2.3" || c.hits != 1 {
		t.Fatalf("candidate = %+v hits %d", e, c.hits)
	}
}

// TestVersionPartialNotSearched: a minimized copy (partial) is not a
// version question: no search.
func TestVersionPartialNotSearched(t *testing.T) {
	codec, hasher := classBytes(t)
	c := versionCentral(t, []string{"2.3", "2.4"}, "2.4")
	c.jars["org.other:embedded:2.3"] = jarBytes(t, map[string][]byte{codecEntry: codec, hasherEntry: hasher, "a/A.class": codec, "a/B.class": codec, "a/C.class": codec})
	_, e := resolveHost(t, c, "2.3")
	if e.Status != EmbeddedPartial || c.hits != 1 {
		t.Fatalf("candidate = %+v hits %d", e, c.hits)
	}
}

func TestVersionWindow(t *testing.T) {
	pub := []string{"1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7", "1.8", "1.9", "2.0", "2.1"}
	got, _ := versionWindow(pub, "1.5")
	if want := []string{"1.1", "1.2", "1.3", "1.4", "1.6", "1.7", "1.8", "1.9"}; !reflect.DeepEqual(got, want) {
		t.Errorf("around 1.5 = %v, want %v", got, want)
	}
	got, _ = versionWindow(pub, "2.1")
	if want := []string{"1.3", "1.4", "1.5", "1.6", "1.7", "1.8", "1.9", "2.0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("at the end = %v, want %v", got, want)
	}
	// a declared version that is not published: its neighbours
	got, _ = versionWindow([]string{"1.0", "1.2", "1.4"}, "1.3")
	if want := []string{"1.0", "1.2", "1.4"}; !reflect.DeepEqual(got, want) {
		t.Errorf("unpublished = %v", got)
	}
	if got, how := versionWindow(pub, ""); got != nil || !strings.Contains(how, "not searched") {
		t.Errorf("no version, 12 published = %v %q", got, how)
	}
	if got, _ := versionWindow(pub[:8], ""); len(got) != 8 {
		t.Errorf("no version, 8 published = %v", got)
	}
}

func TestPublishedVersionsOrderAndSnapshots(t *testing.T) {
	c := &fakeCentral{files: map[string][]byte{MetadataURL(testBase, lib("g", "a")): metadata("1.10", "1.9", "2.0-SNAPSHOT", "1.9", "2.0-beta-1")}}
	got, err := publishedVersions(context.Background(), c.fetch, testBase, lib("g", "a"))
	if err != nil || !reflect.DeepEqual(got, []string{"1.9", "1.10", "2.0-beta-1"}) {
		t.Errorf("versions = %v %v", got, err)
	}
}

const shadeIncludesPOM = `<project>
  <parent><groupId>org.host</groupId><artifactId>parent</artifactId><version>1.0</version></parent>
  <artifactId>fat</artifactId>
  <dependencies>
    <dependency><groupId>org.other</groupId><artifactId>embedded</artifactId><version>${emb.version}</version></dependency>
    <dependency><groupId>org.other</groupId><artifactId>testonly</artifactId><version>1</version><scope>test</scope></dependency>
    <dependency><groupId>org.other</groupId><artifactId>opt</artifactId><version>1</version><optional>true</optional></dependency>
    <dependency><groupId>org.unrelated</groupId><artifactId>x</artifactId><version>1</version></dependency>
    <dependency><groupId>org.other</groupId><artifactId>managed</artifactId></dependency>
  </dependencies>
  <build><plugins><plugin>
    <artifactId>maven-shade-plugin</artifactId>
    <executions><execution><configuration>
      <artifactSet><includes><include>org.other:*</include></includes><excludes>org.other:excluded</excludes></artifactSet>
    </configuration></execution></executions>
  </plugin></plugins></build>
</project>`

const parentPOM = `<project><groupId>org.host</groupId><artifactId>parent</artifactId><version>1.0</version>
  <properties><emb.version>2.3</emb.version><mgd.version>5</mgd.version></properties>
  <dependencyManagement><dependencies>
    <dependency><groupId>org.other</groupId><artifactId>managed</artifactId><version>${mgd.version}</version></dependency>
  </dependencies></dependencyManagement>
</project>`

// TestBuildConfigCandidates: a jar without pom.properties for its shaded
// dependencies — the shade includes ∩ the pom's dependencies name them,
// the version property is resolved from the parent pom, a managed version
// from dependencyManagement, transitive dependencies are followed through
// their poms; test/optional/unincluded/excluded ones are not candidates.
// Candidates are still confirmed only by their classes.
func TestBuildConfigCandidates(t *testing.T) {
	codec, hasher := classBytes(t)
	deps := t.TempDir()
	writeJar(t, filepath.Join(deps, "fat-1.0.jar"),
		[]string{"META-INF/maven/org.host/fat/pom.properties", "META-INF/maven/org.host/fat/pom.xml", codecEntry, hasherEntry},
		map[string][]byte{"META-INF/maven/org.host/fat/pom.properties": pom("org.host", "fat", "1.0"),
			"META-INF/maven/org.host/fat/pom.xml": []byte(shadeIncludesPOM), codecEntry: codec, hasherEntry: hasher})
	inv, h := buildWithH(t, deps)
	c := &fakeCentral{
		jars: map[string][]byte{
			"org.other:embedded:2.3": jarBytes(t, map[string][]byte{codecEntry: codec}),
			"org.other:tree:7":       jarBytes(t, map[string][]byte{hasherEntry: hasher}),
		},
		files: map[string][]byte{
			POMURL(testBase, input.GAV{Group: "org.host", Artifact: "parent", Version: "1.0"}): []byte(parentPOM),
			POMURL(testBase, input.GAV{Group: "org.other", Artifact: "embedded", Version: "2.3"}): []byte(`<project><groupId>org.other</groupId><artifactId>embedded</artifactId><version>2.3</version>
<dependencies><dependency><groupId>org.other</groupId><artifactId>tree</artifactId><version>7</version></dependency>
<dependency><groupId>org.other</groupId><artifactId>excluded</artifactId><version>1</version></dependency></dependencies></project>`),
		},
	}
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, testBase, h, nil); err != nil {
		t.Fatal(err)
	}
	got := map[string]Embedded{}
	for _, e := range inv.Embedded {
		got[e.GAV.String()] = e
	}
	for _, g := range []string{"org.other:testonly:1", "org.other:opt:1", "org.unrelated:x:1", "org.other:excluded:1"} {
		if _, ok := got[g]; ok {
			t.Errorf("%s is a candidate", g)
		}
	}
	emb, tree, mgd := got["org.other:embedded:2.3"], got["org.other:tree:7"], got["org.other:managed:5"]
	if emb.Source != SourceBuildConfig || emb.Status != EmbeddedConfirmed || emb.Instance == "" {
		t.Errorf("embedded = %+v", emb)
	}
	if tree.Status != EmbeddedConfirmed || !strings.Contains(tree.SourceDetail, "org.other:embedded -> org.other:tree") {
		t.Errorf("tree = %+v", tree)
	}
	if mgd.Source != SourceBuildConfig || mgd.Status != EmbeddedReferenceUnavailable || mgd.Instance != "" {
		t.Errorf("managed = %+v", mgd)
	}
	in := instanceByID(inv, emb.Instance)
	if loc := inv.InstanceLocation(in.ID); loc != "deps:fat-1.0.jar!/build-config/org.other/embedded" || in.Evidence[0].Kind != "build-config" {
		t.Errorf("instance %+v at %s", in, loc)
	}
	if host := instanceByID(inv, emb.HostInstance); host.Classes != 0 {
		t.Errorf("host keeps %d classes", host.Classes)
	}
}

// TestBuildConfigWithoutIncludes: "shade everything" names every
// dependency — no candidates, a note.
func TestBuildConfigWithoutIncludes(t *testing.T) {
	codec, _ := classBytes(t)
	deps := t.TempDir()
	noInc := strings.Replace(shadeIncludesPOM, "<includes><include>org.other:*</include></includes>", "", 1)
	writeJar(t, filepath.Join(deps, "fat-1.0.jar"),
		[]string{"META-INF/maven/org.host/fat/pom.properties", "META-INF/maven/org.host/fat/pom.xml", codecEntry},
		map[string][]byte{"META-INF/maven/org.host/fat/pom.properties": pom("org.host", "fat", "1.0"),
			"META-INF/maven/org.host/fat/pom.xml": []byte(noInc), codecEntry: codec})
	inv, h := buildWithH(t, deps)
	c := &fakeCentral{}
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, testBase, h, nil); err != nil {
		t.Fatal(err)
	}
	if len(inv.Embedded) != 0 || c.hits != 0 || !strings.Contains(strings.Join(inv.BuildConfigNotes, "\n"), "no artifactSet includes") {
		t.Errorf("embedded %+v hits %d notes %v", inv.Embedded, c.hits, inv.BuildConfigNotes)
	}
}

// TestBuildConfigUnresolvedVersion: the property is nowhere (parent
// unavailable) — the candidate has no version; with more published
// versions than the search bound it stays version_unknown, no jar read.
func TestBuildConfigUnresolvedVersion(t *testing.T) {
	codec, _ := classBytes(t)
	deps := t.TempDir()
	writeJar(t, filepath.Join(deps, "fat-1.0.jar"),
		[]string{"META-INF/maven/org.host/fat/pom.properties", "META-INF/maven/org.host/fat/pom.xml", codecEntry},
		map[string][]byte{"META-INF/maven/org.host/fat/pom.properties": pom("org.host", "fat", "1.0"),
			"META-INF/maven/org.host/fat/pom.xml": []byte(shadeIncludesPOM), codecEntry: codec})
	inv, h := buildWithH(t, deps)
	c := &fakeCentral{files: map[string][]byte{
		MetadataURL(testBase, lib("org.other", "embedded")): metadata("1", "2", "3", "4", "5", "6", "7", "8", "9"),
	}}
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, testBase, h, nil); err != nil {
		t.Fatal(err)
	}
	e := embeddedByGAV(inv, "org.other:embedded:")
	if e.Status != EmbeddedVersionUnknown || e.Instance != "" || len(e.VersionsChecked) != 0 || !strings.Contains(e.Reason, "not searched") {
		t.Errorf("candidate = %+v", e)
	}
	for _, u := range c.urls {
		if strings.HasSuffix(u, ".jar") {
			t.Errorf("fetched %s", u)
		}
	}
}

func writeSBOM(t *testing.T, body string) *SBOM {
	t.Helper()
	p := filepath.Join(t.TempDir(), "bom.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadSBOM(p)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestSBOMCandidates: the SBOM component whose SHA-256 is the host jar's
// is tied to it; its nested component becomes a candidate confirmed by
// its classes; a component nested in an untied one, and a component
// listed at top level with no hash, are not used; a disagreeing identity
// is an issue, never a replacement.
func TestSBOMCandidates(t *testing.T) {
	codec, hasher := classBytes(t)
	deps := t.TempDir()
	jar := filepath.Join(deps, "fat-1.0.jar")
	writeJar(t, jar, []string{"META-INF/maven/org.host/fat/pom.properties", codecEntry, hasherEntry},
		map[string][]byte{"META-INF/maven/org.host/fat/pom.properties": pom("org.host", "fat", "1.0"), codecEntry: codec, hasherEntry: hasher})
	sha := fileSHA(t, jar)
	s := writeSBOM(t, `{"bomFormat":"CycloneDX","specVersion":"1.5","serialNumber":"urn:uuid:1",
 "metadata":{"component":{"type":"application","name":"app"}},
 "components":[
  {"type":"library","bom-ref":"fat","purl":"pkg:maven/org.host/fat@1.1","hashes":[{"alg":"SHA-256","content":"`+sha+`"}],
   "components":[{"type":"library","bom-ref":"emb","purl":"pkg:maven/org.other/embedded@2.3"}]},
  {"type":"library","bom-ref":"untied","purl":"pkg:maven/org.x/y@1","components":[{"type":"library","purl":"pkg:maven/org.x/z@1"}]},
  {"type":"library","purl":"pkg:maven/org.other/listed@9"}
 ]}`)
	inv, h := buildWithH(t, deps)
	inv.ApplySBOM(s)
	c := &fakeCentral{jars: map[string][]byte{"org.other:embedded:2.3": jarBytes(t, map[string][]byte{codecEntry: codec, hasherEntry: hasher})}}
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, testBase, h, nil); err != nil {
		t.Fatal(err)
	}
	if len(inv.Embedded) != 1 {
		t.Fatalf("embedded = %+v", inv.Embedded)
	}
	e := inv.Embedded[0]
	if e.Source != SourceSBOM || e.Status != EmbeddedConfirmed || e.Instance == "" || !strings.Contains(e.SourceDetail, "emb nested") {
		t.Errorf("candidate = %+v", e)
	}
	sb := inv.SBOM
	if sb.Components != 6 || sb.Linked != 1 || sb.Candidates != 1 || sb.Unlinked != 3 {
		t.Errorf("summary = %+v", sb)
	}
	issues := strings.Join(sb.Issues, "\n")
	if !strings.Contains(issues, "disagrees with the identity read from the jar") || !strings.Contains(issues, "3 Maven component(s) are not tied") {
		t.Errorf("issues = %v", sb.Issues)
	}
	host := instanceByID(inv, e.HostInstance)
	if host.GAV.String() != "org.host:fat:1.0" {
		t.Errorf("host identity replaced: %s", host.GAV)
	}
	found := false
	for _, ev := range host.Evidence {
		found = found || ev.Kind == "sbom"
	}
	if !found {
		t.Errorf("no sbom evidence on host: %+v", host.Evidence)
	}
}

func TestLoadSBOMRejects(t *testing.T) {
	for name, body := range map[string]string{
		"spdx":     `{"spdxVersion":"SPDX-2.3"}`,
		"old spec": `{"bomFormat":"CycloneDX","specVersion":"1.2"}`,
		"not json": `<bom/>`,
	} {
		p := filepath.Join(t.TempDir(), "bom.json")
		os.WriteFile(p, []byte(body), 0o644)
		if _, err := LoadSBOM(p); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestMavenPURL(t *testing.T) {
	for in, want := range map[string]string{
		"pkg:maven/org.ow2.asm/asm@9.0":          "org.ow2.asm:asm:9.0",
		"pkg:maven/org.ow2.asm/asm@9.0?type=jar": "org.ow2.asm:asm:9.0",
		"pkg:maven/org.a/b@1%2B2":                "org.a:b:1+2",
		"pkg:maven/org.a/b@1?classifier=sources": "",
		"pkg:npm/left-pad@1.0":                   "",
		"pkg:maven/b@1":                          "",
	} {
		g, ok := mavenPURL(in)
		if got := ""; ok {
			got = g.String()
			if got != want {
				t.Errorf("%s = %s, want %s", in, got, want)
			}
		} else if want != "" {
			t.Errorf("%s rejected, want %s", in, want)
		}
	}
}
