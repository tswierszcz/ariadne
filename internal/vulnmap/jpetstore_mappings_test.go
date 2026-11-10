package vulnmap

import (
	"testing"

	"ariadne/internal/mavenversion"
)

// TestHSQLDBAndLog4j1Mappings pins the seven sinks added for JPetStore's
// HSQLDB 2.3.4 and log4j 1.2.17 (GRAPH_PRECISION:P4, step 3): every
// release checked in the bytecode (see the comments in vuln_methods.yaml)
// is covered by exactly one variant with the expected methods, and a
// version outside the checked span selects nothing
// (MAPPING_VERSION_UNCOVERED, never a guess) — HSQLDB 2.7.1 is the fixed
// release, log4j below 1.2.8 was not checked, and log4j-core (which OSV
// also lists for CVE-2023-26464) has no variant at all.
func TestHSQLDBAndLog4j1Mappings(t *testing.T) {
	meta, err := LoadManualMeta("")
	if err != nil {
		t.Fatal(err)
	}
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
		return out
	}
	const (
		hsqldb = "org.hsqldb:hsqldb"
		log4j  = "log4j:log4j"
	)
	want := map[string]struct {
		pkg     string
		methods []string
	}{
		"CVE-2022-41853": {hsqldb, []string{"org.hsqldb.Routine#getMethods", "org.hsqldb.Routine#invokeJavaMethod"}},
		"CVE-2019-17571": {log4j, []string{"org.apache.log4j.net.SocketNode#run"}},
		"CVE-2021-4104":  {log4j, []string{"org.apache.log4j.net.JMSAppender#activateOptions", "org.apache.log4j.net.JMSAppender#lookup"}},
		"CVE-2022-23302": {log4j, []string{"org.apache.log4j.net.JMSSink#<init>", "org.apache.log4j.net.JMSSink#onMessage"}},
		"CVE-2022-23305": {log4j, []string{"org.apache.log4j.jdbc.JDBCAppender#getLogStatement", "org.apache.log4j.jdbc.JDBCAppender#flushBuffer", "org.apache.log4j.jdbc.JDBCAppender#execute"}},
		"CVE-2022-23307": {log4j, []string{"org.apache.log4j.chainsaw.LoggingReceiver$Slurper#run"}},
		"CVE-2023-26464": {log4j, []string{"org.apache.log4j.net.SocketNode#run", "org.apache.log4j.chainsaw.LoggingReceiver$Slurper#run"}},
	}
	checked := map[string][]string{
		hsqldb: {"2.0.0", "2.2.9", "2.3.4", "2.5.0", "2.6.1", "2.7.0"},
		log4j:  {"1.2.8", "1.2.9", "1.2.12", "1.2.14", "1.2.15", "1.2.16", "1.2.17"},
	}
	outside := map[string][]string{
		hsqldb: {"2.7.1", "2.7.3", "1.8.0.10"},
		log4j:  {"1.2.6", "1.3alpha-8"},
	}
	for id, w := range want {
		m := meta[id]
		if m.Role != RoleSink || m.Coverage != CoveragePartial || m.Review != ReviewReviewed {
			t.Errorf("%s: %+v, want sink/partial/reviewed", id, m)
		}
		for _, v := range checked[w.pkg] {
			got := pick(id, w.pkg, v)
			if len(got) != len(w.methods) {
				t.Errorf("%s %s = %v, want %v", id, v, got, w.methods)
				continue
			}
			for i := range got {
				if got[i] != w.methods[i] {
					t.Errorf("%s %s = %v, want %v", id, v, got, w.methods)
					break
				}
			}
		}
		for _, v := range outside[w.pkg] {
			if got := pick(id, w.pkg, v); len(got) != 0 {
				t.Errorf("%s %s outside the checked span = %v", id, v, got)
			}
		}
	}
	if got := pick("CVE-2023-26464", "org.apache.logging.log4j:log4j-core", "2.0-beta9"); len(got) != 0 {
		t.Errorf("CVE-2023-26464 log4j-core = %v, want no variant", got)
	}
}
