package mn.internal;

import jakarta.annotation.PostConstruct;

/** (x) no bean annotation: Micronaut generates no definition, nothing runs it. */
class Unmanaged {
    @PostConstruct
    void never() {
        Expand.expand("never ${java:hardware}");
    }
}
