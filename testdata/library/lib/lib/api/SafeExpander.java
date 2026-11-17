package lib.api;

import org.apache.commons.text.StringSubstitutor;

/** (h) Extends a dependency class and overrides replace(String): that one is not inherited. */
public class SafeExpander extends StringSubstitutor {
    @Override
    public String replace(String source) {
        return source;
    }
}
