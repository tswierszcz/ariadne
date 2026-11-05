package serviceload

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/analysis"
	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/models"
)

// oracleProviders are every provider class the fixtures declare.
var oracleProviders = map[string]bool{
	"fixture.provider.DecoderProvider": true,
	"fixture.jpms.impl.ModuleCodec":    true, "fixture.jpms.impl.LegacyCodec": true, "fixture.jpms.impl.FactoryCodec": true,
	"fixture.jpms.auto.AutoCodec": true, "fixture.jpms.auto.AutoHasher": true,
}

func readTSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var rows [][]string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows
}

// modelProviders runs the real pipeline from root and returns the
// provider classes the model materializes (BasisModel edges to a
// provider's <init> or provider()) anywhere in the reachable graph.
func modelProviders(t *testing.T, app string, deps []input.DepJar, cfg ScenarioConfig, root string) ([]string, bool) {
	t.Helper()
	h, _, err := hierarchy.Build(app, deps)
	if err != nil {
		t.Fatal(err)
	}
	regs, _, err := Inventory(h, app, deps)
	if err != nil {
		t.Fatal(err)
	}
	inv, _, err := InventoryModules(h, app, deps, 0, regs)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AppJar = app
	view, _ := ApplyScenario(cfg, regs, &inv, h)
	loads, _, _ := FindLoadSites(h)
	consumptions, _, _ := FindConsumptionSites(h, loads)
	eval := NewModelEvaluator(h, loads, FindMaterializations(h, loads, view.Registrations), consumptions)
	eval.SetScenario(view)
	g := callgraph.Build(h, nil)

	cls, name, _ := strings.Cut(root, "#")
	ci, ok := h.Lookup(cls)
	if !ok {
		t.Fatalf("no class %s", cls)
	}
	ms := ci.MethodsByName(name)
	if len(ms) != 1 {
		t.Fatalf("%s: want one method, got %d", root, len(ms))
	}
	rootID, _ := g.ID(cls, name, ms[0].Descriptor)
	registry := models.NewRegistry()
	RegisterModel(registry)
	RegisterDispatchModels(registry, consumptions)
	res, err := analysis.Run(h, g, registry, eval, []uint32{rootID}, analysis.Options{MaxIterations: 4})
	if err != nil {
		t.Fatal(err)
	}

	seen := map[uint32]bool{rootID: true}
	queue := []uint32{rootID}
	found := map[string]bool{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, e := range g.Edges(id) {
			if e.Basis == callgraph.BasisModel {
				key := g.Key(e.To)
				owner, method, _ := strings.Cut(key, "#")
				if oracleProviders[owner] && (strings.HasPrefix(method, "<init>") || strings.HasPrefix(method, "provider(")) {
					found[owner] = true
				}
			}
			if !seen[e.To] {
				seen[e.To] = true
				queue = append(queue, e.To)
			}
		}
	}
	var out []string
	for p := range found {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, res.Incomplete
}

// TestServiceLoaderOracle (SERVICE_LOADING S4, ROADMAP B4) compares the
// model with what the real JVM instantiated for the same entry method
// and launch (testdata/oracle/expected.tsv, from
// scripts/serviceloader-oracle.sh). Every provider the JVM created must
// be modeled; a modeled provider the JVM did not create is allowed only
// when testdata/oracle/model-extras.tsv names it with a reason (a static
// may-analysis over-approximates), and every listed extra must still
// occur. A provider the JVM created but the model misses is allowed only
// when testdata/oracle/model-gaps.tsv names it as a known gap and the
// analysis reports itself Incomplete for that root — never a silent miss.
func TestServiceLoaderOracle(t *testing.T) {
	expected := readTSV(t, "../../testdata/oracle/expected.tsv")
	extras := map[string]bool{}
	for _, r := range readTSV(t, "../../testdata/oracle/model-extras.tsv") {
		extras[r[0]+"\t"+r[1]+"\t"+r[2]] = true
	}
	usedExtras := map[string]bool{}
	gaps := map[string]bool{}
	for _, r := range readTSV(t, "../../testdata/oracle/model-gaps.tsv") {
		gaps[r[0]+"\t"+r[1]+"\t"+r[2]] = true
	}
	usedGaps := map[string]bool{}
	checked, missed, extraCount := 0, 0, 0

	slPrepared, err := input.PrepareAppInput(filepath.Join("..", "..", "testdata", "serviceloader", "classes"))
	if err != nil {
		t.Fatal(err)
	}
	defer slPrepared.Cleanup()
	j := buildJPMSJars(t)
	nouses := zipFixture(t, filepath.Join(jpmsClasses, "mods", "fixture.jpms.nouses"), t.TempDir(), "fixture-nouses-1.0.jar", nil)
	jpmsDeps := []input.DepJar{{Path: j.api}, {Path: j.impl}, {Path: j.auto}, {Path: nouses}}
	scenarios := map[string]ScenarioConfig{
		"classpath":   {Scenario: ScenarioClasspath},
		"module-path": {Scenario: ScenarioModulePath},
		"mixed":       {Scenario: ScenarioMixed, ModulePathJars: []string{"app", "fixture-api-*.jar", "fixture-impl-*.jar"}},
	}

	for _, row := range expected {
		scenario, entry, want := row[0], row[1], row[2]
		app, deps := j.app, jpmsDeps
		if strings.HasPrefix(entry, "fixture.app.") {
			app, deps = slPrepared.AppClassesJar, nil
		}
		if strings.HasPrefix(entry, "fixture.jpms.nouses.") {
			continue // its root is outside the app jar; covered by TestScenario_missingUsesBlocksProviders
		}
		got, incomplete := modelProviders(t, app, deps, scenarios[scenario], entry)
		checked++
		wantSet := map[string]bool{}
		if want != "-" {
			for _, p := range strings.Split(want, ",") {
				wantSet[p] = true
			}
		}
		gotSet := map[string]bool{}
		for _, p := range got {
			gotSet[p] = true
			if !wantSet[p] {
				key := scenario + "\t" + entry + "\t" + p
				if extras[key] {
					usedExtras[key] = true
					extraCount++
				} else {
					t.Errorf("%s %s: model materializes %s, which the JVM did not (not in model-extras.tsv)", scenario, entry, p)
				}
			}
		}
		for p := range wantSet {
			if gotSet[p] {
				continue
			}
			key := scenario + "\t" + entry + "\t" + p
			switch {
			case gaps[key] && incomplete:
				usedGaps[key] = true
				missed++
			case gaps[key]:
				t.Errorf("%s %s: known gap for %s, but the analysis does not report itself incomplete", scenario, entry, p)
			default:
				t.Errorf("%s %s: the JVM instantiated %s but the model has no edge to it (missed target)", scenario, entry, p)
			}
		}
	}
	for key := range extras {
		if !usedExtras[key] {
			t.Errorf("model-extras.tsv lists %q, which the model no longer produces", key)
		}
	}
	for key := range gaps {
		if !usedGaps[key] {
			t.Errorf("model-gaps.tsv lists %q, which the model now covers or the oracle no longer expects", key)
		}
	}
	t.Logf("oracle: %d entry/scenario pairs checked, %d known missed targets (declared gaps), %d declared over-approximations", checked, missed, extraCount)
}
