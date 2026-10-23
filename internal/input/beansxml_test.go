package input

import (
	"path/filepath"
	"testing"
)

func TestFindBeansXML_findsSpringBeansFile(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "app-classes.jar")
	writeZip(t, jar, map[string][]byte{
		"com/example/App.class": []byte("fake class bytes"),
		"applicationContext.xml": []byte(
			`<beans xmlns="http://www.springframework.org/schema/beans"><bean id="x" class="demo.X"/></beans>`),
		"not-a-beans-file.xml": []byte(`<config><setting>1</setting></config>`),
		"META-INF/MANIFEST.MF": []byte("Manifest-Version: 1.0\n"),
	})

	found, warnings, err := FindBeansXML(jar)
	if err != nil {
		t.Fatalf("FindBeansXML: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if len(found) != 1 {
		t.Fatalf("found = %v, want exactly 1 entry", found)
	}
	if _, ok := found["applicationContext.xml"]; !ok {
		t.Errorf("found = %v, want applicationContext.xml", found)
	}
}

func TestFindBeansXML_noBeansFilesIsEmpty(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "app-classes.jar")
	writeZip(t, jar, map[string][]byte{
		"com/example/App.class": []byte("fake class bytes"),
	})

	found, _, err := FindBeansXML(jar)
	if err != nil {
		t.Fatalf("FindBeansXML: %v", err)
	}
	if len(found) != 0 {
		t.Errorf("found = %v, want empty", found)
	}
}

func TestIsSpringBeansRoot(t *testing.T) {
	cases := []struct {
		data []byte
		want bool
	}{
		{[]byte(`<beans><bean id="x" class="demo.X"/></beans>`), true},
		{[]byte(`<?xml version="1.0"?><beans xmlns="http://www.springframework.org/schema/beans"></beans>`), true},
		{[]byte(`<config><setting>1</setting></config>`), false},
		{[]byte(`not xml at all`), false},
		{[]byte(``), false},
	}
	for _, c := range cases {
		if got := isSpringBeansRoot(c.data); got != c.want {
			t.Errorf("isSpringBeansRoot(%q) = %v, want %v", c.data, got, c.want)
		}
	}
}
