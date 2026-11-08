package fx.other;

import fx.trace.Ran;
import javax.ariadnefx.Codec;
import javax.ariadnefx.spi.CodecProvider;

/** Implements the service but is not registered: never a target. */
public final class RogueProvider implements CodecProvider {
    public Codec create() {
        Ran.here();
        return new RogueCodec();
    }
}
