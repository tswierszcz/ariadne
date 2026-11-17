package lib.api;

import org.apache.commons.text.StringSubstitutor;
import org.apache.commons.text.TextStringBuilder;

/**
 * (i) Extends a dependency class and overrides the protected step every
 * inherited replace calls on this: it never substitutes, so on a
 * StrictExpander the inherited replace overloads never reach a lookup.
 */
public class StrictExpander extends StringSubstitutor {
    @Override
    protected boolean substitute(TextStringBuilder buf, int offset, int length) {
        return false;
    }
}
