package vulnmap

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/mavenversion"
)

// TestMigratedManualMappings (CVE:F9b) pins the two entries migrated to
// schema v2: for each released version checked in the bytecode (see the
// comments in vuln_methods.yaml), exactly one variant covers it and
// selects the methods that release declares; versions outside the
// advisory's span select nothing (MAPPING_VERSION_UNCOVERED, never a
// guess).
func TestMigratedManualMappings(t *testing.T) {
	meta, err := LoadManualMeta("")
	if err != nil {
		t.Fatal(err)
	}
	const (
		jndiLE  = "org.apache.logging.log4j.core.lookup.JndiLookup#lookup(Lorg/apache/logging/log4j/core/LogEvent;Ljava/lang/String;)Ljava/lang/String;"
		jndiStr = "org.apache.logging.log4j.core.lookup.JndiLookup#lookup(Ljava/lang/String;)Ljava/lang/String;"
	)
	pick := func(id, pkg, v string) []string {
		var out []string
		n := 0
		for _, vr := range meta[id].Variants {
			if vr.Package == pkg && vr.Range.Contains(v) == mavenversion.In {
				n++
				out = append(out, vr.Methods...)
			}
		}
		if n > 1 {
			t.Errorf("%s %s: %d variants cover it", id, v, n)
		}
		sort.Strings(out)
		return out
	}
	for _, v := range []string{"2.0-beta9", "2.0-rc1", "2.0-rc2", "2.0", "2.0.1", "2.0.2"} {
		if got := pick("CVE-2021-44228", "org.apache.logging.log4j:log4j-core", v); !reflect.DeepEqual(got, []string{jndiStr, jndiLE}) {
			t.Errorf("log4j %s = %v", v, got)
		}
	}
	for _, v := range []string{"2.1", "2.2", "2.3", "2.3.1", "2.8.2", "2.11.1", "2.12.1", "2.12.2", "2.13.0", "2.14.1"} {
		if got := pick("CVE-2021-44228", "org.apache.logging.log4j:log4j-core", v); !reflect.DeepEqual(got, []string{jndiLE}) {
			t.Errorf("log4j %s = %v", v, got)
		}
	}
	for _, v := range []string{"2.0-beta8", "2.15.0", "2.17.1"} {
		if got := pick("CVE-2021-44228", "org.apache.logging.log4j:log4j-core", v); len(got) != 0 {
			t.Errorf("log4j %s outside the span = %v", v, got)
		}
	}
	for _, v := range []string{"1.5", "1.6", "1.7", "1.8", "1.9"} {
		got := pick("CVE-2022-42889", "org.apache.commons:commons-text", v)
		if len(got) != 11 {
			t.Errorf("commons-text %s: %d methods", v, len(got))
		}
		for _, m := range got {
			if !strings.HasPrefix(m, "org.apache.commons.text.StringSubstitutor#replace(") || strings.Contains(m, "Ljava/util/Map;") || strings.Contains(m, "Ljava/util/Properties;") {
				t.Errorf("commons-text %s: unexpected %s", v, m)
			}
		}
	}
	for _, v := range []string{"1.4", "1.10.0"} {
		if got := pick("CVE-2022-42889", "org.apache.commons:commons-text", v); len(got) != 0 {
			t.Errorf("commons-text %s outside the span = %v", v, got)
		}
	}
	for _, id := range []string{"CVE-2021-44228", "CVE-2022-42889"} {
		if m := meta[id]; m.Review != ReviewReviewed || m.Coverage != CoveragePartial {
			t.Errorf("%s meta %+v", id, m)
		}
	}
}
