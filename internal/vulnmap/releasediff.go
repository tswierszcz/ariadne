package vulnmap

// CVE:F10 part C: a mapping from the released bytecode when an advisory
// has no fix commit — the methods whose code changed between the last
// release before the fix and the fixing release, restricted to the
// classes the advisory itself names. A heuristic, labelled as such
// (source "release-diff", unreviewed): a release carries other changes
// too, and an advisory that names no class gets nothing.

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"ariadne/internal/classfile"
	"ariadne/internal/cves"
	"ariadne/internal/httpcache"
	"ariadne/internal/mavenversion"
)

// SourceReleaseDiff marks a mapping read from the bytecode of two
// releases (CVE:F10).
const SourceReleaseDiff = "release-diff"

// ReleaseDiff is one component's evidence for a release-diff mapping.
type ReleaseDiff struct {
	Package string `json:"package"` // group:artifact
	Before  string `json:"before"`  // the last release before Fixed
	Fixed   string `json:"fixed"`   // the first release with the fix after the analyzed version
	// Named are the advisory's identifiers that name a class of the
	// component; Classes the classes they matched.
	Named   []string `json:"named"`
	Classes []string `json:"classes"`
	// Changed/Removed count the selectors taken (methods whose code
	// differs / that the fixed release no longer declares).
	Changed int `json:"changed"`
	Removed int `json:"removed"`
	// Skipped says why this component gave nothing (no fixed release
	// after the analyzed version, no earlier release, jar unavailable,
	// no named class in it).
	Skipped string `json:"skipped,omitempty"`
}

// releaseDiffMaxMethods bounds one mapping: a named class rewritten in
// the release says little about which method was the vulnerability.
const releaseDiffMaxMethods = 40

var camelRe = regexp.MustCompile(`\b[A-Z][a-z0-9]+(?:[A-Z][A-Za-z0-9]*)+\b`)

// codeRe matches an identifier marked as code (`Name`, or the last
// segment of `a.b.Name`).
var codeRe = regexp.MustCompile("`(?:[a-z][\\w]*\\.)*([A-Z][A-Za-z0-9]+)(?:[#.(][^`]*)?`")

// advisoryClassNames returns the identifiers of an advisory's text that
// may be simple class names: CamelCase words with at least two humps (so
// "Spring" or "Java" never count), and any capitalized identifier the
// text marks as code with backticks.
func advisoryClassNames(cve cves.CVE) []string {
	seen := map[string]bool{}
	var out []string
	text := cve.Summary + "\n" + cve.Details
	words := camelRe.FindAllString(text, -1)
	for _, m := range codeRe.FindAllStringSubmatch(text, -1) {
		words = append(words, m[1])
	}
	for _, w := range words {
		if !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	sort.Strings(out)
	return out
}

// fixedAfter returns the smallest "fixed" event of pkg's OSV ranges
// greater than version (the release a user of version upgrades to).
func fixedAfter(cve cves.CVE, pkg, version string) string {
	best := ""
	for _, s := range cve.Sources {
		if s.Withdrawn != "" {
			continue
		}
		for _, r := range s.Ranges {
			if r.Package != pkg {
				continue
			}
			for _, ev := range r.Events {
				f := ev["fixed"]
				if f == "" || mavenversion.Compare(f, version) <= 0 {
					continue
				}
				if best == "" || mavenversion.Compare(f, best) < 0 {
					best = f
				}
			}
		}
	}
	return best
}

func centralPath(pkg string) (string, string, bool) {
	g, a, ok := strings.Cut(pkg, ":")
	if !ok || g == "" || a == "" {
		return "", "", false
	}
	return strings.ReplaceAll(g, ".", "/") + "/" + a, a, true
}

// releaseBefore returns the published name of fixed (OSV writes "2.2.10"
// for Maven's "2.2.10.Final") and the highest final release of pkg below
// it in the same major.minor line, from Maven Central's metadata. A fix
// first released in a new minor line has no such predecessor: comparing
// across a minor release would take every change of that release.
func (r *Resolver) releaseBefore(ctx context.Context, pkg, fixed string) (string, string, error) {
	path, _, ok := centralPath(pkg)
	if !ok {
		return "", "", fmt.Errorf("package %q is not group:artifact", pkg)
	}
	body, err := r.fetchArtifact(ctx, path+"/maven-metadata.xml")
	if err != nil {
		return "", "", err
	}
	var md struct {
		Versions []string `xml:"versioning>versions>version"`
	}
	if err := xml.Unmarshal(body, &md); err != nil {
		return "", "", fmt.Errorf("maven-metadata.xml: %w", err)
	}
	published := ""
	for _, v := range md.Versions {
		if mavenversion.Compare(v, fixed) == 0 && (published == "" || v == fixed) {
			published = v
		}
	}
	if published == "" {
		return "", "", fmt.Errorf("fixed release %s of %s not published", fixed, pkg)
	}
	line := minorLine(published)
	best := ""
	pre := preRelease(published)
	for _, v := range md.Versions {
		if !mavenversion.Regular(v) || mavenversion.Compare(v, published) >= 0 || (preRelease(v) && !pre) || minorLine(v) != line {
			continue
		}
		if best == "" || mavenversion.Compare(v, best) > 0 {
			best = v
		}
	}
	if best == "" {
		return published, "", fmt.Errorf("no earlier release of %s in the %s line: the fix came with a new minor release", pkg, line)
	}
	return published, best, nil
}

var minorLineRe = regexp.MustCompile(`^(\d+)\.(\d+)`)

// minorLine is a version's "major.minor" ("" when it has no such prefix).
func minorLine(v string) string {
	m := minorLineRe.FindStringSubmatch(v)
	if m == nil {
		return ""
	}
	return m[1] + "." + m[2]
}

// centralGate serializes Maven Central requests of the release-diff
// source and spaces them out: Central rate-limits clients that fetch many
// artifacts in parallel (HTTP 429), and a cached or local artifact never
// reaches it.
var centralGate = struct {
	sync.Mutex
	last time.Time
}{}

// centralDelay is the minimum time between two Central requests.
var centralDelay = 750 * time.Millisecond

// fetchArtifact returns an artifact file (path relative to the
// repository root) from Maven Central through the cache — so a snapshot
// records it like any other lookup — one request at a time. Offline, a
// jar the cache lacks is read from the local Maven repository instead;
// a snapshot replay never falls back (it must not depend on the machine).
func (r *Resolver) fetchArtifact(ctx context.Context, rel string) ([]byte, error) {
	b, err := r.fetchCentral(ctx, rel)
	if err != nil && errors.Is(err, httpcache.ErrOffline) && strings.HasSuffix(rel, ".jar") {
		if dir := r.localMavenRepo(); dir != "" {
			if lb, lerr := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel))); lerr == nil {
				noteEvidence(ctx, lb)
				return lb, nil
			}
		}
	}
	return b, err
}

func (r *Resolver) fetchCentral(ctx context.Context, rel string) ([]byte, error) {
	url := r.centralBase() + "/" + rel
	hdr := http.Header{"User-Agent": []string{"ariadne (CVE reachability analysis)"}}
	if r.Cache != nil {
		// A cached artifact is served by the cache's own policy: no need
		// to wait for the Central gate.
		if _, ok, _ := r.Cache.Get(httpcache.Key("GET", url)); ok {
			return r.fetch(ctx, url, hdr)
		}
	}
	centralGate.Lock()
	defer centralGate.Unlock()
	if wait := centralDelay - time.Since(centralGate.last); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	defer func() { centralGate.last = time.Now() }()
	return r.fetch(ctx, url, hdr)
}

// localMavenRepo is LocalMavenRepo, or ~/.m2/repository when unset; ""
// when neither exists.
func (r *Resolver) localMavenRepo() string {
	if r.LocalMavenRepo != "" {
		return r.LocalMavenRepo
	}
	if home, err := os.UserHomeDir(); err == nil {
		dir := filepath.Join(home, ".m2", "repository")
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir
		}
	}
	return ""
}

var preReleaseRe = regexp.MustCompile(`(?i)[-.](alpha|beta|a\d|b\d|m\d+|milestone|rc\d*|cr\d*|snapshot|ea|preview|pr\d*)\b`)

// preRelease reports a milestone/release-candidate/snapshot version:
// never "the release before the fix" of a final release.
func preRelease(v string) bool { return preReleaseRe.MatchString(v) }

func (r *Resolver) centralBase() string {
	if r.CentralBaseURL != "" {
		return r.CentralBaseURL
	}
	return "https://repo1.maven.org/maven2"
}

// releaseClasses reads the classes of pkg@version whose simple name (the
// part after the last '.' or '$') is one of names.
func (r *Resolver) releaseClasses(ctx context.Context, pkg, version string, names map[string]bool) (map[string]*classfile.ClassFile, error) {
	path, artifact, ok := centralPath(pkg)
	if !ok {
		return nil, fmt.Errorf("package %q is not group:artifact", pkg)
	}
	body, err := r.fetchArtifact(ctx, fmt.Sprintf("%s/%s/%s-%s.jar", path, version, artifact, version))
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}
	out := map[string]*classfile.ClassFile{}
	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".class") || strings.HasPrefix(f.Name, "META-INF/") {
			continue
		}
		simple := strings.TrimSuffix(f.Name[strings.LastIndex(f.Name, "/")+1:], ".class")
		if i := strings.LastIndex(simple, "$"); i >= 0 {
			simple = simple[i+1:]
		}
		if !names[simple] {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		cf, err := classfile.Parse(io.LimitReader(rc, 16<<20))
		rc.Close()
		if err != nil {
			continue
		}
		out[cf.ThisClass] = cf
	}
	return out, nil
}

func methodPrints(cf *classfile.ClassFile) map[string]string {
	out := map[string]string{}
	for _, m := range cf.Methods {
		fp, err := cf.MethodFingerprint(m)
		if err != nil {
			fp = "unreadable"
		}
		out[m.Name+m.Descriptor] = fp
	}
	return out
}

// resolveFromReleaseDiff maps cve from the bytecode of its components'
// releases. Only a contract error is returned.
func (r *Resolver) resolveFromReleaseDiff(ctx context.Context, cve cves.CVE) ([]string, []ReleaseDiff, error) {
	nameList := advisoryClassNames(cve)
	if len(nameList) == 0 {
		return nil, nil, nil
	}
	names := map[string]bool{}
	for _, n := range nameList {
		names[n] = true
	}
	seen := map[string]bool{}
	var methods []string
	var diffs []ReleaseDiff
	done := map[string]bool{}
	for _, a := range cve.Affected {
		if a.Version == "" || done[a.Package+"@"+a.Version] {
			continue
		}
		done[a.Package+"@"+a.Version] = true
		rd := ReleaseDiff{Package: a.Package}
		rd.Fixed = fixedAfter(cve, a.Package, a.Version)
		if rd.Fixed == "" {
			rd.Skipped = "no fixed release after " + a.Version
			diffs = append(diffs, rd)
			continue
		}
		published, before, err := r.releaseBefore(ctx, a.Package, rd.Fixed)
		if published != "" {
			rd.Fixed = published
		}
		if err != nil {
			if isContractError(err) {
				return nil, nil, err
			}
			rd.Skipped = err.Error()
			diffs = append(diffs, rd)
			continue
		}
		rd.Before = before
		oldCls, err := r.releaseClasses(ctx, a.Package, before, names)
		if err == nil {
			var newCls map[string]*classfile.ClassFile
			newCls, err = r.releaseClasses(ctx, a.Package, rd.Fixed, names)
			if err == nil {
				var taken []string
				named := map[string]bool{}
				for cls, oc := range oldCls {
					var regular, init []string
					rd.Classes = append(rd.Classes, cls)
					simple := cls[strings.LastIndexAny(cls, ".$")+1:]
					named[simple] = true
					newPrints := map[string]string{}
					if nc, ok := newCls[cls]; ok {
						newPrints = methodPrints(nc)
					}
					for sig, fp := range methodPrints(oc) {
						nfp, present := newPrints[sig]
						if present && nfp == fp {
							continue
						}
						if !present {
							rd.Removed++
						} else {
							rd.Changed++
						}
						name, desc, _ := strings.Cut(sig, "(")
						sel := cls + "#" + name + "(" + desc
						if name == "<init>" || name == "<clinit>" {
							init = append(init, sel)
						} else {
							regular = append(regular, sel)
						}
					}
					// A constructor or static initializer that changed next
					// to ordinary methods is usually a new field or constant
					// of the fix, not the vulnerable code — and it is
					// reachable whenever the class is used. It counts only
					// when nothing else in the class changed.
					if len(regular) > 0 {
						taken = append(taken, regular...)
					} else {
						taken = append(taken, init...)
					}
				}
				for n := range named {
					rd.Named = append(rd.Named, n)
				}
				sort.Strings(rd.Named)
				sort.Strings(rd.Classes)
				sort.Strings(taken)
				switch {
				case len(rd.Classes) == 0:
					rd.Skipped = "no class named by the advisory in " + a.Package
				case len(taken) == 0:
					rd.Skipped = "named classes unchanged between " + before + " and " + rd.Fixed
				case len(taken) > releaseDiffMaxMethods:
					rd.Skipped = fmt.Sprintf("%d methods changed in the named classes: too broad to say which is vulnerable", len(taken))
					rd.Changed, rd.Removed = 0, 0
				default:
					for _, s := range taken {
						if !seen[s] {
							seen[s] = true
							methods = append(methods, s)
						}
					}
				}
			}
		}
		if err != nil {
			if isContractError(err) {
				return nil, nil, err
			}
			rd.Skipped = "jar unavailable: " + err.Error()
		}
		diffs = append(diffs, rd)
	}
	sort.Strings(methods)
	return methods, diffs, nil
}
