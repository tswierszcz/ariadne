package sp.api;

import org.apache.commons.text.lookup.StringLookupFactory;
import sp.supplied.Resolver;

/**
 * (dd) The library calls a root, Resolver#resolve, with its own lookup:
 * the parameter is caller-supplied only when a caller enters at resolve.
 */
public class Wrapper {
    public String wrap(String key) {
        return new Resolver().resolve(StringLookupFactory.INSTANCE.constantStringLookup(), key);
    }
}
