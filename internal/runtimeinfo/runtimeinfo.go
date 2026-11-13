// Package runtimeinfo identifies the Java runtime an application is
// deployed on (CVE:F8a): from the installation's own `release` file
// (-jdk-home) or from a declared description (-runtime-file). It never
// runs the target `java` and never assumes the scanner's own JDK.
//
// Raw values are kept as read. Normalization is deliberately cautious:
// a vendor or distribution is named only for an IMPLEMENTOR value this
// package knows, and a field the input does not establish is left empty
// with a reason in Unresolved — an assessment that needs it is then not
// performed, never guessed.
package runtimeinfo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Version of the identification rules (snapshots and v2 record it).
const Version = "f8a-1"

// Sources of a description.
const (
	SourceJDKHome     = "jdk-home"
	SourceRuntimeFile = "runtime-file"
)

// Runtime is one identified (or partly identified) runtime.
type Runtime struct {
	Source string `json:"source" yaml:"-"`
	// Path is the -jdk-home directory or the -runtime-file path, as given.
	Path string `json:"path" yaml:"-"`
	// Raw holds every key of the release file (or the declared fields),
	// unmodified.
	Raw map[string]string `json:"raw" yaml:"-"`

	Vendor       string `json:"vendor,omitempty" yaml:"vendor"`
	Distribution string `json:"distribution,omitempty" yaml:"distribution"`
	// VersionString is the full runtime version as reported
	// (JAVA_RUNTIME_VERSION, else JAVA_VERSION).
	VersionString string   `json:"version_string,omitempty" yaml:"version"`
	Version       *Release `json:"version,omitempty" yaml:"-"`
	OS            string   `json:"os,omitempty" yaml:"os"`
	Arch          string   `json:"arch,omitempty" yaml:"arch"`
	Libc          string   `json:"libc,omitempty" yaml:"libc"`
	// Package is a system package the runtime was installed from, when
	// the input names one (declared, or a distribution build).
	Package *Package `json:"package,omitempty" yaml:"package"`
	// Unresolved lists what could not be established and why.
	Unresolved []string `json:"unresolved,omitempty" yaml:"-"`
}

// Package is a system package.
type Package struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	// Distro is the package's distribution release, e.g. "Debian:12".
	Distro string `json:"distro,omitempty" yaml:"distro"`
}

// Release is a parsed Java runtime version (JEP 322, or the legacy
// 1.8.0_NNN-bBB form).
type Release struct {
	Feature int `json:"feature"`
	Interim int `json:"interim"`
	Update  int `json:"update"`
	Patch   int `json:"patch"`
	// Build is the +N (or -bNN) build number, 0 when absent.
	Build int `json:"build,omitempty"`
	// Opt is what follows the build ("LTS", "1-Debian", …), unparsed.
	Opt string `json:"opt,omitempty"`
}

// String renders feature.interim.update[.patch][+build].
func (r Release) String() string {
	s := fmt.Sprintf("%d.%d.%d", r.Feature, r.Interim, r.Update)
	if r.Patch > 0 {
		s += fmt.Sprintf(".%d", r.Patch)
	}
	if r.Build > 0 {
		s += fmt.Sprintf("+%d", r.Build)
	}
	return s
}

var (
	jep322Re = regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:\.(\d+))?(?:\.\d+)*(?:-[A-Za-z0-9.]+)?(?:\+(\d+))?(?:-(.+))?$`)
	legacyRe = regexp.MustCompile(`^1\.(\d+)\.0(?:_(\d+))?(?:-b(\d+))?`)
)

// ParseVersion parses a Java runtime version string; ok is false when it
// fits neither form.
func ParseVersion(s string) (Release, bool) {
	s = strings.TrimSpace(s)
	if m := legacyRe.FindStringSubmatch(s); m != nil {
		r := Release{Feature: atoi(m[1]), Update: atoi(m[2]), Build: atoi(m[3])}
		return r, r.Feature > 0
	}
	m := jep322Re.FindStringSubmatch(s)
	if m == nil {
		return Release{}, false
	}
	r := Release{Feature: atoi(m[1]), Interim: atoi(m[2]), Update: atoi(m[3]), Patch: atoi(m[4]), Build: atoi(m[5]), Opt: m[6]}
	return r, r.Feature > 0
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// implementor maps an exact IMPLEMENTOR value to the vendor and, when the
// value alone determines it, the distribution. An IMPLEMENTOR that several
// products share (Oracle builds both Oracle JDK and the jdk.java.net
// OpenJDK builds) names no distribution.
var implementor = map[string]struct{ vendor, distribution string }{
	"Eclipse Adoptium":   {"Eclipse Adoptium", "Temurin"},
	"Amazon.com Inc.":    {"Amazon", "Corretto"},
	"Azul Systems, Inc.": {"Azul", "Zulu"},
	"BellSoft":           {"BellSoft", "Liberica"},
	"Microsoft":          {"Microsoft", "Microsoft Build of OpenJDK"},
	"SAP SE":             {"SAP", "SapMachine"},
	"Red Hat, Inc.":      {"Red Hat", ""},
	"Oracle Corporation": {"Oracle", ""},
	"Debian":             {"Debian", "OpenJDK package"},
	"Ubuntu":             {"Ubuntu", "OpenJDK package"},
	"Eclipse Foundation": {"Eclipse Foundation", ""},
	"International Business Machines Corporation": {"IBM", ""},
}

// FromJDKHome reads dir/release. It reads nothing else and runs nothing.
func FromJDKHome(dir string) (Runtime, error) {
	f, err := os.Open(filepath.Join(dir, "release"))
	if err != nil {
		return Runtime{}, fmt.Errorf("-jdk-home %s: %w (a JDK/JRE installation has a release file at its root)", dir, err)
	}
	defer f.Close()
	raw := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		raw[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"`)
	}
	if err := sc.Err(); err != nil {
		return Runtime{}, fmt.Errorf("-jdk-home %s: reading release: %w", dir, err)
	}
	r := Runtime{Source: SourceJDKHome, Path: dir, Raw: raw}
	r.VersionString = raw["JAVA_RUNTIME_VERSION"]
	if r.VersionString == "" {
		r.VersionString = raw["JAVA_VERSION"]
	}
	r.OS, r.Arch, r.Libc = raw["OS_NAME"], raw["OS_ARCH"], raw["LIBC"]
	switch impl := raw["IMPLEMENTOR"]; {
	case impl == "" || impl == "N/A":
		r.Unresolved = append(r.Unresolved, "vendor: the release file has no IMPLEMENTOR")
	default:
		if known, ok := implementor[impl]; ok {
			r.Vendor, r.Distribution = known.vendor, known.distribution
			if known.distribution == "" {
				r.Unresolved = append(r.Unresolved, fmt.Sprintf("distribution: IMPLEMENTOR %q is shared by several products", impl))
			}
		} else {
			r.Unresolved = append(r.Unresolved, fmt.Sprintf("vendor: IMPLEMENTOR %q is not a known vendor", impl))
		}
	}
	// A distribution package's own version is in the runtime version
	// (25.0.5+8-1-Debian → package version 25.0.5+8-1); its distribution
	// release (Debian 12, 13…) is not in the release file.
	if r.Vendor == "Debian" || r.Vendor == "Ubuntu" {
		if rel, ok := ParseVersion(r.VersionString); ok {
			pv := strings.TrimSuffix(r.VersionString, "-"+r.Vendor)
			r.Package = &Package{Name: fmt.Sprintf("openjdk-%d", rel.Feature), Version: pv}
			r.Unresolved = append(r.Unresolved, "package distro: the release file does not name the "+r.Vendor+" release (declare it with -runtime-file)")
		}
	}
	r.finish()
	return r, nil
}

// FromFile reads a declared runtime description (YAML): vendor,
// distribution, version, os, arch, libc, package{name, version, distro}.
// Unknown fields are an error.
func FromFile(path string) (Runtime, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Runtime{}, fmt.Errorf("-runtime-file: %w", err)
	}
	var r Runtime
	dec := yaml.NewDecoder(strings.NewReader(string(b)))
	dec.KnownFields(true)
	if err := dec.Decode(&r); err != nil {
		return Runtime{}, fmt.Errorf("-runtime-file %s: %w", path, err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(b, &raw); err == nil {
		r.Raw = map[string]string{}
		for k, v := range raw {
			r.Raw[k] = fmt.Sprint(v)
		}
	}
	r.Source, r.Path = SourceRuntimeFile, path
	if r.Vendor == "" {
		r.Unresolved = append(r.Unresolved, "vendor: not declared")
	}
	if r.Distribution == "" {
		r.Unresolved = append(r.Unresolved, "distribution: not declared")
	}
	r.finish()
	return r, nil
}

func (r *Runtime) finish() {
	if r.VersionString == "" {
		r.Unresolved = append(r.Unresolved, "version: none given")
	} else if v, ok := ParseVersion(r.VersionString); ok {
		r.Version = &v
		if v.Build == 0 {
			r.Unresolved = append(r.Unresolved, "build: the version has no build number")
		}
	} else {
		r.Unresolved = append(r.Unresolved, fmt.Sprintf("version: %q is not a Java runtime version", r.VersionString))
	}
	if r.OS == "" {
		r.Unresolved = append(r.Unresolved, "os: not given")
	}
	if r.Arch == "" {
		r.Unresolved = append(r.Unresolved, "arch: not given")
	}
	sort.Strings(r.Unresolved)
}

// Describe is a one-line description for the report and the log.
func (r Runtime) Describe() string {
	var parts []string
	name := strings.TrimSpace(r.Vendor + " " + r.Distribution)
	if name == "" {
		name = "unknown vendor"
	}
	parts = append(parts, name)
	if r.VersionString != "" {
		parts = append(parts, r.VersionString)
	}
	if r.OS != "" || r.Arch != "" {
		parts = append(parts, strings.TrimSpace(r.OS+" "+r.Arch))
	}
	if r.Package != nil {
		p := "package " + r.Package.Name + " " + r.Package.Version
		if r.Package.Distro != "" {
			p += " (" + r.Package.Distro + ")"
		}
		parts = append(parts, p)
	}
	return strings.Join(parts, ", ") + " — from " + r.Source + " " + r.Path
}
