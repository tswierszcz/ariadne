package serviceload

import (
	"archive/zip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/analysis"
	"ariadne/internal/callback"
	"ariadne/internal/callgraph"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/models"
	"ariadne/internal/reflectinvoke"
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
	if app := descriptorFor(t, inv, j.app); strings.Join(app.Info.Uses, ",") != "fixture.jpms.api.Codec,fixture.jpms.api.Hasher" {
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
	want := []string{"fixture.jpms.auto.AutoCodec", "fixture.jpms.auto.AutoHasher", "fixture.jpms.impl.LegacyCodec", "fixture.jpms.impl.ModuleCodec"}
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

// runJPMSScenario runs the real pipeline (hierarchy → S0 inventory →
// module inventory → scenario view → load/consumption sites → B3
// coordinator) from rootFQCN.main for one scenario.
func runJPMSScenario(t *testing.T, j jpmsJars, nousesJar string, scenario ModuleScenario, rootFQCN string) (*callgraph.Graph, uint32, analysis.Result, []string) {
	t.Helper()
	return runJPMS(t, j, nousesJar, ScenarioConfig{Scenario: scenario, AppJar: j.app}, rootFQCN, "main", "([Ljava/lang/String;)V")
}

func runJPMS(t *testing.T, j jpmsJars, nousesJar string, cfg ScenarioConfig, rootFQCN, rootName, rootDesc string) (*callgraph.Graph, uint32, analysis.Result, []string) {
	t.Helper()
	scenario := cfg.Scenario
	deps := append(j.deps(), input.DepJar{Path: nousesJar})
	h, _, err := hierarchy.Build(j.app, deps)
	if err != nil {
		t.Fatal(err)
	}
	regs, _, err := Inventory(h, j.app, deps)
	if err != nil {
		t.Fatal(err)
	}
	inv, warnings, err := InventoryModules(h, j.app, deps, 0, regs)
	if err != nil {
		t.Fatal(err)
	}
	view, scenarioWarnings := ApplyScenario(cfg, regs, &inv, h)
	warnings = append(warnings, scenarioWarnings...)
	if inv.Scenario != scenario {
		t.Fatalf("inventory scenario = %q, want %q", inv.Scenario, scenario)
	}
	loads, _, err := FindLoadSites(h)
	if err != nil {
		t.Fatal(err)
	}
	consumptions, _, err := FindConsumptionSites(h, loads)
	if err != nil {
		t.Fatal(err)
	}
	mats := FindMaterializations(h, loads, view.Registrations)
	eval := NewModelEvaluator(h, loads, mats, consumptions)
	eval.SetScenario(view)
	g := callgraph.Build(h, nil)
	root, ok := g.ID(rootFQCN, rootName, rootDesc)
	if !ok {
		t.Fatalf("no %s#%s%s in graph", rootFQCN, rootName, rootDesc)
	}
	registry := models.NewRegistry()
	RegisterModel(registry)
	RegisterDispatchModels(registry, consumptions)
	res, err := analysis.Run(h, g, registry, eval, []uint32{root}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatal(err)
	}
	return g, root, res, warnings
}

func hasModelEdgeFrom(g *callgraph.Graph, from uint32, owner, name, desc string) bool {
	to, ok := g.ID(owner, name, desc)
	if !ok {
		return false
	}
	for _, e := range g.Edges(from) {
		if e.To == to && e.Basis == callgraph.BasisModel {
			return true
		}
	}
	return false
}

// TestScenario_providerSetsDiffer is SERVICE_LOADING S3 slice 2's core
// case: the same bytecode materializes different providers on the class
// path and on the module path.
func TestScenario_providerSetsDiffer(t *testing.T) {
	j := buildJPMSJars(t)
	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)
	const app = "fixture.jpms.app.Main"
	const encode = "(Ljava/lang/String;)Ljava/lang/String;"
	const providerDesc = "()Lfixture/jpms/api/Codec;"

	g, root, _, warnings := runJPMSScenario(t, j, nouses, ScenarioClasspath, app)
	if len(warnings) != 0 {
		t.Errorf("classpath warnings: %v", warnings)
	}
	for _, c := range []string{"fixture.jpms.impl.ModuleCodec", "fixture.jpms.impl.LegacyCodec", "fixture.jpms.auto.AutoCodec"} {
		if !hasModelEdgeFrom(g, root, c, "<init>", "()V") {
			t.Errorf("classpath: missing constructor edge to %s", c)
		}
	}
	if hasModelEdgeFrom(g, root, "fixture.jpms.impl.FactoryCodec", "provider", providerDesc) {
		t.Error("classpath: provides-only FactoryCodec must not be materialized")
	}

	g, root, res, warnings := runJPMSScenario(t, j, nouses, ScenarioModulePath, app)
	if len(warnings) != 0 {
		t.Errorf("module-path warnings: %v", warnings)
	}
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
		t.Error("module-path: missing ModuleCodec constructor (declared in provides)")
	}
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.FactoryCodec", "provider", providerDesc) {
		t.Error("module-path: missing FactoryCodec.provider() edge")
	}
	if hasModelEdgeFrom(g, root, "fixture.jpms.impl.FactoryCodec", "<init>", "()V") {
		t.Error("module-path: FactoryCodec must be created by provider(), not its private constructor")
	}
	if !hasModelEdgeFrom(g, root, "fixture.jpms.auto.AutoCodec", "<init>", "()V") {
		t.Error("module-path: missing automatic-module AutoCodec constructor")
	}
	if hasModelEdgeFrom(g, root, "fixture.jpms.impl.LegacyCodec", "<init>", "()V") {
		t.Error("module-path: LegacyCodec (explicit module's META-INF/services only) must be ignored")
	}
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "encode", encode) {
		t.Error("module-path: missing exact dispatch to ModuleCodec.encode")
	}
	if hasModelEdgeFrom(g, root, "fixture.jpms.impl.FactoryCodec", "encode", encode) || hasModelEdgeFrom(g, root, "fixture.jpms.impl.LegacyCodec", "encode", encode) {
		t.Error("module-path: no exact model dispatch to a provider()-created or ignored provider")
	}
	if !res.Incomplete || !strings.Contains(strings.Join(res.Reasons, "\n"), "provider() method") {
		t.Errorf("module-path: want an explicit provider() dispatch gap, got incomplete=%v reasons=%v", res.Incomplete, res.Reasons)
	}
}

func TestScenario_missingUsesBlocksProviders(t *testing.T) {
	j := buildJPMSJars(t)
	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)
	const caller = "fixture.jpms.nouses.NoUses"

	g, root, res, _ := runJPMSScenario(t, j, nouses, ScenarioModulePath, caller)
	for _, c := range []string{"fixture.jpms.impl.ModuleCodec", "fixture.jpms.auto.AutoCodec"} {
		if hasModelEdgeFrom(g, root, c, "<init>", "()V") {
			t.Errorf("module-path without `uses`: unexpected edge to %s", c)
		}
	}
	if !res.Incomplete || !strings.Contains(strings.Join(res.Reasons, "\n"), "does not declare `uses fixture.jpms.api.Codec`") {
		t.Errorf("want an explicit missing-uses reason, got %v", res.Reasons)
	}

	// The class path has no `uses` rule: the same caller loads providers.
	g, root, _, _ = runJPMSScenario(t, j, nouses, ScenarioClasspath, caller)
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
		t.Error("classpath: NoUses should still materialize ModuleCodec")
	}
}

func TestApplyScenario_missingRequiresWarns(t *testing.T) {
	j := buildJPMSJars(t)
	deps := []input.DepJar{{Path: j.impl}} // fixture.jpms.api deliberately absent
	h, _, err := hierarchy.Build(j.app, deps)
	if err != nil {
		t.Fatal(err)
	}
	regs, _, _ := Inventory(h, j.app, deps)
	inv, _, err := InventoryModules(h, j.app, deps, 0, regs)
	if err != nil {
		t.Fatal(err)
	}
	_, warnings := ApplyScenario(ScenarioConfig{Scenario: ScenarioModulePath}, regs, &inv, h)
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "requires fixture.jpms.api") {
		t.Errorf("warnings = %v, want a missing-requires warning", warnings)
	}
	if _, cpWarnings := ApplyScenario(ScenarioConfig{Scenario: ScenarioClasspath}, regs, &inv, h); len(cpWarnings) != 0 {
		t.Errorf("classpath scenario must not check requires: %v", cpWarnings)
	}
}

func TestParseScenario(t *testing.T) {
	for _, ok := range []string{"classpath", "module-path"} {
		if _, err := ParseScenario(ok); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	if _, err := ParseScenario("modulepath"); err == nil {
		t.Error("want an error for an unknown scenario")
	}
}

const (
	layersFQCN   = "fixture.jpms.app.Layers"
	codecProvide = "()Lfixture/jpms/api/Codec;"
)

// TestScenario_bootLayer: load(ModuleLayer.boot(), S) sees only named
// modules on the module path — nothing from the input on the class path,
// automatic and explicit modules on the module path, and only the
// module-path part of a mixed launch.
func TestScenario_bootLayer(t *testing.T) {
	j := buildJPMSJars(t)
	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)

	g, root, res, _ := runJPMS(t, j, nouses, ScenarioConfig{Scenario: ScenarioClasspath, AppJar: j.app}, layersFQCN, "boot", "()V")
	for _, c := range []string{"fixture.jpms.impl.ModuleCodec", "fixture.jpms.impl.LegacyCodec", "fixture.jpms.auto.AutoCodec"} {
		if hasModelEdgeFrom(g, root, c, "<init>", "()V") {
			t.Errorf("classpath boot layer: unexpected edge to %s", c)
		}
	}
	reasons := strings.Join(res.Reasons, "\n")
	if !strings.Contains(reasons, "boot layer holds only JDK modules") || strings.Contains(reasons, "module-layer loader") {
		t.Errorf("classpath boot layer reasons = %v", res.Reasons)
	}

	g, root, res, _ = runJPMS(t, j, nouses, ScenarioConfig{Scenario: ScenarioModulePath, AppJar: j.app}, layersFQCN, "boot", "()V")
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") ||
		!hasModelEdgeFrom(g, root, "fixture.jpms.impl.FactoryCodec", "provider", codecProvide) ||
		!hasModelEdgeFrom(g, root, "fixture.jpms.auto.AutoCodec", "<init>", "()V") {
		t.Error("module-path boot layer: want ModuleCodec, FactoryCodec.provider() and automatic AutoCodec")
	}
	if strings.Contains(strings.Join(res.Reasons, "\n"), "module-layer loader") {
		t.Errorf("boot layer must not keep the generic layer visibility gap: %v", res.Reasons)
	}

	mixed := ScenarioConfig{Scenario: ScenarioMixed, AppJar: j.app, ModulePathJars: []string{"app", "fixture-api-*.jar", "fixture-impl-*.jar"}}
	g, root, _, _ = runJPMS(t, j, nouses, mixed, layersFQCN, "boot", "()V")
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
		t.Error("mixed boot layer: missing module-path ModuleCodec")
	}
	if hasModelEdgeFrom(g, root, "fixture.jpms.auto.AutoCodec", "<init>", "()V") {
		t.Error("mixed boot layer: AutoCodec is on the class path, in no layer")
	}
}

func TestScenario_unknownLayerKeepsGap(t *testing.T) {
	j := buildJPMSJars(t)
	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)
	g, root, res, _ := runJPMS(t, j, nouses, ScenarioConfig{Scenario: ScenarioClasspath, AppJar: j.app}, layersFQCN, "given", "(Ljava/lang/ModuleLayer;)V")
	// Never a confident negative: the candidate providers stay, with a gap.
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
		t.Error("unknown layer: candidate provider edge must remain")
	}
	if !res.Incomplete || !strings.Contains(strings.Join(res.Reasons, "\n"), "module-layer loader") {
		t.Errorf("unknown layer: want the visibility gap, got %v", res.Reasons)
	}
}

// TestScenario_mixed: module-path explicit modules provide through
// `provides`, class-path jars through META-INF/services, both visible to
// ServiceLoader.load(Class).
func TestScenario_mixed(t *testing.T) {
	j := buildJPMSJars(t)
	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)
	cfg := ScenarioConfig{Scenario: ScenarioMixed, AppJar: j.app, ModulePathJars: []string{"app", "fixture-api-*.jar", "fixture-impl-*.jar"}}
	g, root, _, warnings := runJPMS(t, j, nouses, cfg, "fixture.jpms.app.Main", "main", "([Ljava/lang/String;)V")
	if len(warnings) != 0 {
		t.Errorf("warnings: %v", warnings)
	}
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.FactoryCodec", "provider", codecProvide) {
		t.Error("mixed: missing module-path FactoryCodec.provider()")
	}
	if hasModelEdgeFrom(g, root, "fixture.jpms.impl.LegacyCodec", "<init>", "()V") {
		t.Error("mixed: impl is on the module path, its META-INF/services must be ignored")
	}
	if !hasModelEdgeFrom(g, root, "fixture.jpms.auto.AutoCodec", "<init>", "()V") {
		t.Error("mixed: class-path AutoCodec must stay visible")
	}

	// The class-path half has no `uses` rule; the module-path caller does.
	cfgNoUses := ScenarioConfig{Scenario: ScenarioMixed, AppJar: j.app, ModulePathJars: []string{"fixture-nouses-*.jar", "fixture-api-*.jar"}}
	_, _, res, _ := runJPMS(t, j, nouses, cfgNoUses, "fixture.jpms.nouses.NoUses", "main", "([Ljava/lang/String;)V")
	if !strings.Contains(strings.Join(res.Reasons, "\n"), "mixed scenario: caller module fixture.jpms.nouses does not declare `uses") {
		t.Errorf("mixed no-uses reasons = %v", res.Reasons)
	}
}

func TestApplyScenario_mixedWarnings(t *testing.T) {
	j := buildJPMSJars(t)
	h, _, err := hierarchy.Build(j.app, j.deps())
	if err != nil {
		t.Fatal(err)
	}
	regs, _, _ := Inventory(h, j.app, j.deps())
	inv, _, err := InventoryModules(h, j.app, j.deps(), 0, regs)
	if err != nil {
		t.Fatal(err)
	}
	cfg := ScenarioConfig{Scenario: ScenarioMixed, AppJar: j.app, ModulePathJars: []string{"fixture-impl-*.jar", "nothing-*.jar"}}
	view, warnings := ApplyScenario(cfg, regs, &inv, h)
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "requires fixture.jpms.api, which is on the class path") {
		t.Errorf("want a class-path requires warning, got %v", warnings)
	}
	if !strings.Contains(joined, `pattern "nothing-*.jar" matches no input jar`) {
		t.Errorf("want an unmatched-pattern warning, got %v", warnings)
	}
	for _, d := range inv.Descriptors {
		if d.OnModulePath != (d.Jar == j.impl) {
			t.Errorf("%s OnModulePath=%v", filepath.Base(d.Jar), d.OnModulePath)
		}
	}
	for _, r := range view.Registrations {
		if r.SourceJar == j.auto && r.Module != "" {
			t.Errorf("class-path registration got a module: %+v", r)
		}
	}
}

// TestLoadSites_explicitLoaderIdentity: load(Class, ClassLoader) with a
// loader recognized from the method's own bytecode has no visibility
// gap and keeps its class-path provider edges; a parameter loader keeps
// both the edges and the gap.
func TestLoadSites_explicitLoaderIdentity(t *testing.T) {
	j := buildJPMSJars(t)
	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)
	h, _, err := hierarchy.Build(j.app, j.deps())
	if err != nil {
		t.Fatal(err)
	}
	loads, _, err := FindLoadSites(h)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]LoaderIdentity{"context": LoaderIdentContext, "system": LoaderIdentSystem, "nullLoader": LoaderIdentSystem, "defining": LoaderIdentDefining, "given": ""}
	seen := 0
	for _, l := range loads {
		if l.CallerFQCN != "fixture.jpms.app.Loaders" {
			continue
		}
		seen++
		if l.LoaderKind != LoaderExplicit || l.LoaderIdentity != want[l.CallerName] {
			t.Errorf("%s: kind=%s identity=%q detail=%q, want explicit/%q", l.CallerName, l.LoaderKind, l.LoaderIdentity, l.LoaderDetail, want[l.CallerName])
		}
		if l.CallerName == "context" && l.LoaderDetail != "Thread.currentThread().getContextClassLoader()" {
			t.Errorf("context detail = %q", l.LoaderDetail)
		}
		if l.CallerName == "defining" && l.LoaderDetail != "fixture.jpms.app.Loaders.class.getClassLoader()" {
			t.Errorf("defining detail = %q", l.LoaderDetail)
		}
	}
	if seen != 5 {
		t.Fatalf("Loaders load sites = %d, want 5", seen)
	}

	for name, desc := range map[string]string{"context": "()V", "system": "()V", "nullLoader": "()V", "defining": "()V"} {
		g, root, res, _ := runJPMS(t, j, nouses, ScenarioConfig{Scenario: ScenarioClasspath, AppJar: j.app}, "fixture.jpms.app.Loaders", name, desc)
		if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
			t.Errorf("%s: missing class-path provider edge", name)
		}
		if strings.Contains(strings.Join(res.Reasons, "\n"), "explicit loader") {
			t.Errorf("%s: recognized loader must not keep the visibility gap: %v", name, res.Reasons)
		}
	}
	g, root, res, _ := runJPMS(t, j, nouses, ScenarioConfig{Scenario: ScenarioClasspath, AppJar: j.app}, "fixture.jpms.app.Loaders", "given", "(Ljava/lang/ClassLoader;)V")
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
		t.Error("given: candidate edge must remain")
	}
	if !strings.Contains(strings.Join(res.Reasons, "\n"), "provider visibility for explicit loader") {
		t.Errorf("given: want the explicit-loader gap, got %v", res.Reasons)
	}
}

// TestLoadSites_parametersFromCallers covers DATA_FLOW D3's "parameters
// from callers" slice through its first consumer, ServiceLoader.load.
func TestLoadSites_parametersFromCallers(t *testing.T) {
	j := buildJPMSJars(t)
	h, _, err := hierarchy.Build(j.app, j.deps())
	if err != nil {
		t.Fatal(err)
	}
	loads, _, err := FindLoadSites(h)
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]ServiceLoadSite{}
	for _, l := range loads {
		if strings.HasPrefix(l.CallerFQCN, "fixture.jpms.app.Params") {
			by[l.CallerFQCN+"#"+l.CallerName] = l
		}
	}
	const p = "fixture.jpms.app.Params"

	if l := by[p+"#loadFor"]; l.Interface != "" || strings.Join(l.Interfaces, ",") != "fixture.jpms.api.Codec,java.lang.Runnable" || l.UnresolvedReason != "" || !l.ClassParamSet {
		t.Errorf("loadFor: %+v, want the complete set {Codec, Runnable}", l)
	}
	if l := by[p+"#secure"]; l.Interface != "fixture.jpms.api.Codec" || l.LoaderIdentity != LoaderIdentSystem || !strings.Contains(l.LoaderDetail, "null") {
		t.Errorf("secure (Jackson shape): %+v, want Codec and the system loader via null two levels up", l)
	}
	if l := by[p+"#rec"]; l.Interface != "fixture.jpms.api.Codec" || l.UnresolvedReason != "" {
		t.Errorf("rec: %+v, want recursion to converge on Codec", l)
	}
	for name, want := range map[string]string{
		p + "#viaLambda": "lambda/method-reference implementation",
		p + "#orphan":    "no caller in the analyzed input",
		p + "$Base#go":   "is overridden in fixture.jpms.app.Params$Sub",
	} {
		l := by[name]
		if len(l.Candidates()) != 0 || !strings.Contains(l.UnresolvedReason, want) {
			t.Errorf("%s: %+v, want unresolved with %q", name, l, want)
		}
	}
}

// TestParameters_noCrossCallerPath: the helper's load site has the set
// {Codec, Runnable}, but each caller's consumption is narrowed to its own
// argument, so the Runnable caller gets no Codec provider edge.
func TestParameters_noCrossCallerPath(t *testing.T) {
	j := buildJPMSJars(t)
	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)
	cp := ScenarioConfig{Scenario: ScenarioClasspath, AppJar: j.app}

	g, root, _, _ := runJPMS(t, j, nouses, cp, "fixture.jpms.app.Params", "viaCodec", "()V")
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
		t.Error("viaCodec: missing Codec provider edge")
	}
	g, root, _, _ = runJPMS(t, j, nouses, cp, "fixture.jpms.app.Params", "viaRunnable", "()V")
	if hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
		t.Error("viaRunnable: Codec provider edge from the other caller's argument")
	}
	g, root, res, _ := runJPMS(t, j, nouses, cp, "fixture.jpms.app.Params", "findAll", "(Ljava/lang/ClassLoader;)V")
	if !hasModelEdgeFrom(g, root, "fixture.jpms.impl.ModuleCodec", "<init>", "()V") {
		t.Error("findAll: missing Codec provider edge through secure()")
	}
	if strings.Contains(strings.Join(res.Reasons, "\n"), "explicit loader") {
		t.Errorf("findAll: the loader parameter resolves to null; no visibility gap expected: %v", res.Reasons)
	}
}

func TestResolveParameter_limits(t *testing.T) {
	j := buildJPMSJars(t)
	h, _, err := hierarchy.Build(j.app, j.deps())
	if err != nil {
		t.Fatal(err)
	}
	idx := dataflow.NewCallerIndex(h)
	defer idx.Close()
	pv := idx.ResolveParameter(dataflow.MethodID{Owner: "fixture.jpms.app.Params", Name: "loadFor", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"}, 0)
	if !pv.Complete() || len(pv.Values) != 2 || pv.Callers != 2 {
		t.Errorf("loadFor param = %+v", pv)
	}
	// secure's loader parameter: findAll passes its own parameter, whose
	// only caller (modules) passes null — two caller levels.
	pv = idx.ResolveParameter(dataflow.MethodID{Owner: "fixture.jpms.app.Params", Name: "secure", Desc: "(Ljava/lang/Class;Ljava/lang/ClassLoader;)Ljava/util/ServiceLoader;"}, 1)
	if !pv.Complete() || len(pv.Values) != 1 || pv.Values[0].Kind != "null" || pv.Callers != 2 {
		t.Errorf("secure loader param = %+v", pv)
	}
}

// TestLoadSites_finalFieldsFromConstructorParameters covers DATA_FLOW
// D3's "final fields set from constructor parameters" slice.
func TestLoadSites_finalFieldsFromConstructorParameters(t *testing.T) {
	j := buildJPMSJars(t)
	h, _, err := hierarchy.Build(j.app, j.deps())
	if err != nil {
		t.Fatal(err)
	}
	loads, _, err := FindLoadSites(h)
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]ServiceLoadSite{}
	for _, l := range loads {
		if strings.HasPrefix(l.CallerFQCN, "fixture.jpms.app.Captures") {
			by[l.CallerFQCN] = l
		}
	}
	if l := by["fixture.jpms.app.Captures$1"]; l.Interface != "fixture.jpms.api.Codec" || l.LoaderIdentity != LoaderIdentSystem || !strings.Contains(l.LoaderDetail, "final field") || !strings.Contains(l.ParamSource, "final field") {
		t.Errorf("anonymous capture (Jackson shape): %+v, want Codec and the system loader from val$ fields", l)
	}
	if l := by["fixture.jpms.app.Captures$Box"]; strings.Join(l.Interfaces, ",") != "fixture.jpms.api.Codec,java.lang.Runnable" || l.UnresolvedReason != "" {
		t.Errorf("Box: %+v, want the set {Codec, Runnable} from two new sites", l)
	}
	if l := by["fixture.jpms.app.Captures$Mutable"]; len(l.Candidates()) != 0 || !strings.Contains(l.UnresolvedReason, "not a final instance field") {
		t.Errorf("Mutable: %+v, want unresolved (not final)", l)
	}
}

// TestDoPrivileged covers the AccessController.doPrivileged model: the
// ASYNC_CALLBACKS catalog edge to the action's run() and the DATA_FLOW
// D3 summary "result = what run() returns", seen by the ServiceLoader
// consumption linking and by reflection.
func TestDoPrivileged(t *testing.T) {
	j := buildJPMSJars(t)
	h, _, err := hierarchy.Build(j.app, j.deps())
	if err != nil {
		t.Fatal(err)
	}
	const p = "fixture.jpms.app.Privileged"

	// Callback edges to run().
	sites := callback.FindRegistrationSites(h)
	resolved, _ := callback.ResolveCallbacks(h, sites)
	edges := callback.EdgeTargets(callback.FindCallbackEdges(h, resolved))
	targets := map[string]string{}
	for _, e := range edges {
		if e.CallerFQCN == p {
			targets[e.CallerName] = e.TargetFQCN + "#" + e.TargetMethod + e.TargetMethodDesc
		}
	}
	if targets["anon"] != p+"$1#run()Ljava/lang/Object;" || !strings.HasPrefix(targets["lambda"], p+"#lambda$") || !strings.HasPrefix(targets["exception"], p+"#lambda$") || targets["reflective"] != p+"$2#run()Ljava/lang/Object;" {
		t.Errorf("callback edges = %v", targets)
	}
	if _, ok := targets["given"]; ok {
		t.Errorf("given: an unknown action must not get a callback edge: %v", targets)
	}

	// ServiceLoader consumption linking through the doPrivileged result.
	loads, _, _ := FindLoadSites(h)
	consumptions, _, err := FindConsumptionSites(h, loads)
	if err != nil {
		t.Fatal(err)
	}
	linked := map[string]ServiceConsumptionSite{}
	for _, c := range consumptions {
		if c.CallerFQCN == p && c.CallOwner == "java.util.Iterator" && c.CallName == "next" {
			linked[c.CallerName] = c
		}
	}
	for _, name := range []string{"anon", "lambda", "exception"} {
		c := linked[name]
		if strings.Join(c.Interfaces, ",") != "fixture.jpms.api.Codec" || c.UnresolvedReason != "" {
			t.Errorf("%s: next() = %+v, want linked to Codec through doPrivileged", name, c)
		}
	}
	var givenIterator *ServiceConsumptionSite
	for i, c := range consumptions {
		if c.CallerFQCN == p && c.CallerName == "given" && c.CallOwner == serviceLoaderOwner && c.CallName == "iterator" {
			givenIterator = &consumptions[i]
		}
	}
	if givenIterator == nil || len(givenIterator.Interfaces) != 0 || !strings.Contains(givenIterator.UnresolvedReason, "doPrivileged action is") {
		t.Errorf("given: iterator() = %+v, want an explicit doPrivileged gap", givenIterator)
	}

	// Reflection: a getDeclaredMethod performed inside the action.
	var site reflectinvoke.InvokeSite
	for _, s := range reflectinvoke.FindInvokeSites(h) {
		if s.CallerFQCN == p && s.CallerName == "reflective" {
			site = s
		}
	}
	ri, _ := reflectinvoke.ResolveInvokes(h, []reflectinvoke.InvokeSite{site})
	if len(ri) != 1 || ri[0].OwnerFQCN != p+"$Target" || ri[0].MethodName != "hit" {
		t.Errorf("reflective invoke = %+v, want Target#hit via doPrivileged", ri)
	}
}

// TestRegression_duplicatesAndTwoLoaders: S4's compiled-fixture cases.
// A provider listed twice in one file and shipped in two jars is
// materialized once, with a duplicate warning; two loaders for two
// services in one method do not mix their providers.
func TestRegression_duplicatesAndTwoLoaders(t *testing.T) {
	j := buildJPMSJars(t)
	copyJar := zipFixture(t, filepath.Join(jpmsClasses, "auto"), t.TempDir(), "fixture-auto-copy-1.0.jar", nil)
	deps := append(j.deps(), input.DepJar{Path: copyJar})
	h, _, err := hierarchy.Build(j.app, deps)
	if err != nil {
		t.Fatal(err)
	}
	regs, warnings, err := Inventory(h, j.app, deps)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "fixture.jpms.auto.AutoCodec") {
		t.Errorf("want a duplicate-entry warning for AutoCodec, got %v", warnings)
	}
	loads, _, _ := FindLoadSites(h)
	n := 0
	for _, m := range FindMaterializations(h, loads, regs) {
		if m.Registration.ProviderClass == "fixture.jpms.auto.AutoCodec" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("AutoCodec materialized %d times, want once despite two jars and a repeated line", n)
	}

	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)
	g, root, _, _ := runJPMS(t, j, nouses, ScenarioConfig{Scenario: ScenarioClasspath, AppJar: j.app}, "fixture.jpms.app.Regression", "twoLoaders", "()V")
	consumptions, _, _ := FindConsumptionSites(h, loads)
	var codecNext, hasherNext int = -1, -1
	for _, c := range consumptions {
		if c.CallerFQCN != "fixture.jpms.app.Regression" || c.CallName != "next" {
			continue
		}
		switch strings.Join(c.Interfaces, ",") {
		case "fixture.jpms.api.Codec":
			codecNext = c.PC
		case "fixture.jpms.api.Hasher":
			hasherNext = c.PC
		default:
			t.Errorf("next() at pc=%d linked to %v, want exactly one service", c.PC, c.Interfaces)
		}
	}
	if codecNext < 0 || hasherNext < 0 {
		t.Fatalf("want one Codec and one Hasher next() site, got codec=%d hasher=%d", codecNext, hasherNext)
	}
	codecCtor, _ := g.ID("fixture.jpms.auto.AutoCodec", "<init>", "()V")
	hasherCtor, _ := g.ID("fixture.jpms.auto.AutoHasher", "<init>", "()V")
	if !hasModelEdgeAt(g, root, codecCtor, codecNext) || hasModelEdgeAt(g, root, hasherCtor, codecNext) {
		t.Error("Codec next() must reach AutoCodec and not AutoHasher")
	}
	if !hasModelEdgeAt(g, root, hasherCtor, hasherNext) || hasModelEdgeAt(g, root, codecCtor, hasherNext) {
		t.Error("Hasher next() must reach AutoHasher and not AutoCodec")
	}
}
