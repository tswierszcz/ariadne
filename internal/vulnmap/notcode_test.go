package vulnmap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ariadne/internal/cves"
)

// A not_code entry (CVE:F10, NOT_CODE_MAPPABLE) carries a reason and no
// methods; it excludes methods, variants, role and coverage; the resolver
// returns it from the manual table (also by alias) without trying a fix
// diff, so the advisory is not MAPPING_MISSING.
func TestManualNotCode(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "m.yaml")
	os.WriteFile(p, []byte("CVE-1: {not_code: \"an installer\", review: reviewed}\nCVE-2: [a.B#m]\n"), 0o644)
	meta, err := LoadManualMeta(p)
	if err != nil {
		t.Fatal(err)
	}
	methods, _ := LoadManual(p)
	if meta["CVE-1"].NotCode != "an installer" || len(methods["CVE-1"]) != 0 {
		t.Fatalf("meta %+v methods %v", meta["CVE-1"], methods["CVE-1"])
	}
	for _, bad := range []string{
		"CVE-1: {not_code: x, methods: [a.B#m]}",
		"CVE-1: {not_code: x, role: vulnerable}",
		"CVE-1: {not_code: x, coverage: partial}",
		"CVE-1: {not_code: x, variants: [{package: a:b, versions: \"[1,2)\", methods: [a.B#m]}]}",
	} {
		os.WriteFile(p, []byte(bad+"\n"), 0o644)
		if _, err := LoadManualMeta(p); err == nil || !strings.Contains(err.Error(), "not_code excludes") {
			t.Errorf("%s: err = %v, want a not_code error", bad, err)
		}
	}

	r := NewResolver(methods, mustCache(t))
	r.ManualMeta = meta
	e, ok, err := r.Resolve(context.Background(), cves.CVE{ID: "GHSA-1", Aliases: []string{"CVE-1"}})
	if err != nil || !ok || e.NotCode != "an installer" || e.Source != SourceManual || e.Mapped() {
		t.Fatalf("Resolve = (%+v, %v, %v), want the not_code entry", e, ok, err)
	}

	// The embedded table's group-C entry.
	em, err := LoadManualMeta("")
	if err != nil || !strings.Contains(em["CVE-2025-49124"].NotCode, "installer") {
		t.Errorf("CVE-2025-49124 = %+v (%v)", em["CVE-2025-49124"], err)
	}
}
