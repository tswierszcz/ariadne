package jdkapi

import (
	"reflect"
	"testing"

	"ariadne/internal/hierarchy"
)

// TestSubinterfaces: the table is read transitively, from the parent down.
func TestSubinterfaces(t *testing.T) {
	cases := map[string][]string{
		"java.sql.Statement":         {"java.sql.CallableStatement", "java.sql.PreparedStatement"},
		"java.sql.PreparedStatement": {"java.sql.CallableStatement"},
		"java.sql.CallableStatement": nil,
		"java.sql.ResultSet":         {"javax.sql.RowSet"},
		"javax.sql.CommonDataSource": {"javax.sql.ConnectionPoolDataSource", "javax.sql.DataSource", "javax.sql.XADataSource"},
		"java.lang.Runnable":         nil,
	}
	for iface, want := range cases {
		if got := Subinterfaces(iface); !reflect.DeepEqual(got, want) {
			t.Errorf("Subinterfaces(%s) = %v, want %v", iface, got, want)
		}
	}
	w := Subinterfaces("java.sql.Wrapper")
	for _, must := range []string{"java.sql.Connection", "java.sql.Statement", "java.sql.PreparedStatement", "java.sql.CallableStatement", "java.sql.ResultSet", "javax.sql.DataSource", "javax.sql.RowSet"} {
		found := false
		for _, s := range w {
			found = found || s == must
		}
		if !found {
			t.Errorf("Subinterfaces(java.sql.Wrapper) = %v lacks %s", w, must)
		}
	}
}

// TestSubtypes: a class implementing only PreparedStatement (or
// CallableStatement) is a subtype of Statement; one implementing Statement
// directly is listed once.
func TestSubtypes(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "d.Plain", SuperClass: "java.lang.Object", Interfaces: []string{"java.sql.Statement"}},
		{FQCN: "d.Prepared", SuperClass: "java.lang.Object", Interfaces: []string{"java.sql.PreparedStatement"}},
		{FQCN: "d.Callable", SuperClass: "java.lang.Object", Interfaces: []string{"java.sql.CallableStatement"}},
		{FQCN: "d.Both", SuperClass: "java.lang.Object", Interfaces: []string{"java.sql.Statement", "java.sql.PreparedStatement"}},
	})
	got := map[string]int{}
	for _, s := range Subtypes(h, "java.sql.Statement") {
		got[s]++
	}
	want := map[string]int{"d.Plain": 1, "d.Prepared": 1, "d.Callable": 1, "d.Both": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Subtypes(Statement) = %v, want %v", got, want)
	}
	if got := Subtypes(h, "java.sql.CallableStatement"); !reflect.DeepEqual(got, []string{"d.Callable"}) {
		t.Errorf("Subtypes(CallableStatement) = %v", got)
	}
}
