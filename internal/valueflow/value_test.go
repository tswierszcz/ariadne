package valueflow

import "testing"

func TestMerge_identicalValuesCollapseToOne(t *testing.T) {
	a := Const("java.lang.String", "x")
	b := Const("java.lang.String", "x")
	got := Merge(a, b)
	if got.Kind != KindConst || got.Const != "x" {
		t.Errorf("Merge(identical) = %+v, want the same single constant", got)
	}
}

func TestMerge_eitherUnknownPropagates(t *testing.T) {
	known := Const("java.lang.String", "x")
	unknown := Unknown("some reason")
	if got := Merge(known, unknown); got.Kind != KindUnknown {
		t.Errorf("Merge(known, unknown) = %+v, want KindUnknown", got)
	}
	if got := Merge(unknown, known); got.Kind != KindUnknown {
		t.Errorf("Merge(unknown, known) = %+v, want KindUnknown", got)
	}
}

func TestMerge_setPlusNewCandidateGrows(t *testing.T) {
	set := Merge(Const("T", "a"), Const("T", "b")) // KindSet of 2
	grown := Merge(set, Const("T", "c"))
	if grown.Kind != KindSet || len(grown.Candidates) != 3 {
		t.Fatalf("Merge(set, new candidate) = %+v, want a 3-candidate KindSet", grown)
	}
}

func TestMerge_duplicateCandidateDoesNotGrowTheSet(t *testing.T) {
	set := Merge(Const("T", "a"), Const("T", "b"))
	same := Merge(set, Const("T", "a"))
	if len(same.Candidates) != 2 {
		t.Errorf("Merge(set, duplicate) = %+v, want still 2 candidates", same)
	}
}

func TestUnknown_requiresAReason(t *testing.T) {
	v := Unknown("")
	if v.Reason == "" {
		t.Errorf("Unknown(\"\") produced an empty Reason; the constructor must fill one in")
	}
}
