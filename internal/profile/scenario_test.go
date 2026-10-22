package profile

import "testing"

// TestDetectAll_separatesPresenceFromActiveSelection is FRAMEWORKS F0's
// own "Gotowe": library availability and the active scenario are
// separate pieces of information — two profiles can both be detected
// (and merged by Select into one "a+b" active Profile), but DetectAll
// must still report each one's own individual Present fact, and a
// profile whose marker is absent must still appear with Present=false,
// not be silently omitted.
func TestDetectAll_separatesPresenceFromActiveSelection(t *testing.T) {
	var c Catalog
	c.put(Profile{Name: "a", Detect: []string{"fixture.App"}})
	c.put(Profile{Name: "b", Detect: []string{"fixture.Util"}})
	c.put(Profile{Name: "c", Detect: []string{"no.such.Class"}})
	c.put(Profile{Name: "generic"})

	h, _ := buildFixture(t)
	got := DetectAll(c, h)

	want := map[string]bool{"a": true, "b": true, "c": false}
	if len(got) != len(want) {
		t.Fatalf("DetectAll = %+v, want exactly %d entries (generic excluded)", got, len(want))
	}
	for _, d := range got {
		if d.Present != want[d.Name] {
			t.Errorf("Detection[%q].Present = %v, want %v", d.Name, d.Present, want[d.Name])
		}
	}

	// Confirm this is genuinely a separate fact from Select's own
	// merged result, not a duplicate of it.
	selected, err := Select(c, "auto", h)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if selected.Name != "a+b" {
		t.Fatalf("Select merged Name = %q, want \"a+b\"", selected.Name)
	}
}

func TestDetectAll_excludesGeneric(t *testing.T) {
	var c Catalog
	c.put(Profile{Name: "generic"})
	h, _ := buildFixture(t)

	if got := DetectAll(c, h); len(got) != 0 {
		t.Errorf("DetectAll = %+v, want none (generic is never a detection target)", got)
	}
}
