package hierarchy

import (
	"archive/zip"
	"fmt"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/input"
)

// readJarClasses reads every .class entry from a jar, parses it, and
// returns one ClassInfo per class. module-info.class and the alternate
// per-version classes inside a multi-release jar's META-INF/versions/ are
// skipped: ariadne always analyzes the jar's base (lowest-version)
// classes, which is what actually runs on the JVM ariadne itself targets
// as a baseline, and avoids reading the same class twice under two names.
func readJarClasses(jarPath string, gav *input.GAV, isApp bool) ([]*ClassInfo, error) {
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", jarPath, err)
	}
	defer zr.Close()

	var out []*ClassInfo
	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".class") {
			continue
		}
		if f.Name == "module-info.class" || strings.HasPrefix(f.Name, "META-INF/versions/") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening %s in %s: %w", f.Name, jarPath, err)
		}
		cf, err := classfile.Parse(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("parsing %s in %s: %w", f.Name, jarPath, err)
		}

		out = append(out, &ClassInfo{
			FQCN:        cf.ThisClass,
			SuperClass:  cf.SuperClass,
			Interfaces:  cf.Interfaces,
			AccessFlags: cf.AccessFlags,
			Methods:     cf.Methods,
			Annotations: cf.Annotations,
			SourceJar:   jarPath,
			GAV:         gav,
			IsApp:       isApp,
		})
	}
	return out, nil
}
