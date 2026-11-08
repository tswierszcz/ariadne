package fx.other;

import fx.trace.Ran;
import javax.ariadnefx.Codec;

/** Implements the product in a jar with no registered provider: never a target. */
public final class RogueCodec implements Codec {
    public String encode(String s) {
        Ran.here();
        return "rogue";
    }
}
