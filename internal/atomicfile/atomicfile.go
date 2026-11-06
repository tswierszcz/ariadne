// Package atomicfile writes a run's output artifacts all or nothing
// (CVE:F9.9): every file is first written in full to a temporary file in
// its own directory, synced and closed; only when all of them are, each
// temporary file is renamed over its destination. A failure while
// writing replaces nothing — an earlier run's outputs stay as they were,
// and no new, partial artifact can look complete. A rename is atomic per
// file on POSIX file systems; a failure between two renames (rare: the
// temporary files are already in place) is reported with the files that
// were already replaced.
package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Set is a group of files committed together.
type Set struct {
	files []file
}

type file struct {
	path string
	data []byte
}

// Add schedules path to be written with data. A later Add of the same
// path replaces the earlier one.
func (s *Set) Add(path string, data []byte) {
	for i := range s.files {
		if s.files[i].path == path {
			s.files[i].data = data
			return
		}
	}
	s.files = append(s.files, file{path: path, data: data})
}

// Paths are the scheduled destinations, in Add order.
func (s *Set) Paths() []string {
	var out []string
	for _, f := range s.files {
		out = append(out, f.path)
	}
	return out
}

// createTemp is os.CreateTemp, replaceable in tests.
var createTemp = os.CreateTemp

// rename is os.Rename, replaceable in tests.
var rename = os.Rename

// Commit writes every scheduled file as described in the package doc.
func (s *Set) Commit() error {
	temps := make([]string, len(s.files))
	cleanup := func(from int) {
		for _, t := range temps[from:] {
			if t != "" {
				os.Remove(t)
			}
		}
	}
	for i, f := range s.files {
		dir, base := filepath.Split(f.path)
		if dir == "" {
			dir = "."
		}
		tmp, err := createTemp(dir, "."+base+".*.tmp")
		if err != nil {
			cleanup(0)
			return fmt.Errorf("writing %s: %w (no output file was replaced)", f.path, err)
		}
		temps[i] = tmp.Name()
		_, werr := tmp.Write(f.data)
		serr := tmp.Sync()
		cerr := tmp.Close()
		if err := firstErr(werr, serr, cerr, os.Chmod(tmp.Name(), 0o644)); err != nil {
			cleanup(0)
			return fmt.Errorf("writing %s: %w (no output file was replaced)", f.path, err)
		}
	}
	var replaced []string
	for i, f := range s.files {
		if err := rename(temps[i], f.path); err != nil {
			cleanup(i)
			msg := "no output file was replaced"
			if len(replaced) > 0 {
				msg = "already replaced: " + strings.Join(replaced, ", ")
			}
			return fmt.Errorf("writing %s: %w (%s)", f.path, err, msg)
		}
		temps[i] = ""
		replaced = append(replaced, f.path)
	}
	return nil
}

// WriteFile writes one file atomically.
func WriteFile(path string, data []byte) error {
	var s Set
	s.Add(path, data)
	return s.Commit()
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
