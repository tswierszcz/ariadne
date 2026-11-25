package main

// RESEARCH_INTERFACE RI1–RI3: `ariadne research index` and
// `ariadne research query` (docs/research-interface/CONTRACT.md).
// index runs the same static pipeline as the CVE scan (analyze), stops
// before any advisory lookup, and publishes the frozen graph as a bundle;
// query answers one request from a bundle without the JARs.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ariadne/internal/analysis"
	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/component"
	"ariadne/internal/hierarchy"
	"ariadne/internal/libapi"
	"ariadne/internal/reach"
	"ariadne/internal/research"
)

const researchUsage = `usage:
  ariadne research index -app JAR -roots all|profile|api [-deps DIR] [-deps-list FILE]
                         [-profile NAME] [-target-java N] ... -out BUNDLE_DIR
  ariadne research query -bundle BUNDLE_DIR -request REQUEST.json -out RESPONSE.json [-verify operation|full]
  ariadne research check -bundle BUNDLE_DIR -request REQUEST.json -response RESPONSE.json [-replay]
  ariadne research trace-config -bundle BUNDLE_DIR -request TARGETS.json -out CONFIG.json -jfc SETTINGS.jfc
  ariadne research correlate -bundle BUNDLE_DIR -trace RUN.trace [-config CONFIG.json] -out CORRELATION.json

index builds the call graph of the built input with every model the CVE scan
uses, without advisories and without network access (Maven Central reference
data only from the cache and ARIADNE_MAVEN_LOCAL), and publishes it as a
research bundle (contract ` + research.ContractRevision + `). query answers one request
(operations: entries, methods, calls, callers, callees, reachability) from the
bundle. Exit status: 0 a result (also missing / not_found / incomplete),
2 an invalid argument or request, 1 a bundle, I/O or internal failure.
query verifies the datasets its operation reads in full and the others for
presence and size; -verify full verifies every dataset.
check is the producer's verdict on a response (RESEARCH_HARNESS RH-A0): it
writes {"valid", "violations", "replay"} to stdout; exit 0 valid, 1 a broken
rule (or, with -replay, a response this producer would not give), 2 an
invalid argument or an unreadable bundle.
trace-config and correlate (RI6, experimental, outside contract
1.0.0-draft.1) turn targets into a JFR jdk.MethodTrace filter and attach an
ariadne-runtime-trace-1 trace to the bundle by class content.
See docs/research-interface/README.md, ACCEPTANCE.md and RUNTIME.md.
`

func researchMain(args []string, stdout, stderr io.Writer, getenv func(string) string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stderr, researchUsage)
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	switch args[0] {
	case "index":
		return researchIndexMain(args[1:], stderr, getenv)
	case "query":
		return researchQueryMain(args[1:], stderr)
	case "check":
		return researchCheckMain(args[1:], stdout, stderr)
	case "trace-config":
		return researchTraceConfigMain(args[1:], stderr)
	case "correlate":
		return researchCorrelateMain(args[1:], stderr)
	}
	fmt.Fprintf(stderr, "ariadne research: unknown command %q\n%s", args[0], researchUsage)
	return 2
}

func researchIndexMain(args []string, stderr io.Writer, getenv func(string) string) int {
	fs := flag.NewFlagSet("ariadne research index", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var cfg config
	fs.StringVar(&cfg.app, "app", "", "the built input: a library or application jar, a directory, a WAR/EAR or a Spring Boot fat jar")
	fs.StringVar(&cfg.deps, "deps", "", "directory of dependency jars")
	fs.StringVar(&cfg.depsList, "deps-list", "", "optional `mvn dependency:list` output naming the dependency jars")
	fs.StringVar(&cfg.roots, "roots", "", `required: "all", "profile" or "api" (as for the CVE scan)`)
	fs.StringVar(&cfg.profile, "profile", "auto", "profile: auto, generic, spring, servlet, jakarta, micronaut, quarkus, stripes, or a -profile-file name")
	fs.StringVar(&cfg.profileFile, "profile-file", "", "load/override one profile from a YAML file")
	fs.IntVar(&cfg.targetJava, "target-java", 0, "Java feature version for multi-release jars (0: base classes)")
	fs.StringVar(&cfg.activeProfiles, "active-profiles", "", "comma-separated active Spring profiles")
	fs.StringVar(&cfg.jpmsScenario, "jpms-scenario", "classpath", `"classpath", "module-path" or "mixed"`)
	fs.StringVar(&cfg.jpmsModulePath, "jpms-module-path", "", "with -jpms-scenario mixed: jar name patterns on the module path")
	fs.StringVar(&cfg.apiExclude, "api-exclude", "", "with -roots api: package prefixes left out of the API")
	fs.StringVar(&cfg.libraryPOM, "library-pom", "", "with -roots api: the library's pom.xml")
	fs.StringVar(&cfg.callgraphMode, "callgraph-mode", callgraph.ModeCHA, `"cha", "rta" or "pta"`)
	fs.StringVar(&cfg.embeddedComps, "embedded-components", "", "declared embedded components (YAML)")
	fs.StringVar(&cfg.sbom, "sbom", "", "CycloneDX JSON SBOM for embedded-component candidates")
	fs.StringVar(&cfg.cacheDir, "cache-dir", ".cache", "cache directory read for Maven Central reference data (never written from the network)")
	var out, scope, sources string
	fs.StringVar(&out, "out", "", "bundle directory to publish; must not exist")
	fs.StringVar(&sources, "sources", "", "optional source tree (directory) or sources jar of the input: each method's source file, found by its package path, gets its SHA-256 (binding stays declared)")
	fs.StringVar(&scope, "scope", "all", `"all" (every graph node) or "reachable" (only what the roots and the input's own methods reach: the same reachability answers, a smaller bundle; other methods resolve as missing)`)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "ariadne research index: unexpected argument %q\n", fs.Arg(0))
		return 2
	}
	if out == "" || cfg.roots == "" {
		fmt.Fprintln(stderr, "ariadne research index: -out and -roots are required")
		return 2
	}
	if scope != "all" && scope != "reachable" {
		fmt.Fprintf(stderr, "ariadne research index: -scope must be \"all\" or \"reachable\", got %q\n", scope)
		return 2
	}
	if _, err := os.Lstat(out); err == nil {
		fmt.Fprintf(stderr, "ariadne research index: %s already exists: a bundle is never overwritten\n", out)
		return 2
	}
	if fi, err := os.Stat(filepath.Dir(filepath.Clean(out))); err != nil || !fi.IsDir() {
		fmt.Fprintf(stderr, "ariadne research index: -out %s: its parent directory does not exist\n", out)
		return 2
	}
	cfg.offline = true // no network: lookups are answered from the cache or fail
	cfg.centralURL = getenv("ARIADNE_MAVEN_CENTRAL")
	cfg.mavenLocal = mavenLocal(getenv("ARIADNE_MAVEN_LOCAL"))
	cfg.dataTTL = 1
	cfg.out, cfg.jsonOut, cfg.jsonFormat, cfg.runtimeStatus = "-", "-", "v1", defaultRuntimeFilterStatus
	cfg.altVersions, cfg.altVersionsMode = "off", "approx"
	if err := validate(cfg); err != nil {
		fmt.Fprintln(stderr, "ariadne research index:", err)
		return 2
	}
	cfg.research = &researchOut{dir: out, scope: scope, sourcesArg: sources}
	if sources != "" {
		t, err := openSources(sources)
		if err != nil {
			fmt.Fprintln(stderr, "ariadne research index:", err)
			return 2
		}
		defer t.close()
		cfg.research.sources = t
	}
	static, err := prepare(context.Background(), cfg, stderr)
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research index:", err)
		return 1
	}
	defer static.cleanup()
	log := func(format string, a ...any) { fmt.Fprintf(stderr, format+"\n", a...) }
	id, err := writeResearchBundle(static.research(cfg), log)
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research index: research bundle:", err)
		return 1
	}
	fmt.Fprintf(stderr, "research bundle %s published (analysis %s)\n", out, id)
	return 0
}

func researchQueryMain(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("ariadne research query", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var bundleDir, reqPath, out, verify string
	fs.StringVar(&bundleDir, "bundle", "", "research bundle directory")
	fs.StringVar(&reqPath, "request", "", "request JSON file")
	fs.StringVar(&out, "out", "", "response JSON file (replaced atomically)")
	fs.StringVar(&verify, "verify", "operation", `"operation": verify the datasets the operation reads in full, the others for presence and size; "full": every dataset`)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if bundleDir == "" || reqPath == "" || out == "" || fs.NArg() > 0 {
		fmt.Fprintln(stderr, "ariadne research query: -bundle, -request and -out are required")
		return 2
	}
	if verify != "operation" && verify != "full" {
		fmt.Fprintln(stderr, `ariadne research query: -verify is "operation" or "full"`)
		return 2
	}
	raw, rerr := os.ReadFile(reqPath)
	fail := func(e *research.Error) int {
		fmt.Fprintln(stderr, "ariadne research query:", e.Error())
		if werr := writeResponse(out, research.ErrorEnvelope(raw, e)); werr != nil {
			fmt.Fprintln(stderr, "ariadne research query: writing the error response:", werr)
		}
		return e.ExitCode()
	}
	if rerr != nil {
		return fail(&research.Error{Code: "IO_ERROR", Message: rerr.Error(), Location: "-request", Reason: "cannot read the request"})
	}
	// The operation decides which datasets are read and verified in full;
	// the others are checked for presence and size (research.LoadFor).
	var head struct {
		Operation string `json:"operation"`
	}
	json.Unmarshal(raw, &head)
	if verify == "full" {
		head.Operation = "" // every dataset
	}
	b, err := research.LoadFor(bundleDir, head.Operation)
	if err != nil {
		var re *research.Error
		if errors.As(err, &re) {
			if re.Location == "" {
				re.Location, re.Reason = "-bundle", re.Message
			}
			return fail(re)
		}
		return fail(&research.Error{Code: "INTERNAL_ERROR", Message: err.Error()})
	}
	req, qerr := research.ParseRequest(b, raw)
	if qerr != nil {
		return fail(qerr)
	}
	resp, qerr := research.Query(b, req)
	if qerr != nil {
		return fail(qerr)
	}
	if err := writeResponse(out, resp); err != nil {
		fmt.Fprintln(stderr, "ariadne research query: writing the response:", err)
		return 1
	}
	return 0
}

func researchCheckMain(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("ariadne research check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var bundleDir, reqPath, respPath string
	var replay bool
	fs.StringVar(&bundleDir, "bundle", "", "research bundle directory (verified in full)")
	fs.StringVar(&reqPath, "request", "", "the request JSON file the response answers")
	fs.StringVar(&respPath, "response", "", "response JSON file to check")
	fs.BoolVar(&replay, "replay", false, "also require the response to be byte-identical to this producer's answer")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if bundleDir == "" || reqPath == "" || respPath == "" || fs.NArg() > 0 {
		fmt.Fprintln(stderr, "ariadne research check: -bundle, -request and -response are required")
		return 2
	}
	reqRaw, err := os.ReadFile(reqPath)
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research check:", err)
		return 2
	}
	respRaw, err := os.ReadFile(respPath)
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research check:", err)
		return 2
	}
	b, err := research.Load(bundleDir)
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research check: bundle:", err)
		return 2
	}
	v := research.CheckResponse(b, reqRaw, respRaw)
	verdict := struct {
		Valid      bool                 `json:"valid"`
		Violations []research.Violation `json:"violations"`
		Replay     string               `json:"replay"`
	}{Violations: append([]research.Violation{}, v...), Replay: "skipped"}
	if replay {
		verdict.Replay = "identical"
		if r := research.Replay(b, reqRaw, respRaw); r != nil {
			verdict.Replay = "differs"
		}
	}
	verdict.Valid = len(v) == 0
	out, _ := json.MarshalIndent(verdict, "", "  ")
	fmt.Fprintln(stdout, string(out))
	if !verdict.Valid || verdict.Replay == "differs" {
		return 1
	}
	return 0
}

// RI6 (experimental): docs/research-interface/RUNTIME.md.
func researchTraceConfigMain(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("ariadne research trace-config", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var bundleDir, reqPath, out, jfc string
	fs.StringVar(&bundleDir, "bundle", "", "research bundle directory")
	fs.StringVar(&reqPath, "request", "", "trace request ("+research.TraceRequestSchema+")")
	fs.StringVar(&out, "out", "", "trace configuration ("+research.TraceConfigSchema+")")
	fs.StringVar(&jfc, "jfc", "", "JFR settings file to record with")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if bundleDir == "" || reqPath == "" || out == "" || jfc == "" || fs.NArg() > 0 {
		fmt.Fprintln(stderr, "ariadne research trace-config: -bundle, -request, -out and -jfc are required")
		return 2
	}
	raw, err := os.ReadFile(reqPath)
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research trace-config:", err)
		return 2
	}
	b, err := research.LoadFor(bundleDir, "reachability")
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research trace-config: bundle:", err)
		return 1
	}
	req, qerr := research.ParseTraceRequest(b, raw)
	if qerr != nil {
		fmt.Fprintln(stderr, "ariadne research trace-config:", qerr.Error())
		return 2
	}
	cfg := b.TraceConfigFor(req)
	if err := writeResponse(out, cfg); err != nil {
		fmt.Fprintln(stderr, "ariadne research trace-config:", err)
		return 1
	}
	if err := os.WriteFile(jfc, cfg.JFC(), 0o644); err != nil {
		fmt.Fprintln(stderr, "ariadne research trace-config:", err)
		return 1
	}
	fmt.Fprintf(stderr, "trace configuration %s: %d filters for %d targets; record with -XX:StartFlightRecording=settings=profile,settings=%s (JDK 25+)\n", out, len(cfg.Filters), len(cfg.Targets), jfc)
	return 0
}

func researchCorrelateMain(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("ariadne research correlate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var bundleDir, tracePath, cfgPath, out string
	fs.StringVar(&bundleDir, "bundle", "", "research bundle directory")
	fs.StringVar(&tracePath, "trace", "", "runtime trace (ariadne-runtime-trace-1, docs/RUNTIME_TRACE.MD)")
	fs.StringVar(&cfgPath, "config", "", "optional trace configuration naming the targets")
	fs.StringVar(&out, "out", "", "correlation ("+research.CorrelationSchema+")")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if bundleDir == "" || tracePath == "" || out == "" || fs.NArg() > 0 {
		fmt.Fprintln(stderr, "ariadne research correlate: -bundle, -trace and -out are required")
		return 2
	}
	t, err := loadRuntimeTrace(tracePath)
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research correlate:", err)
		return 2
	}
	var cfg *research.TraceConfig
	if cfgPath != "" {
		raw, err := os.ReadFile(cfgPath)
		if err == nil {
			cfg = &research.TraceConfig{}
			err = json.Unmarshal(raw, cfg)
		}
		if err != nil || cfg.SchemaVersion != research.TraceConfigSchema {
			fmt.Fprintf(stderr, "ariadne research correlate: -config %s: not a %s document (%v)\n", cfgPath, research.TraceConfigSchema, err)
			return 2
		}
	}
	b, err := research.Load(bundleDir)
	if err != nil {
		fmt.Fprintln(stderr, "ariadne research correlate: bundle:", err)
		return 1
	}
	c, qerr := b.Correlate(t, cfg)
	if qerr != nil {
		fmt.Fprintln(stderr, "ariadne research correlate:", qerr.Error())
		return 2
	}
	if err := writeResponse(out, c); err != nil {
		fmt.Fprintln(stderr, "ariadne research correlate:", err)
		return 1
	}
	fmt.Fprintf(stderr, "correlation %s: classes matched %d, other bytes %d, not in the bundle %d; calls in the bundle %d, not %d, unmatched %d\n",
		out, c.Classes.Matched, c.Classes.HashDiffers, c.Classes.NotInBundle, c.Calls["in_bundle"], c.Calls["not_in_bundle"], c.Calls["unmatched"])
	return 0
}

// writeResponse publishes a response file: a temporary file in the same
// directory, then a rename over the destination.
func writeResponse(path string, v any) error {
	b, err := research.MarshalResponse(v)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".partial-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// researchOut is set on a research index run: prepare's state goes to
// writeResearchBundle instead of the CVE stage.
type researchOut struct {
	dir        string
	scope      string // "all" or "reachable"
	sources    *sourceTree
	sourcesArg string
}

// researchState is the static pipeline's result (RESEARCH_INTERFACE RI1):
// everything the CVE stage would start from, before any advisory.
type researchState struct {
	cfg          config
	h            *hierarchy.Hierarchy
	g            *callgraph.Graph
	roots        []uint32
	inv          *component.Inventory
	apiInherited map[uint32][]string
	apiAnnounced map[uint32][]string
	gapSites     []reach.GapSite
	reachable    func(fqcn, name, desc string) bool
	dispatch     func(u uint32, e callgraph.Edge) (client, supplied bool)
	framework    *reach.FrameworkMatcher
	profile      string
	activeProfs  []string
	fixedPoint   analysis.Result
	disabled     []string
}

// research selects what the bundle needs from the static pipeline's state.
func (s *staticRun) research(cfg config) researchState {
	return researchState{cfg: cfg, h: s.h, g: s.g, roots: s.roots, inv: s.components,
		apiInherited: s.apiInherited, apiAnnounced: s.apiAnnounced,
		gapSites: s.gapSites, reachable: s.engine.MethodReachable, dispatch: s.engine.DispatchFlags, framework: s.framework, profile: s.prof.Name,
		activeProfs: s.activeProfiles, fixedPoint: s.modelResult, disabled: analysis.Labels(s.disabledParts)}
}

// embeddedLimit names the embedded library candidates that were not
// identified: their classes stay in the host's component, so nothing in the
// records tells them apart. reference_unavailable is the offline case — the
// reference jar is neither in -cache-dir nor in ARIADNE_MAVEN_LOCAL.
func embeddedLimit(inv *component.Inventory) string {
	byStatus := map[string][]string{}
	n := 0
	for _, e := range inv.Embedded {
		if e.Kind != component.EmbeddedLibrary || e.Status == component.EmbeddedConfirmed || e.Status == component.EmbeddedVersionCandidates {
			continue
		}
		byStatus[e.Status] = append(byStatus[e.Status], e.GAV.String())
		n++
	}
	if n == 0 {
		return ""
	}
	var parts []string
	for st, gavs := range byStatus {
		parts = append(parts, st+": "+strings.Join(sortedUnique(gavs), ", "))
	}
	sort.Strings(parts)
	l := fmt.Sprintf("Embedded components: %d library candidates not identified (%s); whatever classes of theirs the host carries are attributed to the host component.", n, strings.Join(parts, "; "))
	if len(byStatus[component.EmbeddedReferenceUnavailable]) > 0 {
		l += " reference_unavailable: the reference jar is in neither -cache-dir nor ARIADNE_MAVEN_LOCAL (research index does not use the network)."
	}
	return l
}

func rid(prefix string, parts ...string) string {
	s := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return prefix + hex.EncodeToString(s[:12])
}

func sha(hexsum string) string { return "sha256:" + hexsum }

func siteKind(k classfile.CallKind) string {
	switch k {
	case classfile.CallStatic, classfile.CallSpecial, classfile.CallVirtual, classfile.CallInterface, classfile.CallDynamic:
		return string(k)
	case classfile.CallInit:
		return "implicit"
	}
	return "model"
}

func splitKey(key string) (fqcn, name, desc string) {
	fqcn, rest, _ := strings.Cut(key, "#")
	i := strings.IndexByte(rest, '(')
	if i < 0 {
		return fqcn, rest, ""
	}
	return fqcn, rest[:i], rest[i:]
}

const (
	accNative   = 0x0100
	accAbstract = 0x0400
)

// writeResearchBundle turns the static pipeline's state into research
// records and publishes the bundle (RESEARCH_INTERFACE RI2).
func writeResearchBundle(st researchState, log func(string, ...any)) (string, error) {
	h, g, inv := st.h, st.g, st.inv
	recs := research.Records{}
	var limits []string

	// Components and their artifacts.
	artifactSHA := map[string]string{}
	for _, a := range inv.Artifacts {
		artifactSHA[a.ID] = a.SHA256
	}
	artifactLoc := map[string]string{}
	for _, a := range inv.Artifacts {
		artifactLoc[a.ID] = a.Location
	}
	if l := embeddedLimit(inv); l != "" {
		limits = append(limits, l)
	}
	frameworkInst := map[string]bool{}
	for _, in := range inv.Instances {
		c := research.Component{RecordType: research.TypeComponent, InstanceID: in.ID, Role: "dependency",
			ArtifactSHA256: sha(artifactSHA[in.ArtifactID]), Location: artifactLoc[in.ArtifactID]}
		if in.EmbeddedPath != "" {
			c.Location += "!/" + in.EmbeddedPath
		}
		switch {
		case in.Detection == component.DetectedApp && st.cfg.roots == "api":
			c.Role = "library"
		case in.Detection == component.DetectedApp:
			c.Role = "application"
		}
		switch {
		case len(in.VersionCandidates) > 0:
			c.Identity = "version_candidates"
		case in.GAV != nil:
			c.Identity = "identified"
			s := in.GAV.String()
			c.GAV = &s
		default:
			c.Identity = "unidentified"
		}
		recs.Components = append(recs.Components, c)
	}

	// Selected classes, and the shadowed copies of conflicts.
	lookup, closeJars := classLookup(h)
	defer closeJars()
	classIDOf := map[string]string{} // fqcn -> selected class id
	classSHA := map[string]string{}
	classInst := map[string]string{}
	noHash := 0
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		inst := inv.InstanceForClass(ci)
		cs, _, ok := lookup(fqcn)
		if !ok || cs == "" || inst == "" {
			noHash++
			continue
		}
		id := rid("k-", inst, fqcn)
		classIDOf[fqcn], classSHA[fqcn], classInst[fqcn] = id, cs, inst
		recs.Classes = append(recs.Classes, research.Class{RecordType: research.TypeClass, ClassID: id, InstanceID: inst, Class: fqcn,
			ClassSHA256: sha(cs), Selected: true, SelectedClassID: id, VariantJava: ci.VariantVersion})
	}
	shadowed := 0
	seenClass := map[string]bool{}
	for _, c := range recs.Classes {
		seenClass[c.ClassID] = true
	}
	for _, cf := range inv.Conflicts {
		kept, ok := classIDOf[cf.FQCN]
		if !ok || cf.ShadowSHA256 == "" || cf.ShadowedInstance == "" {
			continue
		}
		id := rid("k-", cf.ShadowedInstance, cf.FQCN)
		if seenClass[id] {
			continue
		}
		seenClass[id] = true
		shadowed++
		recs.Classes = append(recs.Classes, research.Class{RecordType: research.TypeClass, ClassID: id, InstanceID: cf.ShadowedInstance, Class: cf.FQCN,
			ClassSHA256: sha(cf.ShadowSHA256), Selected: false, SelectedClassID: kept})
	}
	if noHash > 0 {
		limits = append(limits, fmt.Sprintf("%d classes have no class record (their class file could not be hashed); their methods keep a null class hash", noHash))
	}

	// Nodes: every graph node, receiver-specialized copies included.
	n := g.NodeCount()
	nodeID := make([]string, n)
	for id := 0; id < n; id++ {
		key := g.Key(uint32(id))
		if o := g.Origin(uint32(id)); o != uint32(id) {
			recv := strings.Join(st.apiInherited[uint32(id)], ",")
			nodeID[id] = rid("n-", key, "@", recv)
		} else {
			nodeID[id] = rid("n-", key)
		}
	}
	// -scope reachable: only the nodes the roots or the input's own
	// methods reach (with the original method of every kept specialized
	// copy). Every kept node keeps all its outgoing edges, so every path
	// from a root is in the bundle and reachability answers are unchanged.
	keep := make([]bool, n)
	if st.cfg.research.scope == "reachable" {
		var queue []uint32
		push := func(id uint32) {
			if !keep[id] {
				keep[id] = true
				queue = append(queue, id)
			}
		}
		for _, r := range st.roots {
			push(r)
		}
		for id := 0; id < n; id++ {
			fqcn, _, _ := splitKey(g.Key(uint32(id)))
			if ci, ok := h.Lookup(fqcn); ok && ci.IsApp {
				push(uint32(id))
			}
		}
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			push(g.Origin(u))
			for _, e := range g.Edges(u) {
				push(e.To)
			}
		}
		kept := 0
		for _, k := range keep {
			if k {
				kept++
			}
		}
		limits = append(limits, fmt.Sprintf("Scope reachable: %d of %d graph nodes, those the roots or the input's own methods reach, with all their edges and calls; a method outside the scope resolves as missing (index with -scope all to see it), and callers lists omit callers outside the scope.", kept, n))
	} else {
		for i := range keep {
			keep[i] = true
		}
	}
	srcPaths, srcHashed := 0, 0
	srcCache, srcUnmatched := map[string]string{}, map[string]bool{}
	methodIDOf := map[string]string{} // key -> method id
	for id := 0; id < n; id++ {
		if !keep[id] {
			continue
		}
		key := g.Key(uint32(id))
		fqcn, name, desc := splitKey(key)
		if !research.ValidMethodDescriptor(desc) {
			return "", fmt.Errorf("graph node %s has no JVM method descriptor", key)
		}
		node := research.Node{RecordType: research.TypeNode, NodeID: nodeID[id], NodeKind: "method", CanonicalNodeID: nodeID[id],
			Body: "external", Source: research.Source{Binding: "unavailable"}}
		m := &research.Method{Class: fqcn, Name: name, Descriptor: desc}
		if ci, ok := h.Lookup(fqcn); ok {
			if inst := classInst[fqcn]; inst != "" {
				m.InstanceID = &inst
				s := sha(classSHA[fqcn])
				m.ClassSHA256 = &s
			}
			for _, hm := range ci.Methods {
				if hm.Name != name || hm.Descriptor != desc {
					continue
				}
				f := int(hm.AccessFlags)
				node.AccessFlags = &f
				switch {
				case hm.AccessFlags&accNative != 0:
					node.Body = "native"
				case hm.AccessFlags&accAbstract != 0:
					node.Body = "abstract"
				case hm.Code != nil:
					node.Body = "bytecode"
				}
				if len(hm.LineNumbers) > 0 {
					lo, hi := hm.LineNumbers[0].Line, hm.LineNumbers[0].Line
					for _, ln := range hm.LineNumbers {
						lo, hi = min(lo, ln.Line), max(hi, ln.Line)
					}
					if lo >= 1 {
						node.Source.LineStart, node.Source.LineEnd = &lo, &hi
						node.Source.Binding = "declared"
					}
				}
				break
			}
			// RI4: the source file the class declares, by package path;
			// with -sources, the hash of that file.
			if sp := sourcePath(fqcn, ci.SourceFile); sp != "" {
				node.Source.Path = &sp
				node.Source.Binding = "declared"
				srcPaths++
				if t := st.cfg.research.sources; t != nil {
					if _, h, ok := t.lookup(sp, srcCache); ok {
						hs := sha(h)
						node.Source.SourceSHA256 = &hs
						srcHashed++
					} else {
						srcUnmatched[sp] = true
					}
				}
			}
		}
		inst := ""
		if m.InstanceID != nil {
			inst = *m.InstanceID
		}
		m.MethodID = rid("m-", inst, key)
		methodIDOf[key] = m.MethodID
		node.Method = m
		if o := g.Origin(uint32(id)); o != uint32(id) {
			node.NodeKind = "receiver_specialization"
			node.CanonicalNodeID = nodeID[o]
			recv := strings.Join(st.apiInherited[uint32(id)], ",")
			if recv == "" {
				recv = "unknown"
			}
			node.ReceiverType = &recv
		}
		recs.Nodes = append(recs.Nodes, node)
	}

	if t := st.cfg.research.sources; t != nil {
		log("source files (RESEARCH_INTERFACE RI4): %d of %d nodes name a source file; %d hashed from -sources (%d files), %d source paths not found there or ambiguous",
			srcPaths, len(recs.Nodes), srcHashed, t.files, len(srcUnmatched))
		if len(srcUnmatched) > 0 {
			limits = append(limits, fmt.Sprintf("-sources: %d source paths the classes declare are absent from the given sources or match more than one file; their methods have no source hash.", len(srcUnmatched)))
		}
	}

	// Edges, with client/supplied dispatch by library mode's rules.
	type siteKey struct {
		key string
		pc  int
	}
	builder := g.Config().Builder
	edgeAt := map[string]int{}
	targetsAt := map[siteKey][]string{} // caller key + pc -> target node ids (canonical callers only)
	for from := 0; from < n; from++ {
		if !keep[from] {
			continue
		}
		fromKey := g.Key(uint32(from))
		canonical := g.Origin(uint32(from)) == uint32(from)
		for _, e := range g.Edges(uint32(from)) {
			id := rid("e-", nodeID[from], nodeID[e.To], string(e.Kind), string(e.Basis))
			sites := []research.Site{}
			client, supplied := st.dispatch(uint32(from), e)
			for _, s := range e.Sites {
				site := research.Site{CallerNodeID: nodeID[from], PC: s.PC, Kind: siteKind(e.Kind)}
				if e.Basis == callgraph.BasisModel || e.Basis == callgraph.BasisSynthetic {
					site.Kind = "model"
				}
				if s.Condition != "" {
					c := s.Condition
					site.Condition = &c
				}
				sites = append(sites, site)
				sk := siteKey{fromKey, s.PC}
				if canonical {
					targetsAt[sk] = append(targetsAt[sk], nodeID[e.To])
				}
			}
			if i, dup := edgeAt[id]; dup {
				recs.Edges[i].Sites = append(recs.Edges[i].Sites, sites...)
				continue
			}
			edgeAt[id] = len(recs.Edges)
			recs.Edges = append(recs.Edges, research.Edge{RecordType: research.TypeEdge, EdgeID: id, From: nodeID[from], To: nodeID[e.To],
				Kind: string(e.Kind), Basis: string(e.Basis), Character: reach.BasisCharacter(e.Basis), Rule: string(e.Basis) + "@" + builder,
				Sites: sites, ClientDispatch: client, SuppliedDispatch: supplied})
		}
	}

	// Roots.
	for _, r := range st.roots {
		basis := map[string]string{"all": "application_method", "profile": "profile_entry", "api": "public_api"}[st.cfg.roots]
		inh, ann := sortedUnique(st.apiInherited[r]), sortedUnique(st.apiAnnounced[r])
		switch {
		case st.cfg.roots == "api" && len(inh) > 0:
			basis = "inherited_api"
		case st.cfg.roots == "api" && len(ann) > 0:
			basis = "announced_api"
		}
		recs.Roots = append(recs.Roots, research.Root{RecordType: research.TypeRoot, RootID: rid("r-", nodeID[r]), NodeID: nodeID[r],
			Basis: basis, InheritedBy: inh, AnnouncedBy: ann})
	}
	recs.Roots = dedupeRoots(recs.Roots)

	// Gaps: the model-gap sites, anchored at their calling method.
	seenGap := map[string]bool{}
	unplaced := 0
	for _, s := range st.gapSites {
		caller, ok := g.ID(s.CallerFQCN, s.CallerName, s.CallerDesc)
		if !ok {
			unplaced++
			continue
		}
		if !keep[caller] {
			continue // outside the scope: the roots never reach it
		}
		var cands []string
		reason := fmt.Sprintf("Unresolved %s site in %s#%s: no bounded set of targets.", s.Kind, s.CallerFQCN, s.CallerName)
		if s.Owners != nil {
			set := map[string]bool{}
			for _, o := range s.Owners {
				if ci, ok := h.Lookup(o); ok {
					if inst := inv.InstanceForClass(ci); inst != "" {
						set[inst] = true
					}
				}
			}
			cands = []string{}
			for k := range set {
				cands = append(cands, k)
			}
			sort.Strings(cands)
			reason = fmt.Sprintf("Unresolved %s site in %s#%s: targets bounded to %d classes.", s.Kind, s.CallerFQCN, s.CallerName, len(s.Owners))
			if len(s.Owners) == 0 && (s.Kind == "client" || s.Kind == "supplied") {
				reason = fmt.Sprintf("Call in %s#%s may run code the library's caller supplies (%s code, LIBRARY_MODE LB2); that code is not analyzed and bounds no library target.", s.CallerFQCN, s.CallerName, s.Kind)
			}
		}
		id := rid("g-", s.Kind, g.Key(caller), strings.Join(s.Owners, ","))
		if seenGap[id] {
			continue
		}
		seenGap[id] = true
		recs.Gaps = append(recs.Gaps, research.Gap{RecordType: research.TypeGap, GapID: id, Mechanism: s.Kind, Reason: reason,
			Site:                     &research.Site{CallerNodeID: nodeID[caller], PC: 0, Kind: "implicit"},
			ReachableFromIndexedRoot: st.reachable(s.CallerFQCN, s.CallerName, s.CallerDesc),
			CandidateInstanceIDs:     cands, Rule: "gaps." + s.Kind + "@b0"})
	}
	if unplaced > 0 {
		limits = append(limits, fmt.Sprintf("%d gap sites have no graph node for their calling method and are not recorded", unplaced))
	}
	fp := research.FixedPoint{State: "converged", Reasons: []string{}}
	if st.fixedPoint.Incomplete {
		fp.State = "incomplete"
		fp.Reasons = append(fp.Reasons, st.fixedPoint.Reasons...)
		if len(fp.Reasons) == 0 {
			fp.Reasons = []string{"the model coordinator's iteration budget ended before a fixed point"}
		}
		recs.Gaps = append(recs.Gaps, research.Gap{RecordType: research.TypeGap, GapID: rid("g-", "fixed_point"), Mechanism: "fixed_point",
			Reason: "Model coordination stopped before a fixed point: " + strings.Join(fp.Reasons, "; "), ReachableFromIndexedRoot: true, Rule: "analysis.fixed_point@b3"})
	}

	// Calls: every bytecode call site of the analyzed methods.
	seenCall := map[string]bool{}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, hm := range ci.Methods {
			key := fqcn + "#" + hm.Name + hm.Descriptor
			caller, ok := g.IDOfKey(key)
			if !ok || !keep[caller] {
				continue
			}
			for _, c := range hm.Calls {
				if !research.ValidMethodDescriptor(c.Desc) || c.Owner == "" || c.Name == "" {
					continue
				}
				id := rid("s-", nodeID[caller], fmt.Sprint(c.PC), string(c.Kind), c.Owner, c.Name, c.Desc)
				if seenCall[id] {
					continue
				}
				seenCall[id] = true
				body := false
				if decl, ok := h.ResolveExact(c.Owner, c.Name, c.Desc); ok {
					if dci, ok := h.Lookup(decl); ok {
						for _, dm := range dci.Methods {
							if dm.Name == c.Name && dm.Descriptor == c.Desc && dm.Code != nil {
								body = true
							}
						}
					}
				}
				recs.Calls = append(recs.Calls, research.Call{RecordType: research.TypeCall, CallID: id,
					Site:          research.Site{CallerNodeID: nodeID[caller], PC: c.PC, Kind: siteKind(c.Kind)},
					DeclaredClass: c.Owner, DeclaredName: c.Name, DeclaredDescriptor: c.Desc,
					TargetNodeIDs: sortedUnique(targetsAt[siteKey{key, c.PC}]), BodyInInput: body})
			}
		}
	}

	// Identity.
	fwIDs := []string{}
	if st.framework != nil {
		for _, fqcn := range h.SortedFQCNs() {
			ci, _ := h.Lookup(fqcn)
			if st.framework.Match(ci.GAV) {
				if inst := classInst[fqcn]; inst != "" && !frameworkInst[inst] {
					frameworkInst[inst] = true
					fwIDs = append(fwIDs, inst)
				}
			}
		}
	}
	sort.Strings(fwIDs)
	var inputs []research.Input
	var order []string
	for _, a := range inv.Artifacts {
		inputs = append(inputs, research.Input{ID: a.ID, Role: string(a.Role), LogicalLocation: a.Location, SHA256: sha(a.SHA256)})
		order = append(order, a.ID)
	}
	for _, f := range []struct{ role, path string }{{"deps-list", st.cfg.depsList}, {"profile-file", st.cfg.profileFile}, {"embedded-components", st.cfg.embeddedComps}, {"sbom", st.cfg.sbom}, {"library-pom", st.cfg.libraryPOM}} {
		if f.path == "" {
			continue
		}
		b, err := os.ReadFile(f.path)
		if err != nil {
			return "", fmt.Errorf("hashing %s %s: %w", f.role, f.path, err)
		}
		s := sha256.Sum256(b)
		inputs = append(inputs, research.Input{ID: "in-" + f.role, Role: f.role, LogicalLocation: baseName(f.path), SHA256: sha(hex.EncodeToString(s[:]))})
	}
	if t := st.cfg.research.sources; t != nil {
		inputs = append(inputs, research.Input{ID: "in-sources", Role: "sources", LogicalLocation: baseName(st.cfg.research.sourcesArg), SHA256: sha(t.digest)})
	}
	exe := (*string)(nil)
	if p, err := os.Executable(); err == nil {
		if b, err := os.ReadFile(p); err == nil {
			s := sha(fmt.Sprintf("%x", sha256.Sum256(b)))
			exe = &s
		}
	}
	gc := g.Config()
	ident := research.Identity{
		SchemaVersion: research.IdentitySchema, ContractRevision: research.ContractRevision,
		Provenance: research.Provenance{Kind: "analysis", Producer: "ariadne", Revision: version()},
		Tool: research.Tool{Name: "ariadne", Version: version(), BinarySHA256: exe, Algorithms: map[string]string{
			"callgraph.builder": gc.Builder, "callgraph.skip_rule": gc.SkipRule, "callgraph.unknown_allocation": nonEmpty(gc.UnknownAllocationPolicy),
			"libapi": libapi.Version, "research.paths": research.PathsRule, "research.relevance": research.RelevanceRule,
			"research.work": research.WorkRule, "research.dispatch": research.DispatchRule,
		}},
		Configuration: research.Configuration{RootsMode: st.cfg.roots, GraphMode: gc.Mode, TargetJava: st.cfg.targetJava, Profile: st.profile,
			ActiveProfiles: sortedUnique(st.activeProfs), JPMSScenario: st.cfg.jpmsScenario, APIExclude: sortedUnique(splitList(st.cfg.apiExclude)),
			Settings: map[string]any{research.FrameworkSetting: strings.Join(fwIDs, ","), "participants": strings.Join(gc.Participants, ","),
				"central_lookups": "offline: cache and ARIADNE_MAVEN_LOCAL only", "scope": st.cfg.research.scope}},
		Inputs: inputs, ClasspathOrder: order, Capabilities: research.Capabilities, FixedPoint: fp,
		Coverage: research.Coverage{Limits: append([]string{
			"Classes form one flat class path (" + inv.ClasspathOrder + "); class-loader isolation is not modeled.",
			"Gap sites are anchored at their calling method (pc 0, kind implicit): the model-gap rule records the method of an unresolved site, not its bytecode offset.",
			"Calls list every bytecode call site of the analyzed methods; edges added by models, CHA and declarations are edges, not calls.",
			"Roots carry their selection mode; under -roots all/profile, roots added from web.xml and bean lifecycles are not told apart.",
			"Edges carry client_dispatch by library mode's LB2 rule and supplied_dispatch only for caller-supplied field values (on every path); a caller-supplied parameter of the entry method itself (lb2-7) is not yet reported per entry.",
			"Source locations: path is the package path of the file the class's SourceFile attribute names (a nested class names its outer file), lines the method's LineNumberTable range; with -sources, source_sha256 is the hash of the one file in the given sources that ends with that path. The binding stays declared: nothing ties the given sources to the analyzed build.",
			"No advisory, CVE mapping or network lookup is part of the index; Maven Central reference data comes only from the cache and ARIADNE_MAVEN_LOCAL.",
		}, limits...), DisabledParticipants: sortedUnique(st.disabled)},
	}
	id, err := research.Publish(st.cfg.research.dir, ident, recs, time.Now())
	if err != nil {
		return "", err
	}
	log("research bundle (RESEARCH_INTERFACE RI2, contract %s): %d components, %d classes (%d shadowed copies), %d nodes, %d edges, %d roots, %d gaps, %d calls",
		research.ContractRevision, len(recs.Components), len(recs.Classes), shadowed, len(recs.Nodes), len(recs.Edges), len(recs.Roots), len(recs.Gaps), len(recs.Calls))
	return id, nil
}

func sortedUnique(xs []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, x := range xs {
		if x != "" && !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func dedupeRoots(rs []research.Root) []research.Root {
	seen := map[string]bool{}
	out := rs[:0]
	for _, r := range rs {
		if !seen[r.RootID] {
			seen[r.RootID] = true
			out = append(out, r)
		}
	}
	return out
}

func nonEmpty(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
