package libapi

import (
	"fmt"
	"reflect"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const suppliedFixture = "../../testdata/library-supplied/"

// TestSuppliedSites (lb2-7) follows testdata/library-supplied/EXPECTED.MD:
// the StringLookup#lookup calls on a root's parameter (y) and on a field
// written only from a root's parameter (z) are caller-supplied; the
// library's own lookup (aa), a field also set to the library's default
// (bb) are not, nor is handing one to a dependency constructor (cc).
func TestSuppliedSites(t *testing.T) {
	h, _, err := hierarchy.Build(suppliedFixture+"lib.jar", []input.DepJar{
		{Path: fixture + "deps/commons-text-1.9.jar"}, {Path: fixture + "deps/commons-lang3-3.11.jar"}})
	if err != nil {
		t.Fatal(err)
	}
	g := callgraph.Build(h, callgraph.DefaultSkip)
	roots, _, _ := Roots(g, h, nil, nil, Options{})
	keys := map[string]bool{}
	for _, r := range roots {
		keys[g.Key(r)] = true
	}
	isRoot := func(owner, name, desc string) bool { return keys[owner+"#"+name+desc] }
	sites, st := SuppliedSites(h, isRoot)
	var got []string
	for _, s := range sites {
		got = append(got, fmt.Sprintf("%s#%s %s field=%v", s.CallerFQCN, s.CallerName, s.Called, s.Field))
	}
	want := []string{
		"sp.supplied.Holder#get org.apache.commons.text.lookup.StringLookup#lookup(Ljava/lang/String;)Ljava/lang/String; field=true",
		"sp.supplied.Resolver#resolve org.apache.commons.text.lookup.StringLookup#lookup(Ljava/lang/String;)Ljava/lang/String; field=false",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sites\n got %q\nwant %q", got, want)
	}
	if st.Fields != 1 || st.Unanalyzed != 0 {
		t.Errorf("stats %+v, want 1 field (Holder.lookup), 0 unanalyzed", st)
	}
	// Not a root: a parameter is the library's own value.
	if s, _ := SuppliedSites(h, func(string, string, string) bool { return false }); len(s) != 0 {
		t.Errorf("without roots: %d sites, want 0", len(s))
	}
}
