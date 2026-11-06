package component

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/classfile"
)

// CVE:F7a fixtures: the jpms "auto" classes relocated from
// fixture/jpms/auto to shaded/auto, as a shading tool would.

const (
	shadedCodec  = "shaded/auto/AutoCodec.class"
	shadedHasher = "shaded/auto/AutoHasher.class"
)

func relocated(t *testing.T, b []byte) []byte {
	t.Helper()
	r, err := classfile.RelocateClassForTesting(b, "fixture/jpms/auto", "shaded/auto")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func fileSHA(t *testing.T, p string) string {
	t.Helper()
	s := sha256.Sum256(readFile(t, p))
	return hex.EncodeToString(s[:])
}

func writeDecl(t *testing.T, body string) *Declarations {
	t.Helper()
	p := filepath.Join(t.TempDir(), "decl.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := LoadDeclarations(p)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

const shadePOM = `<project><build><plugins>
  <plugin><artifactId>maven-compiler-plugin</artifactId><configuration><relocations><relocation>
    <pattern>fixture.jpms.auto</pattern><shadedPattern>wrong.place</shadedPattern></relocation></relocations></configuration></plugin>
  <plugin><groupId>org.apache.maven.plugins</groupId><artifactId>maven-shade-plugin</artifactId>
    <executions><execution><configuration><relocations>
      <relocation><pattern>fixture.jpms.auto</pattern><shadedPattern>shaded.auto</shadedPattern></relocation>
      <relocation><pattern>${unresolved}</pattern><shadedPattern>x.y</shadedPattern></relocation>
    </relocations></configuration></execution></executions></plugin>
</plugins></build></project>`

// relocatedHost writes org.host:fat:1.0 with the relocated classes, the
// embedded candidate's pom.properties and, when pom is set, its own
// pom.xml.
func relocatedHost(t *testing.T, path, pomXML string) {
	t.Helper()
	codec, hasher := classBytes(t)
	names := []string{"META-INF/maven/org.host/fat/pom.properties", shadedCodec, shadedHasher, "META-INF/maven/org.other/embedded/pom.properties"}
	entries := map[string][]byte{
		names[0]: pom("org.host", "fat", "1.0"), shadedCodec: relocated(t, codec), shadedHasher: relocated(t, hasher),
		names[3]: pom("org.other", "embedded", "2.3"),
	}
	if pomXML != "" {
		names = append(names, "META-INF/maven/org.host/fat/pom.xml")
		entries[names[len(names)-1]] = []byte(pomXML)
	}
	writeJar(t, path, names, entries)
}

func refCentral(t *testing.T) *fakeCentral {
	codec, hasher := classBytes(t)
	return &fakeCentral{jars: map[string][]byte{"org.other:embedded:2.3": jarBytes(t, map[string][]byte{codecEntry: codec, hasherEntry: hasher})}}
}

// TestRelocationDeclared: without a relocation the candidate's classes
// are absent; declared for the jar's SHA-256, they are found under the
// relocated names, compared through the inverse relocation, and owned by
// the embedded instance, which records the rule and its source.
func TestRelocationDeclared(t *testing.T) {
	deps := t.TempDir()
	jar := filepath.Join(deps, "fat-1.0.jar")
	relocatedHost(t, jar, "")

	inv, h := buildWithH(t, deps)
	if err := inv.ResolveEmbedded(context.Background(), refCentral(t).fetch, "https://central.test", h, nil); err != nil {
		t.Fatal(err)
	}
	if e := embeddedByGAV(inv, "org.other:embedded:2.3"); e.Status != EmbeddedClassesAbsent || !strings.Contains(e.Reason, "declare the relocation") {
		t.Fatalf("undeclared: %+v", e)
	}

	inv, h = buildWithH(t, deps)
	d := writeDecl(t, `version: 1
jars:
  - sha256: `+fileSHA(t, jar)+`
    components:
      - gav: org.other:embedded:2.3
        relocations:
          - {from: fixture.jpms.auto, to: shaded.auto}
`)
	if err := inv.ResolveEmbedded(context.Background(), refCentral(t).fetch, "https://central.test", h, d); err != nil {
		t.Fatal(err)
	}
	e := embeddedByGAV(inv, "org.other:embedded:2.3")
	if e.Status != EmbeddedConfirmed || e.Identical != 2 || !e.Declared || len(e.Relocations) != 1 || e.Relocations[0].To != "shaded.auto" || !strings.HasPrefix(e.Relocations[0].Source, "declared") {
		t.Fatalf("declared: %+v", e)
	}
	in := instanceByID(inv, e.Instance)
	if in.Classes != 2 || len(in.Relocations) != 1 {
		t.Errorf("instance = %+v", in)
	}
	for _, ci := range h.All() {
		if strings.HasPrefix(ci.FQCN, "shaded.auto.") && inv.InstanceForClass(ci) != in.ID {
			t.Errorf("%s owned by %s", ci.FQCN, inv.InstanceForClass(ci))
		}
	}
}

// TestRelocationFromBuildConfig: the host's own pom.xml carries a
// maven-shade relocation (a relocation of another plugin and a
// placeholder pattern are ignored): no declaration is needed.
func TestRelocationFromBuildConfig(t *testing.T) {
	deps := t.TempDir()
	relocatedHost(t, filepath.Join(deps, "fat-1.0.jar"), shadePOM)
	inv, h := buildWithH(t, deps)
	if err := inv.ResolveEmbedded(context.Background(), refCentral(t).fetch, "https://central.test", h, nil); err != nil {
		t.Fatal(err)
	}
	e := embeddedByGAV(inv, "org.other:embedded:2.3")
	if e.Status != EmbeddedConfirmed || e.Declared || len(e.Relocations) != 1 || !strings.HasPrefix(e.Relocations[0].Source, "build-config: META-INF/maven/org.host/fat/pom.xml") {
		t.Fatalf("build config: %+v", e)
	}
	notes := strings.Join(inv.BuildConfigNotes, "\n")
	if !strings.Contains(notes, "${unresolved}") || !strings.Contains(notes, "1 maven-shade relocation(s)") {
		t.Errorf("notes = %s", notes)
	}
}

// TestRelocationNegative: a wrong map finds nothing; a declaration for
// another hash is reported and ignored; a build-config rule contradicted
// by a declaration is a conflict; a minimized relocated copy is partial.
func TestRelocationNegative(t *testing.T) {
	deps := t.TempDir()
	jar := filepath.Join(deps, "fat-1.0.jar")
	relocatedHost(t, jar, "")
	sha := fileSHA(t, jar)

	inv, h := buildWithH(t, deps)
	d := writeDecl(t, `version: 1
jars:
  - sha256: `+sha+`
    components:
      - gav: org.other:embedded:2.3
        relocations: [{from: fixture.jpms.auto, to: shaded.elsewhere}]
  - sha256: `+strings.Repeat("0", 64)+`
    name: fat-1.0.jar
    components: [{gav: "org.x:y:1"}]
  - sha256: `+strings.Repeat("1", 64)+`
    components: [{gav: "org.x:y:1"}]
`)
	if err := inv.ResolveEmbedded(context.Background(), refCentral(t).fetch, "https://central.test", h, d); err != nil {
		t.Fatal(err)
	}
	if e := embeddedByGAV(inv, "org.other:embedded:2.3"); e.Status != EmbeddedClassesAbsent || e.Instance != "" {
		t.Errorf("wrong map: %+v", e)
	}
	issues := strings.Join(inv.DeclarationIssues, "\n")
	if !strings.Contains(issues, "fat-1.0.jar ignored: hash mismatch") || !strings.Contains(issues, "no analyzed dependency jar has that hash") || len(inv.DeclarationIssues) != 2 {
		t.Errorf("issues = %s", issues)
	}
	if embeddedByGAV(inv, "org.x:y:1").Kind != "" {
		t.Error("a declaration for another jar created a candidate")
	}

	deps2 := t.TempDir()
	jar2 := filepath.Join(deps2, "fat-1.0.jar")
	relocatedHost(t, jar2, shadePOM)
	inv, h = buildWithH(t, deps2)
	d = writeDecl(t, "version: 1\njars:\n  - sha256: "+fileSHA(t, jar2)+"\n    components:\n      - gav: org.other:embedded:2.3\n        relocations: [{from: fixture.jpms.auto, to: other.auto}]\n")
	inv.ResolveEmbedded(context.Background(), refCentral(t).fetch, "https://central.test", h, d)
	if e := embeddedByGAV(inv, "org.other:embedded:2.3"); e.Status != EmbeddedRelocationConflict {
		t.Errorf("conflict: %+v", e)
	}

	// minimized: one of the two classes kept
	deps3 := t.TempDir()
	codec, _ := classBytes(t)
	writeJar(t, filepath.Join(deps3, "fat-1.0.jar"), []string{"META-INF/maven/org.host/fat/pom.properties", shadedCodec, "META-INF/maven/org.host/fat/pom.xml", "META-INF/maven/org.other/embedded/pom.properties"},
		map[string][]byte{"META-INF/maven/org.host/fat/pom.properties": pom("org.host", "fat", "1.0"), shadedCodec: relocated(t, codec),
			"META-INF/maven/org.host/fat/pom.xml": []byte(shadePOM), "META-INF/maven/org.other/embedded/pom.properties": pom("org.other", "embedded", "2.3")})
	inv, h = buildWithH(t, deps3)
	c := refCentral(t)
	codecB, hasher := classBytes(t)
	c.jars["org.other:embedded:2.3"] = jarBytes(t, map[string][]byte{codecEntry: codecB, hasherEntry: hasher, "fixture/jpms/auto/Third.class": hasher})
	inv.ResolveEmbedded(context.Background(), c.fetch, "https://central.test", h, nil)
	if e := embeddedByGAV(inv, "org.other:embedded:2.3"); e.Status != EmbeddedPartial || e.Present != 1 || !strings.Contains(e.Reason, "not proof") {
		t.Errorf("minimized: %+v", e)
	}
}

// TestRelocationCollision: two candidates whose (relocated) classes are
// the same entries cannot both own them: the second is a collision.
func TestRelocationCollision(t *testing.T) {
	deps := t.TempDir()
	codec, hasher := classBytes(t)
	writeJar(t, filepath.Join(deps, "fat-1.0.jar"),
		[]string{"META-INF/maven/org.host/fat/pom.properties", shadedCodec, shadedHasher, "META-INF/maven/org.host/fat/pom.xml", "META-INF/maven/org.a/one/pom.properties", "META-INF/maven/org.b/two/pom.properties"},
		map[string][]byte{"META-INF/maven/org.host/fat/pom.properties": pom("org.host", "fat", "1.0"), shadedCodec: relocated(t, codec), shadedHasher: relocated(t, hasher),
			"META-INF/maven/org.host/fat/pom.xml": []byte(shadePOM), "META-INF/maven/org.a/one/pom.properties": pom("org.a", "one", "1"), "META-INF/maven/org.b/two/pom.properties": pom("org.b", "two", "1")})
	inv, h := buildWithH(t, deps)
	ref := jarBytes(t, map[string][]byte{codecEntry: codec, hasherEntry: hasher})
	c := &fakeCentral{jars: map[string][]byte{"org.a:one:1": ref, "org.b:two:1": ref}}
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, "https://central.test", h, nil); err != nil {
		t.Fatal(err)
	}
	one, two := embeddedByGAV(inv, "org.a:one:1"), embeddedByGAV(inv, "org.b:two:1")
	if one.Status != EmbeddedConfirmed || two.Status != EmbeddedCollision || two.Instance != "" {
		t.Errorf("one %+v\ntwo %+v", one, two)
	}
}

// TestLoadDeclarationsRejects: a bad hash, GAV, version, rule or unknown
// field stops the run.
func TestLoadDeclarationsRejects(t *testing.T) {
	sha := strings.Repeat("a", 64)
	for name, body := range map[string]string{
		"version": "version: 2\njars: []\n",
		"sha":     "version: 1\njars: [{sha256: abc, components: []}]\n",
		"gav":     "version: 1\njars: [{sha256: " + sha + ", components: [{gav: a:b}]}]\n",
		"rule":    "version: 1\njars: [{sha256: " + sha + ", components: [{gav: a:b:1, relocations: [{from: 'a.*', to: b}]}]}]\n",
		"field":   "version: 1\njars: [{sha256: " + sha + ", hash: x}]\n",
	} {
		p := filepath.Join(t.TempDir(), "d.yaml")
		os.WriteFile(p, []byte(body), 0o644)
		if _, err := LoadDeclarations(p); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
