package analysis

import (
	"strings"
	"testing"
)

// TestSelect (ROADMAP B3 b3-3): a profile disables participants by ID;
// both lists keep the catalog's order, labels are id@version, and an ID
// the catalog does not know is an error naming it and the known ones.
func TestSelect(t *testing.T) {
	catalog := []Participant{
		{ID: "m:a", Version: "1", Kind: KindModel},
		{ID: "g:b", Version: "x2", Kind: KindGate},
		{ID: "m:c", Version: "3", Kind: KindModel},
	}
	active, disabled, err := Select(catalog, nil)
	if err != nil || len(active) != 3 || len(disabled) != 0 {
		t.Fatalf("no selection: %v %v %v", active, disabled, err)
	}
	active, disabled, err = Select(catalog, []string{"m:c", "g:b", "g:b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].ID != "m:a" || len(disabled) != 2 || disabled[0].ID != "g:b" || disabled[1].ID != "m:c" {
		t.Errorf("active %v, disabled %v", active, disabled)
	}
	if got := strings.Join(Labels(disabled), ","); got != "g:b@x2,m:c@3" {
		t.Errorf("labels = %s", got)
	}
	_, _, err = Select(catalog, []string{"m:a", "jdbc:object"})
	if err == nil || !strings.Contains(err.Error(), `"jdbc:object"`) || !strings.Contains(err.Error(), "m:a, g:b, m:c") {
		t.Errorf("unknown ID: %v", err)
	}
}
