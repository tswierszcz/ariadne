package component

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"

	"ariadne/internal/input"
)

// CVE:F6b: embedded candidates named by the host jar's build
// configuration rather than by a pom.properties of their own — a shading
// tool that copies a dependency does not always copy its metadata
// (jruby-core embeds ASM without one). The host's embedded pom.xml is
// read for the maven-shade-plugin <artifactSet><includes>, and every
// dependency of that pom (compile or runtime scope, not optional) — and,
// transitively, of the dependencies' own poms on Central — that an
// include pattern names and no exclude removes is a candidate. A version
// written as a property is resolved from the pom's properties and its
// parents' (fetched from Central), or from dependencyManagement. The
// build configuration describes the build, not the result: a candidate
// is only a claim, verified by its classes like every other one. A shade
// configuration without explicit includes (shade everything) yields no
// candidates — it would name every dependency — and is noted.

// newPOMDecoder is a lenient XML decoder that also reads poms declared
// as ISO-8859-1 / US-ASCII (common in older poms), not only UTF-8.
func newPOMDecoder(r io.Reader) *xml.Decoder {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	dec.CharsetReader = func(label string, in io.Reader) (io.Reader, error) {
		switch strings.ToLower(label) {
		case "iso-8859-1", "iso8859-1", "latin1", "latin-1", "us-ascii", "ascii", "windows-1252", "cp1252":
			b, err := io.ReadAll(in)
			if err != nil {
				return nil, err
			}
			rs := make([]rune, len(b))
			for i, c := range b {
				rs[i] = rune(c)
			}
			return strings.NewReader(string(rs)), nil
		}
		return nil, fmt.Errorf("pom encoding %q not supported", label)
	}
	return dec
}

// maxPOMDepth bounds parent chains and transitive dependency levels.
const maxPOMDepth = 4

// POMURL is the Central URL of g's pom.
func POMURL(base string, g input.GAV) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s-%s.pom", base, strings.ReplaceAll(g.Group, ".", "/"), g.Artifact, g.Version, g.Artifact, g.Version)
}

type pomDep struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
	Type       string `xml:"type"`
	Classifier string `xml:"classifier"`
	Optional   string `xml:"optional"`
}

type pomProp struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type pomFile struct {
	Parent struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"parent"`
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Properties struct {
		Items []pomProp `xml:",any"`
	} `xml:"properties"`
	Dependencies []pomDep `xml:"dependencies>dependency"`
	Managed      []pomDep `xml:"dependencyManagement>dependencies>dependency"`
}

// pomCtx is a pom with its effective properties and managed versions
// (its parent chain's, overridden by its own).
type pomCtx struct {
	pom     *pomFile
	props   map[string]string
	managed map[string]string // group:artifact -> raw version
	notes   []string
}

func parsePOM(b []byte) (*pomFile, error) {
	var p pomFile
	dec := newPOMDecoder(bytes.NewReader(b))
	if err := dec.Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// pomResolver loads poms from Central (through the run's cache) and
// builds their effective contexts.
type pomResolver struct {
	ctx   context.Context
	fetch ReferenceFetcher
	base  string
	seen  map[string]*pomCtx
}

func (r *pomResolver) load(g input.GAV, depth int) (*pomCtx, error) {
	key := g.String()
	if c, ok := r.seen[key]; ok {
		if c == nil {
			return nil, fmt.Errorf("pom %s unavailable", key)
		}
		return c, nil
	}
	r.seen[key] = nil
	b, err := r.fetch(r.ctx, POMURL(r.base, g))
	if err != nil {
		return nil, err
	}
	p, err := parsePOM(b)
	if err != nil {
		return nil, fmt.Errorf("pom %s: %w", key, err)
	}
	c := r.effective(p, depth)
	r.seen[key] = c
	return c, nil
}

// effective merges p over its parent chain (fetched while depth allows).
func (r *pomResolver) effective(p *pomFile, depth int) *pomCtx {
	c := &pomCtx{pom: p, props: map[string]string{}, managed: map[string]string{}}
	if p.Parent.ArtifactID != "" && p.Parent.Version != "" {
		pg := input.GAV{Group: p.Parent.GroupID, Artifact: p.Parent.ArtifactID, Version: p.Parent.Version}
		if depth >= maxPOMDepth {
			c.notes = append(c.notes, fmt.Sprintf("parent %s not read (depth limit %d)", pg, maxPOMDepth))
		} else if pc, err := r.load(pg, depth+1); err != nil {
			c.notes = append(c.notes, fmt.Sprintf("parent %s unavailable: %v", pg, err))
		} else {
			for k, v := range pc.props {
				c.props[k] = v
			}
			for k, v := range pc.managed {
				c.managed[k] = v
			}
			c.notes = append(c.notes, pc.notes...)
		}
	}
	for _, pp := range p.Properties.Items {
		c.props[pp.XMLName.Local] = strings.TrimSpace(pp.Value)
	}
	group, version := p.GroupID, p.Version
	if group == "" {
		group = p.Parent.GroupID
	}
	if version == "" {
		version = p.Parent.Version
	}
	for _, k := range []string{"project.groupId", "pom.groupId", "groupId"} {
		c.props[k] = group
	}
	for _, k := range []string{"project.version", "pom.version", "version"} {
		c.props[k] = version
	}
	c.props["project.artifactId"] = p.ArtifactID
	c.props["project.parent.version"] = p.Parent.Version
	c.props["project.parent.groupId"] = p.Parent.GroupID
	for _, d := range p.Managed {
		if d.Scope == "import" {
			continue
		}
		c.managed[c.resolve(d.GroupID)+":"+c.resolve(d.ArtifactID)] = d.Version
	}
	return c
}

// resolve substitutes ${...} properties; an unresolved one leaves "".
func (c *pomCtx) resolve(s string) string {
	s = strings.TrimSpace(s)
	for i := 0; i < 8 && strings.Contains(s, "${"); i++ {
		a := strings.Index(s, "${")
		b := strings.Index(s[a:], "}")
		if b < 0 {
			return ""
		}
		v, ok := c.props[s[a+2:a+b]]
		if !ok {
			return ""
		}
		s = s[:a] + v + s[a+b+1:]
	}
	if strings.Contains(s, "${") {
		return ""
	}
	return s
}

// version is d's resolved version: its own, else the managed one.
func (c *pomCtx) version(d pomDep, g, a string) string {
	if d.Version != "" {
		return c.resolve(d.Version)
	}
	if m, ok := c.managed[g+":"+a]; ok {
		return c.resolve(m)
	}
	return ""
}

// shadeArtifactSet returns the <artifactSet> include and exclude
// patterns of every maven-shade-plugin in the pom, and whether a shade
// plugin was seen at all.
func shadeArtifactSet(r io.Reader) (includes, excludes []string, shade bool, err error) {
	dec := newPOMDecoder(r)
	type plugin struct {
		artifact   string
		incs, excs []string
	}
	var (
		stack   []string
		plugins []*plugin
		text    strings.Builder
	)
	for {
		tok, terr := dec.Token()
		if terr == io.EOF {
			break
		}
		if terr != nil {
			return nil, nil, false, terr
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			text.Reset()
			if t.Name.Local == "plugin" {
				plugins = append(plugins, &plugin{})
			}
		case xml.CharData:
			text.Write(t)
		case xml.EndElement:
			name := t.Name.Local
			at := func(i int) string {
				if len(stack) > i {
					return stack[len(stack)-1-i]
				}
				return ""
			}
			v := strings.TrimSpace(text.String())
			switch {
			case name == "artifactId" && at(1) == "plugin" && len(plugins) > 0:
				plugins[len(plugins)-1].artifact = v
			case name == "include" && at(1) == "includes" && at(2) == "artifactSet" && len(plugins) > 0:
				plugins[len(plugins)-1].incs = append(plugins[len(plugins)-1].incs, v)
			case name == "excludes" && at(1) == "artifactSet" && len(plugins) > 0 && v != "":
				// <excludes>g:a</excludes> written without <exclude> children
				plugins[len(plugins)-1].excs = append(plugins[len(plugins)-1].excs, v)
			case name == "exclude" && at(1) == "excludes" && at(2) == "artifactSet" && len(plugins) > 0:
				plugins[len(plugins)-1].excs = append(plugins[len(plugins)-1].excs, v)
			case name == "plugin" && len(plugins) > 0:
				p := plugins[len(plugins)-1]
				plugins = plugins[:len(plugins)-1]
				if p.artifact == "maven-shade-plugin" {
					shade = true
					includes = append(includes, p.incs...)
					excludes = append(excludes, p.excs...)
				}
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			text.Reset()
		}
	}
	return includes, excludes, shade, nil
}

// artifactPatternMatch: a shade artifactSet pattern "group:artifact"
// (either part may use * wildcards; extra :type/:classifier parts are
// matched loosely by their first two parts).
func artifactPatternMatch(pattern, group, artifact string) bool {
	p := strings.Split(strings.TrimSpace(pattern), ":")
	if len(p) < 2 {
		ok, _ := path.Match(p[0], group)
		return ok
	}
	g, _ := path.Match(p[0], group)
	a, _ := path.Match(p[1], artifact)
	return g && a
}

func matchesAny(patterns []string, group, artifact string) bool {
	for _, p := range patterns {
		if artifactPatternMatch(p, group, artifact) {
			return true
		}
	}
	return false
}

// shaded: a dependency the shade plugin copies into the jar.
func shadedScope(d pomDep) bool {
	switch strings.TrimSpace(d.Scope) {
	case "", "compile", "runtime":
	default:
		return false
	}
	if strings.TrimSpace(d.Optional) == "true" || d.Classifier != "" {
		return false
	}
	return d.Type == "" || d.Type == "jar" || d.Type == "bundle"
}

// addBuildConfigCandidates adds the host's build-config candidates.
func (inv *Inventory) addBuildConfigCandidates(ctx context.Context, fetch ReferenceFetcher, base string, host Instance, files []*zip.File) {
	pomPath := "META-INF/maven/" + host.GAV.Group + "/" + host.GAV.Artifact + "/pom.xml"
	var body []byte
	for _, f := range files {
		if f.Name != pomPath {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return
		}
		body, err = io.ReadAll(r)
		r.Close()
		if err != nil {
			return
		}
	}
	if body == nil {
		return
	}
	incs, excs, shade, err := shadeArtifactSet(bytes.NewReader(body))
	if err != nil || !shade {
		return
	}
	if len(incs) == 0 {
		inv.BuildConfigNotes = append(inv.BuildConfigNotes, fmt.Sprintf("%s: maven-shade-plugin in %s has no artifactSet includes (shades every dependency): no build-config candidates derived", host.GAV, pomPath))
		return
	}
	p, err := parsePOM(body)
	if err != nil {
		inv.BuildConfigNotes = append(inv.BuildConfigNotes, fmt.Sprintf("%s: %s unreadable: %v", host.GAV, pomPath, err))
		return
	}
	res := &pomResolver{ctx: ctx, fetch: fetch, base: base, seen: map[string]*pomCtx{}}
	root := res.effective(p, 0)
	have := map[string]bool{host.GAV.PackageName(): true}
	for _, e := range inv.Embedded {
		if e.HostInstance == host.ID {
			have[e.GAV.PackageName()] = true
		}
	}
	type item struct {
		ctx   *pomCtx
		chain string
		depth int
	}
	queue := []item{{ctx: root, chain: pomPath, depth: 0}}
	added, unresolved := 0, 0
	notes := map[string]bool{}
	for _, n := range root.notes {
		notes[n] = true
	}
	for len(queue) > 0 {
		it := queue[0]
		queue = queue[1:]
		for _, d := range it.ctx.pom.Dependencies {
			if !shadedScope(d) {
				continue
			}
			g, a := it.ctx.resolve(d.GroupID), it.ctx.resolve(d.ArtifactID)
			if g == "" || a == "" || !matchesAny(incs, g, a) || matchesAny(excs, g, a) || have[g+":"+a] {
				continue
			}
			have[g+":"+a] = true
			v := it.ctx.version(d, g, a)
			gav := input.GAV{Group: g, Artifact: a, Version: v}
			chain := it.chain + " -> " + g + ":" + a
			if v == "" {
				unresolved++
				chain += " (version " + d.Version + " unresolved)"
			}
			inv.Embedded = append(inv.Embedded, Embedded{HostInstance: host.ID, GAV: gav, Source: SourceBuildConfig,
				SourceDetail: "maven-shade-plugin artifactSet includes " + strings.Join(incs, ", ") + ": " + chain})
			added++
			if v == "" || it.depth+1 >= maxPOMDepth {
				continue
			}
			dc, err := res.load(gav, 1)
			if err != nil {
				notes[fmt.Sprintf("pom of %s unavailable, its own dependencies not followed: %v", gav, err)] = true
				continue
			}
			for _, n := range dc.notes {
				notes[n] = true
			}
			queue = append(queue, item{ctx: dc, chain: chain, depth: it.depth + 1})
		}
	}
	if added > 0 {
		inv.BuildConfigNotes = append(inv.BuildConfigNotes, fmt.Sprintf("%s: %d build-config candidate(s) from maven-shade artifactSet includes in %s (%d with an unresolved version)", host.GAV, added, pomPath, unresolved))
	}
	for _, n := range sortedKeys(notes) {
		inv.BuildConfigNotes = append(inv.BuildConfigNotes, fmt.Sprintf("%s: build-config candidates: %s", host.GAV, n))
	}
}

func sortedKeys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
