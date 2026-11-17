package vulnmap

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

// A jackson-databind blocklist fix (a class name added to
// SubTypeValidator's static set, nothing else) maps to the API that
// enables default typing, role entrypoint, with the initializer kept as
// evidence — not to SubTypeValidator#<clinit>, which runs on any
// deserialization.
func TestFixDiff_jacksonBlocklistRedirect(t *testing.T) {
	const path = "src/main/java/com/fasterxml/jackson/databind/jsontype/impl/SubTypeValidator.java"
	const old = `package com.fasterxml.jackson.databind.jsontype.impl;

public class SubTypeValidator {
    protected final static java.util.Set<String> DEFAULT_NO_DESER_CLASS_NAMES;
    static {
        java.util.Set<String> s = new java.util.HashSet<String>();
        s.add("org.apache.commons.collections.functors.InvokerTransformer");
        DEFAULT_NO_DESER_CLASS_NAMES = s;
    }

    public void validateSubType(Object ctxt, Object type) {
        check(type);
    }
}
`
	fixed := strings.Replace(old, `        DEFAULT_NO_DESER_CLASS_NAMES = s;`,
		`        s.add("org.apache.xalan.lib.sql.JNDIConnectionPool");
        DEFAULT_NO_DESER_CLASS_NAMES = s;`, 1)
	diff := `diff --git a/` + path + ` b/` + path + `
--- a/` + path + `
+++ b/` + path + `
@@ -6,6 +6,7 @@ public class SubTypeValidator {
         java.util.Set<String> s = new java.util.HashSet<String>();
         s.add("org.apache.commons.collections.functors.InvokerTransformer");
+        s.add("org.apache.xalan.lib.sql.JNDIConnectionPool");
         DEFAULT_NO_DESER_CLASS_NAMES = s;
     }

`
	s := newF5Server(t)
	s.api("o/r", f5Fix, f5Parent)
	s.serve("/o/r/commit/"+f5Fix+".diff", diff)
	s.serve("/o/r/"+f5Parent+"/"+path, old)
	s.serve("/o/r/"+f5Fix+"/"+path, fixed)
	e, ok, err := s.resolver(t, nil).Resolve(context.Background(), f5CVE("CVE-JACKSON-GADGET", f5Fix))
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if !reflect.DeepEqual(e.Methods, initializerRedirects[0].entrypoints) {
		t.Errorf("methods = %v, want the default-typing entry points", e.Methods)
	}
	if e.Role != RoleEntrypoint || e.Source != SourceDiff || len(e.FixCommits) != 1 {
		t.Errorf("role=%q source=%q commits=%d, want entrypoint/diff/1", e.Role, e.Source, len(e.FixCommits))
	}
	if e.InitializerRedirect == nil || !reflect.DeepEqual(e.InitializerRedirect.Initializers,
		[]string{"com.fasterxml.jackson.databind.jsontype.impl.SubTypeValidator#<clinit>"}) {
		t.Errorf("redirect = %+v", e.InitializerRedirect)
	}
	if e.InitializerOnly() {
		t.Error("a redirected mapping names no initializer: InitializerOnly must be false")
	}
}

func TestRedirectInitializers(t *testing.T) {
	const stv = "com.fasterxml.jackson.databind.jsontype.impl.SubTypeValidator#<clinit>()V"
	const bdf = "com.fasterxml.jackson.databind.deser.BeanDeserializerFactory#<clinit>"
	for name, tc := range map[string]struct {
		in   []string
		want bool
	}{
		"SubTypeValidator":       {[]string{stv}, true},
		"both blocklists":        {[]string{bdf, stv}, true},
		"next to a method":       {[]string{stv, "com.fasterxml.jackson.databind.jsontype.impl.SubTypeValidator#validateSubType"}, false},
		"other class <clinit>":   {[]string{"com.fasterxml.jackson.databind.ext.DOMDeserializer#<clinit>()V"}, false},
		"blocklist class <init>": {[]string{"com.fasterxml.jackson.databind.deser.BeanDeserializerFactory#<init>"}, false},
		"empty":                  {nil, false},
	} {
		got, r, ok := redirectInitializers(tc.in)
		if ok != tc.want {
			t.Errorf("%s: ok=%v, want %v", name, ok, tc.want)
			continue
		}
		if !ok && !reflect.DeepEqual(got, tc.in) {
			t.Errorf("%s: methods changed without a redirect: %v", name, got)
		}
		if ok && (r == nil || !reflect.DeepEqual(r.Initializers, tc.in) || len(got) != len(initializerRedirects[0].entrypoints)) {
			t.Errorf("%s: methods=%v redirect=%+v", name, got, r)
		}
	}
}

func TestInitializerOnly(t *testing.T) {
	for name, tc := range map[string]struct {
		e    Entry
		want bool
	}{
		"diff <clinit>":         {Entry{Source: SourceDiff, Methods: []string{"p.A#<clinit>()V", "p.B#<clinit>"}}, true},
		"release-diff <clinit>": {Entry{Source: SourceReleaseDiff, Methods: []string{"p.A#<clinit>"}}, true},
		"next to a method":      {Entry{Source: SourceDiff, Methods: []string{"p.A#<clinit>", "p.A#run"}}, false},
		"constructors":          {Entry{Source: SourceDiff, Methods: []string{"p.A#<init>"}}, false},
		"manual <clinit>":       {Entry{Source: SourceManual, Methods: []string{"p.A#<clinit>"}}, false},
		"no methods":            {Entry{Source: SourceDiff}, false},
	} {
		if got := tc.e.InitializerOnly(); got != tc.want {
			t.Errorf("%s: InitializerOnly = %v, want %v", name, got, tc.want)
		}
	}
}
