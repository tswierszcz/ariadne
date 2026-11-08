package fx.impl;

import fx.trace.Ran;
import javax.ariadnefx.Codec;
import javax.ariadnefx.spi.CodecProvider;

/** The registered provider (META-INF/services/javax.ariadnefx.spi.CodecProvider). */
public final class Base64Provider implements CodecProvider {
    public Codec create() {
        Ran.here();
        return new Base64Codec();
    }
}
