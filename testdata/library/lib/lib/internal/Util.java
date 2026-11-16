package lib.internal;

import org.apache.commons.text.StringSubstitutor;

/**
 * (f) Public by the JVM's rules, "internal" by convention: API unless the
 * user excludes lib.internal.
 */
public class Util {
    public static String expand(String s) {
        return StringSubstitutor.createInterpolator().replace(s);
    }
}
