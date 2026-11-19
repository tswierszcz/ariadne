package sp.supplied;

import org.apache.commons.text.lookup.StringLookup;

/** (y) The caller supplies the lookup as a parameter. */
public class Resolver {
    public String resolve(StringLookup lookup, String key) {
        return lookup.lookup(key);
    }
}
