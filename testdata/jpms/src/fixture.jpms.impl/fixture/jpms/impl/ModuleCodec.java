package fixture.jpms.impl;

import fixture.jpms.api.Codec;

// Declared both in module-info `provides` and in META-INF/services.
public class ModuleCodec implements Codec {
    public ModuleCodec() {}
    public String encode(String s) { return "module:" + s; }
}
