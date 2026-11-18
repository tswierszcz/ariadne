package component

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"ariadne/internal/input"
	"ariadne/internal/mavenversion"
)

// LIBRARY_MODE LB3 (lb3-1): who supplies each dependency of a library to
// its users. The library's own pom (embedded in its jar, or -library-pom)
// is read with its parents, and the dependency tree a *consumer* of the
// library gets is computed as Maven does it: the library's compile and
// runtime declarations that are not optional, then, transitively, every
// compile/runtime, non-optional declaration of their poms (fetched from
// Central through the run's cache; versions managed in the pom, its
// parents or the BOMs they import), nearest first, a declaration's
// exclusions applying below it. Those dependencies are supplied by the
// library. The library's provided, system and optional declarations —
// and what reaches the tree only through them — are supplied by the user
// (provided dependencies are not transitive; optional ones only exist
// when the user adds them). Test scope is not published. Each declared
// dependency is matched by group:artifact to a dependency instance of
// the input and its version compared with the declared one; an input
// instance the tree does not contain is undeclared.

// Who supplies a declared dependency (DeclaredDependency.SuppliedBy).
const (
	SuppliedByLibrary  = "library"
	SuppliedByUser     = "user"
	SuppliedUndeclared = "undeclared"
)

// Version check of an input instance against the declaration
// (DeclaredDependency.Match).
const (
	MatchDeclared     = "declared"     // the declared exact version
	MatchDiffers      = "differs"      // another version than the declared exact one
	MatchInRange      = "in_range"     // inside the declared range
	MatchOutOfRange   = "out_of_range" // outside the declared range
	MatchUndetermined = "undetermined" // the instance's version cannot be ordered
	MatchUnresolved   = "unresolved"   // the declared version is not known (unresolved property, unread parent)
	MatchAbsent       = "absent"       // declared, but no instance in the input
	// MatchUnidentified: no identified instance, but an unidentified jar
	// of the input is named <artifact>-<version>.jar (a jar without
	// pom.properties, e.g. built by Gradle; -deps-list identifies it).
	MatchUnidentified = "unidentified"
)

// maxLibraryDepth bounds the levels of the consumer tree followed.
const maxLibraryDepth = 6

// DeclaredDependency is one dependency of the library's tree, or an
// input instance the tree does not declare.
type DeclaredDependency struct {
	Package    string `json:"package"` // group:artifact
	SuppliedBy string `json:"supplied_by"`
	// Scope is the dependency's scope in the library's tree (compile,
	// runtime, provided, system); empty when undeclared.
	Scope    string `json:"scope,omitempty"`
	Optional bool   `json:"optional,omitempty"`
	// Via are the dependencies (group:artifact) through which it is
	// declared, nearest the library first; empty for a direct one.
	Via []string `json:"via,omitempty"`
	// Declared is the version as written in the declaring pom ("" when it
	// comes from dependencyManagement); Version the resolved version or
	// range ("" when unresolved).
	Declared string `json:"declared,omitempty"`
	Version  string `json:"version,omitempty"`
	// Instance and InstanceVersion name the matching input instance.
	Instance        string `json:"instance,omitempty"`
	InstanceVersion string `json:"instance_version,omitempty"`
	Match           string `json:"version_match,omitempty"`
}

// LibraryDependencies is what the library's pom declares and how the
// input's instances match it.
type LibraryDependencies struct {
	// POM names where the declarations were read: "<jar>!/<entry>" or the
	// -library-pom file; "" when none was found (Notes say why).
	POM     string `json:"pom"`
	Library string `json:"library,omitempty"` // group:artifact:version of the pom
	// Declared is sorted by package; undeclared input instances are
	// included with SuppliedBy "undeclared".
	Declared []DeclaredDependency `json:"declared,omitempty"`
	Notes    []string             `json:"notes,omitempty"`
}

// ForInstance returns the declaration matched to instance id, or nil.
func (d *LibraryDependencies) ForInstance(id string) *DeclaredDependency {
	if d == nil || id == "" {
		return nil
	}
	for i := range d.Declared {
		if d.Declared[i].Instance == id {
			return &d.Declared[i]
		}
	}
	return nil
}

// Count returns how many declared dependencies have the given supplier
// and, of those, how many have no instance in the input.
func (d *LibraryDependencies) Count(suppliedBy string) (n, absent int) {
	if d == nil {
		return 0, 0
	}
	for _, x := range d.Declared {
		if x.SuppliedBy == suppliedBy {
			n++
			if x.Match == MatchAbsent {
				absent++
			}
		}
	}
	return n, absent
}

// LibraryDeclarations reads the library's pom — pomFile when set, else
// the one embedded in appJar (appName is the input's file name, used to
// choose among several) — and matches the input's instances to the
// consumer tree. It never fails: what cannot be read is a note.
func (inv *Inventory) LibraryDeclarations(ctx context.Context, fetch ReferenceFetcher, base, appJar, appName, pomFile string) *LibraryDependencies {
	if base == "" {
		base = CentralBaseURL
	}
	out := &LibraryDependencies{}
	body, where, err := libraryPOM(appJar, appName, pomFile)
	if err != nil {
		out.Notes = append(out.Notes, err.Error())
		inv.matchUndeclared(out, map[string]bool{})
		return out
	}
	p, err := parsePOM(body)
	if err != nil {
		out.Notes = append(out.Notes, fmt.Sprintf("%s unreadable: %v", where, err))
		inv.matchUndeclared(out, map[string]bool{})
		return out
	}
	out.POM = where
	res := &pomResolver{ctx: ctx, fetch: fetch, base: base, seen: map[string]*pomCtx{}, imports: true}
	root := res.effective(p, 0)
	out.Library = root.props["project.groupId"] + ":" + p.ArtifactID + ":" + root.props["project.version"]
	notes := map[string]bool{}
	for _, n := range root.notes {
		notes[n] = true
	}

	instances := map[string]*Instance{}
	for i := range inv.Instances {
		in := &inv.Instances[i]
		if in.GAV == nil || in.Detection == DetectedApp || in.Detection == DetectedEmbedded || len(in.VersionCandidates) > 0 {
			continue
		}
		key := in.GAV.PackageName()
		if prev, ok := instances[key]; ok {
			notes[fmt.Sprintf("%s: several instances in the input (%s, %s); the first is matched", key, prev.GAV.Version, in.GAV.Version)] = true
			continue
		}
		instances[key] = in
	}

	type node struct {
		ctx      *pomCtx
		via      []string
		scope    string
		optional bool
		user     bool
		excl     []pomExclusion
		depth    int
	}
	seen := map[string]bool{}
	var queue []node
	// visit records the declarations of n's pom and queues their poms.
	visit := func(n node, direct bool, wantUser bool) {
		for _, d := range n.ctx.pom.Dependencies {
			g, a := n.ctx.resolve(d.GroupID), n.ctx.resolve(d.ArtifactID)
			if g == "" || a == "" {
				notes[fmt.Sprintf("a dependency of %s has unresolved coordinates %s:%s", ctxName(n.ctx), d.GroupID, d.ArtifactID)] = true
				continue
			}
			key := g + ":" + a
			scope := strings.TrimSpace(d.Scope)
			if scope == "" {
				scope = n.ctx.managedScope[key]
			}
			if scope == "" {
				scope = "compile"
			}
			optional := strings.TrimSpace(d.Optional) == "true"
			var user bool
			switch {
			case direct:
				switch scope {
				case "compile", "runtime":
					user = optional
				case "provided", "system":
					user = true
				default: // test, import: not published
					continue
				}
				if user != wantUser {
					continue
				}
			default:
				// Below the library only compile/runtime, non-optional
				// declarations are transitive.
				if optional || (scope != "compile" && scope != "runtime") {
					continue
				}
				switch n.scope {
				case "provided", "system":
					scope = n.scope
				case "runtime":
					scope = "runtime"
				}
				optional = n.optional
				user = n.user
			}
			if excluded(n.excl, g, a) || seen[key] {
				continue
			}
			seen[key] = true
			dd := DeclaredDependency{Package: key, Scope: scope, Optional: optional, Via: n.via,
				Declared: strings.TrimSpace(d.Version), Version: n.ctx.version(d, g, a), SuppliedBy: SuppliedByLibrary}
			if user {
				dd.SuppliedBy = SuppliedByUser
			}
			pomVersion := ""
			if in := instances[key]; in != nil {
				dd.Instance, dd.InstanceVersion = in.ID, in.GAV.Version
				dd.Match = versionMatch(dd.Version, in.GAV.Version)
				if dd.Match == MatchDeclared || dd.Match == MatchInRange {
					pomVersion = in.GAV.Version
				}
			} else if id := inv.unidentifiedJar(a, dd.Version); id != "" {
				dd.Instance, dd.Match = id, MatchUnidentified
			} else {
				dd.Match = MatchAbsent
			}
			if pomVersion == "" && dd.Version != "" && !isRange(dd.Version) {
				pomVersion = dd.Version
			}
			out.Declared = append(out.Declared, dd)
			switch {
			case pomVersion == "":
				if dd.Version != "" {
					notes[fmt.Sprintf("%s: range %s with no instance inside it in the input; its dependencies not followed", key, dd.Version)] = true
				} else {
					notes[fmt.Sprintf("%s: version unresolved; its dependencies not followed", key)] = true
				}
				continue
			case n.depth+1 >= maxLibraryDepth:
				notes[fmt.Sprintf("dependencies below depth %d not followed", maxLibraryDepth)] = true
				continue
			}
			gav := input.GAV{Group: g, Artifact: a, Version: pomVersion}
			dc, err := res.load(gav, 0)
			if err != nil {
				notes[fmt.Sprintf("pom of %s unavailable, its dependencies not followed: %v", gav, err)] = true
				continue
			}
			for _, nt := range dc.notes {
				notes[nt] = true
			}
			via := append(append([]string{}, n.via...), key)
			excl := append(append([]pomExclusion{}, n.excl...), d.Exclusions...)
			queue = append(queue, node{ctx: dc, via: via, scope: scope, optional: optional, user: user, excl: excl, depth: n.depth + 1})
		}
	}
	// The consumer tree first (nearest wins within it), then what the
	// user supplies.
	for _, wantUser := range []bool{false, true} {
		visit(node{ctx: root}, true, wantUser)
		for len(queue) > 0 {
			n := queue[0]
			queue = queue[1:]
			visit(n, false, wantUser)
		}
	}
	inv.matchUndeclared(out, seen)
	sort.SliceStable(out.Declared, func(i, j int) bool { return out.Declared[i].Package < out.Declared[j].Package })
	out.Notes = append(out.Notes, sortedKeys(notes)...)
	return out
}

// matchUndeclared adds the input's identified dependency instances the
// tree does not declare.
func (inv *Inventory) matchUndeclared(out *LibraryDependencies, seen map[string]bool) {
	for _, in := range inv.Instances {
		if in.GAV == nil || in.Detection == DetectedApp || in.Detection == DetectedEmbedded || len(in.VersionCandidates) > 0 {
			continue
		}
		key := in.GAV.PackageName()
		if seen[key] {
			continue
		}
		seen[key] = true
		out.Declared = append(out.Declared, DeclaredDependency{Package: key, SuppliedBy: SuppliedUndeclared,
			Instance: in.ID, InstanceVersion: in.GAV.Version})
	}
	sort.SliceStable(out.Declared, func(i, j int) bool { return out.Declared[i].Package < out.Declared[j].Package })
}

// unidentifiedJar returns the unidentified dependency instance whose jar
// is named <artifact>-<version>.jar (or <artifact>-<digit>… when the
// version is a range or unresolved), "" when none or several.
func (inv *Inventory) unidentifiedJar(artifact, version string) string {
	found := ""
	for _, in := range inv.Instances {
		if in.Detection != DetectedUnidentified {
			continue
		}
		name := strings.TrimPrefix(inv.ArtifactLocation(in.ArtifactID), "deps:")
		name = name[strings.LastIndex(name, "/")+1:]
		ok := false
		if version != "" && !isRange(version) {
			ok = name == artifact+"-"+version+".jar"
		} else if rest, cut := strings.CutPrefix(name, artifact+"-"); cut && rest != "" && rest[0] >= '0' && rest[0] <= '9' {
			ok = strings.HasSuffix(rest, ".jar")
		}
		if ok {
			if found != "" {
				return ""
			}
			found = in.ID
		}
	}
	return found
}

// libraryPOM returns the library's pom and where it was read.
func libraryPOM(appJar, appName, pomFile string) ([]byte, string, error) {
	if pomFile != "" {
		b, err := os.ReadFile(pomFile)
		if err != nil {
			return nil, "", fmt.Errorf("-library-pom: %w", err)
		}
		return b, pomFile, nil
	}
	zr, err := zip.OpenReader(appJar)
	if err != nil {
		return nil, "", fmt.Errorf("library pom: %w", err)
	}
	defer zr.Close()
	var poms []*zip.File
	for _, f := range zr.File {
		parts := strings.Split(f.Name, "/")
		if len(parts) == 5 && parts[0] == "META-INF" && parts[1] == "maven" && parts[4] == "pom.xml" {
			poms = append(poms, f)
		}
	}
	jar := appName
	if jar == "" {
		jar = path.Base(appJar)
	}
	if len(poms) > 1 {
		// Several poms (an uber jar, or embedded metadata): the library's
		// is the one whose artifactId names the jar.
		stem := strings.TrimSuffix(jar, ".jar")
		var named []*zip.File
		for _, f := range poms {
			if a := strings.Split(f.Name, "/")[3]; stem == a || strings.HasPrefix(stem, a+"-") {
				named = append(named, f)
			}
		}
		if len(named) != 1 {
			var names []string
			for _, f := range poms {
				names = append(names, f.Name)
			}
			return nil, "", fmt.Errorf("library pom: %s has %d poms (%s) and none is named by the jar; pass -library-pom", jar, len(poms), strings.Join(names, ", "))
		}
		poms = named
	}
	if len(poms) == 0 {
		return nil, "", fmt.Errorf("library pom: %s embeds no META-INF/maven/*/*/pom.xml (pass -library-pom for a library built without one); dependency declarations not read", jar)
	}
	rc, err := poms[0].Open()
	if err != nil {
		return nil, "", fmt.Errorf("library pom: %w", err)
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, "", fmt.Errorf("library pom: %w", err)
	}
	return b, jar + "!/" + poms[0].Name, nil
}

func ctxName(c *pomCtx) string {
	return c.props["project.groupId"] + ":" + c.pom.ArtifactID + ":" + c.props["project.version"]
}

func isRange(v string) bool { return strings.HasPrefix(v, "[") || strings.HasPrefix(v, "(") }

func excluded(excl []pomExclusion, g, a string) bool {
	for _, e := range excl {
		eg, ea := strings.TrimSpace(e.GroupID), strings.TrimSpace(e.ArtifactID)
		if (eg == "*" || eg == g) && (ea == "*" || ea == a) {
			return true
		}
	}
	return false
}

// versionMatch compares an input instance's version with a declared
// version or range.
func versionMatch(declared, have string) string {
	switch {
	case declared == "":
		return MatchUnresolved
	case isRange(declared):
		r, err := mavenversion.ParseRange(declared)
		if err != nil {
			return MatchUnresolved
		}
		switch r.Contains(have) {
		case mavenversion.In:
			return MatchInRange
		case mavenversion.Out:
			return MatchOutOfRange
		}
		return MatchUndetermined
	case mavenversion.Compare(declared, have) == 0:
		return MatchDeclared
	}
	return MatchDiffers
}
