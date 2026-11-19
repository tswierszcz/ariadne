package cd.internal;

import cd.api.Startup;
import jakarta.enterprise.context.Dependent;
import jakarta.enterprise.event.Observes;

/** (u) A package-private observer method. */
@Dependent
class AuditObserver {
    void onStartup(@Observes Startup event) {
        Expand.expand("${java:os}");
    }
}
