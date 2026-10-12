// Command ariadne reports which known vulnerabilities (CVE/GHSA) in a JVM
// application's dependencies are reachable from the application's own code.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"ariadne/internal/callgraph"
	"ariadne/internal/cves"
	"ariadne/internal/hierarchy"
	"ariadne/internal/httpcache"
	"ariadne/internal/input"
	"ariadne/internal/reach"
	"ariadne/internal/report"
	"ariadne/internal/vulnmap"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
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
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("ariadne", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var cfg config
	fs.StringVar(&cfg.app, "app", "", "the application's classes: a directory, a plain jar, or a Spring Boot fat jar")
	fs.StringVar(&cfg.deps, "deps", "", "directory of dependency jars (not needed for a Spring Boot fat jar, which carries its own)")
	fs.StringVar(&cfg.depsList, "deps-list", "", "optional `mvn dependency:list` output, used as a fallback when a jar has no pom.properties")
	fs.StringVar(&cfg.vulnMethods, "vuln-methods", "", "override vuln_methods.yaml (default: ariadne's embedded table)")
	fs.StringVar(&cfg.out, "out", "", "report.md output path (required)")
	fs.StringVar(&cfg.jsonOut, "json", "", "results.json output path (required)")
	fs.StringVar(&cfg.cacheDir, "cache-dir", ".cache", "disk cache directory for OSV/GitHub HTTP responses")
	fs.BoolVar(&cfg.offline, "offline", false, "serve only from the cache; fail clearly on a miss instead of touching the network")
	fs.BoolVar(&cfg.showVersion, "version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if cfg.showVersion {
		fmt.Fprintf(stdout, "ariadne %s\n", version())
		return 0
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

	log("building call graph")
	g := callgraph.Build(h, nil)
	log("call graph: %d nodes, %d edges (%d direct, %d added by CHA), %d unresolved dynamic call sites",
		g.NodeCount(), g.EdgeCount(), g.PreCHAEdgeCount(), g.CHAEdgeCount(), g.UnresolvedDynamicCount())

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
	vm, err := vulnmap.NewResolver(manual, cache).ResolveAll(ctx, cveList)
	if err != nil {
		return fmt.Errorf("mapping CVEs to vulnerable methods: %w", err)
	}
	log("mapped %d/%d CVEs to vulnerable methods", len(vm), len(cveList))

	log("analyzing reachability")
	engine := reach.New(g, h, nil)
	results := engine.Analyze(cveList, vm)

	counts := map[reach.Status]int{}
	for _, r := range results {
		counts[r.Status]++
	}
	log("REACHABLE=%d FRAMEWORK=%d NOT_REACHABLE=%d UNKNOWN=%d",
		counts[reach.StatusReachable], counts[reach.StatusFramework], counts[reach.StatusNotReachable], counts[reach.StatusUnknown])

	header := report.Header{
		App:             cfg.app,
		GeneratedAt:     time.Now(),
		AriadneVersion:  version(),
		Profile:         "generic (no profile system yet)",
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
