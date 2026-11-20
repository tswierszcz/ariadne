package mn.internal;

import jakarta.annotation.PostConstruct;
import jakarta.inject.Singleton;
import mn.api.Store;

/** (v) a @PostConstruct the container runs when a client uses a Store. */
@Singleton
class CachedStore implements Store {
    private String warmed = "";

    @PostConstruct
    void warm() {
        warmed = Expand.expand("warm ${java:vm}");
    }

    @Override
    public String get(String key) {
        return key + warmed.length();
    }
}
