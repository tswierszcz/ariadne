package component

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const (
	codecEntry  = "fixture/jpms/auto/AutoCodec.class"
	hasherEntry = "fixture/jpms/auto/AutoHasher.class"
)

// hostJar writes a host jar org.host:fat:1.0 holding classes and the
// given embedded pom.properties (group:artifact:version).
func hostJar(t *testing.T, path string, classes map[string][]byte, embedded ...string) {
	t.Helper()
	names := []string{"META-INF/maven/org.host/fat/pom.properties"}
	entries := map[string][]byte{names[0]: pom("org.host", "fat", "1.0")}
	for n, b := range classes {
		names = append(names, n)
		entries[n] = b
	}
	for _, gav := range embedded {
		p := strings.Split(gav, ":")
		n := "META-INF/maven/" + p[0] + "/" + p[1] + "/pom.properties"
		names = append(names, n)
		entries[n] = pom(p[0], p[1], p[2])
	}
	writeJar(t, path, names, entries)
}

func jarBytes(t *testing.T, classes map[string][]byte) []byte {
	t.Helper()
	p := filepath.Join(t.TempDir(), "ref.jar")
	var names []string
	for n := range classes {
		names = append(names, n)
	}
	writeJar(t, p, names, classes)
	return readFile(t, p)
}

// fakeCentral serves reference jars by URL suffix and counts requests.
type fakeCentral struct {
	jars  map[string][]byte // "group:artifact:version" -> jar
	files map[string][]byte // any other URL (poms, maven-metadata.xml)
	hits  int
	urls  []string
}

func (f *fakeCentral) fetch(_ context.Context, url string) ([]byte, error) {
	f.hits++
	f.urls = append(f.urls, url)
	if b, ok := f.files[url]; ok {
		return b, nil
	}
	for gav, b := range f.jars {
		var g input.GAV
		p := strings.Split(gav, ":")
		g.Group, g.Artifact, g.Version = p[0], p[1], p[2]
		if url == ReferenceURL("https://central.test", g) {
			return b, nil
		}
	}
	return nil, errors.New("HTTP 404")
}

func classBytes(t *testing.T) (codec, hasher []byte) {
	return readFile(t, filepath.Join(autoClasses, codecEntry)), readFile(t, filepath.Join(autoClasses, hasherEntry))
}

// padUtf8 appends an unused Utf8 constant to a class file's pool (other
// bytes, same content), as a shading tool's re-serialization would.
func padUtf8(t *testing.T, b []byte) []byte {
	t.Helper()
	count := int(binary.BigEndian.Uint16(b[8:10]))
	pos := 10
	for i := 1; i < count; i++ {
		tag := b[pos]
		pos++
		switch tag {
		case 1:
			pos += 2 + int(binary.BigEndian.Uint16(b[pos:]))
		case 3, 4, 9, 10, 11, 12, 17, 18:
			pos += 4
		case 5, 6:
			pos += 8
			i++
		case 7, 8, 16, 19, 20:
			pos += 2
		case 15:
			pos += 3
		default:
			t.Fatalf("tag %d", tag)
		}
	}
	var out bytes.Buffer
	out.Write(b[:8])
	binary.Write(&out, binary.BigEndian, uint16(count+1))
	out.Write(b[10:pos])
	out.Write([]byte{1, 0, 3, 'p', 'a', 'd'})
	out.Write(b[pos:])
	return out.Bytes()
}

func embeddedByGAV(inv *Inventory, gav string) Embedded {
	for _, e := range inv.Embedded {
		if e.GAV.String() == gav {
			return e
		}
	}
	return Embedded{}
}

func instanceByID(inv *Inventory, id string) *Instance {
	for i := range inv.Instances {
		if inv.Instances[i].ID == id {
			return &inv.Instances[i]
		}
	}
	return nil
}

// TestEmbeddedConfirmed: a library embedded with its packages becomes its
// own instance — confirmed by classes whose bytes differ but whose
// content is the reference's — owns those classes, and is located
// inside its host.
func TestEmbeddedConfirmed(t *testing.T) {
	codec, hasher := classBytes(t)
	deps := t.TempDir()
	hostJar(t, filepath.Join(deps, "fat-1.0.jar"), map[string][]byte{codecEntry: padUtf8(t, codec), hasherEntry: hasher}, "org.other:embedded:2.3")
	inv, h := buildWithH(t, deps)
	c := &fakeCentral{jars: map[string][]byte{"org.other:embedded:2.3": jarBytes(t, map[string][]byte{codecEntry: codec, hasherEntry: hasher})}}
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, "https://central.test", h, nil); err != nil {
		t.Fatal(err)
	}
	e := embeddedByGAV(inv, "org.other:embedded:2.3")
	if e.Kind != EmbeddedLibrary || e.Status != EmbeddedConfirmed || e.Present != 2 || e.Identical != 2 || e.Instance == "" {
		t.Fatalf("candidate = %+v", e)
	}
	in := instanceByID(inv, e.Instance)
	host := instanceByID(inv, e.HostInstance)
	if in.Detection != DetectedEmbedded || in.Classes != 2 || in.HostInstance != host.ID || host.Classes != 0 {
		t.Errorf("embedded %+v host %+v", in, host)
	}
	if loc := inv.InstanceLocation(in.ID); loc != "deps:fat-1.0.jar!/META-INF/maven/org.other/embedded" {
		t.Errorf("location = %q", loc)
	}
	for _, ci := range h.All() {
		if strings.HasPrefix(ci.FQCN, "fixture.jpms.auto.") && inv.InstanceForClass(ci) != in.ID {
			t.Errorf("%s owned by %s", ci.FQCN, inv.InstanceForClass(ci))
		}
	}
	if gs := inv.ConfirmedEmbeddedGAVs(); len(gs) != 1 || gs[0].String() != "org.other:embedded:2.3" {
		t.Errorf("lookup GAVs = %v", gs)
	}
	if conf, unconf, mods := inv.EmbeddedCounts(); conf != 1 || unconf != 0 || mods != 0 {
		t.Errorf("counts = %d %d %d", conf, unconf, mods)
	}
}

// TestEmbeddedNotConfirmed: stale metadata without classes, other content
// under the same names, a minimized copy, a missing reference and a
// module of the host's own project never become an instance.
func TestEmbeddedNotConfirmed(t *testing.T) {
	codec, hasher := classBytes(t)
	deps := t.TempDir()
	hostJar(t, filepath.Join(deps, "fat-1.0.jar"), map[string][]byte{codecEntry: codec},
		"org.stale:gone:1.0", "org.swap:lib:1.0", "org.min:lib:1.0", "org.nowhere:lib:1.0", "org.host.modules:core:1.0")
	inv, h := buildWithH(t, deps)
	c := &fakeCentral{jars: map[string][]byte{
		"org.stale:gone:1.0": jarBytes(t, map[string][]byte{"other/Gone.class": hasher}),
		"org.swap:lib:1.0":   jarBytes(t, map[string][]byte{codecEntry: hasher}),
		"org.min:lib:1.0":    jarBytes(t, map[string][]byte{codecEntry: codec, hasherEntry: hasher, "other/X.class": hasher}),
	}}
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, "https://central.test", h, nil); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"org.stale:gone:1.0":  EmbeddedClassesAbsent,
		"org.swap:lib:1.0":    EmbeddedContentDiffers,
		"org.min:lib:1.0":     EmbeddedPartial,
		"org.nowhere:lib:1.0": EmbeddedReferenceUnavailable,
	}
	for gav, st := range want {
		if e := embeddedByGAV(inv, gav); e.Kind != EmbeddedLibrary || e.Status != st || e.Instance != "" {
			t.Errorf("%s = %+v, want %s", gav, e, st)
		}
	}
	if e := embeddedByGAV(inv, "org.host.modules:core:1.0"); e.Kind != EmbeddedModule || e.Status != "" {
		t.Errorf("module = %+v", e)
	}
	// 4 reference jars + the version list of the content_differs one
	// (CVE:F6b; unavailable here, so its status stays content_differs).
	if c.hits != 5 {
		t.Errorf("Central asked %d times, want 5 (never for a module)", c.hits)
	}
	if len(inv.Instances) != 2 || len(inv.ConfirmedEmbeddedGAVs()) != 0 {
		t.Errorf("instances = %d", len(inv.Instances))
	}
	if conf, unconf, mods := inv.EmbeddedCounts(); conf != 0 || unconf != 4 || mods != 1 {
		t.Errorf("counts = %d %d %d", conf, unconf, mods)
	}
}

// TestEmbeddedTwoVersions: the same library embedded in two jars in two
// versions gives two instances, never one; the class conflict between
// the copies is attributed to the embedded instances.
func TestEmbeddedTwoVersions(t *testing.T) {
	codec, _ := classBytes(t)
	deps := t.TempDir()
	hostJar(t, filepath.Join(deps, "a-fat.jar"), map[string][]byte{codecEntry: codec}, "org.other:embedded:2.3")
	writeJar(t, filepath.Join(deps, "b-fat.jar"), []string{"META-INF/maven/org.b/fat/pom.properties", codecEntry, "META-INF/maven/org.other/embedded/pom.properties"},
		map[string][]byte{"META-INF/maven/org.b/fat/pom.properties": pom("org.b", "fat", "7.0"), codecEntry: codec,
			"META-INF/maven/org.other/embedded/pom.properties": pom("org.other", "embedded", "2.4")})
	inv, h := buildWithH(t, deps)
	ref := jarBytes(t, map[string][]byte{codecEntry: codec})
	c := &fakeCentral{jars: map[string][]byte{"org.other:embedded:2.3": ref, "org.other:embedded:2.4": ref}}
	if err := inv.ResolveEmbedded(context.Background(), c.fetch, "https://central.test", h, nil); err != nil {
		t.Fatal(err)
	}
	a, b := embeddedByGAV(inv, "org.other:embedded:2.3"), embeddedByGAV(inv, "org.other:embedded:2.4")
	if a.Instance == "" || b.Instance == "" || a.Instance == b.Instance {
		t.Fatalf("instances %q %q", a.Instance, b.Instance)
	}
	if len(inv.Conflicts) != 1 || inv.Conflicts[0].KeptInstance != a.Instance || inv.Conflicts[0].ShadowedInstance != b.Instance {
		t.Errorf("conflicts = %+v, want kept %s shadowed %s", inv.Conflicts, a.Instance, b.Instance)
	}
	if in := instanceByID(inv, b.Instance); in.Shadowed != 1 || in.Classes != 0 {
		t.Errorf("shadowed copy = %+v", in)
	}
}

// buildWithH builds the inventory of the serviceloader fixture app with
// the jars of deps, and returns the hierarchy too.
func buildWithH(t *testing.T, deps string) (*Inventory, *hierarchy.Hierarchy) {
	t.Helper()
	app := "../../testdata/serviceloader/classes"
	prepared, err := input.PrepareAppInput(app)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { prepared.Cleanup() })
	djs, err := input.ResolveDepJars(deps, nil)
	if err != nil {
		t.Fatal(err)
	}
	h, _, err := hierarchy.Build(prepared.AppClassesJar, djs)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := Build(Inputs{AppInput: app, AppJar: prepared.AppClassesJar, DepsDir: deps, DepJars: djs}, h)
	if err != nil {
		t.Fatal(err)
	}
	return inv, h
}
