package input

import (
	"path/filepath"
	"testing"
)

func TestListServiceFiles_readsEntriesAndHashesContent(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "provider.jar")
	content := []byte("com.example.FooProvider\n")
	writeZip(t, jar, map[string][]byte{
		"META-INF/services/com.example.Foo": content,
		"com/example/FooProvider.class":     []byte("not real bytecode, irrelevant here"),
	})

	files, err := ListServiceFiles(jar)
	if err != nil {
		t.Fatalf("ListServiceFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d service files, want 1: %+v", len(files), files)
	}
	f := files[0]
	if f.Interface != "com.example.Foo" {
		t.Errorf("Interface = %q, want %q", f.Interface, "com.example.Foo")
	}
	if string(f.Content) != string(content) {
		t.Errorf("Content = %q, want %q", f.Content, content)
	}
	if f.SHA256 == "" {
		t.Errorf("SHA256 is empty")
	}
}

func TestListServiceFiles_ignoresEverythingElse(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "plain.jar")
	writeZip(t, jar, map[string][]byte{
		"com/example/Foo.class": []byte("irrelevant"),
		"META-INF/MANIFEST.MF":  []byte("Manifest-Version: 1.0\n"),
	})

	files, err := ListServiceFiles(jar)
	if err != nil {
		t.Fatalf("ListServiceFiles: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("got %d service files, want 0: %+v", len(files), files)
	}
}

func TestListServiceFiles_multipleInterfaces(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "multi.jar")
	writeZip(t, jar, map[string][]byte{
		"META-INF/services/com.example.A": []byte("com.example.AImpl\n"),
		"META-INF/services/com.example.B": []byte("com.example.BImpl\n"),
	})

	files, err := ListServiceFiles(jar)
	if err != nil {
		t.Fatalf("ListServiceFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d service files, want 2: %+v", len(files), files)
	}
}
