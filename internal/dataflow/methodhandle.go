package dataflow

import (
	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/valueflow"
)

// lookupOwner is java.lang.invoke.MethodHandles$Lookup — recognized
// directly here (not in a separate consumer package like
// internal/callback's own catalog) because this is core JVM invocation
// machinery, the same tier invokedynamic/LambdaMetafactory already get
// in stepCall, not an arbitrary third-party API.
const lookupOwner = "java.lang.invoke.MethodHandles$Lookup"

// lookupFindSpec describes one MethodHandles.Lookup find* overload:
// which argument (if any) is the method name, and which JVMS 5.4.3.5
// reference kind it implies.
type lookupFindSpec struct {
	nameArgIndex int // -1 for findConstructor, which takes no name
	refKind      byte
}

var lookupFindCatalog = map[[2]string]lookupFindSpec{
	{"findStatic", "(Ljava/lang/Class;Ljava/lang/String;Ljava/lang/invoke/MethodType;)Ljava/lang/invoke/MethodHandle;"}:                   {1, classfile.RefInvokeStatic},
	{"findVirtual", "(Ljava/lang/Class;Ljava/lang/String;Ljava/lang/invoke/MethodType;)Ljava/lang/invoke/MethodHandle;"}:                  {1, classfile.RefInvokeVirtual},
	{"findSpecial", "(Ljava/lang/Class;Ljava/lang/String;Ljava/lang/invoke/MethodType;Ljava/lang/Class;)Ljava/lang/invoke/MethodHandle;"}: {1, classfile.RefInvokeSpecial},
	{"findConstructor", "(Ljava/lang/Class;Ljava/lang/invoke/MethodType;)Ljava/lang/invoke/MethodHandle;"}:                                {-1, classfile.RefNewInvokeSpecial},
}

// lookupHandleValue recognizes a MethodHandles.Lookup find* call and
// builds a HandleValue from its already-resolved Class/name arguments —
// DYNAMIC_LINKAGE L2's own "stały cel" case. ok is false for any other
// call (including a find* call whose Class/name argument isn't a
// provable literal — L2's own "nieznany lookup", left for the caller to
// widen to Unknown with a reason, never guessed at).
func lookupHandleValue(site bytecodeir.Site, op classfile.LocalOp, args []valueflow.Value) (valueflow.HandleValue, bool) {
	if op.CallOwner != lookupOwner {
		return valueflow.HandleValue{}, false
	}
	spec, ok := lookupFindCatalog[[2]string{op.CallName, op.CallDesc}]
	if !ok {
		return valueflow.HandleValue{}, false
	}
	if len(args) == 0 {
		return valueflow.HandleValue{}, false
	}
	ownerVal := args[0]
	if ownerVal.Kind != valueflow.KindConst || ownerVal.Type != "java.lang.Class" {
		return valueflow.HandleValue{}, false
	}
	owner, ok := ownerVal.Const.(string)
	if !ok {
		return valueflow.HandleValue{}, false
	}

	name := "<init>"
	if spec.nameArgIndex >= 0 {
		if len(args) <= spec.nameArgIndex {
			return valueflow.HandleValue{}, false
		}
		nameVal := args[spec.nameArgIndex]
		if nameVal.Kind != valueflow.KindConst || nameVal.Type != "java.lang.String" {
			return valueflow.HandleValue{}, false
		}
		n, ok := nameVal.Const.(string)
		if !ok {
			return valueflow.HandleValue{}, false
		}
		name = n
	}

	return valueflow.HandleValue{Site: site, RefKind: spec.refKind, RequestedOwner: owner, Name: name}, true
}
