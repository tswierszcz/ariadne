package input

import (
	"archive/zip"
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// pomPropertiesPath matches META-INF/maven/<group>/<artifact>/pom.properties
// inside a jar, however deep the group/artifact path segments are (a
// groupId itself contains dots, not slashes, so this is always exactly
// 4 path segments).
var pomPropertiesPath = regexp.MustCompile(`^META-INF/maven/[^/]+/[^/]+/pom\.properties$`)

// webJarPrefix is where the WebJars convention puts a jar's web assets.
const webJarPrefix = "META-INF/resources/webjars/"

// inspectJar reads every META-INF/maven/*/*/pom.properties entry from a
// jar and returns the GAV each one declares. A normal jar has exactly one;
// a shaded/uber jar can have several, which the caller should treat as a
// warning (see DepJar.Warnings) rather than a hard error, since we can
// still identify at least one real dependency inside it. It also reports
// whether the jar is a WebJar with no JVM bytecode (see DepJar.WebAssetsOnly).
func inspectJar(path string) (gavs []GAV, webAssetsOnly bool, err error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, false, fmt.Errorf("opening %s as a zip: %w", path, err)
	}
	defer zr.Close()

	hasClass, hasWebAssets := false, false
	for _, f := range zr.File {
		switch {
		case strings.HasSuffix(f.Name, ".class"):
			hasClass = true
		case strings.HasPrefix(f.Name, webJarPrefix):
			hasWebAssets = true
		}
		if !pomPropertiesPath.MatchString(f.Name) {
			continue
		}
		gav, err := readPomProperties(f)
		if err != nil {
			return nil, false, fmt.Errorf("reading %s in %s: %w", f.Name, path, err)
		}
		if gav != nil {
			gavs = append(gavs, *gav)
		}
	}
	return gavs, hasWebAssets && !hasClass, nil
}

func readPomProperties(f *zip.File) (*GAV, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	props := make(map[string]string)
	sc := bufio.NewScanner(rc)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		props[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	group, artifact, version := props["groupId"], props["artifactId"], props["version"]
	if group == "" || artifact == "" || version == "" {
		// Incomplete pom.properties; treat as not found rather than erroring
		// the whole jar out.
		return nil, nil
	}
	return &GAV{Group: group, Artifact: artifact, Version: version}, nil
}
