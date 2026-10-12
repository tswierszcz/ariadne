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

		// ExtractCalls needs cf's constant pool and bootstrap methods,
		// which classfile.ClassFile deliberately doesn't expose, so this
		// is the only point where it can run; the result (not cf itself)
		// is what ClassInfo keeps.
		methods := make([]Method, len(cf.Methods))
		for i, m := range cf.Methods {
			methods[i].Method = m
			if m.Code == nil {
				continue
			}
			calls, unresolved, err := cf.ExtractCalls(m.Code)
			if err != nil {
				return nil, fmt.Errorf("extracting calls from %s %s#%s%s: %w", jarPath, cf.ThisClass, m.Name, m.Descriptor, err)
			}
			methods[i].Calls = calls
			methods[i].UnresolvedDynamic = unresolved
		}

		out = append(out, &ClassInfo{
			FQCN:         cf.ThisClass,
			SuperClass:   cf.SuperClass,
			Interfaces:   cf.Interfaces,
			AccessFlags:  cf.AccessFlags,
			MajorVersion: cf.MajorVersion,
			Methods:      methods,
			Annotations:  cf.Annotations,
			SourceJar:    jarPath,
			GAV:          gav,
			IsApp:        isApp,
		})
	}
	return out, nil
}
