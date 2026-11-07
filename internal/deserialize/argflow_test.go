package deserialize

import (
	"reflect"
	"testing"

	"ariadne/internal/flowsum"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

// TestArgumentFlow (DATA_FLOW D3 consumer): the stream an entry reads is
// `new ObjectInputStream(in)` with `in` the method's parameter 0, so the
// data read depends on parameter 0 — through the constructor that wraps
// it, not by a name match.
func TestArgumentFlow(t *testing.T) {
	h, _, err := hierarchy.Build(oracleDir+"scenarios.jar", []input.DepJar{{Path: oracleDir + "lib.jar"}, {Path: oracleDir + "shaded.jar"}})
	if err != nil {
		t.Fatal(err)
	}
	sites := FindSites(h)
	e := flowsum.New(h, flowsum.DefaultOptions)
	defer e.Close()
	e.Summarize(SummaryRoots(sites))
	AttachArgumentFlow(sites, e)
	if len(sites) != 2 {
		t.Fatalf("%d sites", len(sites))
	}
	for _, s := range sites {
		if !reflect.DeepEqual(s.DataFromRaw, []string{"param:0"}) || !reflect.DeepEqual(s.DataFrom, []string{"parameter 0 (java.io.InputStream)"}) || len(s.DataFlowIncomplete) > 0 {
			t.Errorf("%s: data from %v (%v), incomplete %v", s.CallerName, s.DataFromRaw, s.DataFrom, s.DataFlowIncomplete)
		}
	}
}
