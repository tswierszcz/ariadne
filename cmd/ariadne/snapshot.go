package main

import (
	"ariadne/internal/deserialize"
	"ariadne/internal/flowsum"
	"ariadne/internal/samcall"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"ariadne/internal/alloc"
	"ariadne/internal/bootfeatures"
	"ariadne/internal/callgraph"
	"ariadne/internal/component"
	"ariadne/internal/cves"
	"ariadne/internal/exploit"
	"ariadne/internal/httpcache"
	"ariadne/internal/jdbcdispatch"
	"ariadne/internal/log4jconfig"
	"ariadne/internal/logbackconfig"
	"ariadne/internal/proxycatalog"
	"ariadne/internal/reach"
	"ariadne/internal/resultsv2"
	"ariadne/internal/runtimeevents"
	"ariadne/internal/servletdispatch"
	"ariadne/internal/snapshot"
	"ariadne/internal/spidispatch"
	"ariadne/internal/taint"
	"ariadne/internal/vulnmap"
)

// snapshotRun holds a run's CVE:F3 state: a recorder (-snapshot-out), a
// replayed manifest (-snapshot-in) and an earlier manifest to compare with
// (-snapshot-compare).
type snapshotRun struct {
	recorder *snapshot.Recorder
	replay   *snapshot.Manifest
	replayer *snapshot.Replayer
	compare  *snapshot.Manifest
	tool     snapshot.Tool
	algos    map[string]string
	config   map[string]string
}

func algorithmVersions(mode string) map[string]string {
	v := map[string]string{
		"callgraph.builder":      callgraph.BuilderVersion,
		"reach.engine":           reach.EngineVersion,
		"vulnmap.resolver":       vulnmap.ResolverVersion,
		"vulnmap.llm_prompt":     strconv.Itoa(vulnmap.PromptVersion),
		"exploit.prompt":         strconv.Itoa(exploit.PromptVersion),
		"framework_check.prompt": strconv.Itoa(exploit.FrameworkPromptVersion),
		"results.v2":             resultsv2.SchemaVersion,
		"snapshot.format":        strconv.Itoa(snapshot.FormatVersion),
		// DESERIALIZATION Z4: the deserializer models/policy rules and the
		// gadget-chain catalog the run's deserialization results come from.
		"deserialize.models": deserialize.ModelCatalogVersion,
		"deserialize.chains": deserialize.ChainCatalogVersion,
		// DATA_FLOW D3: the dependency-summary rules.
		"flowsum.summaries": flowsum.SummaryVersion,
		// DYNAMIC_LINKAGE L1: SAM-call materialization rules.
		"samcall.rules": samcall.Version,
		// DATA_FLOW D4: the taint sources, guards and rules (-taint).
		"taint.catalog": taint.CatalogVersion,
		// The servlet container dispatch model (signatures, registration
		// sources, Spring Boot catalog): it adds call-graph edges in every mode.
		"servlet.dispatch": servletdispatch.Version,
		"spi.dispatch":     spidispatch.Version,
		// JDBC objects: creator-gated edges in every mode.
		"jdbc.objects": jdbcdispatch.Version,
		// Framework proxies: creator-gated edges in every mode.
		"framework.proxies": proxycatalog.Version,
		// Undertow features a Spring Boot application leaves off: CHA/RTA
		// targets in every mode.
		"boot.features": bootfeatures.Version,
		// log4j 1.2 appenders the packaged configuration leaves out:
		// CHA/RTA targets in every mode.
		"log4j.appenders": log4jconfig.Version,
		// logback appenders the configuration leaves out: CHA/RTA targets
		// in every mode.
		"logback.appenders": logbackconfig.Version,
	}
	// GRAPH_PRECISION P4: with RTA the created-type catalog and the fixed
	// point decide edges, so their rule versions are part of the
	// derivation (CHA: report-only, left out so CHA snapshots don't churn).
	if mode == callgraph.ModeRTA {
		v["alloc.catalog"] = alloc.CatalogVersion
		v["callgraph.rta"] = callgraph.RTAVersion
	}
	return v
}

func toolInfo() snapshot.Tool {
	t := snapshot.Tool{Version: version(), GoVersion: runtime.Version(), Reproducible: true}
	if exe, err := os.Executable(); err == nil {
		if f, err := os.Open(exe); err == nil {
			h := sha256.New()
			if _, err := io.Copy(h, f); err == nil {
				t.BinarySHA256 = hex.EncodeToString(h.Sum(nil))
			}
			f.Close()
		}
	}
	if strings.HasSuffix(t.Version, "dirty") || t.Version == "devel" || t.Version == "unknown" {
		t.Dirty, t.Reproducible = true, false
		t.Note = "locally modified or unversioned build: the version does not identify the source; only this exact binary (binary_sha256) can replay the snapshot"
	}
	return t
}

// analysisConfig is every flag value that shapes the result (output
// paths, cache location and secrets excluded).
func analysisConfig(cfg config) map[string]string {
	return map[string]string{
		"profile": cfg.profile, "roots": cfg.roots, "target-java": strconv.Itoa(cfg.targetJava),
		"active-profiles": cfg.activeProfiles, "jpms-scenario": cfg.jpmsScenario, "jpms-module-path": cfg.jpmsModulePath,
		"llm": strconv.FormatBool(cfg.llm), "llm-model": cfg.llmConfig.Model, "llm-reask-changed": strconv.FormatBool(cfg.llmReaskChanged),
		"exploit-check": strconv.FormatBool(cfg.exploitCheck), "framework-check": strconv.FormatBool(cfg.frameworkCheck),
		"data-ttl": cfg.dataTTL.String(), "max-data-age": cfg.maxDataAge.String(),
		"refresh": strconv.FormatBool(cfg.refresh), "taint": strconv.FormatBool(cfg.taint), "callgraph-mode": cfg.callgraphMode, "allow-stale": strconv.FormatBool(cfg.allowStale),
	}
}

func newSnapshotRun(cfg config) (*snapshotRun, error) {
	sr := &snapshotRun{tool: toolInfo(), algos: algorithmVersions(cfg.callgraphMode), config: analysisConfig(cfg)}
	// RUNTIME_NATIVE N1: with a trace, the importer's rules and the trace
	// file are part of the derivation (the file is an input, above).
	if cfg.runtimeTrace != "" {
		sr.algos["runtime.import"] = runtimeevents.ImportVersion
		sr.algos["runtime.gap_causes"] = runtimeevents.GapCauseVersion
		sr.config["runtime-trace"] = baseName(cfg.runtimeTrace)
	}
	if cfg.snapshotOut != "" {
		r, err := snapshot.NewRecorder(cfg.snapshotOut)
		if err != nil {
			return nil, err
		}
		sr.recorder = r
	}
	if cfg.snapshotIn != "" {
		m, err := snapshot.Read(cfg.snapshotIn)
		if err != nil {
			return nil, err
		}
		if err := snapshot.CheckTool(m, sr.tool, sr.algos); err != nil {
			return nil, err
		}
		sr.replay, sr.replayer = &m, snapshot.NewReplayer(cfg.snapshotIn, m)
	}
	if cfg.snapshotCompare != "" {
		m, err := snapshot.Read(cfg.snapshotCompare)
		if err != nil {
			return nil, fmt.Errorf("-snapshot-compare: %w", err)
		}
		sr.compare = &m
	}
	return sr, nil
}

// attach connects a cache to the snapshot recorder or replay.
func (sr *snapshotRun) attach(c *httpcache.Cache, namespace string) {
	c.SetNamespace(namespace)
	if sr.recorder != nil {
		c.SetRecorder(sr.recorder)
	}
	if sr.replayer != nil {
		c.SetReplay(sr.replayer)
	}
}

// inputs lists the analyzed files by stable location.
func snapshotInputs(cfg config, inv *component.Inventory) ([]snapshot.Input, error) {
	var out []snapshot.Input
	for _, a := range inv.Artifacts {
		out = append(out, snapshot.Input{Role: string(a.Role), Location: a.Location, SHA256: a.SHA256})
	}
	for _, f := range []struct{ role, path string }{{"deps-list", cfg.depsList}, {"vuln-methods", cfg.vulnMethods}, {"profile-file", cfg.profileFile}, {"embedded-components", cfg.embeddedComps}, {"sbom", cfg.sbom}, {"runtime-trace", cfg.runtimeTrace}} {
		if f.path == "" {
			continue
		}
		b, err := os.ReadFile(f.path)
		if err != nil {
			return nil, fmt.Errorf("hashing %s %s: %w", f.role, f.path, err)
		}
		out = append(out, snapshot.Input{Role: f.role, Location: baseName(f.path), SHA256: snapshot.HashBytes(b)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Role+out[i].Location < out[j].Role+out[j].Location })
	return out, nil
}

func baseName(p string) string {
	if i := strings.LastIndexAny(p, `/\`); i >= 0 {
		return p[i+1:]
	}
	return p
}

// derivations builds one derivation key per advisory (CVE_IMPLEMENTATION_PLAN F3.4).
func derivations(list []cves.CVE, vm map[string]vulnmap.Entry, results []reach.Result, inv *component.Inventory, llmModel string) []snapshot.Derivation {
	locByInstance := map[string]string{}
	if inv != nil {
		for _, in := range inv.Instances {
			locByInstance[in.ID] = inv.InstanceLocation(in.ID)
		}
	}
	byCVE := map[string]reach.Result{}
	for _, r := range results {
		byCVE[r.CVE] = r
	}
	out := make([]snapshot.Derivation, 0, len(list))
	for _, c := range list {
		d := snapshot.Derivation{Advisory: c.ID, AdvisoryHash: snapshot.HashJSON(c), Resolver: vulnmap.ResolverVersion}
		for _, a := range c.Affected {
			d.Affected = append(d.Affected, a.Package+"@"+a.Version)
		}
		sort.Strings(d.Affected)
		for _, f := range byCVE[c.ID].Findings {
			d.Instances = append(d.Instances, locByInstance[f.Instance])
		}
		sort.Strings(d.Instances)
		e := vm[c.ID]
		methods := append([]string(nil), e.Methods...)
		sort.Strings(methods)
		d.Source = e.Source
		var vars []any
		for _, v := range e.Variants {
			vars = append(vars, []any{v.Package, v.Versions, v.Methods})
		}
		d.MappingHash = snapshot.HashJSON([]any{methods, e.Role, e.Coverage, e.Review, vars, e.MethodRoles})
		d.SourceHashes = e.SourceHashes
		if e.Source == vulnmap.SourceLLM {
			d.LLM = &snapshot.LLMDerivation{Model: llmModel, PromptVersion: strconv.Itoa(vulnmap.PromptVersion), Reused: e.Reused}
		}
		d.ComputeKey()
		out = append(out, d)
	}
	return out
}

// manifest assembles this run's manifest.
func (sr *snapshotRun) manifest(inputs []snapshot.Input, obs []httpcache.Observation, ders []snapshot.Derivation, results []reach.Result, freshness string) snapshot.Manifest {
	verdicts := map[string]string{}
	for _, r := range results {
		verdicts[r.CVE] = string(r.Status)
	}
	return snapshot.Manifest{
		CreatedAt: time.Now().UTC(), Tool: sr.tool, Algorithms: sr.algos, Config: sr.config,
		Inputs: inputs, Observations: obs, Derivations: ders, Verdicts: verdicts, DataFreshness: freshness,
	}
}
