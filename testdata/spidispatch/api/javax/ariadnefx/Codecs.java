package javax.ariadnefx;

import java.util.ServiceLoader;
import javax.ariadnefx.spi.CodecProvider;

/** The API's bootstrap: finds the registered provider and asks it for a codec. */
public final class Codecs {
    public static Codec open() {
        CodecProvider p = ServiceLoader.load(CodecProvider.class).findFirst().orElseThrow();
        return p.create();
    }
}
