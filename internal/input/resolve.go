package input

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveDepJars identifies every *.jar in dir, preferring the GAV embedded
// in each jar's own META-INF/maven/*/*/pom.properties (works regardless of
// build tool) and falling back to matching the jar's file name against
// depsListGAVs (typically parsed from `mvn dependency:list` via
// ParseDepsList). A jar neither source can identify is still returned,
// with GAV == nil, so the report can list it as "unidentified".
func ResolveDepJars(dir string, depsListGAVs []GAV) ([]DepJar, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading dependency dir %s: %w", dir, err)
	}

	byFileStem := indexByFileStem(depsListGAVs)

	var result []DepJar
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jar") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		dj, err := resolveOne(path, e.Name(), byFileStem)
		if err != nil {
			return nil, err
		}
		result = append(result, dj)
	}
	return result, nil
}

func resolveOne(path, name string, byFileStem map[string]GAV) (DepJar, error) {
	gavs, webAssetsOnly, err := inspectJar(path)
	if err != nil {
		return DepJar{}, err
	}
	stem := strings.TrimSuffix(name, ".jar")

	if len(gavs) > 0 {
		// Prefer the candidate whose own "artifact-version" matches the
		// jar's file name: for a jar that bundles several pom.properties
		// (a fat/uber jar, e.g. ehcache's), the other entries are internal
		// modules that were never published under their own coordinates,
		// so picking an arbitrary one (gavs[0]) can produce a GAV that OSV
		// has no record of at all.
		best := 0
		for i, g := range gavs {
			if g.Artifact+"-"+g.Version == stem {
				best = i
				break
			}
		}
		dj := DepJar{Path: path, GAV: &gavs[best], Source: SourcePOMProperties, WebAssetsOnly: webAssetsOnly}
		if len(gavs) > 1 {
			others := make([]GAV, 0, len(gavs)-1)
			for i, g := range gavs {
				if i != best {
					others = append(others, g)
				}
			}
			dj.Warnings = append(dj.Warnings, fmt.Sprintf(
				"%d pom.properties entries found inside this jar (shaded/uber jar?); using %s, ignoring %v",
				len(gavs), gavs[best], others))
		}
		return dj, nil
	}

	if gav, ok := byFileStem[stem]; ok {
		g := gav
		return DepJar{Path: path, GAV: &g, Source: SourceDepsList, WebAssetsOnly: webAssetsOnly}, nil
	}

	return DepJar{Path: path, GAV: nil, WebAssetsOnly: webAssetsOnly}, nil
}

// indexByFileStem maps the file-name stem Maven would use for a dependency
// jar ("artifactId-version" or "artifactId-version-classifier") to its GAV,
// so an unidentified jar can be matched by name against a `deps.txt` list.
func indexByFileStem(gavs []GAV) map[string]GAV {
	idx := make(map[string]GAV, len(gavs))
	for _, g := range gavs {
		stem := g.Artifact + "-" + g.Version
		if g.Classifier != "" {
			stem += "-" + g.Classifier
		}
		idx[stem] = g
	}
	return idx
}
