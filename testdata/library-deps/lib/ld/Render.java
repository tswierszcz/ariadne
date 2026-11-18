package ld;

import org.apache.commons.text.StringSubstitutor;

/** The control library's one API method (LIBRARY_MODE LB3): expands ${...} lookups. */
public final class Render {
    private Render() {}

    public static String render(String template) {
        return StringSubstitutor.createInterpolator().replace(template);
    }
}
