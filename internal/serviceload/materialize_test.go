package serviceload

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func ctorMethod(name string, access uint16) hierarchy.Method {
	return hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: "()V", AccessFlags: access}}
}

func testHierarchy(t *testing.T) *hierarchy.Hierarchy {
	t.Helper()
	return hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.GoodProvider", Methods: []hierarchy.Method{ctorMethod("<init>", classfile.AccPublic)}},
		{FQCN: "test.AbstractProvider", AccessFlags: classfile.AccAbstract, Methods: []hierarchy.Method{ctorMethod("<init>", classfile.AccPublic)}},
		{FQCN: "test.NoCtorProvider", Methods: []hierarchy.Method{ctorMethod("doStuff", classfile.AccPublic)}},
		{FQCN: "test.PrivateCtorProvider", Methods: []hierarchy.Method{ctorMethod("<init>", 0)}},
	})
}

func TestFindMaterializations_resolvesPublicNoArgConstructor(t *testing.T) {
	h := testHierarchy(t)
	sites := []ServiceLoadSite{{Interface: "test.MyService"}}
	regs := []ServiceRegistration{{Interface: "test.MyService", ProviderClass: "test.GoodProvider", ProviderFound: true}}

	got := FindMaterializations(h, sites, regs)
	if len(got) != 1 {
		t.Fatalf("got %+v, want exactly 1 materialization", got)
	}
	if got[0].ConstructorMethod != "test.GoodProvider#<init>()V" {
		t.Errorf("ConstructorMethod = %q, want test.GoodProvider#<init>()V", got[0].ConstructorMethod)
	}
	if got[0].Gap != "" {
		t.Errorf("Gap = %q, want empty", got[0].Gap)
	}
}

func TestFindMaterializations_gapCases(t *testing.T) {
	h := testHierarchy(t)
	sites := []ServiceLoadSite{{Interface: "test.MyService"}}
	cases := []struct {
		provider string
	}{
		{"test.AbstractProvider"},
		{"test.NoCtorProvider"},
		{"test.PrivateCtorProvider"},
	}
	for _, c := range cases {
		t.Run(c.provider, func(t *testing.T) {
			regs := []ServiceRegistration{{Interface: "test.MyService", ProviderClass: c.provider, ProviderFound: true}}
			got := FindMaterializations(h, sites, regs)
			if len(got) != 1 {
				t.Fatalf("got %+v, want exactly 1 materialization", got)
			}
			if got[0].ConstructorMethod != "" {
				t.Errorf("ConstructorMethod = %q, want empty (a gap case)", got[0].ConstructorMethod)
			}
			if got[0].Gap == "" {
				t.Errorf("Gap is empty, want a non-empty explanation")
			}
		})
	}
}

// TestFindMaterializations_unusedInterfaceIsNotAGap is S2's own declared
// negative case: a registration whose interface no resolved site ever
// names gets no materialization at all — that is the correct result,
// not a gap to report.
func TestFindMaterializations_unusedInterfaceIsNotAGap(t *testing.T) {
	h := testHierarchy(t)
	sites := []ServiceLoadSite{{Interface: "test.SomeOtherService"}}
	regs := []ServiceRegistration{{Interface: "test.MyService", ProviderClass: "test.GoodProvider", ProviderFound: true}}

	got := FindMaterializations(h, sites, regs)
	if len(got) != 0 {
		t.Errorf("got %+v, want none (the interface was never resolved at a load site)", got)
	}
}

// TestFindMaterializations_duplicateRegistrationMaterializedOnce is S2's
// own "powtórzony wpis" case: the same provider registered for the same
// interface more than once (e.g. from two separate META-INF/services
// files) must not produce two materializations or two edges.
func TestFindMaterializations_duplicateRegistrationMaterializedOnce(t *testing.T) {
	h := testHierarchy(t)
	sites := []ServiceLoadSite{{Interface: "test.MyService"}}
	regs := []ServiceRegistration{
		{Interface: "test.MyService", ProviderClass: "test.GoodProvider", ProviderFound: true, SourceJar: "a.jar"},
		{Interface: "test.MyService", ProviderClass: "test.GoodProvider", ProviderFound: true, SourceJar: "b.jar"},
	}

	got := FindMaterializations(h, sites, regs)
	if len(got) != 1 {
		t.Fatalf("got %+v, want exactly 1 (same provider/interface pair, deduplicated)", got)
	}
}

func TestEdgeTargets_joinsSitesAndMaterializations(t *testing.T) {
	sites := []ServiceLoadSite{
		{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 5, Interface: "test.MyService"},
		{CallerFQCN: "test.Other", CallerName: "n", CallerDesc: "()V", PC: 9, Interface: ""}, // unresolved: contributes nothing
	}
	materializations := []ProviderMaterialization{
		{Registration: ServiceRegistration{Interface: "test.MyService", ProviderClass: "test.GoodProvider"}, ConstructorMethod: "test.GoodProvider#<init>()V"},
		{Registration: ServiceRegistration{Interface: "test.MyService", ProviderClass: "test.GapProvider"}, Gap: "no public no-argument constructor"},
	}

	got := EdgeTargets(sites, materializations)
	if len(got) != 1 {
		t.Fatalf("got %+v, want exactly 1 edge (the gap provider contributes nothing, the unresolved site contributes nothing)", got)
	}
	want := MaterializedEdge{CallerFQCN: "test.Caller", CallerName: "m", CallerDesc: "()V", PC: 5, ProviderFQCN: "test.GoodProvider"}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
}
