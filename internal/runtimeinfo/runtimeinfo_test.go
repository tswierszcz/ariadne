package runtimeinfo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseVersion(t *testing.T) {
	cases := map[string]string{
		"17.0.9+9":          "17.0.9+9",
		"25.0.5+8-1-Debian": "25.0.5+8",
		"21.0.1+12-LTS":     "21.0.1+12",
		"11.0.21.1+1":       "11.0.21.1+1",
		"21":                "21.0.0",
		"17-ea+35":          "17.0.0+35",
		"1.8.0_391-b13":     "8.0.391+13",
		"1.8.0_392":         "8.0.392",
	}
	for in, want := range cases {
		r, ok := ParseVersion(in)
		if !ok || r.String() != want {
			t.Errorf("ParseVersion(%q) = %v %v, want %s", in, r, ok, want)
		}
	}
	for _, bad := range []string{"", "openjdk", "x17"} {
		if _, ok := ParseVersion(bad); ok {
			t.Errorf("ParseVersion(%q) accepted", bad)
		}
	}
}

// TestFromJDKHome: the real Debian package release file, and constructed
// Temurin, Oracle JDK 8 and IMPLEMENTOR-less files. A shared IMPLEMENTOR
// names no distribution; a missing one names no vendor; Debian's package
// version comes from the runtime version and its release stays
// unresolved.
func TestFromJDKHome(t *testing.T) {
	deb, err := FromJDKHome("testdata/debian-25")
	if err != nil {
		t.Fatal(err)
	}
	if deb.Vendor != "Debian" || deb.Version == nil || deb.Version.Feature != 25 || deb.Version.Build != 8 ||
		deb.Package == nil || deb.Package.Name != "openjdk-25" || deb.Package.Version != "25.0.5+8-1" || deb.Package.Distro != "" {
		t.Errorf("debian = %+v %+v %+v", deb, deb.Version, deb.Package)
	}
	if !has(deb.Unresolved, "package distro") || deb.OS != "Linux" || deb.Arch != "x86_64" || deb.Raw["JAVA_VERSION_DATE"] != "2026-10-20" {
		t.Errorf("debian unresolved/os/raw = %v %s %s", deb.Unresolved, deb.OS, deb.Arch)
	}
	tem, _ := FromJDKHome("testdata/temurin-17")
	if tem.Distribution != "Temurin" || tem.Version.String() != "17.0.9+9" || len(tem.Unresolved) != 0 {
		t.Errorf("temurin = %+v unresolved %v", tem, tem.Unresolved)
	}
	ora, _ := FromJDKHome("testdata/jdk8-oracle")
	if ora.Vendor != "Oracle" || ora.Distribution != "" || !has(ora.Unresolved, "shared by several products") || ora.Version.Feature != 8 || ora.Version.Update != 391 {
		t.Errorf("oracle = %+v unresolved %v", ora, ora.Unresolved)
	}
	none, _ := FromJDKHome("testdata/no-implementor")
	if none.Vendor != "" || !has(none.Unresolved, "no IMPLEMENTOR") || !has(none.Unresolved, "no build number") || !has(none.Unresolved, "arch") {
		t.Errorf("no implementor unresolved = %v", none.Unresolved)
	}
	if _, err := FromJDKHome(t.TempDir()); err == nil || !strings.Contains(err.Error(), "release file") {
		t.Errorf("missing release file: %v", err)
	}
}

func TestFromFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "rt.yaml")
	os.WriteFile(p, []byte("vendor: Debian\ndistribution: OpenJDK package\nversion: 17.0.13+11-1~deb12u1\nos: Linux\narch: x86_64\npackage: {name: openjdk-17, version: 17.0.13+11-1~deb12u1, distro: \"Debian:12\"}\n"), 0o644)
	r, err := FromFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Source != SourceRuntimeFile || r.Version == nil || r.Version.String() != "17.0.13+11" || r.Package.Distro != "Debian:12" || len(r.Unresolved) != 0 {
		t.Errorf("declared = %+v %+v unresolved %v", r, r.Version, r.Unresolved)
	}
	os.WriteFile(p, []byte("vendor: X\nversoin: 17\n"), 0o644)
	if _, err := FromFile(p); err == nil || !strings.Contains(err.Error(), "versoin") {
		t.Errorf("unknown field: %v", err)
	}
}

func has(list []string, sub string) bool {
	for _, s := range list {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
