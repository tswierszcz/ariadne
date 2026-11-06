package relocate

import "testing"

func mustSet(t *testing.T, rules ...Rule) Set {
	t.Helper()
	s, err := New(rules)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestNameBoundaryAndLongest: a prefix matches only at a package
// boundary, in both name forms, and the longest rule wins.
func TestNameBoundaryAndLongest(t *testing.T) {
	s := mustSet(t, Rule{From: "org.objectweb", To: "x.org.objectweb"}, Rule{From: "org.objectweb.asm.tree", To: "y.tree"})
	cases := map[string]string{
		"org.objectweb.asm.Type":        "x.org.objectweb.asm.Type",
		"org.objectweb":                 "x.org.objectweb",
		"org.objectwebx.Y":              "org.objectwebx.Y",
		"org.objectweb.asm.tree.Insn$1": "y.tree.Insn$1",
		"org/objectweb/asm/Type":        "x/org/objectweb/asm/Type",
		"org/objectwebx/Y":              "org/objectwebx/Y",
		"com.other.T":                   "com.other.T",
	}
	for in, want := range cases {
		if got := s.Name(in); got != want {
			t.Errorf("Name(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestDescriptorAndSelector: class names inside descriptors (arrays,
// nested classes, several parameters) and selectors are rewritten;
// primitives and other packages are not.
func TestDescriptorAndSelector(t *testing.T) {
	s := mustSet(t, Rule{From: "a.b", To: "shaded.a.b"})
	if got, want := s.Descriptor("([[La/b/C$D;ILa/bx/E;Ljava/lang/String;)[La/b/F;"),
		"([[Lshaded/a/b/C$D;ILa/bx/E;Ljava/lang/String;)[Lshaded/a/b/F;"; got != want {
		t.Errorf("Descriptor = %q, want %q", got, want)
	}
	for in, want := range map[string]string{
		"a.b.C#m":                  "shaded.a.b.C#m",
		"a.b.C$D#<init>(La/b/C;)V": "shaded.a.b.C$D#<init>(Lshaded/a/b/C;)V",
		"x.Y#m(La/b/C;)La/b/C;":    "x.Y#m(Lshaded/a/b/C;)Lshaded/a/b/C;",
		"a.bx.C#m":                 "a.bx.C#m",
	} {
		if got := s.Selector(in); got != want {
			t.Errorf("Selector(%q) = %q, want %q", in, got, want)
		}
	}
	if got := s.String("a/b/Resource"); got != "shaded/a/b/Resource" {
		t.Errorf("String = %q", got)
	}
	if got := s.String("a.b.C is broken"); got != "a.b.C is broken" {
		t.Errorf("a sentence was rewritten: %q", got)
	}
}

// TestInverseAndConflicts: the inverse maps back; one package relocated
// to two places, or a non-package name, is rejected.
func TestInverseAndConflicts(t *testing.T) {
	s := mustSet(t, Rule{From: "a.b", To: "z.a.b"})
	if got := s.Inverse().Name(s.Name("a.b.C")); got != "a.b.C" {
		t.Errorf("round trip = %q", got)
	}
	if _, err := New([]Rule{{From: "a.b", To: "x"}, {From: "a.b", To: "y"}}); err == nil {
		t.Error("conflicting rules accepted")
	}
	if _, err := Merge(s, mustSet(t, Rule{From: "a.b", To: "q"})); err == nil {
		t.Error("conflicting merge accepted")
	}
	for _, bad := range []Rule{{From: "a.*", To: "b"}, {From: "${p}", To: "b"}, {From: "a.", To: "b"}, {From: "", To: "b"}} {
		if _, err := New([]Rule{bad}); err == nil {
			t.Errorf("rule %+v accepted", bad)
		}
	}
	if r, ok := s.Under("z/a/b/C"); !ok || r.From != "a.b" {
		t.Errorf("Under = %+v %v", r, ok)
	}
	if _, ok := s.Under("z.a.bc.C"); ok {
		t.Error("Under matched across a package boundary")
	}
}
