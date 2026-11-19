package cd.internal;

import org.apache.commons.text.StringSubstitutor;

/** Package-private helper: the only caller of the vulnerable replace. */
final class Expand {
    private Expand() {}
    static String expand(String s) {
        return StringSubstitutor.createInterpolator().replace(s);
    }
}
