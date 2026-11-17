package lib.api;

import org.apache.commons.text.StringSubstitutor;

/** (b) The vulnerable call is in a private method nothing calls. */
public final class Formatter {
    public String format(String s) {
        return s.trim();
    }

    private String unsafe(String s) {
        return StringSubstitutor.createInterpolator().replace(s);
    }

    /** Protected in a final class: no client can call it. */
    protected String hook(String s) {
        return s;
    }

    /** A private member class: its public method is not API. */
    private static class Cache {
        public String get(String s) {
            return s;
        }
    }
}
