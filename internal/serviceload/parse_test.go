package serviceload

import "testing"

func TestParse_commentsAndWhitespaceAndBlankLines(t *testing.T) {
	content := []byte("" +
		"# a leading comment\n" +
		"\n" +
		"  com.example.FooProvider  \n" +
		"com.example.BarProvider # trailing comment\n" +
		"   \n" +
		"#\n")
	res := Parse("com.example.Foo", content)

	got := res.ValidProviders()
	want := []string{"com.example.FooProvider", "com.example.BarProvider"}
	if len(got) != len(want) {
		t.Fatalf("ValidProviders() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ValidProviders()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if len(res.Warnings) != 0 {
		t.Errorf("Warnings = %v, want none", res.Warnings)
	}
}

func TestParse_duplicatesAreFlaggedNotDropped(t *testing.T) {
	content := []byte("com.example.FooProvider\ncom.example.BarProvider\ncom.example.FooProvider\n")
	res := Parse("com.example.Foo", content)

	if len(res.Entries) != 3 {
		t.Fatalf("got %d entries, want 3 (duplicates are recorded, not silently removed from Entries)", len(res.Entries))
	}
	if !res.Entries[2].Duplicate {
		t.Errorf("third entry (repeat of FooProvider) not flagged Duplicate: %+v", res.Entries[2])
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly 1 (one duplicate)", res.Warnings)
	}

	got := res.ValidProviders()
	want := []string{"com.example.FooProvider", "com.example.BarProvider"}
	if len(got) != len(want) {
		t.Fatalf("ValidProviders() = %v, want %v (duplicate collapsed)", got, want)
	}
}

func TestParse_malformedNamesAreWarnedNotFatal(t *testing.T) {
	content := []byte("com.example.Good\n123.not.a.valid.name\ncom..example\n\ncom.example.AlsoGood\n")
	res := Parse("com.example.Foo", content)

	got := res.ValidProviders()
	want := []string{"com.example.Good", "com.example.AlsoGood"}
	if len(got) != len(want) {
		t.Fatalf("ValidProviders() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ValidProviders()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if len(res.Warnings) != 2 {
		t.Fatalf("Warnings = %v, want exactly 2 (one per malformed line)", res.Warnings)
	}
}

func TestParse_emptyFileProducesNoEntries(t *testing.T) {
	res := Parse("com.example.Foo", []byte(""))
	if len(res.Entries) != 0 || len(res.Warnings) != 0 {
		t.Errorf("Parse(empty) = %+v, want no entries and no warnings", res)
	}
}

func TestIsBinaryClassName(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"com.example.Foo", true},
		{"com.example.Foo$Inner", true},
		{"_underscore.$dollar", true},
		{"", false},
		{"com..example", false},
		{"com.example.", false},
		{".com.example", false},
		{"1com.example", false},
		{"com.1example", false},
	}
	for _, tc := range cases {
		if got := isBinaryClassName(tc.name); got != tc.want {
			t.Errorf("isBinaryClassName(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}
