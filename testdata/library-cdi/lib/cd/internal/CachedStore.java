package cd.internal;

import cd.api.Store;
import jakarta.annotation.PostConstruct;
import jakarta.enterprise.context.ApplicationScoped;

/** (v) A package-private @PostConstruct method. */
@ApplicationScoped
class CachedStore implements Store {
    private String value = "";

    @PostConstruct
    void warm() {
        value = Expand.expand("${java:vm}");
    }

    @Override
    public String get() {
        return value;
    }
}
