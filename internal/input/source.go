package input

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PreparedApp is what PrepareAppInput turns any supported -app value
// into: a single jar hierarchy.Build can read the application's own
// classes from, and (for a fat archive: a Spring Boot fat jar, a WAR, or
// an EAR) a directory of extracted dependency jars to resolve alongside
// whatever -deps names.
type PreparedApp struct {
	AppClassesJar string // always a real jar on disk, ready for hierarchy.Build
	ExtractedLibs string // a directory of dependency jars extracted from the archive's own lib entries, or "" if appPath was a plain jar
	Cleanup       func() error
}

// PrepareAppInput accepts any of the application inputs section 3.1
// describes:
//   - a directory of .class files (e.g. target/classes): packed into a
//     temp jar;
//   - a plain jar (e.g. a hand-built app-classes.jar): used as-is;
//   - a Spring Boot fat jar (detected by a BOOT-INF/ entry): its
//     BOOT-INF/classes/ is repacked into a temp jar and every
//     BOOT-INF/lib/*.jar is extracted to a temp directory;
//   - a WAR (detected by a WEB-INF/ entry): WEB-INF/classes/ and
//     WEB-INF/lib/*.jar are handled the same way;
//   - an EAR (detected by META-INF/application.xml, or a nested *.war
//     entry when that file is missing): every nested WAR module's
//     WEB-INF/classes/WEB-INF/lib are merged the same way, and every
//     top-level *.jar is treated as a dependency jar (see Known
//     limitations: an EAR's own EJB-module jars are not distinguished
//     from shared libraries, so an EJB module's own classes are not
//     picked up as app code) —
//
// all without the caller needing to know which case it was.
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

// archiveKind is what detectArchiveKind tells a plain jar apart as.
type archiveKind int

const (
	archivePlain archiveKind = iota
	archiveBootJar
	archiveWAR
	archiveEAR
)

func prepareFromJar(jarPath string) (PreparedApp, error) {
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return PreparedApp{}, fmt.Errorf("opening %s: %w", jarPath, err)
	}
	defer zr.Close()

	switch detectArchiveKind(zr.File) {
	case archiveEAR:
		return prepareFromEar(zr.File)
	case archiveWAR:
		return prepareFromFatArchive(zr.File, "WEB-INF/classes/", "WEB-INF/lib/")
	case archiveBootJar:
		return prepareFromFatArchive(zr.File, "BOOT-INF/classes/", "BOOT-INF/lib/")
	default:
		return PreparedApp{AppClassesJar: jarPath, Cleanup: noopCleanup}, nil
	}
}

// detectArchiveKind looks at an archive's own entries to tell a plain
// jar apart from a Spring Boot fat jar, a WAR, or an EAR. EAR is checked
// first: an EAR's own top-level entries are nested *.war/*.jar modules,
// not a WEB-INF/BOOT-INF prefix of their own, so it has to be
// recognized before a WAR nested inside it would otherwise be mistaken
// for the outer archive (a nested WAR's own WEB-INF/ prefix is only
// visible once the EAR entry is opened, not from the outer listing).
func detectArchiveKind(files []*zip.File) archiveKind {
	hasNestedWar := false
	for _, f := range files {
		if f.Name == "META-INF/application.xml" {
			return archiveEAR
		}
		if !f.FileInfo().IsDir() && strings.HasSuffix(f.Name, ".war") {
			hasNestedWar = true
		}
	}
	if hasNestedWar {
		return archiveEAR
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name, "WEB-INF/") {
			return archiveWAR
		}
		if strings.HasPrefix(f.Name, "BOOT-INF/") {
			return archiveBootJar
		}
	}
	return archivePlain
}

// prepareFromFatArchive repacks classesPrefix into a temp jar (prefix
// stripped) and extracts every *.jar directly under libPrefix to a temp
// directory — the shared shape a Spring Boot fat jar (BOOT-INF/classes,
// BOOT-INF/lib) and a WAR (WEB-INF/classes, WEB-INF/lib) both have.
// Extraction (not reading the nested jars straight out of the outer zip
// in memory) is a deliberate simplification: it reuses hierarchy.Build
// and ResolveDepJars completely unchanged, at the cost of touching disk,
// which the original internal/input design (see PLAN.md Phase 6 status)
// wanted to avoid.
func prepareFromFatArchive(files []*zip.File, classesPrefix, libPrefix string) (PreparedApp, error) {
	return buildPreparedApp("ariadne-fat-archive-*", func(zw *zip.Writer, written map[string]bool, libDir string) error {
		return extractModule(files, classesPrefix, libPrefix, zw, written, libDir)
	})
}

// prepareFromEar merges every nested WAR module's WEB-INF/classes +
// WEB-INF/lib (via extractModule, same as a standalone WAR) and every
// top-level jar into one app-classes jar and one lib directory, the same
// shape a WAR or a Boot fat jar produces. A top-level jar is always
// treated as a dependency jar, never as app code: an EAR's own top-level
// modules can in principle be EJB jars carrying real application
// classes, but there is no reliable signal in the zip listing alone to
// tell one of those apart from a shared library jar without parsing
// META-INF/application.xml (which, since Java EE 5, is itself optional)
// — see Known limitations.
func prepareFromEar(files []*zip.File) (PreparedApp, error) {
	return buildPreparedApp("ariadne-ear-*", func(zw *zip.Writer, written map[string]bool, libDir string) error {
		for _, f := range files {
			switch {
			case !f.FileInfo().IsDir() && strings.HasSuffix(f.Name, ".war"):
				nested, err := readNestedZip(f)
				if err != nil {
					return fmt.Errorf("opening nested WAR %s: %w", f.Name, err)
				}
				if err := extractModule(nested, "WEB-INF/classes/", "WEB-INF/lib/", zw, written, libDir); err != nil {
					return err
				}

			case !f.FileInfo().IsDir() && strings.HasSuffix(f.Name, ".jar"):
				dest := filepath.Join(libDir, filepath.Base(f.Name))
				if _, err := os.Stat(dest); err == nil {
					continue // already extracted by an earlier module with the same jar name
				}
				if err := extractZipEntry(f, dest); err != nil {
					return fmt.Errorf("extracting %s: %w", f.Name, err)
				}
			}
		}
		return nil
	})
}

// buildPreparedApp creates a temp dir with a lib/ subdirectory and an
// open app-classes.jar, lets fill populate both, and closes everything
// into a PreparedApp — the scaffolding prepareFromFatArchive and
// prepareFromEar both need, so neither has to repeat the temp-dir/
// zip-writer open-close-cleanup dance by hand.
func buildPreparedApp(tmpDirPattern string, fill func(zw *zip.Writer, written map[string]bool, libDir string) error) (PreparedApp, error) {
	tmpDir, err := os.MkdirTemp("", tmpDirPattern)
	if err != nil {
		return PreparedApp{}, fmt.Errorf("creating temp dir: %w", err)
	}
	cleanup := func() error { return os.RemoveAll(tmpDir) }

	libDir := filepath.Join(tmpDir, "lib")
	if err := os.Mkdir(libDir, 0o755); err != nil {
		cleanup()
		return PreparedApp{}, fmt.Errorf("creating %s: %w", libDir, err)
	}

	classesJarPath := filepath.Join(tmpDir, "app-classes.jar")
	classesJar, err := os.Create(classesJarPath)
	if err != nil {
		cleanup()
		return PreparedApp{}, fmt.Errorf("creating %s: %w", classesJarPath, err)
	}
	zw := zip.NewWriter(classesJar)

	if err := fill(zw, map[string]bool{}, libDir); err != nil {
		zw.Close()
		classesJar.Close()
		cleanup()
		return PreparedApp{}, err
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

// extractModule repacks every file under classesPrefix (prefix stripped)
// into zw, skipping a name already written by an earlier call — an EAR
// with several WAR modules sharing a class name, where the first module
// read wins, the same "first one wins" rule hierarchy.go already applies
// to a class duplicated across two dependency jars — and extracts every
// *.jar directly under libPrefix into libDir, skipping a file name
// that's already there (two modules bundling a same-named dependency
// jar, assumed to be the same jar; see Known limitations).
func extractModule(files []*zip.File, classesPrefix, libPrefix string, zw *zip.Writer, written map[string]bool, libDir string) error {
	for _, f := range files {
		switch {
		case strings.HasPrefix(f.Name, classesPrefix) && !f.FileInfo().IsDir():
			name := strings.TrimPrefix(f.Name, classesPrefix)
			if name == "" || written[name] {
				continue
			}
			written[name] = true
			if err := copyZipEntry(zw, f, name); err != nil {
				return fmt.Errorf("repacking %s: %w", f.Name, err)
			}

		case strings.HasPrefix(f.Name, libPrefix) && strings.HasSuffix(f.Name, ".jar"):
			name := strings.TrimPrefix(f.Name, libPrefix)
			dest := filepath.Join(libDir, filepath.Base(name))
			if _, err := os.Stat(dest); err == nil {
				continue
			}
			if err := extractZipEntry(f, dest); err != nil {
				return fmt.Errorf("extracting %s: %w", f.Name, err)
			}
		}
	}
	return nil
}

// readNestedZip reads a zip.File fully into memory and opens it as its
// own zip.Reader — needed for a WAR nested inside an EAR, since
// archive/zip can only open a *zip.File as a one-shot stream, not re-seek
// into a nested archive in place.
func readNestedZip(f *zip.File) ([]*zip.File, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	return zr.File, nil
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
