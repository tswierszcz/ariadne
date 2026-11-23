package component

import (
	"reflect"
	"testing"
)

// TestChooseNewest (LIBRARY_MODE LB3 lb3-2): the "newest" strategy picks
// the highest published version inside a declared range and the newest of
// the analyzed version's major.minor line, never the analyzed version.
func TestChooseNewest(t *testing.T) {
	published := []string{"3.9", "3.10", "3.11", "3.12.0", "3.17.0", "3.18.0", "3.19.0"}
	type pick struct {
		v  string
		by []string
	}
	for _, c := range []struct {
		name, have, declared string
		want                 []pick
	}{
		{"range below a fix", "3.11", "[3.11,3.18.0)", []pick{{"3.17.0", []string{ChosenRangeMax}}}},
		{"range up to the newest", "3.11", "[3.11,)", []pick{{"3.19.0", []string{ChosenRangeMax}}}},
		{"range max is the analyzed version", "3.17.0", "[3.11,3.18.0)", nil},
		{"exact version, line has nothing newer", "3.11", "3.11", nil},
		{"line latest", "3.12.0", "3.12.0", nil},
	} {
		var got []pick
		for _, a := range chooseNewest(published, c.have, c.declared) {
			got = append(got, pick{a.Version, a.ChosenBy})
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	// Both reasons for one version are listed once.
	got := chooseNewest([]string{"2.0.0", "2.0.1", "2.0.5"}, "2.0.1", "[2.0,2.1)")
	if len(got) != 1 || got[0].Version != "2.0.5" || !reflect.DeepEqual(got[0].ChosenBy, []string{ChosenRangeMax, ChosenLineLatest}) {
		t.Errorf("range max and line latest at once: %+v", got)
	}
}

func TestMajorMinor(t *testing.T) {
	for v, want := range map[string]string{"3.11": "3.11", "3.12.0": "3.12", "2.0.1-RC1": "2.0", "5": "", "a.b": ""} {
		got, ok := majorMinor(v)
		if (want == "") == ok || got != want {
			t.Errorf("majorMinor(%q) = %q, %v; want %q", v, got, ok, want)
		}
	}
}
