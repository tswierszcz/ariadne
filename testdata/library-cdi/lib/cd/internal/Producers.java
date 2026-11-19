package cd.internal;

import cd.api.Banner;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.inject.Produces;

/** (t) A package-private producer method of a package-private bean. */
@ApplicationScoped
class Producers {
    @Produces
    Banner banner() {
        return new Banner(Expand.expand("${java:version}"));
    }
}
