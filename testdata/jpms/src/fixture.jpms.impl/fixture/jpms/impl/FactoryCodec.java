package fixture.jpms.impl;

import fixture.jpms.api.Codec;

// Declared only in module-info `provides`; on the module path ServiceLoader
// calls the static provider() method instead of a constructor.
public class FactoryCodec implements Codec {
    private FactoryCodec() {}
    public static Codec provider() {
        System.out.println("ORACLE-PROVIDER fixture.jpms.impl.FactoryCodec");
        return new FactoryCodec();
    }
    public String encode(String s) { return "factory:" + s; }
}
