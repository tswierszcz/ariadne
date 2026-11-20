package mn.internal;

import org.apache.commons.text.StringSubstitutor;

/** The library's only caller of the vulnerable StringSubstitutor#replace. */
final class Expand {
    private Expand() {}

    static String expand(String template) {
        return StringSubstitutor.createInterpolator().replace(template);
    }
}
