package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		return "<missing>"
	}
	return string(b)
}

func leftovers(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	es, _ := os.ReadDir(dir)
	for _, e := range es {
		if strings.HasSuffix(e.Name(), ".tmp") {
			out = append(out, e.Name())
		}
	}
	return out
}

func TestCommitWritesAll(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "report.md"), filepath.Join(dir, "results.json")
	os.WriteFile(a, []byte("old report"), 0o644)
	var s Set
	s.Add(a, []byte("new report"))
	s.Add(b, []byte("draft"))
	s.Add(b, []byte("new json"))
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	if read(t, a) != "new report" || read(t, b) != "new json" || len(leftovers(t, dir)) != 0 {
		t.Errorf("a=%q b=%q leftovers=%v", read(t, a), read(t, b), leftovers(t, dir))
	}
	if st, _ := os.Stat(b); st.Mode().Perm() != 0o644 {
		t.Errorf("mode %v", st.Mode().Perm())
	}
}

// TestWriteFailureReplacesNothing: the second file cannot be written (its
// directory does not exist) — the first destination keeps its old
// content and no temporary file is left behind.
func TestWriteFailureReplacesNothing(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "report.md")
	os.WriteFile(a, []byte("old report"), 0o644)
	var s Set
	s.Add(a, []byte("new report"))
	s.Add(filepath.Join(dir, "missing", "results.json"), []byte("new json"))
	err := s.Commit()
	if err == nil || !strings.Contains(err.Error(), "no output file was replaced") {
		t.Fatalf("err = %v", err)
	}
	if read(t, a) != "old report" || len(leftovers(t, dir)) != 0 {
		t.Errorf("a=%q leftovers=%v", read(t, a), leftovers(t, dir))
	}
}

// TestRenameFailureNamesReplaced: a failure between renames says which
// files were already replaced and removes the remaining temporaries.
func TestRenameFailureNamesReplaced(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	calls := 0
	rename = func(from, to string) error {
		calls++
		if calls == 2 {
			return errors.New("boom")
		}
		return os.Rename(from, to)
	}
	defer func() { rename = os.Rename }()
	var s Set
	s.Add(a, []byte("A"))
	s.Add(b, []byte("B"))
	err := s.Commit()
	if err == nil || !strings.Contains(err.Error(), "already replaced: "+a) {
		t.Fatalf("err = %v", err)
	}
	if read(t, a) != "A" || read(t, b) != "<missing>" || len(leftovers(t, dir)) != 0 {
		t.Errorf("a=%q b=%q leftovers=%v", read(t, a), read(t, b), leftovers(t, dir))
	}
}
