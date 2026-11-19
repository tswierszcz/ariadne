package cd.internal;

import cd.api.Banner;
import jakarta.annotation.PostConstruct;
import jakarta.enterprise.inject.Produces;

/**
 * (x) No bean-defining annotation: in a bean archive with
 * bean-discovery-mode="annotated" this is not a bean, so CDI never runs
 * its methods.
 */
class Unmanaged {
    @PostConstruct
    void never() {
        Expand.expand("${java:locale}");
    }

    @Produces
    Banner neverProduced() {
        return new Banner(Expand.expand("${java:hardware}"));
    }
}
