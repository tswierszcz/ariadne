package mavenversion

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestAgainstMaven checks Compare on every pair of the 197 versions in
// testdata/maven-3.9.12-order.tsv — the specification's examples plus
// every version of the reference apps' dependencies — against their rank
// in Maven 3.9.12's own ComparableVersion order (38809 comparisons).
func TestAgainstMaven(t *testing.T) {
	f, err := os.Open("testdata/maven-3.9.12-order.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	type ranked struct {
		rank int
		v    string
	}
	var vs []ranked
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "#") {
			continue
		}
		r, v, _ := strings.Cut(sc.Text(), "\t")
		n, _ := strconv.Atoi(r)
		vs = append(vs, ranked{n, v})
	}
	if len(vs) != 197 {
		t.Fatalf("read %d versions", len(vs))
	}
	bad := 0
	for _, a := range vs {
		for _, b := range vs {
			want := 0
			if a.rank < b.rank {
				want = -1
			} else if a.rank > b.rank {
				want = 1
			}
			if got := Compare(a.v, b.v); got != want {
				bad++
				if bad <= 25 {
					t.Errorf("Compare(%q, %q) = %d, Maven says %d", a.v, b.v, got, want)
				}
			}
		}
	}
	if bad > 0 {
		t.Errorf("%d of %d comparisons disagree with Maven", bad, len(vs)*len(vs))
	}
}

// render prints items the way Maven's ComparableVersion main prints tokens.
func render(l []item) string {
	var parts []string
	for _, it := range l {
		switch it.kind {
		case kNumber:
			parts = append(parts, it.num.String())
		case kString:
			if it.str != "" { // Maven's display omits release qualifiers
				parts = append(parts, it.str)
			}
		default:
			parts = append(parts, render(it.list))
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func TestTokensMatchMaven(t *testing.T) {
	b, err := os.ReadFile("testdata/maven-3.9.12-tokens.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		v, rest, _ := strings.Cut(line, " -> ")
		_, tokens, _ := strings.Cut(rest, "tokens: ")
		if got := render(parse(v)); got != tokens {
			t.Errorf("%s: tokens %s, Maven %s", v, got, tokens)
		}
	}
}

// TestRangesAgainstMaven: Range.Contains against Maven 3.9.12's
// VersionRange.containsVersion (testdata/maven-3.9.12-ranges.tsv).
func TestRangesAgainstMaven(t *testing.T) {
	b, err := os.ReadFile("testdata/maven-3.9.12-ranges.tsv")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		p := strings.Split(line, "\t")
		r, err := ParseRange(p[0])
		if err != nil {
			t.Fatalf("ParseRange(%q): %v", p[0], err)
		}
		if got := r.Contains(p[1]).String(); got != p[2] {
			t.Errorf("%s contains %s = %s, Maven says %s", p[0], p[1], got, p[2])
		}
		n++
	}
	if n != 396 {
		t.Errorf("read %d cases", n)
	}
}

func TestRegularAndUndetermined(t *testing.T) {
	for _, v := range []string{"1.0", "2.0.0.RC1", "31.1-jre", "1.1.4c", "3.27.0-GA", "1.0_1"} {
		if !Regular(v) {
			t.Errorf("%q should be regular", v)
		}
	}
	r, _ := ParseRange("[1.0,2.0)")
	for _, v := range []string{"1.5-SNAPSHOT", "", "v1.5", "latest", "1.5 beta", "${project.version}"} {
		if r.Contains(v) != Undetermined {
			t.Errorf("%q: want undetermined, got %s", v, r.Contains(v))
		}
	}
	for _, bad := range []string{"", "1.0", "[1.0", "[2.0,1.0]", "[,1.0]", "[1.0,2.0),", "[1,2)(3,4)", "(1.0)", "[1.0,1.5-SNAPSHOT]", "[1,2,3]"} {
		if _, err := ParseRange(bad); err == nil {
			t.Errorf("ParseRange(%q) accepted", bad)
		}
	}
}

// OSV ECOSYSTEM events, including unsorted events and last_affected.
func TestEvalOSV(t *testing.T) {
	ev := func(kv ...string) []map[string]string {
		var out []map[string]string
		for i := 0; i < len(kv); i += 2 {
			out = append(out, map[string]string{kv[i]: kv[i+1]})
		}
		return out
	}
	cases := []struct {
		typ    string
		events []map[string]string
		v      string
		want   Decision
	}{
		{"ECOSYSTEM", ev("introduced", "0", "fixed", "2.13.4"), "2.11.4", In},
		{"ECOSYSTEM", ev("introduced", "0", "fixed", "2.13.4"), "2.13.4", Out},
		{"ECOSYSTEM", ev("introduced", "2.0", "fixed", "2.3", "introduced", "0", "fixed", "1.5"), "1.6", Out},
		{"ECOSYSTEM", ev("fixed", "2.3", "introduced", "2.0", "fixed", "1.5", "introduced", "0"), "2.1", In},
		{"ECOSYSTEM", ev("introduced", "1.0", "last_affected", "1.4"), "1.4", In},
		{"ECOSYSTEM", ev("introduced", "1.0", "last_affected", "1.4"), "1.4.1", Out},
		{"ECOSYSTEM", ev("introduced", "0", "fixed", "9.0.46"), "9.0.100", Out},
		{"SEMVER", ev("introduced", "0"), "1.0", Undetermined},
		{"ECOSYSTEM", ev("introduced", "0", "fixed", "2.0"), "2.0-SNAPSHOT", Undetermined},
		{"ECOSYSTEM", ev("introduced", "0", "unknown", "2.0"), "1.0", Undetermined},
	}
	for _, c := range cases {
		if got := EvalOSV(c.typ, c.events, c.v); got != c.want {
			t.Errorf("%s %v %s = %s, want %s", c.typ, c.events, c.v, got, c.want)
		}
	}
}
