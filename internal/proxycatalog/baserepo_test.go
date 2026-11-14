package proxycatalog

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func baseMeth(name, desc string, calls []classfile.Call, inits []classfile.InitRef) hierarchy.Method {
	return hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc}, Calls: calls, InitRefs: inits}
}

// TestBaseRepositoryEdges (x2, spring-data:base-repository): a call on an
// application repository interface reaches the body each base repository
// class runs for it, gated by the base class's creators — the methods that
// `new` it, or every method of a class naming it as a literal; a concrete
// Repository class nothing creates or names gets no edge, nor does a query
// method declared on the repository itself (no base-class method of that
// name), one annotated *.Query, or a default method. A repository method
// declared with concrete types (void save(Pet)) matches the base class's
// erased save(Object) by name and parameter assignability.
func TestBaseRepositoryEdges(t *testing.T) {
	const (
		repo = "org.springframework.data.repository.Repository"
		crud = "org.springframework.data.repository.CrudRepository"
	)
	iface := uint16(classfile.AccInterface | classfile.AccAbstract)
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: repo, AccessFlags: iface},
		{FQCN: crud, AccessFlags: iface, Interfaces: []string{repo}, Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "count", Descriptor: "()J", AccessFlags: classfile.AccAbstract}}}},
		{FQCN: "app.Pet", SuperClass: "java.lang.Object", IsApp: true},
		{FQCN: "app.PetRepository", AccessFlags: iface, Interfaces: []string{crud}, IsApp: true, Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "findByName", Descriptor: "(Ljava/lang/String;)Ljava/util/List;", AccessFlags: classfile.AccAbstract}},
			{Method: classfile.Method{Name: "save", Descriptor: "(Lapp/Pet;)V", AccessFlags: classfile.AccAbstract}},
			{Method: classfile.Method{Name: "findById", Descriptor: "(Ljava/lang/Long;)Lapp/Pet;", AccessFlags: classfile.AccAbstract,
				Annotations: []string{"org.springframework.data.jpa.repository.Query"}}},
			{Method: classfile.Method{Name: "findAll", Descriptor: "()Ljava/util/List;"}}}},
		{FQCN: "store.NewedBase", SuperClass: "java.lang.Object", Interfaces: []string{crud}, Methods: []hierarchy.Method{
			baseMeth("count", "()J", nil, nil),
			baseMeth("save", "(Ljava/lang/Object;)Ljava/lang/Object;", nil, nil),
			baseMeth("findById", "(Ljava/lang/Object;)Ljava/util/Optional;", nil, nil),
			baseMeth("findAll", "()Ljava/util/List;", nil, nil)}},
		{FQCN: "store.NamedBase", SuperClass: "java.lang.Object", Interfaces: []string{crud}, Methods: []hierarchy.Method{baseMeth("count", "()J", nil, nil)}},
		{FQCN: "store.Orphan", SuperClass: "java.lang.Object", Interfaces: []string{crud}, Methods: []hierarchy.Method{baseMeth("count", "()J", nil, nil)}},
		{FQCN: "store.Factory", SuperClass: "java.lang.Object", NamedClasses: []string{"store.NamedBase"}, Methods: []hierarchy.Method{
			baseMeth("target", "()Ljava/lang/Object;", nil, []classfile.InitRef{{Kind: classfile.InitNew, Owner: "store.NewedBase"}}),
			baseMeth("baseClass", "()Ljava/lang/Class;", nil, nil)}},
		{FQCN: "app.Service", SuperClass: "java.lang.Object", IsApp: true, Methods: []hierarchy.Method{
			baseMeth("run", "()V", []classfile.Call{
				{Kind: classfile.CallInterface, Owner: "app.PetRepository", Name: "count", Desc: "()J", PC: 1},
				{Kind: classfile.CallInterface, Owner: "app.PetRepository", Name: "findByName", Desc: "(Ljava/lang/String;)Ljava/util/List;", PC: 7},
				{Kind: classfile.CallInterface, Owner: "app.PetRepository", Name: "save", Desc: "(Lapp/Pet;)V", PC: 13},
				{Kind: classfile.CallInterface, Owner: "app.PetRepository", Name: "findById", Desc: "(Ljava/lang/Long;)Lapp/Pet;", PC: 19},
				{Kind: classfile.CallInterface, Owner: "app.PetRepository", Name: "findAll", Desc: "()Ljava/util/List;", PC: 25},
			}, nil)}},
	})
	r := Result{ByProxy: map[string]int{}}
	creators := map[string][]callgraph.MethodRef{"store.NewedBase": {{FQCN: "store.Factory", Name: "target", Desc: "()Ljava/lang/Object;"}}}
	edges := baseRepositoryEdges(h, []string{"app.PetRepository"}, creators, &r)
	got := map[string][]string{}
	for _, e := range edges {
		if e.CallerFQCN != "app.Service" {
			t.Errorf("unexpected edge %+v", e)
		}
		var cr []string
		for _, c := range e.Creators {
			cr = append(cr, c.FQCN+"#"+c.Name)
		}
		sort.Strings(cr)
		got[fmt.Sprintf("pc%d %s#%s%s", e.PC, e.TargetFQCN, e.TargetName, e.TargetDesc)] = cr
	}
	newed := []string{"store.Factory#target"}
	named := []string{"store.Factory#baseClass", "store.Factory#target"}
	want := map[string][]string{
		"pc1 store.NewedBase#count()J":                                    newed,
		"pc1 store.NamedBase#count()J":                                    named,
		"pc13 store.NewedBase#save(Ljava/lang/Object;)Ljava/lang/Object;": newed,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("edges (site target → creators) = %v, want %v", got, want)
	}
	if r.ByProxy[BaseRepositoryEntry] != 3 {
		t.Errorf("ByProxy = %v", r.ByProxy)
	}
}
