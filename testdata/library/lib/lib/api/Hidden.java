package lib.api;

import org.apache.commons.text.StringSubstitutor;

/** (c) A package-private class with a public method; no API method uses it. */
class Hidden {
    public String run(String s) {
        return StringSubstitutor.createInterpolator().replace(s);
    }
}
