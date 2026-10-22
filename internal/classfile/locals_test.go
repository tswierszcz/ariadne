package classfile

import (
	"reflect"
	"testing"
)

// TestArgumentSlots_wideTypes is DATA_FLOW D1's own "wide types" example:
// a long/double parameter occupies two local-variable slots, shifting
// every later parameter's slot number by one more than its own index
// would suggest.
func TestArgumentSlots_wideTypes(t *testing.T) {
	// (ILjava/lang/String;JD)V — int, String, long, double, static.
	got, err := ArgumentSlots("(ILjava/lang/String;JD)V", true)
	if err != nil {
		t.Fatalf("ArgumentSlots: %v", err)
	}
	want := []Argument{
		{Slot: 0, Width: 1, Type: "I"},
		{Slot: 1, Width: 1, Type: "Ljava/lang/String;"},
		{Slot: 2, Width: 2, Type: "J"}, // occupies slots 2 AND 3
		{Slot: 4, Width: 2, Type: "D"}, // starts at 4, not 3, because J took two slots
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ArgumentSlots = %+v, want %+v", got, want)
	}
}

func TestArgumentSlots_instanceMethodReservesThis(t *testing.T) {
	got, err := ArgumentSlots("(I)V", false)
	if err != nil {
		t.Fatalf("ArgumentSlots: %v", err)
	}
	want := []Argument{{Slot: 1, Width: 1, Type: "I"}} // slot 0 is `this`
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ArgumentSlots = %+v, want %+v", got, want)
	}
}

func TestArgumentSlots_arrayOfWideTypeIsOneSlot(t *testing.T) {
	got, err := ArgumentSlots("([J[D)V", true)
	if err != nil {
		t.Fatalf("ArgumentSlots: %v", err)
	}
	want := []Argument{
		{Slot: 0, Width: 1, Type: "[J"}, // a reference, not a wide primitive
		{Slot: 1, Width: 1, Type: "[D"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ArgumentSlots = %+v, want %+v", got, want)
	}
}

func TestArgumentSlots_noParameters(t *testing.T) {
	got, err := ArgumentSlots("()V", true)
	if err != nil {
		t.Fatalf("ArgumentSlots: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ArgumentSlots(()V) = %v, want none", got)
	}
}

func TestArgumentSlots_malformedDescriptor(t *testing.T) {
	cases := []string{"", "I)V", "(Q)V", "(L", "(["}
	for _, d := range cases {
		if _, err := ArgumentSlots(d, true); err == nil {
			t.Errorf("ArgumentSlots(%q) succeeded, want an error", d)
		}
	}
}
