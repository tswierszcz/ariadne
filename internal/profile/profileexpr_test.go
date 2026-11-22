package profile

import "testing"

// TestEvalProfileExpression (FRAMEWORKS F5 f5-3): Spring 5.1+ profile
// expressions, as Profiles.of evaluates them.
func TestEvalProfileExpression(t *testing.T) {
	active := map[string]bool{"eu": true, "cloud": true}
	for _, c := range []struct {
		expr  string
		v, ok bool
	}{
		{"eu", true, true},
		{"us", false, true},
		{"!eu", false, true},
		{"!us", true, true},
		{"eu & cloud", true, true},
		{"eu & us", false, true},
		{"eu | us", true, true},
		{"us | legacy", false, true},
		{"!(eu & cloud)", false, true},
		{"(eu | us) & !legacy", true, true},
		{"eu&cloud", true, true},
		{"eu & cloud | us", false, false}, // mixed without parentheses: Spring rejects it
		{"(eu", false, false},
		{"", false, false},
		{"eu &", false, false},
	} {
		v, ok := evalProfileExpression(c.expr, active)
		if v != c.v || ok != c.ok {
			t.Errorf("%q: got %v,%v want %v,%v", c.expr, v, ok, c.v, c.ok)
		}
	}
}

// TestBeanName_EffectiveName: the explicit name, else the decapitalized
// simple name (java.beans.Introspector: two leading capitals stay).
func TestBeanName_EffectiveName(t *testing.T) {
	for _, c := range []struct {
		b    BeanName
		want string
	}{
		{BeanName{FQCN: "fx.beans.ColdStore"}, "coldStore"},
		{BeanName{FQCN: "fx.beans.ColdStore", Name: "cold"}, "cold"},
		{BeanName{FQCN: "fx.URLStore"}, "URLStore"},
		{BeanName{FQCN: "fx.Outer$Inner"}, "outer.Inner"},
		{BeanName{FQCN: "A"}, "a"},
	} {
		if got := c.b.EffectiveName(); got != c.want {
			t.Errorf("%+v: %q, want %q", c.b, got, c.want)
		}
	}
}
