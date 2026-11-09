package main

import (
	"reflect"
	"testing"
)

// TestAppendNewRoots: web.xml roots join the application roots without
// duplicates (a web.xml class that is an application class is already
// rooted) and in their own order, so the BFS stays deterministic.
func TestAppendNewRoots(t *testing.T) {
	got, n := appendNewRoots([]uint32{3, 1, 2}, []uint32{2, 7, 5, 7})
	if want := []uint32{3, 1, 2, 7, 5}; !reflect.DeepEqual(got, want) || n != 2 {
		t.Errorf("got %v (%d new), want %v (2 new)", got, n, want)
	}
	if got, n := appendNewRoots([]uint32{1}, nil); !reflect.DeepEqual(got, []uint32{1}) || n != 0 {
		t.Errorf("no extra roots: got %v (%d)", got, n)
	}
}
