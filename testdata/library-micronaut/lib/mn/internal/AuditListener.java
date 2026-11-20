package mn.internal;

import io.micronaut.runtime.event.annotation.EventListener;
import jakarta.inject.Singleton;
import mn.api.Startup;

/** (u) an event listener the container runs when a client publishes Startup. */
@Singleton
class AuditListener {
    @EventListener
    void onStartup(Startup event) {
        Expand.expand("audit ${java:os}");
    }
}
