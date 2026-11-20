package mn.internal;

import io.micronaut.context.annotation.Factory;
import jakarta.inject.Singleton;
import mn.api.Banner;

/** (t) a factory method the container runs when a client asks for a Banner. */
@Factory
class Producers {
    @Singleton
    Banner banner() {
        return new Banner(Expand.expand("banner ${java:version}"));
    }
}
