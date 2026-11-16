package vulnmap

import (
	"reflect"
	"testing"

	"ariadne/internal/mavenversion"
)

// TestServerManualMappings pins the version split of the 2026-11-16
// server/Jackson/logback entries: where the fix-commit diff named a class
// that only a newer line has, the older line (the one the control apps
// ship: tomcat-embed-core 9.0.45, Jackson 2.11.4, logback 1.2.3, Undertow
// 2.2.4, XNIO 3.8.0) selects the method that holds the same code there,
// and the fixing release selects nothing.
func TestServerManualMappings(t *testing.T) {
	meta, err := LoadManualMeta("")
	if err != nil {
		t.Fatal(err)
	}
	pick := func(id, pkg, v string) []string {
		var out []string
		n := 0
		for _, vr := range meta[id].Variants {
			if vr.Package == pkg && vr.Range.Contains(v) == mavenversion.In {
				n++
				out = append(out, vr.Methods...)
			}
		}
		if n > 1 {
			t.Errorf("%s %s %s: %d variants cover it", id, pkg, v, n)
		}
		return out
	}
	const (
		tomcat = "org.apache.tomcat.embed:tomcat-embed-core"
		core   = "com.fasterxml.jackson.core:jackson-core"
		db     = "com.fasterxml.jackson.core:jackson-databind"
		lb     = "ch.qos.logback:logback-core"
		ut     = "io.undertow:undertow-core"
	)
	for _, c := range []struct {
		id, pkg, v string
		want       []string
	}{
		{"CVE-2024-56337", tomcat, "9.0.45", []string{"org.apache.catalina.servlets.DefaultServlet#doPut"}},
		{"CVE-2024-56337", tomcat, "9.0.98", nil},
		{"CVE-2026-65905", tomcat, "9.0.45", []string{"org.apache.catalina.authenticator.DigestAuthenticator$NonceInfo#nonceCountValid"}},
		{"CVE-2026-65905", tomcat, "10.1.58", nil},
		{"CVE-2026-68494", core, "2.11.4", []string{
			"com.fasterxml.jackson.core.json.async.NonBlockingJsonParser#_startPositiveNumber",
			"com.fasterxml.jackson.core.json.async.NonBlockingJsonParser#_startNegativeNumber",
			"com.fasterxml.jackson.core.json.async.NonBlockingJsonParser#_finishNumberIntegralPart"}},
		{"CVE-2026-68494", core, "2.18.8", nil},
		{"CVE-2026-50193", db, "2.11.4", []string{
			"com.fasterxml.jackson.databind.node.InternalNodeMapper#nodeToString",
			"com.fasterxml.jackson.databind.node.InternalNodeMapper#nodeToPrettyString"}},
		{"CVE-2026-50193", db, "2.14.0", nil},
		{"CVE-2026-54514", db, "2.11.4", []string{"com.fasterxml.jackson.databind.deser.std.FromStringDeserializer$Std#_deserialize"}},
		{"CVE-2026-54514", db, "2.3.2", []string{"com.fasterxml.jackson.databind.deser.std.InetSocketAddressDeserializer#_deserialize"}},
		{"CVE-2026-54514", db, "2.2.0", nil},
		{"CVE-2026-54515", db, "2.11.4", []string{"com.fasterxml.jackson.databind.deser.BeanDeserializerBase#createContextual"}},
		{"CVE-2025-11226", lb, "1.2.3", []string{"ch.qos.logback.core.joran.conditional.IfAction#begin"}},
		{"CVE-2025-11226", lb, "1.5.18", []string{"ch.qos.logback.core.model.processor.conditional.IfModelHandler#handle"}},
		{"CVE-2025-11226", lb, "1.3.16", nil},
		{"CVE-2026-1225", lb, "1.2.3", []string{"ch.qos.logback.core.joran.action.NestedComplexPropertyIA#begin"}},
		{"CVE-2026-1225", lb, "1.5.24", []string{"ch.qos.logback.core.model.processor.ImplicitModelHandler#doComplex"}},
		{"CVE-2026-10532", lb, "1.2.3", []string{"ch.qos.logback.core.net.HardenedObjectInputStream#<init>"}},
		{"CVE-2026-10532", lb, "1.5.34", nil},
		{"CVE-2021-3597", ut, "2.2.4.Final", []string{"io.undertow.server.protocol.framed.AbstractFramedStreamSinkChannel#flushComplete"}},
		{"CVE-2021-3597", ut, "2.2.9.Final", nil},
		{"CVE-2021-3629", ut, "2.2.4.Final", []string{"io.undertow.protocols.http2.Http2Channel#updateSettings"}},
		{"CVE-2026-3260", ut, "2.2.4.Final", []string{"io.undertow.server.handlers.form.MultiPartParserDefinition$MultiPartUploadHandler#parseBlocking"}},
		{"CVE-2020-14340", "org.jboss.xnio:xnio-nio", "3.8.0.Final", []string{"org.xnio.nio.NioXnio#getSelector"}},
		{"CVE-2020-14340", "org.jboss.xnio:xnio-nio", "3.8.2.Final", nil},
	} {
		if got := pick(c.id, c.pkg, c.v); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s %s %s = %v, want %v", c.id, c.pkg, c.v, got, c.want)
		}
	}
}
