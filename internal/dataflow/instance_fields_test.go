package dataflow

import (
	"testing"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// instanceFieldPool mirrors fieldWriterPool but with a Fieldref sized for
// an instance field named NAME, plus a Sink#accept(String) methodref to
// read a getfield's loaded value back out (same pattern
// TestAnalyze_staticFieldReadProducesKindField uses for a static field).
func instanceFieldPool() []classfile.PoolEntry {
	return []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/Caller"},                        // 1
		{Tag: classfile.TagClass, NameIndex: 1},                             // 2
		{Tag: classfile.TagUtf8, Str: "NAME"},                               // 3
		{Tag: classfile.TagUtf8, Str: "Ljava/lang/String;"},                 // 4
		{Tag: classfile.TagNameAndType, NameIndex: 3, DescriptorIndex: 4},   // 5
		{Tag: classfile.TagFieldref, ClassIndex: 2, NameAndTypeIndex: 5},    // 6: test.Caller#NAME
		{Tag: classfile.TagUtf8, Str: "x"},                                  // 7
		{Tag: classfile.TagString, NameIndex: 7},                            // 8
		{Tag: classfile.TagUtf8, Str: "test/Sink"},                          // 9
		{Tag: classfile.TagClass, NameIndex: 9},                             // 10
		{Tag: classfile.TagUtf8, Str: "accept"},                             // 11
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/String;)V"},              // 12
		{Tag: classfile.TagNameAndType, NameIndex: 11, DescriptorIndex: 12}, // 13
		{Tag: classfile.TagMethodref, ClassIndex: 10, NameAndTypeIndex: 13}, // 14: test.Sink#accept(String)V
	}
}

// TestAnalyze_instanceFieldReadProducesKindField is DATA_FLOW D3's
// instance-field increment, first example: a getfield must identify
// *which* field it read, same as getstatic already does.
//
// Bytecode (void m()):
//
//	pc0: aload_0
//	pc1: getfield #6     (test.Caller#NAME:Ljava/lang/String;)
//	pc4: invokestatic #14 (test.Sink#accept(String)V)
//	pc7: return
func TestAnalyze_instanceFieldReadProducesKindField(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", instanceFieldPool())
	code := []byte{
		0x2a,             // pc0 aload_0 (this)
		0xb4, 0x00, 0x06, // pc1 getfield #6
		0xb8, 0x00, 0x0e, // pc4 invokestatic #14
		0xb1, // pc7 return
	}
	m := classfile.Method{Name: "m", Descriptor: "()V", Code: code, MaxLocals: 1}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	val := facts.ArgumentAt(bytecodeir.Site{Method: "test.Caller#m", Descriptor: "()V", PC: 4, Op: bytecodeir.OpCall}, 0)
	if val.Kind != valueflow.KindField {
		t.Fatalf("ArgumentAt = %+v, want KindField", val)
	}
	want := valueflow.FieldRef{Owner: "test.Caller", Name: "NAME", Desc: "Ljava/lang/String;"}
	if *val.Field != want {
		t.Errorf("Field = %+v, want %+v", *val.Field, want)
	}
}

// TestAnalyze_instanceFieldSelfWriteInConstructorRecordsFieldValueAt is
// D3's instance-field increment, second example: a `this.field = ...`
// putfield inside a constructor must have its stored value recoverable,
// same as a putstatic already does.
//
// Bytecode (<init>()V):
//
//	pc0: aload_0  (this)
//	pc1: aload_0  (this, objectref for putfield)
//	pc2: ldc #8   ("x")
//	pc4: putfield #6 (test.Caller#NAME)
//	pc7: return
func TestAnalyze_instanceFieldSelfWriteInConstructorRecordsFieldValueAt(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", instanceFieldPool())
	code := []byte{
		0x2a,       // pc0 aload_0 (this, objectref)
		0x12, 0x08, // pc1 ldc #8 ("x")
		0xb5, 0x00, 0x06, // pc3 putfield #6
		0xb1, // pc6 return
	}
	m := classfile.Method{Name: "<init>", Descriptor: "()V", Code: code, MaxLocals: 1}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	site := bytecodeir.Site{Method: "test.Caller#<init>", Descriptor: "()V", PC: 3, Op: bytecodeir.OpFieldPut}
	v, ok := facts.fieldWrites[site]
	if !ok {
		t.Fatalf("fieldWrites[%+v] not recorded, want a self-write", site)
	}
	want := valueflow.Const("java.lang.String", "x")
	if !v.Equal(want) {
		t.Errorf("fieldWrites = %+v, want %+v", v, want)
	}
}

// TestAnalyze_instanceFieldWriteOutsideConstructorNotRecorded confirms
// the narrow scope: a putfield in a method other than <init> (even
// through `this`) is not recorded, since D3's instance-field increment
// trusts finality + constructor-only to make the value instance-
// independent, and a non-constructor write would break that.
func TestAnalyze_instanceFieldWriteOutsideConstructorNotRecorded(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", instanceFieldPool())
	code := []byte{
		0x2a,       // pc0 aload_0 (this, objectref)
		0x12, 0x08, // pc1 ldc #8 ("x")
		0xb5, 0x00, 0x06, // pc3 putfield #6
		0xb1, // pc6 return
	}
	m := classfile.Method{Name: "setName", Descriptor: "()V", Code: code, MaxLocals: 1}

	facts, err := Analyze(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if aborted, reason := facts.Aborted(); aborted {
		t.Fatalf("Analyze aborted: %s", reason)
	}

	site := bytecodeir.Site{Method: "test.Caller#setName", Descriptor: "()V", PC: 3, Op: bytecodeir.OpFieldPut}
	if _, ok := facts.fieldWrites[site]; ok {
		t.Errorf("fieldWrites[%+v] recorded, want absent (not a constructor)", site)
	}
}

func instanceWriterMethod(name string) hierarchy.Method {
	code := []byte{
		0x2a,       // pc0 aload_0 (this, objectref)
		0x12, 0x08, // pc1 ldc #8 ("x")
		0xb5, 0x00, 0x06, // pc3 putfield #6
		0xb1, // pc6 return
	}
	return hierarchy.Method{
		Method: classfile.Method{Name: name, Descriptor: "()V", Code: code, MaxLocals: 1},
	}
}

func TestMergeConstructorFieldWriters_singleWriterResolves(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", instanceFieldPool())
	methods := []hierarchy.Method{instanceWriterMethod("<init>")}

	v, warnings := mergeConstructorFieldWriters(cf, "test.Caller", methods, "NAME", "Ljava/lang/String;")
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	want := valueflow.Const("java.lang.String", "x")
	if !v.Equal(want) {
		t.Errorf("mergeConstructorFieldWriters = %+v, want %+v", v, want)
	}
}

func TestMergeConstructorFieldWriters_noWriterIsHonestUnknown(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", instanceFieldPool())
	methods := []hierarchy.Method{{Method: classfile.Method{Name: "<init>", Descriptor: "()V", Code: []byte{0xb1}}}}

	v, _ := mergeConstructorFieldWriters(cf, "test.Caller", methods, "NAME", "Ljava/lang/String;")
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("mergeConstructorFieldWriters(no writer) = %+v, want KindUnknown", v)
	}
}

func TestMergeConstructorFieldWriters_nonConstructorWriterIgnored(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", instanceFieldPool())
	methods := []hierarchy.Method{instanceWriterMethod("setName")}

	v, _ := mergeConstructorFieldWriters(cf, "test.Caller", methods, "NAME", "Ljava/lang/String;")
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("mergeConstructorFieldWriters(non-constructor writer) = %+v, want KindUnknown", v)
	}
}

func TestResolveInstanceField_unresolvableOwnerIsUnknown(t *testing.T) {
	h := hierarchy.NewForTesting(nil)
	v, _ := ResolveInstanceField(h, valueflow.FieldRef{Owner: "test.NoSuchClass", Name: "X", Desc: "I"})
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("ResolveInstanceField(unresolvable owner) = %+v, want KindUnknown", v)
	}
}

func TestResolveInstanceField_nonFinalFieldIsUnknown(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN:   "test.Caller",
			Fields: []classfile.Field{{Name: "NAME", Descriptor: "Ljava/lang/String;", AccessFlags: 0}},
		},
	})
	v, _ := ResolveInstanceField(h, valueflow.FieldRef{Owner: "test.Caller", Name: "NAME", Desc: "Ljava/lang/String;"})
	if v.Kind != valueflow.KindUnknown {
		t.Errorf("ResolveInstanceField(non-final field) = %+v, want KindUnknown", v)
	}
}

// TestResolveField_dispatchesByAccessFlagsNotTrialAndError confirms
// ResolveField picks ResolveInstanceField/ResolveStaticField from the
// field's own access flags, rather than trying static first and falling
// through to instance on any Unknown. The two cases below have no
// SourceJar to read (not needed for this check), so each one's failure
// reason, left distinct, proves which branch actually ran:
// ResolveInstanceField's non-final short-circuit never touches a jar at
// all, while ResolveStaticField's does (and fails to open one) — a
// trial-and-error caller would mask the static field's real "could not
// re-read" reason with the instance branch's "not a final field"
// message, which is wrong for a field that is static in the first
// place.
func TestResolveField_dispatchesByAccessFlagsNotTrialAndError(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{
			FQCN: "test.Caller",
			Fields: []classfile.Field{
				{Name: "STATIC_NAME", Descriptor: "Ljava/lang/String;", AccessFlags: classfile.AccStatic | classfile.AccFinal},
				{Name: "INSTANCE_NAME", Descriptor: "Ljava/lang/String;", AccessFlags: 0},
			},
		},
	})

	staticVal, staticWarnings := ResolveField(h, valueflow.FieldRef{Owner: "test.Caller", Name: "STATIC_NAME", Desc: "Ljava/lang/String;"})
	if staticVal.Kind != valueflow.KindUnknown {
		t.Fatalf("ResolveField(static) = %+v, want KindUnknown", staticVal)
	}
	if len(staticWarnings) == 0 {
		t.Errorf("ResolveField(static) warnings = %v, want a jar-read warning (proves ResolveStaticField ran)", staticWarnings)
	}

	instanceVal, instanceWarnings := ResolveField(h, valueflow.FieldRef{Owner: "test.Caller", Name: "INSTANCE_NAME", Desc: "Ljava/lang/String;"})
	if instanceVal.Kind != valueflow.KindUnknown {
		t.Fatalf("ResolveField(instance) = %+v, want KindUnknown", instanceVal)
	}
	if len(instanceWarnings) != 0 {
		t.Errorf("ResolveField(instance) warnings = %v, want none (non-final short-circuits before any jar read)", instanceWarnings)
	}
	if staticVal.Reason == instanceVal.Reason {
		t.Errorf("static and instance fields produced the same Unknown reason %q — dispatch did not actually differ", staticVal.Reason)
	}
}
