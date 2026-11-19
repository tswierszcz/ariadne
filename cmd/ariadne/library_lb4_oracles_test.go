package main

import (
	"bufio"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// TestLibraryLB4Oracles (LIBRARY_MODE LB4) checks the recorded JVM
// oracles of the facade, factory/subclass and SPI libraries
// (testdata/library-lb4/EXPECTED.MD, scripts/library-lb4-jvm-oracle.sh)
// against the verdicts recorded with them (mapped.txt): no mapped method
// of a NOT_REACHABLE finding ever ran (no false negative shown), and the
// REACHABLE findings whose mapped code ran are exactly the documented
// ones. The analysis itself needs work/libs and is not re-run here.
func TestLibraryLB4Oracles(t *testing.T) {
	notShown := map[string][]string{
		"configuration2": nil,
		"jackson-yaml":   {"CVE-2026-68494", "CVE-2026-89425"},
		"jjwt-jackson":   {"CVE-2026-68494"},
	}
	for lib, wantNotShown := range notShown {
		dir := "../../testdata/library-lb4/" + lib + "/"
		ran := map[string]bool{}
		for _, l := range lines(t, dir+"observed.txt") {
			ran[l] = true
		}
		verdict := map[string]string{}
		shown := map[string]bool{}
		for _, l := range lines(t, dir+"mapped.txt") {
			f := strings.Split(l, "\t")
			if len(f) != 3 {
				t.Fatalf("%s mapped.txt: bad line %q", lib, l)
			}
			id, v, sel := f[0], f[1], f[2]
			verdict[id] = v
			for r := range ran {
				if r == sel || !strings.Contains(sel, "(") && strings.HasPrefix(r, sel+"(") {
					shown[id] = true
					if v != "REACHABLE" {
						t.Errorf("%s: %s is %s but %s ran on the JVM", lib, id, v, r)
					}
				}
			}
		}
		var got []string
		for id, v := range verdict {
			if v == "REACHABLE" && !shown[id] {
				got = append(got, id)
			}
		}
		sort.Strings(got)
		if !reflect.DeepEqual(got, wantNotShown) && !(len(got) == 0 && len(wantNotShown) == 0) {
			t.Errorf("%s: REACHABLE findings the oracle does not show = %v, want %v", lib, got, wantNotShown)
		}
	}
}

func lines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if l := sc.Text(); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out
}
