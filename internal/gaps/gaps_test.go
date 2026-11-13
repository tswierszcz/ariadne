package gaps

import (
	"os"
	"strings"
	"testing"

	"ariadne/internal/alloc"
	"ariadne/internal/runtimeevents"
	"ariadne/internal/samcall"
)

// TestEveryProducerCodeMapped: every gap code the analyzers define maps
// to a mechanism, so no gap falls back to "unclassified" silently. The
// lists are each analyzer's own constants; the site kinds are the ones
// cmd/ariadne's collectGapSites and rtaGapSites emit.
func TestEveryProducerCodeMapped(t *testing.T) {
	produced := map[string][]string{
		SourceSite: {"reflection", "methodhandle", "callback", "serviceloader", "deserialization", "allocation"},
		SourceAllocation: {alloc.GapReflection, alloc.GapProxy, alloc.GapUnsafe, alloc.GapHandle, alloc.GapService,
			alloc.GapContainerProxy, alloc.GapExternalInput},
		SourceLambda: {samcall.GapParameter, samcall.GapSAMArg, samcall.GapField, samcall.GapCallResult, samcall.GapInstance,
			samcall.GapValue, samcall.GapDepth, samcall.GapBootstrap, samcall.GapBudget},
		SourceRuntime: {runtimeevents.GapClassInit, runtimeevents.GapInvokeDynamic, runtimeevents.GapReflection, runtimeevents.GapMethodHandle,
			runtimeevents.GapProxy, runtimeevents.GapAsync, runtimeevents.GapSkippedReceiver, runtimeevents.GapRTAPruned, runtimeevents.GapInherited,
			runtimeevents.GapSAMCall, runtimeevents.GapDispatchMissing, runtimeevents.GapIndirect, runtimeevents.GapNoCallAtSite, runtimeevents.GapUnclassified},
	}
	for src, codes := range produced {
		if len(Mapping[src]) != len(codes) {
			t.Errorf("%s: mapping has %d codes, the analyzer %d", src, len(Mapping[src]), len(codes))
		}
		for _, c := range codes {
			if _, ok := Mapping[src][c]; !ok {
				t.Errorf("%s code %q has no mechanism", src, c)
			}
		}
	}
	if Classify(SourceCoordinator, "servlet-container:dispatch") != ModelIncomplete || Classify(SourceSite, "new-kind") != Unclassified {
		t.Error("coordinator / unknown code classification")
	}
	// The same mechanism under three names before b0-3 now reads the same.
	if Classify(SourceSite, "reflection") != Classify(SourceAllocation, alloc.GapReflection) || Classify(SourceSite, "reflection") != Classify(SourceRuntime, runtimeevents.GapReflection) {
		t.Error("reflection is not one mechanism")
	}
}

// TestVocabularyDocumented: the contract lists every mechanism and every
// source code with its mechanism.
func TestVocabularyDocumented(t *testing.T) {
	b, err := os.ReadFile("../../docs/RESULTS_CONTRACT.MD")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	i := strings.Index(doc, "## Gap vocabulary")
	if i < 0 {
		t.Fatal("no \"## Gap vocabulary\" section")
	}
	sec := doc[i:]
	if j := strings.Index(sec[3:], "\n## "); j >= 0 {
		sec = sec[:j+3]
	}
	for _, m := range Mechanisms {
		if !strings.Contains(sec, "| `"+string(m)+"` |") {
			t.Errorf("mechanism %s not documented", m)
		}
	}
	for src, codes := range Mapping {
		for c, m := range codes {
			if !strings.Contains(sec, src+"/`"+c+"`") || !strings.Contains(sec, "| `"+string(m)+"` |") {
				t.Errorf("%s/%s (%s) not documented", src, c, m)
			}
		}
	}
}

func TestFromCountsAndSummary(t *testing.T) {
	g := FromCounts(SourceAllocation, map[string]int{alloc.GapProxy: 3, alloc.GapReflection: 2, alloc.GapUnsafe: 0})
	if len(g) != 2 || g[0].Mechanism != Reflection || g[1].Mechanism != DynamicProxy {
		t.Errorf("gaps = %+v", g)
	}
	if s := ByMechanism(append(g, Gap{Mechanism: Reflection, Count: 5})); s != "reflection 7, dynamic_proxy 3" {
		t.Errorf("summary = %q", s)
	}
}
