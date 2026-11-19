package sp.api;

import org.apache.commons.text.StringSubstitutor;
import org.apache.commons.text.lookup.StringLookup;

/**
 * (cc) The caller supplies the lookup; the library hands it to a
 * dependency object, which stores it and calls it later.
 */
public class Registrar {
    public String expand(StringLookup lookup, String template) {
        return new StringSubstitutor(lookup).replace(template);
    }
}
