package fixture.jpms.impl;

import fixture.jpms.api.Codec;

// Declared only in META-INF/services; an explicit module's
// META-INF/services is ignored on the module path.
public class LegacyCodec implements Codec {
    public LegacyCodec() {
        System.out.println("ORACLE-PROVIDER fixture.jpms.impl.LegacyCodec");
    }
    public String encode(String s) { return "legacy:" + s; }
}
