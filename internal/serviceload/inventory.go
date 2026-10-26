package serviceload

import (
	"fmt"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

// ServiceRegistration is one provider class declared, in one
// META-INF/services/<Interface> file, to implement that service
// interface — a fact read from a jar's own metadata. It is not a
// call-graph edge: a registration exists whether or not anything in the
// analyzed app ever calls ServiceLoader.load for this interface, and
// whether or not the provider is ever actually instantiated. Those are
// SERVICE_LOADING S1 (FindLoadSites — which call sites name this
// interface) and S2 (whether a supported, linked consumption can reach
// a constructible provider) respectively, neither of which this phase
// (S0) attempts.
type ServiceRegistration struct {
	Interface     string // the service interface FQCN
	ProviderClass string // the FQCN this registration names
	SourceJar     string
	GAV           *input.GAV // nil for the app, or an unidentified dependency jar
	SourceFile    string     // "META-INF/services/" + Interface
	SourceSHA256  string     // content hash of the service file itself (not just the provider name)
	Line          int
	// ProviderFound is whether ProviderClass actually exists anywhere in
	// the analyzed hierarchy (app + resolved dependency jars) — not a
	// claim that it's reachable or ever instantiated, just that the
	// class the file names is part of what ariadne read at all.
	ProviderFound bool
}

// LoaderKind distinguishes which java.util.ServiceLoader factory a call
// site used — SERVICE_LOADING S1's own explicit requirement ("odróżnić
// classpath, system loader i jawnie podany loader. Nie przyjmować bez
// dowodu, że dowolny loader widzi wszystkie artefakty"): a caller-named
// ClassLoader or a ModuleLayer is not guaranteed to see the same
// artifacts as the thread context class loader every registration in
// ServiceRegistration was found under.
type LoaderKind string

const (
	LoaderClasspath   LoaderKind = "classpath"    // load(Class) — the thread context class loader
	LoaderExplicit    LoaderKind = "explicit"     // load(Class, ClassLoader) — a caller-named loader
	LoaderModuleLayer LoaderKind = "module-layer" // load(ModuleLayer, Class)
	LoaderInstalled   LoaderKind = "installed"    // loadInstalled(Class) — the extension class loader
)

// ServiceLoadSite is SERVICE_LOADING S1's own concept: one app call to
// ServiceLoader.load(...)/loadInstalled(...), with the service interface
// it names (when resolvable), which loader it used, and the calling
// method. S0 defined this type as a stable contract; FindLoadSites (S1)
// is what actually constructs it, using DATA_FLOW D0-D2's local-value
// analysis to resolve the Class argument at each call site.
type ServiceLoadSite struct {
	// Interface is the service interface's FQCN, resolved from the
	// call's Class argument — set only when that argument was a
	// provably exact Class literal (valueflow.KindConst). "" otherwise;
	// see UnresolvedReason for why.
	Interface    string
	CallerMethod string // "fqcn#name(desc)", for display — see CallerFQCN/CallerName/CallerDesc for the same identity split for programmatic use (SERVICE_LOADING S2's exact call-site lookup needs the parts, not the joined string)
	CallerFQCN   string
	CallerName   string
	CallerDesc   string
	LoaderKind   LoaderKind
	// PC is the invoke* instruction's bytecode offset within
	// CallerMethod — evidence for which call site this is, same
	// reasoning as JVM_GRAPH G0's own Call.PC.
	PC int
	// UnresolvedReason explains why Interface is "" — the argument's
	// value.Kind (e.g. "the argument is KindTyped: known type, no known
	// constant" for an unresolvable variable) or, if the whole method's
	// local-value analysis gave up, that reason instead. Empty when
	// Interface is resolved.
	UnresolvedReason string
}

// ServiceConsumptionSite is one exact iterator/stream/provider API call
// with the load-site identity recovered from bounded local call provenance.
// Multiple interfaces are possible at a control-flow join and are retained
// as a union; an empty set with UnresolvedReason is an explicit gap.
type ServiceConsumptionSite struct {
	CallerFQCN, CallerName, CallerDesc string
	PC                                 int
	CallOwner, CallName, CallDesc      string
	ProviderMethodCall                 bool
	Interfaces                         []string
	LoaderKinds                        []LoaderKind
	UnresolvedReason                   string
}

// ProviderMaterialization is SERVICE_LOADING S2's resolved candidate:
// it records whether a provider can be constructed for a service
// interface. It becomes a call-graph target only when a supported
// consumption site is linked to a load site for that interface. S0
// defined this type as a stable contract; FindMaterializations builds it.
type ProviderMaterialization struct {
	Registration ServiceRegistration
	// ConstructorMethod is "fqcn#<init>()V" when the provider has a
	// public no-argument constructor — the only shape the legacy
	// META-INF/services mechanism actually instantiates (the Java 9+
	// module-provider "public static provider()" factory method is a
	// different mechanism, SERVICE_LOADING S3's own JPMS scope, not
	// this one). "" when Gap explains why no target could be modeled.
	ConstructorMethod string
	// Gap explains why ConstructorMethod is "" — never a silent drop
	// (S2's own requirement: "brak implementacji daje lukę, nie brak
	// podatności").
	Gap string
}

// Inventory scans every META-INF/services/* file in the app and every
// dependency jar, parses each one (Parse), and cross-references each
// declared provider class against h. It never creates a call-graph edge
// or any ServiceLoadSite/ProviderMaterialization — just the declarative
// inventory and diagnostics S0's own "Gotowe" asks for ("poprawna lista
// kandydatów i diagnostyka bez tworzenia fikcyjnych wywołań").
//
// warnings accumulates every Parse warning (prefixed with the source
// jar/file) plus one warning per registration whose provider class isn't
// found anywhere in the analyzed hierarchy. It never errors: a jar that
// can't be opened is itself a real error (propagated), but a malformed
// service file inside an openable jar is a diagnostic, not a failure.
func Inventory(h *hierarchy.Hierarchy, appJarPath string, depJars []input.DepJar) ([]ServiceRegistration, []string, error) {
	type jarRef struct {
		path string
		gav  *input.GAV
	}
	jars := make([]jarRef, 0, len(depJars)+1)
	jars = append(jars, jarRef{path: appJarPath})
	for _, dj := range depJars {
		jars = append(jars, jarRef{path: dj.Path, gav: dj.GAV})
	}

	var regs []ServiceRegistration
	var warnings []string
	for _, j := range jars {
		files, err := input.ListServiceFiles(j.path)
		if err != nil {
			return nil, warnings, fmt.Errorf("listing service files in %s: %w", j.path, err)
		}
		for _, f := range files {
			parsed := Parse(f.Interface, f.Content)
			for _, w := range parsed.Warnings {
				warnings = append(warnings, fmt.Sprintf("%s (%s): %s", j.path, f.Interface, w))
			}
			for _, e := range parsed.Entries {
				if !e.Valid || e.Duplicate {
					continue
				}
				_, found := h.Lookup(e.ProviderClass)
				if !found {
					warnings = append(warnings, fmt.Sprintf(
						"%s (%s): provider %q (line %d) not found in the analyzed classes/dependencies",
						j.path, f.Interface, e.ProviderClass, e.Line))
				}
				regs = append(regs, ServiceRegistration{
					Interface:     f.Interface,
					ProviderClass: e.ProviderClass,
					SourceJar:     j.path,
					GAV:           j.gav,
					SourceFile:    "META-INF/services/" + f.Interface,
					SourceSHA256:  f.SHA256,
					Line:          e.Line,
					ProviderFound: found,
				})
			}
		}
	}
	return regs, warnings, nil
}
