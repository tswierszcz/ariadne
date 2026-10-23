package models

import (
	"testing"

	"ariadne/internal/input"
)

func TestRegistry_lookupMatchesExactSignatureOnly(t *testing.T) {
	r := NewRegistry()
	r.Register(&Model{
		ID:         "test:foo",
		Signatures: []Signature{{Owner: "test.Foo", Name: "bar", Desc: "()V"}},
	})

	if got := r.Lookup(Signature{Owner: "test.Foo", Name: "bar", Desc: "()V"}, nil); len(got) != 1 {
		t.Errorf("Lookup(exact match) = %+v, want exactly 1", got)
	}
	// A same-named method on a different class is never matched —
	// B3's own "pinned to the concrete call, not every method of the
	// same name" requirement.
	if got := r.Lookup(Signature{Owner: "test.Other", Name: "bar", Desc: "()V"}, nil); len(got) != 0 {
		t.Errorf("Lookup(different owner) = %+v, want none", got)
	}
	// A different descriptor (a different overload) is never matched.
	if got := r.Lookup(Signature{Owner: "test.Foo", Name: "bar", Desc: "(I)V"}, nil); len(got) != 0 {
		t.Errorf("Lookup(different descriptor) = %+v, want none", got)
	}
}

func TestRegistry_artifactAndVersionGateTheMatch(t *testing.T) {
	r := NewRegistry()
	sig := Signature{Owner: "org.example.Foo", Name: "bar", Desc: "()V"}
	r.Register(&Model{
		ID: "test:foo", Artifact: "org.example:foo-lib",
		Versions:   VersionRange{Min: "2.0.0", Max: "2.9.9"},
		Signatures: []Signature{sig},
	})

	inRange := &input.GAV{Group: "org.example", Artifact: "foo-lib", Version: "2.3.1"}
	if got := r.Lookup(sig, inRange); len(got) != 1 {
		t.Errorf("Lookup(in-range version) = %+v, want exactly 1", got)
	}

	tooOld := &input.GAV{Group: "org.example", Artifact: "foo-lib", Version: "1.9.0"}
	if got := r.Lookup(sig, tooOld); len(got) != 0 {
		t.Errorf("Lookup(too-old version) = %+v, want none", got)
	}

	tooNew := &input.GAV{Group: "org.example", Artifact: "foo-lib", Version: "3.0.0"}
	if got := r.Lookup(sig, tooNew); len(got) != 0 {
		t.Errorf("Lookup(too-new version) = %+v, want none", got)
	}

	wrongArtifact := &input.GAV{Group: "org.example", Artifact: "other-lib", Version: "2.3.1"}
	if got := r.Lookup(sig, wrongArtifact); len(got) != 0 {
		t.Errorf("Lookup(wrong artifact) = %+v, want none", got)
	}

	// No GAV at all (an application class) never matches a
	// Model with an Artifact restriction — never guessed.
	if got := r.Lookup(sig, nil); len(got) != 0 {
		t.Errorf("Lookup(nil gav, artifact-restricted model) = %+v, want none", got)
	}
}

func TestRegistry_noArtifactModelMatchesRegardlessOfGAV(t *testing.T) {
	r := NewRegistry()
	sig := Signature{Owner: "java.lang.Class", Name: "forName", Desc: "(Ljava/lang/String;)Ljava/lang/Class;"}
	r.Register(&Model{ID: "jdk:class-fornamee", Signatures: []Signature{sig}})

	if got := r.Lookup(sig, nil); len(got) != 1 {
		t.Errorf("Lookup(no-artifact model, nil gav) = %+v, want exactly 1", got)
	}
	gav := &input.GAV{Group: "anything", Artifact: "whatsoever", Version: "1.0"}
	if got := r.Lookup(sig, gav); len(got) != 1 {
		t.Errorf("Lookup(no-artifact model, any gav) = %+v, want exactly 1", got)
	}
}

// TestInclusiveVersionInRange_unparseableVersionWidensRatherThanExcludes
// confirms B0's own "a limit widens to unknown, never silently
// discards a candidate" applied to version comparison: a version this
// package's own minimal comparator can't parse is treated as inside
// the range, not excluded by a guess it cannot actually prove.
func TestInclusiveVersionInRange_unparseableVersionWidensRatherThanExcludes(t *testing.T) {
	r := VersionRange{Min: "1.0.0", Max: "2.0.0"}
	if !inclusiveVersionInRange("5.3.21-RELEASE", r) {
		t.Error("inclusiveVersionInRange(unparseable qualifier) = false, want true (widen, don't exclude)")
	}
}

func TestInclusiveVersionInRange_boundaryValuesAreInclusive(t *testing.T) {
	r := VersionRange{Min: "1.2.0", Max: "1.2.5"}
	if !inclusiveVersionInRange("1.2.0", r) {
		t.Error("inclusiveVersionInRange(min boundary) = false, want true")
	}
	if !inclusiveVersionInRange("1.2.5", r) {
		t.Error("inclusiveVersionInRange(max boundary) = false, want true")
	}
	if inclusiveVersionInRange("1.1.9", r) {
		t.Error("inclusiveVersionInRange(just below min) = true, want false")
	}
	if inclusiveVersionInRange("1.2.6", r) {
		t.Error("inclusiveVersionInRange(just above max) = true, want false")
	}
}

// TestInclusiveVersionInRange_missingTrailingSegmentReadsAsZero
// confirms "5.3" == "5.3.0" for comparison purposes, the same
// convention Maven's own version comparator uses.
func TestInclusiveVersionInRange_missingTrailingSegmentReadsAsZero(t *testing.T) {
	r := VersionRange{Min: "5.3", Max: "5.3"}
	if !inclusiveVersionInRange("5.3.0", r) {
		t.Error("inclusiveVersionInRange(5.3.0 against 5.3..5.3) = false, want true")
	}
}
