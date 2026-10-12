package input

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PreparedApp is what PrepareAppInput turns any supported -app value
// into: a single jar hierarchy.Build can read the application's own
// classes from, and (for a Spring Boot fat jar) a directory of extracted
// dependency jars to resolve alongside whatever -deps names.
type PreparedApp struct {
	AppClassesJar string // always a real jar on disk, ready for hierarchy.Build
	ExtractedLibs string // a directory of dependency jars extracted from BOOT-INF/lib/, or "" if appPath wasn't a fat jar
	Cleanup       func() error
}

// PrepareAppInput accepts any of the three application inputs section 3.1
// describes:
//   - a directory of .class files (e.g. target/classes): packed into a
//     temp jar;
//   - a plain jar (e.g. a hand-built app-classes.jar): used as-is;
//   - a Spring Boot fat jar (detected by a BOOT-INF/ entry): its
//     BOOT-INF/classes/ is repacked into a temp jar and every
//     BOOT-INF/lib/*.jar is extracted to a temp directory, both without
//     the caller needing to know which case it was.
//
// The caller must call the returned PreparedApp.Cleanup when done, which
// removes any temp files PrepareAppInput created (a no-op for the plain
// jar case, which created none).
func PrepareAppInput(appPath string) (PreparedApp, error) {
	info, err := os.Stat(appPath)
	if err != nil {
		return PreparedApp{}, fmt.Errorf("stat %s: %w", appPath, err)
	}

	if info.IsDir() {
		return prepareFromDir(appPath)
	}
	return prepareFromJar(appPath)
}

func noopCleanup() error { return nil }

func prepareFromDir(dir string) (PreparedApp, error) {
	tmp, err := os.CreateTemp("", "ariadne-app-classes-*.jar")
	if err != nil {
		return PreparedApp{}, fmt.Errorf("creating temp jar: %w", err)
	}
	tmp.Close()

	if err := packDirAsJar(dir, tmp.Name()); err != nil {
		os.Remove(tmp.Name())
		return PreparedApp{}, err
	}
	return PreparedApp{
		AppClassesJar: tmp.Name(),
		Cleanup:       func() error { return os.Remove(tmp.Name()) },
	}, nil
}

func prepareFromJar(jarPath string) (PreparedApp, error) {
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return PreparedApp{}, fmt.Errorf("opening %s: %w", jarPath, err)
	}
	defer zr.Close()

	isBootJar := false
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "BOOT-INF/") {
			isBootJar = true
			break
		}
	}
	if !isBootJar {
		return PreparedApp{AppClassesJar: jarPath, Cleanup: noopCleanup}, nil
	}
	return prepareFromBootJar(zr)
}

// prepareFromBootJar repacks BOOT-INF/classes/ into a temp jar and
// extracts every BOOT-INF/lib/*.jar to a temp directory. Extraction (not
// reading the nested jars straight out of the outer zip in memory) is a
// deliberate simplification: it reuses hierarchy.Build and
// ResolveDepJars completely unchanged, at the cost of touching disk,
// which the original internal/input design (see PLAN.md Phase 6 status)
// wanted to avoid.
func prepareFromBootJar(zr *zip.ReadCloser) (PreparedApp, error) {
	tmpDir, err := os.MkdirTemp("", "ariadne-boot-jar-*")
	if err != nil {
		return PreparedApp{}, fmt.Errorf("creating temp dir: %w", err)
	}
	cleanup := func() error { return os.RemoveAll(tmpDir) }

	classesJarPath := filepath.Join(tmpDir, "app-classes.jar")
	libDir := filepath.Join(tmpDir, "lib")
	if err := os.Mkdir(libDir, 0o755); err != nil {
		cleanup()
		return PreparedApp{}, fmt.Errorf("creating %s: %w", libDir, err)
	}

	classesJar, err := os.Create(classesJarPath)
	if err != nil {
		cleanup()
		return PreparedApp{}, fmt.Errorf("creating %s: %w", classesJarPath, err)
	}
	zw := zip.NewWriter(classesJar)

	for _, f := range zr.File {
		switch {
		case strings.HasPrefix(f.Name, "BOOT-INF/classes/") && !f.FileInfo().IsDir():
			name := strings.TrimPrefix(f.Name, "BOOT-INF/classes/")
			if name == "" {
				continue
			}
			if err := copyZipEntry(zw, f, name); err != nil {
				zw.Close()
				classesJar.Close()
				cleanup()
				return PreparedApp{}, fmt.Errorf("repacking %s: %w", f.Name, err)
			}

		case strings.HasPrefix(f.Name, "BOOT-INF/lib/") && strings.HasSuffix(f.Name, ".jar"):
			name := strings.TrimPrefix(f.Name, "BOOT-INF/lib/")
			if err := extractZipEntry(f, filepath.Join(libDir, name)); err != nil {
				zw.Close()
				classesJar.Close()
				cleanup()
				return PreparedApp{}, fmt.Errorf("extracting %s: %w", f.Name, err)
			}
		}
	}

	if err := zw.Close(); err != nil {
		classesJar.Close()
		cleanup()
		return PreparedApp{}, fmt.Errorf("finishing %s: %w", classesJarPath, err)
	}
	if err := classesJar.Close(); err != nil {
		cleanup()
		return PreparedApp{}, fmt.Errorf("closing %s: %w", classesJarPath, err)
	}

	return PreparedApp{AppClassesJar: classesJarPath, ExtractedLibs: libDir, Cleanup: cleanup}, nil
}

func copyZipEntry(zw *zip.Writer, f *zip.File, newName string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	w, err := zw.Create(newName)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, rc)
	return err
}

func extractZipEntry(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

// packDirAsJar walks dir and writes every file into a new jar at
// jarPath, with paths relative to dir — the same shape
// `jar cf x.jar -C dir .` produces.
func packDirAsJar(dir, jarPath string) error {
	out, err := os.Create(jarPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", jarPath, err)
	}
	defer out.Close()
	zw := zip.NewWriter(out)

	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		return err
	})
	if err != nil {
		zw.Close()
		return fmt.Errorf("packing %s: %w", dir, err)
	}
	return zw.Close()
}
