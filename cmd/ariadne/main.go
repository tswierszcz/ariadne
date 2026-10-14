// Command ariadne reports which known vulnerabilities (CVE/GHSA) in a JVM
// application's dependencies are reachable from the application's own code.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"ariadne/internal/callgraph"
	"ariadne/internal/cves"
	"ariadne/internal/exploit"
	"ariadne/internal/hierarchy"
	"ariadne/internal/httpcache"
	"ariadne/internal/input"
	"ariadne/internal/profile"
	"ariadne/internal/reach"
	"ariadne/internal/report"
	"ariadne/internal/vulnmap"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}

type config struct {
	app         string
	deps        string
	depsList    string
	vulnMethods string
	out         string
	jsonOut     string
	cacheDir    string
	offline     bool
	showVersion bool

	profile     string
	profileFile string
	roots       string

	llm        bool
	llmModel   string
	llmBaseURL string
	llmConfig  vulnmap.LLMConfig // resolved from the flags above and the environment; set when llm or exploitCheck is true

	exploitCheck   bool
	frameworkCheck bool
}

func run(args []string, stdout, stderr io.Writer, getenv func(string) string) int {
	fs := flag.NewFlagSet("ariadne", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var cfg config
	fs.StringVar(&cfg.app, "app", "", "the application's classes: a directory, a plain jar, or a Spring Boot fat jar")
	fs.StringVar(&cfg.deps, "deps", "", "directory of dependency jars (not needed for a Spring Boot fat jar, which carries its own)")
	fs.StringVar(&cfg.depsList, "deps-list", "", "optional `mvn dependency:list` output, used as a fallback when a jar has no pom.properties")
	fs.StringVar(&cfg.vulnMethods, "vuln-methods", "", "override vuln_methods.yaml (default: ariadne's embedded table)")
	fs.StringVar(&cfg.profile, "profile", "auto", "which profile's framework knowledge to use: auto, generic, spring, or a name from -profile-file")
	fs.StringVar(&cfg.profileFile, "profile-file", "", "load/override one profile from a YAML file (same shape as the embedded defaults), in addition to ariadne's own spring/generic")
	fs.StringVar(&cfg.roots, "roots", "all", `BFS roots: "all" (every application method, the default) or "profile" (only the active profile's entry points: annotated methods/classes, main)`)
	fs.StringVar(&cfg.out, "out", "", "report.md output path (required)")
	fs.StringVar(&cfg.jsonOut, "json", "", "results.json output path (required)")
	fs.StringVar(&cfg.cacheDir, "cache-dir", ".cache", "disk cache directory for OSV/GitHub HTTP responses")
	fs.BoolVar(&cfg.offline, "offline", false, "serve only from the cache; fail clearly on a miss instead of touching the network")
	fs.BoolVar(&cfg.llm, "llm", false, "also map CVEs no other source could to vulnerable methods by asking an LLM (needs ARIADNE_LLM_API_KEY and a model)")
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
	if cfg.roots != "all" && cfg.roots != "profile" {
		return fmt.Errorf("-roots must be %q or %q, got %q", "all", "profile", cfg.roots)
	}
	return nil
}

func analyze(ctx context.Context, cfg config, stdout io.Writer) error {
	log := func(format string, a ...any) { fmt.Fprintf(stdout, format+"\n", a...) }

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
	h, warnings, err := hierarchy.Build(prepared.AppClassesJar, depJars)
	if err != nil {
		return fmt.Errorf("building hierarchy: %w", err)
	}
	maxBytecode := 0
	for _, ci := range h.All() {
		if int(ci.MajorVersion) > maxBytecode {
			maxBytecode = int(ci.MajorVersion)
		}
	}
	log("hierarchy: %d classes (max bytecode version %d)", h.Len(), maxBytecode)

	catalog, err := profile.LoadDefaultCatalog()
	if err != nil {
		return fmt.Errorf("loading default profile catalog: %w", err)
	}
	if cfg.profileFile != "" {
		if err := catalog.AddFile(cfg.profileFile); err != nil {
			return fmt.Errorf("loading -profile-file %s: %w", cfg.profileFile, err)
		}
	}
	prof, err := profile.Select(catalog, cfg.profile, h)
	if err != nil {
		return fmt.Errorf("selecting profile: %w", err)
	}
	log("profile: %s", prof.Name)

	log("building call graph")
	g := callgraph.Build(h, nil)
	log("call graph: %d nodes, %d edges (%d direct, %d added by CHA), %d unresolved dynamic call sites",
		g.NodeCount(), g.EdgeCount(), g.PreCHAEdgeCount(), g.CHAEdgeCount(), g.UnresolvedDynamicCount())

	if len(prof.SyntheticEdges) > 0 {
		var synthetic []callgraph.SyntheticEdge
		for _, se := range prof.SyntheticEdges {
			synthetic = append(synthetic, callgraph.SyntheticEdge{FromSubtypesOf: se.FromSubtypesOf, To: se.To})
		}
		added := callgraph.ApplySyntheticEdges(g, h, synthetic)
		log("applied %d synthetic edges from profile %q (e.g. Spring Data repository proxies)", added, prof.Name)
	}

	cache, err := httpcache.New(cfg.cacheDir, cfg.offline)
	if err != nil {
		return fmt.Errorf("opening cache dir %s: %w", cfg.cacheDir, err)
	}

	var gavs []input.GAV
	for _, dj := range depJars {
		if dj.GAV != nil {
			gavs = append(gavs, *dj.GAV)
		}
	}
	log("querying OSV for %d resolved dependencies", len(gavs))
	cveList, err := cves.NewClient(cache).Resolve(ctx, gavs)
	if err != nil {
		return fmt.Errorf("resolving CVEs: %w", err)
	}
	log("found %d CVEs", len(cveList))

	manual, err := vulnmap.LoadManual(cfg.vulnMethods)
	if err != nil {
		return fmt.Errorf("loading vuln_methods.yaml: %w", err)
	}
	resolver := vulnmap.NewResolver(manual, cache)
	var llmSrc *vulnmap.LLMSource
	if cfg.llm {
		llmCache, err := httpcache.New(filepath.Join(cfg.cacheDir, "llm"), cfg.offline)
		if err != nil {
			return fmt.Errorf("opening LLM cache dir: %w", err)
		}
		llmSrc = vulnmap.NewLLMSource(cfg.llmConfig, llmCache, nil)
		resolver.LLM = llmSrc
		resolver.MethodExists = methodExists(h)
		log("LLM source enabled: model %s", cfg.llmConfig.Model)
	}
	vm, err := resolver.ResolveAll(ctx, cveList)
	if err != nil {
		return fmt.Errorf("mapping CVEs to vulnerable methods: %w", err)
	}
	bySource := map[string]int{}
	for _, e := range vm {
		bySource[e.Source]++
	}
	log("mapped %d/%d CVEs to vulnerable methods (manual=%d diff=%d llm=%d)",
		len(vm), len(cveList), bySource[vulnmap.SourceManual], bySource[vulnmap.SourceDiff], bySource[vulnmap.SourceLLM])
	if llmSrc != nil {
		st := llmSrc.Stats()
		log("llm: %d mapped, %d answered with no usable method, %d skipped (request failed), %d invalid reply, %d methods dropped as not in bytecode",
			st.Mapped, st.Empty, st.Skipped, st.Invalid, st.Dropped)
		warnings = append(warnings, llmSrc.Warnings()...)
	}

	log("analyzing reachability")
	framework := reach.NewFrameworkMatcher(prof.ResolvePatterns(h))
	var engine *reach.Engine
	if cfg.roots == "profile" {
		roots := prof.AppRoots(g, h)
		log("roots: profile %q (%d entry-point methods)", prof.Name, len(roots))
		engine = reach.NewWithRoots(g, h, framework, roots)
	} else {
		engine = reach.New(g, h, framework)
	}
	webAssets := map[string]bool{}
	for _, dj := range depJars {
		if dj.WebAssetsOnly && dj.GAV != nil {
			webAssets[dj.GAV.Group+":"+dj.GAV.Artifact] = true
		}
	}
	engine.SetWebAssetPackages(webAssets)
	results := engine.Analyze(cveList, vm)

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

	profileLabel := prof.Name
	if cfg.roots == "profile" {
		profileLabel += " (roots: profile)"
	}
	header := report.Header{
		App:             cfg.app,
		GeneratedAt:     time.Now(),
		AriadneVersion:  version(),
		Profile:         profileLabel,
		MaxBytecodeVer:  maxBytecode,
		GraphNodes:      g.NodeCount(),
		GraphEdges:      g.EdgeCount(),
		PreCHAEdges:     g.PreCHAEdgeCount(),
		CHAEdges:        g.CHAEdgeCount(),
		UnresolvedDyn:   g.UnresolvedDynamicCount(),
		DepsTotal:       len(depJars),
		DepsIdentified:  identified,
		UnidentifiedJar: unidentified,
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
	data := report.Build(header, cveList, results, warnings)

	if err := report.WriteMarkdown(cfg.out, data); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}
	log("wrote %s", cfg.out)

	if err := report.WriteResultsJSON(cfg.jsonOut, results); err != nil {
		return fmt.Errorf("writing results JSON: %w", err)
	}
	log("wrote %s", cfg.jsonOut)

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
