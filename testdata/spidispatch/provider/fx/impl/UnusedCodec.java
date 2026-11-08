package fx.impl;

import fx.trace.Ran;
import javax.ariadnefx.Codec;

/** In the provider's jar but never created: a target the model keeps (jar-bounded, not allocation-aware). */
public final class UnusedCodec implements Codec {
    public String encode(String s) {
        Ran.here();
        return s;
    }
}
