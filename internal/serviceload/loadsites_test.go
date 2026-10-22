package serviceload

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// serviceLoaderPool builds a constant pool with a Class literal for
// test.MyService and a Methodref for java.util.ServiceLoader#load(Class),
// shared by the tests below. Index numbers are documented at each use
// (classfile.PoolEntry is 1-indexed, same as the real constant pool).
func serviceLoaderPool() []classfile.PoolEntry {
	return []classfile.PoolEntry{
		{Tag: classfile.TagUtf8, Str: "test/MyService"},                               // 1
		{Tag: classfile.TagClass, NameIndex: 1},                                       // 2: test.MyService
		{Tag: classfile.TagUtf8, Str: "java/util/ServiceLoader"},                      // 3
		{Tag: classfile.TagClass, NameIndex: 3},                                       // 4
		{Tag: classfile.TagUtf8, Str: "load"},                                         // 5
		{Tag: classfile.TagUtf8, Str: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"}, // 6
		{Tag: classfile.TagNameAndType, NameIndex: 5, DescriptorIndex: 6},             // 7
		{Tag: classfile.TagMethodref, ClassIndex: 4, NameAndTypeIndex: 7},             // 8: ServiceLoader#load(Class)
	}
}

func TestLoadSitesInMethod_literalClassResolves(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", serviceLoaderPool())
	code := []byte{
		0x12, 0x02, // pc0 ldc #2 (test.MyService.class)
		0xb8, 0x00, 0x08, // pc2 invokestatic #8 (ServiceLoader#load)
		0x57, // pc5 pop
		0xb1, // pc6 return
	}
	m := hierarchy.Method{
		Method: classfile.Method{Name: "m", Descriptor: "()V", AccessFlags: classfile.AccStatic, Code: code, MaxLocals: 1},
		Calls: []classfile.Call{
			{Kind: classfile.CallStatic, Owner: "java.util.ServiceLoader", Name: "load",
				Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;", PC: 2},
		},
	}

	sites, err := loadSitesInMethod(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("loadSitesInMethod: %v", err)
	}
	if len(sites) != 1 {
		t.Fatalf("sites = %+v, want exactly 1", sites)
	}
	s := sites[0]
	if s.Interface != "test.MyService" {
		t.Errorf("Interface = %q, want %q", s.Interface, "test.MyService")
	}
	if s.LoaderKind != LoaderClasspath {
		t.Errorf("LoaderKind = %q, want %q", s.LoaderKind, LoaderClasspath)
	}
	if s.PC != 2 {
		t.Errorf("PC = %d, want 2", s.PC)
	}
	if s.CallerMethod != "test.Caller#m()V" {
		t.Errorf("CallerMethod = %q, want test.Caller#m()V", s.CallerMethod)
	}
	if s.UnresolvedReason != "" {
		t.Errorf("UnresolvedReason = %q, want empty for a resolved site", s.UnresolvedReason)
	}
}

// TestLoadSitesInMethod_unknownArgumentStaysUnresolved: a parameter
// passed straight through must never be fabricated into a specific
// interface name.
func TestLoadSitesInMethod_unknownArgumentStaysUnresolved(t *testing.T) {
	cf := classfile.NewClassFileForTesting("test.Caller", serviceLoaderPool())
	code := []byte{
		0x2a,             // pc0 aload_0 (the Class parameter)
		0xb8, 0x00, 0x08, // pc1 invokestatic #8 (ServiceLoader#load)
		0x57, // pc4 pop
		0xb1, // pc5 return
	}
	m := hierarchy.Method{
		Method: classfile.Method{Name: "m", Descriptor: "(Ljava/lang/Class;)V", Code: code, MaxLocals: 1},
		Calls: []classfile.Call{
			{Kind: classfile.CallStatic, Owner: "java.util.ServiceLoader", Name: "load",
				Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;", PC: 1},
		},
	}

	sites, err := loadSitesInMethod(cf, "test.Caller", m)
	if err != nil {
		t.Fatalf("loadSitesInMethod: %v", err)
	}
	if len(sites) != 1 {
		t.Fatalf("sites = %+v, want exactly 1", sites)
	}
	if sites[0].Interface != "" {
		t.Errorf("Interface = %q, want \"\" (the argument was an unknown parameter)", sites[0].Interface)
	}
	if sites[0].UnresolvedReason == "" {
		t.Errorf("UnresolvedReason is empty, want a non-empty explanation")
	}
}

func TestAnyServiceLoaderCall(t *testing.T) {
	yes := []hierarchy.Method{{Calls: []classfile.Call{
		{Owner: "java.util.ServiceLoader", Name: "load", Desc: "(Ljava/lang/Class;)Ljava/util/ServiceLoader;"},
	}}}
	if !anyServiceLoaderCall(yes) {
		t.Errorf("anyServiceLoaderCall(load(Class)) = false, want true")
	}

	no := []hierarchy.Method{{Calls: []classfile.Call{
		{Owner: "java.util.ServiceLoader", Name: "stream", Desc: "()Ljava/util/stream/Stream;"},
		{Owner: "java.lang.String", Name: "valueOf", Desc: "(I)Ljava/lang/String;"},
	}}}
	if anyServiceLoaderCall(no) {
		t.Errorf("anyServiceLoaderCall(unrelated calls) = true, want false")
	}
}
