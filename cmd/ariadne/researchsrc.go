package main

// RESEARCH_INTERFACE RI4 (source locations): the source file of a method,
// from its class's SourceFile attribute and package, and — with -sources —
// the hash of that file in the given source tree or sources jar.

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// sourcePath is the package-relative path of a class's source file
// ("fx/research/Extractor.java"), "" without a SourceFile attribute or with
// one that is not a plain .java/.kt/... file name.
func sourcePath(fqcn, sourceFile string) string {
	if sourceFile == "" || strings.ContainsAny(sourceFile, `/\`) || sourceFile == "." || sourceFile == ".." {
		return ""
	}
	if i := strings.LastIndexByte(fqcn, '.'); i >= 0 {
		return strings.ReplaceAll(fqcn[:i], ".", "/") + "/" + sourceFile
	}
	return sourceFile
}

// sourceTree indexes the files of -sources by every path suffix that
// starts at a directory boundary, so a package-relative path finds its file
// whatever the source root is (src/main/java/..., a sources jar's root).
type sourceTree struct {
	// bySuffix: package-relative path -> files that end with it.
	bySuffix map[string][]string
	read     func(name string) ([]byte, error)
	close    func()
	// digest is the SHA-256 over the sorted (path, file hash) list: the
	// input's identity in the bundle.
	digest string
	files  int
}

func openSources(path string) (*sourceTree, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("-sources: %w", err)
	}
	t := &sourceTree{bySuffix: map[string][]string{}, close: func() {}}
	var names []string
	if fi.IsDir() {
		err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type().IsRegular() && isSourceName(p) {
				rel, _ := filepath.Rel(path, p)
				names = append(names, filepath.ToSlash(rel))
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("-sources %s: %w", path, err)
		}
		t.read = func(name string) ([]byte, error) { return os.ReadFile(filepath.Join(path, filepath.FromSlash(name))) }
	} else {
		zr, err := zip.OpenReader(path)
		if err != nil {
			return nil, fmt.Errorf("-sources %s: not a directory or a jar/zip: %w", path, err)
		}
		files := map[string]*zip.File{}
		for _, f := range zr.File {
			if !f.FileInfo().IsDir() && isSourceName(f.Name) && !strings.Contains(f.Name, "..") {
				files[f.Name] = f
				names = append(names, f.Name)
			}
		}
		t.read = func(name string) ([]byte, error) {
			rc, err := files[name].Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
		t.close = func() { zr.Close() }
	}
	sort.Strings(names)
	h := sha256.New()
	for _, n := range names {
		b, err := t.read(n)
		if err != nil {
			t.close()
			return nil, fmt.Errorf("-sources: reading %s: %w", n, err)
		}
		s := sha256.Sum256(b)
		fmt.Fprintf(h, "%s\x00%x\n", n, s)
		for i := 0; ; {
			t.bySuffix[n[i:]] = append(t.bySuffix[n[i:]], n)
			j := strings.IndexByte(n[i:], '/')
			if j < 0 {
				break
			}
			i += j + 1
		}
	}
	t.digest = hex.EncodeToString(h.Sum(nil))
	t.files = len(names)
	return t, nil
}

func isSourceName(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".java", ".kt", ".scala", ".groovy":
		return true
	}
	return false
}

// lookup returns the file a package-relative path names and its hash; ok
// is false when no file or more than one file ends with that path.
func (t *sourceTree) lookup(pkgPath string, cache map[string]string) (name, hash string, ok bool) {
	c := t.bySuffix[pkgPath]
	if len(c) != 1 {
		return "", "", false
	}
	if h, done := cache[c[0]]; done {
		return c[0], h, h != ""
	}
	b, err := t.read(c[0])
	if err != nil {
		cache[c[0]] = ""
		return "", "", false
	}
	s := sha256.Sum256(b)
	cache[c[0]] = hex.EncodeToString(s[:])
	return c[0], cache[c[0]], true
}
