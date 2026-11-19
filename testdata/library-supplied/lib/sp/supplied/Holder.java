package sp.supplied;

import org.apache.commons.text.lookup.StringLookup;

/** (z) The caller supplies the lookup through a setter; a later call uses the field. */
public class Holder {
    private StringLookup lookup;

    public void setLookup(StringLookup lookup) {
        this.lookup = lookup;
    }

    public String get(String key) {
        return lookup.lookup(key);
    }
}
