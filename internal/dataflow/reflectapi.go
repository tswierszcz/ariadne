package dataflow

import (
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

// classOwner is java.lang.Class — recognized directly here, same
// reasoning as lookupOwner in methodhandle.go: core JVM reflection
// machinery, not an arbitrary third-party API.
const classOwner = "java.lang.Class"

// forNameOverloads maps every `Class.forName` overload ariadne
// recognizes to the 0-based argument index holding the class name —
// always 0, but kept as a set (not a bare bool) so a future overload
// (e.g. the Java 9+ `forName(Module, String)`, whose name argument is
// at index 1) has an obvious place to go without restructuring the
// recognizer itself.
var forNameOverloads = map[[2]string]int{
	{"forName", "(Ljava/lang/String;)Ljava/lang/Class;"}:                         0,
	{"forName", "(Ljava/lang/String;ZLjava/lang/ClassLoader;)Ljava/lang/Class;"}: 0,
}

// classForNameValue recognizes a `Class.forName` call whose class-name
// argument is a provable string literal — REFLECTION R1's own
// "jednoznaczna nazwa" case — and resolves it directly to the same
// KindConst/"java.lang.Class" shape a `Foo.class` literal already
// produces (constValue's own ConstKindClass case), rather than the
// generic KindCallResult every other invokestatic call falls through
// to. This is deliberate reuse, not a new Value kind: `Class.forName`
// and `Foo.class` both resolve to "the Class object named by this
// FQCN" — the same fact, just reached two different ways — and any
// existing consumer of a Class-typed KindConst (e.g.
// SERVICE_LOADING S1's own `ServiceLoader.load(Class)` argument
// reading) benefits immediately, with no changes of its own.
//
// `Class.forName`'s own string argument is already a dotted binary
// name (JLS 13.1) — "com.example.Outer$Inner" for a nested class,
// identical to ariadne's own FQCN convention — so it is used as-is,
// unlike a `Foo.class` literal's internal name (slash-separated),
// which constValue's own ConstKindClass case gets pre-converted by
// the classfile reader before this analysis ever sees it.
//
// ok is false for any other call, including a `forName` call whose
// own name argument isn't a provable literal (a runtime-computed
// class name — REFLECTION R1's own explicit non-goal, "poza pierwszym
// zakresem: dowolne runtime-computed nazwy") — left for the caller to
// fall through to the generic KindCallResult, never guessed at.
func classForNameValue(op classfile.LocalOp, args []valueflow.Value) (valueflow.Value, bool) {
	if op.CallOwner != classOwner {
		return valueflow.Value{}, false
	}
	argIndex, ok := forNameOverloads[[2]string{op.CallName, op.CallDesc}]
	if !ok {
		return valueflow.Value{}, false
	}
	if len(args) <= argIndex {
		return valueflow.Value{}, false
	}
	nameVal := args[argIndex]
	if nameVal.Kind != valueflow.KindConst || nameVal.Type != "java.lang.String" {
		return valueflow.Value{}, false
	}
	fqcn, ok := nameVal.Const.(string)
	if !ok {
		return valueflow.Value{}, false
	}
	return valueflow.Const("java.lang.Class", fqcn), true
}
