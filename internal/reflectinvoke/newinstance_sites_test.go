package reflectinvoke

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

func TestFindNewInstanceSites(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "test.Caller", Methods: []hierarchy.Method{
			{Method: classfile.Method{Name: "m", Descriptor: "()V"}, Calls: []classfile.Call{
				{Owner: constructorOwner, Name: "newInstance", Desc: newInstanceDesc, PC: 11},
				{Owner: constructorOwner, Name: "setAccessible", Desc: "(Z)V", PC: 3},
				{Owner: "java.lang.String", Name: "valueOf", Desc: "(I)Ljava/lang/String;", PC: 1},
			}},
		}},
	})

	sites := FindNewInstanceSites(h)
	if len(sites) != 1 {
		t.Fatalf("sites = %+v, want exactly 1", sites)
	}
	if sites[0].PC != 11 || sites[0].CallerFQCN != "test.Caller" {
		t.Errorf("sites[0] = %+v, want pc=11 on test.Caller", sites[0])
	}
}
