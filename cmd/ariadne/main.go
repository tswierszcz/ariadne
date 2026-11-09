// Command ariadne reports which known vulnerabilities (CVE/GHSA) in a JVM
// application's dependencies are reachable from the application's own code.
package main

import (
	"ariadne/internal/flowsum"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"ariadne/internal/alloc"
	"ariadne/internal/analysis"
	"ariadne/internal/aop"
	"ariadne/internal/atomicfile"
	"ariadne/internal/callback"
	"ariadne/internal/callgraph"
	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/deserialize"
	"ariadne/internal/exploit"
	"ariadne/internal/hierarchy"
	"ariadne/internal/httpcache"
	"ariadne/internal/input"
	"ariadne/internal/methodhandle"
	"ariadne/internal/models"
	"ariadne/internal/profile"
	"ariadne/internal/reach"
	"ariadne/internal/reflectinvoke"
	"ariadne/internal/report"
	"ariadne/internal/resultsv2"
	"ariadne/internal/runtimeevents"
	"ariadne/internal/samcall"
	"ariadne/internal/serviceload"
	"ariadne/internal/servletdispatch"
	"ariadne/internal/snapshot"
	"ariadne/internal/spidispatch"
	"ariadne/internal/springxml"
	"ariadne/internal/taint"
	"ariadne/internal/vulnmap"
	"ariadne/internal/webxml"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}

type config struct {
	githubToken     string // GITHUB_TOKEN: authenticates fix-commit metadata requests (CVE:F5)
	centralURL      string // ARIADNE_MAVEN_CENTRAL: repository root for embedded-component reference jars (CVE:F6a); default Maven Central
	app             string
	deps            string
	depsList        string
	vulnMethods     string
	embeddedComps   string // -embedded-components: declared embedded components and relocations (CVE:F7a)
	sbom            string // -sbom: CycloneDX JSON SBOM, embedded-component candidates tied to jars by SHA-256 (CVE:F6b)
	out             string
	jsonOut         string
	jsonFormat      string
	llmReaskChanged bool
	snapshotOut     string
	snapshotIn      string
	snapshotCompare string
	refresh         bool
	allowStale      bool
	dataTTL         time.Duration
	maxDataAge      time.Duration
	cacheDir        string
	offline         bool
	showVersion     bool

	profile        string
	profileFile    string
	roots          string
	targetJava     int
	activeProfiles string
	jpmsScenario   string
	jpmsModulePath string
	taint          bool
	callgraphMode  string
	graphOut       string
	graphCompare   string
	runtimeTrace   string
	runtimeFilter  string
	runtimeStatus  string

	llm        bool
	llmModel   string
	llmBaseURL string
	llmConfig  vulnmap.LLMConfig // resolved from the flags above and the environment; set when llm or exploitCheck is true

	exploitCheck   bool
	frameworkCheck bool
}

// moduleLabel names a PreparedApp.WebXMLs key for an error message —
// "" (a standalone WAR, the common case) reads as "the app" rather
// than an empty, confusing string.
func moduleLabel(module string) string {
	if module == "" {
		return "the app"
	}
	return module
}

func run(args []string, stdout, stderr io.Writer, getenv func(string) string) int {
	fs := flag.NewFlagSet("ariadne", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var cfg config
	fs.StringVar(&cfg.app, "app", "", "the application's classes: a directory, a plain jar, or a Spring Boot fat jar")
	fs.StringVar(&cfg.deps, "deps", "", "directory of dependency jars (not needed for a Spring Boot fat jar, which carries its own)")
	fs.StringVar(&cfg.depsList, "deps-list", "", "optional `mvn dependency:list` output, used as a fallback when a jar has no pom.properties")
	fs.StringVar(&cfg.vulnMethods, "vuln-methods", "", "override vuln_methods.yaml (default: ariadne's embedded table)")
	fs.StringVar(&cfg.embeddedComps, "embedded-components", "", "YAML file declaring components embedded in dependency jars and their package relocations, pinned to each jar's SHA-256 (CVE:F7a); every declaration is verified against the classes")
	fs.StringVar(&cfg.profile, "profile", "auto", "which profile's framework knowledge to use: auto, generic, spring, or a name from -profile-file")
	fs.StringVar(&cfg.sbom, "sbom", "", "CycloneDX JSON SBOM (spec 1.4-1.6): components nested in a component whose SHA-256 is an analyzed dependency jar's become embedded-component candidates of that jar, verified by their classes (CVE:F6b); nothing else from the SBOM is used")
	fs.StringVar(&cfg.profileFile, "profile-file", "", "load/override one profile from a YAML file (same shape as the embedded defaults), in addition to ariadne's own spring/generic")
	fs.StringVar(&cfg.roots, "roots", "all", `BFS roots: "all" (every application method, the default) or "profile" (only the active profile's entry points: annotated methods/classes, main)`)
	fs.IntVar(&cfg.targetJava, "target-java", 0, "Java feature version (e.g. 11, 17) to resolve multi-release JAR variants against; unset falls back to every jar's base classes, not the highest bytecode version seen (JVM_GRAPH G4)")
	fs.StringVar(&cfg.activeProfiles, "active-profiles", "", "comma-separated Spring profile names to treat as active when evaluating @Profile conditions (FRAMEWORKS F1); unset reports every @Profile condition as unknown, never a guessed false")
	fs.StringVar(&cfg.jpmsScenario, "jpms-scenario", "classpath", `JPMS launch scenario for ServiceLoader providers (SERVICE_LOADING S3): "classpath" (every jar in the unnamed module; module-info ignored) or "module-path" (every jar, the app included, on the module path: explicit modules provide only through "provides", automatic modules through META-INF/services, explicit callers need "uses"), or "mixed" (only the -jpms-module-path jars on the module path, assumed resolved as with --add-modules ALL-MODULE-PATH)`)
	fs.StringVar(&cfg.jpmsModulePath, "jpms-module-path", "", `with -jpms-scenario mixed: comma-separated jar file-name patterns (path.Match syntax, e.g. "commons-text-*.jar") for the jars on the module path; "app" names the analyzed application; every other jar stays on the class path`)
	fs.StringVar(&cfg.out, "out", "", "report.md output path (required)")
	fs.StringVar(&cfg.jsonOut, "json", "", "results.json output path (required)")
	fs.StringVar(&cfg.jsonFormat, "json-format", "v1", `results.json schema: "v1" (default: one row per advisory) or "v2" (CVE:F1c: components, findings per advisory × component instance, coverage, inputs reference; docs/RESULTS_CONTRACT.MD §5)`)
	fs.StringVar(&cfg.cacheDir, "cache-dir", ".cache", "disk cache directory for OSV/GitHub HTTP responses")
	fs.StringVar(&cfg.snapshotOut, "snapshot-out", "", "write a data snapshot to this new directory (CVE:F3): every OSV/fix-commit/LLM lookup outcome of the run, input hashes, configuration, analyzer and algorithm versions, derivation keys and verdicts")
	fs.StringVar(&cfg.snapshotIn, "snapshot-in", "", "replay a data snapshot: every lookup is answered from it (no cache, no network); inputs, analyzer and algorithm versions must match it")
	fs.StringVar(&cfg.snapshotCompare, "snapshot-compare", "", "compare this run with an earlier snapshot: list the advisories whose derived results were invalidated, and why")
	fs.BoolVar(&cfg.refresh, "refresh", false, "revalidate every cached OSV/advisory response now, whatever its age (CVE:F2); never re-asks an LLM; not with -offline")
	fs.StringVar(&cfg.callgraphMode, "callgraph-mode", callgraph.ModeCHA, `call-graph strategy for virtual/interface dispatch (GRAPH_PRECISION P0/P2): "cha" (class hierarchy analysis, the default) or "rta" (opt-in: CHA narrowed to the classes created by code reachable from the roots, with the unknown-allocation fallback)`)
	fs.StringVar(&cfg.graphOut, "graph-out", "", "write the final call graph (every edge with a stable identity, its call-site PCs, the graph's settings and the findings' paths) as sorted text to this file, for -graph-compare")
	fs.StringVar(&cfg.runtimeTrace, "runtime-trace", "", "import a runtime trace (RUNTIME_NATIVE N1, format ariadne-runtime-trace-1, docs/RUNTIME_TRACE.MD; e.g. from scripts/JfrToTrace.java): loaded classes, methods seen executing and measured caller → callee relations, matched to the analyzed classes by content hash and shown next to the static results; it never changes the graph or a verdict, and the absence of an event proves nothing")
	fs.StringVar(&cfg.runtimeFilter, "runtime-filter-out", "", "write a JFR settings file (.jfc) that traces, on entry, the mapped vulnerable methods of this run's findings (jdk.MethodTrace, JDK 25+), to record the next run with: java -XX:StartFlightRecording=settings=profile,settings=FILE ...; then import the converted recording with -runtime-trace")
	fs.StringVar(&cfg.runtimeStatus, "runtime-filter-status", defaultRuntimeFilterStatus, "with -runtime-filter-out: comma-separated verdicts whose mapped methods are traced")
	fs.StringVar(&cfg.graphCompare, "graph-compare", "", "compare this run's call graph with one written earlier by -graph-out: edges added/removed by basis, findings whose verdict or path changed and the edges that explain it, and any difference in the graph settings")
	fs.BoolVar(&cfg.taint, "taint", false, "DATA_FLOW D4 (opt-in, report-only): assess whether request data (Spring MVC/JAX-RS handler parameters, ServletRequest accessors) reaches the data each deserialization entry reads; adds a taint result per entry, never changes a verdict")
	fs.BoolVar(&cfg.allowStale, "allow-stale", false, "when an online refresh fails, use the cached copy and mark the report as built on stale data (never past -max-data-age)")
	fs.DurationVar(&cfg.dataTTL, "data-ttl", httpcache.DefaultTTL, "how long a fetched OSV/advisory response counts as fresh before it is fetched again (CVE:F2)")
	fs.DurationVar(&cfg.maxDataAge, "max-data-age", 0, "hard limit on the age of any OSV/advisory data used, offline included; data of unknown age (cached before CVE:F2) fails it. 0 = no limit")
	fs.BoolVar(&cfg.offline, "offline", false, "serve only from the cache; fail clearly on a miss instead of touching the network")
	fs.BoolVar(&cfg.llm, "llm", false, "also map CVEs no other source could to vulnerable methods by asking an LLM (needs ARIADNE_LLM_API_KEY and a model)")
	fs.BoolVar(&cfg.llmReaskChanged, "llm-reask-changed", false, "with -llm: ask again when a CVE's advisory text or fix diff changed since its cached answer (default: reuse the previous answer, with a warning, so a data refresh never triggers paid requests by itself)")
	fs.StringVar(&cfg.llmModel, "llm-model", "", "LLM model id; overrides ARIADNE_LLM_MODEL")
	fs.StringVar(&cfg.llmBaseURL, "llm-base-url", "", "OpenAI-compatible base URL; overrides ARIADNE_LLM_BASE_URL (default "+vulnmap.DefaultLLMBaseURL+")")
	fs.BoolVar(&cfg.exploitCheck, "exploit-check", false, "ask an LLM, for each REACHABLE CVE, whether its call path plausibly carries attacker-controlled input (PLAN.md Phase 8; needs the same ARIADNE_LLM_* settings as -llm, and makes one extra LLM request per REACHABLE CVE)")
	fs.BoolVar(&cfg.frameworkCheck, "framework-check", false, "ask an LLM, for each FRAMEWORK CVE, whether its call path looks like a real framework call chain or a CHA over-approximation artifact (needs the same ARIADNE_LLM_* settings as -llm, and makes one extra LLM request per FRAMEWORK CVE)")
	fs.BoolVar(&cfg.showVersion, "version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if cfg.showVersion {
		fmt.Fprintf(stdout, "ariadne %s\n", version())
		return 0
	}
	if cfg.snapshotIn != "" && !cfg.refresh {
		cfg.offline = true // a replay never touches the network
	}

	cfg.githubToken = getenv("GITHUB_TOKEN")
	cfg.centralURL = getenv("ARIADNE_MAVEN_CENTRAL")
	if cfg.llm || cfg.exploitCheck || cfg.frameworkCheck {
		cfg.llmConfig = vulnmap.LLMConfigFromEnv(getenv)
		if cfg.llmModel != "" {
			cfg.llmConfig.Model = cfg.llmModel
		}
		if cfg.llmBaseURL != "" {
			cfg.llmConfig.BaseURL = cfg.llmBaseURL
		}
	}

	if err := validate(cfg); err != nil {
		fmt.Fprintln(stderr, "ariadne:", err)
		fs.Usage()
		return 2
	}

	if err := analyze(context.Background(), cfg, stdout); err != nil {
		fmt.Fprintln(stderr, "ariadne:", err)
		return 1
	}
	return 0
}

func validate(cfg config) error {
	if cfg.app == "" {
		return fmt.Errorf("-app is required")
	}
	if cfg.out == "" {
		return fmt.Errorf("-out is required")
	}
	if cfg.jsonOut == "" {
		return fmt.Errorf("-json is required")
	}
	if cfg.llm || cfg.exploitCheck || cfg.frameworkCheck {
		if err := cfg.llmConfig.Validate(cfg.offline); err != nil {
			return err
		}
	}
	if cfg.snapshotIn != "" && (cfg.refresh || cfg.snapshotOut != "") {
		return fmt.Errorf("-snapshot-in replays recorded data: it cannot be combined with -refresh or -snapshot-out")
	}
	if cfg.refresh && cfg.offline {
		return fmt.Errorf("-refresh and -offline cannot be combined: a refresh needs the network")
	}
	if cfg.dataTTL <= 0 {
		return fmt.Errorf("-data-ttl must be positive, got %s", cfg.dataTTL)
	}
	if cfg.maxDataAge < 0 {
		return fmt.Errorf("-max-data-age must not be negative, got %s", cfg.maxDataAge)
	}
	if cfg.jsonFormat != "v1" && cfg.jsonFormat != "v2" {
		return fmt.Errorf("-json-format must be %q or %q, got %q", "v1", "v2", cfg.jsonFormat)
	}
	if cfg.roots != "all" && cfg.roots != "profile" {
		return fmt.Errorf("-roots must be %q or %q, got %q", "all", "profile", cfg.roots)
	}
	for _, s := range strings.Split(cfg.runtimeStatus, ",") {
		switch reach.Status(strings.TrimSpace(s)) {
		case reach.StatusReachable, reach.StatusFramework, reach.StatusNotReachable, reach.StatusUnknown, reach.StatusNotApplicable:
		default:
			return fmt.Errorf("-runtime-filter-status: unknown verdict %q", s)
		}
	}
	if _, err := callgraph.StrategyFor(cfg.callgraphMode); err != nil {
		return fmt.Errorf("-callgraph-mode: %w", err)
	}
	if cfg.targetJava < 0 {
		return fmt.Errorf("-target-java must be a positive Java feature version, got %d", cfg.targetJava)
	}
	scenario, err := serviceload.ParseScenario(cfg.jpmsScenario)
	if err != nil {
		return fmt.Errorf("-jpms-scenario: %w", err)
	}
	patterns, err := serviceload.ParseModulePathJars(cfg.jpmsModulePath)
	if err != nil {
		return fmt.Errorf("-jpms-module-path: %w", err)
	}
	if scenario == serviceload.ScenarioMixed && len(patterns) == 0 {
		return fmt.Errorf("-jpms-scenario mixed needs -jpms-module-path")
	}
	if scenario != serviceload.ScenarioMixed && len(patterns) > 0 {
		return fmt.Errorf("-jpms-module-path is only valid with -jpms-scenario mixed")
	}
	return nil
}

func analyze(ctx context.Context, cfg config, stdout io.Writer) error {
	log := func(format string, a ...any) { fmt.Fprintf(stdout, format+"\n", a...) }

	sr, err := newSnapshotRun(cfg)
	if err != nil {
		return err
	}
	// GRAPH_PRECISION P0: the earlier graph is read first, so a bad file
	// fails before the analysis.
	var graphBefore *callgraph.Export
	if cfg.graphCompare != "" {
		f, err := os.Open(cfg.graphCompare)
		if err != nil {
			return fmt.Errorf("-graph-compare: %w", err)
		}
		ex, err := callgraph.Read(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("-graph-compare %s: %w", cfg.graphCompare, err)
		}
		graphBefore = &ex
	}
	// RUNTIME_NATIVE N1: the trace is validated before the analysis too.
	var runtimeTrace *runtimeevents.Trace
	if cfg.runtimeTrace != "" {
		t, err := loadRuntimeTrace(cfg.runtimeTrace)
		if err != nil {
			return err
		}
		runtimeTrace = t
	}
	if sr.replay != nil {
		log("snapshot: replaying %s (data set %s, created %s)", cfg.snapshotIn, sr.replay.DataSetID, sr.replay.CreatedAt.Format(time.RFC3339))
	}

	log("preparing application input from %s", cfg.app)
	prepared, err := input.PrepareAppInput(cfg.app)
	if err != nil {
		return fmt.Errorf("preparing -app %s: %w", cfg.app, err)
	}
	defer prepared.Cleanup()

	depsDir := cfg.deps
	if prepared.ExtractedLibs != "" {
		depsDir = prepared.ExtractedLibs // a Spring Boot fat jar carries its own BOOT-INF/lib
	}

	// FRAMEWORKS F1: parse every captured WEB-INF/web.xml (a WAR has
	// one keyed "", an EAR can have one per nested WAR module) — sorted
	// by module name, since map iteration order is random and this
	// codebase's own byte-identical-results.json determinism contract
	// does not tolerate that.
	var webXMLInput report.WebXMLInput
	// servletClasses/filterClasses/listenerClasses (deduplicated, in
	// first-seen order) are FRAMEWORKS F2's own input to
	// profile.WebXMLRoots/WebXMLAppRoots, collected here since they're
	// the one raw fact (which classes web.xml names) that only exists
	// at this point, before h/g exist.
	var servletClasses, filterClasses, listenerClasses []string
	if len(prepared.WebXMLs) > 0 {
		modules := make([]string, 0, len(prepared.WebXMLs))
		for module := range prepared.WebXMLs {
			modules = append(modules, module)
		}
		sort.Strings(modules)
		skippedSeen := map[string]bool{}
		servletClassSeen := map[string]bool{}
		filterClassSeen := map[string]bool{}
		listenerClassSeen := map[string]bool{}
		for _, module := range modules {
			parsed, err := webxml.Parse(bytes.NewReader(prepared.WebXMLs[module]))
			if err != nil {
				return fmt.Errorf("parsing %s's web.xml: %w", moduleLabel(module), err)
			}
			for _, s := range parsed.Servlets {
				webXMLInput.Servlets = append(webXMLInput.Servlets, report.ServletRow{Module: module, Name: s.Name, Class: s.Class})
				if !servletClassSeen[s.Class] {
					servletClassSeen[s.Class] = true
					servletClasses = append(servletClasses, s.Class)
				}
			}
			for _, m := range parsed.ServletMappings {
				webXMLInput.ServletMappings = append(webXMLInput.ServletMappings, report.ServletMappingRow{Module: module, Name: m.Name, URLPattern: m.URLPattern})
			}
			for _, f := range parsed.Filters {
				webXMLInput.Filters = append(webXMLInput.Filters, report.FilterRow{Module: module, Name: f.Name, Class: f.Class})
				if !filterClassSeen[f.Class] {
					filterClassSeen[f.Class] = true
					filterClasses = append(filterClasses, f.Class)
				}
			}
			for _, m := range parsed.FilterMappings {
				webXMLInput.FilterMappings = append(webXMLInput.FilterMappings, report.FilterMappingRow{Module: module, Name: m.Name, URLPattern: m.URLPattern, ServletName: m.ServletName})
			}
			for _, l := range parsed.Listeners {
				webXMLInput.Listeners = append(webXMLInput.Listeners, report.ListenerRow{Module: module, Class: l.Class})
				if !listenerClassSeen[l.Class] {
					listenerClassSeen[l.Class] = true
					listenerClasses = append(listenerClasses, l.Class)
				}
			}
			for _, p := range parsed.ContextParams {
				webXMLInput.ContextParams = append(webXMLInput.ContextParams, report.ContextParamRow{Module: module, Name: p.Name, Value: p.Value})
			}
			for _, elem := range parsed.Skipped {
				if !skippedSeen[elem] {
					skippedSeen[elem] = true
					webXMLInput.Skipped = append(webXMLInput.Skipped, elem)
				}
			}
		}
		log("web.xml: %d module(s), %d servlets, %d filters, %d listeners, %d context-params found (%d unmodeled element kinds)",
			len(modules), len(webXMLInput.Servlets), len(webXMLInput.Filters), len(webXMLInput.Listeners), len(webXMLInput.ContextParams), len(webXMLInput.Skipped))
	}

	// FRAMEWORKS F1: Spring classic XML bean definitions found on the
	// app's own classpath (never a dependency jar).
	var springXMLBeans []report.XMLBeanRow
	beansXMLResources, beansXMLWarnings, err := input.FindBeansXML(prepared.AppClassesJar)
	if err != nil {
		return fmt.Errorf("scanning %s for Spring XML bean definitions: %w", prepared.AppClassesJar, err)
	}
	if len(beansXMLResources) > 0 {
		resources := make([]string, 0, len(beansXMLResources))
		for path := range beansXMLResources {
			resources = append(resources, path)
		}
		sort.Strings(resources)
		for _, path := range resources {
			parsed, err := springxml.Parse(bytes.NewReader(beansXMLResources[path]))
			if err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}
			for _, b := range parsed.Beans {
				springXMLBeans = append(springXMLBeans, report.XMLBeanRow{Source: path, ID: b.ID, Class: b.Class, Scope: b.Scope, Primary: b.Primary})
			}
		}
		log("Spring XML bean definitions: %d resource(s), %d beans found", len(resources), len(springXMLBeans))
	}

	var depsList []input.GAV
	if cfg.depsList != "" {
		f, err := os.Open(cfg.depsList)
		if err != nil {
			return fmt.Errorf("opening -deps-list %s: %w", cfg.depsList, err)
		}
		depsList, err = input.ParseDepsList(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("parsing -deps-list %s: %w", cfg.depsList, err)
		}
	}

	var depJars []input.DepJar
	if depsDir != "" {
		depJars, err = input.ResolveDepJars(depsDir, depsList)
		if err != nil {
			return fmt.Errorf("resolving dependency jars in %s: %w", depsDir, err)
		}
	}
	var identified int
	var unidentified []string
	for _, dj := range depJars {
		if dj.GAV == nil {
			unidentified = append(unidentified, dj.Path)
		} else {
			identified++
		}
	}
	log("resolved %d/%d dependency jars to Maven coordinates", identified, len(depJars))

	log("building class hierarchy")
	h, warnings, err := hierarchy.BuildWithTarget(prepared.AppClassesJar, depJars, cfg.targetJava)
	if err != nil {
		return fmt.Errorf("building hierarchy: %w", err)
	}
	warnings = append(warnings, beansXMLWarnings...)
	maxBytecode := 0
	mrVariants := 0
	for _, ci := range h.All() {
		if int(ci.MajorVersion) > maxBytecode {
			maxBytecode = int(ci.MajorVersion)
		}
		if ci.VariantVersion > 0 {
			mrVariants++
		}
	}
	if cfg.targetJava > 0 {
		log("hierarchy: %d classes (max bytecode version %d, target Java %d, %d multi-release variants selected)",
			h.Len(), maxBytecode, cfg.targetJava, mrVariants)
	} else {
		log("hierarchy: %d classes (max bytecode version %d)", h.Len(), maxBytecode)
	}

	// CVE:F1a: artifacts (content hash, stable location) and component
	// instances with identity evidence — an inventory only; verdicts are
	// unchanged (docs/RESULTS_CONTRACT.MD, "Component identity").
	components, err := component.Build(component.Inputs{
		AppInput: cfg.app, AppJar: prepared.AppClassesJar,
		DepsDir: depsDir, Extracted: prepared.ExtractedLibs != "", LibEntries: prepared.LibEntries,
		DepJars: depJars,
	}, h)
	if err != nil {
		return fmt.Errorf("building the component inventory: %w", err)
	}
	snapInputs, err := snapshotInputs(cfg, components)
	if err != nil {
		return err
	}
	if sr.replay != nil {
		if err := snapshot.CheckInputs(*sr.replay, snapInputs); err != nil {
			return err
		}
		log("snapshot: %d inputs match the manifest", len(snapInputs))
	}
	nArtifacts, nIdentified, nUnidentified, nConflicts := components.Counts()
	log("components: %d artifacts, %d identified instances, %d unidentified, %d class conflicts (first copy wins)", nArtifacts, nIdentified, nUnidentified, nConflicts)

	// SERVICE_LOADING S0: a declarative inventory of META-INF/services
	// provider registrations, never a call-graph edge (see
	// serviceload.ServiceRegistration's own doc comment).
	svcRegs, svcWarnings, err := serviceload.Inventory(h, prepared.AppClassesJar, depJars)
	if err != nil {
		return fmt.Errorf("inventorying ServiceLoader providers: %w", err)
	}
	warnings = append(warnings, svcWarnings...)
	log("ServiceLoader: %d provider registrations found", len(svcRegs))

	// SERVICE_LOADING S1: resolve the service-interface Class argument at
	// every ServiceLoader.load(...)/loadInstalled(...) call site, using
	// DATA_FLOW D0-D2's local-value analysis — still never a call-graph
	// edge (see serviceload.ServiceLoadSite's own doc comment).
	svcLoadSites, svcLoadWarnings, err := serviceload.FindLoadSites(h)
	if err != nil {
		return fmt.Errorf("finding ServiceLoader.load call sites: %w", err)
	}
	warnings = append(warnings, svcLoadWarnings...)
	resolvedLoadSites := 0
	for _, s := range svcLoadSites {
		if len(s.Candidates()) > 0 {
			resolvedLoadSites++
		}
	}
	log("ServiceLoader: %d load() call sites found, %d resolved to a service interface", len(svcLoadSites), resolvedLoadSites)
	svcConsumptionSites, svcConsumptionWarnings, err := serviceload.FindConsumptionSites(h, svcLoadSites)
	if err != nil {
		return fmt.Errorf("finding ServiceLoader consumption call sites: %w", err)
	}
	warnings = append(warnings, svcConsumptionWarnings...)
	linkedConsumptionSites := 0
	for _, site := range svcConsumptionSites {
		if len(site.Interfaces) > 0 && site.UnresolvedReason == "" {
			linkedConsumptionSites++
		}
	}
	log("ServiceLoader: %d consumption/dispatch sites found, %d linked to a load site", len(svcConsumptionSites), linkedConsumptionSites)

	// SERVICE_LOADING S3 (first slice): module descriptors and `provides`
	// declarations, interpreted under the class-path scenario only —
	// inventory and diagnostics, never registrations or edges (see
	// serviceload.ScenarioClasspath's own doc comment).
	moduleInventory, moduleWarnings, err := serviceload.InventoryModules(h, prepared.AppClassesJar, depJars, cfg.targetJava, svcRegs)
	if err != nil {
		return fmt.Errorf("inventorying JPMS module descriptors: %w", err)
	}
	warnings = append(warnings, moduleWarnings...)
	jpmsScenario, _ := serviceload.ParseScenario(cfg.jpmsScenario)           // validated in parseFlags
	modulePathJars, _ := serviceload.ParseModulePathJars(cfg.jpmsModulePath) // validated in parseFlags
	scenarioView, scenarioWarnings := serviceload.ApplyScenario(serviceload.ScenarioConfig{Scenario: jpmsScenario, ModulePathJars: modulePathJars, AppJar: prepared.AppClassesJar}, svcRegs, &moduleInventory, h)
	warnings = append(warnings, scenarioWarnings...)
	scenarioNote := "module-info ignored"
	if jpmsScenario != serviceload.ScenarioClasspath {
		scenarioNote = fmt.Sprintf("%d provider registrations in effect", len(scenarioView.Registrations))
	}
	explicitModules, automaticModules, moduleProvides, moduleMismatches := moduleInventory.Counts()
	log("JPMS: scenario %s (%s); %d explicit module descriptors, %d jars without one, %d provides declarations, %d module-info/META-INF/services mismatches",
		moduleInventory.Scenario, scenarioNote, explicitModules, automaticModules, moduleProvides, moduleMismatches)

	// ASYNC_CALLBACKS A0: an inventory of callback-registration call
	// sites (Thread/Executor so far) — not yet a call-graph edge, and
	// not yet resolving which concrete object the callback is (see
	// callback.RegistrationSite's own doc comment).
	callbackSites := callback.FindRegistrationSites(h)
	unsupportedCallbackAPIs := 0
	for _, s := range callbackSites {
		if s.Outcome == callback.OutcomeUnsupportedAPI {
			unsupportedCallbackAPIs++
		}
	}
	log("callbacks: %d registration call sites found (%d a known API's uncataloged overload)", len(callbackSites), unsupportedCallbackAPIs)

	// ASYNC_CALLBACKS A1: resolve each cataloged site's callback
	// argument via DATA_FLOW D0-D2 — a concrete class for a direct `new`
	// allocation, or a lambda/method reference's own implementation
	// method (DYNAMIC_LINKAGE L0/L1).
	resolvedCallbacks, callbackResolveWarnings := callback.ResolveCallbacks(h, callbackSites)
	warnings = append(warnings, callbackResolveWarnings...)
	resolvedCallbackCount := 0
	for _, rc := range resolvedCallbacks {
		if rc.TargetFQCN != "" || rc.ImplOwner != "" {
			resolvedCallbackCount++
		}
	}
	log("callbacks: %d of %d cataloged sites resolved to a concrete class or lambda/method reference", resolvedCallbackCount, len(resolvedCallbacks))

	catalog, err := profile.LoadDefaultCatalog()
	if err != nil {
		return fmt.Errorf("loading default profile catalog: %w", err)
	}
	if cfg.profileFile != "" {
		if err := catalog.AddFile(cfg.profileFile); err != nil {
			return fmt.Errorf("loading -profile-file %s: %w", cfg.profileFile, err)
		}
	}
	// FRAMEWORKS F0: which frameworks' own marker classes are present —
	// a fact about the dependency closure, kept separate from which
	// profile ends up active below (Select may merge several; before
	// F0 there was no way to see the individual matches that produced
	// a merged name like "spring+jakarta").
	detections := profile.DetectAll(catalog, h)
	var detected []string
	for _, d := range detections {
		if d.Present {
			detected = append(detected, d.Name)
		}
	}
	log("frameworks detected: %s", strings.Join(detected, ", "))

	prof, err := profile.Select(catalog, cfg.profile, h)
	if err != nil {
		return fmt.Errorf("selecting profile: %w", err)
	}
	log("profile: %s", prof.Name)

	// FRAMEWORKS F1: real Spring MVC route/bean-name/@Profile values,
	// sourced from the annotation element values classfile.Annotation
	// now carries (previously only the bare presence of a mapping/
	// stereotype/@Profile annotation was visible). -active-profiles
	// unset means activeProfiles stays nil, so every @Profile condition
	// reports ConditionUnknown, never a guessed ConditionFalse.
	var activeProfiles []string
	if cfg.activeProfiles != "" {
		activeProfiles = strings.Split(cfg.activeProfiles, ",")
	}
	springRoutes := profile.Routes(h)
	springBeanNames := profile.BeanNames(h)
	springBeanMethods := profile.BeanMethods(h)
	springRepositoryBeans := profile.RepositoryBeans(h)
	springProfileConditions := profile.ProfileConditions(h, activeProfiles)
	log("frameworks: %d Spring MVC routes, %d @Component-family bean names, %d @Bean factory methods, %d Spring Data repository beans, %d @Profile conditions found (active profiles: %s)",
		len(springRoutes), len(springBeanNames), len(springBeanMethods), len(springRepositoryBeans), len(springProfileConditions), cfg.activeProfiles)

	// FRAMEWORKS F2: first DI — every @Autowired field/constructor
	// injection point, resolved against springBeanNames plus
	// springBeanMethods (F2's own "known factories" slice: a @Bean
	// method's declared return type counts as a bean exactly like a
	// stereotype class does) by type, @Qualifier, then @Primary (the
	// same order Spring's own container uses). Unresolved means no
	// candidate or an undisambiguated one, never a guess.
	injectionPoints := profile.InjectionPoints(h, springBeanNames)
	allBeans := append(append(append([]profile.BeanName{}, springBeanNames...), springBeanMethods...), springRepositoryBeans...)
	resolvedInjections := profile.ResolveInjections(h, injectionPoints, allBeans)
	resolvedInjectionCount := 0
	for _, inj := range resolvedInjections {
		if inj.Resolved != "" {
			resolvedInjectionCount++
		}
	}
	log("frameworks: %d DI injection points found, %d resolved to exactly one bean", len(injectionPoints), resolvedInjectionCount)

	// FRAMEWORKS F2: inventory of web.xml-declared lifecycle methods
	// that WebXMLAppRoots would add under -roots profile — shown
	// regardless of which -roots mode this run actually used, same as
	// every other inventory section in this report.
	webXMLRootsFound := profile.WebXMLRoots(h, servletClasses, filterClasses, listenerClasses)
	if len(webXMLRootsFound) > 0 {
		log("frameworks: %d web.xml-declared lifecycle methods found (roots under -roots profile)", len(webXMLRootsFound))
	}

	// FRAMEWORKS F2: entry point catalog — the above HTTP/web.xml
	// entry points, labeled with basis and activation condition.
	// Purely a reporting pass; no new root or call-graph edge.
	externalEntries := profile.ExternalEntries(h, activeProfiles, servletClasses, filterClasses, listenerClasses)
	log("frameworks: %d external entries cataloged (http/servlet/filter/listener)", len(externalEntries))

	// Servlet container dispatch (FRAMEWORKS F2 / JVM_GRAPH): the servlets
	// and filters the app registers — the bounded target set of the B3
	// model that expands Servlet#service/Filter#doFilter (javax./jakarta.
	// receivers the call graph never CHA-expands).
	var servletBeans []servletdispatch.Bean
	for _, b := range springBeanNames {
		servletBeans = append(servletBeans, servletdispatch.Bean{Class: b.FQCN, Detail: "stereotype"})
	}
	for _, b := range springBeanMethods {
		servletBeans = append(servletBeans, servletdispatch.Bean{Class: b.FQCN, Detail: "@Bean " + b.Name, Method: "@Bean " + b.Name})
	}
	servletRegs, servletUnresolved := servletdispatch.Find(h, servletdispatch.Inputs{WebXMLServlets: servletClasses, WebXMLFilters: filterClasses, Beans: servletBeans})
	log("servlet dispatch (%s): %d registered servlets/filters%s, %d registrations of unknown class",
		servletdispatch.Version, len(servletRegs), servletRegistrationSummary(servletRegs), len(servletUnresolved))

	// FRAMEWORKS F3's own first slice: proxy/target/interceptor
	// identity only — no new call-graph edge (see aop.FindProxyHandlers'
	// own doc comment for why: that needs REFLECTION R2, not built).
	proxyHandlers := aop.FindProxyHandlers(h)
	if len(proxyHandlers) > 0 {
		log("frameworks: %d InvocationHandler implementations found", len(proxyHandlers))
	}

	strategy, _ := callgraph.StrategyFor(cfg.callgraphMode) // validated in parseFlags
	log("building call graph (mode %s)", strategy.Mode())
	g := callgraph.BuildWith(h, strategy)
	g.SetAnalysis(cfg.roots, cfg.targetJava)
	log("call graph: %d nodes, %d edges (%d direct, %d added by CHA, %d class-init), %d unresolved dynamic call sites, %d calls resolved outside the analyzed input (informational, see report)",
		g.NodeCount(), g.EdgeCount(), g.PreCHAEdgeCount(), g.CHAEdgeCount(), g.InitEdgeCount(), g.UnresolvedDynamicCount(), g.UnresolvedOwnerCount())

	if len(prof.SyntheticEdges) > 0 {
		var synthetic []callgraph.SyntheticEdge
		for _, se := range prof.SyntheticEdges {
			synthetic = append(synthetic, callgraph.SyntheticEdge{FromSubtypesOf: se.FromSubtypesOf, To: se.To})
		}
		added := callgraph.ApplySyntheticEdges(g, h, synthetic)
		log("applied %d synthetic edges from profile %q (e.g. Spring Data repository proxies)", added, prof.Name)
	}

	// SERVICE_LOADING S2: connect a load site to its registered providers
	// only at a linked lazy-consumption operation (Iterator.hasNext/next or
	// Provider.get), plus the provider's own <clinit>, if any.
	svcMaterializations := serviceload.FindMaterializations(h, svcLoadSites, scenarioView.Registrations)
	serviceModelEvaluator := serviceload.NewModelEvaluator(h, svcLoadSites, svcMaterializations, svcConsumptionSites)
	serviceModelEvaluator.SetScenario(scenarioView)

	// ASYNC_CALLBACKS A2: connect each resolved callback identity (A1)
	// to the class that actually declares its run()/call() method — a
	// "may run" edge (BasisCallback), never claiming the thread/task
	// actually executes.
	callbackEdgeResults := callback.FindCallbackEdges(h, resolvedCallbacks)
	callbackEdgeTargets := callback.EdgeTargets(callbackEdgeResults)
	var callbackGraphEdges []callgraph.CallbackEdge
	for _, e := range callbackEdgeTargets {
		callbackGraphEdges = append(callbackGraphEdges, callgraph.CallbackEdge{
			CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc,
			PC: e.PC, TargetFQCN: e.TargetFQCN, TargetMethod: e.TargetMethod, TargetMethodDesc: e.TargetMethodDesc,
		})
	}
	callbackEdgesAdded := callgraph.ApplyCallbackEdges(g, callbackGraphEdges)
	log("callbacks: %d resolved identities modeled as call-graph targets (%d edges added)", len(callbackEdgeTargets), callbackEdgesAdded)

	// DYNAMIC_LINKAGE L2: find every MethodHandle#invoke/invokeExact
	// call site, read its receiver back via DATA_FLOW D2's
	// Facts.ReceiverAt, and connect it to every overload of the
	// resolved target method name — a "may run" edge, like
	// ASYNC_CALLBACKS A2's own BasisCallback.
	handleInvokeSites := methodhandle.FindInvokeSites(h)
	resolvedHandles, handleResolveWarnings := methodhandle.ResolveInvokes(h, handleInvokeSites)
	warnings = append(warnings, handleResolveWarnings...)
	handleEdgeResults := methodhandle.FindHandleEdges(h, resolvedHandles)
	handleEdgeTargets := methodhandle.EdgeTargets(handleEdgeResults)
	var handleGraphEdges []callgraph.HandleEdge
	for _, e := range handleEdgeTargets {
		handleGraphEdges = append(handleGraphEdges, callgraph.HandleEdge{
			CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc,
			PC: e.PC, TargetFQCN: e.TargetFQCN, TargetMethod: e.TargetMethod,
		})
	}
	handleEdgesAdded := callgraph.ApplyMethodHandleEdges(g, handleGraphEdges)
	log("method handles: %d invoke/invokeExact call sites found, %d resolved identities modeled as call-graph targets (%d edges added)",
		len(handleInvokeSites), len(handleEdgeTargets), handleEdgesAdded)

	// DYNAMIC_LINKAGE L1 (SAM-call materialization): a lambda/method
	// reference created in one method and run through its functional
	// interface in another — connect the calling site to the
	// implementation, following the value with DATA_FLOW (callers,
	// captured values, fields, helper returns). Unresolved receivers add
	// no edge and are counted as gaps.
	samStart := time.Now()
	samSites := samcall.Find(h)
	samResolver := samcall.NewResolver(h)
	samResults := samResolver.Resolve(samSites)
	samResolver.Close()
	var samGraphEdges []callgraph.SAMEdge
	for _, e := range samcall.Edges(samResults) {
		var creators []callgraph.MethodRef
		for _, c := range e.Creators {
			creators = append(creators, callgraph.MethodRef{FQCN: c.Owner, Name: c.Name, Desc: c.Desc})
		}
		samGraphEdges = append(samGraphEdges, callgraph.SAMEdge{CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc,
			PC: e.PC, Kind: e.CallKind, TargetFQCN: e.TargetFQCN, TargetName: e.TargetName, TargetDesc: e.TargetDesc, Virtual: e.Virtual, Creators: creators})
	}
	// The edges wait for the roots (samGate.Apply, with the library
	// models): each needs a reachable method creating the value.
	samGate := callgraph.NewSAMGate(g, h, samGraphEdges)
	samStats := samcall.Summarize(samResults)
	log("SAM calls (DYNAMIC_LINKAGE L1, %s): %s; %d edge candidates, in %s", samcall.Version, samSiteSummary(samStats), len(samGraphEdges), time.Since(samStart).Round(time.Millisecond))

	// REFLECTION R2: find every java.lang.reflect.Method#invoke call
	// site, read its Method receiver back via REFLECTION R1's own
	// KindReflectiveMethod, and connect it to every overload of the
	// resolved target method name on its real dispatch target — a "may
	// run" edge, like DYNAMIC_LINKAGE L2's own BasisMethodHandle.
	reflectInvokeSites := reflectinvoke.FindInvokeSites(h)
	resolvedReflectInvokes, reflectResolveWarnings := reflectinvoke.ResolveInvokes(h, reflectInvokeSites)
	warnings = append(warnings, reflectResolveWarnings...)
	reflectEdgeResults := reflectinvoke.FindInvokeEdges(h, resolvedReflectInvokes)
	reflectEdgeTargets := reflectinvoke.EdgeTargets(reflectEdgeResults)
	var reflectGraphEdges []callgraph.ReflectInvokeEdge
	for _, e := range reflectEdgeTargets {
		reflectGraphEdges = append(reflectGraphEdges, callgraph.ReflectInvokeEdge{
			CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc,
			PC: e.PC, TargetFQCN: e.TargetFQCN, TargetMethod: e.TargetMethod,
		})
	}
	reflectEdgesAdded := callgraph.ApplyReflectInvokeEdges(g, reflectGraphEdges)
	log("reflection: %d Method#invoke call sites found, %d resolved identities modeled as call-graph targets (%d edges added)",
		len(reflectInvokeSites), len(reflectEdgeTargets), reflectEdgesAdded)

	// REFLECTION R2's own second half: the same reasoning for
	// java.lang.reflect.Constructor#newInstance, resolving its
	// Constructor receiver via REFLECTION R2's own
	// KindReflectiveConstructor and connecting it to the exact "<init>"
	// its resolved parameter types declared — never every overload,
	// since a constructor is never inherited.
	reflectNewInstanceSites := reflectinvoke.FindNewInstanceSites(h)
	resolvedReflectNewInstances, reflectNewInstanceResolveWarnings := reflectinvoke.ResolveNewInstances(h, reflectNewInstanceSites)
	warnings = append(warnings, reflectNewInstanceResolveWarnings...)
	reflectNewInstanceEdgeResults := reflectinvoke.FindNewInstanceEdges(h, resolvedReflectNewInstances)
	reflectNewInstanceEdgeTargets := reflectinvoke.NewInstanceEdgeTargets(reflectNewInstanceEdgeResults)
	var reflectNewInstanceGraphEdges []callgraph.ReflectNewInstanceEdge
	for _, e := range reflectNewInstanceEdgeTargets {
		reflectNewInstanceGraphEdges = append(reflectNewInstanceGraphEdges, callgraph.ReflectNewInstanceEdge{
			CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc,
			PC: e.PC, TargetFQCN: e.TargetFQCN, TargetDesc: e.TargetDesc,
		})
	}
	reflectNewInstanceEdgesAdded := callgraph.ApplyReflectNewInstanceEdges(g, reflectNewInstanceGraphEdges)
	log("reflection: %d Constructor#newInstance call sites found, %d resolved identities modeled as call-graph targets (%d edges added)",
		len(reflectNewInstanceSites), len(reflectNewInstanceEdgeTargets), reflectNewInstanceEdgesAdded)

	// DESERIALIZATION Z1: every Java serialization entry
	// (ObjectInputStream#readObject/readUnshared, ObjectInput#readObject)
	// gets a "may run" edge to every callback serialization may run on
	// a concrete serializable class of the input (readObject,
	// readObjectNoData, readResolve, readExternal, the constructor it
	// calls). XStream entries are listed, not connected (entry-only).
	deserSites := deserialize.FindSites(h)
	warnings = append(warnings, deserialize.ResolvePolicies(h, deserSites)...) // DESERIALIZATION Z2
	deserJava, deserSummary := deserialize.ClassTargets(h, deserialize.FormatJava)
	deserXStream, _ := deserialize.ClassTargets(h, deserialize.FormatXStream)
	deserTargets, _ := deserialize.Candidates(h)
	var deserGraphEdges []callgraph.DeserializationEdge
	for _, e := range deserialize.Connect(deserSites, map[string]map[string][]deserialize.Target{deserialize.FormatJava: deserJava, deserialize.FormatXStream: deserXStream}) {
		deserGraphEdges = append(deserGraphEdges, callgraph.DeserializationEdge{
			CallerFQCN: e.CallerFQCN, CallerName: e.CallerName, CallerDesc: e.CallerDesc, PC: e.PC,
			TargetFQCN: e.TargetFQCN, TargetName: e.TargetName, TargetDesc: e.TargetDesc,
		})
	}
	deserEdgesAdded := callgraph.ApplyDeserializationEdges(g, deserGraphEdges)
	deserStates := map[string]int{}
	for _, s := range deserSites {
		deserStates[s.Filter]++
	}
	log("deserialization (DESERIALIZATION Z1/Z2): %d entry sites (policy: %s); %d serializable candidate classes (%d Externalizable; %d enums, %d records excluded; %d of unknown serializability), %d Java serialization callback targets, %d edges added",
		len(deserSites), report.GapCounts(deserStates), deserSummary.Classes, deserSummary.Externalizable, deserSummary.Enums, deserSummary.Records, deserSummary.Unknown, len(deserTargets), deserEdgesAdded)

	cache, err := httpcache.New(cfg.cacheDir, cfg.offline)
	if err != nil {
		return fmt.Errorf("opening cache dir %s: %w", cfg.cacheDir, err)
	}
	sr.attach(cache, "main")
	cache.SetPolicy(httpcache.Policy{TTL: cfg.dataTTL, MaxAge: cfg.maxDataAge, Refresh: cfg.refresh, AllowStale: cfg.allowStale})

	embeddedDecls, err := component.LoadDeclarations(cfg.embeddedComps)
	if err != nil {
		return err
	}
	sbom, err := component.LoadSBOM(cfg.sbom)
	if err != nil {
		return err
	}
	components.ApplySBOM(sbom)
	if sb := components.SBOM; sb != nil {
		log("sbom (CVE:F6b): %s, CycloneDX %s: %d components, %d tied to an analyzed jar by SHA-256, %d nested candidates, %d Maven components not tied to the input (not used)",
			sb.Path, sb.SpecVersion, sb.Components, sb.Linked, sb.Candidates, sb.Unlinked)
		for _, issue := range sb.Issues {
			warnings = append(warnings, "-sbom: "+issue)
		}
	}
	// CVE:F6a: embedded components, confirmed by their classes against
	// the Maven Central reference jar of the declared version.
	if err := components.ResolveEmbedded(ctx, func(ctx context.Context, url string) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		return httpcache.Fetch(http.DefaultClient, cache, httpcache.Key("GET", url), httpcache.KindForURL(url), req)
	}, cfg.centralURL, h, embeddedDecls); err != nil {
		return fmt.Errorf("resolving embedded components: %w", err)
	}
	embConfirmed, embUnconfirmed, embModules := components.EmbeddedCounts()
	embRelocated := 0
	for _, e := range components.Embedded {
		if (e.Status == component.EmbeddedConfirmed || e.Status == component.EmbeddedVersionCandidates) && len(e.Relocations) > 0 {
			embRelocated++
		}
	}
	embCandidateSets, embFromClasses := 0, 0
	for _, e := range components.Embedded {
		switch {
		case e.Status == component.EmbeddedVersionCandidates:
			embCandidateSets++
		case e.Status == component.EmbeddedConfirmed && e.Version != e.GAV.Version:
			embFromClasses++
		}
	}
	log("embedded components (CVE:F6a/F6b/F7a): %d confirmed by their classes (%d relocated, %d with the version read from the classes, %d with a set of candidate versions), %d unconfirmed (not looked up), %d modules of their host",
		embConfirmed, embRelocated, embFromClasses, embCandidateSets, embUnconfirmed, embModules)
	for _, n := range components.BuildConfigNotes {
		log("embedded components: %s", n)
	}
	for _, issue := range components.DeclarationIssues {
		warnings = append(warnings, "-embedded-components: "+issue)
	}

	// DESERIALIZATION Z3a: catalog gadget chains against every entry
	// site, with the identity of the classes from the component
	// inventory (after F6a/F7a, so embedded and relocated copies have
	// their owner). Report-only: no edge, no verdict changes.
	// DATA_FLOW D3: dependency summaries of the methods with entries (and
	// what they call), so each entry states what the data it reads
	// depends on in its method. Report-only.
	flowStart := time.Now()
	flowOpts := flowsum.DefaultOptions
	flowRoots := deserialize.SummaryRoots(deserSites)
	// DATA_FLOW D4 (-taint): the entries' methods and their application
	// callers, within the search depth, are summarized in the same run;
	// the source accessors stay opaque (their results are sources).
	var taintAnalyzer *taint.Analyzer
	if cfg.taint {
		taintAnalyzer = taint.New(h, taint.DefaultOptions)
		flowRoots = deserialize.TaintRoots(deserSites, taintAnalyzer)
		flowOpts.Opaque = taint.Opaque
	}
	flowEngine := flowsum.New(h, flowOpts)
	flowStats := flowEngine.Summarize(flowRoots)
	deserialize.AttachArgumentFlow(deserSites, flowEngine)
	if taintAnalyzer != nil {
		taintStart := time.Now()
		deserialize.AttachTaint(deserSites, taintAnalyzer, flowEngine)
		ts := taintAnalyzer.Stats
		log("taint (DATA_FLOW D4, %s, opt-in): %d deserialization entries assessed: %s; %d searches, %d inputs visited, %d depth-limit and %d search-limit stops, in %s (summaries above)",
			taint.CatalogVersion, len(deserSites), report.GapCounts(deserialize.TaintStateCounts(deserSites)),
			ts.Sinks, ts.Nodes, ts.DepthLimits, ts.NodeLimits, time.Since(taintStart).Round(time.Microsecond))
	}
	flowEngine.Close()
	withParams := 0
	for _, s := range deserSites {
		for _, d := range s.DataFromRaw {
			if strings.HasPrefix(d, "param:") {
				withParams++
				break
			}
		}
	}
	log("dependency summaries (DATA_FLOW D3, %s): %d methods, %d evaluations (converged %v, %d cut by the budget, %d incomplete, %d calls answered conservatively) in %s; %d of %d deserialization entries read data depending on a parameter of their method",
		flowsum.SummaryVersion, flowStats.Methods, flowStats.Evaluations, flowStats.Converged, flowStats.BudgetCut, flowStats.Incomplete, flowStats.Widened,
		time.Since(flowStart).Round(time.Millisecond), withParams, len(deserSites))

	deserChains := deserialize.EvaluateChains(chainInputs(h, components, cfg.targetJava), deserSites)
	chainFlow := "data flow into the entries not analyzed (no -taint)"
	if cfg.taint {
		chainFlow = "flow per chain: none with an entry site"
		if fc := deserialize.ChainFlowCounts(deserChains); len(fc) > 0 {
			chainFlow = "flow per chain: " + report.GapCounts(fc)
		}
	}
	log("deserialization chains (DESERIALIZATION Z3, catalog %s, %d chains): %s; %s",
		deserialize.ChainCatalogVersion, len(deserialize.ChainCatalog), report.GapCounts(deserialize.ChainStateCounts(deserChains)), chainFlow)

	var gavs []input.GAV
	seenGAV := map[input.GAV]bool{}
	for _, dj := range depJars {
		if dj.GAV != nil && !seenGAV[*dj.GAV] {
			seenGAV[*dj.GAV] = true
			gavs = append(gavs, *dj.GAV)
		}
	}
	for _, g := range components.ConfirmedEmbeddedGAVs() {
		if !seenGAV[g] {
			seenGAV[g] = true
			gavs = append(gavs, g)
		}
	}
	log("querying OSV for %d resolved dependencies", len(gavs))
	resolution, err := cves.NewClient(cache).ResolveAll(ctx, gavs)
	if err != nil {
		return fmt.Errorf("resolving CVEs: %w", err)
	}
	cveList := resolution.CVEs
	log("found %d CVEs (%d withdrawn advisories not analyzed)", len(cveList), len(resolution.Withdrawn))

	manual, err := vulnmap.LoadManual(cfg.vulnMethods)
	if err != nil {
		return fmt.Errorf("loading vuln_methods.yaml: %w", err)
	}
	resolver := vulnmap.NewResolver(manual, cache)
	resolver.GithubToken = cfg.githubToken
	resolver.Bytecode = declaredOverloads(h)
	if resolver.ManualMeta, err = vulnmap.LoadManualMeta(cfg.vulnMethods); err != nil {
		return fmt.Errorf("loading vuln_methods mapping metadata: %w", err)
	}
	var llmSrc *vulnmap.LLMSource
	if cfg.llm {
		llmCache, err := httpcache.New(filepath.Join(cfg.cacheDir, "llm"), cfg.offline)
		if err != nil {
			return fmt.Errorf("opening LLM cache dir: %w", err)
		}
		sr.attach(llmCache, "llm")
		llmSrc = vulnmap.NewLLMSource(cfg.llmConfig, llmCache, nil)
		llmSrc.ReaskChanged = cfg.llmReaskChanged
		resolver.LLM = llmSrc
		resolver.MethodExists = methodExists(h)
		log("LLM source enabled: model %s", cfg.llmConfig.Model)
	}
	vm, err := resolver.ResolveAll(ctx, cveList)
	if err != nil {
		return fmt.Errorf("mapping CVEs to vulnerable methods: %w", err)
	}
	bySource := map[string]int{}
	mapped, newSideOnly := 0, 0
	for _, e := range vm {
		if !e.Mapped() {
			bySource["unavailable"]++
			continue
		}
		mapped++
		bySource[e.Source]++
		if e.Source == vulnmap.SourceDiff && e.NewSideOnly() {
			newSideOnly++
		}
	}
	log("mapped %d/%d CVEs to vulnerable methods (manual=%d diff=%d llm=%d; diff read on the post-fix side only=%d; fix commits unreadable=%d)",
		mapped, len(cveList), bySource[vulnmap.SourceManual], bySource[vulnmap.SourceDiff], bySource[vulnmap.SourceLLM], newSideOnly, bySource["unavailable"])
	log("fix commits vs the analyzed release (CVE:F5c): %s", releaseLineSummary(vm))
	if llmSrc != nil {
		st := llmSrc.Stats()
		log("llm: %d mapped, %d answered with no usable method, %d skipped (request failed), %d invalid reply, %d methods dropped as not in bytecode, %d previous answers reused for changed advisories",
			st.Mapped, st.Empty, st.Skipped, st.Invalid, st.Dropped, st.Reused)
		warnings = append(warnings, llmSrc.Warnings()...)
	}

	log("analyzing reachability")
	framework := reach.NewFrameworkMatcher(prof.ResolvePatterns(h))
	var roots []uint32
	webXMLRootCount := 0
	if cfg.roots == "profile" {
		roots = prof.AppRoots(g, h)
		webRoots := profile.WebXMLAppRoots(g, h, servletClasses, filterClasses, listenerClasses)
		webXMLRootCount = len(webRoots)
		roots = append(roots, webRoots...)
	} else {
		roots = reach.AppRoots(g, h)
	}
	var modelGapReasons []string
	registry := models.NewRegistry()
	serviceload.RegisterModel(registry)
	serviceload.RegisterDispatchModels(registry, svcConsumptionSites)
	servletdispatch.Register(registry)
	// Calls on a JDK/javax service-provider interface (a receiver CHA never
	// expands) go to the providers registered for it in META-INF/services.
	var spiRegs []spidispatch.Registration
	for _, r := range svcRegs {
		spiRegs = append(spiRegs, spidispatch.Registration{Service: r.Interface, Provider: r.ProviderClass, Source: r.SourceFile})
	}
	spiCatalog := spidispatch.Find(h, spiRegs, callgraph.DefaultSkip)
	spidispatch.Register(registry, spiCatalog)
	spiServices := 0
	for range spiCatalog.Providers {
		spiServices++
	}
	log("service-provider dispatch (%s %s): %d skipped service interfaces with registered providers, %d call signatures in the input (%d registrations of services CHA already expands, %d providers not in the input)",
		spidispatch.ModelID, spidispatch.Version, spiServices, len(spiCatalog.Signatures), spiCatalog.NotSkipped, spiCatalog.ProviderMissing)
	modelEvaluator := models.Router{
		ByID: map[string]models.Evaluator{
			servletdispatch.ModelID: servletdispatch.NewEvaluator(h, servletRegs, servletUnresolved),
			spidispatch.ModelID:     spidispatch.NewEvaluator(h, spiCatalog),
		},
		Default: serviceModelEvaluator,
	}
	modelBudget := len(svcLoadSites) + 1 + servletdispatch.MaxChain + 2
	// Servlet request/response/chain objects (servlet-container:dispatch
	// s2): calls on them reach the bodies of the implementations reachable
	// code creates — gated like the SAM edges.
	objStart := time.Now()
	objEdges := servletdispatch.ObjectEdges(h)
	objGate := callgraph.NewCreatorGate(g, h, objEdges, callgraph.BasisModel)
	objTime := time.Since(objStart)
	applyGates := func() (callgraph.SAMStats, callgraph.SAMStats) {
		var sam, obj callgraph.SAMStats
		for {
			before := sam.Added + obj.Added
			sam = samGate.Apply(g, roots)
			obj = objGate.Apply(g, roots)
			if sam.Added+obj.Added == before {
				return sam, obj
			}
		}
	}
	applyGates() // DYNAMIC_LINKAGE L1 and the servlet objects, before the models see the graph
	modelResult, err := analysis.Run(h, g, registry, modelEvaluator, roots, analysis.Options{MaxIterations: modelBudget})
	if err != nil {
		return fmt.Errorf("coordinating library models: %w", err)
	}
	log("models: %d facts, %d edges added, %d iterations (incomplete=%t)", len(modelResult.Facts), modelResult.AddedEdges, modelResult.Iterations, modelResult.Incomplete)
	// Model edges can make more creators reachable.
	samApplied, objApplied := applyGates()
	log("servlet objects (%s %s): %d candidates, %d edges added, %d already in the graph, %d left out (no method creating the class is reachable from the roots), in %s",
		servletdispatch.ModelID, servletdispatch.Version, len(objEdges), objApplied.Added, objApplied.Covered, objApplied.Gated, objTime.Round(time.Millisecond))
	samLine := samSummary(samStats, samApplied)
	log("SAM-call edges (DYNAMIC_LINKAGE L1): %d added, %d already in the graph, %d left out (no method creating the value is reachable from the roots), %d without a node",
		samApplied.Added, samApplied.Covered, samApplied.Gated, samApplied.Missing)
	for _, reason := range modelResult.Reasons {
		warnings = append(warnings, "model analysis: "+reason)
		modelGapReasons = append(modelGapReasons, reason)
	}
	// GRAPH_PRECISION P1: the catalog of created types, over the final
	// graph (every model edge added) and the models' facts. Report-only
	// in CHA mode; in RTA mode (P2) it feeds the fixed point below.
	allocIn := allocInputs(h, g, cfg, roots, springBeanNames, springBeanMethods, springRepositoryBeans,
		deserJava, deserXStream, deserSites, svcLoadSites, reflectNewInstanceSites, reflectNewInstanceEdgeTargets)
	for _, r := range servletRegs {
		allocIn.Container = append(allocIn.Container, alloc.ContainerBean{Class: r.Class, Detail: string(r.Kind) + " registration (" + r.Basis + ")"})
	}
	allocCatalog := alloc.Build(allocIn)
	allocSummary := alloc.Summarize(allocCatalog, g)
	log("created types (GRAPH_PRECISION P1, %s): %d classes (%s; %d only through a model), %d lambda sites, gaps: %s; %d of %d CHA edges lead to a class not in the catalog (informational)",
		alloc.CatalogVersion, allocSummary.Types, report.GapCounts(allocSummary.ByOrigin), allocSummary.OnlyModeled, allocSummary.Lambdas,
		report.GapCounts(allocSummary.Gaps), allocSummary.CHATargetsNotCreated, allocSummary.CHATargets)

	// GRAPH_PRECISION P2 (opt-in -callgraph-mode rta): narrow the CHA
	// edges to the classes created by reachable code, after every model
	// edge is in and before reachability is analyzed.
	var rtaResult *callgraph.RTAResult
	if strategy.Mode() == callgraph.ModeRTA {
		rtaStart := time.Now()
		providers := []string{}
		for _, r := range svcRegs {
			providers = append(providers, r.ProviderClass)
		}
		sort.Strings(providers)
		res := callgraph.ApplyRTA(g, h, rtaInputs(h, g, allocCatalog, roots, providers))
		rtaResult = &res
		log("RTA (GRAPH_PRECISION P2, %s): %d reachable methods, %d live classes (%d only by the unknown-allocation fallback); %d of %d CHA edges removed, %d kept only by the fallback; active unknown-allocation gaps: %s; in %s",
			res.Version, res.ReachableMethods, res.LiveClasses, res.FallbackClasses, res.CHAEdgesRemoved, res.CHAEdgesBefore, res.CHAEdgesFallback,
			report.GapCounts(res.ActiveGaps), time.Since(rtaStart).Round(time.Millisecond))
	}

	var engine *reach.Engine
	if cfg.roots == "profile" {
		log("roots: profile %q (%d entry-point methods, %d from web.xml)", prof.Name, len(roots), webXMLRootCount)
		engine = reach.NewWithRoots(g, h, framework, roots)
	} else {
		engine = reach.NewWithRoots(g, h, framework, roots)
	}
	engine.SetComponents(components) // CVE:F1b: findings per affected component instance
	gapSites := collectGapSites(h, gapInputs{
		reflectInvokes: resolvedReflectInvokes, newInstances: resolvedReflectNewInstances, handles: resolvedHandles,
		callbackSites: callbackSites, callbacks: resolvedCallbacks,
		loadSites: svcLoadSites, consumptions: svcConsumptionSites, registrations: svcRegs,
		deserialization: deserSites,
	})
	if rtaResult != nil {
		gapSites = append(gapSites, rtaGapSites(allocCatalog)...)
	}
	engine.SetGaps(gapSites) // CVE:F1b: model-gap rule, docs/RESULTS_CONTRACT.MD §4
	reachableGaps, unboundedGaps := engine.ReachableGaps()
	log("model gaps: %d unresolved dynamic sites, %d reachable from the roots (%d with no bounded candidate set)", len(gapSites), reachableGaps, unboundedGaps)
	webAssets := map[string]bool{}
	for _, dj := range depJars {
		if dj.WebAssetsOnly && dj.GAV != nil {
			webAssets[dj.GAV.Group+":"+dj.GAV.Artifact] = true
		}
	}
	engine.SetWebAssetPackages(webAssets)
	results := engine.Analyze(cveList, vm)

	// RUNTIME_NATIVE N1: the exact-entry tracing selection for a recorded
	// run (-runtime-filter-out; the file is written with the other outputs).
	var runtimeFilterSel RuntimeFilter
	if cfg.runtimeFilter != "" {
		runtimeFilterSel = runtimeFilter(results, h, cfg.runtimeStatus)
		if len(runtimeFilterSel.Filters) == 0 {
			warnings = append(warnings, "-runtime-filter-out: no mapped vulnerable method of the selected verdicts is in the analyzed input; the settings file traces nothing")
		}
	}

	// RUNTIME_NATIVE N1: correlate the imported trace with the final graph.
	// Evidence next to the results; no edge or verdict changes.
	var runtimeResult *runtimeevents.Result
	var runtimeObs map[string]map[string][]runtimeevents.MethodObservation
	var runtimeTracedBy map[string]map[string]bool
	if runtimeTrace != nil {
		lookup, closeJars := classLookup(h)
		rt := runtimeevents.Correlate(runtimeTrace, runtimeevents.Inputs{G: g, Lookup: lookup, Classify: gapClassifier(h, g.Config().Mode)})
		closeJars()
		runtimeResult = &rt
		runtimeObs = runtimeObserved(rt, results, h, components)
		runtimeTracedBy = runtimeTraced(rt, results, h, components)
		log("runtime trace (RUNTIME_NATIVE N1, %s): run %s, %s, sampling %s, complete=%t; classes %s; loaded %s; executed (usable) %s, %d not usable; calls %s; %d advisories with an observed vulnerable method",
			cfg.runtimeTrace, rt.RunID, rt.Recorder, rt.Sampling, rt.Complete, report.GapCounts(rt.ClassCounts), report.GapCounts(rt.LoadedCounts),
			report.GapCounts(rt.ExecutedCounts), rt.ExecutedOther, report.GapCounts(rt.CallCounts), len(runtimeObs))
	}

	counts := map[reach.Status]int{}
	for _, r := range results {
		counts[r.Status]++
	}
	log("REACHABLE=%d FRAMEWORK=%d NOT_REACHABLE=%d UNKNOWN=%d NOT_APPLICABLE=%d",
		counts[reach.StatusReachable], counts[reach.StatusFramework], counts[reach.StatusNotReachable], counts[reach.StatusUnknown],
		counts[reach.StatusNotApplicable])

	var cveByID map[string]cves.CVE
	if cfg.exploitCheck || cfg.frameworkCheck {
		cveByID = make(map[string]cves.CVE, len(cveList))
		for _, c := range cveList {
			cveByID[c.ID] = c
		}
	}

	var exploitStats exploit.Stats
	if cfg.exploitCheck {
		log("exploit-check: asking %s about %d REACHABLE CVEs", cfg.llmConfig.Model, counts[reach.StatusReachable])
		exploitCache, err := httpcache.New(filepath.Join(cfg.cacheDir, "exploit"), cfg.offline)
		if err != nil {
			return fmt.Errorf("opening exploit-check cache dir: %w", err)
		}
		sr.attach(exploitCache, "exploit")
		checker := exploit.NewChecker(cfg.llmConfig, exploitCache, nil)
		findings, err := checker.CheckAll(ctx, results, cveByID)
		if err != nil {
			return fmt.Errorf("exploit-check: %w", err)
		}
		for i := range results {
			if f, ok := findings[results[i].CVE]; ok {
				results[i].Exploitability = string(f.Verdict)
				results[i].ExploitReason = f.Reason
				results[i].ExploitModel = f.Model
			}
		}
		exploitStats = checker.Stats()
		log("exploit-check: %d checked, %d skipped (request failed), %d invalid reply",
			exploitStats.Checked, exploitStats.Skipped, exploitStats.Invalid)
		warnings = append(warnings, checker.Warnings()...)
	}

	var frameworkCheckStats exploit.Stats
	if cfg.frameworkCheck {
		log("framework-check: asking %s about %d FRAMEWORK CVEs", cfg.llmConfig.Model, counts[reach.StatusFramework])
		frameworkCache, err := httpcache.New(filepath.Join(cfg.cacheDir, "framework-check"), cfg.offline)
		if err != nil {
			return fmt.Errorf("opening framework-check cache dir: %w", err)
		}
		sr.attach(frameworkCache, "framework-check")
		checker := exploit.NewFrameworkChecker(cfg.llmConfig, frameworkCache, nil)
		findings, err := checker.CheckAll(ctx, results, cveByID)
		if err != nil {
			return fmt.Errorf("framework-check: %w", err)
		}
		for i := range results {
			if f, ok := findings[results[i].CVE]; ok {
				results[i].PathPlausibility = string(f.Verdict)
				results[i].PathPlausibilityReason = f.Reason
				results[i].PathPlausibilityModel = f.Model
			}
		}
		frameworkCheckStats = checker.Stats()
		log("framework-check: %d checked, %d skipped (request failed), %d invalid reply",
			frameworkCheckStats.Checked, frameworkCheckStats.Skipped, frameworkCheckStats.Invalid)
		warnings = append(warnings, checker.Warnings()...)
	}

	// Every network fetch of the run is done by now (OSV, fix diffs).
	freshness := cache.Freshness()
	freshnessState, freshnessLines, freshnessWarnings := report.DescribeFreshness(freshness, cfg.dataTTL)
	if sr.replay != nil {
		recorded := sr.replay.DataFreshness
		freshnessState = report.FreshnessReplayed
		freshnessLines = []string{fmt.Sprintf("snapshot data set %s, created %s, recorded as %q", sr.replay.DataSetID, sr.replay.CreatedAt.Format(time.RFC3339), recorded)}
		freshnessWarnings = nil
		for _, d := range snapshot.ConfigDiff(sr.replay.Config, sr.config) {
			warnings = append(warnings, "snapshot: configuration differs from the recorded run: "+d)
		}
		if !sr.replay.Tool.Reproducible {
			warnings = append(warnings, "snapshot: recorded by a locally modified build — the data replays, the analyzer source is identified only by the binary hash")
		}
	} else {
		warnings = append(warnings, freshnessWarnings...)
	}
	log("data freshness: %s (%s)", freshnessState, strings.Join(freshnessLines, "; "))

	if sr.replayer != nil {
		if f := sr.replayer.Failures(); len(f) > 0 {
			return fmt.Errorf("the snapshot could not answer %d lookups, so the replay is not a reproduction: %s", len(f), strings.Join(f, "; "))
		}
	}
	// CVE:F3: derivation keys, the manifest of this run, and the
	// comparison with an earlier snapshot.
	ders := derivations(cveList, vm, results, components, cfg.llmConfig.Model)
	var observations []httpcache.Observation
	if sr.recorder != nil {
		if observations, err = sr.recorder.Observations(); err != nil {
			return fmt.Errorf("writing snapshot objects: %w", err)
		}
	} else if sr.replay != nil {
		observations = sr.replay.Observations
	}
	current := sr.manifest(snapInputs, observations, ders, results, freshnessState)
	var changes []snapshot.Change
	if sr.compare != nil {
		changes = snapshot.Compare(*sr.compare, current)
		log("snapshot compare: %d advisories with invalidated derivations or changed verdicts since %s", len(changes), sr.compare.DataSetID)
	}
	profileLabel := prof.Name
	if cfg.roots == "profile" {
		profileLabel += " (roots: profile)"
	}
	// GRAPH_PRECISION P0: the final graph (every model edge added) with
	// the findings' paths, for -graph-out and -graph-compare.
	var graphFindings []callgraph.FindingPath
	for _, r := range results {
		graphFindings = append(graphFindings, callgraph.FindingPath{ID: r.CVE, Status: string(r.Status), Path: r.Path})
	}
	graphNow := callgraph.Snapshot(g, graphFindings)
	graphNow.Types = allocExport(allocCatalog)
	var graphCmp *callgraph.Comparison
	if graphBefore != nil {
		c := callgraph.Compare(*graphBefore, graphNow)
		graphCmp = &c
		log("call graph comparison with %s: %d → %d edges, %d added, %d removed, %d with other call-site evidence; %d findings changed; settings: %s",
			cfg.graphCompare, c.EdgesBefore, c.EdgesAfter, len(c.Added), len(c.Removed), c.EvidenceChanged, len(c.Findings), graphSettings(c.ConfigDifferences))
	}
	header := report.Header{
		App:                cfg.app,
		GeneratedAt:        time.Now(),
		AriadneVersion:     version(),
		DataFreshness:      freshnessState,
		FreshnessLines:     freshnessLines,
		Profile:            profileLabel,
		FrameworksDetected: strings.Join(detected, ", "),
		MaxBytecodeVer:     maxBytecode,
		JPMSScenario:       string(jpmsScenario),
		GraphConfig:        g.Config(),
		GraphNodes:         g.NodeCount(),
		GraphEdges:         g.EdgeCount(),
		PreCHAEdges:        g.PreCHAEdgeCount(),
		CHAEdges:           g.CHAEdgeCount(),
		InitEdges:          g.InitEdgeCount(),
		ModelEdges:         g.ModelEdgeCount(),
		SAMEdges:           g.SAMEdgeCount(),
		SAMLine:            samLine,
		UnresolvedDyn:      g.UnresolvedDynamicCount(),
		UnresolvedOwner:    g.UnresolvedOwnerCount(),
		GapSites:           len(gapSites),
		GapsReachable:      reachableGaps,
		GapsUnbounded:      unboundedGaps,
		DepsTotal:          len(depJars),
		DepsIdentified:     identified,
		UnidentifiedJar:    unidentified,
	}
	if llmSrc != nil {
		st := llmSrc.Stats()
		header.LLMModel = cfg.llmConfig.Model
		header.LLMMapped = st.Mapped
		header.LLMSkipped = st.Skipped + st.Invalid
	}
	if cfg.exploitCheck {
		header.ExploitModel = cfg.llmConfig.Model
		header.ExploitChecked = exploitStats.Checked
		header.ExploitSkipped = exploitStats.Skipped + exploitStats.Invalid
	}
	if cfg.frameworkCheck {
		header.FrameworkCheckModel = cfg.llmConfig.Model
		header.FrameworkChecked = frameworkCheckStats.Checked
		header.FrameworkCheckSkipped = frameworkCheckStats.Skipped + frameworkCheckStats.Invalid
	}
	data := report.Build(header, cveList, results, warnings, svcRegs, svcLoadSites, svcMaterializations, callbackSites, resolvedCallbacks, callbackEdgeResults, resolvedHandles, handleEdgeResults, springRoutes, springBeanNames, springProfileConditions, cfg.activeProfiles, webXMLInput, springXMLBeans, resolvedInjections, webXMLRootsFound, externalEntries, proxyHandlers, resolvedReflectInvokes, reflectEdgeResults, resolvedReflectNewInstances, reflectNewInstanceEdgeResults, svcConsumptionSites)
	data.Modules = moduleInventory
	data.Components = components
	data.Deserialization = &report.Deserialization{Models: deserialize.Catalog, Sites: deserSites, Summary: deserSummary, Targets: len(deserTargets), EdgesAdded: deserEdgesAdded,
		Chains: deserChains, ChainCatalogVersion: deserialize.ChainCatalogVersion,
		ModelCatalogVersion: deserialize.ModelCatalogVersion, TaintCatalog: taintCatalog(cfg.taint)}
	data.GraphComparison, data.GraphCompareFile = graphCmp, filepath.Base(cfg.graphCompare)
	data.CreatedTypes, data.CreatedTypeGaps, data.RTA = &allocSummary, allocCatalog.Gaps, rtaResult
	if runtimeResult != nil {
		data.Runtime = &report.Runtime{Result: *runtimeResult, Findings: runtimeFindings(runtimeObs, results), TracedNoEntry: tracedNoEntry(runtimeTracedBy, runtimeObs, results), NotInGraph: runtimeResult.NotInGraph()}
	}
	data.Withdrawn = resolution.Withdrawn
	data.SnapshotChanges, data.SnapshotComparedWith = changes, compareID(sr)

	// CVE:F9.9: every artifact is rendered in memory first and written
	// at the end, all or nothing (internal/atomicfile): an error anywhere
	// before — including the v1/v2 consistency check — leaves no new
	// report that looks complete. The report, the v2 document and the
	// snapshot manifest carry the same analysis_id; the report also names
	// the results file's SHA-256 (v1 has no field for it).
	//
	// The v2 document is built on every run so that its v1 export is
	// checked against the engine's v1 rows (contract §5) either way.
	doc := resultsv2.Build(resultsv2.Inputs{
		ToolVersion: version(), Offline: cfg.offline, Inventory: components, CVEs: cveList, Mappings: vm, Results: results,
		Freshness:         resultsv2.DataFreshness{State: freshnessState, TTL: cfg.dataTTL.String(), MaxAge: maxAgeString(cfg.maxDataAge), Kinds: freshness},
		Withdrawn:         resolution.Withdrawn,
		DataSet:           dataSetRef(sr, cfg, observations),
		Changes:           changes,
		Gaps:              resultsv2.GlobalGaps{UnresolvedSites: len(gapSites), Reachable: reachableGaps, Unbounded: unboundedGaps, ModelAnalysis: modelGapReasons},
		DeserModelCatalog: deserialize.ModelCatalogVersion, DeserChainCatalog: deserialize.ChainCatalogVersion,
		DeserSites: deserSites, DeserChains: deserChains, TaintCatalog: taintCatalog(cfg.taint),
		CallGraph: &resultsv2.CallGraph{Config: g.Config(), Nodes: g.NodeCount(), Edges: g.EdgeCount(), Comparison: graphCmp, CreatedTypes: &allocSummary, RTA: rtaResult},
		Runtime:   runtimeResult, RuntimeObserved: runtimeObs, RuntimeTraced: runtimeTracedBy,
	})
	if mismatch := resultsv2.CheckV1(doc, results); len(mismatch) > 0 {
		return fmt.Errorf("v2 findings do not aggregate to the v1 results (internal error): %s", strings.Join(mismatch, "; "))
	}
	log("results: %d advisories, %d findings per component instance (%d advisories without an instance); v1 export from v2 consistent",
		doc.Coverage.Counts.Advisories, doc.Coverage.Counts.Findings, len(doc.Coverage.AdvisoriesWithoutInstance))
	if doc, err = resultsv2.WithAnalysisID(doc); err != nil {
		return err
	}
	var resultsBytes []byte
	if cfg.jsonFormat == "v2" {
		resultsBytes, err = resultsv2.Marshal(doc)
	} else {
		resultsBytes, err = report.ResultsJSON(results)
	}
	if err != nil {
		return fmt.Errorf("rendering results JSON: %w", err)
	}
	data.AnalysisID = doc.AnalysisID
	// DESERIALIZATION Z4: the advisory row names the catalog chains
	// linked to its findings (same advisory, same instance).
	// Rows are per advisory and package; key the notes the same way.
	pkgOf := map[string]string{}
	for _, c := range doc.Components {
		if parts := strings.Split(c.GAV, ":"); len(parts) >= 2 {
			pkgOf[c.ID] = parts[0] + ":" + parts[1]
		}
	}
	chainNotes := map[string][]string{}
	for _, f := range doc.Findings {
		for _, c := range f.DeserializationChains {
			n := c.Chain + " " + c.State
			if c.Copy != "" {
				n = c.Chain + " (copy in " + c.Copy + ") " + c.State
			}
			if c.Flow != "" && c.Flow != deserialize.FlowNotAnalyzed {
				n += ", request data " + c.Flow // DESERIALIZATION Z3 (-taint)
			}
			k := f.AdvisoryID + "\x00" + pkgOf[f.ComponentInstanceID]
			chainNotes[k] = append(chainNotes[k], n)
		}
	}
	for i := range data.Rows {
		if ns := chainNotes[data.Rows[i].CVE+"\x00"+data.Rows[i].Package]; len(ns) > 0 {
			data.Rows[i].Chains = strings.Join(ns, "; ")
		}
	}
	data.ResultsFile = filepath.Base(cfg.jsonOut)
	data.ResultsFormat = cfg.jsonFormat
	if data.ResultsFormat == "" {
		data.ResultsFormat = "v1"
	}
	data.ResultsSHA256 = snapshot.HashBytes(resultsBytes)
	rendered, err := report.RenderMarkdown(data)
	if err != nil {
		return fmt.Errorf("writing report: %w", err)
	}
	var outputs atomicfile.Set
	outputs.Add(cfg.out, []byte(rendered))
	outputs.Add(cfg.jsonOut, resultsBytes)
	if cfg.graphOut != "" {
		var gb strings.Builder
		if err := callgraph.Write(&gb, graphNow); err != nil {
			return fmt.Errorf("writing call graph: %w", err)
		}
		outputs.Add(cfg.graphOut, []byte(gb.String()))
	}
	if cfg.runtimeFilter != "" {
		outputs.Add(cfg.runtimeFilter, runtimeFilterSel.jfcFile(cfg.runtimeStatus))
	}
	if sr.recorder != nil {
		current.AnalysisID = doc.AnalysisID
		manifest, err := snapshot.Marshal(current)
		if err != nil {
			return fmt.Errorf("writing snapshot: %w", err)
		}
		outputs.Add(snapshot.ManifestPath(cfg.snapshotOut), manifest)
	}
	if err := outputs.Commit(); err != nil {
		return err
	}
	log("wrote %s", cfg.out)
	log("wrote %s", cfg.jsonOut)
	if cfg.runtimeFilter != "" {
		log("wrote runtime filter %s: exact entries of %d mapped vulnerable methods of %d advisories (%s; %d mapped methods not in the input); record with -XX:StartFlightRecording=settings=profile,settings=%s (JDK 25+)",
			cfg.runtimeFilter, len(runtimeFilterSel.Filters), runtimeFilterSel.Advisories, cfg.runtimeStatus, runtimeFilterSel.Skipped, cfg.runtimeFilter)
	}
	if cfg.graphOut != "" {
		log("wrote call graph %s (%d edges, %s)", cfg.graphOut, len(graphNow.Edges), callgraph.ExportFormat)
	}
	if sr.recorder != nil {
		log("wrote snapshot %s (data set %s, %d observations)", cfg.snapshotOut, snapshot.DataSetID(observations), len(observations))
	}
	log("analysis %s", doc.AnalysisID)
	return nil
}

func compareID(sr *snapshotRun) string {
	if sr.compare == nil {
		return ""
	}
	return sr.compare.DataSetID
}

// dataSetRef is v2's reference to the data snapshot of the run, if any.
func dataSetRef(sr *snapshotRun, cfg config, obs []httpcache.Observation) *resultsv2.DataSet {
	switch {
	case sr.replay != nil:
		return &resultsv2.DataSet{ID: sr.replay.DataSetID, Replayed: true, ConfigMatches: len(snapshot.ConfigDiff(sr.replay.Config, sr.config)) == 0}
	case sr.recorder != nil:
		return &resultsv2.DataSet{ID: snapshot.DataSetID(obs), ConfigMatches: true}
	}
	return nil
}

// methodExists grounds an LLM-suggested method in the analyzed bytecode: a
// constructor must be declared on exactly that class; any other method may
// be inherited, found by the JVM's own lookup order.
func methodExists(h *hierarchy.Hierarchy) func(fqcn, method string) bool {
	return func(fqcn, method string) bool {
		if method == "<init>" || method == "<clinit>" {
			ci, ok := h.Lookup(fqcn)
			if !ok {
				return false
			}
			_, declared := ci.MethodByName(method)
			return declared
		}
		_, found := h.ResolveMethod(fqcn, method)
		return found
	}
}

// declaredOverloads answers CVE:F5b's bytecode verification: the
// descriptors of the methods named name that class fqcn itself declares,
// bridge and synthetic methods excluded (a covariant-return bridge is not
// a separate overload of the source method).
func declaredOverloads(h *hierarchy.Hierarchy) func(fqcn, name string) ([]string, bool) {
	const accBridge, accSynthetic = 0x0040, 0x1000
	return func(fqcn, name string) ([]string, bool) {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			return nil, false
		}
		var descs []string
		for _, m := range ci.Methods {
			if m.Name == name && m.AccessFlags&(accBridge|accSynthetic) == 0 {
				descs = append(descs, m.Descriptor)
			}
		}
		return descs, true
	}
}

// version returns the module version, or the VCS revision for local builds.
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	rev, dirty := "", false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return "devel"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if dirty {
		rev += "-dirty"
	}
	return "devel+" + rev
}

func maxAgeString(d time.Duration) string {
	if d == 0 {
		return ""
	}
	return d.String()
}

// releaseLineSummary counts readable fix commits by release line and the
// commits left out of mappings (CVE:F5c).
func releaseLineSummary(vm map[string]vulnmap.Entry) string {
	var descends, inRelease, other, unknown, exLine, exFit int
	for _, e := range vm {
		for _, c := range e.FixCommits {
			if c.Error != "" {
				continue
			}
			switch c.ReleaseLine {
			case vulnmap.LineDescends:
				descends++
			case vulnmap.LineInRelease:
				inRelease++
			case vulnmap.LineOther:
				other++
			default:
				unknown++
			}
			switch c.Excluded {
			case vulnmap.ExcludedReleaseLine:
				exLine++
			case vulnmap.ExcludedReleaseFit:
				exFit++
			}
		}
	}
	return fmt.Sprintf("descends from its tag=%d, already in it=%d, another line=%d, unknown=%d; left out of mappings: by release line=%d, by release fit=%d",
		descends, inRelease, other, unknown, exLine, exFit)
}

// taintCatalog is the taint catalog version when -taint ran, else "".
func taintCatalog(on bool) string {
	if on {
		return taint.CatalogVersion
	}
	return ""
}

// graphSettings renders a comparison's settings differences.
func graphSettings(diffs []string) string {
	if len(diffs) == 0 {
		return "identical"
	}
	return "DIFFERENT — " + strings.Join(diffs, "; ")
}

// servletRegistrationSummary renders the registered servlets/filters with
// their basis for the log: " (class [kind, basis]; ...)".
func servletRegistrationSummary(regs []servletdispatch.Registration) string {
	if len(regs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(regs))
	for _, r := range regs {
		parts = append(parts, fmt.Sprintf("%s [%s, %s]", r.Class, r.Kind, r.Basis))
	}
	return " (" + strings.Join(parts, "; ") + ")"
}
