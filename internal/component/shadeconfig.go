package component

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"ariadne/internal/input"
	"ariadne/internal/relocate"
)

// shadeRelocations reads the maven-shade-plugin <relocation> rules from
// the host jar's own META-INF/maven/<group>/<artifact>/pom.xml (CVE:F7a,
// build configuration as evidence). The pom describes the build, not the
// result — rules of every execution and profile are taken, and the
// classes decide (a rule that renames nothing has no effect). A rule
// with a property placeholder or wildcard, or not a package name, is
// skipped with a note. Returns the rules and notes appended.
func shadeRelocations(files []*zip.File, host *input.GAV, notes []string) ([]relocate.Rule, []string) {
	if host == nil {
		return nil, notes
	}
	path := "META-INF/maven/" + host.Group + "/" + host.Artifact + "/pom.xml"
	var f *zip.File
	for _, zf := range files {
		if zf.Name == path {
			f = zf
		}
	}
	if f == nil {
		return nil, notes
	}
	r, err := f.Open()
	if err != nil {
		return nil, append(notes, fmt.Sprintf("%s in %s: %v", path, host, err))
	}
	defer r.Close()
	pairs, err := parseShadeRelocations(r)
	if err != nil {
		return nil, append(notes, fmt.Sprintf("%s in %s: %v", path, host, err))
	}
	source := "build-config: " + path + " (maven-shade-plugin)"
	var rules []relocate.Rule
	seen := map[[2]string]bool{}
	for _, p := range pairs {
		if seen[p] {
			continue
		}
		seen[p] = true
		if strings.ContainsAny(p[0]+p[1], "${}*") {
			notes = append(notes, fmt.Sprintf("%s: relocation %s -> %s skipped (placeholder or wildcard)", host, p[0], p[1]))
			continue
		}
		from, to := strings.ReplaceAll(p[0], "/", "."), strings.ReplaceAll(p[1], "/", ".")
		if _, err := relocate.New([]relocate.Rule{{From: from, To: to}}); err != nil {
			notes = append(notes, fmt.Sprintf("%s: relocation %s -> %s skipped: %v", host, p[0], p[1], err))
			continue
		}
		rules = append(rules, relocate.Rule{From: from, To: to, Source: source})
	}
	if len(rules) > 0 {
		notes = append(notes, fmt.Sprintf("%s: %d maven-shade relocation(s) read from %s", host, len(rules), path))
	}
	return rules, notes
}

// parseShadeRelocations returns the (pattern, shadedPattern) pairs of
// every <relocation> inside a <plugin> whose <artifactId> is
// maven-shade-plugin.
func parseShadeRelocations(r io.Reader) ([][2]string, error) {
	dec := newPOMDecoder(r)
	type plugin struct {
		artifact string
		pairs    [][2]string
	}
	var (
		stack   []string
		plugins []*plugin
		cur     [2]string
		text    strings.Builder
		out     [][2]string
	)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			text.Reset()
			switch t.Name.Local {
			case "plugin":
				plugins = append(plugins, &plugin{})
			case "relocation":
				cur = [2]string{}
			}
		case xml.CharData:
			text.Write(t)
		case xml.EndElement:
			name := t.Name.Local
			parent := ""
			if len(stack) >= 2 {
				parent = stack[len(stack)-2]
			}
			v := strings.TrimSpace(text.String())
			switch {
			case name == "artifactId" && parent == "plugin" && len(plugins) > 0:
				plugins[len(plugins)-1].artifact = v
			case name == "pattern" && parent == "relocation":
				cur[0] = v
			case name == "shadedPattern" && parent == "relocation":
				cur[1] = v
			case name == "relocation" && len(plugins) > 0:
				if cur[0] != "" && cur[1] != "" {
					p := plugins[len(plugins)-1]
					p.pairs = append(p.pairs, cur)
				}
			case name == "plugin" && len(plugins) > 0:
				p := plugins[len(plugins)-1]
				plugins = plugins[:len(plugins)-1]
				if p.artifact == "maven-shade-plugin" {
					out = append(out, p.pairs...)
				}
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			text.Reset()
		}
	}
	return out, nil
}
