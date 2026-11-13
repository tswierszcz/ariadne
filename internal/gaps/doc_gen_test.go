package gaps

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

// TestPrintDocTable writes the contract table when GAPS_DOC is set (a
// maintenance helper; the table in docs/RESULTS_CONTRACT.MD is checked by
// TestVocabularyDocumented).
func TestPrintDocTable(t *testing.T) {
	out := os.Getenv("GAPS_DOC")
	if out == "" {
		t.Skip("GAPS_DOC not set")
	}
	meaning := map[Mechanism]string{
		Reflection: "Method#invoke, Constructor#newInstance, Class.forName with an unresolved value", MethodHandle: "MethodHandle#invoke*, Lookup#find* with an unresolved value",
		CallbackAsync: "a task or callback handed to a thread, executor or framework", ServiceLoader: "ServiceLoader use whose service or providers are unknown",
		Deserialization: "a deserialization entry whose input classes are open", DynamicProxy: "a JDK Proxy / generated InvocationHandler dispatch",
		ContainerProxy: "a container-generated implementation (repository, AOP proxy)", UnknownAllocation: "an object created where the type is unknown (Unsafe, constructor handles)",
		ExternalInput: "`-roots all`: a root's parameters may be any subtype", Lambda: "a functional-interface call whose lambda / method reference is unresolved",
		SkippedReceiver: "a call on `java.`/`javax.`/`jdk.`/`sun.` types, never CHA-expanded", VirtualDispatch: "a virtual call target the graph has no edge for (RTA pruning, inherited bodies)",
		ClassInit: "a class initializer the JVM runs implicitly", ModelIncomplete: "a library model that could not resolve a site (B3 coordinator; detail: model id)",
		TraceAlignment: "a measured call not attributable to one site (dropped frames, inlining)", Unclassified: "no mechanism known",
	}
	var b strings.Builder
	b.WriteString("| Mechanism | Meaning | Source codes |\n|---|---|---|\n")
	for _, m := range Mechanisms {
		var codes []string
		for _, src := range Sources {
			var cs []string
			for c, mm := range Mapping[src] {
				if mm == m {
					cs = append(cs, src+"/`"+c+"`")
				}
			}
			sort.Strings(cs)
			codes = append(codes, cs...)
		}
		if m == ModelIncomplete {
			codes = append(codes, "coordinator/*model id*")
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", m, meaning[m], strings.Join(codes, ", "))
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
