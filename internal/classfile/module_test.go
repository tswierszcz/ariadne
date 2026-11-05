package classfile

import (
	"os"
	"reflect"
	"testing"
)

func parseFixtureFile(t *testing.T, path string) *ClassFile {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cf, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse(%s): %v", path, err)
	}
	return cf
}

// TestParse_moduleInfoFromCompiledFixture decodes real javac output
// (testdata/jpms, see its README) — SERVICE_LOADING S3's Module attribute.
func TestParse_moduleInfoFromCompiledFixture(t *testing.T) {
	cf := parseFixtureFile(t, "../../testdata/jpms/classes/mods/fixture.jpms.impl/module-info.class")
	if !cf.IsModule() || cf.Module == nil {
		t.Fatalf("IsModule=%v Module=%v, want a module descriptor", cf.IsModule(), cf.Module)
	}
	m := cf.Module
	if m.Name != "fixture.jpms.impl" || m.IsOpen() {
		t.Errorf("Name=%q open=%v", m.Name, m.IsOpen())
	}
	var api *ModuleRequires
	for i := range m.Requires {
		if m.Requires[i].Module == "fixture.jpms.api" {
			api = &m.Requires[i]
		}
	}
	if api == nil || api.Flags&ModuleAccTransitive == 0 {
		t.Errorf("requires = %+v, want transitive fixture.jpms.api", m.Requires)
	}
	want := []ModuleProvides{{Service: "fixture.jpms.api.Codec", With: []string{"fixture.jpms.impl.ModuleCodec", "fixture.jpms.impl.FactoryCodec"}}}
	if !reflect.DeepEqual(m.Provides, want) {
		t.Errorf("Provides = %+v, want %+v", m.Provides, want)
	}

	api2 := parseFixtureFile(t, "../../testdata/jpms/classes/mods/fixture.jpms.api/module-info.class").Module
	if len(api2.Exports) != 1 || api2.Exports[0].Package != "fixture.jpms.api" || len(api2.Exports[0].To) != 0 {
		t.Errorf("Exports = %+v, want unqualified fixture.jpms.api", api2.Exports)
	}

	app := parseFixtureFile(t, "../../testdata/jpms/classes/mods/fixture.jpms.app/module-info.class").Module
	if !reflect.DeepEqual(app.Uses, []string{"fixture.jpms.api.Codec"}) {
		t.Errorf("Uses = %v", app.Uses)
	}
}

func TestParse_ordinaryClassHasNoModule(t *testing.T) {
	cf := parseFixtureFile(t, "../../testdata/jpms/classes/mods/fixture.jpms.api/fixture/jpms/api/Codec.class")
	if cf.IsModule() || cf.Module != nil {
		t.Errorf("ordinary interface reported as module: %+v", cf.Module)
	}
}

func TestParseModuleAttribute_truncatedIsError(t *testing.T) {
	p := pool{{}, {tag: tagUtf8, str: "m"}, {tag: tagModule, nameIndex: 1}}
	// name, flags, version, then requires_count=1 with no entry bytes.
	if _, err := parseModuleAttribute([]byte{0, 2, 0, 0, 0, 0, 0, 1}, p); err == nil {
		t.Fatal("want an error for a truncated Module attribute")
	}
}
