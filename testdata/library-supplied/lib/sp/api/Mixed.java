package sp.api;

import org.apache.commons.text.lookup.StringLookup;
import org.apache.commons.text.lookup.StringLookupFactory;

/** (bb) The field holds either the caller's lookup or the library's own default. */
public class Mixed {
    private final StringLookup lookup;

    public Mixed(StringLookup lookup) {
        this.lookup = lookup;
    }

    public Mixed() {
        this.lookup = StringLookupFactory.INSTANCE.constantStringLookup();
    }

    public String get(String key) {
        return lookup.lookup(key);
    }
}
