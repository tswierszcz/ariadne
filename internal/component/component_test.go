package component

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const autoClasses = "../../testdata/jpms/classes/auto"

func readFile(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// writeJar writes entries (name -> content) to path, in the given order.
func writeJar(t *testing.T, path string, names []string, entries map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for _, n := range names {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		w.Write(entries[n])
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
}

func pom(g, a, v string) []byte {
	return []byte("groupId=" + g + "\nartifactId=" + a + "\nversion=" + v + "\n")
}

func libJar(t *testing.T, path string, extraPom bool) {
	t.Helper()
	names := []string{"fixture/jpms/auto/AutoCodec.class", "META-INF/maven/org.example/lib/pom.properties"}
	entries := map[string][]byte{
		names[0]: readFile(t, filepath.Join(autoClasses, "fixture/jpms/auto/AutoCodec.class")),
		names[1]: pom("org.example", "lib", "1.0"),
	}
	if extraPom {
		n := "META-INF/maven/org.other/embedded/pom.properties"
		names = append(names, n)
		entries[n] = pom("org.other", "embedded", "2.3")
	}
	writeJar(t, path, names, entries)
}

func build(t *testing.T, appInput, depsDir string) *Inventory {
	t.Helper()
	prepared, err := input.PrepareAppInput(appInput)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { prepared.Cleanup() })
	dir := depsDir
	if prepared.ExtractedLibs != "" {
		dir = prepared.ExtractedLibs
	}
	var deps []input.DepJar
	if dir != "" {
		deps, err = input.ResolveDepJars(dir, nil)
		if err != nil {
			t.Fatal(err)
		}
	}
	h, _, err := hierarchy.Build(prepared.AppClassesJar, deps)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := Build(Inputs{AppInput: appInput, AppJar: prepared.AppClassesJar, DepsDir: dir,
		Extracted: prepared.ExtractedLibs != "", LibEntries: prepared.LibEntries, DepJars: deps}, h)
	if err != nil {
		t.Fatal(err)
	}
	return inv
}

func ids(inv *Inventory) []string {
	var out []string
	for _, in := range inv.Instances {
		out = append(out, in.ID+"="+inv.ArtifactLocation(in.ArtifactID))
	}
	return out
}

// TestIDsStableAcrossTempPaths: a classes directory and a Spring Boot fat
// jar are repacked/extracted into fresh temp directories on every run;
// identifiers and locations must not change.
func TestIDsStableAcrossTempPaths(t *testing.T) {
	// Directory input.
	a := strings.Join(ids(build(t, "../../testdata/serviceloader/classes", "")), ",")
	b := strings.Join(ids(build(t, "../../testdata/serviceloader/classes", "")), ",")
	if a != b || !strings.Contains(a, "=app:classes") {
		t.Errorf("directory input ids differ or bad location:\n%s\n%s", a, b)
	}

	// Fat jar input, built twice in two different directories.
	mk := func() string {
		d := t.TempDir()
		lib := filepath.Join(d, "lib-1.0.jar")
		libJar(t, lib, false)
		fat := filepath.Join(d, "demo-app.jar")
		writeJar(t, fat, []string{"BOOT-INF/classes/fixture/service/Decoder.class", "BOOT-INF/lib/lib-1.0.jar"}, map[string][]byte{
			"BOOT-INF/classes/fixture/service/Decoder.class": readFile(t, "../../testdata/serviceloader/classes/fixture/service/Decoder.class"),
			"BOOT-INF/lib/lib-1.0.jar":                       readFile(t, lib),
		})
		return fat
	}
	x := strings.Join(ids(build(t, mk(), "")), ",")
	y := strings.Join(ids(build(t, mk(), "")), ",")
	if x != y {
		t.Errorf("fat jar ids differ across runs:\n%s\n%s", x, y)
	}
	if !strings.Contains(x, "=app:demo-app.jar!/BOOT-INF/lib/lib-1.0.jar") {
		t.Errorf("fat jar lib location = %s, want app:demo-app.jar!/BOOT-INF/lib/lib-1.0.jar", x)
	}
}

// TestTwoInstancesOfOneComponent: the same component in two jars gets two
// instances with their own evidence (and a class conflict), not one.
func TestTwoInstancesOfOneComponent(t *testing.T) {
	deps := t.TempDir()
	libJar(t, filepath.Join(deps, "lib-1.0.jar"), false)
	libJar(t, filepath.Join(deps, "lib-copy.jar"), true)
	inv := build(t, "../../testdata/serviceloader/classes", deps)

	var libs []Instance
	for _, in := range inv.Instances {
		if in.GAV != nil && in.GAV.Artifact == "lib" {
			libs = append(libs, in)
		}
	}
	if len(libs) != 2 || libs[0].ID == libs[1].ID || libs[0].ArtifactID == libs[1].ArtifactID {
		t.Fatalf("instances = %+v, want two distinct instances of org.example:lib:1.0", libs)
	}
	for _, in := range libs {
		if in.Detection != DetectedPOMProperties || len(in.Evidence) == 0 || in.Evidence[0].Kind != "pom.properties" {
			t.Errorf("%s: detection/evidence = %s %+v", in.ID, in.Detection, in.Evidence)
		}
	}
	// lib-copy carries a second pom.properties: an embedded candidate, not
	// an extra instance.
	var copyInst *Instance
	for i := range libs {
		if strings.HasSuffix(inv.ArtifactLocation(libs[i].ArtifactID), "lib-copy.jar") {
			copyInst = &libs[i]
		}
	}
	if copyInst == nil || len(copyInst.Evidence) != 2 || copyInst.Evidence[1].Kind != "embedded-candidate" || !strings.Contains(copyInst.Evidence[1].Detail, "org.other:embedded:2.3") {
		t.Errorf("lib-copy evidence = %+v, want pom.properties + embedded candidate", copyInst)
	}
	if len(inv.Instances) != 3 {
		t.Errorf("want app + 2 instances, got %d", len(inv.Instances))
	}

	// The shared class is a conflict: first copy (by file name) wins.
	if len(inv.Conflicts) != 1 || inv.Conflicts[0].FQCN != "fixture.jpms.auto.AutoCodec" || !inv.Conflicts[0].SameContent {
		t.Fatalf("conflicts = %+v", inv.Conflicts)
	}
	kept, shadow := inv.Conflicts[0].KeptInstance, inv.Conflicts[0].ShadowedInstance
	for _, in := range libs {
		loc := inv.ArtifactLocation(in.ArtifactID)
		switch {
		case strings.HasSuffix(loc, "lib-1.0.jar") && (in.ID != kept || in.Classes != 1 || in.Shadowed != 0):
			t.Errorf("lib-1.0 should be selected: %+v (kept=%s)", in, kept)
		case strings.HasSuffix(loc, "lib-copy.jar") && (in.ID != shadow || in.Classes != 0 || in.Shadowed != 1):
			t.Errorf("lib-copy should be shadowed: %+v (shadowed=%s)", in, shadow)
		}
	}
}

func TestUnidentifiedJar(t *testing.T) {
	deps := t.TempDir()
	writeJar(t, filepath.Join(deps, "mystery.jar"), []string{"fixture/jpms/auto/AutoCodec.class"}, map[string][]byte{
		"fixture/jpms/auto/AutoCodec.class": readFile(t, filepath.Join(autoClasses, "fixture/jpms/auto/AutoCodec.class")),
	})
	inv := build(t, "../../testdata/serviceloader/classes", deps)
	_, identified, unidentified, _ := inv.Counts()
	if identified != 0 || unidentified != 1 {
		t.Errorf("counts identified=%d unidentified=%d", identified, unidentified)
	}
	for _, in := range inv.Instances {
		if in.Detection == DetectedUnidentified && (in.GAV != nil || len(in.Evidence) != 0 || in.Classes != 1) {
			t.Errorf("unidentified instance = %+v", in)
		}
	}
}
