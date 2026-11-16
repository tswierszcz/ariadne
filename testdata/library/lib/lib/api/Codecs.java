package lib.api;

import java.util.ServiceLoader;

/** (e) Public API that runs every registered Codec. */
public class Codecs {
    public static String decodeAll(String s) {
        for (Codec c : ServiceLoader.load(Codec.class)) {
            s = c.decode(s);
        }
        return s;
    }
}
