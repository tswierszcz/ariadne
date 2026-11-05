package serviceload

import (
	"archive/zip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const jpmsClasses = "../../testdata/jpms/classes"

// zipFixture packs dir into dest/name. extra adds or replaces entries
// (e.g. a manifest); skip drops entries by zip path.
func zipFixture(t *testing.T, dir, dest, name string, extra map[string]string, skip ...string) string {
	t.Helper()
	path := filepath.Join(dest, name)
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(out)
	skipped := map[string]bool{}
	for _, s := range skip {
		skipped[s] = true
	}
	for k := range extra {
		skipped[k] = true
	}
	err = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		rel = filepath.ToSlash(rel)
		if skipped[rel] {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range extra {
		w, err := zw.Create(k)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(v))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	out.Close()
	return path
}

type jpmsJars struct {
	app, api, impl, auto, plain, mr, mrNotMR string
}

func buildJPMSJars(t *testing.T) jpmsJars {
	t.Helper()
	d := t.TempDir()
	return jpmsJars{
		app:  zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.app"), d, "app.jar", nil),
		api:  zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.api"), d, "fixture-api-1.0.jar", nil),
		impl: zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.impl"), d, "fixture-impl-1.0.jar", nil),
		auto: zipFixture(t, filepath.Join(jpmsClasses, "auto"), d, "fixture-auto-1.2.3.jar",
			map[string]string{"META-INF/MANIFEST.MF": "Manifest-Version: 1.0\r\nAutomatic-Module-Name: fixture.jpms.automati\r\n c\r\n\r\n"}),
		plain:   zipFixture(t, filepath.Join(jpmsClasses, "auto"), d, "fixture-plain_lib-2.0.1.jar", nil, "META-INF/services/fixture.jpms.api.Codec"),
		mr:      zipFixture(t, filepath.Join(jpmsClasses, "mr"), d, "fixture-mr-3.0.jar", nil),
		mrNotMR: zipFixture(t, filepath.Join(jpmsClasses, "mr"), d, "fixture-mr-nomr-3.0.jar", nil, "META-INF/MANIFEST.MF"),
	}
}

func (j jpmsJars) deps() []input.DepJar {
	return []input.DepJar{{Path: j.api}, {Path: j.impl}, {Path: j.auto}, {Path: j.plain}, {Path: j.mr}, {Path: j.mrNotMR}}
}

func descriptorFor(t *testing.T, inv ModuleInventory, jar string) ModuleDescriptor {
	t.Helper()
	for _, d := range inv.Descriptors {
		if d.Jar == jar {
			return d
		}
	}
	t.Fatalf("no descriptor for %s in %+v", jar, inv.Descriptors)
	return ModuleDescriptor{}
}

func inventoryJPMS(t *testing.T, j jpmsJars, targetJava int) (ModuleInventory, []ServiceRegistration, *hierarchy.Hierarchy) {
	t.Helper()
	h, _, err := hierarchy.BuildWithTarget(j.app, j.deps(), targetJava)
	if err != nil {
		t.Fatalf("hierarchy: %v", err)
	}
	regs, _, err := Inventory(h, j.app, j.deps())
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	inv, warnings, err := InventoryModules(h, j.app, j.deps(), targetJava, regs)
	if err != nil {
		t.Fatalf("InventoryModules: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings: %v", warnings)
	}
	return inv, regs, h
}

// TestInventoryModules_compiledFixture checks SERVICE_LOADING S3's first
// slice on real javac output: explicit vs automatic modules, provides
// declarations, the provider() factory flag and both mismatch directions.
func TestInventoryModules_compiledFixture(t *testing.T) {
	j := buildJPMSJars(t)
	inv, _, _ := inventoryJPMS(t, j, 0)
	if inv.Scenario != ScenarioClasspath {
		t.Fatalf("Scenario = %q", inv.Scenario)
	}

	for jar, want := range map[string]string{j.app: "fixture.jpms.app", j.api: "fixture.jpms.api", j.impl: "fixture.jpms.impl"} {
		d := descriptorFor(t, inv, jar)
		if d.Kind != ModuleExplicit || d.Name != want || d.NameSource != "module-info.class" || d.Info == nil {
			t.Errorf("%s: %+v, want explicit %s from module-info.class", filepath.Base(jar), d, want)
		}
	}
	if app := descriptorFor(t, inv, j.app); len(app.Info.Uses) != 1 || app.Info.Uses[0] != "fixture.jpms.api.Codec" {
		t.Errorf("app uses = %v", app.Info.Uses)
	}

	auto := descriptorFor(t, inv, j.auto)
	if auto.Kind != ModuleAutomatic || auto.Name != "fixture.jpms.automatic" || auto.NameSource != "Automatic-Module-Name" {
		t.Errorf("auto: %+v, want Automatic-Module-Name (continuation line joined)", auto)
	}
	plain := descriptorFor(t, inv, j.plain)
	if plain.Kind != ModuleAutomatic || plain.Name != "fixture.plain.lib" || plain.NameSource != "file name" {
		t.Errorf("plain: %+v, want file-name-derived fixture.plain.lib", plain)
	}

	mr := descriptorFor(t, inv, j.mr)
	if mr.Kind != ModuleExplicit || mr.Name != "fixture.jpms.mr" || mr.NameSource != "META-INF/versions/9/module-info.class" || !strings.Contains(mr.DescriptorNote, "-target-java unset") {
		t.Errorf("mr (target unset): %+v", mr)
	}
	if nomr := descriptorFor(t, inv, j.mrNotMR); nomr.Kind != ModuleAutomatic || nomr.Name != "fixture.mr.nomr" {
		t.Errorf("versioned descriptor without Multi-Release manifest must be inert: %+v", nomr)
	}

	byProvider := map[string]ModuleProviderDecl{}
	for _, p := range inv.Provides {
		byProvider[p.Provider] = p
	}
	if len(inv.Provides) != 2 {
		t.Fatalf("Provides = %+v, want ModuleCodec and FactoryCodec", inv.Provides)
	}
	if p := byProvider["fixture.jpms.impl.ModuleCodec"]; !p.ProviderFound || p.HasProviderMethod || !p.InServicesFile || p.Module != "fixture.jpms.impl" {
		t.Errorf("ModuleCodec decl = %+v", p)
	}
	if p := byProvider["fixture.jpms.impl.FactoryCodec"]; !p.ProviderFound || !p.HasProviderMethod || p.InServicesFile {
		t.Errorf("FactoryCodec decl = %+v, want provider() and not in META-INF/services", p)
	}

	var got []string
	for _, m := range inv.Mismatches {
		got = append(got, m.Provider+"|"+m.Side)
	}
	sort.Strings(got)
	want := []string{"fixture.jpms.impl.FactoryCodec|module-info only", "fixture.jpms.impl.LegacyCodec|META-INF/services only"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Mismatches = %v, want %v", got, want)
	}

	explicit, automatic, provides, mismatches := inv.Counts()
	if explicit != 4 || automatic != 3 || provides != 2 || mismatches != 2 {
		t.Errorf("Counts = %d/%d/%d/%d, want 4/3/2/2", explicit, automatic, provides, mismatches)
	}
}

func TestInventoryModules_multiReleaseTarget(t *testing.T) {
	j := buildJPMSJars(t)

	inv8, _, _ := inventoryJPMS(t, j, 8)
	if d := descriptorFor(t, inv8, j.mr); d.Kind != ModuleAutomatic || !strings.Contains(d.DescriptorNote, "above -target-java 8") {
		t.Errorf("target 8: %+v, want automatic with a note", d)
	}
	inv11, _, _ := inventoryJPMS(t, j, 11)
	if d := descriptorFor(t, inv11, j.mr); d.Kind != ModuleExplicit || d.NameSource != "META-INF/versions/9/module-info.class" || d.DescriptorNote != "" {
		t.Errorf("target 11: %+v, want explicit from versions/9 without a note", d)
	}
}

// TestInventoryModules_classpathSemanticsUnchanged: a provider declared
// only through module-info `provides` must not become a constructor
// candidate in the class-path scenario, while META-INF/services-only
// providers of the same explicit module still do.
func TestInventoryModules_classpathSemanticsUnchanged(t *testing.T) {
	j := buildJPMSJars(t)
	_, regs, h := inventoryJPMS(t, j, 0)
	loads, _, err := FindLoadSites(h)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, m := range FindMaterializations(h, loads, regs) {
		got = append(got, m.Registration.ProviderClass)
	}
	sort.Strings(got)
	want := []string{"fixture.jpms.auto.AutoCodec", "fixture.jpms.impl.LegacyCodec", "fixture.jpms.impl.ModuleCodec"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("materializations = %v, want %v (FactoryCodec is module-path only)", got, want)
	}
}

func TestInventoryModules_malformedDescriptorIsWarning(t *testing.T) {
	d := t.TempDir()
	bad := zipFixture(t, filepath.Join(jpmsClasses, "auto"), d, "bad-1.0.jar", map[string]string{"module-info.class": "not a class"})
	app := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.api"), d, "app.jar", nil)
	h, _, err := hierarchy.Build(app, []input.DepJar{{Path: bad}})
	if err != nil {
		t.Fatal(err)
	}
	inv, warnings, err := InventoryModules(h, app, []input.DepJar{{Path: bad}}, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "bad-1.0.jar") {
		t.Errorf("warnings = %v, want one for bad-1.0.jar", warnings)
	}
	for _, desc := range inv.Descriptors {
		if desc.Jar == bad {
			t.Errorf("malformed descriptor must not be reported as a module: %+v", desc)
		}
	}
}

func TestAutomaticModuleName(t *testing.T) {
	for _, tc := range []struct{ file, name string }{
		{"jackson-core-2.11.4.jar", "jackson.core"},
		{"spring-boot-starter-web-2.4.5.jar", "spring.boot.starter.web"},
		{"foo_bar--baz.jar", "foo.bar.baz"},
		{"commons-lang3-3.12.0.jar", "commons.lang3"},
		{"guava-30.1-jre.jar", "guava"},
		{"1-2.jar", ""},
		{"lib.9x-1.0.jar", ""},
	} {
		got, reason := automaticModuleName(tc.file)
		if got != tc.name {
			t.Errorf("%s: got %q (%s), want %q", tc.file, got, reason, tc.name)
		}
		if tc.name == "" && reason == "" {
			t.Errorf("%s: invalid name without a reason", tc.file)
		}
	}
}
