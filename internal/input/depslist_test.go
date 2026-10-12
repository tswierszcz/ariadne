package input

import (
	"strings"
	"testing"
)

func TestParseDepsList(t *testing.T) {
	// Real-world sample, trimmed from petclinic's and WebGoat's deps.txt:
	// a plain header, lines with and without ANSI color codes, lines with
	// and without a classifier, and the " -- module x [auto]"/"(auto)"
	// suffixes newer dependency-plugin versions append.
	input := "" +
		"\n" +
		"The following files have been resolved:\n" +
		"   org.springframework.boot:spring-boot-starter:jar:2.4.5:compile -- module spring.boot.starter [auto]\n" +
		"\x1b[36m   ch.qos.logback:logback-classic:jar:1.2.3:compile\x1b[0;1;33m (auto)\x1b[m\n" +
		"   com.thoughtworks.xstream:xstream:jar:1.4.5:compile\n" +
		"   org.some:thing:jar:tests:1.0:test\n" +
		"\n"

	got, err := ParseDepsList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseDepsList: %v", err)
	}

	want := []GAV{
		{Group: "org.springframework.boot", Artifact: "spring-boot-starter", Version: "2.4.5"},
		{Group: "ch.qos.logback", Artifact: "logback-classic", Version: "1.2.3"},
		{Group: "com.thoughtworks.xstream", Artifact: "xstream", Version: "1.4.5"},
		{Group: "org.some", Artifact: "thing", Classifier: "tests", Version: "1.0"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d GAVs, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GAV[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseDepsList_empty(t *testing.T) {
	got, err := ParseDepsList(strings.NewReader("The following files have been resolved:\n   None\n"))
	if err != nil {
		t.Fatalf("ParseDepsList: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d GAVs, want 0: %+v", len(got), got)
	}
}
