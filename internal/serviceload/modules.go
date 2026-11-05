package serviceload

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

// ModuleScenario names the launch configuration ariadne assumes when it
// interprets module declarations (SERVICE_LOADING S3: "brak scenariusza
// nie oznacza poprawnego odtworzenia pełnej konfiguracji runtime'u").
type ModuleScenario string

// ScenarioClasspath is the only scenario this slice models: every input
// jar is on the class path, so each one is part of the unnamed module,
// every module-info.class is ignored by the JVM, and ServiceLoader sees
// exactly the META-INF/services declarations (S0-S2). Module descriptors
// are inventoried as facts about what would change on a module path,
// never as registrations or call-graph edges.
const ScenarioClasspath ModuleScenario = "classpath"

// ScenarioModulePath puts every input jar, the app included, on the
// application module path (SERVICE_LOADING S3 slice 2): explicit modules
// declare providers only through `provides` (their META-INF/services is
// ignored), automatic modules through META-INF/services with a
// constructor, and a caller in an explicit module must declare `uses`.
// A mixed class path/module path launch and module layers are not
// modeled; every module is assumed to be resolved into the boot layer
// (service binding resolves provider modules of a used service).
const ScenarioModulePath ModuleScenario = "module-path"

// ParseScenario validates a -jpms-scenario value.
func ParseScenario(s string) (ModuleScenario, error) {
	switch ModuleScenario(s) {
	case ScenarioClasspath, ScenarioModulePath:
		return ModuleScenario(s), nil
	}
	return "", fmt.Errorf("unknown JPMS scenario %q (want %q or %q)", s, ScenarioClasspath, ScenarioModulePath)
}

// ModuleKind is how a jar would be treated if it were placed on the
// module path — not how it is treated in ScenarioClasspath (unnamed).
type ModuleKind string

const (
	ModuleExplicit  ModuleKind = "explicit"  // has a module-info.class
	ModuleAutomatic ModuleKind = "automatic" // no descriptor: name from Automatic-Module-Name or the file name
)

// ModuleDescriptor is one input jar's module identity (SERVICE_LOADING
// S3), kept outside the class index: hierarchy.Build deliberately skips
// module-info.class.
type ModuleDescriptor struct {
	Jar  string
	GAV  *input.GAV // nil for the app or an unidentified dependency
	Kind ModuleKind
	// Name is the module name; "" for an automatic module whose file
	// name yields no valid name (NameSource then says why — the JDK's
	// ModuleFinder would reject such a jar on the module path).
	Name string
	// NameSource is where Name came from: the descriptor entry path
	// ("module-info.class", "META-INF/versions/9/module-info.class"),
	// "Automatic-Module-Name", or "file name".
	NameSource string
	// DescriptorNote explains a non-obvious choice of descriptor entry
	// (multi-release selection with -target-java unset); "" otherwise.
	DescriptorNote string
	Info           *classfile.ModuleInfo // explicit modules only
}

// ModuleProviderDecl is one `provides Service with Provider` pair from an
// explicit module's descriptor. A declaration, not a registration: in
// ScenarioClasspath ServiceLoader never reads it.
type ModuleProviderDecl struct {
	Module, Jar       string
	Service, Provider string
	// ProviderFound: the class exists in the analyzed hierarchy.
	ProviderFound bool
	// HasProviderMethod: the class declares `public static provider()`
	// with no parameters — on the module path ServiceLoader would call it
	// instead of a constructor (java.util.ServiceLoader, "Deploying
	// service providers as modules"). Not modeled as an edge yet.
	HasProviderMethod bool
	// InServicesFile: the same jar's META-INF/services/<Service> also
	// names this provider, so the class-path view agrees.
	InServicesFile bool
}

// DeclarationMismatch is a provider that one of an explicit module's two
// declaration mechanisms names and the other does not, so which
// providers ServiceLoader sees depends on the launch scenario.
type DeclarationMismatch struct {
	Module, Jar       string
	Service, Provider string
	// Side is "module-info only" or "META-INF/services only".
	Side   string
	Effect string
}

// ModuleInventory is SERVICE_LOADING S3's first slice: module identity
// and declarations for every input jar, under an explicit scenario.
type ModuleInventory struct {
	Scenario    ModuleScenario
	Descriptors []ModuleDescriptor
	Provides    []ModuleProviderDecl
	Mismatches  []DeclarationMismatch
}

// AutomaticCount is the number of inventoried jars without a descriptor.
func (m ModuleInventory) AutomaticCount() int {
	_, automatic, _, _ := m.Counts()
	return automatic
}

// Counts summarizes the inventory for the CLI log line.
func (m ModuleInventory) Counts() (explicit, automatic, provides, mismatches int) {
	for _, d := range m.Descriptors {
		if d.Kind == ModuleExplicit {
			explicit++
		} else {
			automatic++
		}
	}
	return explicit, automatic, len(m.Provides), len(m.Mismatches)
}

// InventoryModules reads each input jar's module descriptor (or derives
// its automatic-module name), lists `provides` declarations, and compares
// them with the META-INF/services registrations S0 already found (regs).
// It never adds registrations or edges. targetJava selects among
// multi-release descriptors the same way JVM_GRAPH G4 selects classes.
//
// A jar that cannot be opened is an error. A malformed descriptor inside
// an openable jar is a warning, and that jar is left out of Descriptors
// entirely: reporting it as automatic would be a guess (the JDK rejects
// such a jar on the module path).
func InventoryModules(h *hierarchy.Hierarchy, appJarPath string, depJars []input.DepJar, targetJava int, regs []ServiceRegistration) (ModuleInventory, []string, error) {
	inv := ModuleInventory{Scenario: ScenarioClasspath} // ApplyScenario sets the run's actual scenario
	var warnings []string

	type jarRef struct {
		path string
		gav  *input.GAV
	}
	jars := []jarRef{{path: appJarPath}}
	for _, dj := range depJars {
		jars = append(jars, jarRef{path: dj.Path, gav: dj.GAV})
	}

	servicesByJar := map[string]map[string]map[string]bool{} // jar -> service -> provider
	for _, r := range regs {
		if servicesByJar[r.SourceJar] == nil {
			servicesByJar[r.SourceJar] = map[string]map[string]bool{}
		}
		if servicesByJar[r.SourceJar][r.Interface] == nil {
			servicesByJar[r.SourceJar][r.Interface] = map[string]bool{}
		}
		servicesByJar[r.SourceJar][r.Interface][r.ProviderClass] = true
	}

	for _, j := range jars {
		desc, err := readModuleDescriptor(j.path, targetJava)
		if err != nil {
			if _, ok := err.(descriptorError); ok {
				warnings = append(warnings, fmt.Sprintf("%s: %v (module descriptor not inventoried)", j.path, err))
				continue
			}
			return inv, warnings, err
		}
		desc.GAV = j.gav
		inv.Descriptors = append(inv.Descriptors, desc)
		if desc.Kind != ModuleExplicit {
			continue
		}

		declared := map[string]map[string]bool{}
		for _, p := range desc.Info.Provides {
			if declared[p.Service] == nil {
				declared[p.Service] = map[string]bool{}
			}
			for _, impl := range p.With {
				declared[p.Service][impl] = true
				decl := ModuleProviderDecl{
					Module: desc.Name, Jar: j.path,
					Service: p.Service, Provider: impl,
					InServicesFile: servicesByJar[j.path][p.Service][impl],
				}
				if ci, ok := h.Lookup(impl); ok {
					decl.ProviderFound = true
					decl.HasProviderMethod = providerFactoryDesc(ci) != ""
				} else {
					warnings = append(warnings, fmt.Sprintf(
						"%s (module %s): provides %s with %q, not found in the analyzed classes/dependencies",
						j.path, desc.Name, p.Service, impl))
				}
				inv.Provides = append(inv.Provides, decl)
				if !decl.InServicesFile {
					inv.Mismatches = append(inv.Mismatches, DeclarationMismatch{
						Module: desc.Name, Jar: j.path, Service: p.Service, Provider: impl,
						Side:   "module-info only",
						Effect: "seen by ServiceLoader only when this jar is on the module path; not seen on the class path (this run's scenario)",
					})
				}
			}
		}
		services := servicesByJar[j.path]
		for _, svc := range sortedKeys(services) {
			for _, prov := range sortedKeys(services[svc]) {
				if declared[svc][prov] {
					continue
				}
				inv.Mismatches = append(inv.Mismatches, DeclarationMismatch{
					Module: desc.Name, Jar: j.path, Service: svc, Provider: prov,
					Side:   "META-INF/services only",
					Effect: "seen on the class path (this run's scenario); ignored on the module path, where an explicit module's META-INF/services is not consulted",
				})
			}
		}
	}
	return inv, warnings, nil
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// descriptorError marks a malformed descriptor inside a readable jar —
// a diagnostic, not a run failure.
type descriptorError struct{ error }

const moduleInfoEntry = "module-info.class"
const versionsPrefix = "META-INF/versions/"

func readModuleDescriptor(jarPath string, targetJava int) (ModuleDescriptor, error) {
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return ModuleDescriptor{}, fmt.Errorf("opening %s as a zip: %w", jarPath, err)
	}
	defer zr.Close()

	var root *zip.File
	versioned := map[int]*zip.File{}
	var manifest *zip.File
	for _, f := range zr.File {
		switch {
		case f.Name == moduleInfoEntry:
			root = f
		case f.Name == "META-INF/MANIFEST.MF":
			manifest = f
		case strings.HasPrefix(f.Name, versionsPrefix) && strings.HasSuffix(f.Name, "/"+moduleInfoEntry):
			ver := strings.TrimSuffix(strings.TrimPrefix(f.Name, versionsPrefix), "/"+moduleInfoEntry)
			if n, err := strconv.Atoi(ver); err == nil && n >= 9 && !strings.Contains(ver, "/") {
				versioned[n] = f
			}
		}
	}
	attrs := map[string]string{}
	if manifest != nil {
		data, err := readZipEntry(manifest)
		if err != nil {
			return ModuleDescriptor{}, fmt.Errorf("reading %s in %s: %w", manifest.Name, jarPath, err)
		}
		attrs = parseManifestMain(data)
	}
	if !strings.EqualFold(attrs["Multi-Release"], "true") {
		versioned = nil // versioned entries are inert outside a multi-release jar
	}

	chosen, note := pickDescriptor(root, versioned, targetJava)
	if chosen != nil {
		data, err := readZipEntry(chosen)
		if err != nil {
			return ModuleDescriptor{}, fmt.Errorf("reading %s in %s: %w", chosen.Name, jarPath, err)
		}
		cf, err := classfile.Parse(bytes.NewReader(data))
		if err != nil {
			return ModuleDescriptor{}, descriptorError{fmt.Errorf("%s: %w", chosen.Name, err)}
		}
		if !cf.IsModule() || cf.Module == nil {
			return ModuleDescriptor{}, descriptorError{fmt.Errorf("%s has no Module attribute", chosen.Name)}
		}
		return ModuleDescriptor{
			Jar: jarPath, Kind: ModuleExplicit, Name: cf.Module.Name,
			NameSource: chosen.Name, DescriptorNote: note, Info: cf.Module,
		}, nil
	}

	if name := strings.TrimSpace(attrs["Automatic-Module-Name"]); name != "" {
		return ModuleDescriptor{Jar: jarPath, Kind: ModuleAutomatic, Name: name, NameSource: "Automatic-Module-Name", DescriptorNote: note}, nil
	}
	name, reason := automaticModuleName(filepath.Base(jarPath))
	src := "file name"
	if reason != "" {
		src = "file name (" + reason + ")"
	}
	return ModuleDescriptor{Jar: jarPath, Kind: ModuleAutomatic, Name: name, NameSource: src, DescriptorNote: note}, nil
}

// pickDescriptor mirrors JVM_GRAPH G4's rule for classes: with a target,
// the highest versioned descriptor N <= target, else the root one; with
// no target, the root one. A jar that ships a descriptor only under
// META-INF/versions/ is a named module on every runtime >= that N, so
// with no target the lowest such N is used and the choice is noted
// rather than silently treating the jar as automatic.
func pickDescriptor(root *zip.File, versioned map[int]*zip.File, targetJava int) (*zip.File, string) {
	if targetJava > 0 {
		best := -1
		for n := range versioned {
			if n <= targetJava && n > best {
				best = n
			}
		}
		if best >= 0 {
			return versioned[best], ""
		}
		if root == nil && len(versioned) > 0 {
			return nil, fmt.Sprintf("only META-INF/versions/ descriptors above -target-java %d: not a named module for that target", targetJava)
		}
		return root, ""
	}
	if root != nil || len(versioned) == 0 {
		return root, ""
	}
	lowest := -1
	for n := range versioned {
		if lowest < 0 || n < lowest {
			lowest = n
		}
	}
	return versioned[lowest], fmt.Sprintf("-target-java unset: used the lowest versioned descriptor (Java %d+); the jar has no root module-info.class", lowest)
}

func readZipEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// parseManifestMain reads a manifest's main section (up to the first
// blank line), joining 72-byte continuation lines (JAR File
// Specification: a line starting with a single space continues the
// previous one).
func parseManifestMain(data []byte) map[string]string {
	attrs := map[string]string{}
	var lines []string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(raw, "\r")
		if line == "" {
			break
		}
		if strings.HasPrefix(line, " ") && len(lines) > 0 {
			lines[len(lines)-1] += line[1:]
			continue
		}
		lines = append(lines, line)
	}
	for _, line := range lines {
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		attrs[strings.TrimSpace(name)] = strings.TrimSpace(value)
	}
	return attrs
}

var (
	automaticVersionRE = regexp.MustCompile(`-(\d+(\.|$))`)
	nonAlnumRE         = regexp.MustCompile(`[^A-Za-z0-9]`)
	repeatedDotsRE     = regexp.MustCompile(`\.{2,}`)
)

// automaticModuleName derives a name the way java.lang.module.ModuleFinder
// does for a jar without Automatic-Module-Name: drop ".jar", cut at the
// first "-<digit>" version start, map every non-alphanumeric character
// to '.', collapse repeated dots and trim them from both ends. A result
// that is empty or has a segment starting with a digit is not a legal
// module name (the JDK would reject the jar); reason then says so. Java
// keywords as segments are not checked.
func automaticModuleName(fileName string) (name, reason string) {
	base := strings.TrimSuffix(fileName, ".jar")
	if loc := automaticVersionRE.FindStringIndex(base); loc != nil {
		base = base[:loc[0]]
	}
	name = nonAlnumRE.ReplaceAllString(base, ".")
	name = repeatedDotsRE.ReplaceAllString(name, ".")
	name = strings.Trim(name, ".")
	if name == "" {
		return "", "no valid name can be derived"
	}
	for _, seg := range strings.Split(name, ".") {
		if seg[0] >= '0' && seg[0] <= '9' {
			return "", fmt.Sprintf("derived %q has a segment starting with a digit, not a legal module name", name)
		}
	}
	return name, ""
}

// ScenarioView is the provider set and module context ServiceLoader
// would use in one launch scenario. In ScenarioClasspath it is exactly
// the S0 registrations and no module checks apply.
type ScenarioView struct {
	Scenario      ModuleScenario
	Registrations []ServiceRegistration
	moduleByJar   map[string]ModuleDescriptor
}

// ApplyScenario derives the scenario's registrations from the S0
// inventory and the module inventory, and records the scenario on inv.
// Module-path warnings name configurations the JVM would reject at
// startup (an unnamed automatic module, a `requires` on a module absent
// from the input); the analysis still proceeds, as a declared gap.
func ApplyScenario(scenario ModuleScenario, regs []ServiceRegistration, inv *ModuleInventory, h *hierarchy.Hierarchy) (ScenarioView, []string) {
	inv.Scenario = scenario
	view := ScenarioView{Scenario: scenario}
	if scenario != ScenarioModulePath {
		view.Registrations = regs
		return view, nil
	}
	var warnings []string
	view.moduleByJar = map[string]ModuleDescriptor{}
	names := map[string]bool{}
	for _, d := range inv.Descriptors {
		view.moduleByJar[d.Jar] = d
		if d.Name != "" {
			names[d.Name] = true
		} else {
			warnings = append(warnings, fmt.Sprintf("%s: module-path scenario: no valid automatic module name (%s); the JVM would reject this jar", d.Jar, d.NameSource))
		}
	}
	for _, d := range inv.Descriptors {
		if d.Kind != ModuleExplicit {
			continue
		}
		for _, r := range d.Info.Requires {
			if names[r.Module] || r.Flags&classfile.ModuleAccStaticPh != 0 || strings.HasPrefix(r.Module, "java.") || strings.HasPrefix(r.Module, "jdk.") {
				continue
			}
			warnings = append(warnings, fmt.Sprintf("%s: module-path scenario: module %s requires %s, which is not among the input modules; module resolution would fail at startup", d.Jar, d.Name, r.Module))
		}
	}
	for _, r := range regs {
		d, ok := view.moduleByJar[r.SourceJar]
		if ok && d.Kind == ModuleExplicit {
			continue // ignored on the module path; the module's `provides` replaces it
		}
		r.Module = d.Name
		view.Registrations = append(view.Registrations, r)
	}
	for _, p := range inv.Provides {
		d := view.moduleByJar[p.Jar]
		view.Registrations = append(view.Registrations, ServiceRegistration{
			Interface: p.Service, ProviderClass: p.Provider,
			SourceJar: p.Jar, GAV: d.GAV, SourceFile: d.NameSource,
			ProviderFound: p.ProviderFound,
			Mechanism:     MechanismModuleProvides, Module: p.Module,
		})
	}
	return view, warnings
}

// usesCheck reports whether a ServiceLoader caller class may load
// service iface in this scenario: always on the class path; on the
// module path only if the caller's module is automatic or its
// descriptor declares `uses iface` (otherwise ServiceLoader throws
// ServiceConfigurationError). The reason explains a refusal.
func (v ScenarioView) usesCheck(h *hierarchy.Hierarchy, callerFQCN, iface string) (bool, string) {
	if v.Scenario != ScenarioModulePath {
		return true, ""
	}
	ci, ok := h.Lookup(callerFQCN)
	if !ok {
		return true, ""
	}
	d, ok := v.moduleByJar[ci.SourceJar]
	if !ok || d.Kind != ModuleExplicit {
		return true, ""
	}
	for _, u := range d.Info.Uses {
		if u == iface {
			return true, ""
		}
	}
	return false, fmt.Sprintf("module-path scenario: caller module %s does not declare `uses %s`; ServiceLoader would throw ServiceConfigurationError", d.Name, iface)
}
