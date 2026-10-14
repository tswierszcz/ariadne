package classfile

import (
	"archive/zip"
	"testing"
)

// readFixtureClass opens testdata/fixture/fixture-<release>.jar (committed
// by scripts/build-fixtures.sh) and parses the given class entry, e.g.
// readFixtureClass(t, 8, "fixture/App.class").
func readFixtureClass(t *testing.T, release int, entry string) *ClassFile {
	t.Helper()
	path := fixtureJarPath(release)
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("opening %s: %v (run `make fixtures` if testdata/fixture jars are missing)", path, err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name != entry {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening %s in %s: %v", entry, path, err)
		}
		defer rc.Close()
		cf, err := Parse(rc)
		if err != nil {
			t.Fatalf("parsing %s from %s: %v", entry, path, err)
		}
		return cf
	}
	t.Fatalf("%s not found in %s", entry, path)
	return nil
}

func fixtureJarPath(release int) string {
	switch release {
	case 8:
		return "../../testdata/fixture/fixture-8.jar"
	case 21:
		return "../../testdata/fixture/fixture-21.jar"
	default:
		panic("unknown fixture release")
	}
}

func methodByName(cf *ClassFile, name string) *Method {
	for i := range cf.Methods {
		if cf.Methods[i].Name == name {
			return &cf.Methods[i]
		}
	}
	return nil
}

func callCounts(calls []Call) map[Call]int {
	m := make(map[Call]int)
	for _, c := range calls {
		m[c]++
	}
	return m
}

// TestParseAndExtractCalls_handleMethod cross-checks ariadne's own
// bytecode decoder against ground truth read by hand from
// `javap -c -p -v fixture/App.class` for both the Java 8 and Java 21
// compiles of testdata/fixture (see scripts/build-fixtures.sh). The two
// compiles are genuinely different bytecode, not just the same
// instructions with a different version number, which is exactly why
// both are worth testing:
//   - targeting --release 8, javac emits invokespecial for a call to the
//     class's own private method (greetWith) and compiles string
//     concatenation (`+`) the old way, as explicit
//     java.lang.StringBuilder calls — Java 8 has no StringConcatFactory.
//   - targeting --release 21 (compiled by a JDK 17 javac), the same
//     source compiles the greetWith call via invokevirtual instead
//     (nest-based private access, JEP 181), and string concatenation
//     becomes a single invokedynamic through
//     StringConcatFactory#makeConcatWithConstants — a bootstrap ariadne
//     does not resolve (it isn't LambdaMetafactory), so it must show up
//     as one unresolved dynamic call site rather than an error or a
//     silently wrong/missing edge.
//
// Both compiles agree on everything else: the constructor calls
// (invokespecial), the interface calls through Greeter (invokeinterface),
// the lambda resolved through LambdaMetafactory to its synthetic method,
// and the method reference (Util::shout) resolved the same way to a real
// static method.
func TestParseAndExtractCalls_handleMethod(t *testing.T) {
	const greetWithDesc = "(Ljava/lang/String;Ljava/util/function/Function;)Ljava/lang/String;"
	common := map[Call]int{
		{CallSpecial, "fixture.FormalGreeter", "<init>", "()V"}:                                   1,
		{CallSpecial, "fixture.CasualGreeter", "<init>", "()V"}:                                   1,
		{CallInterface, "fixture.Greeter", "greet", "(Ljava/lang/String;)Ljava/lang/String;"}:     2,
		{CallDynamic, "fixture.App", "lambda$handle$0", "(Ljava/lang/String;)Ljava/lang/String;"}: 1,
		{CallDynamic, "fixture.Util", "shout", "(Ljava/lang/String;)Ljava/lang/String;"}:          1,
	}

	cases := []struct {
		release        int
		extra          map[Call]int
		wantUnresolved int
	}{
		{
			release: 8,
			extra: map[Call]int{
				{CallSpecial, "fixture.App", "greetWith", greetWithDesc}:                                            2,
				{CallSpecial, "java.lang.StringBuilder", "<init>", "()V"}:                                           1,
				{CallVirtual, "java.lang.StringBuilder", "append", "(Ljava/lang/String;)Ljava/lang/StringBuilder;"}: 4,
				{CallVirtual, "java.lang.StringBuilder", "toString", "()Ljava/lang/String;"}:                        1,
			},
			wantUnresolved: 0,
		},
		{
			release: 21,
			extra: map[Call]int{
				{CallVirtual, "fixture.App", "greetWith", greetWithDesc}: 2,
			},
			// StringConcatFactory#makeConcatWithConstants: not
			// LambdaMetafactory, so it must be reported as unresolved
			// rather than produce a wrong or missing edge.
			wantUnresolved: 1,
		},
	}

	for _, tc := range cases {
		t.Run(releaseName(tc.release), func(t *testing.T) {
			cf := readFixtureClass(t, tc.release, "fixture/App.class")

			if cf.ThisClass != "fixture.App" {
				t.Errorf("ThisClass = %q, want fixture.App", cf.ThisClass)
			}
			if cf.SuperClass != "java.lang.Object" {
				t.Errorf("SuperClass = %q, want java.lang.Object", cf.SuperClass)
			}

			handle := methodByName(cf, "handle")
			if handle == nil {
				t.Fatalf("no handle method found")
			}
			if handle.Descriptor != "(Ljava/lang/String;)Ljava/lang/String;" {
				t.Errorf("handle descriptor = %q", handle.Descriptor)
			}
			if len(handle.Annotations) != 1 || handle.Annotations[0] != "fixture.Entry" {
				t.Errorf("handle annotations = %v, want [fixture.Entry]", handle.Annotations)
			}

			calls, unresolved, err := cf.ExtractCalls(handle.Code)
			if err != nil {
				t.Fatalf("ExtractCalls: %v", err)
			}
			if unresolved != tc.wantUnresolved {
				t.Errorf("unresolved dynamic call sites = %d, want %d", unresolved, tc.wantUnresolved)
			}

			want := map[Call]int{}
			for k, v := range common {
				want[k] = v
			}
			for k, v := range tc.extra {
				want[k] = v
			}

			got := callCounts(calls)
			for call, wantCount := range want {
				if got[call] != wantCount {
					t.Errorf("call %+v: got count %d, want %d", call, got[call], wantCount)
				}
			}
			for call, gotCount := range got {
				if _, known := want[call]; !known {
					t.Errorf("unexpected call %+v (count %d)", call, gotCount)
				}
			}
		})
	}
}

func TestParse_parameterAnnotations(t *testing.T) {
	for _, release := range []int{8, 21} {
		t.Run(releaseName(release), func(t *testing.T) {
			cf := readFixtureClass(t, release, "fixture/Observer.class")

			onEvent := methodByName(cf, "onEvent")
			if onEvent == nil {
				t.Fatalf("no onEvent method found")
			}
			if len(onEvent.ParameterAnnotations) != 2 {
				t.Fatalf("ParameterAnnotations = %v, want 2 entries (one per parameter)", onEvent.ParameterAnnotations)
			}
			if got := onEvent.ParameterAnnotations[0]; len(got) != 1 || got[0] != "fixture.Param" {
				t.Errorf("parameter 0 annotations = %v, want [fixture.Param]", got)
			}
			if got := onEvent.ParameterAnnotations[1]; len(got) != 0 {
				t.Errorf("parameter 1 annotations = %v, want none", got)
			}
		})
	}
}

func TestParseAndExtractCalls_greetWithIsInvokeinterfaceOnFunction(t *testing.T) {
	cf := readFixtureClass(t, 8, "fixture/App.class")
	m := methodByName(cf, "greetWith")
	if m == nil {
		t.Fatalf("no greetWith method found")
	}
	calls, _, err := cf.ExtractCalls(m.Code)
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}
	want := Call{CallInterface, "java.util.function.Function", "apply", "(Ljava/lang/Object;)Ljava/lang/Object;"}
	if callCounts(calls)[want] != 1 {
		t.Errorf("calls = %+v, want exactly one %+v", calls, want)
	}
}

func TestParseInterfaceAndImplementations(t *testing.T) {
	greeter := readFixtureClass(t, 8, "fixture/Greeter.class")
	if !greeter.IsInterface() {
		t.Errorf("Greeter.IsInterface() = false, want true")
	}
	greet := methodByName(greeter, "greet")
	if greet == nil {
		t.Fatalf("Greeter has no greet method")
	}
	if greet.AccessFlags&AccAbstract == 0 {
		t.Errorf("Greeter#greet AccessFlags = %#x, want ACC_ABSTRACT set", greet.AccessFlags)
	}

	for _, impl := range []string{"FormalGreeter", "CasualGreeter"} {
		cf := readFixtureClass(t, 8, "fixture/"+impl+".class")
		if cf.IsInterface() {
			t.Errorf("%s.IsInterface() = true, want false", impl)
		}
		if cf.SuperClass != "java.lang.Object" {
			t.Errorf("%s.SuperClass = %q, want java.lang.Object", impl, cf.SuperClass)
		}
		if len(cf.Interfaces) != 1 || cf.Interfaces[0] != "fixture.Greeter" {
			t.Errorf("%s.Interfaces = %v, want [fixture.Greeter]", impl, cf.Interfaces)
		}
		m := methodByName(cf, "greet")
		if m == nil {
			t.Fatalf("%s has no greet method", impl)
		}
		if len(m.Code) == 0 {
			t.Errorf("%s#greet has no Code (expected a concrete implementation)", impl)
		}
	}
}

func TestMainMethodCallsHandle(t *testing.T) {
	cf := readFixtureClass(t, 8, "fixture/App.class")
	m := methodByName(cf, "main")
	if m == nil {
		t.Fatalf("no main method found")
	}
	if !m.IsStatic() {
		t.Errorf("main is not static")
	}
	calls, _, err := cf.ExtractCalls(m.Code)
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}
	want := Call{CallVirtual, "fixture.App", "handle", "(Ljava/lang/String;)Ljava/lang/String;"}
	if callCounts(calls)[want] != 1 {
		t.Errorf("calls = %+v, want exactly one %+v", calls, want)
	}
	ctor := Call{CallSpecial, "fixture.App", "<init>", "()V"}
	if callCounts(calls)[ctor] != 1 {
		t.Errorf("calls = %+v, want exactly one %+v", calls, ctor)
	}
}

func releaseName(r int) string {
	switch r {
	case 8:
		return "release8"
	case 21:
		return "release21"
	default:
		return "unknown"
	}
}
