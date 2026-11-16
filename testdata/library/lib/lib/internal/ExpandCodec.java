package lib.internal;

import lib.api.Codec;
import org.apache.commons.text.StringSubstitutor;

/** (e) The library's own provider, registered in META-INF/services. */
public class ExpandCodec implements Codec {
    public String decode(String s) {
        return StringSubstitutor.createInterpolator().replace(s);
    }
}
