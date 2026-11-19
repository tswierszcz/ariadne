package sp.api;

import org.apache.commons.text.lookup.StringLookupFactory;

/** (aa) The library obtains the lookup itself. */
public class Constants {
    public String constant(String key) {
        return StringLookupFactory.INSTANCE.constantStringLookup().lookup(key);
    }
}
