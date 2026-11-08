package fx.impl;

import fx.trace.Ran;
import java.util.Base64;
import javax.ariadnefx.Codec;

public final class Base64Codec implements Codec {
    public String encode(String s) {
        Ran.here();
        return Base64.getEncoder().encodeToString(s.getBytes());
    }
}
