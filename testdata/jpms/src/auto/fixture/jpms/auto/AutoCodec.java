package fixture.jpms.auto;

import fixture.jpms.api.Codec;

// Packaged without module-info: an automatic module on the module path,
// whose META-INF/services entries are honored there too.
public class AutoCodec implements Codec {
    public AutoCodec() {}
    public String encode(String s) { return "auto:" + s; }
}
