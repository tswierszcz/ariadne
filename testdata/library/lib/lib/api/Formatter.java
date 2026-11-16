package lib.api;

import org.apache.commons.text.StringSubstitutor;

/** (b) The vulnerable call is in a private method nothing calls. */
public class Formatter {
    public String format(String s) {
        return s.trim();
    }

    private String unsafe(String s) {
        return StringSubstitutor.createInterpolator().replace(s);
    }
}
