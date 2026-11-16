package lib.api;

import org.apache.commons.text.StringSubstitutor;

/** Package-private: reached from the public Templates#render. */
class Engine {
    static String expand(String template) {
        return StringSubstitutor.createInterpolator().replace(template);
    }
}
